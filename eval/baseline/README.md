# eval/baseline

Arnés de medición de referencia del change `opsx-baseline` (ficha 02).

- **Externo al binario; no importado por `cmd/` ni `internal/`.** Decisión D6: directorio de evaluación separado, en Go, solo biblioteca estándar, sin tocar `go.mod`.
- **Nombre del directorio: `eval/baseline/` (TBD · renombrable).** No está decidido; renombrarlo solo exige mover la carpeta, porque nada fuera de `eval/` lo importa.

## Paquetes

| Paquete | Tarea | Contenido |
|---|---|---|
| `manifest` | 1.1 | Manifiesto versionado (JSON), matriz de cohortes con tareas y pesos, validación previa por celda (`ejecutable` / `incompatible` con motivo), plan con celdas fuera del denominador y enum de estados de corrida. |
| `review` | 1.4 | Regla de emparejamiento de la revisión de un diff sembrado (verdadero positivo, duplicado, falso positivo, hallazgo nuevo a adjudicación, omisiones). Desde la tanda 2 rechaza con error los defectos con archivo o mecanismo vacío y las disposiciones desconocidas (DEUDA 6). |
| `fixture` | 1.3 | Carga y validación de fixtures (huella del prompt, huella del estado inicial, criterios obligatorios con chequeo conocido), checkout descartable en un temporal con `Reset`/`Close`, evaluación por criterio y veredicto (`aceptada`, `rechazada_por_calidad`, `pendiente_de_revisor`). |

## Fixtures (`fixtures/<familia>/`)

Cada familia tiene `task.json` (id, familia, versión de rúbrica, fuente congelada con huella, prompt con huella sha256, herramientas permitidas, límites, criterios), `prompt.md`, la clave cuando la familia se evalúa por hechos, y una referencia y una variante defectuosa.

| Familia | Fuente | Referencia / defectuosa |
|---|---|---|
| `exploracion` | `git archive` de `1b4639b537a4d0c04dc8c25c0220732095a772d3` | respuestas JSON |
| `cambio-pequeno` | mini-proyecto Go sintético `project/` | parches (`git apply`) |
| `cambio-multiarchivo` | mini-proyecto Go sintético `project/` | parches |
| `diagnostico` | mini-proyecto Go sintético con falla sembrada | respuestas JSON |
| `revision` | la misma revisión + `seeded.patch` | respuestas JSON con hallazgos |

- Los mini-proyectos tienen su propio `go.mod`, así que `go test ./...` del repo no los incluye (el de diagnóstico falla a propósito). Las pruebas ocultas viven en `hidden/` con sufijo `.go.txt` y el evaluador las copia sin el sufijo.
- La huella de árbol es el sha256 de las líneas `sha256sum` ordenadas por ruta: `find . -type f | sed 's|^\./||' | LC_ALL=C sort | xargs sha256sum --text | sha256sum`.
- Cambiar un prompt o un proyecto exige recalcular su huella en `task.json`: es congelar una versión nueva del fixture.
- Los límites figuran `(TBD · D3)` hasta que se acuerden.

## Pruebas

```
go test ./eval/baseline/...
```

Los datos de los tests son sintéticos. No hay adaptadores de captura de clientes reales todavía (2.2 y 2.3).

## Arnés (tanda 3, parte A)

- `run/` (2.1): `Executor.Execute` prepara un checkout descartable con `fixture`, llama a un `Runner` (interfaz; el cliente real llega con 2.2), evalúa y reintenta desde el fixture reiniciado según `Policy`. Cada intento guarda estado (`manifest.RunStatus`; `pendiente_de_revisor` no es final), causa, código de salida bruto y los eventos `turn.failed`/`error` como JSON bruto, sin semántica asumida. `ExecutePlan` corre solo las cohortes ejecutables de `manifest.Plan`.
- `trace/` (2.4): lecturas y llamadas a herramientas por tipo y rol; una señal que la interfaz no expone es `no disponible`, no cero.
- `timing/` (2.5): reloj monotónico inyectable, esperas de aprobación aparte, unión de intervalos solapados; el ajustado es no disponible si alguna espera no se puede delimitar.
- `checkout_intacto` compara la huella de archivos y además rechaza directorios vacíos agregados y entradas no regulares (symlinks): la huella solo cuenta archivos regulares, así que esas adiciones se detectan aparte.
- Duración: el paquete `fixture` tarda unos 100 s en Windows; el tiempo es casi todo `os.Open` de ~650 archivos por cada `Reset`/`TreeHash` (antivirus) y `go test` de los mini-proyectos. Las pruebas pesadas corren en paralelo (`t.Parallel`); no se recortó la verificación de la huella.

## Consumo, registros y resumen (tanda 3, parte B)

- `run/`: si el llamador cancela durante un intento, la corrida parcial conserva ese intento con estado `interrumpida_por_llamador` (no final), causa, código y eventos; `Execute` devuelve además el error del contexto.
- `usage/` (2.6): `FromClaude` (entrada lógica = entrada + lectura + creación de caché) y `FromCodex` (la entrada cacheada ya está en la entrada; sin caché por diferencia). `Probe` separa la entrada de una solicitud del total del turno; el total solo se usa como aproximación identificada si se comprobó que fue una única solicitud, y el estado de caché observado nunca se reescribe para coincidir con el previsto. `Reconcile` aplica la precedencia de design.md (60.000 en el ejemplo de Claude, 48.000 en el de Codex): con el agregado comprobado como inclusivo no se suman padre e hijos; detalle menor = sin atribuir y parcial; detalle mayor = ambas cifras y fuera de cobertura completa.
- `record/` (2.7): `Builder.Records` emite un registro por intento y uno por corrida (`WriteJSONL`), con cobertura por señal y conteo de redacciones por categoría. Lista sensible mínima aplicada (la lista completa es de 05): correo, UUID de cuenta, id de usuario, rutas del host, detalle de herramientas y cuerpos de solicitud. Los eventos de error se conservan como JSON con esas claves y textos redactados; un evento que no es JSON válido se reduce a su tipo. Huellas: `sha256` con el algoritmo en el campo (definición de 06). La referencia a evidencia local se conserva (una ruta absoluta se reduce a `local:<archivo>`) o se registra su ausencia con motivo. Los UUID de sesión y agente son identificadores opacos y se conservan.
- `summary/` (3.1): `Compute` y `Markdown`. El cálculo está documentado en el comentario del paquete y en la sección "Cálculo" del Markdown: el costo total suma todos los intentos de todas las corridas iniciadas; el cociente por aceptada queda sin valor con cero aceptadas o cobertura incompleta; las incompatibles quedan fuera y se listan aparte; las pendientes de revisor figuran aparte y dejan el resumen provisional; cada cohorte (cliente, versión, interfaz real, modalidad, variante) es una fila propia.
