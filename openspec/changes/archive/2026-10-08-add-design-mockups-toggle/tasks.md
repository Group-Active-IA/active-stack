# Tasks

> Strict TDD activo. Cada grupo de código sigue SAFETY NET → RED → GREEN → TRIANGULATE → REFACTOR, ejecutando `go test` del paquete en cada gate. No correr `go build` salvo pedido explícito (regla del operador). Governance: `harness/config` y `verify` = MEDIO (checkpoints); `catalog` = BAJO. `filemerge`/`backup`/`pipeline` (ALTO) NO se modifican: solo se usa su API.
> Specs: `specs/design-mockups-harness/spec.md`. Decisiones: `design.md` (D1–D7).

## 1. Safety net y decisiones abiertas

- [x] 1.1 SAFETY NET: correr `go test ./internal/harness/config/... ./internal/catalog/... ./internal/verify/... ./internal/uninstall/... ./internal/install/...` y registrar la línea base ("N tests passing"). Si algo falla, PARAR y reportar como falla preexistente. Verificar: salida anotada en esta task.
  - Resultado: baseline 1344 `--- PASS` (incl. subtests) en `go test ./...`; todos los paquetes `ok`, sin fallas preexistentes (los 5 paquetes del alcance: config, catalog, verify, uninstall, install verdes).
- [x] 1.2 CHECKPOINT (usuario): resolver `install_modes` (`[full]` recomendado vs `[lite, full]`, design.md D6) y el modelo de la fila `design` (`sonnet` vs `opus`). Verificar: respuestas registradas en design.md (Open Questions).
  - Resuelto: `install_modes: [full]`, modelo `sonnet`, y una línea condicional en el fragmento para `emil-design-eng`/skill de taste (sin dependencia dura, sin catálogo).

## 2. Catálogo (BAJO)

- [x] 2.1 RED: en `internal/catalog/catalog_test.go`, `TestDesignMockups_InCatalog`: `ByID("design-mockups")` existe, `Type == config`, `SupportsAgent(claude)` true, `SupportsAgent(opencode|codex|gemini)` false, `DependsOn == [sdd-orchestrator]`, `InMode` según 1.2. Actualizar los conteos de `TestForMode_CustomReturnsOnlyGlobal` y `TestForMode_ExcludesStarterOnlyHarnesses` (Full/Custom 15→16; Lite 6→7 solo si 1.2 eligió `[lite, full]`). Verificar: fallan por la razón esperada.
- [x] 2.2 GREEN: agregar la entrada `design-mockups` en `internal/catalog/harnesses.yaml` (sección Configuración, después de `permissions`) con `description`/`long_description` es/en (paso de diseño de mockups con aprobación humana, solo Claude, opt-in, requiere la herramienta Artifact). Verificar: 2.1 pasa y la suite de catálogo (incluido `localized_validation_test.go`) sigue verde vía `catalog.Load()`.
- [x] 2.3 TRIANGULATE: test en `internal/install` que `CustomPickerHarnesses` para `[claude]` incluye `design-mockups` y para `[opencode]`/`[codex]` no; y que `BuildPlan` Custom con solo `design-mockups` + claude resuelve `sdd-orchestrator` como auto-dependencia ordenada antes. Verificar: pasan.

## 3. Fragmento y registro (harness/config, MEDIO)

- [x] 3.1 RED: `TestDesignMockupsFragment_Content` en `internal/harness/config`: el fragmento cargado contiene `## Design Stage — UI Mockups`, el orden `propose → **design** → apply`, `openspec/changes/<name>/mockups/`, las tres respuestas (yes / no / I already have them), `## UI Mockups`, la frase de excepción ("never block on phase gates"), `quickstart` + `intent "design"` y el fallback HTML estático. Verificar: falla (el asset no existe).
- [x] 3.2 GREEN: crear `internal/harness/config/assets/design-mockups.md` a partir del borrador de design.md (con el modelo resuelto en 1.2) y el registro de fragmentos por harness (D2). Verificar: 3.1 pasa. Revisar que ninguna aserción sea tautológica (cada una falla si se borra la línea del fragmento).
- [x] 3.3 RED: `TestPurgeStaleSections_KeepsDesignMockups` (en `purge_test.go`): un contenido con secciones `sdd-orchestrator`, `design-mockups` y una legacy (`persona`) → tras purgar quedan las dos primeras. Verificar: falla.
- [x] 3.4 GREEN: registrar `design-mockups` en `ownedSectionIDs` (D3). Verificar: 3.3 pasa y `purge_test.go` completo verde.

