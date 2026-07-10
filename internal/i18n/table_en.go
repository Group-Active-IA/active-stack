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
	"component.fallback.security":     "Basic protection for safer setup. This is always installed.",
	"component.fallback.external":     "Downloads and configures an external tool.",
	"component.fallback.config":       "Applies recommended configuration.",
	"component.fallback.skill":        "Adds guided workflow helpers.",
	"component.fallback.generic_fmt":  "Installs %s.",

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
	"tier.estricto.label": "Strict",
	"tier.estricto.desc":  "Agent must ask for every operation. Highest friction, highest security.",
	"tier.estricto.long":  "Strict configures the agent's permissions to ask for confirmation before every operation. It applies the most restrictive rule set on top of the security floor: nothing runs without your sign-off. It is the highest-friction, highest-security level. Choose it for sensitive projects or when you're starting out and want to review everything before it happens; if it slows you down, Balanced loosens the repetitive parts.",

	"tier.balanceado.label": "Balanced",
	"tier.balanceado.desc":  "Curated allow-list for safe, repetitive operations. Recommended starting point.",
	"tier.balanceado.long":  "Balanced configures a curated allow-list so safe, repetitive operations run without interruption while riskier actions still ask for confirmation. It is the middle ground between friction and autonomy on top of the security floor. It is the recommended starting point for most projects. Choose it when you want flow without giving up the important controls; move to Strict if you need to review more, or to Bypass for less friction.",

	"tier.bypass.label":   "Bypass",
	"tier.bypass.desc":    "Full autonomy opt-in. The security floor deny-list still applies.",
	"tier.bypass.long":    "Bypass grants the agent full autonomy to act without asking. The security floor deny-list still blocks the most dangerous operations—that layer never turns off—but everything else runs unattended. Choose it only when you fully trust the flow and friction is holding you back more than protecting you. Use it with caution: it is the lowest-control level.",
	"tier.bypass.warning": "Bypass: autonomous mode — the security floor still applies (C-21)",

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
}

func init() {
	registerTable(LangEN, tableEN)
}
