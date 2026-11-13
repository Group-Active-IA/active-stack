# Proposal: claude-native-opsx

> Propuesta 03 del roadmap del equipo Active Stack (etapa 02, prioridad P1, cliente Claude Code).
> Responsable: Juani. Depende de la Propuesta 02 (medición de referencia, integrada en `main`) y
> se coordina con la Propuesta 04 (Codex, Matías). Revisión 2026-11-11: reescrita según el informe
> unificado de revisión OPSX (acuerdos del 2026-11-10); el detalle está en
> `revision-2026-11-11.md`. Base `35a89f8` (`main` con 01 y 02 integradas).

## Why

En Claude Code el instalador inyecta el bloque completo del orquestador (rol de coordinador,
delegación, model-routing, governance, Engram y TDD) en `~/.claude/CLAUDE.md`
(`internal/agents/claude/adapter.go:28-30`, `:125-127`; `internal/harness/config/install.go:62-75`).
Ese archivo lo lee todo: la sesión normal, cualquier agente de sesión y cada subagente
(experimento del 2026-11-06, ver design.md D0). El propio asset dice "Bind this to the dedicated
`sdd-orchestrator` agent only. Do NOT apply it to executor agents"
(`internal/harness/config/assets/claude/sdd-orchestrator.md:3`), pero en Claude ese agente no
existe. Resultado: nadie puede decir qué contexto recibe cada especialista, y el especialista
recibe instrucciones de coordinador que lo contradicen ("NEVER do apply or propose work inline",
`sdd-orchestrator.md:208`). Además, toda sesión de Claude se comporta como coordinador OPSX aunque
el usuario no lo haya pedido.

## What Changes

- Separar el contenido que hoy viaja en un solo bloque en tres capas con dueño explícito:
  **reglas universales** (bloque mínimo con markers en `CLAUDE.md`), **definición del rol
  coordinador** (skill de Claude Code invocable sólo por el usuario) y **módulos por acción**
  (archivos de apoyo de esa skill que el coordinador carga según la acción OPSX). La matriz de
  reglas por cliente, interfaz, rol y acción se mantiene en conjunto con la Propuesta 04.
- **Activación explícita en la sesión actual (acuerdo del grupo, 2026-11-10).** El coordinador
  se entrega como la skill `sdd-orchestrator` con `disable-model-invocation: true`. El usuario
  la activa con `/sdd-orchestrator` y el coordinador sigue en el hilo actual. El modelo no la
  puede invocar por su cuenta y no se puede precargar en subagentes. Se descarta el agente de
  sesión con `claude --agent` que proponía la versión anterior.
- Agregar en el adapter de Claude un modo de entrega propio para el coordinador (directorio de
  skill con `SKILL.md` y módulos, con marca de propiedad), distinto de
  `ConfigDeliveryPrimaryAgent`, que escribe JSON de OpenCode.
- **BREAKING (comportamiento)**: una sesión normal de Claude Code deja de comportarse como
  coordinador OPSX. Quien quiera coordinador lo invoca con `/sdd-orchestrator`. Requiere nota de
  migración y aviso en install/update.
- Migrar instalaciones existentes: purgar sólo las secciones administradas del bloque legado en
  `CLAUDE.md`, con snapshot previo, preservando todo contenido del usuario fuera de markers;
  reinstalación y update idempotentes; colisiones con skills propias del usuario sin
  sobrescribir; rollback durante la operación con la guarda de 01 y reversión posterior que no
  pisa ediciones del usuario.
- Extender `verify` y `uninstall` para la skill del coordinador y el bloque universal.
- Cómo se comunica un control obligatorio ausente en Claude (acuerdo 07/08): el coordinador
  detiene sólo la acción que depende del módulo o control faltante, dice cuál falta y sigue con
  lo demás.
- Preservar permisos de archivos y directorios del usuario al reescribir (hoy
  `internal/filemerge/writer.go:40` y `:63` fuerzan `0o755`/`0o644`). Por tocar `filemerge`, el
  cambio se coordina con 01 y 08 en una sola revisión ALTO (design.md D6).
- Las decisiones que siguen abiertas están en design.md (tabla G). Cada requisito que depende de
  las decisiones abiertas se marcaban con "(depende de Gn, recomendación)"; el 2026-11-13 quedaron cerradas y se quitaron esas marcas.

