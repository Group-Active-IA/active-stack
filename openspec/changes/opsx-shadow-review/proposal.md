# Proposal: opsx-shadow-review

> Change `opsx-shadow-review` · ficha 10 del roadmap · Responsable: Neyen · Agentes: Claude y Codex · Prioridad P3 · Gobernanza MEDIO
> Base revisada: `1b4639b`. Depende de 08 (`opsx-native-controls`) y 09 (`opsx-model-cache-pilots`, `89f1fa5`); usa el protocolo de 02 (`240e1b2`) y toma como referencia el contrato de 06 (`opsx-verifiable-evidence`).
> Revisión 2 (10/11/2026): incorpora la auditoría técnica del 10/11 y la decisión confirmada de alcance de la primera ronda; ver `design.md`, "Cambios por la auditoría".
> Revisión 3 (13/11/2026): actualizada sobre `7a60089`. 02 ya está en `main` con el fixture de revisión sembrada (`eval/baseline/fixtures/revision/`) y la regla de emparejamiento (`eval/baseline/review`), que los casos preparados reutilizan. Borrador de la tarea 1.1 en `muestra-candidatos.md`. 08 y 09 siguen sin cambios desde el 10/11.
> Revisión 4 (13/11/2026): sobre `0005a6f` (03, 04, 05 y 06 integradas; 07 en PR #13, 08 en PR #15, 09 en PR #14). Adopta el recibo de ejecución y los tres estados de 06 en la referencia. Implementa lo que ya no depende de 08 ni de 09: muestra preparada (1.2, parcial) y paquete con examen previo (3.1, 3.2) en `eval/shadow/`; ver `design.md`, "Cambios por la integración de 03 a 06".
> Ronda `propose`: el piloto no corre y no modifica el instalador. Las tareas implementadas son preparación externa al binario (`eval/shadow/`).

## Why

La revisión de un change hoy descansa en criterios escritos (specs, escenarios, recibos de 06, controles de 08) y en una persona. No sabemos si un juez externo —un modelo o servicio que revisa el mismo change sin participar del flujo— encuentra defectos que esa revisión deja pasar, cuántas falsas alarmas agrega ni cuánto cuesta. Incorporarlo sin medir tiene dos riesgos: pagar por ruido y sacar datos del entorno sin necesidad.

### Evidencia

- **No hay medición.** El informe ejecutivo (04/11/2026) sólo dice "juez externo solo en sombra y con criterio de salida" y "Un juez opcional en sombra puede evaluarse contra reglas", y lo califica como "un diseño de experimento razonable, sin prueba local de ahorro, precisión ni conveniencia de Jev". No define el criterio de salida, ni exposición de datos, ni falsos positivos o negativos, ni costo.
- **Jev no está definido.** El informe lo nombra una vez sin descripción ni enlace. Esta propuesta no presupone qué es ni qué soporta.
- **El único reporte del repo no sirve de referencia.** `internal/verify/report.go:38` marca `Ready = Failed == 0`; los `skipped` y `warning` no bloquean y un reporte sin checks queda listo. `internal/verify/scenario_test.go` sólo prueba salud de instalación. Una referencia de revisión tiene que salir de los requisitos y recibos de 06, no de `verify`.
- **El repo es público** (`gh repo view Group-Active-IA/active-stack` → `PUBLIC`). Sus diffs ya están expuestos; lo sensible es lo que no está en el repo: salidas originales, configuración del usuario, memoria de Engram, rutas del home.
- **Hay material para una muestra real.** `main` tiene 126 commits, 33 de ellos `fix`, y 25 changes archivados en `openspec/changes/archive/`: defectos que el flujo dejó pasar y se corrigieron después.
- **El roadmap fija el límite.** Una integración permanente del juez requiere una decisión de arquitectura separada y el arnés de evaluación queda fuera del binario (`CHANGES.md`, Límites compartidos).

## What Changes

- **Piloto en sombra, primera ronda acotada.** Sólo casos históricos y preparados, con material público revisado y datos ficticios; sin changes vivos ni evidencia real (decisión confirmada). Los hallazgos del juez se guardan con huella en un almacén de acceso restringido a un custodio, que los abre sólo cuando el caso está cerrado definitivamente.
- **Muestra etiquetada.** Estados previos a fixes conocidos (sin el fix ni pistas que revelen la respuesta), defectos sembrados (incluidos los casos de rechazo de 06) y controles inspeccionados o fixtures limpios. Dos personas etiquetan de forma independiente y un tercero resuelve; las etiquetas se congelan.
- **Revisión de referencia.** La revisión con criterios del flujo: requisitos y escenarios del change, recibos de ejecución de 06 con su estado por requisito (verificado, rechazado, sin verificar) y resultados de controles de 08, más la persona que acepta.
- **Métricas.** Contra etiquetas congeladas: verdaderos positivos, falsos positivos, falsos negativos; señal adicional como cantidad de defectos mayores que sólo el juez encontró; proporción de controles con falsa alarma; costo por caso en tokens, dinero o cuota, latencia y triage. Los hallazgos sin etiqueta van a un arbitraje aparte, ciego al origen.
- **Minimización de datos.** Lista de inclusión, examen previo de todo el paquete (diffs públicos incluidos), registro de envío y condiciones de retención y entrenamiento del proveedor verificadas por separado.
- **Candidatos.** Al menos un juez de base y Jev como candidato opcional, sólo si cumple criterios de entrada documentados.
- **Criterio de salida.** Cerrar, repetir o descartar según umbrales acordados, con un informe de decisión. Los resultados no entran en la aceptación de ningún change durante el piloto.

## Capabilities

### New Capabilities

- `shadow-review-pilot`: muestra etiquetada, revisión de referencia, ejecución en sombra con resultados sellados, métricas, candidatos y criterio de cierre.
- `review-data-exposure`: clasificación de datos, paquete de revisión minimizado, redacción, huellas y condiciones del proveedor.

### Modified Capabilities

(ninguna)

## Alcance

- Diseño del piloto, de la muestra y del etiquetado.
- Definición de la revisión de referencia a partir de 06 y 08.
- Métricas, emparejamiento de hallazgos y contabilidad de costo con el protocolo de 02/09.
- Política de datos que puede recibir el juez.
- Criterios de entrada de candidatos, incluido Jev.
- Criterio para cerrar, repetir o descartar.

## Exclusiones

- Integración permanente del juez, ledger firmado o runtime nuevo: decisión de arquitectura separada.
- Cualquier cambio en `cmd/`, `internal/` o en la configuración que instala el binario.
- Usar el veredicto del juez para aceptar o rechazar changes durante el piloto.
- Routing de modelos con Jev (09): esta ficha evalúa sólo el caso de revisión.
- Contrato de recibos (06) y controles nativos (08): se consumen.
- Implementación y corridas: etapa posterior.

## Impact

- **Código del instalador**: ninguno. El arnés vive fuera del binario.
- **Datos**: define qué sale del entorno; requiere acuerdo del grupo antes de la primera corrida.
- **Gobernanza**: MEDIO por exposición de datos, no por riesgo sobre la configuración del usuario.
- **Otras fichas**: consume requisitos y recibos de 06, la tabla de cobertura de 08 y la contabilidad y elección de modelos de 09.
