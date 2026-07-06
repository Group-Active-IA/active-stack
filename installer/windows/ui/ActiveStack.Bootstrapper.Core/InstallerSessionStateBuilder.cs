using System.Text.Json;
using System.Text.Json.Serialization;

namespace ActiveStack.Bootstrapper.Core;

public static class InstallerSessionStateBuilder
{
    private static readonly JsonSerializerOptions JsonOptions = new()
    {
        PropertyNameCaseInsensitive = true
    };

    public static InstallerSessionState BuildFromJson(string detectJson, string optionsJson)
    {
        var detect = JsonSerializer.Deserialize<WindowsDetectResponse>(detectJson, JsonOptions)
            ?? new WindowsDetectResponse();
        var options = JsonSerializer.Deserialize<WindowsOptionsResponse>(optionsJson, JsonOptions)
            ?? new WindowsOptionsResponse();

        var assistants = (detect.DetectedAgents ?? [])
            .Where(static id => !string.IsNullOrWhiteSpace(id))
            .Select(static id => new AssistantChoice(id, ToAssistantLabel(id)))
            .ToList();

        var installTypes = (options.Modes ?? [])
            .Select(static mode => new InstallTypeChoice(mode.Id ?? string.Empty, mode.Label ?? string.Empty, mode.Description ?? string.Empty))
            .ToList();

        var forced = (options.ForcedComponents ?? [])
            .Select(static item => new ComponentChoice(item.Id ?? string.Empty, item.Label ?? string.Empty, item.Description ?? string.Empty, item.Recommended))
            .ToList();

        var forcedIds = forced
            .Select(static item => item.Id)
            .Where(static id => !string.IsNullOrWhiteSpace(id))
            .ToHashSet(StringComparer.OrdinalIgnoreCase);

        var custom = (options.CustomComponents ?? [])
            .Where(item => string.IsNullOrWhiteSpace(item.Id) || !forcedIds.Contains(item.Id))
            .Select(static item => new ComponentChoice(item.Id ?? string.Empty, item.Label ?? string.Empty, item.Description ?? string.Empty, item.Recommended))
            .ToList();

        var recommendedModeId = installTypes.Any(static m => m.Id == "full")
            ? "full"
            : installTypes.FirstOrDefault()?.Id;

        return new InstallerSessionState(
            assistants,
            assistants.FirstOrDefault()?.Id,
            installTypes,
            recommendedModeId,
            forced,
            custom);
    }

    private static string ToAssistantLabel(string id) => id switch
    {
        "claude" => "Claude",
        "codex" => "Codex",
        "opencode" => "OpenCode",
        "cursor" => "Cursor",
        "windsurf" => "Windsurf",
        _ => string.Concat(id[..1].ToUpperInvariant(), id[1..])
    };
}

public sealed record InstallerSessionState(
    IReadOnlyList<AssistantChoice> AssistantChoices,
    string? DefaultAssistantId,
    IReadOnlyList<InstallTypeChoice> InstallTypeChoices,
    string? RecommendedModeId,
    IReadOnlyList<ComponentChoice> ForcedComponents,
    IReadOnlyList<ComponentChoice> CustomComponents);

public sealed record AssistantChoice(string Id, string Label);

public sealed record InstallTypeChoice(string Id, string Label, string Description);

public sealed record ComponentChoice(string Id, string Label, string Description, bool Recommended);

internal sealed class WindowsDetectResponse
{
    [JsonPropertyName("detected_agents")]
    public List<string>? DetectedAgents { get; init; }
}

internal sealed class WindowsOptionsResponse
{
    [JsonPropertyName("modes")]
    public List<WindowsModeOption>? Modes { get; init; }

    [JsonPropertyName("forced_components")]
    public List<WindowsComponentOption>? ForcedComponents { get; init; }

    [JsonPropertyName("custom_components")]
    public List<WindowsComponentOption>? CustomComponents { get; init; }
}

internal sealed class WindowsModeOption
{
    [JsonPropertyName("id")]
    public string? Id { get; init; }

    [JsonPropertyName("label")]
    public string? Label { get; init; }

    [JsonPropertyName("description")]
    public string? Description { get; init; }
}

internal sealed class WindowsComponentOption
{
    [JsonPropertyName("id")]
    public string? Id { get; init; }

    [JsonPropertyName("label")]
    public string? Label { get; init; }

    [JsonPropertyName("description")]
    public string? Description { get; init; }

    [JsonPropertyName("recommended")]
    public bool Recommended { get; init; }
}
