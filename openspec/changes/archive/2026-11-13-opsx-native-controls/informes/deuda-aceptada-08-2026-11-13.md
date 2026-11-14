# Deuda aceptada de 08 al archivar `opsx-native-controls`

13 de noviembre de 2026. Responsable de 08: Gianfranco; revisor e integrador: Neyén (revisor ALTO). Sigue la convención de 06 (`openspec/changes/archive/2026-11-13-opsx-verifiable-evidence/informes/deuda-aceptada-06-2026-11-13.md`), 03 y 04: lo que depende de implementación, de mediciones del piloto o de acuerdos del grupo sale del checklist activo, queda registrado con su criterio de cierre y **no se marca como hecho**.

Base: `origin/main` `0005a6f` (01 a 06 integrados). 07 se cita en `origin/propose/opsx-selective-delegation` `4922498` (no mergeada), por sección. Verificación de lo entregado: `verificacion-08-2026-11-13.md`.

## Qué significa «propuesta de 08»

Al archivar, la spec `native-controls-pilot` se sincroniza a `openspec/specs/native-controls-pilot/spec.md`. **Eso no la convierte en un acuerdo del grupo.** Las decisiones que lleva (clasificación en tres, política ante fallo, regla de control obligatorio ausente, preservación de entradas del usuario, archivo ilegible = abortar, reversión con `GuardRollback`, revisión humana de `filemerge`) son el estado de partida que 08 propone, no reglas aprobadas por 01, 03, 04, 05, 06, 07 ni por el grupo. Lo mismo vale para `design.md`, `proposal.md` y `tasks.md` archivados. **08 es de gobernanza ALTO: ninguna implementación empieza sin la revisión humana de Neyén.**

## Lo entregado (no es deuda)

1.1, 1.2a, 7.1, 7.2a, 7.3 y 7.5a (ver `tasks.md`): la propuesta (`proposal.md`, `design.md`, `tasks.md`, spec), citas re-verificadas contra `0005a6f`, la validación strict y estos informes. No se escribió ni se midió nada del piloto, y no se tocó ningún `.go`.

## Registro de deuda

**Estado de todas las filas:** deuda aceptada por Gianfranco el 2026-11-13 para cerrar el circuito. Ninguna fila se cierra por existir una propuesta o un informe. «Mientras tanto» describe qué rige hasta que se cierre la fila.

### Parte pendiente de unidades entregadas

| ID | Tarea | Dueño / proveedor | Qué falta | Criterio de cierre | Mientras tanto |
|---|---|---|---|---|---|
| D08-0102 | 1.2a | Gianfranco con 01 (Juan Cruz) | Registrar resultados medidos de las celdas de Codex en un registro único con 01 | Resultados de las celdas de Codex registrados una sola vez, con superficie, versión y estado final del archivo de prueba | La política y el método figuran como hechos de `main` en design D11; ninguna celda de Codex está probada |
| D08-0702 | 7.2a | Gianfranco | Comprobar los criterios de aceptación de la ficha (permitido, denegado, timeout, JSON inválido, handler ausente, rutas alternativas, shells reales, límites por superficie, revisión antes de escribir) en Windows y Linux | Cada criterio con comando, resultado y plataforma | Solo se verificó la propuesta; el piloto **no se corrió** |
| D08-0705 | 7.5a | Gianfranco; revisa Neyén | Abrir el PR (`gh` no está instalado) y la revisión de Neyén como ALTO | PR abierto y «contrato y evidencia sobre el último commit» comprobado por Neyén | La rama `propose/opsx-native-controls` se publica en esta ronda; el cuerpo del PR está preparado fuera del repo |

### Grupo 1: acuerdos y reproducciones

| ID | Tarea | Dueño / proveedor | Qué falta | Criterio de cierre | Mientras tanto |
|---|---|---|---|---|---|
| D08-0103 | 1.3 | Grupo; revisor ALTO Neyén | Cerrar las preguntas abiertas de 08 (ver filas D08-DEC) y pasar por la revisión humana | Acta del grupo con cada decisión y revisión de Neyén registrada | Todas figuran como (TBD) o propuesta de 08 |
| D08-0104 | 1.4 | Gianfranco | Reproducir en un `HOME` temporal el caso del 02/11 (`defaultMode` `"auto"` y un permiso propio en `allow`) con cada tier | Informe con qué cambia en cada tier, sin tocar el entorno real | El caso del 02/11 está observado, **no reproducido de forma aislada** (proposal, Evidencia); D9 sigue como propuesta |
| D08-0105 | 1.5 | Juani (03) y revisor ALTO | Confirmar que D12 sigue vigente y acordar la integración serializada de los arreglos de `WriteFileAtomic` (03 G6) y `MergeJSONObjects` (D9) | Confirmación de 03 y orden de integración acordado con Neyén | D12 y D9 se tratan como propuesta |

