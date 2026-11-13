# Revisión independiente del consumo Codex de 05 y 06

13/11/2026. Se revisó la entrega nueva de 04 desde un agente distinto del implementador. Alcance: textos y composición Codex, política explícita, lifecycle y límites de evidencia. No se implementó 03, no se modificó código de producción desde esta revisión y no se ejecutaron build de producto, commit ni push.

## Revisiones y alcance

- Base de 04 al iniciar: `91fbb70950babe3b694cfa1ddee141fd959acb40`, más el diff local de módulos. Los resultados se refieren al árbol con ese diff, no al commit base sin modificar.
- 05: `e4ec0f54536760f7fca6692e27bafcb8cd380df1`, `entrega-03-04.md`, §2–3. Se extrajeron los cuatro bloques literales para comparar sus bytes con los assets.
- 06: `f62cf00c09dc8a3aa96ddb4648801250e1a67a8c`, design D2–D5 y respuesta a 04/05. D2–D4 aportan normas corregidas; D5 y sus entregas literales siguen pendientes.
- El diff de esta entrega no cambia adapters/assets Claude, `ComposeLayers` ni `layered_install.go`. Un registro común propuesto por 05 no se impone como cambio de 03.

## Resultado de la revisión

No se encontraron bloqueantes nuevos en la entrega técnica inspeccionada. La conclusión acredita distribución y preservación de archivos; no acredita que el modelo siga instrucciones ni que exista captura independiente operativa.

| Condición | Evidencia y resultado |
|---|---|
| Cuatro módulos de 05, sin toggle propio | Presentes con toggles vacíos y completos. Comparación independiente de bytes contra las secciones originales: coincidencia exacta, incluida nueva línea final. |
| Apoyos, no skills adicionales | Los módulos empiezan con `##`, sin frontmatter de skill. La sonda real de descubrimiento encuentra la skill instalada y no sus módulos como skills propias. |
| Tabla de acciones | Se conservan las cinco acciones, delegación e inicio/cierre. Se adaptan rutas relativas Codex y toggles; no quedan paths Claude ni `model-assignments.md`. |
| Ausente/vacío | La tabla trata cero bytes o solo whitespace como ausencia. La verificación estructural real rechaza esos estados y archivo faltante para cada módulo; una reinstalación no adopta silenciosamente el archivo alterado. |
| Fallo limitado a la acción afectada | El texto exige parar la acción dependiente y permite terminar las partes no afectadas. La falta opcional continúa con aviso; sin delegación no se delega. Esto es contrato entregado, no prevención de runtime demostrada. |
| Reanudación | La tabla exige reinvocar la skill y releer módulos tras compactación/reanudación. No se ejecutó ese escenario en un cliente real durante esta revisión. |
| Reglas de 06 | D2–D4 se consumen inline, con pin y etiqueta explícita de requisitos parciales. No se fabrica `modules/evidence.md`, esquema, mecanismo de huella de árbol ni validador. |
| Ejecución directa/delegada | El texto mantiene los mismos criterios, separa registro de instalación de ejecución y no elimina evidencia al apagar delegación. |
| Sin falso verificado | Huella coherente sin corroboración, cero recibos, suficiencia desconocida o detalle redactado necesario mantienen el estado sin verificar. La ausencia de contrato/captura/revisor obligatorio detiene solo la acción dependiente. |
| Herramientas | El módulo literal distingue una recomendación de un bloqueo del cliente. No se implementó un nuevo control preventivo ni se presenta el prompt como aislamiento efectivo. |
| Policy | La política instalada sigue declarando explícitamente `allow_implicit_invocation: false`; pruebas de YAML y descubrimiento aprobadas. No equivale a probar activación implícita/manual por interfaz. |
| Lifecycle nuevo | Pruebas de instalación, dos cambios de toggles, verificación y reinstalación idempotente conservan primer snapshot y nota ajena; el retiro elimina solo módulos propios y mantiene reglas del usuario. |
| Actualización desde paquete anterior real | Sonda adicional usa el instalador auténtico de `91fbb70`, anterior a 05, para generar ownership y snapshot. El código nuevo agrega e inventaría los cuatro módulos, conserva identidad/archive original, reinstala sin reescritura y retira módulos propios preservando nota. Otro perfil comprueba restauración posterior desde manifiesto recargado. |

### Identidad de los cuatro módulos

