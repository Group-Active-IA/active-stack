namespace ActiveStack.Bootstrapper.Host.Navigation;

/// <summary>
/// Pure navigation graph for the Uninstall wizard: linear
/// Hub → UninstallAgents → UninstallMode → UninstallStrategy →
/// UninstallConfirm → Installing → Complete, mirroring
/// <c>internal/tui/uninstall_screen.go</c>. No I/O, no mutation — kept
/// separate from <see cref="WizardFlow"/> (D1, design.md) since the two
/// flows share no branching predicates.
/// </summary>
public static class UninstallFlow
{
    public static WizardPageId NextPage(WizardPageId current, UninstallSelection selection) => current switch
    {
        WizardPageId.Hub => WizardPageId.UninstallAgents,
        WizardPageId.UninstallAgents => WizardPageId.UninstallMode,
        WizardPageId.UninstallMode => WizardPageId.UninstallStrategy,
        WizardPageId.UninstallStrategy => WizardPageId.UninstallConfirm,
        WizardPageId.UninstallConfirm => WizardPageId.Installing,
        WizardPageId.Installing => WizardPageId.Complete,
        WizardPageId.Complete => WizardPageId.Complete,
        _ => current
    };

    public static WizardPageId PreviousPage(WizardPageId current, UninstallSelection selection) => current switch
    {
        WizardPageId.Hub => WizardPageId.Hub,
        WizardPageId.UninstallAgents => WizardPageId.Hub,
        WizardPageId.UninstallMode => WizardPageId.UninstallAgents,
        WizardPageId.UninstallStrategy => WizardPageId.UninstallMode,
        WizardPageId.UninstallConfirm => WizardPageId.UninstallStrategy,
        WizardPageId.Installing => WizardPageId.UninstallConfirm,
        WizardPageId.Complete => WizardPageId.Complete,
        _ => current
    };
}
