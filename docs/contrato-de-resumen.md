# Contrato del resumen de salidas de comando

> Change `opsx-task-context`, tarea 2.11 · capacidad `bounded-command-output` · `design.md` — D3, D8, D9 y D20.
> El ejemplo de este documento es el fixture `fail-middle` de `eval/taskcontext/fixtures` y su resumen
> esperado; un test (`eval/taskcontext/summary`) falla si el ejemplo y el resumen esperado se desvían.
> El texto que el agente recibe vive en el módulo `bounded-command-output`.

## Cuándo se resume

Una salida de más de **300 líneas** o de más de **10.000 caracteres**, lo que ocurra primero, se entrega
como resumen más referencia. Es una condición de entrada, no un tope de recorte: el umbral decide
*cuándo* resumir, nunca *cuánto* cortar. Los dos números salen del catálogo
(`internal/catalog/taskcontext.yaml`, `output_threshold`). Una salida más corta se entrega entera.

## Qué conserva el resumen

- Cada fallo que la salida registra y el código de salida. Un código distinto de cero es siempre un fallo.
- Los fallos se identifican con los patrones por herramienta del catálogo (`failure_patterns`): para
  `go test`, `--- FAIL`, `FAIL` + tabulador y `panic:`; para `go build` y `go vet`,
  `<archivo>:<línea>:<columna>:`. Con código de salida cero, una línea con forma de error **no** es un
  fallo. Una herramienta sin patrones declarados se evalúa solo por código de salida y el resumen lo dice
  con la línea `failures identified by exit code only`.
- Se descarta la repetición sin fallos. Nunca se descarta una línea de fallo para ahorrar espacio.

## Las líneas de forma fija

```text
exit: <code>
failure: <line as captured>
ref: cmd=<command> run=<run-id> path=<redacted copy> sha256:<hex> expires=<YYYY-MM-DD>
capture: <harness | client:<source> | agent> · redactor: <same set> · producer: <who wrote this summary>
status: unverified — <what is missing>
partial — unverified: <reason>
```

Sin fallos, la línea es exactamente `failure: none`, sin agregados; cada `failure:` es la línea
capturada tal cual. El evaluador ignora el BOM y los CRLF del archivo del resumen (PowerShell los
escribe), nunca los de la copia, cuya huella es de bytes.

La referencia lleva cinco campos y la huella nombra el algoritmo. El estado `status: unverified` es
obligatorio cuando la captura es del agente: un resultado declarado por el agente queda sin verificar
hasta que otra parte corrobore la ejecución con captura independiente (D20).

## Ejemplo: entrada y salida

**Entrada.** El comando `go test -v ./fixture/pkg/` del fixture `fail-middle` imprime 3.000 líneas de
relleno, con `--- FAIL: TestNeedle_fail_middle (0.00s)` en la línea 1.500 y `FAIL	fixture/pkg	0.512s` al
final, y sale con código 1. Un cliente que recorta por cabeza y cola pierde la línea 1.500.

**Salida.** El resumen conforme conserva el fallo central y el código, y lleva la referencia de la copia
redactada que guardó el arnés:

```text
exit: 1
failure: --- FAIL: TestNeedle_fail_middle (0.00s)
failure: FAIL	fixture/pkg	0.512s
ref: cmd=go test -v ./fixture/pkg/ run=20261113-example path=<cache>/active-stack/runs/_sin-change/20261113-example/output.redacted.txt sha256:64988d51744504333f91c0ff0ba75e4f059f476b34a5b457166eccf5b712e073 expires=2026-11-20
capture: harness · redactor: harness · producer: agent
```

## Cómo lo evalúa el arnés

El evaluador (`eval/taskcontext/summary`) devuelve uno de tres veredictos:

| Veredicto | Cuándo |
|---|---|
| **aceptado** | conserva cada fallo y el código, no inventa fallos, lleva la referencia de cinco campos con huella coincidente y declara capturador, redactor y productor |
| **insuficiente** | no contradice la salida pero falta información: referencia incompleta, capturador, redactor o productor sin declarar, captura del agente sin `status: unverified`, herramienta sin patrones sin la frase correspondiente |
| **rechazado** | contradice la salida conocida: fallo omitido o inventado, código de salida distinto, huella distinta a la de la copia, o un valor sensible sin redactar. El rechazo registra el productor declarado |

Un original **nunca capturado** se entrega `partial — unverified: never captured`, sin ruta ni huella.
Los otros motivos válidos son `copy missing`, `expired on <fecha>` y `no persistent disk`. Pedir su
original informa «nunca capturado», no «vencido».
