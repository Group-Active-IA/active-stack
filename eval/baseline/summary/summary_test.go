package summary

import (
	"math"
	"math/rand"
	"reflect"
	"strings"
	"testing"

	"github.com/Group-Active-IA/active-stack/eval/baseline/manifest"
	"github.com/Group-Active-IA/active-stack/eval/baseline/run"
	"github.com/Group-Active-IA/active-stack/eval/baseline/usage"
)

func tok(v float64) Amount   { return Amount{Available: true, Value: v} }
func gone(why string) Amount { return Amount{Reason: why} }
func tokCost(v float64) Cost { return Cost{Tokens: tok(v), Money: gone("n/a"), Quota: gone("n/a")} }
func costs(vs ...float64) []Cost {
	var out []Cost
	for _, v := range vs {
		out = append(out, tokCost(v))
	}
	return out
}

func cohort(id string, cl manifest.Client, ver string, ifc manifest.Interface, mod manifest.Modality) manifest.Cohort {
	return manifest.Cohort{ID: id, Cell: manifest.Cell{Client: cl, Version: ver, Interface: ifc, Modality: mod}}
}

func setup(cs ...manifest.Cohort) (manifest.Manifest, manifest.Plan) {
	m := manifest.Manifest{ProtocolVersion: "1", Variant: "actual", Cohorts: cs}
	var p manifest.Plan
	for _, c := range cs {
		p.Executable = append(p.Executable, manifest.Planned{CohortID: c.ID, Interface: c.Cell.Interface, DeclaredInterface: c.Cell.Interface})
	}
	return m, p
}

func rr(c, task string, rep int, st manifest.RunStatus, cs ...float64) RunResult {
	return RunResult{CohortID: c, TaskID: task, Repetition: rep, Status: st, Attempts: costs(cs...)}
}

func one(t *testing.T, m manifest.Manifest, p manifest.Plan, runs []RunResult) CohortSummary {
	t.Helper()
	r, err := Compute(m, p, runs)
	if err != nil || len(r.Cohorts) != 1 {
		t.Fatalf("Compute = %+v, %v", r, err)
	}
	return r.Cohorts[0]
}

func TestCompute_RejectionsBlocksTimeoutsAndRetriesAddToTotalCost(t *testing.T) {
	m, p := setup(cohort("c1", manifest.ClientClaude, "2.1.310", manifest.InterfaceTerminal, manifest.ModalityAPI))
	s := one(t, m, p, []RunResult{
		rr("c1", "t1", 1, manifest.RunAccepted, 100),
		rr("c1", "t2", 1, manifest.RunRejectedQuality, 200),
		rr("c1", "t3", 1, manifest.RunBlockedPermissions, 50),
		rr("c1", "t4", 1, manifest.RunTimeout, 30),
		rr("c1", "t5", 1, manifest.RunAccepted, 60, 40), // el primer intento falló y se reintentó
		rr("c1", "t6", 1, manifest.RunClientFailure, 10),
		rr("c1", "t7", 1, run.StatusInterrupted, 5),
	})
	if s.Started != 7 || s.Accepted != 2 || s.Rejected != 1 || s.Blocked != 1 || s.TimedOut != 1 || s.ClientFailed != 1 || s.Interrupted != 1 || s.Pending != 0 || s.Attempts != 8 {
		t.Errorf("conteos = %+v", s)
	}
	// 100+200+50+30+60+40+10+5 = 495 sobre 2 aceptadas
	if !s.Tokens.Total.Available || s.Tokens.Total.Value != 495 {
		t.Errorf("costo total = %+v, want 495 con todo lo iniciado", s.Tokens.Total)
	}
	if !s.Tokens.Complete || !s.Tokens.PerAccepted.Available || s.Tokens.PerAccepted.Value != 247.5 {
		t.Errorf("costo por aceptada = %+v, want 247.5", s.Tokens.PerAccepted)
	}
	if s.Provisional {
		t.Error("sin pendientes no es provisional")
	}
}

