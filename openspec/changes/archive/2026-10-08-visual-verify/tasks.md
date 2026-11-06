# Tasks

> Strict TDD: en cada grupo, primero RED (test que falla por la razón correcta), después GREEN mínimo, TRIANGULATE y REFACTOR. Antes de tocar archivos existentes, correr el SAFETY NET del paquete y anotar la línea base; si hay fallas preexistentes, reportarlas y no arreglarlas. No hacer build salvo pedido explícito. No commitear salvo pedido explícito. No tocar `filemerge`, `backup` ni `pipeline` (ALTO): si surge la necesidad, frenar y reportar.

## 1. Catálogo: harness `playwright` (BAJO)

- [x] 1.1 SAFETY NET: correr `go test ./internal/catalog/...` y anotar la línea base (N tests en verde).
- [x] 1.2 RED: en `internal/catalog/catalog_test.go`, agregar `TestPlaywrightHarness_Declared`. Debe exigir:
  - `ByID("playwright")` existe;
  - `Type == external` y `External.Method == "mcp"`;
  - `External.URL == ""`;
  - `External.MCP.Name == "playwright"` y `Command == "npx"`;
  - `Agents == [claude]`, `InstallModes == [full]` y `DependsOn` vacío.

  Verificar que falla con "playwright not found".
- [x] 1.3 RED: agregar `TestPlaywrightHarness_VersionPinned`. Exactamente un arg debe matchear `^@playwright/mcp@\d+\.\d+\.\d+$` y ningún arg debe contener `@latest` ni `@next`. Agregar también `TestPlaywrightHarness_ModeAndAgentScope`, que exige:
  - `ForMode(full)` contiene `playwright` y `ForMode(lite)` no lo contiene;
  - `ForAgent(opencode)` y `ForAgent(codex)` no lo contienen;
  - `ForAgent(claude)` sí lo contiene.

  Verificar que fallan.
- [x] 1.4 RED: en `TestForMode_ExcludesStarterOnlyHarnesses` y en `wantFoundationGlobal`, sumar +1 al literal vigente de Full/Custom (design D7): 16 → 17 si `design-skills-harnesses` no se aplicó, 18 → 19 si ya se aplicó. Lite queda en 6. Comentar `+1 playwright (visual-verify)`. Verificar que falla por el conteo.
- [x] 1.5 Reconfirmar la versión con `npm view @playwright/mcp dist-tags` (al 2026-10-08, `latest` = `0.0.83`). Si cambió, usar el `latest` estable, nunca `next`/alpha.
- [x] 1.6 GREEN: agregar la entrada `playwright` a `internal/catalog/harnesses.yaml`, en la sección de externos después de `context7`, con la forma de design D1 y `description`/`long_description` bilingües (es/en). La `long_description` debe nombrar a Playwright MCP de Microsoft y el registro stdio vía `npx` con versión fija; decir que es solo para Claude, que requiere Node ≥ 18 y que no instala navegadores, y que lo usa la verificación visual de `design-mockups`. Verificar con `go test ./internal/catalog/...` en verde (`catalog.Load()` valida).
- [x] 1.7 TRIANGULATE: agregar en `TestDesignMockups*` (o en uno nuevo, `TestDesignMockups_NoHardDependencyOnPlaywright`) que `design-mockups.DependsOn` contiene `sdd-orchestrator` y no contiene `playwright`. Verificar en verde.
- [x] 1.8 Revisar que no haya otros tests rotos por el conteo o por listas de harnesses, por ejemplo pickers de TUI/GUI o tests de `internal/install` que enumeran Full: `go test ./internal/install/... ./internal/tui/...`. Si alguno enumera explícitamente, ajustarlo por delta (+1 `playwright`). Verificar en verde.

## 2. Alcance por agente en `external.Install` (MEDIO, checkpoint)

- [x] 2.1 SAFETY NET: correr `go test ./internal/harness/external/...` y anotar la línea base.
- [x] 2.2 RED: en `internal/harness/external/mcp_test.go` (o en un nuevo `agent_scope_test.go`), agregar `TestInstall_SkipsUnsupportedAgents`. Un harness MCP stdio con `Agents: [claude]` se instala con adapters fake de claude y opencode. Debe exigir que el config de opencode quede byte a byte igual (o inexistente si no existía) y que el de claude tenga `mcpServers.playwright` de `type stdio` con `command npx` y los args pinneados. Verificar que falla porque opencode recibe `{"type":"remote","url":"/mcp"}`.
- [x] 2.3 RED: agregar `TestInstall_AgentAgnosticExternalStillInstallsForAll`. Un harness sin `Agents` (forma `context7`) con adapters de claude y opencode debe escribir en ambos configs. Verificar que pasa: es una regresión que fija el comportamiento actual.
- [x] 2.4 GREEN: en `internal/harness/external/install.go`, filtrar los adapters con `h.SupportsAgent(a.Agent())` antes del dispatch por método, con un helper `supportedAdapters` (design D3). Verificar 2.2 y 2.3 en verde.
- [x] 2.5 TRIANGULATE: agregar `TestInstall_Playwright_Idempotent`. Dos installs de `playwright` para claude: la segunda devuelve `AlreadyInstalled == true` y hay una sola entrada `playwright`. Agregar también `TestInstall_Playwright_PreservesOtherServers`: un `~/.claude.json` fake con `context7` conserva esa entrada intacta. Verificar en verde.
- [x] 2.6 REFACTOR: dejar comentado el helper con la referencia al patrón de `config.Install`. Verificar `go test ./internal/harness/...` en verde.
- [x] 2.7 Checkpoint MEDIO: resumir el diff del filtro al orquestador antes de seguir (comportamiento sin cambios para externos sin `agents`).

