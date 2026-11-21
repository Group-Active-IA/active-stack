# Acciones de ClickUp

Mostrá el nombre y la frase breve de cada acción. Si el usuario elige **Ver detalle**, pedile que elija una acción si todavía no la indicó. Mostrá su ejemplo, qué lee, qué cambia y cuándo confirma. Volvé a ofrecer la elección después del detalle. Consultar un detalle no ejecuta la acción.

## Publicar fichas

**Frase breve:** Creá en ClickUp tareas nuevas a partir de las fichas que elijas.

- **Ejemplo:** «De estas dos fichas, subí solo la del informe».
- **Qué lee:** Las fichas de `CHANGES.md`, la lista de destino y sus tareas para evitar duplicados. Si hay responsables mapeados, verifica sus cuentas.
- **Qué cambia:** Crea solo las tareas nuevas elegidas y guarda el vínculo local con ClickUp.
- **Cuándo confirma:** Después de mostrar la lista y las fichas que propone crear. Elegir la acción abre la vista previa.
- **Herramientas de sesión:** `clickup_get_list`, `clickup_filter_tasks`, `clickup_get_task` y `clickup_create_task`. Para validar responsables mapeados, verificá que exista el mecanismo oficial de lectura que requiere el flujo.

## Actualizar una tarea vinculada

**Frase breve:** Pasá los cambios de una ficha a la tarea de ClickUp que ya tiene vinculada.

- **Ejemplo:** «Agregué un requisito a la ficha. Mostrame cómo quedaría su tarea».
- **Qué lee:** La ficha, el vínculo guardado y el título y la descripción actuales de la tarea.
- **Qué cambia:** Actualiza el título y la parte de la descripción que administra Active Stack. Conserva el resto de la descripción, el estado y los responsables.
- **Cuándo confirma:** Después de mostrar el título y la descripción resultantes, antes de enviarlos a ClickUp.
- **Herramientas de sesión:** `clickup_get_list`, `clickup_filter_tasks`, `clickup_get_task` y `clickup_update_task`.

## Ver y sincronizar avance

**Frase breve:** Mirá en qué estado está la tarea y quién la tiene asignada. Guardá ese dato en el proyecto si querés actualizar el seguimiento.

- **Ejemplo:** «Alguien puso el informe en En curso. Quiero ver cómo quedó».
- **Qué lee:** El vínculo de la ficha, el estado y los responsables actuales en ClickUp.
- **Qué cambia:** Consultar solo muestra los datos. Si pedís sincronizar, guarda esos datos en el seguimiento local. ClickUp queda tal como estaba.
- **Cuándo confirma:** La consulta no pide confirmación. Si la elección fue ambigua, preguntá si también quiere guardar el resultado localmente.
- **Herramientas de sesión:** `clickup_get_list` y `clickup_get_task`.

## Configurar ClickUp

**Frase breve:** Revisá qué falta para conectar este proyecto con la lista correcta de ClickUp.

- **Ejemplo:** «Todavía no conecté ClickUp. Ayudame a elegir la lista de pruebas».
- **Qué lee:** El modo del proyecto, la lista configurada y el diagnóstico de conexión.
- **Qué cambia:** El diagnóstico solo lee. Si elegís conectar o cambiar el destino, se prepara el cambio de configuración.
- **Cuándo confirma:** Antes de registrar la conexión o guardar un destino. La autorización de ClickUp la completa el usuario.

## Seguir con el archivo local

**Frase breve:** Revisá las fichas del proyecto en `CHANGES.md` sin usar ClickUp.

- **Ejemplo:** «Mostrame qué fichas hay y de cuáles depende el informe».
- **Qué lee:** `CHANGES.md` y, si hace falta, las referencias locales de la ficha.
- **Qué cambia:** La consulta solo muestra datos. Una edición de Markdown requiere un pedido concreto.
- **Cuándo confirma:** Elegir esta opción no cambia archivos. Acordá cualquier edición según el flujo local del proyecto.

## Entrega en equipo

Para proyectos con `team.json`, el [flujo de equipo](team-roadmap.md) define
tres acciones adicionales y mantiene la asignación conciliada separada de las
lecturas de avance.

- **Tomar una ficha.** Lee el responsable efectivo y previsualiza asignarla al
  integrante identificado. Confirma la lista completa de responsables que
  quedará tras el cambio.
- **Vincular un PR.** Lee el PR y la tarea vinculada. Confirma el enlace que se
  incorporará al bloque de Active Stack, conservando el resto de la tarea.
- **Cerrar una ficha.** Lee el merge y el estado actual de la tarea y muestra el
  estado de cierre, el PR, el SHA y el traspaso. Confirma esa fila antes de
  actualizarla. El cierre local sigue disponible desde la ficha y el PR.