func TestCompute_IncompatibleCellsAreOutsideTheDenominatorAndListedApart(t *testing.T) {
	m, p := setup(cohort("c1", manifest.ClientClaude, "2.1.310", manifest.InterfaceTerminal, manifest.ModalityAPI))
	m.Cohorts = append(m.Cohorts, cohort("c2", manifest.ClientCodex, "0.1", manifest.InterfaceIDE, manifest.ModalityAPI))
	p.Incompatible = []manifest.Incompatible{{CohortID: "c2", Reasons: []string{"versión menor que la mínima"}}}
	r, err := Compute(m, p, []RunResult{rr("c1", "t1", 1, manifest.RunAccepted, 100)})
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Cohorts) != 1 || r.Cohorts[0].CohortID != "c1" || r.Cohorts[0].Started != 1 || r.Cohorts[0].Tokens.PerAccepted.Value != 100 {
		t.Errorf("cohortes = %+v", r.Cohorts)
	}
	if len(r.Incompatible) != 1 || r.Incompatible[0].CohortID != "c2" || r.Incompatible[0].Reasons[0] != "versión menor que la mínima" {
		t.Errorf("incompatibles = %+v", r.Incompatible)
	}
	if _, err := Compute(m, p, []RunResult{rr("c2", "t1", 1, manifest.RunAccepted, 1)}); err == nil {
		t.Error("una corrida en una celda incompatible debe ser error: no tiene consumo ni entra al denominador")
	}
	if _, err := Compute(m, p, []RunResult{rr("zz", "t1", 1, manifest.RunAccepted, 1)}); err == nil {
		t.Error("cohorte desconocida: error")
	}
}

func TestCompute_APISubscriptionAndInterfacesAreSeparate(t *testing.T) {
	m, p := setup(
		cohort("api-term", manifest.ClientClaude, "2.1.310", manifest.InterfaceTerminal, manifest.ModalityAPI),
		cohort("sub-term", manifest.ClientClaude, "2.1.310", manifest.InterfaceTerminal, manifest.ModalitySubscription),
		cohort("sub-ide", manifest.ClientClaude, "2.1.310", manifest.InterfaceIDE, manifest.ModalitySubscription),
	)
	r, err := Compute(m, p, []RunResult{
		rr("api-term", "t1", 1, manifest.RunAccepted, 100),
		rr("sub-term", "t1", 1, manifest.RunAccepted, 1000),
		rr("sub-ide", "t1", 1, manifest.RunAccepted, 10000),
	})
	if err != nil || len(r.Cohorts) != 3 {
		t.Fatalf("= %+v %v", r, err)
	}
	got := map[string]float64{}
	mod := map[string]string{}
	ifc := map[string]string{}
	for _, c := range r.Cohorts {
		got[c.CohortID] = c.Tokens.Total.Value
		mod[c.CohortID], ifc[c.CohortID] = c.Modality, c.Interface
	}
	if got["api-term"] != 100 || got["sub-term"] != 1000 || got["sub-ide"] != 10000 {
		t.Errorf("totales mezclados: %v", got)
	}
	if mod["api-term"] != "api" || mod["sub-term"] != "subscription" || ifc["sub-ide"] != "ide" || ifc["sub-term"] != "terminal" {
		t.Errorf("modalidad/interfaz = %v %v", mod, ifc)
	}
}

func TestCompute_RealInterfaceFromPlanIsReported(t *testing.T) {
	m, p := setup(cohort("c1", manifest.ClientClaude, "2.1.310", manifest.InterfaceIDE, manifest.ModalityAPI))
	p.Executable[0].Interface = manifest.InterfaceTerminal // CLI en la terminal integrada del IDE
	s := one(t, m, p, []RunResult{rr("c1", "t1", 1, manifest.RunAccepted, 1)})
	if s.Interface != "terminal" {
		t.Errorf("interfaz = %q, want la real del plan (terminal)", s.Interface)
	}
}

func TestCompute_ZeroAcceptedPublishesTotalAndQuotientHasNoValue(t *testing.T) {
	m, p := setup(cohort("c1", manifest.ClientCodex, "1", manifest.InterfaceTerminal, manifest.ModalityAPI))
	s := one(t, m, p, []RunResult{
		rr("c1", "t1", 1, manifest.RunRejectedQuality, 300),
		rr("c1", "t2", 1, manifest.RunBlockedPermissions, 20),
	})
	if s.Accepted != 0 || !s.Tokens.Total.Available || s.Tokens.Total.Value != 320 {
		t.Errorf("total = %+v", s.Tokens.Total)
	}
	q := s.Tokens.PerAccepted
	if q.Available || q.Value != 0 || !strings.Contains(q.Reason, "sin tareas aceptadas") {
		t.Errorf("cociente = %+v, want sin valor y motivo", q)
	}
	// sin ninguna corrida: nada que publicar
	e := one(t, m, p, nil)
	if e.Started != 0 || e.Tokens.Total.Available || e.Tokens.PerAccepted.Available {
		t.Errorf("cohorte vacía = %+v", e)
	}
}

