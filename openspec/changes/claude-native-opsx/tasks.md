# Tasks: claude-native-opsx

## Estado del workflow

- Nivel: ALTO (HIGH) con tensión ALTO vs CRITICAL abierta para la persona que aprueba
  (proposal.md, Clasificación).
- Fase: propose. Intento: 3 (corrección de N1-N8 del plan-reviewer, intento 2). Es el último
  intento antes de escalar a la persona: si el plan-reviewer vuelve a devolverlo, no hay cuarto
  intento automático y se ofrecen las tres opciones del flujo (alcance nuevo aprobado, deuda
  justificada o detener el change).
- Rol activo: planner (correcciones del intento 3 entregadas); próximo rol: plan-reviewer, intento 3.
- Revisión base: rama `claude-native-opsx`, base `97d7f6c` (= main remoto). Sin diff de código.
  Los artefactos están ignorados por git (`.gitignore:36`); la revisión se ata a hashes sha256 de
  los artefactos (los del intento 1 están en el informe del plan-reviewer).
- Informes originales:
  - `informes/propose-1-planner.md` (intento 1, planner).
  - `informes/propose-1-plan-reviewer.md` (intento 1, veredicto VUELVE AL PLANNER: 1 BLOQUEANTE,
    9 MAYOR, 6 MENOR).
  - `informes/propose-2-planner.md` (intento 2, planner: tabla hallazgo → cambio).
  - `informes/propose-2-plan-reviewer.md` (intento 2, veredicto VUELVE AL PLANNER, ronda corta:
    N1-N3 MAYOR, N4-N8 MENOR).
  - `informes/propose-3-planner.md` (intento 3, planner: tabla N → cambio y hashes).
- Bloqueantes para construir: (1) revisión del plan intento 3; (2) decisiones de grupo G1-G12
  (design.md); (3) aprobación humana fechada en proposal.md `## Aprobación`, incluido el nivel ALTO
  o CRITICAL; (4) matriz acordada con la Propuesta 04; (5) resultado de 1.10 (B1) antes de dar por
  cerrada la vía D1.
- Pendiente humano fuera del plan: compartir los artefactos con el equipo (`openspec/` ignorado por
  git; `AGENTS.md:150` lo declara intencional). No se tocó git ni `.gitignore`.
- Próxima acción: plan-reviewer revisa el intento 3.
- Lecciones L-NNN: el repo no tiene base de lecciones (sin `kb/` ni archivos de lecciones,
  verificado 2026-11-06). Constructor, breaker y auditor deben registrar esa ausencia.

Reglas para toda tarea de código: TDD obligatorio. RED = test de comportamiento que falla por
aserción (no por compilación) antes de tocar producción; GREEN mínimo; refactor con tests focales.
Por ser ALTO: regresiones de consumidores (OpenCode, Codex, Antigravity, `design-mockups`) y un
mutante temporal sobre cada aserción clave marcada con (M), revertido antes de entregar. Si la
persona elige CRITICAL, toda aserción de seguridad o datos lleva (M). Guardar bajo cada tarea
comando, timestamp, revisión y salida de RED y GREEN. Ninguna tarea se marca hecha sin esa
evidencia. Nunca build salvo pedido explícito (regla del repo).

## 1. Verificación de suposiciones (sin código de producción)

Registrar cada resultado como EVIDENCIA o SUPOSICIÓN en
`openspec/changes/claude-native-opsx/investigacion/contexto-efectivo.md`, con versión de Claude,
modelo, comando, timestamp y testigo usado. Scripts en `investigacion/`, nunca en `internal/`.
Toda corrida con modelo sigue D9 (sin copiar credenciales, sólo nombres de variables, borrar el
directorio temporal). Orden: 1.12 y 1.11 primero, porque las demás dependen de ellas.