## Capabilities

### New Capabilities
- `claude-instruction-layers`: qué reglas recibe cada consumidor de Claude Code (sesión normal,
  coordinador, especialista) y cómo se preserva el contenido propio del usuario en `CLAUDE.md`.
- `claude-coordinator-skill`: entrega del coordinador OPSX como skill de Claude Code invocable sólo
  por el usuario, su continuidad en la sesión (compactación y reanudación), la carga de módulos
  por acción y la compatibilidad por versión y target.
- `claude-orchestrator-migration`: ciclo de vida de la entrega para Claude (migración desde el
  bloque legado, reinstalación, update, colisiones, rollback, fallo de escritura, verify y
  uninstall).

### Modified Capabilities
- Ninguna. Las specs integradas en `main` (`opsx-benchmarking`, de 02, y las de 01) no cambian:
  esta ficha sólo las referencia.

## Impact

- Código (futuro, no en esta propuesta): `internal/agents/claude/adapter.go`,
  `internal/model/harness.go` (nuevo `ConfigDelivery`), `internal/harness/config/{install.go,
  compose.go,inject.go}` y un inyector nuevo de skill administrada,
  `internal/harness/config/assets/claude/` (partición del asset),
  `internal/harness/skill/idempotent.go` (reutilizar el reemplazo atómico con chequeo de
  propiedad), `internal/verify/harness_checks.go` y `projectVerifyAdapter` en
  `cmd/active-stack/main.go`, `internal/uninstall/{steps.go,plan.go}` (sólo Machine; Project queda
  como DEUDA, design.md D5), `internal/install/plan.go` (rutas por target y
  `projectTargetAdapter`), `cmd/active-stack/update.go` e `internal/update/state.go` (camino
  `update --sync` y proyectos registrados), `internal/backup` (reutiliza `GuardRollback` de 01 y
  agrega el registro persistente para la reversión posterior), posiblemente
  `internal/filemerge/writer.go` (permisos, ALTO, revisión compartida con 01 y 08).
- Consumidores fuera del área: OpenCode (`ConfigDeliveryPrimaryAgent` no debe cambiar), Codex y
  Antigravity (comparten `Compose` y los fragmentos; la Propuesta 04 define su lado de la matriz),
  goldens `internal/harness/config/testdata/compose_all_toggles_*.golden`, la sección
  `design-mockups` (registrada en `sectionFragments`, no debe purgarse), la skill de fundación
  `active-orchestrator` (otro ID), y el usuario final con contenido propio en `~/.claude/CLAUDE.md`
  y skills propias en `~/.claude/skills/`.
- Fichas que entregan contenido a la 03: 05 (texto de sus módulos, D17), 06 (contrato de
  evidencia, D5) y 07 (regla de decisión de delegar en la skill y sobre de retorno, D1 y D6).
  La 03 los ubica y no los redefine.
- Medición: el manifiesto, la rúbrica y el protocolo de 02 están en `main`
  (`openspec/changes/opsx-baseline/`); la comparación antes y después la corre 02 en su tarea 4.1
  con la variante que entregue esta ficha.
- Fuera de alcance: cambiar la entrega de OpenCode, Codex u otros agentes; reescribir los
  contenidos de governance/TDD/Engram; hooks, output styles, plugins y el setting `agent` (ver
  alternativas en design.md).

## Clasificación

**ALTO (HIGH).** Justificación por riesgo y alcance, no por tamaño:
- Escribe y purga `~/.claude/CLAUDE.md`, el archivo de instrucciones global del usuario, y crea
  un directorio en `~/.claude/skills/`, donde el usuario puede tener skills propias. Toca dominios
  que la governance del proyecto declara ALTO: `backup`, `filemerge`, `pipeline` (rollback)
  (`CLAUDE.md` del repo, §4).
