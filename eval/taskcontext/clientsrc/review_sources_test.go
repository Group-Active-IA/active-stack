package clientsrc_test

import (
	"testing"

	"github.com/Group-Active-IA/active-stack/eval/taskcontext/clientsrc"
)

// F-SRC-01: solo se quita la nota del cliente (última línea completa); una línea
// real del comando que contiene la misma frase se conserva.
func TestReviewSourcePreservesRealStderrMatchingCwdSuffix(t *testing.T) {
	cases := []struct {
		name, payload, want string
	}{
		{"línea real con la frase", `{"tool_response":{"stdout":"ok\n","stderr":"FAIL: Shell cwd was reset to unexpected-directory; expected project-root\n"}}`,
			"ok\nFAIL: Shell cwd was reset to unexpected-directory; expected project-root\n"},
		{"frase en medio de stderr real", `{"tool_response":{"stdout":"","stderr":"x Shell cwd was reset to a\nmás\n"}}`,
			"x Shell cwd was reset to a\nmás\n"},
		{"nota real tras stderr real", `{"tool_response":{"stdout":"a\n","stderr":"warning: x\nShell cwd was reset to C:\\proj"}}`,
			"a\nwarning: x\n"},
		{"nota real sola", `{"tool_response":{"stdout":"a\n","stderr":"\nShell cwd was reset to C:\\proj\n"}}`, "a\n"},
	}
	for _, c := range cases {
		obs, err := clientsrc.FromClaudePostToolUse([]byte(c.payload))
		if err != nil {
			t.Fatalf("%s: %v", c.name, err)
		}
		if obs.Output != c.want {
			t.Errorf("%s: salida = %q, want %q", c.name, obs.Output, c.want)
		}
	}
}

// F-SRC-03: una respuesta interrumpida no inventa código 0.
func TestReviewInterruptedClaudeDoesNotInventExitZero(t *testing.T) {
	obs, err := clientsrc.FromClaudePostToolUse([]byte(`{"hook_event_name":"PostToolUse","tool_name":"Bash","tool_response":{"stdout":"partial output\n","stderr":"","interrupted":true,"isImage":false}}`))
	if err != nil {
		t.Fatal(err)
	}
	if obs.ExitCode != nil || !obs.Interrupted || !obs.Present || obs.Output != "partial output\n" {
		t.Errorf("interrumpida = %+v, want Present, Interrupted, sin ExitCode", obs)
	}
	obs, _ = clientsrc.FromClaudePostToolUse([]byte(`{"tool_response":{"stdout":"done\n","stderr":"","interrupted":false}}`))
	if obs.ExitCode == nil || *obs.ExitCode != 0 || obs.Interrupted {
		t.Errorf("no interrumpida = %+v, want exit 0 e Interrupted=false", obs)
	}
}

// F-SRC-02 (App Server): salida y exit se correlacionan por itemId; con más de un
// comando la entrada es ambigua y se rechaza.
func TestReviewAppServerDoesNotMixCommandIdentities(t *testing.T) {
	mixed := [][]byte{
		[]byte(`{"method":"item/commandExecution/outputDelta","params":{"itemId":"failed-command","delta":"FAIL: first command\n"}}`),
		[]byte(`{"method":"item/completed","params":{"item":{"type":"commandExecution","id":"failed-command","exitCode":1}}}`),
		[]byte(`{"method":"item/completed","params":{"item":{"type":"commandExecution","id":"successful-command","exitCode":0}}}`),
	}
	if obs, err := clientsrc.FromCodexAppServer(mixed); err == nil {
		t.Errorf("dos comandos deben dar error de entrada ambigua; obtuvo %+v", obs)
	}
	// Un solo comando, con deltas e item final del mismo id.
	one := [][]byte{
		[]byte(`{"method":"item/commandExecution/outputDelta","params":{"itemId":"c1","delta":"a\n"}}`),
		[]byte(`{"method":"item/commandExecution/outputDelta","params":{"itemId":"c1","delta":"b\n"}}`),
		[]byte(`{"method":"item/completed","params":{"item":{"type":"commandExecution","id":"c1","exitCode":1}}}`),
	}
	obs, err := clientsrc.FromCodexAppServer(one)
	if err != nil || obs.Output != "a\nb\n" || obs.ExitCode == nil || *obs.ExitCode != 1 {
		t.Errorf("un comando = %+v, %v; want a\nb\n y exit 1", obs, err)
	}
	// Deltas de un comando y final de otro: ambiguo aunque cada uno sea único.
	cross := [][]byte{one[0], []byte(`{"method":"item/completed","params":{"item":{"type":"commandExecution","id":"c2","exitCode":0}}}`)}
	if _, err := clientsrc.FromCodexAppServer(cross); err == nil {
		t.Error("delta de c1 y final de c2 deben dar error")
	}
}

// F-SRC-02 (exec --json): varios command_execution no se descartan en silencio.
func TestReviewExecDoesNotSilentlyDiscardEarlierCommands(t *testing.T) {
	two := [][]byte{
		[]byte(`{"type":"item.completed","item":{"id":"first","type":"command_execution","aggregated_output":"FAIL: first\n","exit_code":1}}`),
		[]byte(`{"type":"item.completed","item":{"id":"second","type":"command_execution","aggregated_output":"ok\n","exit_code":0}}`),
	}
	if obs, err := clientsrc.FromCodexExecJSON(two); err == nil {
		t.Errorf("dos comandos deben dar error; obtuvo %+v", obs)
	}
	// Dos finales sin id tampoco se pueden distinguir.
	noID := [][]byte{
		[]byte(`{"type":"item.completed","item":{"type":"command_execution","aggregated_output":"a\n","exit_code":1}}`),
		[]byte(`{"type":"item.completed","item":{"type":"command_execution","aggregated_output":"b\n","exit_code":0}}`),
	}
	if _, err := clientsrc.FromCodexExecJSON(noID); err == nil {
		t.Error("dos comandos sin id deben dar error")
	}
	// Un comando y otros items no-comando: sigue aceptado.
	ok := [][]byte{
		[]byte(`{"type":"item.completed","item":{"id":"m","type":"agent_message","text":"hola"}}`),
		two[0],
	}
	obs, err := clientsrc.FromCodexExecJSON(ok)
	if err != nil || obs.Output != "FAIL: first\n" || obs.ExitCode == nil || *obs.ExitCode != 1 {
		t.Errorf("un comando = %+v, %v", obs, err)
	}
}
