# Dependencias 05, 06 y 07 para cerrar 04

13 de noviembre de 2026. Análisis en lectura de referencias remotas disponibles en el repositorio. No se implementan 03/05/06/07, no se incorporan sus ramas y no se marcan tareas de 04. El informe unificado y la revisión de Juani se usan como requisitos y evidencia atribuida a sus autores; no autorizan cambios adicionales de alcance ni aceptación de deuda.

## Resultado

**05 avanzó y ya aporta una entrega textual concreta que 04 puede integrar con pruebas. 06 todavía no entrega el contrato corregido requerido por el informe unificado. 07 define un sobre utilizable como referencia, pero su implementación y la definición exacta de `recibos` siguen pendientes.** Por eso este avance permite trabajo parcial de 1.4/2.5, pero no permite declararlas completas ni llamar verificado a un resultado autoconsistente.

| Change | Revisión inspeccionada | Qué aporta ahora | Límite |
|---|---|---|---|
| 05 `opsx-task-context` | `e4ec0f54536760f7fca6692e27bafcb8cd380df1`, anterior `f372681` | `entrega-03-04.md`: cuatro módulos, tabla de acciones y reglas de ausencia/vacío/compactación. Nueva matriz de captura y sondas Windows. | El commit `e4ec0f5` modifica diez archivos documentales. Su cabecera dice «texto listo para integrar; integración en código pendiente de apply». El arnés, adaptadores, catálogo y pruebas de aceptación propias siguen en tasks pendientes. |
| 06 `opsx-verifiable-evidence` | `f3f005e3b09ee835b02c1b966fd8360676f85f48` | Borrador de seis campos por comando, tres estados y revisión de suficiencia. | No incorporó las correcciones del informe unificado ni la nueva propuesta de datos/captura de 05. |
| 07 `opsx-selective-delegation` | `dde9d7b6621efddea83f4a4b19dc2fcdc36bd3e7` | D6 define sobre, veredicto sin verificar hasta corroboración y aplicación equivalente en ejecución directa. D1 prioriza independencia. | Tasks de implementación y piloto pendientes; `recibos` y huella del estado probado dependen de 06. |

El listado grande del diff desde `f372681` incluye una sincronización de 05 con main y trabajo previo de 01/02. Eso no es implementación nueva de captura de 05: la entrega propia de `e4ec0f5` es documental. No se confunde disponer del código de baseline con haber demostrado toda la captura de 05.

## Qué de 05 se puede consumir en 04

Fuente exacta: `e4ec0f5:openspec/changes/opsx-task-context/entrega-03-04.md`, §1–4.

| Entrega | Integración propia de 04 que habilita | Prueba necesaria |
|---|---|---|
| `modules/directed-code-reading.md` | Añadir contenido de lectura dirigida como archivo de apoyo, sin frontmatter de skill. | Archivo inventariado, referencia correcta y reglas conservadas; no aparece como skill independiente. |
| `modules/bounded-command-output.md` | Añadir reglas de captura previa, redacción, resumen y referencia; distinguir captura del agente de independiente. | Composición y límites; resultado del agente permanece sin verificar hasta corroboración. No inferir una salida íntegra de la vista recortada del cliente. |
| `modules/task-tool-selection.md` | Entregar herramientas por acción y límites declarados cuando faltan opcionales. | Lista recibida por especialista, uso recomendado distinto de bloqueo efectivo y ausencia opcional con aviso. |
| `modules/concise-agent-reporting.md` | Entregar reporte breve con procedencia y líneas fijas de otros módulos. | Texto y procedencia conservados sin omitir fallos para acortar. |
| Tabla acción → módulos | Actualizar la tabla Codex y fijar la revisión fuente; los cuatro módulos nuevos no tienen toggle y se entregan siempre. Codex sigue omitiendo `model-assignments` hasta 09. | Acciones explore/propose/apply/verify/archive, módulos requeridos y opcionales, toggles y referencias sin archivos ausentes. |
| Ausente/vacío/compactación | Tratar archivo de cero bytes o solo espacios como faltante; detener solo la acción dependiente; reinvocar y releer tras compactación/reanudación. | Casos positivos y negativos. Un módulo opcional ausente no puede habilitar delegación ni suprimir independencia obligatoria. |

