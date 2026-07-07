using ActiveStack.Bootstrapper.Host.Navigation;

namespace ActiveStack.Bootstrapper.Host.Pages;

/// <summary>
/// Capability shared by every page that, when advanced, starts a long-running
/// engine stream (Install's Review, Uninstall's Confirm, Starter's Review):
/// a footer label and a method returning the
/// <see cref="InstallProgressSnapshot"/> stream to hand to the shared
/// <see cref="ProgressPageViewModel"/> (D9, design.md). Lets
/// <see cref="ShellViewModel"/> run any of the three flows' terminal pages
/// through one generic "trigger page → stream → progress → complete" method
/// instead of three near-duplicate run methods.
/// </summary>
public interface IStreamTriggerPage : IWizardPage
{
    /// <summary>The shell's primary footer action label while this page is current.</summary>
    string PrimaryLabel { get; }

    IAsyncEnumerable<InstallProgressSnapshot> StartStream(CancellationToken cancellationToken = default);
}
