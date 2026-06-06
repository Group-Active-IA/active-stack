# CHANGES.md — JR Stack v2: Post-Starters

> Roadmap operativo del epic **"JR Stack v2 — post-starters"** (2026-06-04).
> El epic anterior (installer base + starters) está archivado en
> `docs/archive/CHANGES-2026-06-04-installer-y-starters.md` — leerlo para
> historial de decisiones de C-01..C-32 y los changes descriptivos intermedios.
>
> Este documento cubre los 6 changes del nuevo epic. Convención de IDs: kebab
> descriptivo (sin `C-NN` estricto — convención establecida desde `harness-scope-model`).
> Atomicidad: cada change implementable en una sesión, ≤ 12 tareas.
> **Fuente de verdad del estado real: el `openspec` CLI** (`openspec list`,
> `openspec status`). Este documento es el plan.

---

## Árbol de dependencias

```
openspec-init-cleanup (XS, repo de skills)
  — independiente, cualquier ola —

opencode-orchestrator-parity (PRIORIDAD #1)
  — independiente, cualquier ola —

uninstall-subcommand
  — independiente de los anteriores —
          │
          ▼
   tui-menu-hub
   (reestructura ScreenWelcome en hub multi-opción)
          │
          ├────────────────────┐
          ▼                    ▼
   tui-update-stack     tui-configure-models
   (opción "Update      (opción "Configure models":
    stack" en el menú)   model-routing claude + opencode)

claude-agent-switch-research
  — independiente, cualquier ola —
```

> Notas de dependencia:
> - `openspec-init-cleanup` NO toca este repo Go — vive en el `SKILL.md` del
>   skill `jr-orchestrator` (repo de skills separado). Se puede hacer en cualquier
>   ola sin coordinación con los demás.
> - `opencode-orchestrator-parity` es trabajo de auditoría + golden test sobre
>   assets existentes; no bloquea ni es bloqueado por nada de este epic.
> - `uninstall-subcommand` es wiring puro (el engine `internal/uninstall` ya
>   existe y está testeado desde C-12); no depende de paridad ni de cleanup.
> - `tui-menu-hub` depende del executor headless de uninstall (`uninstall-subcommand`)
>   porque la pantalla ScreenUninstall lo invoca.
> - `tui-update-stack` depende del menú hub (agrega una entrada al hub).
> - `tui-configure-models` también depende del menú hub (agrega la entrada
>   "Configure models"); es hermano paralelo de `tui-update-stack` (ambos cuelgan
>   del hub y no se bloquean entre sí).
> - `claude-agent-switch-research` es un change de investigación pura; puede
>   ejecutarse en cualquier ola, en paralelo con cualquier otro.

---

## Plan óptimo de paralelización

| Ola | Changes en paralelo | Bloquea a | Comentario |
|---|---|---|---|
| 0 | `openspec-init-cleanup`, `opencode-orchestrator-parity` | nada / nada | Independientes; `openspec-init-cleanup` es XS en repo de skills |
| 1 | `uninstall-subcommand`, `claude-agent-switch-research` | `tui-menu-hub` / — | `uninstall-subcommand` habilita el menú; research corre en paralelo libre |
| 2 | `tui-menu-hub` | `tui-update-stack`, `tui-configure-models` | Punto de convergencia TUI; necesita executor del uninstall (ola 1) |
| 3 | `tui-update-stack`, `tui-configure-models` | — | Cierre del epic TUI; ambos cuelgan del hub en paralelo |

**Camino crítico del epic**: `uninstall-subcommand → tui-menu-hub → tui-update-stack` (3 changes).
Es la cadena más larga; el resto corre en paralelo o es independiente.

**Máximo paralelismo útil**: 2 agentes (ola 0 y ola 1 tienen 2 changes independientes cada una).

---

## Plan con 2 agentes