La propuesta de **compositor común con registro de módulos** (05 D23) sigue marcada «propuesto a 03/04». No es condición técnica indispensable para consumir el texto: la entrega describe un camino temporal por compositor. Adoptar el registro común exige acuerdo sobre alcance y pruebas; no autoriza cambiar la implementación de Claude desde 04. La afirmación de 05 de que ninguno de ambos cambios está en main quedó histórica tras el merge de PR #8; al integrar 04 se debe partir del main real, conservando el recorrido de 03.

### Alcance seguro propuesto para una entrega posterior, solo de 04

La siguiente es una propuesta concreta para revisar después de conciliar 04 con main; no se implementó aquí ni se declara aprobada por disponer del texto de 05.

| Ruta de 04 que afectaría | Cambio acotado propuesto | Pruebas a extender |
|---|---|---|
| `internal/harness/config/assets/codex/native/modules/<id>.md` (cuatro archivos nuevos candidatos) | Guardar los cuatro textos fijados en `e4ec0f5`, con origen y revisión identificados. Se propone una ruta Codex para no cambiar el compositor Claude ni fingir implementación del registro común de 05. El destino final sigue siendo `modules/<id>.md` dentro de la skill. | `codex_package_test.go`: contenido esperado, ausencia de frontmatter y entrega sin toggle para los cuatro módulos. |
| `internal/harness/config/codex_package.go` | Cargar archivos en `CodexPackage.Files` y sustituir su tabla `actions` por la entrega Codex adaptada. No incorporar `model-assignments`, no ampliar targets ni rutas del usuario. | `codex_package_test.go`: todas las acciones y combinaciones de toggles, referencias completas, política explícita intacta y composición determinista. |
| `internal/harness/config/assets/codex/native/SKILL.md` | Identificar revisión de 05 y conservar el cierre seguro por contrato/captura ausentes. La tabla de 05 no debe borrar esa condición ni afirmar que 06 ya está corregido. | Casos de texto compuesto para evidencia directa/delegada, contrato ausente e independencia; luego sondas reales de lectura y ausencia/vacío. |
| Tests existentes de `internal/harness/config/` y `internal/verify/` | Extender instalación/update/ownership/retiro y verify al nuevo conjunto esperado. Ajustar producción adicional solo si un test revela una necesidad dentro del alcance aprobado. | `managed_skill_test.go`, pruebas de lifecycle y `internal/verify/managed_skill_test.go`: cuatro archivos inventariados, actualización idempotente, ausencia/vacío/modificación detectados y nota ajena preservada. |

El bloque normativo de `SKILL.md` que detiene acciones con evidencia si faltan **contrato corregido o captura independiente** debe permanecer. Añadir `bounded-command-output` no satisface por sí solo esas condiciones: una captura hecha por el agente sigue siendo declarada, no corroborada. No se deben habilitar esas acciones por la mera existencia de los nuevos archivos. La integración de 06/07 y su aceptación funcional constituye una entrega posterior distinta.

## Captura y datos: avance útil, cierre aún pendiente

05 D8 separa capturador, redactor y productor del resumen. D9 propone referencia `{comando, id de corrida, ruta, sha256 de copia redactada, vencimiento}` y registro de captura con fuente, redactor, cliente/interfaz/sistema/versión, exit code, cobertura previa al truncamiento y redacciones por categoría. Propone caché del usuario fuera del repo, acceso local y retención hasta archivo + siete días; esos siete días son supuesto y D9 sigue **propuesto a 06**, no acuerdo cerrado entre ambos responsables.

