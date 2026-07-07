using ActiveStack.Bootstrapper.Core;

namespace ActiveStack.Bootstrapper.Host;

public interface IInstallerEngineClient
{
    Task<InstallerSessionState> LoadSessionAsync(CancellationToken cancellationToken = default);

    IAsyncEnumerable<InstallProgressSnapshot> RunInstallAsync(
        IReadOnlyList<string> agents,
        string mode,
        IReadOnlyList<string> customIds,
        string? tier,
        CancellationToken cancellationToken = default);

    Task<IReadOnlyList<StarterChoice>> ListStartersAsync(CancellationToken cancellationToken = default);

    IAsyncEnumerable<InstallProgressSnapshot> RunStarterInstallAsync(
        string starterId,
        string projectPath,
        IReadOnlyList<string> agents,
        bool dryRun,
        CancellationToken cancellationToken = default);

    Task<IReadOnlyList<BackupEntry>> ListBackupsAsync(CancellationToken cancellationToken = default);

    Task<BackupActionResult> RunBackupActionAsync(
        string action,
        string id,
        string? description,
        CancellationToken cancellationToken = default);

    Task<UninstallOptions> LoadUninstallOptionsAsync(CancellationToken cancellationToken = default);

    IAsyncEnumerable<InstallProgressSnapshot> RunUninstallAsync(
        IReadOnlyList<string> agents,
        string mode,
        string strategy,
        string? backupId,
        CancellationToken cancellationToken = default);
}
