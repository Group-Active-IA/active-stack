namespace ActiveStack.Bootstrapper.Host.Navigation;

/// <summary>
/// Mutable model shared across every Install-flow page view-model: agents,
/// mode, custom component ids and permission tier. Created once when the
/// Install flow starts (from the Hub) and threaded through
/// <see cref="ShellViewModel"/> and the page view-models (D9, design.md).
/// <see cref="WizardFlow"/> only ever reads it — it never mutates it.
/// </summary>
public sealed class InstallSelection
{
    public List<string> Agents { get; set; } = [];

    /// <summary>
    /// Empty until <see cref="Pages.Install.InstallTypePageViewModel"/> is
    /// first constructed, which seeds it from the session's recommended
    /// mode. Deliberately not defaulted to "full" here — that would make it
    /// indistinguishable from an explicit user choice of "full" when the
    /// page is reconstructed after a Back navigation (see
    /// InstallTypePageViewModel's existing-selection honoring).
    /// </summary>
    public string Mode { get; set; } = string.Empty;

    public List<string> CustomIds { get; set; } = [];

    public string? Tier { get; set; }
}
