package fixture

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// run prepara (o reinicia) el checkout, aplica el parche de change si lo hay
// y evalua. Las familias de respuesta usan el archivo como Answer.
func run(t *testing.T, f *Fixture, c *Checkout, variant string, out Output) Verdict {
	t.Helper()
	if err := c.Reset(); err != nil {
		t.Fatal(err)
	}
	out.Dir = c.Dir
	if strings.HasSuffix(variant, ".patch") {
		if err := ApplyPatch(c.Dir, f.path(variant)); err != nil {
			t.Fatal(err)
		}
	} else if variant != "" {
		b, err := os.ReadFile(f.path(variant))
		if err != nil {
			t.Fatal(err)
		}
		out.Answer = b
	}
	v, err := f.Evaluate(out)
	if err != nil {
		t.Fatalf("Evaluate(%s): %v", variant, err)
	}
	return v
}

func prepared(t *testing.T, fam Family) (*Fixture, *Checkout) {
	t.Helper()
	f := mustLoad(t, fam)
	c, err := f.Prepare()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { c.Close() })
	return f, c
}

func outcomes(v Verdict) map[string]Outcome {
	m := map[string]Outcome{}
	for _, r := range v.Criteria {
		m[r.ID] = r.Outcome
	}
	return m
}

func reviewerYes(f *Fixture) map[string]bool {
	d := map[string]bool{}
	for _, c := range f.Task.Criteria {
		if c.Check == checkReviewer {
			d[c.ID] = true
		}
	}
	return d
}

// Tarea 1.3: los chequeos de cada familia aceptan la referencia y rechazan la
// variante defectuosa con el criterio incumplido esperado.
func TestEvaluate_ReferenceAcceptedDefectiveRejected(t *testing.T) {
	t.Parallel()
	cases := []struct {
		fam        Family
		wantFailed []string
		review     map[string]ReviewCall // decision del revisor para la defectuosa
	}{
		{FamilyExploration, []string{"E1", "E2"}, nil},
		{FamilySmallChange, []string{"S3"}, nil},
		{FamilyMultiFile, []string{"M2"}, nil},
		{FamilyDiagnosis, []string{"G2", "G3"}, nil},
		{FamilyReview, []string{"R1", "R3"}, map[string]ReviewCall{"F2": {Disposition: "rejected"}}},
	}
	for _, tc := range cases {
		t.Run(string(tc.fam), func(t *testing.T) {
			f, c := prepared(t, tc.fam)
			ref := run(t, f, c, f.Task.Reference, Output{Decisions: reviewerYes(f)})
			if ref.Status != StatusAccepted || len(ref.Failed) != 0 {
				t.Errorf("referencia: %+v, want aceptada", ref)
			}
			if len(ref.Criteria) != len(f.Task.Criteria) {
				t.Errorf("referencia: %d resultados, want uno por criterio (%d)", len(ref.Criteria), len(f.Task.Criteria))
			}
			def := run(t, f, c, f.Task.Defective, Output{Decisions: reviewerYes(f), Review: tc.review})
			if def.Status != StatusRejected || !reflect.DeepEqual(def.Failed, tc.wantFailed) {
				t.Errorf("defectuosa: status=%s failed=%v, want rechazada %v; criterios=%+v", def.Status, def.Failed, tc.wantFailed, def.Criteria)
			}
		})
	}
}

