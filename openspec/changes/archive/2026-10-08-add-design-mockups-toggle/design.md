# Design

## Context

Motivación: ver `proposal.md` (Why). Requirements: ver `specs/design-mockups-harness/spec.md`.

### Estado actual relevante (verificado en el código)

- **Instalación config**: `install/plan.go:528-541` despacha `HarnessConfig` → `permissionsStep` si `h.ID == "permissions"`, si no → `configStep` → `config.Install` (`install/steps.go:176-198`).
- **`config.Install` está acoplado al orquestador**: siempre llama `Compose(h.Toggles, adapter.VariantKey())` (`harness/config/install.go:46,63`) e `Inject` escribe siempre en la sección `"sdd-orchestrator"` (`harness/config/inject.go:100`). Un segundo harness config pasado por `Install` hoy inyectaría **el bloque del orquestador**: hace falta despacho por harness.
- **Adapters sin filtrar por harness**: `BuildPlan` pasa a cada step **todos** los adapters elegidos (`install/plan.go:113,494-501`); `filterByAgents` (`install/plan.go:241-255`) solo decide si el harness entra. Con claude+opencode, un harness `agents: [claude]` recibiría también el adapter opencode y `Install` iría por la rama primary-agent (`install.go:37-54`), registrando un agente `design-mockups` en `opencode.json`. Lo mismo en `verify` (`verify/hook.go:29-30`) y en uninstall (`uninstall/plan.go:64-66`). Hay que filtrar por `h.SupportsAgent(adapter.Agent())`.
- **Purge de secciones stale**: `PurgeStaleSections` (`inject.go:51-58`) borra toda sección `active-stack:` cuyo ID no esté en `ownedSectionIDs` (`inject.go:33-37`). Se llama en install (`inject.go:97`), en primary-agent (`primary_agent.go:109`) y en uninstall de **cualquier** config harness (`uninstall/steps.go:73-89,158`). Sin registrar `design-mockups`, reinstalar o desinstalar el orquestador **borraría** la sección del harness nuevo.
- **Desinstalación genérica**: `markerRemovalStep` quita la sección `h.ID` (`uninstall/steps.go:153`) y después purga stale. Para `design-mockups` funciona sin código nuevo. En el adapter opencode (si llega sin filtrar) `primaryAgentRemovalFn(settings, "design-mockups")` es no-op (`steps.go:119-121`).
- **Verify**: `checkConfig` busca siempre el marker `sdd-orchestrator` (`verify/harness_checks.go:19,105`). Para `design-mockups` daría un falso positivo (si el orquestador está) o un falso negativo. Como el ID de sección del orquestador == su ID de harness, usar `h.ID` no cambia nada para `sdd-orchestrator`.
- **Catálogo**: no hay validación específica de `config` (`catalog/catalog.go:88-131`); un harness global-picker exige `long_description` bilingüe (`catalog.go:123-127`). `depends_on` lo resuelve el planner como auto-dependencia (`planner/resolver.go:64-68`, `planner/types.go:50`). Conteos hardcodeados: `catalog_test.go:341,368-370` (Lite 6, Full 15, Custom 15).
- **UI**: el picker Custom es data-driven (`install/custom_picker.go:14`: `filterByAgents(cat.ForMode(Custom), agents)`); la GUI Windows usa `description` localizada (`headless/windows.go:530-545`). No hay goldens de UI con la lista de harnesses. El harness aparece en TUI/GUI/review sin cambios de UI.
- **Diagrama OPSX** del base claude (`assets/claude/sdd-orchestrator.md` L48-61): `explore → propose → apply → archive`, sin `verify`. Existe la skill `openspec-verify`, pero no el slash command `/opsx:verify`.
- El base claude dice "NEVER block on phase gates": el gate de este harness es una excepción deliberada y se declara como tal.
- `openspec/` está en `.gitignore` en ESTE repo (dogfooding local). En proyectos destino, la convención es versionarlo.

### Investigación verificada: Artifact tipo "Design"

Verificado en la sesión de propose, ejecutando la herramienta y leyendo su descripción:

