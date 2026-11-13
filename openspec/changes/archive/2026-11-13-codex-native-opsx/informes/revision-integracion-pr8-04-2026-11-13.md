# Revisión independiente de la integración de main y cambio 04

13/11/2026. Revisión técnica independiente del integrador, en un agente distinto. El alcance es preservar la entrega Claude ya integrada en main y comprobar que 04 selecciona su propio recorrido Codex, sin implementar funcionalidades nuevas de 03.

## Revisión inspeccionada

- HEAD de 04 previo a la integración: `e4c7e4846ecbe3311ad8e089a2d0345460027d5b`.
- `origin/main` y `MERGE_HEAD`: `7a600895e8f1fe1f3728cffa1fcb57dca5663a59`, con PR #8 integrado.
- Se revisó el árbol combinado resuelto antes del commit de merge. Los resultados no se atribuyen solamente a ninguno de los dos SHA anteriores. La publicación debe identificar el commit final que contiene esta resolución y los tests.
- No se ejecutaron build, commit ni push en esta revisión. Los escenarios usan homes temporales.

## Resultado de la lectura del contrato

No se identificaron bloqueantes nuevos en el código de conciliación inspeccionado. La revisión cubrió estas condiciones concretas:

| Área | Comprobación |
|---|---|
| Interfaz y enum | Se conserva un único `ConfigDeliveryManagedSkill`. `SkillsDir` se obtiene como capacidad opcional para entrega nativa; la rama Claude hace el cast antes de utilizarla. No se exige a un adapter de instrucciones una capacidad que no usa. |
| `InstallWith` | Conserva wrapper, `Home`, `Target`, `ShadowSkillsDir`, `DetectVersion` y `Layers` de main. Codex no usa el detector/versionado o la composición Claude. Claude mantiene su selección skill/legacy por versión. |
| Resultados parciales | `Files`, `Warnings`, `Modes` y `ManagedWrites` sobreviven a un error de adapter posterior. `AllAlready` no anuncia éxito en ese retorno de error. |
| Pipeline y rollback | El paso registra estados y padres y coloca guardas antes de devolver el error de instalación. Conserva guardas de existencia/ausencia, detección de padres sustituidos y protección de notas nuevas de Codex en snapshots compartidos. |
| Machine y Project | Machine selecciona skill propia por cliente. Codex Project mantiene instrucciones legadas y diagnóstico; Claude Project conserva target, ubicación de backups y chequeo de shadowing de main. Las rutas se resuelven mediante adapters. |
| Avisos | El modo `skill` de Codex no emite el aviso BREAKING de Claude. Project Codex soportado continúa avisando por stderr sin convertirse en evento `Degraded`. Los notices Claude conservan su recorrido anterior. |
| Snapshot | Las rutas de la skill Claude mantienen el colector de main; las rutas e inventario Codex usan su colector propio. No se aplica la operación de ownership Codex a Claude. |
| Verify | Claude comprueba su policy y marker `universal-rules`; Codex comprueba inventario, ownership y `allow_implicit_invocation: false`. Codex Project conserva el chequeo de instrucciones legado mediante el adapter de target. |
| Retiro | El dispatcher no envía Claude al retiro Codex ni viceversa. El retiro Codex conserva archivos ajenos y el contenido del otro cliente. |
| Documentación | `AGENTS.md` y `CLAUDE.md` resultaron idénticos por `cmp`; el índice no conserva conflictos sin resolver en el corte revisado. |

Se comprobó con `git diff origin/main` que `internal/agents/claude/`, `layered_install.go`, `skill_inject.go` y `compose_layers.go` no cambian respecto del main integrado. La adaptación común en `installManaged` sustituye el acceso directo a `SkillsDir` por su capacidad opcional; no cambia el mecanismo de entrega Claude.

El retiro Claude de main usa `removeLayered` y elimina su directorio administrado completo cuando la marca de propiedad corresponde. Esta revisión conserva ese comportamiento heredado; no extiende a Claude la garantía de preservación de notas internas que 04 implementa para Codex. Cualquier cambio propio de Claude corresponde a su responsable y no se incorpora bajo esta conciliación.

## Ejecución independiente

| Comando | Resultado y alcance |
|---|---|
| `go test -count=1 -json ./internal/harness/config ./internal/install ./internal/uninstall ./internal/verify` | Exit 0: 461 eventos de tests/subtests PASS — config 205, install 123, uninstall 70, verify 63 — y 1 skip de la sonda opt-in. Este corte precede a la triangulación final agregada por el integrador. |
| `go vet ./internal/harness/config ./internal/install ./internal/uninstall ./internal/verify` | Exit 0. |
| `go test -count=1 -json ./internal/harness/config ./internal/install ./internal/uninstall -run '^TestSharedManagedDelivery'` | Exit 0: 18 eventos PASS después de la triangulación del integrador. Cubre instalación/reinstalación/actualización en ambos órdenes; resultados parciales; fallback Claude; Machine/Project; verify y retiro por cliente. |
| `go test -count=1 -json ./internal/install -run '^TestSharedManagedDeliveryRollbackPreservesConcurrentCodexEdits$'` | Exit 0: 3 eventos PASS — test y dos órdenes. Prueba adicional descrita debajo, agregada después de la corrida anterior. |
| `go vet ./internal/install` | Exit 0 después de añadir la prueba de rollback combinado. |
| `ACTIVE_STACK_CODEX_PROBE=1 go test ./internal/harness/config -run TestManagedSkillCodexDiscovery -count=1 -v` | PASS con Codex CLI `0.167.0-alpha.3.2`, macOS Intel. El cliente descubre el paquete realmente instalado en el fixture temporal y los apoyos no aparecen como skills adicionales. No se inicia un turno de modelo. |
| `git diff --check`, `git diff origin/main --check` y `cmp AGENTS.md CLAUDE.md` | Exit 0 para el árbol de trabajo, el alcance de 04 respecto de main y el espejo. `git diff --cached --check` informa una línea vacía final en `openspec/changes/archive/2026-11-13-claude-native-opsx/investigacion/pruebas-claude-real.md:56`, heredada de main/03. No se editó ese informe histórico ni se presenta el chequeo global del índice como verde. |

