using System.Diagnostics;
using System.Text.Json;
using System.Text.Json.Serialization;
using ActiveStack.Bootstrapper.Core;

namespace ActiveStack.Bootstrapper.Host;

public sealed class ProcessInstallerEngineClient : IInstallerEngineClient
{
    private readonly string _enginePath;

    public ProcessInstallerEngineClient(string enginePath)
    {
        _enginePath = enginePath;
    }

    public async Task<InstallerSessionState> LoadSessionAsync(CancellationToken cancellationToken = default)
    {
        var detectJson = await RunJsonCommandAsync("windows detect", cancellationToken);
        var assistantId = PickFirstAssistantId(detectJson);
        var optionsJson = await RunJsonCommandAsync($"windows options --agent {assistantId}", cancellationToken);
        return InstallerSessionStateBuilder.BuildFromJson(detectJson, optionsJson);
    }

    public async IAsyncEnumerable<InstallProgressSnapshot> RunInstallAsync(
        string assistantId,
        string installTypeId,
        [System.Runtime.CompilerServices.EnumeratorCancellation] CancellationToken cancellationToken = default)
    {
        var startInfo = new ProcessStartInfo
        {
            FileName = _enginePath,
            Arguments = $"windows install --agent {assistantId} --mode {installTypeId}",
            RedirectStandardOutput = true,
            RedirectStandardError = true,
            UseShellExecute = false,
            CreateNoWindow = true
        };

        using var process = new Process { StartInfo = startInfo };
        process.Start();

        var errorTask = process.StandardError.ReadToEndAsync(cancellationToken);

        while (!process.StandardOutput.EndOfStream)
        {
            cancellationToken.ThrowIfCancellationRequested();

            var line = await process.StandardOutput.ReadLineAsync(cancellationToken);
            if (string.IsNullOrWhiteSpace(line))
            {
                continue;
            }

            yield return InstallProgressSnapshotParser.Parse(line);
        }

        await process.WaitForExitAsync(cancellationToken);
        var error = await errorTask;

        if (process.ExitCode != 0)
        {
            throw new InvalidOperationException(string.IsNullOrWhiteSpace(error) ? $"Installer engine exited with code {process.ExitCode}." : error.Trim());
        }
    }

    private async Task<string> RunJsonCommandAsync(string arguments, CancellationToken cancellationToken)
    {
        var startInfo = new ProcessStartInfo
        {
            FileName = _enginePath,
            Arguments = arguments,
            RedirectStandardOutput = true,
            RedirectStandardError = true,
            UseShellExecute = false,
            CreateNoWindow = true
        };

        using var process = new Process { StartInfo = startInfo };
        process.Start();

        var outputTask = process.StandardOutput.ReadToEndAsync(cancellationToken);
        var errorTask = process.StandardError.ReadToEndAsync(cancellationToken);
        await process.WaitForExitAsync(cancellationToken);

        var output = await outputTask;
        var error = await errorTask;

        if (process.ExitCode != 0)
        {
            throw new InvalidOperationException(string.IsNullOrWhiteSpace(error) ? $"Installer engine exited with code {process.ExitCode}." : error.Trim());
        }

        return output;
    }

    private static string PickFirstAssistantId(string detectJson)
    {
        var state = InstallerSessionStateBuilder.BuildFromJson(detectJson, """{"modes":[],"forced_components":[],"custom_components":[]}""");
        return state.DefaultAssistantId ?? "claude";
    }
}
