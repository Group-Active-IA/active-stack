## Purpose

Define cómo Active Stack materializa el flujo OPSX en Codex. Una sesión normal no carga el bloque de coordinación. El orquestador arranca solo cuando se lo invoca explícitamente. Los especialistas y los módulos por acción se entregan por separado, con sus reglas. Cada superficie declarada documenta su invocación, el contexto de cada rol y la alternativa.

> Borrador condicionado a G1–G13 y a la rama D6 de `design.md`. Las frases "según G…" y "según el spike" se reemplazan por la decisión registrada antes de `apply`.

## ADDED Requirements

### Requirement: La sesión normal de Codex no recibe el bloque de coordinación

Cuando se instala `sdd-orchestrator` para Codex, todo bloque del ID `sdd-orchestrator` en el archivo de instrucciones que resuelve el adapter SHALL contener como máximo un arranque breve dentro del presupuesto de G7. Cuentan tanto el bloque bajo el prefijo actual como cualquier prefijo heredado. El arranque no lleva reglas de coordinación, delegación, TDD, governance, Engram ni asignación de modelos. El contenido del usuario fuera de esos markers MUST quedar intacto.

#### Scenario: Instalación limpia en máquina

- **GIVEN** un home de prueba sin `~/.codex/AGENTS.md`
- **WHEN** el usuario ejecuta `active-stack install --agent codex --mode lite`
- **THEN** `~/.codex/AGENTS.md` contiene el marker `active-stack:sdd-orchestrator` exactamente una vez
- **AND** no contiene ningún marker `jr-stack:sdd-orchestrator`
- **AND** la sección no supera el presupuesto de bytes acordado
- **AND** la sección no contiene ninguno de los títulos de firma del bloque anterior: `## Delegation Rules`, `## OPSX Workflow`, `## Model Assignments`, `Strict TDD Mode`, `## Agent Governance`, `## Engram Persistent Memory`

#### Scenario: Sin arranque en un home sin AGENTS.md (si G7 elige "sin arranque")

- **GIVEN** un home de prueba sin `~/.codex/AGENTS.md`
- **WHEN** se instala para Codex
- **THEN** no se crea `~/.codex/AGENTS.md`

#### Scenario: Contenido del usuario preservado en instalación limpia

- **GIVEN** un `~/.codex/AGENTS.md` con reglas del usuario y sin markers de Active Stack
- **WHEN** se instala `sdd-orchestrator` para Codex
- **THEN** los bytes del usuario se conservan idénticos antes de la sección
- **AND** se crea un snapshot antes de la escritura

#### Scenario: Títulos de firma en contenido del usuario

- **GIVEN** un `AGENTS.md` cuyo contenido del usuario, fuera de los markers, contiene `## Delegation Rules` y `## OPSX Workflow`
- **WHEN** se instala y después corre la verificación
- **THEN** la verificación pasa
- **AND** los bytes del usuario quedan intactos

#### Scenario: Sesión normal en una superficie declarada

- **GIVEN** una instalación completa en una superficie declarada en G1, con la versión objetivo de Codex `(TBD)`
- **WHEN** el usuario abre una sesión sin invocar al orquestador y pide una tarea común
- **THEN** las instrucciones efectivas de la sesión no contienen los títulos de firma, observadas con el método que fije el spike y medidas con el protocolo de 02
- **AND** no queda registrada ninguna invocación del orquestador

#### Scenario: Composición fuera de presupuesto

- **WHEN** el arranque compuesto supera el presupuesto acordado o contiene un título de firma
- **THEN** la composición falla con un error que nombra la causa
- **AND** no se escribe ningún archivo

### Requirement: El orquestador se instala para invocación explícita

El instalador SHALL instalar el orquestador `sdd-orchestrator` como skill de Codex en el directorio de skills que resuelve el adapter. La skill MUST incluir el rol de coordinador, la consulta de estado al CLI de OpenSpec y la tabla que asigna módulos a cada acción. Cuando el spike confirme que el cliente permite deshabilitar la invocación implícita, la skill MUST declararlo.

