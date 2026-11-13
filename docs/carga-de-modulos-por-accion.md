# Carga de módulos por acción

> Generado desde `internal/catalog/taskcontext.yaml` (`modules`). No se edita a mano:
> `go test ./internal/catalog -update` lo regenera y un test falla si se desvía.
> Decisión de origen: `design.md` — D19 de `opsx-task-context`. Es una propuesta a acordar con 03, 04 y 07.

Columnas por acción: **O** obligatorio, **op** opcional, **—** no se carga. Un módulo de toggle inactivo no figura; si el toggle de un módulo obligatorio está inactivo, la acción se detiene.

| Módulo | Toggle | explore | propose | apply | verify | archive | Delegación | Especialista |
|---|---|---|---|---|---|---|---|---|
| `artifacts` | — | — | O | O | — | O | — | sí, en el encargo |
| `delegation` | `delegation` | — | — | — | O | — | op, fila `delegate` | sí, cuando se delega |
| `model-assignments` | `model-routing` | — | — | — | O | — | op, fila `delegate` | no |
| `directed-code-reading` | — | op | op | op | op | — | — | sí, si lee código |
| `bounded-command-output` | — | op (if you run commands) | — | O | O | — | — | sí, si corre comandos |
| `tdd` | `tdd` | — | — | O (when Strict TDD is on and a test runner exists) | — | — | — | sí, en apply |
| `task-tool-selection` | — | op | op | op | op | op | — | no: el coordinador le pasa la lista |
| `concise-agent-reporting` | — | op | op | op | op | op | — | sí |
| `engram-opsx` | `engram` | — | — | — | — | — | op, fila `session start and close` | no |

## Forma de la ruta por cliente

| Cliente | Forma de la ruta |
|---|---|
| Claude Code | `${CLAUDE_SKILL_DIR}/modules/<id>.md` |
| Codex | `modules/<id>.md`, relativa al directorio de la skill |

## Límites declarados

- `model-assignments` omitido en Codex hasta la 09: límite declarado, no ausencia.
- Los cuatro módulos de la 05 (`directed-code-reading`, `bounded-command-output`, `task-tool-selection`, `concise-agent-reporting`) no tienen toggle: figuran siempre.
- `bounded-command-output` es obligatorio en `apply` y `verify`: sin él no hay captura ni referencia y el resultado no puede verificarse.
- Modo legado de Claude (anterior a 2.1.110): no hay skill ni tabla, así que ningún módulo de la 05 se entrega; todo resultado de `apply` o `verify` queda *sin verificar*.
- Una superficie sin subagentes deja `verify` detenida como límite declarado de esa superficie.

## Paridad por cliente y superficie

Superficies declaradas por la 04 (`codex-native-opsx/design.md:59-63`): CLI Windows / Linux (`cli`); Extensión IDE Windows / Linux (`ide`); Escritorio / cloud (`desktop-cloud`).

Un módulo que un cliente no puede recibir figura como **límite**, no como entrega:

| Cliente | Ámbito | Módulos | Motivo |
|---|---|---|---|
| Claude | `no-subagents` | `delegation` | superficie sin subagentes: verify queda detenida como límite de la superficie |
| Codex | `no-subagents` | `delegation` | superficie sin subagentes: verify queda detenida como límite de la superficie |
| Codex | `all` | `model-assignments` | omitido en Codex hasta la 09: límite declarado, no ausencia |
| Claude | `legacy-mode` | `directed-code-reading`, `bounded-command-output`, `task-tool-selection`, `concise-agent-reporting` | modo legado (Claude anterior a 2.1.110): no hay skill ni tabla; todo resultado de apply o verify queda sin verificar |

Pendiente de otra ficha:

- Codex: tarea 2.5 de la 04 — verificar en Codex la paridad por superficie (verify y contrato de evidencia por acción); la tabla de superficies ya está declarada.
