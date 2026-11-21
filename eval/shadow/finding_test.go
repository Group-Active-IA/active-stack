package shadow

import (
	"strings"
	"testing"
)

// 4.3: formato de hallazgo de D4 {severidad, ruta, rango de líneas, requisito o escenario, descripción}.
func TestParseFindings_Valid(t *testing.T) {
	raw := `[{"severidad":"bloqueante","ruta":"internal/x.go","linea_desde":3,"linea_hasta":7,"requisito":"Requirement: Preservación","descripcion":"pisa la entrada del usuario"}]`
	got, err := ParseFindings([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Severity != "bloqueante" || got[0].From != 3 || got[0].To != 7 {
		t.Fatalf("hallazgo mal leído: %+v", got)
	}
}

func TestParseFindings_Empty(t *testing.T) {
	got, err := ParseFindings([]byte(`[]`))
	if err != nil || len(got) != 0 {
		t.Fatalf("lista vacía: %v %v", got, err)
	}
}

func TestParseFindings_Invalid(t *testing.T) {
	base := `"ruta":"a.go","linea_desde":1,"linea_hasta":2,"requisito":"R","descripcion":"d"`
	for name, raw := range map[string]string{
		"severidad fuera de D4": `[{"severidad":"critica",` + base + `}]`,
		"sin ruta":              `[{"severidad":"mayor","ruta":"","linea_desde":1,"linea_hasta":2,"requisito":"R","descripcion":"d"}]`,
		"rango invertido":       `[{"severidad":"mayor","ruta":"a.go","linea_desde":5,"linea_hasta":2,"requisito":"R","descripcion":"d"}]`,
		"línea cero":            `[{"severidad":"mayor","ruta":"a.go","linea_desde":0,"linea_hasta":2,"requisito":"R","descripcion":"d"}]`,
		"sin requisito":         `[{"severidad":"mayor","ruta":"a.go","linea_desde":1,"linea_hasta":2,"requisito":" ","descripcion":"d"}]`,
		"sin descripción":       `[{"severidad":"mayor","ruta":"a.go","linea_desde":1,"linea_hasta":2,"requisito":"R","descripcion":""}]`,
		"campo desconocido":     `[{"severidad":"mayor",` + base + `,"confianza":0.9}]`,
		"no es lista":           `{"severidad":"mayor"}`,
		"basura al final":       `[] extra`,
	} {
		if _, err := ParseFindings([]byte(raw)); err == nil {
			t.Fatalf("%s: aceptado", name)
		}
	}
}

// 4.3: el prompt es fijo y versionado. Cambiar su texto sin subir la versión rompe este test.
func TestPrompt_FixedAndVersioned(t *testing.T) {
	if PromptVersion != "revision-v1" {
		t.Fatalf("versión = %q", PromptVersion)
	}
	fp := PromptFingerprint()
	const pinned = "35fce05c48468980763782faa8f690800b158b3410c3b1952514c5e5c62a002c"
	if fp.Value != pinned {
		t.Fatalf("el prompt %s cambió (huella %s): subí la versión y fijá la huella nueva", PromptVersion, fp.Value)
	}
}

// El prompt pide el formato de D4 y no trae pistas ni el razonamiento de otros roles.
func TestPrompt_AsksD4AndHasNoClues(t *testing.T) {
	p := Prompt()
	for _, field := range []string{"severidad", "ruta", "linea_desde", "linea_hasta", "requisito", "descripcion", "bloqueante", "mayor", "menor"} {
		if !strings.Contains(p, field) {
			t.Fatalf("el prompt no nombra %q", field)
		}
	}
	for _, clue := range []string{"esperado", "key.json", "etiqueta", "fix"} {
		if strings.Contains(strings.ToLower(p), clue) {
			t.Fatalf("el prompt contiene la pista %q", clue)
		}
	}
}
