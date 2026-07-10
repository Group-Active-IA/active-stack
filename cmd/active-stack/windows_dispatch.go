package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/Group-Active-IA/active-stack/assets"
	"github.com/Group-Active-IA/active-stack/cmd/active-stack/headless"
	"github.com/Group-Active-IA/active-stack/internal/i18n"
	"github.com/Group-Active-IA/active-stack/internal/install"
	"github.com/Group-Active-IA/active-stack/internal/model"
	"github.com/Group-Active-IA/active-stack/internal/uninstall"
	"github.com/Group-Active-IA/active-stack/internal/verify"
)

// registerLangFlag registers --lang on fs (default "en", shown in --help)
// and returns a pointer to the raw parsed string. Callers register every
// other flag first, call fs.Parse once, then resolve the pointer's value via
// resolveLangFlag. Shared by every windows subcommand so the --lang wiring
// cannot drift (task 7.3 REFACTOR).
func registerLangFlag(fs *flag.FlagSet) *string {
	lang := "en"
	fs.StringVar(&lang, "lang", "en", "output language: en|es")
	return &lang
}

// resolveLangFlag parses the raw string captured by registerLangFlag into an
// i18n.Lang. Returns a non-nil error naming the invalid value; callers must
// print it to w and return a non-zero exit code before any subcommand work
// runs (D3, windows-install-contract / windows-hub-contract specs).
func resolveLangFlag(raw *string) (i18n.Lang, error) {
	return i18n.Parse(*raw)
}

func runWindowsDispatch(args []string, cat install.Catalog, reg install.Registry, w io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(w, "usage: active-stack windows <detect|options|install|uninstall|starters|backups|uninstall-options>")
		return 1
	}

	switch args[0] {
	case "detect":
		fs := flag.NewFlagSet("windows detect", flag.ContinueOnError)
		fs.SetOutput(w)
		homeDir := ""
		fs.StringVar(&homeDir, "home", "", "override home directory")
		if err := fs.Parse(args[1:]); err != nil {
			return 1
		}
		if homeDir == "" {
			var err error
			homeDir, err = os.UserHomeDir()
			if err != nil {
				fmt.Fprintf(w, "error: resolve home dir: %v\n", err)
				return 1
			}
		}
		if err := headless.RunWindowsDetect(homeDir, reg, w); err != nil {
			fmt.Fprintf(w, "error: %v\n", err)
			return 1
		}
		return 0

	case "options":
		fs := flag.NewFlagSet("windows options", flag.ContinueOnError)
		fs.SetOutput(w)
		var agent string
		fs.StringVar(&agent, "agent", "", "comma-separated list of target agents (e.g. claude,opencode)")
		langRaw := registerLangFlag(fs)
		if err := fs.Parse(args[1:]); err != nil {
			return 1
		}
		lang, err := resolveLangFlag(langRaw)
		if err != nil {
			fmt.Fprintf(w, "error: %v\n", err)
			return 1
		}
		var agents []model.Agent
		for _, raw := range strings.Split(agent, ",") {
			raw = strings.TrimSpace(raw)
			if raw == "" {
				continue
			}
			agents = append(agents, model.Agent(raw))
		}
		if len(agents) == 0 {
			fmt.Fprintln(w, "error: windows options requires --agent")
			return 1
		}
		if err := headless.RunWindowsOptions(cat, agents, lang, w); err != nil {
			fmt.Fprintf(w, "error: %v\n", err)
			return 1
		}
		return 0

	case "install":
		parsed, err := headless.ParseInstallFlags(args[1:])
		if err != nil {
			fmt.Fprintf(w, "error: %v\n", err)
			return 1
		}
		if parsed.TUI {
			fmt.Fprintln(w, "error: windows install requires headless intent such as --mode or --agent")
			return 1
		}

		parsed.HomeDir = resolveHomeDir(parsed.HomeDir, nil)
		if parsed.BinaryPath == "" {
			if binPath, err := os.Executable(); err == nil {
				parsed.BinaryPath = binPath
			}
		}
		if parsed.VerifyHookFn == nil {
			verifyBase := parsed.HomeDir
			if parsed.Target == model.Project {
				verifyBase = parsed.ProjectRoot
			}
			if registry, ok := reg.(agentRegistryAdapter); ok && registry.r != nil {
				verifyAdapters := resolveVerifyAdaptersForTarget(parsed.Intent.Agents, registry.r, verifyBase, parsed.Target)
				selectedHarnesses := collectSelectedHarnesses(cat, parsed.Intent)
				parsed.VerifyHookFn = verify.BuildHook(selectedHarnesses, verifyAdapters, verifyBase)
			}
		}
		parsed.BuildPlanFn = func(c install.Catalog, intent install.Intent, opts install.Options) (install.Plan, error) {
			opts = install.WithEmbeddedSkillsFS(opts, assets.SkillsFS)
			return install.BuildPlan(c, intent, opts)
		}
		return headless.RunWindowsInstall(parsed, cat, reg, w)

	case "uninstall":
		parsed, err := headless.ParseUninstallFlags(args[1:])
		if err != nil {
			fmt.Fprintf(w, "error: %v\n", err)
			return 1
		}

		var uninstallReg uninstall.Registry
		switch typed := any(reg).(type) {
		case uninstall.Registry:
			uninstallReg = typed
		case agentRegistryAdapter:
			uninstallReg = uninstallRegistryAdapter{r: typed.r}
		default:
			fmt.Fprintln(w, "error: windows uninstall requires an uninstall-capable registry")
			return 1
		}

		return headless.RunWindowsUninstall(parsed, cat, uninstallReg, w)

	case "starters":
		return runWindowsStartersDispatch(args[1:], cat, reg, w)

	case "backups":
		return runWindowsBackupsDispatch(args[1:], w)

	case "uninstall-options":
		fs := flag.NewFlagSet("windows uninstall-options", flag.ContinueOnError)
		fs.SetOutput(w)
		homeDir := ""
		fs.StringVar(&homeDir, "home", "", "override home directory")
		langRaw := registerLangFlag(fs)
		if err := fs.Parse(args[1:]); err != nil {
			return 1
		}
		lang, err := resolveLangFlag(langRaw)
		if err != nil {
			fmt.Fprintf(w, "error: %v\n", err)
			return 1
		}
		if homeDir == "" {
			var err error
			homeDir, err = os.UserHomeDir()
			if err != nil {
				fmt.Fprintf(w, "error: resolve home dir: %v\n", err)
				return 1
			}
		}
		if err := headless.RunWindowsUninstallOptions(homeDir, lang, w); err != nil {
			fmt.Fprintf(w, "error: %v\n", err)
			return 1
		}
		return 0

	default:
		fmt.Fprintf(w, "unknown windows subcommand: %q\n", args[0])
		return 1
	}
}

