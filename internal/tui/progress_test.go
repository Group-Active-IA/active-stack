package tui

import (
	"testing"
	"time"

	"github.com/JuanCruzRobledo/jr-stack/internal/pipeline"
)

// TestProgressBridgeDrainsWithoutBlocking verifies that the progress bridge
// channel does not block the caller: buffered events can be sent even when
// the consumer (Bubbletea) has not yet read them.
func TestProgressBridgeDrainsWithoutBlocking(t *testing.T) {
	const bufferSize = 8
	bridge := newProgressBridge(bufferSize)

	events := []pipeline.ProgressEvent{
		{StepID: "snapshot", Stage: pipeline.StagePrepare, Status: pipeline.StepStatusRunning},
		{StepID: "snapshot", Stage: pipeline.StagePrepare, Status: pipeline.StepStatusSucceeded},
		{StepID: "sdd-orchestrator", Stage: pipeline.StageApply, Status: pipeline.StepStatusRunning},
		{StepID: "sdd-orchestrator", Stage: pipeline.StageApply, Status: pipeline.StepStatusSucceeded},
	}

	// Send all events synchronously (simulate pipeline calling OnProgress).
	for _, ev := range events {
		bridge.OnProgress(ev)
	}

	// Drain all events.
	for i, want := range events {
		got := bridge.recv()
		if got.StepID != want.StepID || got.Status != want.Status {
			t.Errorf("event[%d]: got {%s %s}, want {%s %s}",
				i, got.StepID, got.Status, want.StepID, want.Status)
		}
	}
}

// TestProgressBridgeCloseSignalsDone verifies that closing the bridge unblocks
// the consumer (simulated as reading from Done channel).
func TestProgressBridgeCloseSignalsDone(t *testing.T) {
	bridge := newProgressBridge(4)

	done := make(chan struct{})
	go func() {
		bridge.close()
		close(done)
	}()

	select {
	case <-done:
		// expected
	case <-time.After(time.Second):
		t.Fatal("close() timed out — bridge did not unblock")
	}
}

// TestProgressBridgeDoneMsg verifies that doneMsg wraps the ExecutionResult.
func TestProgressBridgeDoneMsg(t *testing.T) {
	result := pipeline.ExecutionResult{
		Err: nil,
	}
	msg := doneMsg{result: result}
	if msg.result.Err != nil {
		t.Errorf("doneMsg.result.Err = %v, want nil", msg.result.Err)
	}
}

// TestProgressMsgCarriesEvent verifies that progressMsg carries ProgressEvent.
func TestProgressMsgCarriesEvent(t *testing.T) {
	ev := pipeline.ProgressEvent{StepID: "foo", Status: pipeline.StepStatusFailed}
	msg := progressMsg{event: ev}
	if msg.event.StepID != "foo" {
		t.Errorf("progressMsg.event.StepID = %q, want %q", msg.event.StepID, "foo")
	}
}
