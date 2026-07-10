namespace ActiveStack.Bootstrapper.Host;

internal sealed record BootstrapperLaunchOptions(
    bool AutoInstall,
    bool AutoCloseWhenFinished,
    bool TraceEnabled,
    string? AssistantId,
    string? InstallModeId,
    string? LanguageOverride)
{
    private static readonly string[] SupportedLanguages = ["en", "es"];

    public static BootstrapperLaunchOptions ReadFromEnvironment(Func<string, string?> environmentReader)
    {
        return new BootstrapperLaunchOptions(
            AutoInstall: ParseBoolean(environmentReader("ACTIVE_STACK_AUTO_INSTALL")),
            AutoCloseWhenFinished: ParseBoolean(environmentReader("ACTIVE_STACK_AUTO_CLOSE")),
            TraceEnabled: ParseBoolean(environmentReader("ACTIVE_STACK_BOOTSTRAPPER_TRACE")),
            AssistantId: Normalize(environmentReader("ACTIVE_STACK_INSTALL_ASSISTANT")),
            InstallModeId: Normalize(environmentReader("ACTIVE_STACK_INSTALL_MODE")),
            LanguageOverride: NormalizeLanguage(environmentReader("ACTIVE_STACK_UI_LANG")));
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

    /// <summary>
    /// Surfaces <c>ACTIVE_STACK_UI_LANG</c> as the automation override for
    /// the Language page's preselection, when set to a supported value
    /// (<c>es|en</c>) — unsupported values are ignored so automation typos
    /// fall back to the normal persisted/OS-culture resolution.
    /// </summary>
    private static string? NormalizeLanguage(string? value)
    {
        var normalized = Normalize(value)?.ToLowerInvariant();
        return normalized is not null && SupportedLanguages.Contains(normalized) ? normalized : null;
    }
}
