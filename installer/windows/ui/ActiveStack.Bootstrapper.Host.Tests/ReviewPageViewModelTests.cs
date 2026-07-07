using ActiveStack.Bootstrapper.Core;
using ActiveStack.Bootstrapper.Host.Navigation;
using ActiveStack.Bootstrapper.Host.Pages.Install;
using Xunit;

namespace ActiveStack.Bootstrapper.Host.Tests;

public sealed class ReviewPageViewModelTests
{
    [Fact]
    public void Constructor_SummaryReflectsAgentsModeComponentsAndTier()
    {
        var session = BuildSession();
        var selection = new InstallSelection
        {
            Agents = ["claude"],
            Mode = "custom",
            CustomIds = ["permissions", "openspec"],
            Tier = "balanceado"
        };
        var page = new ReviewPageViewModel(session, selection, new RecordingInstallerEngineClient());

        Assert.Contains("Claude", page.AgentsSummary);
        Assert.DoesNotContain("Codex", page.AgentsSummary);
        Assert.Equal("Custom", page.ModeSummary);
        Assert.Contains("Basic protection", page.ComponentsSummary);
        Assert.Contains("OpenSpec", page.ComponentsSummary);
        Assert.Equal("Balanceado", page.TierSummary);
        Assert.Equal("Install", page.PrimaryLabel);
        Assert.True(page.CanAdvance);
    }

    [Fact]
    public async Task StartInstall_InvokesEngineClientWithTheFullSelection()
    {
        var session = BuildSession();
        var selection = new InstallSelection
        {
            Agents = ["claude", "codex"],
            Mode = "custom",
            CustomIds = ["permissions", "openspec"],
            Tier = "bypass"
        };
        var client = new RecordingInstallerEngineClient();
        var page = new ReviewPageViewModel(session, selection, client);

        var snapshots = new List<InstallProgressSnapshot>();
        await foreach (var snapshot in page.StartInstall())
        {
            snapshots.Add(snapshot);
        }

        Assert.Equal(["claude", "codex"], client.CapturedAgents);
        Assert.Equal("custom", client.CapturedMode);
        Assert.Equal(["permissions", "openspec"], client.CapturedCustomIds);
        Assert.Equal("bypass", client.CapturedTier);
        Assert.Single(snapshots);
    }

    private static InstallerSessionState BuildSession() =>
        new(
            AssistantChoices:
            [
                new AssistantChoice("claude", "Claude"),
                new AssistantChoice("codex", "Codex")
            ],
            DefaultAssistantId: "claude",
            InstallTypeChoices:
            [
                new InstallTypeChoice("full", "Complete", "Full recommended setup with all key tools."),
                new InstallTypeChoice("custom", "Custom", "Choose exactly what to install.")
            ],
            RecommendedModeId: "full",
            ForcedComponents:
            [
                new ComponentChoice("permissions", "Basic protection", "Helps avoid unsafe changes.", false)
            ],
            CustomComponents:
            [
                new ComponentChoice("openspec", "OpenSpec", "Plan and organize changes.", true)
            ],
            TierCapable: true,
            TierCapableAgents: ["claude"],
            PermissionTierChoices:
            [
                new PermissionTierChoice("balanceado", "Balanceado", "Ask for risky changes only.", true, null),
                new PermissionTierChoice("bypass", "Bypass", "Never ask.", false, "Bypass warning")
            ]);

    private sealed class RecordingInstallerEngineClient : IInstallerEngineClient
    {
        public IReadOnlyList<string>? CapturedAgents { get; private set; }
        public string? CapturedMode { get; private set; }
        public IReadOnlyList<string>? CapturedCustomIds { get; private set; }
        public string? CapturedTier { get; private set; }

        public Task<InstallerSessionState> LoadSessionAsync(CancellationToken cancellationToken = default)
            => throw new NotSupportedException();

        public async IAsyncEnumerable<InstallProgressSnapshot> RunInstallAsync(
            IReadOnlyList<string> agents,
            string mode,
            IReadOnlyList<string> customIds,
            string? tier,
            [System.Runtime.CompilerServices.EnumeratorCancellation] CancellationToken cancellationToken = default)
        {
            CapturedAgents = agents;
            CapturedMode = mode;
            CapturedCustomIds = customIds;
            CapturedTier = tier;

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
