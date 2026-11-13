# Deuda aceptada de 05 · `opsx-task-context`

13 de noviembre de 2026. Responsable de la 05: Emiliano. Sigue la convención de 03, 04, 06 y 07 y el acuerdo del grupo: lo que depende de otros changes o de corridas que la 05 no puede cerrar sola sale del checklist activo. Queda registrado con su criterio de cierre y **no se marca como hecho**. Se resuelve en el change común de deuda.

Base: `origin/main` `d5e95bd` (03 y 04 integradas y archivadas), traído a la rama en `c6eb351`.

## Estado del change

El change **queda sin archivar**: el grupo archiva cuando todas las tareas están completas y verificadas, y quedan 10 en deuda. Cuando se archive, las specs de la 05 se sincronizan a `openspec/specs/` y describen el comportamiento completo objetivo. Lo que sigue en este registro **no está entregado** aunque la spec lo describa. Mientras tanto rige la columna "Comportamiento mientras tanto".

## Lo entregado y verificado (no es deuda)

- **Tareas marcadas en `tasks.md`.** Entre ellas: los cuatro textos de módulo y la tabla acción → módulos; el contrato de captura y resumen; el capturador, el redactor y las recetas; los adaptadores de fuente y el comparador de 3.6 con muestras reales; la matriz D22 en código; los fixtures de los grupos 2, 4 y 5; las corridas 8.2 a 8.5, 8.10, 9.1 y 9.2.
- **8.1, medido el 13/11** (n = 1 por celda, fixtures `fail-middle` y `stderr-only`):
  - Claude CLI y Codex CLI, en Windows y Linux.
  - Codex App Server, en Windows y Linux.
  - 3.3 y 3.8 en Linux.
  - Registrado en `eval/taskcontext/clientsrc/matrix.go` y en `design.md` (D22).
- **X12 (`entrega-03-04.md` §5):** Codex lee los módulos desde la skill instalada.

Verificación del 13/11 sobre la rama con `main` traído:
- `go vet ./...` sin errores.
- `go test ./...` con los mismos FAIL que `main`: 9 en Linux, de `cmd/active-stack` y `eval/baseline/fixture`; 6 en Windows, de permisos y symlinks de la 04, D04-BASE01.
- `openspec validate opsx-task-context --strict` válido.

## Registro de deuda

**Estado de todas las filas:** deuda aceptada por Emiliano el 2026-11-13. Ninguna fila se cierra por existir una propuesta o un informe.

| ID | Tarea | Dueño | Qué falta | Criterio de cierre | Comportamiento mientras tanto |
|---|---|---|---|---|---|
| D05-0605 | 6.5, 6.6, 6.7 | Change de deuda; toca código de la 03 (`ComposeLayers`) | Integrar en la skill de Claude los cuatro módulos de la 05 y la tabla de 6.2 (un módulo sin toggle hoy se descarta), con la regla universal D21 y la referencia de invocación. El diff y los tests están en `entrega-03-04.md` §8. | Diff de §8 integrado sobre `main` y tests de 6.5 a 6.7 en verde en Claude y Codex | Claude **no recibe** los cuatro módulos en `/sdd-orchestrator`. Codex sí los recibe, por el paquete de la 04, con textos viejos (ver D05-0610) |
| D05-0610 | — (nueva) | Change de deuda; toca código y tests de la 04 | El paquete de Codex en `main` trae **copias** de los cuatro módulos en `assets/codex/native/modules/`, fijadas a la revisión `e4ec0f5`. `codex_context_modules_test.go` compara su sha256 y `action-table.md` cita esa revisión. Las cuatro difieren de los textos vigentes de la 05 (`assets/modules/`, iteración v3). | `ComposeCodexPackage` lee los textos de `assets/modules/` (una sola copia), o las copias se actualizan con un test que exija igualdad byte a byte con `assets/modules/` | Codex recibe la versión anterior de los textos, sin las correcciones de G1, G7, G8 y G6 |
| D05-0801 | 8.1 (celdas de IDE) | Emiliano | Claude Code en la extensión de VS Code y en la app de escritorio (Windows): declarada por la documentación (*"fires the same hook events wherever it runs: … IDE extensions, the Desktop app"*), sin corrida. IDE en Linux: no hay IDE en WSL2. La extensión y el escritorio no envían un pedido sin una persona: el link `prompt`/`q` solo lo precarga. | Dos pedidos enviados a mano (`fail-middle` y `stderr-only`) con los hooks de la carpeta de prueba, comparados con 3.6 y registrados en `matrix.go` | La matriz declara esas celdas. Ninguna captura del IDE cuenta como independiente; rige la del arnés |
| D05-0806 | 8.6 | Emiliano | En Claude, la lectura por símbolo dio 0 de 2: abre el archivo entero. Las corridas de Codex quedan sin hacer. | Lectura por símbolo aceptada en Claude y Codex con `readfix.Evaluate`, o límite declarado por el grupo | El módulo `directed-code-reading` rige como texto. La lectura por símbolo no está demostrada |
| D05-0807 | 8.7 | Emiliano | En Claude no se escriben `unavailable:` ni `out-of-class use:`. Las corridas de Codex quedan sin hacer. | Registros aceptados por `toolfix.Evaluate` en Claude y Codex | El módulo `task-tool-selection` rige como texto. Esos registros no están demostrados |
| D05-0808 | 8.8 | Change de deuda (03) y D05-0605 | Después de `/compact`, el coordinador no relee los módulos. El texto universal de la 03 le pide invocar una skill que tiene `disable-model-invocation: true`. Además, `apply` sin `delegation` no escribe la línea fija. | Texto universal ajustado (propuesta a la 03) y corridas de 8.8 aceptadas | Tras compactar, la persona tiene que reinvocar `/sdd-orchestrator` |
| D05-0809 | 8.9 | Emiliano, tras D05-0605 | Comparar una sesión normal con una coordinada, en instalación limpia y en actualización, en cada cliente | D05-0605 cerrada y corridas por cliente registradas | Sin evidencia de qué módulos recibe cada sesión después de actualizar |
| D05-0903 | 9.3, 9.4 | Emiliano, tras D05-0605 | Correr con y sin cada módulo y medir el costo de las reglas universales de memoria. Hoy está medido solo sobre la integración propuesta. | D05-0605 cerrada y las dos corridas sobre `main` | La matriz común conserva sus filas 20 y 22 sin el insumo de la 05 |

## Respuesta a la deuda que otros changes registraron a nombre de la 05

- **D07-0002e (07: nombre del disparador «la acción requiere coordinación de subagentes»).** La 05 lo confirma: el disparador es la fila `delegate (any action)` de la tabla acción → módulos, que ya usan la 03 y la 04. No hay nombre nuevo.
- **D04-0501 (04: captura independiente operativa).** La fuente por interfaz quedó medida en 8.1:
  - Codex CLI (`aggregated_output`): completa en Windows y Linux.
  - App Server: completo en Windows; en Linux trae todas las líneas, pero no conserva el orden entre stdout y stderr.
  - Claude: truncada en el modelo y en los hooks cuando el comando falla.

  Mientras tanto la captura independiente en Claude es la del arnés.
- **Umbral de S1 (07, hallazgo H1 de la revisión del PR #13).** La salida verbosa es **más de 300 líneas o más de 10.000 caracteres** (D3 de la 05, no D19).
