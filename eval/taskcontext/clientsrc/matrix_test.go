package clientsrc_test

import (
	"strings"
	"testing"

	"github.com/Group-Active-IA/active-stack/eval/taskcontext/clientsrc"
)

// Tarea 3.7 (opsx-task-context, D22): matriz de cobertura de captura por
// cliente, interfaz, sistema y fuente.

func TestMatrix_EveryRequiredCellHasAnEntry(t *testing.T) {
	m := clientsrc.DeclaredMatrix()
	for _, client := range []string{"claude", "codex"} {
		for _, iface := range []string{"cli", "ide"} {
			for _, sys := range []string{"windows", "linux"} {
				if len(m.Cell(client, iface, sys)) == 0 {
					t.Errorf("la celda %s · %s · %s no tiene ninguna entrada", client, iface, sys)
				}
			}
		}
	}
	if errs := m.Validate(); len(errs) != 0 {
		t.Errorf("la matriz declarada no valida: %v", errs)
	}
}

func TestMatrix_ProbadaRequiresAGroup8RunID(t *testing.T) {
	base := clientsrc.Entry{Client: "codex", Interface: "cli", System: "windows", Source: "exec --json", Backing: "X7"}
	tests := []struct {
		name    string
		mutate  func(*clientsrc.Entry)
		wantErr string
	}{
		{"probada sin id de corrida", func(e *clientsrc.Entry) { e.State = clientsrc.StateProbed }, "id de corrida"},
		{"probada con un id que no es del grupo 8", func(e *clientsrc.Entry) { e.State = clientsrc.StateProbed; e.RunID = "sonda-X7" }, "grupo 8"},
		{"probada con id de la tarea 8.1", func(e *clientsrc.Entry) { e.State = clientsrc.StateProbed; e.RunID = "8.1/2026-11-20-win-codex" }, ""},
		{"sonda no necesita id y no cuenta como independiente", func(e *clientsrc.Entry) { e.State = clientsrc.StateProbe }, ""},
		{"estado desconocido", func(e *clientsrc.Entry) { e.State = "casi" }, "estado"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := base
			tt.mutate(&e)
			err := e.Validate()
			switch {
			case tt.wantErr == "" && err != nil:
				t.Errorf("rechazó una entrada válida: %v", err)
			case tt.wantErr != "" && (err == nil || !strings.Contains(err.Error(), tt.wantErr)):
				t.Errorf("err = %v, want mención de %q", err, tt.wantErr)
			}
		})
	}
	probed := base
	probed.State, probed.RunID = clientsrc.StateProbed, "8.1/run-1"
	probe := base
	probe.State = clientsrc.StateProbe
	if !probed.Independent() || probe.Independent() {
		t.Error("solo una celda probada con id del grupo 8 cuenta como independiente (D20)")
	}
}

func TestMatrix_CLIOnlySourceIsUnavailableInIDE(t *testing.T) {
	m := clientsrc.DeclaredMatrix()
	var cliOnly []clientsrc.Entry
	for _, e := range m.Entries {
		if e.CLIOnly && e.Interface == "cli" {
			cliOnly = append(cliOnly, e)
		}
	}
	if len(cliOnly) == 0 {
		t.Fatal("la matriz debe declarar al menos una fuente que solo existe en CLI (codex exec --json)")
	}
	for _, c := range cliOnly {
		found := false
		for _, e := range m.Entries {
			if e.Client == c.Client && e.Source == c.Source && e.Interface == "ide" && e.System == c.System {
				found = true
				if e.State != clientsrc.StateUnavailable {
					t.Errorf("la fuente %q solo existe en CLI y en IDE debe figurar no disponible, figura %q", e.Source, e.State)
				}
			}
		}
		if !found {
			t.Errorf("falta la entrada IDE de la fuente solo-CLI %q (%s)", c.Source, c.System)
		}
	}
}

func TestMatrix_ValidateCatchesProblems(t *testing.T) {
	m := clientsrc.Matrix{Entries: []clientsrc.Entry{
		{Client: "claude", Interface: "cli", System: "windows", Source: "hook", State: clientsrc.StateDeclared},
	}}
	errs := m.Validate()
	if len(errs) == 0 {
		t.Fatal("una matriz con celdas exigidas vacías no debe validar")
	}
	joined := strings.Join(errs, "\n")
	for _, want := range []string{"codex · cli · windows", "claude · ide · linux"} {
		if !strings.Contains(joined, want) {
			t.Errorf("falta la mención de la celda %q en:\n%s", want, joined)
		}
	}
}

// Estado al 2026-11-13, después de 8.1: lo medido figura con su corrida o su respaldo.
func TestMatrix_RecordsWhatWasMeasured(t *testing.T) {
	m := clientsrc.DeclaredMatrix()
	find := func(client, iface, sys, source string) (clientsrc.Entry, bool) {
		for _, e := range m.Cell(client, iface, sys) {
			if strings.Contains(e.Source, source) {
				return e, true
			}
		}
		return clientsrc.Entry{}, false
	}
	// Codex CLI: aggregated_output completo con los fixtures, en los dos sistemas.
	for _, sys := range []string{"windows", "linux"} {
		e, ok := find("codex", "cli", sys, "aggregated_output")
		if !ok || e.State != clientsrc.StateProbed || !e.Independent() {
			t.Errorf("Codex CLI %s aggregated_output = %+v, want probada con id de corrida de 8.1", sys, e)
		}
		e, ok = find("claude", "cli", sys, "PostToolUseFailure")
		if !ok || e.State != clientsrc.StateTruncated {
			t.Errorf("Claude CLI %s hook PostToolUseFailure = %+v, want truncada", sys, e)
		}
		e, ok = find("claude", "cli", sys, "tool_result")
		if !ok || e.State != clientsrc.StateTruncated {
			t.Errorf("Claude CLI %s lo que recibe el modelo = %+v, want truncada", sys, e)
		}
	}
	// App Server: completo en Windows; en Linux trae todo pero no conserva el orden.
	e, ok := find("codex", "cli", "windows", "App Server")
	if !ok || e.State != clientsrc.StateProbed || !e.Independent() {
		t.Errorf("Codex App Server Windows = %+v, want probada", e)
	}
	e, ok = find("codex", "cli", "linux", "App Server")
	if !ok || e.State != clientsrc.StateReordered || e.Independent() {
		t.Errorf("Codex App Server Linux = %+v, want reordenada y no independiente", e)
	}
	// Claude IDE y escritorio: declarada por la documentación, no medida.
	e, ok = find("claude", "ide", "windows", "hooks")
	if !ok || e.State != clientsrc.StateDeclared || e.Reason == "" {
		t.Errorf("Claude IDE Windows = %+v, want declarada con motivo", e)
	}
	// IDE en Linux: no hay IDE en el entorno.
	for _, client := range []string{"claude", "codex"} {
		for _, e := range m.Cell(client, "ide", "linux") {
			if e.State == clientsrc.StateProbed {
				t.Errorf("%s IDE Linux %q figura probada sin corrida", client, e.Source)
			}
		}
	}
	// Toda celda probada lleva un id de corrida del grupo 8.
	for _, e := range m.Entries {
		if e.State == clientsrc.StateProbed && !strings.HasPrefix(e.RunID, "8.1") {
			t.Errorf("la celda %s · %s · %s (%s) figura probada sin una corrida de 8.1", e.Client, e.Interface, e.System, e.Source)
		}
	}
}
