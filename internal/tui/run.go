package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/JuanCruzRobledo/jr-stack/internal/install"
	"github.com/JuanCruzRobledo/jr-stack/internal/pipeline"
)

// defaultRunPlan executes the install plan inside a goroutine. It wires
// plan.OnProgress through the bridge and sends a doneMsg when done.
//
// The send function is tea.Program.Send — it must be set by the caller.
// When send is nil (e.g. in tests that inject a fake RunPlanFn), the goroutine
// still runs but cannot deliver messages; the fake is expected to do that.
func defaultRunPlan(plan install.Plan, bridge *progressBridge, send func(tea.Msg)) {
	go func() {
		defer bridge.close()

		orch := pipeline.NewOrchestrator(
			pipeline.DefaultRollbackPolicy(),
			pipeline.WithProgressFunc(plan.OnProgress),
		)
		result := orch.Execute(plan.StagePlan)

		if send != nil {
			send(doneMsg{result: result})
		}
	}()
}

// NewProgram creates a tea.Program for the install TUI. The program wires
// tea.Program.Send into RunPlanFn so the goroutine can deliver doneMsg back.
func NewProgram(deps ModelDeps) *tea.Program {
	m := newModel(deps)
	var p *tea.Program

	// Wrap RunPlanFn to inject tea.Program.Send at runtime.
	originalRun := deps.RunPlanFn
	if originalRun == nil {
		originalRun = defaultRunPlan
	}
	deps.RunPlanFn = func(plan install.Plan, bridge *progressBridge, _ func(tea.Msg)) {
		// p is set below; by the time RunPlanFn is called the program exists.
		originalRun(plan, bridge, func(msg tea.Msg) {
			if p != nil {
				p.Send(msg)
			}
		})
	}
	m.deps = deps

	p = tea.NewProgram(m)
	return p
}
