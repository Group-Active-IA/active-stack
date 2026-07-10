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

    [Fact]
    public void DetailBody_ReflectsSelectedStartersLongDescriptionWhenPresent()
    {
        var starters = BuildStarters();
        var selection = new StarterSelection();
        var page = new StarterCatalogPageViewModel(starters, selection);

        page.SelectedStarterId = "web";

        Assert.Equal("Web Starter", page.DetailTitle);
        Assert.Equal("Scaffolds a full-stack web app with auth and billing wired in.", page.DetailBody);
    }

    [Fact]
    public void DetailBody_FallsBackToShortDescriptionWhenLongDescriptionIsEmpty()
    {
        var starters = BuildStarters();
        var selection = new StarterSelection();
        var page = new StarterCatalogPageViewModel(starters, selection);

        page.SelectedStarterId = "cli";

        Assert.Equal("CLI Starter", page.DetailTitle);
        Assert.Equal("A CLI tool starter.", page.DetailBody);
    }

    [Fact]
    public void SelectingADifferentStarter_RaisesPropertyChangedForDetailTitleAndBody()
    {
        var starters = BuildStarters();
        var selection = new StarterSelection();
        var page = new StarterCatalogPageViewModel(starters, selection);

        var raised = new List<string>();
        page.PropertyChanged += (_, e) => raised.Add(e.PropertyName!);

        page.SelectedStarterId = "web";

        Assert.Contains(nameof(page.DetailTitle), raised);
        Assert.Contains(nameof(page.DetailBody), raised);
    }

    private static IReadOnlyList<StarterChoice> BuildStarters() =>
    [
        new StarterChoice("web", "Web Starter", "A web app starter.", [], ["claude", "opencode"], 2, "Scaffolds a full-stack web app with auth and billing wired in."),
        new StarterChoice("cli", "CLI Starter", "A CLI tool starter.", [], ["claude"], 1)
    ];
}
