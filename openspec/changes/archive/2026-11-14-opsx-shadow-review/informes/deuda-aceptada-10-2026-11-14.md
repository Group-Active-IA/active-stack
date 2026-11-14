# Deuda aceptada · opsx-shadow-review (10) · 14/11/2026

Responsable: Neyen. El change se archiva con alcance parcial: la propuesta, la spec y la preparación
externa al binario están entregadas; **el piloto no corrió**. Ningún resultado de juez, métrica ni
decisión de salida existe todavía. Nada de esta lista se marca como hecho si no lo está.

Entregadas (6): 3.1, 3.2, 3.3, 4.1, 4.2 y 4.3. Parciales (4): 1.1, 1.2, 3.4 y 4.4. Abiertas (14).

## Decisiones del grupo

| ID | Tarea | Dueño | Qué falta | Criterio de cierre | Qué rige mientras tanto |
|---|---|---|---|---|---|
| D10-0001 | 0.1 | Grupo | Proveedor, paquete de datos, acceso, métricas, tamaño de muestra, U1–U3, acuerdo mínimo, custodio y etiquetadores | Acta con cada valor y las personas designadas | Recomendaciones de D2 y D7 como propuesta; sin custodio, `NewStore` no se puede usar en una corrida real |
| D10-0002 | 0.2 | Neyen con Matías (09) | Documentación de Jev o registro de que no se evalúa | Una entrada compartida por 09 y 10 | Jev queda fuera; `CheckEntry` lo descartaría hoy por los seis motivos |
| D10-0003 | 1.1 y 1.2 | Grupo | Decidir si los fixes de revisión dentro del PR sirven como casos preparados, y el estrato del instalador en C# | Decisión registrada en `muestra-candidatos.md` | Fuera del estrato histórico |

## Muestra y etiquetado

| ID | Tarea | Dueño | Qué falta | Criterio de cierre | Qué rige mientras tanto |
|---|---|---|---|---|---|
| D10-0101 | 1.1 | Etiquetador ajeno a la tabla (TBD) | Confirmación manual por fila y columna de requisito afectado | Tabla fix → origen → archivos → requisito confirmada | Borrador SZZ; quien la armó no puede etiquetar |
| D10-0102 | 1.2 | Neyen | Defectos de lógica propios en la cantidad de 0.1 | Casos sumados a `preparados.json` con huella | Cinco casos (02 y 06) |
| D10-0103 | 1.3 | Etiquetadores (TBD) | Controles inspeccionados y fixtures limpios | Controles con inspección registrada | Ningún control entra sin inspección |
| D10-0104 | 1.4 | Etiquetadores y tercero (TBD) | Doble etiquetado, acuerdo y congelado | Commit de congelado con acuerdo calculado | Sin muestra congelada no hay corrida |
| D10-0105 | 1.5 | Neyen; revisa el custodio | Paquetes históricos desde el estado previo, sin pistas | Paquetes con huella y revisión sin pistas | — |

## Referencia, envío y calibración

| ID | Tarea | Dueño | Qué falta | Criterio de cierre | Qué rige mientras tanto |
|---|---|---|---|---|---|
| D10-0201 | 2.1 | Neyen | Referencia por caso; la columna de cobertura depende de la matriz de 08 (D08-0204) | Referencia registrada por caso | Formato D13 de 08 con cobertura «no probada» |
| D10-0202 | 2.2 | Tercer etiquetador | Medir la referencia contra etiquetas | VP/FP/FN por severidad | — |
| D10-0301 | 3.4 | Neyen | Verificar retención y entrenamiento del proveedor de 0.1 | Condiciones documentadas por separado | La compuerta de `SendLog` y `CheckEntry` bloquea el envío |
| D10-0401 | 4.4 | Neyen | Aplicar `CheckEntry` a los candidatos reales | Decisión por candidato en el informe | — |
| D10-0402 | 4.5 | Neyen | Calibración; 09 sólo entregó la parte offline | Prompt, modelo y esfuerzo fijados con conjunto propio | Sin calibración no hay corrida |

## Corrida y cierre

| ID | Tarea | Dueño | Qué falta | Criterio de cierre | Qué rige mientras tanto |
|---|---|---|---|---|---|
| D10-0501 | 5.1–5.4 | Neyen, tercer etiquetador y árbitro (TBD) | Correr, emparejar, arbitrar y calcular | Métricas de D4 por candidato | No hay métricas |
| D10-0601 | 6.1–6.2 | Neyen | Informe y salida (seguir explorando, repetir, descartar) | Informe con datos crudos y decisión | Ninguna integración de juez; la aceptación sigue su flujo |
