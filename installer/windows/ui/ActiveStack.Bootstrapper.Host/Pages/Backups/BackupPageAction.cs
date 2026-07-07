namespace ActiveStack.Bootstrapper.Host.Pages.Backups;

/// <summary>
/// Which inline confirmation (if any) is active on the Backups page,
/// mirroring <c>internal/tui/backups_screen.go</c>'s <c>backupAction</c>
/// (D7, design.md). Only one is ever active at a time.
/// </summary>
public enum BackupPageAction
{
    None,
    Restore,
    Delete,
    Rename
}
