# Sonda local de descubrimiento de skill · 13 de noviembre de 2026

Matías aprobó el alcance del 04 y autorizó iniciar las tareas independientes el 13/11/2026.

## Entorno

- Codex CLI 0.167.0-alpha.3.2, macOS Intel.
- Cliente App Server por stdio. Solicitudes `initialize`, `initialized` y `skills/list` con `cwds` del fixture temporal y `forceReload: true`.
- Fixture con `.agents/skills/opsx04-probe/SKILL.md`, política `agents/openai.yaml` con `allow_implicit_invocation: false` y `modules/example.md` sin metadatos de skill.
- El proceso se terminó y el fixture se eliminó al finalizar. La sonda realizó descubrimiento sin ejecutar una tarea de modelo.

## Resultado observado

`initialize` respondió correctamente. `skills/list` devolvió `opsx04-probe` con `enabled: true`. El archivo de apoyo no apareció como una skill adicional. La respuesta filtrada conservó únicamente los campos del fixture.

Este resultado respalda el descubrimiento del paquete y la distinción estructural entre skill y apoyo en esta versión local. No permite concluir bloqueo de activación implícita, ejecución manual, herencia de reglas o continuidad de contexto.

## Comprobaciones pendientes

- Tarea 1.2 y 4.2, CLI y extensión IDE reales en Windows y Linux con versiones identificadas.
- Activación manual y pedido implícito, aplicación de módulos y permisos, especialista y reanudación.
- 03 mantiene la matriz compartida y el contrato técnico de entrega.
- 05/06 mantienen captura y procedencia. La integración de evidencia espera sus contratos corregidos.