| Celda o afirmación | Evidencia que declara 05 | Qué falta para aceptación de 04 |
|---|---|---|
| Codex CLI Windows 0.167.1, `codex exec --json` | X7: `aggregated_output`/`exit_code` entregó salida de 129.047 caracteres, fallo central/final, stdout y solo stderr, exit 0/1. Sonda `n=1`. | Tarea 8.1 de 05 con fixtures conocidos del arnés y adaptador 3.6. D22 exige esa corrida para clasificar fuente independiente aceptada, aunque X7 use ya esa palabra. No extrapolar a sesión interactiva o IDE. |
| Activación Codex Windows | X10/X11: `$sdd-orchestrator` carga cuerpo y pedido sin `$` no la invoca; App Server lista metadata en X9. | Registrar reproducción sobre paquete/revisión finales de 04, permisos efectivos y alcance de la sonda. Es evidencia nueva atribuida a 05; no son pruebas ejecutadas en este análisis. |
| Lectura de módulos, ausencia/vacío y herencia en subagentes Codex | X12 no pudo leer por política de shell; subagentes tampoco comprobados. | Sondas P2 de 04: lectura real de módulos y comportamiento de error, encargo/herencia y permisos. El bypass usado en temporal no acredita protección equivalente. |
| Claude CLI Windows | Vista fallida truncada; copia recuperable del cliente en éxito; hooks no probados. | No usar transcripción/copia sin redactar como original administrado. Captura del agente no prueba procedencia independiente. |
| IDE Windows y CLI/IDE Linux | App Server interno de extensión Codex no disponible como fuente externa; otras fuentes declaradas o no probadas. | Cobertura por celda con fuente real, o límite/no disponible explícito. No extrapolar sondas CLI a extensión. |

La revisión de Juani sobre 04 en Linux prueba instalador y preservación, pero declara no ejecutar Codex real. Las sondas Claude reales de 03 son de otro cliente. Ninguna de esas evidencias cierra la cobertura de captura Codex Linux/IDE que falta en 05/04.

## Correcciones que 06 todavía debe entregar

Fuente exacta: `f3f005e:openspec/changes/opsx-verifiable-evidence/{design.md,specs/execution-receipts/spec.md,tasks.md}`.

| Necesidad del informe unificado | Estado observado de 06 | Contrato/caso que falta |
|---|---|---|
| Separar integridad, procedencia y suficiencia | D3 llama «procedencia mecánica» a comparar revisión/directorio, existencia, huella y exit coherente. | La misma parte puede inventar salida y recibo coherentes. Añadir origen independiente o reejecución que corrobore ejecución y código, manteniendo integridad como resultado separado. |
| Estado exacto probado | `revision` solo exige HEAD y si había cambios sin confirmar. | Identidad inequívoca del árbol probado, incluidos cambios preparados, sin preparar y nuevos. Caso: mismo HEAD con dos diffs distintos. |
| Identidad de ejecución | Un recibo por comando con exit y fecha declarados por subagente. | ID de corrida y asociación demostrable a captura, comando, entorno y árbol. La hora declarada no prueba ejecución. |
| Evidencia directa y delegada | D5 y requisitos de entrega se restringen al módulo/prompt de delegación. | Cargar contrato cuando apply/verify directo requiere evidencia; no hacer depender su carga de que se delegue. |
| Copia redactada por defecto | D6/spec conservan original completo, dos huellas y redacción antes de compartir; ubicación bajo `openspec/changes/.../evidence/`. | Acordar con 05 ubicación, acceso, vencimiento, registro de captura y persistencia redactada antes de disco. No conservar crudo por adoptar literalmente el borrador viejo. |
| Huellas y LF/CRLF | Tarea 4.3 espera igual huella para archivo LF y CRLF usando `--no-filters`. | Huella de bytes: mismos bytes producen misma huella; cambiar LF por CRLF cambia bytes. Fijar algoritmo y comprobar comportamiento con/sin filtros sin afirmar normalización inexistente. |
| Estados consistentes | Spec «invented receipt» rechaza original ausente; «missing field» deja original ilegible sin verificar. | Definir diferencia entre evidencia demostrablemente inventada/adulterada y evidencia faltante/vencida/no accesible. Regla de duda → sin verificar; no aceptación por falta de datos. |
| Suficiencia por requisito | Existe separación normativa y tests verdes insuficientes como escenario. | Quién evalúa y criterios identificables; evidencias que pasen integridad/procedencia pero no cubran el requisito no cierran aceptación. |

