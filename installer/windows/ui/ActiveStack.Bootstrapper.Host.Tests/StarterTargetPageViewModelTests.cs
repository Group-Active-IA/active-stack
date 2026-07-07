using ActiveStack.Bootstrapper.Host.Navigation;
using ActiveStack.Bootstrapper.Host.Pages.Starters;
using Xunit;

namespace ActiveStack.Bootstrapper.Host.Tests;

public sealed class StarterTargetPageViewModelTests
{
    [Fact]
    public void PickingAFolder_SetsProjectPathAndCanAdvanceWithDefaultAgents()
    {
        var selection = new StarterSelection();
        var picker = new FakeFolderPicker("C:\\Users\\dev\\My Project");
        var page = new StarterTargetPageViewModel(selection, picker);

        page.PickFolder();

        Assert.Equal("C:\\Users\\dev\\My Project", page.ProjectPath);
        Assert.Equal("C:\\Users\\dev\\My Project", selection.ProjectPath);
        Assert.True(page.CanAdvance);
        Assert.Equal(["claude", "opencode"], selection.Agents);
    }

    [Fact]
    public void NoDirectoryChosen_BlocksAdvancing()
    {
        var selection = new StarterSelection();
        var picker = new FakeFolderPicker(null);
        var page = new StarterTargetPageViewModel(selection, picker);

        Assert.False(page.CanAdvance);

        page.PickFolder();

        Assert.False(page.CanAdvance);
        Assert.Equal(string.Empty, page.ProjectPath);
    }

    private sealed class FakeFolderPicker : IFolderPicker
    {
        private readonly string? _path;

        public FakeFolderPicker(string? path)
        {
            _path = path;
        }

        public string? PickFolder() => _path;
    }
}
