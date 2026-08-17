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
            .Select(static mode => new InstallTypeChoice(mode.Id ?? string.Empty, mode.Label ?? string.Empty, mode.Description ?? string.Empty, mode.LongDescription ?? string.Empty))
            .ToList();

        var forced = (options.ForcedComponents ?? [])
            .Select(static item => new ComponentChoice(item.Id ?? string.Empty, item.Label ?? string.Empty, item.Description ?? string.Empty, item.Recommended, item.LongDescription ?? string.Empty))
            .ToList();

        var forcedIds = forced
            .Select(static item => item.Id)
            .Where(static id => !string.IsNullOrWhiteSpace(id))
            .ToHashSet(StringComparer.OrdinalIgnoreCase);

        var custom = (options.CustomComponents ?? [])
            .Where(item => string.IsNullOrWhiteSpace(item.Id) || !forcedIds.Contains(item.Id))
            .Select(static item => new ComponentChoice(item.Id ?? string.Empty, item.Label ?? string.Empty, item.Description ?? string.Empty, item.Recommended, item.LongDescription ?? string.Empty))
            .ToList();

        var recommendedModeId = installTypes.Any(static m => m.Id == "full")
            ? "full"
            : installTypes.FirstOrDefault()?.Id;

        var tierCapableAgents = (options.TierCapableAgents ?? [])
            .Where(static id => !string.IsNullOrWhiteSpace(id))
            .ToList();

        var dependencies = (detect.Dependencies?.Dependencies ?? [])
            .Select(static d => new DependencyChoice(
                d.Name ?? string.Empty,
                d.Required,
                d.Installed,
                d.Version ?? string.Empty,
                d.InstallHint ?? string.Empty,
                d.InstallCommand ?? string.Empty))
            .ToList();

        var permissionTierChoices = (options.PermissionTiers ?? [])
            .Select(static tier => new PermissionTierChoice(
                tier.Id ?? string.Empty,
                tier.Label ?? string.Empty,
                tier.Description ?? string.Empty,
                tier.Default,
                tier.Warning,
                tier.LongDescription ?? string.Empty))
            .ToList();

        return new InstallerSessionState(
            assistants,
            assistants.FirstOrDefault()?.Id,
            installTypes,
            recommendedModeId,
            forced,
            custom,
            options.TierCapable,
            tierCapableAgents,
            permissionTierChoices,
            dependencies);
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
    IReadOnlyList<ComponentChoice> CustomComponents,
    bool TierCapable = false,
    IReadOnlyList<string>? TierCapableAgents = null,
    IReadOnlyList<PermissionTierChoice>? PermissionTierChoices = null,
    IReadOnlyList<DependencyChoice>? Dependencies = null)
{
    public IReadOnlyList<string> TierCapableAgents { get; init; } = TierCapableAgents ?? [];

    public IReadOnlyList<PermissionTierChoice> PermissionTierChoices { get; init; } = PermissionTierChoices ?? [];

    public IReadOnlyList<DependencyChoice> Dependencies { get; init; } = Dependencies ?? [];
}

public sealed record AssistantChoice(string Id, string Label);

public sealed record InstallTypeChoice(string Id, string Label, string Description, string LongDescription = "");

public sealed record ComponentChoice(string Id, string Label, string Description, bool Recommended, string LongDescription = "");

public sealed record PermissionTierChoice(string Id, string Label, string Description, bool IsDefault, string? Warning, string LongDescription = "");

public sealed record DependencyChoice(string Name, bool Required, bool Installed, string Version, string InstallHint, string InstallCommand);

internal sealed class WindowsDetectResponse
{
    [JsonPropertyName("detected_agents")]
    public List<string>? DetectedAgents { get; init; }

    [JsonPropertyName("dependencies")]
    public WindowsDependenciesResponse? Dependencies { get; init; }
}

internal sealed class WindowsDependenciesResponse
{
    [JsonPropertyName("dependencies")]
    public List<WindowsDependencyEntry>? Dependencies { get; init; }

    [JsonPropertyName("all_present")]
    public bool AllPresent { get; init; }

    [JsonPropertyName("missing_required")]
    public List<string>? MissingRequired { get; init; }

    [JsonPropertyName("missing_optional")]
    public List<string>? MissingOptional { get; init; }
}

internal sealed class WindowsDependencyEntry
{
    [JsonPropertyName("name")]
    public string? Name { get; init; }

    [JsonPropertyName("required")]
    public bool Required { get; init; }

    [JsonPropertyName("installed")]
    public bool Installed { get; init; }

    [JsonPropertyName("version")]
    public string? Version { get; init; }

    [JsonPropertyName("install_hint")]
    public string? InstallHint { get; init; }

    [JsonPropertyName("install_command")]
    public string? InstallCommand { get; init; }
}

internal sealed class WindowsOptionsResponse
{
    [JsonPropertyName("modes")]
    public List<WindowsModeOption>? Modes { get; init; }

    [JsonPropertyName("forced_components")]
    public List<WindowsComponentOption>? ForcedComponents { get; init; }

    [JsonPropertyName("custom_components")]
    public List<WindowsComponentOption>? CustomComponents { get; init; }

    [JsonPropertyName("tier_capable")]
    public bool TierCapable { get; init; }

    [JsonPropertyName("tier_capable_agents")]
    public List<string>? TierCapableAgents { get; init; }

    [JsonPropertyName("permission_tiers")]
    public List<WindowsPermissionTierOption>? PermissionTiers { get; init; }
}

internal sealed class WindowsPermissionTierOption
{
    [JsonPropertyName("id")]
    public string? Id { get; init; }

    [JsonPropertyName("label")]
    public string? Label { get; init; }

    [JsonPropertyName("description")]
    public string? Description { get; init; }

    [JsonPropertyName("default")]
    public bool Default { get; init; }

    [JsonPropertyName("warning")]
    public string? Warning { get; init; }

    [JsonPropertyName("long_description")]
    public string? LongDescription { get; init; }
}

internal sealed class WindowsModeOption
{
    [JsonPropertyName("id")]
    public string? Id { get; init; }

    [JsonPropertyName("label")]
    public string? Label { get; init; }

    [JsonPropertyName("description")]
    public string? Description { get; init; }

    [JsonPropertyName("long_description")]
    public string? LongDescription { get; init; }
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

    [JsonPropertyName("long_description")]
    public string? LongDescription { get; init; }
}