| Paso | Agente A (Go / infra) | Agente B (skills / research) |
|------|------------------------|-------------------------------|
| Ola 0 | `opencode-orchestrator-parity` (auditoría assets + golden test) | `openspec-init-cleanup` (SKILL.md del jr-orchestrator) |
| Ola 1 | `uninstall-subcommand` (wiring CLI Go) | `claude-agent-switch-research` (investigación) |
| Ola 2 | `tui-menu-hub` (reestructura TUI Bubbletea) | — |
| Ola 3 | `tui-update-stack` (port internal/update + opción menú) | `tui-configure-models` (port model-routing picker claude+opencode) |

---

## Fichas de change

### `opencode-orchestrator-parity` — Auditar paridad del orquestador opencode vs claude

- **Estado**: COMPLETADO.
- **Scope**:
  - Auditar `internal/harness/config/assets/opencode/sdd-orchestrator.md`: identificar qué secciones del
    orquestador SDD (delegation, governance, model-routing, engram protocol,
    session-close, etc.) existen en `internal/harness/config/assets/claude/sdd-orchestrator.md` pero faltan
    o están degradadas en el asset de opencode.
  - Portar el contenido faltante al asset de opencode, conservando diferencias
    legítimas de formato entre agentes (opencode usa `task` tool nativo; claude usa Agent tool).
  - Implementar el golden test ausente: `TestCompose_AllToggles_opencode` (hoy solo
    existe `TestCompose_AllToggles` para claude) en
    `internal/harness/config/compose_test.go` con testdata en `internal/harness/config/testdata/`.
  - El golden test falla rojo si el asset de opencode diverge silenciosamente
    en el futuro (blind spot que este change cierra).
- **Decisiones / notas clave**:
  - El código de `ConfigDeliveryPrimaryAgent` y `mode:primary` en opencode.json está
    implementado y testeado; el gap era de CONTENIDO, no de mecanismo.
  - Única divergencia justificada: delegación via `task` tool (opencode nativo) vs Agent tool (claude).
    El `Template:` block y todo lo demás son idénticos entre variantes.
  - Golden generado en `internal/harness/config/testdata/compose_all_toggles_opencode.golden`.
- **Dependencias**: ninguna. Habilita: nada directo (mejora de calidad).
- **Governance**: BAJO (auditoría) / MEDIO (edita asset en `internal/harness/config/assets/opencode/`).
- **Leer antes**:
  - `internal/harness/config/assets/opencode/sdd-orchestrator.md` — asset opencode (corregido).
  - `internal/harness/config/assets/claude/sdd-orchestrator.md` — referencia canónica.
  - `internal/harness/config/compose_test.go` — golden test modelo para claude y opencode.
  - `internal/harness/config/testdata/compose_all_toggles_claude.golden` — golden existente de claude.
  - `CLAUDE.md` §3 regla "NUNCA hardcodear paths de agente" y §4 governance.

---

### `uninstall-subcommand` — Exponer `jr-stack uninstall` en la CLI

- **Estado**: IMPLEMENTADO (2026-06-06, rama `feat/uninstall-subcommand`). Wiring completo
  (flags + executor exportado + dispatch + `case "uninstall"`). **Bonus fix de engine**:
  `internal/uninstall` no manejaba `model.HarnessCommand` → `uninstall --mode lite/full`
  crasheaba con el catálogo real (`starter-add-command` es type:command). Se agregó
  `commandRemovalStep` + `CommandsDir`/`VariantKey` en la interfaz + `RelPathForVariant`
  exportado de `internal/harness/command`. TDD estricto, suite completa verde. Pendiente: archive OPSX.
