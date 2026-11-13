# Informe original: propose, intento 1, planner

Autor: subagente nativo `role-planner` (Claude Code, despachado por el orquestador el 2026-11-06). Revisión: 97d7f6c, rama `claude-native-opsx`.

Plan completo escrito; `openspec validate claude-native-opsx` pasa (también `--strict`). Dos cosas antes de enviarlo al equipo:

- **Compartir por git no funciona tal cual:** `openspec/` está en `.gitignore:36`; los artefactos no aparecen en `git status`. Para que el equipo de siete los vea hace falta `git add -f` o cambiar el ignore; no hice ninguna de las dos.
- **Una premisa del brief no se sostiene:** las skills `openspec-*` embebidas son marcadores vacíos, así que los módulos por acción todavía no existen como skills (ver C2).

## Rutas escritas
- openspec/changes/claude-native-opsx/proposal.md
- openspec/changes/claude-native-opsx/design.md
- openspec/changes/claude-native-opsx/specs/claude-instruction-layers/spec.md
- openspec/changes/claude-native-opsx/specs/claude-coordinator-agent/spec.md
- openspec/changes/claude-native-opsx/specs/claude-orchestrator-migration/spec.md
- openspec/changes/claude-native-opsx/tasks.md

## EVIDENCIA (rama claude-native-opsx, revisión 97d7f6c, 2026-11-06; sin cambios de código)
- `openspec status --change claude-native-opsx --json`: `isPlanningComplete: true`, los cuatro artefactos `done`.
- `openspec validate claude-native-opsx` → "Change 'claude-native-opsx' is valid" (exit 0); `--strict` exit 0.
- `openspec list --specs` → "No specs found.": las tres capacidades son nuevas, no hay modificadas.
- Cero em dashes en los artefactos.
- Clasificación ALTO con justificación en proposal.md (no CRITICAL porque el CLAUDE.md del repo §4 fija ALTO como techo; si el grupo considera que settings.json puede guardar credenciales, se revisa). `## Aprobación` quedó vacía para la persona.
- Contradicciones del código (en design.md, "Contradicciones"):
  - C1: `sdd-orchestrator.md:3` pide atarse a un agente dedicado, pero en Claude se inyecta en el archivo compartido (`adapter.go:125-127`).
  - C2: el coordinador delega en `openspec-apply-change`/`openspec-archive-change` (`sdd-orchestrator.md:100-105`); las skills embebidas se llaman `openspec-apply`/`openspec-archive`, tienen 4 líneas y empiezan "PLACEHOLDER: port from" (`assets/skills/openspec-apply/SKILL.md:3`); no figuran en `harnesses.yaml`; `CHANGES.md:375-390` documenta que se borran tras `openspec init`.
  - C3: `InjectPrimaryAgent` no sirve tal cual para Claude: escribiría `{"agent":{"<id>":{...}}}` en `~/.claude/settings.json`, y según `claude --help`, `--agent` "Overrides the 'agent' setting", o sea esa clave selecciona el agente de la sesión. Por eso se planifica un modo de entrega nuevo basado en archivo de agente.
  - C4: `WriteFileAtomic` fuerza `0o644` (`writer.go:63`; llamadores `inject.go:125`, `primary_agent.go:83,:118`) y hace `os.Chmod(dir, 0o755)` (`writer.go:40`); puede abrir un `~/.claude` que estaba en `0700`.
  - C5: `PurgeStaleSections` (`inject.go:66-73`) borra toda sección `active-stack:` no registrada; un bloque universal sin registrar se borraría solo en la próxima instalación.
  - C6: el CLAUDE.md del repo dice que el orquestador "ya vive en tu `~/.claude/CLAUDE.md`"; con este cambio deja de ser cierto.
- Colisión de nombres: el ID sugerido `active-orchestrator` ya es una skill del catálogo (`harnesses.yaml:92`). En esta máquina hay agentes propios `~/.claude/agents/role-*.md` que el repo no administra; el plan exige no pisarlos.
- Lecciones L-NNN: no hay base de lecciones ni `kb/` en el repo, ni `obsidian.md`. Declarado en tasks.md.

