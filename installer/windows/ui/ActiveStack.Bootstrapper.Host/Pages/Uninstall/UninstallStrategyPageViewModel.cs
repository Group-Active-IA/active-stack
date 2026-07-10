using System.Collections.ObjectModel;
using ActiveStack.Bootstrapper.Core;
using ActiveStack.Bootstrapper.Core.Localization;
using ActiveStack.Bootstrapper.Host.Navigation;

namespace ActiveStack.Bootstrapper.Host.Pages.Uninstall;

/// <summary>
/// Third Uninstall-flow page: single-select radio group over the uninstall
/// options' strategies, preselecting the strategy marked default
/// ("targeted"). When the selected strategy requires a manifest ("restore"),
/// reveals a backup selector fed by <see cref="IInstallerEngineClient.ListBackupsAsync"/>
/// and gates advancing until a backup is chosen.
/// </summary>
public sealed class UninstallStrategyPageViewModel : WizardPageViewModelBase
{
    private readonly UninstallSelection _selection;
    private readonly IInstallerEngineClient _engineClient;
    private string _selectedStrategyId;

    public UninstallStrategyPageViewModel(UninstallOptions options, UninstallSelection selection, IInstallerEngineClient engineClient, string lang = "en")
        : base(UiStrings.Get(lang, "page.uninstallstrategy.title"), UiStrings.Get(lang, "page.uninstallstrategy.subtitle"), lang)
    {
        _selection = selection;
        _engineClient = engineClient;
        ChooseBackupHeading = UiStrings.Get(lang, "template.choosebackuptorestore");
        Strategies = new ObservableCollection<UninstallStrategyChoice>(options.Strategies);

        var hasExistingSelection = !string.IsNullOrEmpty(selection.Strategy) &&
            Strategies.Any(s => string.Equals(s.Id, selection.Strategy, StringComparison.OrdinalIgnoreCase));

        var defaultStrategy = Strategies.FirstOrDefault(static s => s.IsDefault) ?? Strategies.FirstOrDefault();

        _selectedStrategyId = hasExistingSelection ? selection.Strategy : defaultStrategy?.Id ?? string.Empty;
        _selection.Strategy = _selectedStrategyId;
        _selection.RequiresManifest = Strategies.FirstOrDefault(s => string.Equals(s.Id, _selectedStrategyId, StringComparison.OrdinalIgnoreCase))?.RequiresManifest ?? false;
    }

    public ObservableCollection<UninstallStrategyChoice> Strategies { get; }

    public ObservableCollection<BackupEntry> Backups { get; } = [];

    /// <summary>Localized "Choose a backup to restore" section heading (bound by the template).</summary>
    public string ChooseBackupHeading { get; }

    public string SelectedStrategyId
    {
        get => _selectedStrategyId;
        set => _ = SelectStrategyAsync(value);
    }

    public BackupEntry? SelectedBackup
    {
        get => _selection.SelectedBackup;
        set
        {
            if (Equals(_selection.SelectedBackup, value))
            {
                return;
            }

            _selection.SelectedBackup = value;
            RaisePropertyChanged(nameof(SelectedBackup));
            RaiseCanAdvanceChanged();
        }
    }

    public bool ShowBackupSelector => _selection.RequiresManifest;

    /// <summary>The selected strategy's label, for the shared detail panel.</summary>
    public string DetailTitle => SelectedStrategy?.Label ?? string.Empty;

    /// <summary>The selected strategy's long description, falling back to its short description (D5, design.md).</summary>
    public string DetailBody => !string.IsNullOrEmpty(SelectedStrategy?.LongDescription)
        ? SelectedStrategy!.LongDescription
        : SelectedStrategy?.Description ?? string.Empty;

    private UninstallStrategyChoice? SelectedStrategy => Strategies.FirstOrDefault(s => string.Equals(s.Id, _selectedStrategyId, StringComparison.Ordinal));

    public override bool CanAdvance => !_selection.RequiresManifest || _selection.SelectedBackup is not null;

    public async Task SelectStrategyAsync(string strategyId, CancellationToken cancellationToken = default)
    {
        if (string.Equals(_selectedStrategyId, strategyId, StringComparison.Ordinal))
        {
            return;
        }

        var choice = Strategies.FirstOrDefault(s => string.Equals(s.Id, strategyId, StringComparison.OrdinalIgnoreCase));

        _selectedStrategyId = strategyId;
        _selection.Strategy = strategyId;
        _selection.RequiresManifest = choice?.RequiresManifest ?? false;
        _selection.SelectedBackup = null;

        Backups.Clear();
        RaisePropertyChanged(nameof(SelectedStrategyId));
        RaisePropertyChanged(nameof(ShowBackupSelector));
        RaisePropertyChanged(nameof(SelectedBackup));
        RaisePropertyChanged(nameof(DetailTitle));
        RaisePropertyChanged(nameof(DetailBody));

        if (_selection.RequiresManifest)
        {
            var backups = await _engineClient.ListBackupsAsync(cancellationToken);
            foreach (var backup in backups)
            {
                Backups.Add(backup);
            }
        }

        RaiseCanAdvanceChanged();
    }
}
