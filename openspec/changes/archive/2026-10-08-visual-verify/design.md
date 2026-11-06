# Design

## Context

La motivación está en proposal.md (sección Why). Este documento parte del estado actual del código:

- **Harnesses externos MCP.** El catálogo declara `context7` como `type: external`, `method: mcp`, con `url` remota más un bloque `mcp` stdio (`npx -y --package=@upstash/context7-mcp@2.2.5 -- context7-mcp`). La versión ya está pinneada, así que existe un precedente de pin. No declara `agents`, por lo que aplica a todos los agentes.
- **`installMCP`** (`internal/harness/external/mcp.go`) hace backup, `filemerge.MergeJSONObjects` y `WriteFileAtomic` contra `adapter.MCPConfigPath`. Para Claude escribe en `~/.claude.json` → `mcpServers.<mcp.Name>` = `{type: stdio, command, args}` (con `__replace__`). Para Codex usa TOML con el bloque `mcp`. Para OpenCode siempre usa `buildOverlay`, que arma una entrada remota a partir de `External.URL`.
- **El paso de install externo no filtra por agente.** `install.buildHarnessStep` le pasa a `externalStep` todos los adapters seleccionados, y `external.Install` no consulta `h.SupportsAgent`. `internal/harness/config/install.go` sí lo hace, en la línea `if !h.SupportsAgent(adapter.Agent()) { continue }`. Hoy esto no muerde porque ningún externo declara `agents`. Un harness MCP sin `url` y restringido a `claude` sí mordería: con `claude` y `opencode` seleccionados, OpenCode recibiría `{"type":"remote","url":"/mcp"}`.
- **Verificación post-install.** `verify.checkExternal` con `method: mcp` agrega un check duro de config MCP. El check de endpoint solo aparece si `URL != ""`. `harness_checks.go` ya filtra por `SupportsAgent`.
- **Uninstall.** En externos MCP para Claude/OpenCode el resultado es `externalSkipStep`: la entrada MCP no se remueve, igual que hoy con `context7`. Solo Codex tiene un paso de remoción.
- **Sección `design-mockups`.** `sectionFragments` (`assets.go`) mapea `design-mockups` → `design-mockups.md`, e `isOwnedSection` protege la sección del purge del orquestador. Install compara el contenido y reemplaza entre markers. `TestDesignMockupsFragment_DoesNotHardcodeTypeURL` exige que el fragmento no contenga `https://` ni `claude.ai/`.
- **Playwright MCP (fuentes oficiales, verificadas el 2026-10-08).**
  - README de `github.com/microsoft/playwright-mcp`. Comando para Claude Code: `claude mcp add playwright npx @playwright/mcp@latest`. Config estándar: `{"command":"npx","args":["@playwright/mcp@latest"]}`. Requisito: Node.js 18 o superior.
  - Herramientas: `browser_navigate`, `browser_resize`, `browser_take_screenshot`, `browser_console_messages`, `browser_network_requests` y `browser_snapshot`.
  - Flags: `--headless` (por defecto abre el navegador visible, "headed"), `--viewport-size`, `--device`, `--isolated`, `--output-dir`, `--caps`.
  - Registro npm (`npm view @playwright/mcp`): `latest` = `0.0.83` (publicado el 2026-09-28), licencia Apache-2.0, `engines.node >=18`, bin `playwright-mcp`.
  - La guía "Synthetic Architect" no se usó como fuente. Lo que diga y no figure arriba se considera no verificado.

## Goals / Non-Goals

**Goals:**
- Registrar Playwright MCP para Claude con una versión reproducible.
- Que `verify` compare de verdad lo construido contra los mockups aprobados, y que, cuando no pueda hacerlo, lo diga.
- Mantener `design-mockups` instalable sin Playwright.
- Cero cambios en `filemerge`, `backup` y `pipeline`.

