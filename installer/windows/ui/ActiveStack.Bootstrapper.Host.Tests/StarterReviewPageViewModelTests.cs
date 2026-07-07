using ActiveStack.Bootstrapper.Core;
using ActiveStack.Bootstrapper.Host.Navigation;
using ActiveStack.Bootstrapper.Host.Pages.Starters;
using Xunit;

namespace ActiveStack.Bootstrapper.Host.Tests;

public sealed class StarterReviewPageViewModelTests
{
    [Fact]
    public void Constructor_SummaryReflectsStarterTargetAndAgents()
    {
        var starters = BuildStarters();
        var selection = new StarterSelection { StarterId = "web", ProjectPath = "C:\\My Project", Agents = ["claude"] };
        var page = new StarterReviewPageViewModel(starters, selection, new RecordingClient());

        Assert.Equal("Web Starter", page.StarterSummary);
        Assert.Equal("C:\\My Project", page.TargetSummary);
        Assert.Contains("claude", page.AgentsSummary);
        Assert.Equal("Install", page.PrimaryLabel);
        Assert.True(page.CanAdvance);
    }

    [Fact]
    public async Task StartStarterInstall_InvokesEngineClientWithFullSelectionAndIntactPath()
    {
        var starters = BuildStarters();
        var selection = new StarterSelection { StarterId = "web", ProjectPath = "C:\\Users\\dev\\My Project", Agents = ["claude", "opencode"] };
        var client = new RecordingClient();
        var page = new StarterReviewPageViewModel(starters, selection, client);

        await foreach (var _ in page.StartStarterInstall())
        {
        }

        Assert.Equal("web", client.CapturedStarterId);
        Assert.Equal("C:\\Users\\dev\\My Project", client.CapturedProjectPath);
        Assert.Equal(["claude", "opencode"], client.CapturedAgents);
    }

    private static IReadOnlyList<StarterChoice> BuildStarters() =>
    [
        new StarterChoice("web", "Web Starter", "A web app starter.", [], ["claude", "opencode"], 2)
    ];

    private sealed class RecordingClient : IInstallerEngineClient
    {
        public string? CapturedStarterId { get; private set; }
        public string? CapturedProjectPath { get; private set; }
        public IReadOnlyList<string>? CapturedAgents { get; private set; }

        public Task<InstallerSessionState> LoadSessionAsync(CancellationToken cancellationToken = default) => throw new NotSupportedException();
        public IAsyncEnumerable<InstallProgressSnapshot> RunInstallAsync(IReadOnlyList<string> agents, string mode, IReadOnlyList<string> customIds, string? tier, CancellationToken cancellationToken = default) => throw new NotSupportedException();
        public Task<IReadOnlyList<StarterChoice>> ListStartersAsync(CancellationToken cancellationToken = default) => throw new NotSupportedException();

        public async IAsyncEnumerable<InstallProgressSnapshot> RunStarterInstallAsync(
            string starterId,
            string projectPath,
            IReadOnlyList<string> agents,
            bool dryRun,
            [System.Runtime.CompilerServices.EnumeratorCancellation] CancellationToken cancellationToken = default)
        {
            CapturedStarterId = starterId;
            CapturedProjectPath = projectPath;
            CapturedAgents = agents;

            yield return new InstallProgressSnapshot("starter_finished", "starter", null, "Starter install finished successfully.", true);
            await Task.CompletedTask;
        }

        public Task<IReadOnlyList<BackupEntry>> ListBackupsAsync(CancellationToken cancellationToken = default) => throw new NotSupportedException();
        public Task<BackupActionResult> RunBackupActionAsync(string action, string id, string? description, CancellationToken cancellationToken = default) => throw new NotSupportedException();
        public Task<UninstallOptions> LoadUninstallOptionsAsync(CancellationToken cancellationToken = default) => throw new NotSupportedException();
        public IAsyncEnumerable<InstallProgressSnapshot> RunUninstallAsync(IReadOnlyList<string> agents, string mode, string strategy, string? backupId, CancellationToken cancellationToken = default) => throw new NotSupportedException();
    }
}
