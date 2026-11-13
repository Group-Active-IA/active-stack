# Design

## Context

**Base y fuentes (revisadas el 2026-11-13).** El código y los changes mergeados se citan contra `origin/main` `7a60089` (01, 02 y 03 integrados; 03 en el PR #8). 04 (`codex-native-opsx`) sigue sin mergear y se cita contra `origin/propose/codex-native-opsx` `9a6eb2a`. 06 se cita contra la rama `opsx-verifiable-evidence` (local, pendiente de push), D2, D3 y D9. Reparto: 08 es de Gianfranco y la revisa e integra Neyén; 06 la revisa Juan Cruz; 03 es de Juani; 04, de Matías; 05, de Emiliano; 07, de Neyén.

El stack hoy ofrece tres cosas distintas que conviene no mezclar:

1. **Texto** (orquestador, reglas, skills): orienta al agente, no impide nada.
2. **Permisos del cliente**: el harness `permissions` (tipo `config`, no se puede desactivar ni en Custom) escribe, por agente y nivel, un overlay en la configuración del cliente. En Claude se mezcla con `filemerge.MergeJSONObjects` en `~/.claude/settings.json` (`internal/harness/config/permissions/install.go:180`); en Codex se escriben `approval_policy` y `sandbox_mode` en `~/.codex/config.toml` con `UpsertTOMLTopLevel` (`install.go:127`). Hay tres niveles: Estricto, Balanceado y Bypass; un nivel vacío se normaliza a Balanceado, nunca a Bypass (`install.go:54-60`). Antes de escribir se crea un snapshot en `~/.active-stack/backups/permissions-<agente>` y, si el backup falla, no se escribe (`install.go:164-167`).
3. **Hook de verificación del instalador** (`install.Options.VerifyHook`, `internal/install/types.go:108-110`; constructor en `internal/verify/hook.go`): una función Go que corre al final de la instalación y hace rollback si un check duro falla. **No es un hook del cliente.**

**Codex Estricto ya está resuelto en `main` (01).** El instalador escribe Estricto como `approval_policy = "on-request"` con `sandbox_mode = "read-only"` (`install.go:120-126`); Balanceado, `on-request` con `workspace-write`; Bypass, `never` con `danger-full-access`. Es la política que 01 fijó (`openspec/changes/codex-strict-compat/design.md:20`) y todas sus tareas figuran completas (`tasks.md`, `[x]`). Este change la toma como un hecho de `main` (D11).

Lo que hoy cubre `permissions` en Claude es el piso `deny` (lectura y edición de `.env`, `rm -rf /` y `~`, `git push --force`: `internal/harness/config/permissions/overlays.go:16-28`). Claude evalúa deny → ask → allow y el piso se aplica en todos los niveles, Bypass incluido (`overlays.go:9-12`).

`active-stack update` reinyecta configuraciones con el pipeline de instalación, registra la selección por agente y exige adopción explícita cuando encuentra material sin ownership registrado. Cualquier control nuevo tiene que entrar en ese circuito.

## Goals / Non-Goals

**Goals**

- Medir qué cubren de verdad los controles nativos (por cliente, versión, shell y superficie) y qué pasa cuando fallan.
- Dejar escrito qué queda garantizado por control nativo y qué solo por instrucción.
- Que ninguna escritura sobre la configuración del usuario ocurra sin diff, revisión, backup y reversión, sin pisar entradas ajenas y sin pasar por encima de un archivo que no se pudo leer.

**Non-Goals**

- Prometer protección que no se probó.
- Resolver el tier estricto de Codex (01, ya en `main`) o redefinir el contrato de evidencia (06).
- Subcomandos nuevos del binario, ledger firmado o runtime nuevo.
- Modificar `filemerge`, `backup` o `pipeline` en esta ronda: es solo propuesta; cualquier cambio posterior requiere el revisor ALTO (D15).

## Decisions

### D1 — Un vocabulario sin ambigüedad

En todos los artefactos de este change:

- **Hook del instalador** = `VerifyHook` de la instalación (Go, del binario). Fuera del piloto, salvo para decir qué *no* comprueba. En el caso del 02/11 terminó ✓ con "0 passed, 0 failed" sin comprobar ningún control.
- **Hook del cliente** = comando que el cliente ejecuta alrededor de una acción (Claude Code en `settings.json`; Codex en `hooks.json` o en una tabla `[hooks]` de `config.toml`, D6).
- **Recibo de ejecución** = el registro de 06 (una ejecución de un comando). No es el "registro de instalación" de 04.

Cada regla del piloto se clasifica en **una** de tres clases:

| Clase | Qué hace | Ejemplo |
|---|---|---|
| **Orienta** | Le dice al agente qué hacer (texto). No impide nada. | Una instrucción del orquestador |
| **Restringe** | El cliente bloquea o pide confirmación antes de la acción, mediante permisos o rules. | `deny` de lectura de `.env` |
| **Verifica** | Un hook del cliente evalúa un evento y puede validar o frenar. | Un hook posterior contra el recibo de ejecución de 06 |

### D2 — Controles previos para bloquear, posteriores para validar

- **Previos:** para lo que no debe ocurrir sin autorización (leer `.env`, git destructivo). Impiden el efecto. Si fallan, hay que decidir qué pasa (D6).
- **Posteriores:** para validar el resultado, como el contrato de 06. No deshacen lo ya hecho, así que no son la única defensa para acciones destructivas. Un control posterior solo puede comprobar de forma mecánica los campos del recibo de ejecución; la corroboración de procedencia y la suficiencia son de 06 (D3 de 06), no de 08.

### D3 — Cobertura declarada, no supuesta

Para cada cliente, versión, shell y superficie se registra qué acciones cubre el control nativo en una **matriz** (cliente × superficie × herramienta × shell). Una acción sin cobertura comprobada se documenta como *no controlada*. No se afirma protección que no se probó. El formato de la tabla que consumirá 10 está en D13.

### D4 — Cómo se entregan los hooks del cliente

Los permisos ya siguen un camino probado: resolver la ruta por adapter, snapshot, merge, escritura atómica y omitir si no hay cambios. Los hooks de Claude viven en el mismo `settings.json`, así que la propuesta es reutilizar ese camino y agregar a los adapters una **capacidad opcional de hooks**, que cada cliente declare si soporta. Codex documenta hooks (D6), pero la ubicación (`hooks.json` o `[hooks]` en `config.toml`) y el modelo de confianza difieren, así que se decide por cliente.

El JSON no admite comentarios, así que hace falta una forma de reconocer qué entradas son del stack para actualizarlas y desinstalarlas sin tocar las del usuario. Opciones: un identificador distintivo dentro de cada entrada o un registro de ownership junto a los backups. **Decisión del grupo (TBD)**: toca `filemerge`, nivel ALTO (D15).

### D5 — Los handlers no pueden ser un subcomando del binario

Un hook ejecuta un comando. El roadmap excluye comandos nuevos de runtime, así que el handler no puede ser algo como `active-stack check`. Opciones, de menor a mayor alcance:

1. Solo reglas declarativas de permisos, sin hooks.
2. Hooks con comandos de shell en línea, simples y revisables en el diff.
3. Scripts instalados por el stack: exigen una variante por sistema (shell POSIX y PowerShell) y su propio ciclo de `update` y `uninstall`. Codex no distribuye los scripts de hooks gestionados (D6).
4. Subcomando del binario: requiere la decisión de arquitectura separada del roadmap.

Se propone pilotear las dos primeras y evaluar la tercera según los resultados (TBD, grupo). Los comandos se prueban en los shells reales de cada plataforma (D16).

### D6 — Comportamiento ante fallo del control (según la documentación oficial)

Se distinguen dos fallos: el de un **hook del cliente** (esta decisión) y el del `VerifyHook` del instalador (D1), que no es un control. Hay dos políticas posibles: **fallar cerrado** (se bloquea la acción) o **fallar abierto** (se permite y se avisa). El piloto **registra lo que cada cliente hace realmente**; lo siguiente es lo que la documentación afirma, con fuente y fecha. Todo se consultó el **2026-11-13**; ninguna de las dos páginas muestra fecha de publicación. Lo que no se pudo verificar queda **(TBD)**.

**Claude Code** — fuente: `https://code.claude.com/docs/en/hooks` (sin fecha en la página; la versión que cita para `onFailure` está abajo).

- *Por defecto un hook que falla no detiene la acción.* "Non-blocking error: your hook exits with any other code, or fails in some other way, such as not starting or printing invalid JSON. The action goes ahead, except for hooks whose failure blocks the action…". También: "Exit code 1 without valid JSON output is a non-blocking error…".
- *Timeout:* "On `PreToolUse`, a timed-out `command`, `http`, or `mcp_tool` hook doesn't block the tool call, except for hooks whose failure blocks the action." Tiempo por defecto: 600 s para `command`, `http` y `mcp_tool` (30 s para `prompt`, 60 s para `agent`); configurable con `timeout`.
- *Bloqueo explícito:* el código de salida 2 bloquea en los eventos que pueden bloquear ("a `PreToolUse` hook blocks the tool call, for example"). **No verificado en esta ronda:** la tabla de códigos de salida por evento no se releyó de forma textual; **(TBD)** por evento.
- *`onFailure`:* existe: la tabla de campos de `command` lo describe como "Whether the action that triggered the hook goes ahead when the hook fails: `"continue"` or `"block"`". La página nombra `"continue"` como valor por defecto solo en la tabla de campos de los hooks `http`; para `command` no nombra un valor por defecto, y la sección "Exit code output" describe que la acción sigue si el hook falla. "set `"onFailure": "block"` on a `command` or `http` hook. Requires Claude Code v2.1.295 or later." Para `mcp_tool`, `prompt` y `agent` la página no lo dice: **(TBD)**. No tiene efecto en `Stop`, `SubagentStop`, `TaskCompleted`, `TeammateIdle` ni en hooks de comando en segundo plano.
- *Bypass:* "PreToolUse hooks run before every tool call, whether or not it needs permission." Si un `PreToolUse` sigue ejecutándose o puede bloquear específicamente bajo `bypassPermissions`, la página **no lo dice**: **(TBD)**, se mide (D10). Las reglas deny y ask "are still evaluated regardless of what the hook returns".
- *Gestión (no verificado en esta ronda; recogido de la lectura previa de la página, sin releerlo textual):* `disableAllHooks` en configuración de usuario, proyecto o local no desactivaría hooks gestionados, y `allowManagedHooksOnly` bloquearía los hooks de usuario, proyecto, local y plugin. **(TBD)** hasta releerlo.

**Codex** — fuente: `https://developers.openai.com/codex/hooks`, que redirige (308) a `https://learn.chatgpt.com/docs/hooks`, y la referencia de configuración. Sin fecha y sin versión de Codex en las páginas.

- *Existencia:* "Hooks are an extensibility framework for Codex", "enabled by default". Se configuran en `hooks.json` o en una tabla `[hooks]` de `config.toml`, junto a las capas de configuración activas (`~/.codex/…` y `<repo>/.codex/…`). Eventos, entre otros: `PreToolUse`, `PermissionRequest`, `PostToolUse`, `Stop`, `SessionStart`. Desde qué versión de Codex y si aplican a la extensión de IDE: **(TBD)**; la página nombra la CLI (`/hooks`) y la app de escritorio.
- *Confianza:* "Non-managed hooks must be reviewed and trusted before they run"; "new or changed hooks are marked for review and skipped until trusted"; "Codex records trust against the hook's current hash". Un hook que el instalador escribe **no se ejecuta** hasta que el usuario lo confíe. "Project-local hooks load only when the project `.codex/` layer is trusted."
- *Fallo:* "Exit `0` with no output is treated as success and Codex continues." Para un `PreToolUse` local, la documentación afirma solo esto: ante salida **no soportada o inválida**, "Codex marks that hook run as failed, reports the error, and continues the tool call." El comportamiento ante **error o timeout** de un hook local de Codex **no está afirmado: (TBD)**, se mide. La frase "A `PreToolUse` error, timeout, or malformed response can let the tool run" pertenece a la sección de orquestación en la nube / hooks gestionados, no al `PreToolUse` local, y no se generaliza. Tiempo por defecto: 600 s. El comportamiento de otros eventos ante fallo o timeout tampoco está dicho (TBD).
- *Alcance:* "Treat tool hooks as a useful guardrail, not a complete enforcement boundary." Las herramientas alojadas, como `WebSearch`, no pasan por `PreToolUse`/`PostToolUse`, y "Some specialized tool paths can opt out of the default hook path".
- *Gestión:* hooks gestionados desde `requirements.toml`, sistema, MDM o nube; "Codex doesn't distribute the scripts in `managed_dir`".

**Consecuencias para la política (propuesta de partida, TBD del grupo).** (1) En Claude, un hook de bloqueo necesita `onFailure: "block"` (v2.1.295 o posterior) para fallar cerrado; sin eso falla abierto. El piso real de bloqueo sigue siendo `permissions.deny`, que no depende de hooks. (2) En Codex, la documentación solo afirma que ante salida no soportada o inválida de un `PreToolUse` local el hook se marca como fallido y la llamada continúa; el comportamiento ante error o timeout de un hook local es (TBD, se mide). Hasta medirlo, un hook de Codex no se toma como barrera única y se apoya en el sandbox y la política de aprobación (01). (3) Para validaciones posteriores, abierto con aviso. El roadmap marca el punto como pendiente de acuerdo; esta propuesta no lo cierra.

### D7 — Regla de control obligatorio ausente (acordada)

Acuerdo del informe unificado (07/08): si falta un **control obligatorio**, se detiene **solo la acción afectada**, se explica qué protección falta y se reanuda cuando hay una **protección equivalente verificada**. Un **control opcional** ausente produce una **advertencia** y se continúa.

- 08 define el *contenido* del mensaje de control ausente y qué cuenta como protección efectiva (03 D12, `openspec/changes/archive/2026-11-13-claude-native-opsx/design.md:346`: "Mensaje de control ausente y protección efectiva | 08"). 03 solo fija dónde aparece (respuesta del coordinador en el hilo, `design.md:333-335` del mismo archivo) y su regla para el **módulo** obligatorio ausente (D11 de 03, `:322-332`); 04 la replica en `openspec/specs/codex-opsx-delivery/spec.md:64-78` (requisito «Controles y módulos faltantes» y sus tres escenarios; en el delta archivado, `openspec/changes/archive/2026-11-13-codex-native-opsx/specs/codex-opsx-delivery/spec.md:63-77`). 08 no redefine ninguna de las dos.
- Qué controles son obligatorios y cuáles opcionales es una decisión del grupo (TBD). Propuesta de partida: el piso `deny` y la política de Estricto son obligatorios; un hook posterior es opcional.
- *Protección equivalente verificada* significa una comprobación hecha por el método de la matriz (D3), no la presencia de una clave en un archivo.
- Casos de aceptación: control obligatorio presente; obligatorio ausente (se detiene solo esa acción); opcional ausente (advertencia). Están en el spec y en las tareas.

### D8 — Revisión antes de escribir y reversión sin pisar al usuario

El instalador ya ofrece `install --dry-run` y el planner genera un review payload. El piloto agrega a esa vista el **diff exacto** de permisos y hooks. Ninguna escritura ocurre sin backup previo; se preservan claves y perfiles ajenos.

**Reversión: se reutiliza el `rollbackGuard` de 01 y 04, no se reimplementa.** `Manifest.GuardRollback(path, expected [sha256.Size]byte)` (`internal/backup/restore.go:22-29`) asocia a un snapshot el sha256 de los bytes que escribió el instalador; `Restore` (`:49-99`) compara el archivo actual y, si cambió después de la instalación, lo **conserva**, nombra el backup y devuelve el conflicto (`:64-66`). 04 sumó `GuardRollbackAbsent` (`:31-37`, protege un archivo retirado contra su recreación) y `GuardPersistentRestore` (`:39-47`; guardas persistidas en el manifiesto, `internal/backup/persistent_restore.go`), y la desinstalación tiene su propia reversión acotada (`internal/uninstall/managed_rollback.go:45-58`). **Quién registra guardas hoy** (leído en `d5e95bd`, no supuesto): (a) el paso de permisos, solo para el `config.toml` de Codex en Estricto: `permissions.Install` llena `Result.CodexStrictWrites` únicamente con `tier == model.TierEstricto` (`internal/harness/config/permissions/install.go:144-148`) y `permissionsStep.Run` las registra (`internal/install/steps.go:371-378`); (b) la entrega administrada: los archivos de `ManagedWrites` (`internal/install/steps.go:223-230`; `internal/harness/config/managed_skill.go:339-370`, que cubren skill, inventario y archivo de instrucciones) y su guarda persistente (`:370`); (c) `CLAUDE.md` en la entrega por capas (`internal/harness/config/layered_install.go:297`); (d) la reversión de `uninstall` para el archivo de instrucciones y las rutas propias (`managed_rollback.go:45-58`). Ninguna de esas rutas de código registra `~/.claude/settings.json` ni el `config.toml` de Codex fuera de Estricto (revisión de todas las llamadas a `GuardRollback*` y `GuardPersistentRestore` fuera de pruebas). Por lo tanto, **"protege solo Codex Estricto y `CLAUDE.md`" ya no es exacto** (ahora también cubre los archivos de la entrega administrada de Claude y de Codex), pero **sigue siendo cierto que el `settings.json` de Claude no está protegido**. Además, la restauración por manifiesto excluye `settings.json` de los agentes de entrega administrada cuando el manifiesto es el de la migración registrada (`internal/uninstall/plan.go:54-59`, `internal/uninstall/steps.go:401-404`). Registrar `settings.json` en el paso de permisos y de hooks es una decisión para el revisor ALTO (Neyén). Si se decide, se registra con la misma guarda (sin reimplementarla) cada archivo que se escriba (`~/.claude/settings.json`, `~/.codex/config.toml` y los archivos de hooks). Hay un hueco que la tarea 3.4 debe cerrar: el paso de desinstalación de `permissions` hoy **no hace nada** (`internal/uninstall/steps.go:362-384`, con el comentario "(TBD) There is no dedicated mechanism yet to selectively strip just the permissions overlay" en `:363`; `Run` devuelve `nil`), así que la remoción de claves necesita un mecanismo propio y acotado. Quién lo implementa y revisa: D15.

### D9 — Preservar las entradas del usuario y no pisar un archivo ilegible

**Lo que hace hoy `MergeJSONObjects`** (`internal/filemerge/json_merge.go`, `origin/main`):

- Solo los objetos se mezclan de forma recursiva (`:224-226`); **los escalares y las listas del overlay reemplazan a los del usuario** (`:229`), salvo el sentinel `__replace__` (`:180`, `:204-206`), que hace lo mismo de forma explícita. Para `permissions` eso significa que `defaultMode`, `deny` y `allow` del usuario pueden perderse (el overlay fija `defaultMode` en `overlays.go:77-83`).
- **Un `settings.json` malformado se trata como `{}`:** si no se puede interpretar ni tras quitar comentarios y comas finales, `base = map[string]any{}` (`:10-15`, con el comentario "Malformed base is treated as {} — overlay prevalence is safe and preferable to aborting the install"). Resultado: el contenido del usuario se descarta y se escribe solo el overlay.
- Los comentarios del archivo se eliminan al leer (`:49-52`) y el resultado se vuelve a serializar sin ellos (`:23`).

**Contraste:** `UpsertTOMLTopLevel` valida el TOML antes de tocar nada y devuelve error si es inválido (`internal/filemerge/toml.go:12-21`, `:26-28`), y 01 pide "ante TOML inválido, abortar e informar el diagnóstico con el archivo intacto" (`codex-strict-compat/design.md:33`).

**Requisito de este change:** no se pasa por encima de un archivo que no se puede leer. Un `settings.json` existente que no se pueda interpretar aborta la escritura, deja el archivo intacto e informa la ruta y el motivo; nunca se trata como `{}`. Arreglarlo toca `filemerge` (ALTO) y se decide en el grupo (D15). Antes de sumar hooks al mismo archivo hay que resolver también cómo se mezclan las **listas** (unión con ownership, o reemplazo solo de lo propio) y cómo se trata un escalar que el usuario fijó (por ejemplo `defaultMode`). Se parte de reproducir el caso del 02/11 de forma aislada (tarea 1.4) y de una prueba con un archivo malformado (tarea 4.5).

### D10 — Relación con los niveles de permisos

Hay que definir si los hooks dependen del nivel (Estricto, Balanceado, Bypass) o son independientes. En Claude, el piso `deny` se aplica en todos los niveles; en Bypass el resto se auto-aprueba (`overlays.go:79-81`). La documentación no dice si un `PreToolUse` bloquea bajo `bypassPermissions` (D6, TBD): se mide. En Codex, Bypass es `never` + `danger-full-access` (`install.go:124-125`), sin sandbox, y los hooks son "a useful guardrail". Si Bypass desactiva o debilita los controles de bloqueo, no puede usarse como base de la validación.

### D11 — Codex: la política y el método de aceptación son hechos de `main` (01)

01 está integrado. El piloto **no** define ni espera la política de Estricto de Codex: usa la que está en `main` (`sandbox_mode = "read-only"`, `approval_policy = "on-request"`, `install.go:120-126`) y reutiliza su método de aceptación (`openspec/changes/codex-strict-compat/design.md:22`; `tasks.md:12`, tarea 2.1, completa):

- perfil aislado y cliente real, nunca el `~/.codex` del usuario;
- una lectura permitida y una escritura de prueba con la escalación denegada;
- verificar el estado del archivo después del intento y registrar superficie, versión y resultado;
- Windows con Codex CLI 0.166.0 es la primera superficie de aceptación; las demás versiones e interfaces están **(TBD)** en 01 (`design.md:41`) y el piloto las hereda. Terminal interactiva y `codex exec` se registran por separado (`design.md:43`: en `codex exec` la sesión observada aplicó `approval: never` con `sandbox: read-only`).

Lo propio de 08 son las celdas que 01 no cubre: los hooks y la política ante fallo (D6, D7). Los resultados de 01 y los de la matriz comparten método y se registran una sola vez.

### D12 — Claude: activación explícita de 03, sin `agent` ni `--agent`

03 está en `main` (archivado en `openspec/changes/archive/2026-11-13-claude-native-opsx/`). El coordinador se activa por **skill explícita**: `/sdd-orchestrator` en Claude (`design.md:150-154`, decisión G2 en `:354`) y `$sdd-orchestrator` en Codex (04). La skill lleva `disable-model-invocation: true`. No se usa `claude --agent` ni el ajuste `agent` de `settings.json`: "El instalador sigue sin escribir `agent`; se conserva sólo como regresión" (`design.md:164`); `--agent` y el setting quedaron como alternativa descartada (`:168-169`).

**No colisión.** Las claves que escribe este change en `settings.json` son `permissions.defaultMode`, `permissions.deny`, `permissions.allow` y, si el piloto los instala, `hooks`. La activación de 03 es una skill en `<skills>/sdd-orchestrator/` más un bloque en `CLAUDE.md`; no toca `settings.json`, que 03 solo lee y no entra a su manifiesto de backup (`design.md:208`). 03 descarta los hooks `SessionStart` (`:168-169`), así que la clave `hooks` sigue libre. La clave `agent` del usuario, si existe, debe quedar idéntica tras la escritura (regresión).

**Punto de contacto: `WriteFileAtomic`.** `permissions.Install` escribe con `filemerge.WriteFileAtomic` en `install.go:143` (Codex) y `:185` (Claude), ambos con modo `0o644`. `WriteFileAtomic` sigue llamando a `writeFileAtomic` con `repairParent == true` (`internal/filemerge/writer.go:16-18`): crea el directorio con `0o755` y hace `os.Chmod(dir, 0o755)` sobre el padre en cada escritura que cambia (`:43-58`), y aplica a la temporal el modo que pasa el llamador (`tmp.Chmod(perm)`, `:78-81`; `perm == 0` pasa a `0o644`, `:27-29`). 03 lo documentó (D6, `openspec/changes/archive/2026-11-13-claude-native-opsx/design.md:246-254`) y por G6 (`:366`, decisión de Juani, 2026-11-13) eligió un alcance chico: solo las rutas de 03 conservan su modo, **sin cambios en `internal/filemerge` global**. **Cambio respecto de la base anterior (`7a60089`):** 04 agregó `WriteFileAtomicPreservingParent` (`writer.go:20-24`: no cambia el modo de un directorio existente y crea los que faltan con `0o700`), que usan la entrega administrada (`internal/harness/config/managed_skill.go:26`, que conserva el modo del archivo existente, `:354-358`) y la restauración de respaldos (`internal/backup/restore.go:198`). `WriteFileAtomic` no cambió de comportamiento y la escritura de permisos de 08 no usa el helper nuevo. Consecuencia para 08: una escritura de permisos todavía puede ampliar un `~/.claude` o `~/.codex` en `0700` y reemplazar el modo `0600` de un archivo. 04 dejó anotado que el helper "puede evaluarse como mecanismo común" (`openspec/changes/archive/2026-11-13-codex-native-opsx/informes/respuesta-bloqueantes-pr9-2026-11-13.md:71`); adoptarlo en la ruta de permisos es una decisión del revisor ALTO (TBD). Se mide en Windows y Linux (D13), y 03 D6 (`:256-259`) pide que este arreglo y el de D9 se integren serializados con un único revisor ALTO (D15).

### D13 — Inventario de escrituras y tabla de cobertura

**Escrituras que el change haría** (ninguna en esta ronda). El nivel lo marca el paquete (`CLAUDE.md` §4, `:122-130`: `filemerge`, `backup` y `pipeline` son ALTO; `agents` y `harness/*`, MEDIO). Como todas las escrituras de configuración del usuario pasan por `filemerge`/`backup`, se marcan ALTO.

| # | Archivo | Qué se escribe | Función | Nivel |
|---|---|---|---|---|
| 1 | `~/.claude/settings.json` | overlay de permisos (existente) | `filemerge.MergeJSONObjects` + `WriteFileAtomic` (`install.go:180,185`) | ALTO |
| 2 | `~/.claude/settings.json` | clave `hooks` (nueva, si el piloto la instala) | ídem | ALTO |
| 3 | `~/.codex/config.toml` | `approval_policy`, `sandbox_mode` y, en Balanceado, `sandbox_workspace_write.network_access` | `UpsertTOMLTopLevel`, `UpsertTOMLTableValues` + `WriteFileAtomic` (`install.go:127-143`) | ALTO |
| 4 | `~/.codex/hooks.json` o `[hooks]` en `config.toml` (TBD según D4) | hooks de Codex (nuevo) | `UpsertTOMLSection` (`toml.go:80`) o escritura JSON + `WriteFileAtomic` | ALTO |
| 5 | `~/.active-stack/backups/permissions-<agente>/` | snapshot previo | `backup.NewSnapshotter().Create` (`install.go:17-20,164-166`) | ALTO |
| 6 | registro de ownership junto a los backups (si D4 lo elige) | qué entradas son del stack | `backup` / `filemerge` | ALTO |
| 7 | scripts de handlers (solo si D5 llega a la opción 3) | ejecutables en el HOME del usuario | por definir | (TBD) |
| 8 | capas de proyecto (`<repo>/.claude/`, `<repo>/.codex/`) | fuera de alcance salvo acuerdo; Codex las carga solo si el proyecto es confiable | — | — |

**Riesgos a medir en Windows y Linux** (tarea 4.8): (a) `UpsertTOMLTopLevel` convierte todo `\r\n` en `\n` en el archivo completo (`toml.go:29`) y recorta con `TrimSpace` (`:72`): un `config.toml` con CRLF cambia de fines de línea al escribir; `RemoveTOMLSection`, que usa `UpsertTOMLSection`, hace lo mismo (`:103`). (b) `WriteFileAtomic` hace `chmod 0755` del directorio padre (`writer.go:54-58`) y deja el archivo en el `0644` que pasa `permissions` (`install.go:143,185`; aplicado en `writer.go:78`). (c) `MergeJSONObjects` re-serializa y pierde los comentarios (`json_merge.go:23`). (d) en Windows el significado de `chmod` es limitado (TBD, se mide). Cada medición: archivo de prueba con fines de línea y modo conocidos en un `HOME` temporal, escritura, comparación de bytes y de modos.

**Formato de la tabla de cobertura (para 10).** Un archivo de filas, una por celda de la matriz, con un conjunto cerrado de estados, para que un juez externo la lea sin interpretar prosa:

| Columna | Valores |
|---|---|
| `regla` | id estable de la regla |
| `clase` | `orienta` / `restringe` / `verifica` |
| `momento` | `previo` / `posterior` |
| `cliente`, `version`, `superficie` | texto (por ejemplo `codex`, `0.166.0`, `cli`) |
| `herramienta`, `shell`, `so` | texto (por ejemplo `Bash`, `powershell`, `windows`) |
| `mecanismo` | `permiso` / `sandbox` / `hook` / `ninguno` |
| `ante_fallo` | `bloquea` / `continua` / `no_medido` |
| `estado` | `cubierta` / `no_cubierta` / `no_probada` |
| `evidencia` | referencia a la corrida (método de 01 y arnés de 02) |
| `fecha` | fecha de la corrida |

Una celda sin evidencia nunca figura como `cubierta`.

### D14 — Entradas que el change instala: dueño, reversión y precedencia

El piloto (medición en un `HOME` temporal, sin escribir en el entorno real) es distinto de la instalación permanente (solo después de la revisión humana ALTO y de un criterio de adopción, TBD). Para la instalación permanente, por cada entrada:

| Entrada | Dueño | Reversión | Precedencia frente al usuario | Bypass | Actualización desatendida | Restauración sin pisar ediciones |
|---|---|---|---|---|---|---|
| `permissions.*` en `settings.json` | stack para las claves que fija; hoy sin ownership por entrada (D9) | hoy `uninstall` no la quita (`steps.go:362-384`); necesita remoción acotada | el overlay reemplaza escalares y listas del usuario (`json_merge.go:229`); la precedencia entre capas (usuario/proyecto/local/gestionada) no se verificó: (TBD) | el piso `deny` rige en todos los niveles (`overlays.go:9-12`) | `update` reaplica la selección registrada; sin TTY no se puede pedir aprobación: (TBD) omitir y avisar | `GuardRollback` por archivo (D8) |
| `approval_policy`/`sandbox_mode` en `config.toml` | stack; política de 01 | rollback de 01 (`install.go:144-148`) | 01 limita la edición a la selección explícita de Estricto o su reaplicación registrada (`design.md:27`) | Bypass = `never` + `danger-full-access` (`install.go:124-125`) | replay de `update` con selección registrada (01, `design.md:35`) | guarda de 01 |
| `hooks` en `settings.json` | (TBD) ownership por identificador o registro (D4) | quitar solo entradas propias; requiere el mecanismo de D8 | las del usuario se preservan; el orden de ejecución entre hooks de distintas capas: (TBD) | si se ejecutan bajo Bypass: (TBD, D6) | igual que arriba | `GuardRollback` |
| hooks de Codex | (TBD) | ídem | un hook nuevo o cambiado se omite hasta que el usuario lo confíe (hash); la capa de proyecto exige proyecto confiable | los hooks son "a useful guardrail, not a complete enforcement boundary" | el cambio de un hook invalida su confianza y vuelve a pedir revisión | `GuardRollback` |

### D15 — Revisor ALTO único para `filemerge`

El PR de este change toca (o puede tocar) `filemerge`, nivel ALTO. **Lo revisa Neyén**, según el reparto, y **cualquier cambio a `filemerge` requiere ese revisor humano**. Coincide con 03 D6 (`design.md:256-259`: los arreglos de `WriteFileAtomic` y de `MergeJSONObjects` se integran "serializados, con un único revisor ALTO") y con `CLAUDE.md` §4 (ALTO: se propone y se espera review). Si el diseño final exige tocar `backup` o `pipeline`, se frena y se reporta.

### D16 — Plataformas

Windows y Linux son el mínimo. La matriz distingue tres columnas por plataforma:

| Columna | Qué significa | Fuente | Estado hoy |
|---|---|---|---|
| *distribuida* | el instalador publica binario | `.github/workflows/release.yml:71-76` (`GOOS=windows` `:71-73` y `GOOS=linux` `:74-76`, ambos `amd64`) | Windows amd64, Linux amd64 |
| *declarada* | la documentación dice que se soporta | `CLAUDE.md:13` ("Windows, macOS, Linux, WSL y Termux") | las cinco |
| *probada* | lo que efectivamente se corrió | tarea 4.9 | (TBD): ninguna todavía; 01 registró Windows CLI 0.166.0 para su aceptación (`codex-strict-compat/design.md:41`) |

macOS, WSL y Termux figuran como *declaradas*; no hay binario publicado para ellas en `release.yml` y no se promete aquí que estén probadas.

## Risks / Trade-offs

- **Falsa sensación de seguridad.** Un control parcial hace creer que todo está protegido; en Codex la propia documentación lo llama "guardrail". Mitigación: límites documentados y pruebas de rutas alternativas.
- **Configuración equivocada en la máquina del usuario** (ya observada el 02/11). Mitigación: diff previo, backup obligatorio, preservación de entradas ajenas, negativa a escribir sobre un archivo ilegible, reversión con guarda.
- **Hooks propios mezclados con los del usuario.** Mitigación: ownership explícito y pruebas de preservación.
- **Handler roto que detiene el trabajo al fallar cerrado.** Mitigación: pruebas de handler ausente y timeout, y mensajes claros.
- **Un hook de Codex escrito no corre** hasta que el usuario lo confíe (hash). Mitigación: la matriz mide "instalado" y "ejecutado" por separado.
- **Diferencias entre versiones, shells y superficies.** Mitigación: matriz por combinación.
- **Superposición con 01.** Ambas verifican permisos efectivos de Codex. Mitigación: se reutiliza el método de 01 (D11) y los resultados se registran una sola vez.
- **Deuda de desinstalación.** El paso de `permissions` de `uninstall` no quita nada (`steps.go:362-384`) y quitar un MCP externo no borra su entrada de `~/.claude.json` (`docs/archive/CHANGES-2026-11-05-post-starters.md:511`). Mitigación: requisito del diseño, no mejora posterior (D8).
- **Tocar `filemerge`** puede afectar a todos los harnesses. Mitigación: cambio mínimo, suite completa y revisión humana ALTO (D15).

## Migration Plan

No hay migración en esta ronda: no se escribe nada. Para la etapa de implementación, los controles entran con snapshot previo, son idempotentes al reinstalar, forman parte de `update` y se quitan por completo con `uninstall` (con el mecanismo acotado de D8). Una instalación existente sin controles sigue funcionando igual hasta que el usuario acepte el diff.

## Open Questions

Quedan vivas las que dependen del grupo o de otras propuestas; ninguna se cierra aquí.

1. **Política ante fallo** (TBD, grupo; punto pendiente del roadmap): D6 documenta lo que hace cada cliente; falta decidir la política.
2. **Alcance de handlers** (TBD, grupo): hasta qué opción de D5 se pilotea.
3. **Ownership de hooks** (TBD, grupo y revisor ALTO): identificador dentro de la entrada o registro aparte (D4).
4. **Bypass** (TBD, se mide): si un `PreToolUse` de Claude bloquea bajo `bypassPermissions` (D6, D10).
5. **Hooks en Codex**: desde qué versión y en qué interfaces (TBD, inventario 2.1); la documentación solo confirma que existen.
6. **Listas y escalares del usuario, y archivo malformado** (D9) y quién arregla `filemerge` (TBD, grupo y revisor ALTO).
7. **Qué controles son obligatorios y cuáles opcionales** (D7, TBD, grupo con 03 y 04).
8. **Actualización desatendida** sin posibilidad de aprobar el diff (D14, TBD).
9. **Remoción acotada de `permissions` en `uninstall`** (D8): quién la implementa (TBD).
10. **Mínimo de cobertura** para considerar adoptable el piloto (TBD, grupo).
11. **Nombre de la rama:** hoy es `opsx-native-controls`; renombrarla a `propose/opsx-native-controls` se decide con Juan Cruz (TBD).

Resueltas o pasadas a otro change: *agentes incluidos* (el roadmap fija "Ambos", Claude y Codex: `CHANGES.md` ficha 08), *política de Estricto de Codex* (01, en `main`), *activación de 03* (skill explícita, en `main`).