- **Scope**:
  - Nuevo `cmd/jr-stack/headless/uninstall_flags.go`: tipo `ParsedUninstallFlags` +
    función `ParseUninstallFlags(args []string)`. Flags: `--mode`, `--agent`,
    `--custom` (lista de IDs), `--strategy targeted|restore`, `--restore-manifest`,
    `--dry-run`, `--yes`, `--home`. Uninstall es MACHINE-scope (sin `--project`).
  - Nuevo `cmd/jr-stack/headless/uninstall_executor.go`: función `RunHeadlessUninstall`
    que llama `uninstall.BuildPlan` (engine C-12, intacto) y lo ejecuta. Construida
    como executor COMPARTIDO — también podrá ser invocado desde la TUI (change
    `tui-menu-hub`).
  - Nuevo `cmd/jr-stack/uninstall_dispatch.go`: función `runUninstallDispatch` con
    lógica flat (espeja `runStarterAdd`). Adapter `uninstallRegistryAdapter` que
    satisface la interfaz que `uninstall.BuildPlan` espera.
  - Nuevo `case "uninstall":` en el `switch os.Args[1]` de `cmd/jr-stack/main.go`.
  - Tests: `uninstall_flags_test.go`, `uninstall_executor_test.go`,
    `uninstall_dispatch_test.go`. Usar fakes/stubs para el engine (nunca el entorno real).
  - **SIN tocar `internal/uninstall`** — el engine está completo y testeado (C-12).
  - ~200–250 LOC producción + tests proporcionados. TDD estricto activo.
- **Decisiones / notas clave**:
  - `RunHeadlessUninstall` debe ser una función exportada limpia (no solo usado por
    main.go) para que `tui-menu-hub` pueda invocarla en la pantalla ScreenUninstall.
  - El uninstall es MACHINE-scope por diseño (no hay `--project` aquí): deshace
    harnesses del home del usuario, no de un proyecto.
  - La operación es destructiva — el engine ya tiene snapshot+rollback mandatorio
    (C-12, governance ALTO); el wiring es aditivo pero hereda esa governance.
- **Dependencias**: ninguna (el engine `internal/uninstall` ya existe). Habilita:
  `tui-menu-hub` (pantalla ScreenUninstall usa `RunHeadlessUninstall`).
- **Governance**: **ALTO** (expone destrucción de config del usuario; la operación
  es destructiva). Proponer y esperar review antes del apply.
- **Leer antes**:
  - `internal/uninstall/` — engine completo (C-12, archivado `2026-05-27-c12-uninstall`).
  - `cmd/jr-stack/headless/executor.go` + `flags.go` — modelo para el executor/flags del install (espejarlo).
  - `cmd/jr-stack/starter_add.go` + `cmd/jr-stack/starter_dispatch.go` — modelo del dispatch flat para starters (espejarlo).
  - `cmd/jr-stack/main.go` — punto de entrada del switch de dispatch.
  - `CLAUDE.md` §3 regla "NUNCA pisar config del usuario sin backup".

---

### `tui-menu-hub` — TUI con menú descriptivo multi-opción

- **Estado**: PENDIENTE.
- **Scope**:
  - Reestructurar `internal/tui/` para convertir `ScreenWelcome` en un menú hub:
    opciones Install / Starters / Manage backups / Uninstall / Update stack / Quit.
  - Nueva pantalla `ScreenStarters`: lista los starters del catálogo → al confirmar,
    llama `runStarterAdd` (binario en PATH con `$ARGUMENTS`, o invoca el executor
    headless de starter directamente).
  - Nueva pantalla `ScreenBackups`: lista backups existentes (`internal/backup`
    snapshot index) → acciones restore / rename / delete.
  - Nueva pantalla `ScreenUninstall`: formulario de opciones (modo, agente, strategy)
    → llama `RunHeadlessUninstall` (del change `uninstall-subcommand`).
  - La pantalla Install existente (`ScreenInstall`) se preserva sin cambios internos;
    solo se agrega el routing desde el hub.
  - Usar Lipgloss inline mínimo (igual que hoy). **NO portar theme cosmético del
    legacy** (`styles/`, logo, frames, persona-marketing) — prohibido por `CLAUDE.md` §1.
  - Tests Bubbletea con `teatest` (skill `go-testing`).
- **Decisiones / notas clave**:
  - Los backends de cada pantalla YA existen (`internal/backup`, `internal/uninstall`,
    `catalog.Starters`); el trabajo es exclusivamente pantallas TUI + wiring.
  - La entrada "Update stack" aparece en el menú pero puede mostrar "coming soon"
    hasta que `tui-update-stack` esté terminado — el hub no bloquea.
  - El routing entre pantallas debe ser limpio (patrón Bubbletea model/update/view);
    no poner lógica de negocio en la TUI.
