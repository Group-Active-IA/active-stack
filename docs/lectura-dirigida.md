# Lectura dirigida y procedencia

> Change `opsx-task-context`, tarea 4.7 · capacidad `directed-code-reading` · `design.md` — D11, D13 y D15.
> Cada ejemplo documentado corresponde a un fixture de `eval/taskcontext/readfix` (marcado como
> `fixture: <id>`); un test falla si un ejemplo no tiene fixture o si un fixture queda sin documentar.
> El texto que el agente recibe vive en el módulo `directed-code-reading`.

## Cuándo corresponde cada modo de lectura

| Situación | Modo | Fixture |
|---|---|---|
| El archivo tiene 300 líneas o menos | lectura completa | `fixture: full-read-under-threshold` |
| La tarea va a reescribir el archivo entero | lectura completa | `fixture: full-read-rewrite` |
| El archivo es el artefacto que la acción activa produce o edita | lectura completa | `fixture: full-read-artifact` |
| Una lectura acotada previa de ese archivo devolvió vacío | lectura completa | `fixture: full-read-after-empty-bounded` |
| Ningún rango cubre lo que la tarea necesita | lectura completa **con motivo declarado** | `fixture: fifth-condition-with-reason` |
| Cualquier otro archivo de más de 300 líneas | por ruta y rango de líneas, o por símbolo | `fixture: full-read-unjustified` (el contraejemplo) |
| Dónde se usa un símbolo | búsqueda y `path:line` por coincidencia, sin abrir cada archivo | `fixture: orientation-by-search` |
| Dónde vive una capacidad | una búsqueda acotada sobre el repositorio; rutas candidatas ordenadas por tipo | `fixture: walk-file-by-file` (el contraejemplo) |

## El umbral: 300 líneas, medido

El valor vive en el catálogo (`internal/catalog/taskcontext.yaml`, `read_threshold`: unidad líneas,
valor 300). Sale de medir los archivos `.go` versionados bajo `internal/` y `cmd/` que no terminan en
`_test.go` (130 en `1b4639b`), con `wc -l` y percentiles por rango más cercano:

| | líneas |
|---|---|
| p50 | 85,5 |
| p75 | 178 |
| p85 | 244 |
| p90 | 311 |
| p92 | 375 |

A 300 se leen enteros 116 de 130 archivos (89 %) y la lectura acotada queda para los 14 grandes. A 400 la
regla tocaría solo seis archivos; a 200 forzaría lectura acotada donde leer entero sale más barato. El test
`TestReadThreshold_BetweenP85AndP92` (en `internal/catalog`) falla si 300 deja de estar entre el p85 y el
p92; entonces se elige el valor nuevo con el mismo criterio. El umbral de lectura de archivos y el de
salidas de comando son dos reglas distintas, aunque compartan las 300 líneas.

## Las cinco condiciones de lectura completa (lista cerrada)

1. el archivo no supera el umbral;
2. la tarea va a reescribir el archivo entero;
3. el archivo es el artefacto que la acción activa produce o edita;
4. una lectura acotada previa sobre ese archivo devolvió vacío;
5. los rangos no alcanzan, y el motivo queda en el reporte: `full read: <path> — <reason>`.

Cualquier otra lectura completa no está justificada. Una quinta sin motivo cuenta como no justificada
(`fixture: fifth-condition-without-reason`). El arnés cuenta aparte las lecturas que entran por la
quinta, con sus motivos: es el insumo para recalibrar el umbral.

## Lectura por símbolo

Una lectura por símbolo devuelve el símbolo completo, de su primera a su última línea, más la **firma de
la declaración que lo contiene** cuando está anidado. Nada más: ni el resto del cuerpo del contenedor ni
los símbolos vecinos (`fixture: nested-symbol`, `readfix.NestedSymbol`).

## Procedencia y relectura

Toda afirmación sobre el código lleva su procedencia: `path:line` o `path#symbol`. Una afirmación sin
procedencia no cuenta como evidencia. Releer el mismo archivo y el mismo rango sin una escritura en el
medio se registra con `unnecessary call: <tool> <input>` (`fixture: reread-without-write-recorded`; sin
el registro, `fixture: reread-without-write-unrecorded`); después de una escritura la relectura es
legítima (`fixture: reread-after-write`).

## Orden de las rutas candidatas

Una definición pesa más que cualquier cantidad de referencias, y una referencia más que cualquier
cantidad de coincidencias de texto; la cantidad solo ordena dentro del mismo tipo (`readfix.Rank`). Cada
candidato se explica en una línea y no se abren los archivos descartados.
