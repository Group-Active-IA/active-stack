package manifest

import (
	"reflect"
	"testing"
)

func cohort(id string, mutate func(*Cell)) Cohort {
	c := baseCell()
	if mutate != nil {
		mutate(&c)
	}
	return Cohort{ID: id, Cell: c, Tasks: []TaskWeight{{TaskID: "exploracion", Weight: 1}}}
}

func TestManifestPlan_ExcludesIncompatibleFromDenominator(t *testing.T) {
	m := validManifest()
	m.Cohorts = []Cohort{
		cohort("claude-terminal-linux", nil),
		cohort("claude-terminal-windows-vieja", func(c *Cell) { c.OS = OSWindows; c.Version = "2.0.1" }),
		cohort("claude-ide-linux-otel", func(c *Cell) {
			c.Interface = InterfaceIDE
			c.Observed.Settings = map[string]string{"OTEL_LOG_USER_PROMPTS": "1"}
		}),
		cohort("claude-terminal-windows", func(c *Cell) { c.OS = OSWindows }),
	}
	got := mustPlan(t, m)

	wantExec := []string{"claude-terminal-linux", "claude-terminal-windows"}
	if !reflect.DeepEqual(plannedIDs(got), wantExec) {
		t.Errorf("Executable = %v, want %v", got.Executable, wantExec)
	}
	if len(got.Incompatible) != 2 {
		t.Fatalf("Incompatible = %+v, want 2 entries", got.Incompatible)
	}
	if got.Incompatible[0].CohortID != "claude-terminal-windows-vieja" || len(got.Incompatible[0].Reasons) != 1 {
		t.Errorf("Incompatible[0] = %+v", got.Incompatible[0])
	}
	if got.Incompatible[1].CohortID != "claude-ide-linux-otel" || len(got.Incompatible[1].Reasons) != 1 {
		t.Errorf("Incompatible[1] = %+v", got.Incompatible[1])
	}
}

func TestManifestPlan_AllExecutableAndAllIncompatible(t *testing.T) {
	ok := validManifest()
	ok.Cohorts = []Cohort{cohort("a", nil)}
	if p := mustPlan(t, ok); len(p.Executable) != 1 || len(p.Incompatible) != 0 {
		t.Errorf("all-executable plan = %+v", p)
	}
	bad := validManifest()
	bad.Cohorts = []Cohort{cohort("a", func(c *Cell) { c.Version = "1.0.0" })}
	if p := mustPlan(t, bad); len(p.Executable) != 0 || len(p.Incompatible) != 1 {
		t.Errorf("all-incompatible plan = %+v", p)
	}
	if _, err := (Manifest{}).Plan(); err == nil {
		t.Error("Plan() de un manifiesto vacio debe fallar por Validate")
	}
}

func TestRunStatus_DistinctAndValid(t *testing.T) {
	all := []RunStatus{RunAccepted, RunRejectedQuality, RunBlockedPermissions, RunTimeout, RunClientFailure}
	seen := map[RunStatus]bool{}
	for _, s := range all {
		if seen[s] {
			t.Errorf("estado duplicado %q", s)
		}
		seen[s] = true
		if !s.Valid() {
			t.Errorf("%q debería ser válido", s)
		}
	}
	if RunStatus("").Valid() || RunStatus("incompatible").Valid() || RunStatus("otro").Valid() {
		t.Error("estados desconocidos o de celda no deben ser válidos como estado de corrida")
	}
	if RunRejectedQuality == RunBlockedPermissions {
		t.Error("rechazo por calidad y bloqueo por permisos deben ser estados distintos")
	}
}

func mustPlan(t *testing.T, m Manifest) Plan {
	t.Helper()
	p, err := m.Plan()
	if err != nil {
		t.Fatalf("Plan() error = %v", err)
	}
	return p
}

func plannedIDs(p Plan) []string {
	var ids []string
	for _, e := range p.Executable {
		ids = append(ids, e.CohortID)
	}
	return ids
}
