package evidence

import (
	"encoding/hex"
	"fmt"
	"maps"
	"slices"
	"time"

	"github.com/Group-Active-IA/active-stack/eval/taskcontext/capture"
)

// ReceiptFromCapture translates the capture record without asserting provenance
// or sufficiency. The reviewer collects those observations separately.
func ReceiptFromCapture(record capture.Record, tree Tree, at time.Time, requirements []string) (Receipt, error) {
	if err := record.Validate(); err != nil {
		return Receipt{}, err
	}
	digest, err := hex.DecodeString(record.SHA256)
	if err != nil || len(digest) != 32 || at.IsZero() {
		return Receipt{}, fmt.Errorf("capture receipt: invalid output fingerprint or timestamp")
	}
	exit := record.ExitCode
	capturedExit := record.ExitCode
	receipt := Receipt{
		Run: record.RunID, Tree: tree, Command: record.Command,
		Result:       Result{ExitCode: &exit, At: at.UTC().Format(time.RFC3339Nano)},
		Requirements: slices.Clone(requirements),
		Output: Output{
			Fingerprint: &Fingerprint{Algorithm: "sha256", Value: hex.EncodeToString(digest)},
			Capture: &Capture{
				Source: record.Source, Redactor: record.Redactor, Client: record.Client,
				Interface: record.Interface, System: record.System, Version: record.Version,
				ExitCode: &capturedExit, Coverage: record.Coverage, Redactions: maps.Clone(record.Redactions),
			},
		},
	}
	if missing := receipt.Missing(); len(missing) > 0 {
		return Receipt{}, fmt.Errorf("capture receipt: missing %v", missing)
	}
	return receipt, nil
}
