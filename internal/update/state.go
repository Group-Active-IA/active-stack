package update

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/Group-Active-IA/active-stack/internal/agents"
	"github.com/Group-Active-IA/active-stack/internal/backup"
	"github.com/Group-Active-IA/active-stack/internal/catalog"
	"github.com/Group-Active-IA/active-stack/internal/install"
	"github.com/Group-Active-IA/active-stack/internal/model"
)

// AgentSelection stores the choices needed to reapply an install for one agent.
type AgentSelection struct {
	Mode   model.InstallMode    `json:"mode"`
	Custom []string             `json:"custom,omitempty"`
	Tier   model.PermissionTier `json:"tier,omitempty"`
}

// State records managed selections. Older installs can be migrated by scanning
// installed skills, but their mode cannot be safely inferred.
type State struct {
	Version       int                            `json:"version"`
	Agents        map[model.Agent]AgentSelection `json:"agents"`
	Projects      []string                       `json:"projects,omitempty"`
	ManagedSkills map[model.Agent][]string       `json:"managed_skills,omitempty"`
}

func statePath(home string) string { return filepath.Join(home, ".active-stack", "installation.json") }

func LoadState(home string) (State, error) {
	data, err := os.ReadFile(statePath(home))
	if os.IsNotExist(err) {
		return State{Version: 1, Agents: make(map[model.Agent]AgentSelection), ManagedSkills: make(map[model.Agent][]string)}, nil
	}
	if err != nil {
		return State{}, err
	}
	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		return State{}, fmt.Errorf("parse installation state: %w", err)
	}
	if state.Version != 1 {
		return State{}, fmt.Errorf("unsupported installation state version %d", state.Version)
	}
	if state.Agents == nil {
		state.Agents = make(map[model.Agent]AgentSelection)
	}
	if state.ManagedSkills == nil {
		state.ManagedSkills = make(map[model.Agent][]string)
	}
	return state, nil
}

// SaveIntent merges a successful machine install into the state file. It takes
// a durable backup before replacing an existing receipt and never discards
// selections belonging to other agents.
func SaveIntent(home string, intent install.Intent) error {
	if !intent.Mode.IsValid() {
		return fmt.Errorf("invalid install mode %q", intent.Mode)
	}
	state, err := LoadState(home)
	if err != nil {
		return err
	}
	cat, err := catalog.Load()
	if err != nil {
		return err
	}
	reg, err := agents.NewDefaultRegistry()
	if err != nil {
		return err
	}
	var selected []model.Harness
	if intent.Mode == model.ModeCustom {
		for _, id := range intent.Custom {
			if h, ok := cat.ByID(id); ok {
				selected = append(selected, h)
			}
		}
	} else {
		selected = cat.ForMode(intent.Mode)
	}
	for _, agent := range intent.Agents {
		selection, exists := state.Agents[agent]
		if !exists || intent.Mode == model.ModeFull || (intent.Mode == model.ModeLite && selection.Mode != model.ModeFull) {
			selection.Mode = intent.Mode
		}
		if intent.Mode == model.ModeCustom {
			seen := make(map[string]bool, len(selection.Custom)+len(intent.Custom))
			for _, id := range selection.Custom {
				seen[id] = true
			}
			for _, id := range intent.Custom {
				seen[id] = true
			}
			selection.Custom = selection.Custom[:0]
			for id := range seen {
				selection.Custom = append(selection.Custom, id)
			}
			sort.Strings(selection.Custom)
		}
		selection.Tier = intent.Tier
		state.Agents[agent] = selection
		adapter, ok := reg.Get(agent)
		if !ok {
			continue
		}
		for _, h := range selected {
			if h.Type != model.HarnessSkill || !h.SupportsAgent(agent) {
				continue
			}
			path := filepath.Join(adapter.SkillsDir(home), h.ID)
			if info, err := os.Stat(filepath.Join(path, "SKILL.md")); err == nil && info.Mode().IsRegular() {
				state.ManagedSkills[agent] = appendUniquePath(state.ManagedSkills[agent], path)
			}
		}
	}
	return saveState(home, state)
}