1. `Artifact` con `action: "quickstart"` e `intent: "design"` es read-only. Devuelve el tipo core **"Design"** (*"Design canvas for websites, landing pages, screens, UI mockups, wireframes…"*) con su `type_url`, que es por cuenta/catálogo. **No se hardcodea**: se obtiene siempre del quickstart.
2. Paso siguiente: publicar con `type_url`, un `title`, sin archivos y con `auto_open: "after_first_write"`. El resultado trae las instrucciones del tipo y cómo llenarlo. El quickstart también reporta los design systems de la cuenta.
3. Los Artifacts son **privados por defecto**; la URL no sirve al equipo hasta que se comparte.
4. `Artifact` `action: "read"` con `path`/`paths` + `out_dir` descarga archivos publicados. En un artifact tipado se obtienen los archivos de datos propios, no la página del tipo.
5. Sub-agentes: `general-purpose` tiene `Artifact`; `Explore` y `Plan` no.
6. **DesignSync** no existe en este entorno (solo herramientas MCP `stitch__*`, de Google Stitch). El fragmento no la menciona.

Sin verificar:

- `(TBD)` Formato de los archivos de datos de un artifact "Design" y si se exportan como HTML/PNG autocontenido. El fragmento delega en las instrucciones que devuelve la creación del artifact.
- `(TBD)` Disponibilidad de `Artifact` en todos los modos de autenticación de Claude Code. Por eso hay fallback.
- `(TBD)` Si `openspec archive` mueve `mockups/` junto con el change (se asume que sí; se verifica en apply).

## Goals / Non-Goals

**Goals:**
- Un harness `design-mockups` opt-in, visible y elegible en la instalación Custom, exclusivo de claude.
- Inyección como sección propia, idempotente, con backup previo, independiente de la sección del orquestador (instalar/desinstalar uno no rompe al otro).
- Instrucciones del paso de diseño basadas en lo verificado de Artifact; lo no verificado se delega a las instrucciones que la herramienta devuelve en runtime.
- No modificar `filemerge`, `backup` ni `pipeline` (governance ALTO): solo usar su API.
- Salida de `Compose` para codex/opencode byte a byte igual.

**Non-Goals:**
- Exponer los toggles del `sdd-orchestrator` en TUI/GUI.
- Corregir para `sdd-orchestrator` el leak latente de adapters sin filtrar (por ejemplo, gemini recibiendo el bloque si se elige junto a claude). Se registra como deuda; este change solo filtra en la ruta nueva por sección y en verify.
- Soporte de Stitch/OpenCode (harness hermano `stitch-design`, change futuro).
- Editar `roadmap-generator`/`active-orchestrator` (repos externos).
- Agregar `/opsx:verify` como slash command, o una fila `verify` en la tabla de modelos.

## Decisions

### D1. Harness propio `design-mockups` (reemplaza la decisión anterior "toggle")

- **Elegido**: entrada nueva en `harnesses.yaml`:
  ```yaml
  - id: design-mockups
    name: Design Mockups (Claude)
    description: { es: "...", en: "..." }
    long_description: { es: "...", en: "..." }
    type: config
    install_modes: [full]   # decidido: ver D6
    agents: [claude]
    depends_on: [sdd-orchestrator]
  ```
- **Por qué**: el usuario lo elige explícitamente. Los harnesses aparecen en el picker Custom, en la review y en la GUI sin cambios de UI; los toggles no.
- **Decisión anterior descartada**: toggle `design-mockups` dentro de `sdd-orchestrator`, aportando texto solo con `variantKey == "claude"` en `Compose`. Se descarta porque los toggles salen literales del catálogo y no se exponen en TUI/GUI: el paso quedaba activo para **todos** los usuarios de claude, sin opt-in. Con esto desaparecen: la entrada en `knownToggles`, la condición `variantKey` en `Compose`, los tests de toggle on/off y el golden de claude con el fragmento.
- **`depends_on: [sdd-orchestrator]`**: el fragmento extiende el flujo OPSX del orquestador. El planner lo auto-incluye y lo ordena antes, así la sección del orquestador queda primero en `CLAUDE.md` en una instalación limpia.
- **Alternativa**: un campo `agents` por toggle + exponer toggles en UI. Se descarta: cambia `model.Harness`, parser, validación y dos UIs para un solo caso.

### D2. Instalación por sección dentro de `internal/harness/config`

