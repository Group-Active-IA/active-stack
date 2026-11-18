# Prácticas de desarrolladores avanzados y harnesses para mejorar tiempo, código y calidad

Investigación ampliada al **2 de noviembre de 2026**. Complementa el [informe por agente](harnesses-ia-por-agente-2026-11-02.md) y el [anexo de routing, memoria y tokens](routing-y-optimizacion-ia-2026-11-02.md). Son recomendaciones para revisar; no se instalaron candidatos ni se modificó el catálogo.

La ampliación busca eficiencia completa: llegar antes a un cambio correcto, revisable y mantenible. Incluye herramientas de búsqueda, codemods, contratos, pruebas, entornos, revisión y organización del trabajo. Reducir tokens puede ayudar, pero también puede ser más eficiente gastar más en una comprobación que evita horas de retrabajo.

## 1. Qué evidencia se usó

Se combinaron documentación y código de autores, artículos de profesionales, charlas y conversaciones públicas. **Documentación** confirma capacidades; **experiencias** sugieren usos y fallos; **experimentos** aportan resultados limitados a sus condiciones. Ninguno demuestra automáticamente productividad en Active Stack.

Los enlaces de videos son publicaciones reales del autor, organizador o entrevistador. Cuando YouTube solo devolvió metadatos, capítulos o no permitió recuperar contenido, se indica: no se afirma haber visto el video o leído una transcripción completa. Las propuestas de harness propias se distinguen de paquetes existentes.

## 2. Qué hacen desarrolladores con experiencia

### Mitchell Hashimoto: mejorar el entorno después de cada fallo repetido

