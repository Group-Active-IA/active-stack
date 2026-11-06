## Why

Este repositorio (clon fresco de `jr-stack`) va a republicarse bajo la organización GitHub `Group-Active-IA` como `active-stack`. El producto pasa a llamarse **Active Stack**: cambia el dueño del repo, el module path de Go, el binario, el directorio del entrypoint, la marca en prosa y los comandos expuestos al usuario. Hacerlo ahora —sobre un working tree limpio que matchea `origin/main`— evita arrastrar la marca vieja a la primera release de la organización.

## What Changes

- **BREAKING (module path)**: `github.com/JuanCruzRobledo/jr-stack` → `github.com/Group-Active-IA/active-stack` en `go.mod` y en cada `import` de cada `.go`.
- **BREAKING (binario)**: el ejecutable `jr-stack` → `active-stack`.
- **BREAKING (entrypoint)**: directorio `cmd/jr-stack/` → `cmd/active-stack/` (incluye subpaquete `cmd/jr-stack/headless/`).
- **Scripts de build**: `build.bat` y `build.sh` apuntan al nuevo path del binario y entrypoint.
- **Marca en prosa**: "JR Stack" → "Active Stack" en `README.md`, `CLAUDE.md`, `AGENTS.md`, `ARCHITECTURE.md`, `CHANGES.md` y `assets/.../banner.svg`.
- **BREAKING (comandos expuestos)**:
  - CLI: `jr-stack install` / `jr-stack starter add` / `jr-stack uninstall` → `active-stack ...`.
  - Slash-command Claude: `jr:starter-add` → `active:starter-add` (renombra el archivo/dir de asset).
  - Slash-command opencode: `jr-starter-add` → `active-starter-add` (renombra el archivo de asset).
- **Catálogo y assets bundleados**: `internal/catalog/harnesses.yaml` (descripciones, IDs de comando) y los `internal/harness/config/assets/<agente>/sdd-orchestrator.md` que mencionan el producto.
- **Golden files**: `internal/harness/config/testdata/compose_all_toggles_{claude,opencode}.golden` deben regenerarse/editarse para reflejar la marca y/o el prefijo de marker resultante.
- **Marker de inyección idempotente** (`jr-stack:` → decisión a tomar en design): prefijo usado por `internal/filemerge` para inyectar bloques sin pisar. Tocarlo tiene impacto en idempotencia de reinstalación; se trata como decisión explícita, NO como rename mecánico ciego.
- **NO se tocan**: las referencias a **repos-fuente externos** que el instalador clona (`JuanCruzRobledo/jr-orchestrator`, `JuanCruzRobledo/skill-registry`, starters). Esas NO son este módulo y renombrarlas rompe la instalación.

## Capabilities

### New Capabilities
<!-- Ninguna. Este change es un rename/refactor mecánico: no introduce comportamiento nuevo ni capabilities nuevas. -->

### Modified Capabilities
<!-- Ninguna. Un rename de producto no cambia REQUERIMIENTOS a nivel spec: el comportamiento observable (instalar harnesses, inyectar el orquestador, starters, uninstall) es idéntico. Por eso este change no produce delta specs; la red de seguridad es la suite de tests existente, que debe quedar verde con el MISMO conteo que el baseline. -->

## Impact

- **Magnitud medida** (grep case-insensitive `jr-stack|jrstack|jr_stack`): ~804 ocurrencias en ~212 archivos; la grep estricta sobre `jr-stack` da 453 en 176 archivos. La mayoría es el module path repetido en cada `import`.
- **Código afectado**: prácticamente todo `cmd/` e `internal/` (imports del module path), `internal/catalog`, `internal/harness/config` (assets + golden), `internal/filemerge` (marker), `assets/commands/{claude,opencode}`.
- **Build/CI**: cualquier pipeline o doc que invoque `go build`/el binario por nombre. `go.mod` cambia → consumidores externos (si hubiera) deben re-importar el nuevo path.
- **Usuarios existentes con instalaciones previas**: si se cambia el prefijo de marker, una reinstalación con el nuevo binario podría no reconocer los bloques viejos (riesgo de duplicación). Ver design.md.
- **Preservación deliberada**: referencias a `JuanCruzRobledo/<otros-repos>` (orchestrator, skill-registry, starters) se quedan EXACTAMENTE como están.
