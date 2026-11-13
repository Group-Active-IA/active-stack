package catalog

import (
	_ "embed"
	"fmt"
	"slices"
	"strings"
)

// Generación de texto desde el catálogo (opsx-task-context, tareas 5.5, 6.2 y
// 6.4). El agente no lee el catálogo embebido en el binario: lo que necesita
// saber llega como texto generado (tabla del SKILL.md) y los documentos de
// docs/ se regeneran y se contrastan en tests.

//go:embed taskcontext_actions_intro.txt
var actionTableIntro string

//go:embed taskcontext_actions_rules.txt
var actionTableRules string

// skillModulePath devuelve la forma de ruta de un módulo para un cliente
// (entrega-03-04 §3, «Rutas por cliente»).
func skillModulePath(client, id string) (string, error) {
	switch client {
	case "claude":
		return "`${CLAUDE_SKILL_DIR}/modules/" + id + ".md`", nil
	case "codex":
		return "`modules/" + id + ".md`", nil
	}
	return "", fmt.Errorf("catalog: no module path form declared for client %q", client)
}

// RenderActionTable genera el texto «Action to Modules» que va en el SKILL.md
// del cliente: la tabla acción → módulos con su regla de módulo ausente o
// vacío y la de compactación debajo.
func (tc *TaskContext) RenderActionTable(client string) (string, error) {
	return tc.renderActionTable(client, nil)
}

// RenderActionTableFor es la tabla de un cliente para una instalación concreta:
// un módulo cuyo toggle está inactivo no figura y una fila que queda sin
// módulos se omite (matriz común, «Reglas de la tabla»). Con todos los toggles
// activos coincide byte a byte con RenderActionTable.
func (tc *TaskContext) RenderActionTableFor(client string, toggles []string) (string, error) {
	if toggles == nil {
		toggles = []string{}
	}
	return tc.renderActionTable(client, toggles)
}

// AssignmentRule es la regla de D21 sobre el encargo: quién no ve CLAUDE.md
// recibe las reglas obligatorias y governance repetidas en el encargo.
func (tc *TaskContext) AssignmentRule(client string) (string, error) {
	if client != "claude" {
		return "", fmt.Errorf("catalog: no assignment rule declared for client %q (Codex pending on 04 task 2.5)", client)
	}
	return assignmentRuleClaude, nil
}

const assignmentRuleClaude = `### Assignments to agents without CLAUDE.md

Explore, Plan and any agent with ` + "`omitClaudeMd: true`" + ` start without CLAUDE.md, so the universal block does not reach them. In their assignment, repeat the mandatory rules of that block and its governance levels.
`

func (tc *TaskContext) renderActionTable(client string, toggles []string) (string, error) {
	active := func(m Module) bool {
		return toggles == nil || m.Toggle == "" || slices.Contains(toggles, m.Toggle)
	}
	var b strings.Builder
	b.WriteString(actionTableIntro)
	b.WriteString("| Action | Mandatory | Optional |\n|--------|-----------|----------|\n")

	cell := func(id, condition string) (string, error) {
		p, err := skillModulePath(client, id)
		if err != nil {
			return "", err
		}
		if condition != "" {
			p += " " + condition
		}
		return p, nil
	}
	join := func(items []string) string {
		if len(items) == 0 {
			return "—"
		}
		return strings.Join(items, "; ")
	}
	omitted := func(m Module) bool { return slices.Contains(m.OmittedIn, client) || !active(m) }

	for _, action := range Actions {
		var mandatory, optional []string
		for _, id := range tc.ModuleOrder {
			m := tc.module(id)
			level := m.Actions[action]
			if level == "" || omitted(m) {
				continue
			}
			c, err := cell(id, m.Conditions[action])
			if err != nil {
				return "", err
			}
			if level == LevelMandatory {
				mandatory = append(mandatory, c)
			} else {
				optional = append(optional, c)
			}
		}
		if toggles != nil && len(mandatory) == 0 && len(optional) == 0 {
			continue
		}
		fmt.Fprintf(&b, "| %s | %s | %s |\n", action, join(mandatory), join(optional))
	}
	for _, row := range []struct{ name, via string }{
		{"delegate (any action)", "delegate-row"},
		{"session start and close", "session-row"},
	} {
		var optional []string
		for _, id := range tc.ModuleOrder {
			m := tc.module(id)
			if m.OptionalVia != row.via || omitted(m) {
				continue
			}
			c, err := cell(id, "")
			if err != nil {
				return "", err
			}
			optional = append(optional, c)
		}
		if toggles != nil && len(optional) == 0 {
			continue
		}
		fmt.Fprintf(&b, "| %s | — | %s |\n", row.name, join(optional))
	}
	b.WriteString("\n")
	b.WriteString(actionTableRules)
	return b.String(), nil
}

