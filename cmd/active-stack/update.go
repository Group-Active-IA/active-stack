package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"

	"github.com/Group-Active-IA/active-stack/assets"
	"github.com/Group-Active-IA/active-stack/internal/agents"
	"github.com/Group-Active-IA/active-stack/internal/catalog"
	"github.com/Group-Active-IA/active-stack/internal/install"
	"github.com/Group-Active-IA/active-stack/internal/model"
	"github.com/Group-Active-IA/active-stack/internal/planner"
	"github.com/Group-Active-IA/active-stack/internal/update"
)

type updateFlags struct {
	Check       bool
	SkillsOnly  bool
	Sync        bool
	Finish      bool
	ParentPID   int
	Target      string
	AdoptLegacy bool
	CleanupBin  string
	Home        string
	Project     string
}

const updateUsage = `Usage: active-stack update [--check | --skills-only] [--project PATH] [--home PATH]

  --check        Check the latest published binary version without changes
  --skills-only  Refresh installed catalog skills using this binary's catalog
  --project PATH Include an older project with starter skills
  --adopt-legacy-skills  Replace matching skill names from installs without a receipt
  --home PATH    Override the user home directory
`

// buildVersion is stamped by release builds and build.bat. Unstamped source
// builds report "dev" and never replace themselves with a possibly older release.
var buildVersion = "dev"

func versionParts(tag string) ([]int, bool) {
	if !strings.HasPrefix(tag, "v") {
		return nil, false
	}
	parts := strings.Split(strings.TrimPrefix(tag, "v"), ".")
	if len(parts) < 3 || len(parts) > 4 {
		return nil, false
	}
	values := make([]int, 4)
	for i, part := range parts {
		value, err := strconv.Atoi(part)
		if err != nil || value < 0 {
			return nil, false
		}
		values[i] = value
	}
	return values, true
}

func shouldUpdateBinary(current, latest string) bool {
	a, okA := versionParts(current)
	b, okB := versionParts(latest)
	if !okA || !okB {
		return false
	}
	for i := range a {
		if b[i] > a[i] {
			return true
		}
		if b[i] < a[i] {
			return false
		}
	}
	return false
}

func parseUpdateFlags(args []string) (updateFlags, error) {
	var parsed updateFlags
	flags := flag.NewFlagSet("update", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	flags.BoolVar(&parsed.Check, "check", false, "check for a published binary update without installing")
	flags.BoolVar(&parsed.SkillsOnly, "skills-only", false, "refresh installed managed skills using the current catalog")
	flags.BoolVar(&parsed.Sync, "sync", false, "internal: reapply recorded selection after binary update")
	flags.BoolVar(&parsed.Finish, "finish", false, "internal: complete Windows executable replacement")
	flags.IntVar(&parsed.ParentPID, "parent-pid", 0, "internal: parent process to wait for")
	flags.StringVar(&parsed.Target, "target", "", "internal: executable to replace")
	flags.StringVar(&parsed.CleanupBin, "cleanup-bin", "", "internal: remove verified helper after handoff")
	flags.BoolVar(&parsed.AdoptLegacy, "adopt-legacy-skills", false, "explicitly adopt matching legacy skill directories")
	flags.StringVar(&parsed.Home, "home", "", "override home directory")
	flags.StringVar(&parsed.Project, "project", "", "also refresh skills in this project (for older starter installs)")
	if err := flags.Parse(args); err != nil {
		return parsed, err
	}
	if flags.NArg() != 0 {
		return parsed, fmt.Errorf("unexpected arguments: %v", flags.Args())
	}
	if parsed.Check && (parsed.SkillsOnly || parsed.Sync) {
		return parsed, fmt.Errorf("--check cannot be combined with --skills-only or --sync")
	}
	if parsed.Sync && parsed.SkillsOnly {
		return parsed, fmt.Errorf("--sync cannot be combined with --skills-only")
	}
	if parsed.Finish && (parsed.ParentPID <= 0 || parsed.Target == "" || parsed.Check || parsed.Sync || parsed.SkillsOnly) {
		return parsed, fmt.Errorf("--finish requires --parent-pid and --target")
	}
	if parsed.CleanupBin != "" && (!parsed.Sync || parsed.ParentPID <= 0) {
		return parsed, fmt.Errorf("--cleanup-bin requires --sync and --parent-pid")
	}
	return parsed, nil
}

type gitRunner struct{}

func (gitRunner) Run(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("empty command")
	}
	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, output)
	}
	return nil
}

