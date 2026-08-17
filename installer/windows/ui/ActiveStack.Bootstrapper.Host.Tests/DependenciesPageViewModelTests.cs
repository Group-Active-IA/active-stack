using ActiveStack.Bootstrapper.Core;
using ActiveStack.Bootstrapper.Host.Pages.Install;
using Xunit;

namespace ActiveStack.Bootstrapper.Host.Tests;

public sealed class DependenciesPageViewModelTests
{
    [Fact]
    public void Constructor_ExposesSessionDependenciesAndCanAlwaysAdvance()
    {
        var session = BuildSession(
        [
            new DependencyChoice("git", true, true, "2.43.0", "", ""),
            new DependencyChoice("node", true, false, "", "winget install OpenJS.NodeJS.LTS", "winget install --id OpenJS.NodeJS.LTS -e")
        ]);

        var page = new DependenciesPageViewModel(session);

        Assert.Equal(2, page.Rows.Count);
        Assert.Contains(page.Rows, r => r.Name == "git");
        Assert.Contains(page.Rows, r => r.Name == "node");
        Assert.True(page.CanAdvance);
    }

    [Fact]
    public void CanAdvance_StaysTrueEvenWithMissingRequiredDependency()
    {
        var session = BuildSession(
        [
            new DependencyChoice("git", true, false, "", "install git from https://git-scm.com/", "")
        ]);

        var page = new DependenciesPageViewModel(session);

        Assert.True(page.CanAdvance);
    }

    [Fact]
    public void MissingDependency_ExposesInstallHintAndCommandAsDisplayStrings()
    {
        var session = BuildSession(
        [
            new DependencyChoice("git", true, false, "", "install git from https://git-scm.com/", "winget install --id Git.Git -e --accept-source-agreements --accept-package-agreements")
        ]);

        var page = new DependenciesPageViewModel(session);

        var git = Assert.Single(page.Rows);
        Assert.Equal("install git from https://git-scm.com/", git.InstallHint);
        Assert.Equal("winget install --id Git.Git -e --accept-source-agreements --accept-package-agreements", git.InstallCommand);
        Assert.True(git.HasInstallHint);
        Assert.True(git.HasInstallCommand);
    }

    [Fact]
    public void InstalledDependency_HasNoInstallHintOrCommand()
    {
        var session = BuildSession(
        [
            new DependencyChoice("git", true, true, "2.43.0", "install git from https://git-scm.com/", "")
        ]);

        var page = new DependenciesPageViewModel(session);

        var git = Assert.Single(page.Rows);
        Assert.False(git.HasInstallHint);
        Assert.False(git.HasInstallCommand);
    }

    private static InstallerSessionState BuildSession(System.Collections.Generic.IReadOnlyList<DependencyChoice> dependencies) =>
        new(
            AssistantChoices: [],
            DefaultAssistantId: null,
            InstallTypeChoices: [],
            RecommendedModeId: null,
            ForcedComponents: [],
            CustomComponents: [],
            Dependencies: dependencies);
}
