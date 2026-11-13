package catalog

import (
	"reflect"
	"slices"
	"strings"
	"testing"
)

// Tarea 6.1 (opsx-task-context, D19): tabla de carga de módulos por acción y
// función pura que la resuelve (specs/task-context-loading).

func loadTC(t *testing.T) *TaskContext {
	t.Helper()
	c, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	return c.TaskContext
}

func entryIDs(es []Entry) []string {
	var out []string
	for _, e := range es {
		out = append(out, e.Module)
	}
	return out
}

var allToggles = []string{"tdd", "engram", "model-routing", "delegation", "governance"}

func TestResolve_MandatoryPerAction(t *testing.T) {
	tc := loadTC(t)
	tests := []struct {
		action string
		want   []string // obligatorios, en el orden canónico
	}{
		{"explore", nil},
		{"propose", []string{"artifacts"}},
		{"apply", []string{"artifacts", "bounded-command-output", "tdd"}},
		{"verify", []string{"delegation", "model-assignments", "bounded-command-output"}},
		{"archive", []string{"artifacts"}},
	}
	for _, tt := range tests {
		t.Run(tt.action, func(t *testing.T) {
			r, err := tc.Resolve(Query{Action: tt.action, Role: RoleCoordinator, Toggles: allToggles, StrictTDD: true})
			if err != nil {
				t.Fatal(err)
			}
			var got []string
			for _, e := range r.Entries {
				if e.Level == LevelMandatory {
					got = append(got, e.Module)
				}
			}
			if !slices.Equal(got, tt.want) {
				t.Errorf("obligatorios en %s = %v, want %v", tt.action, got, tt.want)
			}
			if len(r.Blocked) != 0 {
				t.Errorf("con todos los toggles no debe haber bloqueos: %v", r.Blocked)
			}
		})
	}
}

func TestResolve_Deterministic(t *testing.T) {
	tc := loadTC(t)
	q := Query{Action: "apply", Role: RoleCoordinator, Toggles: allToggles, StrictTDD: true}
	a, err := tc.Resolve(q)
	if err != nil {
		t.Fatal(err)
	}
	b, _ := tc.Resolve(q)
	if !reflect.DeepEqual(a, b) {
		t.Errorf("dos resoluciones iguales difieren:\n%+v\n%+v", a, b)
	}
	// El orden de los toggles no altera el resultado.
	rev := slices.Clone(allToggles)
	slices.Reverse(rev)
	q.Toggles = rev
	c, _ := tc.Resolve(q)
	if !reflect.DeepEqual(a, c) {
		t.Errorf("el orden de los toggles alteró el resultado:\n%+v\n%+v", a, c)
	}
	// El orden de declaración de los módulos en el catálogo tampoco.
	flipped := *tc
	flipped.Modules = slices.Clone(tc.Modules)
	slices.Reverse(flipped.Modules)
	d, _ := flipped.Resolve(Query{Action: "apply", Role: RoleCoordinator, Toggles: allToggles, StrictTDD: true})
	if !reflect.DeepEqual(a, d) {
		t.Errorf("el orden de declaración de módulos alteró el resultado:\n%+v\n%+v", a, d)
	}
}

func TestResolve_TDDOnlyWhenRequired(t *testing.T) {
	tc := loadTC(t)
	// Sin requisito de TDD no aparece tdd, aunque el toggle esté activo.
	r, err := tc.Resolve(Query{Action: "apply", Role: RoleCoordinator, Toggles: allToggles, StrictTDD: false})
	if err != nil {
		t.Fatal(err)
	}
	if slices.Contains(entryIDs(r.Entries), "tdd") {
		t.Errorf("tdd figura sin requisito de TDD: %v", entryIDs(r.Entries))
	}
	// Con requisito y toggle inactivo, apply se detiene.
	r, _ = tc.Resolve(Query{Action: "apply", Role: RoleCoordinator, Toggles: []string{"delegation"}, StrictTDD: true})
	if !slices.Equal(r.Blocked, []string{"tdd"}) {
		t.Errorf("apply con TDD exigido y sin toggle: Blocked = %v, want [tdd]", r.Blocked)
	}
	if slices.Contains(entryIDs(r.Entries), "tdd") {
		t.Error("un módulo de toggle inactivo no figura")
	}
	// tdd no se carga en otras acciones.
	r, _ = tc.Resolve(Query{Action: "verify", Role: RoleCoordinator, Toggles: allToggles, StrictTDD: true})
	if slices.Contains(entryIDs(r.Entries), "tdd") {
		t.Error("tdd solo se carga en apply")
	}
}

func TestResolve_DelegatingBringsModelAssignments(t *testing.T) {
	tc := loadTC(t)
	r, err := tc.Resolve(Query{Action: "apply", Role: RoleCoordinator, Toggles: allToggles, Delegating: true})
	if err != nil {
		t.Fatal(err)
	}
	got := entryIDs(r.Entries)
	if !slices.Contains(got, "delegation") || !slices.Contains(got, "model-assignments") {
		t.Errorf("delegar debe traer delegation con model-assignments: %v", got)
	}
	// Sin delegar no figuran en apply.
	r, _ = tc.Resolve(Query{Action: "apply", Role: RoleCoordinator, Toggles: allToggles})
	for _, id := range []string{"delegation", "model-assignments"} {
		if slices.Contains(entryIDs(r.Entries), id) {
			t.Errorf("%s figura sin delegar en apply", id)
		}
	}
	// Sin el toggle de delegation no se delega: no figura aunque se pida.
	r, _ = tc.Resolve(Query{Action: "apply", Role: RoleCoordinator, Toggles: []string{"engram"}, Delegating: true})
	if slices.Contains(entryIDs(r.Entries), "delegation") {
		t.Error("delegation figura con su toggle inactivo")
	}
}