### Grupo 2: inventario

| ID | Tarea | Dueño / proveedor | Qué falta | Criterio de cierre | Mientras tanto |
|---|---|---|---|---|---|
| D08-0201 | 2.1 | Gianfranco | Inventario de mecanismos de permisos y hooks por cliente y versión (Codex: desde qué versión y en qué interfaces, dónde se configuran, confianza por hash) | Tabla por cliente y versión con fuente consultada | design D6 resume la documentación consultada el 2026-11-13; Codex por versión es (TBD) |
| D08-0202 | 2.2 | Gianfranco | Superficies declaradas por cliente y shells reales por sistema operativo | Lista por cliente y SO | (TBD) |
| D08-0203 | 2.3 | Gianfranco con 07 y 05 | Clasificar acciones en «debe bloquearse» y «se valida después» y cada regla en orienta, restringe o verifica | Clasificación revisada por el grupo | Solo se conoce el piso `deny` de hoy (`overlays.go:16-28`) |
| D08-0204 | 2.4 | Gianfranco | Matriz cliente × superficie × herramienta × shell con estado por celda | Matriz en el formato de D13, sin celdas «cubierta» sin corrida | **No hay matriz.** Toda acción figura como no controlada o no probada |
| D08-0205 | 2.5 | Gianfranco; revisa Neyén | Inventario de escrituras (archivo, función de `filemerge`, nivel) | Inventario con las ALTO marcadas | design D13 lo propone; sin revisión |
| D08-0206 | 2.6 | Gianfranco | Verificar con el cliente real lo que la documentación no afirma: `PreToolUse` bajo `bypassPermissions`, efecto de `onFailure` en `mcp_tool`/`prompt`/`agent`, error y timeout de un hook local de Codex, códigos de salida por evento, `disableAllHooks` y `allowManagedHooksOnly` | Cada punto medido y registrado | Todo figura como (TBD); no se asume que el control falle cerrado |

### Grupo 3: diseño del piloto

| ID | Tarea | Dueño / proveedor | Qué falta | Criterio de cierre | Mientras tanto |
|---|---|---|---|---|---|
| D08-0301 | 3.1 | Grupo | Política de fallo por tipo de acción | Política acordada | Ver D08-DEC-01 |
| D08-0302 | 3.2 | Grupo | Contrato de entrada y salida de handlers y opción de D5 | Contrato y opción acordados | Ver D08-DEC-02 |
| D08-0303 | 3.3 | Grupo y revisor ALTO | Capacidad opcional de hooks en adapters y ownership | Mecanismo acordado | Ver D08-DEC-03; los adapters no exponen hooks |
| D08-0304 | 3.4 | Gianfranco; revisa Neyén | Integración a `update` y `uninstall` (remoción acotada) y registro con `Manifest.GuardRollback` | Diseño revisado e implementado con pruebas | `uninstall` no quita el overlay de permisos (`steps.go:362-384`) |
| D08-0305 | 3.5 | Revisor ALTO | Mezcla de listas y escalares y aborto ante archivo ilegible; quién revisa `filemerge` | Decisión y cambio en `filemerge` revisado | `MergeJSONObjects` sigue tratando el JSON malformado como `{}` (`json_merge.go:10-15`) |
| D08-0306 | 3.6 | Gianfranco | Diff de permisos y hooks en `install --dry-run` | Vista implementada y probada | No existe |
| D08-0307 | 3.7 | Grupo con 03 y 04 | Mensaje de control ausente, qué cuenta como protección equivalente, lista de controles obligatorios y opcionales | Mensaje y lista acordados con 03 y 04 | Ver D08-DEC-07; rige la regla de 04 (`codex-opsx-delivery/spec.md:64-78`) |
| D08-0308 | 3.8 | Gianfranco | Tabla D14 completa para cada entrada que se instale | Tabla completa | Parcial en design D14 |

