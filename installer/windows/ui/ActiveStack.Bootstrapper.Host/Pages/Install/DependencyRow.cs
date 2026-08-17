namespace ActiveStack.Bootstrapper.Host.Pages.Install;

/// <summary>
/// One dependency row on the Dependencies page (gui-installer-dependency-detection).
/// Display-only — no execution capability anywhere on this type. InstallCommand
/// is pre-joined display text from the engine (system.InstallCommandsForDep),
/// never argv; the page renders it as read-only, selectable text.
/// </summary>
public sealed record DependencyRow(
    string Name,
    bool Required,
    bool Installed,
    string Version,
    string StatusLabel,
    string RequiredLabel,
    string InstallHint,
    string InstallCommand)
{
    public bool HasInstallCommand => !string.IsNullOrEmpty(InstallCommand);

    public bool HasInstallHint => !string.IsNullOrEmpty(InstallHint);
}
