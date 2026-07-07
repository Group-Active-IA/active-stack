// Package headless implements the non-interactive (headless) install mode for
// the active-stack binary. It is extracted into its own package so that flag
// parsing and execution logic can be unit-tested without running main().
package headless

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/Group-Active-IA/active-stack/internal/install"
	"github.com/Group-Active-IA/active-stack/internal/model"
	"github.com/Group-Active-IA/active-stack/internal/pipeline"
)

// ParsedFlags is the result of parsing the install sub-command flags.
// When TUI is true, no headless execution should happen; the caller must
// launch the interactive TUI instead.
type ParsedFlags struct {
	// TUI is true when no headless intent was expressed (no --headless,
	// no --mode, no --agent). The caller must launch the interactive TUI.
	TUI bool

	// Intent is the resolved install intent (agents, mode, custom harnesses).
	// Valid only when TUI == false.
	Intent install.Intent

	// HomeDir is the resolved home directory override.
	// When --home is not provided this is set to os.UserHomeDir().
	HomeDir string

	// DryRun, when true, means the caller should print the plan steps and
	// exit without executing anything.
	DryRun bool

	// Yes means "confirm without prompting" (--yes / -y).
	Yes bool

	// VerifyHookFn, when non-nil, overrides the default verify hook wired by
	// the executor. Used only by tests to inject a fake verify hook.
	VerifyHookFn func() error

	// BuildPlanFn, when non-nil, overrides the default install.BuildPlan call
	// inside RunHeadless. The entry point (main.go) uses this to inject
	// install.WithEmbeddedSkillsFS into opts before calling install.BuildPlan.
	// Tests leave it nil to use the default install.BuildPlan directly.
	BuildPlanFn func(cat install.Catalog, intent install.Intent, opts install.Options) (install.Plan, error)

	// ProgressEventFn, when non-nil, receives raw pipeline progress events from
	// RunHeadless. This is used by the Windows GUI contract to expose structured
	// step lifecycle events without parsing plaintext output.
	ProgressEventFn func(pipeline.ProgressEvent)
	// DownloadEventFn, when non-nil, receives structured download lifecycle
	// events emitted by external installers.
	DownloadEventFn func(install.DownloadEvent)

	// WindowsEventsFile, when non-empty, receives the same line-delimited JSON
	// stream as stdout for `active-stack windows install`. This lets a native
	// Windows bundle UI consume detailed progress while Burn owns the package
	// execution lifecycle.
	WindowsEventsFile string

	// ── C-29 starter add fields ───────────────────────────────────────────────

	// Target selects whether harness writes go to the machine home or to a
	// project root. Zero-value is model.Machine, preserving the pre-C-29
	// behavior for all existing install call sites.
	Target model.InstallTarget

	// ProjectRoot is the project directory used when Target == model.Project.
	// Ignored when Target is Machine (zero-value).
	ProjectRoot string

	// Starter is an optional starter whose MCPs should be written into the
	// project config. When nil, no MCP write steps are emitted.
	Starter *model.Starter

	// ── binary-self-install ───────────────────────────────────────────────────

	// NoSelfInstall, when true, skips the self-install step so the running
	// binary is NOT copied into the PATH bin dir. Default (false) = self-install ON.
	// Set by --no-self-install for CI / reproducible builds (D3).
	NoSelfInstall bool

	// BinaryPath is the resolved path of the running binary (os.Executable result),
	// forwarded to install.Options.SelfInstallBinaryPath. Set by main.go at entry,
	// empty means the self-install step resolves it via os.Executable internally.
	BinaryPath string
}

// ParseInstallFlags parses the raw argument list for the "install" sub-command.
//
// Rules (from design D4):
//   - No flags → TUI mode (ParsedFlags.TUI == true).
//   - --headless, --mode, or --agent → headless mode.
//   - --mode must be "lite", "full", or "custom".
//   - --custom requires --mode custom.
//   - --home defaults to os.UserHomeDir().
//
// The returned error is non-nil for validation failures (invalid mode, --custom
// without --mode custom). Callers must write the error to stderr and exit != 0.
func ParseInstallFlags(args []string) (ParsedFlags, error) {
	fs := flag.NewFlagSet("install", flag.ContinueOnError)
	// Suppress the default output so callers control stderr.
	fs.SetOutput(os.Stderr)

	var (
		headless      bool
		mode          string
		agent         string
		custom        string
		tier          string
		dryRun        bool
		yes           bool
		yShort        bool
		home          string
		noSelfInstall bool
		windowsEventsFile string
	)

	fs.BoolVar(&headless, "headless", false, "non-interactive install; implied by --mode or --agent")
	fs.StringVar(&mode, "mode", "", "install mode: lite|full|custom")
	fs.StringVar(&agent, "agent", "", "comma-separated list of agents (e.g. claude,opencode)")
	fs.StringVar(&custom, "custom", "", "comma-separated harness IDs (only with --mode custom)")
	fs.StringVar(&tier, "tier", "", "permission tier: estricto|balanceado|bypass")
	fs.BoolVar(&dryRun, "dry-run", false, "print plan steps; do not execute")
	fs.BoolVar(&yes, "yes", false, "confirm without prompt")
	fs.BoolVar(&yShort, "y", false, "alias for --yes")
	fs.StringVar(&home, "home", "", "override home directory (default: os.UserHomeDir())")
	fs.BoolVar(&noSelfInstall, "no-self-install", false, "skip copying the running binary into the PATH bin dir (CI / reproducible builds)")
	fs.StringVar(&windowsEventsFile, "windows-events-file", "", "internal: write windows install JSON events to this file")

	if err := fs.Parse(args); err != nil {
		return ParsedFlags{}, err
	}

	// Merge -y into yes.
	if yShort {
		yes = true
	}

	// Determine if headless mode is active.
	// --custom alone (without --mode) should also trigger headless so we can
	// validate and reject it rather than silently falling back to TUI.
	// --tier alone triggers headless the same way --custom does (D1, design.md).
	isHeadless := headless || mode != "" || agent != "" || custom != "" || tier != ""

	// No intent expressed → TUI.
	if !isHeadless {
		return ParsedFlags{TUI: true}, nil
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
			return ParsedFlags{}, fmt.Errorf("invalid --mode %q: must be lite, full, or custom", mode)
		}
	}

	// ── Validate --custom ──────────────────────────────────────────────────
	if custom != "" && installMode != model.ModeCustom {
		return ParsedFlags{}, fmt.Errorf("--custom requires --mode custom (got --mode %q)", mode)
	}

	// ── Validate --tier ─────────────────────────────────────────────────────
	// Empty stays at the zero value; the engine's PermissionTier.Normalize()
	// turns that into TierBalanceado downstream (D1, design.md).
	var permTier model.PermissionTier
	if tier != "" {
		switch tier {
		case "estricto":
			permTier = model.TierEstricto
		case "balanceado":
			permTier = model.TierBalanceado
		case "bypass":
			permTier = model.TierBypass
		default:
			return ParsedFlags{}, fmt.Errorf("invalid --tier %q: must be estricto, balanceado, or bypass", tier)
		}
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
			return ParsedFlags{}, fmt.Errorf("resolve home dir: %w", err)
		}
	}

	return ParsedFlags{
		TUI:           false,
		DryRun:        dryRun,
		Yes:           yes,
		HomeDir:       homeDir,
		NoSelfInstall: noSelfInstall,
		WindowsEventsFile: windowsEventsFile,
		Intent: install.Intent{
			Mode:   installMode,
			Agents: agents,
			Custom: customIDs,
			Tier:   permTier,
		},
	}, nil
}
