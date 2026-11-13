package fixture

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Group-Active-IA/active-stack/eval/baseline/review"
)

// Evaluate aplica los chequeos de la rubrica a la salida de una corrida y
// decide el veredicto con Decide. Un criterio requiere_revisor queda
// pendiente hasta que Decisions traiga su decision. Las fallas de la salida
// del agente (respuesta sin JSON, pruebas que fallan) son criterios
// incumplidos; los errores son del arnes o de las decisiones del revisor
// (criterio desconocido, disposicion desconocida, hallazgo inexistente).
func (f *Fixture) Evaluate(out Output) (Verdict, error) {
	byID := map[string]Criterion{}
	for _, c := range f.Task.Criteria {
		byID[c.ID] = c
	}
	for id := range out.Decisions {
		c, ok := byID[id]
		if !ok {
			return Verdict{}, fmt.Errorf("decisión sobre criterio desconocido %q", id)
		}
		if c.Check != checkReviewer {
			return Verdict{}, fmt.Errorf("decisión sobre %q, que no requiere revisor", id)
		}
	}
	e := &evaluation{f: f, out: out}
	defer e.cleanup()
	var rs []CriterionResult
	for _, c := range f.Task.Criteria {
		o, detail, err := e.check(c)
		if err != nil {
			return Verdict{}, fmt.Errorf("criterio %s: %w", c.ID, err)
		}
		rs = append(rs, CriterionResult{ID: c.ID, Mandatory: c.Mandatory, Outcome: o, Detail: detail})
	}
	return Decide(rs)
}

type evaluation struct {
	f       *Fixture
	out     Output
	temps   []string
	matched *review.Result
	badAns  string
}

func (e *evaluation) cleanup() {
	for _, d := range e.temps {
		os.RemoveAll(d)
	}
}

func outcome(ok bool, detail string) (Outcome, string, error) {
	if ok {
		return Met, "", nil
	}
	return NotMet, detail, nil
}

func (e *evaluation) check(c Criterion) (Outcome, string, error) {
	switch c.Check {
	case checkReviewer:
		d, ok := e.out.Decisions[c.ID]
		if !ok {
			return Pending, "sin decisión del revisor", nil
		}
		return outcome(d, "el revisor no lo corrobora")
	case "checkout_intacto":
		ok, detail, err := intact(e.out.Dir, e.f.Task.Source.SHA256)
		if err != nil {
			return "", "", err
		}
		return outcome(ok, detail)
	case "respuesta_entrada", "respuesta_archivos":
		return e.exploration(c.Check)
	case "go_test":
		ok, log, err := goTest(e.out.Dir)
		if err != nil {
			return "", "", err
		}
		return outcome(ok, "go test falla: "+tail(log))
	case "pruebas_ocultas":
		return e.hiddenTests(c.Args[0])
	case "prueba_regresion":
		return e.regression()
	case "archivos_permitidos":
		return e.touched(c.Args)
	case "diag_reproduccion", "diag_causa_raiz", "diag_descarte":
		return e.diagnosis(c.Check)
	case "rev_sin_omisiones", "rev_evidencia", "rev_sin_falsos_positivos":
		return e.review(c.Check)
	}
	return "", "", fmt.Errorf("chequeo desconocido %q", c.Check)
}

// --- respuestas JSON ---

// decodeAnswer lee la respuesta como JSON entero o, si no, el ultimo bloque
// ```json de la respuesta.
func decodeAnswer(ans []byte, v any) error {
	s := strings.TrimSpace(string(ans))
	if !strings.HasPrefix(s, "{") {
		i := strings.LastIndex(s, "```json")
		if i < 0 {
			return errors.New("la respuesta no trae un bloque JSON")
		}
		s = s[i+len("```json"):]
		j := strings.Index(s, "```")
		if j < 0 {
			return errors.New("bloque JSON sin cerrar")
		}
		s = s[:j]
	}
	return json.Unmarshal([]byte(s), v)
}

