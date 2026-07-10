using ActiveStack.Bootstrapper.Core;
using ActiveStack.Bootstrapper.Core.Localization;
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

    public ReviewPageViewModel(InstallerSessionState session, InstallSelection selection, IInstallerEngineClient engineClient, string lang = "en")
        : base(UiStrings.Get(lang, "page.review.title"), UiStrings.Get(lang, "page.review.subtitle"), lang)
    {
        _selection = selection;
        _engineClient = engineClient;

        AssistantsHeading = UiStrings.Get(lang, "review.heading.assistants");
        InstallTypeHeading = UiStrings.Get(lang, "review.heading.installtype");
        ComponentsHeading = UiStrings.Get(lang, "review.heading.components");
        PermissionTierHeading = UiStrings.Get(lang, "review.heading.permissiontier");
        PrimaryLabel = UiStrings.Get(lang, "shell.install");

        AgentsSummary = string.Join(", ", session.AssistantChoices
            .Where(a => _selection.Agents.Contains(a.Id, StringComparer.OrdinalIgnoreCase))
            .Select(a => a.Label));

        ModeSummary = session.InstallTypeChoices
            .FirstOrDefault(m => string.Equals(m.Id, _selection.Mode, StringComparison.OrdinalIgnoreCase))?.Label
            ?? _selection.Mode;

        ComponentsSummary = BuildComponentsSummary(session, selection);

        TierSummary = session.PermissionTierChoices
            .FirstOrDefault(t => string.Equals(t.Id, _selection.Tier, StringComparison.OrdinalIgnoreCase))?.Label
            ?? _selection.Tier
            ?? "Not applicable";
    }

    public string AgentsSummary { get; }

    public string ModeSummary { get; }

    public string ComponentsSummary { get; }

    public string TierSummary { get; }

    /// <summary>Localized section headings bound by the Review page's template.</summary>
    public string AssistantsHeading { get; }

    public string InstallTypeHeading { get; }

    public string ComponentsHeading { get; }

    public string PermissionTierHeading { get; }

    /// <summary>The shell's primary footer action label while Review is current (localized "Install").</summary>
    public string PrimaryLabel { get; }

    public override bool CanAdvance => true;

    /// <summary>
    /// Mode-aware components summary (D2, design.md). In <c>custom</c> mode the
    /// set is authoritative from <see cref="InstallSelection.CustomIds"/>
    /// (populated by the Components page). In a non-custom mode (lite/full) the
    /// Components page is skipped and <c>CustomIds</c> stays legitimately empty,
    /// so the set is derived instead from the session's forced components plus
    /// the custom components the session marks as recommended — the mode's
    /// implied default set.
    /// </summary>
    private static string BuildComponentsSummary(InstallerSessionState session, InstallSelection selection)
    {
        var selectedComponents = string.Equals(selection.Mode, "custom", StringComparison.OrdinalIgnoreCase)
            ? session.ForcedComponents.Concat(session.CustomComponents)
                .Where(c => selection.CustomIds.Contains(c.Id, StringComparer.OrdinalIgnoreCase))
            : session.ForcedComponents.Concat(session.CustomComponents.Where(c => c.Recommended));
        return string.Join(", ", selectedComponents.Select(c => c.Label));
    }

    public IAsyncEnumerable<InstallProgressSnapshot> StartInstall(CancellationToken cancellationToken = default) =>
        _engineClient.RunInstallAsync(_selection.Agents, _selection.Mode, _selection.CustomIds, _selection.Tier, cancellationToken);

    IAsyncEnumerable<InstallProgressSnapshot> IStreamTriggerPage.StartStream(CancellationToken cancellationToken) =>
        StartInstall(cancellationToken);
}
