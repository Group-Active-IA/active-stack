using System.Linq;
using ActiveStack.Bootstrapper.Core;
using Xunit;

namespace ActiveStack.Bootstrapper.Core.Tests;

public sealed class UninstallOptionsParserTests
{
    [Fact]
    public void BuildFromJson_ParsesDetectedAgentsModesAndStrategies()
    {
        const string json = """
        {
          "detected_agents": ["claude"],
          "modes": [
            { "id": "lite", "label": "Quick", "description": "Fast setup to start working right away." },
            { "id": "full", "label": "Complete", "description": "Full recommended setup with all key tools." },
            { "id": "custom", "label": "Custom", "description": "Choose exactly what to install." }
          ],
          "strategies": [
            { "id": "targeted", "label": "Targeted", "description": "Reverse each installed harness individually.", "default": true, "requires_manifest": false },
            { "id": "restore", "label": "Restore from backup", "description": "Restore the full pre-install state from a backup manifest.", "default": false, "requires_manifest": true }
          ]
        }
        """;

        var options = UninstallOptionsParser.BuildFromJson(json);

        Assert.Equal(["claude"], options.DetectedAgents);
        Assert.Equal(3, options.Modes.Count);

        var targeted = options.Strategies.Single(s => s.Id == "targeted");
        Assert.True(targeted.IsDefault);
        Assert.False(targeted.RequiresManifest);

        var restore = options.Strategies.Single(s => s.Id == "restore");
        Assert.False(restore.IsDefault);
        Assert.True(restore.RequiresManifest);
    }

    [Fact]
    public void BuildFromJson_MissingCollections_DefaultToEmptyNotNull()
    {
        const string json = """{"detected_agents":[],"modes":[],"strategies":[]}""";

        var options = UninstallOptionsParser.BuildFromJson(json);

        Assert.NotNull(options.DetectedAgents);
        Assert.Empty(options.DetectedAgents);
        Assert.NotNull(options.Modes);
        Assert.Empty(options.Modes);
        Assert.NotNull(options.Strategies);
        Assert.Empty(options.Strategies);
    }
}
