using ActiveStack.Bootstrapper.Core;
using ActiveStack.Bootstrapper.Host.Navigation;

namespace ActiveStack.Bootstrapper.Host.Pages.Starters;

/// <summary>
/// Last Starter-flow page: a read-only summary of the chosen starter,
/// target directory, and agents. Its primary footer label is always
/// "Install". Advancing starts the install by calling
/// <see cref="IInstallerEngineClient.RunStarterInstallAsync"/> with the
/// starter id, project path (as a single value), and selected agents.
/// </summary>
public sealed class StarterReviewPageViewModel : WizardPageViewModelBase, IStreamTriggerPage
{
    private readonly StarterSelection _selection;
    private readonly IInstallerEngineClient _engineClient;

    public StarterReviewPageViewModel(IReadOnlyList<StarterChoice> starters, StarterSelection selection, IInstallerEngineClient engineClient)
        : base("Review your starter", "Confirm the selection before Active Stack scaffolds it.")
    {
        _selection = selection;
        _engineClient = engineClient;

        StarterSummary = starters
            .FirstOrDefault(s => string.Equals(s.Id, selection.StarterId, StringComparison.OrdinalIgnoreCase))?.Name
            ?? selection.StarterId;

        TargetSummary = selection.ProjectPath;
        AgentsSummary = string.Join(", ", selection.Agents);
    }

    public string StarterSummary { get; }

    public string TargetSummary { get; }

    public string AgentsSummary { get; }

    /// <summary>The shell's primary footer action label while Review is current.</summary>
    public string PrimaryLabel => "Install";

    public override bool CanAdvance => true;

    public IAsyncEnumerable<InstallProgressSnapshot> StartStarterInstall(CancellationToken cancellationToken = default) =>
        _engineClient.RunStarterInstallAsync(_selection.StarterId, _selection.ProjectPath, _selection.Agents, dryRun: false, cancellationToken);

    IAsyncEnumerable<InstallProgressSnapshot> IStreamTriggerPage.StartStream(CancellationToken cancellationToken) =>
        StartStarterInstall(cancellationToken);
}