- **Dependencias**: `uninstall-subcommand` (la pantalla ScreenUninstall invoca
  `RunHeadlessUninstall`). Habilita: `tui-update-stack` (agrega la entrada Update
  al hub).
- **Governance**: BAJO (pantallas TUI) + MEDIO (wiring uninstall que toca config
  del usuario). Implementar con checkpoints; superficiar decisiones no obvias al
  operador.
- **Leer antes**:
  - `internal/tui/model.go`, `internal/tui/screen.go` — estructura actual de la TUI.
  - `internal/tui/gate.go` — lógica de preflight y selección actual.
  - `internal/backup/` — API de backups para la pantalla ScreenBackups.
  - `cmd/jr-stack/headless/uninstall_executor.go` — executor a invocar desde ScreenUninstall (creado en `uninstall-subcommand`).
  - `CLAUDE.md` §1 (prohibición del theme cosmético) y skill `go-testing` (patrones teatest).

---

### `claude-agent-switch-research` — Investigar agent-switching en Claude (RESEARCH)

- **Estado**: PENDIENTE. (**Change de investigación — no produce código de producción.**)
- **Scope**:
  - Research profundo: ¿existe en Claude Code un mecanismo nativo de "Tab entre
    agentes" (equivalente a `mode:primary` de opencode)? ¿Hay roadmap público o
    feature flag conocido?
  - Explorar workarounds vía slash-commands o campo `"agent"` en `settings.json`
    que aproximen el UX de switching.
  - Verificar empíricamente qué produce instalar el orquestador como subagent
    definition (`"agent"` en settings.json) — ¿es tab-switch real o solo define
    el agente default del proyecto?
  - Output obligatorio: documento de veredicto (`docs/research/claude-agent-switch.md`
    o similar) con: conclusión (posible / no posible / workaround limitado), evidencia
    consultada, reframe de qué sí es alcanzable hoy, y recomendación de si requiere
    un change de implementación posterior o se cierra como "no entregable documentado".
  - Puede terminar en "no entregable, documentado" — ese también es un resultado
    válido y valioso.
- **Decisiones / notas clave**:
  - La exploración previa ya confirmó que Claude Code NO tiene UI nativa de tab-switch;
    este change profundiza y documenta formalmente ese hallazgo + busca alternativas.
  - Lo más cercano hoy: instalar el orquestador como subagent definition + `"agent"`
    default en settings.json (el proyecto arranca en modo orquestador, NO es tab-switch).
  - El veredicto puede desencadenar un change de implementación posterior (si se
    encuentra un workaround viable) o archivarse como investigación definitiva.
- **Dependencias**: ninguna (research independiente). No bloquea ni habilita ningún
  otro change del epic.
- **Governance**: BAJO (research + config mínima para pruebas empíricas).
- **Leer antes**:
  - `assets/claude/sdd-orchestrator.md` — lo que hoy se inyecta como orquestador en Claude.
  - `internal/agents/claude/adapter.go` — cómo el installer escribe en `settings.json` de Claude.
  - Documentación oficial de Claude Code (slash-commands, subagent definitions, settings.json schema).
  - `ARCHITECTURE.md` §2.1 — descripción del harness `sdd-orchestrator`.

---

### `tui-update-stack` — Update rápido del stack desde la TUI

