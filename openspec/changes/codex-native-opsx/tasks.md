# Tasks — codex-native-opsx

> **Nivel HIGH.**
>
> **Puerta B2.** Ninguna tarea de los grupos 2 en adelante empieza sin estas tres condiciones:
> - (a) aprobación humana fechada en `proposal.md ## Aprobación`, con G1–G13 y la rama D6 resueltas;
> - (b) informe del spike (grupo 1) incorporado a specs, design y tasks;
> - (c) una nueva ronda **obligatoria** del plan-reviewer sobre el plan actualizado.
>
> Cuando se decide una G, se borran las tareas y escenarios de la rama descartada (por ejemplo, 5.4 o 5.5 según G12; 3.3 y 5.3 si G10 = no).
>
> **TDD obligatorio en toda tarea de código.** Primero RED (falla de aserción esperada, no de compilación), después GREEN y refactor.
> - **[TDD]** registra bajo la tarea los comandos RED y GREEN, el timestamp, la revisión y la salida relevante.
> - **[DOC]** registra validación de formato o contrato.
> - **[MANUAL]** requiere Codex instalado y registra versión, superficie y evidencia.
>
> **Rigor HIGH.** Regresiones de consumidores (install, update, verify, uninstall; Claude, OpenCode, Antigravity) y un **mutante** sobre cada aserción clave de 9.2.
>
> **Lecciones `L-NNN`.** No existen. El constructor, el breaker y el auditor registran esa ausencia.
>
> **Build.** No hacer `go build` sin pedido explícito (CLAUDE.md §3). La verificación es `go test` sobre los paquetes tocados.

## 1. Spike de verificación del cliente (sin código de producción)

- [ ] 1.1 **[MANUAL][DOC]** Instalar Codex con la versión objetivo de G8 (la que decida 01 sobre versiones y entornos), en un entorno aislado con HOME y CODEX_HOME temporales. Registrar `codex --version`, el SO y la superficie. Verificar: `informes/spike-1-codex.md` con comandos y salidas.
- [ ] 1.2 **[MANUAL][DOC]** En CLI y en cada superficie candidata de G1, confirmar cada punto y registrarlo como EVIDENCIA o SUPOSICIÓN:
  - carga de `~/.codex/AGENTS.md` y `<root>/AGENTS.md` en la sesión principal y en los subagentes;
  - descubrimiento de skills en `~/.agents/skills` y `<root>/.agents/skills`;
  - si la metadata de skills (y de módulos, si fueran skills) aparece en toda sesión;
  - la invocación explícita `$sdd-orchestrator`;
  - el archivo y la clave para deshabilitar la invocación implícita, o su inexistencia (rama M5);
  - el esquema TOML de roles y su descubrimiento o registro (rama D6 a/b/c);
  - el parámetro de selección de rol al hacer spawn;
  - si `~` se interpreta en `developer_instructions`;
  - los feature flags;
  - el método para observar las instrucciones efectivas (transcripción, rollout o `codex.skill_invocation`).
- [ ] 1.3 **[DOC]** Incorporar los resultados a **specs, design y tasks**: rama D6 elegida, rama M5 y matriz por superficie. Quitar la redacción condicional que el spike resuelva y devolver el plan al plan-reviewer. Esta ronda es **obligatoria**, aunque el enfoque no cambie. Verificar: `openspec validate codex-native-opsx --strict` pasa y queda registrada la disposición del plan-reviewer.

## 2. Línea base de regresión de consumidores (antes de tocar producción)

- [ ] 2.1 **[TDD-characterization]** Reutilizar los goldens existentes `internal/harness/config/testdata/compose_all_toggles_{claude,opencode}.golden` (`compose_test.go:13-54`, flag `-update`) y agregar `antigravity`. Solo hay 4 adapters (`agents/factory.go:16-46`). gemini, cursor, windsurf y generic son variantes de asset: se agregan goldens de esas variantes solo para fijar sus bytes, aclarado en el test. Es caracterización y nace en verde. Para probar que el golden detecta cambios, se aplica un mutante temporal (alterar un byte de `assets/claude/sdd-orchestrator.md`), se observa la falla y se revierte. Verificar: `go test ./internal/harness/config/ -run 'TestCompose_AllToggles'` en verde y el mutante registrado.
- [ ] 2.2 **[TDD-characterization]** Fijar el comportamiento actual de OpenCode (agente primario) y de Claude (instrucciones) en `install_test.go` y `primary_agent_test.go`, si no está cubierto. Verificar: la lista de tests reutilizados y los nuevos queda registrada, en verde sobre la base.
- [ ] 2.3 Correr la suite relevante como baseline (`go test ./internal/... ./cmd/...`) y registrar las fallas preexistentes. Verificar: comando, timestamp y veredicto quedan anotados acá.