### Grupo 4: pruebas del piloto (todas pendientes, ninguna corrió)

| ID | Tarea | Dueño / proveedor | Qué falta | Criterio de cierre | Mientras tanto |
|---|---|---|---|---|---|
| D08-0401 | 4.1 | Gianfranco con 01 | Acción permitida y denegada en cada cliente (Codex con el método de 01) | Resultados por cliente y plataforma | Sin probar |
| D08-0402 | 4.2 | Gianfranco con 03 | Instalar y actualizar permisos con `/sdd-orchestrator` y un `agent` del usuario; `agent` y hooks idénticos | Prueba en `HOME` temporal en verde | Sin probar |
| D08-0403 | 4.3 | Gianfranco | Timeout, JSON inválido y handler ausente, con comportamiento real por cliente | Comportamiento registrado y contrastado con D6 | D6 es hipótesis |
| D08-0404 | 4.4 | Gianfranco | Rutas alternativas en shells reales por SO | Resultados por shell | Sin probar |
| D08-0405 | 4.5 | Gianfranco; revisa Neyén | Archivo ilegible: aborta y queda byte-idéntico (`settings.json` y `config.toml`) | Prueba que falla primero por `json_merge.go:10-15` y luego pasa | El comportamiento actual es tratarlo como `{}` |
| D08-0406 | 4.6 | Gianfranco; revisa Neyén | Escritura con revisión, backup, fallo de backup, preservación, idempotencia, `update`, `uninstall`, reversión y ediciones posteriores | Pruebas con fakes en verde | **No se escribe ninguna entrada nueva de permisos ni hooks** |
| D08-0407 | 4.7 | Gianfranco con 03 y 04 | Control obligatorio presente, ausente y opcional ausente | Los tres casos probados | Sin probar |
| D08-0408 | 4.8 | Gianfranco | Medir en Windows y Linux CRLF a LF (`toml.go:29`), `chmod 0755` (`writer.go:54-58`), `0644` (`install.go:143,185`) y pérdida de comentarios | Mediciones por SO | Riesgos sin medir |
| D08-0409 | 4.9 | Gianfranco | Matriz de plataformas (Windows y Linux como mínimo) con distribuidas, declaradas y probadas | Tabla con lo corrido | Declaradas en `CLAUDE.md:13`; no se probó ninguna |
| D08-0410 | 4.10 | Gianfranco | Pruebas por superficie, por separado | Resultados por superficie | Superficies sin probar quedan «no controladas» |
| D08-0411 | 4.11 | Gianfranco | Suite Go completa en verde y revisión de paquetes ALTO | `go test ./...` exit 0 y `git diff --stat` | Esta ronda no corrió la suite Go: no cambió código. 06 registró 4 paquetes de 04 que fallan en `main` limpio en Windows (D06-0411) |

### Grupos 5 a 7: documentación, integración y cierre

