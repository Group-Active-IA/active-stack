# Evidencia del piloto F5

- Repo. Codex CLI 0.166.0, sesión nueva 01a1eb4b-3ca8-7ff0-944e-0ce29cf3fb81. Leyó F5-PILOT-ROADMAP.md en el worktree aislado y devolvió el mismo bloque Mermaid, comprobado por comparación exacta después de normalizar saltos de línea. La relación fue F1 → F4 y se conservaron los acentos.
- ClickUp. Sesión nueva 01a1eb48-bf5c-70d3-b92e-a578a6740633. Leyó la tarea 1dfevggpuyx y devolvió el bloque exacto de markdown_description, junto con el enlace, el estado To Do y la descripción. El usuario había confirmado la vista visual del piloto inicial de ClickUp.
- Codex Desktop. El usuario abrió un chat nuevo, confirmó que ambos diagramas se veían bien y compartió el bloque Mermaid recuperado por get_task. El bloque crudo pertenece a markdown_description y es la fuente editable de la vista.
- Verificación automatizada. Pasaron 261 pruebas Go focalizadas y las 16 pruebas Python de active-roadmap.
- Claude se verifica en el bloque final acordado.

## Ampliación del piloto

- Discovery. Usar [DISCOVERY.md](pilot-fixtures/DISCOVERY.md) para comprobar que el agente propone Mermaid editable cuando la secuencia lo justifica. OpenCode CLI 1.18.32 produjo el flujo con las cuatro etapas de la fuente. Claude Code propuso una secuencia e identificó una inferencia propia. El criterio de la ampliación exige omitir esa inferencia del dibujo.
- Entrega. Usar [DELIVERY.md](pilot-fixtures/DELIVERY.md). Pedir "mostrame el diagrama" y verificar que se ofrecen Intake, Approval y Handoff antes de mostrar Mermaid. Elegir Handoff y verificar que solo aparece el bloque Confirmation a Handoff. Si el nombre coincide directamente, verificar que se devuelve el bloque sin una pregunta adicional.
- OpenCode 1.18.32. El CLI no interactivo devolvió la lista numerada. El servidor de la misma versión emitio una pregunta nativa con las tres opciones y devolvió el bloque exacto después de responder Handoff mediante su API de preguntas. La prueba observa el intercambio con el servidor.
- Antigravity Desktop. Se observaron tres opciones nativas de diagrama y la opción libre que agrega el cliente. En el ensayo con los tres bloques incluidos en el mensaje, seleccionar Handoff produjo una única vista grafica de Confirmation a Handoff. La fuente de este ensayo fue el mensaje.
- Claude Code. Con la sesión autenticada, se recuperó el bloque exacto por titulo, se ofrecieron tres opciones en modo no interactivo y se identificó el archivo ausente. El MCP de ClickUp requiere autenticación propia para completar la lectura en vivo.
- Codex CLI 0.166.0. Los dos archivos se incluyeron en el prompt de una sesión nueva tras un fallo del ejecutor de comandos de Windows. Codex propuso Mermaid fiel para Discovery y enumeró las tres opciones de Entrega sin mostrar sus bloques.
- Discovery con aclaraciones. Claude Code represento el regreso a revision y su salida hacia la confirmacion de alcance. El caso [DISCOVERY-CLARIFICATION.md](pilot-fixtures/DISCOVERY-CLARIFICATION.md) explicita el envio del pedido revisado, un paso que Claude habia identificado como inferencia en la primera redaccion.
- Cierre de Claude y ClickUp (2026-11-18). Una sesión nueva de Claude Code leyó en vivo el piloto de arquitectura en el workspace 1600300000005032 con task.get_many e includeContent. El contenido estructurado marcaba las líneas como Mermaid. El bloque recuperado coincide con la lectura independiente de Codex, desde sequenceDiagram hasta Plan ordenado para revisar. El estado observado fue to do. La operación fue de solo lectura.
