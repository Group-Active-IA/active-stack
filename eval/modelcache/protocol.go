package modelcache

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/Group-Active-IA/active-stack/eval/baseline/manifest"
)

type Parameters struct {
	Models map[string]string `json:"models"`
	Effort map[string]string `json:"effort"`
}

type Arm struct {
	ID                string     `json:"id"`
	Pilot             string     `json:"pilot"`
	Factor            string     `json:"factor"`
	Role              string     `json:"role,omitempty"`
	ControlID         string     `json:"control_id,omitempty"`
	Requested         Parameters `json:"requested"`
	Effective         Parameters `json:"effective"`
	Layout            Layout     `json:"layout"`
	CachePolicy       string     `json:"cache_policy"`
	DelegationPolicy  string     `json:"delegation_policy"`
	SemanticReviewRef string     `json:"semantic_review_ref,omitempty"`
}

// Decisions are supplied explicitly after team agreement. Nil stays TBD; these
// numbers never grant provenance, authorize runs or enable automatic adoption.
type Decisions struct {
	Metric                 string   `json:"metric"`
	MinimumPairs           *int     `json:"minimum_pairs"`
	MinImprovementFraction *float64 `json:"min_improvement_fraction"`
	QualityMargin          *float64 `json:"quality_margin"`
	MoneyBudget            *float64 `json:"money_budget"`
	QuotaBudget            *float64 `json:"quota_budget"`
	TimeBudgetMS           *float64 `json:"time_budget_ms"`
}

type Protocol struct {
	Version            string            `json:"version"`
	Phase              string            `json:"phase"`
	BaselineRevision   string            `json:"baseline_revision"`
	DelegationRevision string            `json:"delegation_revision"`
	BaselineStable     bool              `json:"baseline_stable"`
	DelegationResult   string            `json:"delegation_result"`
	Baseline           manifest.Manifest `json:"baseline"`
	Arms               []Arm             `json:"arms"`
	Decisions          *Decisions        `json:"decisions"`
}

func revision(s string) bool { b, e := hex.DecodeString(s); return e == nil && len(b) == 20 }

func (p Protocol) Validate() error {
	if p.Version != "modelcache/v1" {
		return fmt.Errorf("unsupported pilot protocol version")
	}
	if p.Phase != "calibration" {
		return fmt.Errorf("confirmatory/adoption phase unavailable: independent 06/07 assessment and team decisions are not integrated")
	}
	if !revision(p.BaselineRevision) || !revision(p.DelegationRevision) {
		return fmt.Errorf("explicit 02 and 07 source revisions required")
	}
	if err := p.Baseline.Validate(); err != nil {
		return err
	}
	if len(p.Arms) == 0 {
		return fmt.Errorf("no arms declared")
	}
	seen := map[string]Arm{}
	for _, a := range p.Arms {
		if strings.TrimSpace(a.ID) == "" {
			return fmt.Errorf("arm id missing")
		}
		if _, exists := seen[a.ID]; exists {
			return fmt.Errorf("arm id repeated")
		}
		if a.Pilot != "A" && a.Pilot != "B" {
			return fmt.Errorf("unknown pilot for %s", a.ID)
		}
		switch a.Factor {
		case "control", "model", "effort", "prefix", "combined":
		default:
			return fmt.Errorf("unknown intervention for %s", a.ID)
		}
		if a.Pilot == "A" && a.Factor == "prefix" || a.Pilot == "B" && a.Factor != "control" && a.Factor != "prefix" {
			return fmt.Errorf("factor %s does not belong to pilot %s", a.Factor, a.Pilot)
		}
		if a.CachePolicy == "" || a.DelegationPolicy == "" {
			return fmt.Errorf("cache/delegation controls missing for %s", a.ID)
		}
		if err := a.Layout.Validate(); err != nil {
			return fmt.Errorf("arm %s: %w", a.ID, err)
		}
		seen[a.ID] = a
	}
	for _, a := range p.Arms {
		if a.Factor != "control" {
			c, ok := seen[a.ControlID]
			if !ok || c.Factor != "control" || c.Pilot != a.Pilot {
				return fmt.Errorf("arm %s has no control in its own pilot", a.ID)
			}
		}
	}
	if d := p.Decisions; d != nil {
		if d.Metric != "" && d.Metric != "tokens" && d.Metric != "money" && d.Metric != "quota" {
			return fmt.Errorf("unknown objective metric")
		}
		if d.MinimumPairs != nil && *d.MinimumPairs <= 0 {
			return fmt.Errorf("minimum pairs must be positive or TBD")
		}
		for _, v := range []*float64{d.MinImprovementFraction, d.QualityMargin, d.MoneyBudget, d.QuotaBudget, d.TimeBudgetMS} {
			if v != nil && !validNumber(*v) {
				return fmt.Errorf("invalid decision value")
			}
		}
		if d.MinImprovementFraction != nil && *d.MinImprovementFraction > 1 {
			return fmt.Errorf("improvement fraction exceeds one")
		}
		if d.QualityMargin != nil && *d.QualityMargin > 1 {
			return fmt.Errorf("acceptance-rate quality margin exceeds one")
		}
	}
	return nil
}
