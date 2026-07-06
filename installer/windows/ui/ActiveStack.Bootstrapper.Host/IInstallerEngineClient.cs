using ActiveStack.Bootstrapper.Core;

namespace ActiveStack.Bootstrapper.Host;

public interface IInstallerEngineClient
{
    Task<InstallerSessionState> LoadSessionAsync(CancellationToken cancellationToken = default);

    IAsyncEnumerable<InstallProgressSnapshot> RunInstallAsync(
        string assistantId,
        string installTypeId,
        CancellationToken cancellationToken = default);
}
