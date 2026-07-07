namespace ActiveStack.Bootstrapper.Host.Navigation;

/// <summary>
/// Pure navigation graph for the Install wizard: given the current page,
/// the shared <see cref="InstallSelection"/>, and the session's tier-capable
/// agents, computes the next/previous page identity. No I/O, no mutation —
/// mirrors the TUI transitions (internal/tui/model.go:278-318) so drift
/// between the two UIs fails a test instead of shipping silently (D2,
/// design.md).
/// </summary>
public static class WizardFlow
{
    public static WizardPageId NextPage(
        WizardPageId current,
        InstallSelection selection,
        IReadOnlyList<string> tierCapableAgents)
    {
        return current switch
        {
            WizardPageId.Hub => WizardPageId.Assistants,
            WizardPageId.Assistants => WizardPageId.InstallType,
            WizardPageId.InstallType => IsCustom(selection)
                ? WizardPageId.Components
                : NextAfterModeDecision(selection, tierCapableAgents),
            WizardPageId.Components => NextAfterModeDecision(selection, tierCapableAgents),
            WizardPageId.Permissions => WizardPageId.Review,
            WizardPageId.Review => WizardPageId.Installing,
            WizardPageId.Installing => WizardPageId.Complete,
            WizardPageId.Complete => WizardPageId.Complete,
            _ => current
        };
    }

    public static WizardPageId PreviousPage(
        WizardPageId current,
        InstallSelection selection,
        IReadOnlyList<string> tierCapableAgents)
    {
        return current switch
        {
            WizardPageId.Hub => WizardPageId.Hub,
            WizardPageId.Assistants => WizardPageId.Hub,
            WizardPageId.InstallType => WizardPageId.Assistants,
            WizardPageId.Components => WizardPageId.InstallType,
            WizardPageId.Permissions => IsCustom(selection) ? WizardPageId.Components : WizardPageId.InstallType,
            WizardPageId.Review => PreviousBeforeReview(selection, tierCapableAgents),
            WizardPageId.Installing => WizardPageId.Review,
            WizardPageId.Complete => WizardPageId.Complete,
            _ => current
        };
    }

    private static WizardPageId NextAfterModeDecision(InstallSelection selection, IReadOnlyList<string> tierCapableAgents) =>
        AnyTierCapable(selection, tierCapableAgents) ? WizardPageId.Permissions : WizardPageId.Review;

    private static WizardPageId PreviousBeforeReview(InstallSelection selection, IReadOnlyList<string> tierCapableAgents)
    {
        if (AnyTierCapable(selection, tierCapableAgents))
        {
            return WizardPageId.Permissions;
        }

        return IsCustom(selection) ? WizardPageId.Components : WizardPageId.InstallType;
    }

    private static bool IsCustom(InstallSelection selection) =>
        string.Equals(selection.Mode, "custom", StringComparison.OrdinalIgnoreCase);

    private static bool AnyTierCapable(InstallSelection selection, IReadOnlyList<string> tierCapableAgents) =>
        selection.Agents.Any(agent => tierCapableAgents.Contains(agent, StringComparer.OrdinalIgnoreCase));
}