func TestCompute_QuotaUnavailableIsNotComputableWithReason(t *testing.T) {
	// D5: con la cuenta Claude Pro compartida, la cuota no se puede atribuir a la corrida.
	why := "cuenta compartida con el uso normal: la cuota no es atribuible a la corrida (D5)"
	m, p := setup(cohort("sub", manifest.ClientClaude, "2.1.310", manifest.InterfaceTerminal, manifest.ModalitySubscription))
	c := func(v float64) Cost {
		return Cost{Tokens: tok(v), Money: gone("no aplica a suscripción"), Quota: gone(why)}
	}
	s := one(t, m, p, []RunResult{
		{CohortID: "sub", TaskID: "t1", Repetition: 1, Status: manifest.RunAccepted, Attempts: []Cost{c(100)}},
		{CohortID: "sub", TaskID: "t2", Repetition: 1, Status: manifest.RunRejectedQuality, Attempts: []Cost{c(50)}},
	})
	if s.Quota.Total.Available || s.Quota.Total.Reason != why || s.Quota.PerAccepted.Available || s.Quota.PerAccepted.Reason == "" || s.Quota.Complete {
		t.Errorf("cuota = %+v, want no calculable con motivo", s.Quota)
	}
	if !s.Tokens.PerAccepted.Available || s.Tokens.PerAccepted.Value != 150 {
		t.Errorf("los tokens (consumo técnico) siguen publicados: %+v", s.Tokens)
	}
	if s.Money.Total.Available {
		t.Errorf("dinero en suscripción: %+v", s.Money)
	}
}

func TestCompute_IncompleteCoverageHasNoQuotientButPublishesCoverage(t *testing.T) {
	m, p := setup(cohort("c1", manifest.ClientClaude, "2.1.310", manifest.InterfaceTerminal, manifest.ModalityAPI))
	partial := Cost{Tokens: Amount{Available: true, Value: 40, Partial: true}}
	s := one(t, m, p, []RunResult{
		{CohortID: "c1", TaskID: "t1", Repetition: 1, Status: manifest.RunAccepted, Attempts: []Cost{tokCost(100)}},
		{CohortID: "c1", TaskID: "t2", Repetition: 1, Status: manifest.RunAccepted, Attempts: []Cost{{Tokens: gone("la interfaz no expone consumo")}}},
		{CohortID: "c1", TaskID: "t3", Repetition: 1, Status: manifest.RunRejectedQuality, Attempts: []Cost{partial}},
	})
	m1 := s.Tokens
	if m1.Expected != 3 || m1.Counted != 2 || m1.Partial != 1 || m1.Complete {
		t.Errorf("cobertura = %+v, want esperados 3, contados 2, parciales 1", m1)
	}
	if !m1.Total.Available || m1.Total.Value != 140 {
		t.Errorf("total publicado de lo observado = %+v", m1.Total)
	}
	if m1.PerAccepted.Available || !strings.Contains(m1.PerAccepted.Reason, "cobertura") {
		t.Errorf("cociente = %+v, want sin valor por cobertura incompleta", m1.PerAccepted)
	}
}

func TestCompute_PartialTotalAloneLeavesFullCoverageIndicators(t *testing.T) {
	m, p := setup(cohort("c1", manifest.ClientClaude, "2.1.310", manifest.InterfaceTerminal, manifest.ModalityAPI))
	s := one(t, m, p, []RunResult{{CohortID: "c1", TaskID: "t1", Repetition: 1, Status: manifest.RunAccepted,
		Attempts: []Cost{{Tokens: Amount{Available: true, Value: 10, Partial: true}}}}})
	if s.Tokens.Complete || s.Tokens.PerAccepted.Available {
		t.Errorf("= %+v", s.Tokens)
	}
}

func TestCompute_PendingReviewIsNeitherAcceptedNorRejected(t *testing.T) {
	m, p := setup(cohort("c1", manifest.ClientClaude, "2.1.310", manifest.InterfaceTerminal, manifest.ModalityAPI))
	s := one(t, m, p, []RunResult{
		rr("c1", "t1", 1, manifest.RunAccepted, 100),
		rr("c1", "t2", 1, run.StatusPendingReview, 60),
		rr("c1", "t3", 1, manifest.RunRejectedQuality, 40),
	})
	if s.Started != 3 || s.Accepted != 1 || s.Rejected != 1 || s.Pending != 1 {
		t.Errorf("conteos = %+v", s)
	}
	if s.Tokens.Total.Value != 200 || !s.Provisional {
		t.Errorf("el costo de la pendiente entra al total y el resumen queda provisional: %+v", s)
	}
}

