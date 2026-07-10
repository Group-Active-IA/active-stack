using System.Diagnostics;
using System.Text;
using ActiveStack.Bootstrapper.Core;

namespace ActiveStack.Bootstrapper.Host;

public sealed class ProcessInstallerEngineClient : IInstallerEngineClient
{
    private readonly string _enginePath;

    /// <summary>
    /// Test-only seam: when set, one-shot JSON commands are resolved through
    /// this delegate instead of spawning the real engine process. This lets
    /// Host.Tests exercise the full <c>ListBackupsAsync</c> /
    /// <c>ListStartersAsync</c> / <c>LoadUninstallOptionsAsync</c> /
    /// <c>RunBackupActionAsync</c> / <c>LoadSessionAsync</c> code paths
    /// (argv building + Core parsing) with a fed JSON string, with no live
    /// process (task 7.4). Mirrors the Go codebase's <c>SetXxxFn</c>
    /// test-seam convention (e.g. <c>install.SetExternalInstallFnWithDownload</c>).
    /// Internal — production callers never set this.
    /// </summary>
    internal Func<IReadOnlyList<string>, CancellationToken, Task<string>>? JsonCommandRunnerOverride { get; set; }

    /// <inheritdoc />
    public string Language { get; set; } = "en";

    public ProcessInstallerEngineClient(string enginePath)
    {
        _enginePath = enginePath;
    }

    public async Task<InstallerSessionState> LoadSessionAsync(CancellationToken cancellationToken = default)
    {
        // "windows detect" is the one subcommand that emits no user-facing
        // text and therefore never carries --lang (frozen L1 contract).
        var detectJson = await RunJsonCommandAsync(["windows", "detect"], cancellationToken);
        var agentsCsv = ResolveDetectedAgentsCsv(detectJson);
        var optionsJson = await RunJsonCommandAsync(["windows", "options", "--agent", agentsCsv, "--lang", Language], cancellationToken);
        return InstallerSessionStateBuilder.BuildFromJson(detectJson, optionsJson);
    }

    public IAsyncEnumerable<InstallProgressSnapshot> RunInstallAsync(
        IReadOnlyList<string> agents,
        string mode,
        IReadOnlyList<string> customIds,
        string? tier,
        CancellationToken cancellationToken = default)
    {
        return RunStreamingCommandAsync(BuildRunInstallArgs(agents, mode, customIds, tier), cancellationToken);
    }

    public async Task<IReadOnlyList<StarterChoice>> ListStartersAsync(CancellationToken cancellationToken = default)
    {
        var json = await RunJsonCommandAsync(["windows", "starters", "list", "--lang", Language], cancellationToken);
        return StarterCatalogParser.BuildFromJson(json);
    }

    public IAsyncEnumerable<InstallProgressSnapshot> RunStarterInstallAsync(
        string starterId,
        string projectPath,
        IReadOnlyList<string> agents,
        bool dryRun,
        CancellationToken cancellationToken = default)
    {
        return RunStreamingCommandAsync(BuildRunStarterInstallArgs(starterId, projectPath, agents, dryRun), cancellationToken);
    }

    public async Task<IReadOnlyList<BackupEntry>> ListBackupsAsync(CancellationToken cancellationToken = default)
    {
        var json = await RunJsonCommandAsync(["windows", "backups", "list", "--lang", Language], cancellationToken);
        return BackupListParser.BuildFromJson(json);
    }

    public async Task<BackupActionResult> RunBackupActionAsync(
        string action,
        string id,
        string? description,
        CancellationToken cancellationToken = default)
    {
        var args = InstallArguments.BuildBackupActionArgs(action, id, description);
        var json = await RunJsonCommandAsync(args, cancellationToken);
        return BackupActionResultParser.BuildFromJson(json);
    }

    public async Task<UninstallOptions> LoadUninstallOptionsAsync(CancellationToken cancellationToken = default)
    {
        var json = await RunJsonCommandAsync(["windows", "uninstall-options", "--lang", Language], cancellationToken);
        return UninstallOptionsParser.BuildFromJson(json);
    }

    public IAsyncEnumerable<InstallProgressSnapshot> RunUninstallAsync(
        IReadOnlyList<string> agents,
        string mode,
        string strategy,
        string? backupId,
        CancellationToken cancellationToken = default)
    {
        return RunStreamingCommandAsync(BuildRunUninstallArgs(agents, mode, strategy, backupId), cancellationToken);
    }