func runUpdate(ctx context.Context, args []string, out io.Writer) error {
	flags, err := parseUpdateFlags(args)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			fmt.Fprint(out, updateUsage)
			return nil
		}
		return err
	}
	home := flags.Home
	if home == "" {
		home, err = os.UserHomeDir()
		if err != nil {
			return err
		}
	}
	if flags.Finish {
		return finishUpdate(ctx, flags, home, out)
	}
	if flags.SkillsOnly {
		return refreshSkills(ctx, home, flags.Project, flags.AdoptLegacy, out)
	}
	if flags.Sync {
		if flags.CleanupBin != "" {
			if err := cleanupWindowsHelper(flags.ParentPID, flags.CleanupBin); err != nil {
				fmt.Fprintf(out, "Warning: update helper cleanup failed: %v; continuing synchronization.\n", err)
			}
		}
		return syncRecordedInstall(ctx, home, flags.Project, flags.AdoptLegacy, out)
	}
	client := update.ReleaseClient{}
	release, err := client.Latest(ctx)
	if err != nil {
		if flags.Check {
			return err
		}
		fmt.Fprintf(out, "No se pudo consultar el binario publicado: %v\n", err)
		return errors.Join(err, refreshSkills(ctx, home, flags.Project, flags.AdoptLegacy, out))
	}
	needsBinary := shouldUpdateBinary(buildVersion, release.Tag)
	if !update.ReleaseBinaryAvailable(runtime.GOOS, runtime.GOARCH) {
		if flags.Check {
			fmt.Fprintf(out, "Sin binario de release para %s/%s; las skills se pueden actualizar.\n", runtime.GOOS, runtime.GOARCH)
			return nil
		}
		fmt.Fprintf(out, "Sin binario de release para %s/%s; se sincroniza el stack actual.\n", runtime.GOOS, runtime.GOARCH)
		return syncRecordedInstall(ctx, home, flags.Project, flags.AdoptLegacy, out)
	}
	if flags.Check {
		if needsBinary {
			fmt.Fprintf(out, "Actualización disponible: %s → %s.\n", buildVersion, release.Tag)
		} else if _, valid := versionParts(buildVersion); !valid {
			fmt.Fprintf(out, "Build local %q: no se reemplaza automáticamente. Última release: %s.\n", buildVersion, release.Tag)
		} else {
			fmt.Fprintf(out, "Binario al día (%s); última release: %s.\n", buildVersion, release.Tag)
		}
		return nil
	}
	if !needsBinary {
		fmt.Fprintf(out, "Binario conservado (%s); última release: %s.\n", buildVersion, release.Tag)
		return syncRecordedInstall(ctx, home, flags.Project, flags.AdoptLegacy, out)
	}
	current, err := os.Executable()
	if err != nil {
		return err
	}
	current, err = filepath.Abs(current)
	if err != nil {
		return err
	}
	stageDir, err := os.MkdirTemp("", "active-stack-update-*")
	if err != nil {
		return err
	}
	if runtime.GOOS != "windows" {
		defer os.RemoveAll(stageDir)
	}
	verified, err := client.DownloadBinary(ctx, release, runtime.GOOS, runtime.GOARCH, stageDir)
	if err != nil {
		fmt.Fprintf(out, "No se pudo descargar/verificar el binario: %v\n", err)
		return errors.Join(err, syncRecordedInstall(ctx, home, flags.Project, flags.AdoptLegacy, out))
	}
	backupDir := filepath.Join(home, ".active-stack", "backups", "binaries")
	if runtime.GOOS == "windows" {
		return handoffWindowsUpdate(verified, current, home, flags.Project, flags.AdoptLegacy, out)
	}
	changed, err := update.ReplaceBinary(verified, current, backupDir)
	if err != nil {
		return err
	}
	if changed {
		fmt.Fprintf(out, "Binario actualizado a %s.\n", release.Tag)
	} else {
		fmt.Fprintf(out, "Binario ya coincide con %s.\n", release.Tag)
	}
	// The current process still contains the old embedded catalog. A fresh
	// subprocess reads the newly installed binary before refreshing skills.
	childArgs := []string{"update", "--sync", "--home", home}
	if flags.AdoptLegacy {
		childArgs = append(childArgs, "--adopt-legacy-skills")
	}
	if flags.Project != "" {
		childArgs = append(childArgs, "--project", flags.Project)
	}
	cmd := exec.CommandContext(ctx, current, childArgs...)
	cmd.Stdout = out
	cmd.Stderr = out
	return cmd.Run()
}

