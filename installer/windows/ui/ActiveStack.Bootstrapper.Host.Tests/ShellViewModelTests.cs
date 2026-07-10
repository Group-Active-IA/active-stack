using ActiveStack.Bootstrapper.Core;
using ActiveStack.Bootstrapper.Host.Navigation;
using ActiveStack.Bootstrapper.Host.Pages;
using ActiveStack.Bootstrapper.Host.Pages.Backups;
using ActiveStack.Bootstrapper.Host.Pages.Install;
using ActiveStack.Bootstrapper.Host.Pages.Starters;
using ActiveStack.Bootstrapper.Host.Pages.Uninstall;
using Xunit;

namespace ActiveStack.Bootstrapper.Host.Tests;

public sealed class ShellViewModelTests
{
    [Fact]
    public void Constructor_StartsOnLanguageWithBackDisabled()
    {
        var shell = new ShellViewModel(new FakeInstallerEngineClient(BuildSession(["claude"])), persistLanguage: NoopPersist);

        Assert.IsType<LanguagePageViewModel>(shell.CurrentPage);
        Assert.False(shell.CanGoBack);
    }

    [Fact]
    public async Task AdvancingFromLanguage_MovesToHubWithBackEnabledAndNextLabel()
    {
        var shell = new ShellViewModel(new FakeInstallerEngineClient(BuildSession(["claude"])), persistLanguage: NoopPersist);

        await shell.AdvanceAsync();

        Assert.IsType<HubPageViewModel>(shell.CurrentPage);
        Assert.True(shell.CanGoBack);
        Assert.Equal("Next", shell.PrimaryLabel);
        Assert.False(shell.PrimaryEnabled);
    }

    [Fact]
    public async Task GoBack_FromHub_ReturnsToLanguage()
    {
        var shell = new ShellViewModel(new FakeInstallerEngineClient(BuildSession(["claude"])), persistLanguage: NoopPersist);
        await shell.AdvanceAsync();
        Assert.IsType<HubPageViewModel>(shell.CurrentPage);

        shell.GoBack();

        Assert.IsType<LanguagePageViewModel>(shell.CurrentPage);
        Assert.False(shell.CanGoBack);
    }

    [Fact]
    public void GoBack_AtLanguage_IsANoOp()
    {
        var shell = new ShellViewModel(new FakeInstallerEngineClient(BuildSession(["claude"])), persistLanguage: NoopPersist);

        shell.GoBack();

        Assert.IsType<LanguagePageViewModel>(shell.CurrentPage);
        Assert.False(shell.CanGoBack);
    }

    [Fact]
    public async Task AdvancingFromLanguage_SetsActiveLanguagePersistsItAndConfiguresTheClient()
    {
        var client = new FakeInstallerEngineClient(BuildSession(["claude"]));
        string? persisted = null;
        var shell = new ShellViewModel(client, persistLanguage: lang => persisted = lang);

        ((LanguagePageViewModel)shell.CurrentPage).SelectedLanguageId = "es";
        await shell.AdvanceAsync();

        Assert.Equal("es", persisted);
        Assert.Equal("es", client.Language);
    }

    [Fact]
    public async Task ChangingLanguageOnReturnToLanguagePage_InvalidatesCachedEngineData()
    {
        var client = new FakeInstallerEngineClient(BuildSession(["claude"]));
        var shell = new ShellViewModel(client, initialLanguage: "en", persistLanguage: NoopPersist);

        await shell.AdvanceAsync(); // Language(en) -> Hub, no change yet
        ((HubPageViewModel)shell.CurrentPage).SelectedEntryId = "install";
        await shell.AdvanceAsync(); // Hub -> Assistants (loads session once)
        Assert.Equal(1, client.LoadSessionCallCount);

        shell.GoBack(); // Assistants -> Hub
        shell.GoBack(); // Hub -> Language

        ((LanguagePageViewModel)shell.CurrentPage).SelectedLanguageId = "es";
        await shell.AdvanceAsync(); // Language(es, changed) -> Hub

        ((HubPageViewModel)shell.CurrentPage).SelectedEntryId = "install";
        await shell.AdvanceAsync(); // Hub -> Assistants: session was invalidated, re-fetched

        Assert.Equal(2, client.LoadSessionCallCount);
    }

