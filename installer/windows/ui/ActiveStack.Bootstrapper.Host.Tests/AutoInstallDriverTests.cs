using ActiveStack.Bootstrapper.Core;
using ActiveStack.Bootstrapper.Host.Navigation;
using ActiveStack.Bootstrapper.Host.Pages;
using ActiveStack.Bootstrapper.Host.Pages.Install;
using Xunit;

namespace ActiveStack.Bootstrapper.Host.Tests;

/// <summary>
/// Regression coverage for the headless/CI automation path
/// (ACTIVE_STACK_AUTO_INSTALL). Since the shell now starts on the Language
/// page instead of the Hub, driving the automated loop must traverse the
/// preselected Language page and only THEN select "Install" on the (newly
/// constructed) Hub page — selecting it on the pre-advance Hub reference
/// would be silently discarded when the language advance swaps in a new
/// <see cref="HubPageViewModel"/> instance (gui-language-page, L4, task 8.1).
/// </summary>
public sealed class AutoInstallDriverTests
{
    [Fact]
    public async Task RunAsync_StartingOnLanguage_TraversesItReachesReviewAndConfirmsWithoutStalling()
    {
        var client = new FakeInstallerEngineClient(BuildSession(["claude"]));
        var shell = new ShellViewModel(client, initialLanguage: "en", persistLanguage: static _ => { });

        await AutoInstallDriver.RunAsync(shell, assistantOverride: null, installModeOverride: null);

        // Matches the original MainWindow behavior: the loop stops AT Review,
        // then one more AdvanceAsync confirms it, running the install stream
        // through to Complete. Reaching Complete (not stalling on Hub with no
        // entry selected) is exactly the regression this test guards.
        Assert.IsType<CompletePageViewModel>(shell.CurrentPage);
        Assert.Equal(1, client.RunInstallCallCount);
    }

    [Fact]
    public async Task RunAsync_AppliesAssistantAndModeOverridesAlongTheWay()
    {
        var client = new FakeInstallerEngineClient(BuildSession(["claude", "codex"]));
        var shell = new ShellViewModel(client, initialLanguage: "en", persistLanguage: static _ => { });

        await AutoInstallDriver.RunAsync(shell, assistantOverride: "codex", installModeOverride: "lite");

        Assert.IsType<CompletePageViewModel>(shell.CurrentPage);
        Assert.Equal(["codex"], client.CapturedAgents);
        Assert.Equal("lite", client.CapturedMode);
    }

    [Fact]
    public async Task RunAsync_WithAcActiveStackUiLangOverride_StartsInTheOverriddenLanguageAndReachesComplete()
    {
        var client = new FakeInstallerEngineClient(BuildSession(["claude"]));
        // Simulates BootstrapperLaunchOptions.LanguageOverride ("es") being
        // resolved by the composition root before constructing the shell.
        var shell = new ShellViewModel(client, initialLanguage: "es", persistLanguage: static _ => { });

        await AutoInstallDriver.RunAsync(shell, assistantOverride: null, installModeOverride: null);

        Assert.IsType<CompletePageViewModel>(shell.CurrentPage);
        Assert.Equal("es", client.Language);
    }

    private static InstallerSessionState BuildSession(IReadOnlyList<string> detectedAgents) =>
        new(
            AssistantChoices: detectedAgents.Select(id => new AssistantChoice(id, id)).ToList(),
            DefaultAssistantId: detectedAgents.FirstOrDefault(),
            InstallTypeChoices:
            [
                new InstallTypeChoice("lite", "Quick", "Fast setup to start working right away."),
                new InstallTypeChoice("full", "Complete", "Full recommended setup with all key tools.")
            ],
            RecommendedModeId: "full",
            ForcedComponents: [],
            CustomComponents: [],
            TierCapable: false,
            TierCapableAgents: [],
            PermissionTierChoices: []);

    private sealed class FakeInstallerEngineClient : IInstallerEngineClient
    {
        private readonly InstallerSessionState _session;

        public FakeInstallerEngineClient(InstallerSessionState session) => _session = session;

        public string Language { get; set; } = "en";

        public int RunInstallCallCount { get; private set; }

        public IReadOnlyList<string>? CapturedAgents { get; private set; }

        public string? CapturedMode { get; private set; }

        public Task<InstallerSessionState> LoadSessionAsync(CancellationToken cancellationToken = default)
            => Task.FromResult(_session);

        public async IAsyncEnumerable<InstallProgressSnapshot> RunInstallAsync(
            IReadOnlyList<string> agents,
            string mode,
            IReadOnlyList<string> customIds,
            string? tier,
            [System.Runtime.CompilerServices.EnumeratorCancellation] CancellationToken cancellationToken = default)
        {
            RunInstallCallCount++;
            CapturedAgents = agents;
            CapturedMode = mode;
            yield return new InstallProgressSnapshot("install_finished", "install", null, "Installation finished successfully.", true);
            await Task.CompletedTask;
        }

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
