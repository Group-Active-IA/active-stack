package evidence

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Group-Active-IA/active-stack/eval/baseline/fixture"
)

// Cada caso de testdata/casos es un recibo (o ninguno) más lo que observaron la fuente
// independiente, la reejecución y el revisor, y el veredicto que el contrato (design.md D3 y D4)
// exige. Las pruebas recorren todos los casos: agregar un caso es agregar un JSON.

func TestCases_VerdictMatchesContract(t *testing.T) {
	files, err := filepath.Glob("testdata/casos/*.json")
	if err != nil || len(files) == 0 {
		t.Fatalf("sin casos en testdata/casos: %v", err)
	}
	for _, f := range files {
		c, err := LoadCase(f)
		if err != nil {
			t.Fatalf("%s: %v", f, err)
		}
		t.Run(c.ID, func(t *testing.T) {
			got := EvaluateRequirement(c.Requirement, c.Receipts, c.Env)
			if got.Status != c.Want.Status {
				t.Errorf("estado = %q, esperado %q (motivos: %v)", got.Status, c.Want.Status, got.Reasons)
			}
			if got.Integrity != c.Want.Integrity {
				t.Errorf("integridad = %q, esperada %q", got.Integrity, c.Want.Integrity)
			}
			if got.Provenance != c.Want.Provenance {
				t.Errorf("procedencia = %q, esperada %q", got.Provenance, c.Want.Provenance)
			}
			if got.Sufficiency != c.Want.Sufficiency {
				t.Errorf("suficiencia = %q, esperada %q", got.Sufficiency, c.Want.Sufficiency)
			}
		})
	}
}

// Los cinco casos de la ficha (CHANGES.md:194) más "recibo incompleto" deben estar cubiertos.
func TestCases_CoverTheCardAndTheDefaults(t *testing.T) {
	need := map[string]Status{
		"inventado-copia-ausente":           Rejected,
		"viejo":                             Rejected,
		"comando-insuficiente":              Rejected,
		"truncado-sin-original":             Rejected,
		"tests-verdes-requisito-incumplido": Rejected,
		"incompleto":                        Unverified,
		"cero-recibos":                      Unverified,
		"valido":                            Verified,
	}
	for id, want := range need {
		c, err := LoadCase(filepath.Join("testdata", "casos", id+".json"))
		if err != nil {
			t.Fatalf("%s: %v", id, err)
		}
		if c.Want.Status != want {
			t.Errorf("%s: el fixture espera %q, la ficha exige %q", id, c.Want.Status, want)
		}
	}
}

// Una ejecución sin recibos nunca queda verificada, sea cual sea el entorno.
func TestEvaluateRequirement_NoReceiptsNeverVerified(t *testing.T) {
	for _, req := range []string{"R1", "R2"} {
		got := EvaluateRequirement(req, nil, Env{})
		if got.Status != Unverified {
			t.Errorf("%s sin recibos: estado %q, esperado %q", req, got.Status, Unverified)
		}
	}
}

// Un recibo que declara otro requisito no cubre este.
func TestEvaluateRequirement_ReceiptForAnotherRequirementDoesNotCount(t *testing.T) {
	c, err := LoadCase("testdata/casos/valido.json")
	if err != nil {
		t.Fatal(err)
	}
	if got := EvaluateRequirement("R2", c.Receipts, c.Env); got.Status != Unverified {
		t.Errorf("R2 con un recibo de R1: estado %q, esperado %q", got.Status, Unverified)
	}
}

// El mismo texto con LF y con CRLF son bytes distintos: huellas distintas. Los valores son los
// que midieron sha256sum --binary, sha256sum --text y Get-FileHash en Windows y Linux (tarea 1.3).
func TestFingerprint_LFAndCRLFDiffer(t *testing.T) {
	const (
		lf   = "73aa9cd9e0edf14ab73c7754e590d31eb155e386acbfd3f7927b7a6921392b44"
		crlf = "83a6a5f9d382700099a71b77a0031c124dede6db5da853f0dd28fa5f41abff2b"
	)
	if got := Fingerprint256([]byte("linea\n")); got.Algorithm != "sha256" || got.Value != lf {
		t.Errorf("LF = %+v, esperado sha256 %s", got, lf)
	}
	if got := Fingerprint256([]byte("linea\r\n")); got.Value != crlf {
		t.Errorf("CRLF = %s, esperado %s", got.Value, crlf)
	}
}