En **My AI Adoption Journey**, publicado el 5 de febrero de 2026, describe una adopción gradual: reproducir trabajo propio para conocer los límites, dividir tareas y preparar investigación para la siguiente sesión. La parte más transferible es construir herramientas e instrucciones que permitan verificar y corregir errores recurrentes. También reconoce que saber cuándo evitar un agente ahorra trabajo. Fuente primaria: [artículo de Mitchell](https://mitchellh.com/writing/my-ai-adoption-journey).

**Aplicación propuesta:** `feedback-loop-repair`, una skill que, al cerrar un bug, distingue entre error aislado y problema recurrente del entorno. Para lo recurrente propone un test, mensaje de error, fixture o comando mejor; no agrega automáticamente una regla permanente a AGENTS.md. Conviene registrar el fallo original y comprobar si la mejora lo detecta en un caso negativo.

### Addy Osmani: trabajar por porciones pequeñas con una definición verificable

Su artículo del 4 de enero de 2026 combina especificación, plan, tareas acotadas, contexto y responsabilidad del desarrollador. La utilidad para el stack está en reducir el tamaño del salto entre intención y resultado, conservando comprensión del código. Fuente: [workflow de Addy](https://addyosmani.com/blog/ai-coding-workflow/).

Su repositorio ofrece skills de rendimiento, simplificación, observabilidad, migraciones y otros workflows. No conviene instalar el paquete entero encima de OPSX: hay solapamientos metodológicos y automatismos de commits que deben adaptarse a las reglas del proyecto. El README también documenta una limitación al instalar skills individuales sin referencias compartidas. Fuente: [Agent Skills de Addy](https://github.com/addyosmani/agent-skills).

**Aplicación propuesta:** importar selectivamente `performance-optimization`, `deprecation-and-migration` u `observability-and-instrumentation` tras revisar contenido, licencia, recursos incluidos y conflictos. La skill de simplificación debe preservar comportamiento y contratos; menos líneas no siempre significa mejor arquitectura.

### Armin Ronacher: herramientas pequeñas y ciclos rápidos

Sus recomendaciones de junio de 2025 favorecen herramientas que el agente pueda ejecutar directamente y feedback claro. El artículo distingue casos donde un CLI basta de aquellos donde MCP aporta una interfaz más fiable. Sus elecciones de modelos y permisos son experiencias históricas personales, no recomendaciones actuales para copiar. Fuente: [Agentic Coding Recommendations](https://lucumr.pocoo.org/2025/6/12/agentic-coding/).

En octubre de 2026 analiza **codemode**: el agente compone herramientas mediante código, con control de flujo y transformación de datos, en lugar de encadenar intercambios que cargan resultados intermedios en la conversación. Fuente: [What is Codemode](https://lucumr.pocoo.org/2026/10/6/codemode/).

**Aplicación propuesta:** `tool-composition`, para operaciones de lectura repetitivas y resultados estructurados. Debe conservar errores, permisos y trazabilidad; ejecutar código que llama herramientas no le da autorización adicional. No exige construir un nuevo runtime si el host ya permite composición segura.

### Mario Zechner: menos funciones por defecto y más control explícito

Su artículo de noviembre de 2025 explica la construcción de Pi alrededor de un núcleo pequeño y control del contexto. Es un documento histórico de decisiones: las capacidades y distribución actuales deben verificarse en el proyecto vigente. Fuente: [artículo del creador de Pi](https://mariozechner.at/posts/2025-11-30-pi-coding-agent/).

La charla **Building pi in a World of Slop**, de abril de 2026, plantea el costo de producir más código del que una persona puede revisar y mantener. La página del organizador contiene resumen y referencias; la recuperación de la transcripción completa falló durante esta investigación. Fuente: [charla en AI Engineer](https://ai.engineer/talks/RjfbvDXpFls-pi-coding-agent).

**Aplicación propuesta:** perfiles pequeños por tarea y límites de tamaño de entrega. Un cambio grande se divide por comportamiento o contrato; no se acepta solo porque el agente siguió generando. Extender el agente puede ser útil, pero cada extensión agrega una dependencia y otro punto de fallo.

### Simon Willison: demostrar el trabajo y conservar recetas que funcionan

Su guía reúne prácticas como baseline de tests, red/green TDD, pruebas manuales con agentes y explicaciones del código. Fuente: [Agentic Engineering Patterns](https://simonwillison.net/guides/agentic-engineering-patterns/).

El patrón de pruebas manuales propone que el agente ejecute la aplicación y registre lo observado. **Showboat** permite capturar comandos, sus salidas e imágenes en un documento ejecutable. Esto aporta un rastro más verificable que una narración escrita de memoria. Fuentes: [pruebas manuales](https://simonwillison.net/guides/agentic-engineering-patterns/agentic-manual-testing/) y [Showboat](https://github.com/simonw/showboat).

**Aplicación propuesta:** `executable-evidence`, complementario a `verification-evidence`: genera un documento corto con el escenario, comando, resultado y captura cuando corresponda. La evidencia debe poder regenerarse con datos de prueba; nunca registrar credenciales. Para revisión, `linear-walkthrough` explica entradas, transformaciones, salidas y puntos de fallo del cambio. Fuente: [linear walkthroughs](https://simonwillison.net/guides/agentic-engineering-patterns/linear-walkthroughs/).

### Dex Horthy: investigación revisable y compactación intencional

El material primario de **Advanced Context Engineering** desarrolla investigación, planificación e implementación con handoffs deliberados. El valor no está en exigir tres fases a cada tarea mínima, sino en evitar que una investigación enorme y ruidosa acompañe toda la ejecución. Un resumen equivocado también puede propagar el error. Fuente: [material del autor](https://github.com/humanlayer/advanced-context-engineering-for-coding-agents/blob/main/ace-fca.md).

Una entrevista de julio de 2026 describe cambios de práctica y límites de experimentos con código no revisado. Es una experiencia profesional en evolución, no prueba de que cierto número de agentes sea óptimo. Fuente: [entrevista y timestamps del anfitrión](https://newsletter.pragmaticengineer.com/p/context-engineering-with-dex-horthy).

**Aplicación propuesta:** agregar a los contratos de delegación un apartado «hechos verificados / hipótesis / archivos / preguntas». Antes de ejecutar un plan basado en investigación compactada, revisar las decisiones de mayor impacto contra el código y la spec.

## 3. Videos y sesiones que conviene estudiar

| Material primario | Fecha verificable o alcance | Qué observar | Acceso en esta investigación |
|---|---|---|---|
| [Karpathy — Software Is Changing (Again), YC](https://www.youtube.com/watch?v=LCEmiRjPEtQ) | Publicado 18/06/2025; keynote del 17/06 | Autonomía parcial y colaboración; capítulos 18:22 y 23:40 | Metadatos y capítulos; no reproducción completa |
| [Armin — Claude Code Fixes Two MiniJinja Issues](https://www.youtube.com/watch?v=sQYXZCUvpIc) | Enlazado por el artículo del autor del 12/06/2025; fecha del video no confirmada | Sesión concreta: localizar, reproducir y resolver issues | Enlace primario confirmado; transcripción no recuperada |
| [Mario — Building pi in a World of Slop](https://www.youtube.com/watch?v=RjfbvDXpFls) | Abril de 2026; página del organizador consultada | Núcleo pequeño, extensibilidad y revisión humana | Resumen del organizador; YouTube/transcripción con errores de acceso |
| [Addy — AI-Assisted Developer Workflow, JSNation](https://www.youtube.com/watch?v=SpKtpW9TGF0) | Publicado 28/11/2024 | Iteración, debugging y pruebas de interfaz | Descripción y capítulos; herramientas mostradas pueden haber cambiado |
| [Addy — From Chrome DevTools to AI Engineering](https://www.youtube.com/watch?v=2fyPnxKu8ZM) | Entrevista de 2026; fecha exacta no consolidada | Capítulos de agentes, deuda cognitiva y loop engineering | Landing del anfitrión y timestamps consultados |
| [Dex — Advanced Context Engineering, YC Root Access](https://www.youtube.com/watch?v=IS_y40zY-hc) | Publicado 25/08/2025 según metadatos; material refiere charla del 20/08 | Research/plan/implement y handoffs | Material escrito primario; video no reproducido |
| [Dex — Context engineering, Pragmatic Engineer](https://www.youtube.com/watch?v=Usufn8IQJgw) | 15/07/2026 en landing del anfitrión | Evolución del workflow y revisión del trabajo asíncrono | Artículo/timestamps; no reproducción completa |
| [Thoughtworks — conversación vinculada a maintainability sensors](https://www.youtube.com/watch?v=uLWOLmeHOSE) | Enlace desde artículo de mayo de 2026; fecha de video pendiente | Feedback de arquitectura y mantenimiento | Artículo completo; recuperación de video falló |

La charla de Karpathy es una fuente conceptual sobre autonomía graduable, no una comparativa de ocho agentes ni un benchmark de productividad. La sesión oficial [Boris Cherny y Jarred Sumner](https://claude.com/code-with-claude/session/sf-live-coding-session-with-boris-cherny-and-jarred-sumner) sirve como material adicional de workflows cotidianos, pero su página no permitió establecer aquí una transcripción utilizable.

## 4. Lo que revelan los foros y los issues

Las experiencias públicas se usan para formular pruebas y riesgos. No se consideran muestras representativas, y los votos no miden calidad técnica.

| Conversación abierta | Observación útil | Consecuencia para el stack |
|---|---|---|
| [HN: My AI Adoption Journey](https://news.ycombinator.com/item?id=46903558) | Conviven mejoras personales con reportes de más revisión, drift y costo de integración | Medir tiempo total hasta aceptación, incluyendo correcciones humanas |
| [HN: Agentic Coding Recommendations](https://news.ycombinator.com/item?id=44255608) | Hay éxitos con checks concretos y preguntas sobre la dificultad real de las tareas | Evaluar en casos propios; no extrapolar arreglar tipos a diseñar negocio |
| [HN: charla de Karpathy](https://news.ycombinator.com/item?id=44314423) | Debate sobre structured outputs y límites semánticos aun con JSON válido | Validar contratos y significado; parsear no prueba corrección |
| [Reddit: setup compartido de Boris](https://www.reddit.com/r/AgentsOfAI/comments/1q2s83g/boris_cherny_creator_of_claude_code_shares_his/) | Repost de un workflow con paralelismo, permisos seleccionados y feedback; comentarios preguntan costo y transferibilidad | No convertir una rutina personal en preset universal |
| [GitHub: referencias perdidas en skills de Addy, #361](https://github.com/addyosmani/agent-skills/issues/361) | La instalación por carpeta no incluye referencias en la raíz; issue cerrado documentando la limitación | Verificar payload completo y referencias después de instalar |
| [GitHub: no debilitar tests para arreglar flakiness, #2555](https://github.com/fullsend-ai/fullsend/issues/2555) | Un reporte pide impedir sustituciones que reducen lo que un test verifica | Revisión explícita de cambios de tests; conservar el escenario original |
| [Reddit: benchmark de cinco herramientas de tokens](https://www.reddit.com/r/ClaudeCode/comments/1vilwwm/i_benchmarked_5_token_saving_tools_across_codex/) | Autor declara afiliación a repowise; resultados sugieren diferencias de uso efectivo entre agentes | Medir invocaciones reales y calidad, no solo disponibilidad o salida reducida |

El post original de Boris enlazado desde Reddit no pudo abrirse por un 403; el contenido del repost se trata como **reproducción secundaria**, no como verificación directa de todas las cifras o prácticas. Tampoco se trasladan sus comandos de publicación al stack: aquí los commits y builds requieren la autorización aplicable.

El issue de Addy es especialmente relevante para un instalador. Cerrar un issue como completado no siempre significa que el comportamiento haya cambiado: en este caso se documentó una limitación de portabilidad. La comprobación necesaria es instalar el payload en un directorio limpio y resolver sus referencias, con la versión fijada.

## 5. Harnesses que reducen el trabajo equivocado

### Arquitectura verificable y contratos ejecutables

**`architecture-fitness` — propuesta propia, `skill` + `config` + herramientas por stack.** Traduce un conjunto pequeño de decisiones aprobadas en sensores: dependencias permitidas, ciclos prohibidos, límites de paquetes y convenciones de interfaces. Los errores deberían explicar el arreglo esperado. La experiencia de Birgitta Böckeler combina sensores deterministas y revisión semántica, y advierte que más feedback también puede provocar ruido y refactors innecesarios. Fuente: [Maintainability sensors for coding agents](https://www.martinfowler.com/articles/sensors-for-coding-agents.html).

Para TypeScript/JavaScript, [dependency-cruiser](https://github.com/sverweij/dependency-cruiser) permite validar reglas de dependencias. Para otros lenguajes hay que seleccionar herramientas propias; no instalar un ecosistema Node de análisis para todo proyecto Go. El harness define pocas reglas de alto valor y excepciones justificadas, dejando las decisiones de diseño no mecanizables a revisión.

**`api-contracts` — propuesta propia, `skill` + herramientas externas opcionales.** Comparte contratos antes de paralelizar frontend/backend y comprueba compatibilidad de respuestas, errores y versiones. [Spectral](https://github.com/stoplightio/spectral) valida reglas sobre especificaciones; [Specmatic](https://github.com/Specmatic/specmatic) es candidato de contratos ejecutables. Son funciones diferentes: una spec bien formada no prueba que el servidor la cumpla. Ediciones, runtime y licencias deben revisarse antes de empaquetar.

**`approved-fixtures` — propuesta propia, `skill` y assets del proyecto.** Mantiene un conjunto mínimo de entradas, errores y resultados aprobados: payloads de API, árboles de configuración, golden files o capturas de referencia. El agente puede ampliar casos, pero cambiar una expectativa existente requiere explicar el cambio de requisito. La propuesta se inspira en convertir feedback en infraestructura verificable, no en un paquete encontrado con ese nombre. Fuente conceptual: [Harness engineering for coding agent users](https://martinfowler.com/articles/harness-engineering.html).

### Transformaciones deterministas y código de referencia

**`deterministic-codemods` — propuesta propia, `skill` con CLI opcional.** El modelo propone una transformación y sus excepciones; una herramienta aplica el patrón de forma repetible. Es apropiado para cambios de importaciones, firmas o APIs repetidas. Flujo: buscar coincidencias, probar sobre ejemplos positivos/negativos, revisar un diff pequeño y recién entonces ampliar. [ast-grep](https://ast-grep.github.io/guide/introduction.html) ofrece búsqueda y reescritura estructural; regex sigue siendo útil para texto simple.

**`reference-recipes` — propuesta propia, `skill` + assets.** Reúne ejemplos pequeños ya ejecutados de las operaciones frecuentes del proyecto: agregar un endpoint, un adapter o una migración. Un ejemplo incluye versión y prueba, no solo una plantilla bonita. Esto permite recombinar soluciones conocidas sin copiar código viejo indiscriminadamente. Fuente: [Hoard things you know how to do](https://simonwillison.net/guides/agentic-engineering-patterns/hoard-things-you-know-how-to-do/).

### Entornos y feedback más rápidos

**`reproducible-dev-env` — propuesta propia, `skill` y configuración opt-in.** Documenta versiones, setup mínimo, servicios y datos de prueba. Contenedor, devcontainer o sandbox dependen del proyecto y del host. El objetivo es reducir fallos de entorno y permitir restaurar un caso, no imponer Docker donde basta un binario y un fixture.

**`fast-feedback` — propuesta propia, `skill` + scripts del repo.** Mantiene comandos dirigidos por paquete o cambio y una comprobación de integración posterior. Ejecuta primero el control barato que puede detectar el problema: formato, tipos, test focalizado o contrato. No reemplaza el conjunto requerido de verificaciones. Una caché de tests solo ayuda si el comando conserva sus condiciones de validez.

**`test-integrity` — propuesta propia, `config` + `skill`.** Revisa tests borrados, skips, assertions reducidas y expectativas alteradas. Cada test importante debe poder fallar con una implementación incorrecta pertinente. Mutation/property/fuzz testing se usan selectivamente donde aporten información; su costo puede ser alto. Este harness complementa el TDD ya incluido en Active Stack.

### Evaluar el propio harness

**`skill-activation-evals` — propuesta propia, `skill` + laboratorio.** Scott Spence evaluó activación de skills con Claude Code real en sandboxes; distingue de su prueba anterior vía API. Sus resultados cambian al introducir prompts más difíciles y falsos positivos. El experimento mide activación, no corrección del desarrollo completo; tampoco permite afirmar un 100% universal. Fuente: [evaluación de Scott](https://scottspence.com/posts/measuring-claude-code-skill-activation-with-sandboxed-evals).

[UiPath/coder_eval](https://github.com/UiPath/coder_eval) es un candidato concreto para suites de tareas que usan skills, MCP y CLI, con sandboxes, YAML y comparaciones A/B. Puede complementar Promptfoo según el objeto a evaluar. Su incorporación requeriría fijar runtime, runners, presupuesto y datos, evitando que el evaluador apruebe solo la narración final del agente.

Los estudios también cuestionan la inflación de instrucciones. **Evaluating AGENTS.md**, revisado en septiembre de 2026, encontró que los archivos de contexto no mejoraban generalmente el éxito en sus muestras y aumentaban costo promedio. Esto no invalida reglas de seguridad o convenciones particulares; sugiere evaluar qué instrucciones aportan información que el agente no obtendría mejor del repositorio. Fuente: [estudio primario](https://arxiv.org/abs/2602.11988).

## 6. Más harness no siempre significa más eficiencia

Naresh Jain comparó diez ejecuciones por condición al construir un mismo servicio API con evaluación externa. El modelo fuerte pasó todas con y sin harness; agregarlo aumentó aproximadamente 10,5 veces tokens, tres veces tiempo y siete veces costo. El modelo económico pasó 3/10 sin harness y 10/10 con él. Es un experimento pequeño y específico, pero demuestra por qué no se debe asumir que toda estructura reduce consumo o mejora el resultado de cualquier modelo. Fuente: [Are Harnesses Worth the Cost?](https://martinfowler.com/articles/are-harnesses-worth-the-cost.html).

Propuesta de evaluación para Active Stack: comparar **sin harness / harness mínimo / paquete completo** en las mismas tareas y revisiones del repositorio. Medir tiempo humano, tiempo total, regresiones, calidad de tests, cambios ajenos al alcance, costo y mantenimiento del entorno. Solo conservar el módulo que resuelve un fallo observable o una necesidad clara.

No conviene habilitar todos los expertos en cada cambio ni encadenar varios revisores que repiten el mismo análisis. La revisión independiente agrega valor si tiene criterios diferentes y evidencia; si solo pide «encontrá más problemas», puede producir trabajo de validación de falsos positivos.

## 7. Herramientas y workflows adyacentes

### Búsqueda, mapas y contexto dirigido

**ast-grep + `structural-search`.** La [skill del proyecto](https://github.com/ast-grep/agent-skill) y su [guía para agentes](https://ast-grep.github.io/advanced/prompting.html) proporcionan un proceso de patrón mínimo, inspección de AST y fixtures antes de reescribir. El [MCP oficial](https://github.com/ast-grep/ast-grep-mcp) está marcado experimental; empezaría con CLI y skill. Complementa `rg`, que sigue siendo la primera herramienta para búsqueda textual. Windows tiene distribución nativa, todavía sin prueba local de este informe.

**Mapas de repositorio.** [Aider](https://aider.chat/docs/repomap.html) genera mapas de símbolos y firmas con ranking y presupuesto de contexto. Es una función nativa de Aider, no una herramienta que un prompt habilite automáticamente en Codex. Propuesta: `repo-map-budget`, una skill con generador explícito por lenguaje, actualización por revisión y referencias a archivos. Todos los agentes pueden consumir el artefacto, pero hay que implementar o elegir el generador; no se encontró un wrapper universal certificado.

**Repomix + `context-pack-on-demand`.** Empaqueta un subconjunto del repositorio, permite exclusiones y compresión estructural. Su [compresión](https://repomix.com/guide/code-compress) sirve para orientación; al omitir cuerpos de implementación no basta para probar un bug lógico. El [MCP](https://repomix.com/guide/mcp-server) tiene modo sandbox; sin él no debe asumirse confinamiento al workspace. Requiere runtime y revisión de exclusiones. Conviene usarlo para entregar contexto acotado, evitando cargar automáticamente el repo completo o incluir secretos.

### Navegador por CLI y evidencia

**Agent Browser** dispone de CLI, referencias de snapshot, sesiones y capturas; actualmente también ofrece MCP con perfiles de herramientas. No debe describirse como una alternativa exclusivamente sin MCP. Hay binarios Windows x64 y otras plataformas, sujetos a requisitos del navegador. Fuente: [proyecto de Vercel](https://github.com/vercel-labs/agent-browser).

**Playwright CLI** ofrece otra ruta con skills y herramientas de diagnóstico. Fuente: [instalación oficial](https://playwright.dev/agent-cli/installation). Propuesta `browser-cli`: opt-in para tareas donde el agente opera mejor por shell, comparándolo con Playwright MCP ya incluido. No instalar tres navegadores por defecto. Los snapshots cambian al cambiar la página; las referencias antiguas no deberían usarse sin reinspección.

La [demo de Cole Medin](https://www.youtube.com/watch?v=7aEQnTsI6zs), publicada el 18 de enero de 2026 según metadatos, es material práctico adicional. La investigación accedió a descripción, no a una transcripción completa; no se trasladan sus porcentajes personales a una comparativa independiente.

### Loops acotados y tareas persistentes

**Ralph** es una familia de técnicas de repetición, no un único paquete. El [material de Geoffrey Huntley](https://ghuntley.com/ralph/) describe un loop headless; distintas herramientas usan el nombre con estrategias de sesión diferentes. Propuesta propia `bounded-repair-loop`: criterio de aceptación externo, máximo de intentos/costo/tiempo, detención por falta de progreso y prohibición de debilitar tests. Usarlo dentro de una tarea OPSX; no como otro controlador de estado. El Bash original requiere adaptación para PowerShell; una CLI headless disponible no demuestra por sí sola compatibilidad nativa con todos los agentes.

**Beads** permite tareas con dependencias y reclamación de trabajo. El [upstream actual](https://github.com/gastownhall/beads) utiliza Dolt; no debe describirse con el almacenamiento de versiones históricas. Tiene distribución para Windows y otros sistemas, con dependencias adicionales. Solo lo probaría si varios agentes sufren colisiones o pierden tareas: OpenSpec conserva autoridad sobre changes y Beads, si se elige, registra subtareas derivadas con un mapeo explícito. Su inicialización puede modificar instrucciones/hooks, por lo que no debe ejecutarse como una inspección inocua.

### Metodologías y runtimes alternativos

**GSD Core** es el destino actual del proyecto que antes se distribuía como get-shit-done. Sus waves, agentes frescos y verificación aportan referencias útiles, pero el flujo se solapa fuertemente con la fundación y OPSX. El [runtime actual](https://github.com/open-gsd/gsd-core) y sus adapters deben instalarse con sus transformaciones de herramientas y esquemas, no copiando prompts de Claude a todos. Evaluarlo como alternativa excluyente o extraer técnicas; evitar dos fuentes autoritativas de tareas.

**Pi / Oh My Pi.** Pi es un agente extensible, no una skill portable; [su distribución actual](https://github.com/earendil-works/pi) debe distinguirse de los enlaces históricos. [Oh My Pi](https://github.com/can1357/oh-my-pi) es otro runtime, con herramientas y dependencias propias. Para el stack serían nuevos targets/adapters, no módulos que cambian mágicamente el editor interno de Codex. Sus benchmarks de formatos de edición son evidencia del autor y dependen del modelo; no prueban una mejora universal.

### Compatibilidad práctica de estos candidatos

**S = portable mediante shell habilitada y runtime instalado; A = necesita adapter o extensión del host; T = target/metodología alternativa.** S no significa que Active Stack ya lo instale ni que todos los comandos funcionen en Windows/Termux. La compatibilidad de producto general se detalla en el informe principal.

| Capacidad o candidato | Claude | Codex | OpenCode | Gemini CLI | Cursor | VS Code/Copilot | Antigravity | Windsurf/Devin |
|---|---|---|---|---|---|---|---|---|
| ast-grep, scripts de contratos y sensores | S | S | S | S | S | S | S | S; identificar runtime |
| Mapas y fixtures como archivos | S | S | S | S | S | S | S | S |
| Repomix / navegador CLI / Showboat | S | S | S | S | S | S | S | S; validar migración |
| Automatizar esos comandos con hooks | A | A | A/plugin | A | A; local/cloud | A; por harness | A; por superficie | A; nuevo esquema |
| Loop headless con límites | S/A | S/A | S/A | S/A | A; CLI concreto | A; CLI no editor genérico | A; CLI concreto | A; CLI concreto |
| Beads como ledger de subtareas | S/A | S/A | S/A | S/A | S/A | S/A | S/A | S/A |
| GSD como metodología completa | T | T | T | Validar adapter | T | T | T; CLI | T; versión objetivo |
| Aider, Pi u Oh My Pi como runtime | T | T | T | T | T | T | T | T |

La matriz expresa **vías de integración**, no soporte nativo probado. Las rutas CLI permiten reutilizar herramientas sin añadir MCP si ya son fiables. Los hooks requieren fixtures de eventos reales; un JSON común no garantiza compatibilidad. Para GitHub Copilot, Cursor cloud y Antigravity IDE, distinguir entornos y límites antes de prometer interceptación de argumentos o resultados.

### Qué ya cubre Active Stack y qué sería nuevo

| Grupo | Situación | Recomendación |
|---|---|---|
| TDD, debugging y review genérico | Ya en catálogo | Mejorar aplicación y evidencia, no duplicar packs |
| Playwright y verificación visual | Ya presentes | CLI/browser alternativo solo con necesidad medida |
| Arquitectura/clean architecture | Hay skills conceptuales | Añadir sensores ejecutables específicos del proyecto |
| Contratos de tareas y worktrees | Base de delegación existente | Formalizar ownership, límites e integración |
| Código de referencia y fixtures aprobados | No identificados como harness dedicado | Candidatos propios pequeños de alto valor |
| Evaluación de activación y utilidad de skills | No basta catalog.Load ni tests Go | Laboratorio de comportamiento con agentes reales |
| GSD, Beads, Pi/OMP | Nuevos sistemas con solapamiento | Laboratorio o alternativa explícita, fuera del Full automático |

## 8. Plan de adopción propuesto

| Orden | Piloto concreto | Medida para decidir si se conserva |
|---|---|---|
| 1 | `fast-feedback` + `approved-fixtures` | Menos tiempo hasta reproducir/corregir, sin perder escenarios |
| 2 | `test-integrity` + `executable-evidence` | Tests detectan fallos y revisor reproduce la demostración |
| 3 | ast-grep + `structural-search`/codemods | Refactor acotado con menos omisiones y diff revisable |
| 4 | `architecture-fitness` + `api-contracts` | Detecta violaciones reales antes de integración, con poco ruido |
| 5 | `repo-map-budget`/contexto dirigido | Menos exploración irrelevante manteniendo calidad del arreglo |

Primero probar `fast-feedback`, `approved-fixtures`, `test-integrity` y `executable-evidence` en uno o dos proyectos. Son mejoras del entorno que pueden beneficiar a modelos distintos y no exigen otro orquestador.

Después, agregar búsqueda estructural, mapas de repositorio y empaquetado selectivo donde la exploración sea el cuello de botella. En proyectos con varios servicios, priorizar contratos e invariantes de arquitectura antes de multiplicar agentes escritores.

Finalmente evaluar loops autónomos y gestores alternativos de tareas. Deben tener criterio de aceptación externo, presupuesto, máximo de intentos, detección de falta de progreso y preservación del trabajo. La existencia de un comando «auto» no equivale a que su uso esté autorizado para publicar o cambiar configuración del usuario.

Todas las incorporaciones son propuestas pendientes: requerirán revisión de licencias y versiones, pruebas del payload y adapter, y cambios OPSX cuando se decida implementarlas.
