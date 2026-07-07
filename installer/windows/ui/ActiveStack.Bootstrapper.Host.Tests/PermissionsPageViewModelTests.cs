using ActiveStack.Bootstrapper.Core;
using ActiveStack.Bootstrapper.Host.Navigation;
using ActiveStack.Bootstrapper.Host.Pages.Install;
using Xunit;

namespace ActiveStack.Bootstrapper.Host.Tests;

public sealed class PermissionsPageViewModelTests
{
    [Fact]
    public void Constructor_BalanceadoIsPreselectedByDefault()
    {
        var session = BuildSession();
        var selection = new InstallSelection();

        var page = new PermissionsPageViewModel(session, selection);

        Assert.Equal("balanceado", page.SelectedTierId);
        Assert.Equal("balanceado", selection.Tier);
        Assert.Null(page.WarningText);
        Assert.True(page.CanAdvance);
    }

    [Fact]
    public void SelectingBypass_SurfacesItsWarningTextAndWritesTier()
    {
        var session = BuildSession();
        var selection = new InstallSelection();
        var page = new PermissionsPageViewModel(session, selection);

        page.SelectedTierId = "bypass";

        Assert.Equal("bypass", selection.Tier);
        Assert.Equal("Bypass: autonomous mode — the security floor still applies (C-21)", page.WarningText);
    }

    [Fact]
    public void ExistingTier_IsHonoredInsteadOfResettingToDefault()
    {
        var session = BuildSession();
        var selection = new InstallSelection { Tier = "bypass" };

        var page = new PermissionsPageViewModel(session, selection);

        Assert.Equal("bypass", page.SelectedTierId);
        Assert.Equal("Bypass: autonomous mode — the security floor still applies (C-21)", page.WarningText);
    }

    private static InstallerSessionState BuildSession() =>
        new(
            AssistantChoices: [new AssistantChoice("claude", "Claude")],
            DefaultAssistantId: "claude",
            InstallTypeChoices: [new InstallTypeChoice("full", "Complete", "Full recommended setup with all key tools.")],
            RecommendedModeId: "full",
            ForcedComponents: [],
            CustomComponents: [],
            TierCapable: true,
            TierCapableAgents: ["claude"],
            PermissionTierChoices:
            [
                new PermissionTierChoice("estricto", "Estricto", "Ask before every change.", false, null),
                new PermissionTierChoice("balanceado", "Balanceado", "Ask for risky changes only.", true, null),
                new PermissionTierChoice("bypass", "Bypass", "Never ask.", false, "Bypass: autonomous mode — the security floor still applies (C-21)")
            ]);
}
