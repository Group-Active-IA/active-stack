package modelcache

import (
	"testing"
	"testing/fstest"
)

func integer(v int64) *int64 { return &v }

func TestObserveCacheDoesNotInferPrefixFromGlobalTokensOrNewSession(t *testing.T) {
	l, _ := FreezeLayout(fstest.MapFS{"p": {Data: []byte("synthetic prefix")}}, "materialized-prompt", []string{"p"})
	s := CacheObservation{Expected: "cold", Attribution: "aggregate", PrefixFingerprint: l.Fingerprint, ReadTokens: integer(100), EligibleTokens: integer(100), NewSession: true, SignalRef: "synthetic-log"}
	got, err := ObserveCache(l, s)
	if err != nil || got.Observed != "unknown" {
		t.Fatalf("global usage falsely proves a prefix hit: %+v %v", got, err)
	}
	s.Attribution = "prefix"
	got, err = ObserveCache(l, s)
	if err != nil || got.Observed != "full_read" || !got.ColdContaminated {
		t.Fatalf("cold contamination hidden: %+v %v", got, err)
	}
}

func TestObserveCacheRetainsWarmMissPartialAndUnknown(t *testing.T) {
	l, _ := FreezeLayout(fstest.MapFS{"p": {Data: []byte("synthetic")}}, "materialized-prompt", []string{"p"})
	for _, tc := range []struct {
		read *int64
		want string
	}{{integer(0), "miss"}, {integer(20), "partial_read"}, {nil, "unknown"}} {
		r, err := ObserveCache(l, CacheObservation{Expected: "warm", Attribution: "prefix", PrefixFingerprint: l.Fingerprint, ReadTokens: tc.read, EligibleTokens: integer(100), SignalRef: "synthetic-log"})
		if err != nil || r.Observed != tc.want || r.WarmMiss != (tc.want == "miss") {
			t.Fatalf("warm observation rewritten: %+v %v", r, err)
		}
	}
	l.SourceScope = "repo-source"
	r, err := ObserveCache(l, CacheObservation{Expected: "cold", Attribution: "prefix", PrefixFingerprint: l.Fingerprint, ReadTokens: integer(0), EligibleTokens: integer(100), SignalRef: "synthetic-log", NewSession: true})
	if err != nil || r.Observed != "unknown" {
		t.Fatal("repository source falsely proves realized prefix state")
	}
}
