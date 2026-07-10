using System.Linq;
using ActiveStack.Bootstrapper.Core;
using Xunit;

namespace ActiveStack.Bootstrapper.Core.Tests;

public sealed class InstallerSessionStateBuilderTests
{
    [Fact]
    public void BuildFromJson_MapsDetectedAgentsIntoAssistantChoices()
    {
        const string detectJson = """
        {
          "detected_agents": ["claude", "codex"]
        }
        """;

        const string optionsJson = """
        {
          "modes": [
            { "id": "lite", "label": "Quick", "description": "Fast setup to start working right away." },
            { "id": "full", "label": "Complete", "description": "Full recommended setup with all key tools." },
            { "id": "custom", "label": "Custom", "description": "Choose exactly what to install." }
          ],
          "forced_components": [],
          "custom_components": []
        }
        """;

        var state = InstallerSessionStateBuilder.BuildFromJson(detectJson, optionsJson);

        Assert.Collection(
            state.AssistantChoices,
            first => Assert.Equal("claude", first.Id),
            second => Assert.Equal("codex", second.Id));
        Assert.Equal("claude", state.DefaultAssistantId);
    }

    [Fact]
    public void BuildFromJson_MapsModesAndForcedComponentsIntoUiSummary()
    {
        const string detectJson = """
        {
          "detected_agents": ["claude"]
        }
        """;

        const string optionsJson = """
        {
          "modes": [
            { "id": "lite", "label": "Quick", "description": "Fast setup to start working right away." },
            { "id": "full", "label": "Complete", "description": "Full recommended setup with all key tools." },
            { "id": "custom", "label": "Custom", "description": "Choose exactly what to install." }
          ],
          "forced_components": [
            { "id": "permissions", "label": "Basic protection", "description": "Helps avoid unsafe changes." }
          ],
          "custom_components": [
            { "id": "openspec", "label": "OpenSpec", "description": "Plan and organize changes.", "recommended": true }
          ]
        }
        """;

        var state = InstallerSessionStateBuilder.BuildFromJson(detectJson, optionsJson);

        Assert.Equal("full", state.RecommendedModeId);
        Assert.Collection(
            state.InstallTypeChoices,
            first => Assert.Equal("Quick", first.Label),
            second => Assert.Equal("Complete", second.Label),
            third => Assert.Equal("Custom", third.Label));
        Assert.Single(state.ForcedComponents);
        Assert.Equal("Basic protection", state.ForcedComponents[0].Label);
        Assert.Single(state.CustomComponents);
        Assert.True(state.CustomComponents[0].Recommended);
    }

    [Fact]
    public void BuildFromJson_RemovesForcedComponentsFromOptionalSummary()
    {
        const string detectJson = """
        {
          "detected_agents": ["claude"]
        }
        """;

        const string optionsJson = """
        {
          "modes": [
            { "id": "full", "label": "Complete", "description": "Full recommended setup with all key tools." }
          ],
          "forced_components": [
            { "id": "permissions", "label": "Basic protection", "description": "Helps avoid unsafe changes." }
          ],
          "custom_components": [
            { "id": "permissions", "label": "Basic protection", "description": "Helps avoid unsafe changes." },
            { "id": "openspec", "label": "OpenSpec", "description": "Plan and organize changes.", "recommended": true }
          ]
        }
        """;

        var state = InstallerSessionStateBuilder.BuildFromJson(detectJson, optionsJson);

        Assert.Single(state.ForcedComponents);
        Assert.Single(state.CustomComponents);
        Assert.Equal("openspec", state.CustomComponents[0].Id);
    }

    [Fact]
    public void BuildFromJson_ParsesTierMetadataWhenPresent()
    {
        const string detectJson = """
        {
          "detected_agents": ["claude"]
        }
        """;

        const string optionsJson = """
        {
          "modes": [
            { "id": "lite", "label": "Quick", "description": "Fast setup to start working right away." },
            { "id": "full", "label": "Complete", "description": "Full recommended setup with all key tools." },
            { "id": "custom", "label": "Custom", "description": "Choose exactly what to install." }
          ],
          "forced_components": [],
          "custom_components": [],
          "tier_capable": true,
          "tier_capable_agents": ["claude"],
          "permission_tiers": [
            { "id": "estricto", "label": "Estricto", "description": "Agent must ask for every operation. Highest friction, highest security.", "default": false },
            { "id": "balanceado", "label": "Balanceado", "description": "Curated allow-list for safe, repetitive operations. Recommended starting point.", "default": true },
            { "id": "bypass", "label": "Bypass", "description": "Full autonomy opt-in. The security floor deny-list still applies.", "default": false, "warning": "Bypass: autonomous mode — the security floor still applies (C-21)" }
          ]
        }
        """;

        var state = InstallerSessionStateBuilder.BuildFromJson(detectJson, optionsJson);

        Assert.True(state.TierCapable);
        Assert.Equal(["claude"], state.TierCapableAgents);
        Assert.Equal(3, state.PermissionTierChoices.Count);

        var balanceado = state.PermissionTierChoices.Single(t => t.Id == "balanceado");
        Assert.True(balanceado.IsDefault);
        Assert.Null(balanceado.Warning);

        var estricto = state.PermissionTierChoices.Single(t => t.Id == "estricto");
        Assert.False(estricto.IsDefault);

        var bypass = state.PermissionTierChoices.Single(t => t.Id == "bypass");
        Assert.False(bypass.IsDefault);
        Assert.Equal("Bypass: autonomous mode — the security floor still applies (C-21)", bypass.Warning);
    }

