package headless

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/JuanCruzRobledo/jr-stack/internal/model"
	"github.com/JuanCruzRobledo/jr-stack/internal/uninstall"
)

// ParsedUninstallFlags is the result of parsing the uninstall sub-command flags.
// It is a slimmer sibling of ParsedFlags — no TUI field, no install-only fields
// (D3: separate type keeps the uninstall surface clear).
type ParsedUninstallFlags struct {
	// Intent is the resolved uninstall intent (agents, mode, custom harnesses, strategy).
	Intent uninstall.Intent

	// HomeDir is the resolved home directory (from --home or os.UserHomeDir()).
	HomeDir string

	// DryRun, when true, means the caller should print the plan steps and exit
	// without executing anything.
	DryRun bool

	// Yes means "confirm without prompting" (--yes / -y).
	Yes bool

	// RestoreManifestPath is the path supplied via --restore-manifest.
	// Only meaningful when Intent.Strategy == StrategyRestore.
	RestoreManifestPath string
}

// ParseUninstallFlags parses the raw argument list for the "uninstall" sub-command.
//
// Rules (from design D4 and spec):
//   - --mode must be "lite", "full", or "custom" when provided.
//   - --custom requires --mode custom.
//   - --strategy must be "targeted" or "restore"; defaults to "targeted" (D5).
//   - --home defaults to os.UserHomeDir().
//   - No --project flag (uninstall is machine-scope, D7/proposal).
//
// The returned error is non-nil for validation failures (invalid mode,
// --custom without --mode custom, invalid strategy). Callers must write the
// error to stderr and exit non-zero.
func ParseUninstallFlags(args []string) (ParsedUninstallFlags, error) {
	fs := flag.NewFlagSet("uninstall", flag.ContinueOnError)
	// Suppress the default output so callers control stderr.
	fs.SetOutput(os.Stderr)

	var (
		mode             string
		agent            string
		custom           string
		strategy         string
		restoreManifest  string
		dryRun           bool
		yes              bool
		yShort           bool
		home             string
	)

	fs.StringVar(&mode, "mode", "", "uninstall mode: lite|full|custom")
	fs.StringVar(&agent, "agent", "", "comma-separated list of agents (e.g. claude,opencode)")
	fs.StringVar(&custom, "custom", "", "comma-separated harness IDs (only with --mode custom)")
	fs.StringVar(&strategy, "strategy", "targeted", "reversal strategy: targeted|restore (default: targeted)")
	fs.StringVar(&restoreManifest, "restore-manifest", "", "path to install-time backup manifest (required with --strategy restore)")
	fs.BoolVar(&dryRun, "dry-run", false, "print plan steps; do not execute")
	fs.BoolVar(&yes, "yes", false, "confirm without prompt")
	fs.BoolVar(&yShort, "y", false, "alias for --yes")
	fs.StringVar(&home, "home", "", "override home directory (default: os.UserHomeDir())")

	if err := fs.Parse(args); err != nil {
		return ParsedUninstallFlags{}, err
	}

	// Merge -y into yes.
	if yShort {
		yes = true
	}

	// ── Validate strategy ──────────────────────────────────────────────────
	var uninstallStrategy uninstall.Strategy
	switch strategy {
	case "targeted", "":
		uninstallStrategy = uninstall.StrategyTargeted
	case "restore":
		uninstallStrategy = uninstall.StrategyRestore
	default:
		return ParsedUninstallFlags{}, fmt.Errorf("invalid --strategy %q: must be targeted or restore", strategy)
	}

	// ── Validate mode ──────────────────────────────────────────────────────
	var installMode model.InstallMode
	if mode != "" {
		switch mode {
		case "lite":
			installMode = model.ModeLite
		case "full":
			installMode = model.ModeFull
		case "custom":
			installMode = model.ModeCustom
		default:
			return ParsedUninstallFlags{}, fmt.Errorf("invalid --mode %q: must be lite, full, or custom", mode)
		}
	}

	// ── Validate --custom ──────────────────────────────────────────────────
	if custom != "" && installMode != model.ModeCustom {
		return ParsedUninstallFlags{}, fmt.Errorf("--custom requires --mode custom (got --mode %q)", mode)
	}

	// ── Parse agents ───────────────────────────────────────────────────────
	var agents []model.Agent
	if agent != "" {
		for _, raw := range strings.Split(agent, ",") {
			raw = strings.TrimSpace(raw)
			if raw == "" {
				continue
			}
			agents = append(agents, model.Agent(raw))
		}
	}

	// ── Parse custom harness IDs ───────────────────────────────────────────
	var customIDs []string
	if custom != "" {
		for _, raw := range strings.Split(custom, ",") {
			raw = strings.TrimSpace(raw)
			if raw != "" {
				customIDs = append(customIDs, raw)
			}
		}
	}

	// ── Resolve home dir ───────────────────────────────────────────────────
	homeDir := home
	if homeDir == "" {
		var err error
		homeDir, err = os.UserHomeDir()
		if err != nil {
			return ParsedUninstallFlags{}, fmt.Errorf("resolve home dir: %w", err)
		}
	}

	return ParsedUninstallFlags{
		DryRun:              dryRun,
		Yes:                 yes,
		HomeDir:             homeDir,
		RestoreManifestPath: restoreManifest,
		Intent: uninstall.Intent{
			Mode:     installMode,
			Agents:   agents,
			Custom:   customIDs,
			Strategy: uninstallStrategy,
		},
	}, nil
}