// runWindowsStartersDispatch routes "windows starters <list|install>"
// (windows-contract-hub-operations, design D7). args is the subcommand
// argument list with "starters" already stripped.
func runWindowsStartersDispatch(args []string, cat install.Catalog, reg install.Registry, w io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(w, "usage: active-stack windows starters <list|install>")
		return 1
	}

	sc, ok := cat.(headless.StarterCatalog)
	if !ok {
		fmt.Fprintln(w, "error: windows starters requires a starter-capable catalog")
		return 1
	}

	switch args[0] {
	case "list":
		fs := flag.NewFlagSet("windows starters list", flag.ContinueOnError)
		fs.SetOutput(w)
		langRaw := registerLangFlag(fs)
		if err := fs.Parse(args[1:]); err != nil {
			return 1
		}
		lang, err := resolveLangFlag(langRaw)
		if err != nil {
			fmt.Fprintf(w, "error: %v\n", err)
			return 1
		}
		if err := headless.RunWindowsStartersList(sc, lang, w); err != nil {
			fmt.Fprintf(w, "error: %v\n", err)
			return 1
		}
		return 0

	case "install":
		fs := flag.NewFlagSet("windows starters install", flag.ContinueOnError)
		fs.SetOutput(w)
		var starterID, project, agent string
		var dryRun, yes bool
		fs.StringVar(&starterID, "starter", "", "starter id to install")
		fs.StringVar(&project, "project", "", "target project root")
		fs.StringVar(&agent, "agent", "", "comma-separated list of agents (e.g. claude,opencode)")
		fs.BoolVar(&dryRun, "dry-run", false, "print plan steps; do not execute")
		fs.BoolVar(&yes, "yes", false, "confirm without prompt")
		langRaw := registerLangFlag(fs)
		if err := fs.Parse(args[1:]); err != nil {
			return 1
		}
		lang, err := resolveLangFlag(langRaw)
		if err != nil {
			fmt.Fprintf(w, "error: %v\n", err)
			return 1
		}
		if starterID == "" {
			fmt.Fprintln(w, "error: windows starters install requires --starter")
			return 1
		}
		if project == "" {
			fmt.Fprintln(w, "error: windows starters install requires --project")
			return 1
		}

		var agents []model.Agent
		for _, raw := range strings.Split(agent, ",") {
			raw = strings.TrimSpace(raw)
			if raw == "" {
				continue
			}
			agents = append(agents, model.Agent(raw))
		}
		if len(agents) == 0 {
			fmt.Fprintln(w, "error: windows starters install requires --agent")
			return 1
		}

		flags := headless.ParsedStarterAddFlags{
			StarterID:   starterID,
			ProjectPath: project,
			DryRun:      dryRun,
			Yes:         yes,
			Agents:      agents,
			Lang:        lang,
		}
		buildPlanFn := func(c install.Catalog, intent install.Intent, opts install.Options) (install.Plan, error) {
			opts = install.WithEmbeddedSkillsFS(opts, assets.SkillsFS)
			return install.BuildPlan(c, intent, opts)
		}
		return headless.RunWindowsStartersInstall(flags, sc, reg, buildPlanFn, w)

	default:
		fmt.Fprintf(w, "unknown windows starters subcommand: %q\n", args[0])
		return 1
	}
}

