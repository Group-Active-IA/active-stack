package modelcache

import (
	"maps"
	"reflect"
)

type Comparison struct {
	Comparable bool     `json:"comparable"`
	Reasons    []string `json:"reasons"`
	Basis      string   `json:"basis"`
}

func changedRoles(a, b map[string]string) []string {
	keys := map[string]bool{}
	for k := range a {
		keys[k] = true
	}
	for k := range b {
		keys[k] = true
	}
	var changed []string
	for k := range keys {
		if a[k] != b[k] || a[k] == "" || b[k] == "" {
			changed = append(changed, k)
		}
	}
	return changed
}

func validParameters(p Parameters) bool {
	if len(p.Models) == 0 || len(p.Models) != len(p.Effort) {
		return false
	}
	for role, model := range p.Models {
		if role == "" || model == "" || p.Effort[role] == "" {
			return false
		}
	}
	return true
}

// CompareArms checks declared controls only. A semantic-review reference and
// matching bytes are not corroboration that a real client read that context.
func CompareArms(c, a Arm) Comparison {
	r := Comparison{Comparable: true, Basis: "conditional on declared controls; no independently verified execution"}
	bad := func(reason string) { r.Comparable = false; r.Reasons = append(r.Reasons, reason) }
	if c.Factor != "control" || a.ControlID != c.ID || c.Pilot != a.Pilot {
		bad("control does not belong to the same pilot")
	}
	if !validParameters(c.Effective) || !validParameters(a.Effective) {
		bad("effective role parameters unavailable")
	}
	if !reflect.DeepEqual(c.Requested, c.Effective) || !reflect.DeepEqual(a.Requested, a.Effective) {
		bad("requested/effective parameters differ or fallback occurred")
	}
	if c.CachePolicy != a.CachePolicy || c.DelegationPolicy != a.DelegationPolicy {
		bad("cache or delegation policy changed")
	}
	if c.Layout.SourceScope != a.Layout.SourceScope {
		bad("layout source scope changed")
	}
	models, effort := changedRoles(c.Effective.Models, a.Effective.Models), changedRoles(c.Effective.Effort, a.Effective.Effort)
	switch {
	case a.Pilot == "A" && a.Factor == "model":
		if len(models) != 1 || models[0] != a.Role || len(effort) != 0 {
			bad("model contrast must change one role model only")
		}
		if c.Layout.Fingerprint != a.Layout.Fingerprint {
			bad("model pilot changed frozen layout")
		}
	case a.Pilot == "A" && a.Factor == "effort":
		if len(effort) != 1 || effort[0] != a.Role || len(models) != 0 {
			bad("effort contrast must change one role effort only")
		}
		if c.Layout.Fingerprint != a.Layout.Fingerprint {
			bad("effort pilot changed frozen layout")
		}
	case a.Pilot == "B" && a.Factor == "prefix":
		if len(models) != 0 || len(effort) != 0 {
			bad("prefix pilot changed model or effort")
		}
		left, right := map[string]string{}, map[string]string{}
		for _, f := range c.Layout.Files {
			left[f.Path] = f.Fingerprint.Value
		}
		for _, f := range a.Layout.Files {
			right[f.Path] = f.Fingerprint.Value
		}
		if !maps.Equal(left, right) {
			bad("prefix comparison changed module bytes or membership")
		}
		if c.Layout.Fingerprint == a.Layout.Fingerprint {
			bad("prefix order has not changed")
		}
		if c.SemanticReviewRef == "" || c.SemanticReviewRef != a.SemanticReviewRef {
			bad("semantic equivalence review reference missing or different")
		}
	default:
		bad("combined or unsupported factor is not an isolated contrast")
	}
	return r
}
