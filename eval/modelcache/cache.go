package modelcache

import (
	"fmt"

	"github.com/Group-Active-IA/active-stack/eval/baseline/record"
)

type CacheObservation struct {
	Expected          string             `json:"expected"`
	Attribution       string             `json:"attribution"`
	PrefixFingerprint record.Fingerprint `json:"prefix_fingerprint"`
	ReadTokens        *int64             `json:"read_tokens"`
	EligibleTokens    *int64             `json:"eligible_tokens"`
	SignalRef         string             `json:"signal_ref"`
	IsolationRef      string             `json:"isolation_ref"`
	NewSession        bool               `json:"new_session"`
}

type CacheResult struct {
	Expected         string `json:"expected"`
	Observed         string `json:"observed"`
	ColdContaminated bool   `json:"cold_contaminated"`
	WarmMiss         bool   `json:"warm_miss"`
	Reason           string `json:"reason,omitempty"`
}

// ObserveCache classifies declared prefix-attributed signals. It never turns a
// repository snapshot, global usage, latency or a new session into prefix proof.
func ObserveCache(l Layout, s CacheObservation) (CacheResult, error) {
	r := CacheResult{Expected: s.Expected, Observed: "unknown"}
	if s.Expected != "" && s.Expected != "cold" && s.Expected != "warm" {
		return r, fmt.Errorf("unknown planned cache condition")
	}
	if s.ReadTokens != nil && *s.ReadTokens < 0 || s.EligibleTokens != nil && *s.EligibleTokens < 0 {
		return r, fmt.Errorf("negative cache observation")
	}
	if l.SourceScope != "materialized-prompt" || s.Attribution != "prefix" || s.SignalRef == "" || s.PrefixFingerprint != l.Fingerprint || s.ReadTokens == nil || s.EligibleTokens == nil || *s.EligibleTokens == 0 {
		r.Reason = "no attributable signal for the declared materialized prefix"
		return r, nil
	}
	if *s.ReadTokens > *s.EligibleTokens {
		return r, fmt.Errorf("prefix read exceeds declared eligible prefix")
	}
	switch {
	case *s.ReadTokens == 0:
		r.Observed = "miss"
	case *s.ReadTokens < *s.EligibleTokens:
		r.Observed = "partial_read"
	default:
		r.Observed = "full_read"
	}
	r.ColdContaminated = s.Expected == "cold" && *s.ReadTokens > 0
	r.WarmMiss = s.Expected == "warm" && *s.ReadTokens == 0
	if s.Expected == "cold" && s.IsolationRef == "" {
		r.Reason = "cold isolation unavailable; a new session does not establish it"
	}
	return r, nil
}
