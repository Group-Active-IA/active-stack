using ActiveStack.Bootstrapper.Core;
using ActiveStack.Bootstrapper.Host;
using Xunit;

namespace ActiveStack.Bootstrapper.Host.Tests;

public sealed class MainWindowViewModelTests
{
    [Fact]
    public async Task LoadAsync_PopulatesAssistantsAndModesFromEngineClient()
    {
        var session = new InstallerSessionState(
            AssistantChoices:
            [
                new AssistantChoice("claude", "Claude"),
                new AssistantChoice("codex", "Codex")
            ],
            DefaultAssistantId: "claude",
            InstallTypeChoices:
            [
                new InstallTypeChoice("lite", "Quick", "Fast setup to start working right away."),
                new InstallTypeChoice("full", "Complete", "Full recommended setup with all key tools.")
            ],
            RecommendedModeId: "full",
            ForcedComponents:
            [
                new ComponentChoice("permissions", "Basic protection", "Helps avoid unsafe changes.", false)
            ],
            CustomComponents:
            [
                new ComponentChoice("openspec", "OpenSpec", "Plan and organize changes.", true)
            ]);

        var viewModel = new MainWindowViewModel(new FakeInstallerEngineClient(session));

        await viewModel.LoadAsync();

        Assert.False(viewModel.IsLoading);
        Assert.Equal("Install Active Stack", viewModel.Title);
        Assert.Equal(2, viewModel.AssistantChoices.Count);
        Assert.Equal("claude", viewModel.SelectedAssistantId);
        Assert.Equal("full", viewModel.SelectedInstallTypeId);
        Assert.Single(viewModel.ForcedComponents);
        Assert.Single(viewModel.CustomComponents);
        Assert.Contains("AI coding workspace", viewModel.Subtitle);
    }

    [Fact]
    public async Task LoadAsync_OnFailureSetsFriendlyErrorMessage()
    {
        var viewModel = new MainWindowViewModel(new FailingInstallerEngineClient("detect failed"));

        await viewModel.LoadAsync();

        Assert.False(viewModel.IsLoading);
        Assert.Equal("We couldn't prepare your setup yet.", viewModel.ErrorTitle);
        Assert.Contains("detect failed", viewModel.ErrorDetails);
    }

