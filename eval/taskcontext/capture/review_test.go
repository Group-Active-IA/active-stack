package capture_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/Group-Active-IA/active-stack/eval/taskcontext/capture"
)

// Regresiones de la revisión del PR #10 (change opsx-task-context): el store
// no sale de su raíz, no adopta destinos ajenos y los metadatos persistidos
// pasan por el redactor. Solo datos ficticios bajo t.TempDir().

func mustSymlink(t *testing.T, target, link string) {
	t.Helper()
	if err := os.Symlink(target, link); err != nil {
		if runtime.GOOS == "windows" {
			t.Skipf("sin privilegio para crear symlinks en Windows (corre en Linux): %v", err)
		}
		t.Fatal(err)
	}
}

func TestReviewStoreCannotPurgeOutsideItsRoot(t *testing.T) {
	bad := []struct{ name, change, run string }{
		{"change con ..", "../../../victim", "run"},
		{"change con separador", "a/b", "run"},
		{"run con ..", "ok", "../../x"},
		{"run vacío", "ok", ""},
		{"change punto", ".", "run"},
		{"run con barra invertida", "ok", `..\x`},
		{"run absoluto", "ok", "/abs"},
	}
	for _, tt := range bad {
		t.Run("Save rechaza "+tt.name, func(t *testing.T) {
			s, _, cache := newStore(t)
			victim := filepath.Join(cache, "active-stack", "victim", "run")
			if err := os.MkdirAll(victim, 0o700); err != nil {
				t.Fatal(err)
			}
			note := filepath.Join(victim, "user-note.txt")
			if err := os.WriteFile(note, []byte("ajeno"), 0o600); err != nil {
				t.Fatal(err)
			}
			if _, err := s.Save(tt.change, tt.run, goodRecord(), []byte("x\n")); err == nil {
				t.Fatal("Save aceptó una identidad que escapa de Root")
			}
			if _, err := os.Stat(note); err != nil {
				t.Errorf("la nota ajena desapareció: %v", err)
			}
		})
	}

	t.Run("Purge ignora claves de índice que escapan", func(t *testing.T) {
		s, c, cache := newStore(t)
		victim := filepath.Join(cache, "active-stack", "victim", "run")
		if err := os.MkdirAll(victim, 0o700); err != nil {
			t.Fatal(err)
		}
		note := filepath.Join(victim, "user-note.txt")
		if err := os.WriteFile(note, []byte("ajeno"), 0o600); err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(s.Root, 0o700); err != nil {
			t.Fatal(err)
		}
		idx := `{"runs":{"../victim/run":{"captured_at":"2026-01-01T00:00:00Z"}},"archived":{}}`
		if err := os.WriteFile(filepath.Join(s.Root, "index.json"), []byte(idx), 0o600); err != nil {
			t.Fatal(err)
		}
		c.advance(400 * day)
		if _, err := s.Purge(); err == nil {
			t.Error("Purge debe informar el índice con claves que escapan")
		}
		if _, err := os.Stat(note); err != nil {
			t.Errorf("Purge borró un archivo fuera de Root: %v", err)
		}
	})

	t.Run("identidades legítimas siguen funcionando", func(t *testing.T) {
		s, _, _ := newStore(t)
		for _, ch := range []string{"", "opsx-task-context", "change_1.v2"} {
			if _, err := s.Save(ch, "20261113-example", goodRecord(), []byte("x\n")); err != nil {
				t.Errorf("Save(%q): %v", ch, err)
			}
		}
	})

	t.Run("Purge no sigue un symlink de directorio fuera de Root", func(t *testing.T) {
		s, c, cache := newStore(t)
		outside := filepath.Join(cache, "outside")
		if err := os.MkdirAll(filepath.Join(outside, "r1"), 0o700); err != nil {
			t.Fatal(err)
		}
		note := filepath.Join(outside, "r1", "user-note.txt")
		if err := os.WriteFile(note, []byte("ajeno"), 0o600); err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(s.Root, 0o700); err != nil {
			t.Fatal(err)
		}
		mustSymlink(t, outside, filepath.Join(s.Root, "_sin-change"))
		if _, err := s.Save("", "r1", goodRecord(), []byte("x\n")); err == nil {
			t.Error("Save debió rechazar el directorio de change que es un symlink")
		}
		c.advance(30 * day)
		_, _ = s.Purge()
		if _, err := os.Stat(note); err != nil {
			t.Errorf("el symlink de directorio permitió tocar fuera de Root: %v", err)
		}
	})
}

