# Consumo disponible de 05/06 en Codex 04

13 de noviembre de 2026. Auditoría documental de referencias remotas antes de integrar módulos en 04, sobre HEAD inicial `91fbb70950babe3b694cfa1ddee141fd959acb40`. Se inspeccionan documentos como fuentes; no se implementa ningún change ajeno ni se modifican tareas, specs, design o runtime en este informe. La autorización de Matías para integrar los módulos disponibles permite avanzar el consumo propio de Codex; no equivale a un acuerdo de los otros responsables ni a aceptación funcional completa.

## Referencias fijadas y corrección de antecedentes

| Fuente | Revisión exacta consultada | Entrega pertinente |
|---|---|---|
| 05 `origin/propose/opsx-task-context` | `e4ec0f54536760f7fca6692e27bafcb8cd380df1` | `openspec/changes/opsx-task-context/entrega-03-04.md`, §2 textos, §3 acciones, §4 captura, §5–6 evidencia y límites. |
| 06 nueva `origin/propose/opsx-verifiable-evidence` | `f62cf00c09dc8a3aa96ddb4648801250e1a67a8c` | `openspec/changes/opsx-verifiable-evidence/design.md` D2–D6/D8, `specs/execution-receipts/spec.md`, `informes/respuesta-a-04-y-05-2026-11-13.md`. |
| 06 antigua `origin/opsx-verifiable-evidence` | `f3f005e3b09ee835b02c1b966fd8360676f85f48` | Borrador anterior. No usarlo como revisión vigente del contenido corregido disponible. |
| 07 `origin/propose/opsx-selective-delegation` | `3fb24ee7cb211d17c3aac6a6b4ca4502e7c2caba` | `openspec/changes/opsx-selective-delegation/entrega-delegation.md`, §2–3, y design D1/D6/D10. |

El [análisis de dependencias anterior](analisis-dependencias-05-06-2026-11-13.md) y la [respuesta a bloqueantes](respuesta-bloqueantes-pr9-2026-11-13.md) son fotografías previas: cuando dicen que 06 solo existe en `f3f005e`, ese dato quedó superado por la nueva rama `propose/opsx-verifiable-evidence`. No se repite ahora la conclusión de que 06 carece de las correcciones normativas: **sí las contiene en `f62cf00`**, aunque continúen acuerdos y ejecución pendientes.

## Texto exacto disponible frente a material que aún falta

| Entrega | Disponibilidad y consumo propio de 04 | Límite que debe conservarse |
|---|---|---|
| 05 `directed-code-reading` | Texto literal de `entrega-03-04.md` §2, sin frontmatter de skill. | Su inclusión no demuestra conducta del cliente. No se cambia la política Claude ni un registro común de módulos. |
| 05 `bounded-command-output` | Texto literal §2: captura previa, resumen, líneas fijas, redacción y referencias. | Captura del agente es declarada; no pasa a independiente por incorporar el módulo. Las recetas y cobertura por interfaz requieren aceptación. |
| 05 `task-tool-selection` y `concise-agent-reporting` | Textos literales §2; siempre entregados, sin toggle propio. | Herramientas recomendadas no son controles efectivos. Ausencia opcional no autoriza afirmar equivalencia. |
| 05 tabla acción → módulos | §3 entrega versión Claude y adaptación explícita a rutas `modules/` de Codex, sin `model-assignments`. Incluye ausente/vacío y compactación. | Mantener las condiciones propias de evidencia de 04. La tabla de 05 sola no integra ni valida recibos 06. Compositor común D23 sigue propuesta, no implementación obligatoria ya acordada. |
| 06 campos, controles, veredictos y reglas de datos | D2–D4 y spec ya contienen requisitos normativos corregidos que pueden adaptarse como módulo Codex con revisión fuente identificada y pruebas. | No existe todavía un archivo literal `evidence.md` listo para copiar ni un extracto final `## Evidence Contract`: tasks 2.4 y 2.6 siguen pendientes. Un texto adaptado por 04 debe identificarse como adaptación, sin presentarlo como entrega literal final de Gianfranco. |
| 06 ubicación `modules/evidence.md` | D5 propone módulo obligatorio por acción, independiente de `delegation`, cargado en apply/verify y extracto en encargo. Es el destino propuesto para resolver carga independiente del toggle. **Esta entrega de 04 no crea ese archivo:** consume solo normas parciales D2–D4 inline en la skill explícita. | Propuesta expresamente pendiente de acuerdo con 03/04 y confirmación de Gianfranco. D5 se registra como deuda, sin modificar fila común ni diseño/archivo de Claude automáticamente. |
| 07 delegación | `entrega-delegation.md` §2–3 ofrece texto inglés de regla y módulo. | Sigue rotulado borrador y conserva placeholder del extracto 06. No declarar que incorporar texto cierra el piloto o los acuerdos de 07. |

