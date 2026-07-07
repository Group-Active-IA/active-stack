using ActiveStack.Bootstrapper.Core;
using ActiveStack.Bootstrapper.Host.Navigation;

namespace ActiveStack.Bootstrapper.Host.Pages.Uninstall;

/// <summary>
/// Last Uninstall-flow page: a read-only summary of the shared
/// <see cref="UninstallSelection"/> plus the destructive warning matching
/// the TUI ("This will modify your agent configuration."). Advancing starts
/// the uninstall by calling <see cref="IInstallerEngineClient.RunUninstallAsync"/>
/// with the selected agents, mode, strategy and — for "restore" — the chosen
/// backup's manifest path threaded through the client's <c>backupId</c>
/// parameter (D5, design.md).
/// </summary>
public sealed class UninstallConfirmPageViewModel : WizardPageViewModelBase, IStreamTriggerPage
{
    private const string ModificationWarning = "This will modify your agent configuration.";

    private readonly UninstallSelection _selection;
    private readonly IInstallerEngineClient _engineClient;

    public UninstallConfirmPageViewModel(UninstallOptions options, UninstallSelection selection, IInstallerEngineClient engineClient)
        : base("Review the uninstall", "Confirm the selection before Active Stack removes it.")
    {
        _selection = selection;
        _engineClient = engineClient;

        AgentsSummary = string.Join(", ", selection.Agents);

        ModeSummary = options.Modes
            .FirstOrDefault(m => string.Equals(m.Id, selection.Mode, StringComparison.OrdinalIgnoreCase))?.Label
            ?? selection.Mode;

        StrategySummary = options.Strategies
            .FirstOrDefault(s => string.Equals(s.Id, selection.Strategy, StringComparison.OrdinalIgnoreCase))?.Label
            ?? selection.Strategy;

        BackupSummary = selection.Strategy.Equals("restore", StringComparison.OrdinalIgnoreCase)
            ? selection.SelectedBackup?.DisplayLabel
            : null;
    }

    public string AgentsSummary { get; }

    public string ModeSummary { get; }

    public string StrategySummary { get; }

    public string? BackupSummary { get; }

    public string WarningText => ModificationWarning;

    /// <summary>The shell's primary footer action label while Confirm is current.</summary>
    public string PrimaryLabel => "Uninstall";

    public override bool CanAdvance => true;

    public IAsyncEnumerable<InstallProgressSnapshot> StartStream(CancellationToken cancellationToken = default) =>
        _engineClient.RunUninstallAsync(
            _selection.Agents,
            _selection.Mode,
            _selection.Strategy,
            string.Equals(_selection.Strategy, "restore", StringComparison.OrdinalIgnoreCase) ? _selection.SelectedBackup?.ManifestPath : null,
            cancellationToken);
}