    [Fact]
    public async Task ReselectingTheSameLanguage_DoesNotInvalidateCachedEngineData()
    {
        var client = new FakeInstallerEngineClient(BuildSession(["claude"]));
        var shell = new ShellViewModel(client, initialLanguage: "en", persistLanguage: NoopPersist);

        await shell.AdvanceAsync(); // Language(en) -> Hub
        ((HubPageViewModel)shell.CurrentPage).SelectedEntryId = "install";
        await shell.AdvanceAsync(); // Hub -> Assistants (loads session once)
        Assert.Equal(1, client.LoadSessionCallCount);

        shell.GoBack(); // Assistants -> Hub
        shell.GoBack(); // Hub -> Language

        ((LanguagePageViewModel)shell.CurrentPage).SelectedLanguageId = "en"; // same as before
        await shell.AdvanceAsync(); // Language(en, unchanged) -> Hub

        ((HubPageViewModel)shell.CurrentPage).SelectedEntryId = "install";
        await shell.AdvanceAsync(); // Hub -> Assistants: no re-fetch needed

        Assert.Equal(1, client.LoadSessionCallCount);
    }

    [Fact]
    public async Task AdvancingFromHub_LoadsSessionAndMovesToAssistants()
    {
        var client = new FakeInstallerEngineClient(BuildSession(["claude"]));
        var shell = new ShellViewModel(client, persistLanguage: NoopPersist);
        await shell.AdvanceAsync(); // Language -> Hub

        ((HubPageViewModel)shell.CurrentPage).SelectedEntryId = "install";
        Assert.True(shell.PrimaryEnabled);

        await shell.AdvanceAsync();

        Assert.IsType<AssistantsPageViewModel>(shell.CurrentPage);
        Assert.True(shell.CanGoBack);
        Assert.Equal(1, client.LoadSessionCallCount);
    }

    [Fact]
    public async Task FullClaudeFlow_GoesAssistantsInstallTypePermissionsReview_WithInstallPrimaryLabelOnReview()
    {
        var client = new FakeInstallerEngineClient(BuildSession(["claude"], tierCapableAgents: ["claude"]));
        var shell = new ShellViewModel(client, persistLanguage: NoopPersist);
        await shell.AdvanceAsync(); // Language -> Hub
        ((HubPageViewModel)shell.CurrentPage).SelectedEntryId = "install";
        await shell.AdvanceAsync();

        Assert.IsType<AssistantsPageViewModel>(shell.CurrentPage);
        await shell.AdvanceAsync(); // -> InstallType (full is default and preselected)

        Assert.IsType<InstallTypePageViewModel>(shell.CurrentPage);
        await shell.AdvanceAsync(); // full + claude tier-capable -> Permissions

        Assert.IsType<PermissionsPageViewModel>(shell.CurrentPage);
        await shell.AdvanceAsync(); // -> Review

        Assert.IsType<ReviewPageViewModel>(shell.CurrentPage);
        Assert.Equal("Install", shell.PrimaryLabel);
    }

    [Fact]
    public async Task GeminiOnlyFlow_SkipsPermissionsStraightToReview()
    {
        var client = new FakeInstallerEngineClient(BuildSession(["gemini"], tierCapableAgents: ["claude"]));
        var shell = new ShellViewModel(client, persistLanguage: NoopPersist);
        await shell.AdvanceAsync(); // Language -> Hub
        ((HubPageViewModel)shell.CurrentPage).SelectedEntryId = "install";
        await shell.AdvanceAsync();
        await shell.AdvanceAsync(); // -> InstallType

        await shell.AdvanceAsync(); // full + gemini (not tier-capable) -> Review directly

        Assert.IsType<ReviewPageViewModel>(shell.CurrentPage);
    }

    [Fact]
    public async Task GoBack_FromReview_SkipsPagesNotInThePathForGeminiOnlyFlow()
    {
        var client = new FakeInstallerEngineClient(BuildSession(["gemini"], tierCapableAgents: ["claude"]));
        var shell = new ShellViewModel(client, persistLanguage: NoopPersist);
        await shell.AdvanceAsync(); // Language -> Hub
        ((HubPageViewModel)shell.CurrentPage).SelectedEntryId = "install";
        await shell.AdvanceAsync();
        await shell.AdvanceAsync();
        await shell.AdvanceAsync();
        Assert.IsType<ReviewPageViewModel>(shell.CurrentPage);

        shell.GoBack();

        Assert.IsType<InstallTypePageViewModel>(shell.CurrentPage);
    }

