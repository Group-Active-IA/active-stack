using ActiveStack.Bootstrapper.Core;
using ActiveStack.Bootstrapper.Host.Navigation;
using ActiveStack.Bootstrapper.Host.Pages.Uninstall;
using Xunit;

namespace ActiveStack.Bootstrapper.Host.Tests;

public sealed class UninstallAgentsPageViewModelTests
{
    [Fact]
    public void DetectedAgents_ArePreCheckedAndCanAdvance()
    {
        var options = BuildOptions(["claude"]);
        var selection = new UninstallSelection();

        var page = new UninstallAgentsPageViewModel(options, selection);

        var claude = Assert.Single(page.Choices);
        Assert.True(claude.IsSelected);
        Assert.True(page.CanAdvance);
        Assert.Equal(["claude"], selection.Agents);
    }

    [Fact]
    public void UncheckingEveryAgent_BlocksAdvancing()
    {
        var options = BuildOptions(["claude", "opencode"]);
        var selection = new UninstallSelection();
        var page = new UninstallAgentsPageViewModel(options, selection);

        foreach (var choice in page.Choices)
        {
            choice.IsSelected = false;
        }

        Assert.False(page.CanAdvance);
        Assert.Empty(selection.Agents);
    }

    private static UninstallOptions BuildOptions(IReadOnlyList<string> detectedAgents) =>
        new(detectedAgents,
            [new InstallTypeChoice("full", "Complete", "Full removal.")],
            [new UninstallStrategyChoice("targeted", "Targeted", "Undo only what Active Stack added.", true, false)]);
}
