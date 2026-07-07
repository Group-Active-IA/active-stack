using ActiveStack.Bootstrapper.Core;
using ActiveStack.Bootstrapper.Host.Navigation;

namespace ActiveStack.Bootstrapper.Host.Pages.Install;

/// <summary>
/// Last Install-flow page: a read-only summary of the shared
/// <see cref="InstallSelection"/>. Its primary footer label is always
/// "Install" (D4, design.md — <see cref="ShellViewModel"/> reads
/// <see cref="PrimaryLabel"/> off the current page). Confirming starts the
/// install by calling <see cref="IInstallerEngineClient.RunInstallAsync"/>
/// with the full selection; the returned stream is handed to a
/// <see cref="ProgressPageViewModel"/> by the shell.
/// </summary>
public sealed class ReviewPageViewModel : WizardPageViewModelBase, IStreamTriggerPage
{
    private readonly InstallSelection _selection;
    private readonly IInstallerEngineClient _engineClient;

    public ReviewPageViewModel(InstallerSessionState session, InstallSelection selection, IInstallerEngineClient engineClient)
        : base("Review your setup", "Confirm the selection before Active Stack installs it.")
    {
        _selection = selection;
        _engineClient = engineClient;

        AgentsSummary = string.Join(", ", session.AssistantChoices
            .Where(a => _selection.Agents.Contains(a.Id, StringComparer.OrdinalIgnoreCase))
            .Select(a => a.Label));

        ModeSummary = session.InstallTypeChoices
            .FirstOrDefault(m => string.Equals(m.Id, _selection.Mode, StringComparison.OrdinalIgnoreCase))?.Label
            ?? _selection.Mode;

        var allComponents = session.ForcedComponents.Concat(session.CustomComponents);
        ComponentsSummary = string.Join(", ", allComponents
            .Where(c => _selection.CustomIds.Contains(c.Id, StringComparer.OrdinalIgnoreCase))
            .Select(c => c.Label));

        TierSummary = session.PermissionTierChoices
            .FirstOrDefault(t => string.Equals(t.Id, _selection.Tier, StringComparison.OrdinalIgnoreCase))?.Label
            ?? _selection.Tier
            ?? "Not applicable";
    }

    public string AgentsSummary { get; }

    public string ModeSummary { get; }

    public string ComponentsSummary { get; }

    public string TierSummary { get; }

    /// <summary>The shell's primary footer action label while Review is current.</summary>
    public string PrimaryLabel => "Install";

    public override bool CanAdvance => true;

    public IAsyncEnumerable<InstallProgressSnapshot> StartInstall(CancellationToken cancellationToken = default) =>
        _engineClient.RunInstallAsync(_selection.Agents, _selection.Mode, _selection.CustomIds, _selection.Tier, cancellationToken);

    IAsyncEnumerable<InstallProgressSnapshot> IStreamTriggerPage.StartStream(CancellationToken cancellationToken) =>
        StartInstall(cancellationToken);
}