Los conteos son eventos de tests/subtests, no tareas OpenSpec. No se suman corridas repetidas para presentar una cantidad de casos únicos. La suite global y race del integrador constituyen evidencia separada; esta revisión no los ejecutó ni los presenta como propios.

### Triangulación de rollback combinado

Se añadió únicamente `internal/install/mixed_rollback_integration_test.go`, sin cambios de producción. Se seleccionan Claude y Codex en ambos órdenes, se comprueba que ambas entregas llegaron al verify posterior, se edita después de instalar el `AGENTS.md` Codex y se crea una nota ajena dentro de su skill. El verify falla deliberadamente.

Los dos casos comprueban que:

- El error original permanece en `Apply.Err` y el conflicto de preservación se informa en `Rollback.Err` con la ruta Codex.
- El archivo Codex editado y su nota nueva conservan exactamente sus bytes.
- Las instrucciones Claude no editadas vuelven exactamente a su contenido original.
- Los archivos `SKILL.md` nuevos de ambos clientes se revierten, sin borrar la nota Codex.

La primera ejecución del test falló por una expectativa incorrecta del propio test: buscaba el error de verify en `ExecutionResult.Err`, que prioriza el error de rollback cuando lo hay. Se corrigió la aserción para revisar `Apply.Err` y `Rollback.Err` por separado. No fue un fallo de producción ni se presenta como un nuevo RED del dispatcher.

### Evidencia del ciclo TDD

| Trabajo | Archivo | Safety net | RED | GREEN | Triangulación / refactor |
|---|---|---|---|---|---|
| Revisión y reejecución del despacho mixto | `shared_delivery_integration_test.go` en config/install/uninstall | 461 PASS y 1 skip, corte anterior a triangulación final | RED del integrador documentado en su informe; no repetido por esta revisión | 18 eventos PASS dirigidos | Órdenes, targets, fallback, resultados parciales y bytes del otro cliente |
| Rollback mixto con edición posterior | `internal/install/mixed_rollback_integration_test.go` | Suite de install verde dentro del corte anterior | No hay nuevo cambio de producción; triangulación del contrato ya sometido a RED por el integrador | 3 eventos PASS dirigidos | Ambos órdenes; archivo editado, nota ajena, conflicto y restauración Claude; gofmt/vet verdes |

## Logs locales y huellas

Los logs siguientes son artefactos locales de reproducción, no un nuevo esquema de recibos de 06. Sus huellas acreditan la integridad de los archivos indicados; no sustituyen la procedencia ni la suficiencia de la evidencia.

| Archivo local | SHA-256 |
|---|---|
| `/tmp/opsx04-independent-integration.jsonl` | `e6202cf4cf17acdcc469ff8c5d96df3f0ca965a478799770d0dbf0ecc77565e2` |
| `/tmp/opsx04-independent-integration-vet.log` | `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` |
| `/tmp/opsx04-independent-mixed-final.jsonl` | `b29a4cfdb9d70bdfec264495506ce52d8694a1d49608fde202cfd0f7b49fb4ba` |
| `/tmp/opsx04-independent-mixed-rollback-final.jsonl` | `6c440a601c1464181a1c02b64d300107c3dbc98407bc3754e796a57e23ef8338` |
| `/tmp/opsx04-independent-mixed-rollback-vet.log` | `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` |
| `/tmp/opsx04-independent-discovery-final.log` | `f476a597d21c145ed684e9ab3371c0d95e745682b37b4beb4a41f8bce5bb74f6` |

### Identidad del runtime revisado

| Ruta | SHA-256 |
|---|---|
| `internal/harness/config/install.go` | `b91a4cf6a11d3a85d63c086c51db0fd3b6e6f53a32a5a16e879c36ae4857cf3d` |
| `internal/harness/config/types.go` | `09d0059f2361036de233569f9e1f8766df9c2bc80de01bebdacfdaa3e230fe0c` |
| `internal/install/steps.go` | `9750c19368e976e298487daca39280aa7a748c9e23efe70c2bf5430bf70697a0` |
| `internal/install/plan.go` | `260520127aa4b6ca3235940f36c4a33666ff417a135327971c3e39a7dd0653ae` |
| `internal/uninstall/steps.go` | `2cbb2408fdfe41aa336a919d25aa3a69dc66eaaaa81c205a06788efbbc490dfe` |
| `internal/verify/harness_checks.go` | `09b1e533c7176a25483e1f0d4888c1e99e174edd9d208fb69984a73de35b6da6` |

## Alcance de la conclusión

Esta revisión aporta evidencia de compatibilidad técnica del instalador combinado y de preservación de las guardas de 04. No acredita activación funcional Codex, especialistas, permisos efectivos ni CLI/IDE Windows/Linux. Tampoco cierra contratos corregidos de 05/06, la comparación de 02 o la disposición humana de revisión 01/03/08. La sonda local es de descubrimiento; la revisión técnica independiente no sustituye la confirmación del responsable de 03 ni la aceptación humana de la migración.

No se marcaron tareas OpenSpec como completas desde esta revisión. El cierre global de 04 debe conservar los límites y pendientes de sus artefactos.