## 3. Verify post-install de `playwright` (MEDIO)

- [x] 3.1 SAFETY NET: correr `go test ./internal/verify/...` y anotar la línea base.
- [x] 3.2 Tests de comportamiento (design D6; deberían pasar sin código de producción). En `internal/verify/harness_checks_test.go`, agregar `TestHarnessChecks_Playwright`. Con el harness del catálogo y adapters de claude y opencode, debe haber exactamente un check con ID `external:mcp-config:playwright:claude` (no soft), ningún check `external:mcp-endpoint:playwright` y ningún check para opencode. Si falla, frenar y reportar: indicaría un cambio de comportamiento, no un test a ajustar. Verificar en verde.

## 4. Fragmento `design-mockups.md`: paso de verificación visual (MEDIO)

- [x] 4.1 SAFETY NET: correr `go test ./internal/harness/config/... ./internal/install/... ./internal/uninstall/...` y anotar la línea base. Incluye los golden de compose, que no deben cambiar.
- [x] 4.2 RED: en `internal/harness/config/design_mockups_fragment_test.go`, extender la tabla de `TestDesignMockupsFragment_Content` con substrings para:
  - `Visual verify`, `Playwright`, `dev server`, `mobile`, `desktop`;
  - `console`, `failed network requests`, `mockups/verify/`;
  - `Approval: approved`, `blocker`, `reproduction steps`;
  - `Never claim`, `NOT automated`, `manual checklist`;
  - el harness `playwright`, `soft dependency`, `passed | differences | manual`.

  Verificar que falla.
- [x] 4.3 RED: en `internal/harness/config/design_mockups_install_test.go`, agregar `TestDesignMockups_UpgradeReplacesOldSectionInPlace`. Un `CLAUDE.md` con la sección `design-mockups` vieja (contenido distinto entre markers), más la sección del orquestador y contenido de usuario, debe quedar tras el install con exactamente un marker `design-mockups`, el contenido nuevo (contiene `Visual verify`) y la sección del orquestador y el contenido de usuario byte a byte iguales. Verificar que falla porque el contenido nuevo todavía no existe.
- [x] 4.4 GREEN: en `internal/harness/config/assets/design-mockups.md`, agregar el paso 7 según el borrador de design D4 y la línea opcional `Visual verify: ...` en el formato del paso 2. Sin `https://`, sin `claude.ai/`, sin prefijos `mcp__` y sin puertos fijos. No tocar `sectionFragments` ni `isOwnedSection`. Verificar 4.2 y 4.3 en verde, y `TestDesignMockupsFragment_DoesNotHardcodeTypeURL` en verde.
- [x] 4.5 TRIANGULATE: confirmar en verde los tests existentes de idempotencia (doble install sin cambios), purge (sdd-orchestrator no purga la sección), uninstall limpio y `verify` (marker exactamente una vez). Confirmar también que `testdata/compose_all_toggles_claude.golden` y el de opencode no cambiaron: `go test ./internal/harness/config/... ./internal/install/... ./internal/uninstall/... ./internal/verify/...`.
- [x] 4.6 GREEN (catálogo): sumar a la `long_description` es/en de `design-mockups` la frase de design D5 (la verificación visual usa Playwright MCP si está instalado; sin él, checklist manual). Verificar `go test ./internal/catalog/...` en verde.

## 5. Documentación

- [x] 5.1 `README.md`:
  - fila **Playwright MCP** en la tabla del catálogo (external · full · solo Claude, versión pinneada, Node ≥ 18, no instala navegadores; desinstalar no remueve la entrada de `~/.claude.json`, igual que los otros MCP);
  - en la fila de **Design Mockups**, mencionar la verificación visual en `verify` con degradación manual;
  - nombrar a Playwright en la descripción del modo Full.

  Verificar releyendo que los nombres coinciden con el catálogo.
- [x] 5.2 `ARCHITECTURE.md`: sumar `playwright` (`external`, `method: mcp`, `agents: [claude]`, `[full]`) al ejemplo de catálogo §3, y una nota de una línea en §2 sobre el paso visual dentro de la sección `design-mockups` (sin sección nueva). Verificar releyendo.
- [x] 5.3 `CLAUDE.md` y `AGENTS.md` (espejo): en el párrafo de `design-mockups` §2, agregar que la sección incluye la verificación visual en `verify` vía el harness `playwright` (dependencia suave). Verificar con un diff que ambos párrafos quedaron equivalentes.
- [x] 5.4 `CHANGES.md`: marcar `visual-verify` como `✅ IMPLEMENTADO (pend. archive)` en la ficha y en la tabla resumen. Actualizar la ficha con las decisiones: harness `playwright` pinneado, paso 7 en el fragmento existente, dependencia suave y filtro por agente en `external.Install`. Registrar la deuda (`collectWritePaths` sin filtrar y uninstall de MCP en Claude). Dejar Figma MCP como opcional futuro. Verificar releyendo.

## 6. Integración

- [x] 6.1 Correr la suite completa `go test ./...` en verde. Comparar contra las líneas base de 1.1, 2.1, 3.1 y 4.1, y reportar cualquier falla preexistente por separado.
- [x] 6.2 Correr `openspec validate visual-verify --strict` sin errores, y marcar las tasks completadas.
