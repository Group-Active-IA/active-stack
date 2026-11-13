package catalog

import (
	"bytes"
	_ "embed"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"
)

// Declaraciones del change opsx-task-context (ficha 05): umbrales, patrones
// sensibles, patrones de fallo, tabla de herramientas y tabla de carga de
// módulos. Viven en taskcontext.yaml, junto a harnesses.yaml, y Load() las
// valida: una declaración malformada es un error de build/release.

//go:embed taskcontext.yaml
var rawTaskContext []byte

// TaskContext es la declaración parseada y validada de taskcontext.yaml.
type TaskContext struct {
	OutputThreshold    OutputThreshold    `yaml:"output_threshold"`
	ReadThreshold      ReadThreshold      `yaml:"read_threshold"`
	FullReadConditions []string           `yaml:"full_read_conditions"`
	FailurePatterns    []FailurePattern   `yaml:"failure_patterns"`
	SensitivePatterns  []SensitivePattern `yaml:"sensitive_patterns"`
	ToolSelection      ToolSelection      `yaml:"tool_selection"`
	ModuleOrder        []string           `yaml:"module_order"`
	Modules            []Module           `yaml:"modules"`
	Parity             Parity             `yaml:"parity"`
}

// OutputThreshold es el umbral de salidas de comando (D3): se resume una
// salida que supere cualquiera de las dos unidades.
type OutputThreshold struct {
	Lines      int `yaml:"lines"`
	Characters int `yaml:"characters"`
}

// ReadThreshold es el umbral de lectura completa de un archivo (D11).
type ReadThreshold struct {
	Unit  string `yaml:"unit"`
	Value int    `yaml:"value"`
}

// FailurePattern declara los patrones de fallo de una herramienta (D3).
type FailurePattern struct {
	Tool     string   `yaml:"tool"`
	Patterns []string `yaml:"patterns"`
}

// SensitivePattern es un patrón de valor sensible (D14).
type SensitivePattern struct {
	ID       string   `yaml:"id"`
	Category string   `yaml:"category"`
	Kind     string   `yaml:"kind"` // line | block
	Regex    string   `yaml:"regex"`
	Begin    string   `yaml:"begin"`
	End      string   `yaml:"end"`
	Covers   []string `yaml:"covers"`
}

// ToolSelection es la tabla de herramientas por acción (D18).
type ToolSelection struct {
	Actions      []ToolAction  `yaml:"actions"`
	Integrations []Integration `yaml:"integrations"`
}

// ToolAction es una fila de la tabla de herramientas por acción.
type ToolAction struct {
	ID       string         `yaml:"id"`
	Base     []string       `yaml:"base"`
	Optional []OptionalTool `yaml:"optional"`
}

// OptionalTool es una integración opcional habilitada en una acción.
type OptionalTool struct {
	Integration string `yaml:"integration"`
	Label       string `yaml:"label"`
}

// Integration es una integración opcional con su nivel por cliente.
type Integration struct {
	ID              string                 `yaml:"id"`
	Default         string                 `yaml:"default"`
	Kind            string                 `yaml:"kind"`
	Actions         []string               `yaml:"actions"`
	RequiresHarness string                 `yaml:"requires_harness"`
	Levels          map[string]ClientLevel `yaml:"levels"`
}

// ClientLevel es el nivel de control alcanzable por acción en un cliente.
type ClientLevel struct {
	Level     string `yaml:"level"`
	Mechanism string `yaml:"mechanism"`
}

// Module es una fila de la tabla de carga de módulos (D19).
type Module struct {
	ID           string            `yaml:"id"`
	Toggle       string            `yaml:"toggle"`
	Actions      map[string]string `yaml:"actions"`
	Conditions   map[string]string `yaml:"conditions"`
	OptionalVia  string            `yaml:"optional_via"`
	OmittedIn    []string          `yaml:"omitted_in"`
	With         string            `yaml:"with"`
	RequiresFlag string            `yaml:"requires_flag"`
	Specialist   Specialist        `yaml:"specialist"`
}

// Specialist dice si el especialista recibe el módulo en el encargo.
type Specialist struct {
	Receives bool     `yaml:"receives"`
	Actions  []string `yaml:"actions"`
	Note     string   `yaml:"note"`
}

// Actions es el conjunto cerrado de acciones del ciclo OPSX (D10).
var Actions = []string{"explore", "propose", "apply", "verify", "archive"}

// requiredFullReadConditions es la lista cerrada de D15.
var requiredFullReadConditions = []string{
	"within-threshold",
	"rewrite-whole-file",
	"action-artifact",
	"empty-bounded-read",
	"no-range-covers",
}

// ParseTaskContext parsea y valida una declaración de task-context.
func ParseTaskContext(data []byte) (*TaskContext, error) {
	var tc TaskContext
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	if err := dec.Decode(&tc); err != nil {
		return nil, fmt.Errorf("catalog: task-context parse: %w", err)
	}
	if err := tc.validate(); err != nil {
		return nil, err
	}
	return &tc, nil
}

func (tc *TaskContext) validate() error {
	if tc.OutputThreshold.Lines <= 0 || tc.OutputThreshold.Characters <= 0 {
		return fmt.Errorf("catalog: output_threshold needs both lines and characters (> 0), got %+v", tc.OutputThreshold)
	}
	if tc.ReadThreshold.Unit != "lines" || tc.ReadThreshold.Value <= 0 {
		return fmt.Errorf("catalog: read_threshold must be expressed in lines with a positive value, got %+v", tc.ReadThreshold)
	}
	if !sameSet(tc.FullReadConditions, requiredFullReadConditions) {
		return fmt.Errorf("catalog: full_read_conditions must be exactly the closed list %v, got %v", requiredFullReadConditions, tc.FullReadConditions)
	}
	if err := tc.validateFailurePatterns(); err != nil {
		return err
	}
	if err := tc.validateSensitivePatterns(); err != nil {
		return err
	}
	if err := tc.validateToolSelection(); err != nil {
		return err
	}
	if err := tc.validateModules(); err != nil {
		return err
	}
	return tc.validateParity()
}

