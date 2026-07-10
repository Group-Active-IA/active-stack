using ActiveStack.Bootstrapper.Core;
using ActiveStack.Bootstrapper.Core.Localization;
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

    public StarterReviewPageViewModel(IReadOnlyList<StarterChoice> starters, StarterSelection selection, IInstallerEngineClient engineClient, string lang = "en")
        : base(UiStrings.Get(lang, "page.starterreview.title"), UiStrings.Get(lang, "page.starterreview.subtitle"), lang)
    {
        _selection = selection;
        _engineClient = engineClient;

        StarterHeading = UiStrings.Get(lang, "starterreview.heading.starter");
        TargetFolderHeading = UiStrings.Get(lang, "starterreview.heading.targetfolder");
        AgentsHeading = UiStrings.Get(lang, "starterreview.heading.agents");
        PrimaryLabel = UiStrings.Get(lang, "shell.install");

        StarterSummary = starters
            .FirstOrDefault(s => string.Equals(s.Id, selection.StarterId, StringComparison.OrdinalIgnoreCase))?.Name
            ?? selection.StarterId;

        TargetSummary = selection.ProjectPath;
        AgentsSummary = string.Join(", ", selection.Agents);
    }

    public string StarterSummary { get; }

    public string TargetSummary { get; }

    public string AgentsSummary { get; }

    /// <summary>Localized section headings bound by the Starter Review template.</summary>
    public string StarterHeading { get; }

    public string TargetFolderHeading { get; }

    public string AgentsHeading { get; }

    /// <summary>The shell's primary footer action label while Review is current (localized "Install").</summary>
    public string PrimaryLabel { get; }

    public override bool CanAdvance => true;

    public IAsyncEnumerable<InstallProgressSnapshot> StartStarterInstall(CancellationToken cancellationToken = default) =>
        _engineClient.RunStarterInstallAsync(_selection.StarterId, _selection.ProjectPath, _selection.Agents, dryRun: false, cancellationToken);

    IAsyncEnumerable<InstallProgressSnapshot> IStreamTriggerPage.StartStream(CancellationToken cancellationToken) =>
        StartStarterInstall(cancellationToken);
}
