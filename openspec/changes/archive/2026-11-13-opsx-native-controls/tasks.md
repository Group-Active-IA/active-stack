# Tasks

> Estado al archivar: ver «Alcance de 08 al archivar y deuda aceptada» más abajo (6 unidades entregadas, 38 en deuda). Esta ronda es solo de propuesta y de cierre: no implementa nada y solo toca `openspec/changes/opsx-native-controls/` y su archivo. 08 es de gobernanza **ALTO**: la implementación espera la revisión humana de Neyén. Flujo del equipo: **contrato → evidencia → decisión**: se corrige la propuesta, se implementa con `apply`, se verifican los criterios de aceptación y se abre el PR. Este change es de gobernanza **ALTO**: antes de implementar hace falta la revisión humana que exige `CLAUDE.md` (revisa e integra Neyén). Cuando llegue la etapa de implementación, TDD estricto, con fakes para el snapshotter y la lectura de archivos, y **nunca** el entorno real del usuario: RED, GREEN, TRIANGULATE y REFACTOR (esquema tomado de los changes archivados, no de una regla escrita en `CHANGES.md`). Antes de tocar archivos existentes, correr el SAFETY NET del paquete y anotar la línea base. No hacer build ni commitear salvo pedido explícito. Si surge la necesidad de tocar `filemerge`, `backup` o `pipeline`, frenar y reportar para revisión. Base revisada: `origin/main` `0005a6f` (01, 02, 03, 04, 05 y 06 integrados). Revisión consultada (2026-11-13): 04 en `main` `0005a6f` (PR #9; `openspec/changes/archive/2026-11-13-codex-native-opsx/`, `openspec/specs/codex-opsx-delivery/spec.md:64-78` y `informes/traspaso-lifecycle-2026-11-13.md`); 06 en `main` `0005a6f` (PR #12 mergeado, con el fix `d3ab8b3`; ruta archivada `openspec/changes/archive/2026-11-13-opsx-verifiable-evidence/` y spec sincronizada `openspec/specs/execution-receipts/spec.md`): D2, D3 y D9 del `design.md` y §6 de `contrato-recibo-de-ejecucion.md`; 07 en `origin/propose/opsx-selective-delegation` `4922498`: D4 del `design.md` archivado; 05 en `main` `0005a6f` (PR #10 mergeado; ruta archivada `openspec/changes/archive/2026-11-13-opsx-task-context/` y specs sincronizadas, entre ellas `openspec/specs/bounded-command-output/spec.md`): D8, D14 y D18 del `design.md` archivado.

## 1. Antes de empezar: acuerdos y reproducciones (ALTO)

- [x] 1.1 Leer el contrato de 06 (ya en `main`: PR #12 mergeado, `0005a6f`, con `d3ab8b3`; D2 campos del recibo de ejecución, D3 tres controles, D9 hooks y §6 de `contrato-recibo-de-ejecucion.md`) y confirmar con su revisor (Juan Cruz), sobre el PR #12, que no cambia antes de integrarlo en 6.1. No se redefine nada de 06. *Hecho: el PR #12 está mergeado en `main` (`0005a6f`, merge de `d3ab8b3`); el contrato y la spec `openspec/specs/execution-receipts/spec.md` se citan desde `main`. El fix `d3ab8b3` cambió qué cuenta como fuente independiente: el mismo arnés puede capturar y corroborar, y una fuente vacía o `agente` no cuenta (`eval/evidence/judge.go:246-249`); 08 no lo redefine (design D2).*
- [x] 1.2a Registrar como hechos de `main` la política de Estricto de Codex (`sandbox_mode = "read-only"`, `approval_policy = "on-request"`, `install.go:120-126`) y el método de aceptación de 01 (perfil aislado, cliente real, lectura permitida y escritura con escalación denegada, Windows CLI 0.166.0 primero), y registrar los resultados una sola vez con 01 (D11). *Hecho: la política y el método quedan registrados como hechos de `main` en design D11. Pendiente: registrar resultados medidos con 01 (no se corrió ninguna celda de Codex).*
- [ ] 1.3 Cerrar en el grupo las preguntas abiertas del design que dependan solo de 08 y pasar por la revisión humana que exige el nivel ALTO (Neyén). Las que dependen de 03, 04, 05, 06 o 07 siguen (TBD).
- [ ] 1.4 Reproducir en un `HOME` temporal, con el binario y un `settings.json` de prueba, el caso del 02/11: `defaultMode` en `"auto"` y un permiso propio en `allow`. Registrar qué cambia al instalar con cada tier (sin tocar el entorno real). El resultado decide si D9 pasa a requisito de implementación y qué se arregla en `filemerge`.
- [ ] 1.5 Con 03 (Juani): confirmar que D12 sigue vigente (activación por `/sdd-orchestrator`, `agent` sin escribir) y acordar con el revisor ALTO la integración serializada de los arreglos de `WriteFileAtomic` (03 G6, alcance chico, sin cambios en `filemerge` global) y de `MergeJSONObjects` (D9).

## 2. Inventario (BAJO, solo lectura)

- [ ] 2.1 Listar los mecanismos de permisos y hooks de cada cliente y versión soportada. Claude: `onFailure: "block"` desde v2.1.295 para hooks `command`/`http`. Codex: confirmar desde qué versión hay hooks y en qué interfaces (la documentación no lo dice, TBD), dónde se configuran (`hooks.json` o `[hooks]`) y el modelo de confianza por hash.
- [ ] 2.2 Listar las superficies declaradas de cada cliente y los shells reales por sistema operativo.
- [ ] 2.3 Clasificar las acciones en "debe bloquearse" y "se valida después", partiendo de lo que hoy cubre `permissions` (el piso `deny`, `overlays.go:16-28`), y clasificar cada regla en orienta / restringe / verifica.
- [ ] 2.4 Armar la matriz cliente × superficie × herramienta × shell con el estado de cada celda (cubierta, no cubierta, no probada) en el formato de la tabla de cobertura del design D13.
- [ ] 2.5 Armar el inventario de escrituras de D13 (archivo, función de `filemerge`, nivel) y marcar las ALTO.
- [ ] 2.6 Verificar con el cliente real lo que la documentación no afirma y D6 deja (TBD): si un `PreToolUse` de Claude bloquea bajo `bypassPermissions`, el efecto de `onFailure` en `mcp_tool`/`prompt`/`agent`, el error y el timeout de un hook local de Codex (incluido `PreToolUse`; la documentación solo afirma la salida no soportada o inválida), el comportamiento de los demás eventos de Codex ante fallo o timeout, la tabla de códigos de salida por evento de Claude, y releer textual `disableAllHooks` y `allowManagedHooksOnly`.

## 3. Diseño del piloto (ALTO)

- [ ] 3.1 Definir la política de fallo por tipo de acción, partiendo de D6 (TBD del grupo).
- [ ] 3.2 Definir el contrato de entrada y salida de los handlers y la opción de handler a pilotear (D5).
- [ ] 3.3 Definir la capacidad opcional de hooks en los adapters y el mecanismo de ownership (D4).
- [ ] 3.4 Definir cómo se integran los controles a `update` y `uninstall`: remoción acotada del overlay de permisos y de las claves de hooks (hoy el paso de `permissions` de `uninstall` no hace nada, `internal/uninstall/steps.go:362-384`), y registro de cada archivo escrito con `Manifest.GuardRollback` (reutilizar, no reimplementar; D8). Definir la actualización desatendida (D14).
- [ ] 3.5 Definir cómo se mezclan listas y escalares del usuario y cómo se aborta ante un archivo ilegible (D9), y quién revisa el cambio en `filemerge` (D15).
- [ ] 3.6 Extender la vista de `install --dry-run` con el diff de permisos y hooks.
- [ ] 3.7 Definir el contenido del mensaje de control ausente y qué cuenta como protección equivalente verificada, alineado con 03 D11/D12 y con la regla de 04 (`openspec/specs/codex-opsx-delivery/spec.md:64-78`) sin redefinirlas, y la lista de controles obligatorios y opcionales (TBD, grupo).
- [ ] 3.8 Completar la tabla D14 (dueño, reversión, precedencia, Bypass, actualización desatendida, restauración) para cada entrada que se instale.

## 4. Pruebas previstas (ALTO)

- [ ] 4.1 Acción permitida y acción denegada en cada cliente. En Codex, con el método de 01 (D11).
- [ ] 4.2 Claude: instalar y actualizar permisos sobre un `settings.json` de prueba con la skill `/sdd-orchestrator` de 03 instalada y un `agent` fijado por el usuario; la clave `agent` y los hooks del usuario quedan idénticos (D12).
- [ ] 4.3 Timeout, JSON inválido y handler ausente, con el comportamiento real de cada cliente registrado y contrastado con la documentación de D6; en Codex el error y el timeout de un hook local son (TBD) y se miden, sin asumir que fallan abierto.
- [ ] 4.4 Rutas alternativas en shells reales por sistema operativo.
- [ ] 4.5 **Archivo ilegible:** con un `settings.json` malformado en un `HOME` temporal, la escritura aborta, el archivo queda byte-idéntico y se informa ruta y motivo (hoy `MergeJSONObjects` lo trata como `{}`, `json_merge.go:10-15`: la prueba debe fallar primero por esa razón). Lo mismo con un `config.toml` inválido.
- [ ] 4.6 Escritura con revisión, backup, fallo de backup, preservación de entradas ajenas (lista `allow` del usuario, `defaultMode` fijado por el usuario, hooks del usuario), reinstalación idempotente, `update`, `uninstall`, reversión y reversión con ediciones posteriores del usuario (el archivo actual se conserva).
- [ ] 4.7 **Control obligatorio:** (a) obligatorio presente, la acción procede; (b) obligatorio ausente, se detiene solo esa acción y el mensaje nombra la protección y cómo restablecerla; (c) opcional ausente, advertencia y la acción continúa.
- [ ] 4.8 **CRLF y modos:** medir en Windows y Linux la conversión de CRLF a LF de `UpsertTOMLTopLevel` (`toml.go:29`), el `chmod 0755` del directorio padre (`writer.go:54-58`) y el `0644` del archivo que `permissions` pasa a `WriteFileAtomic` (`install.go:143,185`; `writer.go:78`) y la pérdida de comentarios de `MergeJSONObjects`.
- [ ] 4.9 Matriz de plataformas con Windows y Linux como mínimo: distribuidas (`.github/workflows/release.yml:71-76`), declaradas (`CLAUDE.md:13`) y probadas (lo que se corrió).
- [ ] 4.10 Pruebas por superficie, por separado.
- [ ] 4.11 Suite Go completa en verde y revisión de qué paquetes ALTO cambiaron (`git diff --stat`).

## 5. Documentación

- [ ] 5.1 Tabla de cobertura y límites por cliente, versión, shell y superficie, en el formato de D13.
- [ ] 5.2 Guía breve de qué queda garantizado por control nativo y qué solo por instrucción.
- [ ] 5.3 Glosario corto: hook del instalador vs. hook del cliente, y recibo de ejecución (06) vs. registro de instalación (04).

## 6. Integración con otras propuestas

- [ ] 6.1 Con 06: apuntar al contrato de 06 sin redefinirlo. Un control posterior valida de forma mecánica los campos del recibo de ejecución de 06 (D2: `corrida`, `arbol`, `comando`, `resultado`, `requisitos`, `salida`) y la integridad de la copia redactada; la procedencia (fuente independiente o reejecución) y la suficiencia son de 06 (D3). Con `d3ab8b3` en `main`, el mismo arnés puede capturar y corroborar, y una fuente vacía o `agente` no cuenta; 08 no lo redefine. Los hooks complementan la evidencia pero no son un registro exhaustivo, y la ausencia de un rastro no verifica ni rechaza un recibo (06 D9, §6 de `contrato-recibo-de-ejecucion.md` y requisito "Hooks complement evidence" de `specs/execution-receipts/spec.md`). La contraparte en 06 es su tarea 5.3 (D06-0503, abierta como deuda aceptada en su `tasks.md` archivado).
- [ ] 6.2 Con 01: casos de permisos efectivos compartidos y registro único de resultados (D11).
- [ ] 6.3 Con 03: punto de contacto en `WriteFileAtomic` (G6) y clave `agent` (D12); el mensaje de control ausente que 03 D12 le asigna a 08.
- [ ] 6.4 Con 04 (Matías): que su regla de controles y módulos faltantes (`openspec/specs/codex-opsx-delivery/spec.md:64-78`) use el mensaje de D7, y qué hooks de Codex entran al piloto.
- [ ] 6.5 Con 10: línea de base de cobertura para comparar con un juez en sombra, con la tabla de D13.
- [ ] 6.6 Con 07 (Neyén) y 05 (Emiliano), pedidos recibidos que 08 no decidió (TBD, grupo): 07 pide sumar a la tarea 2.3 "escritura fuera del conjunto de escritura del encargo" como candidata a "debe bloquearse" (07 D4, último párrafo; hoy 07 la declara no controlada preventivamente); 05 pide a 01 u 08 un piso equivalente al de Claude en Codex (05 D14) y menciona un hook `PreToolUse` de captura como candidato a una decisión separada, con 08 como dueña de los hooks (05 D8).

## 7. Cierre del change (flujo: entrega, comprobación, pendientes, traspaso, PR)

- [x] 7.1 **Qué entrega.** Listar lo entregado (documentación, matriz, tabla de cobertura, código y pruebas si los hubo) con rutas. **Declarar alcance parcial** si algo depende de 06 o de una decisión (TBD) abierta (04, 05 y 06 ya están en `main`; 07 sigue en su rama). *Hecho, con alcance parcial: se entregan `proposal.md`, `design.md`, `tasks.md` y la spec `native-controls-pilot` (propuesta, sin implementación) más `informes/`. No se entrega ni se mide nada del piloto.*
- [x] 7.2a **Cómo se comprobó.** Para cada criterio de aceptación de la ficha (permitido, denegado, timeout, JSON inválido, handler ausente, rutas alternativas, shells reales, límites por superficie, revisión antes de escribir), registrar comando, resultado y plataforma (Windows y Linux). Comprobar los artefactos con `openspec status --change opsx-native-controls` y `openspec validate opsx-native-controls`. *Hecho solo para la propuesta: `informes/verificacion-08-2026-11-13.md` (validación strict, `git diff --stat`, citas contra `0005a6f`). Pendiente: los criterios de aceptación de la ficha, que requieren el piloto.*
- [x] 7.3 **Qué queda pendiente.** Listar los `(TBD)` abiertos (política ante fallo, handlers, ownership, Bypass, hooks de Codex por versión, controles obligatorios, actualización desatendida, remoción acotada en `uninstall`, mínimo de cobertura) y las dependencias de 03, 04, 06 y 10. *Hecho: `informes/deuda-aceptada-08-2026-11-13.md`.*
- [ ] 7.4 **Informe de traspaso.** Escribir `openspec/changes/opsx-native-controls/informes/traspaso-<fecha>.md` con el modelo de `openspec/changes/archive/2026-11-13-codex-native-opsx/informes/traspaso-lifecycle-2026-11-13.md` (en `main`): qué podrá usar el siguiente responsable y coordinación pendiente por change (03, 04, 06, 10).
- [x] 7.5a **Archivar y abrir el PR.** El PR se abre **después de archivar el change, con todas las tareas completas y verificadas** (o con el alcance parcial declarado en 7.1). El revisor e integrador es **Neyén**, que comprueba **"contrato y evidencia sobre el último commit"**; los hallazgos se resuelven en la misma rama. La rama local es `opsx-native-controls`; al publicarla será `propose/opsx-native-controls` (`CHANGES.md:10`). La rama vieja `opsx-native-controls` del remoto (`f3ba5d3`) no se toca. *Hecho: archivado con `openspec archive` en `openspec/changes/archive/2026-11-13-opsx-native-controls/` (alcance parcial declarado en 7.1) y push sin force de `opsx-native-controls` a `propose/opsx-native-controls`. Pendiente: abrir el PR (`gh` no está instalado; el cuerpo está preparado fuera del repo) y que Neyén lo revise como ALTO.*

## Alcance de 08 al archivar y deuda aceptada

08 se archiva con **alcance parcial** (13/11), igual que 03, 04 y 06: lo que depende de implementación, de mediciones del piloto o de acuerdos del grupo pasa a deuda aceptada y no se marca como hecho. Las unidades con sufijo `a` son subalcances: completar `1.2a` no significa completar `1.2` entero. El detalle por tarea, con dueño, qué falta, criterio de cierre y comportamiento mientras tanto, está en el [registro de deuda](informes/deuda-aceptada-08-2026-11-13.md); la verificación de lo hecho, con comandos y salidas reales, en [verificacion-08-2026-11-13.md](informes/verificacion-08-2026-11-13.md).

**44 unidades: 6 entregadas y 38 pendientes (deuda aceptada).** Entregadas: 1.1, 1.2a, 7.1, 7.2a, 7.3 y 7.5a. Pendientes: 1.2 (parte), 1.3 a 1.5, 2.1 a 2.6, 3.1 a 3.8, 4.1 a 4.11, 5.1 a 5.3, 6.1 a 6.6 y 7.4. Los sufijos `a` no cambian el denominador.

- **Nada del piloto se corrió.** No hay matriz de cobertura medida, ni pruebas en Windows y Linux, ni escrituras de permisos o hooks, ni reproducción del caso del 02/11.
- **Spec `native-controls-pilot`:** al archivar se sincroniza a `openspec/specs/native-controls-pilot/spec.md`, pero las decisiones que lleva son **propuesta de 08, no acuerdos del grupo**.
- **Gobernanza ALTO:** ninguna implementación empieza sin la revisión humana de Neyén; tocar `filemerge`, `backup` o `pipeline` requiere ese revisor.
- No se tocó ningún `.go`, ni `filemerge`, `backup` ni `pipeline`.

## Evidencia de las unidades entregadas

| Unidades | Evidencia |
|---|---|
| 1.1 | PR #12 mergeado en `main` (`0005a6f`); `openspec/specs/execution-receipts/spec.md`; `eval/evidence/judge.go:246-249`. |
| 1.2a | design D11 (política de Estricto de Codex y método de aceptación de 01 como hechos de `main`). |
| 7.1, 7.2a, 7.3, 7.5a | `informes/verificacion-08-2026-11-13.md`, `informes/deuda-aceptada-08-2026-11-13.md`, carpeta de archivo. |

## DEUDA aceptada

Las 38 unidades pendientes más la parte pendiente de 1.2a, 7.2a y 7.5a quedan como deuda aceptada de Gianfranco el 2026-11-13 para cerrar el circuito, igual que 03, 04 y 06. Cada una tiene su fila (D08-xxxx) con dueño o proveedor, qué falta, criterio de cierre y comportamiento mientras tanto en [informes/deuda-aceptada-08-2026-11-13.md](informes/deuda-aceptada-08-2026-11-13.md). Nada de eso se marca como hecho.
