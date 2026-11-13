# Informe original: propose, intento 1, plan-reviewer

Autor: subagente plan-reviewer (Claude Code, despachado por el orquestador el 2026-11-06). Identidad del ejecutor sin registro de despacho propio: SUPOSICIÓN.
Revisión de código: rama `claude-native-opsx`, 97d7f6c. Los artefactos están ignorados por git (`.gitignore:36`, confirmado con `git check-ignore -v`), así que la revisión se ata a hashes sha256 (primeros 16 hex):
proposal.md fba7ec864b3d66a2 · design.md 60be3c6af255a3cf · tasks.md 43ac8b7802b16666
- specs/claude-coordinator-agent/spec.md 5ea28253d1645a7f
- specs/claude-instruction-layers/spec.md 1786accdd222da75
- specs/claude-orchestrator-migration/spec.md f92188cf909c61c4
- informes/propose-1-planner.md 522df88298c8779b

## Veredicto: VUELVE AL PLANNER

Un BLOQUEANTE y ocho MAYOR. Las citas de código del plan son correctas (C1 a C5 verificadas), la clasificación ALTO está justificada y `## Aprobación` está vacía. Los defectos son de diseño y verificabilidad, no de citas.

## Hallazgos ordenados por severidad

### B1 (BLOQUEANTE) El agente de archivo queda disponible como subagente de la sesión normal: contradice "sólo explícita"
- Fuente: spec `claude-coordinator-agent`, Requirement "Invocación explícita activa el coordinador" ("MUST activarse sólo cuando el usuario lo invoca de forma explícita"); design D1/D3; ficha (migrar "cualquier activación automática").
- Observación: un archivo en `~/.claude/agents/<id>.md` no sólo sirve para `--agent`; también aparece como tipo de subagente para la herramienta Agent de cualquier sesión. El propio D0 caso C lo demuestra para agentes `--agents` JSON: la sesión normal lanzó un agente definido con la herramienta Agent. Claude Code además delega por `description` sin pedido explícito (documentación pública, no verificado en 2.1.318: SUPOSICIÓN). Resultado posible: la sesión normal lanza al coordinador como subagente; ahí el coordinador corre sin poder delegar (los subagentes no lanzan subagentes según la documentación pública, SUPOSICIÓN) y su regla "NEVER do apply or propose work inline" (`sdd-orchestrator.md:208`) lo deja sin salida.
- Impacto: se reintroduce una activación implícita por otra vía y el criterio de cierre 2 deja de ser verificable como "sólo explícita". Ninguna tarea de la sección 1 lo investiga y ningún escenario lo cubre.
- Criterio incumplido: ficha ("cualquier activación automática"), requisito propio de la spec.
- Corrección verificable: (1) tarea 1.x que verifique en 2.1.318 si un agente de archivo aparece en la lista de subagentes de la sesión normal, si se delega solo y si un subagente puede usar Agent; (2) escenario "la sesión normal no lanza el coordinador como subagente sin pedido" con su mitigación (description que lo excluya, campo de frontmatter si existe, o cambio de diseño); (3) si no hay forma de excluirlo, llevarlo como decisión de grupo nueva (G9) antes de compartir como cerrada la vía D1.

### M1 (MAYOR) Las specs cierran como MUST decisiones que design declara abiertas
- Fuente: design.md "No se cierran acá" (G1-G8) frente a: spec `claude-coordinator-agent` "MUST NOT fijar el setting `agent`" (G2), "Fallback ... skill o comando" (G2), escenario "Especialista de apply recibe el módulo TDD" escrito para G3 opción (a), spec de permisos (G6), `design-mockups` intacta (G7). El planner lo reconoce sólo en su informe ("Decidido sin el grupo"), no en los artefactos que lee el equipo.
- Impacto: las siete personas leen MUST normativos y pueden tomar como decididas cosas que se les pide decidir; retrabajo de specs si eligen otra opción.
- Criterio incumplido: decisiones de grupo no cerradas unilateralmente (ficha).
- Corrección verificable: cada requisito o escenario que depende de una G lleva una marca visible "(depende de Gn, recomendación)", y design o proposal agrega una línea explícita "las specs están escritas sobre las recomendaciones". Chequeo: `rg -n "depende de G" specs/` devuelve al menos un resultado por G1, G2, G3, G6 y G7.