func (tc *TaskContext) module(id string) Module {
	for _, m := range tc.Modules {
		if m.ID == id {
			return m
		}
	}
	return Module{}
}

// RenderToolTable genera la tabla de herramientas por acción que lleva el
// módulo task-tool-selection (D18).
func (tc *TaskContext) RenderToolTable() string {
	var b strings.Builder
	b.WriteString("| Action | Base tools | Optional integrations enabled |\n|---|---|---|\n")
	for _, a := range tc.ToolSelection.Actions {
		opt := "none"
		if len(a.Optional) > 0 {
			var labels []string
			for _, o := range a.Optional {
				labels = append(labels, o.Label)
			}
			opt = strings.Join(labels, "; ")
		}
		fmt.Fprintf(&b, "| %s | %s | %s |\n", a.ID, strings.Join(a.Base, ", "), opt)
	}
	return b.String()
}

// RenderToolsMarkdown genera docs/herramientas-por-accion.md.
func (tc *TaskContext) RenderToolsMarkdown() string {
	var b strings.Builder
	b.WriteString("# Herramientas por acción\n\n")
	b.WriteString("> Generado desde `internal/catalog/taskcontext.yaml` (`tool_selection`). No se edita a mano:\n")
	b.WriteString("> `go test ./internal/catalog -update` lo regenera y un test falla si se desvía.\n")
	b.WriteString("> Decisión de origen: `design.md` — D18 de `opsx-task-context`.\n\n")
	b.WriteString("## Qué corresponde a cada acción\n\n")
	b.WriteString("Las herramientas base del cliente no son opcionales: la tabla dice cuáles corresponden a cada acción, no las instala ni las apaga. Toda integración opcional está **desactivada por defecto** y se habilita por acción; estar instalada no la habilita.\n\n")
	b.WriteString("| Acción | Herramientas base | Integraciones opcionales habilitadas |\n|---|---|---|\n")
	for _, a := range tc.ToolSelection.Actions {
		opt := "ninguna"
		if len(a.Optional) > 0 {
			var ids []string
			for _, o := range a.Optional {
				ids = append(ids, "`"+o.Integration+"`")
			}
			opt = strings.Join(ids, ", ")
		}
		fmt.Fprintf(&b, "| %s | %s | %s |\n", a.ID, strings.Join(a.Base, ", "), opt)
	}
	b.WriteString("\n## Estado por defecto y nivel por cliente\n\n")
	b.WriteString("«No habilitada» tiene tres niveles con nombre: **(a) solo recomendada** (el módulo le pide al agente que no la use), **(b) fuera del contexto** (el agente no ve que existe) y **(c) restringida por el cliente** (aunque lo intente, la llamada no ocurre). La columna dice el nivel **alcanzable por acción** dentro de una sesión. `allowed-tools` de una skill de Claude no restringe y no cuenta como (c). Una integración *solo recomendada* sigue invocable: si el agente la usa, queda registrado como uso fuera de clase y el reporte nunca la presenta como bloqueada.\n\n")
	b.WriteString("| Integración | Por defecto | Acciones | Requiere harness | Claude | Codex |\n|---|---|---|---|---|---|\n")
	for _, in := range tc.ToolSelection.Integrations {
		req := "—"
		if in.RequiresHarness != "" {
			req = "`" + in.RequiresHarness + "`"
		}
		state := "desactivada"
		fmt.Fprintf(&b, "| `%s` | %s | %s | %s | (%s) %s | (%s) %s |\n",
			in.ID, state, strings.Join(in.Actions, ", "), req,
			in.Levels["claude"].Level, LevelName(in.Levels["claude"].Level),
			in.Levels["codex"].Level, LevelName(in.Levels["codex"].Level))
	}
	b.WriteString("\n### Mecanismo por integración\n\n")
	for _, in := range tc.ToolSelection.Integrations {
		fmt.Fprintf(&b, "- `%s`\n", in.ID)
		for _, c := range toolClients {
			fmt.Fprintf(&b, "  - %s: %s\n", c, in.Levels[c].Mechanism)
		}
	}
	return b.String()
}