- [ ] 1.1 Subagente definido como archivo en el directorio de agentes (no sólo `--agents` JSON): repetir el caso C de D0 con testigos únicos en `CLAUDE.md` de proyecto y global; verificar con el oráculo de 1.11 en la transcripción propia del subagente.
- [ ] 1.2 Agentes integrados (Explore, Plan, general-purpose): mismo procedimiento; registrar si reciben `CLAUDE.md`.
- [ ] 1.3 Frontmatter admitido por archivos de agente (nombre, descripción, tools, model, carga de skills, exclusión de delegación) y si el prompt de un agente de sesión (`--agent`) reemplaza o se suma al system prompt; resultado alimenta el requisito "Precedencia de capas".
- [ ] 1.4 Versión mínima: identificar la primera versión de Claude Code con `--agent` y agentes de archivo (changelog oficial o prueba con versión anterior); el valor reemplaza el `(TBD)` de design.md D7.
- [ ] 1.5 Target Project: verificar que Claude lee `<root>/.claude/agents/<id>.md` y `<root>/.claude/CLAUDE.md` en una sesión abierta en `<root>`, y cuál prevalece frente a un agente del mismo ID en `<home>/.claude/agents/`; resultado alimenta el escenario "Agente Project y Machine con el mismo ID".
- [ ] 1.6 Origen real de las skills `openspec-*` en Claude (contradicción C2): qué crea `openspec init` 1.14.3, qué instala Active Stack y qué nombres invoca el coordinador; dejar la tabla de nombres reales.
- [ ] 1.7 Coordinador como agente de sesión (`--agent`) lanza subagentes con la herramienta Agent: verificar que puede y que el subagente no recibe el testigo del cuerpo del coordinador (oráculo 1.11).
- [ ] 1.8 Tipo y efecto del setting `agent` en `settings.json` (contradicción C3): verificar con una copia aislada de settings, nunca con el archivo real del usuario.
- [ ] 1.9 Skills: qué contexto recibe una skill invocada por un especialista y si una skill puede llevar el módulo por acción (insumo de G3).
- [ ] 1.10 (B1) En 2.1.318 y en la versión mínima: (a) si un agente de archivo aparece como tipo de subagente disponible en una sesión normal; (b) si la sesión lo delega sola ante un pedido OPSX que no lo nombra; (c) si un subagente puede usar la herramienta Agent; (d) si alguna mitigación de D1 (frontmatter, `description`, guarda en el cuerpo) evita (a) o (b). Verificación: tabla caso → resultado con transcripción. Si ninguna mitigación excluye al coordinador, marcar G9 como bloqueante y devolver al planner.
- [ ] 1.11 (M4) Validar el oráculo de D8: con testigos conocidos en `CLAUDE.md`, cuerpo de agente de sesión y prompt de subagente, comprobar en qué entradas de la transcripción aparece cada uno (`attachment.type=instructions`, `attachment.type=prompt_snapshot` con `--system-prompt-snapshot on`, u otras), en la misma corrida que un testigo ausente a propósito. Verificación: el oráculo distingue presente de ausente para las tres capas; si no cubre alguna, se documenta el reemplazo y se ajustan los THEN.
- [ ] 1.12 (M5) Mecanismo de autenticación con directorio de configuración temporal: verificar qué variable de entorno usa Claude Code para el directorio de configuración y qué variable acepta para la credencial; registrar sólo los nombres. Prohibido copiar archivos de credenciales. Verificación: una sesión `claude -p` en el directorio temporal responde y el directorio no contiene archivos de credenciales copiados; el directorio se borra al final.
- [ ] 1.13 (m5) Hecho a verificar, no decisión: si `settings.json` (por ejemplo clave `env`, `apiKeyHelper`) o `~/.claude.json` pueden contener credenciales en la versión mínima y en 2.1.318, según documentación oficial y sin leer valores del usuario. Resultado al humano para la decisión ALTO vs CRITICAL en `## Aprobación`.

## 2. Coordinación y decisiones de grupo