func handoffWindowsUpdate(verified, target, home, project string, adoptLegacy bool, out io.Writer) error {
	args := []string{"update", "--finish", "--parent-pid", strconv.Itoa(os.Getpid()), "--target", target, "--home", home}
	if adoptLegacy {
		args = append(args, "--adopt-legacy-skills")
	}
	if project != "" {
		args = append(args, "--project", project)
	}
	cmd := exec.Command(verified, args...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start update helper: %w", err)
	}
	fmt.Fprintf(out, "Actualización verificada; el asistente reemplazará este ejecutable al cerrarse (PID %d). Resultado: %s\n", cmd.Process.Pid, filepath.Join(home, ".active-stack", "logs", "update.log"))
	return update.ErrHandoff
}

func openUpdateLog(home string) (*os.File, error) {
	path := filepath.Join(home, ".active-stack", "logs", "update.log")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	return os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
}

func finishUpdate(ctx context.Context, flags updateFlags, home string, out io.Writer) (resultErr error) {
	if runtime.GOOS != "windows" {
		return fmt.Errorf("--finish is only supported on Windows")
	}
	log, err := openUpdateLog(home)
	if err != nil {
		return fmt.Errorf("open update log: %w", err)
	}
	defer log.Close()
	out = io.MultiWriter(out, log)
	fmt.Fprintln(out, "Windows update: completing executable replacement...")
	defer func() {
		if resultErr != nil {
			fmt.Fprintf(log, "Update failed: %v\n", resultErr)
		}
	}()
	if err := waitForParent(flags.ParentPID); err != nil {
		return err
	}
	verified, err := os.Executable()
	if err != nil {
		return err
	}
	target, err := filepath.Abs(flags.Target)
	if err != nil {
		return err
	}
	changed, err := update.ReplaceBinary(verified, target, filepath.Join(home, ".active-stack", "backups", "binaries"))
	if err != nil {
		return err
	}
	if changed {
		fmt.Fprintln(out, "Binario actualizado; sincronizando harnesses y skills...")
	}
	cmdArgs := []string{"update", "--sync", "--home", home}
	cmdArgs = append(cmdArgs, "--parent-pid", strconv.Itoa(os.Getpid()), "--cleanup-bin", verified)
	if flags.AdoptLegacy {
		cmdArgs = append(cmdArgs, "--adopt-legacy-skills")
	}
	if flags.Project != "" {
		cmdArgs = append(cmdArgs, "--project", flags.Project)
	}
	cmd := exec.Command(target, cmdArgs...)
	// A direct file handle survives this helper exiting; an io.MultiWriter
	// would require the helper to keep copying child pipes after handoff.
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, log, log
	if err := cmd.Start(); err != nil {
		return err
	}
	fmt.Fprintf(out, "Sincronización iniciada (PID %d).\n", cmd.Process.Pid)
	return nil
}

