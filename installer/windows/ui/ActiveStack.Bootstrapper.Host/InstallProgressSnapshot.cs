namespace ActiveStack.Bootstrapper.Host;

public sealed record InstallProgressSnapshot(
    string Type,
    string? Phase,
    string? StepId,
    string? Message,
    bool Success,
    string? Details = null,
    string? Timestamp = null);

/// <summary>
/// The three terminal event types the engine's stream can end with — one per
/// operation (install/uninstall/starter). Shared between
/// <see cref="ProcessInstallerEngineClient"/> (deciding whether to trust the
/// stream over the process exit code) and <see cref="Pages.ProgressPageViewModel"/>
/// (deciding which snapshot is the terminal one) so the two cannot drift.
/// </summary>
public static class TerminalEventTypes
{
    public static bool Contains(string? type) =>
        string.Equals(type, "install_finished", StringComparison.OrdinalIgnoreCase) ||
        string.Equals(type, "starter_finished", StringComparison.OrdinalIgnoreCase) ||
        string.Equals(type, "uninstall_finished", StringComparison.OrdinalIgnoreCase);
}
