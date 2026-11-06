## Why

Hoy el orquestador OPSX que `active-stack` inyecta en el `CLAUDE.md` del usuario pasa directo de `propose` a `apply`: en changes con interfaz de usuario, el agente escribe código de pantallas sin que nadie haya visto ni aprobado cómo se van a ver. Claude ya tiene la capacidad de generar mockups visuales con la herramienta **Artifact** (tipo "Design", vía `quickstart` con `intent: "design"`), así que podemos sumar un paso de diseño explícito, con aprobación humana, antes de que se escriba una línea de UI.

El usuario tiene que poder **elegirlo**: es opt-in, con un control visible en la instalación. Los toggles del `sdd-orchestrator` no se exponen en la TUI ni en la GUI (salen literales del catálogo); los harnesses sí. Por eso la capacidad se entrega como un **harness propio**, no como un toggle.

Además, el diagrama OPSX del bloque claude del orquestador omite `verify`, y se corrige en la misma pasada.

> Nota de nombre: el change se llama `add-design-mockups-toggle` porque la primera versión era un toggle. Se mantiene el nombre para no romper la trazabilidad; el entregable es el harness `design-mockups` (ver design.md D1).

## What Changes

- **Nuevo harness `design-mockups`** en `internal/catalog/harnesses.yaml`: `type: config`, `agents: [claude]`, `depends_on: [sdd-orchestrator]`, `install_modes: [full]` (decidido por el usuario; Custom matchea todos; ver design.md D6), `description` y `long_description` bilingües es/en. Aparece en el picker Custom (y en la review) como cualquier otro harness, sin cambios de UI.
- **Nuevo fragmento** `internal/harness/config/assets/design-mockups.md` con las instrucciones del paso de diseño. Se inyecta en el `CLAUDE.md` del agente como **sección propia** con marker `<!-- active-stack:design-mockups -->`, separada de la sección `sdd-orchestrator`, con backup previo y markers idempotentes (API existente de `filemerge`/`backup`).
- El fragmento le define al orquestador:
  - una etapa nueva entre `propose` y `apply` (`explore → propose → [design: mockups + aprobación] → apply → verify → archive`);
  - cuándo preguntar: solo si los artefactos del change (proposal/specs/design/tasks) mencionan UI. Respuestas posibles: sí / no / ya los tengo. La respuesta queda registrada en `design.md` para no volver a preguntar;
  - dónde viven los mockups: `openspec/changes/<name>/mockups/`, referenciados desde `design.md` y `tasks.md`, y tomados como criterio de aceptación visual en `apply` y `verify`;
  - un **gate de aprobación humana obligatorio**: si los mockups no están aprobados explícitamente, `apply` no arranca. Es una excepción declarada a la regla general "NEVER block on phase gates".
- **Instalador config genérico por sección** (`internal/harness/config`): `Install` despacha por harness. `sdd-orchestrator` sigue igual (Compose + sección `sdd-orchestrator`); un harness con fragmento registrado (`design-mockups`) se inyecta en su propia sección `active-stack:<id>`, **solo** en adapters que el harness soporta (`agents`), y nunca por primary-agent.
- Registrar `design-mockups` en `ownedSectionIDs` para que `PurgeStaleSections` (install, uninstall y primary-agent) no la borre como stale.
- `internal/verify`: el check de config usa el ID del harness como marker (hoy hardcodea `sdd-orchestrator`) y solo chequea adapters que el harness soporta.
- Desinstalación: sin código nuevo. `markerRemovalStep` ya quita la sección por `h.ID`; se cubre con tests.
- Corregir el diagrama OPSX del base claude del orquestador (`assets/claude/sdd-orchestrator.md`) para que incluya `verify`, y regenerar **solo** el golden claude del orquestador.
- Tests (TDD estricto): instala solo en claude; idempotencia; desinstalación limpia; no la purga el orquestador (ni al instalar ni al desinstalar); opencode/codex intactos; catálogo carga y los conteos de modo se actualizan.
- Docs: `CLAUDE.md` + `AGENTS.md` (espejo), `README.md`, `ARCHITECTURE.md`, `CHANGES.md`. En `CHANGES.md` se registran como pendientes, sin implementarlos, el harness hermano `stitch-design` (OpenCode + Stitch) y la pista opcional `UI: sí/no` en `roadmap-generator`/`active-orchestrator` (repos externos).

Sin cambios en `internal/filemerge`, `internal/backup` ni `internal/pipeline`: solo se usa su API pública.

## Capabilities

### New Capabilities
- `design-mockups-harness`: harness `config` opt-in, exclusivo de Claude, que inyecta en el `CLAUDE.md` del agente una sección propia con la etapa de diseño de mockups (detección de UI, registro de la decisión, ubicación de los mockups en el change y gate de aprobación previo a `apply`), instalable/desinstalable de forma idempotente e independiente de la sección del orquestador. Incluye la corrección del diagrama OPSX (con `verify`) del bloque claude del orquestador.

### Modified Capabilities
(ninguna. `catalog-localization` no cambia sus requirements: la nueva entrada solo tiene que cumplir la validación bilingüe existente.)

## Impact

- Código: `internal/harness/config/{install.go,inject.go}` (despacho por harness, inyección por sección, `ownedSectionIDs`), nuevo `internal/harness/config/assets/design-mockups.md`, `internal/harness/config/assets/claude/sdd-orchestrator.md` (diagrama), `internal/verify/harness_checks.go`, tests en `internal/harness/config`, `internal/verify`, `internal/uninstall`, `internal/catalog`; `internal/harness/config/testdata/compose_all_toggles_claude.golden` (regenerado solo por la línea `verify`).
- Catálogo: nueva entrada `design-mockups`; conteos hardcodeados de `catalog_test.go` (Full/Custom 15→16; Lite sigue en 6 si se confirma `[full]`).
- Usuarios: solo quien elige el harness (o Full, si se confirma `[full]`) con agente claude recibe la sección nueva. `compose.go` y la salida del orquestador para codex/opencode no cambian.
- Governance: `harness/config` y `verify` = MEDIO (checkpoints). La escritura en `CLAUDE.md` del usuario reutiliza backup + markers (sin tocar paquetes ALTO).
- Dependencia de runtime (no de build): la herramienta `Artifact` de Claude Code con el tipo "Design". Si no está disponible, el fragmento define un fallback (ver design.md).
