namespace ActiveStack.Bootstrapper.Core.Localization;

/// <summary>
/// Resolves the wizard's initial language in priority order: persisted
/// preference &gt; OS UI culture &gt; English. Pure and injectable — the OS
/// culture reader is a delegate so detection is deterministic in tests
/// (gui-language-page, L4). Production callers pass
/// <c>() =&gt; System.Globalization.CultureInfo.CurrentUICulture.Name</c>.
/// </summary>
public static class LanguagePreselector
{
    public static string Resolve(string? persisted, Func<string> osCultureReader)
    {
        if (!string.IsNullOrWhiteSpace(persisted))
        {
            return persisted;
        }

        var culture = osCultureReader();
        return IsSpanish(culture) ? "es" : "en";
    }

    private static bool IsSpanish(string culture) =>
        !string.IsNullOrWhiteSpace(culture) &&
        culture.Length >= 2 &&
        culture[..2].Equals("es", StringComparison.OrdinalIgnoreCase);
}