// DropAgents invalidates selections affected by an uninstall. For partial
// uninstalls we deliberately forget the whole agent selection: reapplying it
// would reinstall harnesses the user intentionally removed.
func DropAgents(home string, agents []model.Agent) error {
	state, err := LoadState(home)
	if err != nil {
		return err
	}
	if len(agents) == 0 {
		clear(state.Agents)
		clear(state.ManagedSkills)
	} else {
		for _, agent := range agents {
			delete(state.Agents, agent)
			delete(state.ManagedSkills, agent)
		}
	}
	return saveState(home, state)
}

func appendUniquePath(paths []string, path string) []string {
	clean, err := filepath.Abs(path)
	if err != nil {
		return paths
	}
	for _, existing := range paths {
		if existing == clean {
			return paths
		}
	}
	return append(paths, clean)
}

// RecordProject keeps starter targets discoverable by a later update. It does
// not read or alter project files.
func RecordProject(home, projectRoot string) error {
	root, err := filepath.Abs(projectRoot)
	if err != nil {
		return err
	}
	state, err := LoadState(home)
	if err != nil {
		return err
	}
	for _, known := range state.Projects {
		if known == root {
			return nil
		}
	}
	state.Projects = append(state.Projects, root)
	return saveState(home, state)
}

// RecordStarter records only skill paths selected by a successful starter run.
func RecordStarter(home, projectRoot, starterID string, installedAgents []model.Agent) error {
	root, err := filepath.Abs(projectRoot)
	if err != nil {
		return err
	}
	cat, err := catalog.Load()
	if err != nil {
		return err
	}
	reg, err := agents.NewDefaultRegistry()
	if err != nil {
		return err
	}
	selected, err := cat.ResolveStarter(starterID)
	if err != nil {
		return err
	}
	state, err := LoadState(home)
	if err != nil {
		return err
	}
	state.Projects = appendUniquePath(state.Projects, root)
	for _, agent := range installedAgents {
		adapter, ok := reg.Get(agent)
		if !ok {
			continue
		}
		for _, h := range selected {
			if h.Type != model.HarnessSkill || !h.SupportsAgent(agent) {
				continue
			}
			path := filepath.Join(adapter.PathsFor(root, model.Project).SkillsDir, h.ID)
			if info, err := os.Stat(filepath.Join(path, "SKILL.md")); err == nil && info.Mode().IsRegular() {
				state.ManagedSkills[agent] = appendUniquePath(state.ManagedSkills[agent], path)
			}
		}
	}
	return saveState(home, state)
}

// RecordAdoptedSkills persists the successful one-time adoption of legacy
// skill directories, so later updates can refresh them without another flag.
func RecordAdoptedSkills(home string, targets []SkillTarget) error {
	if len(targets) == 0 {
		return nil
	}
	state, err := LoadState(home)
	if err != nil {
		return err
	}
	for _, target := range targets {
		state.ManagedSkills[target.Agent] = appendUniquePath(state.ManagedSkills[target.Agent], target.Path)
		if target.ProjectRoot != "" {
			state.Projects = appendUniquePath(state.Projects, target.ProjectRoot)
		}
	}
	return saveState(home, state)
}

func saveState(home string, state State) error {
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	path := statePath(home)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	oldInfo, err := os.Lstat(path)
	exists := err == nil
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if exists && !oldInfo.Mode().IsRegular() {
		return fmt.Errorf("installation state is not a regular file: %s", path)
	}
	if exists {
		backupDir := filepath.Join(home, ".active-stack", "backups", "state")
		snapshotDir := filepath.Join(backupDir, time.Now().UTC().Format("20060102-150405.000000000")+"-installation")
		if _, err := backup.NewSnapshotter().Create(snapshotDir, []string{path}); err != nil {
			return fmt.Errorf("snapshot installation state: %w", err)
		}
	}
	stage, err := os.CreateTemp(filepath.Dir(path), ".installation-*")
	if err != nil {
		return err
	}
	stagePath := stage.Name()
	defer os.Remove(stagePath)
	if _, err := stage.Write(data); err != nil {
		stage.Close()
		return err
	}
	if err := stage.Close(); err != nil {
		return err
	}
	old := stagePath + ".old"
	if exists {
		if err := os.Rename(path, old); err != nil {
			return err
		}
	}
	if err := os.Rename(stagePath, path); err != nil {
		if exists {
			_ = os.Rename(old, path)
		}
		return err
	}
	if exists {
		_ = os.Remove(old)
	}
	return nil
}
