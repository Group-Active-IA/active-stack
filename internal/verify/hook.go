package verify

import (
	"context"
	"fmt"

	"github.com/Group-Active-IA/active-stack/internal/model"
)

// BuildHook constructs a func() error suitable for install.Options.VerifyHook.
//
// It assembles a harness-aware check set from harnesses × adapters, runs them,
// builds the report, and returns:
//   - nil  when report.Ready == true (all hard checks passed; warnings are OK).
//   - error when report.Ready == false (at least one hard check failed).
//
// The returned error triggers rollback of the Apply stage in the pipeline.
//
// Import-cycle guard: BuildHook accepts verify.Adapter (a local minimal
// interface) so that internal/verify does NOT import internal/agents.
// Concrete adapters from internal/agents satisfy verify.Adapter by structural
// typing without modification.
func BuildHook(harnesses []model.Harness, adapters []Adapter, homeDir string) func() error {
	return func() error {
		ctx := context.Background()

		// Collect all checks for every (harness × adapter) combination.
		var checks []Check
		for _, h := range harnesses {
			checks = append(checks, ChecksForHarness(h, adapters, homeDir)...)
		}

		results := RunChecks(ctx, checks)
		report := BuildReport(results)

		if !report.Ready {
			rendered := RenderReport(report)
			return fmt.Errorf("post-install verification failed (%d check(s) failed):\n%s",
				report.Failed, rendered)
		}

		return nil
	}
}
