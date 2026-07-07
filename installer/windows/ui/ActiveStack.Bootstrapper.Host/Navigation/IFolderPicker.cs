namespace ActiveStack.Bootstrapper.Host.Navigation;

/// <summary>
/// Seam over the platform folder-picker so
/// <see cref="Pages.Starters.StarterTargetPageViewModel"/> is unit-testable
/// without an STA UI thread (D6, design.md). Returns <c>null</c> when the
/// user cancels the picker.
/// </summary>
public interface IFolderPicker
{
    string? PickFolder();
}