- [ ] 2.1 Presentar al grupo G1-G12 de design.md con recomendación (G9 con el resultado de 1.10); registrar en design.md la opción elegida, fecha y quién decidió. Verificación: cada fila G tiene decisión o queda `(TBD grupo)` explícito, y `rg -n "depende de G" specs/` coincide con las filas abiertas.
- [ ] 2.2 Acordar con Giuliano (Propuesta 04) la matriz de reglas: inventario de módulos, quién los carga y cuándo; publicar la matriz única (ubicación a acordar) y enlazarla desde design.md. Verificación: escenarios "Ninguna regla sin capa" y "Regla en dos capas".
- [ ] 2.3 Acordar con Matías (Propuesta 02) qué métrica de contexto se compara antes/después y en qué revisión; dejar la referencia en la tarea 8.3.
- [ ] 2.4 Si alguna decisión de 2.1 a 2.3, o un resultado de la sección 1, cambia specs o design, devolver al planner y repetir plan-reviewer antes de seguir.
- [ ] 2.5 Pedir la aprobación humana en proposal.md `## Aprobación` con el resultado de 1.13 (la completa la persona, no un agente). Sin ella no empieza la sección 3.

## 3. Modelo y adapter de Claude

- [ ] 3.1 RED en `internal/agents/claude/adapter_test.go`: `ConfigDelivery()` devuelve el modo nuevo de archivo de agente (D1) y `PathsFor` resuelve el directorio de agentes para Machine (`<home>/.claude/agents`) y Project (`<root>/.claude/agents`). GREEN en `internal/model/harness.go` (constante nueva, sin reordenar las existentes) y `adapter.go`. Consumidores: `internal/install/types.go`, `internal/install/plan.go` (`projectTargetAdapter`, `:556-580`, debe delegar el directorio de agentes a `PathsFor` con target Project), `internal/uninstall/types.go`, `internal/verify/adapter.go`, `cmd/active-stack/main.go` (`projectVerifyAdapter`, `:333-345`, debe delegar el directorio de agentes a `PathsFor` con target Project), `internal/harness/config/types.go`; agregar el método a interfaces y fakes. Verificación: `go test ./internal/agents/... ./internal/model/... ./internal/install/... ./cmd/active-stack/...`.
- [ ] 3.2 Regresión: tests existentes de OpenCode (`ConfigDeliveryPrimaryAgent`), Codex y Antigravity siguen verdes sin cambiar aserciones. Verificación: `go test ./internal/agents/...`.
- [ ] 3.3 RED en test del adapter o de `internal/system`: detección de versión de Claude con salida válida, menor que la mínima, no parseable o con error, y binario ausente; cada caso devuelve un estado distinto (compatible, insuficiente, desconocida, ausente). (M) sobre la comparación de versión.

## 4. Partición del asset y composición

- [ ] 4.0 (m2) Antes de tocar `compose.go`: agregar goldens de `Compose` con todos los toggles para Codex y Antigravity en `internal/harness/config/testdata/` (hoy sólo existen los de Claude y OpenCode), generados desde 97d7f6c. Verificación: `go test ./internal/harness/config/...` verde y los goldens nuevos coinciden con la salida de la base.
- [ ] 4.1 Partir `internal/harness/config/assets/claude/sdd-orchestrator.md` según la matriz acordada (2.2): contenido universal y contenido de coordinador; módulos según G3. No cambiar el texto normativo, sólo reubicarlo.
- [ ] 4.2 RED en `internal/harness/config/compose_test.go`: la composición universal para Claude no contiene encabezados del coordinador (Role, Delegation Rules, Sub-Agent Launch Pattern, Model Assignments) ni el módulo TDD; la del coordinador sí los contiene según toggles; con `tdd` inactivo no contiene `Strict TDD Mode: enabled`; (N4) la composición legado (universal + coordinador + módulos, en el orden de hoy) es byte a byte igual al golden `compose_all_toggles_claude.golden` grabado en 97d7f6c. (M) sobre la aserción de ausencia y sobre la igualdad del legado. GREEN en `compose.go` sin cambiar la firma usada por otros agentes.
- [ ] 4.3 Conservar el golden actual de Claude como golden de la composición legado (sin regrabarlo) y agregar goldens nuevos para universal y coordinador; verificar que los goldens de OpenCode, Codex y Antigravity no cambian (`git diff --stat` de testdata limitado a Claude).
- [ ] 4.4 RED en `internal/harness/config/inject_test.go`: después de inyectar la sección universal, una segunda inyección no la purga (contradicción C5); GREEN registrando la sección como propia (`ownedSectionIDs` o `sectionFragments`). (M) quitando el registro.
- [ ] 4.5 (m3) RED: con un marker `active-stack:` de apertura sin cierre en `CLAUDE.md`, install no elimina el fragmento y emite una advertencia con ruta e ID (hoy `PurgeStaleSections` lo ignora en silencio). Verificación: test focal sobre la salida de advertencias.

