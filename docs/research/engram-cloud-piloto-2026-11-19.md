# R-01 · Piloto de Engram Cloud

**Edición pública de la investigación del 19 de noviembre de 2026.** Las sesiones de Codex sobre macOS probaron el etiquetado y la replicación de observaciones sintéticas. La variante B conservó el contexto de rama en la ruta `mem_save`. Persisten límites en otras cinco rutas de escritura y en la separación de algunas ramas bajo la regla original. Claude y Windows quedan pendientes.

El entorno empleó Codex CLI 0.169.0-alpha.11.1, Engram CLI/MCP 3.6.2, plugin 0.1.5 y servidor de pruebas v2.0.0. Los participantes y los destinos del servicio están anonimizados. Los IDs de sesión, llamada y sincronización de las trazas son alias; las fechas técnicas de los eventos se desplazaron en conjunto. Se conservaron contadores, intervalos, relaciones entre eventos e IDs numéricos locales de bases sintéticas. Los endpoints, capturas de administración y registros de acceso originales quedan en el expediente interno. [Documentación del cliente](https://github.com/Gentleman-Programming/engram/blob/v3.6.2/DOCS.md)

## Resultados por pregunta

| Prueba | Veredicto | Resultado y alcance |
|---|---|---|
| Q1 · Autosync y latencia | Verificado A ↔ B | En cada dirección, el receptor vio los dos registros sin sincronización manual. En los ensayos limpios: project ≤30,14 s, personal ≤30,10 s. |
| Q2 · Claves por rama | Verificada localmente | El MCP separó K y branch/slug/K. Ver la sonda local. |
| Q3 · Coexistencia | Hooks original y A fallan en Codex. B pasa tanto al cambiar de rama como en dos worktrees simultáneos. Claude pendiente | Con los dos hooks reescribiendo, la llamada conservó las etiquetas y omitió el session_id explícito aportado por el plugin. Los registros del padre sí quedaron con su ID correcto en esta base aislada. Ver las trazas anonimizadas de coexistencia. |
| Q4 · Colisión al promover | Colisión reproducida localmente | Al promover por ID sobre K, dos filas quedaron con la misma clave. Un guardado posterior con K actualizó #1 en el ensayo original y #2 en la repetición de QA. El destino no fue estable. |
| Q5 · Organización | Comprobada para las versiones del piloto | El cliente 3.6.2 omitió org en las mutaciones. Tanto las búsquedas MCP como el pull del servidor v2.0.0 devolvieron el marcador con un filtro incompatible. |
| Q6 · Subagentes y cambio de rama | Etiquetado observado en Codex. Claude pendiente | Guardados tras switch y checkout, y escrituras verificadas en rollouts de padre e hijo. El proxy B etiqueta ambos y recibe el ID del padre en los dos casos. |
| Q7 · Herramientas de escritura | Inventario comprobado en MCP 3.6.2 | Seis de 27 herramientas crean o editan texto de observación. B etiqueta mem_save. Las otras cinco rutas y el historial separado de prompts están documentados. |
| Q8 · Nombres de herramientas y matcher | Verificado por registro en Codex | Se observaron el alias directo y una llamada con el prefijo alternativo en el cliente. La evidencia de cada invocación y del handler consta en el anexo. |
| Q9 · Permisos al cierre | Parcial en Codex | Git devolvió la rama en estricto y balanceado con código 0 y cero pedidos de aprobación. El checkout carece del comando memory branch de Fase 4. El cierre completo queda pendiente. |
| Q10 · Clones y worktrees | Verificada localmente | La configuración seleccionó el proyecto previsto en el repositorio, el clon y el worktree. Ver evidencia local. |
| Q11 · Sobrecosto | Medida en tres pares reales de Codex. Aceptación pendiente de DA-02 | Entrada acumulada con variante −0,83 % y entrada sin caché +17.478 tokens. Las seis respuestas fueron correctas y cada sesión hizo tres llamadas MCP. Ver mediciones por corrida. |
| Q12 · Acceso al proyecto | Verificación parcial y discrepancia documentada | Se comprobaron respuestas diferenciadas según el permiso del usuario. La configuración del servicio, el diagnóstico operativo y la evidencia administrativa se conservan en el informe interno. |
| Q13 · Ámbito y autor | Verificada en el piloto | Los datos con ámbito personal se replicaron en ambos sentidos. La búsqueda no mostró campo de autor; el panel agregado mostró 2 y 1 unidades de contribución, sin atribución individual por observación. Las capturas originales se reservan para revisión interna. |
| Q14 · Directorio recibido | Verificado para el proxy B en Codex | Sin fijar cwd en el registro MCP, cada proceso del proxy recibió el directorio del worktree de su sesión. El observador de PreToolUse confirmó cwd en padre e hijo del mismo checkout. Faltan esa captura en dos worktrees y la prueba en Claude. |
| Q15 · Manejo de error y timeout | B preservó el guardado en dos fallos internos. El guardado falló al terminar el proxy | Ante marcador inválido o timeout de Git, Engram guardó la entrada original y B avisó en stderr. Al terminar B entre llamadas, Codex registró Transport closed en la segunda y SQLite conservó solo la primera. El timeout externo de PreToolUse queda sin ejecutar porque el ensayo requería nueva confianza. [Control](r01-prototipo/controles-codex.md) |

## Matriz de aceptación por agente

La columna de Codex indica el mecanismo probado en cada escenario. Los controles de V6, V9 y V11 usaron el cliente y el proxy en perfiles aislados sin plugin. La coexistencia con el plugin se midió en V10.

La columna de Claude Code figura como «No ejecutado» en todos los escenarios por falta del cliente. Los resultados observados en Codex corresponden a ese cliente.

| Escenario | Claude Code | Codex |
|---|---|---|
| V1 · Guardado desde una rama | No ejecutado | Verificado para mem_save |
| V2 · Cambio de rama en la sesión | No ejecutado | Verificado con switch y checkout |
| V3 · Regreso a main | No ejecutado | Verificado. Solo Author en el contenido |
| V4 · Worktree por ficha | No ejecutado | B pasó con dos Codex simultáneos, una base, dos ramas y la misma clave de entrada. Las revisiones avanzaron por separado. |
| V5 · Guardado de subagente | No ejecutado | B pasó para etiquetado: el rollout hijo confirmó #19. La sesión que recibió y conservó el proxy fue la del padre |
| V6 · Otras vías de escritura | No ejecutado | Cobertura documentada. Summary, finalize y capture_passive crearon observaciones sin etiqueta. Save_prompt escribió un prompt separado. B etiqueta mem_save. |
| V7 · Proyecto sin marcador | No ejecutado | Verificado en el control con cuatro guardados |
| V8 · Hook ausente o sin confiar | No ejecutado | Parcial. El runtime detectó siete hooks habilitados sin confiar y la interfaz pidió revisión antes del turno. Falta la comprobación desde verify. |
| V9 · Fallo o espera del hook | No ejecutado | Marcador inválido y timeout interno de Git: guardado confirmado con aviso. Terminación del proxy B: segunda llamada fallida y sin observación. Timeout externo del hook PreToolUse: no ejecutado. |
| V10 · Coexistencia con el hook de Engram | No ejecutado | Falla con los dos hooks. El proxy B y el hook oficial conservaron los campos en 4/4 guardados |
| V11 · Detached HEAD y nombres complejos | No ejecutado | Detached, rama con barra y nombres largos con hashes distintos dieron el formato previsto. Dos ramas cortas con el mismo slug y clave base actualizaron la misma fila #1 (revisión 1 a 2). B original falla en ese caso. Una variante aislada con 12 cifras hexadecimales de SHA-256 del nombre Git completo separó `feat/una` y `feat-una` en dos worktrees simultáneos de Codex: IDs 1 y 2, cada uno con dos revisiones. En una base temporal aparte, una clave anterior de origen conocido se cambió por API sin perder su ID ni el historial. La fila ya compartida se conservó intacta y cada rama continuó con un ID nuevo. [Evidencia de la variante](r01-prototipo/v11-sufijo-en-dos-worktrees.md) |
| V12 · Windows y otro SO | No ejecutado | Se ejecutó macOS x86_64. Windows queda pendiente de ejecución nativa. |
| V13 · Lectura posterior desde main | Diferido | Pasó en una corrida con modelo real: mem_search mostró las dos decisiones incompatibles y el modelo eligió el puerto 4381 de la compartida. |


## Evidencia y controles

- [Sonda local MCP](r01-prototipo/sonda_engram.py) y [sonda de sincronización para un entorno provisto por quien ejecuta](r01-prototipo/sonda_cloud.py).
- [Organización, herramientas y QA](r01-prototipo/alcance-q5-q7-q8.md), [controles de Codex](r01-prototipo/controles-codex.md) y [variante con sufijo e historial](r01-prototipo/v11-sufijo-en-dos-worktrees.md).
- Q11 midió tres pares reales con `gpt-6.2-sol`, esfuerzo bajo. Las tres diferencias relativas de entrada fueron +5,91 %, +2,72 % y −9,70 %. Las sumas de tokens y la separación entre entrada total, entrada en caché y entrada sin caché se preservan en las mediciones. La aceptación queda pendiente de fijar el umbral.
- La siguiente decisión corresponde a la revisión del piloto: cobertura de la variante B, criterio de sobrecosto y ensayos pendientes en Claude y Windows.

### Trazas anonimizadas

- [Ajuste a](r01-prototipo/codex-ajuste-a.json)
- [Autoria control identico](r01-prototipo/codex-autoria-control-identico.json)
- [Autoria pre post](r01-prototipo/codex-autoria-pre-post.json)
- [Coexistencia](r01-prototipo/codex-coexistencia.json)
- [Diagnostico native](r01-prototipo/codex-diagnostico-native.json)
- [Diagnostico observador](r01-prototipo/codex-diagnostico-observador.json)
- [Proxy b subagente](r01-prototipo/codex-proxy-b-subagente.json)
- [Proxy b worktrees](r01-prototipo/codex-proxy-b-worktrees.json)
- [Proxy b](r01-prototipo/codex-proxy-b.json)
- [Q11 tokens](r01-prototipo/codex-q11-tokens.json)
- [V13 lectura main](r01-prototipo/codex-v13-lectura-main.json)

### Evidencia TDD del prototipo

| Tarea | Archivo de prueba | Capa | Línea base | RED | GREEN | Triangulación | Refactor |
|---|---|---|---|---|---|---|---|
| Reescritura de eventos de guardado | `r01-prototipo/test_hook.py` | Contrato de proceso con Git temporal | Archivo nuevo | 1 prueba, 1 error al faltar el hook | 1/1 aprobada | 5 pruebas, 3 fallos y 1 error antes de generalizar. Después pasaron 5/5. | Sin paso de refactor registrado |
| Ajuste A · Tres campos en updatedInput | `r01-prototipo/test_ajuste_a.py` | Proceso y repositorio Git temporal | 5/5 originales | Falló el primer caso al faltar la variante | 1/1 | Falló el caso sin clave; después 7/7 con la suite original | Sin refactor posterior |
| B · Etiquetado y reenvío por stdio | `r01-prototipo/test_proxy_b.py` | Proceso, upstream de prueba y Git temporal | 7/7 originales | Error al faltar el proxy | 1/1 | Los tres casos adicionales fallaron antes de implementarse: paso transparente, traza acotada y continuación fuera de Git. Suite final 11/11 | Sin refactor posterior |
| Autoría PRE/POST · extracción por llamada | `r01-prototipo/test_observador_autoria.py` | Contrato de evento sintético | Suite previa 11/11 | Error al faltar el extractor | 1/1 | Fallaron 2 casos adicionales antes de acotar los campos y excluir IDs ajenos. Después 3/3; suite de laboratorio 14/14 | Sin refactor posterior |
