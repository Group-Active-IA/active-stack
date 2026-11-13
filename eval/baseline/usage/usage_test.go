package usage

import (
	"math/rand"
	"reflect"
	"testing"
)

func i64(n int64) *int64 { return &n }

func TestFromClaude_LogicalInputIncludesCacheReadAndCreation(t *testing.T) {
	for _, c := range []struct {
		name string
		in   ClaudeRequest
		want Logical
	}{
		{"entrada nueva, lectura, creación y salida", ClaudeRequest{Input: 100, CacheRead: 5000, CacheCreation: 300, Output: 40},
			Logical{Input: 5400, Uncached: 100, CacheRead: 5000, CacheCreation: 300, Output: 40}},
		{"sin caché", ClaudeRequest{Input: 7, Output: 3}, Logical{Input: 7, Uncached: 7, Output: 3}},
		{"solo lectura", ClaudeRequest{CacheRead: 10}, Logical{Input: 10, CacheRead: 10}},
	} {
		got, err := FromClaude(c.in)
		if err != nil || got != c.want {
			t.Errorf("%s: FromClaude = %+v, %v; want %+v", c.name, got, err, c.want)
		}
	}
	if _, err := FromClaude(ClaudeRequest{Input: -1}); err == nil {
		t.Error("tokens negativos deben ser error")
	}
}

func TestFromCodex_CachedInputIsPartOfInput(t *testing.T) {
	got, err := FromCodex(CodexTurn{Input: 1000, CachedInput: 900, Output: 50, Reasoning: 20})
	want := Logical{Input: 1000, Uncached: 100, CacheRead: 900, Output: 50, Reasoning: 20}
	if err != nil || got != want {
		t.Errorf("FromCodex = %+v, %v; want %+v (entrada sin caché por diferencia, sin sumar la cacheada)", got, err, want)
	}
	got, _ = FromCodex(CodexTurn{Input: 10, Output: 1})
	if got.Input != 10 || got.Uncached != 10 || got.CacheRead != 0 {
		t.Errorf("sin caché: %+v", got)
	}
	if _, err := FromCodex(CodexTurn{Input: 5, CachedInput: 6}); err == nil {
		t.Error("entrada cacheada mayor que la entrada es inconsistente: error")
	}
	if _, err := FromCodex(CodexTurn{Input: -1}); err == nil {
		t.Error("negativos: error")
	}
}

func TestLogical_TotalIsInputPlusOutputWithoutCountingReasoningTwice(t *testing.T) {
	l := Logical{Input: 1000, Output: 50, Reasoning: 20}
	if l.Total() != 1050 {
		t.Errorf("Total = %d, want 1050 (razonamiento ya dentro de la salida, suposición)", l.Total())
	}
}

func TestProbe_DistinguishesRequestInputFromTurnTotal(t *testing.T) {
	req := Logical{Input: 5400, CacheRead: 5000, CacheCreation: 300, Uncached: 100, Output: 40}
	r := Probe(Sample{Request: &req, TurnTotal: i64(20000)})
	if !r.RequestInput.Available || r.RequestInput.Value != 5400 || r.RequestInput.Basis != BasisRequest {
		t.Errorf("entrada de la solicitud = %+v, want 5400 base solicitud", r.RequestInput)
	}
	if !r.TurnTotal.Available || r.TurnTotal.Value != 20000 || r.TurnTotal.Basis != BasisTurn {
		t.Errorf("total del turno = %+v, want 20000 propio, distinto de la entrada de la solicitud", r.TurnTotal)
	}
}

func TestProbe_TurnTotalOnlyIsApproximationOnlyWhenSingleRequestVerified(t *testing.T) {
	r := Probe(Sample{TurnTotal: i64(900), TurnIsSingleRequest: true})
	if !r.RequestInput.Available || r.RequestInput.Value != 900 || r.RequestInput.Basis != BasisApproxTurn {
		t.Errorf("aproximación identificada = %+v", r.RequestInput)
	}
	r = Probe(Sample{TurnTotal: i64(900)})
	if r.RequestInput.Available || r.RequestInput.Reason == "" || r.RequestInput.Value != 0 {
		t.Errorf("total no comprobado como una solicitud = %+v, want no disponible con motivo", r.RequestInput)
	}
	if !r.TurnTotal.Available || r.TurnTotal.Value != 900 {
		t.Errorf("el total del turno sigue disponible: %+v", r.TurnTotal)
	}
}

