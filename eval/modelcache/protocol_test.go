package modelcache

import (
	"strings"
	"testing"
	"testing/fstest"

	"github.com/Group-Active-IA/active-stack/eval/baseline/manifest"
)

func syntheticProtocol(t *testing.T) Protocol {
	t.Helper()
	l, err := FreezeLayout(fstest.MapFS{"rules.md": {Data: []byte("synthetic rules\n")}}, "repo-source", []string{"rules.md"})
	if err != nil {
		t.Fatal(err)
	}
	params := Parameters{Models: map[string]string{"worker": "synthetic-control"}, Effort: map[string]string{"worker": "synthetic-level"}}
	return Protocol{Version: "modelcache/v1", Phase: "calibration", BaselineRevision: strings.Repeat("a", 40), DelegationRevision: strings.Repeat("b", 40),
		Baseline: manifest.Manifest{ProtocolVersion: "synthetic-02", RubricVersion: "synthetic-rubric", FixtureRevision: "synthetic-fixture", Variant: "synthetic", Cohorts: []manifest.Cohort{{ID: "c", Cell: manifest.Cell{Client: manifest.ClientCodex, Version: "1.0.0", MinVersion: "1.0.0", OS: manifest.OSLinux, Interface: manifest.InterfaceTerminal, Modality: manifest.ModalityAPI, Permissions: manifest.Permissions{Mode: "synthetic"}, Observed: manifest.Observed{Permissions: manifest.Permissions{Mode: "synthetic"}}}, Tasks: []manifest.TaskWeight{{TaskID: "task", Weight: 1}}}}},
		Arms:     []Arm{{ID: "A0", Pilot: "A", Factor: "control", Requested: params, Effective: params, Layout: l, CachePolicy: "synthetic-cold", DelegationPolicy: "synthetic-direct"}},
	}
}

func TestProtocolCalibrationKeepsPendingDecisionsAndBlocksConfirmation(t *testing.T) {
	p := syntheticProtocol(t)
	if err := p.Validate(); err != nil {
		t.Fatal(err)
	}
	if p.Decisions != nil {
		t.Fatal("TBD decisions replaced by defaults")
	}
	p.Phase = "confirmatory"
	if err := p.Validate(); err == nil {
		t.Fatal("confirmation accepted without external evidence and agreements")
	}
}

func TestProtocolRejectsCrossPilotFactorAndMissingRevision(t *testing.T) {
	for _, tc := range []struct{ pilot, factor string }{{"A", "prefix"}, {"B", "model"}} {
		p := syntheticProtocol(t)
		p.Arms[0].Pilot = tc.pilot
		a := candidate(p.Arms[0], tc.factor)
		p.Arms = append(p.Arms, a)
		if err := p.Validate(); err == nil {
			t.Fatalf("accepted %s factor in pilot %s", tc.factor, tc.pilot)
		}
	}
	p := syntheticProtocol(t)
	p.BaselineRevision = ""
	if err := p.Validate(); err == nil {
		t.Fatal("missing source revision accepted")
	}
}