- **Estado**: PENDIENTE.
- **Scope**:
  - Portar `internal/update` + `internal/update/upgrade` del repo legacy
    (`E:\ESCRITORIO\programar\2026\Framework\jr-stack-legacy`), **stripped** de
    código `gga` y del theme cosmético del legacy.
  - Agregar la opción "Update stack" al menú hub (`tui-menu-hub`): al seleccionarla,
    la pantalla ScreenUpdate corre el flujo de actualización.
  - Flujo de actualización:
    - Check de versiones de `jr-stack` + `engram` contra GitHub Releases API.
    - Mostrar qué binarios están desactualizados (diff de versión).
    - Upgrade de binarios: self-replace de `jr-stack` (rename-trick para Windows
      exe-in-use, ya resuelto en el legacy `update/upgrade/executor.go`) + upgrade
      de `engram` vía el método de instalación correcto por OS.
    - Re-sync de configs: re-inyectar los config-harness (`sdd-orchestrator`,
      `permissions`) en los agentes detectados (vía el pipeline existente en modo
      "solo config").
  - Limpieza obligatoria al portar: eliminar todos los leftovers
    `gentle-ai`/`gentle-stack`/`Gentleman.Dots`/GGA/theme del código portado.
  - Tests: unit tests para la lógica de check de versiones y upgrade; mock de la
    GitHub Releases API. TDD estricto activo.
- **Decisiones / notas clave**:
  - **UX: replicar la riqueza del legacy (decisión del operador, 2026-06-06).** El
    hub legacy NO tenía un solo "Update stack" — tenía **tres** entradas separadas
    (`Upgrade tools`, `Sync configs`, `Upgrade + Sync`) más un **badge dinámico ★**
    en el ítem del menú cuando hay updates disponibles (o sufijo `(up to date)` tras
    el chequeo). Portar esa UX: las 3 acciones + el badge ★ vivo en el hub. Hoy el
    hub (de `tui-menu-hub`) muestra solo "Update stack (coming soon)" como placeholder
    inline (`hubNotice`); este change lo reemplaza por la entrada/entradas reales.
  - Referencia legacy: `internal/tui/screens/welcome.go` (`WelcomeOptions` con el
    badge ★ vía `update.HasUpdates`), `screens/upgrade.go`, `screens/sync_screen.go`,
    `screens/upgrade_sync.go`.
  - El rename-trick para Windows (escribir `.new`, mover target a `.old`, renombrar
    `.new` → target, best-effort remove `.old`) ya está resuelto en `binary-self-install`
    (`internal/install/self_install.go`) — reutilizarlo o factorizarlo a `internal/system`.
  - El re-sync de configs debe pasar por el pipeline existente (backup + markers
    idempotentes), no reescribir lógica de inyección.
  - Checkpoint: proponer el diseño del flujo de re-sync al operador antes de
    implementarlo (involucra `internal/install` en modo parcial).
- **Dependencias**: `tui-menu-hub` (el hub debe existir para agregar la entrada
  Update). No habilita nada (es el cierre del epic TUI).
- **Governance**: MEDIO (self-replace de binario + re-inyección en config de
  agentes). Implementar con checkpoints; superficiar el diseño del re-sync al
  operador antes del apply.
- **Leer antes**:
  - Repo legacy `internal/update/` y `internal/update/upgrade/executor.go` — lógica a portar.
  - `internal/install/self_install.go` — rename-trick para Windows ya implementado (reusar).
  - `internal/agents/` — adapters por agente para detectar qué agentes están instalados.
  - `internal/harness/config/` — instalador de config-harnesses (para el re-sync).
  - `tui-menu-hub` change (debe estar terminado — depende de él).
  - `CLAUDE.md` §3 regla "SIEMPRE limpiar leftovers gentle-ai / gentle-stack / Gentleman.Dots".

---

### `tui-configure-models` — Configurar model-routing desde la TUI (claude + opencode)

- **Estado**: PENDIENTE. (Origen: el hub legacy tenía "Configure models" y el
  operador decidió, 2026-06-06, portarlo a v2 como change nuevo — no entró en
  `tui-menu-hub`, que ya está cerrado con 6 opciones.)
