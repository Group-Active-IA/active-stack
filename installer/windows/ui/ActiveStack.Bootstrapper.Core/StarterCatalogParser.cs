using System.Text.Json;
using System.Text.Json.Serialization;

namespace ActiveStack.Bootstrapper.Core;

/// <summary>
/// Parses the "windows starters list" response into <see cref="StarterChoice"/>
/// results, mirroring <see cref="InstallerSessionStateBuilder"/>'s
/// <c>BuildFromJson</c> style (D2, design.md).
/// </summary>
public static class StarterCatalogParser
{
    private static readonly JsonSerializerOptions JsonOptions = new()
    {
        PropertyNameCaseInsensitive = true
    };

    public static IReadOnlyList<StarterChoice> BuildFromJson(string json)
    {
        var response = JsonSerializer.Deserialize<WindowsStartersResponse>(json, JsonOptions)
            ?? new WindowsStartersResponse();

        return (response.Starters ?? [])
            .Select(static entry => new StarterChoice(
                entry.Id ?? string.Empty,
                entry.Name ?? string.Empty,
                entry.Description ?? string.Empty,
                entry.Includes ?? [],
                entry.Harnesses ?? [],
                entry.McpCount,
                entry.LongDescription ?? string.Empty))
            .ToList();
    }

    private sealed class WindowsStartersResponse
    {
        [JsonPropertyName("starters")]
        public List<WindowsStarterEntry>? Starters { get; init; }
    }

    private sealed class WindowsStarterEntry
    {
        [JsonPropertyName("id")]
        public string? Id { get; init; }

        [JsonPropertyName("name")]
        public string? Name { get; init; }

        [JsonPropertyName("description")]
        public string? Description { get; init; }

        [JsonPropertyName("includes")]
        public List<string>? Includes { get; init; }

        [JsonPropertyName("harnesses")]
        public List<string>? Harnesses { get; init; }

        [JsonPropertyName("mcp_count")]
        public int McpCount { get; init; }

        [JsonPropertyName("long_description")]
        public string? LongDescription { get; init; }
    }
}

public sealed record StarterChoice(
    string Id,
    string Name,
    string Description,
    IReadOnlyList<string> Includes,
    IReadOnlyList<string> Harnesses,
    int McpCount,
    string LongDescription = "");
