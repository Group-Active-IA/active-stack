## Context

`active-stack` es un clon fresco de `jr-stack` (remoto `JuanCruzRobledo/jr-stack`) que se republicará como `Group-Active-IA/active-stack`. JR Stack es un instalador methodology-first en Go 1.26 (TUI Bubbletea, binario único cross-platform). El working tree está limpio y matchea `origin/main`.

El rename es **mecánico pero masivo y heterogéneo**: ~453 ocurrencias estrictas de `jr-stack` en 176 archivos, repartidas en cuatro frentes con muy distinta criticidad:

1. **Module path + imports** (`go.mod` + cada `.go`): puro find/replace, validado por el compilador.
2. **Directorio `cmd/jr-stack/` + build scripts**: rename de carpeta + ajuste de paths en `build.bat`/`build.sh`.
3. **Marca en prosa/docs/banner**: "JR Stack" → "Active Stack".
4. **Comandos + assets** (`harnesses.yaml`, `assets/commands/*`, `sdd-orchestrator.md`, golden files, marker de filemerge): el frente de **criterio**, donde un find/replace ciego rompe cosas.

Decisiones de nombre YA fijadas por el operador (no se re-discuten): module path `github.com/Group-Active-IA/active-stack`, binario `active-stack`, dir `cmd/active-stack/`, marca "Active Stack", CLI `active-stack ...`, slash Claude `active:starter-add`, slash opencode `active-starter-add`.

## Goals / Non-Goals

**Goals:**
- Renombrar **este módulo/producto** (`JuanCruzRobledo/jr-stack` → `Group-Active-IA/active-stack`) de forma completa y consistente en código, build, docs y comandos.
- Cero leftovers legítimos de `jr-stack`/`jrstack`/`jr_stack` al final (salvo lo preservado a propósito, que queda documentado).
- Conservar comportamiento idéntico: `go test ./...` verde con el **mismo conteo** que el baseline (un rename no cambia comportamiento).

**Non-Goals:**
- NO renombrar repos-fuente externos que el instalador clona (`JuanCruzRobledo/jr-orchestrator`, `JuanCruzRobledo/skill-registry`, starters). Son terceros respecto de este módulo.
- NO cambiar funcionalidad, flags, ni la estructura de paquetes.
- NO ejecutar el apply en esta tarea: esto es solo la propuesta. El rename lo corre otro operador.
- NO hacer `git build`/release automático (regla dura del proyecto: no build sin pedido explícito; aquí los `go build` son red de seguridad del apply, no releases).

## Decisions

### D1 — Orden de rename en capas, con `go build ./...` entre capas estructurales
Renombrar en este orden: (1) module path en `go.mod` → (2) todos los imports en `.go` → (3) rename de `cmd/jr-stack/` → `cmd/active-stack/` → (4) build scripts → (5) prosa/docs/banner → (6) comandos + assets + golden. Tras cada capa **estructural** (1+2, 3, 4) correr `go build ./...`. Rationale: aísla el frente que rompe la compilación; el compilador es el verificador más barato y exacto para el module path. Alternativa descartada: un único find/replace global sobre todo el repo → mezcla los frentes y hace imposible distinguir un fallo de compilación de una preservación mal hecha.

### D2 — Distinguir "este módulo" de "repos-fuente externos" (TRAMPA CRÍTICA)
El reemplazo de `JuanCruzRobledo/...` NO es global. Solo se reemplaza `JuanCruzRobledo/jr-stack` (este módulo). Se **preservan textualmente**:
- `internal/catalog/harnesses.yaml:62` → `source: { repo: JuanCruzRobledo/jr-orchestrator, ... }`
- `internal/install/best_effort_test.go` → `Source{Repo: "JuanCruzRobledo/jr-orchestrator", ...}`
- `CHANGES.md`, `docs/archive/CHANGES-2026-06-04-installer-y-starters.md` → menciones a `jr-orchestrator` / `skill-registry`.
- Cualquier otro `JuanCruzRobledo/<repo-distinto-de-jr-stack>` (starters incluidos).

Rationale: esas referencias le dicen al instalador **de dónde clonar** harnesses externos. Si se renombran al nuevo owner, el clone apunta a un repo inexistente y la instalación se rompe. Técnica obligatoria del apply: usar reemplazo **anclado al token `jr-stack`** (no a `JuanCruzRobledo/`), y para cualquier match de `JuanCruzRobledo/` que NO sea `jr-stack`, dejarlo intacto. Alternativa descartada: `sed JuanCruzRobledo→Group-Active-IA` global → rompería la instalación (es exactamente la trampa).

### D3 — Prefijo de marker de filemerge (`jr-stack:` → ¿`active-stack:`?) — decisión de governance ALTO
`internal/filemerge/section.go` y los `sdd-orchestrator.md` bundleados usan marcadores `<!-- jr-stack:sdd-delegation -->` … `<!-- /jr-stack:sdd-model-assignments -->` para inyectar/purgar bloques de forma idempotente en el `CLAUDE.md`/`AGENTS.md` del usuario. Tocar este prefijo es el cambio de **mayor riesgo** del rename (dominio `filemerge`, governance ALTO en el proyecto).

