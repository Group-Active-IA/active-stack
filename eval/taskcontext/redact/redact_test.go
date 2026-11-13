package redact_test

import (
	"strings"
	"testing"

	"github.com/Group-Active-IA/active-stack/eval/taskcontext/redact"
	"github.com/Group-Active-IA/active-stack/internal/catalog"
)

// Redactor determinista (design.md D8 y D14): reemplaza el valor de cada
// patrón sensible por una marca, conservando la línea en su posición.

func newRedactor(t *testing.T) *redact.Redactor {
	t.Helper()
	cat, err := catalog.Load()
	if err != nil {
		t.Fatal(err)
	}
	r, err := redact.New(cat.TaskContext.SensitivePatterns)
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func TestRedact_EachPattern(t *testing.T) {
	r := newRedactor(t)
	tests := []struct {
		id      string
		line    string
		secret  string // lo que no puede sobrevivir
		keeps   string // lo que debe sobrevivir de la línea
		pattern string
	}{
		{"env-assignment", "DB_PASSWORD=FICTICIO-hunter2-0001", "FICTICIO-hunter2-0001", "DB_PASSWORD=", "env-assignment"},
		{"env-var-value", "printenv API_REGION=FICTICIO-region-9", "FICTICIO-region-9", "printenv API_REGION=", "env-var-value"},
		{"token", "Authorization: Bearer FICTICIOtoken0123456789abcdef", "FICTICIOtoken0123456789abcdef", "Authorization: Bearer ", "token"},
		{"api-key", "api_key=sk-FICTICIO0123456789abcd", "sk-FICTICIO0123456789abcd", "api_key=", "api-key"},
		{"cloud-credential", "aws key AKIAFICTICIO01234567 in config", "AKIAFICTICIO01234567", "aws key ", "cloud-credential"},
		{"cloud-secret", "aws_secret_access_key = FICTICIOsecret0123456789abcdef", "FICTICIOsecret0123456789abcdef", "aws_secret_access_key = ", "cloud-credential"},
		{"home-path-windows", `open C:\Users\FICTICIO_usuario\proyecto\x.go: denied`, "FICTICIO_usuario", `open C:\Users\`, "home-path"},
		{"home-path-unix", "open /home/FICTICIO_usuario/proyecto/x.go: denied", "FICTICIO_usuario", "open /home/", "home-path"},
		{"account-id", "account_id: FICTICIO-acct-4455", "FICTICIO-acct-4455", "account_id: ", "account-id"},
		{"account-uuid", "account_uuid=00000000-0000-4000-8000-0000F1C71C10", "00000000-0000-4000-8000-0000F1C71C10", "account_uuid=", "account-uuid"},
		{"email", "owner FICTICIO.persona@ejemplo.invalid asked", "FICTICIO.persona@ejemplo.invalid", "owner ", "email"},
	}
	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			got, counts := r.RedactText(tt.line + "\n")
			if strings.Contains(got, tt.secret) {
				t.Errorf("el valor sensible sobrevivió: %q", got)
			}
			if !strings.Contains(got, tt.keeps) {
				t.Errorf("la línea perdió su contexto %q: %q", tt.keeps, got)
			}
			if !strings.Contains(got, redact.Mark(tt.pattern)) {
				t.Errorf("falta la marca de omisión %q en %q", redact.Mark(tt.pattern), got)
			}
			if counts[tt.pattern] != 1 {
				t.Errorf("conteo de %s = %d, want 1 (todos: %v)", tt.pattern, counts[tt.pattern], counts)
			}
			if strings.Count(got, "\n") != 1 {
				t.Errorf("la redacción cambió la cantidad de líneas: %q", got)
			}
		})
	}
}

func TestRedact_PrivateKeyBlockKeepsEveryLinePosition(t *testing.T) {
	r := newRedactor(t)
	in := strings.Join([]string{
		"before",
		"-----BEGIN OPENSSH PRIVATE KEY-----",
		"FICTICIOb3BlbnNzaC1rZXktdjEAAAAABG5vbmU",
		"FICTICIOAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
		"-----END OPENSSH PRIVATE KEY-----",
		"after",
	}, "\n") + "\n"
	got, counts := r.RedactText(in)
	if strings.Contains(got, "FICTICIO") {
		t.Errorf("el cuerpo de la clave sobrevivió:\n%s", got)
	}
	if len(strings.Split(got, "\n")) != len(strings.Split(in, "\n")) {
		t.Errorf("cambió la cantidad de líneas:\n%s", got)
	}
	lines := strings.Split(got, "\n")
	if lines[0] != "before" || lines[5] != "after" || lines[1] != "-----BEGIN OPENSSH PRIVATE KEY-----" {
		t.Errorf("las líneas de contexto no quedaron en su posición:\n%s", got)
	}
	if lines[2] != redact.Mark("ssh-private-key") || lines[3] != redact.Mark("ssh-private-key") {
		t.Errorf("cada línea del cuerpo debe quedar como marca:\n%s", got)
	}
	if counts["ssh-private-key"] != 2 {
		t.Errorf("conteo = %v, want 2 líneas de cuerpo", counts)
	}
}

func TestRedact_LeavesCleanTextByteForByte(t *testing.T) {
	r := newRedactor(t)
	in := "ok  \tfixture/pkg\t0.012s\n--- FAIL: TestX (0.00s)\n    x_test.go:12: expected 4, got 5\nexit=1\n"
	got, counts := r.RedactText(in)
	if got != in {
		t.Errorf("texto sin valores sensibles cambió:\n%q\n%q", in, got)
	}
	if len(counts) != 0 {
		t.Errorf("conteos inesperados: %v", counts)
	}
}

func TestRedact_IsIdempotent(t *testing.T) {
	r := newRedactor(t)
	in := "DB_PASSWORD=FICTICIO-x-1234\nowner a@b.invalid\n"
	once, _ := r.RedactText(in)
	twice, counts := r.RedactText(once)
	if once != twice {
		t.Errorf("redactar dos veces cambió el resultado:\n%q\n%q", once, twice)
	}
	if len(counts) != 0 {
		t.Errorf("la segunda pasada no debe encontrar nada: %v", counts)
	}
}
