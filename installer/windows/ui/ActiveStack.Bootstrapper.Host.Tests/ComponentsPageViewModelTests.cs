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

    [Fact]
    public void SelectedComponentId_DefaultsToFirstRenderedOptionsId()
    {
        var session = BuildSession();
        var selection = new InstallSelection();

        var page = new ComponentsPageViewModel(session, selection);

        Assert.Equal("permissions", page.SelectedComponentId);
    }

    [Fact]
    public void TogglingACheckbox_DoesNotChangeSelectedComponentIdOfADifferentAlreadySelectedRow()
    {
        var session = BuildSession();
        var selection = new InstallSelection();
        var page = new ComponentsPageViewModel(session, selection);

        page.SelectedComponentId = "openspec";
        var extra = Assert.Single(page.Choices, c => c.Id == "extra");
        extra.IsSelected = true;

        Assert.Equal("openspec", page.SelectedComponentId);
    }

    [Fact]
    public void SettingSelectedComponentId_DoesNotChangeAnyIsSelectedValue()
    {
        var session = BuildSession();
        var selection = new InstallSelection();
        var page = new ComponentsPageViewModel(session, selection);

        var before = page.Choices.ToDictionary(c => c.Id, c => c.IsSelected);

        page.SelectedComponentId = "extra";

        foreach (var choice in page.Choices)
        {
            Assert.Equal(before[choice.Id], choice.IsSelected);
        }
    }

    [Fact]
    public void DetailBody_ReflectsSelectedComponentsLongDescriptionOrFallsBackToDescription()
    {
        var session = BuildSession();
        var selection = new InstallSelection();
        var page = new ComponentsPageViewModel(session, selection);

        Assert.Equal("Basic protection", page.DetailTitle);
        Assert.Equal("Permissions enforces a security floor deny-list on every agent.", page.DetailBody);

        page.SelectedComponentId = "extra";

        Assert.Equal("Extra", page.DetailTitle);
        Assert.Equal("Not recommended by default.", page.DetailBody);
    }

    [Fact]
    public void SelectingAForcedRow_StillUpdatesSelectedComponentIdWhileCheckboxStaysLockedChecked()
    {
        // Triangulation fixture per tasks.md 9.4: 3 components — one forced/locked, two custom.
        var session = BuildSession();
        var selection = new InstallSelection();
        var page = new ComponentsPageViewModel(session, selection);

        page.SelectedComponentId = "openspec";
        page.SelectedComponentId = "permissions";

        var permissions = Assert.Single(page.Choices, c => c.Id == "permissions");
        Assert.Equal("permissions", page.SelectedComponentId);
        Assert.True(permissions.IsSelected);
        Assert.True(permissions.IsForced);
        Assert.Equal("Permissions enforces a security floor deny-list on every agent.", page.DetailBody);
    }

    private static InstallerSessionState BuildSession() =>
        new(
            AssistantChoices: [new AssistantChoice("claude", "Claude")],
            DefaultAssistantId: "claude",
            InstallTypeChoices: [new InstallTypeChoice("custom", "Custom", "Choose exactly what to install.")],
            RecommendedModeId: "custom",
            ForcedComponents:
            [
                new ComponentChoice("permissions", "Basic protection", "Helps avoid unsafe changes. Always on.", false, "Permissions enforces a security floor deny-list on every agent.")
            ],
            CustomComponents:
            [
                new ComponentChoice("openspec", "OpenSpec", "Plan and organize changes.", true),
                new ComponentChoice("extra", "Extra", "Not recommended by default.", false)
            ]);
}
