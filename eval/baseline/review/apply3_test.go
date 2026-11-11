package review

import (
	"reflect"
	"testing"
)

// N_omOrder (verify-2-breaker): las omisiones se listan en el orden de la
// clave, no por ID.
func TestMatch_OmissionsFollowKeyOrder(t *testing.T) {
	k := []Defect{
		{ID: "D3", File: "a.go", StartLine: 1, EndLine: 2, Mechanism: "m"},
		{ID: "D1", File: "a.go", StartLine: 10, EndLine: 12, Mechanism: "m"},
		{ID: "D2", File: "a.go", StartLine: 20, EndLine: 22, Mechanism: "m"},
		{ID: "D0", File: "a.go", StartLine: 30, EndLine: 32, Mechanism: "m"},
	}
	r, err := Match(k, []Finding{{ID: "F", File: "a.go", StartLine: 10, EndLine: 10, Mechanism: "m"}})
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"D3", "D2", "D0"}; !reflect.DeepEqual(r.Omissions, want) || r.Counts.Omissions != 3 {
		t.Errorf("omissions = %v (%d), want %v en el orden de la clave", r.Omissions, r.Counts.Omissions, want)
	}
}
