package update

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/Group-Active-IA/active-stack/internal/agents"
	"github.com/Group-Active-IA/active-stack/internal/catalog"
	"github.com/Group-Active-IA/active-stack/internal/harness/skill"
	"github.com/Group-Active-IA/active-stack/internal/model"
)

// SkillTarget is a catalog skill already present in an agent's global skill directory.
type SkillTarget struct {
	Harness     model.Harness
	Agent       model.Agent
	Path        string
	ProjectRoot string
}

// SkillOutcome records a separate result for each target, including failures.
// A failed optional skill must never be reported as updated.
type SkillOutcome struct {
	Target SkillTarget
	Result skill.Result
	Err    error
}

// RefreshInstalledSkills updates only the supplied installed targets to the
// source revision approved by the current catalog. A pinned ref stays pinned.
func RefreshInstalledSkills(ctx context.Context, targets []SkillTarget, reg *agents.Registry, home, backupDir string, runner skill.Runner, embedded fs.FS) []SkillOutcome {
	installer := skill.NewInstaller(runner, embedded)
	outcomes := make([]SkillOutcome, 0, len(targets))
	for _, target := range targets {
		adapter, ok := reg.Get(target.Agent)
		if !ok {
			outcomes = append(outcomes, SkillOutcome{Target: target, Err: fmt.Errorf("agent %s has no adapter", target.Agent)})
			continue
		}
		var skillAdapter skill.AgentAdapter = adapter
		if target.ProjectRoot != "" {
			skillAdapter = projectSkillAdapter{agent: target.Agent, dir: adapter.PathsFor(target.ProjectRoot, model.Project).SkillsDir}
		}
		results, err := installer.Install(ctx, target.Harness, []skill.AgentAdapter{skillAdapter}, home, backupDir)
		outcome := SkillOutcome{Target: target, Err: err}
		if err == nil {
			if len(results) != 1 {
				outcome.Err = fmt.Errorf("skill %s: expected one install result, got %d", target.Harness.ID, len(results))
			} else {
				outcome.Result = results[0]
			}
		}
		outcomes = append(outcomes, outcome)
	}
	return outcomes
}

type projectSkillAdapter struct {
	agent model.Agent
	dir   string
}

func (a projectSkillAdapter) Agent() model.Agent      { return a.agent }
func (a projectSkillAdapter) SkillsDir(string) string { return a.dir }

// ScanInstalledSkills migrates older installs without a receipt by inspecting only
// catalog-owned skill names through each agent adapter. Unrelated user skills are ignored.
func ScanInstalledSkills(cat *catalog.Catalog, reg *agents.Registry, home string) ([]SkillTarget, error) {
	return scanSkills(cat, reg, home, false)
}

// ScanProjectSkills finds catalog skills installed under a project's per-agent
// layout. The adapter remains the sole authority for these paths.
func ScanProjectSkills(cat *catalog.Catalog, reg *agents.Registry, projectRoot string) ([]SkillTarget, error) {
	return scanSkills(cat, reg, projectRoot, true)
}

func scanSkills(cat *catalog.Catalog, reg *agents.Registry, base string, project bool) ([]SkillTarget, error) {
	var targets []SkillTarget
	seen := make(map[string]bool)
	for _, agent := range reg.SupportedAgents() {
		adapter, _ := reg.Get(agent)
		dir := adapter.SkillsDir(base)
		if project {
			dir = adapter.PathsFor(base, model.Project).SkillsDir
		}
		if dir == "" {
			continue
		}
		for _, h := range cat.AllHarnesses() {
			if h.Type != model.HarnessSkill || (!project && h.IsStarterOnly()) || !h.SupportsAgent(agent) {
				continue
			}
			path := filepath.Join(dir, h.ID)
			info, err := os.Lstat(path)
			if os.IsNotExist(err) {
				continue
			}
			if err != nil {
				return nil, fmt.Errorf("inspect %s: %w", path, err)
			}
			if !info.IsDir() {
				return nil, fmt.Errorf("managed skill path %s is not a directory", path)
			}
			skillInfo, err := os.Lstat(filepath.Join(path, "SKILL.md"))
			if os.IsNotExist(err) {
				continue
			}
			if err != nil {
				return nil, fmt.Errorf("inspect skill %s: %w", path, err)
			}
			if skillInfo.Mode().IsRegular() {
				key := filepath.Clean(path)
				if runtime.GOOS == "windows" {
					key = strings.ToLower(key)
				}
				if seen[key] {
					continue
				}
				seen[key] = true
				target := SkillTarget{Harness: h, Agent: agent, Path: path}
				if project {
					target.ProjectRoot = base
				}
				targets = append(targets, target)
			}
		}
	}
	sort.Slice(targets, func(i, j int) bool {
		if targets[i].Agent == targets[j].Agent {
			return targets[i].Harness.ID < targets[j].Harness.ID
		}
		return targets[i].Agent < targets[j].Agent
	})
	return targets, nil
}
