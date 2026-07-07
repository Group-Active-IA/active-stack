using ActiveStack.Bootstrapper.Core;
using ActiveStack.Bootstrapper.Host.Navigation;
using ActiveStack.Bootstrapper.Host.Pages.Install;
using Xunit;

namespace ActiveStack.Bootstrapper.Host.Tests;

public sealed class ComponentsPageViewModelTests
{
    [Fact]
    public void Constructor_RecommendedComponentsArePrecheckedAndPermissionsIsForced()
    {
        var session = BuildSession();
        var selection = new InstallSelection();

        var page = new ComponentsPageViewModel(session, selection);

        var permissions = Assert.Single(page.Choices, c => c.Id == "permissions");
        Assert.True(permissions.IsSelected);
        Assert.True(permissions.IsForced);

        var openspec = Assert.Single(page.Choices, c => c.Id == "openspec");
        Assert.True(openspec.IsSelected);
        Assert.True(openspec.IsRecommended);

        var extra = Assert.Single(page.Choices, c => c.Id == "extra");
        Assert.False(extra.IsSelected);

        Assert.Contains("permissions", selection.CustomIds);
        Assert.Contains("openspec", selection.CustomIds);
        Assert.DoesNotContain("extra", selection.CustomIds);
    }

    [Fact]
    public void TogglingPermissionsOff_StaysCheckedAndKeptInCustomIds()
    {
        var session = BuildSession();
        var selection = new InstallSelection();
        var page = new ComponentsPageViewModel(session, selection);

        var permissions = Assert.Single(page.Choices, c => c.Id == "permissions");
        permissions.IsSelected = false;

        Assert.True(permissions.IsSelected);
        Assert.Contains("permissions", selection.CustomIds);
    }

    [Fact]
    public void ExistingCustomIds_AreHonoredInsteadOfResettingToRecommended()
    {
        var session = BuildSession();
        var selection = new InstallSelection { CustomIds = ["permissions", "extra"] };

        var page = new ComponentsPageViewModel(session, selection);

        var openspec = Assert.Single(page.Choices, c => c.Id == "openspec");
        var extra = Assert.Single(page.Choices, c => c.Id == "extra");
        Assert.False(openspec.IsSelected);
        Assert.True(extra.IsSelected);
    }

    private static InstallerSessionState BuildSession() =>
        new(
            AssistantChoices: [new AssistantChoice("claude", "Claude")],
            DefaultAssistantId: "claude",
            InstallTypeChoices: [new InstallTypeChoice("custom", "Custom", "Choose exactly what to install.")],
            RecommendedModeId: "custom",
            ForcedComponents:
            [
                new ComponentChoice("permissions", "Basic protection", "Helps avoid unsafe changes. Always on.", false)
            ],
            CustomComponents:
            [
                new ComponentChoice("openspec", "OpenSpec", "Plan and organize changes.", true),
                new ComponentChoice("extra", "Extra", "Not recommended by default.", false)
            ]);
}