func TestResolve_VerifyWithoutDelegationIsBlocked(t *testing.T) {
	tc := loadTC(t)
	r, err := tc.Resolve(Query{Action: "verify", Role: RoleCoordinator, Toggles: []string{"engram", "tdd"}})
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(r.Blocked, []string{"delegation", "model-assignments"}) {
		t.Errorf("verify sin delegation ni model-routing: Blocked = %v", r.Blocked)
	}
	if !slices.Contains(entryIDs(r.Entries), "bounded-command-output") {
		t.Error("el módulo obligatorio sin toggle debe seguir figurando")
	}
}

func TestResolve_CodexOmitsModelAssignments(t *testing.T) {
	tc := loadTC(t)
	r, err := tc.Resolve(Query{Action: "verify", Role: RoleCoordinator, Toggles: allToggles, Client: "codex"})
	if err != nil {
		t.Fatal(err)
	}
	if slices.Contains(entryIDs(r.Entries), "model-assignments") {
		t.Error("model-assignments se omite en Codex hasta la 09")
	}
	if len(r.Blocked) != 0 {
		t.Errorf("la omisión declarada en Codex no es un bloqueo: %v", r.Blocked)
	}
}

func TestResolve_Specialist(t *testing.T) {
	tc := loadTC(t)
	r, err := tc.Resolve(Query{Action: "apply", Role: RoleSpecialist, Toggles: allToggles, StrictTDD: true})
	if err != nil {
		t.Fatal(err)
	}
	got := entryIDs(r.Entries)
	for _, want := range []string{"artifacts", "tdd", "directed-code-reading", "bounded-command-output", "concise-agent-reporting"} {
		if !slices.Contains(got, want) {
			t.Errorf("el especialista en apply debe recibir %s: %v", want, got)
		}
	}
	for _, not := range []string{"task-tool-selection", "model-assignments", "engram-opsx"} {
		if slices.Contains(got, not) {
			t.Errorf("el especialista no recibe %s: %v", not, got)
		}
	}
	// tdd solo llega al especialista en apply.
	r, _ = tc.Resolve(Query{Action: "verify", Role: RoleSpecialist, Toggles: allToggles, StrictTDD: true})
	if slices.Contains(entryIDs(r.Entries), "tdd") {
		t.Error("tdd llega al especialista solo en apply")
	}
}

func TestResolve_UnknownAction(t *testing.T) {
	if _, err := loadTC(t).Resolve(Query{Action: "deploy", Role: RoleCoordinator}); err == nil {
		t.Error("una acción fuera del ciclo OPSX debe dar error")
	}
}

func TestParseTaskContext_RejectsInvalidModules(t *testing.T) {
	good := mustTaskContextYAML(t)
	tests := []struct {
		name    string
		mutate  func(string) string
		wantErr string
	}{
		{"módulo sin acciones", func(s string) string {
			return strings.Replace(s, "    actions: { propose: mandatory, apply: mandatory, archive: mandatory }\n", "    actions: {}\n", 1)
		}, "no actions"},
		{"nivel desconocido", func(s string) string {
			return strings.Replace(s, "propose: mandatory, apply: mandatory, archive", "propose: sometimes, apply: mandatory, archive", 1)
		}, "sometimes"},
		{"acción fuera del ciclo", func(s string) string {
			return strings.Replace(s, "actions: { verify: mandatory }\n    optional_via: delegate-row\n    specialist: { receives: true", "actions: { deploy: mandatory }\n    optional_via: delegate-row\n    specialist: { receives: true", 1)
		}, "deploy"},
		{"orden sin un módulo", func(s string) string {
			return strings.Replace(s, "  - engram-opsx\n\nmodules:", "\nmodules:", 1)
		}, "module_order"},
		{"módulo duplicado", func(s string) string {
			return strings.Replace(s, "  - id: delegation\n", "  - id: artifacts\n", 1)
		}, "duplicate"},
		{"with a un módulo inexistente", func(s string) string {
			return strings.Replace(s, "    with: delegation", "    with: nada", 1)
		}, "nada"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mutated := tt.mutate(good)
			if mutated == good {
				t.Fatal("la mutación no cambió el YAML: el test no prueba nada")
			}
			_, err := ParseTaskContext([]byte(mutated))
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("err = %v, want mención de %q", err, tt.wantErr)
			}
		})
	}
}

// Los toggles de los módulos salen del sdd-orchestrator del catálogo: un
// toggle que el orquestador no declara se rechaza.
func TestModuleToggles_AreOrchestratorToggles(t *testing.T) {
	c, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	h, ok := c.ByID("sdd-orchestrator")
	if !ok {
		t.Fatal("el catálogo no declara sdd-orchestrator")
	}
	for _, m := range c.TaskContext.Modules {
		if m.Toggle != "" && !slices.Contains(h.Toggles, m.Toggle) {
			t.Errorf("módulo %q usa el toggle %q que sdd-orchestrator no declara (%v)", m.ID, m.Toggle, h.Toggles)
		}
	}
	bad := *c.TaskContext
	bad.Modules = slices.Clone(c.TaskContext.Modules)
	bad.Modules[0].Toggle = "inexistente"
	if err := bad.validateAgainst(h.Toggles); err == nil {
		t.Error("validateAgainst debe rechazar un toggle que el orquestador no declara")
	}
}
