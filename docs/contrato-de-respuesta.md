# Contrato de respuesta del agente

> Change `opsx-task-context`, tarea 7.3 · capacidad `concise-agent-reporting` · `design.md` — D16.
> El ejemplo de este documento es el par de fixtures de `eval/taskcontext/replyfix` (`single-finding`);
> un test falla si el documento y el fixture se desvían. La rúbrica que lo evalúa es la de ese paquete
> y se aplica a respuestas reales en la tarea 8.10.

## Por qué es una capacidad aparte

Una salida de comando tiene un original completo que se conserva y se consulta por referencia. Una
respuesta del agente **no tiene original**: nunca existió fuera de lo que el modelo escribió. Lo que se
puede perder en ella no es una línea de fallo sino el hallazgo, enterrado en el relato de cómo se llegó.

## Las reglas

1. **Primera línea: el resultado, con su procedencia** (`path:line` o `path#symbol`).
2. **Segunda línea: la consecuencia o el próximo paso.** No el relato.
3. **Varios hallazgos:** una lista, cada ítem con su procedencia, antes de cualquier explicación.
4. **El detalle se produce a pedido.** No se cuenta cómo se llegó al resultado si no se pidió; cuando se
   pide, se produce en ese momento, desde la misma procedencia.
5. **Lo que ya está en un artefacto se señala, no se copia:** se nombra el artefacto y la sección.
6. **Ningún hallazgo:** una línea que dice que no se encontró nada y qué se revisó. Sin narrar la búsqueda.
7. Las líneas de forma fija de los otros módulos (`full read:`, `unnecessary call:`, `ref:`,
   `partial — unverified:`, `unavailable:`, `out-of-class use:`) se conservan tal cual.

## El par de ejemplo: el mismo hallazgo, dos respuestas

Con el **resultado primero** (cumple la rúbrica):

```text
Compose rejects an unknown toggle with an error: internal/harness/config/compose.go:35
Next: add the toggle to knownToggles before wiring a new module.
```

Con el **recorrido primero** (no la cumple: la primera línea es relato y el resultado llega en la tercera):

```text
I started by listing the config package and opened compose.go. Then I read the first 60 lines and followed the toggle loop.
After that I checked how the error is built and where the known toggles are declared.
Compose rejects an unknown toggle with an error: internal/harness/config/compose.go:35
Next: add the toggle to knownToggles before wiring a new module.
```

Las dos describen el mismo hallazgo con la misma procedencia (`internal/harness/config/compose.go:35`);
solo cambia el orden. Quien lee solo la primera línea de la segunda respuesta no se entera del hallazgo.

## Los otros casos de los fixtures

| Caso (fixture) | Qué verifica la rúbrica |
|---|---|
| `multiple-findings` | lista de tres hallazgos, cada uno con `path:line`, antes de cualquier explicación; después solo el próximo paso |
| `no-finding` | una sola línea: no hay hallazgo y qué se revisó |
| `already-in-artifact` | señala `openspec/changes/opsx-task-context/design.md#D3` en lugar de copiar el párrafo |

## Qué no cubre

La rúbrica es mecánica: mira el orden, la procedencia y el relato. No juzga si el hallazgo es correcto ni
si la consecuencia es útil. La conducta real del agente se verifica en la tarea 8.10 con respuestas reales
sobre estos mismos casos.
