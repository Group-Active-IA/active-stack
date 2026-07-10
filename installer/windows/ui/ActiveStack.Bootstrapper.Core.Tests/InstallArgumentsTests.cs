using ActiveStack.Bootstrapper.Core;
using Xunit;

namespace ActiveStack.Bootstrapper.Core.Tests;

public sealed class InstallArgumentsTests
{
    [Fact]
    public void BuildInstallArgs_WithCustomComponentsAndTier_ReturnsExactTokenListIncludingLang()
    {
        var args = InstallArguments.BuildInstallArgs(
            agents: ["claude", "codex"],
            mode: "custom",
            customIds: ["a", "b"],
            tier: "balanceado",
            lang: "en");

        Assert.Equal(
            ["windows", "install", "--agent", "claude,codex", "--mode", "custom", "--custom", "a,b", "--tier", "balanceado", "--lang", "en"],
            args);
    }

    [Fact]
    public void BuildInstallArgs_OmitsCustomAndTierWhenNotApplicable()
    {
        var args = InstallArguments.BuildInstallArgs(
            agents: ["claude"],
            mode: "full",
            customIds: [],
            tier: null,
            lang: "es");

        Assert.Contains("--agent", args);
        Assert.Contains("claude", args);
        Assert.Contains("--mode", args);
        Assert.Contains("full", args);
        Assert.DoesNotContain("--custom", args);
        Assert.DoesNotContain("--tier", args);
        Assert.Contains("--lang", args);
        Assert.Contains("es", args);
    }

    [Fact]
    public void BuildInstallArgs_IgnoresCustomIdsWhenModeIsNotCustom()
    {
        var args = InstallArguments.BuildInstallArgs(
            agents: ["claude"],
            mode: "lite",
            customIds: ["a", "b"],
            tier: null,
            lang: "en");

        Assert.DoesNotContain("--custom", args);
    }

    [Fact]
    public void BuildStarterInstallArgs_KeepsProjectPathAsSingleElement()
    {
        var args = InstallArguments.BuildStarterInstallArgs(
            starterId: "alpha",
            projectPath: "C:/My Projects/app",
            agents: ["claude", "opencode"],
            dryRun: true,
            yes: true,
            lang: "es");

        var projectIndex = args.ToList().IndexOf("--project");
        Assert.True(projectIndex >= 0);
        Assert.Equal("C:/My Projects/app", args[projectIndex + 1]);

        var agentIndex = args.ToList().IndexOf("--agent");
        Assert.Equal("claude,opencode", args[agentIndex + 1]);

        Assert.Contains("--dry-run", args);
        Assert.Contains("--yes", args);

        var langIndex = args.ToList().IndexOf("--lang");
        Assert.True(langIndex >= 0);
        Assert.Equal("es", args[langIndex + 1]);
    }

    [Fact]
    public void BuildStarterInstallArgs_OmitsFlagsWhenNotSet()
    {
        var args = InstallArguments.BuildStarterInstallArgs(
            starterId: "alpha",
            projectPath: "C:/app",
            agents: ["claude"],
            dryRun: false,
            yes: false,
            lang: "en");

        Assert.DoesNotContain("--dry-run", args);
        Assert.DoesNotContain("--yes", args);
        Assert.Contains("--lang", args);
        Assert.Contains("en", args);
    }

    [Fact]
    public void BuildBackupActionArgs_RenameWithDescription_IncludesDescriptionAsSingleElement()
    {
        var args = InstallArguments.BuildBackupActionArgs("rename", "backup-1", "before upgrade");

        var descIndex = args.ToList().IndexOf("--description");
        Assert.True(descIndex >= 0);
        Assert.Equal("before upgrade", args[descIndex + 1]);
    }

    [Fact]
    public void BuildBackupActionArgs_DeleteOrEmptyDescription_OmitsDescription()
    {
        var deleteArgs = InstallArguments.BuildBackupActionArgs("delete", "backup-1", null);
        Assert.Equal(["windows", "backups", "delete", "--id", "backup-1"], deleteArgs);

        var restoreArgs = InstallArguments.BuildBackupActionArgs("restore", "backup-1", "");
        Assert.DoesNotContain("--description", restoreArgs);

        var renameEmptyDescArgs = InstallArguments.BuildBackupActionArgs("rename", "backup-1", "");
        Assert.DoesNotContain("--description", renameEmptyDescArgs);
    }

    [Fact]
    public void BuildUninstallArgs_TargetedStrategy_IncludesStrategyAndLang()
    {
        var args = InstallArguments.BuildUninstallArgs(
            agents: ["claude"],
            mode: "full",
            strategy: "targeted",
            restoreManifestPath: null,
            lang: "en");

        Assert.Equal(["windows", "uninstall", "--agent", "claude", "--mode", "full", "--strategy", "targeted", "--lang", "en"], args);
    }

    [Fact]
    public void BuildUninstallArgs_RestoreStrategy_IncludesRestoreManifestFlagAndLang()
    {
        var args = InstallArguments.BuildUninstallArgs(
            agents: ["claude"],
            mode: "full",
            strategy: "restore",
            restoreManifestPath: "C:/backups/backup-1/manifest.json",
            lang: "es");

        Assert.Contains("--strategy", args);
        Assert.Contains("restore", args);

        var manifestIndex = args.ToList().IndexOf("--restore-manifest");
        Assert.True(manifestIndex >= 0);
        Assert.Equal("C:/backups/backup-1/manifest.json", args[manifestIndex + 1]);

        var langIndex = args.ToList().IndexOf("--lang");
        Assert.True(langIndex >= 0);
        Assert.Equal("es", args[langIndex + 1]);
    }
}