**Non-Goals:**
- Instalar navegadores. Playwright MCP los gestiona en runtime y el instalador no ejecuta `npx playwright install`.
- Diff visual por píxeles o tests de regresión visual automatizados: la comparación la hace el agente mirando las capturas.
- Figma MCP (queda como mejora futura, según CHANGES.md).
- Soporte de `playwright` para OpenCode/Codex/Gemini.
- Remover la entrada MCP de `~/.claude.json` al desinstalar. Es una limitación preexistente que comparten todos los externos MCP de Claude.
- Expone flags de Playwright (`--headless`, `--isolated`) como opciones del instalador.

## Decisions

### D1 — Harness `playwright`: `external` / `method: mcp`, sin `url`, con bloque `mcp` stdio pinneado

```yaml
- id: playwright
  name: Playwright MCP
  description: { es: "...", en: "..." }        # bilingüe
  long_description: { es: "...", en: "..." }   # bilingüe, obligatorio (global-picker)
  type: external
  external:
    method: mcp
    mcp:
      name: playwright
      command: npx
      args: ["-y", "@playwright/mcp@0.0.83"]
  install_modes: [full]
  agents: [claude]
```

- **Por qué `method: mcp` y no `npm`.** Con `npm` el instalador correría `npm install -g` y `verifyBinary` buscaría un binario `mcp` en el PATH, que no existe (el binario se llama `playwright-mcp`). `method: mcp` con el bloque `mcp` es la ruta stdio de `installMCP` que ya usa `context7` para Claude, y `npx` resuelve el paquete en runtime.
- **Por qué sin `url`.** Playwright MCP no tiene un endpoint remoto oficial. Sin `url`, `verify` no agrega el check de endpoint, y `catalog.Load()` no exige `url`: solo valida `method` y `mcp.Validate()`.
- **id = nombre MCP = `playwright`.** Mantiene la paridad con `context7`. Varias rutas usan `h.ID` (paths de config y backups) y otras usan `mcp.Name` (la clave del overlay y el check). Si coinciden no hay ambigüedad, y además el nombre coincide con el que usa el comando oficial (`claude mcp add playwright ...`), con lo que las herramientas quedan como `mcp__playwright__*`.
- **`-y`.** Igual que `context7`: evita el prompt interactivo de `npx` la primera vez.
- **Sin flags extra.** No se agregan `--headless` ni `--isolated`. Durante el desarrollo, ver el navegador es útil y el default oficial es "headed". El usuario puede editar su `~/.claude.json`.
- **Alternativa descartada: ejecutar `claude mcp add` como subproceso.** Rompe la regla de backup y merge propio del instalador, depende de que el CLI `claude` esté en el PATH y no es idempotente con nuestros markers.

### D2 — Pin de versión en vez de `@latest`

Se fija la versión exacta: `@playwright/mcp@0.0.83` (el `latest` al 2026-10-08).

- **A favor del pin.** El build es reproducible: el mismo binario de active-stack produce la misma config. Un release roto upstream no rompe a los usuarios. Respeta la preferencia del responsable y el precedente de `context7@2.2.5`.
- **En contra.** El paquete es `0.0.x` y publica seguido (9 versiones entre mayo y septiembre de 2026). El pin envejece y los usuarios no reciben fixes ni nuevas herramientas hasta el siguiente release del instalador. El cambio de versión obliga a sacar release.
- **Mitigación.** La versión vive en un único lugar (`harnesses.yaml`). La tarea de apply incluye reconfirmarla con `npm view @playwright/mcp dist-tags` y pinnear el `latest` estable de ese momento si cambió. Se elige la versión más reciente que lleve al menos unos días publicada, nunca `next`/alpha.
- Un test exige el formato `MAJOR.MINOR.PATCH` y rechaza `@latest`/`@next`. Así un cambio accidental a `@latest` falla en CI.

### D3 — `agents: [claude]` más filtro por agente en `external.Install`

- **Por qué solo Claude.**
  1. El único consumidor del MCP es el paso visual, que vive en un fragmento exclusivo de Claude.
  2. Con OpenCode, `installMCP` siempre usa `buildOverlay` (entrada remota por URL). Sin `url` generaría una entrada rota. Habilitarlo exigiría tocar la rama de OpenCode, lo que queda fuera de alcance.
  3. La guía de alcance del catálogo es la misma que `design-mockups` (`agents: [claude]`).
