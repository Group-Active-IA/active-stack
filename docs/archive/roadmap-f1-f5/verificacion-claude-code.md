# Cierre de verificación de Claude Code

> Encargo cerrado el 2026-11-18. Ver el [resultado de la verificación](verificacion-claude-code-resultado.md).

**Frase de inicio** · «Implementá y verificá lo pendiente de Claude».

Este encargo revisa el roadmap F1–F5 ya integrado, reproduce sus recorridos en Claude Code y resuelve los pendientes de Claude. Antes de empezar, leer el [roadmap publicado y su flujo histórico](CHANGES-publicado.md), la [copia de continuidad del cierre](CHANGES-copia-local.md) y las reglas vigentes en [CLAUDE.md](../../../CLAUDE.md). El informe de trabajo en equipo sigue siendo el alcance de la próxima mejora.

## Encargo

Resolvé los detalles rutinarios con la evidencia del repositorio y las herramientas disponibles. Si falta un acceso externo o una decisión que condiciona el resultado, completá primero los casos independientes y pedí únicamente ese insumo.

1. Partí de `origin/main` y usá una rama y un worktree aislados. Registrá SHA, plataforma, versiones de Active Stack, Claude Code y OpenSpec, y herramientas disponibles. Leé el catálogo real y la evidencia publicada. Usá subagentes con casos acotados cuando estén disponibles.
2. Prepará proyectos descartables fuera del checkout para probar la instalación Full de Claude. Ejecutá los comandos y abrí procesos nuevos de Claude Code cuando el caso requiera comprobar inicio o reanudación. Comprobá el resultado en archivos y herramientas además de la respuesta del agente.
3. Ejecutá la matriz completa. Para cada falla, reproducila, determiná si es de código, instalación, alcance declarado o acceso externo. Corregí los defectos y verificá nuevamente los recorridos. F3 y F4 se empaquetaron inicialmente para Codex. Para cerrar el equivalente en Claude, revisá qué se puede reutilizar, adaptá el punto de entrada al cliente y probá el comportamiento desde una instalación real. Conservá las garantías de vista previa y confirmación por fila.
4. Usá datos sintéticos y, para escrituras en ClickUp, una lista de pruebas identificada y únicamente las filas aprobadas. Probá el caso sin acceso con una fuente inexistente o un entorno descartable. Registrá el bloqueo concreto si falta una cuenta, permisos o revisión obligatoria. Agotá los casos que puedas verificar con los accesos actuales.
5. Registrá resultados por ficha con escenario, esperado, observado, comando o sesión, evidencia, defectos corregidos y limitaciones. Cuando toda la matriz esté verificada, mové este encargo y el resultado al archivo histórico, actualizá su índice y ajustá los puntos de entrada en ambos archivos de instrucciones. Guardá continuidad en Engram y compactá al cerrar el trabajo.
6. La invocación de este encargo por el responsable autoriza avanzar con la entrega a GitHub. Seguí las reglas vigentes por ficha para pruebas, commit convencional, push, PR y merge a `main` tras los checks y revisiones. Confirmá el SHA fusionado y que el resultado quedó accesible. El merge requiere los permisos y aprobaciones que el repositorio aplique.

## Matriz de aceptación

| Ficha | Qué reproducir | Evidencia previa y criterio de cierre |
|---|---|---|
| F1 · Fundación | Ruta directa para una tarea acotada, checklist de fundación, aprobación, reutilización y reanudación en otra sesión | Integrada en PR #18. Demostrar los dos recorridos y la recuperación del avance en Claude Code. |
| F2 · Adaptación de arneses | En la fundación, comprobar que base de conocimiento, roadmap e instrucciones se ofrecen y se guardan o reutilizan según el flujo vigente | F2 quedó incluida en F1 en el roadmap publicado. Registrar estos casos por separado para tener evidencia de las cinco fichas. |
| F3 · ODD junto a SDD | Resolver un caso acotado con ODD y uno que necesite OpenSpec. Probar continuidad al escalar y al retomar | PR #19. Verificar el soporte funcional de Claude desde sesión nueva, con alcance y pruebas visibles. |
| F4 · ClickUp | Configurar una lista de pruebas, mostrar acciones y detalle, seleccionar una fila de dos, aprobar la escritura, actualizar una vinculada y consultar o sincronizar su avance | PR #20 y #22. Verificar en Claude que abrir menú o detalle no escribe y que el cambio confirmado afecta únicamente la fila aprobada. |
| F5 · Visualizaciones | Obtener Mermaid editable desde repo y ClickUp, elegir solo uno entre varios y avisar si falta la fuente. Ejercitar Discovery y Entrega | PR #21 y #22. La [evidencia publicada](https://github.com/Group-Active-IA/active-stack/blob/8562a85/openspec/changes/f5-transversal-diagrams/pilot-evidence.md) ya cubre lectura local y en vivo en Claude. Reutilizarla como línea base y comprobar que la entrega integrada conserva el comportamiento. Registrar por separado el render y la selección nativa si el cliente los expone. |

El cierre requiere evidencia reproducible para las cinco fichas. La ausencia de una capacidad en el cliente se registra con el comando o respuesta observada y su efecto sobre el criterio de aceptación.