    [Fact]
    public async Task PrimaryEnabled_ReflectsCurrentPageCanAdvance()
    {
        var client = new FakeInstallerEngineClient(BuildSession(["claude"]));
        var shell = new ShellViewModel(client, persistLanguage: NoopPersist);
        await shell.AdvanceAsync(); // Language -> Hub
        ((HubPageViewModel)shell.CurrentPage).SelectedEntryId = "install";
        await shell.AdvanceAsync();

        var assistants = (AssistantsPageViewModel)shell.CurrentPage;
        Assert.True(shell.PrimaryEnabled);

        foreach (var choice in assistants.Choices)
        {
            choice.IsSelected = false;
        }

        Assert.False(shell.PrimaryEnabled);
    }

    [Fact]
    public async Task ConfirmingReview_RunsInstallAndReachesComplete()
    {
        var client = new FakeInstallerEngineClient(BuildSession(["claude"]));
        var shell = new ShellViewModel(client, persistLanguage: NoopPersist);
        await shell.AdvanceAsync(); // Language -> Hub
        ((HubPageViewModel)shell.CurrentPage).SelectedEntryId = "install";
        await shell.AdvanceAsync(); // Assistants
        await shell.AdvanceAsync(); // InstallType (full, claude not tier-capable in this session)
        await shell.AdvanceAsync(); // -> Review (claude not tier-capable here since BuildSession default tierCapableAgents is empty)

        Assert.IsType<ReviewPageViewModel>(shell.CurrentPage);

        await shell.AdvanceAsync(); // Install -> Complete

        Assert.IsType<CompletePageViewModel>(shell.CurrentPage);
        Assert.Equal(CompleteState.Success, ((CompletePageViewModel)shell.CurrentPage).State);
        Assert.Equal(1, client.RunInstallCallCount);
        Assert.True(shell.InstallSucceeded);
    }

    [Fact]
    public async Task ConfirmingReview_ReachesComplete_WithFinishPrimaryLabel()
    {
        var client = new FakeInstallerEngineClient(BuildSession(["claude"]));
        var shell = new ShellViewModel(client, persistLanguage: NoopPersist);
        await shell.AdvanceAsync(); // Language -> Hub
        ((HubPageViewModel)shell.CurrentPage).SelectedEntryId = "install";
        await shell.AdvanceAsync(); // Assistants
        await shell.AdvanceAsync(); // InstallType
        await shell.AdvanceAsync(); // -> Review
        await shell.AdvanceAsync(); // Install -> Complete

        Assert.IsType<CompletePageViewModel>(shell.CurrentPage);
        Assert.Equal("Finish", shell.PrimaryLabel);
    }

    [Fact]
    public async Task ConfirmingReview_ReachesComplete_WithFinalizarPrimaryLabelInSpanish_AndNextLabelUnaffectedElsewhere()
    {
        var client = new FakeInstallerEngineClient(BuildSession(["claude"]));
        var shell = new ShellViewModel(client, initialLanguage: "es", persistLanguage: NoopPersist);

        await shell.AdvanceAsync(); // Language(es) -> Hub
        Assert.IsType<HubPageViewModel>(shell.CurrentPage);
        Assert.Equal("Siguiente", shell.PrimaryLabel); // non-Complete page: special-case must not leak

        ((HubPageViewModel)shell.CurrentPage).SelectedEntryId = "install";
        await shell.AdvanceAsync(); // Assistants
        await shell.AdvanceAsync(); // InstallType
        await shell.AdvanceAsync(); // -> Review
        await shell.AdvanceAsync(); // Install -> Complete

        Assert.IsType<CompletePageViewModel>(shell.CurrentPage);
        Assert.Equal("Finalizar", shell.PrimaryLabel);
    }

    [Fact]
    public async Task AdvancingFromComplete_RaisesCloseRequested_DoesNotThrow_AndDoesNotNavigateOrReenterInstall()
    {
        var client = new FakeInstallerEngineClient(BuildSession(["claude"]));
        var shell = new ShellViewModel(client, persistLanguage: NoopPersist);
        await shell.AdvanceAsync(); // Language -> Hub
        ((HubPageViewModel)shell.CurrentPage).SelectedEntryId = "install";
        await shell.AdvanceAsync(); // Assistants
        await shell.AdvanceAsync(); // InstallType
        await shell.AdvanceAsync(); // -> Review
        await shell.AdvanceAsync(); // Install -> Complete
        Assert.IsType<CompletePageViewModel>(shell.CurrentPage);
        Assert.Equal(1, client.RunInstallCallCount);

        var raised = false;
        shell.CloseRequested += (_, _) => raised = true;

        var exception = await Record.ExceptionAsync(() => shell.AdvanceAsync());

        Assert.Null(exception);
        Assert.True(raised);
        Assert.Equal(1, client.RunInstallCallCount); // advancing on Complete did not re-enter the install flow
        Assert.IsType<CompletePageViewModel>(shell.CurrentPage);
    }