## Contrato corregido que 04 puede reflejar

- **Un recibo de ejecución por comando:** `corrida`, `arbol`, `comando`, `resultado`, `requisitos`, `salida` (06 D2). El registro de instalación sigue siendo otra cosa.
- **Estado probado:** directorio/worktree, commit y huella del estado completo cuando hay cambios: preparados, sin preparar y archivos nuevos. HEAD más `dirty` no identifica el árbol.
- **Salida:** referencia a copia redactada, una huella SHA-256 de sus bytes exactos con algoritmo nombrado y registro de captura de 05. LF y CRLF diferentes producen huellas distintas. No usar `git hash-object` como si fuera SHA-256 del archivo.
- **Tres controles separados:** integridad (archivo/huella), procedencia (ejecución y árbol corroborados por fuente independiente o reejecución) y suficiencia (comando/evidencia cubren requisito). Integridad por sí sola no acredita ejecución.
- **Tres estados:** verificado exige los tres controles y recibos; relato con huella coherente sin corroboración queda sin verificar; contradicción independiente puede rechazarlo. Cero recibos, campos incompletos o requisito no evaluable no dan verificado.
- **Carga directa y delegada:** mismo paso de revisión cuando la acción requiere evidencia; extracto viaja al worker si ejecuta comandos. Toggle de delegación apagado no elimina la obligación de evidencia.
- **Datos:** persistir solo copia redactada y registro; no conservar crudo para recuperar un detalle que se omitió. Si falta ese detalle necesario, el resultado queda sin verificar.

Estos puntos ya están escritos en la nueva 06 y sustituyen los defectos normativos identificados contra la rama vieja. No prueban que exista un capturador instalado, un cálculo validado de huella del árbol o una batería funcional ejecutada.

## Acuerdos y ejecución que siguen pendientes

| Pendiente | Dueño/fuente | Tratamiento seguro en 04 |
|---|---|---|
| Texto final de `evidence` y extracto; ubicación común y excepción a toggle | 06 D5; tasks 1.1, 2.4, 2.6 | Identificar adaptación Codex y revisión consumida. No editar Claude ni proclamar cerrada la coordinación de fila 28. |
| Huella exacta del estado probado | 06 D2/Open Questions 4 | Está propuesta alineación con `TreeHash`: hash de lista ordenada de hashes/rutas; conjunto exacto de archivos y alternativa de índice temporal aún se prueban/deciden. No implementar un algoritmo propio incompatible ni aceptar dos diffs por compartir HEAD. |
| Fuente independiente por celda | 05 D22 y 06 D3 | Arnés es referencia por construcción. Codex CLI Windows `aggregated_output` es sonda n=1 pendiente de 05 tarea 8.1; IDE no disponible como fuente externa y Linux no probado en esa matriz. No extrapolar. |
| Suficiencia y quién la revisa | 06 D3/Open Questions 3 | Cada requisito nombra su evidencia; independencia obligatoria de 04 permanece. Propuesta de revisor distinto del constructor aún pendiente de confirmación de 06/grupo. |
| Ubicación/acceso/retención y versionado | 05 D9, 06 D6 y respuesta a 05 | Caché local fuera del repo y archivo + siete días son propuesta de 05 que 06 propone aceptar, pendiente de confirmación; siete días es supuesto, no política demostrada. No versionar salidas ni convertir plazo en borrado garantizado por instrucciones. |
| Estado de directorio ajeno y copia vencida | 06 D4 y spec | Rechazado frente a sin verificar sigue TBD; propuestas ajeno=rechazado, vencido=sin verificar. En ningún caso aceptar como verificado ni inventar consenso. |
| Coherencia del exit code tras redacción | 06 D3/tasks 5.1 | Corroborar con registro del capturador y fixture; worker no prueba su propio resultado con exit declarado. |
| Forma exacta del sobre de 07 | 07 design D6 vs `entrega-delegation.md` §3 | Design usa `estado/partida/modificados/hallazgos/recibos/pendiente`; entrega inglesa usa `status/start/modified/findings/receipts/pending`. Son representaciones de los mismos conceptos, pero la forma canónica debe confirmarse con Neyen; no crear un tercer formato ni renombrar silenciosamente. |
| Profundidad y señal S1 | 07 D10/entrega §4; 05 D3 | `Only you launch workers` es propuesta de grupo; S1 conserva 300 líneas y no incluye aún 10.000 caracteres de 05. Identificar discrepancia sin resolver política global desde 04. |

