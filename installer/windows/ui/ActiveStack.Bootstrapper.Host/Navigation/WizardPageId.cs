namespace ActiveStack.Bootstrapper.Host.Navigation;

/// <summary>
/// Identity of every page the Install wizard (and its surrounding Hub) can
/// show. Kept as a flat enum so <see cref="WizardFlow"/> can express the
/// navigation graph as pure, typed functions instead of stringly-typed
/// screen names (mirrors the TUI's <c>Screen</c> type, internal/tui).
/// </summary>
public enum WizardPageId
{
    Hub,
    Assistants,
    InstallType,
    Components,
    Permissions,
    Review,
    Installing,
    Complete,
    UninstallAgents,
    UninstallMode,
    UninstallStrategy,
    UninstallConfirm,
    StarterCatalog,
    StarterTarget,
    StarterReview,
    Backups
}
