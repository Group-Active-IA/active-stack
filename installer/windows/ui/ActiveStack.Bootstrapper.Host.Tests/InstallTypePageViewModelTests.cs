using ActiveStack.Bootstrapper.Core;
using ActiveStack.Bootstrapper.Host.Navigation;
using ActiveStack.Bootstrapper.Host.Pages.Install;
using Xunit;

namespace ActiveStack.Bootstrapper.Host.Tests;

public sealed class InstallTypePageViewModelTests
{
    [Fact]
    public void Constructor_FullIsPreselectedByDefault()
    {
        var session = BuildSession();
        var selection = new InstallSelection();

        var page = new InstallTypePageViewModel(session, selection);

        Assert.Equal("full", page.SelectedId);
        Assert.Equal("full", selection.Mode);
        Assert.True(page.CanAdvance);
    }

    [Fact]
    public void SelectingCustomCard_WritesModeToInstallSelection()
    {
        var session = BuildSession();
        var selection = new InstallSelection();
        var page = new InstallTypePageViewModel(session, selection);

        page.SelectedId = "custom";

        Assert.Equal("custom", selection.Mode);
        Assert.True(page.CanAdvance);
    }

    [Fact]
    public void ExistingSelection_IsHonoredInsteadOfResettingToRecommended()
    {
        var session = BuildSession();
        var selection = new InstallSelection { Mode = "custom" };

        var page = new InstallTypePageViewModel(session, selection);

        Assert.Equal("custom", page.SelectedId);
        Assert.Equal("custom", selection.Mode);
    }

    [Fact]
    public void DetailBody_ReturnsLongDescriptionOfSelectedChoiceWhenPresent()
    {
        var session = BuildSession();
        var selection = new InstallSelection();
        var page = new InstallTypePageViewModel(session, selection);

        Assert.Equal("Full recommended setup with all key tools, end to end.", page.DetailBody);
        Assert.Equal("Complete", page.DetailTitle);
    }

    [Fact]
    public void DetailBody_FallsBackToShortDescriptionWhenLongDescriptionIsEmpty()
    {
        var session = BuildSession();
        var selection = new InstallSelection();
        var page = new InstallTypePageViewModel(session, selection);

        page.SelectedId = "lite";

        Assert.Equal("Fast setup to start working right away.", page.DetailBody);
        Assert.Equal("Quick", page.DetailTitle);
    }

    [Fact]
    public void SelectingADifferentCard_RaisesPropertyChangedForDetailTitleAndBody()
    {
        var session = BuildSession();
        var selection = new InstallSelection();
        var page = new InstallTypePageViewModel(session, selection);

        var raised = new List<string>();
        page.PropertyChanged += (_, e) => raised.Add(e.PropertyName!);

        page.SelectedId = "lite";

        Assert.Contains(nameof(page.DetailTitle), raised);
        Assert.Contains(nameof(page.DetailBody), raised);
    }

    private static InstallerSessionState BuildSession() =>
        new(
            AssistantChoices: [new AssistantChoice("claude", "Claude")],
            DefaultAssistantId: "claude",
            InstallTypeChoices:
            [
                new InstallTypeChoice("lite", "Quick", "Fast setup to start working right away."),
                new InstallTypeChoice("full", "Complete", "Full recommended setup with all key tools.", "Full recommended setup with all key tools, end to end."),
                new InstallTypeChoice("custom", "Custom", "Choose exactly what to install.")
            ],
            RecommendedModeId: "full",
            ForcedComponents: [],
            CustomComponents: []);
}
