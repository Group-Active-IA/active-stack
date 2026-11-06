# Design — codex-native-opsx

> Motivación y alcance: ver `proposal.md`. Requisitos: `specs/codex-opsx-delivery/spec.md` y `specs/codex-sdd-migration/spec.md`.
> Las citas `archivo:línea` son de `bd63543` y siguen idénticas en HEAD `1b4639b` para `internal/` y `cmd/`.
> Etiquetas de evidencia:
> - **[EV]** observado en el código o con un comando.
> - **[DOC]** documentación de Codex rust-v0.165.0 según el informe del orquestador; no ejecutado.
> - **[SUP]** suposición sin verificar.
> - **R*n*** evidencia del informe `informes/propose-1-plan-reviewer.md`.

> **Puerta B2.** Antes de compartir esta propuesta como final y antes de `apply`, el grupo tiene que cerrar G1–G13 y la rama D6 (a/b/c). CHANGES.md "Cómo trabajar" §4 lo exige porque estas decisiones cambian alcance, diseño o aceptación. Mientras sigan abiertas, este documento es un borrador condicional: no decide ninguna.

## Context

### Recorrido actual: adapter → plantilla → archivos efectivos

1. **Adapter de Codex** (`internal/agents/codex/adapter.go`) **[EV]**
   - Máquina:
     - `InstructionsPath` = `~/.codex/AGENTS.md` (:18-20)
     - `SkillsDir` = `~/.agents/skills` (:22-24)
     - `SettingsPath` = `~/.codex/config.toml` (:30-32)
     - sin directorio de comandos (:26-28)
   - Proyecto: `<root>/AGENTS.md`, `<root>/.agents/skills`, `<root>/.codex/config.toml` (:44-53).
   - Entrega: `ConfigDeliveryInstructions` (:64-66).
2. **Instalador de config** (`internal/harness/config/install.go`) **[EV]**
   - Rama por defecto (:63-78): `Compose(h.Toggles, "codex")` → `Inject` → `InjectSection(path, "sdd-orchestrator", …)` (`inject.go:85-87`).
   - `InjectSection` hace, en orden: backup (:107-111), purga de lo no propio (:119), inyección (:122) y escritura atómica sin reescritura si no hay cambios (:125).
   - Los directorios de backup son **fijos** y se sobrescriben en cada corrida:
     - `install.go:35-36` → `backups/sdd-orchestrator-<agent>`
     - `install/plan.go:141` → `backups/install` (R6)
3. **Composición** (`compose.go:34-106`) **[EV]**
   - Parte de la base `assets/codex/sdd-orchestrator.md`.
   - Quita las secciones de delegación y de modelos según los toggles (:57-64).
   - Agrega governance, Engram y TDD (:66-103).
   - Con los cinco toggles (`harnesses.yaml:70`) suma **18.010 bytes**. Es aritmética de tamaños, consistente según R17; `Compose` no se ejecutó.
4. **Plantilla de Codex** (181 líneas) **[EV]**
   - :3 dice "Bind this to the dedicated `sdd-orchestrator` agent only", pero ese agente no existe.
   - :17-36 delegación con umbral fijo.
   - :63-67 `/opsx:*`.
   - :71-83 consulta de estado.
   - :87-92 tabla de acción → skill.
   - :94 `$HOME/.agents/skills`.
   - :100 Skill Resolver.
   - :129-159 Engram.
   - :169-181 modelos de Claude (`opus`/`sonnet`/`haiku`, :174-179).
5. **Lo que existe para Codex en un proyecto** **[EV]**
   - Dos sondas independientes de `openspec init --tools codex` con OpenSpec 1.16.0 (planner 13:08:47Z y R3, 13:25:23Z):
     - 6 skills en `<proj>/.agents/skills/`
     - "Commands skipped for: codex (uses skills)"
     - invocación `$openspec-propose` (CLI/IDE) o selector de Skills en la app
   - Active Stack no instala esas skills:
     - el harness `openspec` es solo npm (`harnesses.yaml:16-22`);
     - `assets/skills/openspec-*` son placeholders.
   - **Tener los archivos instalados no activa OPSX.**
