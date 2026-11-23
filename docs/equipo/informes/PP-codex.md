# PP-01 a PP-04 · Codex real en Windows

**Fecha:** 2026-11-23 · **Responsable:** Juan Cruz (dueño), con agente · **Rama:** `feat/PP-codex-real` · **Base:** `origin/main@eb98fc80a28f66d98405cc57f0b8067c36a146a3`.

**Resultado:** las cuatro PP siguen abiertas. Hay evidencia real favorable de lectura y etiquetado, pero fallan la confianza tras reinstalar, la denegación de push, la desinstalación Full y la coexistencia del plugin. No se declara la garantía completa de Codex (DA-12).

## Entorno y alcance

Windows, Codex CLI **0.166.0**, Engram **3.7.2**, plugin Codex Engram **0.1.5**. Active Stack del PATH informa `dev`; su SHA256 queda en [environment.json](../../research/pp-codex/environment.json). No se recompiló ni se reemplazó el binario. El código probado con Go parte de la base indicada; el binario instalado no identifica su commit.

Worktree: `E:/pp-codex-work`; laboratorio: `E:/pp-codex-lab`, ambos fuera del perfil. Cada proceso Active Stack/Codex recibió `USERPROFILE`, `HOME`, `LOCALAPPDATA`, `APPDATA` y `CODEX_HOME` mediante `subprocess.run(..., env=...)`, o variables inline en el comando de la TUI. Ningún shell de esas corridas ejecutó `gh`. Los comandos de GitHub usaron un proceso separado, sin cambiar su HOME.

Se instaló **Full, solo Codex**, con `--no-self-install`. Se capturó el resultado antes de preparar el MCP local. Para V1–V13 se usó una configuración de laboratorio con el proxy real, rutas absolutas de los ejecutables, base aislada, puerto `17439` y `ENGRAM_CLOUD_AUTOSYNC=0`. Es una prueba local: **cloud, dashboard, dos personas y otros sistemas quedan (TBD)**. No se copiaron credenciales de Engram. Los repos y remotos bare son descartables.

Se leyeron las cinco fuentes pedidas, CHANGES y los planes. No se editaron `docs/estado-equipo.md`, `internal/backup`, `internal/filemerge` ni `internal/pipeline`. No hubo merge.

## PP-01 · X0–X9

La tabla se completó también en la [ficha R-16](../fichas/R-16.md). El [ayudante real del instalador](../../research/pp-codex/hooks-manual.json) se ejecutó con adapters y snapshots reales, en dos HOME exclusivos para esa prueba; no se lo presenta como ejecución del agente.

| Caso | Resultado observado |
|---|---|
| X0 | Codex 0.166.0; `CODEX_HOME=E:/pp-codex-lab/home/.codex`, coincide con `<HOME>/.codex`. |
| X1 | `installed=false missing=[session-start pre-read]`, binario en PATH. |
| X2 | `changed=true`, aviso de revisar/confiar; solo se agrega el bloque al fixture del ayudante. |
| X3 | Codex real responde `AUSENTE` antes de confiar; `hooks/list` muestra ambos `untrusted`. El [panel de revisión de la TUI](../../research/pp-codex/X4-tui-review.txt) muestra dos hooks pendientes antes de poder ejecutarlos. |
| X4 | Revisión en la TUI y confianza persistida; `PreToolUse` y `SessionStart` pasan de 0 a 1 activos. [Captura cruda](../../research/pp-codex/X4-tui-trust.txt). |
| X5 | El hook ejecuta la cadena en Windows: contexto `Team memory` recibido. La primera corrida decía `untagged` porque el fixture todavía no tenía remoto/default branch; tras agregar el bare y `origin/HEAD`, [V1-X7](../../research/pp-codex/V1-X7.json) recibe `feat/prueba-r16 (slug: feat-prueba-r16-107728)`. Shell concreto usado internamente por el hook: **(TBD)**. |
| X6 | **FALLA** en el instalador de la base: tras confiar, reinstalar da `changed=true` y elimina `hooks.state`/`trusted_hash`. [Antes](../../research/pp-codex/config-before-reinstall.toml), [después](../../research/pp-codex/config-after-reinstall.toml), [corrida](../../research/pp-codex/hooks-main-reinstall.json). Codex vuelve a listar los hooks propios como `untrusted`. Corrección con TDD en `e89fe83`; no se recompiló ni se repitió E2E con ese código. |
| X7 | `mem_search` real funciona y el agente reproduce `Current branch: feat/prueba-r16 (slug: feat-prueba-r16-107728).` |
| X8 | **FALLA** con el plugin real: el hook bloquea `mem_save` por sesión no registrada. Ver V10. No se pudo confirmar vínculo de sesión y guardado. |
| X9 | El ayudante instala y retira, recupera los **57 bytes originales** y el SHA256 `2fd944be98d0198b0d4b3a6d80a6681089def41f35a77ec6a03d5b7a71f7c500`; también sin PATH. Retiro observado desde `/hooks` en el HOME principal: **(TBD)**. Esto no subsana el fallo de Full en PP-03. |

