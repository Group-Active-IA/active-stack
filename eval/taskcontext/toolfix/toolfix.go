// Package toolfix contiene los fixtures de selección de herramientas y su
// evaluador (change opsx-task-context, D18): qué integraciones habilita cada
// acción, uso fuera de clase, integración habilitada pero ausente, invocación
// repetida y sondeo de una integración ya informada como no disponible. Vive
// fuera del binario (CHANGES.md:49).
package toolfix

import (
	"regexp"
	"sort"
	"strings"

	"github.com/Group-Active-IA/active-stack/internal/catalog"
)

// EnabledFor devuelve, ordenadas, las integraciones opcionales que la tabla del
// catálogo habilita en una acción. Las integraciones de diseño se suman solo si
// el harness design-mockups está activo.
func EnabledFor(tc *catalog.TaskContext, action string, designMockups bool) []string {
	set := map[string]bool{}
	for _, a := range tc.ToolSelection.Actions {
		if a.ID != action {
			continue
		}
		for _, o := range a.Optional {
			set[o.Integration] = true
		}
	}
	if designMockups {
		for _, in := range tc.ToolSelection.Integrations {
			if in.RequiresHarness == "design-mockups" {
				for _, a := range in.Actions {
					if a == action {
						set[in.ID] = true
					}
				}
			}
		}
	}
	var out []string
	for id := range set {
		out = append(out, id)
	}
	sort.Strings(out)
	return out
}

// Env es el entorno de una tarea.
type Env struct {
	Action  string
	Enabled []string
	// Unavailable son las integraciones habilitadas que no responden.
	Unavailable []string
}

func (e Env) enabled(id string) bool {
	for _, x := range e.Enabled {
		if x == id {
			return true
		}
	}
	return false
}

// Call es una invocación de una herramienta o integración.
type Call struct {
	Integration string
	Tool        string
	Input       string
	// State es la versión del estado que la llamada observa; dos llamadas
	// iguales con el mismo estado son una repetición.
	State int
	// Unavailable: la llamada reveló que la integración no responde.
	Unavailable bool
}

// Result es lo que cuenta el evaluador.
type Result struct {
	OutOfClass             []string // usos fuera de clase registrados con la línea fija
	UnrecordedOutOfClass   []string
	MissingUnavailableLine []string
	UnrecordedRepeats      []string
	Reprobed               []string
	PresentedAsBlocked     bool
}

var (
	outOfClassRe   = regexp.MustCompile(`(?m)^out-of-class use: (\S+) in (\S+)$`)
	unavailableRe  = regexp.MustCompile(`(?m)^unavailable: (\S+) — uncovered: ([^;]+); alternative: (.+)$`)
	unnecessaryRe  = regexp.MustCompile(`(?m)^unnecessary call: (\S+) (.+)$`)
	blockedWordsRe = regexp.MustCompile(`(?i)\b(blocked|restricted by the client)\b|bloquead`)
)

// Evaluate cuenta los incumplimientos de una tarea.
func Evaluate(env Env, calls []Call, report string) Result {
	var res Result
	outOfClass := map[string]bool{}
	for _, m := range outOfClassRe.FindAllStringSubmatch(report, -1) {
		outOfClass[m[1]+"|"+m[2]] = true
	}
	unavailableDeclared := map[string]bool{}
	for _, m := range unavailableRe.FindAllStringSubmatch(report, -1) {
		if strings.TrimSpace(m[2]) != "" && strings.TrimSpace(m[3]) != "" {
			unavailableDeclared[m[1]] = true
		}
	}
	unnecessary := unnecessaryRe.FindAllStringSubmatch(report, -1)

	type callKey struct {
		tool, input string
		state       int
	}
	seen := map[callKey]bool{}
	unavailableSeen := map[string]bool{}

	for _, c := range calls {
		if c.Integration != "" && !env.enabled(c.Integration) {
			if outOfClass[c.Integration+"|"+env.Action] {
				res.OutOfClass = append(res.OutOfClass, c.Integration)
			} else {
				res.UnrecordedOutOfClass = append(res.UnrecordedOutOfClass, c.Integration)
			}
		}
		if c.Integration != "" && unavailableSeen[c.Integration] {
			res.Reprobed = append(res.Reprobed, c.Integration)
		}
		if c.Unavailable || (c.Integration != "" && isIn(env.Unavailable, c.Integration)) {
			unavailableSeen[c.Integration] = true
		}
		k := callKey{c.Tool, c.Input, c.State}
		if seen[k] {
			found := false
			for _, m := range unnecessary {
				if m[1] == c.Tool && strings.Contains(m[2], c.Input) {
					found = true
				}
			}
			if !found {
				res.UnrecordedRepeats = append(res.UnrecordedRepeats, c.Tool+" "+c.Input)
			}
		}
		seen[k] = true
	}
	// Una integración habilitada que el entorno declara no disponible exige la
	// línea aunque nunca se haya llamado: un MCP que no arranca no expone
	// herramientas (8.7, corrida real).
	for _, id := range env.Unavailable {
		if env.enabled(id) {
			unavailableSeen[id] = true
		}
	}
	for id := range unavailableSeen {
		if !unavailableDeclared[id] {
			res.MissingUnavailableLine = append(res.MissingUnavailableLine, id)
		}
	}
	sort.Strings(res.MissingUnavailableLine)
	// Una regla solo recomendada nunca se presenta como bloqueo.
	for _, line := range strings.Split(report, "\n") {
		if outOfClassRe.MatchString(line) {
			continue
		}
		for _, in := range uniqueIntegrations(calls) {
			if !env.enabled(in) && strings.Contains(line, in) && blockedWordsRe.MatchString(line) {
				res.PresentedAsBlocked = true
			}
		}
	}
	return res
}

