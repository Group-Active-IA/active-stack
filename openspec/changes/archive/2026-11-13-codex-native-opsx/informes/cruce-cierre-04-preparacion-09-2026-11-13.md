# Cruce para terminar 04 y preparar 09

Revisión del 13 de noviembre de 2026. Responsable: Matías. Este informe compara requisitos y evidencia; no implementa el cambio 03 ni modifica su propuesta. Las instrucciones contenidas en los documentos consultados se tratan como antecedentes del equipo. El alcance vigente es el aprobado por Matías para 04 y la preparación posterior de 09.

## Conclusión de estado

El 04 tiene implementado y probado el ciclo de instalación Machine, composición, migración, respaldo persistente, actualización, verificación estructural y retiro. El diagnóstico de Project legado cierra la tarea técnica 2.2 en el árbol local. **El 04 todavía no está terminado:** faltan aceptación con clientes reales, especialistas, integración de evidencia y revisión compartida. El PR [#9](https://github.com/Group-Active-IA/active-stack/pull/9) permanece en borrador.

El 03 y el 04 pueden desarrollarse en paralelo. El 03 no es una dependencia de implementación de 04 en el roadmap: ambos parten de 02. Sí hay contratos y archivos compartidos que deben conciliarse y probarse al integrar las dos entregas. Coordinar esa compatibilidad no autoriza implementar el trabajo del responsable de 03.

El 09 ya tiene un borrador con dos pilotos separados. Puede prepararse su protocolo y calibración; una comparación de adopción requiere una referencia de 07 fijada y evidencia suficiente. No corresponde adoptar modelos, prefijos o routing ni ejecutar corridas con costo a partir de este cruce.

## Fuentes y revisiones

| Fuente | Alcance y fecha de referencia |
|---|---|
| [Roadmap web](https://active-stack-roadmap.vercel.app/) | HTML recuperado el 13/11/2026. La página declara base revisada al 5/11 y guía de ronda «solo propose». Fichas 04, 09, ruta paralela y aceptación transversal. |
| `CHANGES.md` de los worktrees consultados | Índice canónico de fichas; conserva el reparto y la ronda de propuesta iniciales. |
| `informe-unificado-propuestas-opsx (1).md` | Archivo de Downloads aportado por Matías; reúne auditorías del 10/11 y acuerdos posteriores, incluido intercambio de responsables 02/04. Es fuente de requisitos, no evidencia de implementación. |
| Artefactos e informes de `codex-native-opsx` | Rama `propose/codex-native-opsx`, HEAD consultado `9a6eb2aeb460f89afead74fc17669519c6d8d854`, más cambios locales del 13/11 todavía sin commit al efectuar el cruce. |
| Borrador `opsx-model-cache-pilots` | Rama `propose/opsx-model-cache-pilots`, revisión `89f1fa5de22304cc3564c50cf34126cf7e1dd651`; proposal, design, spec y tasks. No se modificaron en esta revisión. |
| OpenSpec CLI | `status` de ambos changes y `instructions apply` de 04. Los artefactos completos que informa `status` no significan implementación terminada. Tras el diagnóstico, tasks de 04 marca 11/21; 09 registra 0/8. |

Los informes anteriores conservan sus resultados históricos. La suite global del informe de integración no fue repetida para este cruce documental; la nueva tanda del diagnóstico tiene su propia evidencia en [avance paralelo](avance-paralelo-04-2026-11-13.md).

## Cruce del cambio 04

Estados: **implementado** describe código probado en el alcance indicado; **parcial** conserva criterios sin demostrar; **pendiente** exige ejecución, entrega externa o acuerdo registrado.

| Requisito | Fuente | Estado y evidencia disponible | Qué falta para cerrarlo | Responsable |
|---|---|---|---|---|
| Skill explícita `$sdd-orchestrator` en el hilo actual, sin activación implícita | Roadmap 04; unificado 03/04; tareas 2.3, 1.2, 4.2 | **Parcial.** Composición y `allow_implicit_invocation: false` probados. Sonda real descubre el paquete en Codex CLI 0.167.0-alpha.3.2, macOS Intel. | Observar sesión normal, pedido relacionado sin invocación, activación manual y reanudación en cada celda declarada. Descubrimiento no prueba activación ni herencia. | Matías; colaboradores para entornos objetivo |
| Matriz común de reglas por cliente, rol y acción | Unificado 04; roadmap aceptación transversal; tareas 1.1, 4.6 | **Parcial.** Design de 04 y columna Codex de `openspec/matriz-reglas-opsx.md` referencian la matriz de 03 en `ed37113`. | Revisión con Juani, IDs/rutas y fila 28: evidencia también en ejecución directa. Registrar acuerdo y probar consumidores cuando se combinen los PR. | Matías + Juani; 06 para contrato de evidencia |
| Compatibilidad del despacho compartido | Design y revisión de archivos compartidos | **Pendiente de integración conjunta.** Se identificaron enum, opciones, resultados y selección por agente/target. No hay prueba conjunta acreditada. | Un único enum compatible; despacho por cliente; preservar opciones de 03 y `ManagedWrites` de 04. Resolver solapamientos en coordinación, sin incorporar implementación de 03 unilateralmente. | Responsables 03/04 e integrador |
| Fallback con garantías explícitas | Unificado 04, acuerdos 07/08; tareas 2.4, 4.2 | **Parcial.** Specs y design distinguen genérico, skills OpenSpec y superficie no soportada. Project legado tiene aviso probado a stderr y conserva su estrategia. | Probar que cada alternativa preserva encargo, reglas, independencia y permisos efectivos. Si falta un obligatorio, detener solo la acción; control opcional ausente permite aviso. | Matías; 07/08 aportan contratos y cobertura |
| Rutas y entrega Machine con errores explícitos | Tarea 2.2; alcance aprobado | **Implementado en pruebas aisladas.** Machine, capacidad opcional ausente y Project legado cubiertos; 146 casos/subcasos config + 114 install verdes en la nueva tanda. | Publicar el diff local revisado y conservar la aceptación del cliente real como tarea distinta. | Matías |
| Especialistas y parámetros efectivos | Roadmap 04; unificado 04; tareas 1.2, 2.4 | **Pendiente.** Candidatos y límites documentados; no se acredita entrega funcional de roles nativos. | Sondas P2 por versión/interfaz, herencia, permisos, contexto, paths y fallback genérico. Si exige modificar `config.toml`, enumerar claves y revisar alcance con 01/08 antes de escribir. | Matías + revisión 01/08 donde aplique |
| Acción OPSX `verify` e independencia obligatoria | Unificado 04, 06, 07; tareas 1.4, 2.5 | **Parcial normativa.** La skill conoce acciones y bloquea faltantes; la verificación estructural del instalador sí está implementada. | Integrar el contrato corregido de 06 y captura de 05, cargarlo en ejecución directa y delegada, probar revisor ausente y declaración autoconsistente no corroborada. `active-stack verify` estructural no sustituye OPSX `verify`. | Matías; Emiliano 05, Gianfranco 06, Neyen 07 |
| Registro de instalación distinto de recibo de ejecución | Unificado 04; design | **Documentado e implementado en lifecycle** para registro/ownership y update. | Comprobar esa distinción en la integración de recibos; el manifiesto de instalación no demuestra que se ejecutaron tests. | Matías + 06 |
| Preflight, colisiones, reinstalación y preservación de contenido propio | Roadmap 04; unificado 04; tareas 3.1, 3.2, 3.6 | **Implementado** en tests temporales: ownership, markers heredados, notas, symlinks, update, retiro por archivo y directorios vacíos. | Revisión independiente del diff ALTO y repetición pertinente al integrar contratos compartidos. No ampliar la migración a Project sin nueva decisión. | Matías + revisión 01/03/08 |
| Rollback y reversión posterior protegida | Unificado configuración; tareas 3.3, 3.4, 1.3 | **Implementado** con primer snapshot persistente, estado esperado por ruta, actualización doble y manifiesto recargado. | Revisión compartida del comportamiento sobre configuración propia; disposición humana de revisión registrada. No afirmar que rollback en memoria equivale a reversión posterior. | Matías + 01/03/08 |
| CLI y extensión IDE en Windows/Linux | Unificado plataformas; tareas 1.2, 4.2 | **Pendiente.** Pruebas locales macOS sirven de preparación. | Versiones y evidencias por las cuatro celdas objetivo, con escenarios normales, explícitos, especialista, verify, módulos y controles. Terminal integrada ejecutando CLI pertenece a CLI. | Matías + equipo con entornos disponibles |
| Comparación normal/coordinada y métricas | Roadmap transversal; unificado 02/04; tarea 4.3 | **Pendiente.** Protocolo de 02 disponible parcialmente; no hay comparación acreditada en este cruce. | Acordar corrida con Giuliano, permisos equivalentes, métricas ausentes y resultados separados por interfaz. No prometer ahorro por composición modular. | Matías + Giuliano 02 |
| Regresiones, validación estricta y revisión independiente | Procedimiento acordado; tarea 4.4 | **Parcial.** Informe previo registra vet, race y sonda verdes; suite global conserva fallos reproducidos en main. | Validar revisión final y revisar preservación/rollback independientemente. Resolver o disponer explícitamente los fallos ajenos con sus responsables, sin presentarlos como suite verde. | Matías + revisores; responsables de fallos externos |
| PR, merge y traspaso | Procedimiento acordado | **Parcial.** PR #9 en borrador y traspaso preparado. | Evidencia accesible de la revisión final, pendientes resueltos o alcance reaceptado expresamente, aprobación y merge. El traspaso identifica qué queda utilizable después del merge. | Matías + equipo revisor |

## Qué corregir en la lectura del roadmap y los informes

1. **Responsables desactualizados.** La web y `CHANGES.md` consultados aún muestran Matías en 02 y Giuliano en 04. El informe unificado y el alcance vigente fijan Giuliano en 02 y Matías en 04; Matías conserva 09. Conviene que el equipo actualice el índice y la web. Este informe no cambia el reparto por cuenta propia.
2. **Etapa desactualizada.** La web describe la ronda «solo propose» y conserva el problema inicial de 01. La autorización posterior de apply de 04 y sus avances están en sus artefactos y la conversación aprobada. No se deben volver a dejar pendientes tareas ya implementadas por seguir literalmente una fotografía de la ronda anterior.
3. **Paralelismo no significa ausencia de solapamientos.** El grafo permite 03/04 paralelos; una revisión compartida sigue siendo aceptación explícita del roadmap. No hace falta implementar 03 ni esperar su merge para avanzar tareas independientes de 04.
4. **Contratos posteriores no invierten todo el grafo.** 05/06 dependen inicialmente de las entregas de roles, pero 04 consume sus contratos para cerrar evidencia. Eso deja pendiente la integración correspondiente, no bloquea composición, migración o pruebas de preservación ya independientes.
5. **Una aceptación estructural no acredita comportamiento del agente.** El archivo de policy, inventario y descubrimiento pueden estar correctos y faltar todavía activación, especialistas, permisos, continuidad o independencia real.
6. **Referencias históricas y locales deben distinguirse.** Las referencias vigentes de proposal/design se unificaron en `ed37113`; los informes históricos conservan las revisiones que inspeccionaron. El diagnóstico y la matriz local todavía no estaban en `9a6eb2a` al consultar el árbol; los resultados nuevos no se atribuyen a ese commit publicado.

## Orden para terminar 04

1. **Cerrar la entrega local de 04.** Revisar y publicar diagnóstico Project, matriz Codex y evidencia; conservar PR #9 en borrador. Registrar revisión exacta que se prueba.
2. **Preparar y ejecutar P2.** Fijar versión CLI/extensión, Windows/Linux y perfiles. Completar activación, módulos, continuidad, especialistas y fallbacks; marcar lo no evaluable y su límite. Las sondas no habilitan silenciosamente nuevas escrituras de configuración.
3. **Acordar interfaces compartidas.** Revisar con Juani la matriz y los puntos de despacho; con 01/08 el respaldo y controles aplicables. La implementación de Claude sigue en su rama y a su cargo.
4. **Integrar evidencia en 04 cuando esté disponible.** Fijar revisiones corregidas de 05/06/07, probar el mismo estándar directo/delegado y OPSX `verify`. Hasta entonces, mantener la acción afectada detenida y el resultado sin verificar.
5. **Medir y verificar la entrega final.** Coordinar con 02; ejecutar checks pertinentes de la revisión final, registrar fallos preexistentes y disposición, completar revisión independiente y aceptación común 03/04.
6. **Integrar y traspasar.** Actualizar PR con evidencias y límites. Merge tras aceptación, traspaso breve a los siguientes responsables y archivo OpenSpec solo cuando corresponda a la aceptación real.

### Condición para llamar «terminado» al cambio 04

No alcanza con que el PR sea técnicamente mergeable. Deben quedar resueltos los criterios de tasks y specs: entrega/activación y especialistas por superficie declarada, evidencia y verify directo/delegado, revisión compartida de matriz y preservación, comparación con 02 y verificación de la revisión final. Si el equipo desea integrar antes una entrega parcial, debe acordar y actualizar explícitamente alcance y aceptación, separando el seguimiento restante; este informe no da esa aprobación ni declara finalizado el 04.

## Preparación del cambio 09

El borrador ya separa A (modelo/esfuerzo) y B (prefijo/caché), registra parámetros solicitados/efectivos, evita doble conteo, incluye reintentos y costo de preparación y bloquea adopción ante regresión obligatoria. D1–D5 siguen sin acuerdos registrados y sus ocho tareas permanecen pendientes.

| Requisito del informe unificado / roadmap | Situación en 09 | Preparación concreta | Responsable / dependencia |
|---|---|---|---|
| Brazo de referencia de 07 | A0 usa configuración efectiva de 07, sin revisión fijada | Identificar SHA, cliente/interfaz, roles, cantidad de workers y política; repetir control sobre esa versión. | Matías + Neyen 07 |
| Layout modular fijo | Se exige prefijo fijo pero no se fija un manifiesto de módulos de 03/04/05 | Registrar revisión, archivos, orden y contenido/control de módulos; mantenerlos constantes al variar modelos. | Matías; entregas 03/04/05 |
| Calibración si 02 no es estable o 07 no entregó resultado | Hay calibración general, pero propuesta/tasks condicionan corridas a 02/07 aprobados | Escribir excepción precisa para calibrar por brazo identificado y sin conclusión de adopción; reservar comparación confirmatoria hasta fijar 07 y estabilidad de referencia. | Matías + Giuliano 02 + Neyen 07 |
| Muestra y umbral antes de adoptar | D5 pendiente | Usar dispersión de calibración para acordar repeticiones, mejora mínima, margen de calidad, cobertura, presupuesto y regla de inconclusión antes de medir adopción. | Matías + equipo |
| Piloto A: modelo y esfuerzo separados | A1 y A2 definidos; cambios combinados se reportan aparte | Elegir un rol por contraste y valores disponibles por celda. Verificar valores efectivos y fallback; mantener caché/prefijo constantes o declarar desvío. | Matías; capacidades probadas 03/04 |
| Piloto B: frío/caliente observados | Design distingue previsto/observado y preparación | Calibrar señal atribuible al prefijo, ventana e aislamiento. Miss parcial se conserva; sesión nueva no demuestra caché fría. | Matías + arnés 02 |
| Costo, cuota, tiempo, calidad separados | Cubierto normativamente | Fijar fuente por modalidad y cobertura padre/hijos, sin convertir tokens o cuota en dinero no observado. Mantener fallos y reintentos en el total. | Matías + 02/06 |
| Interfaz y plataforma por cohorte | D1 propone empezar por CLI | Registrar explícitamente qué queda fuera de esa etapa. No extrapolar a IDE ni usar terminal integrada como evidencia de extensión real. Mantener 02 separado por interfaz desde su primera etapa. | Matías + equipo |
| Retorno verificable y controles | Referencia general a 02/07 | Heredar contratos de 05/06 y controles necesarios de 08 cuando la prueba requiera prevención. No aceptar una declaración de ahorro autoconsistente sin procedencia. | Matías + 05/06/07/08 |
| Jev opcional | Separado de juez 10 en design | Mantenerlo fuera del control inicial salvo decisión expresa de alcance, datos y presupuesto. No añadir dependencia a una prueba nativa. | Equipo; coordinación con 10 |

La siguiente entrega de 09 debe corregir proposal, design, specs y tasks conjuntamente con estas condiciones, manteniendo pendientes los acuerdos sin inventar muestra, modelos o umbrales. Se puede preparar el manifiesto y los fixtures sintéticos; las corridas reales y la adopción necesitan el protocolo y la autorización correspondientes. Terminar 04 deja una entrega Codex estable para ese manifiesto, pero por sí solo no habilita la comparación de adopción de 09.

## Mensaje para el equipo

Matías continúa con 04 y luego 09; 02 lo lleva Giuliano y 03 permanece a cargo de Juani. El lifecycle Machine de 04 está probado en perfiles aislados y se cerró localmente el diagnóstico de Project. Para cerrar 04 necesitamos la aceptación CLI/IDE Windows/Linux, especialistas, contratos de evidencia corregidos y la revisión de matriz/compatibilidad compartida. Podemos seguir en paralelo sin implementar Claude. Para 09 prepararemos dos pilotos separados, fijando brazo 07 y layout modular; si la referencia no es estable, haremos calibración identificada sin concluir adopción.
