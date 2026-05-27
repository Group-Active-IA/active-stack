package pipeline

import "fmt"

// RollbackPolicy controls whether rollback is triggered on stage failure.
type RollbackPolicy struct {
	OnApplyFailure bool
}

// DefaultRollbackPolicy returns a policy that rolls back on any Apply failure.
func DefaultRollbackPolicy() RollbackPolicy {
	return RollbackPolicy{OnApplyFailure: true}
}

// ShouldRollback returns true when the policy mandates rollback for the
// given stage failure.
func (p RollbackPolicy) ShouldRollback(stage Stage, err error) bool {
	if err == nil {
		return false
	}
	return stage == StageApply && p.OnApplyFailure
}

// ExecuteRollback iterates over succeeded Apply steps in reverse order,
// calling Rollback() on each step that implements RollbackStep.
// Steps that do not implement RollbackStep are silently skipped.
func ExecuteRollback(steps []StepResult, stepIndex map[string]Step) StageResult {
	result := StageResult{Stage: StageRollback, Success: true}

	for i := len(steps) - 1; i >= 0; i-- {
		stepResult := steps[i]
		if stepResult.Status != StepStatusSucceeded {
			continue
		}

		step, ok := stepIndex[stepResult.StepID]
		if !ok {
			continue
		}

		rollbackStep, ok := step.(RollbackStep)
		if !ok {
			// Not a rollback-capable step — skip without error.
			continue
		}

		err := rollbackStep.Rollback()
		item := StepResult{StepID: rollbackStep.ID(), Status: StepStatusRolledBack}
		if err != nil {
			item.Status = StepStatusFailed
			item.Err = err
			result.Steps = append(result.Steps, item)
			result.Success = false
			result.Err = fmt.Errorf("rollback step %q: %w", rollbackStep.ID(), err)
			return result
		}

		result.Steps = append(result.Steps, item)
	}

	return result
}