- **Fix necesario.** El catálogo restringe la oferta (`ForAgent`, `filterByAgents`), pero el paso externo recibe todos los adapters. Se agrega en `external.Install` un filtro `adapters = supportedAdapters(h, adapters)`, que saltea los que cumplen `!h.SupportsAgent(a.Agent())`, antes del dispatch por método. Es el mismo patrón que `config.Install`, y aplica a `npm`, `homebrew` y `mcp`.
  - Los harnesses sin `agents` (`openspec`, `engram`, `context7`) no cambian, porque `SupportsAgent` devuelve `true` con una lista vacía. Lo cubre un test de regresión.
  - Governance: `harness/external` es MEDIO. No toca `filemerge`, `backup` ni `pipeline`.
- **Efecto residual aceptado.** `collectWritePaths` (`internal/install/plan.go`) sigue agregando al snapshot el path MCP de OpenCode. El snapshot de un archivo que no se modifica es inocuo, así que no se toca `install/plan.go`. Queda registrado como deuda junto a la ya anotada en CHANGES.md ("los adapters no se filtran por harness... en el plan de install").

### D4 — El texto vive en el fragmento `design-mockups.md` existente (paso 7), no en una sección nueva

- **Elegida: extender `design-mockups.md`.** Es la opción conservadora. Se reutiliza el mismo marker `active-stack:design-mockups` sin tocar `sectionFragments` ni `isOwnedSection`, e install y purge siguen funcionando sin cambios. Reinstalar reemplaza el contenido entre markers: los usuarios existentes reciben el paso al reinstalar, sin duplicados. El paso visual solo tiene sentido cuando hay mockups aprobados, que es exactamente lo que gestiona esa sección. Por eso no se gana nada con instalarlo por separado.
- **Descartada: sección nueva `active-stack:visual-verify`, propiedad del harness `playwright`.** Un harness `external` no inyecta secciones (eso es de `config`). Habría que crear un tercer harness `config` y registrarlo en `sectionFragments`, con más superficie de install, uninstall y verify, y un acoplamiento de orden entre dos secciones que se referencian entre sí.
- **Restricciones del texto:**
  - nada de `https://` (la URL local se escribe como `localhost`);
  - nada de `claude.ai/`;
  - nombres de herramientas genéricos (`browser_navigate`, ...) sin prefijo `mcp__`, porque el prefijo depende del nombre con que el usuario registró el server;
  - sin hardcodear puertos: el puerto sale del script de dev.

**Borrador del paso** (se agrega después del paso 6 y antes de la línea `Model:`; en apply se ajusta el texto, no la sustancia):

```markdown
7. Visual verify (during verify): run only when `## UI Mockups` records
   `Decision: yes | provided` AND `Approval: approved`. Include these instructions
   in the verify sub-agent brief.
   - Playwright available (tools such as `browser_navigate`, `browser_resize`,
     `browser_take_screenshot`, `browser_console_messages`,
     `browser_network_requests`): start the dev server (script from package.json
     or the project's CLAUDE.md; ask if unknown), open each mockup screen on its
     localhost URL, capture it at mobile (390x844) and desktop (1440x900), read
     console errors and failed network requests, compare against `mockups/`, then
     stop the dev server you started. Save or reference captures under
     `openspec/changes/<name>/mockups/verify/`.
   - Report each difference with priority (blocker / major / minor), screen,
     viewport, expected (mockup ref), observed, and reproduction steps. List which
     screens and viewports were checked and which were not. Never claim a screen
     matches or works unless you opened and checked it.
   - Playwright not available: say explicitly "visual verification was NOT
     automated", give the user a manual checklist (each screen × mobile and
     desktop), and suggest installing the `playwright` harness (active-stack).
     This is a soft dependency: design-mockups works without it.
   - Record `Visual verify: passed | differences | manual (<date>)` under
     `## UI Mockups`.