// Misma base, distinto diff: el estado exacto probado debe diferir (tarea 4.4). Se usa TreeHash de
// 02, el mismo cálculo que la propuesta toma como definición.
func TestStateFingerprint_SameBaseDifferentDiffs(t *testing.T) {
	base := map[string]string{"f.txt": "a\nb\n", "g.txt": "a\r\nb\r\n"}
	dirs := map[string]map[string]string{
		"limpio":        base,
		"f-modificado":  {"f.txt": "a\nb\nc\n", "g.txt": "a\r\nb\r\n"},
		"g-modificado":  {"f.txt": "a\nb\n", "g.txt": "a\r\nb\r\nc\r\n"},
		"archivo-nuevo": {"f.txt": "a\nb\n", "g.txt": "a\r\nb\r\n", "nuevo.txt": "nuevo\n"},
		"f-con-crlf":    {"f.txt": "a\r\nb\r\n", "g.txt": "a\r\nb\r\n"},
		"g-borrado":     {"f.txt": "a\nb\n"},
	}
	seen := map[string]string{}
	for name, files := range dirs {
		h := treeHashOf(t, files)
		if prev, dup := seen[h]; dup {
			t.Errorf("%s y %s dan la misma huella de estado %s", name, prev, h)
		}
		seen[h] = name
	}
	// Valor medido en Windows y en Linux (WSL2) con core.autocrlf true, false e input (tarea 1.3).
	const clean = "1642de00a955de709d3901a362dc9195c1476411c13628a2714e11134b501921"
	if got := treeHashOf(t, base); got != clean {
		t.Errorf("huella del estado limpio = %s, esperada %s", got, clean)
	}
}

func treeHashOf(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	h, err := fixture.TreeHash(dir)
	if err != nil {
		t.Fatal(err)
	}
	return h
}

// Un recibo se carga de JSON y vuelve a la misma forma: los nombres de campo son los de design.md D2.
func TestLoadCase_FieldNamesFollowTheDesign(t *testing.T) {
	c, err := LoadCase("testdata/casos/valido.json")
	if err != nil {
		t.Fatal(err)
	}
	r := c.Receipts[0]
	if r.Run == "" || r.Command == "" || r.Tree.Dir == "" || r.Tree.Commit == "" || len(r.Requirements) == 0 {
		t.Errorf("campos de D2 sin cargar: %+v", r)
	}
	if r.Output.Fingerprint == nil || r.Output.Capture == nil {
		t.Errorf("salida sin huella o registro de captura: %+v", r.Output)
	}
	if missing := r.Missing(); len(missing) != 0 {
		t.Errorf("el recibo válido reporta faltantes: %v", missing)
	}
}

func TestMissing_ListsEachAbsentField(t *testing.T) {
	var empty Receipt
	if got := len(empty.Missing()); got < 6 {
		t.Errorf("un recibo vacío debe reportar al menos los 6 campos de D2, reportó %d", got)
	}
}

// loadValid carga el recibo válido; el recibo original debe seguir verificado (control de las
// regresiones que siguen).
func loadValid(t *testing.T) Case {
	t.Helper()
	c, err := LoadCase("testdata/casos/valido.json")
	if err != nil {
		t.Fatal(err)
	}
	if got := EvaluateRequirement("R1", c.Receipts, c.Env); got.Status != Verified {
		t.Fatalf("el recibo original dejó de estar verificado: %q (%v)", got.Status, got.Reasons)
	}
	return c
}

// Cambiar el comando conservando el id de corrida no permite reutilizar la corroboración ni la
// suficiencia: la fuente independiente vio otro comando (rechazado) y el juicio del revisor no
// corresponde al comando nuevo.
func TestEvaluateRequirement_ChangedCommandKeepingRunIDReusesNoJudgment(t *testing.T) {
	c := loadValid(t)
	r := c.Receipts[0]
	r.Command = "echo no-test"
	got := EvaluateRequirement("R1", []Receipt{r}, c.Env)
	if got.Status == Verified {
		t.Fatalf("comando distinto con el mismo id quedó verificado: %+v", got)
	}
	if got.Provenance == provOK || got.Sufficiency == suffOK {
		t.Errorf("se reutilizaron juicios de otra ejecución: procedencia %q, suficiencia %q", got.Provenance, got.Sufficiency)
	}
	if got.Status != Rejected || got.Provenance != provBad {
		t.Errorf("la fuente independiente vio otro comando: esperado rechazado/contradicha, fue %q/%q", got.Status, got.Provenance)
	}

	// Sin registro de corroboración para el comando nuevo, solo la suficiencia deja de aplicar.
	env := c.Env
	env.Corroboration = nil
	got = EvaluateRequirement("R1", []Receipt{r}, env)
	if got.Status != Unverified || got.Sufficiency != notEvaluable || got.Provenance != provNone {
		t.Errorf("sin evidencia independiente de esa ejecución: esperado sin verificar, fue %+v", got)
	}
}