    [Fact]
    public async Task StartInstallAsync_TracksProgressAndMarksSuccess()
    {
        var session = BuildSession();
        var client = new FakeInstallerEngineClient(session)
        {
            InstallSnapshots =
            [
                new InstallProgressSnapshot("phase_started", "install", null, "Starting installation.", false),
                new InstallProgressSnapshot("step_started", "apply", "openspec", "Installing OpenSpec.", false),
                new InstallProgressSnapshot("install_finished", "install", null, "Installation finished successfully.", true)
            ]
        };
        var viewModel = new MainWindowViewModel(client);
        await viewModel.LoadAsync();

        await viewModel.StartInstallAsync();

        Assert.False(viewModel.IsInstalling);
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
    public async Task StartInstallAsync_MapsGenericEngineEventsToFriendlyStatus()
    {
        var session = BuildSession();
        var client = new FakeInstallerEngineClient(session)
        {
            InstallSnapshots =
            [
                new InstallProgressSnapshot("phase_started", "prepare", null, "Preparing installation.", false),
                new InstallProgressSnapshot("step_started", "apply", "external:engram", "Step started.", false),
                new InstallProgressSnapshot("download_started", "apply", "external:engram", "Downloading engram.", false),
                new InstallProgressSnapshot("step_succeeded", "apply", "external:engram", "Step completed.", false),
                new InstallProgressSnapshot("install_finished", "install", null, "Installation finished successfully.", true)
            ]
        };

        var viewModel = new MainWindowViewModel(client);
        await viewModel.LoadAsync();

        await viewModel.StartInstallAsync();

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
    public async Task StartInstallAsync_OnFailureSetsFriendlyInstallError()
    {
        var session = BuildSession();
        var client = new FakeInstallerEngineClient(session)
        {
            InstallException = new InvalidOperationException("install failed")
        };
        var viewModel = new MainWindowViewModel(client);
        await viewModel.LoadAsync();

        await viewModel.StartInstallAsync();

        Assert.False(viewModel.IsInstalling);
        Assert.False(viewModel.InstallSucceeded);
        Assert.Equal("We couldn't finish the installation.", viewModel.ErrorTitle);
        Assert.Contains("install failed", viewModel.ErrorDetails);
    }

    [Fact]
    public async Task StartInstallAsync_WithExternalRunnerLeavesInstallOpenUntilBurnCompletes()
    {
        var session = BuildSession();
        var client = new FakeInstallerEngineClient(session);
        var viewModel = new MainWindowViewModel(client);
        var runnerCalled = false;

        viewModel.ConfigureExternalInstallRunner((assistantId, installTypeId) =>
        {
            runnerCalled = assistantId == "claude" && installTypeId == "full";
            return Task.CompletedTask;
        });

        await viewModel.LoadAsync();
        await viewModel.StartInstallAsync();

        Assert.True(runnerCalled);
        Assert.True(viewModel.IsInstalling);
        Assert.False(viewModel.InstallSucceeded);
        Assert.False(viewModel.IsInstallActionEnabled);
        Assert.Equal("Installing...", viewModel.InstallButtonText);
        Assert.Equal("Starting installation.", viewModel.ProgressMessage);

        viewModel.CompleteExternalInstall(true, "Installation finished successfully.", "ActiveStackBootstrap");

        Assert.False(viewModel.IsInstalling);
        Assert.True(viewModel.InstallSucceeded);
        Assert.True(viewModel.IsInstallActionEnabled);
        Assert.Equal("Install now", viewModel.InstallButtonText);
        Assert.Equal("Installation finished successfully.", viewModel.ProgressMessage);
        Assert.Equal("ActiveStackBootstrap", viewModel.CurrentStepId);
    }

    [Fact]
    public void ReportExternalProgress_DuringDetectDoesNotEnterInstallingState()
    {
        var viewModel = new MainWindowViewModel(new FakeInstallerEngineClient(BuildSession()));

        viewModel.ReportExternalProgress("Setup is ready to install.", "detect");

        Assert.False(viewModel.IsInstalling);
        Assert.True(viewModel.IsInstallActionEnabled is false); // no selections loaded yet
        Assert.Equal("Install now", viewModel.InstallButtonText);
        Assert.Equal("Setup is ready to install.", viewModel.ProgressMessage);
        Assert.Equal("detect", viewModel.CurrentStepId);
    }

    [Fact]
    public async Task StartInstallAsync_WithoutSelectionsShowsValidationAndKeepsActionEnabled()
    {
        var session = BuildSession();
        var client = new FakeInstallerEngineClient(session);
        var viewModel = new MainWindowViewModel(client);
        await viewModel.LoadAsync();

        viewModel.SelectedAssistantId = null;
        await viewModel.StartInstallAsync();

        Assert.False(viewModel.IsInstalling);
        Assert.False(viewModel.InstallSucceeded);
        Assert.False(viewModel.IsInstallActionEnabled);
        Assert.Equal("Install now", viewModel.InstallButtonText);
        Assert.Equal("We couldn't finish the installation.", viewModel.ErrorTitle);
    }

    private static InstallerSessionState BuildSession() =>
        new(
            AssistantChoices:
            [
                new AssistantChoice("claude", "Claude"),
                new AssistantChoice("codex", "Codex")
            ],
            DefaultAssistantId: "claude",
            InstallTypeChoices:
            [
                new InstallTypeChoice("lite", "Quick", "Fast setup to start working right away."),
                new InstallTypeChoice("full", "Complete", "Full recommended setup with all key tools.")
            ],
            RecommendedModeId: "full",
            ForcedComponents:
            [
                new ComponentChoice("permissions", "Basic protection", "Helps avoid unsafe changes.", false)
            ],
            CustomComponents:
            [
                new ComponentChoice("openspec", "OpenSpec", "Plan and organize changes.", true)
            ]);

    private sealed class FakeInstallerEngineClient : IInstallerEngineClient
    {
        private readonly InstallerSessionState _session;
        public IReadOnlyList<InstallProgressSnapshot> InstallSnapshots { get; init; } = [];
        public Exception? InstallException { get; init; }

        public FakeInstallerEngineClient(InstallerSessionState session)
        {
            _session = session;
        }

        public Task<InstallerSessionState> LoadSessionAsync(CancellationToken cancellationToken = default)
            => Task.FromResult(_session);

        public async IAsyncEnumerable<InstallProgressSnapshot> RunInstallAsync(
            string assistantId,
            string installTypeId,
            [System.Runtime.CompilerServices.EnumeratorCancellation] CancellationToken cancellationToken = default)
        {
            if (InstallException is not null)
            {
                throw InstallException;
            }

            foreach (var snapshot in InstallSnapshots)
            {
                yield return snapshot;
                await Task.Yield();
            }
        }
    }

    private sealed class FailingInstallerEngineClient : IInstallerEngineClient
    {
        private readonly string _message;

        public FailingInstallerEngineClient(string message)
        {
            _message = message;
        }

        public Task<InstallerSessionState> LoadSessionAsync(CancellationToken cancellationToken = default)
            => Task.FromException<InstallerSessionState>(new InvalidOperationException(_message));

        public async IAsyncEnumerable<InstallProgressSnapshot> RunInstallAsync(
            string assistantId,
            string installTypeId,
            [System.Runtime.CompilerServices.EnumeratorCancellation] CancellationToken cancellationToken = default)
        {
            await Task.Yield();
            throw new InvalidOperationException(_message);
#pragma warning disable CS0162
            yield break;
#pragma warning restore CS0162
        }
    }
}
