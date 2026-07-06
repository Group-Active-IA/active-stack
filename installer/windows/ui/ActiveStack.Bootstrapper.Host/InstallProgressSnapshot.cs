namespace ActiveStack.Bootstrapper.Host;

public sealed record InstallProgressSnapshot(
    string Type,
    string? Phase,
    string? StepId,
    string? Message,
    bool Success,
    string? Details = null,
    string? Timestamp = null);
