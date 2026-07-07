using ActiveStack.Bootstrapper.Core;
using ActiveStack.Bootstrapper.Host.Navigation;
using ActiveStack.Bootstrapper.Host.Pages.Uninstall;
using Xunit;

namespace ActiveStack.Bootstrapper.Host.Tests;

public sealed class UninstallStrategyPageViewModelTests
{
    [Fact]
    public void Constructor_DefaultsToTargetedWithNoBackupSelectorAndCanAdvance()
    {
        var options = BuildOptions();
        var selection = new UninstallSelection();
        var client = new FakeClient();

        var page = new UninstallStrategyPageViewModel(options, selection, client);

        Assert.Equal("targeted", page.SelectedStrategyId);
        Assert.False(page.ShowBackupSelector);
        Assert.True(page.CanAdvance);
        Assert.Equal(0, client.ListBackupsCallCount);
    }

    [Fact]
    public async Task SelectingRestore_RevealsBackupSelectorAndGatesUntilChosen()
    {
        var options = BuildOptions();
        var selection = new UninstallSelection();
        var backup = new BackupEntry("b1", "2026-01-01T00:00:00Z", "manual", "desc", 3, false, true, "2026-01-01 manual", "C:\\backups\\b1\\manifest.json");
        var client = new FakeClient([backup]);
        var page = new UninstallStrategyPageViewModel(options, selection, client);

        await page.SelectStrategyAsync("restore");

        Assert.True(page.ShowBackupSelector);
        Assert.False(page.CanAdvance);
        Assert.Equal(1, client.ListBackupsCallCount);
        Assert.Single(page.Backups);

        page.SelectedBackup = backup;

        Assert.True(page.CanAdvance);
        Assert.Equal(backup, selection.SelectedBackup);
        Assert.True(selection.RequiresManifest);
        Assert.Equal("restore", selection.Strategy);
    }

    private static UninstallOptions BuildOptions() =>
        new(["claude"],
            [new InstallTypeChoice("full", "Complete", "Full removal.")],
            [
                new UninstallStrategyChoice("targeted", "Targeted", "Undo only what Active Stack added.", true, false),
                new UninstallStrategyChoice("restore", "Restore from backup", "Restore a previous backup.", false, true)
            ]);

    private sealed class FakeClient : IInstallerEngineClient
    {
        private readonly IReadOnlyList<BackupEntry> _backups;

        public FakeClient(IReadOnlyList<BackupEntry>? backups = null)
        {
            _backups = backups ?? [];
        }

        public int ListBackupsCallCount { get; private set; }

        public Task<InstallerSessionState> LoadSessionAsync(CancellationToken cancellationToken = default) => throw new NotSupportedException();
        public IAsyncEnumerable<InstallProgressSnapshot> RunInstallAsync(IReadOnlyList<string> agents, string mode, IReadOnlyList<string> customIds, string? tier, CancellationToken cancellationToken = default) => throw new NotSupportedException();
        public Task<IReadOnlyList<StarterChoice>> ListStartersAsync(CancellationToken cancellationToken = default) => throw new NotSupportedException();
        public IAsyncEnumerable<InstallProgressSnapshot> RunStarterInstallAsync(string starterId, string projectPath, IReadOnlyList<string> agents, bool dryRun, CancellationToken cancellationToken = default) => throw new NotSupportedException();

        public Task<IReadOnlyList<BackupEntry>> ListBackupsAsync(CancellationToken cancellationToken = default)
        {
            ListBackupsCallCount++;
            return Task.FromResult(_backups);
        }

        public Task<BackupActionResult> RunBackupActionAsync(string action, string id, string? description, CancellationToken cancellationToken = default) => throw new NotSupportedException();
        public Task<UninstallOptions> LoadUninstallOptionsAsync(CancellationToken cancellationToken = default) => throw new NotSupportedException();
        public IAsyncEnumerable<InstallProgressSnapshot> RunUninstallAsync(IReadOnlyList<string> agents, string mode, string strategy, string? backupId, CancellationToken cancellationToken = default) => throw new NotSupportedException();
    }
}