    [Fact]
    public void BuildFromJson_MapsLongDescriptionIntoInstallTypeComponentAndPermissionTierChoices()
    {
        const string detectJson = """
        {
          "detected_agents": ["claude"]
        }
        """;

        const string optionsJson = """
        {
          "modes": [
            { "id": "full", "label": "Complete", "description": "Full recommended setup with all key tools.", "long_description": "The full install sets up every recommended harness end to end." }
          ],
          "forced_components": [
            { "id": "permissions", "label": "Basic protection", "description": "Helps avoid unsafe changes.", "long_description": "Permissions enforces a security floor deny-list on every agent." }
          ],
          "custom_components": [
            { "id": "openspec", "label": "OpenSpec", "description": "Plan and organize changes.", "recommended": true, "long_description": "OpenSpec tracks proposals, specs, and tasks per change." }
          ],
          "permission_tiers": [
            { "id": "balanceado", "label": "Balanceado", "description": "Curated allow-list.", "default": true, "long_description": "Balanceado allows safe, repetitive operations without prompting." }
          ]
        }
        """;

        var state = InstallerSessionStateBuilder.BuildFromJson(detectJson, optionsJson);

        Assert.Equal("The full install sets up every recommended harness end to end.", state.InstallTypeChoices[0].LongDescription);
        Assert.Equal("Permissions enforces a security floor deny-list on every agent.", state.ForcedComponents[0].LongDescription);
        Assert.Equal("OpenSpec tracks proposals, specs, and tasks per change.", state.CustomComponents[0].LongDescription);
        Assert.Equal("Balanceado allows safe, repetitive operations without prompting.", state.PermissionTierChoices[0].LongDescription);
    }

    [Fact]
    public void BuildFromJson_AbsentLongDescriptionDefaultsToEmptyString()
    {
        const string detectJson = """
        {
          "detected_agents": ["claude"]
        }
        """;

        const string optionsJson = """
        {
          "modes": [
            { "id": "full", "label": "Complete", "description": "Full recommended setup with all key tools." }
          ],
          "forced_components": [
            { "id": "permissions", "label": "Basic protection", "description": "Helps avoid unsafe changes." }
          ],
          "custom_components": [
            { "id": "openspec", "label": "OpenSpec", "description": "Plan and organize changes.", "recommended": true }
          ],
          "permission_tiers": [
            { "id": "balanceado", "label": "Balanceado", "description": "Curated allow-list.", "default": true }
          ]
        }
        """;

        var state = InstallerSessionStateBuilder.BuildFromJson(detectJson, optionsJson);

        Assert.Equal(string.Empty, state.InstallTypeChoices[0].LongDescription);
        Assert.Equal(string.Empty, state.ForcedComponents[0].LongDescription);
        Assert.Equal(string.Empty, state.CustomComponents[0].LongDescription);
        Assert.Equal(string.Empty, state.PermissionTierChoices[0].LongDescription);
    }

    [Fact]
    public void BuildFromJson_AbsentTierFieldsDefaultSafely()
    {
        const string detectJson = """
        {
          "detected_agents": ["claude"]
        }
        """;

        const string optionsJson = """
        {
          "modes": [
            { "id": "lite", "label": "Quick", "description": "Fast setup to start working right away." },
            { "id": "full", "label": "Complete", "description": "Full recommended setup with all key tools." }
          ],
          "forced_components": [
            { "id": "permissions", "label": "Basic protection", "description": "Helps avoid unsafe changes." }
          ],
          "custom_components": []
        }
        """;

        var state = InstallerSessionStateBuilder.BuildFromJson(detectJson, optionsJson);

        Assert.False(state.TierCapable);
        Assert.Empty(state.TierCapableAgents);
        Assert.Empty(state.PermissionTierChoices);
        // Existing fields unaffected by the absent tier metadata.
        Assert.Equal("full", state.RecommendedModeId);
        Assert.Single(state.ForcedComponents);
    }
}