6. **Carga de instrucciones en Codex** **[DOC]**
   - `AGENTS.md` por árbol.
   - `developer_instructions` se suma; `model_instructions_file` reemplaza.
   - Precedencia de config: managed > `-c` > proyecto > perfil > usuario > cloud > sistema.
   - Skills: `SKILL.md` + `openai.yaml`.
   - Telemetría `codex.skill_invocation`.
   - Subagentes: spawn / `wait_agent` / `close_agent`.
   - **No confirmado [SUP]:**
     - descubrimiento y esquema de roles TOML;
     - selección de rol al hacer spawn;
     - si un subagente recibe `AGENTS.md`;
     - clave para deshabilitar la invocación implícita;
     - feature flags;
     - si la metadata de skills aparece en toda sesión.
7. **Objetivo de proyecto** **[EV, R9/R10]**
   - `install` no tiene `--project` (`headless/flags.go:134-145`).
   - `Target: model.Project` solo llega desde `starter add`, y ningún starter incluye `sdd-orchestrator`.
   - En proyecto, Codex y Antigravity resuelven el **mismo** `<root>/AGENTS.md` y el **mismo** `<root>/.agents/skills` (`antigravity/adapter.go:61-62` frente a `codex/adapter.go:47-48`).
8. **Uninstall** **[EV, R11]**
   - Es solo de máquina: no tiene wrapper de proyecto (`uninstall/types.go:48-62`).
   - `skillRemovalStep` hace `os.RemoveAll(SkillsDir/<id>)` (`uninstall/steps.go:199-205`).
9. **Backups visibles al usuario** **[EV]**
   - La TUI lista todo subdirectorio de `~/.active-stack/backups/` que tenga `manifest.json` (`internal/tui/backups_screen.go:32` → `backup.ListManifests`, `retention.go:175-181`).
   - La GUI de Windows lista solo los slots fijos `install` y `uninstall` (`cmd/active-stack/headless/windows_backups.go:24`).
   - `backup.Prune` no tiene llamadores en código de producción (grep, 2026-11-06), así que no hay retención automática.

### Consumidores del recorrido (todos [EV])

| Consumidor | Dónde | Qué asume hoy |
|---|---|---|
| Snapshot de install | `install/plan.go:141-160`, `:307-310` | Un harness `config` escribe solo `InstructionsPath`; el directorio del snapshot es fijo. |
| Wrapper de proyecto (install) | `install/plan.go:556-575` | Sin directorio de roles. |
| Verify | `verify/harness_checks.go:97-131` | El marker actual aparece exactamente una vez (`:106,:120`) y solo se cuenta el prefijo actual (R8). |
| Verify de proyecto | `cmd/active-stack/main.go:333-346` | Mismas rutas que el wrapper de install. |
| Uninstall (solo máquina) | `uninstall/plan.go:162-176`, `:229-241`; `steps.go:199-205` | Snapshot de `InstructionsPath`; retiro del marker; `RemoveAll` del directorio de la skill. |
| Update | `cmd/active-stack/update.go:353-413` | Reaplica `install --home` según el recibo, a veces **dos veces** (`installRuns`, :362-368, R7). Sin recibo, solo muestra un mensaje (:375-377). |
| Refresco de skills | `internal/update/skills.go:97` | Filtra `h.Type != HarnessSkill`; nunca toca `sdd-orchestrator` (R12). |
| Purga y markers | `inject.go:34-72`; `section.go:127-139` | La purga salta los IDs propios. `locateSection` prefiere el prefijo actual, de modo que un `jr-stack:` coexistente **sobrevive** (R8). Así lo fija `section_test.go:122-131`. |
| Backup | `snapshot.go:117-160`, `restore.go:35-104` | Un directorio existente se restaura como NO-OP. Cada archivo nuevo dentro de `~/.agents/skills` necesita su propia entrada (R5). |
| Escritura | `filemerge/writer.go:16-45` | Crea el archivo aunque el contenido esté vacío y aplica `chmod 0755` al directorio padre (R15). |
| Precedente | `primary_agent.go:43-122`, `model/harness.go:68-89` | OpenCode ya resolvió la misma fuga con una entrega propia. |

## Goals / Non-Goals

**Goals:**

- Entrega nativa de Codex en cuatro piezas, seleccionada por el adapter y sin rutas hardcodeadas.
- Reutilizar `internal/backup`, `internal/filemerge` e `internal/pipeline` **sin modificarlos**.
- Una única fuente de verdad para las rutas administradas, compartida por install, verify y uninstall.
- Que la aceptación por superficie se pueda ejecutar con el protocolo de 02.

**Non-Goals:**

