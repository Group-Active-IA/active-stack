# Proposal: opsx-native-controls

> Change `opsx-native-controls` · ficha 08 del roadmap · Responsable: Gianfranco · Agentes: Claude y Codex · Prioridad P2 · Gobernanza **ALTO**
> Base revisada: `7a60089` (`origin/main`, con 01, 02 y 03 integrados). Depende de 06 (rama `opsx-verifiable-evidence`, local, pendiente de push; D2, D3 y D9). Usa como hechos de `main` la política de Estricto y el método de aceptación de Codex de 01 (`codex-strict-compat`, design D11) y la activación por skill explícita de 03 (design D12). 04 (`codex-native-opsx`) sigue pendiente y se cita contra `origin/propose/codex-native-opsx` `9a6eb2a`. Habilita 10. Responsable: Gianfranco; revisa e integra Neyén.
> Esta ronda es solo `propose`. Por tocar configuración del usuario, **ninguna implementación empieza sin revisión humana previa** (nivel ALTO).

## Why

Hoy el stack le pide cosas al agente con texto (el orquestador, las reglas, las skills) y fija permisos con el harness `permissions`. El texto orienta al agente, pero no impide una acción. Los permisos y los hooks nativos de cada cliente sí actúan sobre la acción real, pero no se sabe qué cobertura dan en la práctica, qué ocurre cuando un control falla ni dónde están sus límites en cada shell y superficie. Que un hook esté instalado no prueba que se ejecute, que cubra todas las rutas ni que pueda bloquear una acción antes de que ocurra.

Esta propuesta planifica un piloto que mida todo eso **antes** de que el instalador escriba algo nuevo en la configuración de nadie.

### Evidencia

- **El nombre "hook" está ocupado por dos cosas distintas.** En el instalador, `install.Options.VerifyHook` es una función Go que corre al final de la instalación y puede hacer rollback (`internal/verify/hook.go`). Los hooks de un cliente (por ejemplo los de `settings.json` de Claude Code) son comandos que el cliente ejecuta alrededor de una acción. No tienen relación entre sí. Los adapters de hoy exponen `SettingsPath` pero **ninguna capacidad de hooks**.
- **Caso observado el 02/11/2026** (v0.4.0, Windows, modo `lite`, agente `claude`): el `verify-hook` del instalador terminó ✓ con "0 passed, 0 failed", y `~/.claude/settings.json` no tenía ningún hook de cliente. El paso de verificación no comprueba que los controles existan.
- **La escritura de permisos reemplaza valores del usuario** (`origin/main`, `7a60089`). `permissions.Install` mezcla un overlay con `filemerge.MergeJSONObjects`. En esa función solo los objetos se mezclan de forma recursiva: **los valores escalares y las listas del overlay reemplazan a los del usuario** (salvo el sentinel `__replace__`, que hace lo mismo de forma explícita). El overlay de Claude fija `permissions.defaultMode` (`"default"` en los tiers Estricto y Balanceado, `"bypassPermissions"` en Bypass), `permissions.deny` y, en Balanceado, `permissions.allow`. Una entrada que el usuario haya agregado a `allow` desaparece al reinstalar.
- **Observado en la misma instalación del 02/11:** el instalador cambió `defaultMode` de `"auto"` a `"default"` y borró un permiso propio del usuario (`mcp__moodle-tutor__responder_mensaje`) de `allow`. Es coherente con lo anterior; no se reprodujo de forma aislada (tarea 1.4).
- **Un `settings.json` malformado se trata como `{}`.** `MergeJSONObjects` descarta un archivo que no puede interpretar y escribe solo el overlay (`internal/filemerge/json_merge.go:10-15`); además elimina los comentarios (`:49-52`). `UpsertTOMLTopLevel`, en cambio, valida y aborta ante TOML inválido (`toml.go:26-28`) y convierte todo CRLF a LF (`toml.go:29`). `WriteFileAtomic` hace `chmod 0755` del directorio padre y `0644` del archivo (`writer.go:40,62`).
- **El nivel de gobernanza es ALTO** en este dominio: el código de `permissions` lo marca así y `CLAUDE.md` pide proponer y esperar review antes de implementar.
- **Deuda conocida de desinstalación.** El paso de `permissions` de `uninstall` no quita nada (`internal/uninstall/steps.go:333`, "(TBD) There is no dedicated mechanism yet to selectively strip just the permissions overlay"), y quitar un MCP externo no borra su entrada de `~/.claude.json` (`docs/archive/CHANGES-2026-11-05-post-starters.md:511`); arreglarlo toca la remoción de JSON cerca de `filemerge` (ALTO). Quitar hooks de `settings.json` tendría el mismo problema.
- **Qué hace un hook del cliente cuando falla** según la documentación oficial consultada el 2026-11-13: en Claude Code un hook que falla no detiene la acción por defecto, salvo `onFailure: "block"` (v2.1.295 o posterior, hooks `command` y `http`); en Codex, ante salida no soportada o inválida de un `PreToolUse` local, el hook se marca como fallido y la llamada continúa, el comportamiento ante error o timeout de un hook local es (TBD), y la propia documentación los llama "a useful guardrail, not a complete enforcement boundary" (design D6).