## 3. Modelo y adapter (BAJO / MEDIO). **3.1 bloqueada hasta que G13 esté acordado con 03**

- [ ] 3.1 `internal/model/harness.go`: agregar el valor de entrega acordado en G13, sin alterar el valor cero. `internal/model/agentpaths.go`: agregar `RolesDir`. Es un cambio de tipo y no tiene RED de comportamiento propio: lo cubre el RED de despacho de 3.2 y 5.2. Verificar: `go test ./internal/model/` en verde, sin regresión.
- [ ] 3.2 **[TDD]** `internal/agents/codex/adapter.go`: `ConfigDelivery()` devuelve la entrega nativa, y `RolesDir(home)` / `PathsFor(...).RolesDir` siguen la rama D6 (vacío en D6-c). RED: modificar `adapter_test.go:56-67`, que hoy espera `ConfigDeliveryInstructions`, y agregar las aserciones de `RolesDir`. Consumidores: `registry_test.go`, `assertions_test.go`. Verificar: `go test ./internal/agents/...`, con los otros 3 adapters sin cambios.
- [ ] 3.3 **[TDD] (solo si G10 = sí)** `projectTargetAdapter` (`internal/install/plan.go:556-575`) y `projectVerifyAdapter` (`cmd/active-stack/main.go:333-346`) implementan `RolesDir`. Uninstall no tiene wrapper de proyecto (R11): no se crea. RED: extender `internal/install/codex_project_internal_test.go`. Verificar: `go test ./internal/install/ ./cmd/active-stack/`.

## 4. Assets y composición del paquete Codex (MEDIO)

- [ ] 4.1 **[TDD]** Crear `internal/harness/config/assets/codex/native/` con:
  - `bootstrap.md`;
  - `orchestrator/SKILL.md` (con `$openspec-*`, la tabla acción → módulos y, en la rama M5, la descripción restrictiva);
  - `orchestrator/agents/openai.yaml`, si el spike lo confirma;
  - `modules/{tdd,delegation,subagents,governance,engram}.md`;
  - `roles/<id>.toml`, solo en las ramas D6-a/b;
  - en D6-c, las plantillas de especialista dentro del orquestador.
  
  RED: test de contrato de assets que, por pieza, exige existencia, frontmatter `name`, marca `managed-by` y ausencia de `/opsx:` y de `opus|sonnet|haiku` en assets de Codex. Verificar: `go test ./internal/harness/config/ -run CodexAssets`.
- [ ] 4.2 **[TDD]** Función pura de composición del paquete Codex y función de rutas administradas (D8). RED, uno por escenario:
  - toggles completos → 5 módulos y ninguno de model routing;
  - sin `tdd` → sin módulo TDD ni referencias;
  - sin `delegation` → sin delegación ni subagentes, y el orquestador lo declara;
  - `model-routing` solo → ningún módulo;
  - toggle desconocido → error;
  - determinismo frente al orden;
  - arranque ≤ G7 y sin títulos de firma;
  - arranque fuera de presupuesto o con título de firma → error;
  - las rutas administradas son la unión de todos los toggles y de la rama D6.
  
  Verificar: `go test ./internal/harness/config/ -run CodexBundle`.
- [ ] 4.3 **[TDD]** Serialización de las referencias a módulos dentro de `developer_instructions` con `go-toml/v2`, en el estilo decidido en D5 (`~` o relativa). Casos: home de Windows (`C:\Users\x`), ruta con espacios y, si G10 = sí, objetivo de proyecto sin home ni nombre de usuario. RED: parsear el TOML generado y comparar el texto. Verificar: `go test ./internal/harness/config/ -run CodexRoleTOML`.

## 5. Instalador nativo (MEDIO; configuración del usuario)