- La entrega de Claude (03), OpenCode, Antigravity y las demás variantes de asset.
- El contrato estructurado de retorno (06), la lectura dirigida (05), la delegación selectiva (07) y el model routing para Codex (09).
- Exponer toggles en la UI.
- Editar la política de permisos (01).
- Agregar objetivo de proyecto a `install` o `uninstall`, salvo que G10 lo apruebe como alcance.
- El `chmod 0755` del directorio padre (`writer.go:39-42`). Es una observación preexistente que se deriva a 01/08 sin corregirse acá (m12).

## Decisions

### D1 · Una entrega propia para Codex, con nombre y contrato según G13

El adapter de Codex declara una entrega distinta de `ConfigDeliveryInstructions`, y `config.Install` despacha según esa entrega, igual que con `ConfigDeliveryPrimaryAgent` (`install.go:43-61`).

El nombre y el contrato los acuerdan 03 y 04 en **G13** antes de la tarea 3.1. Opciones:

- **(a)** `ConfigDeliveryCodexNative`: específico, sin riesgo de que 03 lo reutilice para Claude y termine en el instalador de Codex.
- **(b)** `ConfigDeliveryNativeRoles` neutral, con un contrato en el que el adapter provee el paquete de piezas y sus rutas, de modo que el instalador no sabe de Codex.

Recomendación: (b) si 03 también adopta una entrega nativa, para compartir install, verify y uninstall; (a) si 03 no la adopta o no hay acuerdo antes de `apply`. **Sin fundar del todo**: depende del diseño de 03, que todavía no existe.

Alternativas rechazadas:

- `if agent == codex` dentro de la rama por defecto: acopla el instalador al agente.
- Un harness de catálogo por pieza: duplica toggles que no se exponen en la UI (ARCHITECTURE.md §2.1) y cambia el picker de Custom.

### D2 · Interfaces opcionales estrechas para las rutas nuevas

`config.AgentAdapter` (`types.go:8-28`) no expone `SkillsDir`. Se define una interfaz opcional `{ SkillsDir(home) string; RolesDir(home) string }` que se resuelve con type assertion. Si el adapter declara la entrega nativa y no implementa la interfaz, el instalador devuelve un **error explícito**.

Cambios de soporte:

- `model.AgentPaths` gana `RolesDir`.
- Si G10 = sí, `projectTargetAdapter` (`plan.go:556-575`) y `projectVerifyAdapter` (`main.go:333-346`) implementan `RolesDir`.
- **Uninstall es solo de máquina** (R11). Su adapter recibe `RolesDir` mediante la misma interfaz opcional, sin wrapper de proyecto. Un objetivo de proyecto en uninstall sería alcance nuevo, fuera de este change.

Alternativa rechazada: agregar `RolesDir` a todas las interfaces. Obliga a tocar los 4 adapters (`agents/factory.go:16-46`, R14) y todos los fakes de test.

### D3 · Arranque bajo el marker `sdd-orchestrator`, con retiro de copias heredadas

El arranque reemplaza el contenido de la sección actual mediante `InjectSection`, que ya hace backup, purga e idempotencia.

**Corrección M1.** Antes de inyectar, el instalador nativo retira, en bucle y dentro del paquete config, todas las ocurrencias del ID `sdd-orchestrator` bajo cualquier prefijo reconocido: llama a `InjectMarkdownSection(…, "")` hasta que `MarkedSectionIDs` no lo devuelva. Después inyecta una sola sección actual. Con esto, `filemerge` y `section_test.go:122` no cambian.

**G7 = "sin arranque" (m11).** Si `AGENTS.md` no existe, no se escribe (`WriteFileAtomic` crearía un archivo vacío, R15). Si existe, se retiran los bloques del ID y no se agrega ninguno.

Borrador del arranque, 273 bytes según R17:

```
OPSX está disponible bajo demanda; esta sesión no es coordinadora.
Para coordinar un change, invocá `$sdd-orchestrator`.
Sin orquestador: `$openspec-propose`, `$openspec-apply-change`, `$openspec-archive-change`
(requieren `openspec init --tools codex` en el proyecto).
```

Alternativa rechazada: un marker nuevo. Exige otra sección propia y otra ruta de migración, sin beneficio.

### D4 · Orquestador como skill de invocación explícita

