# Matriz de reglas OPSX por capa (G4, común a 03 y 04)

Publicada por la 03 (`claude-native-opsx`) el 2026-11-12. La 04 (`codex-native-opsx`) se alinea a
esta tabla y completa la columna Codex donde dice "lo define 04". Cualquier cambio de capa se hace
acá y en los dos changes a la vez; nadie redefine una fila en su propio design.

Capas:

- **Universal**: siempre cargada, en cualquier sesión (Claude: bloque `active-stack:universal-rules`
  en `CLAUDE.md`; Codex: marker en `AGENTS.md`).
- **Coordinador**: sólo con el coordinador invocado (`/sdd-orchestrator` en Claude,
  `$sdd-orchestrator` en Codex). Cuerpo de la skill.
- **Módulo `<x>`**: archivo de apoyo de la skill, se lee sólo para la acción que lo pide y, si se
  delega, se pega en el encargo del especialista.
- **Especialista**: si el subagente lo recibe y por qué vía.

Rangos de línea sobre los assets de `internal/harness/config/assets/` en `main` `35a89f8`.

| # | Origen | Regla o bloque | Capa | Especialista | Claude (03) | Codex (04) |
|---|---|---|---|---|---|---|
| 1 | contenido del usuario fuera de markers | Reglas propias del usuario | universal | sí, salvo Explore y Plan (por encargo) | `CLAUDE.md` del usuario | `AGENTS.md` por árbol |
| 2 | config del cliente | Permisos y sandbox | universal por mecanismo | heredado | `settings.json` | `config.toml` (01) |
| 3 | nuevo | Arranque breve: existe el coordinador y cómo invocarlo | universal | no | bloque universal con `/sdd-orchestrator` | marker `sdd-orchestrator` |
| 4 | `claude/sdd-orchestrator.md:3` | "Bind this to the dedicated agent only" | se elimina (lo cumple la entrega) | no | no se entrega | no se entrega |
| 5 | `claude/sdd-orchestrator.md:5-9` (sin la última oración de `:7`) | Role: coordinador, no ejecutor | coordinador | no | cuerpo de `SKILL.md` | skill `sdd-orchestrator` |
| 6 | `claude/sdd-orchestrator.md:13` y `:7` (última oración) | CLI `openspec` dueño del estado | universal | sí | bloque universal; encargo a Explore/Plan | bloque universal; encargo con reglas aplicables |
| 7 | `claude/sdd-orchestrator.md:14` | Delegar, no inflar el contexto | coordinador | no | `SKILL.md` | skill |
| 8 | `claude/sdd-orchestrator.md:15` | Engram persiste contexto (una línea) | universal | sí | bloque universal (toggle `engram`) | bloque universal (toggle engram) |
| 9 | `claude/sdd-orchestrator.md:17-26` | Starting work, fundación | coordinador | no | `SKILL.md` | skill |
| 10 | `claude/sdd-orchestrator.md:27-47` | Delegation rules (toggle) | módulo `delegation` | sí, en el encargo | `modules/delegation.md` | módulo `delegation` |
| 11 | 07 D1 | Cuándo delegar (regla breve) | coordinador | no | pendiente de texto de 07 | pendiente de texto de 07 |
| 12 | `claude/sdd-orchestrator.md:48-108` | Workflow, comandos, manejo de pedidos y tabla acción → especialista | coordinador | no | `SKILL.md` | skill |
| 13 | `claude/sdd-orchestrator.md:109-166` | Patrón de lanzamiento y plantilla de encargo | coordinador | no (es lo que arma el encargo) | `SKILL.md` | skill |
| 14 | `claude/sdd-orchestrator.md:167-202` | Ciclo de artefactos y comandos CLI | módulo `artifacts` (propose, apply, archive) | sí, en el encargo | `modules/artifacts.md` | modules/artifacts.md |
| 15 | `claude/sdd-orchestrator.md:203-206` | NEVER adivinar estado; NEVER crear `openspec/` a mano | universal | sí | bloque universal | bloque universal |
| 16 | `claude/sdd-orchestrator.md:207-210` | NEVER bloquear, NEVER apply/propose inline, ambigüedad, `/sdd-*` | coordinador | no (lo contradiría) | `SKILL.md` | skill |
| 17 | `claude/sdd-orchestrator.md:212-226` | Model assignments (toggle) | módulo `model-assignments` | no | `modules/model-assignments.md` | se omite hasta 09 |
| 18 | `governance.md:1-13` | Niveles de autonomía por dominio | universal (toggle `governance`) | sí | bloque universal | bloque universal (toggle governance) |
| 19 | `governance.md:14-fin` | Cómo aplicarlos a delegación y checkpoints | coordinador | no | `SKILL.md` (toggle `governance`) | skill |
| 20 | `engram-protocol.md:1-5` | Encabezado "ALWAYS ACTIVE" | universal | sí | bloque universal (toggle `engram`) | bloque universal (toggle engram) |
| 21 | `engram-protocol.md:6-12` | Wiring OPSX | módulo `engram-opsx` | no | `modules/engram-opsx.md` | módulo |
| 22 | `engram-protocol.md:13-59` | Guardado proactivo y búsqueda | universal | sí | bloque universal (toggle `engram`) | bloque universal (toggle engram) |
| 23 | `engram-protocol.md:60-92` | Cierre de sesión y recuperación tras compactación | coordinador | no | `SKILL.md` (toggle `engram`) | skill |
| 24 | `engram-protocol.md:93-fin` | Git sync de la memoria | módulo `engram-opsx` | no | `modules/engram-opsx.md` | módulo |
| 25 | `compose.go:93` | Línea `Strict TDD Mode: enabled` | módulo `tdd` | sí, en apply | `modules/tdd.md` | módulo `tdd` |
| 26 | `strict-tdd.md` | Módulo TDD | módulo `tdd` (apply) | sí, en apply | `modules/tdd.md` | módulo `tdd` |
| 27 | 05 D17 | Lectura dirigida, resumen acotado, herramientas, reporte conciso | módulos nuevos | según 05 D2 | pendiente de texto de 05 | pendiente de texto de 05 |
| 28 | 06 D5 | Contrato de evidencia, extracto `## Evidence Contract` | módulo `delegation` y encargo | sí, si ejecuta comandos | pendiente de texto de 06 | pendiente de texto de 06 |
| 29 | 07 D6 | Sobre de retorno | módulo `delegation` | sí, cuando se delega | pendiente de texto de 07 | pendiente de texto de 07 |
| 30 | `design-mockups.md` | Gates visuales | sección propia, sin cambios (G7 de 03) | sí | sección en `CLAUDE.md` | sección en `AGENTS.md` |

