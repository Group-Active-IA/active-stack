package i18n

// tableEN is the English string table (design D1). Registered into the
// package-level tables map via init(). Keys are dot-namespaced and grouped by
// surface; table_es.go MUST carry the identical key set (enforced by
// TestTableKeyParity in parity_test.go).
var tableEN = map[string]string{
	// windows.install.* — install stream literals.
	"windows.install.starting":      "Starting installation.",
	"windows.install.finished_ok":   "Installation finished successfully.",
	"windows.install.finished_fail": "Installation failed.",

	// windows.uninstall.* — uninstall stream literals.
	"windows.uninstall.starting":      "Starting uninstall.",
	"windows.uninstall.finished_ok":   "Uninstall finished successfully.",
	"windows.uninstall.finished_fail": "Uninstall failed.",
	"windows.uninstall.plan_failed":   "Uninstall plan failed.",

	// windows.phase.* — install phase messages.
	"windows.phase.prepare":  "Preparing installation.",
	"windows.phase.apply":    "Applying installation steps.",
	"windows.phase.rollback": "Rolling back changes.",
	"windows.phase.default":  "Running installation.",

	// windows.uninstall.phase.* — uninstall phase messages.
	"windows.uninstall.phase.prepare":  "Preparing uninstall.",
	"windows.uninstall.phase.apply":    "Removing installed items.",
	"windows.uninstall.phase.rollback": "Restoring previous state.",
	"windows.uninstall.phase.default":  "Running uninstall.",

	// windows.step.* — per-step status messages.
	"windows.step.started":     "Step started.",
	"windows.step.succeeded":   "Step completed.",
	"windows.step.failed":      "Step failed.",
	"windows.step.rolled_back": "Rollback step completed.",
	"windows.step.degraded":    "Step completed with warnings.",

	// windows.dryrun.* — dry-run step messages.
	"windows.dryrun.planned": "Dry-run step planned.",
	"windows.dryrun.listed":  "Dry-run step listed.",

	// component.fallback.* — component-description fallbacks.
	"component.fallback.security":    "Basic protection for safer setup. This is always installed.",
	"component.fallback.external":    "Downloads and configures an external tool.",
	"component.fallback.config":      "Applies recommended configuration.",
	"component.fallback.skill":       "Adds guided workflow helpers.",
	"component.fallback.generic_fmt": "Installs %s.",

	// mode.{lite,full,custom}.{label,desc,long} — install-mode options.
	"mode.lite.label": "Quick",
	"mode.lite.desc":  "Fast setup to start working right away.",
	"mode.lite.long":  "Quick mode installs only the minimal useful substrate so you can start right away. It configures the essential harnesses—OpenSpec, Engram, Context7, the SDD orchestrator, safe permissions, and the starter command—and skips the foundation skills and advanced configuration screens. Choose it when you want to start now and don't need the full guided foundation; you can always add the rest later via Complete or Custom mode.",

	"mode.full.label": "Complete",
	"mode.full.desc":  "Full recommended setup with all key tools.",
	"mode.full.long":  "Complete mode installs the whole recommended kit: the Quick substrate plus the foundation skills (active-orchestrator, kb-creator, roadmap-generator, agent-instruction, skill-registry) and the third-party ones (find-skill, skill-creator). It leaves the project ready for the full OPSX cycle with the guided foundation included. It takes a bit longer than Quick but leaves nothing important out. Choose it when you want everything from day one instead of adding pieces by hand.",

	"mode.custom.label": "Custom",
	"mode.custom.desc":  "Choose exactly what to install.",
	"mode.custom.long":  "Custom mode lets you hand-pick exactly which harnesses to install, one by one. Nothing outside your selection gets installed, except the mandatory security floor (permissions), which always ships. Use it when you already know your stack and want full control over what stays and what doesn't. If you'd rather have the decision made for you, Quick or Complete are more direct.",

	// tier.{estricto,balanceado,bypass}.{label,desc,long} + tier.bypass.warning.
	"tier.estricto.label":      "Strict",
	"tier.estricto.codex.desc": "Codex: Read-only by default",
	"tier.estricto.codex.long": "Codex runs in read-only mode. Actions that require additional permissions need explicit approval when the interface supports requests. Otherwise, those actions remain blocked.",
	"tier.estricto.desc":       "Agent must ask for every operation. Highest friction, highest security.",
	"tier.estricto.long":       "Strict configures the agent's permissions to ask for confirmation before every operation. It applies the most restrictive rule set on top of the security floor: nothing runs without your sign-off. It is the highest-friction, highest-security level. Choose it for sensitive projects or when you're starting out and want to review everything before it happens; if it slows you down, Balanced loosens the repetitive parts.",

	"tier.balanceado.label": "Balanced",
	"tier.balanceado.desc":  "Curated allow-list for safe, repetitive operations. Recommended starting point.",
	"tier.balanceado.long":  "Balanced configures a curated allow-list so safe, repetitive operations run without interruption while riskier actions still ask for confirmation. It is the middle ground between friction and autonomy on top of the security floor. It is the recommended starting point for most projects. Choose it when you want flow without giving up the important controls; move to Strict if you need to review more, or to Bypass for less friction.",

	"tier.bypass.label":   "Bypass",
	"tier.bypass.desc":    "Full autonomy opt-in. The security floor deny-list still applies.",
	"tier.bypass.long":    "Bypass grants the agent full autonomy to act without asking. The security floor deny-list still blocks the most dangerous operations—that layer never turns off—but everything else runs unattended. Choose it only when you fully trust the flow and friction is holding you back more than protecting you. Use it with caution: it is the lowest-control level.",
	"tier.bypass.warning": "Bypass: autonomous mode — the security floor still applies",

	// claude-native-opsx: layered Claude delivery notices.
	"claude.breaking.notice":           "BREAKING (Claude): the OPSX coordinator no longer lives in CLAUDE.md. It is now a skill you invoke with /sdd-orchestrator; normal sessions keep only the universal rules. After compacting the conversation, invoke /sdd-orchestrator again. Your own CLAUDE.md content was preserved.",
	"install.help.claude_user_content": "Claude note: whatever you write in your own CLAUDE.md reaches every session and subagent; Active Stack cannot filter it. The installer only manages its marked sections.",
	"update.breaking.nosync":           "Claude: this install has no selection record, so CLAUDE.md was not modified. BREAKING change: the OPSX coordinator is now the /sdd-orchestrator skill. Run `active-stack install` once with your current mode to migrate.",
	"update.project.legacy":            "Project %s still has the legacy coordinator block in %s; it was not rewritten. Migrate it with: active-stack starter add <starter> --project %s",
	"install.project.manual_cleanup":   "Claude project install: to remove it manually, delete %s and the active-stack:universal-rules section of %s (project uninstall is not supported yet).",
	"uninstall.flag.overwrite_edited":  "with --strategy restore, restore CLAUDE.md and skill files even if they were edited after the migration (their current content is saved first; --yes does not imply this)",

	// strategy.{targeted,restore}.{label,desc,long} — uninstall strategies.
	"strategy.targeted.label": "Targeted",
	"strategy.targeted.desc":  "Reverse each installed harness individually.",
	"strategy.targeted.long":  "Targeted reverses each installed harness individually, undoing only what this tool changed. It doesn't need a backup manifest: it walks through what was installed and removes it piece by piece. It is the recommended default for most uninstalls. Choose it when you want to remove what Active Stack installed without touching the rest of your configuration.",

	"strategy.restore.label": "Restore from backup",
	"strategy.restore.desc":  "Restore the full pre-install state from a backup manifest.",
	"strategy.restore.long":  "Restore rolls the machine back to the exact state captured in a pre-install backup manifest. Instead of undoing harness by harness, it reinstalls the full snapshot taken before your config was touched. It requires that manifest to be available. Choose it when you need a full rollback rather than a targeted removal; if you have no backup, use Targeted.",

	// backup.source.* — localized backup source labels.
	"backup.source.install": "install",
	"backup.source.sync":    "sync",
	"backup.source.upgrade": "upgrade",

	// memory.* — "active-stack memory branch" ([EN] Fase 4).
	"memory.usage":                      "Usage: active-stack memory branch [--project <path>] [--json] [--lang en|es]\n       active-stack memory hook session-start|pre-read  (agent hook: event JSON on stdin)\n       active-stack memory init [--project <path>] [--name <name>] [--yes] [--json] [--lang en|es]\n       active-stack memory proxy [--project <path>] -- <engram command> [args...]  (MCP proxy that tags saves)",
	"memory.proxy.no_command":           "memory proxy needs the Engram server command after \"--\" (for example: active-stack memory proxy -- engram mcp).",
	"memory.label.mode":                 "Memory mode",
	"memory.label.tagged":               "Tagged",
	"memory.label.branch":               "Branch",
	"memory.label.slug":                 "Slug",
	"memory.value.yes":                  "yes",
	"memory.value.no":                   "no",
	"memory.value.no_repo":              "(not a Git repository)",
	"memory.value.none":                 "(none)",
	"memory.warn.marker_invalid":        "Ignoring %s: it is not valid JSON. Memory mode stays off.",
	"memory.warn.marker_unknown_policy": "Ignoring %s: unknown memory_policy %q. Memory mode stays off.",
	"memory.error.unknown_subcommand":   "Unknown memory subcommand %q.",
	"memory.error.no_directory":         "The directory does not exist: %s",
	"memory.error.git_missing":          "Git is not installed or not in PATH.",
	"memory.error.empty_slug":           "Branch %q has no letters or digits to build its tag. Rename the branch.",
	"memory.error.failed":               "Could not read the Git state: %v",
	"memory.hook.unknown":               "Unknown memory hook %q. Available: session-start, pre-read. Nothing was changed.",
	"memory.hook.skipped":               "Memory hook %s skipped (%v). The agent continues without it.",
	"memory.init.project":               "Project: %s",
	"memory.init.name":                  "Project name: %s (%s)",
	"memory.init.source.flag":           "from --name",
	"memory.init.source.remote":         "from the origin remote",
	"memory.init.source.directory":      "from the folder of the main checkout",
	"memory.init.file":                  "  %s: %s",
	"memory.init.action.created":        "created",
	"memory.init.action.updated":        "updated",
	"memory.init.action.unchanged":      "unchanged",
	"memory.init.backup":                "Backup of the previous files: %s",
	"memory.init.warn.dirty":            "Warning: %s already had uncommitted changes of yours. The proposed commit will include them.",
	"memory.init.warn.tracked":          "Warning: other files under .engram/ are already versioned (%s). The .gitignore rules do not un-version them.",
	"memory.init.commit.header":         "Proposed commit (NOT executed: run it yourself, or ask for it). Run the commands from %s:",
	"memory.init.commit.none":           "Nothing to commit: the files are already versioned.",
	"memory.init.prompt":                "Use %q as the project name? [y/N] ",
	"memory.init.error.not_confirmed":   "The project name would be %q (%s). Confirm it with --yes, or choose another with --name <name>. Nothing was changed.",
	"memory.init.error.declined":        "Cancelled: the project name was not confirmed. Nothing was changed.",
	"memory.init.error.conflict_name":   "Stopped: %s already has project_name %q, and active-stack would write %q. Nothing was changed. If %[2]q is the right name, run again with --name %[2]s; otherwise edit the file by hand.",
	"memory.init.error.conflict_policy": "Stopped: %s has %s = %q, and active-stack would write %q. Nothing was changed. Edit the file by hand and run again.",
	"memory.init.error.ignored_dir":     "Stopped: .gitignore has the rule %q, which ignores the whole .engram/ folder, and Git cannot re-include files inside an ignored folder. Narrow that rule (for example to .engram/*) and run again. Nothing was changed.",
	"memory.init.error.still_ignored":   "Stopped: after writing, Git still ignores %s because of another ignore rule (for example a global ignore file, or an earlier line of your .gitignore such as .*). Everything was put back. Remove that rule and run again.",
	"memory.init.error.not_ignored":     "Stopped: after writing, Git would not ignore the rest of .engram/ (%s). Everything was put back.",
	"memory.init.error.invalid_file":    "%s cannot be used: %v. Fix it and run again. Nothing was changed.",
	"memory.init.error.home":            "Refusing to write in your home directory (%s). Run it inside a project repository.",
	"memory.init.error.not_git_repo":    "That directory is not inside a Git repository. Nothing was changed.",
	"memory.init.error.empty_name":      "Could not build a project name from the repository. Pass one with --name <name>.",
	"memory.init.error.failed":          "memory init failed: %v",

	// verify.cloud.* — engram-cloud health checks (R-21).
	"verify.cloud.hooks.desc":       "Engram Cloud memory hooks installed for %s",
	"verify.cloud.hooks.missing":    "memory hooks missing for %s in %q (%s): reinstall the engram-cloud harness",
	"verify.cloud.path.desc":        "active-stack is on PATH so the %s hooks can run",
	"verify.cloud.path.missing":     "active-stack is not on PATH: the %s hooks are installed but cannot run until it is",
	"verify.cloud.trust.desc":       "Codex hook trust (cannot be checked automatically)",
	"verify.cloud.trust.notice":     "Codex: open /hooks and review and trust the active-stack hook; Codex does not run it until you do",
	"verify.cloud.proxy.desc":       "Engram MCP entry runs through the active-stack save proxy for %s",
	"verify.cloud.proxy.missing":    "the Engram MCP entry of %s in %q does not run through the active-stack proxy: reinstall the engram-cloud harness",
	"verify.cloud.proxy.noentry":    "no Engram MCP entry found for %s in %q: the save proxy was not registered",
	"verify.cloud.autosync.desc":    "ENGRAM_CLOUD_AUTOSYNC enabled in the Engram MCP environment for %s",
	"verify.cloud.autosync.off":     "ENGRAM_CLOUD_AUTOSYNC is not set to 1 in the Engram MCP entry of %s (%q): continuous sync is off",
	"verify.cloud.unsupported.desc": "%s is outside the engram-cloud scope",
	"verify.cloud.unsupported.warn": "%s does not support engram-cloud and this project uses the mode: memories saved from %s will not be labeled by branch",
	"verify.cloud.project.desc":     "project files .engram/config.json and .engram/active-stack.json are valid",
	"verify.cloud.project.invalid":  "%s is not valid: %s",
	"verify.cloud.status_error":     "could not read the state: %v",
}

func init() {
	registerTable(LangEN, tableEN)
}
