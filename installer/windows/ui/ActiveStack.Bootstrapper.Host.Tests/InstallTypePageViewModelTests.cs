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

    private static InstallerSessionState BuildSession() =>
        new(
            AssistantChoices: [new AssistantChoice("claude", "Claude")],
            DefaultAssistantId: "claude",
            InstallTypeChoices:
            [
                new InstallTypeChoice("lite", "Quick", "Fast setup to start working right away."),
                new InstallTypeChoice("full", "Complete", "Full recommended setup with all key tools."),
                new InstallTypeChoice("custom", "Custom", "Choose exactly what to install.")
            ],
            RecommendedModeId: "full",
            ForcedComponents: [],
            CustomComponents: []);
}
