using System.Text.Json;
using System.Text.Json.Serialization;

namespace ActiveStack.Bootstrapper.Core;

/// <summary>
/// Parses the "windows backups list" response into <see cref="BackupEntry"/>
/// results, mirroring <see cref="InstallerSessionStateBuilder"/>'s
/// <c>BuildFromJson</c> style (D2, design.md). <c>created_at</c> is kept
/// verbatim as the engine's RFC3339 string.
/// </summary>
public static class BackupListParser
{
    private static readonly JsonSerializerOptions JsonOptions = new()
    {
        PropertyNameCaseInsensitive = true
    };

    public static IReadOnlyList<BackupEntry> BuildFromJson(string json)
    {
        var response = JsonSerializer.Deserialize<WindowsBackupsResponse>(json, JsonOptions)
            ?? new WindowsBackupsResponse();

        return (response.Backups ?? [])
            .Select(static entry => new BackupEntry(
                entry.Id ?? string.Empty,
                entry.CreatedAt ?? string.Empty,
                entry.Source ?? string.Empty,
                entry.Description ?? string.Empty,
                entry.FileCount,
                entry.Pinned,
                entry.Compressed,
                entry.DisplayLabel ?? string.Empty,
                entry.ManifestPath ?? string.Empty))
            .ToList();
    }

    private sealed class WindowsBackupsResponse
    {
        [JsonPropertyName("backups")]
        public List<WindowsBackupEntry>? Backups { get; init; }
    }

    private sealed class WindowsBackupEntry
    {
        [JsonPropertyName("id")]
        public string? Id { get; init; }

        [JsonPropertyName("created_at")]
        public string? CreatedAt { get; init; }

        [JsonPropertyName("source")]
        public string? Source { get; init; }

        [JsonPropertyName("description")]
        public string? Description { get; init; }

        [JsonPropertyName("file_count")]
        public int FileCount { get; init; }

        [JsonPropertyName("pinned")]
        public bool Pinned { get; init; }

        [JsonPropertyName("compressed")]
        public bool Compressed { get; init; }

        [JsonPropertyName("display_label")]
        public string? DisplayLabel { get; init; }

        [JsonPropertyName("manifest_path")]
        public string? ManifestPath { get; init; }
    }
}

public sealed record BackupEntry(
    string Id,
    string CreatedAt,
    string Source,
    string Description,
    int FileCount,
    bool Pinned,
    bool Compressed,
    string DisplayLabel,
    string ManifestPath);