- La skill vive en `<skills-dir>/sdd-orchestrator/SKILL.md`, con frontmatter `name`, `description`, `metadata.managed-by: active-stack` y la versión del asset.
- La política de invocación va en `<skills-dir>/sdd-orchestrator/agents/openai.yaml` **[SUP: archivo y clave, a confirmar en el spike]**.
- Contenido, derivado de la plantilla:
  - rol y principios (:5-15) y flujo (:38-53);
  - estado (:71-83), ciclo de vida y CLI (:102-127), reglas (:161-168);
  - la tabla de acción → skill corregida: `$openspec-*`, sin `/opsx:*`, buscando las skills en el proyecto y en la máquina;
  - la tabla de acción → módulos;
  - la regla "si falta un módulo requerido, informar la ruta esperada y detener la acción".

**Rama sin mecanismo implícito (M5).** Si el spike no confirma cómo deshabilitar la invocación implícita:

- la `description` de la skill se redacta para no corresponder a pedidos comunes; por ejemplo: "Solo cuando el usuario invoca explícitamente $sdd-orchestrator";
- la aceptación 11.1 la mide con un conjunto fijo de pedidos implícitos;
- si alguno la activa, la superficie falla y se pasa a la alternativa de arranque de G2 (perfil `-p`).

Alternativas evaluadas:

- **Perfil de config** (`[profiles.opsx]`, `codex -p opsx`): escribe en `config.toml`, que comparten 01 y 08. Queda como alternativa de G2.
- **Mantener el orquestador en `AGENTS.md`**: es justamente el problema que este change corrige.

### D5 · Módulos por acción y mapeo de toggles (M6)

| Toggle | Módulo(s) | Fuente | El orquestador lo carga en | Lo recibe el especialista |
|---|---|---|---|---|
| `tdd` | TDD | `assets/strict-tdd.md` | apply | constructor |
| `delegation` | Delegación | plantilla :17-36 | toda acción coordinada | — |
| `delegation` | Coordinación de subagentes | plantilla :94-100 + spawn / `wait_agent` / `close_agent` **[DOC]** | toda delegación | — |
| `governance` | Governance | `assets/governance.md` | propose, apply | constructor (G4) |
| `engram` | Engram | plantilla :129-159 + `assets/engram-protocol.md` | inicio y cierre de acción | (G4) |
| `model-routing` | ninguno en Codex hasta 09 (G9) | plantilla :169-181 no se porta | — | — |

Sin `delegation`, el orquestador trabaja sin subagentes y lo declara.

**Ubicación (G5).** Recomendación: archivos `sdd-orchestrator/modules/<id>.md` dentro de la skill. La premisa "no se listan como skills en la sesión normal" es **[SUP]** y la confirma el spike.

**Referencia a módulos desde los especialistas (m6).**

- **Opción 1:** rutas relativas al directorio de skills que resuelve el adapter, expresadas como `<skills-dir>/sdd-orchestrator/modules/<id>.md` con `~` en máquina.
- **Opción 2:** rutas absolutas resueltas al instalar.

Recomendación: `~` o relativas, para que un objetivo de proyecto (si G10 = sí) no versione el home ni el nombre del usuario. Que Codex expanda `~` dentro de `developer_instructions` es **[SUP]**: la ruta es texto que el modelo lee, no la resuelve el cliente, y se confirma en el spike. Se serializa con `go-toml/v2` y se prueba con rutas Windows.

Al cambiar el catálogo entre versiones (m13), los módulos de toggles inactivos se retiran, con snapshot.

### D6 · Especialistas: tres ramas según el spike (M8)

Contenido de cada rol:

- `name` / `description` / `developer_instructions`, esquema **[DOC]**;
- marca `# managed-by: active-stack (sdd-orchestrator)`;
- el encargo, los módulos que le aplican y una sección de retorno mínima (estado, hallazgos, evidencia, módulos cargados), que reemplazará el contrato de 06.

La convención `name = "sdd_explorer"` es **[SUP]**.

| Rama | Condición del spike | Install | Verify | Riesgo |
|---|---|---|---|---|
| **D6-a** | Roles descubiertos por directorio (`~/.codex/agents/`) | Escribe solo archivos propios | Parseo y campos | Medio |
| **D6-b** | Roles registrados en `config.toml` (`[agents.<name>]` con `config_file`) **[SUP]** | `filemerge.UpsertTOMLSection` + backup, coordinado con 01/08 | Parseo + entrada en `config.toml` | Alto; el plan vuelve a revisión |
| **D6-c** | Ninguna superficie declarada admite roles nativos | No escribe roles | No espera roles | Bajo. Los especialistas viven en el orquestador como plantillas de mensaje para subagentes genéricos. |