### M2 (MAYOR) Versión insuficiente o desconocida contradice verify y deja sin definir la instalación nueva
- Fuente: spec `claude-coordinator-agent` "Versión mínima" (no purgar el legado, "entrega previa sin cambios") frente a spec `claude-orchestrator-migration` "Verify refleja la entrega por capas" (MUST fallar si falta el agente o queda `active-stack:sdd-orchestrator`).
- Observación (razonada, sin código): con versión insuficiente el legado queda y no hay agente, así que verify falla siempre. En instalación nueva sin entrega previa (o sin binario `claude`, caso típico de CI, Docker de `e2e/` o instalar antes que Claude) la spec no dice qué se escribe: ¿legado completo (vuelve la fuga), sólo universal, o nada?
- Impacto: verify rojo permanente en entornos legítimos; comportamiento indefinido en el caso más común de primera instalación.
- Criterio incumplido: criterios verificables y sin contradicción.
- Corrección verificable: escenarios para (a) versión insuficiente con legado previo, (b) versión insuficiente o desconocida sin entrega previa, (c) binario ausente; y el resultado esperado de verify en cada uno (por ejemplo un check que informe "modo legado por versión" en lugar de fallar). Tareas 6.4 y 7.1 los cubren con RED.

### M3 (MAYOR) Target Project: uninstall y verify no conocen el target
- Fuente: `internal/uninstall/steps.go:150-170` y `internal/verify/harness_checks.go:98-130` usan `a.InstructionsPath(homeDir)` sin target; `rg -i "project|target"` en `internal/uninstall` y `internal/verify` (sin tests) no devuelve soporte de Project. Install sí lo tiene (`internal/install/plan.go:98-101`, `projectTargetAdapter` en `:556-580`).
- Impacto: los escenarios "Install Project crea el agente", 8.2 ("repetir para target Project ... después de restore") y "Uninstall sin restore limpia sólo lo propio" no tienen camino de código para Project en uninstall/verify. El plan los da por existentes.
- Corrección verificable: declarar el alcance (Project sólo en install, o agregar target a verify/uninstall con tareas RED propias) y ajustar 8.2. Además: `projectTargetAdapter` (`plan.go:556`) falta en la lista de consumidores de la tarea 3.1.
- Observación lateral (preexistente, no del change): `config.Install` arma `snapshotDir` con `homeDir` (`install.go:36-37`), que en Project vale `effectiveBase` = raíz del proyecto; contradice el comentario de `plan.go:138-141` ("must not leave its backup mixed into the project folder"). El inyector nuevo no debería copiar ese patrón: D5 tiene que nombrar dónde vive el manifiesto.

### M4 (MAYOR) El oráculo de las pruebas con testigo depende de lo que el modelo decida repetir
- Fuente: specs `claude-instruction-layers` y `claude-coordinator-agent` ("la respuesta contiene ... y no contiene ..."); tasks 8.1; design Risks.
- Observación: una aserción de ausencia basada en que el modelo no repite un testigo no prueba ausencia (el modelo puede omitirlo u obedecer una instrucción del propio contexto). EVIDENCIA de que existe un oráculo determinista candidato: en transcripciones de Claude Code 2.1.318 de esta máquina, el contenido de `~/.claude/CLAUDE.md` queda registrado como entradas `"attachment":{"type":"instructions"}` (9 entradas con un texto conocido del CLAUDE.md global) y existe `"attachment":{"type":"prompt_snapshot"}` (20 entradas); `claude --help` documenta `--system-prompt-snapshot`. No verifiqué si el cuerpo de un agente aparece en `prompt_snapshot`: SUPOSICIÓN.
- Corrección verificable: tarea 1.x que valide un oráculo determinista (buscar el testigo en las entradas registradas de la transcripción, no en la respuesta) y reescribir los THEN para usarlo; mantener la respuesta del modelo sólo como control positivo. Escenarios de ausencia siempre con un control positivo en la misma corrida.

