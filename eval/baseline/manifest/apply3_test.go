package manifest

import (
	"strings"
	"testing"
)

// R1 (verify-2-breaker): OTEL_LOG_LEVEL regula solo la verbosidad del SDK y no
// es un opt-in de contenido; el resto de la familia OTEL_LOG_* sigue contando.
func TestCheckCell_OTELLogLevelIsNotContentOptIn(t *testing.T) {
	tests := []struct {
		name     string
		settings map[string]string
		want     CellStatus
	}{
		{"OTEL_LOG_LEVEL info", map[string]string{"OTEL_LOG_LEVEL": "info"}, CellExecutable},
		{"otel_log_level debug", map[string]string{"otel_log_level": "debug"}, CellExecutable},
		{"OTEL_LOG_LEVEL con espacios", map[string]string{" Otel_Log_Level ": "warn"}, CellExecutable},
		{"OTEL_LOG_USER_PROMPTS sigue activo", map[string]string{"OTEL_LOG_USER_PROMPTS": "1"}, CellIncompatible},
		{"OTEL_LOG_LEVELS no es la excepcion", map[string]string{"OTEL_LOG_LEVELS": "1"}, CellIncompatible},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := baseCell()
			c.Observed.Settings = tt.settings
			got := CheckCell(c)
			if got.Status != tt.want {
				t.Fatalf("CheckCell = %+v, want %q", got, tt.want)
			}
		})
	}
	c := baseCell()
	c.Observed.Settings = map[string]string{"OTEL_LOG_LEVEL": "info", "OTEL_LOG_USER_PROMPTS": "1"}
	got := CheckCell(c)
	if len(got.Reasons) != 1 || !strings.Contains(got.Reasons[0], "OTEL_LOG_USER_PROMPTS") {
		t.Errorf("motivos = %v, want solo OTEL_LOG_USER_PROMPTS", got.Reasons)
	}
}

// N_modeTrim (verify-2-breaker): un modo declarado solo de espacios es vacio,
// aunque el observado sea igual.
func TestCheckCell_BlankPermissionModeIsEmpty(t *testing.T) {
	for _, mode := range []string{" ", "\t "} {
		c := baseCell()
		c.Permissions = Permissions{Mode: mode}
		c.Observed.Permissions = Permissions{Mode: mode}
		got := CheckCell(c)
		if got.Status != CellIncompatible || len(got.Reasons) != 1 || !strings.Contains(got.Reasons[0], "permisos declarados vacíos") {
			t.Errorf("modo %q: CheckCell = %+v, want incompatible por permisos declarados vacíos", mode, got)
		}
	}
}
