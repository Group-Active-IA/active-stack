# Guía de implementación en equipo

**Frase de inicio** · «Soy <nombre>, quiero implementar lo mío» (o simplemente «continuar» si ya estás trabajando en una ficha).

Esta guía la sigue el agente (Claude Code o Codex) de **cada integrante**. Reemplaza al encargo anterior de verificación de Claude, que quedó [archivado](archive/roadmap-f1-f5/verificacion-claude-code.md). Trabaja sobre el [roadmap integrado](roadmap-implementacion-equipo.md), el [estado](estado-equipo.md) y los dos planes: [trabajo en equipo](plan-trabajo-en-equipo.md) **[EQ]** y [Engram Cloud](plan-engram-cloud-equipo.md) **[EN]**.

---

## Para la persona (qué hacer)

1. Bajá el repo y abrilo con Claude Code o Codex. Si ya lo tenías, el agente hace `git fetch` solo.
2. Escribí: **«Soy Matías, quiero implementar lo mío»** (con tu nombre).
3. El agente te dice qué ficha te toca, en qué orden y, si depende de otra persona que no terminó, **te avisa y no te deja empezar**.
4. Implementás con el agente. Cuando termina, genera tu **informe de cierre** (un `.md` corto).
5. **Pasale ese `.md` a Juan Cruz por el grupo de WhatsApp.** Él revisa y acepta tu PR y actualiza el estado.
6. El agente te dice cuál es tu próxima ficha y de quién espera. Cuando Juan Cruz la libere, repetí desde el paso 2.

---

## 1. Reglas que no se negocian

- **Sin SDD.** No crear changes de OpenSpec ni propuesta/diseño/specs/tareas. Cada ficha se ejecuta con **criterios cortos** (los tests `T-*` de la fase indicada en el roadmap) y **TDD estricto**. OpenSpec solo si un change ya está iniciado o el dueño lo pide.
- **Orden por dependencia.** Nunca se implementa una ficha cuya dependencia no figure `mergeada` (sección 3).
- **Una ficha = una rama = un PR.** Rama `feat/R-xx-slug` desde `origin/main` actualizado.
- **Solo el dueño escribe [estado-equipo.md](estado-equipo.md).** Nadie más lo edita.
- Todas las reglas del proyecto de [CLAUDE.md](../CLAUDE.md) (espejo [AGENTS.md](../AGENTS.md)) siguen vigentes, en especial: NUNCA commitear sin pedido explícito de la persona; *conventional commits* y **sin coautoría de IA**; NUNCA `go build` ni `active-stack install` real; NUNCA pisar config del usuario sin backup; markers idempotentes; paths por el adapter; `(TBD)` donde falte una decisión.
- **Governance ALTO** (R-16, R-17, R-18 y toda ficha que toque `backup`, `filemerge` o `pipeline`): primero se **propone** y se espera la revisión del dueño; no se codifica antes.
- **Hecho = evidencia.** Algo es «verificado» solo con comando, salida y versión. Una prueba que no se pudo ejecutar se informa como **no ejecutada**.
- **Las decisiones no se inventan.** Si falta una, se aplica la sección 4.

## 2. Identificar a la persona

1. Si dijo su nombre («soy Matías»), buscarlo en la tabla **Equipo** de [estado-equipo.md](estado-equipo.md). Si no lo dijo, **preguntarlo**: no adivinar por el nombre de Git.
2. Si hay usuario de GitHub registrado, compararlo con `gh api user` (o `git config user.name`). Si no coincide, mostrar la diferencia y pedir confirmación. Si el usuario figura `(TBD)`, pedirle que lo confirme y avisar al dueño para completarlo.
3. Si el nombre no está en la tabla, **modo consulta**: solo lectura; explicar que el dueño debe asignarle una ficha.
4. El dueño (Juan Cruz) puede pedir «mostrame el estado». Si dice «Soy Juan Cruz, continuá» o «implementá lo mío», su ficha es **R-05**: seguir [equipo/fichas/R-05.md](equipo/fichas/R-05.md) (Windows + Claude Code; el dueño también actualiza el estado).

