# Deudas de 05/06/07 aceptadas para la entrega de 04

13 de noviembre de 2026. **Aceptación: Matías, instrucción explícita en esta conversación de marcar como deuda lo que no pueda avanzar sin 05/06/07.** La autorización permite registrar esos pendientes y publicar el avance de 04 para revisión. No aprueba en nombre de Emiliano, Gianfranco, Neyen o el grupo un contrato técnico propuesto; tampoco declara terminado el change ni mergeable el PR.

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

## Qué no queda aceptado como deuda por esta autorización

- Sondas propias P2, especialistas, herencia, lectura real de módulos y aceptación CLI/IDE Windows/Linux de 04 (tareas 1.2, 2.4 y 4.2, salvo el insumo externo concreto registrado arriba).
- Comparación con 02, resultados de ahorro y su baseline (4.3).
- Revisión humana de snapshot, matriz y migración, ni comprobaciones de integración y regresiones propias de 04 (1.1, 1.3, 4.4 y 4.6).
- Defectos propios de 04 o fallos preexistentes de main: se registran y disponen por su circuito, no se convierten en dependencias 05/06/07.
- Deuda de 03, implementación de otros responsables, archivo del change, aceptación por el equipo o merge automático del PR.

Las tareas mixtas 1.4/2.5 pueden tener una parte implementada y otra sujeta a este registro. Un checkbox completo no sustituye esa separación ni la evidencia. Este documento no modifica su estado; el integrador debe reflejar alcance y referencias en los artefactos que mantiene, sin ocultar los requisitos restantes.

## Reanudación

Cuando llegue una revisión nueva de 05/06/07, comparar el contrato con la revisión fijada, revisar su disposición humana y pruebas, ejecutar en 04 el caso de la fila correspondiente y registrar árbol/revisión/resultados. Solo entonces cerrar la deuda aplicable. El seguimiento no implica implementar los changes proveedores desde esta rama ni extrapolar cobertura entre clientes o interfaces.
