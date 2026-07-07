using System.Text.Json;
using System.Text.Json.Serialization;

namespace ActiveStack.Bootstrapper.Core;

/// <summary>
/// Parses the "windows backups restore|delete|rename" response into a
/// <see cref="BackupActionResult"/>, mirroring
/// <see cref="InstallerSessionStateBuilder"/>'s <c>BuildFromJson</c> style
/// (D2, design.md).
/// </summary>
public static class BackupActionResultParser
{
    private static readonly JsonSerializerOptions JsonOptions = new()
    {
        PropertyNameCaseInsensitive = true
    };

    public static BackupActionResult BuildFromJson(string json)
    {
        var response = JsonSerializer.Deserialize<WindowsBackupActionResponse>(json, JsonOptions)
            ?? new WindowsBackupActionResponse();

        return new BackupActionResult(response.Success, response.Message ?? string.Empty, response.Id ?? string.Empty);
    }

    private sealed class WindowsBackupActionResponse
    {
        [JsonPropertyName("success")]
        public bool Success { get; init; }

        [JsonPropertyName("message")]
        public string? Message { get; init; }

        [JsonPropertyName("id")]
        public string? Id { get; init; }
    }
}

public sealed record BackupActionResult(bool Success, string Message, string Id);