// Escenario "Rechazo por calidad": la prueba pasa pero falta la prueba de
// regresion; el criterio incumplido queda registrado.
func TestEvaluate_SmallChangeTestsPassButRegressionMissing(t *testing.T) {
	t.Parallel()
	f, c := prepared(t, FamilySmallChange)
	v := run(t, f, c, f.Task.Defective, Output{})
	o := outcomes(v)
	if o["S1"] != Met || o["S2"] != Met || o["S3"] != NotMet || o["S4"] != Met {
		t.Errorf("outcomes = %v, want S1,S2,S4 cumplidos y S3 incumplido", o)
	}
	// Una prueba agregada que tambien pasa con el codigo original no es de
	// regresion: S3 incumplido aunque el arreglo este.
	if err := c.Reset(); err != nil {
		t.Fatal(err)
	}
	if err := ApplyPatch(c.Dir, f.path(f.Task.Defective)); err != nil {
		t.Fatal(err)
	}
	os.WriteFile(filepath.Join(c.Dir, "ascii_test.go"), []byte("package texto\n\nimport \"testing\"\n\nfunc TestNoRegresion(t *testing.T) {\n\tif Truncate(\"ab\", 1) != \"a\" {\n\t\tt.Fatal(\"x\")\n\t}\n}\n"), 0o644)
	v, err := f.Evaluate(Output{Dir: c.Dir})
	if o := outcomes(v); err != nil || o["S3"] != NotMet || o["S1"] != Met {
		t.Errorf("prueba nueva que pasa con el original: %v %v, want S3 incumplido", o, err)
	}
	// Sin cambios: go test pasa pero no hay arreglo ni regresion.
	v = run(t, f, c, "", Output{})
	if o := outcomes(v); o["S2"] != NotMet || o["S3"] != NotMet || v.Status != StatusRejected {
		t.Errorf("checkout sin cambios: %v %s, want S2 y S3 incumplidos", o, v.Status)
	}
}

// Control de archivos tocados: tocar o borrar un archivo no permitido incumple.
func TestEvaluate_TouchedFilesOutsideAllowed(t *testing.T) {
	t.Parallel()
	f, c := prepared(t, FamilyMultiFile)
	if err := c.Reset(); err != nil {
		t.Fatal(err)
	}
	ApplyPatch(c.Dir, f.path(f.Task.Reference))
	os.Remove(filepath.Join(c.Dir, "api", "api_test.go"))
	v, err := f.Evaluate(Output{Dir: c.Dir})
	if err != nil {
		t.Fatal(err)
	}
	if o := outcomes(v); o["M3"] != NotMet || o["M1"] != Met {
		t.Errorf("borrar api_test.go: %v, want M3 incumplido y M1 cumplido", o)
	}
	os.WriteFile(filepath.Join(c.Dir, "api", "api_test.go"), []byte("package api\n"), 0o644)
	if err := c.Reset(); err != nil {
		t.Fatal(err)
	}
	ApplyPatch(c.Dir, f.path(f.Task.Reference))
	os.WriteFile(filepath.Join(c.Dir, "extra.txt"), []byte("x"), 0o644)
	if v, _ := f.Evaluate(Output{Dir: c.Dir}); outcomes(v)["M3"] != NotMet {
		t.Errorf("archivo agregado no permitido: %v, want M3 incumplido", outcomes(v))
	}
}

// Criterios requiere_revisor: sin decision el veredicto queda pendiente; con
// decision negativa, rechazado.
func TestEvaluate_ReviewerCriteriaPendingUntilDecided(t *testing.T) {
	t.Parallel()
	for _, fam := range []Family{FamilyExploration, FamilyDiagnosis} {
		t.Run(string(fam), func(t *testing.T) {
			f, c := prepared(t, fam)
			rid := ""
			for id := range reviewerYes(f) {
				rid = id
			}
			v := run(t, f, c, f.Task.Reference, Output{})
			if v.Status != StatusPending || !reflect.DeepEqual(v.Pending, []string{rid}) {
				t.Errorf("sin decision: %s pending=%v, want pendiente [%s]", v.Status, v.Pending, rid)
			}
			v = run(t, f, c, f.Task.Reference, Output{Decisions: map[string]bool{rid: false}})
			if v.Status != StatusRejected || !reflect.DeepEqual(v.Failed, []string{rid}) {
				t.Errorf("decision negativa: %s failed=%v, want rechazada [%s]", v.Status, v.Failed, rid)
			}
		})
	}
}

