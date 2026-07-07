using ActiveStack.Bootstrapper.Core;
using ActiveStack.Bootstrapper.Host.Navigation;
using ActiveStack.Bootstrapper.Host.Pages.Starters;
using Xunit;

namespace ActiveStack.Bootstrapper.Host.Tests;

public sealed class StarterCatalogPageViewModelTests
{
    [Fact]
    public void NothingSelected_BlocksAdvancing()
    {
        var starters = BuildStarters();
        var selection = new StarterSelection();

        var page = new StarterCatalogPageViewModel(starters, selection);

        Assert.False(page.CanAdvance);
        Assert.Equal(2, page.Choices.Count);
    }

    [Fact]
    public void SelectingAStarter_WritesSelectionAndEnablesAdvancing()
    {
        var starters = BuildStarters();
        var selection = new StarterSelection();
        var page = new StarterCatalogPageViewModel(starters, selection);

        page.SelectedStarterId = "web";

        Assert.Equal("web", selection.StarterId);
        Assert.True(page.CanAdvance);
    }

    private static IReadOnlyList<StarterChoice> BuildStarters() =>
    [
        new StarterChoice("web", "Web Starter", "A web app starter.", [], ["claude", "opencode"], 2),
        new StarterChoice("cli", "CLI Starter", "A CLI tool starter.", [], ["claude"], 1)
    ];
}
