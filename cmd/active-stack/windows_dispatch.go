package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/Group-Active-IA/active-stack/assets"
	"github.com/Group-Active-IA/active-stack/cmd/active-stack/headless"
	"github.com/Group-Active-IA/active-stack/internal/install"
	"github.com/Group-Active-IA/active-stack/internal/model"
	"github.com/Group-Active-IA/active-stack/internal/uninstall"
	"github.com/Group-Active-IA/active-stack/internal/verify"
)

func runWindowsDispatch(args []string, cat install.Catalog, reg install.Registry, w io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(w, "usage: active-stack windows <detect|options|install|uninstall>")
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
		fs.StringVar(&agent, "agent", "", "target agent")
		if err := fs.Parse(args[1:]); err != nil {
			return 1
		}
		if agent == "" {
			fmt.Fprintln(w, "error: windows options requires --agent")
			return 1
		}
		if err := headless.RunWindowsOptions(cat, model.Agent(agent), w); err != nil {
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

	default:
		fmt.Fprintf(w, "unknown windows subcommand: %q\n", args[0])
		return 1
	}
}