    [Fact]
    public async Task OnComplete_BackStaysDisabled_AndInvokingPrimaryTwiceRaisesCloseRequestedTwiceWithoutThrowingOrNavigating()
    {
        var client = new FakeInstallerEngineClient(BuildSession(["claude"]));
        var shell = new ShellViewModel(client, persistLanguage: NoopPersist);
        await shell.AdvanceAsync(); // Language -> Hub
        ((HubPageViewModel)shell.CurrentPage).SelectedEntryId = "install";
        await shell.AdvanceAsync(); // Assistants
        await shell.AdvanceAsync(); // InstallType
        await shell.AdvanceAsync(); // -> Review
        await shell.AdvanceAsync(); // Install -> Complete
        Assert.IsType<CompletePageViewModel>(shell.CurrentPage);
        Assert.False(shell.CanGoBack);

        var raisedCount = 0;
        shell.CloseRequested += (_, _) => raisedCount++;

        var firstException = await Record.ExceptionAsync(() => shell.AdvanceAsync());
        var secondException = await Record.ExceptionAsync(() => shell.AdvanceAsync());

        Assert.Null(firstException);
        Assert.Null(secondException);
        Assert.Equal(2, raisedCount);
        Assert.False(shell.CanGoBack);
        Assert.IsType<CompletePageViewModel>(shell.CurrentPage);
    }

    [Fact]
    public async Task AdvancingFromReview_WhenStreamThrowsMidEnumeration_LandsOnCompleteErrorPageInsteadOfPropagating()
    {
        var client = new FakeInstallerEngineClient(BuildSession(["claude"]))
        {
            RunInstallFailureMessage = "Installer engine exited with code 1."
        };
        var shell = new ShellViewModel(client, persistLanguage: NoopPersist);
        await shell.AdvanceAsync(); // Language -> Hub
        ((HubPageViewModel)shell.CurrentPage).SelectedEntryId = "install";
        await shell.AdvanceAsync(); // Assistants
        await shell.AdvanceAsync(); // InstallType (full, claude not tier-capable here)
        await shell.AdvanceAsync(); // -> Review
        Assert.IsType<ReviewPageViewModel>(shell.CurrentPage);

        var exception = await Record.ExceptionAsync(() => shell.AdvanceAsync()); // Review -> stream throws mid-enumeration

        Assert.Null(exception);
        var complete = Assert.IsType<CompletePageViewModel>(shell.CurrentPage);
        Assert.Equal(CompleteState.Error, complete.State);
    }

    [Fact]
    public async Task AdvancingFromReview_WhenStreamThrows_ErrorPageMessageCarriesTheFailureText()
    {
        var client = new FakeInstallerEngineClient(BuildSession(["claude"]))
        {
            RunInstallFailureMessage = "Network unreachable while downloading harness payload."
        };
        var shell = new ShellViewModel(client, persistLanguage: NoopPersist);
        await shell.AdvanceAsync(); // Language -> Hub
        ((HubPageViewModel)shell.CurrentPage).SelectedEntryId = "install";
        await shell.AdvanceAsync(); // Assistants
        await shell.AdvanceAsync(); // InstallType
        await shell.AdvanceAsync(); // -> Review

        await shell.AdvanceAsync(); // Review -> stream throws mid-enumeration

        var complete = Assert.IsType<CompletePageViewModel>(shell.CurrentPage);
        Assert.Equal(CompleteState.Error, complete.State);
        Assert.Equal("Network unreachable while downloading harness payload.", complete.Message);
    }

    [Fact]
    public async Task AdvancingFromHub_WithUninstallSelected_SetsOperationAndMovesToUninstallAgents()
    {
        var client = new FakeInstallerEngineClient(BuildSession(["claude"]))
        {
            UninstallOptions = BuildUninstallOptions(["claude"])
        };
        var shell = new ShellViewModel(client, persistLanguage: NoopPersist);
        await shell.AdvanceAsync(); // Language -> Hub

        ((HubPageViewModel)shell.CurrentPage).SelectedEntryId = "uninstall";
        await shell.AdvanceAsync();

        Assert.IsType<UninstallAgentsPageViewModel>(shell.CurrentPage);
    }

