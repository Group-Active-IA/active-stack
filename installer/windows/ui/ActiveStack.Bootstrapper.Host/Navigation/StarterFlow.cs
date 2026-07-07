namespace ActiveStack.Bootstrapper.Host.Navigation;

/// <summary>
/// Pure navigation graph for the Starter flow: linear
/// Hub → StarterCatalog → StarterTarget → StarterReview → Installing →
/// Complete, mirroring <c>internal/tui/starters_screen.go</c>. No I/O, no
/// mutation — kept separate from <see cref="WizardFlow"/> (D1, design.md).
/// </summary>
public static class StarterFlow
{
    public static WizardPageId NextPage(WizardPageId current, StarterSelection selection) => current switch
    {
        WizardPageId.Hub => WizardPageId.StarterCatalog,
        WizardPageId.StarterCatalog => WizardPageId.StarterTarget,
        WizardPageId.StarterTarget => WizardPageId.StarterReview,
        WizardPageId.StarterReview => WizardPageId.Installing,
        WizardPageId.Installing => WizardPageId.Complete,
        WizardPageId.Complete => WizardPageId.Complete,
        _ => current
    };

    public static WizardPageId PreviousPage(WizardPageId current, StarterSelection selection) => current switch
    {
        WizardPageId.Hub => WizardPageId.Hub,
        WizardPageId.StarterCatalog => WizardPageId.Hub,
        WizardPageId.StarterTarget => WizardPageId.StarterCatalog,
        WizardPageId.StarterReview => WizardPageId.StarterTarget,
        WizardPageId.Installing => WizardPageId.StarterReview,
        WizardPageId.Complete => WizardPageId.Complete,
        _ => current
    };
}
