using ActiveStack.Bootstrapper.Core;

namespace ActiveStack.Bootstrapper.Host.Navigation;

/// <summary>
/// Mutable model shared across every Uninstall-flow page view-model: agents,
/// mode, strategy id, whether the chosen strategy requires a backup manifest,
/// and (for the "restore" strategy) the selected backup. Created once when
/// the Uninstall flow starts (from the Hub) and threaded through
/// <see cref="ShellViewModel"/> and the page view-models, mirroring
/// <see cref="InstallSelection"/> (D4, design.md). <see cref="UninstallFlow"/>
/// only ever reads it — it never mutates it.
/// </summary>
public sealed class UninstallSelection
{
    public List<string> Agents { get; set; } = [];

    public string Mode { get; set; } = string.Empty;

    public string Strategy { get; set; } = string.Empty;

    public bool RequiresManifest { get; set; }

    public BackupEntry? SelectedBackup { get; set; }
}
