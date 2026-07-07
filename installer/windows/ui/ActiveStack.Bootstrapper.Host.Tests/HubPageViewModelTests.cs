using ActiveStack.Bootstrapper.Host.Pages;
using Xunit;

namespace ActiveStack.Bootstrapper.Host.Tests;

public sealed class HubPageViewModelTests
{
    [Fact]
    public void SelectingInstall_EnablesAdvancing()
    {
        var page = new HubPageViewModel();

        page.SelectedEntryId = "install";

        Assert.Equal("install", page.SelectedEntryId);
        Assert.True(page.CanAdvance);
    }

    [Fact]
    public void StartersBackupsUninstall_AreEnabledAndSelectable()
    {
        var page = new HubPageViewModel();

        var starters = Assert.Single(page.Entries, e => e.Id == "starters");
        var backups = Assert.Single(page.Entries, e => e.Id == "backups");
        var uninstall = Assert.Single(page.Entries, e => e.Id == "uninstall");

        Assert.True(starters.IsEnabled);
        Assert.True(backups.IsEnabled);
        Assert.True(uninstall.IsEnabled);

        page.SelectedEntryId = "starters";

        Assert.Equal("starters", page.SelectedEntryId);
        Assert.True(page.CanAdvance);
    }

    [Fact]
    public void UpdateStack_StaysDisabledWithTooltipAndDoesNotNavigate()
    {
        var page = new HubPageViewModel();

        var update = Assert.Single(page.Entries, e => e.Id == "update");

        Assert.False(update.IsEnabled);
        Assert.False(string.IsNullOrWhiteSpace(update.Tooltip));

        page.SelectedEntryId = "update";

        Assert.Null(page.SelectedEntryId);
        Assert.False(page.CanAdvance);
    }

    [Fact]
    public void InstallEntry_IsPresentAndEnabled()
    {
        var page = new HubPageViewModel();

        var install = Assert.Single(page.Entries, e => e.Id == "install");

        Assert.True(install.IsEnabled);
    }
}
