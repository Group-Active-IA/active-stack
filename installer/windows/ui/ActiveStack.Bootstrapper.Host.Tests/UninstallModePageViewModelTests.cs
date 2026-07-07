using ActiveStack.Bootstrapper.Core;
using ActiveStack.Bootstrapper.Host.Navigation;
using ActiveStack.Bootstrapper.Host.Pages.Uninstall;
using Xunit;

namespace ActiveStack.Bootstrapper.Host.Tests;

public sealed class UninstallModePageViewModelTests
{
    [Fact]
    public void Constructor_DefaultsToFullAndCanAlwaysAdvance()
    {
        var options = BuildOptions();
        var selection = new UninstallSelection();

        var page = new UninstallModePageViewModel(options, selection);

        Assert.Equal("full", page.SelectedId);
        Assert.Equal("full", selection.Mode);
        Assert.True(page.CanAdvance);
    }

    [Fact]
    public void SelectingCustom_WritesSelectionMode()
    {
        var options = BuildOptions();
        var selection = new UninstallSelection();
        var page = new UninstallModePageViewModel(options, selection);

        page.SelectedId = "custom";

        Assert.Equal("custom", selection.Mode);
        Assert.True(page.CanAdvance);
    }

    private static UninstallOptions BuildOptions() =>
        new(["claude"],
            [
                new InstallTypeChoice("lite", "Quick", "Fast removal."),
                new InstallTypeChoice("full", "Complete", "Full removal."),
                new InstallTypeChoice("custom", "Custom", "Choose exactly what to remove.")
            ],
            [new UninstallStrategyChoice("targeted", "Targeted", "Undo only what Active Stack added.", true, false)]);
}
