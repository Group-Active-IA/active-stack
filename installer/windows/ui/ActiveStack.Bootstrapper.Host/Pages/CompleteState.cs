namespace ActiveStack.Bootstrapper.Host.Pages;

/// <summary>Terminal state the Complete page surfaces from the install stream's outcome.</summary>
public enum CompleteState
{
    Success,
    Degraded,
    RolledBack,
    Error
}
