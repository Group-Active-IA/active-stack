# R-05 · Matriz Claude Code + Windows

Entorno: Windows 11, Claude Code 2.1.327 (`claude -p`, modelo `haiku`), Engram 3.7.2, Python 3.14.3, Git 2.45.1.windows.1. Base de Engram descartable por corrida (`ENGRAM_DATA_DIR`), proyecto sintético `r05-synthetic`, repos en rutas con espacios, sin plugins ni MCP de usuario (`--strict-mcp-config`, `--setting-sources project,local`). Arnés: `arnes.py`, `escenarios.py`; salidas crudas en `resultados/` y `q3-*.txt`. Ejecutado el 2026-11-19. n=1 por escenario salvo Q11 (6 pares) y salvo lo indicado.

## Límites que condicionan todo

- **El plugin de Engram instalado en esta máquina era 0.1.0 y no traía `PreToolUse`** (R-01 probó 0.1.5 en Codex). El plugin oficial de la etiqueta `v3.7.2` es la **0.1.22** y sí trae `PreToolUse` (`engram hook claude-pre-tool-use`, que ata el `session_id` autoritativo de Claude) sobre `mem_save`, `mem_update`, `mem_session_summary`, `mem_capture_passive` y otras. **Q3 se probó con ese plugin real** cargado con `--plugin-dir` desde un clon temporal (commit `c6ba9e3c`), sin tocar la instalación. Las corridas con hooks sustitutos (`sim_*.py`, `q3-*+sim.txt`) quedan como **antecedente superado**: su conclusión («el hook pisa el `session_id`») era un artefacto del sustituto. Después actualicé el plugin local 0.1.0 → 0.1.23 (ver «Plugin local»).
- `claude -p` no muestra el diálogo de confianza y trata la carpeta como confiable: **V8 (hook sin confiar) no ejecutada** (requiere sesión interactiva).
- La doc oficial no especifica cómo se combinan los `updatedInput` de varios hooks; lo observado vale para esta versión.
- Con `bypassPermissions` salvo Q9.

## Resultados

| Escenario | Resultado en Claude/Windows | Evidencia |
|---|---|---|
| Línea base | 14/14 tests; sonda Engram 3.7.2 coincide con 3.6.2 en Q2/Q4/Q5 | `sonda-local-win-engram-3.7.2.txt` |
| V1 guardado en rama | **Pasa** (hook y proxy): título `[feat-r-05-x]`, `Branch:`/`Author:`, clave `branch/<slug>/k` | `q3-hook-solo.txt` |
| V2/V3 switch y vuelta a main | **Pasa** (proxy): `[feat-r-05-a]`, `[feat-otra]`, y en `main` sin prefijo ni cambio de clave | `resultados/v2_v3-proxy.json` |
| V4 dos worktrees simultáneos | **Pasa** (proxy): 2 filas, `feat-wa` y `feat-wb`, misma clave `k`, sin colisión; cada proceso recibió el cwd de su worktree (Q14) | `resultados/v4-proxy.json` |
| V5 / Q6 subagente | **Pasa**: etiquetado desde el subagente con hook y con proxy; con proxy+sustitutos la sesión atribuida es la del padre | `v5_q6-hook/proxy/proxy+sim.json` |
| V7 sin marcador | **Pasa**: guardado sin tocar | `v7-proxy.json` |
| V8 hook ausente / sin confiar | Ausente: **pasa** (guardado sin etiqueta). Sin confiar: **no ejecutada** | `q11-*` brazo `sin_B` |
| V9 fallo del hook | Marcador inválido: **pasa** (guardado sin etiqueta). **Timeout externo de `PreToolUse`: pasa** (hook de 60 s con `timeout: 2` → corrida de 10,6 s, guardado continúa) | `v9_marcador_invalido-proxy.json`, `v9_timeout-60s.json` |
| V9 proxy terminado entre llamadas | **No ejecutada** en Claude (en Codex falló) | — |
| V10 / Q3 coexistencia | Con el **plugin real 0.1.22**: `hook.py` **6/6** y proxy B **6/6** (secuencial) etiquetan y atribuyen la sesión de Claude. Bajo carga (12 corridas en paralelo) `hook.py` perdió la etiqueta en 2 de 4 y el ajuste A en 1 de 3 (además descarta `type`); el proxy 3/3. Ver abajo | `resultados/q3-real*.json`, `q3_real.py` |
| V6 otras vías de escritura | **Cubiertas ahora** (cambio de R-05): `mem_session_summary`, `mem_session_finalize`, `mem_save_prompt` llevan `Branch:`/`Author:`; `mem_capture_passive` etiqueta cada ítem de la lista (Engram descarta el resto). Verificado en Claude real con hook y proxy. `mem_update` queda fuera (edita una fila ya etiquetada). `finalize` solo con prueba unitaria | `resultados/narrativas-*.json`, tests `test_hook`/`test_proxy_b` |
| V11 nombres complejos | HEAD detached y rama larga: **pasan** (`detached-8471285`, sufijo `-ca7b9b`). `feat/una` y `feat-una` con la misma clave: **colisionan** (una fila, revisión 2), igual que en Codex; la variante con sufijo de R-01 no se re-probó en Claude | `resultados/v11-hook.json` |
| V12 Windows | **Pasa**: `python`, `py`, `cmd /s /c`, Git Bash; ruta con espacios; marcador CRLF. `cmd /c` sin `/s` rompe por el quitado de comillas de cmd | `v12-windows-resultado.txt` |
| V13 lectura desde main | **Pasa**: `mem_search` en `main` devolvió el puerto 4381 y la rama de origen | `v13-proxy.json` |
| Q9 permisos | `git branch --show-current` sin pedido de aprobación (0 `permission_denials`) en `default`, `plan` y `acceptEdits`. Cierre completo: **no medido** | `q9-x.json` |
| Q11 sobrecosto | Ver abajo | `q11-par1..6.json` |

