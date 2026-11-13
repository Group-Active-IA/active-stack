package modelcache

import (
	"maps"
	"testing"
	"testing/fstest"
)

func candidate(control Arm, factor string) Arm {
	a := control
	a.ID = "candidate"
	a.ControlID = control.ID
	a.Factor = factor
	a.Role = "worker"
	a.Requested = Parameters{Models: maps.Clone(control.Requested.Models), Effort: maps.Clone(control.Requested.Effort)}
	a.Effective = Parameters{Models: maps.Clone(control.Effective.Models), Effort: maps.Clone(control.Effective.Effort)}
	return a
}

func TestCompareArmsPrefixRequiresFixedContentAndReviewReference(t *testing.T) {
	p := syntheticProtocol(t)
	c := p.Arms[0]
	c.Pilot = "B"
	c.SemanticReviewRef = "synthetic-review-reference"
	files := fstest.MapFS{"one": {Data: []byte("one")}, "two": {Data: []byte("two")}}
	c.Layout, _ = FreezeLayout(files, "repo-source", []string{"one", "two"})
	a := candidate(c, "prefix")
	a.Layout, _ = FreezeLayout(files, "repo-source", []string{"two", "one"})
	if got := CompareArms(c, a); !got.Comparable {
		t.Fatalf("conditional reorder rejected: %+v", got)
	}
	a.SemanticReviewRef = ""
	if got := CompareArms(c, a); got.Comparable {
		t.Fatal("missing semantic review reference accepted")
	}
	a.SemanticReviewRef = c.SemanticReviewRef
	a.Effective.Models["worker"] = "fallback"
	if got := CompareArms(c, a); got.Comparable {
		t.Fatal("model change in cache pilot accepted")
	}
}

func TestCompareArmsSeparatesModelAndEffort(t *testing.T) {
	c := syntheticProtocol(t).Arms[0]
	a := candidate(c, "model")
	a.Requested.Models["worker"] = "synthetic-other"
	a.Effective.Models["worker"] = "synthetic-other"
	if got := CompareArms(c, a); !got.Comparable {
		t.Fatalf("single model contrast rejected: %+v", got)
	}
	a.Requested.Effort["worker"] = "other"
	a.Effective.Effort["worker"] = "other"
	if got := CompareArms(c, a); got.Comparable {
		t.Fatal("model+effort falsely isolated")
	}
	a = candidate(c, "effort")
	a.Requested.Effort["worker"] = "synthetic-other-effort"
	a.Effective.Effort["worker"] = "synthetic-other-effort"
	if got := CompareArms(c, a); !got.Comparable {
		t.Fatalf("single effort contrast rejected: %+v", got)
	}
	c.Requested.Models["reviewer"] = "synthetic-reviewer"
	c.Effective.Models["reviewer"] = "synthetic-reviewer"
	a = candidate(c, "model")
	for _, role := range []string{"worker", "reviewer"} {
		a.Requested.Models[role] = "synthetic-changed"
		a.Effective.Models[role] = "synthetic-changed"
	}
	if got := CompareArms(c, a); got.Comparable {
		t.Fatal("two role changes falsely isolated")
	}
}
