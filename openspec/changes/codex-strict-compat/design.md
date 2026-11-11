# Design

## Context

El instalador de permisos de Codex resuelve la ruta mediante el adapter, respalda el archivo existente y escribe las claves TOML de nivel superior de forma atómica. El verificador actual comprueba la presencia de las claves. La reproducción con `features list` aporta evidencia de carga en CLI 0.166.0 sobre Windows, con permisos efectivos pendientes de verificar.

## Goals / Non-Goals

**Goals**

- Reutilizar el punto de escritura, el respaldo y el rollback existentes para instalación y actualización.
- Separar la aceptación del TOML de las pruebas de permisos en una sesión real.

**Non-Goals**

- Rediseñar las reglas de ejecución, la confianza por proyecto ni la política de otras superficies.

## Decisions

1. En `internal/harness/config/permissions/install.go`, usar `sandbox_mode = "read-only"` y `approval_policy = "on-request"` para Estricto. La [guía oficial](https://learn.chatgpt.com/docs/agent-approvals-security#migrate-from-the-retired-untrusted-approval-policy) recomienda esta combinación para lectura interactiva. `never` también carga, pero impide pedir escalaciones. La vía de confianza por proyecto tiene efectos sobre la carga de configuración local y requeriría administrar cada proyecto.
2. Usar una descripción específica de Codex en los selectores que muestran Estricto. El texto propuesto es «Solo lectura por defecto. Codex puede pedir autorización para acciones que requieran permisos adicionales». Revisar las variantes en español e inglés y conservar los identificadores del tier.
3. Mantener el chequeo estático como diagnóstico y agregar una prueba de aceptación con el cliente real y un perfil aislado. Probar una lectura permitida y una escritura de prueba con escalación denegada. Verificar el estado del archivo después del intento y registrar la superficie, versión y resultado. Windows CLI 0.166.0 es el primer caso verificable. La matriz adicional queda (TBD) para la revisión del equipo.

## Risks / Trade-offs

- La política de aprobación permite solicitudes de escalación. El revisor efectivo y las restricciones adicionales dependen de la configuración del entorno. Verificarlos en la matriz de aceptación.
- Un valor legado puede ser una elección del usuario. Limitar la edición a la selección explícita de Codex Estricto o a su reaplicación registrada. Conservar tablas, perfiles y preferencias ajenas.
- Restaurar el respaldo recupera el archivo anterior, que podría seguir siendo incompatible con el cliente nuevo. El diagnóstico debe identificar el respaldo y la acción de reparación.

## Migration Plan

- El instalador administra las dos claves de nivel superior del TOML resuelto por el adapter de Codex. Si `verify` identifica `untrusted` en otra capa, informa su origen y el impacto observado sin modificarla. Cuando la política efectiva no se pueda determinar, informa que queda sin verificar.
- Ante TOML inválido, abortar e informar el diagnóstico con el archivo intacto. Para revertir, comparar el estado actual con el escrito por la operación. Si hubo ediciones posteriores, conservar el archivo actual e identificar el respaldo para resolución manual.

- `install` y el replay de `update` con selección registrada usan el mismo paso de permisos. Reutilizar el snapshot y el merge TOML de las dos claves de nivel superior. Si falla el respaldo, abortar la escritura. Si falla una operación posterior que exige revertir el paso, restaurar mediante el mecanismo existente y reportar cualquier fallo de recuperación.
- Si la instalación carece de registro, conservar el aviso de reinstalación con el modo y tier que el usuario elija. El recorrido de solo skills queda fuera de la migración.
- Revisar la propuesta antes de implementar cambios sobre configuración de usuario. Validar con fixtures y perfiles aislados, incluidas reinstalación y recuperación.

## Verificación y traspaso a 02

Juan Cruz confirmó el informe elaborado por el equipo como base de implementación. La primera superficie de aceptación es Codex CLI 0.166.0 en Windows. Las demás versiones e interfaces quedan (TBD).

- Terminal interactiva y ejecución no interactiva se registran por separado. En `codex exec`, la sesión observada aplicó `approval: never` y `sandbox: read-only` aun con el valor `on-request` solicitado. La escritura quedó bloqueada y el archivo testigo quedó intacto. La ejecución directa del sandbox con el perfil `:read-only` permitió leer el testigo y rechazó la escritura.
- En la TUI de Codex CLI, con `--ask-for-approval on-request --sandbox read-only` en un directorio temporal, el operador de la prueba rechazó la primera solicitud con Esc y el testigo conservó `before`. En una segunda solicitud para `Set-Content -LiteralPath witness.txt -Value after`, el operador aprobó una sola ejecución y el testigo pasó a `after`. La comprobación de carga con un `CODEX_HOME` temporal y los valores del instalador devolvió código 0.
- Para `opsx-baseline`, medir Estricto en una serie propia. En tareas de edición usar perfiles aislados con permisos equivalentes, registrar versión, interfaz, política efectiva y quién resuelve las solicitudes. Contabilizar bloqueos y rechazos como resultados de la corrida.
