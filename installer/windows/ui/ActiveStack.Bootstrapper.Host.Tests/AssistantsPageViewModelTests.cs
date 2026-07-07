using ActiveStack.Bootstrapper.Core;
using ActiveStack.Bootstrapper.Host.Navigation;
using ActiveStack.Bootstrapper.Host.Pages.Install;
using Xunit;

namespace ActiveStack.Bootstrapper.Host.Tests;

public sealed class AssistantsPageViewModelTests
{
    [Fact]
    public void Constructor_DetectedAgentIsPrecheckedAndCanAdvance()
    {
        var session = BuildSession();
        var selection = new InstallSelection();

        var page = new AssistantsPageViewModel(session, selection);

        var claude = Assert.Single(page.Choices, c => c.Id == "claude");
        Assert.True(claude.IsSelected);
        Assert.True(claude.IsRecommended);
        Assert.True(page.CanAdvance);
        Assert.Contains("claude", selection.Agents);
    }

    [Fact]
    public void UnselectingEveryAgent_CannotAdvance()
    {
        var session = BuildSession();
        var selection = new InstallSelection();
        var page = new AssistantsPageViewModel(session, selection);

        foreach (var choice in page.Choices)
        {
            choice.IsSelected = false;
        }

        Assert.False(page.CanAdvance);
        Assert.Empty(selection.Agents);
    }

    [Fact]
    public void SelectingASecondAgent_WritesBothToInstallSelection()
    {
        var session = BuildSession();
        var selection = new InstallSelection();
        var page = new AssistantsPageViewModel(session, selection);

        var codex = Assert.Single(page.Choices, c => c.Id == "codex");
        codex.IsSelected = true;

        Assert.True(page.CanAdvance);
        Assert.Equal(["claude", "codex"], selection.Agents);
    }

    [Fact]
    public void ExistingSelection_IsHonoredInsteadOfReprecheckingEveryAgent()
    {
        var session = BuildSession();
        var selection = new InstallSelection { Agents = ["codex"] };

        var page = new AssistantsPageViewModel(session, selection);

        var claude = Assert.Single(page.Choices, c => c.Id == "claude");
        var codex = Assert.Single(page.Choices, c => c.Id == "codex");
        Assert.False(claude.IsSelected);
        Assert.True(codex.IsSelected);
    }

    private static InstallerSessionState BuildSession() =>
        new(
            AssistantChoices:
            [
                new AssistantChoice("claude", "Claude"),
                new AssistantChoice("codex", "Codex")
            ],
            DefaultAssistantId: "claude",
            InstallTypeChoices:
            [
                new InstallTypeChoice("lite", "Quick", "Fast setup to start working right away."),
                new InstallTypeChoice("full", "Complete", "Full recommended setup with all key tools.")
            ],
            RecommendedModeId: "full",
            ForcedComponents: [],
            CustomComponents: []);
}
