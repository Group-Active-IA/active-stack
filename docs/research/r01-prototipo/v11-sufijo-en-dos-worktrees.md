# V11 · Dos ramas cortas que producían la misma clave

**Resultado del piloto del 19 de noviembre de 2026.** Con un sufijo calculado sobre el nombre Git completo, `feat/una` y `feat-una` conservaron dos observaciones y dos historiales al guardar y actualizar una misma clave base. El control anterior con el prototipo original había actualizado una sola fila. La variante se ejecutó en una copia de laboratorio. [Control con la regla original](controles-codex.md)

## Regla ensayada

Para una rama etiquetable se usaron hasta 27 caracteres del prefijo legible y las primeras 12 cifras hexadecimales de SHA-256 del nombre Git en UTF-8, incluyendo mayúsculas y separadores originales. Si el prefijo legible resulta vacío se usa `rama`. La etiqueta completa tiene un máximo de 40 caracteres. La rama principal conserva la política existente. Los 12 caracteres del hash son un parámetro del experimento.

```python
stem = re.sub(r"[^a-z0-9]+", "-", branch.lower()).strip("-")[:27].rstrip("-") or "rama"
slug = stem + "-" + hashlib.sha256(branch.encode("utf-8")).hexdigest()[:12]
```

## Evidencia del cliente y de la base

macOS x86_64, Codex CLI 0.169.0-alpha.11.1, Engram 3.6.2 y Git 2.50.1. Dos procesos reales de `codex exec -C <worktree> --json` compartieron el mismo repositorio, el mismo proxy stdio y la misma base temporal. El proveedor local sintético emitió cuatro llamadas `mem_save`: una creación por rama y una actualización por rama. Todas usaron el proyecto `r01-fixture`, el scope `personal`, el tipo `decision` y la clave base `base-002`.

| Rama Git | Prefijo observado en título y clave | ID de Engram | Sesión observada en Codex y en la fila |
|---|---|---:|---|
| `feat/una` | `feat-una-1e8539184bc0` | 1 | `id-local-027` |
| `feat-una` | `feat-una-172f1270e0d4` | 2 | `id-local-028` |

Para cada llamada se compararon la emisión sintética, los argumentos despachados por Codex, la entrada del proxy con el directorio del worktree, la salida del proxy, el ID devuelto por Engram, la fila SQLite y `mem_get_observation`. En cada fila, `topic_key` fue `branch/<prefijo observado>/<clave base>`; el título comenzaba con el prefijo entre corchetes y el contenido incluía la línea `Branch: <rama Git> · Author: Persona Piloto`. El laboratorio registró previamente en Engram el ID auténtico de cada evento `thread.started` y el emisor sintético lo pasó en sus dos llamadas. Esta prueba verifica su conservación. La vinculación automática del plugin se ensayó por separado. [Coexistencia y sesiones del proxy](codex-proxy-b-worktrees.json)

| Punto de lectura | ID 1, revisión | ID 2, revisión | Comprobación |
|---|---:|---:|---|
| Tras liberar las dos escrituras iniciales | 1 | 1 | Ambos clientes estaban vivos antes de liberar la primera barrera. |
| Después de actualizar solo `feat/una` | 2 | 1 | La fila completa del ID 2 y su lectura MCP coincidían con la captura inicial. |
| Después de actualizar solo `feat-una` | 2 | 2 | La fila completa del ID 1 y su lectura MCP coincidían con la captura anterior. |

SQLite conservó, para cada ID, la versión histórica 1 con el contenido de la primera llamada de su propia rama. Los IDs devueltos por creación y por actualización fueron idénticos dentro de cada sesión. La barrera aseguró concurrencia de sesiones; el orden de las actualizaciones permitió comprobar la ausencia de cambios cruzados.

## TDD de la variante

| Conducta | Prueba | Base | RED | GREEN y triangulación |
|---|---|---|---|---|
| Distinguir barra y guion, manteniendo título y clave coherentes | `test_slash_dash_and_title`, unidad sobre Git/JSON | La suite del prototipo publicado había pasado antes de copiarlo. | Con la regla original se obtuvo `feat-una` en lugar de `feat-una-1e8539184bc0`. | Ambas ramas dieron sus dos etiquetas esperadas. |
| Longitud máxima, nombres sin caracteres ASCII y rama principal | `test_short_unicode_long_and_primary`, unidad sobre Git/JSON | Primer caso verde. | La versión intermedia generó `-a567bdaa1136` para `中`. | Se limitó el prefijo, se aplicó el fallback y las dos pruebas pasaron. También se comprobó `main` sin etiqueta. |

## Compatibilidad y alcance

Agregar el sufijo cambia las claves de las ramas etiquetadas, incluidas las ramas cortas que antes no colisionaban. El sufijo de 48 bits deja una posibilidad residual de colisión. El ensayo concurrente verificó el par indicado en Codex/macOS. Windows, Claude y Cloud quedan fuera de esta prueba.

### Compatibilidad de historial por la API pública

Ensayo adicional con Engram MCP 3.6.2 en una base temporal vacía del proyecto sintético `r01-historia`, scope `project`, tipo `discovery` y autosync desactivado. Se realizaron todas las escrituras con `mem_save` y `mem_update`; se cotejaron los resultados completos de `mem_get_observation(id, include_history=true)`. Los IDs de esta sección pertenecen a esta nueva base.

| Origen | Operaciones | Evidencia |
|---|---|---|
| Una sola rama conocida, destino nuevo | Dos `mem_save` desde `feat/una` con `branch/feat-una/unico`. Luego `mem_update(id=1, expected_project="r01-historia", topic_key="branch/feat-una-1e8539184bc0/unico")`. Finalmente otro `mem_save` con la clave nueva, el título etiquetado y contenido diferente. | Los primeros guardados devolvieron #1, revisión 2. Cambiar solo la clave mantuvo #1 y el título previo, y elevó la revisión a 3. El guardado siguiente devolvió #1, revisión 4, con el nuevo título. Desde la misma fila se recuperaron el contenido actual `three` y los dos contenidos previos `one` y `two`. El cambio exclusivo de clave incrementó el contador sin añadir una instantánea de contenido. |
| Dos ramas ya compartían una fila | `feat/una` y `feat-una` guardaron con `branch/feat-una/mixto`. Se leyó el estado completo de #2. Cada rama creó su propia observación bajo la nueva clave; se actualizó cada una, leyendo también la fila ajena antes y después. | Los dos guardados antiguos devolvieron #2, revisión 2, con `old-dash` actual y `old-slash` en el historial. Cada rama obtuvo un ID nuevo (#3 para `feat/una`, #4 para `feat-una`) y pasó de revisión 1 a 2. Actualizar #3 conservó #4 y viceversa. La lectura completa de #2 con historial coincidió antes y después de las cuatro escrituras. |

En el primer caso la procedencia se conocía por construcción y la clave destino estaba libre. Q4 mostró por separado que mover una clave a un destino ya ocupado puede dejar dos filas con la misma clave. En el segundo caso se preservó la fila histórica compartida sin reasignar sus revisiones. El ensayo demuestra continuidad individual hacia adelante en #3 y #4; no distribuye automáticamente el pasado de #2 entre ambas ramas. [Control de colisión con destino ocupado](alcance-q5-q7-q8.md#qa-macos-de-la-entrega)