func TestEvaluate_InputErrors(t *testing.T) {
	t.Parallel()
	f, c := prepared(t, FamilyExploration)
	if err := c.Reset(); err != nil {
		t.Fatal(err)
	}
	ans, _ := os.ReadFile(f.path(f.Task.Reference))
	for name, d := range map[string]map[string]bool{
		"criterio desconocido":   {"E9": true},
		"criterio no de revisor": {"E1": true},
	} {
		if _, err := f.Evaluate(Output{Dir: c.Dir, Answer: ans, Decisions: d}); err == nil {
			t.Errorf("%s: err = nil, want error", name)
		}
	}
}

// Exploracion: respuesta en prosa con bloque json final se acepta; una
// respuesta sin JSON incumple sin error; tocar el checkout incumple E4.
func TestEvaluate_ExplorationAnswerFormatsAndIntactCheckout(t *testing.T) {
	t.Parallel()
	f, c := prepared(t, FamilyExploration)
	ref, _ := os.ReadFile(f.path(f.Task.Reference))
	if err := c.Reset(); err != nil {
		t.Fatal(err)
	}
	prose := "Recorrí el flujo.\n\n```json\n" + string(ref) + "```\n"
	v, err := f.Evaluate(Output{Dir: c.Dir, Answer: []byte(prose), Decisions: reviewerYes(f)})
	if err != nil || v.Status != StatusAccepted {
		t.Errorf("prosa + bloque json: %+v %v, want aceptada", v, err)
	}
	v, err = f.Evaluate(Output{Dir: c.Dir, Answer: []byte("no sé"), Decisions: reviewerYes(f)})
	if o := outcomes(v); err != nil || o["E1"] != NotMet || o["E2"] != NotMet {
		t.Errorf("sin json: %v %v, want E1 y E2 incumplidos sin error", o, err)
	}
	os.WriteFile(filepath.Join(c.Dir, "internal", "catalog", "catalog.go"), []byte("package catalog\n"), 0o644)
	v, _ = f.Evaluate(Output{Dir: c.Dir, Answer: ref, Decisions: reviewerYes(f)})
	if o := outcomes(v); o["E4"] != NotMet || v.Status != StatusRejected {
		t.Errorf("checkout modificado: %v %s, want E4 incumplido", o, v.Status)
	}
}

// Diagnostico: una reproduccion con una prueba que pasa o no existe, o con un
// fragmento que no aparece en la salida, incumple G1.
func TestEvaluate_DiagnosisReproductionMustFailWithCitedOutput(t *testing.T) {
	t.Parallel()
	f, c := prepared(t, FamilyDiagnosis)
	for name, repro := range map[string]string{
		"prueba que pasa":     `{"test": "TestWrapShortWidth", "observed": "ok"}`,
		"prueba inexistente":  `{"test": "TestNoExiste", "observed": "FAIL"}`,
		"fragmento ausente":   `{"test": "TestWrapExactWidth", "observed": "panic: nil map"}`,
		"fragmento vacio":     `{"test": "TestWrapExactWidth", "observed": ""}`,
		"nombre no es prueba": `{"test": "-v", "observed": "FAIL"}`,
	} {
		ans := `{"reproduction": ` + repro + `, "root_cause": {"file": "report/report.go", "start_line": 15, "end_line": 15, "mechanism": "m"}, "discarded": {"file": "config/config.go", "evidence": "e"}}`
		if err := c.Reset(); err != nil {
			t.Fatal(err)
		}
		v, err := f.Evaluate(Output{Dir: c.Dir, Answer: []byte(ans), Decisions: reviewerYes(f)})
		if o := outcomes(v); err != nil || o["G1"] != NotMet || o["G2"] != Met {
			t.Errorf("%s: %v %v, want G1 incumplido y G2 cumplido", name, o, err)
		}
	}
}

