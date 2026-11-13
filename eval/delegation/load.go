package delegation

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Group-Active-IA/active-stack/eval/evidence"
)

// Los fixtures viven en testdata/{retornos,decisiones,combinacion}. Las copias redactadas se leen
// como bytes, relativas a testdata/. Los recibos usan el JSON de evidence.Receipt (06) sin
// traducir: el módulo en inglés los nombra run, tree, command, result, requirements y output.

// envFile replica el entorno de los casos de eval/evidence (sus tipos de archivo no se exportan).
type envFile struct {
	Dir           string            `json:"directorio"`
	Commit        string            `json:"commit"`
	StateFP       *Fingerprint      `json:"huella_estado,omitempty"`
	Copies        map[string]string `json:"copias"`
	Expired       []string          `json:"vencidas"`
	Corroboration []struct {
		Run      string                 `json:"corrida"`
		Command  string                 `json:"comando"`
		Dir      string                 `json:"directorio"`
		Commit   string                 `json:"commit"`
		StateFP  *Fingerprint           `json:"huella_estado,omitempty"`
		ExitCode *int                   `json:"codigo_de_salida"`
		Source   string                 `json:"fuente"`
		Verdict  evidence.Corroboration `json:"veredicto"`
	} `json:"corroboracion"`
	Sufficiency []struct {
		Run         string               `json:"corrida"`
		Requirement string               `json:"requisito"`
		Command     string               `json:"comando"`
		Verdict     evidence.Sufficiency `json:"veredicto"`
	} `json:"suficiencia"`
	DetailNeeded []string `json:"detalle_descartado_necesario"`
}

func (f envFile) env(testdata string) (evidence.Env, error) {
	env := evidence.Env{
		Dir: f.Dir, Commit: f.Commit, StateFP: f.StateFP,
		Copies: map[string][]byte{}, Expired: set(f.Expired), DetailNeeded: set(f.DetailNeeded),
	}
	for _, c := range f.Corroboration {
		env.Corroboration = append(env.Corroboration, evidence.CorroborationRecord{
			Run: c.Run, Command: c.Command, Dir: c.Dir, Commit: c.Commit, StateFP: c.StateFP,
			ExitCode: c.ExitCode, Source: c.Source, Verdict: c.Verdict,
		})
	}
	for _, s := range f.Sufficiency {
		env.Sufficiency = append(env.Sufficiency, evidence.SufficiencyRecord{
			Run: s.Run, Requirement: s.Requirement, Command: s.Command, Verdict: s.Verdict,
		})
	}
	for run, rel := range f.Copies {
		b, err := os.ReadFile(filepath.Join(testdata, rel))
		if err != nil {
			return evidence.Env{}, fmt.Errorf("copia de %s: %w", run, err)
		}
		env.Copies[run] = b
	}
	return env, nil
}

func set(xs []string) map[string]bool {
	m := make(map[string]bool, len(xs))
	for _, x := range xs {
		m[x] = true
	}
	return m
}

// ReturnCase es un fixture de 4.1: encargo, retorno crudo, lo observado y el veredicto esperado.
type ReturnCase struct {
	ID          string
	Description string
	Assignment  Assignment
	Raw         []byte
	Observed    Observed
	Want        ReturnVerdict
}

// LoadReturnCase lee un fixture de testdata/retornos. El retorno es "retorno" (objeto JSON) o
// "retorno_texto" (prosa libre).
func LoadReturnCase(path string) (ReturnCase, error) {
	var f struct {
		ID          string          `json:"id"`
		Description string          `json:"descripcion"`
		Assignment  Assignment      `json:"encargo"`
		Return      json.RawMessage `json:"retorno"`
		ReturnText  string          `json:"retorno_texto"`
		Observed    struct {
			AtStart     []Change `json:"en_partida"`
			Final       []Change `json:"final"`
			CriteriaMet *bool    `json:"criterios_cumplidos"`
			Evidence    envFile  `json:"evidencia"`
		} `json:"observado"`
		Want struct {
			Status   evidence.Status `json:"estado"`
			Step     string          `json:"paso"`
			Action   string          `json:"accion"`
			Escalate []string        `json:"escala_con"`
		} `json:"esperado"`
	}
	if err := readJSON(path, &f); err != nil {
		return ReturnCase{}, err
	}
	raw := []byte(f.ReturnText)
	if len(f.Return) > 0 {
		raw = f.Return
	}
	env, err := f.Observed.Evidence.env(testdataOf(path))
	if err != nil {
		return ReturnCase{}, fmt.Errorf("%s: %w", path, err)
	}
	return ReturnCase{
		ID: f.ID, Description: f.Description, Assignment: f.Assignment, Raw: raw,
		Observed: Observed{AtStart: f.Observed.AtStart, Final: f.Observed.Final,
			CriteriaMet: f.Observed.CriteriaMet, Evidence: env},
		Want: ReturnVerdict{Status: f.Want.Status, Step: f.Want.Step, Action: f.Want.Action,
			Escalate: f.Want.Escalate},
	}, nil
}

// DecisionCase es un fixture de 4.2.
type DecisionCase struct {
	ID          string
	Description string
	Task        Task
	Want        Decision
}

// LoadDecisionCase lee un fixture de testdata/decisiones.
func LoadDecisionCase(path string) (DecisionCase, error) {
	var f struct {
		ID          string   `json:"id"`
		Description string   `json:"descripcion"`
		Task        Task     `json:"tarea"`
		Want        Decision `json:"esperado"`
	}
	if err := readJSON(path, &f); err != nil {
		return DecisionCase{}, err
	}
	return DecisionCase{ID: f.ID, Description: f.Description, Task: f.Task, Want: f.Want}, nil
}

// CombinationCase es un fixture de 4.3.
type CombinationCase struct {
	ID          string
	Description string
	Assignments []Assignment
	Integration Integration
	Want        CombinedVerdict
}

// LoadCombinationCase lee un fixture de testdata/combinacion.
func LoadCombinationCase(path string) (CombinationCase, error) {
	var f struct {
		ID          string       `json:"id"`
		Description string       `json:"descripcion"`
		Assignments []Assignment `json:"encargos"`
		Integration struct {
			Final     []Change           `json:"final"`
			Scenarios []string           `json:"escenarios"`
			Receipts  []evidence.Receipt `json:"recibos"`
			Evidence  envFile            `json:"evidencia"`
		} `json:"integracion"`
		Want struct {
			Status    evidence.Status `json:"estado"`
			Integrate bool            `json:"integrar"`
			FailedFor []string        `json:"fallo_contra"`
		} `json:"esperado"`
	}
	if err := readJSON(path, &f); err != nil {
		return CombinationCase{}, err
	}
	env, err := f.Integration.Evidence.env(testdataOf(path))
	if err != nil {
		return CombinationCase{}, fmt.Errorf("%s: %w", path, err)
	}
	return CombinationCase{
		ID: f.ID, Description: f.Description, Assignments: f.Assignments,
		Integration: Integration{Final: f.Integration.Final, Scenarios: f.Integration.Scenarios,
			Receipts: f.Integration.Receipts, Evidence: env},
		Want: CombinedVerdict{Status: f.Want.Status, Integrate: f.Want.Integrate, FailedFor: f.Want.FailedFor},
	}, nil
}

func readJSON(path string, v any) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(raw, v); err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	return nil
}

// testdataOf devuelve testdata/ para testdata/<grupo>/<caso>.json.
func testdataOf(path string) string {
	return filepath.Dir(filepath.Dir(path))
}
