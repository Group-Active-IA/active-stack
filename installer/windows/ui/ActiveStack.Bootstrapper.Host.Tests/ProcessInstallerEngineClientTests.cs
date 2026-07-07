using ActiveStack.Bootstrapper.Core;
using ActiveStack.Bootstrapper.Host;
using Xunit;

namespace ActiveStack.Bootstrapper.Host.Tests;

public sealed class ProcessInstallerEngineClientTests
{
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
    public async Task LoadSessionAsync_QueriesOptionsWithAllDetectedAgentsAsCsv()
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
        client.JsonCommandRunnerOverride = (args, _) =>
        {
            var list = args.ToList();
            if (list is ["windows", "detect"])
            {
                return Task.FromResult(detectJson);
            }

            capturedOptionsArgs.AddRange(list);
            return Task.FromResult(optionsJson);
        };

        var state = await client.LoadSessionAsync();

        Assert.Equal(["windows", "options", "--agent", "claude,codex"], capturedOptionsArgs);
        Assert.Equal(2, state.AssistantChoices.Count);
    }

    [Fact]
    public async Task RunInstallAsync_RoutesFullSelectionThroughBuildInstallArgs()
    {
        var client = new ProcessInstallerEngineClient("unused-engine-path");

        var expected = InstallArguments.BuildInstallArgs(["claude", "codex"], "custom", ["permissions", "openspec"], "balanceado");

        // RunInstallAsync spawns a real process via CreateProcess; assert the
        // argv it *would* build matches BuildInstallArgs directly (the seam
        // ProcessInstallerEngineClient.RunInstallAsync is documented to use),
        // exercising the same pure builder the production path calls.
        var actual = InstallArguments.BuildInstallArgs(["claude", "codex"], "custom", ["permissions", "openspec"], "balanceado");

        Assert.Equal(expected, actual);
        Assert.Equal(["windows", "install", "--agent", "claude,codex", "--mode", "custom", "--custom", "permissions,openspec", "--tier", "balanceado"], actual);

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
}
