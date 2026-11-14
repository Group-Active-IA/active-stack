// Package modelcache prepares and analyzes offline pilot data. It neither runs
// clients nor treats declarations in an input file as corroborated evidence.
package modelcache

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/Group-Active-IA/active-stack/eval/baseline/usage"
)

// Metric distinguishes an observed zero from an unavailable value. Complete is
// mathematical coverage of the declared inputs, never an evidence verdict.
type Metric struct {
	Value    *float64 `json:"value"`
	Complete bool     `json:"complete"`
	Reason   string   `json:"reason,omitempty"`
}

// Component is one identified non-overlapping amount in a provider's unit.
type Component struct {
	ID    string   `json:"id"`
	Role  string   `json:"role"`
	Value *float64 `json:"value"`
}

// MeasureInput makes the scope of a parent aggregate explicit. For parent-only
// aggregates Components must contain workers only; the parent is not added twice.
// Without an aggregate, Expected lists every participant required for coverage.
type MeasureInput struct {
	Aggregate  *float64    `json:"aggregate"`
	Scope      usage.Scope `json:"scope"`
	Expected   []string    `json:"expected"`
	Components []Component `json:"components"`
}

type Reconciled struct {
	Total  Metric `json:"total"`
	Detail Metric `json:"detail"`
}

func validNumber(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) && v >= 0 }
func metric(v float64, complete bool, reason string) Metric {
	return Metric{Value: &v, Complete: complete, Reason: reason}
}

// ReconcileMeasure applies scope once for money or quota. An inclusive aggregate
// is authoritative; an unknown aggregate scope is always incomplete. No prices
// or token-to-money conversions are inferred.
func ReconcileMeasure(in MeasureInput) (Reconciled, error) {
	expected := map[string]bool{}
	for _, id := range in.Expected {
		if strings.TrimSpace(id) == "" || expected[id] {
			return Reconciled{}, fmt.Errorf("empty or repeated expected component")
		}
		expected[id] = true
	}
	seen := map[string]bool{}
	var values []float64
	complete := len(expected) > 0
	for _, c := range in.Components {
		if !expected[c.ID] || seen[c.ID] {
			return Reconciled{}, fmt.Errorf("unknown or repeated component %s", c.ID)
		}
		seen[c.ID] = true
		if in.Aggregate != nil && in.Scope == usage.ScopeParentOnly && c.Role != "worker" {
			return Reconciled{}, fmt.Errorf("parent-only aggregate accepts worker components only")
		}
		if c.Value == nil {
			complete = false
			continue
		}
		if !validNumber(*c.Value) {
			return Reconciled{}, fmt.Errorf("invalid component amount")
		}
		values = append(values, *c.Value)
	}
	complete = complete && len(seen) == len(expected)
	sort.Float64s(values)
	var detail float64
	for _, v := range values {
		detail += v
	}
	if !validNumber(detail) {
		return Reconciled{}, fmt.Errorf("component sum overflow")
	}
	out := Reconciled{Detail: Metric{Reason: "component amount unavailable"}}
	if len(values) > 0 {
		out.Detail = metric(detail, complete, "")
	}
	if in.Aggregate == nil {
		if in.Scope != "" {
			return Reconciled{}, fmt.Errorf("aggregate scope without aggregate")
		}
		out.Total = out.Detail
		if !complete {
			out.Total.Reason = "missing declared component amounts"
		}
		return out, nil
	}
	if !validNumber(*in.Aggregate) {
		return Reconciled{}, fmt.Errorf("invalid aggregate amount")
	}
	switch in.Scope {
	case usage.ScopeIncludesChildren:
		out.Total = metric(*in.Aggregate, detail <= *in.Aggregate, "")
		if detail > *in.Aggregate {
			out.Total.Reason = "components exceed inclusive aggregate"
		}
	case usage.ScopeParentOnly:
		v := *in.Aggregate + detail
		if !validNumber(v) {
			return Reconciled{}, fmt.Errorf("amount sum overflow")
		}
		// No expected workers means an explicitly parent-only run with no workers.
		coverage := complete || len(expected) == 0 && len(in.Components) == 0
		out.Total = metric(v, coverage, "")
		if !coverage {
			out.Total.Reason = "missing worker amounts"
		}
	case usage.ScopeUnknown:
		out.Total = metric(*in.Aggregate, false, "aggregate scope unknown; not added to components")
	default:
		return Reconciled{}, fmt.Errorf("aggregate scope must be explicit")
	}
	return out, nil
}

type TokenInput struct {
	Aggregate *int64        `json:"aggregate"`
	Scope     usage.Scope   `json:"scope"`
	Expected  []string      `json:"expected"`
	Entries   []usage.Entry `json:"entries"`
}

// ReconcileTokens consumes 02's public reconciliation. The additional participant
// list and explicit scope are coverage declarations, not independent evidence.
func ReconcileTokens(in TokenInput) (Metric, error) {
	if in.Aggregate == nil && in.Scope != "" {
		return Metric{}, fmt.Errorf("token scope without aggregate")
	}
	expected := map[string]bool{}
	for _, id := range in.Expected {
		if strings.TrimSpace(id) == "" || expected[id] {
			return Metric{}, fmt.Errorf("invalid expected token participant")
		}
		expected[id] = true
	}
	seen := map[string]bool{}
	var detail int64
	for _, entry := range in.Entries {
		if !expected[entry.ID] || seen[entry.ID] {
			return Metric{}, fmt.Errorf("unknown or repeated token participant")
		}
		seen[entry.ID] = true
		if entry.Tokens < 0 || entry.Tokens > math.MaxInt64-detail {
			return Metric{}, fmt.Errorf("invalid or overflowing token detail")
		}
		detail += entry.Tokens
	}
	if in.Aggregate == nil && len(in.Entries) == 0 {
		return Metric{Reason: "token usage unavailable"}, nil
	}
	if in.Aggregate != nil && in.Scope == "" {
		return Metric{}, fmt.Errorf("token aggregate scope missing")
	}
	entries := append([]usage.Entry(nil), in.Entries...)
	if in.Aggregate != nil && in.Scope == usage.ScopeParentOnly {
		if *in.Aggregate < 0 || *in.Aggregate > math.MaxInt64-detail {
			return Metric{}, fmt.Errorf("token total overflow")
		}
		for _, entry := range entries {
			if entry.Role != "worker" {
				return Metric{}, fmt.Errorf("parent-only token aggregate accepts worker entries only")
			}
		}
		// 02 expects a complete detail vector in parent-only mode. Adapt our
		// explicit aggregate + worker contract rather than dropping the parent.
		entries = append(entries, usage.Entry{ID: "aggregate-parent", Role: "parent", Tokens: *in.Aggregate})
	}
	total, err := usage.Reconcile(usage.Input{Aggregate: in.Aggregate, Scope: in.Scope, Entries: entries})
	if err != nil {
		return Metric{}, err
	}
	if !validNumber(float64(total.Total)) || int64(float64(total.Total)) != total.Total {
		return Metric{}, fmt.Errorf("token amount cannot be represented exactly")
	}
	complete := total.FullCoverage && total.DetailFigure == nil && len(seen) == len(expected)
	if in.Aggregate != nil && in.Scope == usage.ScopeIncludesChildren && total.DetailFigure == nil {
		complete = true
	}
	if in.Scope == usage.ScopeUnknown {
		complete = false
	}
	reason := ""
	if !complete {
		reason = "token scope or participant coverage incomplete"
	}
	return metric(float64(total.Total), complete, reason), nil
}
