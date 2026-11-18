# Fortalecer la fundación de Active Stack

El [catálogo](../../../../internal/catalog/harnesses.yaml) describe lo que entrega el instalador y [Arquitectura §4.2](../../../../ARCHITECTURE.md#42-fundación-del-proyecto-una-vez-por-proyecto) describe la secuencia base. El catálogo incluye discovery opcional después de inicializar OpenSpec, un paso que falta en el diagrama de arquitectura. La tabla parte de esas fuentes y deja la validación del comportamiento en agentes reales para el piloto.

| Parte | Qué hay hoy | Alternativas y cuándo evaluarlas |
|---|---|---|
| Inicialización | Active Orchestrator prepara OpenSpec y encadena skills bajo demanda. | Plantilla de repo para proyectos nuevos con stack repetible. Para repos existentes, relevamiento de artefactos y propuesta de incorporación por etapa. Probar reejecución, recuperación de una fase interrumpida y actualización de una fundación anterior. |
| Discovery | Checklist opcional de 11 puntos y ficha comparativa por competidor a partir de una URL. | Entrevistas y observación para problemas con usuarios accesibles. Telemetría, soporte y lectura del repo para productos en marcha. Un árbol de oportunidades vincula resultados, necesidades y experimentos cuando hay muchas hipótesis. Registrar fuente, fecha, hipótesis y qué evidencia podría cambiar cada decisión. [Product Talk](https://www.producttalk.org/opportunity-solution-trees/). |
| Base de conocimiento | kb-creator genera diez Markdown canónicos a partir de fuentes o entrevista. Engram aporta memoria de trabajo persistente. | Markdown/Git para revisión y portabilidad. Un índice local SQLite FTS5 mejora recuperación textual. Una wiki como Confluence puede servir para edición compartida. pgvector sirve si ya se opera PostgreSQL. Qdrant local o administrado permite probar recuperación vectorial o híbrida. Mantener enlaces a la fuente y medir relevancia, vigencia, latencia, privacidad y costo. [SQLite FTS5](https://www.sqlite.org/fts5.html) · [pgvector](https://github.com/pgvector/pgvector) · [Qdrant](https://qdrant.tech/documentation/). |
| Fundación visual | design-system guarda una política por proyecto y, cuando se elige, genera sistema y mockups locales. | Reusar tokens y componentes de producto. Figma para colaboración de diseño o Storybook para documentar y revisar componentes existentes. Probar un mismo flujo en móvil y escritorio, y verificar trazabilidad entre prototipo e implementación. [Figma](https://www.figma.com/variables/) · [Storybook](https://storybook.js.org/docs). |
| Roadmap | roadmap-generator produce CHANGES.md con fichas y dependencias a partir de la KB. OpenSpec registra el estado de los changes. | Git/Markdown para trabajar offline. Trello para coordinar en un tablero con miembros. GitHub Projects para equipos que ya planifican con issues y PR. Comparar acceso de colaboradores, permisos, dependencias, exportación y trabajo operativo. [Trello REST](https://developer.atlassian.com/cloud/trello/rest/api-group-cards/) · [GitHub Projects](https://docs.github.com/en/issues/planning-and-tracking-with-projects/learning-about-projects/about-projects). |
| Búsqueda y registro de skills | find-skill busca capacidades y skill-registry indexa las instaladas. | Lista curada por stack, registro nativo de cada cliente o descubrimiento bajo demanda. Comparar disponibilidad entre clientes, procedencia, versiones reproducibles y contexto cargado. |
| Instrucciones y gobernanza | agent-instruction genera referencias y reglas del proyecto. El instalador entrega módulos del orquestador y políticas por agente. | Reglas mínimas con carga por acción, configuración nativa de clientes y controles ejecutables donde corresponda. Medir el contexto efectivo y comprobar precedencia, preservación de configuración y rollback. Esta línea se cruza con las fichas 03–05 y 08 del roadmap. |

## ODD y SDD por unidad de trabajo

Organic Driven Development, tal como lo documenta Gentle AI, usa una ruta liviana para trabajo entendido y un documento recuperable para trabajo sustancial. El candidato para Active Stack es conservar OpenSpec para contratos públicos, migraciones, cambios transversales o coordinación entre varias personas y ensayar ODD para ajustes acotados y reversibles con aceptación clara. La autorización, gobernanza y evidencia de verificación se mantienen en ambos recorridos. El piloto debe probar cuándo escalar a SDD y cómo trasladar alcance, decisiones y evidencia al escalar. [Uso de ODD en Gentle AI](https://github.com/Gentleman-Programming/gentle-ai/blob/main/docs/usage.md).

## Piloto de roadmap

1. Probar primero con CHANGES.md y elegir un solo tablero remoto para el piloto. Trello tiene sentido si el equipo ya coordina allí. GitHub Projects es el primer candidato si la coordinación ocurre alrededor de issues y PR.
2. En el piloto, Git conserva alcance y dependencias, OpenSpec conserva progreso técnico y el tablero gestiona asignación y estado operativo. Vincular cada ítem con un identificador estable y mantener la correspondencia entre responsables del equipo y cuentas del proveedor.
3. Ensayar publicación optativa y repetible, mapeo de miembros, acceso insuficiente, uso offline, cambios de responsable y discrepancia entre un change completo y una tarjeta abierta. Postergar la edición bidireccional hasta acordar quién controla cada campo.
4. Con un proyecto nuevo y otro existente, medir tiempo hasta el primer trabajo aprobado, repeticiones al retomar, recuperación de evidencia, costo de mantenimiento y conflictos de estado. Proveedor, responsables y umbrales (TBD).

## Decisión y checklist antes de ejecutar

1. Active Orchestrator estima alcance, duración, necesidad de retomar y coordinación. Para un trabajo autocontenido que se puede resolver y verificar en una sesión, recomienda la ruta directa. Si el proyecto lleva varios días, necesita continuidad o involucra a varias personas, requiere una fundación proporcional. Si el alcance crece, vuelve a evaluar la necesidad.
2. Presenta una checklist antes de inicializar, instalar, generar documentos o conectar servicios. Cada opción muestra beneficio, artefacto que deja, si es requerida o elegible, variantes y requisitos. También ofrece reutilizar artefactos existentes. El usuario selecciona varias etapas en una sola interacción y elige una variante donde corresponda.
3. Mostrar el resumen y ejecutar según el plan elegido. Guardar decisiones y avance para poder retomar. En Codex y Claude, probar selección múltiple en terminal y escritorio por separado. Si una superficie carece de controles nativos, permitir confirmar la misma selección con una lista numerada en texto. La disponibilidad de controles nativos queda por verificar.
4. Para roadmap, ofrecer Git/Markdown o tablero compartido con Trello como candidato y GitHub Projects como alternativa. Explicar al elegir qué se guarda localmente, qué se publica y cómo se asignan responsables por cuenta. En proyectos sin UI, ofrecer omitir fundación visual. Cuando haya fuentes suficientes, ofrecer importar o reutilizar la KB.

## Plan por arnés

Primero acordar el contrato de fase, con nombre, beneficio, requisitos, variantes, entrada/salida, estado y reanudación. Después preparar propuestas por arnés o conjunto estrechamente acoplado. Registrar arneses sin cambios cuando una verificación demuestre que ya cumplen el contrato.

| Arnés | Beneficio que debe mostrar la checklist | Trabajo a planificar |
|---|---|---|
| active-orchestrator | Elegir una ruta proporcional y retomar el trabajo | Clasificar alcance, presentar la selección antes de ejecutar, resolver dependencias y reanudar con el mismo plan. |
| discovery-research y web-scraper | Reducir incertidumbre sobre usuarios, problema y competencia | Ofrecer entrevista, fuentes existentes y análisis de mercado según la evidencia disponible. Devolver hallazgos con fuente e hipótesis pendientes. |
| kb-creator | Recuperar dominio y decisiones al retomar | Ofrecer importar, reutilizar o crear KB. Comparar opciones de búsqueda como piloto con el mismo corpus. |
| design-system y design-mockups | Alinear decisiones visuales antes de implementar | Ofrecer las variantes visuales existentes, reutilizar sistema y guardar la elección por proyecto. |
| roadmap-generator | Priorizar, asignar responsables y seguir dependencias | Generar el plan local o preparar publicación optativa al proveedor elegido. Probar mapeo de miembros y evitar duplicados al republicar. |
| find-skill y skill-registry | Tener a mano capacidades pertinentes al proyecto | Distinguir sugerencia, instalación e indexación. Reutilizar el inventario cuando ya existe. |
| agent-instruction | Mantener reglas, referencias y decisiones disponibles al retomar | Generar referencias a las variantes seleccionadas y conservar el contenido del usuario al actualizar. |
| openspec y sdd-orchestrator | Mantener cambios complejos trazables y verificables | Revisar el acoplamiento de la inicialización con OpenSpec y ensayar la convivencia ODD/SDD por unidad de trabajo. |

Validación integrada con un trabajo de una sesión, un proyecto nuevo de varios días y un repo existente. En cada superficie, comprobar selección múltiple, elección de variantes, ausencia de escrituras antes de confirmar, reanudación y reutilización de fases. Pendientes (TBD) — conjunto mínimo obligatorio, señales para casos limítrofes, matriz real de controles por cliente, proveedor remoto y responsables de las propuestas.

## ClickUp y MCP local de Active Stack

La skill tup-campus-navigator usa el MCP oficial de ClickUp para operar tareas y un MCP local separado para Moodle. Esto da un precedente para combinar un servicio de gestión con herramientas propias del dominio.

Probar primero un caso de creación y asignación con el MCP oficial de ClickUp para identificar qué reglas conviene encapsular en Active Stack. Como siguiente opción, un MCP local podría exponer operaciones de metodología como previsualizar tareas desde el roadmap, publicar las seleccionadas y reconciliar sus vínculos. Cada operación debe aplicar el identificador estable, las dependencias y el mapeo de responsables definido por el proyecto. Repetir una publicación debe actualizar o reutilizar la tarea vinculada sin duplicarla.

El contrato puede admitir una salida local en Markdown y una salida remota inicial en ClickUp. Trello y GitHub Projects siguen como candidatos. Git conserva alcance y dependencias, OpenSpec registra avance técnico cuando se usa SDD y el tablero elegido gestiona asignación y seguimiento. El mecanismo de conexión del MCP local con ClickUp y el primer proveedor quedan (TBD). Validar desde Codex y Claude, con previsualización, reintento tras un fallo parcial y recuperación del vínculo desde el change.

## F4 · Vista previa y resultado de la gestión de tareas

- **Antes de publicar**, mostrar proyecto, destino y una tabla con ID, título, responsable, dependencias, estado previsto y acción por fila — crear, actualizar, omitir o resolver un problema. Permitir elegir varias filas. Preparar la vista previa sin escrituras remotas.
- **Después de publicar**, mostrar el resultado por fila — creada, actualizada, sin cambios, omitida o fallida — con enlace a la tarea cuando exista y motivo de cada fallo. Permitir consultar el estado más adelante y reintentar las fallidas conservando sus identificadores.
- **Cómo mostrarlo**, usar una tabla textual en terminal y una tabla visual en escritorio cuando el cliente lo permita. Una vista de tablero puede ayudar a revisar el flujo y un diagrama a entender dependencias. Ambas vistas deben usar los mismos datos de la operación.
- **Investigación al abrir F4**, probar el MCP oficial de ClickUp, comparar proveedores y precisar el alcance del MCP local, su conexión, el modelo de identidad y el método para reconciliar estados. La selección técnica queda (TBD).

## F5 · Diagramas y explicaciones visuales

Incorporar la visualización como capacidad transversal del stack. Ejemplos — mapa de usuarios y procesos en discovery, diagramas de arquitectura o secuencia al explicar un diseño, dependencias del roadmap y estados de publicación en gestión de tareas.

Tomar UMLetini, citado como referencia de la experiencia en Codex Desktop, para definir qué experiencia se quiere reproducir. Al abrir F5, comparar formatos editables como Mermaid o PlantUML y opciones de previsualización en escritorio y terminal. Probar que el diagrama se pueda compartir y regenerar aunque un integrante use otro cliente. Integración, formato y superficies concretas (TBD).

## Alcance de F1 y F3 por agente

F1 integra el diseño del contrato y la adaptación de los arneses de fundación en una sola ficha. El primer paso es inventariar cada arnés, su repositorio o componente, versión empaquetada, consumidores y superficie soportada. Luego adaptar las etapas que necesiten cambios para cumplir el contrato y verificar las que ya lo cumplan. La entrega de F1 se verifica con el recorrido completo. Cada arnés puede llevar su propio PR y actualizar su referencia en el instalador si corresponde.

F3 debe trazar dónde se decide ODD o SDD al iniciar el proyecto y al iniciar una unidad de trabajo. Si la fundación registra una preferencia de proyecto, revisar la skill active-orchestrator. Si el modo se elige por tarea, revisar el orquestador de entrega y cómo consume esa preferencia. Considerar en AGENTS.md, CLAUDE.md y los equivalentes de otros agentes una instrucción breve que indique cómo entrar al flujo. Mantener la lógica en el componente que la ejecuta. La ubicación definitiva queda (TBD).

El catálogo actual declara active-orchestrator para Claude, Codex, OpenCode y Antigravity. Usar esos cuatro como punto de partida y verificar por separado las capacidades de cada arnés en el resto de los agentes y sus superficies. Prioridad, mecanismos nativos y ampliaciones de soporte quedan para la propuesta de cada ficha.