func isIn(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}

func uniqueIntegrations(calls []Call) []string {
	set := map[string]bool{}
	for _, c := range calls {
		if c.Integration != "" {
			set[c.Integration] = true
		}
	}
	var out []string
	for k := range set {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// Want es el registro esperado de un escenario.
type Want struct {
	OutOfClass, UnrecordedOutOfClass, MissingUnavailableLine, UnrecordedRepeats, Reprobed int
	PresentedAsBlocked                                                                    bool
}

// Scenario es un fixture de herramientas con su registro esperado (lo corre la 8.7).
type Scenario struct {
	ID     string
	Env    Env
	Calls  []Call
	Report string
	Want   Want
}

// Scenarios devuelve los fixtures de herramientas derivados del catálogo.
func Scenarios(tc *catalog.TaskContext) []Scenario {
	verify := EnabledFor(tc, "verify", false)
	explore := EnabledFor(tc, "explore", false)
	unavail := func(extra string) string {
		return "unavailable: context7 — uncovered: current library docs; alternative: " + extra + "\n"
	}
	ctx := func(in string, state int) Call {
		return Call{Integration: "context7", Tool: "mcp__context7__query-docs", Input: in, State: state}
	}
	return []Scenario{
		// Instalada y no habilitada: el agente no la llama; el control es solo recomendado (a).
		{ID: "installed-not-enabled", Env: Env{Action: "verify", Enabled: verify},
			Calls: []Call{{Integration: "", Tool: "Bash", Input: "go test ./..."}}, Want: Want{}},
		{ID: "out-of-class-level-a-recorded", Env: Env{Action: "verify", Enabled: verify},
			Calls:  []Call{ctx("context7 docs for testing", 0)},
			Report: "out-of-class use: context7 in verify\n", Want: Want{OutOfClass: 1}},
		{ID: "out-of-class-level-a-unrecorded", Env: Env{Action: "verify", Enabled: verify},
			Calls: []Call{ctx("context7 docs for testing", 0)}, Want: Want{UnrecordedOutOfClass: 1}},
		{ID: "out-of-class-presented-as-blocked", Env: Env{Action: "verify", Enabled: verify},
			Calls:  []Call{ctx("docs", 0)},
			Report: "out-of-class use: context7 in verify\ncontext7 is blocked in verify by the client\n", Want: Want{OutOfClass: 1, PresentedAsBlocked: true}},
		// Habilitada pero ausente.
		{ID: "enabled-absent-with-alternative", Env: Env{Action: "explore", Enabled: explore, Unavailable: []string{"context7"}},
			Calls:  []Call{{Integration: "context7", Tool: "mcp__context7__query-docs", Input: "react hooks", Unavailable: true}},
			Report: unavail("read the vendored docs"), Want: Want{}},
		{ID: "enabled-absent-no-alternative", Env: Env{Action: "explore", Enabled: explore, Unavailable: []string{"context7"}},
			Calls:  []Call{{Integration: "context7", Tool: "mcp__context7__query-docs", Input: "react hooks", Unavailable: true}},
			Report: unavail("none"), Want: Want{}},
		// Invocación repetida.
		{ID: "repeated-call-no-state-change", Env: Env{Action: "explore", Enabled: explore},
			Calls: []Call{ctx("react hooks", 1), ctx("react hooks", 1)}, Want: Want{UnrecordedRepeats: 1}},
		{ID: "repeated-call-no-state-change-recorded", Env: Env{Action: "explore", Enabled: explore},
			Calls:  []Call{ctx("react hooks", 1), ctx("react hooks", 1)},
			Report: "unnecessary call: mcp__context7__query-docs react hooks\n", Want: Want{}},
		{ID: "repeated-call-after-state-change", Env: Env{Action: "explore", Enabled: explore},
			Calls: []Call{ctx("react hooks", 1), ctx("react hooks", 2)}, Want: Want{}},
		// Sondeo de una integración ya informada como no disponible.
		{ID: "probe-of-unavailable", Env: Env{Action: "explore", Enabled: explore, Unavailable: []string{"context7"}},
			Calls: []Call{
				{Integration: "context7", Tool: "mcp__context7__query-docs", Input: "a", Unavailable: true},
				{Integration: "context7", Tool: "mcp__context7__resolve-library-id", Input: "react"},
			},
			Report: unavail("none"), Want: Want{Reprobed: 1}},
	}
}
