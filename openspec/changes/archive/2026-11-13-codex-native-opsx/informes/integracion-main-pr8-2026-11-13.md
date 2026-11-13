# Integración de main tras PR #8 — cambio 04

13 de noviembre de 2026. Base de 04: `e4c7e4846ecbe3311ad8e089a2d0345460027d5b`. Base incorporada: main `7a600895e8f1fe1f3728cffa1fcb57dca5663a59`, con la entrega Claude del PR #8 en `01f944c`. Se usó `git merge --no-commit origin/main`, sin rebase ni force-push. Este informe identifica el árbol combinado antes de su commit de integración.

## Resultado técnico

Se conciliaron once conflictos de tipos, configuración, plan, pasos, verificación, retiro, instrucciones, arquitectura y matriz. La implementación de Claude proviene de main; esta entrega integra la compatibilidad de 04 sin implementar features de 03, sin editar sus assets y sin usar el candidato anterior de integración abandonado.

- Un solo `ConfigDeliveryManagedSkill=2`, con valores previos 0/1 conservados.
- `Install` sigue llamando a `InstallWith`. Se preservan `Home`, `Target`, `ShadowSkillsDir`, `DetectVersion` y `Layers` de Claude; Codex no utiliza su detección ni su compositor de capas.
- La interfaz común mantiene opcional la capacidad de rutas de skills y la comprueba en cada recorrido nativo. Se conserva el error de interfaz opcional ausente de 04.
- Codex Machine usa su instalación administrada; Codex Project mantiene el legado y el diagnóstico. Claude conserva entrega por versión, Project y shadowing. Los assets, adaptador Claude y lógica propia de sus instaladores por capas no se reescribieron.
- Los resultados conservan `Modes` y `Warnings` de Claude y `ManagedWrites` de Codex, también si falla otro adapter después. El pipeline aplica las guardas antes de devolver error.
- Snapshot, verify y uninstall seleccionan por cliente. Claude reconoce `universal-rules`; Codex reconoce su marker `sdd-orchestrator`, inventario y policy. No se interpreta un paquete como si perteneciera al otro cliente.
- El aviso de Project Codex sigue en stderr sin evento Degraded. Un modo `skill` de Codex no dispara un aviso BREAKING de Claude. Se conserva el canal preexistente de avisos Claude.
- AGENTS/CLAUDE permanecen en espejo y la arquitectura y la única matriz describen ambos recorridos.

## TDD Cycle Evidence

| Tarea | Test File | Layer | Safety Net | RED | GREEN | TRIANGULATE | REFACTOR |
|---|---|---|---|---|---|---|---|
| Compatibilidad de instalación y actualización | `internal/harness/config/shared_delivery_integration_test.go` | Integración aislada | Seis paquetes aprobados antes del merge | Ambos órdenes fallaron al enviar Claude al instalador Codex | Ambos órdenes instalan, reinstalan y actualizan módulos | Resultado parcial con tercer adapter fallido en ambos órdenes; fallback insuficiente/desconocido Claude conserva Codex | Selección por cliente; suites, vet y race verdes |
| Verify por cliente | Mismo archivo | Integración aislada | Mismo safety net | Claude falló como paquete Codex | Cada cliente verifica su marker y política | Coexistencia de dos paquetes y rutas distintas | Sin alterar verificadores internos del cliente |
| Rutas y opciones por target | `internal/install/shared_delivery_integration_test.go` | Pipeline aislado | Mismo safety net | No se atribuye un RED adicional: triangulación del despacho anterior | Machine y Project, con ambos órdenes de clientes | Claude nativo por target, Codex nativo solo Machine, Project legado | Sin expansión del alcance Project de 04 |
| Retiro independiente | `internal/uninstall/shared_delivery_integration_test.go` | Plan aislado | Mismo safety net | Retiro Claude falló al enumerar inventario Codex; Codex ya pasaba | Retiro de cada cliente usa su estrategia | Instrucciones y SKILL del otro conservadas byte a byte; notas Codex preservadas | Enumeración de rutas por cliente |

Los tests de triangulación se agregaron después del arreglo de despacho y no se presentan como RED propios. El primer intento de ampliar la prueba de retiro a notas dentro de la skill de Claude detectó que el `removeLayered` preexistente de 03 elimina ese directorio completo. No se modificó esa política desde 04 ni se atribuye a esta integración. La aceptación probada aquí exige notas Codex preservadas y ausencia de cambios en el cliente no seleccionado; no declara preservación de notas ajenas dentro de la skill Claude seleccionada.

## Validación de la revisión combinada

Entorno local: `go version go1.27.3 darwin/amd64`. Se usaron homes temporales y detectores de versión Claude inyectados; los tests de integración no realizaron inferencia de clientes reales.

- Antes del merge: `go test -count=1 ./internal/backup ./internal/harness/config ./internal/filemerge ./internal/install ./internal/verify ./internal/uninstall`, aprobado.
- RED: `go test -count=1 ./internal/harness/config ./internal/uninstall -run '^TestSharedManagedDelivery' -v`, fallos reales de despacho anteriores al arreglo, registrados en `/tmp/opsx04-main-integration-red.log`.
- GREEN/triangulación: `go test -count=1 ./internal/harness/config ./internal/install ./internal/uninstall -run '^TestSharedManagedDelivery' -v`, aprobado, en `/tmp/opsx04-main-integration-triangulate.log`.
- Suite de los seis paquetes afectados, aprobada; race de esos seis paquetes, aprobado.
- `go vet ./...`, aprobado.
- `openspec validate codex-native-opsx --strict`, aprobado.
- Sin conflictos pendientes en el índice; AGENTS/CLAUDE idénticos. `git diff origin/main --check` aprobado para el diff de este PR sobre la base incorporada. El chequeo del merge completo contra el HEAD anterior detecta únicamente una línea vacía al final de `openspec/changes/archive/2026-11-13-claude-native-opsx/investigacion/pruebas-claude-real.md:56`, heredada de main. No se editó ese histórico de 03 ni se presenta el chequeo completo del índice como verde.
- Suite global: `go test -json -count=1 ./... -skip '^TestWindowsBundleThemeContainsMinimalInstallerPages$'`, **2053 eventos de casos/subcasos aprobados, 6 fallidos, 2 omitidos; 36 paquetes aprobados, 2 fallidos**. No se excluyeron silenciosamente los otros fallos.