### Q3 en detalle (la prueba bloqueante)

Plugin real 0.1.22 (hooks `SessionStart` + `PreToolUse` oficiales) más la reescritura de rama (`hook.py`, ajuste A o proxy B), proyecto sintético, una llamada `mem_save` por corrida. Criterio de paso: título etiquetado **y** observación atribuida al `session_id` real de Claude (leído del stream y comparado con SQLite).

| Variante | Secuencial | En paralelo (12 corridas a la vez) |
|---|---|---|
| Solo plugin (control) | guarda y atribuye la sesión, sin etiqueta | — |
| `hook.py` | **6/6** | 2/4 (las otras guardaron **sin etiqueta**) |
| Ajuste A | — | 3/4 (y descarta `type`: queda `manual`) |
| Proxy B | **6/6** | 3/3 (una corrida no terminó de escribirse) |

Lectura honesta: con el plugin real **no hay conflicto de `session_id`** (el hallazgo con el sustituto era un artefacto). Pero bajo carga `hook.py` perdió la etiqueta en silencio.

**Diagnóstico bajo carga (instrumentado, `hook_log.py`, `resultados/carga/`):**

| Mecanismo | Secuencial | 12 corridas en paralelo |
|---|---|---|
| `hook.py` | 6/6 | **9/12** |
| Proxy B | 6/6 | **12/12** |

En las 3 corridas fallidas del hook, **el hook se ejecutó bien**: 0,62–0,68 s, código 0, produjo la reescritura. Aun así la fila quedó sin etiqueta. **No fue el timeout de Git** (hipótesis descartada). La reescritura del hook se descartó porque dos hooks `PreToolUse` del mismo evento corren en paralelo, ambos devuelven `updatedInput` (el del plugin ata el `session_id`) y la documentación no especifica cuál prevalece; con la carga cambian los tiempos y a veces gana el del plugin. No demostré la regla exacta (por ejemplo, «gana el último en terminar»), pero sí que es un orden no determinista y no un error del hook.

**Consecuencia:** un hook de Claude **no puede ser el mecanismo de entrega** de la etiqueta mientras convivan dos hooks que reescriben la misma entrada. El proxy B actúa en la capa MCP, después de todos los hooks, y resistió la carga. DA-05 = B queda confirmada con evidencia; el hook queda descartado como vía principal, no como «alternativa viable».

### Q11 (6 pares, 3 `mem_save` por sesión, mismo prompt)

- Entrada **total** (incl. caché): bimodal ≈132 k o ≈222 k tokens **en ambos brazos** con el mismo número de turnos; no depende de B. No se puede atribuir a B.
- Entrada **sin caché** (estable): con B suma +686, +2.411, +1.172, +565, +650, +865 tokens por sesión (mediana ≈ +770; ≈ +4 % de ≈17,9 k). Parte puede ser variación de redacción del modelo.
- Las 36 llamadas MCP resultaron correctas.

## Pendiente (no ejecutado)

V8 interactivo (hook sin confiar: `claude -p` confía la carpeta); V9 con el proceso del proxy terminado entre llamadas; `mem_session_finalize` en vivo (solo unitaria); cierre completo de Q9; variante con sufijo de V11 en Claude; modelo distinto de `haiku`; la regla exacta de prevalencia entre hooks paralelos (solo se demostró que es no determinista).

## Plugin local (corrección pedida por el dueño)

Causa: el stack instala el binario y el MCP de Engram (`internal/catalog/harnesses.yaml`, harness `engram`, tipo `external`) pero **no instala ni actualiza el plugin de Claude**; el 0.1.0 venía de una instalación manual de abril y el clon del marketplace estaba congelado en esa fecha. Acción en esta máquina, con backup previo en `~/.claude/backups-r05-20261119-224115`: `claude plugin marketplace update engram` y `claude plugin update engram@engram` → **0.1.0 → 0.1.23** (scope usuario; se aplica al reiniciar Claude; la evidencia de esta matriz usó 0.1.22). **Verificación tras reiniciar (2026-11-19):** el plugin 0.1.23 cargó y está habilitado, pero **el primer `mem_save` real fue bloqueado** por su `PreToolUse` («identidad del servidor local sin confirmar»). Causa: un `engram serve` que dejó vivo una de mis corridas anidadas (base temporal, solo el proyecto `r05-synthetic`) ocupaba el puerto 7437; el plugin falla cerrado si el servidor no es el de la base del usuario. Se detuvo ese proceso (verificado antes: solo datos sintéticos) y el guardado real pasó; el plugin levantó el servidor correcto. `arnes.py` ahora cierra los `engram serve` que aparezcan durante una corrida. Dos riesgos que **quedan para R-07/R-08**: (1) cualquier `engram serve` huérfano o de otra base bloquea las escrituras con ese error, sin pista del motivo; (2) en esta máquina falta `jq`, que usan los scripts del plugin (`session-start.sh`), y el stack no lo comprueba. Para repetir las pruebas hay que reclonar `Gentleman-Programming/engram` en la etiqueta `v3.7.2` (commit `c6ba9e3c`) y apuntar `PLUGIN` de `q3_real.py` y `q_narrativas.py` a `plugin/claude-code`.

Pendiente de diseño (no hecho aquí): que `active-stack install`/`verify` fijen y comprueben la versión del plugin; corresponde a R-07/R-08 y el aviso al grupo.
