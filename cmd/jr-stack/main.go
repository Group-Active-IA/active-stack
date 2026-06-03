// Command jr-stack is the JR Stack installer — a methodology-first harness
// installer for AI coding agents.
//
// Usage:
//
//	jr-stack install              Launch the interactive TUI install flow.
//	jr-stack install --headless   Non-interactive install (also implied by --mode/--agent).
//	jr-stack install --dry-run    Print the install plan; do not execute.
//	jr-stack install --help       Show all available flags.
package main

import (
	"fmt"
	"os"

	"github.com/JuanCruzRobledo/jr-stack/assets"
	"github.com/JuanCruzRobledo/jr-stack/cmd/jr-stack/headless"
	"github.com/JuanCruzRobledo/jr-stack/internal/agents"
	"github.com/JuanCruzRobledo/jr-stack/internal/catalog"
	"github.com/JuanCruzRobledo/jr-stack/internal/install"
	"github.com/JuanCruzRobledo/jr-stack/internal/model"
	"github.com/JuanCruzRobledo/jr-stack/internal/tui"
	"github.com/JuanCruzRobledo/jr-stack/internal/verify"
)

func main() {
	// Sin argumentos (caso típico: doble-click del .exe en Windows) → lanzar el
	// flujo interactivo. JR Stack es methodology-first para usuario final: el
	// doble-click debe ABRIR la TUI, no imprimir un usage y cerrarse. runInstall
	// con args vacíos resuelve a TUI == true vía ParseInstallFlags.
	if len(os.Args) < 2 {
		if err := runInstall(nil); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	switch os.Args[1] {
	case "install":
		if err := runInstall(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	case "starter":
		cat, err := catalog.Load()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: load catalog: %v\n", err)
			os.Exit(1)
		}
		reg, err := agents.NewDefaultRegistry()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: create agent registry: %v\n", err)
			os.Exit(1)
		}
		regWrapper := agentRegistryAdapter{r: reg}
		exitCode := runStarterDispatch(os.Args[2:], cat, regWrapper, os.Stderr)
		if exitCode != 0 {
			os.Exit(exitCode)
		}
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %q\n", os.Args[1])
		os.Exit(1)
	}
}

// agentRegistryAdapter wraps *agents.Registry to satisfy install.Registry.
// Both registry types have the same Get signature except for the return type
// name (agents.Adapter vs install.AgentAdapter), and agents.Adapter is a
// structural superset of install.AgentAdapter, so the wrapper is just a cast.
type agentRegistryAdapter struct{ r *agents.Registry }

func (a agentRegistryAdapter) Get(agent model.Agent) (install.AgentAdapter, bool) {
	adapter, ok := a.r.Get(agent)
	if !ok {
		return nil, false
	}
	return adapter, true
}

