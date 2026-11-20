# R-01 · Controles complementarios de Codex

Esta edición presenta resultados obtenidos en contextos de prueba aislados. Las imágenes de consola y los registros de operaciones del despliegue se conservan internamente.

| Control | Resultado |
|---|---|
| Q9 | La consulta local de la rama terminó con código 0 sin solicitudes de aprobación en los dos perfiles ensayados. La operación completa de cierre requiere el comando de Fase 4. |
| Q15 y V9 | Dos fallos internos de etiquetado produjeron un aviso y permitieron el guardado sin reformateo. La pérdida del transporte entre cliente y MCP en otro ensayo dejó el primer guardado y ninguna observación para el segundo. El timeout externo del hook de Codex quedó sin ejecutar. |
| V11 | El prototipo inicial acortó adecuadamente nombres largos, pero dos ramas cortas diferentes colisionaron. La variante de sufijo separó sus dos claves; la continuidad del historial depende de su origen. [Ensayo y compatibilidad](v11-sufijo-en-dos-worktrees.md) |
| V6 | El guardado directo quedó etiquetado. Otras herramientas también crean observaciones o mutan su contenido: no todas usan los mismos campos. El guardado del prompt creó una entidad separada. [Cobertura de las herramientas](alcance-q5-q7-q8.md) |
| V8 | El cliente enumeró siete hooks pendientes de revisión y pidió revisión antes del turno. Queda pendiente el verificador de Fase 10. |
| V12 | Se ejecutó el piloto en macOS x86_64. La evidencia de Windows sigue pendiente. |
| V13 | Al leer desde main, la búsqueda devolvió dos decisiones sintéticas incompatibles. El modelo eligió la compartida en un ensayo. [Resultado](codex-v13-lectura-main.json) |

Las observaciones solo demuestran el comportamiento de las combinaciones ensayadas. [Matriz del piloto](../engram-cloud-piloto-2026-11-19.md)
