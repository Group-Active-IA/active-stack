using ActiveStack.Bootstrapper.Core;
using ActiveStack.Bootstrapper.Host.Navigation;
using ActiveStack.Bootstrapper.Host.Pages.Install;
using Xunit;

namespace ActiveStack.Bootstrapper.Host.Tests;

/// <summary>
/// Bug D diagnosis (D4, design.md): the Components -&gt; Review data flow
/// (<see cref="ComponentsPageViewModel.SyncSelection"/> writes into the
/// shared <see cref="InstallSelection.CustomIds"/>, and
/// <see cref="ReviewPageViewModel"/> reads it back) reads correctly wired
/// from static analysis, but no test exercised it end-to-end — every
/// existing fixture faked <c>InstallSelection.CustomIds</c> independently
/// at each page's construction. This is the missing integration test: a
/// real <see cref="ComponentsPageViewModel"/> and a real
/// <see cref="ReviewPageViewModel"/> constructed over the SAME
/// <see cref="InstallerSessionState"/> and <see cref="InstallSelection"/>,
/// the way <see cref="ShellViewModel"/> actually wires them.
/// </summary>
public sealed class ComponentsToReviewIntegrationTests
{
    [Fact]
    public void TogglingComponentsThenBuildingReview_OverTheSameSelection_SummaryListsExactlyTheToggledComponents()
    {
        var session = BuildSession();
        var selection = new InstallSelection { Agents = ["claude"], Mode = "custom" };

        var componentsPage = new ComponentsPageViewModel(session, selection);

        // permissions is forced-on and stays on; toggle "openspec" on and
        // leave "extra" off (it is not recommended by default); toggle
        // "engram" on explicitly.
        var openspec = Assert.Single(componentsPage.Choices, c => c.Id == "openspec");
        openspec.IsSelected = true;
        var engram = Assert.Single(componentsPage.Choices, c => c.Id == "engram");
        engram.IsSelected = true;
        var extra = Assert.Single(componentsPage.Choices, c => c.Id == "extra");
        extra.IsSelected = false;

        var reviewPage = new ReviewPageViewModel(session, selection, new NotSupportedInstallerEngineClient());

        Assert.Contains("Basic protection", reviewPage.ComponentsSummary);
        Assert.Contains("OpenSpec", reviewPage.ComponentsSummary);
        Assert.Contains("Engram", reviewPage.ComponentsSummary);
        Assert.DoesNotContain("Extra", reviewPage.ComponentsSummary);
    }

    [Fact]
    public void LeavingAllCustomComponentsAtTheirRecommendedDefaults_ReviewSummaryMatchesTheDefaultSelection()
    {
        // Second, differently-shaped fixture: no toggling at all — the
        // recommended defaults ("openspec" recommended-on, "engram" and
        // "extra" recommended-off) flow straight through to Review.
        var session = BuildSession();
        var selection = new InstallSelection { Agents = ["claude"], Mode = "custom" };

        _ = new ComponentsPageViewModel(session, selection);
        var reviewPage = new ReviewPageViewModel(session, selection, new NotSupportedInstallerEngineClient());

        Assert.Contains("Basic protection", reviewPage.ComponentsSummary);
        Assert.Contains("OpenSpec", reviewPage.ComponentsSummary);
        Assert.DoesNotContain("Engram", reviewPage.ComponentsSummary);
        Assert.DoesNotContain("Extra", reviewPage.ComponentsSummary);
    }

    private static InstallerSessionState BuildSession() =>
        new(
            AssistantChoices: [new AssistantChoice("claude", "Claude")],
            DefaultAssistantId: "claude",
            InstallTypeChoices: [new InstallTypeChoice("custom", "Custom", "Choose exactly what to install.")],
            RecommendedModeId: "custom",
            ForcedComponents:
            [
                new ComponentChoice("permissions", "Basic protection", "Always on.", false)
            ],
            CustomComponents:
            [
                new ComponentChoice("openspec", "OpenSpec", "Plan and organize changes.", true),
                new ComponentChoice("engram", "Engram", "Persistent memory.", false),
                new ComponentChoice("extra", "Extra", "Not recommended by default.", false)
            ]);

    private sealed class NotSupportedInstallerEngineClient : IInstallerEngineClient
    {
        public string Language { get; set; } = "en";

        public Task<InstallerSessionState> LoadSessionAsync(CancellationToken cancellationToken = default)
            => throw new NotSupportedException();

        public IAsyncEnumerable<InstallProgressSnapshot> RunInstallAsync(IReadOnlyList<string> agents, string mode, IReadOnlyList<string> customIds, string? tier, CancellationToken cancellationToken = default)
            => throw new NotSupportedException();

        public Task<IReadOnlyList<StarterChoice>> ListStartersAsync(CancellationToken cancellationToken = default)
            => throw new NotSupportedException();

        public IAsyncEnumerable<InstallProgressSnapshot> RunStarterInstallAsync(string starterId, string projectPath, IReadOnlyList<string> agents, bool dryRun, CancellationToken cancellationToken = default)
            => throw new NotSupportedException();

        public Task<IReadOnlyList<BackupEntry>> ListBackupsAsync(CancellationToken cancellationToken = default)
            => throw new NotSupportedException();

        public Task<BackupActionResult> RunBackupActionAsync(string action, string id, string? description, CancellationToken cancellationToken = default)
            => throw new NotSupportedException();

        public Task<UninstallOptions> LoadUninstallOptionsAsync(CancellationToken cancellationToken = default)
            => throw new NotSupportedException();

        public IAsyncEnumerable<InstallProgressSnapshot> RunUninstallAsync(IReadOnlyList<string> agents, string mode, string strategy, string? backupId, CancellationToken cancellationToken = default)
            => throw new NotSupportedException();
    }
}
