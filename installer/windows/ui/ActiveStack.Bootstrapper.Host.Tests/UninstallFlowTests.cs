using ActiveStack.Bootstrapper.Host.Navigation;
using Xunit;

namespace ActiveStack.Bootstrapper.Host.Tests;

public sealed class UninstallFlowTests
{
    [Fact]
    public void NextPage_WalksHubThroughConfirmToInstallingAndComplete()
    {
        var selection = new UninstallSelection();

        Assert.Equal(WizardPageId.UninstallAgents, UninstallFlow.NextPage(WizardPageId.Hub, selection));
        Assert.Equal(WizardPageId.UninstallMode, UninstallFlow.NextPage(WizardPageId.UninstallAgents, selection));
        Assert.Equal(WizardPageId.UninstallStrategy, UninstallFlow.NextPage(WizardPageId.UninstallMode, selection));
        Assert.Equal(WizardPageId.UninstallConfirm, UninstallFlow.NextPage(WizardPageId.UninstallStrategy, selection));
        Assert.Equal(WizardPageId.Installing, UninstallFlow.NextPage(WizardPageId.UninstallConfirm, selection));
        Assert.Equal(WizardPageId.Complete, UninstallFlow.NextPage(WizardPageId.Installing, selection));
    }

    [Fact]
    public void NextPage_FromComplete_StaysAtComplete()
    {
        var selection = new UninstallSelection();

        Assert.Equal(WizardPageId.Complete, UninstallFlow.NextPage(WizardPageId.Complete, selection));
    }

    [Fact]
    public void PreviousPage_IsTheInverseOfNextAcrossTheWholeFlow()
    {
        var selection = new UninstallSelection();

        Assert.Equal(WizardPageId.UninstallStrategy, UninstallFlow.PreviousPage(WizardPageId.UninstallConfirm, selection));
        Assert.Equal(WizardPageId.UninstallMode, UninstallFlow.PreviousPage(WizardPageId.UninstallStrategy, selection));
        Assert.Equal(WizardPageId.UninstallAgents, UninstallFlow.PreviousPage(WizardPageId.UninstallMode, selection));
        Assert.Equal(WizardPageId.Hub, UninstallFlow.PreviousPage(WizardPageId.UninstallAgents, selection));
    }

    [Fact]
    public void PreviousPage_FromInstalling_GoesToConfirm()
    {
        var selection = new UninstallSelection();

        Assert.Equal(WizardPageId.UninstallConfirm, UninstallFlow.PreviousPage(WizardPageId.Installing, selection));
    }
}