Sin binario en PATH: el ayudante detecta `active-stack-in-PATH=false`, instala con aviso y retira sin alterar el fixture. Con Codex real y el MCP configurado por nombre, el servidor no inicia (`program not found`) y el agente avisa que no guardó; no se creó una observación V8. El diagnóstico visible de `verify` sigue sin cubrirse: `install` informa cero checks.

## PP-02 · Denegación de push a main

**FALLA.** R-18 declara que todavía no implementó reglas `deny` de Codex. Se probó la configuración capturada de Full, sin agregar reglas manuales. Para distinguir reglas de denegaciones genéricas del sandbox, se sobreescribieron únicamente modelo/esfuerzo, `approval_policy=never` y `sandbox=danger-full-access`; no se usó `--ignore-rules` ni el flag de bypass de aprobaciones.

Codex ejecutó exactamente `git push origin main`: salida **0**, `main -> main`. Había un commit adelantado para que no fuera un push vacío.

| main remoto antes | HEAD local y main remoto después |
|---|---|
| `1fe3d6fa107a22594a3fc2828f8bd59c115e00c3` | `2f55e9f33882b6aa19b75c34a3917962b1ed43cf` |

[Sesión real](../../research/pp-codex/PP02-push.json) · [Comparación independiente del bare](../../research/pp-codex/PP02-remote.json). No es una prueba de todas las variantes ni de protección de GitHub. Implementar una política nueva requiere un cambio propio; no se agregó en esta prueba.

## PP-03 · Install y uninstall

**Instalación PASA; reversibilidad FALLA.** `active-stack install --mode full --agent codex --yes --home E:/pp-codex-lab/home --no-self-install` termina en 0 y deja `# active-stack:memory-proxy`, los dos hooks y `ENGRAM_CLOUD_AUTOSYNC="1"`. Se partió de un `config.toml` con contenido ajeno y una entrada MCP Engram existente.

`active-stack uninstall --mode full --agent codex --yes --home ...` termina en **1**: `found 0 begin and 1 end marker lines, want exactly one of each`. El rollback también informa fallo al preservar un AGENTS que había cambiado. El archivo final **no** coincide byte a byte con el original. No se hizo restore para aparentar que pasó; la preparación posterior del laboratorio usó otro contenido y está registrada como tal.

[Install](../../research/pp-codex/install-full.json) · [Uninstall](../../research/pp-codex/uninstall-full.json) · [Original](../../research/pp-codex/config-original.toml) · [Instalado](../../research/pp-codex/config-installed.toml) · [Tras retiro fallido](../../research/pp-codex/config-uninstalled.toml). Los hashes están en `environment.json`. No se corrigió: requiere revisar la interacción del retiro de externos, hooks y rollback, fuera del arreglo chico X6 y de los paquetes autorizados.

## PP-04 · Matriz V1–V13

Las filas de la [base aislada](../../research/pp-codex/database-evidence.json) corroboran las escrituras del agente, además de sus respuestas. Una sesión, credencial y autor reales; dos worktrees no equivalen a dos personas.

