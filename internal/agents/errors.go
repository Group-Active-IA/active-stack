package agents

import (
	"errors"
	"fmt"

	"github.com/Group-Active-IA/active-stack/internal/model"
)

// ErrAgentNotSupported is the sentinel error returned when the requested
// agent has no registered adapter (e.g. agents not yet in P0).
var ErrAgentNotSupported = errors.New("agent not supported")

// ErrDuplicateAdapter is the sentinel error returned when the same agent is
// registered twice in a Registry.
var ErrDuplicateAdapter = errors.New("adapter already registered")

// AgentNotSupportedError is the typed error wrapping ErrAgentNotSupported.
// Callers can use errors.As to retrieve the unsupported agent id.
type AgentNotSupportedError struct {
	Agent model.Agent
}

func (e AgentNotSupportedError) Error() string {
	return fmt.Sprintf("agent %q is not supported", e.Agent)
}

func (e AgentNotSupportedError) Is(target error) bool {
	return target == ErrAgentNotSupported
}
