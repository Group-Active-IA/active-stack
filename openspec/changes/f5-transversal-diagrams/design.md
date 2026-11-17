# Diseño F5

## Context

GitHub renderiza bloques Mermaid en Markdown. El piloto de ClickUp devolvió el bloque exacto en markdown_description y el usuario confirmó que pudo ver el diagrama con el contexto de la tarea.

## Goals / Non-Goals

- Hacer que las instrucciones de Codex cubran la documentación en sesiones nuevas y que los arneses de entrega y roadmap reutilicen esa pauta.
- Usar las superficies existentes de Markdown y ClickUp. Mantener las confirmaciones de F4 para escrituras remotas. Verificar Claude en el bloque final.

## Decisions

- Mermaid es la fuente editable. Usar un bloque cercado en el documento o en la descripción Markdown de la tarea. Elegir dependencias o secuencias cuando aclaren relaciones respaldadas por la documentación.
- Al pedir la vista, leer el origen en la sesión actual, recuperar el bloque exacto y entregarlo como Mermaid en Codex. Para ClickUp, leer markdown_description y mencionar el enlace y el contexto disponible de la tarea. La lectura es independiente de la publicación.
- Dar continuidad mediante las instrucciones entregadas a Codex para documentación general y los arneses para sus puntos de entrada específicos. Usar el mecanismo de entrega existente de cada pieza.

## Risks / Trade-offs

- text_content pierde los delimitadores de Mermaid en ClickUp. Verificar la recuperación usando markdown_description.
- Una sesión sin acceso al origen requiere informar qué documento o conexión falta. Una superficie que muestra código requiere conservar y presentar la fuente editable.
