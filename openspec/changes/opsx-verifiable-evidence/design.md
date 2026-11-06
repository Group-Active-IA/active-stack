# Design

## Context

El stack instala el orquestador como texto: se compone desde `internal/harness/config/assets/<agente>/sdd-orchestrator.md` con los toggles `tdd`, `engram`, `model-routing`, `delegation` y `governance`, y se inyecta en `CLAUDE.md` o `AGENTS.md` con markers idempotentes. El instalador no corre mientras el agente trabaja. Por eso el contrato de evidencia es, en esta etapa, **instrucciones más comprobaciones que ejecuta el coordinador con comandos de shell**. No hay forma de impedir una violación sin controles nativos, que son trabajo de la propuesta 08.

El roadmap apunta a un orquestador de invocación explícita, con módulos independientes (TDD, delegación y coordinación) que se cargan según el rol y la acción, y con el bloque global de SDD retirado de la sesión normal. Esta propuesta escribe el contrato para ese modelo. Los puntos que dependen de cómo lo definan 03 y 04 están marcados (TBD).

`openspec/` se versiona desde `1b4639b`. Los recibos y cualquier salida guardada en el change viajan a Git si se los pone ahí.

## Goals / Non-Goals

**Goals**

- Que un "pasó" de un sub-agente se pueda atar a una ejecución real, de esta revisión, con un comando que cubra el requisito.
- Que la duda se resuelva siempre hacia *sin verificar*.
- Mismo contrato y misma aceptación en Claude y en Codex.
- No exponer datos sensibles al versionar evidencia.

**Non-Goals**

- Impedir que un worker mienta (eso es 08 y, más adelante, un juez externo).
- Firmar recibos, ledger, subcomandos del binario o un runtime nuevo.
- Reutilizar `internal/verify`.
- Acotar o resumir salidas (05) o decidir cuándo delegar (07).

## Decisions

### D1 — No se reutiliza `internal/verify`

`internal/verify` es el reporte de health checks posteriores a la instalación. `Report` agrega `CheckResult` con estados `passed`, `failed`, `skipped` y `warning`, y `Ready` es verdadero cuando `Failed == 0`, de modo que los `skipped`, los `warning` y la ausencia total de checks no bloquean. La semántica que pide esta propuesta es la opuesta (lo incompleto no se da por bueno). Mezclar ambas ocultaría la diferencia, y el caso observado el 02/11 (0 checks, ✓) muestra el riesgo. El contrato se define aparte.

*Alternativa descartada:* extender `Report` con recibos. Mezcla salud de instalación con evidencia de tareas y comparte un estado que se comporta al revés.

### D2 — Un recibo por ejecución

Cada recibo describe **una** ejecución de **un** comando. Si un requisito se cubre con varios comandos, hay varios recibos que lo referencian.

| Campo | Contenido | Cómo se obtiene |
|---|---|---|
| `revision` | Commit de partida y si el árbol tenía cambios sin confirmar | `git rev-parse HEAD` y `git status --porcelain` |
| `directorio` | Raíz del repo o del worktree donde corrió | `git rev-parse --show-toplevel` |
| `comando` | Comando completo, tal como se ejecutó | lo declara el sub-agente |
| `resultado` | Código de salida y marca de tiempo | lo declara el sub-agente |
| `requisitos` | IDs de requisitos o escenarios del change que dice cubrir | lo declara el sub-agente |
| `salida` | Ruta de la salida completa, su huella y si fue truncada | `git hash-object --no-filters <archivo>` |

La huella usa `--no-filters`. Sin esa opción, Git puede aplicar conversiones de fin de línea (`core.autocrlf`, `.gitattributes`) y el mismo archivo daría huellas distintas en Windows y en Linux. Falta confirmar el comportamiento en cada plataforma soportada (Windows, macOS, Linux, WSL y Termux) (TBD, tarea 1.3).

### D3 — Dos controles con resultados separados

- **Procedencia (mecánico).** La revisión y el directorio del recibo coinciden con los del change, la salida existe, su huella coincide y el código de salida declarado es coherente con esa salida.
- **Suficiencia (criterio).** El comando ejercita lo que el requisito pide y el resultado respalda el cumplimiento. Cada requisito del change debe nombrar qué evidencia lo satisface; lo que no se pueda evaluar queda *sin verificar*.

Un recibo puede pasar procedencia y fallar suficiencia (tests verdes con un requisito incumplido). Los dos resultados se informan por separado. Quién evalúa la suficiencia es una pregunta abierta para el grupo.

### D4 — Tres estados, con la duda hacia "sin verificar"

- **Verificado:** procedencia y suficiencia aprobadas para cada requisito declarado, y existe al menos un recibo para él.
- **Rechazado:** recibo inventado, viejo, con salida adulterada o ejecutado en otro directorio, o suficiencia fallida.
- **Sin verificar:** falta un campo, falta la salida original, no se pudo evaluar, o no hay ningún recibo para el requisito. Es el estado por defecto y, a diferencia de un `skipped` en los health checks, **no cuenta como aprobado**. Un conjunto vacío de recibos nunca produce *verificado*.