- **Elegido**: un registro de fragmentos por harness en el paquete config (por ejemplo, `sectionFragments = map[string]string{"design-mockups": "design-mockups.md"}`). `config.Install` despacha:
  - harness con fragmento registrado → ruta **por sección**: para cada adapter con `h.SupportsAgent(adapter.Agent())` y `InstructionsPath` no vacío, inyecta `loadFragment(...)` en la sección `h.ID` con backup previo y `PurgeStaleSections`. La rama primary-agent **no** aplica: un harness por sección nunca se registra como agente en settings JSON.
  - cualquier otro (`sdd-orchestrator`) → ruta actual, sin cambios de comportamiento.
- `Inject` se generaliza a un `InjectSection(path, sectionID, content, snapshotDir)` y `Inject` queda como wrapper con `"sdd-orchestrator"` (cero cambio para los callers actuales).
- `install/plan.go` **no** se toca: `configStep` ya llama `config.Install`, y `collectWritePaths` ya snapshotea el `InstructionsPath` de cada adapter para `HarnessConfig` (`install/plan.go:307-310`).
- **Por qué no un step propio en `install/plan.go`**: duplicaría backup/rollback/paths y sumaría otro caso especial por ID como `permissions`. El despacho dentro de config mantiene a `install` agnóstico.

### D3. Registrar `design-mockups` en `ownedSectionIDs`

- Obligatorio: si no, cualquier reinstalación o desinstalación del orquestador (o de otro config harness) borra la sección (`inject.go:51-58`, `uninstall/steps.go:158`). Se suma una línea al mapa, con comentario. Idealmente las claves de `sectionFragments` se derivan al registro (`ownedSectionIDs` incluye todas las secciones registradas), para que agregar un harness por sección no requiera dos ediciones `(se decide en REFACTOR)`.
- Consecuencia: si el usuario desinstala solo el orquestador, la sección `design-mockups` sobrevive (es de otro harness). Se desinstala con su propio harness. Es el comportamiento deseado.

### D4. Verify por ID de harness y por agente soportado

- `checkConfig` usa `"<!-- active-stack:" + h.ID + " -->"` (para `sdd-orchestrator` es idéntico al actual).
- `ChecksForHarness` omite adapters que el harness no soporta. Para harnesses existentes solo puede **quitar** checks (nunca agregar fallos).

### D5. Fragmento: contenido funcional (sin cambios respecto de la versión anterior)

- **Archivo**: `internal/harness/config/assets/design-mockups.md`, en inglés como los demás assets, autocontenido (se lee como sección propia, después del bloque del orquestador).
- **Registro de decisión y aprobación** en `openspec/changes/<name>/design.md`, sección `## UI Mockups`:
  - `Decision: yes | no | provided` (+ fecha)
  - `Mockups: <ruta a mockups/ y/o URL del Artifact>`
  - `Approval: pending | approved (<fecha>, by user)`
  Motivo: `design.md` ya lo exige el schema y lo leen apply y verify; no se edita metadata del CLI (`.openspec.yaml`) a mano.
- **Mockups en** `openspec/changes/<name>/mockups/`: mínimo `mockups/README.md` (URL del Artifact, pantallas, estado de aprobación) y, si se puede, una copia exportada vía `Artifact action:"read"` con `paths`/`out_dir`. `design.md` y `tasks.md` los referencian; son criterio de aceptación visual en apply y verify. En proyectos destino `openspec/` se versiona y los mockups viajan con el change; se advierte que las URLs de Artifact son privadas hasta que se comparten.
- **Fallback sin `Artifact`**: mockups HTML estáticos autocontenidos en `mockups/`; el gate sigue igual.
- **Detección de UI**: señales en los artefactos (página/pantalla/vista/componente/formulario/layout/UI/frontend, `.tsx/.jsx/.vue/.svelte/.html/.css`, frameworks de UI). Si hay duda, se pregunta. Sin señales → no se pregunta y se registra `Decision: no` ("no UI signal"). Respuesta "ya los tengo" (`provided`) → se copian/referencian en `mockups/` y el gate aplica igual.
- **Delegación y gate**: el orquestador delega la generación a un sub-agente `general-purpose` (tiene `Artifact`), que crea el artifact y escribe `mockups/README.md` pero no pide aprobación. El orquestador presenta, pide aprobación explícita y la registra; hasta entonces rechaza arrancar `apply` (excepción declarada a "NEVER block on phase gates"). Cambios → se itera sobre el mismo artifact (republicando por su `url`) y la aprobación vuelve a `pending`.
- **Fila de modelo**: el fragmento dice *"If the Model Assignments table is present, use `design → sonnet`"*. No se toca la tabla del base (evita fase fantasma sin el harness). (decidido: `sonnet`).

