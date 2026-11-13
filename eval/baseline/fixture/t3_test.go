package fixture

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// DEUDA 2 (verify-t2-1-auditor): checkout_intacto ve directorios vacios
// agregados y no archivos regulares; las rutas de Source.Dir y Args[0] pasan
// por la guarda.

func TestEvaluate_IntactCheckoutSeesAddedEmptyDirectories(t *testing.T) {
	t.Parallel()
	f, c := prepared(t, FamilyExploration)
	ref, _ := os.ReadFile(f.path(f.Task.Reference))
	for _, tc := range []struct {
		name string
		add  string
		want Outcome
	}{
		{"sin cambios", "", Met},
		{"directorio vacio", "vacio", NotMet},
		{"directorio vacio anidado", filepath.Join("internal", "catalog", "a", "b"), NotMet},
		{"solo directorios vacios anidados", filepath.Join("x", "y", "z"), NotMet},
	} {
		if err := c.Reset(); err != nil {
			t.Fatal(err)
		}
		if tc.add != "" {
			if err := os.MkdirAll(filepath.Join(c.Dir, tc.add), 0o755); err != nil {
				t.Fatal(err)
			}
		}
		v, err := f.Evaluate(Output{Dir: c.Dir, Answer: ref, Decisions: reviewerYes(f)})
		if err != nil {
			t.Fatalf("%s: %v", tc.name, err)
		}
		if got := outcomes(v)["E4"]; got != tc.want {
			t.Errorf("%s: E4 = %q, want %q", tc.name, got, tc.want)
		}
	}
}

type fakeEntry struct {
	fs.DirEntry
	mode fs.FileMode
}

func (e fakeEntry) Type() fs.FileMode { return e.mode }
func (e fakeEntry) IsDir() bool       { return false }

func TestEntryKind_NonRegularIsSentinel(t *testing.T) {
	for _, tc := range []struct {
		name string
		mode fs.FileMode
		bad  bool
	}{
		{"regular", 0, false},
		{"symlink", fs.ModeSymlink, true},
		{"pipe", fs.ModeNamedPipe, true},
		{"socket", fs.ModeSocket, true},
	} {
		err := entryKind(fakeEntry{mode: tc.mode}, "p")
		if tc.bad && !errors.Is(err, errNonRegular) {
			t.Errorf("%s: err = %v, want errNonRegular", tc.name, err)
		}
		if !tc.bad && err != nil {
			t.Errorf("%s: err = %v, want nil", tc.name, err)
		}
	}
}

func TestHashVerdict(t *testing.T) {
	io := errors.New("disco")
	for _, tc := range []struct {
		name     string
		h        string
		err      error
		wantOK   bool
		wantErr  error
		wantText string
	}{
		{"igual", "a", nil, true, nil, ""},
		{"distinta", "b", nil, false, nil, "modificado"},
		{"symlink", "", fmt.Errorf("%w: p", errNonRegular), false, nil, "no regular"},
		{"error de E/S", "", io, false, io, ""},
	} {
		ok, detail, err := hashVerdict(tc.h, tc.err, "a")
		if ok != tc.wantOK || err != tc.wantErr || !strings.Contains(detail, tc.wantText) {
			t.Errorf("%s: (%v, %q, %v), want (%v, %q, %v)", tc.name, ok, detail, err, tc.wantOK, tc.wantText, tc.wantErr)
		}
	}
}

func TestEvaluate_SymlinkInCheckoutIsNotMetNotError(t *testing.T) {
	t.Parallel()
	f, c := prepared(t, FamilyExploration)
	ref, _ := os.ReadFile(f.path(f.Task.Reference))
	if err := c.Reset(); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(c.Dir, "go.mod"), filepath.Join(c.Dir, "enlace")); err != nil {
		t.Skipf("sin privilegio de symlinks en este host: %v", err)
	}
	v, err := f.Evaluate(Output{Dir: c.Dir, Answer: ref, Decisions: reviewerYes(f)})
	if err != nil || outcomes(v)["E4"] != NotMet {
		t.Errorf("symlink: E4 = %v err = %v, want incumplido sin error", outcomes(v)["E4"], err)
	}
}

func TestLoad_RejectsUnsafeSourceDirAndHiddenArgs(t *testing.T) {
	setDir := func(d string) func(map[string]any) {
		return func(m map[string]any) { m["source"].(map[string]any)["dir"] = d }
	}
	setHidden := func(a string) func(map[string]any) {
		return func(m map[string]any) {
			for _, c := range m["criteria"].([]any) {
				cm := c.(map[string]any)
				if cm["check"] == "pruebas_ocultas" {
					cm["args"] = []any{a}
				}
			}
		}
	}
	abs, _ := filepath.Abs(".")
	type tcase struct {
		name string
		edit func(map[string]any)
		want string
	}
	cases := []tcase{
		{"dir con ..", setDir("../fixtures"), "ruta de fixture inválida"},
		{"dir absoluto", setDir(abs), "ruta de fixture inválida"},
		{"dir con .. interno", setDir("project/../../x"), "ruta de fixture inválida"},
		{"dir con barra inicial", setDir("/etc"), "ruta de fixture inválida"},
		{"ocultas con ..", setHidden("../hidden"), "ruta de fixture inválida"},
		{"ocultas absoluto", setHidden(abs), "ruta de fixture inválida"},
		{"ocultas inexistente", setHidden("no-existe"), "no es un directorio"},
		{"ocultas es un archivo", setHidden("task.json"), "no es un directorio"},
	}
	if runtime.GOOS == "windows" { // en Linux estos son nombres validos
		cases = append(cases,
			tcase{"dir con contrabarra inicial", setDir(`\x`), "ruta de fixture inválida"},
			tcase{"dir con volumen relativo", setDir("C:x"), "ruta de fixture inválida"})
	}
	for _, tc := range cases {
		dir := tamper(t, FamilySmallChange, tc.edit, nil)
		_, err := Load(dir)
		if err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%s: Load = %v, want error con %q", tc.name, err, tc.want)
		}
	}
}
