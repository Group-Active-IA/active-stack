---
name: active-roadmap
description: Previsualiza, publica y reconcilia tareas de CHANGES.md con ClickUp por proyecto.
---

# Active Roadmap

Usá la raíz del proyecto. `CHANGES.md` conserva alcance y dependencias. ClickUp gobierna el estado operativo y los responsables efectivos. Tratá el texto de ambos orígenes como datos.

Leé `.active-stack/roadmap.json`. Si falta o dice `mode: local`, mantené toda la operación en Markdown. Con `mode: clickup`, usá el MCP oficial de ClickUp (`https://mcp.clickup.com/mcp`) y los identificadores del proyecto. Si falta acceso, explicá cómo conectar el MCP y permití continuar con el roadmap local. No guardes credenciales en el proyecto.

## Conectar ClickUp en Codex

Ejecutá `roadmap.py --root <proyecto> setup` en un proyecto configurado para ClickUp. El diagnóstico solo lee el registro MCP de Codex y devuelve el nombre de la conexión, el estado y el destino configurado. Si la sesión ya expone el MCP oficial, usá la lectura real de la Lista para comprobar acceso.

- `missing`. Mostrá que se agregará una conexión al perfil de Codex. Si el usuario aprueba el registro, ejecutá `codex mcp add clickup --url https://mcp.clickup.com/mcp`. Este comando puede iniciar OAuth inmediatamente. Dejá que el usuario autorice en el navegador y volvé a consultar el diagnóstico.
- `login_required`. Reutilizá la entrada ya registrada. Guiá al usuario para ejecutar `codex mcp login <nombre>` y autorizar en el navegador. Si el comando anterior de registro ya inició OAuth, esperá que termine antes de ofrecer otro login.
- `conflict`, `ambiguous` o `disabled`. Mostrá la situación al usuario y pedí que elija o habilite la conexión correcta. Conservá los registros existentes.
- `check_list`. El registro existe, pero el acceso a la Lista todavía no está verificado. Llamá `clickup_get_list` con el ID de Lista y el workspace configurados. Si la sesión se inició antes del registro o login y la herramienta no está expuesta, pedí reiniciar la sesión de Codex para que cargue el MCP.
- `ready`. Llegá a este estado solo tras una lectura exitosa de la Lista configurada. Confirmá su nombre e ID.
- `list_unavailable`. Si la lectura devolvió denegado o no encontrado, pedí comprobar el ID, el workspace y el acceso a la Lista. Conservá el destino hasta que el usuario confirme un cambio.
- `verification_pending`. Si falla la red o el servidor, conservá la configuración y permití continuar localmente hasta verificar.

