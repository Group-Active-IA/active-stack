## Context

`active-stack` (instalador methodology-first, Go 1.26, binario único) materializa skills/harnesses
clonando repos-fuente externos en runtime. El catálogo embebido (`internal/catalog/harnesses.yaml`)
declara cada fuente como `source: { repo: <owner>/<name>, ref: latest, method: clone, path?: ... }`.

Tras el rename `jr-stack → active-stack` (change `rename-product-active-stack`, completado), 2 repos-fuente
todavía llevan el branding `jr` y viven bajo el usuario personal:

- `JuanCruzRobledo/jr-orchestrator` — harness `jr-orchestrator` (catálogo línea ~62).
- `JuanCruzRobledo/jr-skills` — monorepo de las 14 skills curadas propias/fork (catálogo líneas ~140–245),
  cada una con `path: skills/<nombre>`.

El operador los va a republicar bajo `Group-Active-IA`. **Política innegociable**: los repos
`JuanCruzRobledo/*` originales son SAGRADOS — nunca se modifican; solo se leen / se espejan. El strip
de refs `jr` ocurre EXCLUSIVAMENTE sobre la copia ya alojada en la org, jamás sobre el original.

Restricciones del proyecto que aplican:
- Regla dura: "el instalador instala harnesses, NO es un repo de tooling de migración" → ningún script
  de mirror/`gh` se commitea dentro de active-stack.
- Regla dura: "NUNCA editar `harnesses.yaml` sin pasar por `catalog.Load()` validando" → catálogo inválido
  rompe el release.
- Regla dura (operador): "NUNCA build/release salvo red de seguridad explícita".

## Goals / Non-Goals

**Goals:**
- Repuntar las 15 referencias de los 2 repos `jr` de `JuanCruzRobledo/*` → `Group-Active-IA/*`
  en `internal/catalog/harnesses.yaml`, preservando `ref`, `method` y `path` de cada source.
- Garantizar que cada repunte está **gateado** por verificación de existencia/clonabilidad del repo
  destino en la org (no repuntar a la nada).
- Mantener `go build ./...` y `go test ./...` verdes con el conteo baseline (objetivo: 23 paquetes ok / 0 FAIL),
  actualizando las aserciones de test que hardcodean el repo viejo.
- Documentar el flujo mirror-then-strip como runbook operativo (Migration Plan), no como código.

**Non-Goals:**
- NO modificar los repos `JuanCruzRobledo/*` originales (son espejados, no editados).
- NO commitear scripts de automatización de mirror/`gh`/strip dentro de active-stack.
- NO repuntar repos sin `jr` en el nombre (`kb-creator`, `roadmap-generator`, `agent-instruction`,
  `skill-registry`, `find-skill`, `skill-creator`, terceros): quedan en `JuanCruzRobledo`/upstream intactos.
- NO cambiar comportamiento observable del instalador (no hay delta specs).
- NO introducir lógica de verificación de existencia de repo en el código del producto (la verificación
  del gate es una acción del apply / del operador, no una feature runtime).

## Decisions

### D1 — Único archivo de producto que cambia: `harnesses.yaml`
El repunte es puramente de procedencia de la fuente. Se cambia SOLO el sub-campo `repo` de cada `source`
afectado; `ref: latest`, `method: clone` y `path: skills/<...>` quedan idénticos.
*Alternativa descartada*: añadir un campo de "owner por defecto" o indirección configurable — sobre-ingeniería
para un repunte one-shot; viola "no inflar el scope".

### D2 — El gate de existencia/clonabilidad es disciplina del apply, NO código runtime
Antes de repuntar CADA `source.repo`, el apply verifica que el repo destino en la org existe y clona
(`git ls-remote https://github.com/Group-Active-IA/<destino>` → exit 0). Si no clona, NO se repunta esa
referencia y se reporta como bloqueada. Esto es análogo a la trampa D2 del rename anterior (repuntar a un
repo inexistente rompe la instalación).
*Alternativa descartada*: añadir validación de URL viva en `catalog.Load()` — convertiría una validación
de build en una llamada de red, frágil y fuera de scope. El gate vive en el procedimiento, no en el binario.

### D3 — Mirror-then-strip como runbook operativo, fuera del repo
La secuencia crear-mirror → strip-refs-jr ocurre en GitHub sobre los repos de la org. Se documenta en el
Migration Plan. Rationale (regla dura): active-stack instala harnesses, no es tooling de migración; no
commiteamos scripts de `gh`/mirror.
*Alternativa descartada*: un `Makefile`/script `tools/mirror.sh` versionado — viola la regla dura.

### D4 — Tests que hardcodean el repo viejo: actualización obligatoria vs. opcional
- **`internal/catalog/c30_curation_test.go`** (`const wantRepo = "JuanCruzRobledo/jr-skills"`):
  **DEBE actualizarse**. Es un test de producción que valida que las 14 curadas apuntan al monorepo
  correcto. Si el YAML se repunta y el test no, `go test ./internal/catalog/...` FALLA → catálogo no
  valida verde → viola regla dura. No es opcional.
