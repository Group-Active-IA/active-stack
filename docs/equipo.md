# Modo equipo: cómo se funda, cómo se actualiza y qué no garantiza

> Describe lo construido y mergeado en `main`. Contrato técnico: [`team-mode.md`](../assets/skills/active-delivery/references/team-mode.md). Plan: [plan de trabajo en equipo](plan-trabajo-en-equipo.md). Memoria compartida: [engram-cloud-equipo.md](engram-cloud-equipo.md).

El modo equipo es **opcional**. Un proyecto sin `.active-stack/team.json` se comporta exactamente como un proyecto individual.

## Cómo se funda un equipo

En la fundación, `active-orchestrator` ofrece la fase opcional `team-mode`:

1. Pide la **cantidad** de integrantes (obligatoria, mínimo 2). Los nombres son opcionales y quien funda los define: el agente nunca inventa nombres, cuentas ni ids.
2. Pregunta si los nombres se versionan. Con «no», quedan en `.active-stack/team.local.json` (ignorado por git); `team.json` lleva solo los ids.
3. Escribe `.active-stack/team.json` (versionado), las reglas de ignore dentro de marcadores idempotentes en `.gitignore`, una plantilla de PR con los seis campos del traspaso y un ejemplo de `CODEOWNERS`, sin pisar lo existente.
4. Ofrece **una vez** proteger `main`, solo con confirmación. Si el plan de GitHub no lo permite, lo dice y deja la convención escrita (DA-04).
5. `agent-instruction` agrega la sección «Reglas de equipo» (marcadores `active-stack:team-mode`) solo con equipo confirmado.

Opcionalmente, la fase `memory-policy` (`cloud-branch`, `local` o `deferred`) invoca `active-stack memory init` para el modo de memoria compartida y solo **propone** el commit.

Después, cada persona abre el agente en su clon y se identifica en su primera sesión (queda en `.active-stack/me.local.json`, local). `active-delivery` le muestra qué fichas le tocan y sus dependencias, protege la creación de worktrees y avisa de cruces entre personas; el traspaso de seis campos sirve en ODD y SDD. El seguimiento puede ser local o con ClickUp (`active-roadmap`, cada escritura con aprobación de la selección exacta). El validador `team.py` y el verificador `fichas.py` existen solo para Claude y Codex.

## Cómo se actualiza («actualizar equipo»)

Se pide al agente «actualizar equipo» (entrypoint `:team` de `active-orchestrator`): permite sumar un integrante, completar un nombre pendiente o cambiar la cantidad, sin repetir las demás fases y de forma idempotente. Valida el borrador contra `team.py` antes de escribir y no admite credenciales en ningún campo. En OpenCode y Antigravity la fase valida con las reglas escritas en su texto y avisa que no hay validador automático.

## Qué NO garantiza el stack

- **Los permisos reducen riesgo; no son protección.** En Claude, el harness `permissions` agrega `deny` de `git push` a `main` (en los tres niveles, incluidas las variantes con `git -C`/`-c` y sus equivalentes de PowerShell). Es una barrera del agente: una persona, otra herramienta o una variante no contemplada pueden saltarla. En Codex no se agrega `deny` `(TBD)`; OpenCode y Antigravity quedan fuera.
- **La protección real de `main` es la de GitHub (DA-04):** rama protegida, revisión obligatoria y `CODEOWNERS`. La fundación la ofrece con confirmación; si la cuenta es GitHub Free o el plan no lo permite, solo queda la convención escrita.
- No hay cola automática de integración ni se resuelven conflictos entre personas: los avisos son informativos.
- La coordinación sigue necesitando acuerdos del equipo (quién aprueba, cuándo se mergea).
- En la memoria compartida, el etiquetado por rama solo vale en Claude y Codex, y bajo las condiciones de [engram-cloud-equipo.md](engram-cloud-equipo.md).
- ClickUp necesita conexión, autorización y acceso a la lista; sin eso se trabaja en modo local.
- El piloto de extremo a extremo (R-24) y la matriz completa de memoria (R-23) **no se han ejecutado** `(TBD)`.
