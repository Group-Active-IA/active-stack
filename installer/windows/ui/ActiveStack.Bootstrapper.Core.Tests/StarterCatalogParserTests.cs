using ActiveStack.Bootstrapper.Core;
using Xunit;

namespace ActiveStack.Bootstrapper.Core.Tests;

public sealed class StarterCatalogParserTests
{
    [Fact]
    public void BuildFromJson_ParsesStartersInCatalogOrder()
    {
        const string json = """
        {
          "starters": [
            { "id": "alpha", "name": "Alpha", "description": "First starter", "includes": ["beta"], "harnesses": ["harness-a", "harness-b"], "mcp_count": 2 },
            { "id": "beta", "name": "Beta", "description": "Second starter", "includes": [], "harnesses": ["harness-b"], "mcp_count": 1 }
          ]
        }
        """;

        var starters = StarterCatalogParser.BuildFromJson(json);

        Assert.Equal(2, starters.Count);
        Assert.Equal("alpha", starters[0].Id);
        Assert.Equal("beta", starters[1].Id);
        Assert.Equal("Alpha", starters[0].Name);
        Assert.Equal("First starter", starters[0].Description);
        Assert.Equal(["beta"], starters[0].Includes);
        Assert.Equal(2, starters[0].Harnesses.Count);
        Assert.Equal(2, starters[0].McpCount);
        Assert.Equal(1, starters[1].McpCount);
    }

    [Fact]
    public void BuildFromJson_EmptyCatalog_ReturnsEmptyListNotNull()
    {
        const string json = """{"starters":[]}""";

        var starters = StarterCatalogParser.BuildFromJson(json);

        Assert.NotNull(starters);
        Assert.Empty(starters);
    }

    [Fact]
    public void BuildFromJson_MapsLongDescriptionIntoStarterChoice()
    {
        const string json = """
        {
          "starters": [
            { "id": "alpha", "name": "Alpha", "description": "First starter", "includes": ["beta"], "harnesses": ["harness-a"], "mcp_count": 2, "long_description": "Alpha scaffolds a full-stack app with auth and billing wired in." }
          ]
        }
        """;

        var starters = StarterCatalogParser.BuildFromJson(json);

        Assert.Equal("Alpha scaffolds a full-stack app with auth and billing wired in.", starters[0].LongDescription);
    }

    [Fact]
    public void BuildFromJson_AbsentLongDescription_DefaultsToEmptyString()
    {
        const string json = """
        {
          "starters": [
            { "id": "beta", "name": "Beta", "description": "Second starter", "includes": [], "harnesses": ["harness-b"], "mcp_count": 1 }
          ]
        }
        """;

        var starters = StarterCatalogParser.BuildFromJson(json);

        Assert.Equal(string.Empty, starters[0].LongDescription);
    }

    [Fact]
    public void BuildFromJson_MissingIncludesAndHarnesses_DefaultsToEmptyLists()
    {
        const string json = """
        {
          "starters": [
            { "id": "solo", "name": "Solo", "description": "No includes", "mcp_count": 0 }
          ]
        }
        """;

        var starters = StarterCatalogParser.BuildFromJson(json);

        Assert.Single(starters);
        Assert.NotNull(starters[0].Includes);
        Assert.Empty(starters[0].Includes);
        Assert.NotNull(starters[0].Harnesses);
        Assert.Empty(starters[0].Harnesses);
    }
}
