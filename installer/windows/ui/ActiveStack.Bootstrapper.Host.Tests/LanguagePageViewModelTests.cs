using ActiveStack.Bootstrapper.Host.Pages;
using Xunit;

namespace ActiveStack.Bootstrapper.Host.Tests;

public sealed class LanguagePageViewModelTests
{
    [Fact]
    public void Constructor_HasFixedBilingualTitle()
    {
        var page = new LanguagePageViewModel(preselectedLanguageId: "en");

        Assert.Equal("Select your language / Elegí tu idioma", page.Title);
    }

    [Fact]
    public void Constructor_ExposesAnExtensibleChoiceListIncludingEnglishAndSpanish()
    {
        var page = new LanguagePageViewModel(preselectedLanguageId: "en");

        Assert.Contains(page.Choices, c => c.Id == "en" && c.Label == "English");
        Assert.Contains(page.Choices, c => c.Id == "es" && c.Label == "Español");
    }

    [Fact]
    public void Constructor_InitialSelectionFollowsThePreselection()
    {
        var page = new LanguagePageViewModel(preselectedLanguageId: "es");

        Assert.Equal("es", page.SelectedLanguageId);
    }

    [Fact]
    public void CanAdvance_IsTrueWithALanguageSelected()
    {
        var page = new LanguagePageViewModel(preselectedLanguageId: "en");

        Assert.True(page.CanAdvance);
    }
}
