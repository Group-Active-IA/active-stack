# Proposal — codex-native-opsx (change 04 · Roles nativos en Codex)

> Responsable: Giuliano. Rama `feat/codex-native-opsx`; CHANGES.md "Cómo trabajar" §2 pide `propose/<nombre>`, la desviación está registrada en `tasks.md ## Orquestación`. Revisión de partida `bd6354349ca6fce29d4e4786d75ca9a4aed7ea60`. HEAD del checkout: `1b4639b`. `internal/` y `cmd/` son idénticos entre ambas (`git diff --quiet bd63543 HEAD -- internal cmd`, 2026-11-06), así que las citas `archivo:línea` valen para las dos.
> Esta ronda solo cubre `propose`. Todo lo de implementación queda como tareas pendientes en `tasks.md`.

> **Estado: borrador previo a las decisiones de grupo (propose intento 2).** CHANGES.md "Cómo trabajar" §4 pide resolver, antes de compartir la propuesta final, las decisiones que cambian alcance, diseño o aceptación. Esas decisiones son **G1–G13 y la rama D6 (a/b/c)** de `design.md`. Las tiene que cerrar el grupo **antes de compartir esta propuesta como final y antes de `apply`**. Después, el planner elimina la redacción condicional de specs y design y corre una nueva ronda de plan-reviewer. Este documento no decide ninguna.

## Why

La plantilla de Codex le pide al bloque que se aplique "solo al agente dedicado `sdd-orchestrator`" (`internal/harness/config/assets/codex/sdd-orchestrator.md:3`). Sin embargo, el adapter entrega ese bloque como instrucciones compartidas (`internal/agents/codex/adapter.go:64-66`) dentro de `~/.codex/AGENTS.md` (`adapter.go:18-20`), y no registra ningún rol.

Con los cinco toggles del catálogo (`internal/catalog/harnesses.yaml:70`), cada sesión de Codex que lee ese archivo recibe 18.010 bytes de coordinación (cuenta del bloque compuesto). Además, la plantilla anuncia `/opsx:*` (`sdd-orchestrator.md:63-67`), pero ni Active Stack (`adapter.go:26-28`) ni OpenSpec 1.16.0 los instalan para Codex (sondas del 2026-11-06: "Commands skipped for: codex (uses skills)").

Resolverlo ahora es condición para que 05 y 06 tengan una entrega por rol sobre la cual medir y verificar.

## What Changes

- **Nueva entrega de `sdd-orchestrator` para Codex.** El adapter de Codex declara una entrega propia en lugar de `ConfigDeliveryInstructions`. El nombre y el contrato los acuerdan 03 y 04 (G13). El instalador escribe cuatro piezas:
  - un **arranque breve** en la sección `active-stack:sdd-orchestrator` del `AGENTS.md`, o ningún arranque (G7), sin reglas de coordinación;
  - el **orquestador como skill de invocación explícita** (`sdd-orchestrator`) en el directorio de skills que resuelve el adapter;
  - los **módulos por acción** como piezas separadas, con un mapeo explícito de toggle a módulo;
  - los **especialistas** como roles nativos, si la versión objetivo los admite. El esquema TOML `name` / `description` / `developer_instructions` viene de la documentación de Codex y no está verificado [DOC]. Si el cliente no admite roles, la alternativa es la rama D6-c.
- **Migración del bloque administrado.** Instalar o actualizar con recibo reemplaza el bloque de 18 KB por el arranque:
  - con snapshot previo (`internal/backup`) y markers idempotentes (`internal/filemerge`);
  - preservando el contenido del usuario;
  - retirando también las copias bajo el marker heredado `jr-stack:`, aunque coexistan con la actual;
  - con rollback del pipeline si algo falla.
  
  La **reversión explícita** a la configuración previa depende de G12:
  - **(a)** un snapshot de migración dedicado que sobrevive a reinstalaciones;
  - **(b)** un límite documentado: la reversión solo es posible hasta la siguiente instalación o actualización.
- **Propiedad de archivos.** No se sobrescribe ni se borra una skill, un módulo o un TOML que no lleve la marca de Active Stack, ni un archivo del usuario dentro de un directorio administrado.
- **Verify y uninstall** conocen la entrega nativa. Uninstall sigue siendo solo de máquina.
- **Matriz por superficie** (CLI, extensión IDE, app de escritorio, cloud): para cada una, la invocación, el contexto por rol, la alternativa explícita y el límite. Qué superficies se soportan lo decide el grupo (G1).
- **Sin cambios** para Claude, OpenCode ni Antigravity, ni para la sección `design-mockups`. El objetivo de proyecto no se habilita en este change salvo que G10 lo decida: hoy `install` no tiene `--project`, y en proyecto Codex y Antigravity comparten `<root>/AGENTS.md`.
- **BREAKING (observable para usuarios de Codex).** Una sesión normal deja de actuar como coordinador OPSX. Para coordinar hay que invocar `$sdd-orchestrator` de forma explícita, o usar como alternativa las skills `$openspec-*`.

## Capabilities

### New Capabilities

- `codex-opsx-delivery`: cómo se materializa y se verifica OPSX en Codex. Cubre el arranque de sesión normal, el orquestador de invocación explícita, los especialistas con sus reglas (rol nativo o alternativa), los módulos por acción con el corte por módulo faltante, la matriz por superficie con su fallback y la no regresión de los demás agentes.
- `codex-sdd-migration`: cómo se migra y se revierte la configuración Codex previa. Cubre el reemplazo del bloque administrado, incluidos los markers heredados; la idempotencia; el rollback ante fallo; la propiedad de archivos; el punto de reversión (G12); la actualización desde una instalación anterior, con y sin recibo; y la desinstalación.

### Modified Capabilities

