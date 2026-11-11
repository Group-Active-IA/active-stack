package manifest

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

// F2 (verify-1-breaker): dimensiones fuera de la primera comparacion o vacias
// y permisos declarados vacios dejan la celda incompatible con motivo.
func TestCheckCell_UnknownOrMissingDimensions(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Cell)
		reason string
	}{
		{"cliente vacio", func(c *Cell) { c.Client = "" }, "cliente"},
		{"cliente fuera de la matriz", func(c *Cell) { c.Client = "gemini" }, "cliente"},
		{"interfaz vacia", func(c *Cell) { c.Interface = "" }, "interfaz"},
		{"escritorio fuera de la primera comparacion", func(c *Cell) { c.Interface = "desktop" }, "escritorio"},
		{"sistema vacio", func(c *Cell) { c.OS = "" }, "sistema"},
		{"sistema fuera de la matriz", func(c *Cell) { c.OS = "darwin" }, "sistema"},
		{"modalidad vacia", func(c *Cell) { c.Modality = "" }, "modalidad"},
		{"modalidad desconocida", func(c *Cell) { c.Modality = "gratis" }, "modalidad"},
		{"permisos declarados vacios", func(c *Cell) {
			c.Permissions = Permissions{}
			c.Observed.Permissions = Permissions{}
		}, "permisos declarados vacíos"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := baseCell()
			tt.mutate(&c)
			got := CheckCell(c)
			if got.Status != CellIncompatible {
				t.Fatalf("status = %q, want incompatible (reasons %v)", got.Status, got.Reasons)
			}
			if len(got.Reasons) != 1 || !strings.Contains(got.Reasons[0], tt.reason) {
				t.Fatalf("reasons = %v, want one containing %q", got.Reasons, tt.reason)
			}
		})
	}
	c := baseCell()
	c.Client, c.OS, c.Interface, c.Modality = ClientCodex, OSWindows, InterfaceIDE, ModalitySubscription
	if got := CheckCell(c); got.Status != CellExecutable {
		t.Errorf("codex/windows/ide/suscripcion = %+v, want ejecutable", got)
	}
}

// F3: opt-ins por familia (OTEL_LOG_* y claves de contenido de [otel] de Codex).
func TestCheckCell_OptInFamily(t *testing.T) {
	tests := []struct {
		name     string
		settings map[string]string
		want     CellStatus
		reasons  int
	}{
		{"prefijo OTEL_LOG_ no listado", map[string]string{"OTEL_LOG_TOOL_CONTENT": "1"}, CellIncompatible, 1},
		{"prefijo en minusculas", map[string]string{"otel_log_user_prompts": "1"}, CellIncompatible, 1},
		{"nombre con espacios", map[string]string{" OTEL_LOG_USER_PROMPTS ": "1"}, CellIncompatible, 1},
		{"Codex forma de tabla", map[string]string{"otel.log_user_prompt": "true"}, CellIncompatible, 1},
		{"Codex tabla en mayusculas", map[string]string{"OTEL.Log_User_Prompt": "true"}, CellIncompatible, 1},
		{"Codex plana con mayusculas", map[string]string{"Log_User_Prompt": "1"}, CellIncompatible, 1},
		{"dos opt-ins, dos motivos", map[string]string{"OTEL_LOG_A": "1", "OTEL_LOG_B": "1"}, CellIncompatible, 2},
		{"apagado con espacios", map[string]string{"OTEL_LOG_TOOL_CONTENT": " off "}, CellExecutable, 0},
		{"variable OTEL sin LOG", map[string]string{"OTEL_EXPORTER_OTLP_ENDPOINT": "http://x"}, CellExecutable, 0},
		{"otra clave de otel", map[string]string{"otel.exporter": "otlp"}, CellExecutable, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := baseCell()
			c.Observed.Settings = tt.settings
			got := CheckCell(c)
			if got.Status != tt.want || len(got.Reasons) != tt.reasons {
				t.Fatalf("CheckCell = %+v, want %q con %d motivos", got, tt.want, tt.reasons)
			}
		})
	}
	// Orden determinista de motivos.
	c := baseCell()
	c.Observed.Settings = map[string]string{"OTEL_LOG_B": "1", "OTEL_LOG_A": "1"}
	for i := 0; i < 20; i++ {
		got := CheckCell(c)
		if len(got.Reasons) != 2 || !strings.Contains(got.Reasons[0], "OTEL_LOG_A") || !strings.Contains(got.Reasons[1], "OTEL_LOG_B") {
			t.Fatalf("motivos fuera de orden: %v", got.Reasons)
		}
	}
}

