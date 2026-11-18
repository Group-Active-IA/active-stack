# Harnesses profesionales para Active Stack, separados por agente

Fecha de investigación: **2 de noviembre de 2026**. Estado: **informe y recomendaciones; ninguna incorporación aprobada ni implementada**.

La oportunidad principal para Active Stack está en mejorar la coordinación, la verificación y el aislamiento del trabajo, además de sumar herramientas de contexto y diagnóstico. El catálogo ya tiene una base metodológica amplia; agregar más instrucciones de TDD o revisión produciría bastante duplicación. Conviene desarrollar un núcleo portátil de contratos y evidencias, con adaptaciones explícitas para las capacidades reales de cada agente.

## 1. Cómo interpretar la compatibilidad

Este informe separa tres cosas: **capacidad documentada del producto**, **integración existente en Active Stack** y **validación pendiente en un entorno real**. Consultar documentación permite proponer una integración; no demuestra que el instalador ya la implemente o que funcione en cada sistema operativo.

Se investigaron fuentes oficiales y repositorios de sus autores. Las propuestas propias se identifican como tales. No se instalaron herramientas, modificaron configuraciones personales ni ejecutaron builds o pruebas de integración. Las versiones, planes comerciales y políticas de organización pueden condicionar la disponibilidad.

| Agente/producto concreto | Skills | MCP | Hooks o equivalente | Delegación | Active Stack actual |
|---|---|---|---|---|---|
| Claude Code | Nativas | Nativo | Hooks nativos | Subagentes; Teams experimental | Adapter implementado |
| Codex local/app | Nativas | Nativo | Hooks nativos actuales | Subagentes configurables | Adapter implementado |
| OpenCode | Nativas | Nativo | Plugins JS/TS con eventos | Agentes primarios y subagentes | Adapter implementado |
| Antigravity | Nativas | Nativo | Nativos; contrato propio | Custom subagents App/CLI; no extrapolar al IDE | Adapter implementado; auditar superficies |
| Gemini CLI | Nativas | Nativo | Hooks nativos | Subagentes; comprobar versión | Adapter pendiente |
| Cursor | Nativas | Nativo | Nativos; diferencias local/cloud | Subagentes locales y cloud | Adapter pendiente |
| VS Code + GitHub Copilot | Nativas según harness | Nativo según harness | Preview; configuración por harness | Subagentes; herramientas por harness | Adapter pendiente |
| Windsurf/Cascade → Devin Local | Skills y reglas migrables | Configuración Devin actual | Hooks del nuevo harness | No asumir equivalencia con Cascade | Adapter pendiente y migración de producto |

Las fuentes y limitaciones de cada fila se desarrollan en las secciones por agente. Un MCP necesita además transporte, autenticación, runtime y permisos compatibles; la palabra «MCP» por sí sola no certifica todas sus funciones.

### Situación real del repositorio

`internal/agents/factory.go` registra únicamente **Claude, Codex, OpenCode y Antigravity**. Los otros cuatro aparecen en el enum de `internal/model/harness.go`, pero el constructor devuelve un error para ellos. Por tanto, ampliar compatibilidad es trabajo del instalador antes de prometer los ocho agentes.

El catálogo `internal/catalog/harnesses.yaml` ya contiene OpenSpec, Engram, Context7, Playwright MCP, diseño, TDD, debugging sistemático, revisión de código y numerosas skills de stack. Se consideran capacidades existentes, aunque todavía puede mejorarse cómo se aplican y verifican.

El modelo real admite `skill`, `config`, `external` y `command`. Hooks, perfiles de agentes y políticas podrían materializarse mediante `config` con adapters; un servidor MCP mediante `external`; una técnica reusable mediante `skill`. No hace falta inventar un tipo de harness por cada característica.

## 2. Claude Code

