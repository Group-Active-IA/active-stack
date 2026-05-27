package pipeline

import "time"

// StepStatus describes the outcome of a single step execution.
type StepStatus string

const (
	StepStatusPending    StepStatus = "pending"
	StepStatusRunning    StepStatus = "running"
	StepStatusSucceeded  StepStatus = "succeeded"
	StepStatusFailed     StepStatus = "failed"
	StepStatusRolledBack StepStatus = "rolled-back"
	StepStatusSkipped    StepStatus = "skipped"
)

// StepResult captures the outcome and timing of a single step.
type StepResult struct {
	StepID     string
	Status     StepStatus
	StartedAt  time.Time
	FinishedAt time.Time
	Err        error
}

// StageResult captures the outcome of running all steps in one stage.
type StageResult struct {
	Stage   Stage
	Steps   []StepResult
	Success bool
	Err     error
}

// ExecutionResult is the top-level result returned by Orchestrator.Execute.
type ExecutionResult struct {
	Prepare  StageResult
	Apply    StageResult
	Rollback StageResult
	Err      error
}