| ID | Tarea | Dueño / proveedor | Qué falta | Criterio de cierre | Mientras tanto |
|---|---|---|---|---|---|
| D08-0501 | 5.1 | Gianfranco | Tabla de cobertura y límites por cliente, versión, shell y superficie | Tabla en formato D13 con corridas | No existe |
| D08-0502 | 5.2 | Gianfranco | Guía de qué queda garantizado por control nativo y qué solo por instrucción | Guía revisada | No existe |
| D08-0503 | 5.3 | Gianfranco | Glosario hook del instalador vs. hook del cliente, recibo de ejecución vs. registro de instalación | Glosario publicado | Solo en design (Glosario) |
| D08-0601 | 6.1 | Gianfranco con 06 (Juan Cruz) | Validación mecánica de los campos del recibo de ejecución y rastro de hooks; contraparte D06-0503 | Acuerdo con 06 y prueba | 06 en `main` (PR #12): el mismo arnés puede capturar y corroborar, una fuente vacía o `agente` no cuenta (`eval/evidence/judge.go:246-249`); 08 no lo redefine. Los hooks no son registro exhaustivo |
| D08-0602 | 6.2 | Gianfranco con 01 (Juan Cruz) | Casos de permisos efectivos compartidos y registro único | Registro único | Sin hacer |
| D08-0603 | 6.3 | Gianfranco con Juani (03) | `WriteFileAtomic` (G6), clave `agent` y mensaje de control ausente | Confirmación de 03 | D12 propuesta |
| D08-0604 | 6.4 | Gianfranco con Matías (04) | Que la regla de 04 use el mensaje de D7 y qué hooks de Codex entran | Confirmación de 04 | Rige 04 tal como está |
| D08-0605 | 6.5 | Gianfranco con 10 | Línea base de cobertura para el juez en sombra | Tabla D13 con corridas entregada a 10 | Sin tabla medida; dueño de 10 (TBD) |
| D08-0606 | 6.6 | Grupo; Neyén (07) y Emiliano (05) | Pedidos de 07 (escritura fuera del conjunto del encargo como «debe bloquearse») y 05 (piso de Codex, hook `PreToolUse` de captura) | Decisión del grupo | Ver D08-DEC-12 |
| D08-0704 | 7.4 | Gianfranco | Informe de traspaso | `informes/traspaso-<fecha>.md` con el modelo de 04 | Este registro y `verificacion-08-2026-11-13.md` hacen de traspaso provisorio |

### Decisiones del grupo (preguntas abiertas del design)

| ID | Decisión | Dueño | Qué falta | Criterio de cierre | Mientras tanto |
|---|---|---|---|---|---|
| D08-DEC-01 | Política ante fallo (OQ 1) | Grupo | Decidir fallar cerrado o abierto por tipo de acción | Acta del grupo | Se documenta lo que hace cada cliente; no hay política |
| D08-DEC-02 | Alcance de handlers (OQ 2) | Grupo | Hasta qué opción de D5 se pilotea | Acta | (TBD) |
| D08-DEC-03 | Ownership de hooks (OQ 3) | Grupo y Neyén | Identificador dentro de la entrada o registro aparte | Acta | (TBD) |
| D08-DEC-04 | Bypass y hooks de Codex (OQ 4 y 5) | Gianfranco (se mide) | Medir `PreToolUse` bajo `bypassPermissions` y versiones de Codex | Mediciones de D08-0206 | (TBD) |
| D08-DEC-06 | Quién arregla `filemerge` (OQ 6) | Grupo y Neyén | Decidir el arreglo de listas, escalares y archivo malformado y quién lo hace | Decisión y PR revisado | Nadie; `filemerge` no se toca |
| D08-DEC-07 | Controles obligatorios y opcionales (OQ 7) | Grupo con 03 y 04 | Lista cerrada | Acta | (TBD) |
| D08-DEC-08 | Actualización desatendida (OQ 8) | Grupo | Qué hace `update` sin poder aprobar el diff | Acta | (TBD) |
| D08-DEC-09 | Remoción acotada en `uninstall` (OQ 9) | Grupo y Neyén | Quién implementa la remoción | Decisión | `uninstall` no quita el overlay |
| D08-DEC-10 | Adopción mínima (OQ 10) | Grupo | Mínimo de cobertura para considerar adoptable el piloto | Umbral acordado | (TBD); nada es adoptable hoy |
| D08-DEC-12 | Pedidos de 07 y 05 (OQ 12, tarea 6.6) | Grupo | Aceptar o descartar cada pedido | Acta | Quedan registrados, sin decidir |

## Plataformas

| ID | Qué falta | Mientras tanto |
|---|---|---|
| D08-PLAT | Probar en Windows y Linux (nativo o CI); macOS y Termux, si el grupo los incluye | No se probó ninguna plataforma en esta ronda |

## Lo que se le pide a Neyén como revisor ALTO

1. **`GuardRollback` no registra `~/.claude/settings.json`** de Claude (ni el `config.toml` de Codex fuera de Estricto): design D8, revisión de todas las llamadas a `GuardRollback*` y `GuardPersistentRestore`.
2. **`WriteFileAtomic` vs. `WriteFileAtomicPreservingParent`:** cuál usar y cómo evitar el `chmod 0755` del directorio padre (`writer.go:54-58`).
3. **`MergeJSONObjects`** trata el JSON malformado como `{}` y reemplaza listas y escalares del usuario (`json_merge.go:10-15`).
4. **Quién arregla `filemerge`** y en qué orden con 03 (D08-DEC-06).