Es un buen primer destino para contratos de delegación, hooks de calidad y perfiles especializados. Sus skills, MCP, hooks y subagentes están documentados como mecanismos distintos: la skill enseña un procedimiento; el MCP aporta herramientas; el hook reacciona a un evento; el subagente separa una tarea y su contexto. Mantener esa separación facilita diagnosticar fallas. Fuentes: [skills](https://code.claude.com/docs/en/skills), [MCP](https://code.claude.com/docs/en/mcp), [hooks](https://code.claude.com/docs/en/hooks) y [subagentes](https://code.claude.com/docs/en/sub-agents).

**Harnesses recomendados:** `agent-contract`, `verification-evidence`, `quality-hooks`, `isolated-worktrees`, `serena` y `github-mcp`. Para equipos de diseño, `figma-mcp`; para diagnóstico web, `chrome-devtools`.

Los subagentes deberían tener objetivos y permisos concretos: explorador, implementador y revisor. La revisión independiente debe comprobar requisitos y evidencia, además de leer el resultado que entrega el implementador. La selección de herramientas y los permisos del host deben acompañar al prompt. Fuente: [permisos de Claude Code](https://code.claude.com/docs/en/permissions).

**Agent Teams** permite coordinación de compañeros, pero es experimental y requiere habilitación explícita. No conviene convertirlo en dependencia de Lite. También hay que distinguir un subagente con `isolation: worktree` de equipos que trabajan en el checkout principal; la concurrencia no garantiza separación de archivos. Fuente: [Agent Teams](https://code.claude.com/docs/en/agent-teams).

Propuesta de uso: primero delegación jerárquica con propietario único por tarea; Teams como opción avanzada cuando varias tareas necesiten intercambiar información durante la ejecución. Active Stack debería probar los hooks por evento y versión, con timeout y recuperación, antes de distribuirlos.

## 3. Codex

La documentación actual describe hooks de ciclo de vida; sería incorrecto limitar su soporte al antiguo mecanismo `notify`. Se pueden aprovechar eventos de herramientas, sesión, compactación y subagentes. Hay diferencias de comportamiento entre hooks de comando y MCP: los errores de un hook MCP no constituyen por sí solos una barrera que bloquee la acción. Una política importante debe apoyarse en permisos y verificaciones adicionales. Fuente: [hooks oficiales](https://learn.chatgpt.com/docs/hooks).

Los agentes personalizados se pueden definir mediante archivos TOML bajo `.codex/agents/`, con nombre, descripción e instrucciones. Heredan restricciones del padre: escribir «solo lectura» en el prompt no crea un sandbox de solo lectura. Esta diferencia es esencial para diseñar perfiles de auditoría. Fuente: [subagentes oficiales](https://learn.chatgpt.com/docs/agent-configuration/subagents).

**Harnesses recomendados:** `agent-contract`, `context-budget`, `verification-evidence`, `quality-hooks`, `harness-evals`, `serena` y `openai-docs-mcp`. Este último aporta documentación oficial mediante `https://developers.openai.com/mcp`, complementando Context7 en proyectos que usan OpenAI. Fuentes: [construcción de skills](https://learn.chatgpt.com/docs/build-skills) y [Docs MCP](https://developers.openai.com/learn/docs-mcp).

Para escritores concurrentes conviene usar los worktrees gestionados por el producto cuando estén disponibles y respetar su ciclo de vida. Las configuraciones de Codex local, la aplicación y ejecución remota deben validarse por separado; no hay que extrapolar acceso a binarios, archivos o credenciales de una superficie a otra. Fuente: [worktrees](https://learn.chatgpt.com/docs/environments/git-worktrees).

Como origen de ejemplos, el repositorio `openai/skills` actualmente declara su deprecación y remite a **OpenAI Plugins**. No debería incorporarse como una nueva dependencia mantenida. Los ejemplos actuales combinan skills, MCP y otros recursos según el plugin. Fuentes: [repositorio anterior](https://github.com/openai/skills) y [plugins actuales](https://github.com/openai/plugins).

## 4. OpenCode

Tiene skills, agentes configurables y MCP. La automatización por eventos se implementa mediante plugins JavaScript/TypeScript; copiar un JSON de hooks de Claude o Codex no produce una integración equivalente. Fuentes: [skills](https://opencode.ai/docs/skills/), [agentes](https://opencode.ai/docs/agents/), [MCP](https://opencode.ai/docs/mcp-servers/) y [plugins](https://opencode.ai/docs/plugins/).

**Harnesses recomendados:** `agent-contract`, `verification-evidence`, `context-budget`, `quality-hooks` adaptado como plugin, `serena`, `chrome-devtools` y `github-mcp`.

El stack ya conoce una diferencia importante: el orquestador debe registrarse como agente primario dedicado, evitando que sus reglas de coordinación contaminen todos los modos. Ese mismo criterio debería aplicarse a exploradores y revisores: declarar el rol donde OpenCode lo ejecuta, además del contenido del prompt.

Hay documentación V1 y V2 con cambios en términos de herramientas y permisos. Por ejemplo, aparecen diferencias entre `permission`/`permissions`, `bash`/`shell` y `task`/`subagent`. Un adapter profesional necesita detectar la versión y comprobar el esquema efectivo; una transformación textual fija puede dejar una configuración plausible pero inoperante. Fuentes: [permisos V1](https://opencode.ai/docs/permissions/) y [permisos V2](https://dev.opencode.ai/v2/docs/permissions/).

La comunicación debería comenzar por delegación y devolución al primario. Un bus externo solo se justifica para procesos independientes que deban coordinarse fuera de esa sesión. Los plugins requieren pruebas específicas de errores, cancelación y reintentos; no basta con comprobar que el archivo exista.

## 5. Antigravity

La documentación actual presenta skills, hooks y MCP en varias superficies, mientras que los **custom subagents** se documentan para App y CLI. Estos últimos pueden coordinar compañeros mediante identificadores y elegir modalidades de workspace como branch, inherit o share. No se debe prometer esa misma función al IDE por compartir marca. Fuentes: [skills](https://antigravity.google/docs/skills), [hooks](https://antigravity.google/docs/hooks), [MCP](https://antigravity.google/docs/mcp) y [custom subagents](https://antigravity.google/docs/subagents/).

**Harnesses recomendados:** `agent-contract`, `verification-evidence`, `context-budget`, `quality-hooks` con adapter por superficie y `chrome-devtools`. La coordinación entre compañeros necesita asignar rutas y decidir cuándo compartir workspace; los IDs de agentes no evitan conflictos de escritura.

Hay una migración inmediata que atender: la página de workflows del IDE anuncia que los workflows anteriores dejarían de funcionar el **19 de octubre de 2026**, y orienta a migrarlos a skills. Por eso no conviene sumar un harness nuevo basado en workflows antiguos sin identificar la versión y reemplazo efectivo. Fuente: [workflows y migración](https://antigravity.google/docs/ide/workflows/).

El adapter existente resuelve rutas concretas de instrucciones, skills y MCP, pero esta investigación no validó esas rutas contra cada superficie instalada. La recomendación es auditar primero App/CLI/IDE y permisos, conservando compatibilidad solo donde pueda demostrarse. Fuente de producto: [permisos y sandbox](https://antigravity.google/docs/permissions).

## 6. Gemini CLI

Gemini CLI documenta skills, MCP, hooks y subagentes. No debe confundirse con Gemini Code Assist o Antigravity: comparten ecosistema, pero no necesariamente archivos, flags o funciones. Fuentes: [skills](https://geminicli.com/docs/cli/skills/), [MCP](https://geminicli.com/docs/tools/mcp-server/), [hooks](https://geminicli.com/docs/hooks/reference/) y [subagentes](https://geminicli.com/docs/core/subagents/).

**Primera incorporación necesaria:** un adapter de Gemini CLI. Después, `agent-contract`, `verification-evidence`, `quality-hooks`, `serena`, `github-mcp` y `chrome-devtools` serían buenos candidatos.

Existe documentación de subagentes habilitados por defecto y referencias de extensiones que todavía los presentan como preview. El instalador debe detectar versión y capacidades, ofreciendo degradación explícita a ejecución secuencial cuando corresponda. Las extensiones permiten agrupar recursos, pero no deberían llevar un segundo orquestador que compita con OPSX.

Los worktrees se documentan como capacidad experimental mediante configuración y `--worktree`; la preservación del trabajo y el alcance por sesión deben respetarse. No se debe asumir que cada subagente recibe automáticamente uno. Fuente: [git worktrees](https://geminicli.com/docs/cli/git-worktrees/).

Es un destino razonable de expansión porque posee mecanismos cercanos a los otros CLI. Sigue pendiente validar instalación, hooks, MCP y desinstalación en Windows nativo y WSL como entornos diferentes.

## 7. Cursor

Cursor dispone de skills y reglas, MCP, hooks y subagentes. Las reglas deberían contener convenciones pequeñas y estables; las skills, procedimientos cargados cuando se necesitan. Los subagentes aportan contexto separado y permiten especialización y tareas de fondo. Fuentes: [skills](https://cursor.com/docs/skills), [reglas](https://cursor.com/docs/rules), [MCP](https://cursor.com/docs/mcp) y [subagentes](https://cursor.com/docs/subagents).

**Harnesses recomendados:** un adapter propio, seguido por `agent-contract`, `context-budget`, `verification-evidence`, `quality-hooks`, `figma-mcp` y `chrome-devtools`.

Los hooks tienen diferencias concretas entre local y cloud. En cloud se cargan hooks de comando del repositorio y configuraciones administradas compatibles, pero no los hooks personales del home local. Para MCP HTTP ejecutado en el backend, los hooks reciben información limitada: pueden denegar llamadas en los casos documentados, pero no reescribir argumentos y resultados como si fueran herramientas locales. Fuente: [hooks](https://cursor.com/docs/hooks).

Los subagentes cloud trabajan con otra VM y rama; eso debe tratarse como un entorno distinto para setup, secretos y verificación. El perfil del instalador debería indicar editor, CLI o cloud y comprobar la activación de sus recursos. No basta con detectar una carpeta `.cursor`.

## 8. VS Code + GitHub Copilot

VS Code ahora documenta varios harnesses: Copilot, Local, Claude y Codex, además del destino cloud. Seleccionar un modelo Claude dentro de Copilot no cambia al harness Claude. Herramientas, hooks, políticas y lifecycle dependen del harness seleccionado. Por eso `vscode` necesita representar una integración concreta, no todos los agentes que pueden abrirse en el editor. Fuente: [comparación oficial de harnesses](https://code.visualstudio.com/docs/agents/run/agent-harnesses).

Las skills y los subagentes están documentados, pero la herramienta de delegación cambia: Copilot utiliza `task`, mientras que Local utiliza `runSubagent`. Los hooks de VS Code se presentan como **Preview** y su compatibilidad debe comprobarse por harness; no se debe copiar un hook de Claude suponiendo semántica idéntica. Fuentes: [skills](https://code.visualstudio.com/docs/agent-customization/agent-skills), [subagentes](https://code.visualstudio.com/docs/agents/run/subagents) y [hooks](https://code.visualstudio.com/docs/agent-customization/hooks).

**Harnesses recomendados:** adapter Copilot claramente identificado, `agent-contract`, `verification-evidence`, `context-budget` y perfiles de MCP pequeños. Figma y GitHub son útiles para flujos de diseño y revisión; evitar duplicar herramientas que el harness ya ofrezca.

Para aislamiento, evaluar worktrees y Dev Containers según el destino. La política de una organización puede limitar harnesses y herramientas; el instalador debe respetarla y describir la degradación real en vez de anunciar una instalación completa.

## 9. Windsurf / Cascade

Esta fila necesita una actualización de producto: la documentación consultada de Windsurf redirige a **Devin Desktop/Devin Local**. Cascade figura como legacy y la página declara que ya no se inician nuevas conversaciones ni se abren las anteriores. No sería responsable diseñar una nueva integración usando únicamente guías históricas de Cascade. Fuente: [Cascade legacy](https://docs.devin.ai/desktop/cascade).

Devin Local comparte el harness del CLI y configura sus recursos bajo `.devin`; permite incorporar skills heredadas de ubicaciones anteriores. Esa compatibilidad de skills no demuestra que MCP, hooks, reglas y permisos antiguos sean intercambiables. Fuente: [customizaciones actuales de Devin Local](https://docs.devin.ai/desktop/devin-local/customizations).

**Recomendación:** tratar este destino como una migración de adapter con identificación de producto y versión. Empezar por skills de `verification-evidence`, `context-budget` y contrato de tareas; sumar MCP y hooks cuando se haya comprobado el esquema del runtime objetivo. Mantener separado el soporte legacy que todavía requieran usuarios con instalaciones antiguas.

Actualmente no existe adapter registrado de Windsurf en Active Stack. El nombre final del target y su política de compatibilidad quedan **(TBD)** hasta definir qué versiones se quieren soportar.

## 10. Herramientas externas que vale la pena sumar

Estas propuestas son opt-in. «Compatible por MCP» significa candidato técnicamente razonable, sujeto a transporte, permisos y autenticación; no equivale a una prueba realizada por Active Stack.

| ID propuesto | Tipo | Qué aporta | Compatibilidad y límite principal |
|---|---|---|---|
| `serena` | `external` + configuración | Navegación y edición por símbolos, referencias y estructura | Clientes MCP; backend LSP y dependencias según lenguaje |
| `chrome-devtools` | `external` | Traces de performance, red y diagnóstico en Chrome | Guías para los ocho agentes; requiere Node LTS y Chrome |
| `github-mcp` | `external` | Issues, PR, repositorios y diagnóstico de Actions | Cliente MCP; credenciales y toolsets acotados |
| `figma-mcp` | `external` + skills | Contexto del diseño y workflows oficiales de Figma | Clientes autorizados por Figma; no universal por protocolo |
| `semgrep` | `external` + skill | Análisis estático y exploración de hallazgos | CLI/CI; MCP mediante binario oficial actual |
| `sentry-mcp` | `external` | Errores y performance observados en ejecución | HTTP MCP con OAuth y acceso a organización/proyecto |
| `openai-docs-mcp` | `external` | Documentación oficial para desarrollar con OpenAI | Cliente HTTP MCP; complementar documentación general |
| `promptfoo` | `external` opcional | Evaluaciones repetibles de prompts y agentes | Integración por proveedor o wrapper; consumo según ejecución |

**Serena** es especialmente útil en repositorios grandes o refactors: permite recuperar símbolos y referencias sin leer archivos completos. Su backend gratuito usa LSP; otro backend depende de JetBrains. Conviene desactivar funciones redundantes de memoria y shell cuando Engram y el agente ya las cubran. La instalación actual documenta `serena-agent` mediante `uv`, por lo que no debe copiarse una receta antigua de marketplace. Fuentes: [Serena](https://github.com/oraios/serena) y [conexión de clientes](https://oraios.github.io/serena/02-usage/030_clients.html).

**Chrome DevTools** complementa Playwright: interesa para explicar por qué una interfaz tarda, inspeccionar requests o registrar trazas. No hace falta activar dos navegadores para tareas sencillas. El proyecto mantiene configuraciones para los ocho agentes y consideraciones de arranque en Windows; eso todavía requiere validación local. Fuentes: [proyecto oficial](https://github.com/ChromeDevTools/chrome-devtools-mcp) y [configuraciones de clientes](https://github.com/ChromeDevTools/chrome-devtools-mcp/blob/main/docs/client-configurations.md).

**GitHub MCP** debería ofrecer perfiles pequeños, por ejemplo repositorios, issues y PR en solo lectura. El servidor permite seleccionar toolsets y herramientas; su modo read-only excluye escrituras aunque se soliciten explícitamente. Esto reduce superficie y contexto. Es opcional si `gh` ya resuelve el caso con menor complejidad. Fuente: [GitHub MCP oficial](https://github.com/github/github-mcp-server).

**Figma MCP** merece un starter de diseño, no inclusión obligatoria. La documentación presenta instalación para Claude, Codex, Cursor, Gemini CLI y VS Code, pero restringe conexiones a clientes de su catálogo. Para OpenCode, Windsurf y Antigravity la disponibilidad debe confirmarse; no se deduce de tener MCP genérico. Fuente: [instalación remota oficial](https://developers.figma.com/docs/figma-mcp-server/remote-server-installation/).

**Semgrep** debería integrarse primero como verificación determinista; el agente puede ayudar a interpretar y resolver los resultados. El repositorio MCP independiente está deprecado: el desarrollo se trasladó al binario principal. No proponer instalar `semgrep-mcp` antiguo. Fuentes: [aviso de deprecación](https://github.com/semgrep/mcp) y [implementación actual](https://github.com/semgrep/semgrep/tree/develop/cli/src/semgrep/mcp).

**Sentry MCP** conecta diagnóstico con evidencia de ejecución. Usa OAuth y permite acotar la conexión a un proyecto, ocultando herramientas de descubrimiento innecesarias. Su valor depende de que el proyecto ya utilice Sentry; no justifica imponer el servicio. Fuente: [Sentry MCP](https://mcp.sentry.dev/).

## 11. Técnicas profesionales que deberían convertirse en harnesses

### Contratos, contexto y evidencia

`agent-contract` sería una configuración propia, apoyada por una skill de delegación. Toda tarea debe contener objetivo verificable, archivos permitidos, dependencias, restricciones, presupuesto y formato de resultado. El coordinador recibe decisiones y evidencia, evitando llenar su contexto con exploración irrelevante. Debe existir un propietario claro de cada cambio y una vía para resolver bloqueos.

`context-budget` introduciría índices pequeños, lectura bajo demanda, resúmenes con referencias y memoria selectiva. Complementa Engram; no crea otra memoria competidora. La carga progresiva permite reservar contexto para el problema activo. La investigación de Anthropic aborda selección, compactación y delegación como partes de la ingeniería de contexto. Fuente: [context engineering](https://www.anthropic.com/research/effective-context-engineering-for-ai-agents).

`verification-evidence` exige verificar el diff real, ejecutar controles pertinentes y asociar cada requisito con evidencia. La declaración del implementador no basta. Se inspira en [verification-before-completion](https://raw.githubusercontent.com/obra/superpowers/main/skills/verification-before-completion/SKILL.md), adaptada para respetar la prohibición local de builds y commits no solicitados.

### Hooks con alcance limitado

`quality-hooks` debería emitir implementaciones distintas por agente y reutilizar scripts portátiles cuando sea posible. Ejemplos: después de editar, formatear solo archivos propios modificados; antes de una operación relevante, validar el alcance autorizado; al finalizar, informar qué comprobaciones siguen pendientes. Debe definir timeout, código de salida, protección contra recursión y tratamiento de fallas.

Un hook de recomendación puede fallar sin bloquear el trabajo; uno anunciado como bloqueo debe comprobarse con una prueba negativa. Las comprobaciones decisivas también deben poder ejecutarse mediante CLI y CI. Esta propuesta no añade revisión automática de commits ni GGA, excluidos por la arquitectura del proyecto.

### Aislamiento y coordinación de escritores

`isolated-worktrees` separaría ramas y directorios para tareas de escritura concurrentes. El coordinador asigna contratos de API, orden de integración y propietario de archivos compartidos. Los worktrees no separan credenciales, procesos ni permisos del sistema; para ello se necesitan mecanismos reales del host, contenedores o entornos restringidos.

La skill existente de Superpowers aporta una referencia, pero incluye pasos de setup y commit que chocan con reglas de este proyecto. Se recomienda adaptar la técnica a las preferencias del usuario y herramientas nativas disponibles, sin importar el procedimiento completo. Fuente: [using-git-worktrees](https://raw.githubusercontent.com/obra/superpowers/main/skills/using-git-worktrees/SKILL.md).

### Evaluaciones y pruebas más difíciles de engañar

`harness-evals` mediría si un cambio de prompt, skill o hook mejora tareas reales. El laboratorio necesita repositorios restaurables, casos de éxito y fallo, presupuesto explícito y evaluadores independientes. Métricas: cumplimiento de la tarea, regresiones, cambios fuera de alcance, tiempo, consumo y resistencia a instrucciones hostiles incluidas en archivos. La evaluación debe distinguir calidad del resultado y trayectoria de ejecución. Fuente: [evaluaciones de agentes](https://www.anthropic.com/research/demystifying-evals-for-ai-agents).

Promptfoo es un posible motor, no una dependencia obligatoria del método. Permite comparar prompts y modelos con casos y métricas; ejecutar localmente no evita consumo o envío al proveedor configurado. La compatibilidad con cada agente requiere revisar su proveedor o construir un wrapper. Fuente: [Promptfoo](https://www.promptfoo.dev/docs/intro/).

Para ampliar TDD, sumar selectivamente skills de **property-based testing**, fuzzing y mutation testing. Sirven para descubrir invariantes, entradas inesperadas y tests que pasan aunque el comportamiento esté roto. Trail of Bits mantiene skills especializadas; deben elegirse según lenguaje, licencia y herramientas requeridas. Las instalaciones públicas sugieren adopción, no certifican calidad ni portabilidad. Fuente: [skills de Trail of Bits](https://github.com/trailofbits/skills).

### Herramientas mínimas y observabilidad

No conviene habilitar todos los MCP en cada sesión. Un perfil backend, diseño o diagnóstico debería exponer solo sus herramientas necesarias. Filtrar y paginar resultados, procesar datos estructurados mediante código cuando corresponda y evitar reenviar grandes salidas ayuda a controlar costo y errores. Fuente: [ejecución de código con MCP](https://www.anthropic.com/research/code-execution-with-mcp).

Propuesta propia: un registro local de tareas con identificador, agente/modelo, duración, comprobaciones y motivo de bloqueo, sin prompts completos ni secretos. El presupuesto debe decidir cuándo ampliar investigación, repetir pruebas o escalar modelo. No hace falta imponer un servicio de telemetría.

## 12. Comunicación entre agentes: qué agregar y qué posponer

Para subagentes dentro de una sesión, usar primero las herramientas nativas del coordinador. Para procesos independientes, una bandeja compartida puede resolver mensajes y reservas de archivos.

**MCP Agent Mail Rust** ofrece identidades, threads, mensajes y leases de archivos. Las reservas son advisory: expresan intención y no equivalen a aislamiento del filesystem. Su instalación oficial se orienta a Linux/macOS; Windows nativo y Termux quedan pendientes de validación. La licencia incorpora un rider que debe revisarse antes de distribuirlo. El proyecto Python anterior está deprecado. Fuentes: [Rust actual](https://github.com/Dicklesworthstone/mcp_agent_mail_rust) y [Python anterior](https://github.com/Dicklesworthstone/mcp_agent_mail).

**A2A** sirve para interoperabilidad entre servicios de agentes mediante descubrimiento y tareas. No conecta automáticamente Claude Code, Codex u otros CLI, ni es un reemplazo de MCP: harían falta adapters y servicios que hablen el protocolo. Es razonable como investigación futura si aparecen agentes remotos propios. Fuente: [A2A oficial](https://a2a-protocol.org/latest/).

### Prompt portable propuesto para delegar

```text
TAREA: <id y objetivo concreto>
BASE: <commit/revisión y directorio o worktree>
ROL: <explorar | implementar | revisar>
ACEPTACIÓN: <resultados observables>
CONTEXTO: <archivos y decisiones pertinentes>
ALCANCE DE ESCRITURA: <rutas permitidas; otras requieren coordinación>
CONTRATOS: <API/tipos compartidos; responsable y dependencias>
RESTRICCIONES: <reglas de repo, permisos y acciones autorizadas>
PRESUPUESTO: <tiempo, herramientas o costo; cuándo informar límite>

Antes de editar, verificá la base y cambios ajenos. No sobrescribas trabajo
de otros. Ante un conflicto, informá archivo, decisión necesaria e impacto.
Usá las herramientas nativas de coordinación. No inventes resultados.
No publiques, hagas commit ni ejecutes builds sin autorización aplicable.

ENTREGA:
estado: complete | blocked | needs-review
resumen: <resultado y decisiones>
archivos: <rutas modificadas o hallazgos con referencias>
evidencia: <comando, salida relevante y exit code; o no ejecutado y motivo>
pendientes: <riesgos, dependencias y preguntas concretas>
```

El revisor recibe la aceptación, base y diff; comprueba las afirmaciones sin asumir que el implementador terminó correctamente. Este contrato mejora coordinación, pero sus límites de escritura solo son efectivos técnicamente si el host también los impone.

## 13. Prioridad de incorporación y trabajo necesario

| Prioridad | Incorporación | Razón | Condición antes de publicar |
|---|---|---|---|
| P0 | `agent-contract`, `verification-evidence` | Mejora transversal con poca dependencia externa | Evaluaciones de tareas y respeto de reglas locales |
| P0 | Matriz de capacidades y versión por adapter | Evita prometer soporte por un enum | Probes, fixtures y errores explícitos |
| P1 | `quality-hooks`, `context-budget` | Convierte convenciones en workflows verificables | Contrato de eventos por producto y pruebas negativas |
| P1 | `isolated-worktrees` | Reduce conflictos de escritores | Preservar trabajo, rama y lifecycle nativo |
| P1 | `serena`, `github-mcp`, `chrome-devtools` | Contexto y diagnóstico de uso amplio | Dependencias, permisos y prueba real por OS |
| P1 | `harness-evals` | Permite medir cambios del propio stack | Casos restaurables y presupuesto acotado |
| P2 | Adapters Gemini, Cursor y Copilot; auditoría Windsurf → Devin | Amplía compatibilidad real sin fijar un target legacy | Identificar producto y versión; instalación, actualización y remoción idempotentes |
| P2 | Figma, Sentry, skills avanzadas de testing | Alto valor según proyecto | Opt-in, licencia, autenticación y stack apropiado |
| P3 | Agent Teams, Agent Mail y A2A | Valor en coordinación compleja | Necesidad demostrada y soporte de entorno |

Hay dos gaps técnicos concretos antes de ampliar los MCP remotos. `internal/model/starter.go` representa entradas MCP mediante nombre, comando, argumentos y variables de entorno; no modela todos los transportes/autenticaciones remotas. Además, `internal/harness/external/mcp.go` agrega `/mcp` al URL configurado: una URL que ya tenga una ruta puede quedar mal construida. Estos son hallazgos de lectura, no fallos reproducidos en esta investigación.

Antes de sumar integraciones remotas conviene diseñar transporte y endpoint explícitos, autenticación delegada al cliente y capacidades por adapter. Luego validar handshake, listado y una operación inocua; la presencia del archivo de configuración no prueba salud del MCP.

Para Windows, macOS, Linux, WSL y Termux, publicar soporte de cada harness por separado. Chrome, Docker, LSP y binarios auxiliares tienen requisitos distintos. Un instalador Go portable no vuelve portables sus dependencias.

Las incorporaciones deberán pasar por changes OPSX, con aprobación donde corresponda por governance, backup e inyección idempotente. La recomendación es conservar el núcleo metodológico liviano y ofrecer starters profesionales por necesidad: coordinación, diagnóstico web, diseño conectado, seguridad y evaluaciones. Todas las decisiones anteriores quedan propuestas para revisión.

## 14. Ampliación: routing, memoria y optimización de tokens

El [anexo sobre routing y optimización](routing-y-optimizacion-ia-2026-11-02.md) profundiza en Jev, elección de modelos y esfuerzo, clasificación de tareas hacia SDD, alternativas de routing, compresión de salidas, retrieval y memoria persistente. Distingue esas capacidades y sus costos; incluye compatibilidad y propuestas de pilotos antes de incorporarlas al stack.

El [informe de prácticas de la comunidad y harnesses](practicas-comunidad-y-harnesses-2026-11-02.md) amplía a videos, experiencias profesionales y foros, búsqueda estructural, codemods, contratos ejecutables, fixtures, mapas de repositorio, loops acotados y runtimes alternativos. Incluye evidencia crítica sobre cuándo agregar harnesses empeora costo o tiempo.
