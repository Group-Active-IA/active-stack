package install

import (
	"fmt"
	"os"

	"github.com/Group-Active-IA/active-stack/internal/backup"
	"github.com/Group-Active-IA/active-stack/internal/harness/autosync"
	"github.com/Group-Active-IA/active-stack/internal/harness/hooks"
	"github.com/Group-Active-IA/active-stack/internal/model"
	"github.com/Group-Active-IA/active-stack/internal/pipeline"
)

// EngramCloudHarnessID is the harness whose installation also wires the memory
// read hooks (R-16) and ENGRAM_CLOUD_AUTOSYNC (R-20) at user level.
const EngramCloudHarnessID = "engram-cloud"

// CloudStepContext is what an extra engram-cloud step needs to be built.
type CloudStepContext struct {
	Adapters []AgentAdapter // only the agents the harness supports (Claude, Codex)
	Home     string
}

// cloudExtraSteps are further steps that run right after the engram-cloud
// steps. The save-tagging proxy (decision DA-13, ficha R-26) is pending: when
// it exists, adding it to the install is one RegisterCloudStep call from its
// package, with nothing else to change here. Nothing is registered today.
var cloudExtraSteps []func(CloudStepContext) pipeline.Step

// RegisterCloudStep appends a factory for an extra engram-cloud step.
func RegisterCloudStep(f func(CloudStepContext) pipeline.Step) {
	cloudExtraSteps = append(cloudExtraSteps, f)
}

// cloudAgents keeps the agents the engram-cloud harness supports: OpenCode and
// Antigravity are never touched.
func cloudAgents(h model.Harness, adapters []AgentAdapter) []AgentAdapter {
	var out []AgentAdapter
	for _, a := range adapters {
		if h.SupportsAgent(a.Agent()) && (a.Agent() == model.AgentClaude || a.Agent() == model.AgentCodex) {
			out = append(out, a)
		}
	}
	return out
}

var hooksInstallFn = func(home string, adapters []hooks.Adapter) (hooks.Result, error) {
	return hooks.Install(home, adapters)
}

var autosyncSetFn = func(home string, adapters []autosync.Adapter) (autosync.Result, error) {
	return autosync.Set(home, adapters)
}

// cloudNotice emits a non-fatal message the same way configStep does.
func cloudNotice(onProgress pipeline.ProgressFunc, id, msg string) {
	if onProgress != nil {
		onProgress(pipeline.ProgressEvent{StepID: id, Stage: pipeline.StageApply, Status: pipeline.StepStatusDegraded, Err: fmt.Errorf("%s", msg)})
		return
	}
	fmt.Fprintln(os.Stderr, msg)
}

type memoryHooksStep struct {
	adapters   []AgentAdapter
	home       string
	onProgress pipeline.ProgressFunc
	manifest   *backup.Manifest
}

func (s *memoryHooksStep) ID() string                     { return "hooks:" + EngramCloudHarnessID }
func (s *memoryHooksStep) setManifest(m *backup.Manifest) { s.manifest = m }

func (s *memoryHooksStep) Run() error {
	ad := make([]hooks.Adapter, len(s.adapters))
	for i, a := range s.adapters {
		ad[i] = a
	}
	res, err := hooksInstallFn(s.home, ad)
	for _, r := range res.Agents {
		for _, w := range r.Warnings {
			cloudNotice(s.onProgress, s.ID(), "warning: "+w)
		}
	}
	for _, n := range res.Notices {
		cloudNotice(s.onProgress, s.ID(), n)
	}
	return err
}

func (s *memoryHooksStep) Rollback() error {
	if s.manifest == nil {
		return nil
	}
	return restoreFn(*s.manifest)
}

type autosyncStep struct {
	adapters   []AgentAdapter
	home       string
	onProgress pipeline.ProgressFunc
	manifest   *backup.Manifest
}

func (s *autosyncStep) ID() string                     { return "autosync:" + EngramCloudHarnessID }
func (s *autosyncStep) setManifest(m *backup.Manifest) { s.manifest = m }

func (s *autosyncStep) Run() error {
	ad := make([]autosync.Adapter, len(s.adapters))
	for i, a := range s.adapters {
		ad[i] = a
	}
	res, err := autosyncSetFn(s.home, ad)
	for _, r := range res.Agents {
		for _, w := range r.Warnings {
			cloudNotice(s.onProgress, s.ID(), "warning: "+w)
		}
		if r.Skipped && r.Reason == "no engram MCP entry in the file" {
			cloudNotice(s.onProgress, s.ID(), fmt.Sprintf("%s: no engram MCP entry in %s, so %s was not set (the plugin may provide Engram)", r.Agent, r.File, autosync.EnvKey))
		}
	}
	return err
}

func (s *autosyncStep) Rollback() error {
	if s.manifest == nil {
		return nil
	}
	return restoreFn(*s.manifest)
}

// cloudSteps returns the steps that follow the engram-cloud config step. They
// are user-level (decision D12), so a Project target gets none.
func cloudSteps(h model.Harness, adapters []AgentAdapter, opts Options) []pipeline.Step {
	if h.ID != EngramCloudHarnessID || opts.Target == model.Project {
		return nil
	}
	ca := cloudAgents(h, adapters)
	if len(ca) == 0 {
		return nil
	}
	steps := []pipeline.Step{
		&memoryHooksStep{adapters: ca, home: opts.HomeDir, onProgress: opts.OnProgress},
		&autosyncStep{adapters: ca, home: opts.HomeDir, onProgress: opts.OnProgress},
	}
	for _, f := range cloudExtraSteps {
		if st := f(CloudStepContext{Adapters: ca, Home: opts.HomeDir}); st != nil {
			steps = append(steps, st)
		}
	}
	return steps
}

// cloudWritePaths adds the files the cloud steps edit, so the install snapshot
// covers them and a rollback restores them.
func cloudWritePaths(h model.Harness, adapters []AgentAdapter, home string, target model.InstallTarget, add func(string)) {
	if h.ID != EngramCloudHarnessID || target == model.Project {
		return
	}
	for _, a := range cloudAgents(h, adapters) {
		add(a.SettingsPath(home))
		add(a.MCPConfigPath(home, autosync.ServerName))
	}
}
