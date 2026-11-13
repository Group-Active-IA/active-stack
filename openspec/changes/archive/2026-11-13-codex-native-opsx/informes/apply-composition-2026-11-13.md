# Apply · Composición del paquete Codex

13 de noviembre de 2026. Alcance aprobado por Matías en la conversación. Código de referencia de main integrado en el árbol de trabajo desde `35a89f8`; la revisión incluye cambios locales sin commit. Este informe describe la primera tanda independiente del 04.

## Resultado

Se implementó `ComposeCodexPackage` en `internal/harness/config/codex_package.go`. Produce un bloque universal y un mapa de archivos relativos al directorio de skill resuelto por el adapter. La composición es pura, valida toggles y reutiliza los fragmentos embebidos existentes para conservar reglas de governance, memoria y TDD.

El paquete contiene la skill explícita, `agents/openai.yaml` con `policy.allow_implicit_invocation: false`, módulo de artefactos y módulos opcionales de delegación, wiring OPSX de memoria y TDD. Governance y guardado/búsqueda de memoria conservan su capa universal; el cierre y recuperación de sesión quedan en coordinación. Las asignaciones de modelos de Claude se omiten en Codex. Un cambio de encabezados que rompa la partición de fragmentos genera error de composición.

Los módulos opcionales se incluyen según toggles, y las referencias de la skill se producen con el mismo conjunto activo. El módulo de delegación consume el sobre de 07 y conserva el límite de dos workers escritores en árboles separados. La skill declara que las acciones que requieren captura y contrato de evidencia se detienen si esas dependencias faltan.

## Límite de integración

La tarea 2.3 tiene un componente de composición implementado y probado. El adapter sigue declarando la entrega anterior y `Install` conserva su recorrido vigente. Activar la migración completa requiere el preflight, respaldo persistente, escritura, verify, update y retiro coherentes de las siguientes tareas. Este paquete por sí solo todavía no instala un coordinador funcional completo. El retorno de 07 no sustituye los contratos pendientes de 05/06.

La primera tanda no modifica `internal/backup`, `internal/filemerge`, `internal/pipeline`, el catálogo ni archivos de configuración del perfil real.

## Dependencias identificadas

- **03**. Confirmar el contrato compartido de entrega y el mapa final de IDs, especialmente la evidencia requerida en ejecución directa. El compositor expone contenido y rutas relativas sin imponer un enum ni cambiar el despacho del adapter.
- **05**. Integrar sus módulos de contexto y captura cuando estén entregados y se compruebe la captura por superficie. La declaración de captura del agente no acredita por sí sola procedencia independiente.
- **06**. Integrar el contrato corregido de evidencia para ejecución directa y delegada. Hasta entonces las acciones dependientes quedan sin verificar o detenidas según su requisito.
- **P2**. Las pruebas unitarias verifican contenido y YAML. La aceptación de invocación y herencia efectiva en CLI/IDE de Windows/Linux sigue pendiente de comprobaciones con los clientes reales.

## TDD Cycle Evidence

| Tarea | Test File | Layer | Safety Net | RED | GREEN | TRIANGULATE | REFACTOR |
|---|---|---|---|---|---|---|---|
| 2.1 | suites existentes model/adapters/config | Unit e integración local | 298 pruebas pasadas, 0 fallos | No aplica, referencia previa | Referencia capturada | Consumidores y secciones existentes cubiertos por sus suites | Sin cambio de producción |
| 2.3, composición | `internal/harness/config/codex_package_test.go` | Unit | 298/298 | Fallo ejecutado por `ComposeCodexPackage` inexistente, luego fallo ejecutado por módulos, reglas universales y toggle inválido | Compositor implementado y pruebas pasadas | 32 combinaciones de toggles, orden y duplicados, metadatos YAML reales, ausencia de referencias colgantes, toggle desconocido y límites de fragmentos ausentes | gofmt y suite afectada verde |

Comandos ejecutados desde el worktree del 04:

- `go test -json ./internal/model ./internal/agents/... ./internal/harness/config/...` antes y después. Resultado posterior, 336 pruebas pasadas, 0 fallos y 0 omitidas.
- `go test ./internal/harness/config -run TestComposeCodexPackageExplicitActivation -count=1` para RED y primer GREEN.
- `go test ./internal/harness/config -run TestComposeCodexPackage -count=1` para la triangulación RED y GREEN.
- `go test ./internal/harness/config -run 'TestComposeCodexPackage|TestSplitCodexAsset' -count=1` para las combinaciones y validación de assets.
- `go vet ./internal/harness/config`, `git diff --check` y validación estricta de OpenSpec pasaron.

Los logs JSON temporales de referencia y GREEN se encuentran en `/tmp/active-stack-04-safety.jsonl` y `/tmp/active-stack-04-composition-green.jsonl` en este equipo. No se ejecutó un build del producto.

## Actualización de integración

Esta nota conserva la evidencia de la primera tanda. El ciclo completo de Machine y la tarea 3.3 se integraron después. El estado vigente y las limitaciones están en `apply-lifecycle-2026-11-13.md`, `verificacion-integracion-2026-11-13.md` y tasks.md.