### D5 — Dónde vive el contrato: módulo de delegación y coordinación

El contrato es texto. Se escribe como parte del **módulo de delegación y coordinación** del orquestador de cada agente, y un extracto (`## Evidence Contract`) se inyecta en el prompt de cada delegación que ejecute comandos, igual que hoy se inyecta `## Project Standards (auto-resolved)` antes de delegar `apply`. Así se resuelve que el asset de Codex no aplique a los ejecutores.

Si el módulo requerido no está disponible en una sesión, el coordinador informa la limitación y **detiene** la acción que lo necesita (regla transversal del roadmap), en lugar de seguir sin contrato.

La forma exacta del módulo depende del inventario de módulos que acuerden 03 y 04 (TBD). *Alternativa:* un harness `config` propio con sección `active-stack:verifiable-evidence` registrada en `sectionFragments`. Se descarta como primera opción porque el contrato no es opcional para quien delega, pero queda abierta.

### D6 — Dónde se guardan los recibos y las salidas (versionado y datos sensibles)

Los recibos viven en `openspec/changes/<change>/evidence/`. Como `openspec/` se versiona, esa carpeta puede terminar en el repositorio del equipo. Una salida completa de un build o de un test puede contener rutas personales, variables de entorno o secretos.

Propuesta de partida (TBD, la decide el grupo):

- **Se versionan** los recibos (texto pequeño, con la huella y la ruta lógica de la salida).
- **No se versionan** las salidas completas por defecto: viven en un subdirectorio local de la evidencia, fuera de Git, y el recibo guarda su huella para que un revisor que reciba la salida pueda comprobarla.
- **Antes de compartir** una salida (adjuntarla, subirla, pegarla), el coordinador la redacta: valores de variables de entorno, tokens, credenciales y rutas del home del usuario. Una salida redactada cambia de huella; el recibo registra la huella original y la de la versión redactada.
- Retención: el límite lo fija el grupo (por ejemplo, hasta archivar el change) (TBD).

Falta confirmar que el CLI de OpenSpec tolera una carpeta adicional dentro del change (TBD, tarea 1.4).

### D7 — Paridad entre agentes

Los cambios en los assets de Claude y de Codex se prueban con golden files de composición. Hoy existen `TestCompose_AllToggles` (Claude) y `TestCompose_AllToggles_opencode`; para Codex solo hay una comprobación de presencia de texto. Hay que agregar el golden de Codex siguiendo el patrón de los changes de paridad de OpenCode. La batería de recibos de prueba se aplica a ambos agentes y los veredictos tienen que coincidir.

### D8 — Verify en el flujo de Codex

El flujo del asset de Codex no lista `verify`, aunque README y ARCHITECTURE describen el ciclo completo. El asset de Claude ya lo incluye. Hay que acordar con 04 (Giuliano) si se corrige en 04 o dentro de esta propuesta; sin ese paso no hay dónde verificar recibos en Codex (TBD).

## Risks / Trade-offs

- **Recibos solo formales.** Un worker puede completar los campos sin ejecutar nada. Mitigación: huella de la salida y comparación con la revisión actual. No lo elimina; ver 08 y la propuesta de juez externo.
- **Cumplimiento solo por instrucción.** El contrato depende de que el agente lo siga. Mitigación: pruebas de aceptación; controles reales en 08.
- **Suficiencia subjetiva.** Mitigación: cada requisito nombra su evidencia; lo no evaluable queda *sin verificar*.
- **Filtración de datos al versionar.** Mitigación: D6 (no se versionan las salidas completas, redacción antes de compartir).
- **Divergencia entre assets.** OpenCode ya pasó por una auditoría de paridad por una divergencia silenciosa. Mitigación: golden también para Codex.
- **Salidas grandes.** Guardar la salida completa de cada ejecución ocupa espacio. Se coordina con 05.
- **Re-ejecutar siempre el comando** da la garantía más fuerte pero duplica costo y puede fallar en otro entorno. Queda como verificación opcional en cambios de riesgo alto (TBD).

## Migration Plan

No hay migración de datos: el contrato es texto nuevo en los assets. Quien reinstala el orquestador recibe la sección reescrita de forma idempotente, y desinstalar sigue siendo limpio. Los changes que ya estén en curso no tienen recibos; se tratan como *sin verificar* hasta que los generen.

## Open Questions

Para acordar en el grupo antes de enviar a revisión:

1. Dónde se escribe el contrato: módulo de delegación (propuesto) o harness propio. Afecta a 03, 04 y 07.
2. Quién evalúa la suficiencia: el coordinador, una persona o un revisor definido por el change.
3. Criterio de recibo viejo: solo la revisión, o también el estado del árbol de trabajo.
4. Verify en Codex: se corrige en 04 o aquí.
5. Agentes incluidos: confirmar que OpenCode y Antigravity quedan fuera.
6. Texto exacto del extracto en la delegación, con 03 y 04.
7. Política de salidas: qué se versiona, cómo se redacta, cuánto se retiene y quién accede (D6). Se coordina con 05.
8. Por qué el reporte de `verify-hook` quedó con cero checks en el caso del 02/11 (no investigado): informa a 01, 04 y 08.
