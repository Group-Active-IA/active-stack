package clientsrc_test

import (
	"strings"
	"testing"

	"github.com/Group-Active-IA/active-stack/eval/taskcontext/clientsrc"
)

// Regresiones de la segunda revisión del PR #10 (F-SRC-02b): la correlación de
// Codex exige identidad y scope suficientes y unicidad del final. Un replay
// idéntico se tolera; identidad ausente repetida, el mismo id en otro
// thread/turn o un final contradictorio se rechazan sin fabricar observación.
// Secuencias JSON sintéticas en memoria.

func evLines(ss ...string) [][]byte {
	out := make([][]byte, len(ss))
	for i, s := range ss {
		out[i] = []byte(s)
	}
	return out
}

func TestReview2AppServerRejectsUnidentifiableMultipleCompletions(t *testing.T) {
	obs, err := clientsrc.FromCodexAppServer(evLines(
		`{"method":"item/completed","params":{"item":{"type":"commandExecution","aggregatedOutput":"FAIL: lost command\n","exitCode":1}}}`,
		`{"method":"item/completed","params":{"item":{"type":"commandExecution","aggregatedOutput":"ok\n","exitCode":0}}}`,
	))
	if err == nil {
		t.Fatalf("dos finales sin identidad aceptados; el último reemplazó al fallo anterior: %+v", obs)
	}
	if obs.Present || obs.ExitCode != nil {
		t.Errorf("con error no se fabrica observación: %+v", obs)
	}
}

func TestReview2AppServerRejectsSameItemIDAcrossThreads(t *testing.T) {
	obs, err := clientsrc.FromCodexAppServer(evLines(
		`{"method":"item/commandExecution/outputDelta","params":{"threadId":"thread-A","turnId":"turn-A","itemId":"item_0","delta":"FAIL: A\n"}}`,
		`{"method":"item/completed","params":{"threadId":"thread-A","turnId":"turn-A","item":{"id":"item_0","type":"commandExecution","exitCode":1}}}`,
		`{"method":"item/completed","params":{"threadId":"thread-B","turnId":"turn-B","item":{"id":"item_0","type":"commandExecution","exitCode":0}}}`,
	))
	if err == nil {
		t.Fatalf("identidades de distintos threads colapsadas; el fallo quedó con exit 0: %+v", obs)
	}
	if !strings.Contains(err.Error(), "thread-B") {
		t.Errorf("el error no nombra el scope en conflicto: %v", err)
	}
}

func TestReview2AppServerRejectsContradictoryDuplicateCompletion(t *testing.T) {
	_, err := clientsrc.FromCodexAppServer(evLines(
		`{"method":"item/completed","params":{"threadId":"t","turnId":"u","item":{"id":"item_0","type":"commandExecution","command":"one","aggregatedOutput":"FAIL: one\n","exitCode":1}}}`,
		`{"method":"item/completed","params":{"threadId":"t","turnId":"u","item":{"id":"item_0","type":"commandExecution","command":"two","aggregatedOutput":"ok\n","exitCode":0}}}`,
	))
	if err == nil {
		t.Fatal("mismo id con final contradictorio reemplazado en silencio")
	}
}

func TestReview2AppServerToleratesIdenticalReplayAndScopedCommand(t *testing.T) {
	done := `{"method":"item/completed","params":{"threadId":"t","turnId":"u","item":{"id":"item_0","type":"commandExecution","command":"go test","aggregatedOutput":"FAIL: x\n","exitCode":1}}}`
	obs, err := clientsrc.FromCodexAppServer(evLines(
		`{"method":"item/commandExecution/outputDelta","params":{"threadId":"t","turnId":"u","itemId":"item_0","delta":"FAIL: x\n"}}`,
		done, done,
	))
	if err != nil {
		t.Fatalf("replay idéntico de un solo comando rechazado: %v", err)
	}
	if !obs.Present || obs.Output != "FAIL: x\n" || obs.ExitCode == nil || *obs.ExitCode != 1 {
		t.Errorf("observación = %+v", obs)
	}
}

func TestReview2ExecRejectsContradictoryDuplicateCompletion(t *testing.T) {
	obs, err := clientsrc.FromCodexExecJSON(evLines(
		`{"type":"item.completed","item":{"id":"item_0","type":"command_execution","command":"one","aggregated_output":"FAIL: one\n","exit_code":1}}`,
		`{"type":"item.completed","item":{"id":"item_0","type":"command_execution","command":"two","aggregated_output":"ok\n","exit_code":0}}`,
	))
	if err == nil {
		t.Fatalf("mismo id con comando/resultado en conflicto reemplazado en silencio: %+v", obs)
	}
	if obs.Present || obs.ExitCode != nil {
		t.Errorf("con error no se fabrica observación: %+v", obs)
	}
}

func TestReview2ExecToleratesIdenticalReplay(t *testing.T) {
	done := `{"type":"item.completed","item":{"id":"item_0","type":"command_execution","command":"go test","aggregated_output":"FAIL: x\n","exit_code":1}}`
	obs, err := clientsrc.FromCodexExecJSON(evLines(done, done))
	if err != nil {
		t.Fatalf("replay idéntico rechazado: %v", err)
	}
	if obs.Output != "FAIL: x\n" || obs.ExitCode == nil || *obs.ExitCode != 1 {
		t.Errorf("observación = %+v", obs)
	}
}
