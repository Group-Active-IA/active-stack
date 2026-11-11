package manifest

import (
	"strings"
	"testing"
)

func baseCell() Cell {
	return Cell{
		Client:      ClientClaude,
		Version:     "2.1.310",
		MinVersion:  "2.1.310",
		OS:          OSLinux,
		Interface:   InterfaceTerminal,
		Modality:    ModalityAPI,
		Permissions: Permissions{Mode: "acceptEdits", Flags: []string{"--permission-mode", "acceptEdits"}},
		Observed: Observed{
			Permissions: Permissions{Mode: "acceptEdits", Flags: []string{"--permission-mode", "acceptEdits"}},
		},
	}
}

func TestCheckCell_Version(t *testing.T) {
	tests := []struct {
		name       string
		version    string
		min        string
		wantStatus CellStatus
		wantReason string
	}{
		{"menor que la minima", "2.1.309", "2.1.310", CellIncompatible, "versión"},
		{"igual a la minima", "2.1.310", "2.1.310", CellExecutable, ""},
		{"mayor, comparacion numerica no lexicografica", "2.1.1000", "2.1.310", CellExecutable, ""},
		{"menor numericamente aunque lexicografica mayor", "2.1.99", "2.1.310", CellIncompatible, "versión"},
		{"version no interpretable", "latest", "2.1.310", CellIncompatible, "versión"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := baseCell()
			c.Version, c.MinVersion = tt.version, tt.min
			got := CheckCell(c)
			if got.Status != tt.wantStatus {
				t.Fatalf("status = %q, want %q (reasons %v)", got.Status, tt.wantStatus, got.Reasons)
			}
			if tt.wantReason == "" {
				if len(got.Reasons) != 0 {
					t.Fatalf("reasons = %v, want none", got.Reasons)
				}
				return
			}
			if len(got.Reasons) != 1 || !strings.Contains(got.Reasons[0], tt.wantReason) {
				t.Fatalf("reasons = %v, want one containing %q", got.Reasons, tt.wantReason)
			}
		})
	}
}

func TestCheckCell_Permissions(t *testing.T) {
	tests := []struct {
		name     string
		observed Permissions
		want     CellStatus
	}{
		{"mismo modo y flags", Permissions{Mode: "acceptEdits", Flags: []string{"--permission-mode", "acceptEdits"}}, CellExecutable},
		{"flags en otro orden", Permissions{Mode: "acceptEdits", Flags: []string{"acceptEdits", "--permission-mode"}}, CellExecutable},
		{"modo distinto", Permissions{Mode: "default", Flags: []string{"--permission-mode", "acceptEdits"}}, CellIncompatible},
		{"flag extra observado", Permissions{Mode: "acceptEdits", Flags: []string{"--permission-mode", "acceptEdits", "--dangerously-skip-permissions"}}, CellIncompatible},
		{"flag faltante", Permissions{Mode: "acceptEdits"}, CellIncompatible},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := baseCell()
			c.Observed.Permissions = tt.observed
			got := CheckCell(c)
			if got.Status != tt.want {
				t.Fatalf("status = %q, want %q (reasons %v)", got.Status, tt.want, got.Reasons)
			}
			if tt.want == CellIncompatible && (len(got.Reasons) != 1 || !strings.Contains(got.Reasons[0], "permisos")) {
				t.Fatalf("reasons = %v, want one mentioning permisos", got.Reasons)
			}
		})
	}
}

func TestCheckCell_ContentOptIns(t *testing.T) {
	tests := []struct {
		name     string
		settings map[string]string
		want     CellStatus
		reason   string
	}{
		{"sin settings", nil, CellExecutable, ""},
		{"OTEL_LOG_USER_PROMPTS activo", map[string]string{"OTEL_LOG_USER_PROMPTS": "1"}, CellIncompatible, "OTEL_LOG_USER_PROMPTS"},
		{"OTEL_LOG_TOOL_DETAILS activo", map[string]string{"OTEL_LOG_TOOL_DETAILS": "true"}, CellIncompatible, "OTEL_LOG_TOOL_DETAILS"},
		{"OTEL_LOG_RAW_API_BODIES activo", map[string]string{"OTEL_LOG_RAW_API_BODIES": "1"}, CellIncompatible, "OTEL_LOG_RAW_API_BODIES"},
		{"log_user_prompt de Codex activo", map[string]string{"log_user_prompt": "true"}, CellIncompatible, "log_user_prompt"},
		{"explicitamente apagado", map[string]string{"OTEL_LOG_USER_PROMPTS": "0", "log_user_prompt": "false"}, CellExecutable, ""},
		{"valor no reconocido cuenta como activo", map[string]string{"OTEL_LOG_USER_PROMPTS": "maybe"}, CellIncompatible, "OTEL_LOG_USER_PROMPTS"},
		{"variable ajena no afecta", map[string]string{"HOME": "/x"}, CellExecutable, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := baseCell()
			c.Observed.Settings = tt.settings
			got := CheckCell(c)
			if got.Status != tt.want {
				t.Fatalf("status = %q, want %q (reasons %v)", got.Status, tt.want, got.Reasons)
			}
			if tt.reason != "" && (len(got.Reasons) != 1 || !strings.Contains(got.Reasons[0], tt.reason)) {
				t.Fatalf("reasons = %v, want one containing %q", got.Reasons, tt.reason)
			}
		})
	}
}

func TestCheckCell_IDEIntegratedTerminal(t *testing.T) {
	tests := []struct {
		name      string
		iface     Interface
		cliInTerm bool
		want      CellStatus
	}{
		{"declarada IDE pero CLI en terminal integrada: se reclasifica a terminal (apply-2 F8)", InterfaceIDE, true, CellExecutable},
		{"declarada terminal y CLI en terminal integrada", InterfaceTerminal, true, CellExecutable},
		{"IDE con extension real", InterfaceIDE, false, CellExecutable},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := baseCell()
			c.Interface = tt.iface
			c.Observed.CLIInIDETerminal = tt.cliInTerm
			got := CheckCell(c)
			if got.Status != tt.want {
				t.Fatalf("status = %q, want %q (reasons %v)", got.Status, tt.want, got.Reasons)
			}
			if tt.want == CellIncompatible && (len(got.Reasons) != 1 || !strings.Contains(got.Reasons[0], "terminal")) {
				t.Fatalf("reasons = %v, want one mentioning terminal", got.Reasons)
			}
		})
	}
}
