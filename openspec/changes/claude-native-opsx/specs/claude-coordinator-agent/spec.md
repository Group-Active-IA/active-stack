# Spec Delta

## Purpose

Registra el coordinador OPSX como agente nativo de Claude Code, invocable sólo de forma explícita,
que carga sus módulos según la acción y respeta la versión y el target de instalación.

## ADDED Requirements

### Requirement: Coordinador registrado como agente nativo (depende de G1 y G9, recomendación)
Al instalar `sdd-orchestrator` para Claude en una versión compatible, el instalador MUST crear un
archivo de agente administrado, con marca de propiedad de Active Stack, en el directorio de
agentes del target. Actor: instalador. Disparador: install. Resultado: archivo presente.

#### Scenario: Install Machine crea el agente
- **WHEN** se instala `sdd-orchestrator` para Claude con target Machine
- **THEN** existe `<home>/.claude/agents/<id-coordinador>.md` con marca de propiedad de Active Stack y la definición del rol coordinador

#### Scenario: Install Project crea el agente en el proyecto
- **WHEN** se instala con target Project y raíz `<root>`
- **THEN** existe `<root>/.claude/agents/<id-coordinador>.md` y no se escribe nada en `<home>/.claude/agents/`

### Requirement: Invocación explícita activa el coordinador (depende de G2, recomendación)
El coordinador MUST activarse sólo cuando el usuario lo invoca de forma explícita y el instalador
MUST NOT fijar el setting `agent`. Actor: usuario. Disparador: `claude --agent <id-coordinador>`.
Resultado: registro de contexto con testigo del coordinador.

#### Scenario: --agent activa el coordinador
- **WHEN** el usuario ejecuta `claude -p --agent <id-coordinador>`
- **THEN** el registro de contexto contiene el testigo del coordinador y el testigo universal

#### Scenario: El instalador no fija el agente por defecto
- **WHEN** termina un install o update para Claude
- **THEN** `settings.json` del target no tiene la clave `agent` escrita por Active Stack y su valor previo, si existía, queda igual

### Requirement: La sesión normal no lanza el coordinador como subagente (depende de G9, recomendación)
Una sesión normal MUST NOT ejecutar el coordinador como subagente sin que el usuario lo pida por
nombre, y si el coordinador corre como subagente MUST negarse a coordinar e indicar
`claude --agent <id-coordinador>`. Actor: sesión normal. Disparador: pedido OPSX sin nombrar el
coordinador. Resultado: transcripción sin subagente coordinador.

#### Scenario: El coordinador no figura entre los subagentes disponibles (si existe exclusión por frontmatter)
- **WHEN** la tarea 1.10 encontró un campo de frontmatter que excluye al agente de la delegación, el instalador lo escribe, y se abre una sesión sin `--agent`
- **THEN** el registro de la sesión (lista de tipos de subagente disponibles, según el oráculo validado en 1.10 y 1.11) no contiene `<id-coordinador>`; aserción determinista, una corrida alcanza

#### Scenario: Pedido OPSX en sesión normal sin nombrar el coordinador (sólo mitigaciones 2 y 3)
- **WHEN** no existe exclusión por frontmatter, y en sesiones sin `--agent` el usuario pide "proponé un change para X" sin nombrar al coordinador, repetido en N = 20 corridas independientes con el mismo prompt
- **THEN** ninguna de las 20 transcripciones registra un subagente del tipo `<id-coordinador>` (umbral: 0 de 20); con una o más apariciones el requisito se considera incumplido y se aplica el escenario "Exclusión no disponible". N y el umbral son propuesta del plan y los confirma el grupo junto con G9

#### Scenario: Coordinador lanzado como subagente
- **WHEN** el usuario pide explícitamente a una sesión normal lanzar `<id-coordinador>` como subagente
- **THEN** la respuesta del subagente indica que debe usarse `claude --agent <id-coordinador>` y la transcripción no registra delegaciones anidadas ni trabajo de apply o propose hecho por ese subagente

#### Scenario: Exclusión no disponible
- **WHEN** la tarea 1.10 demuestra que un agente de archivo no puede excluirse de la lista de subagentes ni de la delegación automática
- **THEN** la vía de archivo en el directorio descubrible no se construye hasta que el grupo decida G9, y este requisito se reescribe según la opción elegida

### Requirement: Módulos por acción según la acción OPSX (depende de G3, recomendación)
El coordinador MUST cargar o entregar al especialista sólo el módulo correspondiente a la acción
OPSX solicitada, según la matriz de reglas. Actor: coordinador. Disparador: pedido de una acción
(explore, propose, apply, verify, archive). Resultado: módulo presente en el prompt registrado.