    /// <summary>
    /// Internal argv seams for the three streaming subcommands: the
    /// production Run*Async methods route through these so the client's
    /// <see cref="Language"/> threading is assertable in Host.Tests without
    /// spawning a live process (task 6.5; InternalsVisibleTo grants access).
    /// </summary>
    internal IReadOnlyList<string> BuildRunInstallArgs(
        IReadOnlyList<string> agents, string mode, IReadOnlyList<string> customIds, string? tier) =>
        InstallArguments.BuildInstallArgs(agents, mode, customIds, tier, Language);

    internal IReadOnlyList<string> BuildRunStarterInstallArgs(
        string starterId, string projectPath, IReadOnlyList<string> agents, bool dryRun) =>
        InstallArguments.BuildStarterInstallArgs(starterId, projectPath, agents, dryRun, yes: true, Language);

    internal IReadOnlyList<string> BuildRunUninstallArgs(
        IReadOnlyList<string> agents, string mode, string strategy, string? backupId) =>
        InstallArguments.BuildUninstallArgs(agents, mode, strategy, backupId, Language);

    /// <summary>
    /// Shared streaming helper for every JSONL event-stream subcommand
    /// ("windows install", "windows starters install", "windows uninstall").
    /// Differs only in the argument list; parsing is always
    /// <see cref="InstallProgressSnapshotParser"/> (D6, design.md).
    /// </summary>
    private async IAsyncEnumerable<InstallProgressSnapshot> RunStreamingCommandAsync(
        IReadOnlyList<string> args,
        [System.Runtime.CompilerServices.EnumeratorCancellation] CancellationToken cancellationToken)
    {
        using var process = CreateProcess(args);
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

    /// <summary>
    /// Shared one-shot JSON command helper for every non-streaming subcommand
    /// ("windows detect", "windows options", "windows starters list",
    /// "windows backups list", "windows backups &lt;action&gt;",
    /// "windows uninstall-options"). Uses <c>ArgumentList</c> (D6,
    /// design.md) — no string concatenation — so every token, including
    /// paths and free-text with spaces, is passed to the child process
    /// uncorrupted.
    /// </summary>
    private Task<string> RunJsonCommandAsync(IReadOnlyList<string> args, CancellationToken cancellationToken)
    {
        return JsonCommandRunnerOverride is not null
            ? JsonCommandRunnerOverride(args, cancellationToken)
            : RunJsonCommandViaProcessAsync(args, cancellationToken);
    }

    private async Task<string> RunJsonCommandViaProcessAsync(IReadOnlyList<string> args, CancellationToken cancellationToken)
    {
        using var process = CreateProcess(args);
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

    /// <summary>
    /// Internal seam (mirrors the <c>Build*Args</c> seams above; InternalsVisibleTo
    /// grants Host.Tests access): lets tests assert the <see cref="ProcessStartInfo"/>
    /// the client configures — in particular the UTF-8 stdout/stderr encoding
    /// (D2, design.md) — without spawning a live process.
    /// </summary>
    internal Process CreateProcess(IReadOnlyList<string> args)
    {
        var startInfo = new ProcessStartInfo
        {
            FileName = _enginePath,
            RedirectStandardOutput = true,
            RedirectStandardError = true,
            StandardOutputEncoding = Encoding.UTF8,
            StandardErrorEncoding = Encoding.UTF8,
            UseShellExecute = false,
            CreateNoWindow = true
        };

        foreach (var arg in args)
        {
            startInfo.ArgumentList.Add(arg);
        }

        return new Process { StartInfo = startInfo };
    }

    /// <summary>
    /// Resolves the comma-separated agent list to pass to "windows options"
    /// from a "windows detect" response, falling back to "claude" when no
    /// agents were detected (D5, design.md — replaces the old
    /// PickFirstAssistantId single-agent collapse). Extracted as an
    /// internal, pure, testable method so LoadSessionAsync's multi-agent CSV
    /// behavior can be asserted without spawning a live process (Host.Tests,
    /// task 6.3; InternalsVisibleTo already grants Host.Tests access).
    /// </summary>
    internal static string ResolveDetectedAgentsCsv(string detectJson)
    {
        var state = InstallerSessionStateBuilder.BuildFromJson(detectJson, """{"modes":[],"forced_components":[],"custom_components":[]}""");
        var agentIds = state.AssistantChoices
            .Select(static a => a.Id)
            .Where(static id => !string.IsNullOrWhiteSpace(id))
            .ToList();

        return agentIds.Count > 0 ? string.Join(",", agentIds) : "claude";
    }
}
