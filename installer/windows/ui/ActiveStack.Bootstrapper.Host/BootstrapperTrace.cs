using System.Text;
using System.IO;

namespace ActiveStack.Bootstrapper.Host;

internal static class BootstrapperTrace
{
    private static readonly object Sync = new();
    private static readonly string LogPath = Path.Combine(Path.GetTempPath(), "ActiveStack.Bootstrapper.Host.runtime.log");
    private static bool _enabled;

    public static void Configure(bool enabled)
    {
        _enabled = enabled;
    }

    public static void Write(string message)
    {
        if (!_enabled)
        {
            return;
        }

        try
        {
            var line = $"{DateTime.UtcNow:O} [tid:{Environment.CurrentManagedThreadId}] {message}{Environment.NewLine}";
            lock (Sync)
            {
                File.AppendAllText(LogPath, line, Encoding.UTF8);
            }
        }
        catch
        {
        }
    }
}
