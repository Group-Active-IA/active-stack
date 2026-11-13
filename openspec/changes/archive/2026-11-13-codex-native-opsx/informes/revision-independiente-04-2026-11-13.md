# Revisión independiente del cambio 04 — 13/11/2026

## Alcance y conclusión

Se revisaron el diff local del diagnóstico Project y el recorrido de propiedad, snapshot, actualización, retiro y rollback de la entrega Machine. La base inspeccionada fue `9a6eb2aeb460f89afead74fc17669519c6d8d854` más los cambios locales de 04. No se incorporó implementación del 03 ni se ejecutó build.

La revisión detectó un defecto de preservación en la instalación de varios adapters. Se reprodujo antes de corregirlo y el arreglo mínimo pasó las pruebas de las áreas afectadas. La revisión del diagnóstico Project y de los recorridos de preservación inspeccionados queda aprobada con los límites de aceptación indicados abajo. Esta revisión fue realizada por un agente distinto de quienes implementaron originalmente esos recorridos. El mismo revisor corrigió el defecto detectado; el coordinador revisó y aprobó por separado el diff de producción y ambos tests de regresión. Esa revisión técnica no sustituye la aceptación humana compartida y no se presenta como revisión independiente del propio arreglo por su autor.

## Hallazgo corregido

**P1 · Pérdida de estados de escritura nativa si falla un adapter posterior.** En `internal/harness/config/install.go`, la instalación de Codex acumulaba `Files` y `ManagedWrites`, pero las salidas por error de un adapter posterior con entrega de instrucciones o primary-agent devolvían `Result{}`. El paso de instalación recibía el error sin las guardas necesarias para proteger la entrega nativa durante rollback. Una edición posterior podía perder su protección.

El arreglo conserva los resultados parciales acumulados —rutas, estados esperados y avisos— en cada salida por error del bucle. No cambia el despacho de otros clientes, no añade una estrategia de entrega y no amplía las rutas ni la política de rollback aprobada. `AllAlready` permanece falso ante error.

La reproducción usa un adapter nativo seguido de otro cuyo directorio padre es un archivo del usuario. Las dos variantes de inyección fallaron con el mensaje `lost prior native writes on later adapter failure` antes del arreglo. Después se comprobó también el error de una interfaz opcional de rutas ausente.

## Evidencia TDD

| Tarea | Test File | Layer | Safety Net | RED | GREEN | TRIANGULATE | REFACTOR |
|---|---|---|---|---|---|---|---|
| 3.4 · Conservar estados parciales tras fallo posterior | `internal/harness/config/partial_result_test.go` | Integración aislada | Config/install aprobados antes del test nuevo | Dos variantes fallaron por pérdida de `Files` y `ManagedWrites` | Ambas variantes aprobaron tras el arreglo | Tres casos: instrucciones, primary-agent y capacidad ausente | Retorno parcial común; gofmt y suites verdes |
| 3.4 · Protección al llegar a rollback | `internal/install/partial_result_test.go` | Integración aislada | Install aprobado | Cubierto por RED del contrato anterior; este test se añadió tras el arreglo | Instalación falla, rollback informa conflicto y conserva edición posterior | Complementa el contrato con un rollback real sobre snapshot | Sin cambio adicional de producción |

No se agregó ni marcó una nueva tarea por esta corrección de 3.4. La tarea 4.4 conserva su estado pendiente: esta revisión avanza su evidencia, no satisface por sí sola toda la aceptación de entrega.

## Comprobaciones ejecutadas

- `go test -count=1 ./internal/harness/config ./internal/install`: aprobado antes del test de regresión y después del arreglo.
- `go test -count=1 ./internal/backup ./internal/filemerge ./internal/uninstall ./internal/verify`: aprobado.
- `go vet ./internal/harness/config ./internal/install`: aprobado.
- `openspec validate codex-native-opsx --strict`: aprobado después de añadir este informe.
- `go test -race -count=1 ./internal/harness/config ./internal/install`: aprobado.
- `go test -json ./... -skip '^TestWindowsBundleThemeContainsMinimalInstallerPages$'`: 1927 eventos de casos/subcasos aprobados, 6 fallidos y 2 omitidos; 36 paquetes aprobados y 2 fallidos. Los seis eventos corresponden a cuatro tests principales, incluidos sus subtests. Se excluyó expresamente el test de tema Windows con fallo previo en macOS.