| Caso | Resultado y evidencia |
|---|---|
| V1 | **PASA:** `[feat-prueba-r16-107728] V1 rama`, clave `branch/feat-prueba-r16-107728/v1`, `Branch/Author`. [V1-X7](../../research/pp-codex/V1-X7.json). |
| V2 | **PASA:** `git switch -c feat/nueva` y `git checkout -b feat/checkout` desde el agente en la misma sesión; cada guardado toma la rama nueva. [V2V3](../../research/pp-codex/V2V3.json). |
| V3 | **PASA:** vuelve a main, título `V3 main`, clave `v3`, solo `Author: PP Codex`. Mismo log. |
| V4 | **PASA:** dos Codex en paralelo, worktrees `feat/wa` y `feat/wb`, misma clave de entrada `v4shared`, dos claves finales distintas. [wa](../../research/pp-codex/V4-wa.json), [wb](../../research/pp-codex/V4-wb.json). |
| V5 | **PASA sin `--ephemeral`:** subagente real guarda observación 13, `[feat-wa-f42070] V5 subagente`, `Branch/Author`. Primer intento efímero falló por `no rollout found`; no se ocultó. [Reintento](../../research/pp-codex/V5V6-retry.json), [intento inicial](../../research/pp-codex/V5V6.json). |
| V6 | **PASA para las vías ejecutadas:** resumen con `Branch/Author`; prompt con prefijo y `Branch/Author`; captura pasiva de dos ítems con prefijo de rama en título/contenido (sin línea `Branch/Author` separada). El primer texto pasivo no tenía `Key Learnings` y guardó 0; se repitió con el formato correcto y guardó 2. `mem_session_finalize`: **(TBD)**. Logs V5V6 y base aislada. |
| V7 | **PASA:** sin marcador, título `V7 sin marcador`, clave `v7`, contenido sin reescribir. [V7](../../research/pp-codex/V7.json). |
| V8 | **PARCIAL:** sin PATH no arranca MCP y no guarda; hooks sin confianza no ejecutan y `/hooks` los muestra pendientes. **LÍMITE:** no hay diagnóstico CLI visible de verify, ni verificación de confianza en él. [V8](../../research/pp-codex/V8-no-path.json), X3 y `install-full.json`. |
| V9 | **PARCIAL:** marcador inválido no bloquea, guarda sin etiqueta. En stderr capturado de Codex no aparece aviso del proxy. El hook directo sí avisa y devuelve 0 con el binario actual. Timeout: **(TBD)**. [Agente](../../research/pp-codex/V9.json), [hook directo](../../research/pp-codex/V9-direct-hook.json). PP-14 sigue abierta. |
| V10 | **FALLA:** plugin real 0.1.5 instalado mediante marketplace local con manifiesto y paquete existentes, sin sustituir hooks. La TUI confió los ocho hooks; `hooks/list` confirmó hooks del plugin y propios. `PreToolUse` bloquea guardado: `Codex host session is not registered; start a new Codex session to register it`. Repite sin `--ephemeral` y con marcador nativo Windows correcto. No hay nueva observación V10. [Primera sesión](../../research/pp-codex/V10-plugin.json), [nativa](../../research/pp-codex/V10-plugin-native.json), [registro](../../research/pp-codex/hooks-list-after-reinstall-plugin.json). Causa raíz **(TBD)**; no se desactivó el hook para darlo por aprobado. |
| V11 | **PASA:** `/` → `feat-una-1e8539`, guion → `feat-una`, nombre largo → `feat-abcdefghijklmnopqrstuvwxyzab-10f94b` (40 caracteres), detached → `detached-b330328`. Guardados reales distintos. [V11](../../research/pp-codex/V11.json). |
| V12 | **PARCIAL:** Windows probado. Linux/macOS/WSL/Termux: **(TBD)**; corresponden a PP-08. |
| V13 | **PASA:** desde main lee V1, reconoce rama ajena, decisión provisional y falta de evidencia de integración; no la toma como verdad de main. Se cargó la instrucción englobada real `engram-cloud.md` como AGENTS del proyecto. [V13](../../research/pp-codex/V13.json). |

### Escenario 6 · Claude → Codex

