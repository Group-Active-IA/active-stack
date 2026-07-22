using System.Text;
using ActiveStack.Bootstrapper.Core;
using ActiveStack.Bootstrapper.Host;
using Xunit;

namespace ActiveStack.Bootstrapper.Host.Tests;

public sealed class ProcessInstallerEngineClientTests
{
    [Fact]
    public void CreateProcess_ConfiguresUtf8StandardOutputAndErrorEncoding()
    {
        // D2, design.md (Bug B): without explicit UTF-8 encoding, accented
        // text (e.g. Spanish copy) read from the engine's stdout/stderr gets
        // corrupted. Drives the real ProcessStartInfo the client builds
        // rather than a JsonCommandRunnerOverride-fed string.
        var client = new ProcessInstallerEngineClient("unused-engine-path");

        using var process = client.CreateProcess(["windows", "detect"]);

        Assert.Equal(Encoding.UTF8, process.StartInfo.StandardOutputEncoding);
        Assert.Equal(Encoding.UTF8, process.StartInfo.StandardErrorEncoding);
    }

    [Fact]
    public void ResolveDetectedAgentsCsv_JoinsAllDetectedAgents()
    {
        const string detectJson = """{"detected_agents":["claude","codex"]}""";

        var csv = ProcessInstallerEngineClient.ResolveDetectedAgentsCsv(detectJson);

        Assert.Equal("claude,codex", csv);
    }

    [Fact]
    public void ResolveDetectedAgentsCsv_FallsBackToClaudeWhenEmpty()
    {
        const string detectJson = """{"detected_agents":[]}""";

        var csv = ProcessInstallerEngineClient.ResolveDetectedAgentsCsv(detectJson);

        Assert.Equal("claude", csv);
    }

    [Fact]
    public void Language_DefaultsToEnglish()
    {
        var client = new ProcessInstallerEngineClient("unused-engine-path");

        Assert.Equal("en", client.Language);
    }

    [Fact]
    public async Task LoadSessionAsync_QueriesOptionsWithAllDetectedAgentsAsCsv_AndDetectCarriesNoLang()
    {
        const string detectJson = """{"detected_agents":["claude","codex"]}""";
        const string optionsJson = """
        {
          "modes": [{ "id": "full", "label": "Complete", "description": "Full recommended setup with all key tools." }],
          "forced_components": [],
          "custom_components": []
        }
        """;

        var client = new ProcessInstallerEngineClient("unused-engine-path");
        var capturedOptionsArgs = new List<string>();
        var capturedDetectArgs = new List<string>();
        client.JsonCommandRunnerOverride = (args, _) =>
        {
            var list = args.ToList();
            if (list is ["windows", "detect", ..])
            {
                capturedDetectArgs.AddRange(list);
                return Task.FromResult(detectJson);
            }

            capturedOptionsArgs.AddRange(list);
            return Task.FromResult(optionsJson);
        };

        var state = await client.LoadSessionAsync();

        Assert.Equal(["windows", "detect"], capturedDetectArgs);
        Assert.Equal(["windows", "options", "--agent", "claude,codex", "--lang", "en"], capturedOptionsArgs);
        Assert.Equal(2, state.AssistantChoices.Count);
    }

    [Fact]
    public async Task LoadSessionAsync_WithSpanishLanguage_ThreadsLangEsIntoOptionsButNotDetect()
    {
        const string detectJson = """{"detected_agents":["claude"]}""";
        const string optionsJson = """{"modes":[],"forced_components":[],"custom_components":[]}""";

        var client = new ProcessInstallerEngineClient("unused-engine-path") { Language = "es" };
        var capturedOptionsArgs = new List<string>();
        var capturedDetectArgs = new List<string>();
        client.JsonCommandRunnerOverride = (args, _) =>
        {
            var list = args.ToList();
            if (list is ["windows", "detect", ..])
            {
                capturedDetectArgs.AddRange(list);
                return Task.FromResult(detectJson);
            }

            capturedOptionsArgs.AddRange(list);
            return Task.FromResult(optionsJson);
        };

        await client.LoadSessionAsync();

        Assert.DoesNotContain("--lang", capturedDetectArgs);
        Assert.Equal(["windows", "options", "--agent", "claude", "--lang", "es"], capturedOptionsArgs);
    }