- [ ] 5.1 **[TDD]** Función de preflight de propiedad (D7): detecta la marca en `SKILL.md`, en los módulos y en el TOML; un archivo existente sin marca devuelve error nombrando la ruta. RED: fixtures con una skill y un rol homónimos del usuario. Verificar: `go test ./internal/harness/config/ -run CodexOwnership`, con comparación de bytes de los fixtures.
- [ ] 5.2 **[TDD]** Rama de entrega nativa en `config.Install` (`install.go:43`), con error si falta la interfaz opcional (D2). Pasos:
  - preflight;
  - retiro en bucle de los bloques `sdd-orchestrator` bajo todo prefijo e inyección del arranque (D3);
  - escritura de skill, módulos y roles con backup por archivo y `WriteFileAtomic`;
  - retiro de los módulos inactivos.
  
  La escritura pasa por un seam propio del instalador nativo (por ejemplo `writeManagedFileFn`), para inyectar fallos después de reescribir `AGENTS.md` (m5). RED con home temporal, uno por escenario:
  - instalación limpia;
  - contenido del usuario preservado;
  - títulos de firma en contenido del usuario intactos;
  - bloque de 18.010 B migrado con `design-mockups` intacto;
  - `jr-stack:` aislado;
  - `active-stack:` y `jr-stack:` coexistentes → cero `jr-stack:` y una sección;
  - G7 "sin arranque" sin `AGENTS.md` → no se crea el archivo;
  - snapshot fallido → sin escrituras;
  - segunda instalación sin reescrituras (`Result.AllAlready`);
  - catálogo sin `tdd` → retiro del módulo TDD.
  
  Verificar: `go test ./internal/harness/config/ -run CodexNative`.
- [ ] 5.3 **[TDD] (solo si G10 = sí)** Objetivo de proyecto a nivel API (`install.Options{Target: model.Project}`), más coinstalación con Antigravity según la resolución de G10. RED: ambos agentes sobre el mismo `<root>`; la verificación de ambos pasa. Verificar: `go test ./internal/install/ -run CodexProject`.
- [ ] 5.4 **[TDD] (solo si G12 = a)** Snapshot de migración dedicado (D12). Se crea solo al detectar el bloque anterior, en `~/.active-stack/backups/codex-migration-<ts>/`, con `AGENTS.md` y las rutas de D8. RED, uno por escenario:
  - primera migración → un snapshot;
  - reinstalar dos veces → el mismo snapshot intacto y ninguno nuevo;
  - dos `install` encadenados como en `update` (`update.go:362-368`) → exactamente un snapshot, y al restaurarlo vuelve el bloque byte a byte;
  - el snapshot aparece en `backup.ListManifests(~/.active-stack/backups)`;
  - si G12 cubre Windows, `windows_backups.go` lo lista (test en `cmd/active-stack/headless`).
  
  Verificar: `go test ./internal/harness/config/ -run CodexMigrationSnapshot` (y el test de headless, si aplica).
- [ ] 5.5 **[TDD] (solo si G12 = b)** Aviso del límite de reversión en la salida de la instalación que migra el bloque, sin aviso cuando no hay bloque anterior. RED: capturar la salida en ambos casos. Verificar: `go test ./internal/install/ -run CodexReversionNotice`.

## 6. Plan de instalación, snapshot y rollback (ALTO por rollback; sin tocar `backup`, `filemerge` ni `pipeline`)

- [ ] 6.1 **[TDD]** `collectWritePaths` (`internal/install/plan.go:307-310`) agrega cada ruta administrada de D8 como entrada propia y los directorios nuevos como dir hints. RED en `collect_write_paths_test.go`. Consumidor: `snapshot_test.go`. Verificar: `go test ./internal/install/ -run CollectWritePaths`.
- [ ] 6.2 **[TDD]** Rollback end-to-end con el **seam de escritura del instalador nativo** de 5.2, no con `configInstallFn`: un fallo al escribir un rol o un módulo después de reescribir `AGENTS.md` deja `AGENTS.md` (con contenido del usuario y el bloque anterior) idéntico, las piezas nuevas eliminadas y la skill preexistente del usuario intacta. Verificar: `go test ./internal/install/ -run CodexRollback`.
- [ ] 6.3 **[TDD]** Un fallo de verify duro (módulo ausente) dispara el rollback por el hook. Consumidor: `verify_hook_test.go`. Verificar: `go test ./internal/install/ -run CodexVerifyRollback`.
- [ ] 6.4 **[TDD]** Paso de preflight de propiedad en Prepare, antes del snapshot (`plan.go:151-160`). Antes de escribir el test, leer `internal/pipeline` y confirmar que un fallo en Prepare impide todo Apply (esto resuelve la [SUP] de D7). RED: una skill homónima del usuario → el plan falla y ningún archivo, ni `AGENTS.md`, cambia de bytes. Verificar: `go test ./internal/install/ -run CodexPreflight`.

## 7. Verify (MEDIO)

