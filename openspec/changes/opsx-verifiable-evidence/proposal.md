# Proposal: opsx-verifiable-evidence

> Change `opsx-verifiable-evidence` · ficha 06 del roadmap · Responsable: Gianfranco · Agentes: Claude y Codex · Prioridad P1 · Gobernanza MEDIO
> Base revisada: `1b4639b`. Depende de 03 (`claude-native-opsx`) y 04 (`codex-native-opsx`). Habilita 07 y 08.
> Esta ronda es solo `propose`: todas las tareas de implementación quedan pendientes.

## Why

El orquestador delega a sub-agentes el trabajo que ensucia el contexto, incluidas las pruebas y los builds. Lo que el sub-agente devuelve llega como texto libre. En los assets revisados (`internal/harness/config/assets/claude|codex/sdd-orchestrator.md`) no hay un contrato de retorno que permita comprobar que un "los tests pasan" corresponde a una ejecución real, sobre la revisión actual y con un comando que cubra lo pedido. Un resultado bien formado puede venir de otra revisión, de un comando insuficiente o de una ejecución inventada. Y un test verde tampoco prueba por sí solo que se cumpla un requisito.

Las propuestas 07 (delegar con contrato) y 08 (controles nativos) dependen de ese contrato. Sin él, delegar equivale a confiar en el texto del worker.

### Evidencia

- **El único "reporte" que existe hoy es el de salud de la instalación y no sirve como contrato.** En `internal/verify/report.go` (`main`, `1b4639b`), `BuildReport` marca `Ready = (Failed == 0)`. Los checks `skipped` y los `warning` no bloquean. Un reporte con cero checks queda `Ready` y emite "Installation verified. You're ready to start building.".
- **Caso observado el 02/11/2026** en una instalación de v0.4.0 (Windows, modo `lite`, agente `claude`): el paso `verify-hook` terminó con ✓ y el resumen dijo "0 passed, 0 failed". No se ejecutó ninguna comprobación y `~/.claude/settings.json` no tenía ningún hook. No se investigó por qué la lista de checks quedó vacía (queda como tarea en 1.5).
- **La ficha 06 del roadmap** fija los casos de rechazo: recibos inventados o viejos, comandos insuficientes, salidas truncadas sin original y requisitos incumplidos con tests verdes.
- **El asset de Codex** declara en su primera línea "Do NOT apply it to executor agents" y su flujo lista `explore, propose, apply, archive` sin `verify`. El asset de Claude sí incluye `openspec-verify`. Un contrato escrito solo ahí no llega a los ejecutores de Codex.
- **`openspec/` ya se versiona en Git** (commit `1b4639b`). Todo lo que se guarde dentro de un change viaja al repositorio del equipo, salidas originales incluidas.

## What Changes

- Se define el **recibo de ejecución**: un registro, uno por ejecución de un comando, con revisión, directorio, comando, resultado, requisitos cubiertos y referencia a la salida original con su huella.
- Se separan dos controles con resultados independientes: **procedencia** (la ejecución es real y corresponde a este contexto, control mecánico) y **suficiencia** (lo ejecutado alcanza para el requisito, control de criterio).
- Tres estados: **verificado**, **rechazado** y **sin verificar**. Un recibo incompleto, o una verificación con cero recibos, queda *sin verificar*, nunca verificado. Esto invierte a propósito la semántica de `skipped` de los health checks.
- El contrato se entrega como **módulo de delegación y coordinación** de cada orquestador (texto, no código) y se inyecta un extracto breve (`## Evidence Contract`) en el prompt de cada delegación que ejecute comandos. Claude y Codex reciben el mismo contenido y la misma aceptación.
- Se define **qué parte de la evidencia se versiona**: los recibos (texto pequeño) y la huella; las salidas completas se guardan localmente y se redactan antes de compartirse (decisión (TBD) del grupo, ver design D6).
- Se resuelve, junto con 04, que el flujo del asset de Codex incluya el paso de verificación de recibos.

## Capabilities

### New Capabilities

- `execution-receipts`: contrato de recibo de ejecución (esquema, estados, verificación de procedencia, revisión de suficiencia, salida original y su tratamiento, entrega al sub-agente y paridad entre Claude y Codex).

### Modified Capabilities

(ninguna: no existe un spec previo de delegación o de evidencia; el texto de los assets se modifica sin que haya un requisito archivado que lo describa)

## Alcance

- Esquema del recibo y reglas de validación.
- Reglas de procedencia y de suficiencia, y los tres estados.
- Casos de rechazo obligatorios (los cinco de la ficha, más el reporte de cero recibos).
- Dónde viven los recibos dentro del change y qué se versiona.
- Cambios previstos en los assets de Claude y de Codex y su prueba de paridad.
- Redacción de datos sensibles, retención y acceso a las salidas originales.

## Exclusiones

- Ledger firmado, firma de recibos, subcomandos nuevos de runtime en el binario o un juez integrado. El roadmap los deja para una decisión de arquitectura separada.
- Reutilizar o modificar `internal/verify` (ver design D1).
- Lectura dirigida y acotado de salidas (propuesta 05). Aquí solo se exige conservar la referencia al original.
- Cuándo delegar y cuándo ejecutar directo (07). Controles previos y posteriores en el cliente (08).
- OpenCode y Antigravity, salvo que el grupo acuerde incluirlos (ver design, preguntas abiertas).
- Implementación: `apply`, pruebas de implementación, `verify` y `archive` son de una etapa posterior.

## Impact

- **Código y assets previstos** (nada se toca en esta ronda):
  - `internal/harness/config/assets/claude/sdd-orchestrator.md` y `.../codex/sdd-orchestrator.md` (módulo de delegación y coordinación).
  - `internal/harness/config/compose_test.go` (golden de composición; hoy existe `TestCompose_AllToggles` para Claude y `_opencode`, y para Codex solo `TestCompose_VariantCodex`, que comprueba que aparezca la palabra "Codex").
  - Eventualmente `internal/catalog/harnesses.yaml` si el grupo elige un harness propio (BAJO).
- **No se tocan** `backup`, `filemerge` ni `pipeline` (ALTO). Si el diseño final obliga a tocarlos, la propuesta vuelve a revisión.
- **Gobernanza:** `harness/config` es MEDIO (implementar con checkpoints).
- **Otras propuestas:** modifica lo que reciben los roles de 03 y 04 y fija lo que 07 puede exigir y 08 puede validar.
- **Espejo obligatorio:** si se modifica la constitución del repo, `CLAUDE.md` y `AGENTS.md` se actualizan juntos.
