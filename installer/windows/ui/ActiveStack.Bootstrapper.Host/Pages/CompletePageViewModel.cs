namespace ActiveStack.Bootstrapper.Host.Pages;

/// <summary>
/// Terminal Install-flow page: maps the Progress page's outcome (terminal
/// snapshot + whether any step degraded + whether the pipeline rolled back)
/// to one of four states (D4, design.md). No footer nav — offers
/// close/next-steps instead.
/// </summary>
public sealed class CompletePageViewModel : WizardPageViewModelBase
{
    public CompletePageViewModel(InstallProgressSnapshot terminalSnapshot, bool hadDegradedSteps, bool hadRollback)
        : base(TitleFor(terminalSnapshot, hadDegradedSteps, hadRollback), "Here's how your Active Stack setup went.")
    {
        State = DetermineState(terminalSnapshot, hadDegradedSteps, hadRollback);
        Message = terminalSnapshot.Message ?? string.Empty;
    }

    public CompleteState State { get; }

    public string Message { get; }

    /// <summary>Always advanceable — the primary action here is "Close"/"Finish", not blocked by any gate.</summary>
    public override bool CanAdvance => true;

    private static CompleteState DetermineState(InstallProgressSnapshot snapshot, bool hadDegradedSteps, bool hadRollback)
    {
        if (!snapshot.Success)
        {
            return hadRollback ? CompleteState.RolledBack : CompleteState.Error;
        }

        return hadDegradedSteps ? CompleteState.Degraded : CompleteState.Success;
    }

    private static string TitleFor(InstallProgressSnapshot snapshot, bool hadDegradedSteps, bool hadRollback) =>
        DetermineState(snapshot, hadDegradedSteps, hadRollback) switch
        {
            CompleteState.Success => "Installation complete",
            CompleteState.Degraded => "Installation complete (with some best-effort steps skipped)",
            CompleteState.RolledBack => "Installation rolled back",
            CompleteState.Error => "Installation failed",
            _ => "Installation complete"
        };
}