La suite global volvió a registrar los fallos ya documentados en main: `TestDownloadBinary_APIError` (API 404 sin error) y los tres tests principales de `eval/baseline/fixture` relativos a la huella congelada de revisión (`TestCheckout_RepetitionsStartIdenticalAndRestore`, `TestEvaluate_ReviewDispositionsAndMechanism`, `TestEvaluate_ReferenceAcceptedDefectiveRejected`). No aparecieron tests fallidos nuevos en esta corrida. La reproducción en main se conserva en el informe anterior y no se repitió innecesariamente.

Se revisaron `InstallManagedSkill`, `RemoveManagedSkill`, los estados persistentes, la restauración protegida, la transmisión de guardas en `configStep` y el filtrado de rollback del retiro. Los tests existentes cubren conservación de primera copia tras dos actualizaciones, manifiesto recargado, archivo propio, directorio sustituido, symlink, notas, permisos y errores de escritura. No se encontró otro defecto confirmado en este alcance.

## Identidad de la evidencia

Los logs se conservaron localmente en `/tmp`; no son adjuntos permanentes del PR. Los escenarios y resultados anteriores permiten reproducirlos. Las siguientes huellas identifican el contenido local de producción y regresión probado, separado del commit base:

| Archivo | SHA-256 |
|---|---|
| `internal/harness/config/install.go` | `f97ebe8a607646142c4b673bf9f8d28311c0ab78fdd0a77190385f6eea26a5d2` |
| `internal/harness/config/partial_result_test.go` | `7c716b939aa26f571148b84c1624960d7b5d606192cbcb08a7a18dd092dbb998` |
| `internal/install/partial_result_test.go` | `9d06b17e76ee709b3cd0b6cf9dab5180a1549a5d7a1496c344564cade215c871` |
| `/tmp/opsx04-review-partial-result-red.log` | `a1a43e2cf0b4f39320c5f704ce7a1cbb4436ddc33f9fbbc03445dc6bf4fdbe70` |
| `/tmp/opsx04-review-partial-result-green.log` | `dc4bc8fc3c860aecc9ef25e8fd6cfc13d7d63319e30d01ce2e5d8de90be4a1bb` |
| `/tmp/opsx04-review-areas-green.log` | `43a183227fd23fdf364447d4459668cb4feb1b1672e9408f31010d58bdbcde51` |
| `/tmp/opsx04-review-vet.log` | `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` |
| `/tmp/opsx04-review-race.log` | `d92f76d9481cf679aacb74a20a11ced70605d40716b425f36870ff7cc52e1e2d` |
| `/tmp/opsx04-review-openspec.log` | `ca8f686f3ef8538c5197b7759547c2514b4a4e12e2d91e687899324f2f3dae14` |
| `/tmp/opsx04-review-full.jsonl` | `9106e78356740172429e0fe2e222ecad604bb0e0483ad43cbc6918a4df443f35` |

El log RED conserva el nombre inicial `TestInstallRetainsNativeResultWhenLaterComposeFails`; antes de entregar se renombró a `TestInstallRetainsNativeResultWhenLaterAdapterFails`, porque la reproducción efectiva falla en inyección. No se atribuye ese RED a la variante de capacidad ausente ni al test de integración añadido después.

## Límites y pendientes reales

- La suite global sigue con fallos ajenos ya reproducidos en main, registrados en `verificacion-integracion-2026-11-13.md`; no se los presenta como suite verde ni se exige arreglarlos dentro de 04. Su disposición y los demás criterios de entrega mantienen pendiente 4.4.
- La revisión no reemplaza la aceptación de clientes reales CLI/IDE Windows/Linux ni acredita activación, herencia o especialistas.
- La revisión conjunta humana de matriz y migración con 01/03/08 sigue pendiente; inspeccionar sus interfaces no equivale a implementar el 03.
- La evidencia por acción y OPSX `verify` siguen sujetos a los contratos corregidos de 05/06/07 y a la comparación con 02.
- El PR conserva su condición de borrador y no se declara terminado el cambio 04.