// runWindowsBackupsDispatch routes "windows backups <list|restore|delete|rename>"
// (windows-contract-hub-operations, design D7). args is the subcommand
// argument list with "backups" already stripped.
func runWindowsBackupsDispatch(args []string, w io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(w, "usage: active-stack windows backups <list|restore|delete|rename>")
		return 1
	}

	switch args[0] {
	case "list":
		fs := flag.NewFlagSet("windows backups list", flag.ContinueOnError)
		fs.SetOutput(w)
		homeDir := ""
		fs.StringVar(&homeDir, "home", "", "override home directory")
		langRaw := registerLangFlag(fs)
		if err := fs.Parse(args[1:]); err != nil {
			return 1
		}
		lang, err := resolveLangFlag(langRaw)
		if err != nil {
			fmt.Fprintf(w, "error: %v\n", err)
			return 1
		}
		if homeDir == "" {
			var err error
			homeDir, err = os.UserHomeDir()
			if err != nil {
				fmt.Fprintf(w, "error: resolve home dir: %v\n", err)
				return 1
			}
		}
		if err := headless.RunWindowsBackupsList(homeDir, lang, w); err != nil {
			fmt.Fprintf(w, "error: %v\n", err)
			return 1
		}
		return 0

	case "restore", "delete", "rename":
		action := args[0]
		fs := flag.NewFlagSet("windows backups "+action, flag.ContinueOnError)
		fs.SetOutput(w)
		homeDir := ""
		id := ""
		description := ""
		fs.StringVar(&homeDir, "home", "", "override home directory")
		fs.StringVar(&id, "id", "", "backup id")
		fs.StringVar(&description, "description", "", "new description (rename only)")
		langRaw := registerLangFlag(fs)
		if err := fs.Parse(args[1:]); err != nil {
			return 1
		}
		// The backup-action response is machine-oriented (success/message/id);
		// the language is validated at the flag boundary per the hub contract
		// even though RunWindowsBackupsAction emits no localized strings yet.
		if _, err := resolveLangFlag(langRaw); err != nil {
			fmt.Fprintf(w, "error: %v\n", err)
			return 1
		}
		if id == "" {
			fmt.Fprintf(w, "error: windows backups %s requires --id\n", action)
			return 1
		}
		if homeDir == "" {
			var err error
			homeDir, err = os.UserHomeDir()
			if err != nil {
				fmt.Fprintf(w, "error: resolve home dir: %v\n", err)
				return 1
			}
		}
		return headless.RunWindowsBackupsAction(homeDir, action, id, description, w)

	default:
		fmt.Fprintf(w, "unknown windows backups subcommand: %q\n", args[0])
		return 1
	}
}