- [ ] 7.1 **[TDD]** `internal/verify/harness_checks.go`: `checkConfig` despacha la entrega nativa a los chequeos de Codex (D10), y la interfaz opcional de `RolesDir` va en `internal/verify/adapter.go`. RED en `codex_test.go`, un caso por estado:
  - instalación correcta → pasa;
  - `## Delegation Rules` dentro de la sección → falla;
  - `## Delegation Rules` y `## OPSX Workflow` solo en contenido del usuario → pasa;
  - `jr-stack:sdd-orchestrator` remanente → falla y nombra el marker;
  - skill ausente → falla con la ruta;
  - módulo de toggle activo ausente → falla;
  - `model-routing` activo → no exige módulo;
  - rol sin `developer_instructions` (D6-a/b) → falla con archivo y campo;
  - D6-c → no exige roles;
  - arranque fuera de presupuesto → falla;
  - G7 "sin arranque" con marker presente → falla, si aplica.
  
  Verificar: `go test ./internal/verify/`.
- [ ] 7.2 **[TDD]** Si G11 lo aprueba, aviso blando cuando queda el bloque anterior en una instalación sin recibo. Verificar: test con `Soft: true` en `codex_test.go`.
- [ ] 7.3 **[TDD]** Regresión: los chequeos de Claude, de OpenCode (`harness_checks.go:240-279`) y de `design-mockups` no cambian. Verificar: `go test ./internal/verify/` con `harness_checks_test.go` y `scenario_test.go` en verde sin tocar sus expectativas.

## 8. Uninstall y update (ALTO: configuración del usuario)

- [ ] 8.1 **[TDD]** Uninstall en máquina (D11). El snapshot de `internal/uninstall/plan.go:162-176` incluye cada ruta administrada. El retiro de `:229-241` borra el arranque y, archivo por archivo, solo lo que lleva marca; los directorios se eliminan solo si quedan vacíos, sin `RemoveAll` (`steps.go:199-205` no se reutiliza para estas piezas). RED, uno por escenario:
  - desinstalación completa con las reglas del usuario, `design-mockups` y la skill del usuario intactas;
  - `notas.md` del usuario dentro de `sdd-orchestrator/` → intacto tras el uninstall y tras un rollback forzado;
  - home sin piezas → no-op.
  
  Verificar: `go test ./internal/uninstall/`.
- [ ] 8.2 **[TDD]** Regresión de update con recibo: `installRuns` e `installArgs` (`update.go:353-368`) siguen reaplicando Codex. Test de integración a nivel install que simula el replay, con migración e idempotencia en dos corridas. Si G12 = a, también verifica que se conserva un solo snapshot de migración, en coordinación con 5.4. Verificar: `go test ./cmd/active-stack/ -run Update` y `go test ./internal/install/ -run CodexReplay`.
- [ ] 8.3 **[TDD]** Regresión sin recibo: con `update.State` vacío, `syncRecordedInstall` (`update.go:375-377`) no modifica `~/.codex/AGENTS.md` y la salida contiene el mensaje de migración. Ya está comprobado [EV] que `refreshSkills` no toca `sdd-orchestrator` (`update/skills.go:97`, R12); el test lo fija como regresión. Verificar: `go test ./cmd/active-stack/ ./internal/update/`.

## 9. Rigor HIGH transversal

- [ ] 9.1 **[TDD]** Regresión de los otros agentes: los goldens de 2.1 siguen en verde después de los grupos 3–8, y OpenCode conserva `agent.sdd-orchestrator` con `mode: primary`. Verificar: `go test ./internal/harness/config/ -run 'TestCompose_AllToggles|PrimaryAgent'`.
- [ ] 9.2 Mutantes sobre aserciones clave. Para cada una: alterar la producción temporalmente, ver el test en rojo y revertir; nunca entregar el mutante. Aserciones:
  - (a) bytes del usuario preservados (5.2);
  - (b) idempotencia sin reescritura (5.2);
  - (c) rollback byte a byte (6.2);
  - (d) preflight de propiedad (5.1 y 6.4);
  - (e) módulo faltante → falla de verify (7.1);
  - (f) ausencia de títulos de firma por bloque (4.2 y 7.1);
  - (g) uninstall preserva la skill del usuario y `notas.md` dentro del directorio administrado (8.1);
  - (h) retiro de `jr-stack:` coexistente (5.2 y 7.1);
  - (i) update sin recibo no modifica `AGENTS.md` (8.3);
  - (j) `design-mockups` intacto (5.2 y 8.1);
  - (k) punto de reversión conservado tras dos install (5.4, solo G12-a), o aviso del límite (5.5, solo G12-b).
  
  Verificar: tabla con mutante, test, salida en rojo y reversión registrada acá.
- [ ] 9.3 `gofmt -l`, `go vet` y la suite completa sobre la revisión entregada, una sola vez (`go test ./...`). Verificar: comando, timestamp y veredicto registrados. Sin `go build` salvo pedido.

## 10. Catálogo y documentación

