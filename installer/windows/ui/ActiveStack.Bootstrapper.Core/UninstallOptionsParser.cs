using System.Text.Json;
using System.Text.Json.Serialization;

namespace ActiveStack.Bootstrapper.Core;

/// <summary>
/// Parses the "windows uninstall-options" response into an
/// <see cref="UninstallOptions"/> result, mirroring
/// <see cref="InstallerSessionStateBuilder"/>'s <c>BuildFromJson</c> style
/// (D2, design.md). Reuses <see cref="InstallTypeChoice"/> for modes since
/// the shape is identical to the "windows options" modes array.
/// </summary>
public static class UninstallOptionsParser
{
    private static readonly JsonSerializerOptions JsonOptions = new()
    {
        PropertyNameCaseInsensitive = true
    };

    public static UninstallOptions BuildFromJson(string json)
    {
        var response = JsonSerializer.Deserialize<WindowsUninstallOptionsResponse>(json, JsonOptions)
            ?? new WindowsUninstallOptionsResponse();

        var detectedAgents = (response.DetectedAgents ?? [])
            .Where(static id => !string.IsNullOrWhiteSpace(id))
            .ToList();

        var modes = (response.Modes ?? [])
            .Select(static mode => new InstallTypeChoice(mode.Id ?? string.Empty, mode.Label ?? string.Empty, mode.Description ?? string.Empty, mode.LongDescription ?? string.Empty))
            .ToList();

        var strategies = (response.Strategies ?? [])
            .Select(static strategy => new UninstallStrategyChoice(
                strategy.Id ?? string.Empty,
                strategy.Label ?? string.Empty,
                strategy.Description ?? string.Empty,
                strategy.Default,
                strategy.RequiresManifest,
                strategy.LongDescription ?? string.Empty))
            .ToList();

        return new UninstallOptions(detectedAgents, modes, strategies);
    }

    private sealed class WindowsUninstallOptionsResponse
    {
        [JsonPropertyName("detected_agents")]
        public List<string>? DetectedAgents { get; init; }

        [JsonPropertyName("modes")]
        public List<WindowsModeOption>? Modes { get; init; }

        [JsonPropertyName("strategies")]
        public List<WindowsUninstallStrategyOption>? Strategies { get; init; }
    }

    private sealed class WindowsUninstallStrategyOption
    {
        [JsonPropertyName("id")]
        public string? Id { get; init; }

        [JsonPropertyName("label")]
        public string? Label { get; init; }

        [JsonPropertyName("description")]
        public string? Description { get; init; }

        [JsonPropertyName("default")]
        public bool Default { get; init; }

        [JsonPropertyName("requires_manifest")]
        public bool RequiresManifest { get; init; }

        [JsonPropertyName("long_description")]
        public string? LongDescription { get; init; }
    }
}

public sealed record UninstallOptions(
    IReadOnlyList<string> DetectedAgents,
    IReadOnlyList<InstallTypeChoice> Modes,
    IReadOnlyList<UninstallStrategyChoice> Strategies);

public sealed record UninstallStrategyChoice(string Id, string Label, string Description, bool IsDefault, bool RequiresManifest, string LongDescription = "");
