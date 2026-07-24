using ActiveStack.Bootstrapper.Core.Localization;
using ActiveStack.Bootstrapper.Host.Navigation;

namespace ActiveStack.Bootstrapper.Host.Pages;

/// <summary>
/// Terminal Install-flow page: maps the Progress page's outcome (terminal
/// snapshot + whether any step degraded + whether the pipeline rolled back)
/// to one of four states (D4, design.md). No footer nav — offers
/// close/next-steps instead. Title and the short <see cref="StateLabel"/>
/// come from <see cref="UiStrings"/> in the active language; the template
/// binds <see cref="StateLabel"/> rather than the raw enum
/// (gui-language-page, L4). The Error title is further split by
/// <paramref name="operation"/> (install/uninstall/starter) since a failed
/// uninstall showing "Installation failed" reads as nonsensical.
/// </summary>
public sealed class CompletePageViewModel : WizardPageViewModelBase
{
    public CompletePageViewModel(InstallProgressSnapshot terminalSnapshot, bool hadDegradedSteps, bool hadRollback, WizardOperation operation = WizardOperation.Install, string lang = "en")
        : base(TitleFor(terminalSnapshot, hadDegradedSteps, hadRollback, operation, lang), UiStrings.Get(lang, "page.complete.subtitle"), lang)
    {
        State = DetermineState(terminalSnapshot, hadDegradedSteps, hadRollback);
        StateLabel = UiStrings.Get(lang, $"complete.state.{StateKey(State)}.label");
        Message = BuildMessage(terminalSnapshot);
    }

    public CompleteState State { get; }

    /// <summary>Localized short label for <see cref="State"/> — what the template renders instead of the enum.</summary>
    public string StateLabel { get; }

    public string Message { get; }

    /// <summary>Always advanceable — the primary action here is "Close"/"Finish", not blocked by any gate.</summary>
    public override bool CanAdvance => true;

    /// <summary>
    /// Appends the terminal snapshot's <c>Details</c> (the real underlying
    /// reason — a step's Go pipeline error, or an unexpected-crash
    /// exception's message) to its generic localized <c>Message</c>, unless
    /// there is no Details or it merely repeats the message verbatim.
    /// </summary>
    private static string BuildMessage(InstallProgressSnapshot snapshot)
    {
        var message = snapshot.Message ?? string.Empty;
        if (string.IsNullOrWhiteSpace(snapshot.Details) || string.Equals(snapshot.Details, message, StringComparison.Ordinal))
        {
            return message;
        }

        return $"{message}\n\n{snapshot.Details}";
    }

    private static CompleteState DetermineState(InstallProgressSnapshot snapshot, bool hadDegradedSteps, bool hadRollback)
    {
        if (!snapshot.Success)
        {
            return hadRollback ? CompleteState.RolledBack : CompleteState.Error;
        }

        return hadDegradedSteps ? CompleteState.Degraded : CompleteState.Success;
    }

    private static string TitleFor(InstallProgressSnapshot snapshot, bool hadDegradedSteps, bool hadRollback, WizardOperation operation, string lang)
    {
        var state = DetermineState(snapshot, hadDegradedSteps, hadRollback);
        var key = state == CompleteState.Error
            ? $"complete.state.error.title.{OperationTitleSuffix(operation)}"
            : $"complete.state.{StateKey(state)}.title";
        return UiStrings.Get(lang, key);
    }

    /// <summary>Maps to the operation-specific error-title key suffix; unset/unknown operations keep the original "install" wording.</summary>
    private static string OperationTitleSuffix(WizardOperation operation) => operation switch
    {
        WizardOperation.Uninstall => "uninstall",
        WizardOperation.Starter => "starter",
        _ => "install"
    };

    private static string StateKey(CompleteState state) => state switch
    {
        CompleteState.Success => "success",
        CompleteState.Degraded => "degraded",
        CompleteState.RolledBack => "rolledback",
        CompleteState.Error => "error",
        _ => "success"
    };
}