func normPath(p string) string {
	return path.Clean(strings.ReplaceAll(strings.TrimSpace(p), `\`, "/"))
}

func overlaps(a1, a2, b1, b2 int) bool {
	return a1 >= 1 && a2 >= a1 && a1 <= b2 && b1 <= a2
}

func (e *evaluation) key(v any) error {
	b, err := e.f.read(e.f.Task.Key)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(b, v); err != nil {
		return fmt.Errorf("clave: %w", err)
	}
	return nil
}

type location struct {
	File   string `json:"file"`
	Symbol string `json:"symbol"`
}

func (e *evaluation) exploration(check string) (Outcome, string, error) {
	var key, ans struct {
		Entry location `json:"entry"`
		Files []string `json:"files"`
	}
	if err := e.key(&key); err != nil {
		return "", "", err
	}
	if err := decodeAnswer(e.out.Answer, &ans); err != nil {
		return NotMet, "respuesta inválida: " + err.Error(), nil
	}
	if check == "respuesta_entrada" {
		ok := normPath(ans.Entry.File) == normPath(key.Entry.File) && strings.TrimSpace(ans.Entry.Symbol) == key.Entry.Symbol
		return outcome(ok, fmt.Sprintf("entrada %s:%s", ans.Entry.File, ans.Entry.Symbol))
	}
	cited := map[string]bool{}
	for _, p := range ans.Files {
		cited[normPath(p)] = true
	}
	var missing []string
	for _, p := range key.Files {
		if !cited[normPath(p)] {
			missing = append(missing, p)
		}
	}
	return outcome(len(missing) == 0, fmt.Sprintf("faltan %v", missing))
}

// --- cambios de codigo ---

// goTest corre go test en dir. Una salida distinta de cero es prueba fallida;
// no poder ejecutar go es error del arnes.
func goTest(dir string, args ...string) (bool, string, error) {
	cmd := exec.Command("go", append([]string{"test", "-count=1"}, append(args, "./...")...)...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GOWORK=off", "GOFLAGS=", "GOTOOLCHAIN=local")
	out, err := cmd.CombinedOutput()
	var exit *exec.ExitError
	switch {
	case err == nil:
		return true, string(out), nil
	case errors.As(err, &exit):
		return false, string(out), nil
	}
	return false, string(out), err
}

func tail(s string) string {
	if len(s) > 400 {
		return "…" + s[len(s)-400:]
	}
	return s
}

func (e *evaluation) tempCopy(src string) (string, error) {
	d, err := os.MkdirTemp("", "opsx-eval-")
	if err != nil {
		return "", err
	}
	e.temps = append(e.temps, d)
	if src == "" {
		return d, e.f.materialize(filepath.Join(d, "w"))
	}
	return d, copyTree(src, filepath.Join(d, "w"))
}

func (e *evaluation) hiddenTests(hiddenDir string) (Outcome, string, error) {
	d, err := e.tempCopy(e.out.Dir)
	if err != nil {
		return "", "", err
	}
	w := filepath.Join(d, "w")
	src := e.f.path(hiddenDir)
	err = filepath.WalkDir(src, func(p string, de os.DirEntry, err error) error {
		if err != nil || de.IsDir() {
			return err
		}
		rel, _ := filepath.Rel(src, p)
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		target := filepath.Join(w, strings.TrimSuffix(rel, ".txt"))
		os.MkdirAll(filepath.Dir(target), 0o755)
		return os.WriteFile(target, b, 0o644)
	})
	if err != nil {
		return "", "", err
	}
	ok, log, err := goTest(w)
	if err != nil {
		return "", "", err
	}
	return outcome(ok, "pruebas ocultas fallan: "+tail(log))
}

// changes devuelve los archivos agregados, modificados o borrados en el
// checkout respecto del estado congelado.
func (e *evaluation) changes() (map[string]string, []string, error) {
	pristine, err := e.tempCopy("")
	if err != nil {
		return nil, nil, err
	}
	before, err := fileHashes(filepath.Join(pristine, "w"))
	if err != nil {
		return nil, nil, err
	}
	after, err := fileHashes(e.out.Dir)
	if err != nil {
		return nil, nil, err
	}
	var changed []string
	for p, h := range after {
		if before[p] != h {
			changed = append(changed, p)
		}
	}
	for p := range before {
		if _, ok := after[p]; !ok {
			changed = append(changed, p)
		}
	}
	return after, changed, nil
}

func (e *evaluation) touched(allowed []string) (Outcome, string, error) {
	_, changed, err := e.changes()
	if err != nil {
		return "", "", err
	}
	var extra []string
	for _, p := range changed {
		if !contains(allowed, p) {
			extra = append(extra, p)
		}
	}
	return outcome(len(extra) == 0, fmt.Sprintf("archivos no permitidos: %v", extra))
}

// regression: las pruebas agregadas o modificadas, puestas sobre el codigo
// original, deben fallar.
func (e *evaluation) regression() (Outcome, string, error) {
	after, changed, err := e.changes()
	if err != nil {
		return "", "", err
	}
	var tests []string
	for _, p := range changed {
		if _, exists := after[p]; exists && strings.HasSuffix(p, "_test.go") {
			tests = append(tests, p)
		}
	}
	if len(tests) == 0 {
		return NotMet, "no hay prueba agregada ni modificada", nil
	}
	d, err := e.tempCopy("")
	if err != nil {
		return "", "", err
	}
	w := filepath.Join(d, "w")
	for _, p := range tests {
		b, err := os.ReadFile(filepath.Join(e.out.Dir, filepath.FromSlash(p)))
		if err != nil {
			return "", "", err
		}
		if err := os.WriteFile(filepath.Join(w, filepath.FromSlash(p)), b, 0o644); err != nil {
			return "", "", err
		}
	}
	ok, _, err := goTest(w)
	if err != nil {
		return "", "", err
	}
	return outcome(!ok, fmt.Sprintf("las pruebas %v pasan con el código original", tests))
}

// --- diagnostico ---

var testName = regexp.MustCompile(`^Test[A-Za-z0-9_]*$`)

type span struct {
	File      string `json:"file"`
	StartLine int    `json:"start_line"`
	EndLine   int    `json:"end_line"`
}

func (e *evaluation) diagnosis(check string) (Outcome, string, error) {
	var key struct {
		RootCause   span `json:"root_cause"`
		Alternative struct {
			File string `json:"file"`
		} `json:"alternative"`
	}
	if err := e.key(&key); err != nil {
		return "", "", err
	}
	var ans struct {
		Reproduction struct {
			Test     string `json:"test"`
			Observed string `json:"observed"`
		} `json:"reproduction"`
		RootCause span `json:"root_cause"`
		Discarded struct {
			File     string `json:"file"`
			Evidence string `json:"evidence"`
		} `json:"discarded"`
	}
	if err := decodeAnswer(e.out.Answer, &ans); err != nil {
		return NotMet, "respuesta inválida: " + err.Error(), nil
	}
	switch check {
	case "diag_reproduccion":
		r := ans.Reproduction
		if !testName.MatchString(r.Test) || strings.TrimSpace(r.Observed) == "" {
			return NotMet, "reproducción sin prueba válida o sin fragmento observado", nil
		}
		ok, log, err := goTest(e.out.Dir, "-run", "^"+r.Test+"$")
		if err != nil {
			return "", "", err
		}
		return outcome(!ok && strings.Contains(log, "--- FAIL: "+r.Test) && strings.Contains(log, r.Observed),
			"la prueba citada no falla o su salida no contiene el fragmento")
	case "diag_causa_raiz":
		a, k := ans.RootCause, key.RootCause
		ok := normPath(a.File) == normPath(k.File) && overlaps(a.StartLine, a.EndLine, k.StartLine, k.EndLine)
		return outcome(ok, fmt.Sprintf("causa raíz en %s:%d-%d", a.File, a.StartLine, a.EndLine))
	}
	ok := normPath(ans.Discarded.File) == normPath(key.Alternative.File) && strings.TrimSpace(ans.Discarded.Evidence) != ""
	return outcome(ok, "no descarta el síntoma alternativo de la clave con evidencia")
}

// --- revision de un diff sembrado ---

type reviewFinding struct {
	ID        string `json:"id"`
	File      string `json:"file"`
	StartLine int    `json:"start_line"`
	EndLine   int    `json:"end_line"`
	Mechanism string `json:"mechanism"`
	Evidence  string `json:"evidence"`
}

func (e *evaluation) matchReview() error {
	if e.matched != nil || e.badAns != "" {
		return nil
	}
	var key struct {
		Defects []reviewFinding `json:"defects"`
	}
	if err := e.key(&key); err != nil {
		return err
	}
	var ans struct {
		Findings []reviewFinding `json:"findings"`
	}
	if err := decodeAnswer(e.out.Answer, &ans); err != nil {
		e.badAns = "respuesta inválida: " + err.Error()
	}
	ids := map[string]bool{}
	for _, f := range ans.Findings {
		if f.ID == "" || ids[f.ID] || f.StartLine < 1 || f.EndLine < f.StartLine {
			e.badAns = fmt.Sprintf("hallazgo inválido %+v", f)
		}
		ids[f.ID] = true
	}
	for id := range e.out.Review {
		if !ids[id] {
			return fmt.Errorf("decisión del revisor sobre hallazgo inexistente %q", id)
		}
	}
	if e.badAns != "" {
		return nil
	}
	defects := make([]review.Defect, len(key.Defects))
	for i, d := range key.Defects {
		defects[i] = review.Defect{ID: d.ID, File: d.File, StartLine: d.StartLine, EndLine: d.EndLine, Mechanism: d.Mechanism}
	}
	findings := make([]review.Finding, len(ans.Findings))
	for i, f := range ans.Findings {
		call := e.out.Review[f.ID]
		mech := f.Mechanism
		if strings.TrimSpace(call.Mechanism) != "" {
			mech = call.Mechanism // el revisor decide el mecanismo
		}
		findings[i] = review.Finding{ID: f.ID, File: f.File, StartLine: f.StartLine, EndLine: f.EndLine,
			Mechanism: mech, Evidence: f.Evidence, Disposition: review.Disposition(call.Disposition)}
	}
	r, err := review.Match(defects, findings)
	if err != nil {
		return err
	}
	e.matched = &r
	return nil
}

func (e *evaluation) review(check string) (Outcome, string, error) {
	if err := e.matchReview(); err != nil {
		return "", "", err
	}
	if e.badAns != "" {
		return NotMet, e.badAns, nil
	}
	r := e.matched
	switch check {
	case "rev_sin_omisiones":
		return outcome(len(r.Omissions) == 0, fmt.Sprintf("omisiones %v", r.Omissions))
	case "rev_evidencia":
		var bare []string
		for _, c := range r.Findings {
			if c.Class == review.TruePositive && strings.TrimSpace(c.Evidence) == "" {
				bare = append(bare, c.FindingID)
			}
		}
		return outcome(len(bare) == 0, fmt.Sprintf("verdaderos positivos sin evidencia %v", bare))
	}
	switch {
	case r.Counts.FalsePositives > 0:
		return NotMet, fmt.Sprintf("%d falsos positivos", r.Counts.FalsePositives), nil
	case r.Counts.Pending > 0:
		return Pending, fmt.Sprintf("%d hallazgos sin decisión del revisor", r.Counts.Pending), nil
	}
	return Met, "", nil
}