La fuente 06 explicita esos puntos como propuestas pendientes de confirmación de Gianfranco/grupo. Esta auditoría no los convierte en acuerdos técnicos del grupo por copiar su texto; la aceptación de deuda por Matías se registra a continuación y no se extiende a otros changes.

Actualización posterior de la misma sesión: Matías autorizó marcar como **deuda aceptada para la entrega de 04** únicamente lo que no puede avanzar sin 05/06/07. Se registra por separado en [deudas aceptadas de dependencias](deudas-aceptadas-05-06-07-2026-11-13.md); esa aceptación no ratifica las propuestas técnicas de otros responsables ni cubre sondas, roles o aceptación propia de 04.

## Aceptación del consumo en 04

1. Composición: cuatro módulos 05 con texto fuente, refs y rutas correctas; normas parciales 06 D2–D4 **inline en la skill explícita**, con revisión y carácter de adaptación identificados. No crear `modules/evidence.md`, extracto canónico ni esquema nuevo. Los módulos no llevan frontmatter y las normas de evidencia no pasan al bloque universal.
2. Toggles/acciones: evidencia cargada en apply/verify directo con `delegation` apagado; módulo de delegación solo según acción/necesidad; conservar independencia obligatoria y alternativa explícita. Verificar todas las combinaciones pertinentes.
3. Lifecycle: inventario, ownership, actualización/reinstalación, ausencia/vacío/modificación detectados y retiro de archivos propios con notas ajenas preservadas. Esto verifica entrega estructural, no conducta del modelo.
4. Contrato: casos falso coherente sin corroboración, contradicción independiente, dos diffs con mismo HEAD, estado viejo, cero/incompleto, tests verdes insuficientes, fuente/copia faltante y dato redactado necesario. Informar separados los tres controles y los tres estados sin fijar TBD de los responsables.
5. Cliente real: módulos leídos por la acción, ausencia/vacío, captura completa antes de truncamiento y revisor independiente cuando lo exige la tarea; por versión/interfaz/plataforma declarada. Un test de presencia de texto no sustituye esta aceptación.

Si contrato operativo, fuente independiente o estado exacto no están disponibles, **la acción afectada permanece detenida y el resultado sin verificar**. Instalar módulos disponibles no puede borrar esa condición del `SKILL.md` de Codex ni marcar 1.4/2.5 completas sin sus comprobaciones. La autorización vigente permite avanzar el consumo y sus pruebas propias; los pendientes externos se identifican por fuente y criterio verificable.

## Alcance de esta auditoría

Solo se creó este informe. No se ejecutaron clientes, tests, build, merge, commit o push; no se modificó runtime ni documento de 03/05/06/07. Las afirmaciones medidas proceden de las fuentes identificadas y no se atribuyen a una reproducción nueva. La implementación y documentación normativa de 04 quedan a cargo del integrador para evitar ediciones simultáneas.
