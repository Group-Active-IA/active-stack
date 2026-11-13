# Deuda externa aceptada para la entrega de 04

13 de noviembre de 2026. **Aceptación vigente: Matías amplió expresamente la instrucción a dejar como deuda los cambios que no dependen de él.** Sustituye la restricción inicial a 05/06/07 y comprende insumos de 02, revisiones/disposiciones de otros responsables y fallos ajenos. La aceptación permite retirar esos pendientes del checklist activo propio y publicar el avance para revisión. No aprueba contratos en nombre del grupo ni declara terminado el change o mergeable el PR. El nombre de este archivo se conserva por los enlaces existentes; su alcance vigente es toda la deuda externa identificada.

Este registro complementa [consumo disponible de módulos](consumo-modulos-05-06-2026-11-13.md). No marca checks ni archiva OpenSpec. Las normas disponibles se consumen en Codex con su procedencia; lo que falta externamente se conserva con estado y criterio de reanudación verificable.

## Revisiones de partida

| Proveedor | Rama y revisión |
|---|---|
| 05 · Emiliano | `origin/propose/opsx-task-context` · `e4ec0f54536760f7fca6692e27bafcb8cd380df1` |
| 06 · Gianfranco | `origin/propose/opsx-verifiable-evidence` · `f62cf00c09dc8a3aa96ddb4648801250e1a67a8c` |
| 07 · Neyen | `origin/propose/opsx-selective-delegation` · `3fb24ee7cb211d17c3aac6a6b4ca4502e7c2caba` |

La rama antigua de 06 `origin/opsx-verifiable-evidence` en `f3f005e` es antecedente, no la entrega corregida que se consume ahora.

## Registro de deuda autorizada

**Estado de todas las filas:** deuda aceptada por Matías para 04; resultado externo e integración funcional final **no implementados ni verificados**. Cada responsable de 04 debe comprobar el resultado recibido antes de cerrar la fila; la disponibilidad de una propuesta no cuenta como entrega probada.

| ID | Bloqueo y proveedor | Artefacto externo faltante | Criterio de cierre y prueba al retomar | Riesgo y comportamiento mientras tanto |
|---|---|---|---|---|
| D04-0501 | Captura independiente operativa, 05 | Capturador/adaptador y registro según D8/D9/D22; pruebas 3.6/8.1 de 05 y cobertura antes del truncamiento por celda. X7 Windows CLI es sonda n=1, no batería completa. | Fijar revisión y fuente por interfaz; ejecutar fixtures con fallo inicial/central/final, stdout/solo stderr y exit 0/1; cotejar salida conocida, exit y registro. Integrar consumo en 04 sin sustituir su aceptación de cliente real. | Una captura hecha/declarada por el mismo agente podría omitir un fallo. No llamarla independiente ni permitir que una huella por sí sola cierre procedencia. Acción que requiere captura independiente disponible permanece detenida si falta. |
| D04-0601 | Texto final y localización del contrato, 06 con 05/07 y coordinación de roles | Extracto literal `## Evidence Contract`, paso de revisión y acuerdo D5 sobre `modules/evidence.md` independiente del toggle. Tasks 2.4/2.6 de 06 siguen pendientes. | Recibir texto versionado y disposición del responsable; confirmar ubicación y matriz aplicable a Codex, probar carga directa y delegada con `delegation` apagado, encargo con extracto y ausencia/vacío que detiene solo acción dependiente. Seguimiento Claude lo hace su responsable. | Un bloque parcial podría confundirse con contrato completo. 04 solo incorpora requisitos normativos disponibles con etiqueta parcial y revisión; no inventa módulo final ni afirma satisfecha la integración de recibos. |
| D04-0602 | Identidad operativa del estado probado, 06 con 05 | Conjunto exacto de archivos, mecanismo y pruebas de huella de árbol de D2/1.3. Alineación con TreeHash es propuesta; índice temporal es alternativa no decidida. | Entrega y prueba Windows/Linux: mismo HEAD con dos diffs, preparado/sin preparar/nuevo/borrado/renombrado y bytes LF/CRLF; recibo de estado distinto no se acepta para el actual. | HEAD+dirty atribuye pruebas a otro código. No crear algoritmo propio ni aceptar estado insuficientemente identificado. Exigir estado exacto como condición normativa y dejar sin verificar si no se corrobora. |
| D04-0603 | Política de conservación y estados de casos límite, 05/06 | Confirmación de ubicación/acceso/vencimiento, versionado de recibos y veredictos de copia vencida/directorio ajeno. D9 de 05 y D4/D6 de 06 mantienen propuestas/TBD. | Recibir decisión identificada y probar referencia válida, nunca capturada, ausente, vencida, directorio ajeno y copia redactada. Ninguna salida completa entra al repositorio; no guardar crudo para recuperar detalle omitido. | Secreto persistido, copia inaccesible o expiración ficticia. No prometer borrado automático por una instrucción ni fijar siete días como política firme. Ningún caso incierto puede resultar verificado. |
| D04-0604 | Suficiencia y coherencia de evidencia redactada, 06 con 05 | Disposición sobre evaluador/criterios identificables y fixture 5.1 que demuestra exit/fallos conservados tras redacción. | Recibir criterios y correr tests verdes que no cubren requisito, fallo oculto y detalle necesario descartado. Informar integridad, procedencia y suficiencia separadas; resultado falso coherente sin corroboración no se aprueba. | Evidencia auténtica pero insuficiente puede aceptar comportamiento incorrecto. La regla de independencia obligatoria de 04 permanece; sin criterio evaluable o detalle suficiente, resultado sin verificar. |
| D04-0701 | Retorno final con recibos y partida operativos, 07 dependiente de 06 | Sobre validado que usa recibo final de 06 y huella exacta para partida; confirmación de forma canónica entre design español y entrega inglesa, más pruebas de retorno y combinación. | Fijar revisión de 07; probar partida distinta, cambios preexistentes no atribuibles, commit/renombre fuera de alcance, recibos incompletos, dos retornos válidos que fallan combinados y revisión directa equivalente. | Un retorno formal podría atribuir trabajo ajeno o aceptar ejecución no corroborada. Se referencia 07 y no se crea un tercer sobre; sin recibos/partida suficientes no se declara verificado. |
| D04-0702 | Ajustes coordinados de política y texto de delegación, 07 con 05 | Confirmación de S1 frente al umbral líneas/caracteres de 05 y profundidad D10; la entrega de 07 conserva decisiones pendientes. | Recibir disposición de Neyen/grupo y probar regla por pasos, prioridad absoluta de independencia, contención serializada, ausencia de módulo y dos workers con escritura aislados. No adoptar piloto/configuración de referencia de 07 sin resultado. | Un supuesto de política podría convertirse en obligación global o afectar el piloto 09. Consumir solo lo autorizado en Codex, identificar propuestas y conservar bloqueo de revisión independiente si no hay revisor/módulo habilitado. |

