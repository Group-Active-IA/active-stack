## Why

El operador definió un principio de marca firme: **"todo se tiene que llamar active"**. Ya se ejecutaron dos frentes de debranding (producto `jr-stack` → `active-stack`; repunte de `source.repo` de los dos repos `jr` → `Group-Active-IA/active-orchestrator` y `Group-Active-IA/active-skills`). Queda el **último identificador con marca `jr`** vivo en código activo: el harness `jr-orchestrator`. Su `id`/`name` todavía dice `jr`, y ese identificador no es cosmético: define el **namespace del slash-command** de la skill (`/jr-orchestrator:init`, `:kb`, `:rules`, …). Cerrar este rename completa el debranding y deja un nombre coherente de punta a punta.

## What Changes

- Renombrar el identificador del harness `jr-orchestrator` → `active-orchestrator` en `internal/catalog/harnesses.yaml` (`id` + `name`). El `source.repo` ya apunta a `Group-Active-IA/active-orchestrator` (rename previo) — acá solo se alinea el identificador local.
- Actualizar las **~42 ocurrencias del string `jr-orchestrator`** en código activo (tests `.go`, `.md` de docs, goldens, `.sh` de e2e, `sdd-orchestrator.md` bundleados). Excluye `openspec/`, `docs/archive/` y `CHANGES.md` (histórico/planificación).
- Renombrar el identificador Go CamelCase del test `TestForMode_JROrchestratorIsFullOnly` → `TestForMode_ActiveOrchestratorIsFullOnly`, por consistencia.
- Actualizar las aserciones de `internal/catalog/catalog_test.go` (líneas ~309-321) que verifican que el harness con ID `jr-orchestrator` es full-only y NO está en lite → deben asertar el nuevo ID `active-orchestrator`, o el catálogo (regla dura `catalog.Load()`) rompe.
- Regenerar/editar los dos goldens (`compose_all_toggles_{claude,opencode}.golden`) y los `sdd-orchestrator.md` bundleados (la instrucción "invoke `jr-orchestrator`" → "invoke `active-orchestrator`").
- **Coordinación cross-repo (documentada, NO implementada desde acá)**: el `SKILL.md` del repo `Group-Active-IA/active-orchestrator` (org) y el `~/.claude/CLAUDE.md` global del usuario también referencian `jr-orchestrator`. El gate de consistencia es obligatorio (ver Impact).
- **NO** hay cambio de comportamiento: es un rename de identificador. Comportamiento idéntico antes/después.

## Capabilities

### New Capabilities
<!-- Ninguna. Rename de identificador sin nueva capacidad. -->

### Modified Capabilities
<!-- Ninguna. No cambia ningún requirement a nivel spec: el harness se materializa, instala y orquesta exactamente igual; solo cambia su nombre. Sin delta specs. -->

## Impact

**Código activo afectado** (~42 refs string + 1 identificador Go):
- `internal/catalog/harnesses.yaml` — `id` + `name` del harness (regla dura: pasar por `catalog.Load()`).
- 6 tests `.go`: `internal/catalog/catalog_test.go`, `internal/install/best_effort_test.go`, `internal/verify/harness_checks_test.go`, `internal/verify/hook_test.go`, `internal/backup/restore_test.go`, `e2e/starter/starter_hermetic_test.go`.
- 4 `.md`: `README.md`, `ARCHITECTURE.md`, `internal/harness/config/assets/{claude,opencode}/sdd-orchestrator.md`.
- 2 goldens: `internal/harness/config/testdata/compose_all_toggles_{claude,opencode}.golden`.
- 1 `.sh`: `e2e/e2e_test.sh`.
- Identificador Go: `TestForMode_JROrchestratorIsFullOnly` → `…ActiveOrchestratorIsFullOnly`.

**Fuera de scope (NO tocar)** — documentado en design:
- `jr-stack` (refs legacy ya allowlisteadas: repo fuente original / "repo viejo" de porting).
- `Jrh`, `JrbApNNgaTdGDITg` (strings random de fixtures, false positives).
- `JRStackInvocation` (nombres de funciones Go del rename de producto) — decisión cosmética en design.
- `openspec/`, `docs/archive/`, `CHANGES.md` (histórico/planificación).

**Dimensión cross-repo CRÍTICA** (la trampa de este change): el ID es el namespace del slash-command, declarado en el `SKILL.md` del repo de la **org**, no acá. **Gate de consistencia obligatorio**: el `id`/`name` del harness en active-stack DEBE matchear el `name:` del `SKILL.md` de la org. Si active-stack dice `active-orchestrator` pero el SKILL.md sigue diciendo `jr-orchestrator`, la skill instalada no matchea lo que el flujo de fundación invoca → roto. Frentes: (1) este repo, (2) `SKILL.md` de la org [runbook], (3) `~/.claude/CLAUDE.md` global [solo FLAG, config del usuario].

**Red de seguridad**: `go build ./...` + `go test ./...` = 23 paquetes OK / 0 FAIL, idéntico baseline antes y después.

**Governance**: `catalog` es BAJO (autonomía con tests verdes), pero el gate cross-repo (SKILL.md de la org debe alinear) es la salvaguarda crítica de este change.