| Módulo | SHA-256 de los bytes fuente e instalados |
|---|---|
| `directed-code-reading` | `db1edbe36ee928b1dc70c095b2df7d55b35b6d19fcd2ccb557580aa821321cbf` |
| `bounded-command-output` | `fcc5b089881d48ad6e466d250bb9825aa6e5bc81f06b8d98f9e9e5290494c0e9` |
| `task-tool-selection` | `052215169c55cadc2310c1561118b8b6c0f7f818f3213b30ae82ff6181c715bd` |
| `concise-agent-reporting` | `713cf2c124658624c31788cd8e06d9e2b3aae40d2a527be058b7c34e7bdf71f7` |

## Ejecuciones independientes

| Comando | Resultado |
|---|---|
| `go test -count=1 -json ./internal/harness/config ./internal/install ./internal/uninstall ./internal/verify` | Exit 0: 486 eventos PASS, 2 skips opt-in. Config 222, install 131, uninstall 70, verify 63. Corte previo a la nueva triangulación de lifecycle/ausencia/vacío del implementador. |
| `go test -count=1 -json ./internal/harness/config ./internal/verify -run '^TestCodexContextModules'` | Exit 0: 23 eventos PASS tras añadir esa triangulación: config 10 y verify 13. Incluye los cuatro módulos por ausencia, vacío y whitespace, y rechazo de adopción. No se suman a la corrida anterior como casos únicos. |
| `go vet ./internal/harness/config ./internal/install ./internal/uninstall ./internal/verify` | Exit 0. |
| `ACTIVE_STACK_CODEX_PROBE=1 go test ./internal/harness/config -run TestManagedSkillCodexDiscovery -count=1 -v` | PASS. Codex CLI `0.167.0-alpha.3.2`, macOS Intel. Fixture temporal, sin turno de modelo. |
| `ACTIVE_STACK_CODEX_PRE05_UPGRADE_PROBE=1 go test ./internal/harness/config -run '^TestCodexPre05PackageUpgrade$' -count=1 -v` | PASS en 2,22 s. Dos perfiles temporales sembrados mediante el instalador viejo real; ningún cambio en HOME ni en configuración real. |
| Extracción literal de fences de 05 con `git show e4ec0f5:…/entrega-03-04.md` y comparación de bytes | MATCH para los cuatro assets. |
| `git diff --check` y `git diff origin/main --check` | Exit 0 para los cambios de 04 inspeccionados. |

Las 32 combinaciones de toggles están cubiertas por `TestComposeCodexPackageAllToggleCombinations`, incluido en la suite de config: referencias resueltas, toggles, orden y duplicados deterministas. El test de metadata parsea YAML y exige el valor false explícito. Las aserciones de texto de action table y requisitos parciales validan lo que entrega el compositor; no se cuentan como pruebas de comportamiento del agente.

Después del corte independiente de 486 PASS se añadieron cuatro líneas normativas de D4 al `SKILL.md`: verificado exige los tres controles y al menos un recibo por requisito; un estado probado corroborado diferente del actual es viejo y se rechaza; copia ausente o vencimiento sin resolver nunca producen verificado. Se leyeron esas líneas y se comprobó su correspondencia con 06. La huella final de SKILL figura debajo. La suite global y race finales del implementador incluyen ese texto posterior; son evidencia separada y no se atribuyen a esta reejecución independiente. No se repitieron sondas ni suites por ese ajuste documental.

## Reproducción de la actualización pre-05

Se añadió únicamente el test opt-in `internal/harness/config/codex_pre05_upgrade_probe_test.go`. El helper de semilla está incluido como constante Go en ese archivo, de modo que puede reproducirse junto con el test.

1. El test crea un worktree detached temporal en `91fbb70950babe3b694cfa1ddee141fd959acb40` mediante `git worktree add --detach`.
2. Escribe allí un test temporal único `pre05_fixture_export_test.go`. Ese helper invoca el `InstallManagedSkill` de aquella revisión con los cinco toggles y reglas de usuario sintéticas. La ruta se pasa mediante `ACTIVE_STACK_CODEX_OLD_SEED_HOME`; no se reemplaza HOME, CODEX_HOME ni el perfil real.
3. Ejecuta `go test ./internal/harness/config -run '^TestExportPre05Fixture$' -count=1` contra esa fuente vieja. No se construye ni ejecuta el binario del producto.
4. La revisión actual actualiza el perfil producido y prueba ownership SHA-256 de los cuatro nuevos módulos, actualización posterior, reinstalación, identidad/archivo del primer snapshot y retiro con nota propia.
5. Un segundo perfil sembrado de la misma forma se actualiza y restaura con `backup.ReadManifest` y `RestoreService.Restore`. Se recuperan exactamente las instrucciones previas y desaparecen los módulos nuevos.
6. Se elimina el worktree temporal y los perfiles al terminar. `git worktree list` no conserva una entrada nueva de esta sonda.

