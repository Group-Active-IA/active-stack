using ActiveStack.Bootstrapper.Host.Navigation;
using Xunit;

namespace ActiveStack.Bootstrapper.Host.Tests;

public sealed class WizardFlowTests
{
    [Fact]
    public void NextPage_FromInstallType_FullWithTierCapableAgent_GoesToPermissions()
    {
        var selection = new InstallSelection { Mode = "full", Agents = ["claude"] };

        var next = WizardFlow.NextPage(WizardPageId.InstallType, selection, ["claude"]);

        Assert.Equal(WizardPageId.Permissions, next);
    }

    [Fact]
    public void NextPage_FromInstallType_CustomMode_GoesToComponents()
    {
        var selection = new InstallSelection { Mode = "custom", Agents = ["claude"] };

        var next = WizardFlow.NextPage(WizardPageId.InstallType, selection, ["claude"]);

        Assert.Equal(WizardPageId.Components, next);
    }

    [Fact]
    public void NextPage_FromInstallType_NonTierCapableOnly_SkipsPermissionsToReview()
    {
        var selection = new InstallSelection { Mode = "full", Agents = ["gemini"] };

        var next = WizardFlow.NextPage(WizardPageId.InstallType, selection, ["claude"]);

        Assert.Equal(WizardPageId.Review, next);
    }

    [Fact]
    public void NextPage_FromComponents_NonTierCapableOnly_SkipsPermissionsToReview()
    {
        var selection = new InstallSelection { Mode = "custom", Agents = ["gemini"] };

        var next = WizardFlow.NextPage(WizardPageId.Components, selection, ["claude"]);

        Assert.Equal(WizardPageId.Review, next);
    }

    [Fact]
    public void NextPage_FromComponents_TierCapableAgent_GoesToPermissions()
    {
        var selection = new InstallSelection { Mode = "custom", Agents = ["claude"] };

        var next = WizardFlow.NextPage(WizardPageId.Components, selection, ["claude"]);

        Assert.Equal(WizardPageId.Permissions, next);
    }

    [Fact]
    public void NextPage_FromPermissions_AlwaysGoesToReview()
    {
        var selection = new InstallSelection { Mode = "custom", Agents = ["claude"] };

        var next = WizardFlow.NextPage(WizardPageId.Permissions, selection, ["claude"]);

        Assert.Equal(WizardPageId.Review, next);
    }

    [Fact]
    public void PreviousPage_FromReview_GeminiOnlyFull_SkipsPermissionsAndComponentsToInstallType()
    {
        var selection = new InstallSelection { Mode = "full", Agents = ["gemini"] };

        var previous = WizardFlow.PreviousPage(WizardPageId.Review, selection, ["claude"]);

        Assert.Equal(WizardPageId.InstallType, previous);
    }

    [Fact]
    public void PreviousPage_FromReview_CustomTierCapable_GoesToPermissions()
    {
        var selection = new InstallSelection { Mode = "custom", Agents = ["claude"] };

        var previous = WizardFlow.PreviousPage(WizardPageId.Review, selection, ["claude"]);

        Assert.Equal(WizardPageId.Permissions, previous);
    }

    [Fact]
    public void PreviousPage_FromReview_CustomNonTierCapable_GoesToComponents()
    {
        var selection = new InstallSelection { Mode = "custom", Agents = ["gemini"] };

        var previous = WizardFlow.PreviousPage(WizardPageId.Review, selection, ["claude"]);

        Assert.Equal(WizardPageId.Components, previous);
    }

    [Fact]
    public void NextPage_FromHub_GoesToAssistants()
    {
        var selection = new InstallSelection();

        var next = WizardFlow.NextPage(WizardPageId.Hub, selection, []);

        Assert.Equal(WizardPageId.Assistants, next);
    }

    [Fact]
    public void PreviousPage_FromAssistants_GoesToHub()
    {
        var selection = new InstallSelection();

        var previous = WizardFlow.PreviousPage(WizardPageId.Assistants, selection, []);

        Assert.Equal(WizardPageId.Hub, previous);
    }
}