// Revision: el diff del prompt es el parche sembrado; una disposicion
// desconocida es error explicito (DEUDA 6 integrada); un hallazgo sin decidir
// deja R3 pendiente; el revisor decide el mecanismo.
func TestEvaluate_ReviewDispositionsAndMechanism(t *testing.T) {
	t.Parallel()
	f, c := prepared(t, FamilyReview)
	patch, _ := os.ReadFile(f.path(f.Task.Source.Patch))
	if !strings.Contains(string(f.Prompt), string(patch)) {
		t.Error("el prompt de revision no contiene el parche sembrado literal")
	}
	def, _ := os.ReadFile(f.path(f.Task.Defective))
	if err := c.Reset(); err != nil {
		t.Fatal(err)
	}
	if _, err := f.Evaluate(Output{Dir: c.Dir, Answer: def, Review: map[string]ReviewCall{"F2": {Disposition: "rejectd"}}}); err == nil {
		t.Error("disposicion con typo: err = nil, want error")
	}
	if _, err := f.Evaluate(Output{Dir: c.Dir, Answer: def, Review: map[string]ReviewCall{"F9": {Disposition: "rejected"}}}); err == nil {
		t.Error("decision sobre hallazgo inexistente: err = nil, want error")
	}
	v, err := f.Evaluate(Output{Dir: c.Dir, Answer: def})
	if o := outcomes(v); err != nil || o["R3"] != Pending || o["R1"] != NotMet || v.Status != StatusRejected {
		t.Errorf("sin decidir F2: %v %s %v, want R3 pendiente, R1 incumplido, rechazada", o, v.Status, err)
	}
	// F2 rechazado por el revisor es falso positivo: R3 incumplido. Como
	// hallazgo nuevo (miembro benigno) no cuenta como falso positivo.
	v, _ = f.Evaluate(Output{Dir: c.Dir, Answer: def, Review: map[string]ReviewCall{"F2": {Disposition: "rejected"}}})
	if o := outcomes(v); o["R3"] != NotMet {
		t.Errorf("F2 rechazado: %v, want R3 incumplido", o)
	}
	v, _ = f.Evaluate(Output{Dir: c.Dir, Answer: def, Review: map[string]ReviewCall{"F2": {Disposition: "real_defect"}}})
	if o := outcomes(v); o["R3"] != Met {
		t.Errorf("F2 hallazgo nuevo: %v, want R3 cumplido", o)
	}
	// Mecanismo redactado distinto: sin decision no empareja; el revisor lo fija.
	ans := `{"findings": [
	 {"id": "A", "file": "internal/catalog/catalog.go", "start_line": 71, "end_line": 71, "mechanism": "ids repetidos pasan", "evidence": "seen[h.Name]"},
	 {"id": "B", "file": "internal/catalog/catalog.go", "start_line": 115, "end_line": 115, "mechanism": "dependencia_desconocida_aceptada", "evidence": "c.index[h.ID]"}]}`
	v, _ = f.Evaluate(Output{Dir: c.Dir, Answer: []byte(ans)})
	if o := outcomes(v); o["R1"] != NotMet || o["R3"] != Pending {
		t.Errorf("mecanismo sin decidir: %v, want R1 incumplido y R3 pendiente", o)
	}
	v, _ = f.Evaluate(Output{Dir: c.Dir, Answer: []byte(ans), Review: map[string]ReviewCall{"A": {Mechanism: "id_duplicado_no_detectado"}}})
	if v.Status != StatusAccepted {
		t.Errorf("mecanismo decidido por el revisor: %s %v, want aceptada", v.Status, outcomes(v))
	}
	// TP sin evidencia incumple R2.
	noEv := strings.Replace(ans, `"evidence": "c.index[h.ID]"`, `"evidence": " "`, 1)
	v, _ = f.Evaluate(Output{Dir: c.Dir, Answer: []byte(noEv), Review: map[string]ReviewCall{"A": {Mechanism: "id_duplicado_no_detectado"}}})
	if o := outcomes(v); o["R2"] != NotMet {
		t.Errorf("TP sin evidencia: %v, want R2 incumplido", o)
	}
}