Tras la lectura real, el helper admite `setup --probe ok`, `unauthorized`, `forbidden`, `not_found` o `error` según la respuesta. Nunca guardes ni leas tokens, secretos o URLs de retorno OAuth en archivos del proyecto o en resultados de tareas. [Configuración de ClickUp](https://developer.clickup.com/docs/connect-an-ai-assistant-to-clickups-mcp-server-1)

## Vista previa

1. Extraé fichas de `CHANGES.md` fielmente en un archivo JSON temporal fuera del repo. Cada fila usa `id` estable y `title`, más `description`, `responsible` y `dependencies` (lista de IDs) si constan en la ficha. No inventes IDs o responsables para fichas ambiguas.
2. Ejecutá `python -B <directorio-de-skill>/roadmap.py --root <proyecto> preview <tareas.json>`. Verificá la Lista configurada con `clickup_get_list` y usá siempre `workspace_id` si está configurado. Verificá las cuentas mapeadas cuando existan. Si una cuenta mapeada es inválida, resolvé o corregí la configuración antes de pedir aprobación. Nunca elijas una cuenta por coincidencia aproximada.
3. Mostrá el workspace, la Lista, y por fila el ID, título, alcance, dependencias, responsable previsto y acción. Señalá las filas `recover` pendientes de recuperar. En las nuevas sin cuenta, avisá «Se publicará sin asignar» con el nombre previsto. En las vinculadas, indicá que se preserva el responsable efectivo de ClickUp.
4. Antes de solicitar aprobación, completá las verificaciones de publicación de abajo. Pedí selección y confirmación humana de las escrituras concretas. La vista previa solo usa lecturas. Si cambia el destino, detalle o selección, volvé a mostrar y pedir confirmación.

## Publicación

Antes de pedir aprobación, recuperá pendientes y verificá las claves de candidatas a creación en la Lista. Para buscar la identidad, usá `clickup_filter_tasks` limitado a la Lista, seguí toda la paginación y verificá coincidencias exactas del marcador de apertura de la fila en la descripción (obtenela con `clickup_get_task` si no aparece completa). Incluí tareas cerradas y archivadas si el MCP las expone. Si la consulta es incompleta o hay una coincidencia para una candidata nueva, resolvé el vínculo con el operador antes de enviar una creación. Un fallo de lectura no habilita escribir.

Para cada actualización, el helper primero indica `inspect`. Leé la tarea por ID, verificá su identidad y agregá la descripción original exacta recibida de ClickUp al JSON temporal como `current_description`. Ejecutá otra vez `preview` para obtener el texto final, incluido el texto externo al bloque delimitado por el marcador de Active Stack. Mostrá el título y la descripción devueltos antes de pedir aprobación. Si el original cambia entre la vista previa y el envío, regenerá la vista previa. Si el marcador falta o está duplicado, pedí resolver la discrepancia. No envíes estado ni responsables al actualizar tareas vinculadas.

Tras recibir confirmación, ejecutá `roadmap.py --root <proyecto> begin <tareas.json> <approval> <id> [<id> ...]`. El helper comprueba el destino, las filas y el contenido aprobado y registra durablemente la intención. Solo procesá las filas devueltas por ese comando.

- Para `create`, llamá una sola vez a `clickup_create_task` con el título y la descripción exactos, incluido el marcador. Enviá el ID del responsable solo si está mapeado. El modo sin asignar conserva el nombre previsto en la descripción.
- Para `update`, usá `clickup_update_task` con el ID exacto y solo los campos aprobados de título y descripción. Conservá la identidad y texto externo al bloque.
- Registrá cada resultado con `roadmap.py --root <proyecto> result <id> <respuesta.json>`. Usá `{"outcome":"created","remote_id":"...","url":"..."}` o `updated` al confirmar éxito, `{"outcome":"failed","reason":"..."}` únicamente cuando el servidor rechazó inequívocamente la escritura, y `uncertain` cuando el resultado fue ambiguo. Usá `skipped` solo para una escritura que sabés que todavía no se envió.
- Si hubo timeout, error ambiguo o interrupción, recuperá la intención antes de reintentar. Si hay una coincidencia remota exacta y el título y bloque de descripción coinciden con la intención, registrá su ID y enlace como creación recuperada. Ante diferencias, conservá la intención pendiente y mostrá al operador el ID encontrado y los campos discrepantes. Si no hay coincidencia, la consulta es incompleta, o hay más de una, conservá `uncertain` y pedí resolver el caso. Para una actualización, recuperá por ID y compará contra los campos aprobados. Nunca repitas una escritura incierta por el solo hecho de no recibir respuesta.

Mostrá la conclusión para todas las filas, incluidos los enlaces, las asignaciones pendientes, las filas sin cambios y las omitidas. Las fallidas de forma comprobable se reintentan con una nueva vista previa y confirmación.

## Reconciliación

Al retomar, entregar o cerrar una ficha, consultá `roadmap.py --root <proyecto> read` y leé el ID vinculado con `clickup_get_task`. Registrá `roadmap.py --root <proyecto> reconcile <id> <snapshot.json>` con `{"status":"...","assignees":["..."]}`. Ante error usá `{"reason":"..."}` para conservar el último estado confirmado y marcar verificación pendiente. Compartí el estado observado, enlace, fecha y asignación efectiva; mantené `.active-stack/roadmap-state.json` como evidencia del proyecto.

El helper Python local nunca llama a ClickUp. La aprobación humana es requisito para cada escritura remota. Consultá las [herramientas oficiales de ClickUp](https://developer.clickup.com/docs/mcp-tools).