## 5. Inyector del archivo de agente y transacción

- [ ] 5.1 RED en un test nuevo de `internal/harness/config`: escribe `<agents>/<id>.md` con frontmatter y marca de propiedad; segunda ejecución reporta sin cambios y bytes idénticos.
- [ ] 5.2 RED: si existe un archivo con el mismo nombre sin marca de propiedad, en Machine y en Project, devuelve error con la ruta y el archivo queda idéntico; archivos ajenos del directorio no cambian. (M) sobre la verificación de la marca.
- [ ] 5.3 RED: fallo del snapshot (fake de `SnapshotterCreate`) deja todo sin cambios; fallo de escritura del agente después de purgar `CLAUDE.md` restaura `CLAUDE.md` byte a byte (D5 paso 6). (M) sobre la restauración.
- [ ] 5.4 RED: permisos según G6: `CLAUDE.md` en 0600 y directorio en 0700 conservan su modo; (m4) un directorio en 0500 recibe sólo el bit de escritura del dueño (0700) y la escritura completa, preservando la reparación de 6a40f17. Si G6 es global, el test vive en `internal/filemerge/writer_test.go` y cubre los llamadores de `inject.go:125` y `primary_agent.go:83`, `:118`. (M) sobre las aserciones de modo. Saltar en Windows con motivo registrado.
- [ ] 5.5 RED: en target Project el manifiesto de snapshot del inyector nuevo queda bajo `<home>/.active-stack/backups/`, no bajo `<root>` (D5 paso 3, contradicción C8).

## 6. Instalación, migración y update

- [ ] 6.1 RED en `internal/harness/config/install_test.go`: install para Claude sobre un `CLAUDE.md` con bloque legado y texto del usuario antes y después deja snapshot con el contenido previo, sin secciones del coordinador, bloque universal una vez, archivo de agente y texto del usuario idéntico. (M) sobre la comparación byte a byte.
- [ ] 6.2 RED: `design-mockups` presente antes sigue intacta después de la migración.
- [ ] 6.3 RED: install no escribe la clave `agent` en `settings.json` y conserva un valor previo del usuario.
- [ ] 6.4 (M2) RED, un test por caso del requisito "Compatibilidad por versión y entorno sin Claude": (a) versión insuficiente con legado previo: no escribe agente ni purga legado y advierte; (b) versión insuficiente o desconocida sin entrega previa: escribe el bloque legado completo y advierte; (c) binario ausente: entrega por capas y advertencia de versión no comprobada (según G11); (d, N4) update en modo legado con assets nuevos reescribe el bloque legado con la composición legado y una segunda ejecución no cambia nada.
- [ ] 6.5 RED: install con assets de testigo T1 → T2 deja sólo T2 en ambos archivos y el contenido del usuario idéntico.
- [ ] 6.6 Wiring en `internal/install/plan.go` (`collectWritePaths` incluye el archivo de agente y el plan de backup lo cubre) con test en `internal/install`; target Project no escribe en `<home>` salvo el manifiesto. Verificación: `go test ./internal/install/...`.
- [ ] 6.7 Aviso BREAKING en la salida de install/update para Claude y advertencia del contenido del usuario en `install --help` (textos en `internal/i18n`), con test de presencia en ambos idiomas.
- [ ] 6.8 (M9, N6) RED: la primera migración registra la ruta del manifiesto de migración y los hashes resultantes, y crea ese manifiesto con `pinned: true` (verificable con `Prune` y límite 0 o 1 sobre el directorio: el manifiesto sobrevive); una reinstalación posterior genera su snapshot pero no reemplaza ese registro.
- [ ] 6.9 (M7) RED en `cmd/active-stack/update_test.go` o `internal/update`: el camino `update --sync` sobre un HOME con `installation.json` que selecciona Claude y `CLAUDE.md` con bloque legado produce la migración de 6.1; con `projects` registrados que tienen bloque legado en `<root>/.claude/CLAUDE.md`, no escribe en ellos y emite una advertencia por proyecto con el comando de migración (según G10); (N5) con un HOME sin `installation.json` o sin agentes registrados y bloque legado en `CLAUDE.md`, no escribe `CLAUDE.md` y la salida, además del mensaje actual de `update.go:375-378`, nombra el cambio BREAKING de Claude y el comando `active-stack install`.
- [ ] 6.10 (m3) RED: update con ID de coordinador nuevo crea el archivo nuevo y borra el viejo sólo si tiene marca de propiedad (según G5).
- [ ] 6.11 (m3) RED: un archivo de agente con marca editado por el usuario se reemplaza en update, su versión editada queda en un snapshot y la salida nombra la ruta.