Los seis eventos fallidos corresponden al test del error HTTP 404 del descargador y a tres tests principales de la huella congelada de fixtures de revisión, con sus subtests. Se compararon conjuntos de tests fallidos y mensajes contra un checkout limpio del main incorporado `7a600895`: ambos coinciden exactamente. La nueva evidencia del baseline está en `/tmp/opsx04-main-7a60089-areas.jsonl` (SHA-256 `197a8c6184ba09ae936d6eb188d9f86ddf6a6cee8438eb35b0d053e8c8d2f487`). El test WindowsTheme se excluyó expresamente por su fallo preexistente fuera de Windows; no se declara suite global verde.

### Huellas de logs locales

| Log | SHA-256 |
|---|---|
| `/tmp/opsx04-main-integration-baseline.log` | `8416a5d22d05dc1d726a8b1a93a3c879afa0f491907787c9c1cc80bf466c1707` |
| `/tmp/opsx04-main-integration-red.log` | `ed5ab1f960e1ef8dc06099093676ac1c1b64739467aed9c220e7007627623f49` |
| `/tmp/opsx04-main-integration-triangulate.log` | `1cc765a534dcdbc094d773d56e69682d42ffdc080463e5e9705f9b2dae9ec5ee` |
| `/tmp/opsx04-main-integration-race.log` | `8d0effed2e020ca5c6a299f1f142022b2fc7c485aa67ed684c63a5f5fbf8c748` |
| `/tmp/opsx04-main-integration-full.jsonl` | `b14f618e238747b50c3bb1afa5e98cd4a3dbe996268789464056eba9bd57cefe` |
| `/tmp/opsx04-main-integration-vet.log` | `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` |
| `/tmp/opsx04-main-integration-openspec.log` | `ca8f686f3ef8538c5197b7759547c2514b4a4e12e2d91e687899324f2f3dae14` |

Estos logs temporales no son adjuntos permanentes del PR. Los tests y comandos quedan versionados o documentados para repetirlos. La suite global y race anteriores corresponden a la revisión previa al test adicional de rollback combinado solicitado al revisor independiente; ese resultado se registra aparte y no se atribuye retroactivamente a esta corrida.

## Comprobación independiente y rollback combinado posterior

La revisión independiente está en [revisión de integración PR #8/04](revision-integracion-pr8-04-2026-11-13.md). Acredita la conservación del runtime específico de Claude, reejecuta los escenarios dirigidos y agrega `internal/install/mixed_rollback_integration_test.go` como triangulación, sin cambios adicionales de producción.

El test nuevo pasa en ambos órdenes con Claude y Codex seleccionados: una verificación tardía falla después de editar las instrucciones Codex y añadir una nota; el rollback conserva ambos contenidos byte a byte y comunica el conflicto, restaura exactamente las instrucciones previas de Claude y retira las skills nuevas sin perder la nota Codex. Se revisan por separado `Apply.Err` y `Rollback.Err`; no se atribuye un RED nuevo del dispatcher a la corrección inicial de una expectativa del test.

El revisor registró **3 eventos PASS** del test y sus dos subtests en `/tmp/opsx04-independent-mixed-rollback-final.jsonl`, SHA-256 `6c440a601c1464181a1c02b64d300107c3dbc98407bc3754e796a57e23ef8338`. Después se ejecutó además `go test -race -count=1 ./internal/install -run '^TestSharedManagedDeliveryRollbackPreservesConcurrentCodexEdits$' -v`, aprobado en `/tmp/opsx04-main-integration-mixed-race.log`, SHA-256 `0e424966273224bbf62445dc0e17a0195275fdadbe42f696453064f035f6451b`. Estas corridas cubren el test añadido después de la suite global, sin modificar su conteo histórico de 2053 aprobados.

También pasó nuevamente la sonda local de descubrimiento del paquete instalado en Codex CLI macOS, registrada por el revisor. Continúa siendo evidencia de descubrimiento, no de activación, especialistas ni aceptación Windows/Linux o extensión IDE. El coordinador revisó el diff combinado y el test adicional; se mantiene pendiente la revisión humana del árbol integrado por ambos responsables.

## Límites y aceptación pendiente

La integración técnica elimina la colisión de los dos despachos; no acredita CLI o IDE reales, Windows/Linux, especialistas Codex o contratos de 05/06. La aceptación y revisión conjunta del árbol integrado por los responsables 03/04 sigue en 4.6; se mantuvo el alcance ALTO ya aprobado por Matías y su autorización para conciliar la base. La fila 28 y evidencia directa de 06 siguen en 1.1. El resto de condiciones de aceptación mantiene el conteo **11/21** y el PR en borrador. No se archivó, aceptó deuda, ejecutó build ni realizó commit o push durante este trabajo delegado.