No se fabricó un inventario antiguo ni se añadió una interfaz de producción para simularlo. Esta prueba es triangulación de compatibilidad; no se presenta como un nuevo RED de producción. El RED de composición 05 y requisitos parciales 06 pertenece al implementador y se registra en su evidencia.

### TDD Cycle Evidence

| Trabajo | Test | Safety net | RED | GREEN | Triangulación / refactor |
|---|---|---|---|---|---|
| Revisión de composición y contrato entregado | `codex_context_modules_test.go` y suites existentes | Suites independientes verdes | RED ejecutado por implementador, no por esta revisión | 486 PASS / 2 skips en el corte indicado | Fuente literal, 32 toggles, policy y referencias; sin afirmar conducta real |
| Lifecycle/ausencia/vacío | `codex_context_lifecycle_test.go`, verify `codex_context_modules_test.go` | Corte previo verde | Triangulación del contrato existente | 23 PASS dirigidos | Actualizaciones, notas y 12 casos de defecto instalado |
| Actualización desde versión anterior auténtica | `codex_pre05_upgrade_probe_test.go` | Suite de config verde | Sin cambio nuevo de producción | PASS opt-in | Dos perfiles, ownership real, snapshot original, idempotencia, retiro y restauración posterior |

## Logs e identidad del código

Los logs son artefactos locales. Sus hashes identifican esos archivos y no sustituyen el contrato completo de recibos pendiente de 06.

| Archivo | SHA-256 |
|---|---|
| `/tmp/opsx04-modules-independent.jsonl` | `e24b8c251e9b116e676fae2e033ea5a05a622184bfb88bea9810a6d615c73c32` |
| `/tmp/opsx04-modules-independent-vet.log` | `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` |
| `/tmp/opsx04-modules-lifecycle-independent.jsonl` | `40b54b5059bc0c483424d69dc81c80024ec914c6aa1fb3ed2cad3ed7ecc14fde` |
| `/tmp/opsx04-modules-discovery.log` | `2858a01fd7ecdd4054b4d5d166140e4dae12c3a2b5df97cbc751dd6c2f44dba4` |
| `/tmp/opsx04-pre05-upgrade-probe.log` | `f11036589fbff1030a26b00a738e631b5657f3b730102496f7341300d86ffca0` |
| `internal/harness/config/codex_package.go` | `2d8f7d19055ed030d5fdf746c15ba40d4583af42ad94bc70dfa73255d2749215` |
| `assets/codex/native/SKILL.md` bajo config, texto final con aclaración D4 | `f6b27147179e4ced4afb0557e7af82ed37aeaafcf362b2e2fb2da3b0bd07f51c` |
| `assets/codex/native/action-table.md` bajo config | `5c9b940aa4089e377916556a0865220d791bf3f1abad4a239b96b117f77092d9` |
| `internal/harness/config/codex_pre05_upgrade_probe_test.go` | `5a49ec5f8d0013f5c93350f2dd5ea69e8eaf23ac88897a4cbbaaecbc1ce840a3` |

## Límites que permanecen

La deuda aceptada por Matías en esta fase se limita a las dependencias externas 05/06/07 identificadas por sus responsables. No convierte automáticamente P2, especialistas, CLI/IDE Windows/Linux, comparación con 02 o revisión humana de migración en deuda aceptada ni en pruebas realizadas.

La entrega parcial de 06 no es el contrato completo: conserva el bloqueo por acción cuando falta el contrato operativo, corroboración o revisor obligatorio. Siguen pendientes el texto/extracto definitivo, captura independiente por celda, huella validada del estado probado y los acuerdos de ubicación/retención/estados/sobre. Ninguna presencia de módulos, policy parseada, hash coherente o health check estructural permite llamar verificada a una ejecución de tarea.

No se marcaron tareas OpenSpec desde esta revisión. La promoción del PR para revisión y su eventual aprobación de merge deben mostrar esos límites y los escenarios propios de 04 que continúan sin aceptación real.