"No modifiques" es una instrucción, no un sandbox. El control efectivo corresponde a 08.

### D7 · Propiedad de archivos con preflight en Prepare (m4)

El preflight de propiedad corre como **paso de Prepare antes del snapshot**, dentro del plan de install (`plan.go:151-160`). Si una ruta administrada existe **sin** marca, el plan falla antes de cualquier paso Apply, así que no se escribe nada.

Que el pipeline aborte sin ejecutar Apply cuando falla Prepare es **[SUP]**: `internal/pipeline` no fue leído, y se confirma en la tarea 6.4. El preflight también se repite dentro del instalador nativo, como defensa para llamadores directos.

Criterio análogo: `update.preflightRecordedSkills` (`update.go:416-418`).

Alternativa rechazada: saltar los conflictos con una advertencia. Deja el orquestador a medio instalar.

### D8 · Fuente única de rutas administradas

Una función pura del paquete config devuelve la **unión para todos los toggles** y la rama D6. La usan:

- `install.collectWritePaths` (`plan.go:307-310`): cada archivo como entrada propia y los directorios administrados nuevos como dir hint;
- uninstall: snapshot y retiro archivo por archivo;
- verify.

Hay una deriva preexistente del `SettingsPath` de OpenCode en install (`plan.go:307-310` frente a `uninstall/plan.go:172-174`). Es una observación fuera de alcance.

### D9 · Update sin código nuevo

La migración llega por el replay de `install` (`update.go:399-405`). Sin recibo, se mantiene `update.go:375-377`. `refreshSkills` no toca `sdd-orchestrator` **[EV]**: `update/skills.go:97` filtra `h.Type != HarnessSkill` (R12). El test de regresión se conserva igual.

### D10 · Verify (M2)

El alcance del chequeo de títulos de firma es **todo bloque del ID `sdd-orchestrator` bajo cualquier prefijo**, nunca el archivo entero. El contenido del usuario fuera de los markers no se inspecciona.

Chequeos duros:

- marker actual exactamente una vez, o ausencia según G7;
- ningún bloque con prefijo heredado (M1);
- arranque dentro del presupuesto y sin títulos de firma;
- skill del orquestador con su marca;
- módulos según el mapeo D5;
- roles según la rama D6.

La activación real en el cliente no se prueba acá; queda para la aceptación por superficie.

### D11 · Uninstall archivo por archivo (M3)

- Para cada ruta administrada de D8 en máquina, el retiro borra solo los archivos que llevan la marca.
- Después elimina cada directorio administrado **solo si quedó vacío**.
- No se usa `RemoveAll` sobre `sdd-orchestrator/`.
- El snapshot incluye cada archivo administrado, así el rollback los restaura.
- Los archivos del usuario dentro del directorio nunca se tocan, así que no necesitan restaurarse.

### D12 · Punto de reversión según G12 (B1)

Evidencia:

- los snapshots de install y de `InjectSection` se sobrescriben en cada corrida (R6);
- un `update` puede encadenar dos `install` (R7);
- la TUI lista cualquier subdirectorio con manifiesto (`backups_screen.go:32`);
- la GUI de Windows lista solo `install` y `uninstall` (`windows_backups.go:24`);
- no hay retención automática (`Prune` sin llamadores).

Opciones:

- **G12-a · Snapshot de migración dedicado.**
  - Cuándo: solo si se detecta el bloque anterior, es decir, una sección del ID con títulos de firma o con un prefijo heredado.
  - Qué hace: antes de migrar, el paquete config crea `~/.active-stack/backups/codex-migration-<UTC-timestamp>/` con `backup.NewSnapshotter().Create`, sin modificar `internal/backup`.
  - Cobertura: `AGENTS.md` y todas las rutas de D8.
  - Las corridas siguientes no lo recrean ni lo sobrescriben, porque ya no detectan el bloque.
  - Aparece en la TUI sin cambios. Para que aparezca en la GUI de Windows hay que agregar un slot o un listado por prefijo en `windows_backups.go`; si no se hace, se documenta el límite.
  - Costo: un directorio por migración, sin poda. Se crea una sola vez por migración real.
- **G12-b · Límite documentado.** La reversión solo es posible hasta la próxima instalación o actualización. Se avisa en la salida de la instalación que migra y en la guía. No hay código de retención.

