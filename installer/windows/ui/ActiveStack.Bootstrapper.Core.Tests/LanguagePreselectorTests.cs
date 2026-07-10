using ActiveStack.Bootstrapper.Core.Localization;
using Xunit;

namespace ActiveStack.Bootstrapper.Core.Tests;

public sealed class LanguagePreselectorTests
{
    [Fact]
    public void Resolve_NoPreference_OsCultureEsAr_ResolvesToSpanish()
    {
        var lang = LanguagePreselector.Resolve(persisted: null, osCultureReader: () => "es-AR");

        Assert.Equal("es", lang);
    }

    [Fact]
    public void Resolve_NoPreference_OsCultureFrench_ResolvesToEnglish()
    {
        var lang = LanguagePreselector.Resolve(persisted: null, osCultureReader: () => "fr");

        Assert.Equal("en", lang);
    }

    [Fact]
    public void Resolve_PersistedPreferenceWinsOverOsCulture()
    {
        var lang = LanguagePreselector.Resolve(persisted: "en", osCultureReader: () => "es-AR");

        Assert.Equal("en", lang);
    }

    [Fact]
    public void Resolve_NoPreferenceAndNoCultureSignal_DefaultsToEnglish()
    {
        var lang = LanguagePreselector.Resolve(persisted: null, osCultureReader: () => string.Empty);

        Assert.Equal("en", lang);
    }
}