#### Scenario: Skill del orquestador instalada

- **WHEN** se instala `sdd-orchestrator` para Codex en máquina
- **THEN** existe `<skills-dir>/sdd-orchestrator/SKILL.md` con frontmatter `name: sdd-orchestrator`, una descripción no vacía y la marca de propiedad de Active Stack

#### Scenario: Invocación explícita activa el coordinador

- **GIVEN** una instalación completa en una superficie declarada
- **WHEN** el usuario invoca `$sdd-orchestrator`, o el mecanismo explícito que indica la matriz para esa superficie
- **THEN** la sesión ejecuta `openspec list --json` antes de actuar sobre un change
- **AND** la telemetría o la transcripción registra una invocación explícita del orquestador

#### Scenario: Skill del orquestador ausente

- **GIVEN** una instalación en la que falta `<skills-dir>/sdd-orchestrator/SKILL.md`
- **WHEN** corre la verificación post-instalación
- **THEN** el chequeo del orquestador de Codex falla con un error que nombra la ruta esperada

### Requirement: El orquestador no se activa por pedidos implícitos

En una sesión normal, el orquestador SHALL NOT activarse sin una invocación explícita. Si el cliente admite deshabilitar la invocación implícita, se usa ese mecanismo. Si no, la skill MUST tener una descripción que no corresponda a pedidos comunes, y esa protección MUST verificarse con pedidos implícitos medidos. Si la medición falla, se pasa a la alternativa de arranque acordada en G2.

#### Scenario: Pedido implícito con la invocación implícita deshabilitada

- **GIVEN** que el spike confirmó el mecanismo para deshabilitar la invocación implícita
- **WHEN** en una sesión normal el usuario escribe "proponé un change" sin invocar al orquestador
- **THEN** no queda registrada ninguna invocación implícita de `sdd-orchestrator`

#### Scenario: Cliente sin mecanismo para deshabilitar la invocación implícita

- **GIVEN** que el spike no encontró un mecanismo para deshabilitar la invocación implícita
- **WHEN** en la aceptación 11.1 se envía, en sesiones normales, el conjunto acordado de pedidos implícitos, incluidos "proponé un change", "implementá las tareas" y "archivá el change"
- **THEN** ninguno registra una invocación de `sdd-orchestrator`
- **AND** si alguno la registra, la aceptación de esa superficie falla y el arranque pasa a la alternativa de G2

### Requirement: Los especialistas reciben sus reglas y su contrato de retorno

Cada especialista acordado en G6 SHALL recibir, como rol nativo o mediante la alternativa de la rama D6, una descripción y unas instrucciones con:

- el encargo;
- los módulos que le aplican;
- la sección de retorno.

Las instrucciones MUST NOT incluir reglas que no le apliquen.

#### Scenario: Rol nativo bien formado (ramas D6-a y D6-b)

- **WHEN** se instala `sdd-orchestrator` para Codex
- **THEN** cada archivo de rol se parsea como TOML válido
- **AND** su `name` coincide con el nombre acordado
- **AND** `name`, `description` y `developer_instructions` no están vacíos (esquema [DOC], confirmado en el spike)

#### Scenario: Sin roles nativos (rama D6-c)

- **GIVEN** que el spike concluyó que ninguna superficie declarada admite roles nativos
- **WHEN** se instala `sdd-orchestrator` para Codex
- **THEN** no se escribe ningún archivo de rol
- **AND** la verificación no espera archivos de rol
- **AND** el orquestador contiene, por cada especialista, el encargo, las reglas y los módulos que debe pasarle a un subagente genérico

#### Scenario: Especialista recibe sus reglas

- **GIVEN** el orquestador invocado en una superficie declarada
- **WHEN** delega una tarea a un especialista que requiere el módulo TDD
- **THEN** el especialista informa en su retorno los módulos que cargó, y el TDD figura entre ellos
- **AND** las reglas generales del usuario siguen vigentes, observado con el método del spike