func TestProbe_AbsentMetricsAreUnavailableNotZero(t *testing.T) {
	r := Probe(Sample{})
	if r.RequestInput.Available || r.TurnTotal.Available || r.RequestInput.Reason == "" || r.TurnTotal.Reason == "" {
		t.Errorf("sin señales = %+v, want ambas no disponibles con motivo", r)
	}
	if r.Cache.Observed != CacheUnknown {
		t.Errorf("caché observada sin solicitud = %q, want %q", r.Cache.Observed, CacheUnknown)
	}
}

func TestProbe_NewSessionWithCachedReadsKeepsObservedState(t *testing.T) {
	req := Logical{Input: 5100, CacheRead: 5000, Uncached: 100}
	r := Probe(Sample{Request: &req, Expected: CacheCold})
	if r.Cache.Observed != CacheRead || r.Cache.Expected != CacheCold || !r.Cache.Diverges {
		t.Errorf("caché = %+v, want observada con_lectura, prevista fría y divergencia registrada", r.Cache)
	}
}

func TestProbe_CacheStates(t *testing.T) {
	for _, c := range []struct {
		name     string
		req      Logical
		expected CacheState
		obs      CacheState
		diverges bool
	}{
		{"creación sin lectura", Logical{Input: 400, CacheCreation: 300, Uncached: 100}, CacheCold, CacheCreate, false},
		{"sin caché y previsto frío", Logical{Input: 50, Uncached: 50}, CacheCold, CacheNone, false},
		{"lectura prevista", Logical{Input: 60, CacheRead: 50, Uncached: 10}, CacheWarm, CacheRead, false},
		{"previsto cálido y no hubo lectura", Logical{Input: 50, Uncached: 50}, CacheWarm, CacheNone, true},
		{"sin previsión no diverge", Logical{Input: 60, CacheRead: 50}, "", CacheRead, false},
	} {
		req := c.req
		r := Probe(Sample{Request: &req, Expected: c.expected})
		if r.Cache.Observed != c.obs || r.Cache.Diverges != c.diverges {
			t.Errorf("%s: caché = %+v, want observada %q diverge %v", c.name, r.Cache, c.obs, c.diverges)
		}
	}
}

// --- Reconciliación (ejemplos del design) ---

func TestReconcile_ClaudeExample(t *testing.T) {
	// SDK informa 60.000 e incluye al worker; no se suma el cierre del worker (8.000).
	got, err := Reconcile(Input{Aggregate: i64(60000), Scope: ScopeIncludesChildren, Entries: []Entry{
		{Role: "principal", ID: "main", Tokens: 40000},
		{Role: "worker", ID: "w1", Tokens: 12000},
		{Role: "worker", ID: "w1", Tokens: 8000},
	}})
	if err != nil {
		t.Fatal(err)
	}
	if got.Total != 60000 || got.ByRole["principal"] != 40000 || got.ByRole["worker"] != 20000 || got.Unattributed != 0 || got.Partial || !got.FullCoverage {
		t.Errorf("Claude = %+v, want total 60000, principal 40000, worker 20000, sin atribuir 0", got)
	}
}

func TestReconcile_CodexExample(t *testing.T) {
	got, err := Reconcile(Input{Aggregate: i64(48000), Scope: ScopeIncludesChildren, Entries: []Entry{
		{Role: "padre", Tokens: 30000}, {Role: "hijo", Tokens: 18000},
	}})
	if err != nil || got.Total != 48000 || got.ByRole["padre"] != 30000 || got.ByRole["hijo"] != 18000 || got.Unattributed != 0 || got.Partial {
		t.Errorf("Codex = %+v, %v; want 48000 (no 66000)", got, err)
	}
}

func TestReconcile_CodexParentOnlyAggregateSumsBothThreads(t *testing.T) {
	got, err := Reconcile(Input{Aggregate: i64(30000), Scope: ScopeParentOnly, Entries: []Entry{
		{Role: "padre", Tokens: 30000}, {Role: "hijo", Tokens: 18000},
	}})
	if err != nil || got.Total != 48000 || got.Partial || !got.FullCoverage {
		t.Errorf("agregado sin hijos = %+v, %v; want suma de ambos hilos 48000", got, err)
	}
}