## 3. Detectar qué le toca y si puede empezar

1. `git fetch origin`. Leer **[estado-equipo.md](estado-equipo.md) desde `origin/main`** (no la copia local) y el [roadmap](roadmap-implementacion-equipo.md).
2. Listar las fichas de la persona **en el orden del roadmap**. Para cada una calcular su estado real combinando tres fuentes:
   - la tabla del estado (autoridad);
   - Git: ¿existe la rama `feat/R-xx-*` local o remota? ¿commits sin mergear?;
   - `gh pr list` / `gh pr view` (en el repo de la ficha): ¿PR abierto o mergeado?
   Si las fuentes **no coinciden** (por ejemplo, el PR figura mergeado pero el estado dice `PR abierto`), **decirlo** y no continuar hasta que el dueño actualice el estado.
3. Elegir la **primera ficha de la persona** cuyo estado sea `en curso` (retomar) o `pendiente`.
4. **Verificar dependencias.** La ficha puede empezar solo si **todas** sus dependencias están `mergeada` en el estado **y** el merge se confirma en Git (la rama base contiene el SHA registrado, o el PR figura mergeado). Casos:
   - **Todas mergeadas** (las dependencias `omitida` no cuentan) → seguir con la sección 4.
   - **Alguna no está mergeada** → **BLOQUEAR**. Decir de forma concreta: «No podés empezar R-xx todavía: depende de R-yy (de <persona>), que está en estado `<estado>`». **No** escribir código, no crear ramas, no «adelantar lo fácil».
   - Con la ficha bloqueada, ofrecer: (a) otra ficha propia que **no** esté bloqueada, (b) revisar el PR de quien bloquea si está abierto, (c) esperar y avisarle al dueño. Si la persona insiste en empezar igual, responder que solo avanza con una **autorización de excepción** escrita por el dueño en [estado-equipo.md](estado-equipo.md); sin esa línea, no se adelanta.
   - **Estado `bloqueada`** en el estado → leer la nota y avisar.
5. Mostrar el panorama (en pocas líneas): su ficha actual, de qué depende, qué desbloquea y cuál es su siguiente ficha. Si la ficha de otra persona que esperás **no terminó**, decirlo con su nombre («Giuliano todavía no terminó R-13»).

## 4. Decisiones abiertas (antes de implementar)

1. Leer en el [roadmap, sección 3](roadmap-implementacion-equipo.md) las decisiones `DA-xx` que afectan a la ficha, y los `(TBD)` y «No verificado» de la fase del plan que le corresponde.
2. Para **cada decisión abierta**, explicar a la persona **en detalle**: qué es, por qué importa, cada opción con sus ventajas, costos y consecuencias, qué recomienda el plan y qué pasaría si se elige mal. Si hace falta, consultar primero la documentación oficial (sección 7) para no explicar de memoria.
3. Esperar la respuesta. Si la decisión está marcada **dueño**, el implementador **no** la resuelve: se prepara la explicación y la pregunta para el informe y se avanza solo con lo independiente.
4. Registrar cada decisión tomada en el traspaso de la ficha (`docs/equipo/fichas/R-xx.md`, sección «Decisiones») con fecha y quién decidió. Nunca decidir por la persona.
5. Pedir decisión **solo si es indispensable**: antes, completar todo el trabajo independiente que ya esté autorizado.

## 5. Cómo se trabaja (sin SDD)

