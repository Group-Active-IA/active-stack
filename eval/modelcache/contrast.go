package modelcache

type Contrast struct {
	Pilot                string   `json:"pilot"`
	ControlArm           string   `json:"control_arm"`
	CandidateArm         string   `json:"candidate_arm"`
	CohortID             string   `json:"cohort_id"`
	ExpectedCache        string   `json:"expected_cache"`
	Metric               string   `json:"metric"`
	ControlPerAccepted   *float64 `json:"control_per_accepted"`
	CandidatePerAccepted *float64 `json:"candidate_per_accepted"`
	ImprovementFraction  *float64 `json:"improvement_fraction"`
	Status               string   `json:"status"`
	Reasons              []string `json:"reasons"`
}

func contrastResults(p Protocol, rows []Row, pairs []PairResult, runs []Observation) []Contrast {
	arms := map[string]Arm{}
	for _, a := range p.Arms {
		arms[a.ID] = a
	}
	var out []Contrast
	for _, row := range rows {
		a := arms[row.ArmID]
		if a.Factor == "control" {
			continue
		}
		x := Contrast{Pilot: a.Pilot, ControlArm: a.ControlID, CandidateArm: a.ID, CohortID: row.CohortID, ExpectedCache: row.ExpectedCache, Status: "inconclusive"}
		var control *Row
		for i := range rows {
			c := &rows[i]
			if c.ArmID == a.ControlID && c.CohortID == row.CohortID && c.ExpectedCache == row.ExpectedCache && c.MoneyBasis == row.MoneyBasis && c.MoneyUnit == row.MoneyUnit && c.QuotaUnit == row.QuotaUnit {
				control = c
			}
		}
		if control == nil {
			x.Reasons = append(x.Reasons, "matching control row unavailable")
			out = append(out, x)
			continue
		}
		if row.MandatoryRegression || control.MandatoryRegression {
			x.Status = "rejected_quality"
			x.Reasons = append(x.Reasons, "declared mandatory criterion failed")
			out = append(out, x)
			continue
		}
		matched := map[string]bool{}
		candidateCount, controlCount := 0, 0
		allPairs := true
		cacheConfirmed := true
		for _, r := range runs {
			if r.CohortID != row.CohortID || r.Cache.Expected != row.ExpectedCache {
				continue
			}
			if r.ArmID == a.ID || r.ArmID == a.ControlID {
				observed, err := ObserveCache(arms[r.ArmID].Layout, r.Cache)
				cacheConfirmed = cacheConfirmed && err == nil && observed.Observed != "unknown" && !observed.ColdContaminated && observed.Reason == ""
			}
			if r.ArmID == a.ControlID {
				controlCount++
			}
			if r.ArmID != a.ID {
				continue
			}
			candidateCount++
			found := false
			for _, pair := range pairs {
				if pair.CandidateRun == r.ID && pair.Comparison.Comparable && !matched[pair.ControlRun] {
					found = true
					matched[pair.ControlRun] = true
				}
			}
			allPairs = allPairs && found
		}
		if !allPairs || candidateCount == 0 || len(matched) != controlCount {
			x.Reasons = append(x.Reasons, "all initiated control/candidate runs must form comparable pairs; excluded costs remain in report")
			out = append(out, x)
			continue
		}
		if !cacheConfirmed {
			x.Reasons = append(x.Reasons, "cache condition or prefix attribution is unconfirmed; incurred costs remain in report")
			out = append(out, x)
			continue
		}
		d := p.Decisions
		if d == nil || d.Metric == "" {
			x.Reasons = append(x.Reasons, "objective and agreed decision criteria unavailable")
			out = append(out, x)
			continue
		}
		x.Metric = d.Metric
		selectMetric := func(r Row) Metric {
			switch d.Metric {
			case "tokens":
				return r.Tokens.PerAccepted
			case "money":
				if r.Modality == "api" {
					return r.Money.PerAccepted
				}
			case "quota":
				if r.Modality == "subscription" {
					return r.Quota.PerAccepted
				}
			}
			return Metric{Reason: "objective is not available for the modality"}
		}
		left, right := selectMetric(*control), selectMetric(row)
		if left.Value == nil || right.Value == nil || !left.Complete || !right.Complete || *left.Value <= 0 {
			x.Reasons = append(x.Reasons, "objective, accepted denominator or complete cost/preparation coverage unavailable")
			out = append(out, x)
			continue
		}
		x.ControlPerAccepted = left.Value
		x.CandidatePerAccepted = right.Value
		fraction := (*left.Value - *right.Value) / *left.Value
		x.ImprovementFraction = &fraction
		if d.MinimumPairs == nil || d.MinImprovementFraction == nil || d.QualityMargin == nil {
			x.Reasons = append(x.Reasons, "sample, improvement or quality criterion is TBD")
			out = append(out, x)
			continue
		}
		if candidateCount < *d.MinimumPairs || control.Pending > 0 || row.Pending > 0 {
			x.Reasons = append(x.Reasons, "insufficient declared pairs or pending quality results")
			out = append(out, x)
			continue
		}
		if control.Started == 0 || row.Started == 0 || float64(row.Accepted)/float64(row.Started)+*d.QualityMargin < float64(control.Accepted)/float64(control.Started) {
			x.Status = "rejected_quality"
			x.Reasons = append(x.Reasons, "declared acceptance-rate drop exceeds quality margin")
			out = append(out, x)
			continue
		}
		if fraction < *d.MinImprovementFraction {
			x.Status = "below_declared_threshold"
		} else {
			x.Status = "candidate_for_review"
		}
		x.Reasons = append(x.Reasons, "conditional arithmetic only; no verified inference or adoption authorization")
		out = append(out, x)
	}
	return out
}
