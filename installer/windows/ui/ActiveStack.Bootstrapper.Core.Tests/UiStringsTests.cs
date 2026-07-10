using ActiveStack.Bootstrapper.Core.Localization;
using Xunit;

namespace ActiveStack.Bootstrapper.Core.Tests;

public sealed class UiStringsTests
{
    [Fact]
    public void Get_KnownKey_ResolvesInBothLanguages()
    {
        Assert.Equal("Next", UiStrings.Get("en", "shell.next"));
        Assert.Equal("Siguiente", UiStrings.Get("es", "shell.next"));
    }

    [Fact]
    public void Get_ShellFinishKey_ResolvesInBothLanguages()
    {
        Assert.Equal("Finish", UiStrings.Get("en", "shell.finish"));
        Assert.Equal("Finalizar", UiStrings.Get("es", "shell.finish"));
    }

    [Fact]
    public void Get_MissingInRequestedLanguage_FallsBackToEnglish()
    {
        // "zz" is not a supported language table; falling back to English
        // exercises the same fallback path as a key missing only in es.
        Assert.Equal(UiStrings.Get("en", "shell.next"), UiStrings.Get("zz", "shell.next"));
    }

    [Fact]
    public void Get_MissingEverywhere_ReturnsTheKeyItself()
    {
        Assert.Equal("no.such.key", UiStrings.Get("en", "no.such.key"));
        Assert.Equal("no.such.key", UiStrings.Get("es", "no.such.key"));
    }

    [Fact]
    public void Parity_EnAndEsTablesHaveIdenticalKeySetsAndNoEmptyValues()
    {
        var enKeys = UiStrings.EnglishKeys;
        var esKeys = UiStrings.SpanishKeys;

        Assert.Equal(enKeys.OrderBy(k => k, StringComparer.Ordinal), esKeys.OrderBy(k => k, StringComparer.Ordinal));

        foreach (var key in enKeys)
        {
            Assert.False(string.IsNullOrEmpty(UiStrings.Get("en", key)), $"en[{key}] must not be empty");
            Assert.False(string.IsNullOrEmpty(UiStrings.Get("es", key)), $"es[{key}] must not be empty");
        }
    }
}