// La corroboración identifica comando, árbol y resultado: cualquier diferencia con el recibo es una
// ejecución distinta de la que la fuente observó.
func TestEvaluateRequirement_CorroborationMustDescribeTheSameExecution(t *testing.T) {
	other := 1
	cases := map[string]func(*Receipt){
		"otro directorio": func(r *Receipt) { r.Tree.Dir = "/work/otro" },
		"otro commit":     func(r *Receipt) { r.Tree.Commit = "0000000000000000000000000000000000000000" },
		"otra huella":     func(r *Receipt) { r.Tree.Fingerprint = &Fingerprint{Algorithm: "sha256", Value: "aa"} },
		"otro resultado":  func(r *Receipt) { r.Result.ExitCode = &other },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			c := loadValid(t)
			r := c.Receipts[0]
			mutate(&r)
			env := c.Env
			// Aísla la procedencia: el entorno acompaña al recibo mutado salvo en lo que se juzga.
			env.Dir, env.Commit, env.StateFP = r.Tree.Dir, r.Tree.Commit, r.Tree.Fingerprint
			got := EvaluateRequirement("R1", []Receipt{r}, env)
			if got.Status == Verified || got.Provenance != provBad {
				t.Errorf("la corroboración de otra ejecución se reutilizó: %+v", got)
			}
		})
	}
}

// El juicio de suficiencia es de un requisito concreto: el de R1 no respalda R2.
func TestEvaluateRequirement_SufficiencyIsPerRequirement(t *testing.T) {
	c := loadValid(t)
	r := c.Receipts[0]
	r.Requirements = []string{"R1", "R2"}
	got := EvaluateRequirement("R2", []Receipt{r}, c.Env)
	if got.Status != Unverified || got.Sufficiency != notEvaluable {
		t.Errorf("R2 con el juicio de R1: esperado sin verificar/no evaluable, fue %+v", got)
	}
	if got := EvaluateRequirement("R1", []Receipt{r}, c.Env); got.Status != Verified {
		t.Errorf("R1 debe seguir verificado: %+v", got)
	}
}

// Captura declarada por el agente: sin una fuente independiente distinta, el recibo no se corrobora
// aunque la huella coincida, ya sea porque no hay registro o porque el registro lo firma el propio
// agente.
func TestEvaluateRequirement_AgentDeclaredCaptureWithoutIndependentCorroboration(t *testing.T) {
	c := loadValid(t)
	r := c.Receipts[0]
	cap := *r.Output.Capture
	cap.Source = "agente"
	r.Output.Capture = &cap

	t.Run("sin registro", func(t *testing.T) {
		env := c.Env
		env.Corroboration = nil
		got := EvaluateRequirement("R1", []Receipt{r}, env)
		if got.Status != Unverified || got.Provenance != provNone || got.Integrity != intPass {
			t.Errorf("esperado sin verificar/sin corroborar con integridad aprobada, fue %+v", got)
		}
	})
	t.Run("registro de la misma fuente", func(t *testing.T) {
		env := c.Env
		rec := env.Corroboration[0]
		rec.Source = "agente"
		env.Corroboration = []CorroborationRecord{rec}
		got := EvaluateRequirement("R1", []Receipt{r}, env)
		if got.Status != Unverified || got.Provenance != provNone {
			t.Errorf("la autocorroboración del agente no es independiente: %+v", got)
		}
	})
	t.Run("registro de otra corrida", func(t *testing.T) {
		env := c.Env
		rec := env.Corroboration[0]
		rec.Run = "r-999"
		env.Corroboration = []CorroborationRecord{rec}
		got := EvaluateRequirement("R1", []Receipt{r}, env)
		if got.Status != Unverified || got.Provenance != provNone {
			t.Errorf("la corroboración de otra corrida no aplica: %+v", got)
		}
	})
}