    [Fact]
    public async Task ListBackupsAsync_WithSpanishLanguage_AppendsLangEs()
    {
        const string json = """{"backups":[]}""";

        var capturedArgs = new List<string>();
        var client = new ProcessInstallerEngineClient("unused-engine-path") { Language = "es" };
        client.JsonCommandRunnerOverride = (args, _) =>
        {
            capturedArgs.AddRange(args);
            return Task.FromResult(json);
        };

        await client.ListBackupsAsync();

        Assert.Equal(["windows", "backups", "list", "--lang", "es"], capturedArgs);
    }

    [Fact]
    public async Task ListStartersAsync_AppendsTheClientLanguage()
    {
        const string json = """{"starters":[]}""";

        var capturedArgs = new List<string>();
        var client = new ProcessInstallerEngineClient("unused-engine-path");
        client.JsonCommandRunnerOverride = (args, _) =>
        {
            capturedArgs.AddRange(args);
            return Task.FromResult(json);
        };

        await client.ListStartersAsync();

        Assert.Equal(["windows", "starters", "list", "--lang", "en"], capturedArgs);
    }

    [Fact]
    public async Task LoadUninstallOptionsAsync_AppendsTheClientLanguage()
    {
        const string json = """{"detected_agents":[],"modes":[],"strategies":[]}""";

        var capturedArgs = new List<string>();
        var client = new ProcessInstallerEngineClient("unused-engine-path") { Language = "es" };
        client.JsonCommandRunnerOverride = (args, _) =>
        {
            capturedArgs.AddRange(args);
            return Task.FromResult(json);
        };

        await client.LoadUninstallOptionsAsync();

        Assert.Equal(["windows", "uninstall-options", "--lang", "es"], capturedArgs);
    }

    [Fact]
    public void StreamInvocations_BuildArgsWithTheClientLanguage()
    {
        // The streaming subcommands spawn a real process, so their argv
        // construction is exposed through internal Build*Args seams that the
        // production Run*Async methods call with the client's Language.
        var client = new ProcessInstallerEngineClient("unused-engine-path") { Language = "es" };

        var installArgs = client.BuildRunInstallArgs(["claude"], "full", [], null);
        Assert.Equal(["windows", "install", "--agent", "claude", "--mode", "full", "--lang", "es"], installArgs);

        var starterArgs = client.BuildRunStarterInstallArgs("alpha", "C:/app", ["claude"], dryRun: false);
        Assert.Equal(["windows", "starters", "install", "--starter", "alpha", "--project", "C:/app", "--agent", "claude", "--yes", "--lang", "es"], starterArgs);

        var uninstallArgs = client.BuildRunUninstallArgs(["claude"], "full", "targeted", null);
        Assert.Equal(["windows", "uninstall", "--agent", "claude", "--mode", "full", "--strategy", "targeted", "--lang", "es"], uninstallArgs);
    }

    [Fact]
    public void RunInstallAsync_RoutesFullSelectionThroughBuildInstallArgs()
    {
        var client = new ProcessInstallerEngineClient("unused-engine-path");

        var actual = client.BuildRunInstallArgs(["claude", "codex"], "custom", ["permissions", "openspec"], "balanceado");

        Assert.Equal(["windows", "install", "--agent", "claude,codex", "--mode", "custom", "--custom", "permissions,openspec", "--tier", "balanceado", "--lang", "en"], actual);

        // RunInstallAsync itself must accept the full selection shape.
        var stream = client.RunInstallAsync(["claude"], "full", [], null);
        Assert.NotNull(stream);
    }