### D6. `install_modes` — decidido: `[full]`

- **Decidido por el usuario**: `[full]` (NO `[lite, full]`). Custom matchea todos, así que el harness siempre aparece en el picker Custom.
- **Recomendación**: `[full]`. Lite es "sustrato mínimo" y este paso depende de una capacidad runtime exclusiva de Claude (Artifact) que no todos tienen; en Full tiene sentido junto a la fundación guiada. Contra: en Full queda **activo por defecto** (no es opt-in puro). Si se quiere opt-in estricto, la alternativa sería que solo entre en Custom, pero hoy el modelo no tiene un modo "solo Custom" (`install_modes` vacío o un valor nuevo implicaría cambiar `model`/catálogo): se registra como alternativa, no se implementa.
- Impacto en tests: `[full]` → Full 15→16, Custom 15→16, Lite sigue en 6. `[lite, full]` → Lite 6→7 también.

### D7. Diagrama OPSX del base claude del orquestador

Insertar entre apply y archive: `openspec-verify  (validate implementation vs specs/design/tasks)`, referenciando la **skill** (no un slash command inexistente). El paso de design **no** va al diagrama del base: vive en el fragmento del harness, que muestra su diagrama extendido. Así el base es correcto sin el harness. Esto **sí** cambia el golden `compose_all_toggles_claude.golden` (solo esa línea); opencode/codex no cambian.

### Borrador del fragmento (se refina en apply)

```markdown
## Design Stage — UI Mockups (Claude)

Extends the OPSX workflow above. Flow with this stage:
explore → propose → **design** → apply → verify → archive.
Runs AFTER propose (specs define the screens) and BEFORE apply. Never during apply.

1. Detect UI: if the change's proposal/specs/design/tasks mention pages, screens,
   components, forms or other UI → ask the user "Does this change need mockups?"
   (yes / no / I already have them). No UI signal → do not ask; record `Decision: no`.
   If `design.md` already has a `## UI Mockups` decision → honour it, never re-ask.
2. Record in `openspec/changes/<name>/design.md` under `## UI Mockups`:
   Decision / Mockups / Approval.
3. yes → delegate to a general-purpose sub-agent: call the Artifact tool with
   action "quickstart", intent "design"; create the artifact from the returned
   `type_url` (title + auto_open "after_first_write") and follow the instructions
   the creation result returns. Write `mockups/README.md` with the artifact URL
   and screen list; save an exported copy into `mockups/` when the tool allows it.
   Artifact unavailable → write self-contained static HTML mockups into `mockups/`.
   provided → copy/reference the user's mockups into `mockups/`.
4. Approval gate (exception to "never block on phase gates"): do NOT start apply
   until the user explicitly approves and `Approval: approved` is recorded.
5. Reference `mockups/` from design.md and tasks.md; mockups are visual acceptance
   criteria in apply and verify.
6. If design skills such as `emil-design-eng` (Emil Kowalski) or a taste skill are
   installed (soft dependency; the stage works without them), load them for the
   design step.