## Qué sí puede entregar 04 ahora

Los cuatro textos de 05 y su tabla son entrega textual disponible. De 06 están disponibles los requisitos corregidos D2–D4: seis campos, estado exacto, SHA-256 de bytes redactados, tres controles y tres estados, mismo estándar directo/delegado. Su adaptación normativa dentro de la skill explícita de Codex puede avanzar y probarse sin crear un esquema ejecutable o afirmar que existe un contrato final instalado.

El texto de 04 debe identificar origen y carácter parcial, por ejemplo: «Requisitos parciales de evidencia, derivados de 06 en `f62cf00`; no constituyen el contrato completo de recibos». Se conserva la condición: **si faltan contrato completo requerido o captura independiente, detener la acción afectada y mantener el resultado sin verificar**. La presencia del texto o de módulos nuevos no satisface por sí sola esa condición.

## Ampliación a los demás responsables

Los requisitos externos se retiran del checklist activo de Matías, sin marcarlos hechos. Estas filas se suman a D04-0501/0601–0604/0701–0702; no sustituyen sus criterios.

| ID | Origen original y responsable externo | Artefacto/disposición faltante y disparador | Criterio de cierre y prueba al retomar | Riesgo/estado |
|---|---|---|---|---|
| D04-0201 | 4.3; Giuliano, responsable de 02; Matías consumidor de 04 al llegar el insumo | Referencia/protocolo operativo aprobado y disposición de corrida conjunta normal/coordinada. Disparador: entrega identificada de 02 y condiciones de medición disponibles. | Fijar revisión, cohortes CLI/IDE, permisos, esperas y métricas ausentes; ejecutar comparación con misma rúbrica y publicar resultados sin inventar ahorro ni doble contar padre/hijos. | Deuda externa aceptada, medición no ejecutada. Matías conserva integración/ejecución consumidora al disparador; no se presenta baseline ni ahorro concluidos. |
| D04-0301 | 1.1 y 4.6; Juani, responsable de 03, más 06 para la fila de evidencia | Confirmación del responsable sobre matriz/compatibilidad y acuerdo compartido de fila 28. Disparador: revisión del paquete técnico de 04 y texto/acuerdo final de 06. | Disposición humana identificada, IDs/rutas/targets acordados y comprobación del consumo en ambos clientes cuando corresponda; preservar estrategia de Claude. | Deuda externa aceptada, aprobación no recibida. Integración técnica y tests combinados de 04 sí están entregados; no se afirma aprobación de Juani ni se implementa 03. |
| D04-010308 | 1.3; Juan Cruz (01), Juani (03), Gianfranco (08) | Disposición compartida del diff de snapshot/restauración y controles sobre configuración propia. Disparador: revisión del diff/evidencia ya entregados. | Registrar aceptación o hallazgos por responsable y retestar el ajuste pertinente: manifiesto recargado, primer snapshot, ediciones posteriores y conflicto sin pérdida. | Deuda externa aceptada, revisión humana no concluida. La aprobación de Matías del alcance ALTO y los tests técnicos no se atribuyen como aprobación de los otros responsables. |
| D04-BASE01 | 4.4; Giuliano/02 para fixtures; responsable de descargador y test WindowsTheme (TBD equipo) | Corrección o disposición de fallos conocidos de main. Disparador: revisión corregida del proveedor o decisión documentada del equipo. | Repetir tests identificados y comparar con main de la revisión integrada. API404 y tres tests principales de fixture deben dejar de fallar, o documentarse la disposición sin presentar verde la suite. WindowsTheme exige entorno/corrección pertinente. | Deuda externa aceptada. Los checks propios se ejecutaron y documentaron; no se corrige código ajeno ni se ocultan exclusiones/fallos. |
| D04-REV01 | 4.6 y proceso de integración; Juani/03 y revisores del equipo | Disposición humana final del paquete común y revisión del PR #9. Disparador: revisores evalúan la entrega y sus límites actuales. | Registrar aprobación o comentarios, aplicar correcciones propias y revalidar lo afectado. El merge es una acción posterior a la aprobación correspondiente. | Deuda externa aceptada, disposición pendiente. Paquete y evidencia de Matías entregados; no equivale a aceptación por el equipo, cierre de PR ni archivo OpenSpec. |

