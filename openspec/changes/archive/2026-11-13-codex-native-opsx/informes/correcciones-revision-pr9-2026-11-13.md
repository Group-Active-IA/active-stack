# Correcciones de rutas de la revisión del PR #9

13 de noviembre de 2026. Base inspeccionada: `cd4aa3748cbd4642814be99becf965a58242f2ca` más el diff local de esta corrección. Alcance: observación de Juani sobre legibilidad de rutas Windows en los mensajes aportados por 04. No se implementó 03, no se alteraron markers, respaldo, permisos ni restauración, y no se limpiaron líneas en blanco del contenido del usuario.

## Observación resuelta

Los mensajes propios de `managed_skill.go`, `persistent_restore.go` y `managed_rollback.go` usaban `%q` para rutas. Go representa las barras inversas escapadas, por ejemplo `C:\\Users\\Mati Work`, aunque la ruta real tenga una sola barra. Se reemplazó exclusivamente el formato de rutas por comillas literales y `%s`: conserva separación visual, espacios y barras reales.

La búsqueda no encontró un formatter de rutas común reutilizable. No se agregó uno ni se cambió la estructura de errores. Los `%q` de identificadores de harness, toggles, markers y límites de secciones se conservaron. `codex_package.go` no tenía rutas en sus `%q`; la nueva rama de verify de 04 devuelve el chequeo de config y tampoco agrega formatos de rutas propios.

## Evidencia TDD

| Tarea | Test File | Layer | Safety Net | RED | GREEN | TRIANGULATE | REFACTOR |
|---|---|---|---|---|---|---|---|
| Diagnóstico de rutas administradas | `internal/harness/config/path_diagnostic_test.go` | Unidad del preflight real | Config/backup/uninstall/verify aprobados | Ruta Windows con espacios se mostraba escapada | Ruta literal conservada | Ruta Unix con espacios | Solo formato; gofmt y regresión aprobados |
| Diagnóstico de restauración persistente | `internal/backup/path_diagnostic_test.go` | Unidad de RestoreService | Mismo safety net | Guarda ausente mostraba ruta Windows escapada | Ruta validada conservada | Ruta Unix, respetando `filepath.Clean` existente | Sin cambios de política |
| Conflicto de rollback del retiro | `internal/uninstall/path_diagnostic_test.go` | Integración aislada | Mismo safety net | Padre y respaldo con barras se mostraban escapados | Archivo, padre y respaldo conservan su representación literal | Dos padres: con barra inversa y espacios, y nombre Unix con espacios | Sin cambios en selección o restauración |

Los tests ejecutan comportamiento de producción, no comparan un `fmt.Sprintf` aislado. Config y backup tuvieron un caso Windows fallido y uno Unix aprobado antes del arreglo; uninstall tuvo dos casos fallidos antes de reaplicar el cambio de formato a la línea del conflicto. Los seis casos pasan con el arreglo final. Los perfiles y archivos de prueba son temporales.

## Validación ejecutada

- Safety net: `go test -count=1 ./internal/harness/config ./internal/backup ./internal/uninstall ./internal/verify`, aprobado antes de agregar las regresiones.
- RED de config/backup: `go test -count=1 ./internal/harness/config ./internal/backup -run 'Test(ManagedPath|PersistentRestore)DiagnosticPreservesLiteralPath' -v`, fallo esperado de los casos Windows.
- RED de uninstall: `go test -count=1 ./internal/uninstall -run '^TestManagedRollbackDiagnosticPreservesLiteralPaths$' -v`, fallo esperado de las rutas escapadas.
- GREEN y regresión: `go test -count=1 ./internal/harness/config ./internal/backup ./internal/uninstall ./internal/verify`, aprobado.
- `go vet ./internal/harness/config ./internal/backup ./internal/uninstall ./internal/verify`, aprobado.
- `openspec validate codex-native-opsx --strict` y `git diff --check`, aprobados.

El coordinador revisó por separado el diff de los tres archivos de producción y los tests de las tres familias, y aprobó la corrección dentro del alcance de 04. No se repitió la suite global por un cambio exclusivo de formato: la entrega `cd4aa37` ya registra los fallos globales conocidos en `revision-independiente-04-2026-11-13.md`. Se ejecutaron nuevamente las áreas afectadas.

Logs locales: `/tmp/opsx04-paths-baseline.log`, `/tmp/opsx04-paths-red.log`, `/tmp/opsx04-paths-green.log`, `/tmp/opsx04-paths-uninstall-red.log`, `/tmp/opsx04-paths-regression.log`, `/tmp/opsx04-paths-vet.log` y `/tmp/opsx04-paths-openspec.log`. Los comandos y casos anteriores permiten reproducirlos; estos archivos temporales no constituyen adjuntos permanentes del PR.

## Límite preciso

Esto comprueba la representación de rutas, incluidas cadenas con formato Windows, desde el entorno local macOS. No acredita ejecución de Codex en Windows ni aceptación de CLI/IDE. Tampoco convierte en verdes los fallos globales preexistentes documentados.

Los formatos de errores preexistentes de `internal/backup/restore.go`, el writer compartido y los chequeos anteriores de verify no se modificaron por esta observación acotada a 04. Un error anidado mediante `%w`, incluido un error de sistema operativo, conserva su propia representación. No se presenta este arreglo como normalización de todos los errores del instalador.

La sugerencia del documento de archivar o aceptar deuda es material de revisión, no una autorización de Matías. Continúan 11/21 tareas, PR en borrador y aceptación compartida pendiente. No se archivó el change ni se aceptó deuda. La conciliación con 03 continúa a cargo de ambos responsables sin implementar el cambio del compañero.