## What Changes

- Se **pilotean permisos y hooks nativos** en Claude y en Codex, sobre lo que cada cliente soporte en la versión probada.
- Cada regla se **clasifica en tres**: *orienta* al agente (texto), *restringe* una acción (permisos o rules del cliente) o *verifica* un evento (hook).
- Los controles **previos** se usan para lo que debe bloquearse; los **posteriores**, para validar resultados (por ejemplo contra el recibo de la propuesta 06).
- Se define la **cobertura** esperada y el **comportamiento ante fallo** del control (pendiente de acuerdo del grupo según el roadmap).
- Para Codex se usa la política de Estricto que 01 dejó en `main` (`sandbox_mode = "read-only"` con `approval_policy = "on-request"`) y su método de aceptación. Para Claude se verifica que la activación por skill explícita de 03 (`/sdd-orchestrator`, sin `agent` ni `--agent`) no colisione con el overlay de permisos (design D12).
- Se **documentan los límites** por cliente, versión, shell y superficie. Una acción cuya cobertura no se probó figura como *no controlada*.
- Se separa por escrito el **hook de verificación del instalador** de los **hooks del cliente**.
- Se exige que la escritura de permisos y hooks **preserve las entradas del usuario**, con diff previo, revisión, backup y reversión, y que **no pase por encima de un archivo que no pudo leer** (design D9). La reversión reutiliza la guarda de 01 (`Manifest.GuardRollback`) para no pisar ediciones posteriores (D8).
- Regla de **control obligatorio ausente**: se detiene solo la acción afectada, se explica la protección que falta y se reanuda con una protección equivalente verificada; un control opcional ausente produce una advertencia (design D7, alineada con 03 D11 y 04, sin redefinirlas).
- Se entrega un **inventario de escrituras** con su nivel de gobernanza y una **tabla de cobertura** en formato cerrado para 10 (design D13), y se mide el efecto de `UpsertTOMLTopLevel` (CRLF a LF) y de `WriteFileAtomic` (chmod) en Windows y Linux.

## Capabilities

### New Capabilities

- `native-controls-pilot`: inventario y clasificación de controles nativos por cliente, política de fallo, cobertura y límites documentados, y escritura de configuración con revisión, backup, preservación de entradas ajenas, idempotencia y reversión.

### Modified Capabilities

(ninguna archivada: el harness `permissions` existe en el código, pero no hay un spec previo que lo describa en `openspec/specs/`)

## Alcance

- Inventario de mecanismos nativos de permisos y hooks por cliente y versión soportada.
- Clasificación de acciones en "debe bloquearse" y "se valida después", partiendo de lo que hoy cubre `permissions`.
- Matriz de pruebas por cliente, superficie, herramienta y shell: permitido, denegado, timeout, JSON inválido, handler ausente, rutas alternativas y shells reales.
- Diseño de cómo entran los hooks al flujo existente de backup, merge, escritura atómica, `update` y `uninstall`.
- Preservación de las entradas del usuario al escribir permisos.
- Documentación de límites por superficie.

## Exclusiones

- La política y la migración del tier estricto de Codex (propuesta 01). Aquí se reutiliza su política y su método de aceptación para las celdas de Codex (design D11); no se redefinen.
- El contrato de evidencia (06). Esta propuesta lo consume.
- El juez externo en sombra (10).
- Subcomandos nuevos de runtime en el binario o un ledger firmado. El roadmap los deja para una decisión de arquitectura separada.
- OpenCode y Antigravity, salvo que el grupo acuerde incluirlos. Nota: OpenCode evalúa sus reglas con *last-match-wins*, lo opuesto a Claude Code.
- Implementación: `apply`, pruebas de implementación, `verify` y `archive` son de una etapa posterior.

## Impact

- **Gobernanza ALTO.** Toca `harness/config/permissions` (escribe configuración del usuario) y, probablemente, `filemerge` para tratar claves de hooks, listas, archivos ilegibles y su remoción. `CLAUDE.md` pide proponer y esperar review antes de implementar. El PR toca `filemerge` y lo revisa Neyén, según el reparto; cualquier cambio a `filemerge` requiere ese revisor humano (design D15).
- **MEDIO:** `agents` (adapters), si se agrega una capacidad opcional de hooks.
- **Documentación:** una tabla de cobertura y límites, y una guía breve de qué queda garantizado por control nativo y qué solo por instrucción. Es la línea de base de la propuesta 10.
- **Espejo obligatorio:** si se modifica la constitución del repo, `CLAUDE.md` y `AGENTS.md` se actualizan juntos.
