# claude-coordinator-skill Specification

## Purpose
Entrega el coordinador OPSX como una skill de Claude Code que sólo el usuario invoca, que continúa
en la sesión actual, carga sus módulos según la acción y respeta la versión, la interfaz y el
target de instalación.

## Requirements

### Requirement: Coordinador entregado como skill administrada
Al instalar `sdd-orchestrator` para Claude en una versión compatible, el instalador MUST crear el
directorio de skill `sdd-orchestrator` en el directorio de skills del target, con `SKILL.md` que
declara `disable-model-invocation: true` y la marca de propiedad de Active Stack, y con un archivo
de módulo por cada módulo activo. Actor: instalador. Disparador: install. Resultado: directorio
presente.

#### Scenario: Install Machine crea la skill
- **WHEN** se instala `sdd-orchestrator` para Claude con target Machine
- **THEN** existe `<home>/.claude/skills/sdd-orchestrator/SKILL.md` con `disable-model-invocation: true`, marca de propiedad y la definición del rol coordinador, y `modules/` contiene sólo los módulos de toggles activos

#### Scenario: Install Project crea la skill en el proyecto
- **WHEN** se instala con target Project y raíz `<root>`
- **THEN** existe `<root>/.claude/skills/sdd-orchestrator/SKILL.md` y no se escribe nada en `<home>/.claude/skills/`

### Requirement: Invocación explícita en la sesión actual
El coordinador MUST activarse sólo cuando el usuario escribe `/sdd-orchestrator` y MUST continuar
en el mismo hilo de conversación. El instalador MUST NOT escribir el setting `agent`. Actor:
usuario. Disparador: `/sdd-orchestrator` en una sesión abierta. Resultado: registro de contexto
con testigo del coordinador en esa sesión.

#### Scenario: La invocación manual activa el coordinador en el hilo
- **WHEN** en una sesión normal ya iniciada, con un turno previo que contiene el testigo de conversación K, el usuario escribe `/sdd-orchestrator`
- **THEN** el registro de contexto de la misma sesión contiene el testigo del coordinador, el testigo universal y K (la conversación previa sigue disponible)

#### Scenario: Pedido OPSX sin invocar no activa el coordinador
- **WHEN** en una sesión normal el usuario pide "proponé un change para X" sin escribir `/sdd-orchestrator`
- **THEN** el registro de contexto no contiene el testigo del coordinador ni el de ningún módulo, y la respuesta no lanza la skill

#### Scenario: El modelo no ve la skill
- **WHEN** se abre una sesión normal después del install
- **THEN** la lista de skills que el registro de contexto expone al modelo no contiene la descripción de `sdd-orchestrator` (efecto de `disable-model-invocation: true`, validado por la tarea 1.10); aserción determinista, una corrida alcanza

#### Scenario: El instalador no fija el agente por defecto
- **WHEN** termina un install o update para Claude
- **THEN** `settings.json` del target no tiene la clave `agent` escrita por Active Stack y su valor previo, si existía, queda igual

### Requirement: El coordinador no corre como especialista
Un subagente MUST NOT recibir la definición del rol coordinador, ni por precarga de skills ni por
herencia de la conversación. Actor: coordinador. Disparador: lanzamiento con la herramienta Agent.
Resultado: transcripción del subagente sin testigo del coordinador.

#### Scenario: Intento de delegar al coordinador
- **WHEN** el usuario pide a la sesión lanzar un subagente "que actúe como sdd-orchestrator", o una definición de subagente lista `sdd-orchestrator` en su campo `skills`
- **THEN** el registro de contexto del subagente no contiene el testigo del coordinador (la skill no se puede precargar) y la transcripción no registra trabajo de coordinación hecho por ese subagente

### Requirement: Continuidad del coordinador en la sesión
El rol del coordinador MUST seguir disponible después de la compactación de la conversación y al
reanudar la sesión, o el sistema MUST indicar cómo recuperarlo. Actor: usuario. Disparador:
compactación o reanudación. Resultado: testigo del coordinador presente o instrucción de
reinvocar.

#### Scenario: El SKILL.md entra en el presupuesto de reenganche
- **WHEN** se compone el `SKILL.md` del coordinador con los cinco toggles activos
- **THEN** su tamaño queda bajo el umbral de 4000 tokens, verificable sin modelo (criterio de tamaño; no afirma que Claude Code reenganche el cuerpo tras compactar)

#### Scenario: Compactación
- **WHEN** se invocó `/sdd-orchestrator` y la sesión se compacta (`/compact`)
- **THEN** el cuerpo del `SKILL.md` no se reengancha (medido con Claude Code 2.1.318), y el sistema indica cómo recuperarlo: el bloque universal, que sí está en el contexto posterior a la compactación, contiene la instrucción de volver a invocar `/sdd-orchestrator` antes de seguir con trabajo OPSX

#### Scenario: Reanudación
- **WHEN** se invocó `/sdd-orchestrator` en una sesión y se reanuda con `claude --continue` o `--resume`
- **THEN** el registro de contexto de la sesión reanudada contiene el testigo del coordinador; si la tarea 1.14 muestra que no, el bloque universal indica volver a invocar `/sdd-orchestrator` y README lo documenta

### Requirement: Módulos por acción según la acción OPSX
El coordinador MUST cargar sólo los módulos que la tabla acción → módulos de `SKILL.md` asigna a
la acción solicitada y, cuando delega, MUST pasar su contenido en el encargo del especialista.
Actor: coordinador. Disparador: pedido de una acción (explore, propose, apply, verify, archive).
Resultado: testigo del módulo presente en el contexto del hilo o del especialista.

