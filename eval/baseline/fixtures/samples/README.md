# Muestras de captura (tarea 2.2)

**Derivadas de la documentación oficial, no capturas reales.** Ninguna sale de una corrida de `claude` ni de `codex`: no se ejecutó ningún cliente (cuota del usuario, y Codex no está instalado en esta máquina). Los nombres de campos y la estructura siguen la documentación consultada el 2026-11-11; los valores (identificadores, tokens, correos) son inventados, con correo `@example.invalid` y UUID de ceros. Riesgo declarado: si la forma real de un evento difiere de la documentada, el adaptador la tratará como incompleta o la ignorará, no la inventará; hay que repetir estas pruebas con muestras reales en 3.2.

Este directorio no es una familia de fixtures de 1.3: `fixture.Load` carga cada familia por nombre y no recorre `fixtures/`.

## Fuentes (consultadas 2026-11-11)

| Muestra | Basada en |
|---|---|
| `claude/otel-*.jsonl` | https://code.claude.com/docs/en/monitoring-usage: eventos `api_request` (`input_tokens`, `output_tokens`, `cache_read_tokens`, `cache_creation_tokens`, `request_id`, `query_source`, `agent.name`), `tool_result` (`tool_name`, `tool_use_id`), `subagent_completed` (`total_tokens`, `total_tool_uses`) y atributos `agent_id` / `parent_agent_id` (desde 2.1.310). |
| `claude/result-*.json` | https://code.claude.com/docs/en/agent-sdk/typescript (`SDKResultMessage`, `ModelUsage`) y https://code.claude.com/docs/en/agent-sdk/cost-tracking (`usage` excluye subagentes; `total_cost_usd` y `modelUsage` los incluyen). |
| `claude/stream-delegated.jsonl` | https://code.claude.com/docs/en/headless (`stream-json`, `system/init`, `system/api_retry`, `parent_tool_use_id`) y `SDKAssistantMessage` / `SDKSystemMessage` del SDK de TypeScript; mensajes de un mismo paso comparten `message.id` (cost-tracking). |
| `codex/exec-doc-example.jsonl` | https://learn.chatgpt.com/docs/non-interactive-mode: las líneas 1 a 3 y 5 son las del ejemplo de la documentación; la línea 4 (`item.completed` con `agent_message`) sigue la descripción de la página, cuyo texto exacto no se pudo citar completo. |
| `codex/exec-delegated-*.jsonl`, `exec-partial`, `exec-failed`, `exec-duplicates` | Forma de `turn.completed.usage` y de los items de la misma página. **La forma de los hilos hijos, de `error` y de `turn.failed` no está documentada**: son sintéticas. |

## Archivos

| Archivo | Caso |
|---|---|
| `claude/otel-delegated.jsonl` | Corrida delegada completa: principal 9.500 + 30.500 = 40.000; un worker (`agent_id` `a3f2c1b4d5e6f7a80`, `agent.name` `custom`) con solicitudes de 12.000 y 8.000; cierre del worker con `total_tokens` 8.000; 7 `tool_result` (4 principal, 3 worker). |
| `claude/otel-partial-old-version.jsonl` | Parcial y versión anterior a 2.1.310: una solicitud de subagente sin `agent_id` y una solicitud sin campos de tokens. |
| `claude/otel-duplicates.jsonl` | La misma solicitud exportada dos veces (mismo `request_id`). |
| `claude/result-delegated.json` | `claude -p --output-format json` de la corrida delegada: `usage` = 40.000 (solo el bucle principal); `modelUsage` suma 60.000 (650 + 2.300 + 45.000 + 12.050) e incluye al worker. |
| `claude/result-error-max-turns.json` | Resultado `error_max_turns` con una `permission_denials` cuya `tool_input` lleva un texto marcado `SECRET` que el adaptador no debe conservar. |
| `claude/stream-delegated.jsonl` | `stream-json`: init (versión 2.1.310, modo `acceptEdits`), mensajes del asistente repetidos por llamadas en paralelo (mismo `message.id`), mensajes de subagente con `parent_tool_use_id` y `agent_id`, un `api_retry` y el `result`. La salida por paso vale 1 (marcador). |
| `codex/exec-doc-example.jsonl` | Ejemplo de la documentación: turno con entrada 24.763, cacheada 24.448, salida 122. |
| `codex/exec-delegated-parent.jsonl` | Hilo padre, dos turnos: 20.000 con 15.000 cacheadas + 2.000 de salida (22.000) y 7.000 con 6.000 cacheadas + 1.000 (8.000) = 30.000. |
| `codex/exec-delegated-child.jsonl` | Hilo hijo, un turno: 17.000 con 10.000 cacheadas + 1.000 = 18.000. El padre y el hijo juntos reproducen los 48.000 del ejemplo de reconciliación de `design.md`. |
| `codex/exec-partial.jsonl` | Flujo cortado: un turno que nunca se completa. |
| `codex/exec-failed.jsonl` | `error` y `turn.failed` sin `turn.completed`. |
| `codex/exec-duplicates.jsonl` | `thread.started`, `item.completed` y `turn.completed` repetidos. |