#### Scenario: Superficie sin roles nativos

- **GIVEN** una superficie declarada en la que el spike no confirmó roles nativos
- **WHEN** el orquestador necesita un especialista
- **THEN** usa la alternativa documentada en la matriz: un subagente genérico que recibe en el mensaje el encargo, las reglas del rol y los módulos
- **AND** informa al usuario que usó esa alternativa

#### Scenario: Rol con campo faltante

- **GIVEN** un archivo de rol administrado sin `developer_instructions`, en las ramas D6-a o D6-b
- **WHEN** corre la verificación post-instalación
- **THEN** el chequeo del especialista falla con un error que nombra el archivo y el campo

### Requirement: Los módulos siguen un mapeo explícito de toggle a módulo

El instalador SHALL instalar los módulos según el mapeo acordado en G5. La propuesta actual es:

| Toggle | Módulos |
|---|---|
| `tdd` | TDD |
| `delegation` | delegación + coordinación de subagentes |
| `governance` | governance |
| `engram` | Engram |
| `model-routing` | ninguno en Codex hasta 09 (G9) |

Un toggle inactivo MUST omitir sus módulos y sus referencias. El orquestador MUST cargar solo los módulos que requiere la acción activa.

#### Scenario: Toggles completos

- **WHEN** se instala `sdd-orchestrator` con los toggles `tdd, engram, model-routing, delegation, governance`
- **THEN** existen los módulos de TDD, delegación, coordinación de subagentes, governance y Engram, cada uno con su marca de propiedad
- **AND** no existe un módulo de model routing
- **AND** la tabla de acción a módulos del orquestador menciona cada módulo instalado

#### Scenario: Toggle inactivo

- **WHEN** se instala con todos los toggles salvo `tdd`
- **THEN** no existe el módulo TDD
- **AND** ni el orquestador ni los especialistas lo mencionan

#### Scenario: Sin delegation

- **WHEN** se instala con todos los toggles salvo `delegation`
- **THEN** no existen los módulos de delegación ni de coordinación de subagentes
- **AND** el orquestador indica que trabaja sin delegar

#### Scenario: model-routing activo en Codex

- **WHEN** se instala con `model-routing` activo
- **THEN** ningún asset de Codex contiene asignaciones de modelo ni los nombres `opus`, `sonnet` o `haiku`
- **AND** la verificación no espera un módulo de model routing

#### Scenario: Acción que requiere TDD

- **GIVEN** el orquestador invocado con el toggle `tdd` activo
- **WHEN** el usuario pide aplicar las tareas de un change
- **THEN** el orquestador carga el módulo TDD y lo pasa al especialista constructor antes de delegar

#### Scenario: Acción de exploración no carga TDD

- **GIVEN** el orquestador invocado
- **WHEN** el usuario pide explorar un tema
- **THEN** el módulo TDD no se carga en esa acción, según registra el método del spike

### Requirement: Un módulo obligatorio faltante detiene la acción que lo requiere

Si falta un módulo que la acción activa requiere, el orquestador y los especialistas SHALL informar cuál falta y detener esa acción, sin continuar con un sustituto implícito. La verificación post-instalación MUST fallar cuando falta un módulo de un toggle activo.

#### Scenario: Módulo TDD borrado

- **GIVEN** una instalación con el toggle `tdd` activo y el módulo TDD borrado del disco
- **WHEN** el usuario pide al orquestador aplicar tareas
- **THEN** el orquestador informa que falta el módulo TDD y la ruta donde lo esperaba
- **AND** no delega la construcción

#### Scenario: Verificación detecta el módulo faltante

- **GIVEN** el mismo estado
- **WHEN** corre la verificación post-instalación
- **THEN** el chequeo del módulo falla, nombra el módulo y la ruta
- **AND** el resultado global queda en falla

### Requirement: La verificación post-instalación comprueba la entrega nativa de Codex

La verificación SHALL comprobar las piezas Codex instaladas:

- el marker actual del arranque, una vez o ninguna según G7;
- ningún bloque `sdd-orchestrator` heredado;
- ningún título de firma dentro de esos bloques;
- la skill;
- los módulos del mapeo;
- los roles según D6.

Las rutas MUST resolverse vía adapter. Una falla dura MUST disparar el rollback del hook.

#### Scenario: Instalación correcta

- **WHEN** corre la verificación tras una instalación completa de Codex
- **THEN** todos los chequeos de la entrega nativa pasan

#### Scenario: Bloque anterior remanente

- **GIVEN** un `AGENTS.md` cuya sección `sdd-orchestrator` todavía contiene `## Delegation Rules`
- **WHEN** corre la verificación
- **THEN** el chequeo del arranque falla e indica que queda contenido del bloque anterior

#### Scenario: Bloque heredado remanente

- **GIVEN** un `AGENTS.md` con un arranque válido bajo `active-stack:` y un bloque `jr-stack:sdd-orchestrator`
- **WHEN** corre la verificación
- **THEN** el chequeo falla y nombra el marker heredado

#### Scenario: Objetivo de proyecto (solo si G10 = sí)

- **GIVEN** `install.Options{Target: Project}` sobre `<root>`, porque la CLI de `install` no tiene `--project`
- **WHEN** se instala y se verifica para Codex
- **THEN** la verificación busca el arranque en `<root>/AGENTS.md` y las piezas en las rutas de proyecto que resuelve el adapter

#### Scenario: Coinstalación Codex y Antigravity en proyecto (solo si G10 = sí)

- **GIVEN** Codex y Antigravity seleccionados con objetivo de proyecto, ambos sobre `<root>/AGENTS.md`
- **WHEN** se instala
- **THEN** el resultado coincide con la resolución acordada en G10
- **AND** la verificación de ambos agentes pasa sin que ninguno pise la entrega del otro

### Requirement: Cada superficie declarada documenta invocación, contexto por rol y alternativa

Para cada superficie de Codex acordada en G1, la documentación SHALL indicar:

- cómo se invoca el orquestador;
- qué instrucciones y módulos recibe cada rol;
- qué alternativa explícita mantiene OPSX disponible.

Una superficie no soportada MUST declarar el límite y su alternativa, si existe.

#### Scenario: Superficie soportada

- **WHEN** un usuario consulta la matriz para la CLI de Codex
- **THEN** encuentra el mecanismo de invocación, el contexto por rol y la versión verificada
- **AND** encuentra la evidencia de la corrida de aceptación de esa superficie

#### Scenario: Superficie no soportada

- **WHEN** un usuario consulta la matriz para una superficie excluida en G1
- **THEN** encuentra el motivo del límite y la alternativa explícita, o la indicación de que no hay alternativa

#### Scenario: Alternativa sin orquestador

- **GIVEN** un proyecto inicializado con `openspec init --tools codex`, que instala `openspec-propose` y las demás skills en `<root>/.agents/skills/`
- **WHEN** el usuario invoca `$openspec-propose` sin el orquestador
- **THEN** el flujo de propuesta de OpenSpec sigue disponible

### Requirement: La entrega de los demás agentes no cambia

El cambio de entrega SHALL limitarse a Codex. Se mantiene lo siguiente:

- Claude y Antigravity MUST seguir recibiendo el bloque compuesto en su archivo de instrucciones.
- OpenCode MUST seguir registrándolo como agente primario.
- Los bytes compuestos para todas las variantes de asset distintas de Codex MUST coincidir con los de la revisión de partida.

#### Scenario: Regresión de Claude y Antigravity

- **WHEN** se compone `sdd-orchestrator` para `claude` y `antigravity` con los toggles del catálogo
- **THEN** el contenido coincide byte a byte con el golden fijado en `bd63543`

#### Scenario: Regresión de OpenCode

- **WHEN** se instala `sdd-orchestrator` para OpenCode
- **THEN** `opencode.json` contiene `agent.sdd-orchestrator` con `mode: primary`
- **AND** el `AGENTS.md` de OpenCode no contiene la sección del orquestador
