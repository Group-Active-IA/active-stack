## Why

El operador va a republicar bajo la org `Group-Active-IA` los 2 repos-fuente que aún
llevan el branding `jr` (`JuanCruzRobledo/jr-orchestrator` y `JuanCruzRobledo/jr-skills`).
El instalador `active-stack` clona esos repos en runtime (`method: clone`) para
materializar skills/harnesses; mientras el catálogo siga apuntando a `JuanCruzRobledo/*`
el producto queda atado al usuario personal y al naming viejo. Este change repunta esas
referencias a la org — completando la línea de continuidad del rename `jr-stack → active-stack`
para las fuentes externas, que el rename de producto no tocó.

## What Changes

- Repuntar en `internal/catalog/harnesses.yaml` el `source.repo` del harness
  `jr-orchestrator` (1 línea) de `JuanCruzRobledo/jr-orchestrator` → `Group-Active-IA/<destino>`.
- Repuntar en `internal/catalog/harnesses.yaml` los `source.repo` de las 14 skills curadas
  propias/fork (líneas ~140–245), de `JuanCruzRobledo/jr-skills` → `Group-Active-IA/<destino>`,
  **preservando intacto el campo `path: skills/<...>` de cada una**.
- Actualizar las aserciones de test que hardcodean el repo viejo y romperían `go test` al
  repuntar el catálogo:
  - `internal/catalog/c30_curation_test.go` — `const wantRepo = "JuanCruzRobledo/jr-skills"`
    (test de producción que valida que las 14 curadas apuntan al monorepo; **debe** actualizarse
    o el catálogo no valida verde).
  - `internal/install/best_effort_test.go` — fixture `Source{Repo:"JuanCruzRobledo/jr-orchestrator"}`
    (string de test, no clona; decisión de actualizarlo por consistencia se resuelve en design).
  - `e2e/starter/starter_network_test.go` + `e2e/starter/starter_e2e_helpers.go` — comentarios/clones
    reales contra el upstream `jr-skills` (tests de red, no corren en CI por defecto; alcance en design).
- **GATE de seguridad** (no es código del producto, es disciplina del apply): cada `source.repo`
  se repunta SOLO una vez verificado que el repo destino en la org existe y es clonable
  (p. ej. `git ls-remote`). Repuntar a un repo inexistente rompe la instalación (clone a la nada).
- El crear-mirror-en-la-org-y-strip-refs-jr **NO entra como código** en active-stack: se documenta
  como runbook en el Migration Plan del design y se ejecuta en GitHub sobre los repos de la org.

NO entran (quedan en `JuanCruzRobledo`, intactos — no llevan `jr` en el nombre):
`kb-creator`, `roadmap-generator`, `agent-instruction`, `skill-registry`, `find-skill`,
`skill-creator` y todos los terceros (`obra/superpowers`, `wshobson/agents`, etc.).

## Capabilities

### New Capabilities
<!-- Ninguna. Repuntar la URL de un source-repo NO introduce ni modifica ningún
     comportamiento observable del instalador: clona el mismo contenido desde otra ubicación. -->
- (none)

### Modified Capabilities
<!-- Ninguna. No cambia ningún REQUERIMIENTO a nivel spec; es un cambio de procedencia
     de la fuente, transparente para el resto del sistema. No hay delta specs. -->
- (none)

## Impact

- **Código del producto que cambia**: SOLO `internal/catalog/harnesses.yaml` (15 referencias:
  1 de `jr-orchestrator` + 14 de `jr-skills`).
- **Tests que cambian**: `internal/catalog/c30_curation_test.go` (obligatorio), y según design
  `internal/install/best_effort_test.go` y los helpers e2e de `e2e/starter/`.
- **Sin cambios** en: `internal/model`, `internal/install` (lógica), `internal/harness`, TUI,
  ni ningún otro paquete. El comportamiento del instalador es idéntico.
- **Dependencia externa (operativa, fuera del repo)**: existencia y clonabilidad de los 2 repos
  nuevos en `Group-Active-IA`. El apply está **gateado** por esa precondición.
- **Riesgo principal**: repuntar a un repo destino que aún no existe/clona → install roto. Mitigado
  por el gate de verificación por-referencia (análogo a la trampa D2 del rename anterior).
- **Governance**: paquete `catalog` = BAJO; pero el gate (no repuntar a repo inexistente) y la
  política "los `JuanCruzRobledo/*` originales son SAGRADOS, jamás se modifican, solo se espejan"
  son las salvaguardas críticas a respetar.