**Recomendación: G12-a.** Es lo único que cumple el "admite rollback" de CHANGES.md ("Comportamiento objetivo") para el caso real (R7), sin tocar `internal/backup` y con visibilidad inmediata en la TUI. Sigue abierto, para que el grupo decida, si se cubre también la GUI de Windows.

Los specs tienen un requisito por rama y las tareas 5.4 y 5.5 tienen un RED por rama; al decidir G12 se borran el requisito y la tarea de la rama descartada.

### Matriz de reglas compartidas con 03 (borrador para G4)

| Pieza / regla | Sesión normal | Orquestador | Especialista | Codex (este change) | Claude (03) |
|---|---|---|---|---|---|
| Reglas del usuario (fuera de markers) | sí | sí | sí **[SUP, spike]** | `AGENTS.md` por árbol | `(TBD con 03)` |
| Permisos / sandbox | config del cliente | igual | heredado **[SUP]** | `config.toml` (01) | `settings.json` |
| Arranque breve | sí | sí | no | marker `sdd-orchestrator` | `(TBD con 03)` |
| Rol coordinador | no | sí | no | skill `sdd-orchestrator` | `(TBD con 03)` |
| Delegación + coordinación de subagentes | no | sí | no | módulos (`delegation`) | `(TBD con 03)` |
| TDD | no | apply | constructor | módulo (`tdd`) | `(TBD con 03)` |
| Governance | **G4** (`governance.md:3`: "applies to all agents") | sí | constructor | módulo o universal | `(TBD con 03)` |
| Engram | **G4** (`engram-protocol.md:4`: "MANDATORY and ALWAYS ACTIVE") | sí | **G4** | módulo o universal | `(TBD con 03)` |
| `design-mockups` | sí (sección propia) | sí | **G4** | sección en `AGENTS.md` | sección en `CLAUDE.md` |
| Model routing | no | **G9** | — | — | `(TBD con 03/09)` |

### Matriz por superficie (borrador para G1/G2)

| Superficie | Lee `~/.codex/AGENTS.md` | Skills | Roles nativos | Invocación explícita | Recomendación |
|---|---|---|---|---|---|
| CLI | **[DOC]** | **[DOC]** | spike | `$sdd-orchestrator` | Soportar |
| Extensión IDE | **[SUP]** | **[EV indirecta]** (texto de OpenSpec, R3) | spike | `$sdd-orchestrator` | Soportar si el spike lo confirma |
| App de escritorio | **[SUP]** | **[EV indirecta]** (selector de Skills) | spike | selector | Condicional |
| Cloud | **[SUP]** no ve el home | solo el repo | spike | — | Excluir con el límite documentado. La alternativa por proyecto depende de G10. |

### Decisiones de grupo pendientes (bloquean la propuesta final y `apply`)

| ID | Decisión | Opciones | Recomendación (fundamento) | Afecta |
|---|---|---|---|---|
| G1 | Superficies soportadas | CLI / +IDE / +app / +cloud | CLI e IDE; app condicional; cloud excluido. *Parcial*: CLI [DOC], IDE y app [EV indirecta], cloud [SUP]. | aceptación, docs |
| G2 | Arranque del coordinador | skill explícita · perfil `-p` · `AGENTS.md` de proyecto | Skill explícita; perfil si falla la rama M5 (D4). *[SUP]* | D4 |
| G3 | Fallback | `$openspec-*` · restaurar el bloque · ninguno | `$openspec-*`. *[EV]*: R3 y la sonda del planner. | spec |
| G4 | Reglas universales frente a módulo (governance, Engram, design-mockups) | universal · módulo · híbrido | Híbrido, sujeto a la medición de 02. *[SUP]*: 02 no existe. | 03, 05 |
| G5 | Inventario y ubicación de módulos | archivos en la skill · skills separadas | Archivos en la skill (D5). *[SUP]*: listado de metadata. | 03, 05, 07 |
| G6 | Inventario de especialistas | por acción OPSX · cinco roles lean · mínimo | Alinear con 03; si no hay acuerdo, por acción OPSX. *[EV]*: plantilla :87-92. | 03, 06, 07 |
| G7 | Arranque | ninguno · ≤ 1 KiB · ≤ 2 KiB | ≤ 1 KiB. *[EV]*: el borrador mide 273 B (R17). | spec, verify |
| G8 | Versión de Codex y flags experimentales | — | **Referenciar la decisión de 01** sobre versiones y entornos, sin duplicarla. No habilitar flags sin acuerdo. | spike, 01 |
| G9 | Model routing en Codex | omitir hasta 09 · mapear a modelos de Codex | Omitir. *[EV]*: :174-179 usa nombres de Claude. | 09, D5 |
| G10 | Objetivo de proyecto | sí, con resolución de la colisión con Antigravity · no (límite declarado) | **Recomendación previa retirada por no fundada** (R9, R10): `install` no tiene `--project` y la colisión con Antigravity rompe el requisito de no regresión. Si "sí": alcance nuevo (flag o starter) y una regla de coinstalación. Si "no": quitar las tareas 3.3 y 5.3 y los escenarios condicionales. | spec, tests |
| G11 | Instalaciones sin recibo | solo mensaje · + aviso blando en verify · migración automática | Mensaje + aviso blando. *[EV]*: `update.go:375-377`, R12. | update, verify |
| G12 | Retención del punto de reversión | (a) snapshot de migración dedicado · (b) límite documentado | (a) (D12). *[EV]*: R6, R7, `backups_screen.go:32`. Sub-decisión: cobertura en la GUI de Windows. | spec, 5.4/5.5, docs |
| G13 | Nombre y contrato de la entrega, compartido con 03 | (a) `ConfigDeliveryCodexNative` · (b) neutral con paquete provisto por el adapter | Según la adopción de 03 (D1). *Sin fundar*: 03 no existe. **Bloquea la tarea 3.1.** | 03, D1 |
| D6 | Rama de especialistas | a · b · c | La decide el spike. b vuelve a revisión. | spec, 4.1, 7.1 |

