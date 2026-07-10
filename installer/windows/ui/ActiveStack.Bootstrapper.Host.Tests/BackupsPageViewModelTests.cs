using ActiveStack.Bootstrapper.Core;
using ActiveStack.Bootstrapper.Host.Pages.Backups;
using Xunit;

namespace ActiveStack.Bootstrapper.Host.Tests;

public sealed class BackupsPageViewModelTests
{
    [Fact]
    public async Task RefreshAsync_LoadsBackupsByDisplayLabel()
    {
        var backup = BuildBackup("b1", "2026-01-01 manual");
        var client = new FakeClient([backup]);
        var page = new BackupsPageViewModel(client);

        await page.RefreshAsync();

        var item = Assert.Single(page.Items);
        Assert.Equal("2026-01-01 manual", item.DisplayLabel);
        Assert.False(page.IsEmpty);
        Assert.False(page.CanAdvance);
    }

    [Fact]
    public async Task RefreshAsync_EmptyStore_ShowsEmptyStateAndNoActions()
    {
        var client = new FakeClient([]);
        var page = new BackupsPageViewModel(client);

        await page.RefreshAsync();

        Assert.Empty(page.Items);
        Assert.True(page.IsEmpty);
        Assert.False(page.CanAdvance);
    }

    [Fact]
    public async Task RestoreConfirmation_MatchesTuiTextAndInvokesRestore()
    {
        var backup = BuildBackup("b1", "label");
        var client = new FakeClient([backup]);
        var page = new BackupsPageViewModel(client);
        await page.RefreshAsync();

        page.BeginRestore(backup);

        Assert.Equal(BackupPageAction.Restore, page.Action);
        Assert.Equal("This will OVERWRITE your current configuration.", page.ConfirmationText);

        await page.ConfirmActionAsync();

        Assert.Equal("restore", client.CapturedAction);
        Assert.Equal("b1", client.CapturedId);
        Assert.Null(client.CapturedDescription);
        Assert.Equal(BackupPageAction.None, page.Action);
    }

    [Fact]
    public async Task DeleteConfirmation_MatchesTuiTextAndInvokesDelete()
    {
        var backup = BuildBackup("b1", "label");
        var client = new FakeClient([backup]);
        var page = new BackupsPageViewModel(client);
        await page.RefreshAsync();

        page.BeginDelete(backup);

        Assert.Equal("This will PERMANENTLY DELETE this backup.", page.ConfirmationText);

        await page.ConfirmActionAsync();

        Assert.Equal("delete", client.CapturedAction);
        Assert.Equal("b1", client.CapturedId);
        Assert.Null(client.CapturedDescription);
    }

    [Fact]
    public async Task RenameConfirmation_CapturesDescriptionAndInvokesRename()
    {
        var backup = BuildBackup("b1", "label");
        var client = new FakeClient([backup]);
        var page = new BackupsPageViewModel(client);
        await page.RefreshAsync();

        page.BeginRename(backup);
        page.RenameInput = "My new description";

        await page.ConfirmActionAsync();

        Assert.Equal("rename", client.CapturedAction);
        Assert.Equal("b1", client.CapturedId);
        Assert.Equal("My new description", client.CapturedDescription);
    }

    [Fact]
    public async Task CancellingAnAction_InvokesNothingAndReturnsToList()
    {
        var backup = BuildBackup("b1", "label");
        var client = new FakeClient([backup]);
        var page = new BackupsPageViewModel(client);
        await page.RefreshAsync();

        page.BeginDelete(backup);
        page.CancelAction();

        Assert.Equal(BackupPageAction.None, page.Action);
        Assert.Equal(0, client.RunBackupActionCallCount);
    }

    [Fact]
    public async Task DeleteThenRelist_YieldsTheReducedList()
    {
        var kept = BuildBackup("keep", "keep-label");
        var removed = BuildBackup("gone", "gone-label");
        var client = new FakeClient([kept, removed]);
        var page = new BackupsPageViewModel(client);
        await page.RefreshAsync();
        Assert.Equal(2, page.Items.Count);

        client.SetBackupsAfterAction([kept]);
        page.BeginDelete(removed);
        await page.ConfirmActionAsync();

        var remaining = Assert.Single(page.Items);
        Assert.Equal("keep", remaining.Id);
        Assert.NotNull(page.LastMessage);
    }

    private static BackupEntry BuildBackup(string id, string displayLabel) =>
        new(id, "2026-01-01T00:00:00Z", "manual", "desc", 3, false, true, displayLabel, $"C:\\backups\\{id}\\manifest.json");

    private sealed class FakeClient : IInstallerEngineClient
    {
        private IReadOnlyList<BackupEntry> _backups;
        private IReadOnlyList<BackupEntry>? _afterAction;

        public FakeClient(IReadOnlyList<BackupEntry> backups)
        {
            _backups = backups;
        }

        public string Language { get; set; } = "en";
        public int RunBackupActionCallCount { get; private set; }
        public string? CapturedAction { get; private set; }
        public string? CapturedId { get; private set; }
        public string? CapturedDescription { get; private set; }

        public void SetBackupsAfterAction(IReadOnlyList<BackupEntry> backups) => _afterAction = backups;

        public Task<InstallerSessionState> LoadSessionAsync(CancellationToken cancellationToken = default) => throw new NotSupportedException();
        public IAsyncEnumerable<InstallProgressSnapshot> RunInstallAsync(IReadOnlyList<string> agents, string mode, IReadOnlyList<string> customIds, string? tier, CancellationToken cancellationToken = default) => throw new NotSupportedException();
        public Task<IReadOnlyList<StarterChoice>> ListStartersAsync(CancellationToken cancellationToken = default) => throw new NotSupportedException();
        public IAsyncEnumerable<InstallProgressSnapshot> RunStarterInstallAsync(string starterId, string projectPath, IReadOnlyList<string> agents, bool dryRun, CancellationToken cancellationToken = default) => throw new NotSupportedException();

        public Task<IReadOnlyList<BackupEntry>> ListBackupsAsync(CancellationToken cancellationToken = default) => Task.FromResult(_backups);

        public Task<BackupActionResult> RunBackupActionAsync(string action, string id, string? description, CancellationToken cancellationToken = default)
        {
            RunBackupActionCallCount++;
            CapturedAction = action;
            CapturedId = id;
            CapturedDescription = description;

            if (_afterAction is not null)
            {
                _backups = _afterAction;
            }

            return Task.FromResult(new BackupActionResult(true, $"{action} succeeded", id));
        }

        public Task<UninstallOptions> LoadUninstallOptionsAsync(CancellationToken cancellationToken = default) => throw new NotSupportedException();
        public IAsyncEnumerable<InstallProgressSnapshot> RunUninstallAsync(IReadOnlyList<string> agents, string mode, string strategy, string? backupId, CancellationToken cancellationToken = default) => throw new NotSupportedException();
    }
}
