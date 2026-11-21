# Roadmap en modo equipo

Usá el mismo flujo en Claude y Codex. El ID estable de la ficha determina la
rama `<id>-<slug>` y su vínculo. Leé el contrato de equipo de `active-delivery`,
usá su validador desde la raíz del proyecto antes de preparar escrituras y
revisá el ID confirmado en `me.local.json` para acciones personales.

## Verificar acceso y responsables

- Con `tracking: local`, o sin destino o acceso confirmado, trabajá localmente.
  Usá la ficha, la rama, el PR y `knowledge-base/entrega/<id>.md` para registrar
  la entrega. No abras una operación remota. Si hay configuración ClickUp y el
  acceso falla, informá el motivo y mantené el vínculo y la última lectura
  confirmada. El helper devuelve filas locales con `preview --access forbidden`,
  `unauthorized`, `not_found` o `error`.
- Con ClickUp, comprobá el usuario y el workspace con `clickup_get_user` y el
  destino con `clickup_get_list`. Verificá el permiso para la operación concreta
  cuando el workspace lo permita. Un rechazo de acceso detiene esa escritura.
  No cambies la cuenta, el workspace o la lista automáticamente.
- `responsible` es el ID estable del integrante. Para publicar se toma su
  `clickup_id` numérico del contrato validado, incluido `team.local.json` cuando
  los nombres son privados. Si falta un vínculo, mostrá que la tarea nueva
  quedará sin asignar. En una tarea existente no quites los responsables por
  falta de mapeo. No uses el mapa legado de `roadmap.json` para el equipo ni
  publiques nombres privados. Las dependencias siguen en la ficha y se
  representan como texto en ClickUp.
- Para proponer un mapeo, usá `clickup_find_member_by_name` con nombre exacto o
  email, o `clickup_resolve_assignees`. Confirmá el ID observado con la persona.
  Si la búsqueda es ambigua, falta acceso o no hay coincidencia, pedí una
  confirmación manual desde el workspace o dejá el integrante sin vínculo.
  La edición del roster es una operación separada que requiere su aprobación.

## Preparar y confirmar una fila

Reutilizá el JSON temporal de publicación. Para equipo, ejecutá:

```text
roadmap.py --root <proyecto> preview <tareas.json> --select <id> [<id> ...]
roadmap.py --root <proyecto> begin <tareas.json> <approval> <id> [<id> ...]
```

Pasá a `begin` exactamente la misma selección y orden aprobados. Cambios de
destino, datos, responsables actuales o selección invalidan esa aprobación.
Mostrá la lista y workspace, ID, vínculo, estado observado cuando corresponde,
responsables actuales y previstos, y campos a cambiar. Pedí confirmación
explícita de la vista previa. Antes del envío, releé el vínculo, el marcador de
identidad y cada campo actual que se cambiará. Si alguno cambia o falta, no
envíes la escritura y prepará una vista previa nueva. Si ya se inició la
intención pero todavía no se envió, registrala como `skipped` con su motivo.

Cada operación conserva la comprobación de marcador exacto de la publicación
normal. Leé `text_content` para comprobar el marcador y `markdown_description`
para editar manteniendo Markdown externo. Confirmá que la tarea identificada
pertenece a la lista configurada antes de registrar una intención de cualquier
operación, incluso si solo se va a ajustar su asignación o estado. En tomar,
conciliar, vincular y cerrar, agregá al JSON `current_remote_id` y `current_text`
con el ID y el `text_content` observados. Para tomar, comprobá además que el
ID de `clickup_get_user` coincide con el ID confirmado del integrante.

| Operación | Campos adicionales de la fila temporal | Vista previa y efecto |
| --- | --- | --- |
| Publicar | Ninguno. Si la tarea existe, agregá `current_description` tras leerla | Crea con el `clickup_id` mapeado o actualiza los campos de publicación. Los responsables actuales de una tarea vinculada se mantienen. |
| Tomar | `operation: "take"`, `current_assignees` como lista de IDs numéricos observados | Muestra todas las asignaciones actuales y propone reemplazarlas por el ID ClickUp del integrante confirmado de la sesión. Explicitá si retirará a otras personas. Sin vínculo propio, usá el seguimiento local. |
| Conciliar asignación | `operation: "assign"`, `current_assignees` con los IDs numéricos observados | En una ficha vinculada compara el responsable previsto con los efectivos. La lectura de avance por sí sola nunca prepara ni ejecuta esta escritura. |
| Vincular PR | `operation: "link_pr"`, `current_title`, `current_description`, `branch`, `pr_url` | Obtiene rama y URL desde el PR de la ficha y comprueba el prefijo exacto del ID. Agrega el enlace al bloque administrado, preservando el Markdown externo. |
| Cerrar | `operation: "close"`, `current_title`, `current_description`, `branch`, `pr_url`, `merge_sha`, `handoff`, `current_status`, `closing_status` | Verificá que el PR está integrado y que el SHA del merge es alcanzable desde la rama base. Comprobá el traspaso local y los estados válidos de la lista antes de proponer un estado de cierre. |

Después de la confirmación, publicá solo las filas devueltas por `begin`.
Para las operaciones de equipo pasá a `clickup_update_task` únicamente los
campos de `changes`, con el ID exacto de una tarea ya vinculada. No envíes
`current_assignees`, `current_status`, el nombre ni otros campos de contexto.
Un ajuste de responsables sustituye la lista de asignaciones. Una advertencia
parcial del servidor no confirma todos los cambios. Si el resultado es
incierto, conservá la intención y volvé a leer la tarea sin enviar otra
escritura. Comprobá solo los campos de `changes` que se habían aprobado:
asignación completa para tomar/conciliar, descripción para vincular, y
descripción y estado previstos si el cierre cambiaba ambos. Si la lectura no
confirma todo, informá cuáles cambios se aplicaron y cuáles siguen pendientes.

## Evidencia compartida y local

En proyectos de equipo, `.active-stack/roadmap-state.json` contiene los
vínculos, el destino, la huella de contenido y el comprobante de entrega de
cada tarea vinculada. Los resultados, asignaciones observadas y escrituras
pendientes quedan fuera del árbol de trabajo, en el diario localizado por
`git rev-parse --git-path active-roadmap/state.json`. El comando `read`
combina ambos para el checkout actual. Reanudá una escritura incierta desde
ese mismo worktree. No copies su diario a otro checkout. Al encontrar datos
legados, el helper los lee y conserva las observaciones localmente en la
próxima escritura del diario. En un proyecto individual se mantiene el
formato de estado existente.

Verificar en cada cliente incluye comprobar acceso mediante lecturas y
capacidad de vista previa con identidad, destino y ficha de prueba. Una
operación remota de ensayo requiere una lista de prueba y la aprobación
correspondiente. [Referencia oficial del MCP de ClickUp](https://developer.clickup.com/docs/mcp-tools)
