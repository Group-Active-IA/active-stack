using ActiveStack.Bootstrapper.Core;
using ActiveStack.Bootstrapper.Core.Localization;
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
    private readonly UninstallSelection _selection;
    private readonly IInstallerEngineClient _engineClient;

    public UninstallConfirmPageViewModel(UninstallOptions options, UninstallSelection selection, IInstallerEngineClient engineClient, string lang = "en")
        : base(UiStrings.Get(lang, "page.uninstallconfirm.title"), UiStrings.Get(lang, "page.uninstallconfirm.subtitle"), lang)
    {
        _selection = selection;
        _engineClient = engineClient;

        AgentsHeading = UiStrings.Get(lang, "uninstallreview.heading.agents");
        ModeHeading = UiStrings.Get(lang, "uninstallreview.heading.mode");
        StrategyHeading = UiStrings.Get(lang, "uninstallreview.heading.strategy");
        RestoringBackupHeading = UiStrings.Get(lang, "uninstallreview.heading.restoringbackup");
        WarningText = UiStrings.Get(lang, "uninstall.warning");
        PrimaryLabel = UiStrings.Get(lang, "shell.uninstall");

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

    /// <summary>Localized section headings bound by the Confirm page's template.</summary>
    public string AgentsHeading { get; }

    public string ModeHeading { get; }

    public string StrategyHeading { get; }

    public string RestoringBackupHeading { get; }

    /// <summary>Localized destructive warning ("This will modify your agent configuration.") from UiStrings.</summary>
    public string WarningText { get; }

    /// <summary>The shell's primary footer action label while Confirm is current (localized "Uninstall").</summary>
    public string PrimaryLabel { get; }

    public override bool CanAdvance => true;

    public IAsyncEnumerable<InstallProgressSnapshot> StartStream(CancellationToken cancellationToken = default) =>
        _engineClient.RunUninstallAsync(
            _selection.Agents,
            _selection.Mode,
            _selection.Strategy,
            string.Equals(_selection.Strategy, "restore", StringComparison.OrdinalIgnoreCase) ? _selection.SelectedBackup?.ManifestPath : null,
            cancellationToken);
}
