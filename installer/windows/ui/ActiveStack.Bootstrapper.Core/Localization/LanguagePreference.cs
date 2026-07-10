using System.Text.Json;
using System.Text.Json.Nodes;

namespace ActiveStack.Bootstrapper.Core.Localization;

/// <summary>
/// Loads and saves the selected UI language from <c>~/.active-stack/config.json</c>
/// as <c>{"language":"&lt;lang&gt;"}</c>. Merge-friendly: <see cref="Save"/>
/// preserves any other keys already present in the file rather than
/// overwriting the whole document (the file is shared user state — the TUI
/// reads it too per the plan). The home directory is injectable so the
/// round-trip is testable against a temp directory (gui-language-page, L4).
/// </summary>
public static class LanguagePreference
{
    private const string LanguageKey = "language";

    /// <summary>
    /// Returns the persisted language, or <c>null</c> when the file, the
    /// directory, or the <c>language</c> key is absent — and also when the
    /// file exists but is empty or malformed (never throws).
    /// </summary>
    public static string? Load(string homeDir)
    {
        var path = ConfigPath(homeDir);
        if (!File.Exists(path))
        {
            return null;
        }

        var root = TryReadJsonObject(path);
        if (root is null)
        {
            return null;
        }

        return root.TryGetPropertyValue(LanguageKey, out var value) && value is not null
            ? value.GetValue<string>()
            : null;
    }

    /// <summary>
    /// Reads the existing <c>config.json</c> (if any), sets <c>language</c>,
    /// and rewrites the file — preserving every other key already present.
    /// A malformed/empty existing file is treated as an empty document
    /// rather than propagating the parse failure.
    /// </summary>
    public static void Save(string homeDir, string lang)
    {
        var path = ConfigPath(homeDir);
        var root = TryReadJsonObject(path) ?? new JsonObject();

        root[LanguageKey] = lang;

        Directory.CreateDirectory(Path.GetDirectoryName(path)!);
        File.WriteAllText(path, root.ToJsonString(new JsonSerializerOptions { WriteIndented = false }));
    }

    private static string ConfigPath(string homeDir) => Path.Combine(homeDir, ".active-stack", "config.json");

    private static JsonObject? TryReadJsonObject(string path)
    {
        if (!File.Exists(path))
        {
            return null;
        }

        try
        {
            var text = File.ReadAllText(path);
            if (string.IsNullOrWhiteSpace(text))
            {
                return null;
            }

            return JsonNode.Parse(text) as JsonObject;
        }
        catch (JsonException)
        {
            return null;
        }
    }
}