### M5 (MAYOR) Prueba con HOME temporal sin plan de autenticación ni manejo de secretos
- Fuente: tasks 8.1 y 8.2 ("instala en un HOME temporal ... Nunca usar el `~/.claude` real").
- Observación: con `HOME` temporal Claude Code no encuentra credenciales; D0 se corrió con el `~/.claude` real. `e2e/e2e_test.sh` sólo usa el binario `active-stack` (`--dry-run`, `--home "$SANDBOX"`), no `claude -p`, y corre en Docker.
- Impacto: 8.1 no es ejecutable tal como está, y la solución improvisada probable (copiar el archivo de credenciales al HOME temporal) expone un secreto en un directorio de prueba.
- Corrección verificable: la tarea nombra el mecanismo (por ejemplo variable de configuración de directorio o clave por variable de entorno, a verificar en 1.x), prohíbe copiar credenciales a disco, exige borrar el HOME temporal y registrar sólo nombres de variables, nunca valores.

### M6 (MAYOR) Criterio de cierre 4, "precedencia", sin requisito
- Fuente: ficha criterio 4; `rg -i "precedencia|precedence"` en el change sólo devuelve design.md:108 (output style) y tasks 1.5 (agentes del mismo nombre).
- Observación: no hay requisito ni escenario sobre qué gana cuando chocan el texto del usuario en `CLAUDE.md`, el bloque universal y el prompt del coordinador, ni entre `CLAUDE.md` global y de proyecto. La tarea 1.3 investiga si el prompt de `--agent` reemplaza o se suma, pero su resultado no alimenta ninguna spec.
- Corrección verificable: requisito "Precedencia de capas" con al menos un escenario de choque usuario vs coordinador y uno de agente Project vs Machine con el mismo ID, condicionado a 1.3 y 1.5.

### M7 (MAYOR) "Update" del plan no es el `update` real
- Fuente: design D5 ("Update = install con assets nuevos"); tasks 6.5. El comando real reemplaza el binario y reaplica la selección registrada (`cmd/active-stack/update.go:94`, `--sync`; estado en `internal/update/state.go`, que registra `projects`).
- Impacto: el criterio "actualización" se probaría por un camino que el usuario no usa; no queda dicho si `update --sync` migra instalaciones Project de V1 registradas o sólo Machine.
- Corrección verificable: tarea RED en `cmd/active-stack` o `internal/update` que corra el camino de sync sobre un HOME con bloque legado y verifique migración; escenario explícito para proyectos registrados.

### M8 (MAYOR) Fallback exigido por spec sin entrega ni documentación en tareas
- Fuente: spec `claude-coordinator-agent` "Fallback documentado ..." y "Verify informa la vía activa"; tasks 7.1 menciona la salida y 9.1 no menciona el fallback; ninguna tarea instala la skill o comando.
- Impacto: verify debería nombrar una vía que nunca se instala; escenario no verificable.
- Corrección verificable: decidir (G2) si el fallback se entrega en este change; si no, bajar el requisito a documentación y quitar la detección de verify; si sí, agregar tareas RED de instalación y detección.

### M9 (MAYOR) Rollback byte a byte pisa ediciones posteriores del usuario
- Fuente: spec "Rollback completo" (restaurar `CLAUDE.md`, `settings.json` y agentes al snapshot previo a la primera escritura); `internal/backup/restore.go` restaura archivos enteros (Existed=true copia, Existed=false borra).
- Observación (razonada): si el usuario editó `CLAUDE.md` o `settings.json` después de migrar, restore los pierde; además `settings.json` es sólo lectura en este change (D5), así que restaurarlo sólo puede perder cambios ajenos. No queda dicho cuál manifiesto es "el previo a la primera escritura" cuando hubo reinstalaciones o updates que generaron snapshots nuevos.
- Corrección verificable: escenario "usuario editó después de migrar" con el resultado esperado (advertencia y confirmación, o diff), quitar `settings.json` del restore o justificarlo, y definir cómo se identifica el manifiesto de la migración.

### m1 (MENOR) Matriz "exactamente una capa" contra el borrador
- spec `claude-instruction-layers` exige una capa por regla; el borrador asigna `governance.md` a "módulo propose + coordinador" y engram a dos capas. Además falta la línea `Strict TDD Mode: enabled` que agrega `compose.go` al módulo TDD. Corrección: permitir varias capas con justificación o partir la regla; agregar esa línea a la matriz.