**PASA como traspaso histórico de archivos, no como sesión Claude nueva hoy.** Se clonó sin hardlinks el fixture producido por Claude en R-24 (`C:/r24b/P-01-pedidos`) a `E:/pp-codex-lab/handoff`; no se tocó el original. Codex recibió solo el pedido de retomar `knowledge-base/entrega/P-01.md`, leyó el handoff, contrastó Git/CHANGES y ejecutó `python -m unittest`: **5 tests OK**, sin pedir contexto ni implementar cambios. Detectó que CHANGES existe aunque el handoff diga lo contrario y que main avanzó. `origin` de esta copia es local: su referencia publicada no demuestra un push real a GitHub. [Salida](../../research/pp-codex/scenario6-Claude-to-Codex.json).

Nueva sesión de Claude, sentido Codex → Claude y traspaso por servidor cloud: **(TBD)**. Se reutilizó la evidencia Claude existente para no gastar otra corrida ni copiar sus credenciales.

## Arreglo chico separado · X6

Commit `e89fe834bb9cb1a58122e9367ad7dcff59802bc3`, `fix(hooks): preserve Codex trust on unchanged reinstall`. Si las definiciones son semánticamente iguales, ignora únicamente `hooks.state` al compararlas y conserva todos los bytes; una diferencia de comando, matcher o timeout sigue actualizando el bloque. No modifica los paquetes restringidos.

### TDD Cycle Evidence

| Task | Test File | Layer | Safety Net | RED | GREEN | TRIANGULATE | REFACTOR |
|---|---|---|---|---|---|---|---|
| X6 | `internal/harness/hooks/codex_test.go` | Integración del instalador | ✅ 38 tests + 37 subtests | ✅ pérdida de bytes/confianza/snapshot | ✅ pasó | ✅ LF/CRLF, timeout, matcher, comando y tabla extra | ✅ gofmt y suite |

Resultado: **40 tests y 43 subtests PASS; 2 SKIP** (manual sin variables y permisos POSIX). [Evidencia cruda](../../research/pp-codex/x6-tdd.txt). El binario del PATH conserva el comportamiento anterior: **revalidación E2E con un binario que incluya el fix (TBD)**. No se pidió ni ejecutó build.

## Costo y credenciales

Se intentó `gpt-5.1-codex-mini`, rechazado por la cuenta sin completar turno. Una corrida de control usó `gpt-6.1-sol`; las restantes completadas usaron **gpt-6-luna**, con esfuerzo bajo. Se reutilizó Claude histórico. Los JSON registran **18 corridas CLI con turno completado**, incluyendo reintentos y login: **1.579.001 tokens de entrada, 1.340.544 en caché, 5.337 de salida**. La entrada sin caché es **238.457**. Son los tokens observables de esas CLI; no incluyen inferencia del coordinador/ayudante ni uso interno del subagente que no se expone en esas salidas. La TUI advirtió menos del 10 % del cupo semanal restante; no permite atribuir qué parte del consumo total de la cuenta corresponde solo a estas pruebas.

**Costo monetario exacto: (TBD)**. Se usó login ChatGPT, no facturación API; la CLI no devuelve dólares y no corresponde inventar una conversión por tarifas API. [Desglose](../../research/pp-codex/cost.json).

Antes de copiar `auth.json`, se guardó `.bak` fuera del repo. No hubo rotación ni copia más nueva. Verificación final con HOME real: `codex exec ... "di ok"` respondió **`OK.`**, código 0. Se usaron `--ignore-user-config --disable plugins` para probar la autenticación sin arrancar MCP/hooks del perfil real. Solo después se borraron la copia temporal y el `.bak`; no se imprimieron ni versionaron tokens. [Ciclo de credenciales](../../research/pp-codex/credential-lifecycle.json), [login](../../research/pp-codex/real-login-final.json).

## Qué queda pendiente

- Revalidar X6 con el binario actualizado y completar shell del hook/X9 interactivo.
- Resolver la política deny de Codex y repetir PP-02 sin confundir sandbox con regla.
- Investigar retiro Full y rollback de PP-03, sin tocar los tres paquetes restringidos en este trabajo.
- Resolver registro de sesión del plugin; completar V8/verify, V9/timeout/aviso y V12. V6 finalize y traspaso inverso siguen (TBD).
- Cloud real, dos personas, dashboard y otros sistemas permanecen fuera de esta pasada.

**DA-12:** Codex + Windows demuestra etiquetado local y lectura con los límites de la matriz; no garantiza V1–V13 completos. Ninguna PP se marca aprobada.
