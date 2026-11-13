# Apply · Integración del ciclo de entrega Codex

13 de noviembre de 2026. Base de código main 35a89f8 integrada en el árbol de trabajo. Matías autorizó el alcance Machine y esta integración. El estado probado incluye cambios locales sin commit.

## Comportamiento implementado

El adapter Machine de Codex selecciona ConfigDeliveryManagedSkill. Project conserva la entrega anterior mediante su wrapper de target. El instalador despacha la entrega nueva mediante la capacidad opcional ManagedSkillAdapter y compone el paquete probado en la tanda anterior.

Antes de escribir, se comprueban rutas contenidas en el perfil, tipos de archivo, enlaces, propiedad e inventario, huellas de archivos administrados, markers y disponibilidad del primer respaldo. La skill homónima ajena y los archivos administrados editados detienen la operación. Las notas del usuario dentro de la skill se conservan.

El inventario de archivos vive en `.active-stack-managed.json`, dentro del directorio de la skill resuelto por el adapter. El primer respaldo se conserva en el almacén existente de backups con sufijo `-native-migration`, marcado como pinned y de origen install. El registro incluye directorios administrados y estado esperado persistente por archivo. La primera copia no se reemplaza al reinstalar ni al cambiar toggles. Un punto original perdido se informa como conflicto, sin capturar la configuración ya migrada como si fuera la original.

La instalación y el retiro usan un lock de directorio por paquete. Una segunda operación concurrente recibe un error busy antes de escribir. Si el proceso termina abruptamente y deja el lock, corresponde comprobar que no hay una operación activa antes de retirarlo manualmente; no existe recuperación automática del lock.

La operación guarda un snapshot previo propio. Los fallos de escritura restauran los archivos modificados, preservan ediciones concurrentes y permiten reintento cuando la recuperación completa es posible. El manifiesto persistente se actualiza atómicamente. Los módulos deshabilitados se retiran por inventario, conservando archivos ajenos. Los handlers de verify y uninstall usan los mismos helpers de contenido e inventario.

El plan general incluye archivos, inventario, manifiesto de migración y directorios del paquete. El paso config registra hashes escritos y ausencia esperada para rollback. Antes de restaurar un snapshot compartido, identifica archivos desconocidos en directorios nuevos y los preserva como conflicto; también verifica la identidad de los padres registrados para impedir seguir un enlace sustituido durante una verificación tardía. Un rollback que encuentre un archivo ajeno devuelve un conflicto de recuperación y preserva esos datos.

## Pruebas y evidencia TDD

| Trabajo | Safety net | RED ejecutado | GREEN y triangulación |
|---|---|---|---|
| Instalación Machine y rutas | 285 casos de model, adapter Codex, config e install | InstallManagedSkill inexistente y dispatcher sin entrega | Instalación, reinstalación, colisión, toggle retirado, notas propias y respaldo inicial |
| Primer punto persistente | Referencia backup de la tanda previa | Manifiesto original ausente aceptado y padres no registrados | Rechazo de punto perdido, pinned, archivo original conservado y restauración desde disco |
| Fallo y reintento | Suites anteriores | Fallo en módulo dejaba directorios vacíos que impedían reintentar | Recuperación, limpieza sólo de directorios vacíos y fallo atómico al publicar manifiesto |
| Estado ausente en rollback | Referencia backup | GuardRollbackAbsent inexistente | Archivo retirado restaurado y recreación posterior conservada |
| Snapshot compartido | Referencia install | Nota del usuario borrada por rollback del directorio SkillsDir | Nota conservada, instrucciones recuperadas, conflicto informado |
| Identidad de padres | Referencia install | Directorio preexistente sustituido por enlace causaba borrado en destino externo | Validación de identidad antes del rollback, destino externo intacto |
| Exclusión de operación | Referencia config | Segundo instalador aceptado pese a lock existente | Error busy sin escritura |

Comandos de comprobación de esta tanda:

- `go test -json ./internal/harness/config ./internal/agents/codex ./internal/install ./internal/backup ./internal/verify ./internal/uninstall` produjo 471 casos aprobados en seis paquetes, cero fallos. Registro en `/tmp/opsx04-lifecycle-final.jsonl`.
- `go vet` sobre esos paquetes pasó.
- `TestNativeRollbackPreservesReplacedExistingParent` se ejecutó después como regresión adicional RED/GREEN, junto con los otros casos de verificación tardía.
- `git diff --check` pasó. La pasada conjunta final del árbol completo se registra en el informe de integración del responsable principal.

## Cobertura y dependencias

Las pruebas de esta tanda son Go y perfiles temporales en macOS Intel. No acreditan aceptación CLI/IDE en Windows o Linux.

- 03 conserva la revisión conjunta del contrato de entrega y la matriz. El código adopta una capacidad opcional pequeña y una entrega explícita, conservando otros adapters.
- 05 debe entregar captura independiente y módulos de contexto con cobertura por interfaz.
- 06 debe entregar el recibo corregido y su procedencia. La entrega de archivos no convierte la declaración de un worker en evidencia verificada.
- La evidencia y el revisor independiente siguen como condición de las acciones que los necesitan. El paquete instalado informa ese límite.
- La comprobación del legado de Project en una sesión real depende de las sondas por interfaz. El target Project conserva el comportamiento anterior y no se migra silenciosamente.

La actualización con registro reaplica el mismo install. El camino sin registro y la selección de skills se prueban o revisan desde cmd/update por separado. La conservación de la primera copia y el retiro de archivos propios no dependen de elegir un formato provisional de recibos.

## Cierre de revisión operacional

Se agregaron regresiones para impedir seguir un enlace en el directorio del snapshot de operación, y para mantener permisos 0700 en el directorio privado del perfil durante instalación, retiro y restauración protegida. Esos casos fallaron antes de corregirlos y pasaron con el chequeo de rutas y la escritura atómica que preserva los permisos del padre. Los callers de escritura legados conservan su comportamiento anterior. La restauración protegida y la restauración con guardas live usan la variante que preserva directorios existentes.

El rollback de una verificación tardía se comprobó en tres escenarios: un directorio compartido con una nota nueva del usuario, recuperación limpia seguida de reintento y sustitución de un padre preexistente por un enlace hacia un directorio ajeno. En el último, el comportamiento anterior borraba un módulo del destino externo; la validación de identidad lo conserva y reporta el conflicto.

Después de estas correcciones, las suites de config, backup, install, verify, uninstall y adapter Codex y `go vet` sobre los mismos paquetes pasaron. El log de 471 casos corresponde a una pasada anterior a las últimas regresiones; el número definitivo de la pasada global queda en el informe final del responsable principal.

La prueba de conservación del primer snapshot se amplió a instalación inicial y dos actualizaciones distintas (TDD y luego Engram, retirando TDD). Comprueba bytes idénticos de la copia original, visibilidad mediante `backup.ListManifests` con origen install y pinned, y restauración del contenido original desde el manifiesto recargado. `go test ./internal/harness/config -run '^TestManagedSkillPreservesFirstSnapshot$' -count=1` pasó después de esa ampliación, sin cambios adicionales de producción.
