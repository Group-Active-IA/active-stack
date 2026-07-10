using System.ComponentModel;
using ActiveStack.Bootstrapper.Core;
using ActiveStack.Bootstrapper.Core.Localization;
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
    private readonly Action<string> _persistLanguage;
    private readonly InstallSelection _selection = new();
    private readonly UninstallSelection _uninstallSelection = new();
    private readonly StarterSelection _starterSelection = new();
    private InstallerSessionState? _session;
    private UninstallOptions? _uninstallOptions;
    private IReadOnlyList<StarterChoice>? _starters;
    private WizardOperation _operation = WizardOperation.None;
    private WizardPageId _currentPageId = WizardPageId.Language;
    private IWizardPage _currentPage;
    private string _language;

    /// <summary>
    /// <paramref name="initialLanguage"/> is the already-resolved preselected
    /// language (persisted preference &gt; OS UI culture &gt; English —
    /// <see cref="LanguagePreselector"/> — or the <c>ACTIVE_STACK_UI_LANG</c>
    /// override), resolved once by the composition root so this view-model
    /// stays free of OS/filesystem reads and fully unit-testable.
    /// <paramref name="persistLanguage"/> defaults to
    /// <see cref="LanguagePreference.Save"/> against the real user profile;
    /// tests inject a spy instead of touching disk (gui-language-page, L4).
    /// </summary>
    public ShellViewModel(
        IInstallerEngineClient engineClient,
        IFolderPicker? folderPicker = null,
        string initialLanguage = "en",
        Action<string>? persistLanguage = null)
    {
        _engineClient = engineClient;
        _folderPicker = folderPicker ?? new WpfFolderPicker();
        _persistLanguage = persistLanguage ?? DefaultPersistLanguage;
        _language = initialLanguage;
        _engineClient.Language = _language;
        _currentPage = new LanguagePageViewModel(_language);
        _currentPage.PropertyChanged += OnCurrentPagePropertyChanged;
    }

    private static void DefaultPersistLanguage(string lang) =>
        LanguagePreference.Save(Environment.GetFolderPath(Environment.SpecialFolder.UserProfile), lang);

    public event PropertyChangedEventHandler? PropertyChanged;

    /// <summary>
    /// Raised when the primary action is invoked on the terminal Complete
    /// page (D2a, design.md): the wizard is done and the host should close
    /// its window. <see cref="ShellViewModel"/> deliberately does not take a
    /// dependency on the WPF <c>Window</c> — <c>MainWindow</c> subscribes and
    /// calls <c>Close()</c>, consistent with the existing <see cref="IFolderPicker"/>
    /// / injected <c>persistLanguage</c> seam pattern.
    /// </summary>
    public event EventHandler? CloseRequested;

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

    /// <summary>Back is unavailable on the Language page (the wizard's new entry) and on the terminal, footer-less Installing/Complete pages.</summary>
    public bool CanGoBack =>
        _currentPageId is not (WizardPageId.Language or WizardPageId.Installing or WizardPageId.Complete);

    /// <summary>
    /// The localized "Finish" label on the terminal Complete page, else the
    /// stream-trigger page's own label (Install/Uninstall/Install for the
    /// starter Review) while one is current, the localized "Next" everywhere
    /// else (D3, design.md; D9, design.md; gui-language-page L4).
    /// </summary>
    public string PrimaryLabel =>
        OnCompletePage
            ? UiStrings.Get(_language, "shell.finish")
            : (CurrentPage as IStreamTriggerPage)?.PrimaryLabel ?? UiStrings.Get(_language, "shell.next");

    /// <summary>Localized footer "Back" label bound by MainWindow.xaml.</summary>
    public string BackLabel => UiStrings.Get(_language, "shell.back");

    /// <summary>Localized footer "Cancel" label bound by MainWindow.xaml.</summary>
    public string CancelLabel => UiStrings.Get(_language, "shell.cancel");

    public bool PrimaryEnabled => CurrentPage.CanAdvance;

    /// <summary>Shared by <see cref="PrimaryLabel"/> and <see cref="AdvanceAsync"/> (D2a/D3, design.md).</summary>
    private bool OnCompletePage => CurrentPage is CompletePageViewModel;

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

        if (_currentPageId == WizardPageId.Language)
        {
            AdvanceFromLanguage();
            MoveTo(WizardPageId.Hub);
            return;
        }

        if (CurrentPage is IStreamTriggerPage triggerPage)
        {
            await RunStreamAndAdvanceToCompleteAsync(triggerPage, cancellationToken);
            return;
        }

        if (OnCompletePage)
        {
            // D2a, design.md: the Complete page's terminal action is Finish,
            // not "advance into the flow's navigation" — MoveTo(NextPage(Complete, …))
            // self-loops and CreatePage has no Complete case (it throws). Short-circuit
            // before ever reaching the operation switch below.
            CloseRequested?.Invoke(this, EventArgs.Empty);
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
            _ => WizardPageId.Language
        };

        if (previous == WizardPageId.Hub)
        {
            // Returning to the Hub means a different entry can be chosen next (D2).
            _operation = WizardOperation.None;
        }

        MoveTo(previous);
    }

    /// <summary>
    /// Advancing from the Language page (D-language, windows-gui-localization):
    /// sets the shell's active language, persists it, and configures the
    /// engine client's <see cref="IInstallerEngineClient.Language"/>. When
    /// the language actually changed from the previously active one,
    /// invalidates the cached session/uninstall-options/starters so the next
    /// flow re-fetches engine data in the new language; re-selecting the
    /// same language is a no-op cache-wise.
    /// </summary>
    private void AdvanceFromLanguage()
    {
        var languagePage = (LanguagePageViewModel)CurrentPage;
        var newLanguage = languagePage.SelectedLanguageId;
        var changed = !string.Equals(_language, newLanguage, StringComparison.Ordinal);

        _language = newLanguage;
        _persistLanguage(_language);
        _engineClient.Language = _language;

        if (changed)
        {
            _session = null;
            _uninstallOptions = null;
            _starters = null;

            RaisePropertyChanged(nameof(BackLabel));
            RaisePropertyChanged(nameof(CancelLabel));
            RaisePropertyChanged(nameof(PrimaryLabel));
        }
    }

    /// <summary>
    /// Generic "trigger page → stream → progress → complete" run (D9,
    /// design.md): works for Install's Review, Uninstall's Confirm, and
    /// Starter's Review alike since each implements
    /// <see cref="IStreamTriggerPage"/>.
    /// </summary>
    private async Task RunStreamAndAdvanceToCompleteAsync(IStreamTriggerPage triggerPage, CancellationToken cancellationToken)
    {
        var progressPage = new ProgressPageViewModel(_language);
        _currentPageId = WizardPageId.Installing;
        CurrentPage = progressPage;

        // D1, design.md: any exception thrown while consuming the stream
        // (a non-zero engine exit, a parse failure, or anything else) MUST
        // NOT propagate past the shell — it is converted into the existing
        // CompleteState.Error terminal page instead. Without this, the
        // exception crosses into the WPF dispatcher via the async void
        // advance handler and WPF terminates the process silently (Bug A/B).
        try
        {
            await progressPage.ConsumeAsync(triggerPage.StartStream(cancellationToken), cancellationToken);
        }
        catch (Exception ex)
        {
            _currentPageId = WizardPageId.Complete;
            CurrentPage = ToErrorPage(ex);
            return;
        }

        var terminalSnapshot = progressPage.TerminalSnapshot
            ?? new InstallProgressSnapshot("install_finished", null, null, null, progressPage.InstallSucceeded);

        _currentPageId = WizardPageId.Complete;
        CurrentPage = new CompletePageViewModel(terminalSnapshot, progressPage.HadDegradedSteps, progressPage.HadRollback, _language);
    }

    /// <summary>
    /// Builds the terminal Error page for a stream failure caught by
    /// <see cref="RunStreamAndAdvanceToCompleteAsync"/>: a synthesized
    /// failed snapshot with no rollback/degraded steps so
    /// <see cref="CompletePageViewModel.DetermineState"/> resolves to
    /// <see cref="CompleteState.Error"/>, carrying the exception's message.
    /// </summary>
    private CompletePageViewModel ToErrorPage(Exception exception) =>
        new(
            new InstallProgressSnapshot("install_finished", null, null, exception.Message, Success: false),
            hadDegradedSteps: false,
            hadRollback: false,
            _language);

    /// <summary>
    /// Backups (D7, design.md) is a standalone page, not a wizard step, and
    /// needs an async load before it can be shown — handled separately from
    /// the synchronous <see cref="MoveTo"/>/<see cref="CreatePage"/> path
    /// the other flows use.
    /// </summary>
    private async Task MoveToBackupsAsync(CancellationToken cancellationToken)
    {
        var backupsPage = new BackupsPageViewModel(_engineClient, _language);
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
        WizardPageId.Language => new LanguagePageViewModel(_language),
        WizardPageId.Hub => new HubPageViewModel(_language),
        WizardPageId.Assistants => new AssistantsPageViewModel(RequireSession(), _selection, _language),
        WizardPageId.InstallType => new InstallTypePageViewModel(RequireSession(), _selection, _language),
        WizardPageId.Components => new ComponentsPageViewModel(RequireSession(), _selection, _language),
        WizardPageId.Permissions => new PermissionsPageViewModel(RequireSession(), _selection, _language),
        WizardPageId.Review => new ReviewPageViewModel(RequireSession(), _selection, _engineClient, _language),
        WizardPageId.UninstallAgents => new UninstallAgentsPageViewModel(RequireUninstallOptions(), _uninstallSelection, _language),
        WizardPageId.UninstallMode => new UninstallModePageViewModel(RequireUninstallOptions(), _uninstallSelection, _language),
        WizardPageId.UninstallStrategy => new UninstallStrategyPageViewModel(RequireUninstallOptions(), _uninstallSelection, _engineClient, _language),
        WizardPageId.UninstallConfirm => new UninstallConfirmPageViewModel(RequireUninstallOptions(), _uninstallSelection, _engineClient, _language),
        WizardPageId.StarterCatalog => new StarterCatalogPageViewModel(RequireStarters(), _starterSelection, _language),
        WizardPageId.StarterTarget => new StarterTargetPageViewModel(_starterSelection, _folderPicker, _language),
        WizardPageId.StarterReview => new StarterReviewPageViewModel(RequireStarters(), _starterSelection, _engineClient, _language),
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