### m2 (MENOR) Goldens de Codex y Antigravity no existen
- `fdfind -e golden` sólo lista `compose_all_toggles_claude.golden` y `compose_all_toggles_opencode.golden`. El escenario "Goldens de otros agentes intactos" es vacío para Codex y Antigravity. Corrección: tarea que agregue goldens o una aserción de igualdad de `Compose` para esas variantes antes de tocar `compose.go`.

### m3 (MENOR) Escenarios faltantes del brief sin cobertura
- Colisión en `<root>/.claude/agents` (el escenario sólo nombra `~/.claude/agents`).
- Cambio de ID entre versiones (G5): archivo viejo con marca queda huérfano; uninstall por ruta fija no lo encuentra. Escenario: update con ID nuevo borra el archivo viejo con marca.
- Usuario que editó el archivo de agente con marca: update lo pisa (el snapshot lo guarda). Declararlo.
- Desinstalación parcial (falla tras quitar el bloque universal y antes del agente): sin escenario de rollback.
- Marker mal formado: la spec pide advertencia, pero hoy `PurgeStaleSections` lo ignora en silencio (`inject.go:59-60` comentario) y ninguna tarea tiene RED para la advertencia.
- Advertencia del contenido del usuario "en la ayuda de install": la tarea 9.1 sólo cubre README.

### m4 (MENOR) G6 global y el motivo del `Chmod(dir, 0o755)`
- `writer.go:37-41` explica que el chmod repara directorios restringidos por una versión vieja del instalador (introducido en 6a40f17, 2026-05-27). Un arreglo global puede regresionar ese caso. Corrección: el test de 5.4 incluye "directorio sin bit de escritura del dueño sigue siendo reparable" y la regla pasa a "agregar sólo el bit de escritura del dueño, nunca ampliar grupo/otros".

### m5 (MENOR) Clasificación condicionada a una decisión de grupo
- proposal.md: "si el grupo decide que settings.json puede contener credenciales, la clasificación se revisa". Es un hecho a verificar, no una decisión: Claude Code admite `env` en `settings.json` (documentación pública, SUPOSICIÓN no verificada acá; en esta máquina la clave no aparece) y `~/.claude.json` guarda config MCP; G6 global toca la escritura de ambos. El CLAUDE.md del proyecto fija ALTO como techo (§4) y la regla global del usuario dice que un límite de datos sensibles es CRITICAL aun en diffs chicos. Corrección: dejar la tensión explícita para que la resuelva la persona en `## Aprobación`; no la resuelvo yo.

### m6 (MENOR) Inconsistencias de forma
- proposal.md:96 pide decisiones "G1-G7"; design tiene G1-G8.
- AGENTS.md no contiene la frase de C6 "ya vive en tu" (sólo CLAUDE.md:9); AGENTS.md:149 tiene una equivalente. Ajustar la cita.
- Compartir con el equipo: `openspec/` está ignorado; necesita decisión humana (`git add -f` o cambio de ignore). Ya lo informó el planner.

## Factibilidad de TDD (JUICIO)
Las tareas 3.x a 7.x son construibles con RED por aserción con los tests Go existentes: hay costuras (`SnapshotterCreate`, `configInstallFn`, `markerRemovalFn`, `restoreFn`) y los tests base pasan. Lo que no es TDD de Go son los escenarios de contexto efectivo (8.x) y "Especialista de apply recibe el módulo TDD" bajo G3(a), que depende de que el modelo coordinador pegue el módulo; ese escenario prueba obediencia del modelo, no del instalador. Con (c) sería determinista por archivo. Debe quedar dicho en la spec.

