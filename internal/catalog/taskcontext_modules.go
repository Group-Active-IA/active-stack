package catalog

import (
	"fmt"
	"slices"
)

// Carga de módulos por acción (D19 y entrega-03-04 §3): validación de la tabla
// declarada y función pura que la resuelve.

// Role es quién recibe los módulos: el coordinador (lee la skill) o un
// especialista (recibe lo que el coordinador pega en el encargo).
type Role string

const (
	RoleCoordinator Role = "coordinator"
	RoleSpecialist  Role = "specialist"
)

// Niveles de carga de un módulo en una acción.
const (
	LevelMandatory = "mandatory"
	LevelOptional  = "optional"
	// LevelAssigned es lo que recibe un especialista: lo decide el coordinador
	// al armar el encargo, según la condición de la columna Especialista.
	LevelAssigned = "assigned"
)

// Query es la entrada de Resolve.
type Query struct {
	Action string
	Role   Role
	// Toggles activos del sdd-orchestrator (tdd, engram, model-routing,
	// delegation, governance). El orden no importa.
	Toggles []string
	// StrictTDD: el proyecto exige Strict TDD y hay un test runner.
	StrictTDD bool
	// Delegating: el coordinador va a delegar en esta acción.
	Delegating bool
	// Client: "claude" o "codex". Vacío = sin omisiones por cliente.
	Client string
}

// Entry es un módulo resuelto.
type Entry struct {
	Module    string
	Level     string
	Condition string
}

// Resolution es el resultado de Resolve. Blocked nombra los módulos
// obligatorios cuyo toggle está inactivo: la acción se detiene (D19).
type Resolution struct {
	Entries []Entry
	Blocked []string
}

// Resolve devuelve los módulos de una acción para un rol. Es una función pura:
// mismas entradas, mismo resultado, sin depender del orden de los toggles ni
// del orden de declaración de los módulos (se ordena por module_order).
func (tc *TaskContext) Resolve(q Query) (Resolution, error) {
	if !slices.Contains(Actions, q.Action) {
		return Resolution{}, fmt.Errorf("catalog: resolve: unknown action %q (want %v)", q.Action, Actions)
	}
	if q.Role != RoleCoordinator && q.Role != RoleSpecialist {
		return Resolution{}, fmt.Errorf("catalog: resolve: unknown role %q", q.Role)
	}
	active := make(map[string]bool, len(q.Toggles))
	for _, t := range q.Toggles {
		active[t] = true
	}
	byID := make(map[string]Module, len(tc.Modules))
	for _, m := range tc.Modules {
		byID[m.ID] = m
	}

	var res Resolution
	for _, id := range tc.ModuleOrder {
		m, ok := byID[id]
		if !ok {
			continue
		}
		toggleOn := m.Toggle == "" || active[m.Toggle]
		if q.Role == RoleSpecialist {
			if e, ok := specialistEntry(m, q, toggleOn); ok {
				res.Entries = append(res.Entries, e)
			}
			continue
		}
		if slices.Contains(m.OmittedIn, q.Client) {
			continue // límite declarado del cliente, no un bloqueo
		}
		level := m.Actions[q.Action]
		if m.RequiresFlag == "strict-tdd" && !q.StrictTDD {
			continue
		}
		if level == "" && m.OptionalVia == "delegate-row" && q.Delegating {
			level = LevelOptional
			if m.With != "" && !active[m.With] && !active[toggleOf(byID, m.With)] {
				continue
			}
		}
		if level == "" {
			continue
		}
		if !toggleOn {
			if level == LevelMandatory {
				res.Blocked = append(res.Blocked, m.ID)
			}
			continue
		}
		res.Entries = append(res.Entries, Entry{Module: m.ID, Level: level, Condition: m.Conditions[q.Action]})
	}
	return res, nil
}

func toggleOf(byID map[string]Module, id string) string { return byID[id].Toggle }