Opciones:
- **(A) Renombrar el prefijo a `active-stack:`** — coherente con la nueva marca; pero una reinstalación con el binario nuevo sobre un proyecto que ya tiene bloques `<!-- jr-stack:... -->` NO reconocería esos markers → riesgo de **duplicar bloques** (viola la regla dura "reinstalar no debe duplicar bloques") y deja markers viejos huérfanos al desinstalar.
- **(B) Conservar el prefijo `jr-stack:` como identificador técnico interno**, desacoplado de la marca visible — cero impacto en instalaciones existentes; pero deja un leftover `jr-stack` deliberado en código/golden.

**Recomendación para el apply**: como es un repo NUEVO sin instalaciones previas en producción bajo la marca Active, preferir **(A) renombrar a `active-stack:`** por coherencia total de marca, **siempre que** el apply: (i) actualice en bloque `section.go` + los 8 `sdd-orchestrator.md` + ambos golden de forma atómica, y (ii) deje el cambio de prefijo como su **propia capa con review** (governance ALTO: el agent propone y espera). Si el operador decide que habrá usuarios migrando desde instalaciones `jr-stack`, cae a **(B)** o se agrega un step de migración de markers (fuera de scope de este rename). Esta decisión se marca como Open Question para que el operador la confirme en el apply.

### D4 — Golden files: regenerar vs editar a mano
`compose_all_toggles_{claude,opencode}.golden` contienen (i) markers `jr-stack:...` (atados a D3) y (ii) prosa "JR Stack Methodology". Decisión: **regenerar** los golden con el comando/flag de update del paquete (p. ej. `go test ./internal/harness/config -run Compose -update` si existe), y si no existe flag de update, editarlos a mano y verificar con `go test`. Rationale: regenerar garantiza que el golden refleje exactamente el output compuesto real tras D3; editar a mano corre el riesgo de divergencia silenciosa. Tarea explícita en tasks.md.

### D5 — Renames de archivo/directorio de assets de comandos
No alcanza con cambiar contenido: hay que renombrar paths de asset.
- `assets/commands/claude/jr/` → `assets/commands/claude/active/` (dir).
- `assets/commands/opencode/jr-starter-add.md` → `assets/commands/opencode/active-starter-add.md` (archivo).
- Ajustar `assets/assets.go` / `assets/commands_test.go` / `internal/harness/command/*` que referencian esos paths e IDs (`jr:starter-add`, `jr-starter-add`).
Rationale: el embed (`//go:embed`) y los tests resuelven por path/ID; un rename a medias rompe el embed o los tests.

### D6 — Validación del catálogo obligatoria
`internal/catalog/harnesses.yaml` se edita (descripciones + IDs de comando). Regla dura: NUNCA editar `harnesses.yaml` sin que `catalog.Load()` lo valide. El apply debe correr el test que ejercita `catalog.Load()` (`internal/catalog/...`) tras editar el YAML, antes del build final.

## Risks / Trade-offs

- **[Romper la instalación por renombrar repos-fuente externos]** → Mitigación: D2. Reemplazo anclado al token `jr-stack`, nunca a `JuanCruzRobledo/`. Tarea de verificación que liste y confirme intactos `JuanCruzRobledo/jr-orchestrator` y `JuanCruzRobledo/skill-registry`.
- **[Duplicación de bloques en reinstalación por cambiar el prefijo de marker]** → Mitigación: D3. Tratar el cambio de prefijo como capa propia con review (governance ALTO); confirmar con el operador si hay usuarios migrando.
- **[Golden files divergentes o tests rojos por marca/markers]** → Mitigación: D4. Regenerar golden y correr `go test ./internal/harness/config/...`.
- **[Catálogo inválido rompe el release]** → Mitigación: D6. Validar vía `catalog.Load()` (test) tras editar el YAML.
- **[Leftovers olvidados]** → Mitigación: grep final case-insensitive como última tarea, con allowlist documentada de lo preservado a propósito.
- **[Conteo de tests distinto al baseline]** → señal de que el rename cambió comportamiento (no debería). Mitigación: capturar conteo baseline ANTES de tocar nada y comparar al final.

## Migration Plan

1. **Baseline**: `go build ./...` + `go test ./...` verdes; capturar el conteo de tests (p. ej. `ok ... N tests`).
2. Rename por capas (D1), con `go build ./...` tras cada capa estructural.
3. Editar comandos/assets/golden/marker (D3–D5) y validar catálogo (D6).
4. `go test ./...` verde con el mismo conteo que el baseline.
5. Grep de verificación → cero leftovers legítimos.
6. **Rollback**: el working tree estaba limpio en `origin/main`; ante cualquier fallo irrecuperable, `git restore` / `git checkout -- .` revierte todo. Como es una rama nueva, no hay riesgo sobre `main` publicado.

## Open Questions

- **D3 (prefijo de marker)**: ¿el operador confirma renombrar el prefijo a `active-stack:` (opción A) o conservar `jr-stack:` como identificador técnico interno (opción B)? Esto define si quedan leftovers deliberados de `jr-stack` en `section.go`/golden y debe resolverse al inicio del apply.
- ¿Hay flag `-update` para regenerar los golden de `internal/harness/config`, o se editan a mano? (afecta D4).
- ¿El nombre exacto del slash-command Claude es `active:starter-add` con dir `assets/commands/claude/active/`? (asumido por simetría con el viejo `jr/`).