// levelNames asocia cada nivel de control de D18 con su nombre.
var levelNames = map[string]string{
	"a": "solo recomendada",
	"b": "fuera del contexto",
	"c": "restringida por el cliente",
}

// LevelName devuelve el nombre del nivel de control (a, b o c) o "" si no existe.
func LevelName(level string) string { return levelNames[level] }

// toolClients son los clientes para los que toda integración declara nivel.
var toolClients = []string{"claude", "codex"}

func (tc *TaskContext) validateToolSelection() error {
	ts := tc.ToolSelection
	var ids []string
	for _, a := range ts.Actions {
		ids = append(ids, a.ID)
	}
	if !slices.Equal(ids, Actions) {
		return fmt.Errorf("catalog: tool_selection.actions must be exactly %v in that order, got %v", Actions, ids)
	}
	byID := map[string]Integration{}
	for _, in := range ts.Integrations {
		if in.ID == "" {
			return fmt.Errorf("catalog: tool_selection integration with empty id")
		}
		if _, dup := byID[in.ID]; dup {
			return fmt.Errorf("catalog: tool_selection: duplicate integration %q", in.ID)
		}
		byID[in.ID] = in
		if in.Default != "disabled" {
			return fmt.Errorf("catalog: integration %q has default %q, optional integrations are disabled by default", in.ID, in.Default)
		}
		if len(in.Actions) == 0 {
			return fmt.Errorf("catalog: optional integration %q declares no action", in.ID)
		}
		for _, a := range in.Actions {
			if !slices.Contains(Actions, a) {
				return fmt.Errorf("catalog: integration %q lists unknown action %q", in.ID, a)
			}
		}
		for _, client := range toolClients {
			lv, ok := in.Levels[client]
			if !ok {
				return fmt.Errorf("catalog: integration %q has no level for client %q", in.ID, client)
			}
			if LevelName(lv.Level) == "" {
				return fmt.Errorf("catalog: integration %q client %q has invalid level %q (want a|b|c)", in.ID, client, lv.Level)
			}
			if lv.Mechanism == "" {
				return fmt.Errorf("catalog: integration %q client %q has no mechanism", in.ID, client)
			}
			if lv.Level == "c" && strings.Contains(lv.Mechanism, "allowed-tools") {
				return fmt.Errorf("catalog: integration %q client %q: allowed-tools does not restrict tools and cannot be level (c)", in.ID, client)
			}
		}
	}
	for _, a := range ts.Actions {
		for _, o := range a.Optional {
			in, ok := byID[o.Integration]
			if !ok {
				return fmt.Errorf("catalog: action %q enables undeclared integration %q", a.ID, o.Integration)
			}
			if !slices.Contains(in.Actions, a.ID) {
				return fmt.Errorf("catalog: action %q enables integration %q, which does not list that action", a.ID, o.Integration)
			}
		}
	}
	return nil
}

func (tc *TaskContext) validateFailurePatterns() error {
	seen := map[string]bool{}
	for _, fp := range tc.FailurePatterns {
		if fp.Tool == "" || len(fp.Patterns) == 0 {
			return fmt.Errorf("catalog: failure_patterns entry %q needs a tool and at least one pattern", fp.Tool)
		}
		if seen[fp.Tool] {
			return fmt.Errorf("catalog: failure_patterns: duplicate tool %q", fp.Tool)
		}
		seen[fp.Tool] = true
		for _, p := range fp.Patterns {
			if _, err := regexp.Compile(p); err != nil {
				return fmt.Errorf("catalog: failure_patterns %q: invalid pattern %q: %w", fp.Tool, p, err)
			}
		}
	}
	return nil
}

func (tc *TaskContext) validateSensitivePatterns() error {
	seen := map[string]bool{}
	for _, p := range tc.SensitivePatterns {
		if p.ID == "" || p.Category == "" {
			return fmt.Errorf("catalog: sensitive pattern %q needs an id and a category", p.ID)
		}
		if seen[p.ID] {
			return fmt.Errorf("catalog: sensitive_patterns: duplicate id %q", p.ID)
		}
		seen[p.ID] = true
		switch p.Kind {
		case "line":
			if p.Regex == "" {
				return fmt.Errorf("catalog: sensitive pattern %q (kind line) needs a regex", p.ID)
			}
			if _, err := regexp.Compile(p.Regex); err != nil {
				return fmt.Errorf("catalog: sensitive pattern %q: invalid regex: %w", p.ID, err)
			}
		case "block":
			if p.Begin == "" || p.End == "" {
				return fmt.Errorf("catalog: sensitive pattern %q (kind block) needs begin and end", p.ID)
			}
			for _, r := range []string{p.Begin, p.End} {
				if _, err := regexp.Compile(r); err != nil {
					return fmt.Errorf("catalog: sensitive pattern %q: invalid regex: %w", p.ID, err)
				}
			}
		default:
			return fmt.Errorf("catalog: sensitive pattern %q has unknown kind %q (want line|block)", p.ID, p.Kind)
		}
	}
	return nil
}

// sameSet reporta si got tiene exactamente los elementos de want, sin
// repetidos ni extras (el orden no importa).
func sameSet(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	g := slices.Clone(got)
	w := slices.Clone(want)
	slices.Sort(g)
	slices.Sort(w)
	return slices.Equal(g, w)
}