func TestReconcile_DetailBelowTotalGoesToUnattributedAndPartial(t *testing.T) {
	got, _ := Reconcile(Input{Aggregate: i64(60000), Scope: ScopeIncludesChildren, Entries: []Entry{
		{Role: "principal", Tokens: 40000}, {Role: "worker", Tokens: 12000},
	}})
	if got.Total != 60000 || got.Unattributed != 8000 || !got.Partial || got.FullCoverage {
		t.Errorf("= %+v, want sin atribuir 8000, parcial", got)
	}
}

func TestReconcile_DetailAboveTotalKeepsBothFiguresAndLeavesFullCoverage(t *testing.T) {
	got, _ := Reconcile(Input{Aggregate: i64(50000), Scope: ScopeIncludesChildren, Entries: []Entry{
		{Role: "principal", Tokens: 40000}, {Role: "worker", Tokens: 20000},
	}})
	if !got.Partial || got.FullCoverage || got.AggregateFigure == nil || *got.AggregateFigure != 50000 || got.DetailFigure == nil || *got.DetailFigure != 60000 || got.Unattributed != 0 {
		t.Errorf("= %+v, want parcial, ambas cifras, fuera de cobertura completa", got)
	}
}

func TestReconcile_UnknownScopeOnlyTrustsExactMatch(t *testing.T) {
	ok, _ := Reconcile(Input{Aggregate: i64(100), Scope: ScopeUnknown, Entries: []Entry{{Role: "a", Tokens: 100}}})
	if ok.Partial || !ok.FullCoverage || ok.Total != 100 {
		t.Errorf("coincide: %+v", ok)
	}
	bad, _ := Reconcile(Input{Aggregate: i64(100), Scope: ScopeUnknown, Entries: []Entry{{Role: "a", Tokens: 60}}})
	if !bad.Partial || bad.FullCoverage || bad.AggregateFigure == nil || bad.DetailFigure == nil {
		t.Errorf("alcance no comprobado con diferencia: %+v", bad)
	}
}

func TestReconcile_NoAggregateSumsDetail(t *testing.T) {
	got, err := Reconcile(Input{Entries: []Entry{{Role: "x", Tokens: 3}, {Role: "y", Tokens: 4}}})
	if err != nil || got.Total != 7 || got.Partial || got.AggregateFigure != nil {
		t.Errorf("= %+v, %v", got, err)
	}
	empty, err := Reconcile(Input{})
	if err != nil || empty.Total != 0 {
		t.Errorf("vacío = %+v, %v", empty, err)
	}
}

func TestReconcile_InvalidInputs(t *testing.T) {
	for name, in := range map[string]Input{
		"agregado negativo": {Aggregate: i64(-1), Scope: ScopeIncludesChildren},
		"entrada negativa":  {Entries: []Entry{{Role: "a", Tokens: -1}}},
		"alcance inválido":  {Aggregate: i64(1), Scope: "otro"},
		"rol en blanco":     {Entries: []Entry{{Role: " ", Tokens: 1}}},
	} {
		if _, err := Reconcile(in); err == nil {
			t.Errorf("%s: err = nil", name)
		}
	}
}

func TestReconcile_PropertyEntryOrderDoesNotChangeResult(t *testing.T) {
	rng := rand.New(rand.NewSource(20261111))
	scopes := []Scope{ScopeIncludesChildren, ScopeParentOnly, ScopeUnknown}
	for i := 0; i < 300; i++ {
		var es []Entry
		for j := rng.Intn(6); j > 0; j-- {
			es = append(es, Entry{Role: []string{"p", "w", "h"}[rng.Intn(3)], Tokens: int64(rng.Intn(1000))})
		}
		in := Input{Entries: es, Scope: scopes[rng.Intn(3)]}
		if rng.Intn(2) == 0 {
			in.Aggregate = i64(int64(rng.Intn(3000)))
		}
		a, err := Reconcile(in)
		if err != nil {
			t.Fatal(err)
		}
		rng.Shuffle(len(es), func(x, y int) { es[x], es[y] = es[y], es[x] })
		b, _ := Reconcile(in)
		if !reflect.DeepEqual(a, b) {
			t.Fatalf("orden altera el resultado: %+v vs %+v", a, b)
		}
		if !a.Partial {
			var sum int64
			for _, v := range a.ByRole {
				sum += v
			}
			if sum+a.Unattributed != a.Total {
				t.Fatalf("no cuadra: %+v", a)
			}
		}
	}
}
