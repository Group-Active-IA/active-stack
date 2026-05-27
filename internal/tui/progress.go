package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/JuanCruzRobledo/jr-stack/internal/pipeline"
)

// progressMsg is a Bubbletea message carrying a single ProgressEvent emitted
// by the pipeline runner.
type progressMsg struct {
	event pipeline.ProgressEvent
}

// doneMsg is sent when the orchestrator goroutine finishes (success or failure).
type doneMsg struct {
	result pipeline.ExecutionResult
}

// progressBridge is a buffered channel that decouples the synchronous
// pipeline.ProgressFunc callback from the Bubbletea event loop.
//
// The buffer size should be large enough that the runner goroutine is never
// blocked waiting for the TUI to drain events. A size of 64 covers all
// practical install plans.
type progressBridge struct {
	ch chan pipeline.ProgressEvent
}

// newProgressBridge creates a progressBridge with a buffered channel of size n.
func newProgressBridge(n int) *progressBridge {
	return &progressBridge{ch: make(chan pipeline.ProgressEvent, n)}
}

// OnProgress implements pipeline.ProgressFunc. It performs a non-blocking
// send: if the channel is full the event is dropped (prefer liveness over
// perfect accuracy — the final doneMsg carries the full ExecutionResult).
func (b *progressBridge) OnProgress(ev pipeline.ProgressEvent) {
	select {
	case b.ch <- ev:
	default:
	}
}

// recv reads one event from the channel. It blocks until an event or close.
func (b *progressBridge) recv() pipeline.ProgressEvent {
	return <-b.ch
}

// close closes the underlying channel, unblocking any consumer.
func (b *progressBridge) close() {
	close(b.ch)
}

// listenCmd returns a tea.Cmd that reads one ProgressEvent from the bridge
// and emits it as a progressMsg. The model schedules a new listenCmd after
// each event to drain the channel continuously.
func listenCmd(b *progressBridge) tea.Cmd {
	return func() tea.Msg {
		ev, ok := <-b.ch
		if !ok {
			// Channel closed — no more events.
			return nil
		}
		return progressMsg{event: ev}
	}
}
