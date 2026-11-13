package evidence

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Valores medidos con informes/medicion-huella/medir-huella.sh en Windows y en Linux (WSL2) con
// core.autocrlf en true, false e input (tarea 1.3): los primeros 16 hex de la huella del estado.
// El limpio es el valor completo que la tarea 4.4 fija.
const cleanState = "1642de00a955de709d3901a362dc9195c1476411c13628a2714e11134b501921"

func gitRepo(t *testing.T, autocrlf string) string {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git no está disponible")
	}
	dir := t.TempDir()
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
		}
	}
	run("init", "-q")
	run("config", "user.email", "t@example.invalid")
	run("config", "user.name", "t")
	run("config", "core.autocrlf", autocrlf)
	run("config", "commit.gpgsign", "false")
	write(t, dir, "f.txt", "a\nb\n")
	write(t, dir, "g.txt", "a\r\nb\r\n")
	run("add", "-A")
	run("commit", "-q", "-m", "base")
	return dir
}

func write(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func gitIn(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}
}

// reset vuelve a los bytes del commit base sin pasar por git checkout, que traduciría los fines de
// línea según core.autocrlf.
func reset(t *testing.T, dir string) {
	t.Helper()
	gitIn(t, dir, "reset", "-q")
	for _, f := range []string{"nuevo.txt", "ignorado.log", ".gitignore"} {
		_ = os.Remove(filepath.Join(dir, f))
	}
	write(t, dir, "f.txt", "a\nb\n")
	write(t, dir, "g.txt", "a\r\nb\r\n")
}

func TestStateFingerprint_RealWorktree(t *testing.T) {
	scenarios := []struct {
		name   string
		mutate func(t *testing.T, dir string)
		prefix string // 16 hex medidos
	}{
		{"limpio", func(*testing.T, string) {}, "1642de00a955de70"},
		{"archivo modificado sin preparar", func(t *testing.T, d string) { write(t, d, "f.txt", "a\nb\nc\n") }, "1e087e8d0a2a92e6"},
		{"mismo HEAD, otro diff preparado", func(t *testing.T, d string) {
			write(t, d, "g.txt", "a\r\nb\r\nc\r\n")
			gitIn(t, d, "add", "g.txt")
		}, "632cbba40a79a7a1"},
		{"archivo nuevo sin trackear", func(t *testing.T, d string) { write(t, d, "nuevo.txt", "nuevo\n") }, "c7631fed78054154"},
		{"archivo borrado", func(t *testing.T, d string) { _ = os.Remove(filepath.Join(d, "g.txt")) }, "e0b0a3701433b552"},
		{"mismo texto con CRLF", func(t *testing.T, d string) { write(t, d, "f.txt", "a\r\nb\r\n") }, "6f46ba11339a6c55"},
		{"archivo ignorado más .gitignore", func(t *testing.T, d string) {
			write(t, d, "ignorado.log", "temporal\n")
			write(t, d, ".gitignore", "*.log\n")
		}, "abf632b30450a14c"},
		{"g.txt con LF", func(t *testing.T, d string) { write(t, d, "g.txt", "a\nb\n") }, "f6cf7032537f4fac"},
	}
	// El valor no depende de core.autocrlf: es la propiedad por la que se descartó el índice temporal.
	for _, ac := range []string{"true", "false", "input"} {
		t.Run("autocrlf="+ac, func(t *testing.T) {
			dir := gitRepo(t, ac)
			seen := map[string]string{}
			for _, s := range scenarios {
				reset(t, dir)
				s.mutate(t, dir)
				got, err := StateFingerprint(dir)
				if err != nil {
					t.Fatalf("%s: %v", s.name, err)
				}
				if got.Algorithm != "sha256" || !strings.HasPrefix(got.Value, s.prefix) {
					t.Errorf("%s: huella %s, esperada prefijo %s", s.name, got.Value, s.prefix)
				}
				if prev, dup := seen[got.Value]; dup {
					t.Errorf("%s y %s dan la misma huella", s.name, prev)
				}
				seen[got.Value] = s.name
			}
		})
	}
}

func TestStateFingerprint_CleanMatchesMeasuredAndTreeHash(t *testing.T) {
	dir := gitRepo(t, "true")
	reset(t, dir)
	got, err := StateFingerprint(dir)
	if err != nil {
		t.Fatal(err)
	}
	if got.Value != cleanState {
		t.Errorf("limpio = %s, medido %s", got.Value, cleanState)
	}
	// Coincide con TreeHash de 02 sobre una copia sin .git.
	if want := treeHashOf(t, map[string]string{"f.txt": "a\nb\n", "g.txt": "a\r\nb\r\n"}); got.Value != want {
		t.Errorf("distinto de TreeHash: %s frente a %s", got.Value, want)
	}
}

func TestStateFingerprint_NotARepository(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git no está disponible")
	}
	if _, err := StateFingerprint(t.TempDir()); err == nil {
		t.Error("un directorio que no es un repositorio debe dar error, no una huella")
	}
}