- **`internal/install/best_effort_test.go`** (fixture `Source{Repo:"JuanCruzRobledo/jr-orchestrator"}`):
  es un string de test que NO clona (se mockea `SetSkillInstallFn`). El test pasa con cualquier string.
  **Decisión: actualizarlo por consistencia** (que el fixture refleje la realidad post-repunte evita
  confusión futura), pero su no-actualización no rompe el build. Costo cero, beneficio de coherencia.
- **`e2e/starter/starter_network_test.go` + `starter_e2e_helpers.go`**: clones reales contra `jr-skills`,
  con build tag de red que NO corre en CI por defecto. **Decisión: actualizar referencias y comentarios
  por consistencia**, pero quedan fuera de la red de seguridad estándar (`go test ./...` no los ejecuta sin
  el tag). Su corrección real depende de que el mirror en la org exista (mismo gate D2).

### D5 — Sin delta specs
Un repunte de URL de fuente no crea ni modifica requerimientos a nivel spec. `Capabilities` del proposal
quedan vacías a propósito. El archivo `specs/` del change no se materializa.

## Risks / Trade-offs

- **[Repuntar a un repo destino inexistente/no clonable]** → install roto (`method: clone` a la nada).
  *Mitigación*: gate D2 por-referencia (`git ls-remote` exit 0) antes de cada repunte; referencia
  bloqueada no se toca.
- **[Strip de refs `jr` aplicado por error sobre el original `JuanCruzRobledo/*`]** → daño irreversible a
  un repo sagrado. *Mitigación*: el strip ocurre SOLO sobre la copia en la org, y el runbook lo ordena
  explícitamente DESPUÉS del push del mirror; nunca se opera sobre el remoto original.
- **[Catálogo editado pero test `c30_curation_test.go` no]** → `go test ./internal/catalog/...` rojo,
  release bloqueado. *Mitigación*: D4 lo marca como tarea obligatoria del apply, no opcional.
- **[Nombre destino de los repos en la org aún sin decidir]** → no se puede ejecutar el repunte real.
  *Mitigación*: bloqueado por Open Question 1; el apply no procede hasta tener los nombres confirmados.
- **[`ref: latest` apunta a HEAD del mirror, que podría diverger]** → comportamiento de clone variable.
  *Mitigación*: se mantiene `latest` salvo decisión contraria (Open Question 3).

## Migration Plan

**Runbook operativo (se ejecuta en GitHub / por el operador, FUERA de active-stack). Orden estricto, por cada uno de los 2 repos:**

1. **Crear** el repo nuevo vacío en `Group-Active-IA` (nombre destino = Open Question 1).
2. **Espejar** el original: `git clone --mirror https://github.com/JuanCruzRobledo/<jr-repo>` →
   `git push --mirror https://github.com/Group-Active-IA/<destino>`. El original `JuanCruzRobledo/*`
   queda intacto (solo se leyó).
3. **Strip de refs `jr`** SOLO sobre la copia ya alojada en `Group-Active-IA/<destino>` (jamás sobre el
   original): eliminar/renombrar referencias a `jr`/`jr-stack` en el contenido de esa copia, commit + push
   a la org.
4. **Verificar clonabilidad**: `git ls-remote https://github.com/Group-Active-IA/<destino>` exit 0.

**Repunte en active-stack (este change, fase apply — GATEADO por el paso 4):**

5. SOLO cuando el paso 4 verifica verde para un repo, repuntar en `harnesses.yaml` las referencias de ESE
   repo (`jr-orchestrator`: 1 línea; `jr-skills`: 14 líneas, preservando `path:`).
6. Actualizar las aserciones de test del apply (D4): `c30_curation_test.go` (obligatorio),
   `best_effort_test.go` y helpers e2e (consistencia).
7. **Validar catálogo**: `go test ./internal/catalog/...` verde (`catalog.Load()` valida el YAML).
8. **Red de seguridad**: `go build ./...` + `go test ./...` verdes, conteo baseline (23 ok / 0 FAIL).

**Rollback**: el change toca solo `harnesses.yaml` + tests; revertir es un `git checkout` de esos archivos.
Como los repos `JuanCruzRobledo/*` originales nunca se tocan, el catálogo viejo siempre vuelve a clonar
correctamente.

## Open Questions

1. **Nombre destino de los 2 repos en `Group-Active-IA`**: `jr-orchestrator` → `?`, `jr-skills` → `?`.
   (No inventado: el operador lo define. El apply no procede sin estos nombres.)
2. **Fixture `Source{Repo:"JuanCruzRobledo/jr-orchestrator"}` en `best_effort_test.go`**: el design
   recomienda actualizarlo por consistencia (D4), pero queda a confirmación del operador si prefiere
   dejarlo (es solo un string de test que no clona).
3. **`ref: latest`**: ¿se mantiene `latest` (HEAD del mirror) o se pinea a un tag/commit del mirror para
   reproducibilidad?
