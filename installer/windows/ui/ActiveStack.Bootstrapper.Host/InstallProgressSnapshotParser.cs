using System.Text.Json;
using System.Text.Json.Serialization;

namespace ActiveStack.Bootstrapper.Host;

internal static class InstallProgressSnapshotParser
{
    public static InstallProgressSnapshot Parse(string jsonLine)
    {
        var payload = JsonSerializer.Deserialize<InstallProgressPayload>(jsonLine)
            ?? throw new InvalidOperationException("Installer engine returned an empty progress event.");

        return new InstallProgressSnapshot(
            payload.Type ?? "step_output",
            payload.Phase,
            payload.StepId,
            payload.Message,
            payload.Success,
            payload.Details,
            payload.Timestamp);
    }

    private sealed class InstallProgressPayload
    {
        [JsonPropertyName("type")]
        public string? Type { get; init; }

        [JsonPropertyName("phase")]
        public string? Phase { get; init; }

        [JsonPropertyName("step_id")]
        public string? StepId { get; init; }

        [JsonPropertyName("message")]
        public string? Message { get; init; }

        [JsonPropertyName("details")]
        public string? Details { get; init; }

        [JsonPropertyName("success")]
        public bool Success { get; init; }

        [JsonPropertyName("timestamp")]
        public string? Timestamp { get; init; }
    }
}
