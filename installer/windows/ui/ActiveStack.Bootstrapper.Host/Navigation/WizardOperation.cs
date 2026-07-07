namespace ActiveStack.Bootstrapper.Host.Navigation;

/// <summary>
/// Which destination the Hub has routed the shell into. Set from
/// <see cref="Pages.HubPageViewModel.SelectedEntryId"/> when the user
/// advances from the Hub; cleared back to <see cref="None"/> when the user
/// returns Back to the Hub from a flow's first page (D2, design.md).
/// Disambiguates the shared terminal pages (<see cref="WizardPageId.Installing"/>,
/// <see cref="WizardPageId.Complete"/>) so <see cref="ShellViewModel"/> knows
/// which operation's flow functions to dispatch Back/Next to.
/// </summary>
public enum WizardOperation
{
    None,
    Install,
    Uninstall,
    Starter,
    Backups
}
