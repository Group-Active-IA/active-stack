# Inventario de delegación: tareas 1.1 y 1.2 (13/11)

Base: rama `propose/opsx-selective-delegation` en `026af4c` (incluye `main` `d5e95bd`).
Etiquetas: **EVIDENCIA** (comando, salida y hora en este informe), **DOC** (fuente citada, no
observada aquí), **SUPOSICIÓN** (no observado; dice qué falta para confirmarlo).

## 1.1 Obligaciones de delegar por ruta de composición

**EVIDENCIA.** Test temporal en `internal/harness/config` (borrado después de correrlo) que compone
cada ruta con todos los toggles y sin `delegation`, y busca los textos de obligación. Comando:
`go test ./internal/harness/config/ -run TestZZInventoryTmp -v`, 2026-11-13T13:33:56-03:00.
Columnas: línea en el texto compuesto con `delegation` / sin `delegation`; «—» = ausente.

| Ruta | Texto | Con | Sin |
|---|---|---|---|
| Claude por capas (`ComposeLayers`), SKILL.md | «Direct or delegated» (regla de 07) | :14 | :14 |
| Claude por capas, SKILL.md | «Delegation is disabled» | — | :19 |
| Claude por capas, todas las capas y módulos | cualquier obligación de delegar | — | — |
| Codex nativo (`ComposeCodexPackage`), SKILL.md | «Direct or delegated» | :99 | :99 |
| Codex nativo, SKILL.md | «Delegation is disabled» | — | :104 |
| Codex nativo, todos los archivos | cualquier obligación de delegar | — | — |
| Legacy `claude` | «delegate real work» | :7 | :7 |
| Legacy `claude` | «Delegate, don't inflate» | :14 | :14 |
| Legacy `claude` | tabla «Delegation Rules» / «Read 4+ files» | :27 / :34 | — |
| Legacy `claude` | «For each action, delegate» / «You delegate the skill» | :98 / :107 | :76 / :85 |
| Legacy `claude` | «ALWAYS delegate» | :208 | :186 |
| Legacy `opencode` | igual que legacy `claude` | :7, :14, :27, :34, :95, :104, :207 | :7, :14, :73, :82, :185 |
| Legacy `codex` | «Delegate, don't inflate» | :14 | :14 |
| Legacy `codex` | «Delegation Rules», «delegate real work», «Read 4+ files» | :18, :20, :25 | — |
| Legacy `codex` | «For each action, delegate» | :85 | :64 |
| Legacy `generic` | «Delegate, don't inflate», «For each action, delegate» | :14, :83 | :14, :64 |
| Legacy `generic` | tabla | :18, :23 | — |
| Legacy `antigravity` | sólo la tabla | :18, :25 | — |

Conclusión: las rutas de 07 (Claude por capas y Codex nativo) no conservan ninguna obligación,
con o sin el toggle. Todas las rutas legacy la conservan sin el toggle; ese es el límite declarado
en `design.md` D8.

### Quién recibe la ruta legacy (alcance real del límite)

**EVIDENCIA** (lectura de código en `026af4c`):

| Agente | Entrega | Ruta | Fuente |
|---|---|---|---|
| Claude, Claude Code ≥ 2.1.110 | skill gestionada | por capas (07) | `agents/claude/adapter.go:127`; `version.go:45` |
| Claude, versión menor o no detectable | skill gestionada con fallback | **legacy `claude`** | `harness/config/install.go:197-198` |
| Codex, máquina | skill gestionada | nativo (07) | `agents/codex/adapter.go:64`; `install.go:83-88` |
| Codex, instalación de proyecto | instrucciones | **legacy `codex`** | `install.go:78-79` |
| OpenCode | agente primario | **legacy `opencode`** | `agents/opencode/adapter.go:112`; `install.go:133` |
| Antigravity | instrucciones | **legacy `antigravity`** | `agents/antigravity/adapter.go:80` |

El límite de D8 abarca más de lo descrito al decidirlo («agentes sin capas y `primary-agent`»):
también Claude con versión vieja o no detectable y Codex en instalación de proyecto.

## 1.2 Cómo se lanza, espera y cierra un worker, y qué consumo expone

| Cliente | Versión | Lanzar | Esperar | Cerrar | Consumo por worker | Estado |
|---|---|---|---|---|---|---|
| Claude Code, terminal | 2.1.323 (EVIDENCIA: `claude --version`, esta máquina, 13/11) | herramienta Agent; un subagente puede lanzar otros hasta tres niveles, configurable (DOC: 03 `design.md` archivado `:55`) | primer plano bloquea hasta el retorno; segundo plano notifica al terminar | termina al devolver su mensaje final; se puede retomar por mensaje | SUPOSICIÓN: el resultado de Agent informa tokens totales, usos de herramienta y duración por subagente, pero no separa caché ni costo | SUPOSICIÓN en lanzar/esperar/cerrar/consumo: falta una corrida observada |
| Claude Code, `claude -p` | 2.1.323 | igual | igual | igual | SUPOSICIÓN: `--output-format json` informa uso y costo de la sesión entera, no por worker | falta corrida observada |
| Claude Code, IDE | no observado | — | — | — | — | SUPOSICIÓN total; 02 separa terminal e IDE |
| Codex CLI | no instalado aquí (EVIDENCIA: `codex: command not found`, 13/11). 04 observó `0.167.0-alpha.3.2` sólo en macOS (DOC: 04 `plan-aceptacion-clientes-04.md:16`) | DOC: parámetro de rol al hacer spawn, según doc rust-v0.165.0 citada por 04 (`propose-1-planner.md:71,84`), sin observación propia de 04 | SUPOSICIÓN | SUPOSICIÓN | SUPOSICIÓN | bloqueado: depende de la sonda P2 y las celdas C05/C06 de 04 (deuda aceptada de 04) |
| Codex IDE | no observado | — | — | — | — | SUPOSICIÓN total |

### Qué falta para cerrar 1.2

1. Claude terminal: una corrida observada con un subagente trivial y otra en segundo plano. Hay que
   registrar el formato real del consumo por worker y si separa caché. Cuesta una sesión corta de
   `claude -p`.
2. Codex: instalar el CLI en Windows o Linux y correr C05/C06 de 04, o tomar sus resultados cuando
   04 salde la deuda P2.
3. IDE: fuera del alcance mínimo hasta que 02 fije la matriz por interfaz.

Sin estos datos el piloto (sección 5) no puede contabilizar el consumo por worker. La medida que
falte se informa como no disponible, nunca estimada sin decirlo (02).
