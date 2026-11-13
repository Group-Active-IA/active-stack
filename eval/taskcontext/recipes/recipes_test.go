package recipes_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/Group-Active-IA/active-stack/eval/taskcontext/recipes"
)

// Tarea 3.8 (opsx-task-context, D8) y 6.3: las recetas de captura declarada por
// el agente, probadas en cada shell disponible, y su coincidencia byte a byte
// con el módulo bounded-command-output.

const helperEnv = "TASKCONTEXT_RECIPE_HELPER"

// Comando de prueba: escribe una línea por stdout, una por stderr (ambas con
// caracteres no ASCII) y sale con 3. No es 0 ni 1 para distinguirlo del código
// de una tubería o de un envoltorio.
func TestMain(m *testing.M) {
	if os.Getenv(helperEnv) != "" {
		os.Stdout.WriteString("línea stdout ñandú\n")
		os.Stderr.WriteString("línea stderr ¡error!\n")
		os.Exit(3)
	}
	os.Exit(m.Run())
}

const wantOutput = "línea stdout ñandú\nlínea stderr ¡error!\n"

type shell struct {
	name   string
	bin    string
	args   func(script string) []string
	script func(command, f string) string
	quote  func(exe string) string
}

func shells() []shell {
	var out []shell
	if p, err := exec.LookPath("bash"); err == nil {
		out = append(out, shell{
			name: "bash", bin: p,
			args:   func(s string) []string { return []string{"-c", s} },
			script: recipes.BashScript,
			quote:  func(exe string) string { return "'" + filepath.ToSlash(exe) + "'" },
		})
	}
	// En PowerShell el comando va dentro de cmd /c "...": las comillas se duplican.
	psQuote := func(exe string) string { return `""` + exe + `""` }
	psArgs := func(s string) []string { return []string{"-NoProfile", "-NonInteractive", "-Command", s} }
	if p, err := exec.LookPath("powershell"); err == nil && runtime.GOOS == "windows" {
		out = append(out, shell{name: "powershell-5.1", bin: p, args: psArgs, script: recipes.PowerShellScript, quote: psQuote})
	}
	if p, err := exec.LookPath("pwsh"); err == nil {
		out = append(out, shell{name: "pwsh-7", bin: p, args: psArgs, script: recipes.PowerShellScript, quote: psQuote})
	}
	return out
}

func runRecipe(t *testing.T, sh shell) (out []byte, exit string) {
	t.Helper()
	dir := t.TempDir()
	f := filepath.Join(dir, "captura.txt")
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	script := sh.script(sh.quote(exe), f)
	cmd := exec.Command(sh.bin, sh.args(script)...)
	cmd.Env = append(os.Environ(), helperEnv+"=1")
	if combined, err := cmd.CombinedOutput(); err != nil {
		// El script termina con el código del Set-Content / echo; un fallo acá es del shell.
		t.Fatalf("%s: el script falló: %v\n%s", sh.name, err, combined)
	}
	out, err = os.ReadFile(f)
	if err != nil {
		t.Fatalf("%s: no se escribió el archivo de captura: %v", sh.name, err)
	}
	code, err := os.ReadFile(f + ".exit")
	if err != nil {
		t.Fatalf("%s: no se escribió el archivo de código: %v", sh.name, err)
	}
	return out, strings.TrimSpace(string(code))
}

// El archivo queda en bytes del comando, sin BOM ni NativeCommandError, con la
// línea de stderr y el código del comando, no el de una tubería.
func TestRecipes_CaptureIsTheCommandsBytes(t *testing.T) {
	list := shells()
	if len(list) == 0 {
		t.Skip("no hay bash ni PowerShell en esta máquina")
	}
	for _, sh := range list {
		t.Run(sh.name, func(t *testing.T) {
			out, exit := runRecipe(t, sh)
			if string(out) != wantOutput {
				t.Errorf("el archivo no son los bytes del comando:\n got %q\nwant %q", out, wantOutput)
			}
			if bytes.HasPrefix(out, []byte{0xEF, 0xBB, 0xBF}) || bytes.HasPrefix(out, []byte{0xFF, 0xFE}) || bytes.HasPrefix(out, []byte{0xFE, 0xFF}) {
				t.Errorf("el archivo empieza con BOM: % x", out[:4])
			}
			if bytes.Contains(out, []byte("NativeCommandError")) || bytes.Contains(out, []byte("RemoteException")) {
				t.Errorf("el archivo trae el envoltorio de errores de PowerShell:\n%s", out)
			}
			if !strings.Contains(string(out), "línea stderr ¡error!") {
				t.Error("falta la línea de stderr")
			}
			if exit != "exit=3" {
				t.Errorf("código = %q, want exit=3 (el del comando, no el de la tubería)", exit)
			}
		})
	}
}

