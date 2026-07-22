using ActiveStack.Bootstrapper.Host.Pages;
using Xunit;

namespace ActiveStack.Bootstrapper.Host.Tests;

/// <summary>
/// Migrated (D6, design.md) from <c>MainWindowViewModelTests</c>'
/// <c>StartInstallAsync_TracksProgressAndMarksSuccess</c> and
/// <c>StartInstallAsync_MapsGenericEngineEventsToFriendlyStatus</c>.
/// Assertions on the snapshot-to-UI mapping (progress value, friendly
/// message, step label, activity log, terminal success) are kept
/// byte-identical; assertions that belonged to the old dashboard's
/// installing/selection state (not part of the migrated progress logic)
/// are dropped, since that behavior now lives in the wizard flow / Review
/// page, not the Progress page.
/// </summary>
public sealed class ProgressPageViewModelTests
{
    [Fact]
    public async Task ConsumeAsync_TracksProgressAndMarksSuccess()
    {
        var viewModel = new ProgressPageViewModel();

        await viewModel.ConsumeAsync(ToStream(
            new InstallProgressSnapshot("phase_started", "install", null, "Starting installation.", false),
            new InstallProgressSnapshot("step_started", "apply", "openspec", "Installing OpenSpec.", false),
            new InstallProgressSnapshot("install_finished", "install", null, "Installation finished successfully.", true)));

        Assert.True(viewModel.InstallSucceeded);
        Assert.Equal("Installation finished successfully.", viewModel.ProgressMessage);
        Assert.Equal("openspec", viewModel.CurrentStepId);
        Assert.Equal("OpenSpec", viewModel.CurrentStepLabel);
        Assert.Equal(100, viewModel.ProgressValue);
        Assert.True(viewModel.RecentActivity.Count >= 3);
        Assert.Equal("Starting installation.", viewModel.RecentActivity[0]);
        Assert.Contains("Installing OpenSpec.", viewModel.RecentActivity);
    }

    [Fact]
    public async Task ConsumeAsync_MapsGenericEngineEventsToFriendlyStatus()
    {
        var viewModel = new ProgressPageViewModel();

        await viewModel.ConsumeAsync(ToStream(
            new InstallProgressSnapshot("phase_started", "prepare", null, "Preparing installation.", false),
            new InstallProgressSnapshot("step_started", "apply", "external:engram", "Step started.", false),
            new InstallProgressSnapshot("download_started", "apply", "external:engram", "Downloading engram.", false),
            new InstallProgressSnapshot("step_succeeded", "apply", "external:engram", "Step completed.", false),
            new InstallProgressSnapshot("install_finished", "install", null, "Installation finished successfully.", true)));

        Assert.True(viewModel.InstallSucceeded);
        Assert.Equal("external:engram", viewModel.CurrentStepId);
        Assert.Equal("Engram", viewModel.CurrentStepLabel);
        Assert.Equal(100, viewModel.ProgressValue);
        Assert.Contains("Preparing installation.", viewModel.RecentActivity);
        Assert.Contains("Installing Engram.", viewModel.RecentActivity);
        Assert.Contains("Downloading Engram.", viewModel.RecentActivity);
        Assert.Contains("Installed Engram.", viewModel.RecentActivity);
    }

    [Fact]
    public async Task ConsumeAsync_TerminalFailureIsReported()
    {
        var viewModel = new ProgressPageViewModel();

        await viewModel.ConsumeAsync(ToStream(
            new InstallProgressSnapshot("phase_started", "install", null, "Starting installation.", false),
            new InstallProgressSnapshot("install_finished", "install", null, "Installation failed.", false)));

        Assert.False(viewModel.InstallSucceeded);
        Assert.True(viewModel.IsFinished);
        Assert.Equal(100, viewModel.ProgressValue);
    }

    [Fact]
    public async Task ConsumeAsync_TerminalFailureWithoutOwnDetails_IsEnrichedWithTheLastStepFailureDetail()
    {
        var viewModel = new ProgressPageViewModel();

        await viewModel.ConsumeAsync(ToStream(
            new InstallProgressSnapshot("phase_started", "apply", null, "Applying.", false),
            new InstallProgressSnapshot("step_failed", "apply", "openspec", "Failed installing OpenSpec.", false, Details: "git clone failed: repository not found"),
            new InstallProgressSnapshot("install_finished", "install", null, "Installation failed.", false)));

        Assert.NotNull(viewModel.TerminalSnapshot);
        Assert.Equal("git clone failed: repository not found", viewModel.TerminalSnapshot!.Details);
    }

