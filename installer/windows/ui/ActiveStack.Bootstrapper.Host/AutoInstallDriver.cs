using ActiveStack.Bootstrapper.Host.Navigation;
using ActiveStack.Bootstrapper.Host.Pages;
using ActiveStack.Bootstrapper.Host.Pages.Install;

namespace ActiveStack.Bootstrapper.Host;

/// <summary>
/// Drives the headless/CI automation path (ACTIVE_STACK_AUTO_INSTALL):
/// applies the assistant/mode overrides from launch options as each relevant
/// page is reached, advances through the wizard, and stops on Review ready
/// to be confirmed. Extracted out of <c>MainWindow.xaml.cs</c> (no WPF
/// dependency) so the loop is unit-testable without an STA/Application host
/// (gui-language-page, L4, task 8.1).
/// </summary>
public static class AutoInstallDriver
{
    /// <summary>
    /// Walks <paramref name="shell"/> from wherever it currently is (the
    /// Language page on a fresh shell) through to the Review page and
    /// confirms it. Overrides are applied fresh on EVERY iteration —
    /// crucially including the Hub's "install" entry selection — because
    /// advancing off the Language page swaps in a brand-new
    /// <see cref="HubPageViewModel"/> instance; selecting an entry on a page
    /// reference captured before that swap would be silently discarded
    /// (the bug this extraction fixes).
    /// </summary>
    public static async Task RunAsync(
        ShellViewModel shell,
        string? assistantOverride,
        string? installModeOverride,
        Action<string>? trace = null,
        CancellationToken cancellationToken = default)
    {
        while (shell.CurrentPage is not ReviewPageViewModel)
        {
            ApplyOverrides(shell, assistantOverride, installModeOverride);

            if (!shell.PrimaryEnabled)
            {
                trace?.Invoke("AutoInstallDriver stalled: current page cannot advance.");
                return;
            }

            await shell.AdvanceAsync(cancellationToken);
        }

        await shell.AdvanceAsync(cancellationToken);
    }

    private static void ApplyOverrides(ShellViewModel shell, string? assistantOverride, string? installModeOverride)
    {
        if (shell.CurrentPage is HubPageViewModel hub && string.IsNullOrEmpty(hub.SelectedEntryId))
        {
            hub.SelectedEntryId = "install";
        }

        if (shell.CurrentPage is AssistantsPageViewModel assistants &&
            !string.IsNullOrWhiteSpace(assistantOverride))
        {
            foreach (var choice in assistants.Choices)
            {
                choice.IsSelected = string.Equals(choice.Id, assistantOverride, StringComparison.OrdinalIgnoreCase);
            }
        }

        if (shell.CurrentPage is InstallTypePageViewModel installType &&
            !string.IsNullOrWhiteSpace(installModeOverride))
        {
            installType.SelectedId = installModeOverride!;
        }
    }
}
