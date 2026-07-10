using ActiveStack.Bootstrapper.Host.Pages;
using Xunit;

namespace ActiveStack.Bootstrapper.Host.Tests;

public sealed class CompletePageViewModelTests
{
    [Fact]
    public void Success_WhenFinishedSucceededWithNoDegradedSteps()
    {
        var snapshot = new InstallProgressSnapshot("install_finished", "install", null, "Installation finished successfully.", true);

        var page = new CompletePageViewModel(snapshot, hadDegradedSteps: false, hadRollback: false);

        Assert.Equal(CompleteState.Success, page.State);
        Assert.True(page.CanAdvance);
    }

    [Fact]
    public void Degraded_WhenFinishedSucceededButSomeStepDegraded()
    {
        var snapshot = new InstallProgressSnapshot("install_finished", "install", null, "Installation finished successfully.", true);

        var page = new CompletePageViewModel(snapshot, hadDegradedSteps: true, hadRollback: false);

        Assert.Equal(CompleteState.Degraded, page.State);
    }

    [Fact]
    public void RolledBack_WhenFailedAfterEnteringRollback()
    {
        var snapshot = new InstallProgressSnapshot("install_finished", "rollback", null, "Installation failed.", false);

        var page = new CompletePageViewModel(snapshot, hadDegradedSteps: false, hadRollback: true);

        Assert.Equal(CompleteState.RolledBack, page.State);
    }

    [Fact]
    public void Error_WhenFailedWithoutRollback()
    {
        var snapshot = new InstallProgressSnapshot("install_finished", "install", null, "Installation failed.", false);

        var page = new CompletePageViewModel(snapshot, hadDegradedSteps: false, hadRollback: false);

        Assert.Equal(CompleteState.Error, page.State);
    }

    [Fact]
    public void StateLabel_DefaultsToTheEnglishLocalizedShortLabel()
    {
        var snapshot = new InstallProgressSnapshot("install_finished", "install", null, "Installation finished successfully.", true);

        var page = new CompletePageViewModel(snapshot, hadDegradedSteps: false, hadRollback: false);

        Assert.Equal("Success", page.StateLabel);
        Assert.Equal("Installation complete", page.Title);
    }

    [Fact]
    public void StateLabel_SpanishLanguage_LocalizesTheShortLabelAndTitle()
    {
        var snapshot = new InstallProgressSnapshot("install_finished", "install", null, "Instalación finalizada con éxito.", true);

        var page = new CompletePageViewModel(snapshot, hadDegradedSteps: false, hadRollback: false, lang: "es");

        Assert.Equal("Éxito", page.StateLabel);
        Assert.Equal("Instalación completa", page.Title);
    }
}