// F8: la CLI en la terminal integrada del IDE se agrega en la cohorte terminal,
// preservando la interfaz declarada.
func TestPlan_IDEIntegratedTerminalReclassifiedToTerminal(t *testing.T) {
	c := baseCell()
	c.Interface = InterfaceIDE
	c.Observed.CLIInIDETerminal = true
	got := CheckCell(c)
	if got.Status != CellExecutable || got.Interface != InterfaceTerminal {
		t.Fatalf("CheckCell = %+v, want ejecutable en terminal", got)
	}

	m := validManifest()
	m.Cohorts = []Cohort{{ID: "ide-cli", Cell: c, Tasks: []TaskWeight{{TaskID: "t", Weight: 1}}}}
	p := mustPlan(t, m)
	want := []Planned{{CohortID: "ide-cli", Interface: InterfaceTerminal, DeclaredInterface: InterfaceIDE}}
	if !reflect.DeepEqual(p.Executable, want) || len(p.Incompatible) != 0 {
		t.Errorf("Plan = %+v, want %+v", p, want)
	}

	ext := baseCell()
	ext.Interface = InterfaceIDE
	if got := CheckCell(ext); got.Interface != InterfaceIDE {
		t.Errorf("extension real de IDE = %+v, want interfaz ide", got)
	}
}

// F6 y F5: variante obligatoria; task_id vacio o duplicado rechazados; Plan valida.
func TestManifestValidate_VariantAndTaskIDs(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*Manifest)
		wantErr string
	}{
		{"sin variante", func(m *Manifest) { m.Variant = "" }, "variant"},
		{"variante en blanco", func(m *Manifest) { m.Variant = "  " }, "variant"},
		{"task_id vacio", func(m *Manifest) {
			m.Cohorts[0].Tasks = append(m.Cohorts[0].Tasks, TaskWeight{TaskID: "", Weight: 1})
		}, "task_id"},
		{"task_id duplicado", func(m *Manifest) {
			m.Cohorts[0].Tasks = append(m.Cohorts[0].Tasks, TaskWeight{TaskID: "exploracion", Weight: 2})
		}, "duplicad"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := validManifest()
			tt.mutate(&m)
			if err := m.Validate(); err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("Validate() = %v, want error containing %q", err, tt.wantErr)
			}
			if _, err := m.Plan(); err == nil {
				t.Fatal("Plan() debe validar antes de planificar")
			}
		})
	}
	// El mismo task_id en cohortes distintas es valido (validManifest lo usa).
	if err := validManifest().Validate(); err != nil {
		t.Errorf("manifiesto valido: %v", err)
	}
	m := validManifest()
	m.Variant = "modular"
	if err := m.Validate(); err != nil {
		t.Errorf("variante modular: %v", err)
	}
}

// F4: contenido despues del primer objeto JSON.
func TestLoad_RejectsTrailingContent(t *testing.T) {
	data, _ := json.Marshal(validManifest())
	if _, err := Load(strings.NewReader(string(data) + ` {"protocol_version":"OTRO"}`)); err == nil {
		t.Error("Load() debe rechazar un segundo documento")
	}
	if _, err := Load(strings.NewReader(string(data) + ` x`)); err == nil {
		t.Error("Load() debe rechazar basura final")
	}
	if _, err := Load(strings.NewReader(string(data) + "\n\t ")); err != nil {
		t.Errorf("espacios finales deben aceptarse: %v", err)
	}
}

// F7: el rechazo de un campo desconocido se debe a DisallowUnknownFields.
func TestLoad_UnknownFieldOnOtherwiseValidManifest(t *testing.T) {
	data, _ := json.Marshal(validManifest())
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	raw["api_key"] = "secreto"
	withExtra, _ := json.Marshal(raw)
	_, err := Load(strings.NewReader(string(withExtra)))
	if err == nil || !strings.Contains(err.Error(), "unknown field") {
		t.Fatalf("Load() = %v, want error por campo desconocido", err)
	}
}

// F7 (M_neg): un componente negativo de version no se interpreta.
func TestCheckCell_NegativeVersionComponent(t *testing.T) {
	c := baseCell()
	c.Version = "3.-1.0"
	if got := CheckCell(c); got.Status != CellIncompatible {
		t.Errorf("version 3.-1.0 = %+v, want incompatible", got)
	}
}