func TestCompute_MoneyForAPIAndReconciliationExamples(t *testing.T) {
	// Ejemplos del diseño: Claude 60.000 (no 80.000 ni 48.000) y Codex 48.000 (no 66.000).
	i := func(n int64) *int64 { return &n }
	claude, _ := usage.Reconcile(usage.Input{Aggregate: i(60000), Scope: usage.ScopeIncludesChildren, Entries: []usage.Entry{
		{Role: "principal", Tokens: 40000}, {Role: "worker", Tokens: 12000}, {Role: "worker", Tokens: 8000}}})
	codex, _ := usage.Reconcile(usage.Input{Aggregate: i(48000), Scope: usage.ScopeIncludesChildren, Entries: []usage.Entry{
		{Role: "padre", Tokens: 30000}, {Role: "hijo", Tokens: 18000}}})
	m, p := setup(
		cohort("cl", manifest.ClientClaude, "2.1.310", manifest.InterfaceTerminal, manifest.ModalityAPI),
		cohort("cx", manifest.ClientCodex, "1", manifest.InterfaceTerminal, manifest.ModalityAPI))
	cc := Cost{Tokens: TokensFromTotals(claude), Money: Amount{Available: true, Value: 1.5}, Quota: gone("n/a")}
	cx := Cost{Tokens: TokensFromTotals(codex), Money: Amount{Available: true, Value: 0.75}, Quota: gone("n/a")}
	r, err := Compute(m, p, []RunResult{
		{CohortID: "cl", TaskID: "t1", Repetition: 1, Status: manifest.RunAccepted, Attempts: []Cost{cc}},
		{CohortID: "cx", TaskID: "t1", Repetition: 1, Status: manifest.RunAccepted, Attempts: []Cost{cx}},
	})
	if err != nil {
		t.Fatal(err)
	}
	by := map[string]CohortSummary{}
	for _, c := range r.Cohorts {
		by[c.CohortID] = c
	}
	if by["cl"].Tokens.Total.Value != 60000 || by["cx"].Tokens.Total.Value != 48000 {
		t.Errorf("totales = %v %v", by["cl"].Tokens.Total, by["cx"].Tokens.Total)
	}
	if by["cl"].Tokens.PerAccepted.Value != 60000 || by["cl"].Money.PerAccepted.Value != 1.5 || by["cx"].Money.Total.Value != 0.75 {
		t.Errorf("cocientes = %+v %+v", by["cl"], by["cx"])
	}
}

func TestTokensFromTotals_PartialOnlyWhenDetailExceedsOrIsUnverified(t *testing.T) {
	i := func(n int64) *int64 { return &n }
	below, _ := usage.Reconcile(usage.Input{Aggregate: i(100), Scope: usage.ScopeIncludesChildren, Entries: []usage.Entry{{Role: "p", Tokens: 60}}})
	if a := TokensFromTotals(below); !a.Available || a.Value != 100 || a.Partial {
		t.Errorf("detalle menor (atribución parcial, total confiable) = %+v", a)
	}
	above, _ := usage.Reconcile(usage.Input{Aggregate: i(100), Scope: usage.ScopeIncludesChildren, Entries: []usage.Entry{{Role: "p", Tokens: 160}}})
	if a := TokensFromTotals(above); !a.Available || a.Value != 100 || !a.Partial {
		t.Errorf("detalle mayor = %+v, want parcial", a)
	}
}

func TestCompute_InvalidInputs(t *testing.T) {
	m, p := setup(cohort("c1", manifest.ClientClaude, "2.1.310", manifest.InterfaceTerminal, manifest.ModalityAPI))
	for name, runs := range map[string][]RunResult{
		"corrida repetida":   {rr("c1", "t1", 1, manifest.RunAccepted, 1), rr("c1", "t1", 1, manifest.RunAccepted, 1)},
		"estado desconocido": {rr("c1", "t1", 1, "otra", 1)},
		"sin intentos":       {rr("c1", "t1", 1, manifest.RunAccepted)},
		"consumo negativo":   {rr("c1", "t1", 1, manifest.RunAccepted, -1)},
		"consumo NaN":        {rr("c1", "t1", 1, manifest.RunAccepted, math.NaN())},
		"consumo infinito":   {rr("c1", "t1", 1, manifest.RunAccepted, math.Inf(1))},
	} {
		if _, err := Compute(m, p, runs); err == nil {
			t.Errorf("%s: err = nil", name)
		}
	}
}

func TestFromRun_MapsStatusAndRequiresOneCostPerAttempt(t *testing.T) {
	r := run.Run{CohortID: "c1", TaskID: "t1", Repetition: 2, Status: manifest.RunAccepted, Attempts: []run.Attempt{{N: 1}, {N: 2}}}
	got, err := FromRun(r, costs(1, 2))
	if err != nil || got.CohortID != "c1" || got.TaskID != "t1" || got.Repetition != 2 || got.Status != manifest.RunAccepted || len(got.Attempts) != 2 {
		t.Errorf("= %+v, %v", got, err)
	}
	if _, err := FromRun(r, costs(1)); err == nil {
		t.Error("costos de menos: error")
	}
}