## JUICIO (planner; confianza media-alta)
Plan:
- Tres capas: bloque universal mínimo en CLAUDE.md registrado como sección propia; coordinador como archivo de agente nativo (`<base>/.claude/agents/<id>.md`) invocado sólo con `claude --agent <id>` (el instalador nunca escribe el setting `agent`); módulos por acción que por ahora el coordinador pega en el brief del especialista, porque no dependen de las skills de C2.
- Migración: misma transacción que `InjectPrimaryAgent` (chequeo de versión y colisión, snapshot único, escritura, restauración si falla); desinstalar con restore devuelve CLAUDE.md byte a byte.
- Matriz regla por regla en borrador en design.md, para acordar con la Propuesta 04.

Alternativas rechazadas: todo en CLAUDE.md por secciones (el experimento muestra que todas llegan a todos); setting `agent` por defecto; reutilizar `ConfigDeliveryPrimaryAgent` (C3); alias `--agents` JSON; output style; plugin; hooks SessionStart.

Decisiones pendientes de grupo (design.md, con recomendación y qué artefacto cambia si se elige otra):
| ID | Decisión | Recomendación |
|---|---|---|
| G1 | dónde se instala cada pieza | según el target, como hoy |
| G2 | cómo se invoca el coordinador | `--agent` explícito; skill como alternativa |
| G3 | cómo llegan los módulos al especialista | brief del coordinador por ahora; skills o agentes con la 04 |
| G4 | reglas de especialistas (matriz) | según borrador de design.md |
| G5 | ID del coordinador | `sdd-orchestrator` |
| G6 | alcance del arreglo de permisos | arreglo global del writer |
| G7 | design-mockups | queda en CLAUDE.md en este change |
| G8 | dependencia con la 02 | avanzar; comparación antes/después como tarea |

Decidido sin el grupo: las specs se escribieron sobre las recomendaciones. Si el grupo elige otra opción, vuelve a planner y plan-reviewer antes de construir (tarea 2.4).

## SUPOSICIÓN (sin verificar; cada punto tiene tarea en la sección 1 de tasks.md)
- Si subagentes definidos como archivo y los integrados (Explore, Plan, general-purpose) también reciben CLAUDE.md.
- Qué frontmatter acepta un archivo de agente, y si el prompt de `--agent` reemplaza o se suma al system prompt.
- Versión mínima de Claude Code: sólo se probó 2.1.318; D7 queda `(TBD)`.
- En target Project: si Claude lee `.claude/agents` y `.claude/CLAUDE.md`, y cuál gana con dos agentes del mismo nombre.
- Qué tipo de valor acepta el setting `agent`.
- Si un coordinador lanzado con `--agent` puede lanzar subagentes.
- Qué contexto recibe una skill.
- Que este hilo corra como subagente `role-planner` y vea el CLAUDE.md global es sólo un indicio; identidad no verificada.

## Fuentes leídas
CLAUDE.md del repo; internal/agents/claude/adapter.go; internal/harness/config/{compose.go:1-80, primary_agent.go, inject.go, install.go:1-110} y encabezados de los cinco assets; internal/filemerge/writer.go:16-90; internal/verify/harness_checks.go:85-130,236-280; internal/uninstall/steps.go:150-185; internal/install/plan.go (búsqueda); cmd/active-stack/uninstall_dispatch.go:12-30; harnesses.yaml:14-30,60-125; encabezados de assets/skills/*/SKILL.md; CHANGES.md:360-395; búsquedas en ARCHITECTURE.md y README.md; `claude --help`/`--version`; `openspec --version` (1.14.3); sólo nombres de archivo en ~/.claude/{agents,skills,commands}.

## Contaminación
No leí informes de otros roles. El experimento D0 y la evidencia de partida vienen del brief del orquestador; los cito como evidencia ajena y releí el código citado. Sin Engram.

## Pendientes
Aprobación humana fechada en `## Aprobación`, decisión de grupo G1-G8 y acuerdo de matriz con Giuliano (04) antes de construir. Sigue el plan-reviewer; sus hallazgos vuelven al planner.