// Por qué cmd /c: `*>` en Windows PowerShell 5.1 escribe UTF-16 LE con BOM y
// envuelve stderr como NativeCommandError (reproducido por la auditoría del
// 10/11). Si algún día deja de pasar, la receta puede simplificarse.
func TestRecipes_NaivePowerShellRedirectionIsWhyTheRecipeExists(t *testing.T) {
	p, err := exec.LookPath("powershell")
	if err != nil || runtime.GOOS != "windows" {
		t.Skip("Windows PowerShell 5.1 no está disponible")
	}
	dir := t.TempDir()
	f := filepath.Join(dir, "ingenua.txt")
	exe, _ := os.Executable()
	cmd := exec.Command(p, "-NoProfile", "-NonInteractive", "-Command", "& '"+exe+"' *> '"+f+"'")
	cmd.Env = append(os.Environ(), helperEnv+"=1")
	_ = cmd.Run() // sale con 3
	data, err := os.ReadFile(f)
	if err != nil {
		t.Fatal(err)
	}
	badBOM := bytes.HasPrefix(data, []byte{0xFF, 0xFE})
	wrapped := bytes.Contains(data, []byte("NativeCommandError")) || bytes.Contains(data, []byte("RemoteException"))
	utf16 := !bytes.Equal(data, []byte(wantOutput))
	if !badBOM && !wrapped && !utf16 {
		t.Skip("esta versión de PowerShell ya no altera los bytes con `*>`: la receta podría simplificarse")
	}
}

func TestRecipes_ScriptsEmbedTheCommandAndFile(t *testing.T) {
	b := recipes.BashScript("go test ./...", "/tmp/x/out.txt")
	for _, want := range []string{`f='/tmp/x/out.txt'`, `go test ./... > "$f" 2>&1`, `echo "exit=$?" > "$f.exit"`} {
		if !strings.Contains(b, want) {
			t.Errorf("BashScript no contiene %q: %s", want, b)
		}
	}
	p := recipes.PowerShellScript("go test ./...", `C:\tmp\out.txt`)
	for _, want := range []string{`$f='C:\tmp\out.txt'`, `cmd /c "go test ./... > ""$f"" 2>&1"`, `"exit=$LASTEXITCODE" | Set-Content -Encoding ascii "$f.exit"`} {
		if !strings.Contains(p, want) {
			t.Errorf("PowerShellScript no contiene %q: %s", want, p)
		}
	}
}

// 6.3: el módulo bounded-command-output contiene las recetas byte a byte y las
// líneas de forma fija que cuenta el arnés.
func TestModule_BoundedCommandOutput(t *testing.T) {
	path := filepath.Join("..", "..", "..", "internal", "harness", "config", "assets", "modules", "bounded-command-output.md")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("el módulo no existe: %v", err)
	}
	text := string(data)
	if strings.HasPrefix(text, "---") {
		t.Error("el módulo lleva frontmatter de skill")
	}
	for _, recipe := range []string{recipes.Bash, recipes.PowerShell} {
		if !strings.Contains(text, "`"+recipe+"`") {
			t.Errorf("el módulo no contiene la receta byte a byte: %s", recipe)
		}
	}
	for _, fixed := range []string{
		"exit: <code>",
		"failure: <line as captured>",
		"ref: cmd=<command> run=<run-id> path=<redacted copy> sha256:<hex> expires=<YYYY-MM-DD>",
		"capture: <harness | client:<source> | agent> · redactor: <same set> · producer: <who wrote this summary>",
		"status: unverified — <what is missing>",
		"partial — unverified: <reason>",
	} {
		if !strings.Contains(text, fixed) {
			t.Errorf("falta la línea de forma fija %q", fixed)
		}
	}
	// Los motivos de parcial de la forma fija.
	for _, reason := range []string{"never captured", "copy missing", "expired on <date>", "no persistent disk"} {
		if !strings.Contains(text, reason) {
			t.Errorf("falta el motivo de parcial %q", reason)
		}
	}
	// Umbral del contrato: coincide con el catálogo (design.md D3).
	for _, want := range []string{"300 lines", "10,000 characters"} {
		if !strings.Contains(text, want) {
			t.Errorf("el módulo no menciona el umbral %q", want)
		}
	}
}
