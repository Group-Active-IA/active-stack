namespace ActiveStack.Bootstrapper.Host.Navigation;

/// <summary>
/// Pure navigation graph for the Uninstall wizard: Hub → UninstallAgents →
/// UninstallStrategy → (UninstallMode, only for the "targeted" strategy) →
/// UninstallConfirm → Installing → Complete. Strategy comes before Mode
/// (and Mode is skipped entirely for "restore") because Mode only feeds
/// <c>catalog.ForMode</c> for a targeted removal — it is meaningless when
/// restoring from a pre-install backup manifest, where asking it first (the
/// former order) made no sense to the user. No I/O, no mutation — kept
/// separate from <see cref="WizardFlow"/> (D1, design.md) since the two
/// flows share no branching predicates.
/// </summary>
public static class UninstallFlow
{
    public static WizardPageId NextPage(WizardPageId current, UninstallSelection selection) => current switch
    {
        WizardPageId.Hub => WizardPageId.UninstallAgents,
        WizardPageId.UninstallAgents => WizardPageId.UninstallStrategy,
        WizardPageId.UninstallStrategy => IsTargetedStrategy(selection) ? WizardPageId.UninstallMode : WizardPageId.UninstallConfirm,
        WizardPageId.UninstallMode => WizardPageId.UninstallConfirm,
        WizardPageId.UninstallConfirm => WizardPageId.Installing,
        WizardPageId.Installing => WizardPageId.Complete,
        WizardPageId.Complete => WizardPageId.Complete,
        _ => current
    };

    public static WizardPageId PreviousPage(WizardPageId current, UninstallSelection selection) => current switch
    {
        WizardPageId.Hub => WizardPageId.Hub,
        WizardPageId.UninstallAgents => WizardPageId.Hub,
        WizardPageId.UninstallStrategy => WizardPageId.UninstallAgents,
        WizardPageId.UninstallMode => WizardPageId.UninstallStrategy,
        WizardPageId.UninstallConfirm => IsTargetedStrategy(selection) ? WizardPageId.UninstallMode : WizardPageId.UninstallStrategy,
        WizardPageId.Installing => WizardPageId.UninstallConfirm,
        WizardPageId.Complete => WizardPageId.Complete,
        _ => current
    };

    /// <summary>
    /// "restore" is the only strategy that skips Mode. Unset (page not yet
    /// mounted) defaults to true — the safer of the two, since it costs an
    /// extra page rather than silently dropping one that turns out to matter.
    /// </summary>
    private static bool IsTargetedStrategy(UninstallSelection selection) =>
        string.IsNullOrEmpty(selection.Strategy) ||
        !string.Equals(selection.Strategy, "restore", StringComparison.OrdinalIgnoreCase);
}