#### Scenario: propose recibe artifacts y no TDD
- **WHEN** con el coordinador invocado y el toggle `tdd` activo, el usuario pide propose
- **THEN** el contexto de quien ejecuta propose (hilo o especialista) contiene el testigo del módulo `artifacts` (control positivo) y no el de TDD

#### Scenario: Toggle desactivado omite el módulo
- **WHEN** el toggle `tdd` está inactivo y se instala el coordinador
- **THEN** `modules/` no contiene el módulo TDD y ni `SKILL.md` ni otro módulo contienen su testigo (verificable sin modelo)

#### Scenario: Módulo obligatorio ausente
- **WHEN** el archivo de un módulo obligatorio para la acción pedida falta o está vacío
- **THEN** el coordinador no ejecuta esa acción, nombra el módulo y la ruta esperada, indica el comando que lo reinstala, y sigue con las acciones pedidas que no dependen de ese módulo

#### Scenario: Control obligatorio ausente informado por 08
- **WHEN** la acción depende de un control que 08 declara obligatorio y no disponible en la superficie
- **THEN** la respuesta del coordinador en el hilo contiene el mensaje que define 08 para esa superficie y la acción afectada no se ejecuta; las demás siguen

### Requirement: No pisar skills del usuario
El instalador MUST NOT sobrescribir un directorio de skill `sdd-orchestrator` que no tenga la marca
de propiedad de Active Stack, en ningún target, y MUST advertir cuando una skill Machine con el
mismo ID tapa a la del proyecto. Actor: instalador. Disparador: install con colisión. Resultado:
error o advertencia claros y archivos del usuario intactos.

#### Scenario: Colisión con skill propia en Machine
- **WHEN** existe `<home>/.claude/skills/sdd-orchestrator/` sin marca de propiedad y se ejecuta install Machine
- **THEN** el install de ese harness falla para Claude con un mensaje que nombra la ruta y todos los archivos del directorio quedan idénticos

#### Scenario: Colisión con skill propia en Project
- **WHEN** existe `<root>/.claude/skills/sdd-orchestrator/` sin marca de propiedad y se ejecuta install Project
- **THEN** el install falla con un mensaje que nombra la ruta y los archivos quedan idénticos

#### Scenario: Skill Machine tapa a la de Project
- **WHEN** se instala con target Project y existe `<home>/.claude/skills/sdd-orchestrator/` (propia o de Active Stack)
- **THEN** la salida advierte que la skill personal prevalece sobre la del proyecto y nombra ambas rutas; en una sesión abierta en `<root>`, el registro de contexto tras `/sdd-orchestrator` contiene el testigo de la skill que la tarea 1.5 determine como prevalente

#### Scenario: Skills ajenas con otros nombres
- **WHEN** el directorio de skills del target contiene otras skills del usuario
- **THEN** install, update y uninstall no las modifican

### Requirement: Compatibilidad por versión y entorno sin Claude
El instalador MUST comprobar `claude --version` contra una versión mínima declarada antes de
escribir la entrega por capas. Si la versión es menor o no se puede parsear, MUST aplicar el modo
legado e informarlo. Si el binario falta, MUST actuar según G11 (recomendación: entrega por capas
con advertencia). Actor: instalador. Disparador: install o update. Resultado: entrega y estado de
verify definidos para cada caso.

#### Scenario: Versión insuficiente con legado previo
- **WHEN** `claude --version` informa una versión menor que la mínima y `CLAUDE.md` ya tiene el bloque legado
- **THEN** el instalador no escribe la skill, no purga el bloque legado, advierte con la versión mínima, y verify reporta "modo legado por versión" (advertencia, no fallo)

#### Scenario: Versión insuficiente o desconocida sin entrega previa
- **WHEN** la versión es menor que la mínima o no se puede parsear, y no hay entrega previa
- **THEN** el instalador escribe el bloque legado (composición legado: misma salida byte a byte que la composición de 35a89f8 para los mismos toggles, dentro de `active-stack:sdd-orchestrator`), advierte que el coordinador queda en `CLAUDE.md` por compatibilidad, y verify reporta "modo legado por versión"

#### Scenario: Update en modo legado refresca el bloque legado
- **WHEN** la versión sigue siendo insuficiente, `CLAUDE.md` tiene el bloque legado y un update trae assets con texto nuevo
- **THEN** el bloque legado se reescribe con la composición legado de los assets nuevos, no se crea la skill, el texto del usuario queda idéntico, y una segunda ejecución no cambia nada

#### Scenario: Binario claude ausente (CI, Docker de e2e, instalar antes que Claude)
- **WHEN** `claude` no está en el PATH
- **THEN** el instalador escribe la entrega por capas (bloque universal y skill), advierte que no pudo comprobar la versión, y verify reporta "versión no comprobada" (advertencia, no fallo)

#### Scenario: Verify informa la vía activa
- **WHEN** se ejecuta el verify posterior a un install para Claude
- **THEN** la salida nombra si el coordinador está entregado como skill o en modo legado por versión

### Requirement: Pruebas registradas por versión, interfaz y plataforma
Cada prueba de comportamiento de la skill MUST registrar cliente, versión, interfaz (terminal o
IDE), plataforma y permisos efectivos en el manifiesto de 02. Actor: quien prueba. Disparador:
corrida de prueba. Resultado: combinación probada o no disponible con motivo.

#### Scenario: Invocación en terminal y en IDE
- **WHEN** se repite el escenario "La invocación manual activa el coordinador en el hilo" en la terminal y en la extensión de IDE, en Windows y en Linux
- **THEN** cada combinación queda registrada en el manifiesto de 02 como probada con su resultado; una combinación donde el registro de contexto no está disponible queda "no disponible" con motivo, no como éxito
