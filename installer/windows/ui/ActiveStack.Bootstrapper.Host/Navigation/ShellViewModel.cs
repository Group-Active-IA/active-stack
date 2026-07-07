using System.ComponentModel;
using ActiveStack.Bootstrapper.Core;
using ActiveStack.Bootstrapper.Host.Pages;
using ActiveStack.Bootstrapper.Host.Pages.Backups;
using ActiveStack.Bootstrapper.Host.Pages.Install;
using ActiveStack.Bootstrapper.Host.Pages.Starters;
using ActiveStack.Bootstrapper.Host.Pages.Uninstall;

namespace ActiveStack.Bootstrapper.Host.Navigation;

/// <summary>
/// Owns the wizard's current page, footer state, and the active operation's
/// selection; drives navigation via the active <see cref="WizardOperation"/>'s
/// pure flow functions (<see cref="WizardFlow"/>, <see cref="UninstallFlow"/>,
/// <see cref="StarterFlow"/>) (D2/D3, design.md). Advancing from the Hub sets
/// the operation from <see cref="HubPageViewModel.SelectedEntryId"/>; Back to
/// the Hub resets it to <see cref="WizardOperation.None"/>. Each operation's
/// backing data (session / uninstall options / starter catalog) is loaded
/// lazily, once, the first time its flow advances past the Hub.
/// </summary>
public sealed class ShellViewModel : INotifyPropertyChanged
{
    private readonly IInstallerEngineClient _engineClient;
    private readonly IFolderPicker _folderPicker;
    private readonly InstallSelection _selection = new();
    private readonly UninstallSelection _uninstallSelection = new();
    private readonly StarterSelection _starterSelection = new();
    private InstallerSessionState? _session;
    private UninstallOptions? _uninstallOptions;
    private IReadOnlyList<StarterChoice>? _starters;
    private WizardOperation _operation = WizardOperation.None;
    private WizardPageId _currentPageId = WizardPageId.Hub;
    private IWizardPage _currentPage;

    public ShellViewModel(IInstallerEngineClient engineClient, IFolderPicker? folderPicker = null)
    {
        _engineClient = engineClient;
        _folderPicker = folderPicker ?? new WpfFolderPicker();
        _currentPage = new HubPageViewModel();
        _currentPage.PropertyChanged += OnCurrentPagePropertyChanged;
    }

    public event PropertyChangedEventHandler? PropertyChanged;

    public IWizardPage CurrentPage
    {
        get => _currentPage;
        private set
        {
            if (ReferenceEquals(_currentPage, value))
            {
                return;
            }

            _currentPage.PropertyChanged -= OnCurrentPagePropertyChanged;
            _currentPage = value;
            _currentPage.PropertyChanged += OnCurrentPagePropertyChanged;

            RaisePropertyChanged(nameof(CurrentPage));
            RaisePropertyChanged(nameof(CanGoBack));
            RaisePropertyChanged(nameof(PrimaryLabel));
            RaisePropertyChanged(nameof(PrimaryEnabled));
        }
    }

    /// <summary>Back is unavailable on the Hub (flow entry) and on the terminal, footer-less Installing/Complete pages.</summary>
    public bool CanGoBack =>
        _currentPageId is not (WizardPageId.Hub or WizardPageId.Installing or WizardPageId.Complete);

    /// <summary>
    /// The stream-trigger page's own label (Install/Uninstall/Install for
    /// the starter Review) while one is current, "Next" everywhere else
    /// (D9, design.md).
    /// </summary>
    public string PrimaryLabel => (CurrentPage as IStreamTriggerPage)?.PrimaryLabel ?? "Next";

    public bool PrimaryEnabled => CurrentPage.CanAdvance;

    /// <summary>
    /// True once the flow has reached Complete in a success or degraded
    /// (best-effort steps skipped) state. Used by the Burn entry point to
    /// pick its process exit code (D6/D7 constraint: Burn spawn/Process
    /// path unchanged) — false while still mid-flow or on error/rollback.
    /// </summary>
    public bool InstallSucceeded =>
        CurrentPage is CompletePageViewModel complete &&
        complete.State is CompleteState.Success or CompleteState.Degraded;

    public async Task AdvanceAsync(CancellationToken cancellationToken = default)
    {
        if (!CurrentPage.CanAdvance)
        {
            return;
        }

        if (CurrentPage is IStreamTriggerPage triggerPage)
        {
            await RunStreamAndAdvanceToCompleteAsync(triggerPage, cancellationToken);
            return;
        }

        if (_currentPageId == WizardPageId.Hub)
        {
            _operation = ResolveOperation(((HubPageViewModel)CurrentPage).SelectedEntryId);
        }

        switch (_operation)
        {
            case WizardOperation.Install:
                _session ??= await _engineClient.LoadSessionAsync(cancellationToken);
                MoveTo(WizardFlow.NextPage(_currentPageId, _selection, _session.TierCapableAgents));
                break;

            case WizardOperation.Uninstall:
                _uninstallOptions ??= await _engineClient.LoadUninstallOptionsAsync(cancellationToken);
                MoveTo(UninstallFlow.NextPage(_currentPageId, _uninstallSelection));
                break;

            case WizardOperation.Starter:
                _starters ??= await _engineClient.ListStartersAsync(cancellationToken);
                MoveTo(StarterFlow.NextPage(_currentPageId, _starterSelection));
                break;

            case WizardOperation.Backups:
                await MoveToBackupsAsync(cancellationToken);
                break;

            default:
                // No enabled entry selected on the Hub yet — nothing to advance to.
                break;
        }
    }

