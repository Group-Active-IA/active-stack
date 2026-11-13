package catalog

import (
	"os"
	"slices"
	"strings"
	"testing"
)

// Regresiones de la revisión del PR #10 (opsx-task-context): un módulo
// obligatorio apagado por toggle detiene la acción en el renderer, en el
// especialista y el YAML no admite un segundo documento.

func without(toggles []string, off string) []string {
	var out []string
	for _, t := range toggles {
		if t != off {
			out = append(out, t)
		}
	}
	return out
}

// rowFor devuelve la fila de la tabla de una acción.
func rowFor(table, action string) string {
	for _, l := range strings.Split(table, "\n") {
		if strings.HasPrefix(l, "| "+action+" |") {
			return l
		}
	}
	return ""
}

func TestReview05RenderKeepsMandatoryBlock(t *testing.T) {
	tc := loadTC(t)
	for _, client := range []string{"claude", "codex"} {
		t.Run(client+"/verify", func(t *testing.T) {
			got, err := tc.RenderActionTableFor(client, without(allToggles, "delegation"))
			if err != nil {
				t.Fatal(err)
			}
			row := rowFor(got, "verify")
			if !strings.Contains(row, "missing module: delegation — verify stopped") {
				t.Errorf("verify sin delegation debe conservar el bloqueo, fila: %q", row)
			}
		})
		t.Run(client+"/apply-strict-tdd", func(t *testing.T) {
			got, err := tc.RenderActionTableFor(client, without(allToggles, "tdd"))
			if err != nil {
				t.Fatal(err)
			}
			row := rowFor(got, "apply")
			if !strings.Contains(row, "Strict TDD") || !strings.Contains(row, "missing module: tdd — apply stopped") {
				t.Errorf("apply sin tdd debe conservar el bloqueo condicional a Strict TDD, fila: %q", row)
			}
		})
		t.Run(client+"/all-active-unchanged", func(t *testing.T) {
			full, _ := tc.RenderActionTable(client)
			got, err := tc.RenderActionTableFor(client, allToggles)
			if err != nil {
				t.Fatal(err)
			}
			if got != full || strings.Contains(got, "STOP") {
				t.Errorf("con todos los toggles activos no debe haber bloqueo ni cambios")
			}
		})
		t.Run(client+"/unrelated-toggle-off", func(t *testing.T) {
			got, err := tc.RenderActionTableFor(client, without(allToggles, "engram"))
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(got, "STOP") {
				t.Errorf("apagar un toggle de módulo no obligatorio no detiene nada")
			}
		})
	}
}

func TestReview05SpecialistPropagatesStrictTDDBlock(t *testing.T) {
	tc := loadTC(t)
	q := Query{Action: "apply", Role: RoleSpecialist, Client: "codex", StrictTDD: true, Toggles: []string{}}
	res, err := tc.Resolve(q)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(res.Blocked, "tdd") {
		t.Errorf("especialista con Strict TDD y tdd apagado debe quedar bloqueado, Blocked=%v", res.Blocked)
	}
	coord := q
	coord.Role = RoleCoordinator
	cres, _ := tc.Resolve(coord)
	if !slices.Equal(cres.Blocked, res.Blocked) {
		t.Errorf("coordinador %v y especialista %v deben coincidir", cres.Blocked, res.Blocked)
	}
	// Triangulación: sin Strict TDD no hay bloqueo; con tdd activo se recibe.
	q.StrictTDD = false
	if res, _ = tc.Resolve(q); len(res.Blocked) != 0 {
		t.Errorf("sin Strict TDD no hay bloqueo, Blocked=%v", res.Blocked)
	}
	q.StrictTDD, q.Toggles = true, []string{"tdd"}
	res, _ = tc.Resolve(q)
	if len(res.Blocked) != 0 || !slices.ContainsFunc(res.Entries, func(e Entry) bool { return e.Module == "tdd" }) {
		t.Errorf("con tdd activo el especialista recibe tdd sin bloqueo: %+v", res)
	}
}

func TestReview05TaskContextRejectsSecondYAMLDocument(t *testing.T) {
	data, err := os.ReadFile("taskcontext.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ParseTaskContext(data); err != nil {
		t.Fatalf("el YAML válido debe parsear: %v", err)
	}
	for name, extra := range map[string]string{
		"unknown":  "\n---\nunknown_second_document: true\n",
		"explicit": "\n---\n{}\n",
	} {
		if _, err := ParseTaskContext(append(slices.Clone(data), extra...)); err == nil {
			t.Errorf("%s: un segundo documento YAML debe rechazarse", name)
		}
	}
}