#### Scenario: propose recibe governance y no TDD
- **WHEN** el usuario pide propose a un coordinador con `governance` y `tdd` activos
- **THEN** el prompt registrado del especialista de propose contiene el testigo de governance (control positivo) y no el de TDD

#### Scenario: Toggle desactivado omite el módulo
- **WHEN** el toggle `tdd` está inactivo y se instala el coordinador
- **THEN** el archivo del coordinador no contiene el testigo del módulo TDD (verificable sin modelo)

### Requirement: No pisar agentes del usuario
El instalador MUST NOT sobrescribir un archivo de agente con el mismo nombre que no tenga la marca
de propiedad de Active Stack, en ningún target. Actor: instalador. Disparador: install con
colisión de nombre. Resultado: error claro y archivo del usuario intacto.

#### Scenario: Colisión con agente propio en Machine
- **WHEN** existe `<home>/.claude/agents/<id-coordinador>.md` sin marca de propiedad y se ejecuta install Machine
- **THEN** el install de ese harness falla para Claude con un mensaje que nombra la ruta y el archivo queda idéntico

#### Scenario: Colisión con agente propio en Project
- **WHEN** existe `<root>/.claude/agents/<id-coordinador>.md` sin marca de propiedad y se ejecuta install Project
- **THEN** el install falla con un mensaje que nombra la ruta y el archivo queda idéntico

#### Scenario: Agentes ajenos con otros nombres
- **WHEN** el directorio de agentes del target contiene otros archivos del usuario
- **THEN** install, update y uninstall no los modifican

### Requirement: Compatibilidad por versión y entorno sin Claude
El instalador MUST comprobar `claude --version` contra una versión mínima declarada antes de
escribir la entrega por capas. Si la versión es menor o no se puede parsear, MUST aplicar el modo
legado e informarlo. Si el binario falta, MUST actuar según G11 (recomendación: entrega por capas
con advertencia). Actor: instalador. Disparador: install o update. Resultado: entrega y estado de
verify definidos para cada caso.

#### Scenario: Versión insuficiente con legado previo
- **WHEN** `claude --version` informa una versión menor que la mínima y `CLAUDE.md` ya tiene el bloque legado
- **THEN** el instalador no escribe el archivo de agente, no purga el bloque legado, advierte con la versión mínima, y verify reporta el check de Claude como "modo legado por versión" (advertencia, no fallo)

#### Scenario: Versión insuficiente o desconocida sin entrega previa
- **WHEN** la versión es menor que la mínima o no se puede parsear, y no hay entrega previa
- **THEN** el instalador escribe el bloque legado (composición legado: misma salida byte a byte que la composición de 97d7f6c para los mismos toggles, dentro de la sección `active-stack:sdd-orchestrator`), advierte que el coordinador queda en `CLAUDE.md` por compatibilidad, y verify reporta "modo legado por versión" (advertencia, no fallo)

#### Scenario: Update en modo legado refresca el bloque legado
- **WHEN** la versión sigue siendo insuficiente, `CLAUDE.md` tiene el bloque legado y un update trae assets con texto nuevo
- **THEN** el bloque legado se reescribe con la composición legado de los assets nuevos, no se crea archivo de agente, el texto del usuario queda idéntico, y una segunda ejecución no cambia nada

#### Scenario: Binario claude ausente (CI, Docker de e2e, instalar antes que Claude)
- **WHEN** `claude` no está en el PATH
- **THEN** el instalador escribe la entrega por capas (bloque universal y archivo de agente), advierte que no pudo comprobar la versión, y verify reporta "versión no comprobada" (advertencia, no fallo) (depende de G11, recomendación)

### Requirement: Fallback documentado si no hay agente nativo (depende de G2, recomendación)
El sistema MUST documentar en README una forma alternativa de activar el coordinador cuando
`--agent` no esté disponible; en este change el fallback se documenta y no se instala, y verify
nombra sólo vías instaladas. Actor: usuario. Disparador: lectura de README o de verify. Resultado:
vía activa nombrada.

#### Scenario: Verify informa la vía activa
- **WHEN** se ejecuta verify después de un install para Claude
- **THEN** la salida nombra si el coordinador está entregado como agente nativo o en modo legado por versión

#### Scenario: El fallback documentado funciona
- **WHEN** se ejecuta el comando de fallback documentado en README en el entorno de prueba
- **THEN** el registro de contexto de esa sesión contiene el testigo del coordinador
