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

    [Fact]
    public void DetailBody_ReflectsSelectedModesLongDescriptionWhenPresent()
    {
        var options = BuildOptions();
        var selection = new UninstallSelection();
        var page = new UninstallModePageViewModel(options, selection);

        Assert.Equal("Complete", page.DetailTitle);
        Assert.Equal("Removes every Active Stack harness the installer added.", page.DetailBody);
    }

    [Fact]
    public void DetailBody_FallsBackToShortDescriptionWhenLongDescriptionIsEmpty()
    {
        var options = BuildOptions();
        var selection = new UninstallSelection();
        var page = new UninstallModePageViewModel(options, selection);

        page.SelectedId = "lite";

        Assert.Equal("Quick", page.DetailTitle);
        Assert.Equal("Fast removal.", page.DetailBody);
    }

    [Fact]
    public void SelectingADifferentMode_RaisesPropertyChangedForDetailTitleAndBody()
    {
        var options = BuildOptions();
        var selection = new UninstallSelection();
        var page = new UninstallModePageViewModel(options, selection);

        var raised = new List<string>();
        page.PropertyChanged += (_, e) => raised.Add(e.PropertyName!);

        page.SelectedId = "lite";

        Assert.Contains(nameof(page.DetailTitle), raised);
        Assert.Contains(nameof(page.DetailBody), raised);
    }

    private static UninstallOptions BuildOptions() =>
        new(["claude"],
            [
                new InstallTypeChoice("lite", "Quick", "Fast removal."),
                new InstallTypeChoice("full", "Complete", "Full removal.", "Removes every Active Stack harness the installer added."),
                new InstallTypeChoice("custom", "Custom", "Choose exactly what to remove.")
            ],
            [new UninstallStrategyChoice("targeted", "Targeted", "Undo only what Active Stack added.", true, false)]);
}
