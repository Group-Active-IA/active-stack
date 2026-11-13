package summary

import (
	"github.com/Group-Active-IA/active-stack/eval/baseline/manifest"
	"testing"
)

func TestComputeRejectsTaskOutsideCohort(t *testing.T) {
	c := cohort("c1", manifest.ClientCodex, "0.166.0", manifest.InterfaceTerminal, manifest.ModalityAPI)
	c.Tasks = []manifest.TaskWeight{{TaskID: "planned", Weight: 1}}
	m, p := setup(c)
	if _, err := Compute(m, p, []RunResult{rr("c1", "unplanned", 1, manifest.RunAccepted, 100)}); err == nil {
		t.Fatal("accepted a task outside the cohort")
	}
}