## Decisiones tomadas al publicar

- **Governance (filas 18 y 19).** Los niveles van en universal porque `governance.md:3` dice que
  aplican a todos los agentes; coincide con la posición de 05 (núcleo). Cómo aplicarlos a la
  delegación queda en el coordinador.
- **Engram (filas 8, 20 a 24).** Guardar y buscar son universales; lo específico de OPSX y el git
  sync son módulo; el cierre de sesión y la recuperación tras compactar son del coordinador.
- **Ciclo de artefactos (fila 14).** Módulo propio que reciben propose, apply y archive.
- **Model assignments (fila 17).** Módulo con toggle en Claude; en Codex se omite hasta 09.

## Reglas de la tabla

- **Estructura no es regla.** El título del asset (`:1`) se elimina en las capas nuevas, como la
  fila 4. Los encabezados de agrupación (`## Core Principles` `:11`, `## Rules` `:203`) se repiten
  en cada capa que tenga contenido debajo, y la numeración se rehace dentro de cada capa. Los
  markers `active-stack:` son estructura de composición y no viajan a los módulos.
- **Tabla acción → módulos del `SKILL.md`:** propose → `artifacts`; apply → `artifacts`, `tdd`;
  archive → `artifacts`; delegar → `delegation`, `model-assignments`; inicio y cierre de sesión
  → `engram-opsx`. Un módulo de toggle inactivo no figura en la tabla.

- Una fila, una capa. Si un bloque toca dos capas se parte en filas.
- Ninguna regla del asset queda sin fila; ninguna aparece en dos capas.
- Un módulo existe sólo con su toggle activo.

## Aporte del 04 · 13 de noviembre de 2026

La columna Codex se completó desde el compositor y las pruebas del change 04. Las
filas y la columna Claude se conservan como referencia de la matriz publicada
en `ed37113`. Esta actualización no incorpora la implementación del 03.

La fila 28 mantiene pendiente el texto del 06. En Codex, el contrato debe cargarse
cuando la acción exige evidencia, incluso en ejecución directa. La captura y el
esquema corregido se integrarán desde 05/06. La ubicación final en un módulo se
revisa con sus responsables.

El enum compartido `ConfigDeliveryManagedSkill` conserva el valor 2. Al combinar
las entregas, instalación, verificación y retiro deben seleccionar el recorrido
del cliente. La coincidencia del enum no vuelve intercambiables el instalador de
Claude y el de Codex. La prueba conjunta de ambas ramas queda pendiente de esa
reconciliación, con independencia del orden en que se integren sus PR.

## Integración de main tras PR #8 · 13 de noviembre de 2026

El PR #8 se integró en main `7a600895` desde la entrega Claude `01f944c`.
La rama de 04 incorpora esa base mediante merge y concilia el enum único, la
interfaz opcional de rutas y el despacho por agente en install, verify y uninstall.
Claude mantiene sus opciones de versión, capas, Project y shadowing; Codex Machine
mantiene su inventario, snapshot persistente y guardas de escritura. Codex Project
conserva la entrega legada y su aviso. Los markers son propios de cada cliente:
`universal-rules` para Claude y `sdd-orchestrator` para Codex.

Las pruebas aisladas de instalación en ambos órdenes, reaplicación, actualización,
verificación y retiro se registran en el informe de integración de 04. No son
aceptación de CLI/IDE real ni sustituyen la revisión humana de 03/04. La fila 28 y
la evidencia directa de 06 siguen pendientes de su contrato corregido. Este
apartado actualiza el estado técnico; conserva el aporte y pendientes históricos
anteriores como tales.