    [Fact]
    public async Task ConsumeAsync_TerminalFailureWithItsOwnDetails_KeepsThemAsIs()
    {
        var viewModel = new ProgressPageViewModel();

        await viewModel.ConsumeAsync(ToStream(
            new InstallProgressSnapshot("step_failed", "apply", "openspec", "Failed installing OpenSpec.", false, Details: "earlier detail"),
            new InstallProgressSnapshot("install_finished", "install", null, "Installation failed.", false, Details: "authoritative detail")));

        Assert.Equal("authoritative detail", viewModel.TerminalSnapshot!.Details);
    }

    [Fact]
    public async Task ConsumeAsync_TerminalSuccess_IsNeverEnrichedEvenIfAnEarlierStepDegraded()
    {
        var viewModel = new ProgressPageViewModel();

        await viewModel.ConsumeAsync(ToStream(
            new InstallProgressSnapshot("step_degraded", "apply", "context7", "Context7 degraded.", false, Details: "network timeout"),
            new InstallProgressSnapshot("install_finished", "install", null, "Installation finished successfully.", true)));

        Assert.Null(viewModel.TerminalSnapshot!.Details);
    }

    [Fact]
    public async Task ConsumeAsync_CapturesDegradedAndRollbackAndTheTerminalSnapshot()
    {
        var viewModel = new ProgressPageViewModel();

        await viewModel.ConsumeAsync(ToStream(
            new InstallProgressSnapshot("phase_started", "apply", null, "Applying.", false),
            new InstallProgressSnapshot("step_degraded", "apply", "context7", "Context7 degraded.", false),
            new InstallProgressSnapshot("phase_started", "rollback", null, "Rolling back.", false),
            new InstallProgressSnapshot("install_finished", "rollback", null, "Installation failed.", false)));

        Assert.True(viewModel.HadDegradedSteps);
        Assert.True(viewModel.HadRollback);
        Assert.NotNull(viewModel.TerminalSnapshot);
        Assert.Equal("install_finished", viewModel.TerminalSnapshot!.Type);
        Assert.False(viewModel.TerminalSnapshot.Success);
    }

    [Fact]
    public async Task ConsumeAsync_SpanishLanguage_OwnMessagesLocalize_EngineMessagesShownVerbatim()
    {
        var viewModel = new ProgressPageViewModel("es");

        await viewModel.ConsumeAsync(ToStream(
            // phase_started's message is engine-supplied (already localized by L1) and MUST be shown verbatim.
            new InstallProgressSnapshot("phase_started", "install", null, "Iniciando instalación.", false),
            // step_started's friendly message is OWNED by this page and must come from the es UiStrings table.
            new InstallProgressSnapshot("step_started", "apply", "openspec", "Installing OpenSpec.", false)));

        Assert.Equal("Instalando OpenSpec.", viewModel.ProgressMessage);
        Assert.Contains("Iniciando instalación.", viewModel.RecentActivity);
        Assert.Contains("Instalando OpenSpec.", viewModel.RecentActivity);
    }

    [Fact]
    public async Task ConsumeAsync_SpanishLanguage_DefaultInstallFinishedMessageLocalizes()
    {
        var viewModel = new ProgressPageViewModel("es");

        // install_finished with no engine-supplied Message falls back to this page's own default,
        // which must come from the es UiStrings table.
        await viewModel.ConsumeAsync(ToStream(
            new InstallProgressSnapshot("install_finished", "install", null, null, true)));

        Assert.Equal("Instalación finalizada con éxito.", viewModel.ProgressMessage);
    }

    private static async IAsyncEnumerable<InstallProgressSnapshot> ToStream(params InstallProgressSnapshot[] snapshots)
    {
        foreach (var snapshot in snapshots)
        {
            yield return snapshot;
            await Task.Yield();
        }
    }
}