## Risks / Trade-offs

- **[Comportamiento de Codex no verificado]** → Spike obligatorio. La tarea 1.3 actualiza specs, design y tasks, y la re-revisión del plan es **obligatoria**.
- **[BREAKING: sin coordinación implícita]** → El arranque indica cómo invocar, más README, matriz, fallback `$openspec-*` y punto de reversión (G12).
- **[Pérdida del punto de reversión]** (R6, R7) → G12; ambas ramas tienen requisito y RED.
- **[Bloque heredado coexistente sigue cargando 18 KB]** (R8) → D3 (bucle) y D10 (detección), con mutante.
- **[Falso positivo por títulos del usuario]** → Alcance por bloque del ID (D10), con escenario.
- **[Pérdida de contenido del usuario en uninstall]** → D11 archivo por archivo, con escenario y mutante.
- **[Rollback incompleto en directorios existentes]** (R5) → D8, con test end-to-end.
- **[Edición de `config.toml` en D6-b]** → Vuelta a revisión, coordinación con 01/08, `ValidateTOML`.
- **[Choque con 03 en install y verify]** (M7) → G13 resuelto antes de la tarea 3.1, más revisión conjunta.
- **[Colisión con Antigravity en proyecto]** (R9) → G10; fuera de alcance salvo decisión.
- **[Ahorro sin medir]** → Solo se afirma lo que mida 02.
- **[Engram y governance salen de la sesión normal]** → G4.

## Migration Plan

1. **Install o update con recibo:**
   1. Prepare: preflight de propiedad (D7).
   2. Prepare: snapshot del pipeline con todas las rutas de D8.
   3. Apply: si G12 = a y se detecta el bloque anterior, snapshot de migración dedicado.
   4. Apply: retiro de los bloques del ID bajo todo prefijo e inyección del arranque (D3).
   5. Apply: skill, módulos y roles según D6; retiro de los módulos inactivos.
   6. Verify duro. Si algo falla, el pipeline hace rollback al snapshot del paso 2.
2. **Reinstalación:** no escribe archivos sin cambios y no crea otro snapshot de migración.
3. **Reversión explícita:**
   - G12-a: restaurar `codex-migration-<ts>` desde la TUI de backups (y desde la GUI de Windows, si se cubre).
   - G12-b: restaurar el snapshot `install` solo antes de la siguiente instalación; el límite queda documentado.
4. **Sin recibo:** sin cambios; el mensaje existente indica cómo migrar.
5. **Promoción:** después de la aceptación por superficie y de la revisión conjunta con 03.

## Open Questions

Preguntas que no cambian specs ni tareas:

- El nombre del campo de versión del asset en el frontmatter.
- La ruta de la guía en `docs/`.

Todo lo que sí cambia el plan es G1–G13 y D6, y bloquea según la puerta B2.