Ninguna. No hay spec vigente del `sdd-orchestrator` en `openspec/specs/` (`openspec list --specs`, 2026-11-06). `design-mockups-harness` conserva sus requisitos: su sección sigue en el `AGENTS.md` de Codex y no se purga.

## Impact

- **Código (implementación futura).**
  - `internal/model/harness.go`: nuevo valor de `ConfigDelivery` o contrato neutral (G13). Gobernanza BAJO.
  - `internal/agents/codex/adapter.go`: entrega y rutas de roles. Gobernanza MEDIO.
  - `internal/harness/config`: composición del paquete Codex, escritura con propiedad, retiro de markers heredados y, si G12-a, el snapshot de migración. Gobernanza MEDIO.
  - `internal/harness/config/assets/codex/`: assets nuevos.
  - `internal/install/plan.go`: rutas del snapshot y paso de preflight en Prepare.
  - `internal/uninstall`: retiro archivo por archivo. Gobernanza ALTO, porque toca configuración del usuario.
  - `internal/verify/harness_checks.go`: chequeos de la entrega nativa.
  - `cmd/active-stack/main.go`: `projectVerifyAdapter`, solo si G10 = sí.
  - `cmd/active-stack/headless/windows_backups.go`: slots de backup, solo si G12-a decide mostrar el snapshot de migración en la GUI de Windows.
- **Sin cambios previstos** en `internal/backup`, `internal/filemerge` ni `internal/pipeline`: se reutilizan tal cual. Si hiciera falta tocarlos, el plan vuelve a revisión (dominio ALTO, CLAUDE.md §4).
- **Catálogo.** No se prevén entradas nuevas. Como mucho, se ajusta la `long_description` bilingüe de `sdd-orchestrator` (`harnesses.yaml:68`), validando con `catalog.Load()`.
- **Configuración del usuario afectada.**
  - `~/.codex/AGENTS.md`;
  - `~/.agents/skills`;
  - el directorio de roles de Codex (rama D6);
  - `~/.active-stack/backups/` (si G12-a);
  - `~/.codex/config.toml`, solo con D6-b, en coordinación con 01 y 08.
- **Dependencias de otros changes.**
  - 02 `opsx-baseline`: protocolo de medición. Todavía no existe como change.
  - 03 `claude-native-opsx`: matriz de reglas, inventario de módulos y contrato de entrega compartido (G13). Tampoco existe todavía.
  - 01 `codex-strict-compat`: decisión de versiones y entornos, que G8 referencia en lugar de duplicar.
  - Consumidores: 05 y 06.
- **Externas.** Codex CLI con versión objetivo `(TBD, G8 alineada con 01)`. La versión observada en el informe ejecutivo es 0.166.0; Codex no está instalado en esta máquina. OpenSpec CLI 1.16.0.

## Clasificación

**HIGH.**

- **Riesgo.** El change reescribe y migra configuración del usuario: `~/.codex/AGENTS.md`, el directorio de skills compartido `~/.agents/skills` y, eventualmente, `~/.codex/config.toml`. También agrega retiro de piezas en uninstall. Destruir configuración del usuario es el peor daño posible del instalador (CLAUDE.md §4), y CHANGES.md fija gobernanza Alta para 04.
- **Alcance.** Cambia el comportamiento de todas las sesiones de Codex de los usuarios existentes, incluido el arranque del coordinador (BREAKING). Toca install, update (por replay), verify y uninstall, y define contratos que consumen 05 y 06.
- **No es CRITICAL.** No cruza autenticación, billing, secretos de producción ni integridad de auditoría, y el proyecto declara que no tiene dominios CRÍTICO (CLAUDE.md §4).
- **Exigencias que siguen de HIGH:** cinco roles separados; aprobación humana fechada antes de `apply`; TDD con regresiones de consumidores y un mutante sobre cada aserción clave; rondas de breaker hasta quedar limpio; y un auditor con chequeos independientes.

## Aprobación

> Esta sección la completa **solo el responsable humano**, con:
> - la fecha;
> - la opción elegida en cada decisión de grupo (G1–G13 y la rama D6 de `design.md`);
> - el alcance aprobado para `apply`.
>
> Ningún agente la completa ni la firma. Mientras esté vacía, la propuesta no es final y la construcción no puede empezar.

_(pendiente — completar por el humano)_

## Knowledge Base Impact

No hay `knowledge-base/`, ni lecciones `L-NNN`, ni `.atl/skill-registry.md` en el repo. Tampoco hay un `obsidian.md` designado por las instrucciones del proyecto, ni en el repo ni en `~/.claude/`. Documentos canónicos afectados:

| Documento | Acción | Motivo |
|---|---|---|
| `ARCHITECTURE.md` §2.1 (harness `sdd-orchestrator`) | REWRITE | Hoy afirma que el bloque "se inyecta en `CLAUDE.md` / `AGENTS.md`". Para Codex pasa a ser arranque + skill + módulos + roles. |
| `CLAUDE.md` y `AGENTS.md` del repo (§2 modelo de dominio) | REWRITE (ambos, en espejo) | Describen el `sdd-orchestrator` como un bloque inyectado. Tienen que nombrar la nueva entrega. |
| `README.md` (fila `sdd-orchestrator` y uso con Codex) | REWRITE | Cambia la invocación de OPSX en Codex (`$sdd-orchestrator`, fallback `$openspec-*`) y el alcance de la reversión según G12. |
| `CHANGES.md` (ficha 04) | APPEND | Registrar las decisiones de grupo y el resultado de la revisión conjunta 03/04. El estado del change sigue viviendo en OpenSpec. |
| `docs/` (guía de migración y reversión para Codex) | APPEND (documento nuevo) | Matriz por superficie, reversión y límites. Ruta `(TBD)` a acordar en grupo. |