## 7. Verify y uninstall

- [ ] 7.1 RED en `internal/verify/harness_checks_test.go`: en versión compatible falla si falta el archivo de agente, si el marker universal no aparece exactamente una vez o si queda `active-stack:sdd-orchestrator`; pasa en instalación correcta; en modo legado por versión y con versión no comprobada reporta advertencia, no fallo; la salida nombra la vía activa. No cambiar `checkPrimaryAgentConfig` de OpenCode.
- [ ] 7.2 (M9, N6) RED en `internal/uninstall` y `cmd/active-stack/headless` (flags): uninstall Machine sin restore elimina sólo el bloque universal y el archivo con marca propia; con `--strategy restore --restore-manifest <m>` no toca `settings.json` y devuelve `CLAUDE.md` byte a byte si el hash coincide; si difiere, no escribe, guarda el contenido actual en un snapshot nuevo y termina con código no cero; con `--overwrite-edited` (nombre tentativo, flag nuevo con ayuda e i18n) restaura después de guardar ese snapshot; `--yes` solo no saltea la guarda; con un manifiesto distinto del de migración registrado la guarda se aplica igual y la salida lo advierte. (M) sobre el chequeo de hash, sobre la independencia de `--yes` y sobre el borrado condicionado a la marca.
- [ ] 7.3 Regresión: tests de uninstall y verify de OpenCode sin cambios de aserciones. Verificación: `go test ./internal/verify/... ./internal/uninstall/...`.
- [ ] 7.4 (N2) RED: el check nuevo del archivo de agente usa el directorio de agentes que resuelve el adaptador recibido; con `projectVerifyAdapter` (`cmd/active-stack/main.go:320-345`) revisa `<root>/.claude/agents/<id>.md`, y un archivo presente sólo en `<home>` no hace pasar el check Project. Verify ya conoce Project para `CLAUDE.md`; esta tarea no lo reimplementa. Disparador: hook de verify posterior a install (no hay subcomando `verify`).
- [ ] 7.5 (N1) Uninstall Project fuera de alcance (decisión vigente, `cmd/active-stack/headless/uninstall_flags.go:50`). RED: la salida de install Project para Claude nombra `<root>/.claude/agents/<id>.md` y la sección universal de `<root>/.claude/CLAUDE.md` como limpieza manual; regresión: `uninstall_flags_test.go:179-182` sigue verde sin cambiar su aserción. Si el grupo elige la otra opción de G12, esta tarea vuelve al planner.
- [ ] 7.6 (m3) RED: uninstall que falla al borrar el archivo de agente después de quitar el bloque universal restaura el bloque desde el snapshot de esa desinstalación y nombra la ruta.