### Origen de las siete deudas iniciales

| IDs | Tareas originales de las que se separan |
|---|---|
| D04-0501 | 1.4, 2.5; insumo externo de captura para los casos aplicables de 4.2 |
| D04-0601–0604 | 1.4, 2.5; contrato/huella/datos/suficiencia externos para integración posterior de 04 |
| D04-0701–0702 | 1.4, 2.5 y parte dependiente del retorno final de 2.4 |

El [mapa de trazabilidad](trazabilidad-alcance-y-deuda-04-2026-11-13.md) conserva literalmente las 21 tareas originales y el reparto de las diez que estaban abiertas. Los históricos anteriores con exclusiones de 02/revisiones reflejan la autorización inicial y quedan sustituidos en el seguimiento vigente por esta ampliación.

## Trabajo propio que no se convierte en deuda ajena

Permanecen activas **1.2, 2.4 y 4.2**: sondas P2, implementación independiente de especialistas/rutas/encargo y aceptación propia con clientes reales. No se asignaron a otro responsable sin acuerdo; no disponer hoy de un entorno no demuestra que su autoría sea externa. Se registran aparte los insumos de otras fichas que requieren algunos escenarios, sin declarar completa la ejecución propia no realizada.

Las unidades sufijadas `a` en tasks son únicamente partes entregadas de tareas mixtas: texto, pruebas estructurales, conciliación técnica o paquete de revisión. No significan que se implementó el comportamiento funcional completo ni que la aprobación ajena fue recibida. El nuevo conteo es 17/20 unidades activas, con tres propias pendientes; no 17/21 tareas originales completadas.

No se acepta como hecha la implementación de 03 o de otros responsables, no se archiva el change y no se cierra el PR antes de su revisión/disposición y merge. Si una revisión ajena detecta un defecto propio de 04, corregirlo es trabajo propio y vuelve al checklist activo pertinente, no una deuda externa permanente.

## Reanudación

Cuando llegue un insumo, revisión o disposición de cualquiera de los responsables externos, compararlo con la referencia fijada, revisar sus pruebas, ejecutar en 04 el caso de la fila correspondiente y registrar árbol/revisión/resultados. Solo entonces cerrar la deuda aplicable. El seguimiento no implica implementar los changes proveedores desde esta rama ni extrapolar cobertura entre clientes o interfaces.
