using ActiveStack.Bootstrapper.Host.Navigation;
using Xunit;

namespace ActiveStack.Bootstrapper.Host.Tests;

public sealed class StarterFlowTests
{
    [Fact]
    public void NextPage_WalksHubThroughReviewToInstallingAndComplete()
    {
        var selection = new StarterSelection();

        Assert.Equal(WizardPageId.StarterCatalog, StarterFlow.NextPage(WizardPageId.Hub, selection));
        Assert.Equal(WizardPageId.StarterTarget, StarterFlow.NextPage(WizardPageId.StarterCatalog, selection));
        Assert.Equal(WizardPageId.StarterReview, StarterFlow.NextPage(WizardPageId.StarterTarget, selection));
        Assert.Equal(WizardPageId.Installing, StarterFlow.NextPage(WizardPageId.StarterReview, selection));
        Assert.Equal(WizardPageId.Complete, StarterFlow.NextPage(WizardPageId.Installing, selection));
    }

    [Fact]
    public void NextPage_FromComplete_StaysAtComplete()
    {
        var selection = new StarterSelection();

        Assert.Equal(WizardPageId.Complete, StarterFlow.NextPage(WizardPageId.Complete, selection));
    }

    [Fact]
    public void PreviousPage_IsTheInverseOfNextAcrossTheWholeFlow()
    {
        var selection = new StarterSelection();

        Assert.Equal(WizardPageId.StarterTarget, StarterFlow.PreviousPage(WizardPageId.StarterReview, selection));
        Assert.Equal(WizardPageId.StarterCatalog, StarterFlow.PreviousPage(WizardPageId.StarterTarget, selection));
        Assert.Equal(WizardPageId.Hub, StarterFlow.PreviousPage(WizardPageId.StarterCatalog, selection));
    }

    [Fact]
    public void PreviousPage_FromInstalling_GoesToReview()
    {
        var selection = new StarterSelection();

        Assert.Equal(WizardPageId.StarterReview, StarterFlow.PreviousPage(WizardPageId.Installing, selection));
    }
}