    [Fact]
    public async Task AdvancingFromHub_WithStartersSelected_MovesToStarterCatalog()
    {
        var client = new FakeInstallerEngineClient(BuildSession(["claude"]))
        {
            Starters = [new StarterChoice("s1", "Starter One", "desc", [], ["claude"], 0)]
        };
        var shell = new ShellViewModel(client, persistLanguage: NoopPersist);
        await shell.AdvanceAsync(); // Language -> Hub

        ((HubPageViewModel)shell.CurrentPage).SelectedEntryId = "starters";
        await shell.AdvanceAsync();

        Assert.IsType<StarterCatalogPageViewModel>(shell.CurrentPage);
    }

    [Fact]
    public async Task AdvancingFromHub_WithBackupsSelected_MovesToBackupsPage()
    {
        var client = new FakeInstallerEngineClient(BuildSession(["claude"]))
        {
            Backups = []
        };
        var shell = new ShellViewModel(client, persistLanguage: NoopPersist);
        await shell.AdvanceAsync(); // Language -> Hub

        ((HubPageViewModel)shell.CurrentPage).SelectedEntryId = "backups";
        await shell.AdvanceAsync();

        Assert.IsType<BackupsPageViewModel>(shell.CurrentPage);
    }

    [Fact]
    public async Task GoBack_FromFirstUninstallPage_ReturnsToHubAndResetsOperation()
    {
        var client = new FakeInstallerEngineClient(BuildSession(["claude"]))
        {
            UninstallOptions = BuildUninstallOptions(["claude"])
        };
        var shell = new ShellViewModel(client, persistLanguage: NoopPersist);
        await shell.AdvanceAsync(); // Language -> Hub
        ((HubPageViewModel)shell.CurrentPage).SelectedEntryId = "uninstall";
        await shell.AdvanceAsync();
        Assert.IsType<UninstallAgentsPageViewModel>(shell.CurrentPage);

        shell.GoBack();

        Assert.IsType<HubPageViewModel>(shell.CurrentPage);

        ((HubPageViewModel)shell.CurrentPage).SelectedEntryId = "install";
        await shell.AdvanceAsync();

        Assert.IsType<AssistantsPageViewModel>(shell.CurrentPage);
    }

    [Fact]
    public async Task ConfirmingUninstall_RunsUninstallStreamAndReachesComplete()
    {
        var client = new FakeInstallerEngineClient(BuildSession(["claude"]))
        {
            UninstallOptions = BuildUninstallOptions(["claude"])
        };
        var shell = new ShellViewModel(client, persistLanguage: NoopPersist);
        await shell.AdvanceAsync(); // Language -> Hub
        ((HubPageViewModel)shell.CurrentPage).SelectedEntryId = "uninstall";
        await shell.AdvanceAsync(); // UninstallAgents
        await shell.AdvanceAsync(); // UninstallMode
        await shell.AdvanceAsync(); // UninstallStrategy
        await shell.AdvanceAsync(); // UninstallConfirm

        Assert.IsType<UninstallConfirmPageViewModel>(shell.CurrentPage);
        Assert.Equal("Uninstall", shell.PrimaryLabel);

        await shell.AdvanceAsync(); // runs the uninstall stream -> Complete

        Assert.IsType<CompletePageViewModel>(shell.CurrentPage);
        Assert.Equal(1, client.RunUninstallCallCount);
    }

    /// <summary>
    /// Every test that advances past the Language page triggers
    /// <c>ShellViewModel</c>'s persist-language side effect. Without an
    /// explicit override it defaults to the real
    /// <see cref="ActiveStack.Bootstrapper.Core.Localization.LanguagePreference.Save"/>
    /// against the developer's actual <c>~/.active-stack/config.json</c> —
    /// unacceptable in unit tests (and a real contention hazard under
    /// xUnit's parallel test execution). Every construction in this file
    /// passes this no-op instead.
    /// </summary>
    private static void NoopPersist(string lang)
    {
    }

    private static UninstallOptions BuildUninstallOptions(IReadOnlyList<string> detectedAgents) =>
        new(detectedAgents,
            [new InstallTypeChoice("full", "Complete", "Full removal.")],
            [new UninstallStrategyChoice("targeted", "Targeted", "Undo only what Active Stack added.", true, false)]);