- **Scope**:
  - Agregar la opción **"Configure models"** al hub (`tui-menu-hub` ya existe).
  - Pantalla de entrada `ScreenModelConfig`: menú `Configure Claude models` /
    `Configure OpenCode models` / `Back` (espeja `screens/model_config.go` del legacy).
  - Pantalla `ScreenClaudeModelPicker`: presets **balanced / performance / economy /
    custom**; en modo custom, lista por fase OPSX (orchestrator, explore, propose,
    apply, archive, default) cicleando alias **opus → sonnet → haiku** con Enter
    (espeja `screens/claude_model_picker.go` del legacy).
  - Pantalla equivalente para **OpenCode** (su propio set de modelos/aliases — NO
    asumir paridad 1:1 con Claude; opencode tiene su propio catálogo de modelos).
  - Al confirmar, **re-inyectar la tabla de model-routing en el bloque del
    `sdd-orchestrator`** de los agentes detectados, vía el pipeline existente
    (backup + markers idempotentes) en modo "solo config" — NO reescribir lógica
    de inyección. El toggle `model-routing` del `sdd-orchestrator` es el destino real
    del routing configurado ("si se chequea, aparece el routing correcto").
  - Limpieza obligatoria al portar del legacy: eliminar leftovers
    `gentle-ai`/`gentle-stack`/`Gentleman.Dots`/theme cosmético.
  - Tests Bubbletea con `teatest` (skill `go-testing`): presets, ciclado de alias
    en custom, navegación, y el wiring de re-inyección con fakes (nunca el entorno real).
- **Decisiones / notas clave**:
  - Los presets de Claude del legacy (balanced/performance/economy/custom) y el
    ciclado opus→sonnet→haiku son la referencia de UX — reusarlos. Verificar que
    los presets sigan vigentes con el model-routing actual del `sdd-orchestrator` v2.
  - **OpenCode NO es un clon de Claude**: necesita su propio picker con su catálogo
    de modelos. Resolver `(TBD)` qué modelos/aliases expone opencode antes del apply.
  - El destino del routing es el harness `config` `sdd-orchestrator` (toggle
    `model-routing`), no un archivo nuevo — pasa por el mismo install pipeline parcial
    que usará `tui-update-stack` para el re-sync.
- **Dependencias**: `tui-menu-hub` (el hub debe existir para agregar la entrada).
  Hermano paralelo de `tui-update-stack` (ambos cuelgan del hub, no se bloquean).
  No habilita nada (hoja del epic TUI).
- **Governance**: MEDIO (re-inyección en config de agentes — toca `CLAUDE.md`/
  `AGENTS.md` del usuario vía markers + backup). Implementar con checkpoints;
  superficiar el diseño del re-sync al operador antes del apply.
- **Leer antes**:
  - Legacy `internal/tui/screens/model_config.go` — pantalla de entrada (3 opciones).
  - Legacy `internal/tui/screens/claude_model_picker.go` — presets + custom por fase + ciclado.
  - Legacy `internal/tui/screens/model_picker.go` y `model_config_test.go` — picker genérico + tests modelo.
  - `internal/harness/config/` — instalador del `sdd-orchestrator` (toggle `model-routing`), destino del re-sync.
  - `internal/agents/` — adapters por agente (detectar qué agentes reciben el routing).
  - `tui-menu-hub` change (debe estar terminado — depende de él).
  - `CLAUDE.md` §1 (prohibido theme cosmético) y §3 ("SIEMPRE limpiar leftovers gentle-ai…").

---

### `openspec-init-cleanup` — Limpiar basura del `openspec init` en `jr-orchestrator`

- **Estado**: COMPLETADO (2026-06-06). Cleanup `rm -rf .claude/skills/openspec-*`
  agregado al Step 1 del `SKILL.md` del `jr-orchestrator` (justo después de
  `openspec init`), via glob future-proof. Aplicado en el repo fuente
  (`SKILLS/jr-orchestrator`, v2.0) **y** en la copia instalada
  (`~/.claude/skills/jr-orchestrator`, v2.1) que es la que realmente corre.
  ⚠️ Descubierto: el repo fuente (v2.0) está DESACTUALIZADO respecto a la copia
  instalada (v2.1, con checkpoint protocol + skill-registry) — la v2.1 nunca se
  pusheó al repo. Reconciliar source ↔ instalado es follow-up aparte.
