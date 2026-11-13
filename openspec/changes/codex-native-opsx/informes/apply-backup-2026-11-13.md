# Apply · Protección persistente de restauración

13 de noviembre de 2026. Alcance aprobado por Matías. El árbol incluye main 35a89f8 y cambios locales sin commit.

## Implementación y límite

`Manifest.GuardPersistentRestore` registra la huella esperada de cada archivo y activa un esquema versionado de restauración protegida. Los campos se serializan en manifest.json. Al recargarlo, Restore valida el esquema y las guardas antes de restaurar; una edición, eliminación o sustitución se conserva como conflicto. El modo protegido retira directorios creados sólo cuando están vacíos y siguen siendo directorios. La restauración de manifiestos anteriores conserva su recorrido existente.

Este soporte no conserva por sí solo la primera copia de migración: el instalador del 04 debe crearla una vez, persistir las operaciones y registrar las rutas administradas, incluidos sus directorios. Tampoco activa la nueva entrega de Codex. La tarea 3.3 permanece parcial hasta esa integración.

## Evidencia TDD

| Trabajo | Safety net | RED | GREEN y triangulación | Refactor |
|---|---|---|---|---|
| Guardas persistentes | Suite backup existente aprobada | Método inexistente; luego ausencia de guarda y archivos ajenos; sustitución de directorio | Recarga desde disco, edición y eliminación posteriores, archivo/enlace, archivos nuevos del usuario, versión inválida, guardas incompletas o ajenas y reintento | gofmt, suite y vet |

Los casos viven en `internal/backup/persistent_restore_test.go`. Las pruebas usan directorios temporales. La comprobación conjunta y el caso adicional de enlace en directorio intermedio se registran en el resumen de esta tanda al cerrar la revisión.

## Revisión final y resultado conjunto

La revisión añadió un caso de padre registrado sustituido por enlace hacia un directorio externo. La prueba reprodujo una escritura sobre ese destino antes de la corrección. Ahora se valida con Lstat cada directorio registrado antes de restaurar archivos y se aborta ante sustitución. La regresión cubre padres creados y preexistentes. El migrador deberá inventariar los directorios administrados para disponer de esa comprobación.

Tras la corrección, `go test -race ./internal/backup -count=1` y `go vet ./internal/backup` pasaron. La pasada combinada de backup, model, adapters y config pasó con **446 casos en 9 paquetes, cero fallos**. El registro JSON de esa pasada está en `/tmp/opsx04-independent-tests.jsonl` en este equipo.

La tarea 3.3 sigue pendiente de integrar el primer snapshot, su conservación y acceso en el flujo instalador. Las tareas 2.1 y 2.3 están completas y OpenSpec registra 2/21. Las pruebas de los clientes reales en Windows/Linux e IDE y la integración con 03/05/06 siguen identificadas en tasks y design.

## Actualización de integración

Esta nota conserva la evidencia de la primera tanda. El ciclo completo de Machine y la tarea 3.3 se integraron después. El estado vigente y las limitaciones están en `apply-lifecycle-2026-11-13.md`, `verificacion-integracion-2026-11-13.md` y tasks.md.