```

- El paso 2 (el formato del registro en `## UI Mockups`) suma la línea opcional `Visual verify: ...`, para que el formato del registro quede documentado en un solo lugar.

### D5 — Dependencia suave: sin `depends_on` en ningún sentido

- `design-mockups` mantiene `depends_on: [sdd-orchestrator]` y no suma `playwright`. Con una dependencia dura, un Custom con `design-mockups` arrastraría Node/`npx` y un navegador a usuarios que solo quieren la etapa de diseño. Además, el spec archivado define el paso como opcional.
- `playwright` no depende de `design-mockups`: el MCP sirve para más cosas que este paso (debug y exploración de UI).
- **Criterio de instalación:**
  - **Full**: ambos están en `[full]`, así que entran juntos sin acoplamiento en el catálogo.
  - **Custom**: se eligen por separado. El fragmento recomienda `playwright` cuando falta.
  - La `long_description` de `design-mockups` (es/en) suma una frase: "la verificación visual en verify usa Playwright MCP si está instalado (harness `playwright`); sin él degrada a checklist manual".
- **Alternativa descartada: una "recomendación" estructural en el catálogo** (un campo `recommends`). Es un campo nuevo en `model.Harness` con UI en la TUI y la GUI: demasiado para una sugerencia textual.

### D6 — Verify post-install: sin cambios de producción, solo tests que fijan el comportamiento

`checkExternal` ya produce, para `method: mcp` sin URL, solo el check de config. El filtro por agente ya existe. Se agregan tests en `internal/verify` que fijan que, para `playwright` con `claude` y `opencode`, hay un solo check `external:mcp-config:playwright:claude` y ningún check `external:mcp-endpoint:playwright`. Si alguna vez falla, se trata como un cambio de comportamiento y se revisa con el responsable, no como un ajuste del test.

- **Fuera de alcance:** endurecer el check de Claude para que verifique la presencia de `mcpServers.playwright`, y no solo que el JSON sea válido. Cambiaría el contrato también para `context7`/`engram`. Se puede hacer en un change aparte.

### D7 — Conteos de catálogo y coordinación con `design-skills-harnesses`

`playwright` suma exactamente +1 a Full y +1 a Custom (los tests de conteo de `internal/catalog/catalog_test.go` son `wantFoundationGlobal` y `TestForMode_ExcludesStarterOnlyHarnesses`). Lite no cambia (6).

- **Regla independiente del orden.** En apply, leer el literal vigente en el archivo y sumar +1. No se escribe un número absoluto precalculado:
  - si `design-skills-harnesses` todavía no se aplicó: 16 → 17;
  - si ya se aplicó (+2): 18 → 19.
- Cada change ajusta por su propio delta, así que el orden no importa. Si se aplican en ramas paralelas, el merge de ese literal es un conflicto trivial: sumar ambos deltas.
- El comentario del test documenta el bump: `+1 playwright (visual-verify)`.
- Para no depender solo del conteo, se agregan aserciones de pertenencia: `playwright` está en Full y no está en Lite. Esas aserciones son estables ante cualquier orden.

### D8 — Orden TDD

Los tests RED se escriben primero contra el catálogo (fallan con `ByID("playwright")` no encontrado) y contra el fragmento (fallan con substrings ausentes). Después, el test de alcance por agente en `external` falla porque OpenCode recibe la entrada. Recién entonces se escribe el código de producción: YAML, fragmento y filtro.

## Risks / Trade-offs

