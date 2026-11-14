package modelcache

import (
	"encoding/json"
	"os"
	"testing"
)

func TestSyntheticExampleIsOfflineValidAndNeverAuthorizesAdoption(t *testing.T) {
	b, err := os.ReadFile("testdata/synthetic/dataset.json")
	if err != nil {
		t.Fatal(err)
	}
	var d Dataset
	if err = json.Unmarshal(b, &d); err != nil {
		t.Fatal(err)
	}
	if err = ValidateLayouts(d.Protocol, os.DirFS("testdata/synthetic")); err != nil {
		t.Fatal(err)
	}
	r, err := Analyze(d)
	if err != nil {
		t.Fatal(err)
	}
	if d.Protocol.Decisions != nil || len(r.Rows) != 4 || len(r.Pairs) != 2 || r.Adoption.Status != "unavailable" {
		t.Fatalf("synthetic example changed contract: %+v", r)
	}
}