Model: if the Model Assignments table is present, use `design → sonnet`.
Artifact URLs are private until shared.
```

## Risks / Trade-offs

- [Escritura en `CLAUDE.md` del usuario (impacto ALTO)] → Solo API existente: backup previo (`SnapshotterCreate`) + `InjectMarkdownSection` + `WriteFileAtomic`. Checkpoint con el diff real de un `CLAUDE.md` de prueba antes de cerrar.
- [Purge borrando la sección] → Registro en `ownedSectionIDs` + tests de install y uninstall del orquestador con la sección presente.
- [Adapters sin filtrar → opencode recibe un primary-agent `design-mockups`] → Filtro por `SupportsAgent` en la ruta por sección; test con claude+opencode que afirma que `opencode.json` y `AGENTS.md` quedan intactos.
- [Verify falso positivo/negativo] → Marker por `h.ID` + filtro por agente; tests en `internal/verify`.
- [Huérfano si se instala sin orquestador] → `depends_on: [sdd-orchestrator]` (auto-dependencia del planner). Si el usuario desinstala solo el orquestador, la sección queda; el fragmento es autocontenido.
- [En Full queda activo por defecto] → El paso solo se activa con señales de UI y el "no" se registra una vez. Decidido `[full]` (D6).
- [Gate obligatorio vs "NEVER block on phase gates"] → Excepción explícita y acotada (solo yes/provided).
- [Formato de export del artifact Design no verificado] → `mockups/README.md` con la URL es el mínimo garantizado.
- [`Artifact` no disponible] → Fallback HTML estático.
- [Leak latente de adapters en `sdd-orchestrator`] → Fuera de scope; se registra en `CHANGES.md` como deuda.

## Migration Plan

- Reinstalar con el harness elegido inyecta la sección nueva con backup previo. Sin migración de datos.
- Quien instaló el orquestador antes de este change no ve cambios salvo la línea `verify` del diagrama (al reinstalar).
- Rollback: desinstalar el harness `design-mockups` (quita su sección) o restaurar el snapshot.

## Open Questions

- RESUELTO (usuario): `install_modes: [full]` (D6).
- RESUELTO (usuario): modelo de la fase design = `sonnet`.
- RESUELTO (usuario): el fragmento suma UNA línea condicional: si las skills `emil-design-eng` (Emil Kowalski) o una skill de taste están instaladas, se cargan en el paso de diseño; sin dependencia dura. Esas skills NO se agregan al catálogo en este change (ver `design-skills-harnesses` en CHANGES.md).
- ABIERTO por diseño: formato de export de un artifact "Design"; el mínimo garantizado es `mockups/README.md` con la URL y el export se delega a las instrucciones que devuelve la herramienta.
- RESUELTO (criterio conservador): una TUI cuenta como UI dudosa; el fragmento dice que ante la duda se pregunta.
- FUERA DE SCOPE: fila `verify` en la tabla de modelos y slash command `/opsx:verify`.
- RESUELTO (8.1): `openspec archive` SÍ mueve `mockups/` junto con el change (verificado con un change descartable: `archive/<fecha>-<name>/mockups/README.md`; se borró después).
- RESUELTO (REFACTOR 4.5): se derivan. `isOwnedSection` consulta `ownedSectionIDs` (hijos anidados del orquestador) y `sectionFragments` (harnesses por sección); agregar un harness por sección requiere una sola edición.

## Decisiones tomadas en apply (delegadas: "elegí lo mejor", criterio conservador)

- **A1. `ownedSectionIDs` derivado**: se mantiene `ownedSectionIDs` solo para los hijos anidados del orquestador y `isOwnedSection` suma las claves de `sectionFragments`. Una sola edición por harness nuevo.
- **A2. Verify: filtro por agente para TODOS los harnesses** (no solo `design-mockups`) en `ChecksForHarness`. Solo puede quitar checks de agentes que el harness no soporta (catálogo real: p. ej. `sdd-orchestrator` ya no se verifica en gemini, donde nunca se instala). Es más simple y consistente que un caso especial; los tests existentes no cambian.
- **A3. `checkConfig` usa `h.ID` como marker** y se elimina la constante `sddOrchestratorSectionID` (sin otros usos).
- **A4. Los tests de uninstall (6.1) son de caracterización**: la lógica de `markerRemovalStep` ya era genérica por ID. `TestMarkerRemoval_Orchestrator_KeepsDesignMockups` falla si se quita el registro en `sectionFragments` (verificado).
- **A5. Los tests de `internal/install` (2.3) también son de caracterización** sobre el catálogo real, porque dependen solo de la entrada de catálogo y de `depends_on`.
- **A6. Golden claude**: regenerado solo con `-update`; el diff son exactamente 3 líneas (paso `openspec-verify` y su conector). Golden opencode sin diff; codex no tiene golden y su test sigue verde.
- **A7. Fragmento en inglés**, igual que los demás assets (aprobado en el checkpoint 4.6).