## 8. Pruebas de contexto efectivo con testigo

Precondiciones: 1.11 (oráculo) y 1.12 (autenticación) cerradas como EVIDENCIA. Se usa el
directorio de configuración temporal de D9; nunca el `~/.claude` real; sin copiar credenciales;
informes con nombres de variables, no valores; borrar el directorio al final.

- [ ] 8.1 Script reproducible en `investigacion/` (o `e2e/` si el grupo lo pide; hoy `e2e/e2e_test.sh` corre en Docker sin `claude`) que instala con testigos únicos en universal, coordinador, módulos TDD y governance, y contenido del usuario; ejecuta sesión normal, `--agent <id>`, subagente lanzado por el coordinador para propose y para apply, y el caso B1: si 1.10 encontró exclusión por frontmatter, una corrida que verifica que `<id>` no figura entre los subagentes disponibles; si no, N = 20 corridas del pedido OPSX sin nombrar al coordinador con umbral 0 de 20 (escenarios de `claude-coordinator-agent`). Verificación: tabla de presencia/ausencia por oráculo de registro de contexto, con control positivo en cada corrida, que coincide con los escenarios de `claude-instruction-layers` y `claude-coordinator-agent`.
- [ ] 8.2 Repetir 8.1 para target Project (con 7.4 en verde; la limpieza Project se hace a mano según README, no con uninstall), para precedencia (escenarios del requisito "Precedencia de capas") y, en Machine, después de update, reinstalación y restore.
- [ ] 8.3 Comparación de contexto antes/después con la métrica acordada con la Propuesta 02 (2.3); si 02 no entregó línea base, registrar la comparación como pendiente con su dependencia, sin bloquear los criterios funcionales (G8).

## 9. Documentación (Knowledge Base Impact)

- [ ] 9.1 Reescribir `ARCHITECTURE.md` (líneas 56-75), `README.md` (línea 214 y tabla), `CLAUDE.md:9` y la sección 2 del repo, y `AGENTS.md:148-149` (espejo), con la invocación `claude --agent <id>`, el límite del contenido del usuario, el orden de precedencia medido (1.3, 1.5), el modo legado por versión, la nota BREAKING, el manifiesto de migración y `--overwrite-edited` para el rollback, y la limpieza manual en target Project (N1). Verificación: los comandos documentados corren tal cual en el entorno de 8.1.
- [ ] 9.2 Agregar la entrada del change en `CHANGES.md` con dependencias 02 y 04.
- [ ] 9.3 (M8) Documentar en README el fallback de activación del coordinador (según G2) y verificar con el oráculo que el comando documentado carga el testigo del coordinador (escenario "El fallback documentado funciona"). Si el grupo decide instalar el fallback, agregar antes de esta tarea un RED de instalación y uno de detección en verify, y repetir plan-reviewer.

## DEUDA registrada por el plan

- DEUDA-1 (intento 3, N1): uninstall con target Project para la entrega de Claude (archivo de agente y sección universal en `<root>`). Hoy uninstall es Machine-only por decisión vigente (`uninstall_flags.go:50`). Asignada al primer change que necesite desinstalar en Project; mientras tanto, limpieza manual documentada (tareas 7.5 y 9.1). Si el grupo elige revertirla en G12, deja de ser deuda y entra al alcance.

## 10. Integración

- [ ] 10.1 Suite completa una vez sobre la revisión entregada: `go test ./...`, `go vet ./...` y el formateo del repo; registrar comando, timestamp, revisión y veredicto.

## Workflow follow-up

- Plan-reviewer revisa el intento 3; hallazgos vuelven al planner sólo con decisión humana (límite de tres intentos).
- Breaker y auditor sobre la revisión entregada; al cerrar, knowledge-base-report.md y nota del proyecto (destino pendiente: el repo no tiene `obsidian.md`).
- Archivar sólo con APROBADO del auditor y aprobación humana registrada.