## EVIDENCIA
Todas el 2026-11-06, revisión 97d7f6c, sin cambios de código.
- `git rev-parse --short HEAD` → 97d7f6c; rama `claude-native-opsx`; `git status --short` vacío (artefactos ignorados).
- `git check-ignore -v openspec/changes/claude-native-opsx/proposal.md` → `.gitignore:36:openspec/` (exit 0).
- Lectura de código citado: `adapter.go` (InstructionsPath :28-30, ConfigDelivery :125-127, PathsFor :91-118), `install.go:21-89`, `inject.go:22-126` (ownedSectionIDs, PurgeStaleSections, `WriteFileAtomic(..., 0o644)` :125), `primary_agent.go:25-120` (escribe `{"agent":{id:{...}}}`, `0o644` en :83 y :118), `writer.go:16-80` (`MkdirAll 0o755` :35, `Chmod(dir,0o755)` :40, `tmp.Chmod(perm)` :63). C1, C3 (forma del JSON), C4 y C5 coinciden con el código.
- `assets/claude/sdd-orchestrator.md`: :3 "Bind this to the dedicated...", :102-105 tabla con `openspec-apply-change`/`openspec-archive-change`, :144-149 Project Standards, :205-210 Rules, :212-226 Model Assignments. `strict-tdd.md:1` "Apply Phase". Coinciden con la matriz.
- `internal/install/plan.go:98-101, 369-374, 494-580`; `internal/uninstall/steps.go:140-176`; `internal/verify/harness_checks.go:85-130, 236-250`; `internal/backup/restore.go` (Existed/!Existed); `cmd/active-stack/uninstall_dispatch.go:12-40`.
- `claude --version` → 2.1.318; `claude --help` → `--agent <agent> Agent for the current session. Overrides the 'agent' setting`, `--system-prompt-snapshot <on|off>`, `--setting-sources`.
- 09:22:25-03:00 `go test ./internal/harness/config/... ./internal/agents/claude/... ./internal/filemerge/... ./internal/verify/... ./internal/uninstall/... ./internal/install/...` → todos `ok` (línea base, no prueba nada del change).
- `fdfind -e golden internal/harness/config/testdata` → sólo claude y opencode.
- `stat -c %a` (sólo modos, sin contenido): `~/.claude` 755, `settings.json` 644, `CLAUDE.md` 644, `~/.claude.json` 644, `~/.claude/agents` 775. `rg -c '"env"|apiKeyHelper' ~/.claude/settings.json` → sin coincidencias.
- Transcripción de esta sesión (sólo conteos de tipos, sin leer contenido): 9 entradas `attachment.type=instructions` con un texto conocido del CLAUDE.md global; 20 `attachment.type=prompt_snapshot`.
- `openspec validate claude-native-opsx --strict` exit 0: reusado del planner, no reejecutado (artefactos sin cambios según hashes de este informe; no tengo hashes del momento de su corrida, SUPOSICIÓN de igualdad).

## JUICIO
Autor: plan-reviewer. Comparación: ficha fuente y criterios de cierre contra proposal/design/specs/tasks, y citas contra el código en 97d7f6c. Confianza: alta en citas, M2, M3, M7, m2 y m6 (observados en código); media en B1, M4 y M5 (razonados sobre evidencia parcial y documentación pública); media en la severidad elegida para B1 (si 1.x muestra que el agente de archivo no queda expuesto a la sesión normal, baja a MENOR).

## SUPOSICIÓN (no verificado)
- Delegación automática por `description` y prohibición de subagentes anidados en 2.1.318: documentación pública, no probado.
- Que `prompt_snapshot` contenga el cuerpo de un agente de sesión.
- Que `settings.json` de Claude admita `env` con secretos en esta versión.
- Comportamiento de `update --sync` sobre instalaciones Project: no leí el flujo completo.
- No ejecuté ninguna sesión de Claude con modelo ni ningún código del change (no existe).

## Fuentes leídas
proposal.md, design.md, tasks.md, las tres specs, `informes/propose-1-planner.md`; código listado en EVIDENCIA; `CLAUDE.md` del repo (§3-§4, búsquedas), `AGENTS.md` (búsquedas), `ARCHITECTURE.md:74`, `README.md:214`, `internal/catalog/harnesses.yaml:55-130`, `assets/design-mockups.md` (búsqueda), `e2e/e2e_test.sh` (búsqueda), `internal/update/state.go` (búsqueda), `cmd/active-stack/update.go:94`.

## Lecciones L-NNN
El repo no tiene base de lecciones ni `kb/` (coincide con lo declarado por el planner; no lo reverifiqué con listado propio más allá de no encontrarlas en las rutas leídas).

## Contaminación
Leí el informe del planner (`propose-1-planner.md`) por indicación del brief; tomé de ahí la lista C1-C6 y G1-G8 como punto de partida y reverifiqué en código C1, C3, C4, C5 y C6. La evidencia D0 viene del orquestador vía design.md; no la reproduje. No consulté Engram ni informes de otros roles.