- **[El pin queda viejo frente a un upstream 0.0.x que publica seguido]** → La versión está en un solo lugar y se reconfirma en apply. Un test impide `@latest`. Los bumps se hacen como changes chicos.
- **[El navegador no está instalado en la máquina del usuario y la primera llamada a `browser_navigate` falla]** → El instalador no gestiona navegadores (non-goal). El fragmento ordena reportar el error y degradar a checklist manual, diciéndolo explícitamente. El README no documenta qué hace Playwright MCP si falta el navegador; se resolverá con el primer uso real, sin afectar specs ni tasks.
- **[Requiere Node ≥ 18 y `npx` en el PATH]** → Es el mismo requisito de hecho que `context7` (también vía `npx`). La detección de dependencias existente ya cubre Node para `openspec`.
- **[El paso visual depende del criterio del modelo, no de un diff por píxeles]** → Es aceptado por diseño. Lo mitigan la regla "no afirmar sin comprobar", la lista de pantallas y viewports chequeados y no chequeados, y el registro `Visual verify:`.
- **[Levantar el servidor de desarrollo puede dejar procesos colgados]** → El fragmento exige detener el servidor que el agente levantó.
- **[Desinstalar `playwright` no remueve la entrada de `~/.claude.json`]** → Es una limitación preexistente que comparten todos los externos MCP de Claude (`externalSkipStep`). Se documenta en el README. Arreglarla tocaría la remoción JSON, cerca de `filemerge` (ALTO): queda fuera de este change.
- **[Snapshot inocuo del config de OpenCode en `collectWritePaths`]** → No se escribe nada. Queda como deuda ya registrada.
- **[Usuarios con la sección vieja de `design-mockups`]** → Al reinstalar se reemplaza en el lugar (idempotente, con backup). Si no reinstalan, siguen con el flujo anterior sin romper nada.

## Migration Plan

- No hay migración de datos. Reinstalar `design-mockups` actualiza la sección (filemerge reemplaza entre markers, con backup previo). Instalar `playwright` agrega `mcpServers.playwright` a `~/.claude.json` (con backup).
- **Rollback:** revertir el commit. Los usuarios ya instalados conservan la entrada MCP, que es inocua, y la sección extendida, que se reemplaza al reinstalar la versión anterior.

## Open Questions

- Qué hace Playwright MCP la primera vez cuando no hay navegador instalado (lo descarga solo o falla con instrucciones). El README oficial no lo documenta. No cambia specs, enfoque ni tasks: el fragmento ya cubre la falla con la degradación explícita.

## Decisiones tomadas en apply (delegadas: "elegí lo mejor", criterio conservador)

- **A1. Versión re-confirmada**: `npm view @playwright/mcp dist-tags` → `latest` = `0.0.83` (`next` es una alpha, descartada). Se mantiene el pin `@playwright/mcp@0.0.83`; no hubo cambio respecto de D2.
- **A2. Filtro por agente en `external.Install` (checkpoint MEDIO)**: `supportedAdapters(h, adapters)` se aplica una sola vez antes del dispatch por método (`npm`, `homebrew`, `mcp`), con el mismo criterio que `config.Install`. Para harnesses sin `agents` devuelve todos los adapters (regresión cubierta por `TestInstall_AgentAgnosticExternalStillInstallsForAll`). Decidido sin frenar: es MEDIO, no toca `filemerge`/`backup`/`pipeline` y no cambia el comportamiento de `openspec`, `engram` ni `context7`.
- **A3. Test extra de archivo preexistente de OpenCode**: además del caso "no existe", `TestInstall_SkipsUnsupportedAgents_LeavesExistingFileUntouched` fija que un `opencode.json` existente queda byte a byte igual (el spec lo exige con "identical").
- **A4. Conteos**: los literales vigentes eran 18 (Full/Custom) y 6 (Lite); pasan a 19/19/6 con el comentario `+1 playwright (visual-verify)`.
- **A5. Fragmento en el paso 7, con dos agregados menores al borrador de D4**: la frase "otherwise skip this step" (el spec exige que no corra con `Decision: no` ni sin aprobación) y "si el navegador falla al abrir, reportar el error y caer al camino manual" (cubre el riesgo documentado de navegador ausente). La línea `Visual verify:` se documenta también en el formato del paso 2. Sin `https://`, `claude.ai/`, `mcp__` ni puertos.
- **A6. `long_description` de `design-mockups`**: se agrega la frase de D5 en es y en en; un test (`TestDesignMockups_DescribesVisualVerification`) la fija.
- **A7. Golden del orquestador**: sin cambios (el fragmento `design-mockups.md` no entra en el compose del orquestador). Tampoco se tocó `sectionFragments` ni `isOwnedSection`.