No corresponde que 04 invente un tercer esquema de recibo para suplir esos faltantes. Puede conservar sus reglas de no aceptación y preparar casos de consumo; el formato definitivo y las correcciones pertenecen a 06 con coordinación de 05/07.

## Qué conserva y qué espera el sobre de 07

07 D6 define `estado`, `partida`, `modificados`, `hallazgos`, `recibos`, `pendiente`; `estado` admite `completed | blocked | out_of_scope | failed`. `partida` identifica commit y huella, y `modificados` incluye altas, cambios, borrados y renombres. El coordinador verifica partida, alcance atribuible, recibos y criterios de salida. El veredicto queda sin verificar hasta corroboración independiente o reejecución y suficiencia, incluso en ejecución directa.

04 puede referenciar ese sobre sin copiar otro distinto. Pero 07 espera de 06 la huella exacta del estado de trabajo y el contrato de `recibos`, y todas sus tareas de implementación/piloto siguen pendientes. Su señal S1 aún usa más de 300 líneas: 05 pide sumar los 10.000 caracteres y actualizar la prueba de salida verbosa. Eso no se corrige unilateralmente en 04.

## Casos para la integración de 04 cuando los contratos estén fijados

1. Apply directo y retorno delegado con el mismo estándar de captura, identidad del árbol, independencia y suficiencia.
2. Recibo inventado con salida y huella internamente coherentes: no se acepta sin corroboración independiente.
3. Mismo HEAD y dos diffs distintos, incluyendo archivo nuevo: no se reutiliza evidencia del árbol equivocado.
4. Verificación breve que requiere independencia; revisor no habilitado, módulo delegado apagado o superficie sin subagentes: detener esa verificación con motivo.
5. Fallo oculto en medio/final de salida larga, stdout/solo stderr, exit 0/1: resumen no pierde fallos y conserva captura atribuible.
6. Copia redactada ausente, nunca capturada o vencida: estados distintos, todos insuficientes para aceptar el requisito si falta evidencia necesaria.
7. Datos sensibles ficticios y requisito que necesita detalle descartado: el resultado permanece sin verificar; no persistir crudo para evitarlo.
8. Módulo obligatorio ausente/vacío y opcional ausente: solo se detiene la acción dependiente; controles equivalentes deben estar demostrados.
9. Archivo LF frente a CRLF y mismos bytes bajo configuraciones Git distintas: aserciones de huella coherentes con bytes y algoritmo acordados.
10. Tests verdes que no ejercitan el requisito: suficiencia fallida aunque captura e integridad sean correctas.

Estos son criterios de consumo e integración de 04, no una declaración de que 05/06/07 ya los implementaron o probaron.

## Próximo paso dentro del alcance de Matías

Fijar `e4ec0f5` como entrega textual candidata de 05, revisar su incorporación a los assets Codex con TDD y actualizar referencias cuando se integre. Mantener separadas la entrega de módulos y la aceptación de captura. Pedir a 06 revisión corregida y acuerdo explícito con 05 sobre datos/retención; pedir a 07 confirmación del sobre, independencia y actualización de S1. Después, implementar el consumo correspondiente en 04 y ejecutar sus casos. Hasta entonces, 1.4 y 2.5 continúan abiertas y las acciones que necesitan el contrato ausente conservan el límite informado.