- **⚠️ ARTEFACTO: SKILL `jr-orchestrator` — este change NO modifica el repo Go.**
  El fix vive en el `SKILL.md` del skill `jr-orchestrator` (repo
  `JuanCruzRobledo/jr-orchestrator`, archivo `SKILL.md` o equivalente de la skill).
  No tocar `internal/` ni ningún archivo Go de este repo.
- **Scope**:
  - `openspec init` dropea incondicionalmente los dirs
    `<proyecto>/.claude/skills/openspec-*` (skills redundantes — ya están globales
    en el stack). El cleanup no ocurre hoy porque nadie lo dispara.
  - Fix: agregar un step de cleanup en el `SKILL.md` de `jr-orchestrator`, **en el
    Step 1** (justo después de `openspec init`), que borra todos los dirs
    `openspec-*` conocidos bajo `.claude/skills/`:
    ```bash
    rm -rf .claude/skills/openspec-explore \
           .claude/skills/openspec-init \
           .claude/skills/openspec-onboard \
           .claude/skills/openspec-design \
           .claude/skills/openspec-spec \
           .claude/skills/openspec-tasks \
           .claude/skills/openspec-verify \
           .claude/skills/openspec-apply-change \
           .claude/skills/openspec-archive-change \
           .claude/skills/openspec-propose
    ```
  - El comando debe ser future-proof: si `openspec init` agrega dirs `openspec-*`
    nuevos en el futuro, el glob `rm -rf .claude/skills/openspec-*` los cubre
    automáticamente (alternativa más robusta que listar nombres explícitos).
  - **NO tocar** `.claude/commands/opsx/` — es el delivery de slash-commands y se
    mantiene intacto.
  - **NO modificar** el skill `openspec-init` — el fix es en el orquestador, no en
    el init mismo.
  - Tamaño: XS (~5 líneas en el SKILL.md). No requiere tests unitarios.
- **Decisiones / notas clave**:
  - Los dirs `openspec-*` son auto-generados y re-creables; no son archivos
    user-authored. Borrarlos es seguro.
  - La raíz del problema: `openspec init` siempre los genera; el orquestador debe
    limpiar inmediatamente después (es más limpio que parchear `openspec init`).
- **Dependencias**: ninguna (XS, independiente). No bloquea ni habilita nada del epic.
- **Governance**: BAJO (archivos auto-generados, re-creables, no user-authored).
  Autonomía completa — change XS que no requiere review previo.
- **Leer antes**:
  - `SKILL.md` del repo `JuanCruzRobledo/jr-orchestrator` — Step 1 donde se inserta
    el cleanup (fuente a editar).
  - `ARCHITECTURE.md` §4.2 — descripción del flujo de `jr-orchestrator` (lazy-loading).
  - `CLAUDE.md` §1 — confirmación de que `.claude/commands/opsx/` no se toca.

---

## Tabla resumen

| Change | Ola | Governance | Depende de | Habilita |
|---|---|---|---|---|
| `opencode-orchestrator-parity` | 0 (paralelo) | BAJO/MEDIO | — | — |
| `openspec-init-cleanup` | 0 (paralelo) | BAJO | — | — | ✅ COMPLETADO |
| `uninstall-subcommand` | 1 | **ALTO** | — | `tui-menu-hub` | ✅ IMPLEMENTADO |
| `claude-agent-switch-research` | 1+ (libre) | BAJO | — | — |
| `tui-menu-hub` | 2 | BAJO/MEDIO | `uninstall-subcommand` | `tui-update-stack`, `tui-configure-models` | ✅ IMPLEMENTADO (pend. archive) |
| `tui-update-stack` | 3 | MEDIO | `tui-menu-hub` | — |
| `tui-configure-models` | 3 | MEDIO | `tui-menu-hub` | — |

**Camino crítico**: `uninstall-subcommand → tui-menu-hub → {tui-update-stack ‖ tui-configure-models}` (3 changes; los dos últimos en paralelo).

**Próximo paso**: archivar `tui-menu-hub` (implementado, suite verde) y luego encarar
en paralelo `tui-update-stack` y `tui-configure-models` (ambos cuelgan del hub).
