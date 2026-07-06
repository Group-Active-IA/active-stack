namespace ActiveStack.Bootstrapper.Host;

internal sealed record BootstrapperLaunchOptions(
    bool AutoInstall,
    bool AutoCloseWhenFinished,
    bool TraceEnabled,
    string? AssistantId,
    string? InstallModeId)
{
    public static BootstrapperLaunchOptions ReadFromEnvironment(Func<string, string?> environmentReader)
    {
        return new BootstrapperLaunchOptions(
            AutoInstall: ParseBoolean(environmentReader("ACTIVE_STACK_AUTO_INSTALL")),
            AutoCloseWhenFinished: ParseBoolean(environmentReader("ACTIVE_STACK_AUTO_CLOSE")),
            TraceEnabled: ParseBoolean(environmentReader("ACTIVE_STACK_BOOTSTRAPPER_TRACE")),
            AssistantId: Normalize(environmentReader("ACTIVE_STACK_INSTALL_ASSISTANT")),
            InstallModeId: Normalize(environmentReader("ACTIVE_STACK_INSTALL_MODE")));
    }

    public static BootstrapperLaunchOptions ReadFromEnvironment()
        => ReadFromEnvironment(Environment.GetEnvironmentVariable);

    private static bool ParseBoolean(string? value)
    {
        if (string.IsNullOrWhiteSpace(value))
        {
            return false;
        }

        return value.Trim().ToLowerInvariant() is "1" or "true" or "yes" or "on";
    }

    private static string? Normalize(string? value)
        => string.IsNullOrWhiteSpace(value) ? null : value.Trim();
}
