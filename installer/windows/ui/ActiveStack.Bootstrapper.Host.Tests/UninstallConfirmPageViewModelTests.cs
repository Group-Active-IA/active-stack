using ActiveStack.Bootstrapper.Core;
using ActiveStack.Bootstrapper.Host.Navigation;
using ActiveStack.Bootstrapper.Host.Pages.Uninstall;
using Xunit;

namespace ActiveStack.Bootstrapper.Host.Tests;

public sealed class UninstallConfirmPageViewModelTests
{
    [Fact]
    public void Constructor_SummaryShowsAgentsModeStrategyAndWarning()
    {
        var options = BuildOptions();
        var selection = new UninstallSelection { Agents = ["claude"], Mode = "full", Strategy = "targeted" };
        var page = new UninstallConfirmPageViewModel(options, selection, new RecordingClient());

        Assert.Contains("claude", page.AgentsSummary);
        Assert.Contains("Complete", page.ModeSummary);
        Assert.Contains("Targeted", page.StrategySummary);
        Assert.Equal("This will modify your agent configuration.", page.WarningText);
    }

    [Fact]
    public async Task Advancing_WithTargetedStrategy_PassesNoManifestPath()
    {
        var options = BuildOptions();
        var selection = new UninstallSelection { Agents = ["claude"], Mode = "full", Strategy = "targeted", RequiresManifest = false };
        var client = new RecordingClient();
        var page = new UninstallConfirmPageViewModel(options, selection, client);

        await foreach (var _ in page.StartStream())
        {
        }

        Assert.Equal(["claude"], client.CapturedAgents);
        Assert.Equal("full", client.CapturedMode);
        Assert.Equal("targeted", client.CapturedStrategy);
        Assert.Null(client.CapturedBackupId);
    }

    [Fact]
    public async Task Advancing_WithRestoreStrategy_CarriesSelectedBackupManifestPath()
    {
        var options = BuildOptions();
        var backup = new BackupEntry("b1", "2026-01-01T00:00:00Z", "manual", "desc", 3, false, true, "label", "C:\\backups\\b1\\manifest.json");
        var selection = new UninstallSelection { Agents = ["claude"], Mode = "full", Strategy = "restore", RequiresManifest = true, SelectedBackup = backup };
        var client = new RecordingClient();
        var page = new UninstallConfirmPageViewModel(options, selection, client);

        await foreach (var _ in page.StartStream())
        {
        }

        Assert.Equal("restore", client.CapturedStrategy);
        Assert.Equal("C:\\backups\\b1\\manifest.json", client.CapturedBackupId);
    }

    private static UninstallOptions BuildOptions() =>
        new(["claude"],
            [new InstallTypeChoice("full", "Complete", "Full removal.")],
            [
                new UninstallStrategyChoice("targeted", "Targeted", "Undo only what Active Stack added.", true, false),
                new UninstallStrategyChoice("restore", "Restore from backup", "Restore a previous backup.", false, true)
            ]);

    private sealed class RecordingClient : IInstallerEngineClient
    {
        public string Language { get; set; } = "en";
        public IReadOnlyList<string>? CapturedAgents { get; private set; }
        public string? CapturedMode { get; private set; }
        public string? CapturedStrategy { get; private set; }
        public string? CapturedBackupId { get; private set; }

        public Task<InstallerSessionState> LoadSessionAsync(CancellationToken cancellationToken = default) => throw new NotSupportedException();
        public IAsyncEnumerable<InstallProgressSnapshot> RunInstallAsync(IReadOnlyList<string> agents, string mode, IReadOnlyList<string> customIds, string? tier, CancellationToken cancellationToken = default) => throw new NotSupportedException();
        public Task<IReadOnlyList<StarterChoice>> ListStartersAsync(CancellationToken cancellationToken = default) => throw new NotSupportedException();
        public IAsyncEnumerable<InstallProgressSnapshot> RunStarterInstallAsync(string starterId, string projectPath, IReadOnlyList<string> agents, bool dryRun, CancellationToken cancellationToken = default) => throw new NotSupportedException();
        public Task<IReadOnlyList<BackupEntry>> ListBackupsAsync(CancellationToken cancellationToken = default) => throw new NotSupportedException();
        public Task<BackupActionResult> RunBackupActionAsync(string action, string id, string? description, CancellationToken cancellationToken = default) => throw new NotSupportedException();
        public Task<UninstallOptions> LoadUninstallOptionsAsync(CancellationToken cancellationToken = default) => throw new NotSupportedException();

        public async IAsyncEnumerable<InstallProgressSnapshot> RunUninstallAsync(
            IReadOnlyList<string> agents,
            string mode,
            string strategy,
            string? backupId,
            [System.Runtime.CompilerServices.EnumeratorCancellation] CancellationToken cancellationToken = default)
        {
            CapturedAgents = agents;
            CapturedMode = mode;
            CapturedStrategy = strategy;
            CapturedBackupId = backupId;

            yield return new InstallProgressSnapshot("uninstall_finished", "uninstall", null, "Uninstall finished successfully.", true);
            await Task.CompletedTask;
        }
    }
}