## 4. Instalación por sección (harness/config, MEDIO)

- [x] 4.1 RED: en `install_test.go`, `TestInstall_DesignMockups_ClaudeOnly`: `Install(designMockupsHarness, [claudeAdapter, opencodeAdapter(primary-agent), codexAdapter], home)` → el `CLAUDE.md` de claude contiene exactamente una sección `active-stack:design-mockups` con el heading; `AGENTS.md`/`opencode.json` de opencode y `AGENTS.md` de codex quedan byte a byte iguales (o inexistentes); no hay `agent.design-mockups` en ningún settings. Verificar: falla (hoy `Install` compone el orquestador).
- [x] 4.2 GREEN: generalizar `Inject` a `InjectSection(path, sectionID, content, snapshotDir)` dejando `Inject` como wrapper de `"sdd-orchestrator"`; en `Install`, despachar por registro de fragmentos: ruta por sección (filtra `h.SupportsAgent(adapter.Agent())`, salta `InstructionsPath` vacío, nunca primary-agent) vs ruta actual del orquestador. Verificar: 4.1 pasa y toda la suite del paquete sigue verde (incluido `primary_agent_test.go`).
- [x] 4.3 TRIANGULATE: `TestInstall_DesignMockups_Idempotent` (dos instalaciones → una sola sección, segunda `AllAlready == true`); `TestInstall_DesignMockups_PreservesOrchestratorAndUserContent` (archivo con contenido de usuario + sección orquestador → ambos intactos); `TestInstall_DesignMockups_BacksUpFirst` (fake `SnapshotterCreate` invocado antes de escribir; si falla, el archivo no se toca); `TestInstall_Orchestrator_KeepsDesignMockupsSection` (instalar `sdd-orchestrator` sobre un archivo con la sección design → sigue). Verificar: pasan.
- [x] 4.4 TRIANGULATE (invariancia): `go test ./internal/harness/config -run TestCompose` contra el golden opencode sin regenerarlo; `git diff internal/harness/config/testdata/compose_all_toggles_opencode.golden` vacío. `compose.go` sin cambios. Verificar: pasa.
- [x] 4.5 REFACTOR: evaluar derivar `ownedSectionIDs` del registro de fragmentos (D3, Open Question) y extraer helpers solo si mejoran la legibilidad. Verificar: suite verde tras cada paso.
- [x] 4.6 CHECKPOINT (MEDIO, escritura en config del usuario): mostrar al usuario el `CLAUDE.md` resultante de un install de prueba (temp dir) con orquestador + design-mockups, y el diff de reinstalar. Verificar: aprobación registrada.

## 5. Verify (MEDIO)

- [x] 5.1 RED: en `internal/verify/harness_checks_test.go`, `TestChecksForHarness_DesignMockups_ChecksOwnMarker` (CLAUDE.md con ambas secciones → check `config:design-mockups:claude` pasa; con solo la sección orquestador → falla nombrando `active-stack:design-mockups`) y `TestChecksForHarness_SkipsUnsupportedAgents` (adapters claude+opencode → sin checks para opencode). Verificar: fallan.
- [x] 5.2 GREEN: `checkConfig` usa `h.ID` como sección; `ChecksForHarness` filtra adapters por `h.SupportsAgent`. Verificar: 5.1 pasa y los tests existentes de `sdd-orchestrator`/`permissions` siguen verdes.

## 6. Desinstalación (sin código nuevo esperado)

