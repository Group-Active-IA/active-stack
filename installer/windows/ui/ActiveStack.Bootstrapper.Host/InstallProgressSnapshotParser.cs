using System.Text.Json;
using System.Text.Json.Serialization;

namespace ActiveStack.Bootstrapper.Host;

internal static class InstallProgressSnapshotParser
{
    public static InstallProgressSnapshot Parse(string jsonLine)
    {
        InstallProgressPayload? payload;
        try
        {
            payload = JsonSerializer.Deserialize<InstallProgressPayload>(jsonLine);
        }
        catch (JsonException)
        {
            // D2, design.md (Bug B): the engine may interleave a non-JSON
            // diagnostic line on stdout (e.g. a stray warning). One
            // unparseable line must not abort the whole stream with an
            // uncaught exception — treat it as an ignorable/log-type
            // snapshot carrying the raw line instead. A genuine engine
            // failure is still surfaced via the process exit code, not here.
            return ToLogSnapshot(jsonLine);
        }

        if (payload is null)
        {
            throw new InvalidOperationException("Installer engine returned an empty progress event.");
        }

        return new InstallProgressSnapshot(
            payload.Type ?? "step_output",
            payload.Phase,
            payload.StepId,
            payload.Message,
            payload.Success,
            payload.Details,
            payload.Timestamp);
    }

    private static InstallProgressSnapshot ToLogSnapshot(string rawLine) =>
        new("log", null, null, rawLine, false);

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