// specialistEntry decide si un especialista recibe el módulo en su encargo.
func specialistEntry(m Module, q Query, toggleOn bool) (Entry, bool) {
	if !m.Specialist.Receives || !toggleOn {
		return Entry{}, false
	}
	if len(m.Specialist.Actions) > 0 && !slices.Contains(m.Specialist.Actions, q.Action) {
		return Entry{}, false
	}
	if m.RequiresFlag == "strict-tdd" && !q.StrictTDD {
		return Entry{}, false
	}
	// Un módulo que la acción no usa (p. ej. artifacts en explore) no se pega.
	if m.Actions[q.Action] == "" && m.OptionalVia == "" {
		return Entry{}, false
	}
	return Entry{Module: m.ID, Level: LevelAssigned, Condition: m.Specialist.Note}, true
}

func (tc *TaskContext) validateModules() error {
	byID := make(map[string]bool, len(tc.Modules))
	for _, m := range tc.Modules {
		if m.ID == "" {
			return fmt.Errorf("catalog: module with empty id")
		}
		if byID[m.ID] {
			return fmt.Errorf("catalog: modules: duplicate id %q", m.ID)
		}
		byID[m.ID] = true
	}
	// module_order es una permutación exacta de los módulos declarados.
	if len(tc.ModuleOrder) != len(tc.Modules) {
		return fmt.Errorf("catalog: module_order lists %d modules, %d declared", len(tc.ModuleOrder), len(tc.Modules))
	}
	seenOrder := map[string]bool{}
	for _, id := range tc.ModuleOrder {
		if !byID[id] || seenOrder[id] {
			return fmt.Errorf("catalog: module_order has unknown or repeated module %q", id)
		}
		seenOrder[id] = true
	}
	for _, m := range tc.Modules {
		if len(m.Actions) == 0 && m.OptionalVia == "" {
			return fmt.Errorf("catalog: module %q declares no actions (and no optional_via row)", m.ID)
		}
		for a, level := range m.Actions {
			if !slices.Contains(Actions, a) {
				return fmt.Errorf("catalog: module %q lists unknown action %q", m.ID, a)
			}
			if level != LevelMandatory && level != LevelOptional {
				return fmt.Errorf("catalog: module %q action %q has invalid level %q (want mandatory|optional)", m.ID, a, level)
			}
		}
		for a := range m.Conditions {
			if _, ok := m.Actions[a]; !ok {
				return fmt.Errorf("catalog: module %q has a condition for action %q it does not load in", m.ID, a)
			}
		}
		switch m.OptionalVia {
		case "", "delegate-row", "session-row":
		default:
			return fmt.Errorf("catalog: module %q has unknown optional_via %q", m.ID, m.OptionalVia)
		}
		if m.With != "" && !byID[m.With] {
			return fmt.Errorf("catalog: module %q is loaded with unknown module %q", m.ID, m.With)
		}
		for _, c := range m.OmittedIn {
			if !slices.Contains(toolClients, c) {
				return fmt.Errorf("catalog: module %q omitted in unknown client %q", m.ID, c)
			}
		}
		if m.Specialist.Note == "" {
			return fmt.Errorf("catalog: module %q needs a specialist note", m.ID)
		}
		for _, a := range m.Specialist.Actions {
			if !slices.Contains(Actions, a) {
				return fmt.Errorf("catalog: module %q specialist lists unknown action %q", m.ID, a)
			}
		}
	}
	return nil
}

// validateAgainst comprueba que los toggles de los módulos existan en el
// sdd-orchestrator del catálogo (Load() lo llama con sus toggles).
func (tc *TaskContext) validateAgainst(orchestratorToggles []string) error {
	for _, m := range tc.Modules {
		if m.Toggle != "" && !slices.Contains(orchestratorToggles, m.Toggle) {
			return fmt.Errorf("catalog: module %q uses toggle %q, which sdd-orchestrator does not declare (%v)", m.ID, m.Toggle, orchestratorToggles)
		}
	}
	return nil
}