// RenderLoadMarkdown genera docs/carga-de-modulos-por-accion.md.
func (tc *TaskContext) RenderLoadMarkdown() string {
	var b strings.Builder
	b.WriteString("# Carga de módulos por acción\n\n")
	b.WriteString("> Generado desde `internal/catalog/taskcontext.yaml` (`modules`). No se edita a mano:\n")
	b.WriteString("> `go test ./internal/catalog -update` lo regenera y un test falla si se desvía.\n")
	b.WriteString("> Decisión de origen: `design.md` — D19 de `opsx-task-context`. Es una propuesta a acordar con 03, 04 y 07.\n\n")
	b.WriteString("Columnas por acción: **O** obligatorio, **op** opcional, **—** no se carga. Un módulo de toggle inactivo no figura; si el toggle de un módulo obligatorio está inactivo, la acción se detiene.\n\n")
	b.WriteString("| Módulo | Toggle | ")
	b.WriteString(strings.Join(Actions, " | "))
	b.WriteString(" | Delegación | Especialista |\n|---|---|")
	b.WriteString(strings.Repeat("---|", len(Actions)))
	b.WriteString("---|---|\n")
	for _, id := range tc.ModuleOrder {
		m := tc.module(id)
		toggle := "—"
		if m.Toggle != "" {
			toggle = "`" + m.Toggle + "`"
		}
		var cells []string
		for _, a := range Actions {
			c := "—"
			switch m.Actions[a] {
			case LevelMandatory:
				c = "O"
			case LevelOptional:
				c = "op"
			}
			if cond := m.Conditions[a]; cond != "" && c != "—" {
				c += " (" + cond + ")"
			}
			cells = append(cells, c)
		}
		via := "—"
		switch m.OptionalVia {
		case "delegate-row":
			via = "op, fila `delegate`"
		case "session-row":
			via = "op, fila `session start and close`"
		}
		fmt.Fprintf(&b, "| `%s` | %s | %s | %s | %s |\n", m.ID, toggle, strings.Join(cells, " | "), via, m.Specialist.Note)
	}
	b.WriteString("\n## Forma de la ruta por cliente\n\n")
	b.WriteString("| Cliente | Forma de la ruta |\n|---|---|\n")
	b.WriteString("| Claude Code | `${CLAUDE_SKILL_DIR}/modules/<id>.md` |\n")
	b.WriteString("| Codex | `modules/<id>.md`, relativa al directorio de la skill |\n")
	b.WriteString("\n## Límites declarados\n\n")
	for _, m := range tc.Modules {
		for _, c := range m.OmittedIn {
			fmt.Fprintf(&b, "- `%s` omitido en %s hasta la 09: límite declarado, no ausencia.\n", m.ID, clientTitle(c))
		}
	}
	b.WriteString("- Los cuatro módulos de la 05 (`directed-code-reading`, `bounded-command-output`, `task-tool-selection`, `concise-agent-reporting`) no tienen toggle: figuran siempre.\n")
	b.WriteString("- `bounded-command-output` es obligatorio en `apply` y `verify`: sin él no hay captura ni referencia y el resultado no puede verificarse.\n")
	b.WriteString("- Modo legado de Claude (anterior a 2.1.110): no hay skill ni tabla, así que ningún módulo de la 05 se entrega; todo resultado de `apply` o `verify` queda *sin verificar*.\n")
	b.WriteString("- Una superficie sin subagentes deja `verify` detenida como límite declarado de esa superficie.\n")
	b.WriteString(tc.renderParityMarkdown())
	return b.String()
}

func clientTitle(c string) string {
	switch c {
	case "claude":
		return "Claude"
	case "codex":
		return "Codex"
	}
	return c
}
