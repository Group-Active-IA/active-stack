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

// Estado al 2026-11-13: lo medido en la ficha figura con su respaldo.
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
	e, ok := find("codex", "cli", "windows", "aggregated_output")
	if !ok || e.State != clientsrc.StateProbe || e.Backing != "X7" || e.Independent() {
		t.Errorf("Codex CLI Windows aggregated_output = %+v, want sonda X7 sin independencia", e)
	}
	e, ok = find("claude", "cli", "windows", "tool_result")
	if !ok || e.State != clientsrc.StateTruncated {
		t.Errorf("Claude CLI Windows lo que recibe el modelo = %+v, want truncada", e)
	}
	for _, client := range []string{"claude", "codex"} {
		for _, e := range m.Cell(client, "cli", "linux") {
			if e.State != clientsrc.StateNotProbed {
				t.Errorf("%s CLI Linux %q = %q, want no probada (sin cliente nativo en WSL)", client, e.Source, e.State)
			}
		}
	}
	// Ninguna celda figura probada hoy: la tarea 8.1 todavía no corrió.
	for _, e := range m.Entries {
		if e.State == clientsrc.StateProbed {
			t.Errorf("la celda %s · %s · %s (%s) figura probada sin una corrida del grupo 8", e.Client, e.Interface, e.System, e.Source)
		}
	}
}