- Cambia el comportamiento por defecto de toda sesión de Claude Code de quien instale (BREAKING).
- Tensión abierta entre ALTO y CRITICAL, que este plan no resuelve:
  - A favor de ALTO: el `CLAUDE.md` del proyecto (§4) declara que no hay dominios CRÍTICO y que el
    máximo es ALTO; el change no toca autenticación ni facturación.
  - A favor de CRITICAL: la regla global del flujo de trabajo clasifica como CRITICAL todo lo que
    cruce un límite de datos sensibles, aun con diffs chicos. El arreglo de permisos cambia cómo se
    escriben archivos que pueden guardar secretos (`~/.claude.json` guarda config MCP;
    `settings.json` podría admitir `env`, no verificado), y hoy el writer amplía permisos
    (`writer.go:40`, `:63`).
  - Si `settings.json` o `~/.claude.json` pueden contener credenciales es un hecho a verificar, no
    una decisión de grupo: lo cubre la tarea 1.13.
  - La persona que aprueba elige el nivel en `## Aprobación`. Este plan se escribió con la
    rigurosidad de ALTO; si se elige CRITICAL, cada aserción de seguridad o datos lleva mutante y la
    revisión del plan se repite.

Consecuencias del nivel: roles separados, aprobación humana explícita antes de construir, TDD con
regresiones de consumidores y un mutante sobre las aserciones clave, y verificación independiente.
Al cerrar, el PR hacia `main` lo revisa el responsable del change anterior según el circuito del
grupo (contrato, evidencia, decisión).

## Aprobación

<!-- Sólo la completa una persona humana. Ningún agente firma acá. -->
<!-- Formato sugerido: fecha, nombre, nivel elegido (ALTO o CRITICAL, ver Clasificación), opción elegida para cada decisión abierta (design.md, tabla G) y alcance aprobado. -->

- Fecha: 2026-11-13.
- Persona: Juani Sarmiento.
- Nivel elegido: ALTO (no CRITICAL).
- Decisiones: G1, G3, G7, G10, G11 y G13 con la opción recomendada en design.md; G12 queda Machine-only con limpieza manual y DEUDA-1; G6 con alcance chico (sólo las rutas de esta ficha, sin tocar `internal/filemerge` global ni el código de 01/08). Detalle en design.md, "Decisiones".
- Umbral de reenganche del `SKILL.md`: 4000 tokens.
- Alcance aprobado: tareas 3 a 7 y 9 de tasks.md. Siguen fuera de esta aprobación: la sección 1 y la 8 (las corre Juani con Claude real), la 10 y el push.
- Pendiente declarado: los textos de 05, 06 y 07 (tarea 2.3); `MinVersion` sigue en 0.0.0 hasta que Juani lo fije con pruebas reales.

## Knowledge Base Impact

El repo no tiene `obsidian.md` ni knowledge base `kb/` (verificado con listado de la raíz el
2026-11-06); la documentación canónica es `ARCHITECTURE.md`, `README.md`, `CLAUDE.md`/`AGENTS.md`
y `CHANGES.md`.

| Documento | Acción | Motivo |
|---|---|---|
| `ARCHITECTURE.md` (§ harness de configuración, líneas 56-75) | REWRITE | Hoy afirma que el bloque del orquestador "se inyecta en `CLAUDE.md` / `AGENTS.md`"; para Claude pasa a ser skill más bloque universal. |
| `README.md` (línea 214 y tabla de harnesses) | REWRITE + APPEND | Misma afirmación; agregar cómo invocar `/sdd-orchestrator`, la nota BREAKING y los límites (compactación, salida del modo coordinador). |
| `CLAUDE.md` y `AGENTS.md` del repo (§2, "se inyecta en `CLAUDE.md`/`AGENTS.md`"; `CLAUDE.md:9` "ya vive en tu `~/.claude/CLAUDE.md`"; `AGENTS.md:148-149` "El orquestador SDD global (coordinación, delegación, model-routing) vive en el `~/.claude/CLAUDE.md` del usuario") | REWRITE (ambos, son espejo) | Dejan de ser ciertos para Claude. |
| `CHANGES.md` (ficha 03) | UPDATE | "Registrá el orquestador nativo" pasa a "skill invocable"; responsable de 04 actualizado a Matías. |
| Matriz de reglas compartida con 04 (ubicación a acordar con 04) | NUEVO | Entregable acordado en el roadmap y en el informe unificado. |
| `obsidian.md` / nota del vault del proyecto | PENDING | No existe en el repo; destino a definir por el orquestador. |