func cleanupWindowsHelper(parentPID int, binary string) error {
	if runtime.GOOS != "windows" {
		return fmt.Errorf("helper cleanup is only supported on Windows")
	}
	path, err := filepath.Abs(binary)
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	temp, err := filepath.Abs(os.TempDir())
	if err != nil {
		return err
	}
	if !samePath(filepath.Dir(dir), temp) || !strings.HasPrefix(filepath.Base(dir), "active-stack-update-") || !strings.HasPrefix(filepath.Base(path), "active-stack-verified-") {
		return fmt.Errorf("unexpected helper path %q", path)
	}
	if err := waitForParent(parentPID); err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	if err := os.Remove(dir); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func installArgs(agent model.Agent, selection update.AgentSelection, home string) []string {
	args := []string{"install", "--headless", "--yes", "--no-self-install", "--home", home,
		"--agent", string(agent), "--mode", string(selection.Mode), "--tier", string(selection.Tier.Normalize())}
	if selection.Mode == model.ModeCustom {
		args = append(args, "--custom", strings.Join(selection.Custom, ","))
	}
	return args
}

func installRuns(agent model.Agent, selection update.AgentSelection, home string) [][]string {
	runs := [][]string{installArgs(agent, selection, home)}
	if selection.Mode != model.ModeCustom && len(selection.Custom) > 0 {
		runs = append(runs, installArgs(agent, update.AgentSelection{Mode: model.ModeCustom, Custom: selection.Custom, Tier: selection.Tier}, home))
	}
	return runs
}

func syncRecordedInstall(ctx context.Context, home, project string, adoptLegacy bool, out io.Writer) error {
	state, err := update.LoadState(home)
	if err != nil {
		return err
	}
	if len(state.Agents) == 0 {
		fmt.Fprintln(out, "Instalación anterior sin registro de selección: se actualizan las skills detectadas. Para actualizar configuraciones, ejecutá install una vez con tu modo actual.")
		return refreshSkills(ctx, home, project, adoptLegacy, out)
	}
	if err := preflightRecordedSkills(home, state, adoptLegacy); err != nil {
		return err
	}
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	agents := make([]model.Agent, 0, len(state.Agents))
	for agent := range state.Agents {
		agents = append(agents, agent)
	}
	sort.Slice(agents, func(i, j int) bool { return agents[i] < agents[j] })
	var failures []error
	for _, agent := range agents {
		selection := state.Agents[agent]
		if !selection.Mode.IsValid() {
			failures = append(failures, fmt.Errorf("%s: invalid recorded mode %q", agent, selection.Mode))
			continue
		}
		fmt.Fprintf(out, "Reaplicando harnesses de %s (%s)...\n", agent, selection.Mode)
		for _, args := range installRuns(agent, selection, home) {
			cmd := exec.CommandContext(ctx, exe, args...)
			cmd.Stdout, cmd.Stderr = out, out
			if err := cmd.Run(); err != nil {
				failures = append(failures, fmt.Errorf("%s (%s): %w", agent, args[9], err))
			}
		}
	}
	// Also refresh known managed skills from pre-receipt installs. This catches
	// existing skills outside the currently recorded selection without adding
	// a skill the user never installed.
	if err := refreshSkills(ctx, home, project, adoptLegacy, out); err != nil {
		failures = append(failures, err)
	}
	return errors.Join(failures...)
}

// Check every agent before starting any replay, including dependency skills.
// A preset may gain skills in a new release; its receipt is not ownership of
// an existing directory that the user created independently.
func preflightRecordedSkills(home string, state update.State, adoptLegacy bool) error {
	if adoptLegacy {
		return nil
	}
	cat, err := catalog.Load()
	if err != nil {
		return err
	}
	reg, err := agents.NewDefaultRegistry()
	if err != nil {
		return err
	}
	all := make(map[string]model.Harness)
	for _, h := range cat.AllHarnesses() {
		all[h.ID] = h
	}
	for agent, selection := range state.Agents {
		adapter, ok := reg.Get(agent)
		if !ok {
			return fmt.Errorf("unknown recorded agent %q", agent)
		}
		selected, err := install.SelectHarnesses(cat, install.Intent{Agents: []model.Agent{agent}, Mode: selection.Mode, Custom: selection.Custom})
		if err != nil {
			return err
		}
		// Extra Custom harnesses remain additive to a recorded preset.
		for _, id := range selection.Custom {
			if h, ok := all[id]; ok && h.SupportsAgent(agent) {
				selected = append(selected, h)
			}
		}
		resolved, err := planner.NewResolver().Resolve(selected, all)
		if err != nil {
			return err
		}
		for _, id := range resolved.OrderedIDs {
			h := all[id]
			if h.Type != model.HarnessSkill || !h.SupportsAgent(agent) || adapter.SkillsDir(home) == "" {
				continue
			}
			path := filepath.Join(adapter.SkillsDir(home), id)
			if _, err := os.Lstat(path); os.IsNotExist(err) {
				continue
			} else if err != nil {
				return err
			}
			owned := false
			for _, recorded := range state.ManagedSkills[agent] {
				if samePath(recorded, path) {
					owned = true
					break
				}
			}
			if !owned {
				return fmt.Errorf("skill directory has no ownership receipt: %s; review it and rerun update --adopt-legacy-skills to authorize replacement", path)
			}
		}
	}
	return nil
}

func binaryName(goos string) string {
	if goos == "windows" {
		return "active-stack.exe"
	}
	return "active-stack"
}

func samePath(a, b string) bool {
	a, _ = filepath.Abs(a)
	b, _ = filepath.Abs(b)
	if first, err := os.Stat(a); err == nil {
		if second, err := os.Stat(b); err == nil && os.SameFile(first, second) {
			return true
		}
	}
	if runtime.GOOS == "windows" {
		return strings.EqualFold(filepath.Clean(a), filepath.Clean(b))
	}
	return filepath.Clean(a) == filepath.Clean(b)
}

func refreshSkills(ctx context.Context, home, project string, adoptLegacy bool, out io.Writer) error {
	cat, err := catalog.Load()
	if err != nil {
		return err
	}
	reg, err := agents.NewDefaultRegistry()
	if err != nil {
		return err
	}
	targets, err := update.ScanInstalledSkills(cat, reg, home)
	if err != nil {
		return err
	}
	state, err := update.LoadState(home)
	if err != nil {
		return err
	}
	projects := append([]string(nil), state.Projects...)
	if project != "" {
		projects = append(projects, project)
	}
	seenProjects := make(map[string]bool)
	for _, root := range projects {
		absolute, err := filepath.Abs(root)
		if err != nil {
			return err
		}
		if seenProjects[absolute] {
			continue
		}
		seenProjects[absolute] = true
		info, err := os.Stat(absolute)
		if os.IsNotExist(err) {
			fmt.Fprintf(out, "Proyecto registrado no encontrado, omitido: %s\n", absolute)
			continue
		}
		if err != nil {
			return err
		}
		if !info.IsDir() {
			return fmt.Errorf("project %s is not a directory", absolute)
		}
		projectTargets, err := update.ScanProjectSkills(cat, reg, absolute)
		if err != nil {
			return err
		}
		targets = append(targets, projectTargets...)
	}
	owned := make(map[string]bool)
	for _, paths := range state.ManagedSkills {
		for _, path := range paths {
			owned[filepath.Clean(path)] = true
		}
	}
	eligible := targets[:0]
	skipped := 0
	for _, target := range targets {
		if adoptLegacy || owned[filepath.Clean(target.Path)] {
			eligible = append(eligible, target)
		} else {
			skipped++
		}
	}
	targets = eligible
	fmt.Fprintf(out, "Skills gestionadas: %d; coincidencias sin registro omitidas: %d.\n", len(targets), skipped)
	if skipped > 0 && !adoptLegacy {
		fmt.Fprintln(out, "Si instalaste esas skills con una versión anterior, revisalas y ejecutá update --adopt-legacy-skills para adoptarlas.")
	}
	outcomes := update.RefreshInstalledSkills(ctx, targets, reg, home, resolveBackupDir(home), gitRunner{}, assets.SkillsFS)
	var failures []error
	updated, unchanged := 0, 0
	var adopted []update.SkillTarget
	for _, outcome := range outcomes {
		if outcome.Err != nil {
			fmt.Fprintf(out, "ERROR %s/%s: %v\n", outcome.Target.Agent, outcome.Target.Harness.ID, outcome.Err)
			failures = append(failures, outcome.Err)
		} else if outcome.Result.AlreadyInstalled {
			unchanged++
			if adoptLegacy {
				adopted = append(adopted, outcome.Target)
			}
		} else {
			updated++
			if adoptLegacy {
				adopted = append(adopted, outcome.Target)
			}
			fmt.Fprintf(out, "Actualizada %s/%s\n", outcome.Target.Agent, outcome.Target.Harness.ID)
		}
	}
	if err := update.RecordAdoptedSkills(home, adopted); err != nil {
		failures = append(failures, err)
	}
	fmt.Fprintf(out, "Skills: %d actualizadas, %d sin cambios, %d con errores.\n", updated, unchanged, len(failures))
	return errors.Join(failures...)
}
