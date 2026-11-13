# Herramientas por acción

> Generado desde `internal/catalog/taskcontext.yaml` (`tool_selection`). No se edita a mano:
> `go test ./internal/catalog -update` lo regenera y un test falla si se desvía.
> Decisión de origen: `design.md` — D18 de `opsx-task-context`.

## Qué corresponde a cada acción

Las herramientas base del cliente no son opcionales: la tabla dice cuáles corresponden a cada acción, no las instala ni las apaga. Toda integración opcional está **desactivada por defecto** y se habilita por acción; estar instalada no la habilita.

| Acción | Herramientas base | Integraciones opcionales habilitadas |
|---|---|---|
| explore | directed reading, search | `context7` |
| propose | directed reading, OpenSpec CLI | `context7` |
| apply | editing, tests | `context7`, `domain-skills`, `playwright` |
| verify | tests, directed reading | `playwright` |
| archive | OpenSpec CLI | ninguna |

## Estado por defecto y nivel por cliente

«No habilitada» tiene tres niveles con nombre: **(a) solo recomendada** (el módulo le pide al agente que no la use), **(b) fuera del contexto** (el agente no ve que existe) y **(c) restringida por el cliente** (aunque lo intente, la llamada no ocurre). La columna dice el nivel **alcanzable por acción** dentro de una sesión. `allowed-tools` de una skill de Claude no restringe y no cuenta como (c). Una integración *solo recomendada* sigue invocable: si el agente la usa, queda registrado como uso fuera de clase y el reporte nunca la presenta como bloqueada.

| Integración | Por defecto | Acciones | Requiere harness | Claude | Codex |
|---|---|---|---|---|---|
| `context7` | desactivada | explore, propose, apply | — | (a) solo recomendada | (a) solo recomendada |
| `playwright` | desactivada | apply, verify | — | (a) solo recomendada | (a) solo recomendada |
| `domain-skills` | desactivada | apply | — | (b) fuera del contexto | (a) solo recomendada |
| `stitch` | desactivada | propose, apply, verify | `design-mockups` | (a) solo recomendada | (a) solo recomendada |
| `emil-design-eng` | desactivada | propose, apply, verify | `design-mockups` | (b) fuera del contexto | (a) solo recomendada |
| `design-taste-frontend-v1` | desactivada | propose, apply, verify | `design-mockups` | (b) fuera del contexto | (a) solo recomendada |

### Mecanismo por integración

- `context7`
  - claude: CLAUDE.md o cuerpo de la skill (as context, not enforced); level c solo si la acción corre en un subagente con mcpServers o disallowedTools propios
  - codex: AGENTS.md o cuerpo de la skill; el archivo de un rol no fija mcp_servers, así que el control por acción de un MCP es solo recomendado
- `playwright`
  - claude: CLAUDE.md o cuerpo de la skill; level c solo en un subagente con mcpServers o disallowedTools propios
  - codex: AGENTS.md o cuerpo de la skill; el archivo de un rol no fija mcp_servers
- `domain-skills`
  - claude: skill con disable-model-invocation: true (oculta hasta invocarla)
  - codex: policy.allow_implicit_invocation: false deja nombre y descripción listados (medido X9, X11): no es fuera del contexto completo; skills.config en un rol da level c
- `stitch`
  - claude: habilitada solo con el harness design-mockups activo; CLAUDE.md o cuerpo de la skill
  - codex: habilitada solo con el harness design-mockups activo; AGENTS.md o cuerpo de la skill
- `emil-design-eng`
  - claude: skill con disable-model-invocation: true
  - codex: policy.allow_implicit_invocation: false deja nombre y descripción listados
- `design-taste-frontend-v1`
  - claude: skill con disable-model-invocation: true
  - codex: policy.allow_implicit_invocation: false deja nombre y descripción listados