- [ ] 10.1 **[TDD]** Si se aprueba, ajustar la `long_description` bilingüe de `sdd-orchestrator` (`internal/catalog/harnesses.yaml:68`). RED: test de catálogo que busca la mención de la entrega nativa de Codex. Verificar: `go test ./internal/catalog/` (pasa por `catalog.Load()`).
- [ ] 10.2 **[DOC]** `ARCHITECTURE.md` §2.1: REWRITE. `CLAUDE.md` y `AGENTS.md` del repo, §2: REWRITE en espejo. Verificar: `git diff --stat` muestra ambos espejos y una revisión manual confirma que son equivalentes.
- [ ] 10.3 **[DOC]** `README.md` y la guía en `docs/` (ruta `(TBD)`):
  - invocación `$sdd-orchestrator` y fallback `$openspec-*`;
  - el BREAKING;
  - la matriz por superficie y la de reglas;
  - la reversión según G12: ruta del snapshot de migración y su visibilidad en TUI/GUI, o el límite documentado.
  
  Verificar: los comandos documentados coinciden con los flags reales (`active-stack install --help`, `update`).
- [ ] 10.4 **[DOC]** `CHANGES.md`, ficha 04: APPEND de G1–G13 y D6 tal como las registró el humano. Verificar: coincidencia textual con `proposal.md ## Aprobación`.

## 11. Aceptación por superficie (después de la implementación; requiere 02 y Codex)

- [ ] 11.1 **[MANUAL]** Por superficie de G1, por separado, con el protocolo de 02 `opsx-baseline` (bloqueado hasta que 02 exista):
  - sesión normal sin bloque;
  - invocación explícita;
  - conjunto fijo de pedidos implícitos sin activación ("proponé un change", "implementá las tareas", "archivá el change"; rama M5 incluida);
  - especialista con sus módulos (o la alternativa D6-c);
  - apply con TDD;
  - módulo TDD borrado → informe y detención.
  
  Verificar: `informes/verify-<n>-acceptance-<superficie>.md` con evidencia por escenario de spec.
- [ ] 11.2 **[MANUAL]** Por superficie:
  - instalación limpia;
  - update desde una instalación con el bloque de 18 KB, con y sin recibo, incluido el caso de dos `install` encadenados;
  - reinstalación;
  - rollback con contenido del usuario;
  - reversión según G12.
  
  Verificar: informe con hashes de `AGENTS.md` antes y después.
- [ ] 11.3 Revisión conjunta 03/04 con los mismos escenarios antes de promover la migración. El acuerdo de código compartido (G13) ya se registró antes de la tarea 3.1. Verificar: acta enlazada y disposición registrada.

## Workflow follow-up

- Verify: breaker en rondas hasta quedar limpio; después, auditor con chequeos independientes.
- Archive: solo con APROBADO del auditor, los informes originales completos, la DEUDA explícita y `knowledge-base-report.md`, que declare la ausencia de KB y los documentos del Knowledge Base Impact.
- Derivación m12: el `chmod 0755` del directorio padre en `WriteFileAtomic` (`writer.go:39-42`) pasa como observación a 01/08. No se corrige en este change.

## Orquestación

| Campo | Valor |
|---|---|
| Fase | propose |
| Rol activo | plan-reviewer (tras la entrega del planner, intento 2) |
| Intento | 2 |
| Revisión / diff | HEAD `1b4639b` (citas de código de `bd63543`; `internal/` y `cmd/` idénticos). Sin cambios de código: solo artefactos OpenSpec, **sin trackear** (no ignorados), en `openspec/changes/codex-native-opsx/` |
| Rama | `feat/codex-native-opsx`. CHANGES.md "Cómo trabajar" §2 pide `propose/<nombre>`. La desviación viene del orquestador o del responsable; el planner no la decide. Queda registrada para que el humano la justifique o la renombre. |
| Bloqueos | Codex CLI no está instalado (`codex: command not found`, 2026-11-06). No existen KB, `L-NNN` ni `.atl/skill-registry.md`. 02 y 03 no existen como changes. G1–G13 y D6 abiertas (B2). Falta la aprobación humana (HIGH). |
| Siguiente acción | revisión del plan (plan-reviewer, intento 2) sobre las correcciones B1, M1–M8 y m1–m13 |
| Alcance de esta sesión | solo propose; apply en una sesión posterior, después de la aprobación humana |
| Informes originales | `informes/propose-1-planner.md`, `informes/propose-1-plan-reviewer.md`, `informes/propose-2-planner.md` |
| Merge de planificación y revisión | No aplica (HIGH: cinco roles separados) |
