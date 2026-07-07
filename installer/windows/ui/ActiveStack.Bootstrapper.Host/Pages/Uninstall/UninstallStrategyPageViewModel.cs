using System.Collections.ObjectModel;
using ActiveStack.Bootstrapper.Core;
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

    public UninstallStrategyPageViewModel(UninstallOptions options, UninstallSelection selection, IInstallerEngineClient engineClient)
        : base("Choose uninstall strategy", "Pick whether to remove everything or restore a previous backup.")
    {
        _selection = selection;
        _engineClient = engineClient;
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