    public void GoBack()
    {
        if (!CanGoBack)
        {
            return;
        }

        var previous = _operation switch
        {
            WizardOperation.Install => WizardFlow.PreviousPage(_currentPageId, _selection, _session?.TierCapableAgents ?? []),
            WizardOperation.Uninstall => UninstallFlow.PreviousPage(_currentPageId, _uninstallSelection),
            WizardOperation.Starter => StarterFlow.PreviousPage(_currentPageId, _starterSelection),
            WizardOperation.Backups => WizardPageId.Hub,
            _ => WizardPageId.Hub
        };

        if (previous == WizardPageId.Hub)
        {
            // Returning to the Hub means a different entry can be chosen next (D2).
            _operation = WizardOperation.None;
        }

        MoveTo(previous);
    }

    /// <summary>
    /// Generic "trigger page → stream → progress → complete" run (D9,
    /// design.md): works for Install's Review, Uninstall's Confirm, and
    /// Starter's Review alike since each implements
    /// <see cref="IStreamTriggerPage"/>.
    /// </summary>
    private async Task RunStreamAndAdvanceToCompleteAsync(IStreamTriggerPage triggerPage, CancellationToken cancellationToken)
    {
        var progressPage = new ProgressPageViewModel();
        _currentPageId = WizardPageId.Installing;
        CurrentPage = progressPage;

        await progressPage.ConsumeAsync(triggerPage.StartStream(cancellationToken), cancellationToken);

        var terminalSnapshot = progressPage.TerminalSnapshot
            ?? new InstallProgressSnapshot("install_finished", null, null, null, progressPage.InstallSucceeded);

        _currentPageId = WizardPageId.Complete;
        CurrentPage = new CompletePageViewModel(terminalSnapshot, progressPage.HadDegradedSteps, progressPage.HadRollback);
    }

    /// <summary>
    /// Backups (D7, design.md) is a standalone page, not a wizard step, and
    /// needs an async load before it can be shown — handled separately from
    /// the synchronous <see cref="MoveTo"/>/<see cref="CreatePage"/> path
    /// the other flows use.
    /// </summary>
    private async Task MoveToBackupsAsync(CancellationToken cancellationToken)
    {
        var backupsPage = new BackupsPageViewModel(_engineClient);
        await backupsPage.RefreshAsync(cancellationToken);

        _currentPageId = WizardPageId.Backups;
        CurrentPage = backupsPage;
    }

    private void MoveTo(WizardPageId pageId)
    {
        _currentPageId = pageId;
        CurrentPage = CreatePage(pageId);
    }

    private IWizardPage CreatePage(WizardPageId pageId) => pageId switch
    {
        WizardPageId.Hub => new HubPageViewModel(),
        WizardPageId.Assistants => new AssistantsPageViewModel(RequireSession(), _selection),
        WizardPageId.InstallType => new InstallTypePageViewModel(RequireSession(), _selection),
        WizardPageId.Components => new ComponentsPageViewModel(RequireSession(), _selection),
        WizardPageId.Permissions => new PermissionsPageViewModel(RequireSession(), _selection),
        WizardPageId.Review => new ReviewPageViewModel(RequireSession(), _selection, _engineClient),
        WizardPageId.UninstallAgents => new UninstallAgentsPageViewModel(RequireUninstallOptions(), _uninstallSelection),
        WizardPageId.UninstallMode => new UninstallModePageViewModel(RequireUninstallOptions(), _uninstallSelection),
        WizardPageId.UninstallStrategy => new UninstallStrategyPageViewModel(RequireUninstallOptions(), _uninstallSelection, _engineClient),
        WizardPageId.UninstallConfirm => new UninstallConfirmPageViewModel(RequireUninstallOptions(), _uninstallSelection, _engineClient),
        WizardPageId.StarterCatalog => new StarterCatalogPageViewModel(RequireStarters(), _starterSelection),
        WizardPageId.StarterTarget => new StarterTargetPageViewModel(_starterSelection, _folderPicker),
        WizardPageId.StarterReview => new StarterReviewPageViewModel(RequireStarters(), _starterSelection, _engineClient),
        _ => throw new InvalidOperationException(
            $"ShellViewModel cannot construct a page for '{pageId}' via MoveTo; Installing/Complete/Backups are reached via their dedicated async paths.")
    };

    private static WizardOperation ResolveOperation(string? entryId) => entryId switch
    {
        "install" => WizardOperation.Install,
        "uninstall" => WizardOperation.Uninstall,
        "starters" => WizardOperation.Starter,
        "backups" => WizardOperation.Backups,
        _ => WizardOperation.None
    };

    private InstallerSessionState RequireSession() =>
        _session ?? throw new InvalidOperationException("Session must be loaded before navigating past the Hub.");

    private UninstallOptions RequireUninstallOptions() =>
        _uninstallOptions ?? throw new InvalidOperationException("Uninstall options must be loaded before navigating past the Hub.");

    private IReadOnlyList<StarterChoice> RequireStarters() =>
        _starters ?? throw new InvalidOperationException("Starters must be loaded before navigating past the Hub.");

    private void OnCurrentPagePropertyChanged(object? sender, PropertyChangedEventArgs e)
    {
        if (string.Equals(e.PropertyName, nameof(IWizardPage.CanAdvance), StringComparison.Ordinal))
        {
            RaisePropertyChanged(nameof(PrimaryEnabled));
        }
    }

    private void RaisePropertyChanged(string propertyName) =>
        PropertyChanged?.Invoke(this, new PropertyChangedEventArgs(propertyName));
}