    private static InstallerSessionState BuildSession(IReadOnlyList<string> detectedAgents, IReadOnlyList<string>? tierCapableAgents = null) =>
        new(
            AssistantChoices: detectedAgents.Select(id => new AssistantChoice(id, id)).ToList(),
            DefaultAssistantId: detectedAgents.FirstOrDefault(),
            InstallTypeChoices:
            [
                new InstallTypeChoice("lite", "Quick", "Fast setup to start working right away."),
                new InstallTypeChoice("full", "Complete", "Full recommended setup with all key tools."),
                new InstallTypeChoice("custom", "Custom", "Choose exactly what to install.")
            ],
            RecommendedModeId: "full",
            ForcedComponents: [new ComponentChoice("permissions", "Basic protection", "Always on.", false)],
            CustomComponents: [new ComponentChoice("openspec", "OpenSpec", "Plan and organize changes.", true)],
            TierCapable: tierCapableAgents is { Count: > 0 },
            TierCapableAgents: tierCapableAgents ?? [],
            PermissionTierChoices:
            [
                new PermissionTierChoice("balanceado", "Balanceado", "Ask for risky changes only.", true, null),
                new PermissionTierChoice("bypass", "Bypass", "Never ask.", false, "Bypass warning")
            ]);

    private sealed class FakeInstallerEngineClient : IInstallerEngineClient
    {
        private readonly InstallerSessionState _session;

        public FakeInstallerEngineClient(InstallerSessionState session)
        {
            _session = session;
        }

        public string Language { get; set; } = "en";

        public int LoadSessionCallCount { get; private set; }

        public int RunInstallCallCount { get; private set; }

        public int RunUninstallCallCount { get; private set; }

        public UninstallOptions? UninstallOptions { get; set; }

        public IReadOnlyList<StarterChoice>? Starters { get; set; }

        public IReadOnlyList<BackupEntry>? Backups { get; set; }

        /// <summary>
        /// When set, <see cref="RunInstallAsync"/> yields one in-progress
        /// snapshot and then throws mid-enumeration with this message,
        /// simulating a real engine/stream failure (D1, design.md) — drives
        /// the shell's real <c>ReviewPageViewModel</c> (a real
        /// <c>IStreamTriggerPage</c>) through its actual <c>StartStream</c>
        /// path rather than bypassing it.
        /// </summary>
        public string? RunInstallFailureMessage { get; set; }

        public Task<InstallerSessionState> LoadSessionAsync(CancellationToken cancellationToken = default)
        {
            LoadSessionCallCount++;
            return Task.FromResult(_session);
        }

        public async IAsyncEnumerable<InstallProgressSnapshot> RunInstallAsync(
            IReadOnlyList<string> agents,
            string mode,
            IReadOnlyList<string> customIds,
            string? tier,
            [System.Runtime.CompilerServices.EnumeratorCancellation] CancellationToken cancellationToken = default)
        {
            RunInstallCallCount++;
            yield return new InstallProgressSnapshot("phase_started", "prepare", null, "Preparing...", true);

            if (RunInstallFailureMessage is not null)
            {
                throw new InvalidOperationException(RunInstallFailureMessage);
            }

            yield return new InstallProgressSnapshot("install_finished", "install", null, "Installation finished successfully.", true);
            await Task.CompletedTask;
        }

        public Task<IReadOnlyList<StarterChoice>> ListStartersAsync(CancellationToken cancellationToken = default)
            => Task.FromResult(Starters ?? throw new NotSupportedException());

        public IAsyncEnumerable<InstallProgressSnapshot> RunStarterInstallAsync(string starterId, string projectPath, IReadOnlyList<string> agents, bool dryRun, CancellationToken cancellationToken = default)
            => throw new NotSupportedException();

        public Task<IReadOnlyList<BackupEntry>> ListBackupsAsync(CancellationToken cancellationToken = default)
            => Task.FromResult(Backups ?? throw new NotSupportedException());

        public Task<BackupActionResult> RunBackupActionAsync(string action, string id, string? description, CancellationToken cancellationToken = default)
            => throw new NotSupportedException();

        public Task<UninstallOptions> LoadUninstallOptionsAsync(CancellationToken cancellationToken = default)
            => Task.FromResult(UninstallOptions ?? throw new NotSupportedException());

        public async IAsyncEnumerable<InstallProgressSnapshot> RunUninstallAsync(
            IReadOnlyList<string> agents,
            string mode,
            string strategy,
            string? backupId,
            [System.Runtime.CompilerServices.EnumeratorCancellation] CancellationToken cancellationToken = default)
        {
            RunUninstallCallCount++;
            yield return new InstallProgressSnapshot("uninstall_finished", "uninstall", null, "Uninstall finished successfully.", true);
            await Task.CompletedTask;
        }
    }
}
