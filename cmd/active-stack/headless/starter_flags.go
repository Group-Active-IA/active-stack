package headless

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Group-Active-IA/active-stack/internal/i18n"
	"github.com/Group-Active-IA/active-stack/internal/model"
)

// ParsedStarterAddFlags is the result of parsing the "starter add" sub-command
// flags (C-29 D1/D2).
type ParsedStarterAddFlags struct {
	// StarterID is the positional required first argument.
	StarterID string

	// ProjectPath is the resolved project root (from --project or cwd default).
	// It is NOT absolutized or validated here; the handler does that.
	ProjectPath string

	// DryRun, when true, means print the plan and exit without executing.
	DryRun bool

	// Yes means "confirm without prompting".
	Yes bool

	// Agents is the list of targeted agents. When --agent is omitted it
	// defaults to the P0 focal agents [claude, opencode] (D1 TBD resolved:
	// default to "all focal registered" = claude+opencode).
	Agents []model.Agent

	// Lang selects the language for every user-facing string emitted by the
	// "windows starters install" event stream (i18n-engine-locales D3). Zero
	// value is treated as i18n.Default (EN).
	Lang i18n.Lang
}

// defaultFocalAgents are the P0 agents targeted by "starter add" when --agent
// is omitted. Decision made in RED (Task 2.3): default to all focal registered
// agents (claude + opencode) rather than project-detected agents, because
// detection would require reading project files at parse time — which violates
// the separation between flag parsing (pure) and validation/execution (effectful).
var defaultFocalAgents = []model.Agent{model.AgentClaude, model.AgentOpenCode}

// ParseStarterAddFlags parses the raw argument list for the "starter add"
// sub-command. It follows the same style as ParseInstallFlags (flag.FlagSet,
// ContinueOnError, stderr controlled).
//
// Rules (from design D1/D2):
//   - First non-flag argument is the required <starter-id>. It can appear
//     before or after the flags (the id is extracted first, then flags are
//     parsed from the remaining args).
//   - --project <path>: target project root; defaults to cwd when omitted
//     (absolutized by the handler, not here).
//   - --dry-run: print plan steps without executing.
//   - --yes / -y: confirm without prompt.
//   - --agent <csv>: targeted agents; defaults to claude,opencode.
//
// Returns an error for: missing id, id that starts with "-", unknown flags.
func ParseStarterAddFlags(args []string) (ParsedStarterAddFlags, error) {
	// Extract the starter id: the first argument that does not start with "-".
	// This lets the id appear before or after flags without requiring a "--"
	// separator, matching the ergonomic expectation of a positional-first CLI.
	var starterID string
	var remaining []string
	for i, a := range args {
		if !strings.HasPrefix(a, "-") && starterID == "" {
			starterID = a
			remaining = append(remaining, args[i+1:]...)
			break
		}
		remaining = append(remaining, a)
	}

	if starterID == "" {
		if len(args) > 0 && strings.HasPrefix(args[0], "-") {
			return ParsedStarterAddFlags{}, fmt.Errorf("starter add: invalid starter id %q (looks like a flag)", args[0])
		}
		return ParsedStarterAddFlags{}, fmt.Errorf("starter add: missing required <starter-id>")
	}

	// Parse flags from remaining args.
	fs := flag.NewFlagSet("starter add", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	var (
		project string
		dryRun  bool
		yes     bool
		yShort  bool
		agent   string
	)

	fs.StringVar(&project, "project", "", "project root directory (default: current working directory)")
	fs.BoolVar(&dryRun, "dry-run", false, "print plan steps; do not execute")
	fs.BoolVar(&yes, "yes", false, "confirm without prompt")
	fs.BoolVar(&yShort, "y", false, "alias for --yes")
	fs.StringVar(&agent, "agent", "", "comma-separated list of agents (default: claude,opencode)")

	if err := fs.Parse(remaining); err != nil {
		return ParsedStarterAddFlags{}, err
	}

	if yShort {
		yes = true
	}

	// Resolve project path: use --project value or fall back to cwd.
	projectPath := project
	if projectPath == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return ParsedStarterAddFlags{}, fmt.Errorf("starter add: resolve cwd: %w", err)
		}
		projectPath = cwd
	}

	// Parse agent list; default to focal agents when omitted.
	agents := parseAgentCSV(agent)
	if len(agents) == 0 {
		agents = append([]model.Agent(nil), defaultFocalAgents...)
	}

	return ParsedStarterAddFlags{
		StarterID:   starterID,
		ProjectPath: projectPath,
		DryRun:      dryRun,
		Yes:         yes,
		Agents:      agents,
	}, nil
}

// ResolveProjectRoot resolves and validates the project root path for
// "starter add" (design D2). Rules:
//   - Absolutizes the path via filepath.Abs (handles relative paths and ".").
//   - Fails with a clear error when the resolved path does not exist.
//   - NEVER creates the directory. The user must supply an existing path.
//
// The no-marker-required decision (D2 TBD resolved in RED): we accept any
// existing directory — no ".git" check. Rationale: requiring a project marker
// would block legitimate use cases (new projects, monorepo sub-dirs) and the
// backup+rollback system already protects against accidental writes.
func ResolveProjectRoot(raw string) (string, error) {
	abs, err := filepath.Abs(raw)
	if err != nil {
		return "", fmt.Errorf("starter add: absolutize project path %q: %w", raw, err)
	}
	info, err := os.Stat(abs)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("starter add: project path %q does not exist (command will not create it)", abs)
		}
		return "", fmt.Errorf("starter add: stat project path %q: %w", abs, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("starter add: project path %q is not a directory", abs)
	}
	return abs, nil
}

// parseAgentCSV splits a comma-separated agent string into a slice of Agent
// values. Empty string returns nil. Trims spaces around each entry.
// Shared helper for ParseInstallFlags and ParseStarterAddFlags.
func parseAgentCSV(raw string) []model.Agent {
	if raw == "" {
		return nil
	}
	var agents []model.Agent
	for _, s := range strings.Split(raw, ",") {
		s = strings.TrimSpace(s)
		if s != "" {
			agents = append(agents, model.Agent(s))
		}
	}
	return agents
}