func runInstall(args []string) error {
	// Parse flags to determine TUI vs headless mode.
	parsed, err := headless.ParseInstallFlags(args)
	if err != nil {
		return err
	}

	// 1. Load the embedded catalog (needed for both TUI and headless).
	cat, err := catalog.Load()
	if err != nil {
		return fmt.Errorf("load catalog: %w", err)
	}

	// 2. Build the default agent registry (P0: claude + opencode).
	reg, err := agents.NewDefaultRegistry()
	if err != nil {
		return fmt.Errorf("create agent registry: %w", err)
	}

	// Wrap the registry to satisfy install.Registry.
	regWrapper := agentRegistryAdapter{r: reg}

	// ── Headless mode ──────────────────────────────────────────────────────
	if !parsed.TUI {
		// Use the home dir from the parsed flags (may have been --home overridden).
		parsed.HomeDir = resolveHomeDir(parsed.HomeDir, reg)

		// Wire the verify hook (same logic as the TUI BuildPlanFn below).
		if parsed.VerifyHookFn == nil {
			verifyAdapters := resolveVerifyAdapters(parsed.Intent.Agents, reg)
			selectedHarnesses := collectSelectedHarnesses(cat, parsed.Intent)
			parsed.VerifyHookFn = verify.BuildHook(selectedHarnesses, verifyAdapters, parsed.HomeDir)
		}

		// Wire the embedded skills FS into BuildPlan via BuildPlanFn.
		// RunHeadless uses this function instead of calling install.BuildPlan directly,
		// so the FS is injected into opts before the plan is built.
		parsed.BuildPlanFn = func(c install.Catalog, intent install.Intent, opts install.Options) (install.Plan, error) {
			opts = install.WithEmbeddedSkillsFS(opts, assets.SkillsFS)
			return install.BuildPlan(c, intent, opts)
		}

		exitCode := headless.RunHeadless(parsed, cat, regWrapper, os.Stdout)
		if exitCode != 0 {
			os.Exit(exitCode)
		}
		return nil
	}

	// ── TUI (interactive) mode — no flags were passed ─────────────────────

	// 3. Resolve home directory.
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("resolve home dir: %w", err)
	}

	// 4. Detect installed agents; intersect with registered adapters.
	detectedAgents := tui.DetectInstalledAgents(homeDir)
	availableAgents := tui.AvailableAgentsList(detectedAgents, reg.SupportedAgents())

	// 5. Build the TUI deps with the embedded skills FS wired.
	deps := tui.ModelDeps{
		Catalog:         cat,
		Registry:        regWrapper,
		HomeDir:         homeDir,
		AvailableAgents: availableAgents,
		BuildPlanFn: func(c install.Catalog, intent install.Intent, opts install.Options) (install.Plan, error) {
			opts = install.WithEmbeddedSkillsFS(opts, assets.SkillsFS)

			// Wire the post-install verify hook.
			// Resolve adapters for the intent's agents; collect selected harnesses.
			// verify.Adapter is a structural subset of install.AgentAdapter — no cast needed.
			verifyAdapters := resolveVerifyAdapters(intent.Agents, reg)
			selectedHarnesses := collectSelectedHarnesses(c, intent)
			opts.VerifyHook = verify.BuildHook(selectedHarnesses, verifyAdapters, opts.HomeDir)

			return install.BuildPlan(c, intent, opts)
		},
	}

	// 6. Launch the TUI program.
	p := tui.NewProgram(deps)
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("tui: %w", err)
	}
	return nil
}

// resolveHomeDir returns homeDir if non-empty, otherwise falls back to
// os.UserHomeDir(). Any error resolving the system home is silently ignored
// (the caller already validated this path before reaching headless mode).
func resolveHomeDir(homeDir string, _ *agents.Registry) string {
	if homeDir != "" {
		return homeDir
	}
	h, _ := os.UserHomeDir()
	return h
}

// resolveVerifyAdapters resolves the concrete adapters for the given agents from
// the registry and narrows them to verify.Adapter via structural typing.
// agents.Adapter is a superset of verify.Adapter, so the assignment is valid.
func resolveVerifyAdapters(agentList []model.Agent, reg *agents.Registry) []verify.Adapter {
	out := make([]verify.Adapter, 0, len(agentList))
	for _, a := range agentList {
		adapter, ok := reg.Get(a)
		if !ok {
			continue
		}
		out = append(out, adapter)
	}
	return out
}

// collectSelectedHarnesses returns the harnesses selected for the given intent,
// used to wire the post-install verify hook. It DELEGATES to
// install.SelectHarnesses — the single source of truth for the security-first
// rule that forces install.SecurityFirstHarnessID in Custom mode (C-21/C-24).
// The previously duplicated forcing logic (and its own filterHarnessesByAgents)
// was removed.
//
// Intents reaching this point come from validated flags / the catalog, so
// SelectHarnesses should not error; if it does (defensive), we degrade to an
// empty set rather than verifying against an inconsistent selection.
func collectSelectedHarnesses(c install.Catalog, intent install.Intent) []model.Harness {
	harnesses, err := install.SelectHarnesses(c, intent)
	if err != nil {
		return nil
	}
	return harnesses
}
