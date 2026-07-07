namespace ActiveStack.Bootstrapper.Core;

/// <summary>
/// Pure argv builders for every "windows" subcommand exposed by the Go
/// engine. Every builder returns an ordered <see cref="IReadOnlyList{T}"/> of
/// tokens suitable for <c>ProcessStartInfo.ArgumentList</c> — never a
/// concatenated string — so paths and free-text values containing spaces
/// survive as a single element (D3, design.md).
/// </summary>
public static class InstallArguments
{
    /// <summary>
    /// Builds the argv for "windows install". Agents are joined with a comma
    /// into a single <c>--agent</c> value. <c>--custom</c> is included only
    /// when <paramref name="mode"/> is "custom" and <paramref name="customIds"/>
    /// is non-empty. <c>--tier</c> is included only when <paramref name="tier"/>
    /// is non-null/non-empty.
    /// </summary>
    public static IReadOnlyList<string> BuildInstallArgs(
        IEnumerable<string> agents,
        string mode,
        IEnumerable<string> customIds,
        string? tier)
    {
        var args = new List<string>
        {
            "windows", "install",
            "--agent", JoinCsv(agents),
            "--mode", mode
        };

        var customList = customIds.ToList();
        if (string.Equals(mode, "custom", StringComparison.Ordinal) && customList.Count > 0)
        {
            args.Add("--custom");
            args.Add(JoinCsv(customList));
        }

        if (!string.IsNullOrEmpty(tier))
        {
            args.Add("--tier");
            args.Add(tier);
        }

        return args;
    }

    /// <summary>
    /// Builds the argv for "windows starters install". The project path is a
    /// single list element (never concatenated) so spaces survive.
    /// </summary>
    public static IReadOnlyList<string> BuildStarterInstallArgs(
        string starterId,
        string projectPath,
        IEnumerable<string> agents,
        bool dryRun,
        bool yes)
    {
        var args = new List<string>
        {
            "windows", "starters", "install",
            "--starter", starterId,
            "--project", projectPath,
            "--agent", JoinCsv(agents)
        };

        if (dryRun)
        {
            args.Add("--dry-run");
        }

        if (yes)
        {
            args.Add("--yes");
        }

        return args;
    }

    /// <summary>
    /// Builds the argv for "windows backups restore|delete|rename". The
    /// <c>--description</c> flag is included only for the "rename" action
    /// with a non-empty description.
    /// </summary>
    public static IReadOnlyList<string> BuildBackupActionArgs(string action, string id, string? description)
    {
        var args = new List<string>
        {
            "windows", "backups", action,
            "--id", id
        };

        if (string.Equals(action, "rename", StringComparison.Ordinal) && !string.IsNullOrEmpty(description))
        {
            args.Add("--description");
            args.Add(description);
        }

        return args;
    }

    /// <summary>
    /// Builds the argv for "windows uninstall". When <paramref name="strategy"/>
    /// is "restore", includes <c>--restore-manifest &lt;path&gt;</c> — the flag
    /// confirmed by reading the engine's <c>ParseUninstallFlags</c>
    /// (cmd/active-stack/headless/uninstall_flags.go), which resolves Open
    /// Question O1 from design.md.
    /// </summary>
    public static IReadOnlyList<string> BuildUninstallArgs(
        IEnumerable<string> agents,
        string mode,
        string strategy,
        string? restoreManifestPath)
    {
        var args = new List<string>
        {
            "windows", "uninstall",
            "--agent", JoinCsv(agents),
            "--mode", mode,
            "--strategy", strategy
        };

        if (string.Equals(strategy, "restore", StringComparison.Ordinal) && !string.IsNullOrEmpty(restoreManifestPath))
        {
            args.Add("--restore-manifest");
            args.Add(restoreManifestPath);
        }

        return args;
    }

    private static string JoinCsv(IEnumerable<string> values) => string.Join(",", values);
}