    [Fact]
    public async Task ListBackupsAsync_ReturnsCoreParsedResultForFedJson()
    {
        const string json = """
        {
          "backups": [
            { "id": "backup-1", "created_at": "2026-03-15T10:30:00Z", "source": "install", "description": "before upgrade", "file_count": 3, "pinned": true, "compressed": false, "display_label": "backup-1 (install)", "manifest_path": "C:\\backups\\backup-1\\manifest.json" }
          ]
        }
        """;

        var client = new ProcessInstallerEngineClient("unused-engine-path")
        {
            JsonCommandRunnerOverride = (_, _) => Task.FromResult(json)
        };

        var backups = await client.ListBackupsAsync();

        Assert.Equal(BackupListParser.BuildFromJson(json), backups);
    }

    [Fact]
    public async Task ListStartersAsync_ReturnsCoreParsedResultForFedJson()
    {
        const string json = """
        {
          "starters": [
            { "id": "alpha", "name": "Alpha", "description": "First starter", "includes": ["beta"], "harnesses": ["harness-a", "harness-b"], "mcp_count": 2 }
          ]
        }
        """;

        var client = new ProcessInstallerEngineClient("unused-engine-path")
        {
            JsonCommandRunnerOverride = (_, _) => Task.FromResult(json)
        };

        var starters = await client.ListStartersAsync();

        var expected = StarterCatalogParser.BuildFromJson(json);
        Assert.Single(starters);
        Assert.Equal(expected[0].Id, starters[0].Id);
        Assert.Equal(expected[0].Name, starters[0].Name);
        Assert.Equal(expected[0].McpCount, starters[0].McpCount);
        Assert.Equal(expected[0].Includes, starters[0].Includes);
        Assert.Equal(expected[0].Harnesses, starters[0].Harnesses);
    }

    [Fact]
    public async Task LoadUninstallOptionsAsync_ReturnsCoreParsedResultForFedJson()
    {
        const string json = """
        {
          "detected_agents": ["claude"],
          "modes": [{ "id": "full", "label": "Complete", "description": "Full recommended setup with all key tools." }],
          "strategies": [
            { "id": "targeted", "label": "Targeted", "description": "Reverse each installed harness individually.", "default": true, "requires_manifest": false },
            { "id": "restore", "label": "Restore from backup", "description": "Restore the full pre-install state from a backup manifest.", "default": false, "requires_manifest": true }
          ]
        }
        """;

        var client = new ProcessInstallerEngineClient("unused-engine-path")
        {
            JsonCommandRunnerOverride = (_, _) => Task.FromResult(json)
        };

        var options = await client.LoadUninstallOptionsAsync();

        var expected = UninstallOptionsParser.BuildFromJson(json);
        Assert.Equal(expected.DetectedAgents, options.DetectedAgents);
        Assert.Equal(expected.Modes, options.Modes);
        Assert.Equal(expected.Strategies, options.Strategies);
    }

    [Fact]
    public async Task RunBackupActionAsync_ReturnsCoreParsedResultForFedJson()
    {
        const string json = """{"success":true,"message":"restored","id":"backup-1"}""";

        var client = new ProcessInstallerEngineClient("unused-engine-path")
        {
            JsonCommandRunnerOverride = (_, _) => Task.FromResult(json)
        };

        var result = await client.RunBackupActionAsync("restore", "backup-1", null);

        Assert.Equal(BackupActionResultParser.BuildFromJson(json), result);
    }

    [Theory]
    [InlineData("install_finished", true)]
    [InlineData("starter_finished", true)]
    [InlineData("uninstall_finished", true)]
    [InlineData("INSTALL_FINISHED", true)]
    [InlineData("step_failed", false)]
    [InlineData("step_output", false)]
    [InlineData(null, false)]
    public void IsTerminalEventType_RecognizesOnlyTheThreeStreamTerminalTypes(string? type, bool expected)
    {
        // A non-zero engine exit AFTER one of these three event types is an
        // already-reported failure (the terminal snapshot carries its own
        // Success=false + Message) — RunStreamingCommandAsync must not
        // override it with the generic "exited with code N" fallback.
        Assert.Equal(expected, ProcessInstallerEngineClient.IsTerminalEventType(type));
    }
}
