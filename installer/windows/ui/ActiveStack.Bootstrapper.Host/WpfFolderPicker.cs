using ActiveStack.Bootstrapper.Host.Navigation;
using Microsoft.Win32;

namespace ActiveStack.Bootstrapper.Host;

/// <summary>
/// Production <see cref="IFolderPicker"/> wrapping .NET 8's
/// <see cref="OpenFolderDialog"/> (D6, design.md). Needs an STA UI thread —
/// never exercised in unit tests, only wired into the composition root
/// (<see cref="MainWindow"/>) and exercised in the C6 manual smoke.
/// </summary>
public sealed class WpfFolderPicker : IFolderPicker
{
    public string? PickFolder()
    {
        var dialog = new OpenFolderDialog();
        return dialog.ShowDialog() == true ? dialog.FolderName : null;
    }
}