- [x] 6.1 RED→GREEN: en `internal/uninstall` (`marker_step_test.go`/`marker_purge_test.go`), `TestMarkerRemoval_DesignMockups_Clean` (quita solo su sección, deja orquestador y contenido de usuario), `TestMarkerRemoval_DesignMockups_AbsentIsNoop` y `TestMarkerRemoval_Orchestrator_KeepsDesignMockups` (desinstalar solo el orquestador deja la sección design; depende de 3.4). Si alguno pasa de entrada, documentarlo como test de caracterización (la lógica ya existía) y confirmar que falla al revertir 3.4 donde aplique. Verificar: pasan.

## 7. Diagrama OPSX del base claude + golden

- [x] 7.1 RED: `TestCompose_ClaudeDiagramIncludesVerify`: `Compose([]string{}, "claude")` contiene `openspec-verify` entre `/opsx:apply` y `/opsx:archive` (comparar índices). Verificar: falla.
- [x] 7.2 GREEN: editar `internal/harness/config/assets/claude/sdd-orchestrator.md` (diagrama ~L48-61) insertando el paso `openspec-verify` (skill, no slash command; D7). NO tocar la tabla de modelos ni `assets/opencode`/`assets/codex`. Verificar: 7.1 pasa.
- [x] 7.3 TRIANGULATE: regenerar SOLO el golden claude con `go test ./internal/harness/config -run TestCompose_AllToggles$ -update`. El diff debe ser únicamente la línea verify. Verificar: suite verde y golden opencode sin cambios.

## 8. Verificación de supuestos

- [x] 8.1 Verificar el `(TBD)` de archive: crear un change de prueba descartable con un subdirectorio `mockups/`, correr `openspec archive` y confirmar que se mueve al archive. Borrarlo después. Verificar: resultado anotado en design.md (Open Questions).

## 9. Documentación

- [x] 9.1 `CLAUDE.md` y `AGENTS.md` (espejo, misma pasada): en §2 agregar el harness `design-mockups` (config, solo claude, opt-in, sección propia `active-stack:design-mockups`) y aclarar que los toggles del `sdd-orchestrator` no se exponen en UI. Verificar: `grep -n "design-mockups" CLAUDE.md AGENTS.md` con texto equivalente en ambos.
- [x] 9.2 `README.md`: fila nueva en la tabla de harnesses para Design Mockups (solo Claude, gate de aprobación, mockups en `openspec/changes/<name>/mockups/`, requiere Artifact). Verificar: grep.
- [x] 9.3 `ARCHITECTURE.md`: en §2.1 aclarar que hoy los toggles salen del catálogo (no los elige el usuario) y que por eso las capacidades opt-in van como harness `config` propio con sección propia (`ownedSectionIDs`); sumar `design-mockups` al catálogo de ejemplo §3. La nota de toggles fijos ya no condiciona este change. Verificar: grep.
- [x] 9.4 `CHANGES.md`: ficha de `add-design-mockups-toggle` (entregable: harness `design-mockups`; decisión D1 y la descartada; governance MEDIO con escritura ALTO-impacto vía API existente); ficha PENDIENTE `stitch-design` (harness hermano `config` para `opencode` con herramientas MCP `stitch__*`; reutiliza la ruta por sección de D2; depende de validar el flujo en Claude); deuda "adapters sin filtrar por harness en `sdd-orchestrator`/install plan"; nota de mejora opcional: que `roadmap-generator`/`active-orchestrator` (repos externos Group-Active-IA) marquen `UI: sí/no` por change como pista, con la pregunta en runtime como fuente de verdad. Actualizar árbol de dependencias y tabla resumen. Verificar: grep de `design-mockups` y `stitch-design` en CHANGES.md.

## 10. Integración final

- [x] 10.1 Correr `go test ./...` y comparar contra la línea base de 1.1 (mismo conteo + los tests nuevos, sin regresiones). Verificar: salida verde anotada.
- [x] 10.2 `openspec validate add-design-mockups-toggle --strict` sin errores y todas las tasks marcadas. Sin commit salvo pedido explícito del usuario. Verificar: salida del validate.