func TestReviewStoreMustNotOverwriteSymlinkTarget(t *testing.T) {
	for _, name := range []string{capture.CopyFile, capture.RecordFile} {
		t.Run(name, func(t *testing.T) {
			s, _, _ := newStore(t)
			dir := s.RunDir("", "r1")
			if err := os.MkdirAll(dir, 0o700); err != nil {
				t.Fatal(err)
			}
			victim := filepath.Join(t.TempDir(), "unrelated.txt")
			if err := os.WriteFile(victim, []byte("original"), 0o600); err != nil {
				t.Fatal(err)
			}
			mustSymlink(t, victim, filepath.Join(dir, name))
			if _, err := s.Save("", "r1", goodRecord(), []byte("new copy")); err == nil {
				t.Error("Save aceptó un destino symlink")
			}
			got, _ := os.ReadFile(victim)
			if string(got) != "original" {
				t.Errorf("destino ajeno sobrescrito: %q", got)
			}
		})
	}

	t.Run("destino que es un directorio", func(t *testing.T) {
		s, _, _ := newStore(t)
		if err := os.MkdirAll(filepath.Join(s.RunDir("", "r1"), capture.CopyFile), 0o700); err != nil {
			t.Fatal(err)
		}
		if _, err := s.Save("", "r1", goodRecord(), []byte("x")); err == nil {
			t.Error("Save aceptó un directorio como destino")
		}
	})

	t.Run("index.json symlink", func(t *testing.T) {
		s, _, _ := newStore(t)
		if err := os.MkdirAll(s.Root, 0o700); err != nil {
			t.Fatal(err)
		}
		victim := filepath.Join(t.TempDir(), "unrelated.json")
		if err := os.WriteFile(victim, []byte("original"), 0o600); err != nil {
			t.Fatal(err)
		}
		mustSymlink(t, victim, filepath.Join(s.Root, "index.json"))
		_, _ = s.Save("", "r1", goodRecord(), []byte("x"))
		if got, _ := os.ReadFile(victim); string(got) != "original" {
			t.Errorf("index.json sobrescribió el destino ajeno: %q", got)
		}
	})

	t.Run("un archivo regular propio se reemplaza y no quedan temporales", func(t *testing.T) {
		s, _, _ := newStore(t)
		if _, err := s.Save("", "r1", goodRecord(), []byte("v1")); err != nil {
			t.Fatal(err)
		}
		if _, err := s.Save("", "r1", goodRecord(), []byte("v2")); err != nil {
			t.Fatalf("resguardar de nuevo la misma corrida: %v", err)
		}
		dir := s.RunDir("", "r1")
		if got, _ := os.ReadFile(filepath.Join(dir, capture.CopyFile)); string(got) != "v2" {
			t.Errorf("copia = %q, want v2", got)
		}
		entries, _ := os.ReadDir(dir)
		if len(entries) != 2 {
			t.Errorf("quedan %d archivos en la corrida, want 2 (copia y registro)", len(entries))
		}
	})
}

func TestReviewCommandMetadataMustNotPersistSecrets(t *testing.T) {
	const token = "FICTICIO0123456789abcdefghij"
	cases := []struct{ name, command string }{
		{"bearer", "curl -H 'Authorization: Bearer " + token + "' https://example.invalid"},
		{"token=", "tool --opt token=" + token},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			rec, err := capture.Run(capture.Spec{
				Command: tt.command, Cmd: fixtureCmd(t, "fail-middle"), RunDir: dir, RunID: "run-meta",
				Redactor: newRedactor(t),
				Meta:     capture.Meta{Client: "harness", Interface: "cli", System: "test", Version: "0"},
			})
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(rec.Command, token) {
				t.Errorf("el Record devuelto conserva el secreto: %q", rec.Command)
			}
			data, err := os.ReadFile(filepath.Join(dir, capture.RecordFile))
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(string(data), token) {
				t.Errorf("capture.json persiste el secreto del comando")
			}
			if rec.RunID != "run-meta" || rec.Command == "" {
				t.Errorf("la identidad de la ejecución debe conservarse: %+v", rec)
			}
		})
	}
	t.Run("un comando sin secretos queda igual", func(t *testing.T) {
		dir := t.TempDir()
		rec, err := capture.Run(capture.Spec{
			Command: "go test ./pkg", Cmd: fixtureCmd(t, "fail-middle"), RunDir: dir, RunID: "r",
			Redactor: newRedactor(t),
			Meta:     capture.Meta{Client: "harness", Interface: "cli", System: "test", Version: "0"},
		})
		if err != nil {
			t.Fatal(err)
		}
		if rec.Command != "go test ./pkg" {
			t.Errorf("Command = %q", rec.Command)
		}
	})
}

func TestReviewBareEnvSecretsMustNotPersistInRedactedOutput(t *testing.T) {
	if _, err := exec.LookPath("printf"); err != nil {
		t.Skip("sin printf")
	}
	const value = "FICTICIO-secret-value-0123456789"
	for _, name := range []string{"PASSWORD", "SECRET", "TOKEN", "APP_PASSWORD", "PASS", "KEY", "DBPASSWORD1"} {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			_, err := capture.Run(capture.Spec{
				Command: "cat .env", Cmd: exec.Command("printf", "%s\\n", name+"="+value), RunDir: dir, RunID: "r",
				Redactor: newRedactor(t),
				Meta:     capture.Meta{Client: "harness", Interface: "cli", System: "test", Version: "0"},
			})
			if err != nil {
				t.Fatal(err)
			}
			got := readCopy(t, dir)
			if strings.Contains(got, value) {
				t.Errorf("el valor sensible persiste en output.redacted.txt: %q", got)
			}
			if !strings.Contains(got, name+"=") {
				t.Errorf("la línea debe conservarse con el valor omitido identificado: %q", got)
			}
		})
	}
	t.Run("salida normal no se redacta", func(t *testing.T) {
		for _, line := range []string{"PWD=/tmp/x", "PATH=/usr/bin", "ok=true", "Error=failed", "GOOS=linux", "a_b=1"} {
			dir := t.TempDir()
			if _, err := capture.Run(capture.Spec{
				Command: "env", Cmd: exec.Command("printf", "%s\\n", line), RunDir: dir, RunID: "r",
				Redactor: newRedactor(t),
				Meta:     capture.Meta{Client: "harness", Interface: "cli", System: "test", Version: "0"},
			}); err != nil {
				t.Fatal(err)
			}
			if got := readCopy(t, dir); got != line+"\n" {
				t.Errorf("%q se alteró: %q", line, got)
			}
		}
	})
}