1. **Preparar.** Crear `feat/R-xx-slug` desde `origin/main` (si hay trabajo en paralelo en la misma máquina, en un *worktree* aparte). Crear `docs/equipo/fichas/R-xx.md` desde la [plantilla](equipo/fichas/PLANTILLA.md): objetivo, criterios (los `T-*` de la fase), decisiones y punto de reanudación. Es el **único** traspaso de la ficha; no crear otros archivos de handoff.
2. **Orquestar.** El agente principal define el próximo resultado verificable y **delega** el trabajo grande a subagentes (herramienta Agent en Claude Code; los equivalentes en Codex): exploración amplia, implementación de varios archivos y pruebas. Cada encargo lleva **contexto mínimo, archivos a cargo, criterios verificables y límites**. En paralelo **solo** tareas independientes con archivos distintos. El orquestador conserva las decisiones, integra, verifica y actualiza el traspaso.
3. **Resultado de cada subagente:** cambios, comandos y resultados de prueba, **evidencia TDD** (test rojo visto fallar antes de implementar) y bloqueos. Breve.
4. **TDD estricto:** test primero, verlo fallar por la razón correcta, mínimo para pasar, refactor. Al cerrar la fase: los comandos de la puerta de calidad indicados en el plan.
5. **Recomendar compactar** automáticamente (decirlo a la persona) en estos momentos: al cerrar un hito con el contexto muy cargado, antes de cambiar de frente si el historial dejó de servir, o ante un aviso de límite de contexto. **Antes de compactar:** sobrescribir el bloque «Punto de reanudación» del traspaso (evidencia, tareas abiertas, decisiones, siguiente acción), guardar lo durable en Engram (`mem_save`) con referencia a ese archivo, y evitar registros repetidos.
6. **«Continuar» después de compactar:** releer el traspaso de la ficha y el estado, comprobar Git y la rama, verificar si quedó algún subagente trabajando **antes** de reasignar, y ejecutar la siguiente acción. Conservar las autorizaciones y bloqueos vigentes.
7. **Objetivo (goal), según el cliente.** *Codex:* si el chat no tiene objetivo activo y la persona dijo «continuar», crear uno con el texto «Objetivo para el goal» del traspaso. *Claude Code:* no puede fijar `/goal`; mostrar **una sola vez**, en una línea, ese texto listo para que la persona lo pegue, y seguir trabajando.
8. **Qué no hace el agente solo:** commit, push y PR (siempre con el pedido de la persona), escrituras en ClickUp o GitHub, ni cambios de configuración de cuentas. Tampoco crea cuentas, inicia sesión ni acepta permisos de administrador: pide eso a la persona con pasos concretos.

## 6. Repos externos y PR

- Si la ficha dice «repo externo», el trabajo se hace en **ese repositorio** (`Group-Active-IA/active-orchestrator`, `Group-Active-IA/roadmap-generator`, `JuanCruzRobledo/agent-instruction`): clonar o usar un fork, crear la rama allí y **abrir el PR en ese repo**, aunque la persona no tenga permiso de escritura. **Juan Cruz acepta el PR.**
- El cambio de `ref` en `harnesses.yaml` (SHA completo) **no** lo hace el implementador: lo hace el dueño en R-20, en un commit propio y con el test de catálogo. Los tres repos externos ya están fijados por SHA: el cambio de un PR mergeado no llega a nadie hasta ese bump.
- Siempre que se cite un test o decisión de un plan, con prefijo: `[EQ]` o `[EN]`.
- Antes de abrir el PR: criterios verificados con evidencia, `go test ./... -count=1` y `go vet ./...` donde corresponda, espejo `CLAUDE.md`/`AGENTS.md` sincronizado si se tocó alguno, y el PR enlaza la ficha y su informe.

## 7. Documentación oficial primero (cuando algo falla o no se sabe)

Si algo no funciona o no está claro, **antes de probar a ciegas ni inventar**:

1. **Engram:** leer la documentación oficial vigente: [README](https://github.com/Gentleman-Programming/engram/blob/main/README.md), [DOCS.md](https://github.com/Gentleman-Programming/engram/blob/main/DOCS.md), [AGENT-SETUP.md](https://github.com/Gentleman-Programming/engram/blob/main/docs/AGENT-SETUP.md), [TEAM-USAGE.md](https://github.com/Gentleman-Programming/engram/blob/main/docs/TEAM-USAGE.md), [Engram Cloud](https://github.com/Gentleman-Programming/engram/blob/main/docs/engram-cloud/README.md) y los [issues](https://github.com/Gentleman-Programming/engram/issues). Hooks: [Claude Code](https://code.claude.com/docs/en/hooks) y [Codex](https://developers.openai.com/codex/hooks).
2. **Librerías y herramientas** (Go, Bubbletea, GitHub CLI, ClickUp, OpenSpec, etc.): consultar su documentación oficial; si hay Context7 disponible, usarlo.
3. Anotar en el traspaso la **versión** consultada (de Engram, Claude Code, Codex, etc.) y el enlace. Si la documentación **no lo dice**, escribir «no documentado», probarlo con una prueba mínima y registrar evidencia; no suponer.
4. Si la documentación **contradice** un plan, o un componente no funciona como el plan asume, no cambiar el diseño por cuenta propia: aplicar la sección 8.

## 8. Cuando algo no funciona (aviso al grupo)

Detener la ficha afectada y preparar el aviso con la plantilla del plan [EN] sección 4.1 (E3). La persona lo envía **el mismo día al grupo de WhatsApp**; **Juan Cruz decide**. Obligan a avisar, como mínimo: el hook no funciona en Claude o en Codex; falla una prueba bloqueante ([EN] Q3, Q6, Q11); la documentación contradice el plan; una versión nueva cambia lo ya probado; un paso exige tocar `backup`, `filemerge` o `pipeline`; una dependencia que se creía mergeada no lo está. Se puede seguir solo con trabajo independiente y autorizado.

## 9. Cerrar la ficha: informe y siguiente paso

Cuando se cumplen los criterios:

1. Verificar todos los criterios con evidencia y dejar el traspaso `docs/equipo/fichas/R-xx.md` completo.
2. Generar el **informe de cierre** `docs/equipo/informes/R-xx.md` desde la [plantilla](equipo/informes/PLANTILLA.md). Es **corto** (menos de una pantalla): qué ficha, enlace al PR y repo, qué se hizo, cómo verificarlo, **qué tiene que hacer el dueño** (revisar el PR, decidir algo, aceptar), a quién desbloquea y decisiones que necesitan al dueño.
3. Con el pedido de la persona: commit, push y PR (informe incluido). Decirle: **«Pasale este archivo a Juan Cruz por el grupo»**.
4. Actualizar el «Punto de reanudación» del traspaso, guardar lo durable en Engram y **recomendar compactar**.
5. Informar la **siguiente ficha** de la persona, de quién espera y qué debe ocurrir antes (por ejemplo: «tu próxima ficha es R-12; espera a que el dueño acepte tu PR de R-11»). **No empezarla** hasta que el estado figure `mergeada`.

## 10. Para el dueño (Juan Cruz)

1. Recibís el informe `.md` por el grupo. Revisás el PR (en el repo que indique).
2. Si hay algo que decidir, el informe lo dice. Si la ficha es de governance ALTO, revisás **la propuesta antes** de que se codifique.
3. Aceptás/mergeás el PR y, en un commit propio, actualizás **[estado-equipo.md](estado-equipo.md)**: estado `mergeada`, PR, SHA de merge y nota. Eso es lo que **desbloquea** las fichas siguientes.
4. Avisás al grupo a quién le toca, por ejemplo: «R-11 mergeada, Matías puede seguir con R-12».
5. Si tenés que habilitar una ficha adelantándola, lo escribís en «Autorizaciones de excepción».
6. Las puertas **R-05, R-20, R-24, R-25** son tuyas; las pruebas E2E y los pilotos requieren tu autorización expresa.

## 11. Mensaje listo para el grupo

> Equipo: ya está la guía en `main`. Bajen el repo, abran Claude Code o Codex y escriban **«Soy <su nombre>, quiero implementar lo mío»**. El agente los guía. **Arrancan dos en paralelo:** Matías con R-01 (prueba de Engram Cloud, con apoyo de Giuliano) y Emiliano con R-06 (contrato del modo equipo). Gianfranco, tu agente te va a decir que esperes a que Emiliano termine. Al terminar, el agente arma un `.md`: pásenmelo acá y yo reviso y acepto su PR.