func TestCompute_PropertyOrderOfRunsAndCohortsDoesNotChangeTheReport(t *testing.T) {
	rng := rand.New(rand.NewSource(20261111))
	statuses := []manifest.RunStatus{manifest.RunAccepted, manifest.RunRejectedQuality, manifest.RunBlockedPermissions,
		manifest.RunTimeout, manifest.RunClientFailure, run.StatusPendingReview, run.StatusInterrupted}
	for it := 0; it < 100; it++ {
		cs := []manifest.Cohort{
			cohort("a", manifest.ClientClaude, "1", manifest.InterfaceTerminal, manifest.ModalityAPI),
			cohort("b", manifest.ClientClaude, "1", manifest.InterfaceIDE, manifest.ModalitySubscription),
			cohort("c", manifest.ClientCodex, "1", manifest.InterfaceTerminal, manifest.ModalityAPI),
		}
		m, p := setup(cs...)
		var runs []RunResult
		for i, n := 0, rng.Intn(15); i < n; i++ {
			c := cs[rng.Intn(3)].ID
			var at []Cost
			for j, k := 0, 1+rng.Intn(3); j < k; j++ {
				a := Cost{Tokens: tok(float64(rng.Intn(1000))), Money: Amount{Available: true, Value: float64(rng.Intn(100)) / 7},
					Quota: gone("n/a")}
				if rng.Intn(6) == 0 {
					a.Tokens = gone("sin señal")
				}
				at = append(at, a)
			}
			runs = append(runs, RunResult{CohortID: c, TaskID: "t", Repetition: i + 1, Status: statuses[rng.Intn(len(statuses))], Attempts: at})
		}
		base, err := Compute(m, p, runs)
		if err != nil {
			t.Fatal(err)
		}
		rng.Shuffle(len(runs), func(x, y int) { runs[x], runs[y] = runs[y], runs[x] })
		rng.Shuffle(len(m.Cohorts), func(x, y int) { m.Cohorts[x], m.Cohorts[y] = m.Cohorts[y], m.Cohorts[x] })
		rng.Shuffle(len(p.Executable), func(x, y int) { p.Executable[x], p.Executable[y] = p.Executable[y], p.Executable[x] })
		for i := range runs {
			at := runs[i].Attempts
			rng.Shuffle(len(at), func(x, y int) { at[x], at[y] = at[y], at[x] })
		}
		again, err := Compute(m, p, runs)
		if err != nil || !reflect.DeepEqual(base, again) {
			t.Fatalf("el orden altera el resumen:\n%+v\n%+v\n%v", base, again, err)
		}
		// invariante: iniciadas = suma de los estados
		for _, c := range base.Cohorts {
			if c.Started != c.Accepted+c.Rejected+c.Blocked+c.TimedOut+c.ClientFailed+c.Interrupted+c.Pending {
				t.Fatalf("no cuadra: %+v", c)
			}
		}
	}
}

func TestMarkdown_ListsCohortsIncompatiblesAndValuelessQuotients(t *testing.T) {
	m, p := setup(
		cohort("c1", manifest.ClientCodex, "1", manifest.InterfaceTerminal, manifest.ModalityAPI),
		cohort("sub", manifest.ClientClaude, "2.1.310", manifest.InterfaceTerminal, manifest.ModalitySubscription))
	p.Incompatible = []manifest.Incompatible{{CohortID: "x9", Reasons: []string{"opt-in de contenido activo: OTEL_LOG_USER_PROMPTS"}}}
	r, err := Compute(m, p, []RunResult{
		rr("c1", "t1", 1, manifest.RunRejectedQuality, 300),
		{CohortID: "sub", TaskID: "t1", Repetition: 1, Status: manifest.RunAccepted, Attempts: []Cost{
			{Tokens: tok(10), Money: gone("no aplica"), Quota: gone("cuota compartida (D5)")}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	md := Markdown(r)
	for _, want := range []string{"c1", "sub", "300", "sin valor", "sin tareas aceptadas", "cuota compartida (D5)", "Celdas incompatibles", "x9", "OTEL_LOG_USER_PROMPTS", "Cálculo"} {
		if !strings.Contains(md, want) {
			t.Errorf("el Markdown no contiene %q:\n%s", want, md)
		}
	}
	if Markdown(r) != md {
		t.Error("el Markdown no es determinista")
	}
}
