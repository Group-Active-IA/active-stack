# Ficha 05 · `opsx-task-context` — empezá por acá

**Lectura dirigida y herramientas por tarea** · Emiliano · ambos clientes · P1 · gobernanza **MEDIO**
Rama `propose/opsx-task-context` (PR #10) · con `main` integrado (`d5e95bd`, con 03 y 04) · estado: **apply hecho con alcance parcial; sin archivar.** La deuda aceptada está en `informes/deuda-aceptada-05-2026-11-13.md`
Depende de **03** y **04**. Habilita **07**. Hermana de **06**: ninguna depende de la otra.

Revisado el 2026-11-13 contra el informe unificado de revisión OPSX (11/11), la matriz común de
reglas de 03 y 04 (12/11) y pruebas reales con Claude Code y Codex (`entrega-03-04.md` §5). Cada decisión de `design.md` lleva su estado —**acordado**, **propuesto
a <ficha>** o **de la 05**— y lo que todavía depende de otra ficha está en `design.md` —
*Coordinación pendiente*.

Ficha original: [CHANGES.md § 05](../../../CHANGES.md#change-05).

| Archivo | Qué contesta |
|---|---|
| `proposal.md` | por qué hay que hacerlo, con la medición de partida |
| `specs/` | qué tiene que hacer el sistema — 41 reglas con 107 escenarios, en 5 capacidades |
| `design.md` | cómo, y por qué así: 23 decisiones con su estado y su respaldo (medido, documentado o no probado), la coordinación pendiente, los puntos de grupo y las decisiones para el grupo |
| `entrega-03-04.md` | lo que la 05 entrega a 03 y 04: los cuatro módulos en inglés, la tabla acción → módulos, el contrato de captura por referencia, las pruebas reales del 13/11 y qué falta en código por ficha. Estado: textos y tabla en código (`internal/harness/config/assets/modules/`, catálogo); la integración en la skill de Claude es deuda D05-0605 y la actualización de las copias de Codex, D05-0610 |
| `tasks.md` | en qué orden y cómo se verifica cada paso: 64 tareas, 54 hechas y 10 en deuda aceptada, cada una con su ID |
| `research-notes.md` | de dónde salió todo esto: las cuatro fuentes, los hallazgos y el contraste con la documentación oficial de Claude Code y Codex |

## Reglas duras de esta ronda

- **NUNCA especificar subcomandos nuevos del binario** (`active-stack run`, `active-stack
  state`) → el resumen de salida se especifica como **contrato de formato**, y la captura
  independiente vive en el arnés `eval/taskcontext/` o es una fuente del cliente (`design.md` — D8). Los *Límites
  compartidos* del roadmap (`CHANGES.md:49`) dejan un runtime nuevo sujeto a una decisión de
  arquitectura separada y el arnés de evaluación fuera del binario. El informe ejecutivo
  advierte además que un envoltorio de estado no debe crear una segunda autoridad frente a
  OpenSpec, que `ARCHITECTURE.md:165` declara fuente de verdad del estado.
- **NUNCA escribir contenido en `internal/harness/config/assets/*/sdd-orchestrator.md`** →
  ese archivo lo tocan seis fichas (02, 03, 04, 05, 07, 09) y no tiene dueño declarado. Su
  partición en capas es de 03 y 04 (matriz común). Esta ficha **redacta sus cuatro módulos** en
  `entrega-03-04.md` (en código, `internal/harness/config/assets/modules/` en apply) y propone la
  tabla de carga, para acordarla con 03, 04 y 07, y los entrega a 03 y 04 (`design.md` — D23).
- **NUNCA adoptar 1.500 tokens ni 40 líneas como criterio de diseño** → el informe
  ejecutivo los marcó como presupuestos no validados. La frontera núcleo/módulo se define
  **por consecuencia**, no por tamaño (`design.md` — D2). Como hipótesis a medir con 02, sí.
- **NUNCA definir métricas propias de consumo de modelo** → eso es la ficha 02. Los números
  de esta ficha son de **tamaño de texto** y deben etiquetarse como tales.
- **NUNCA tocar `internal/backup`, `internal/filemerge` ni `internal/pipeline`** → son ALTO
  (`CLAUDE.md` §4). Si una tarea los necesitara, se detiene y espera revisión humana fechada.
- **NUNCA cerrar la propuesta con un `(TBD)` que cambie alcance, diseño o aceptación** → se
  resuelve en el grupo antes de enviarla a revisión (`CHANGES.md:16`, paso 4). Si marcás
  `(TBD)`, documentá su impacto.
- **SIEMPRE que una decisión afecte otra ficha** → compartí en el grupo el ID, el cambio
  propuesto, las opciones y las fichas afectadas, y acordá antes de cerrar.
- **NUNCA cerrar de palabra lo que pasó por el grupo** → toda duda, modificación, pedido o
  respuesta se asienta en `tasks.md` — *Workflow follow-up* con fecha, estado,
  con quién se acordó y en qué artefacto quedó incorporada.
- **NO confundir consultar con dejar abierto** → consultar es amplio: va al grupo *toda duda o
  modificación de stack o de implementación*. Lo **excepcional** es dejar algo abierto como
  `(TBD)`, y eso exige acuerdo del grupo más su impacto documentado. Hoy esta ficha tiene
  **cero `(TBD)` propios**: cada supuesto está declarado junto con su decisión y con cuándo se
  recalibra, y lo que depende de otra ficha figura en *Coordinación pendiente*.
- **NUNCA suponer lo que aportan las fichas anteriores** (01, 02, 03, 04) → lo que la 05
  necesita de cada una está pedido explícitamente en `tasks.md` — *Workflow follow-up*. Si algo no
  llegó, se arranca con el estado actual y se declara el punto donde se re-corre.

## Mapa ("Necesito X → Leer Y")

| Necesito… | Leer / mirar |
|---|---|
| La ficha original y sus criterios de cierre | [CHANGES.md § 05](../../../CHANGES.md#change-05) (`CHANGES.md:154-176`) |
| Qué entrega la 05 a 03 y 04 y las pruebas reales | `entrega-03-04.md` |
| La matriz común de reglas por capa (03 y 04) | `origin/claude-native-opsx:openspec/matriz-reglas-opsx.md` |
| Qué se implementa ya y qué está bloqueado | `tasks.md` — *Resumen de bloqueos* |
| El arnés de medición de la 02 que reutiliza esta ficha | `eval/baseline/README.md` |
| El motivo y la medición de partida | `proposal.md` |
| Las decisiones y por qué se descartó cada alternativa | `design.md` |
| Los requisitos y escenarios verificables | `specs/` |
| Qué hay que hacer y en qué orden | `tasks.md` |
| El análisis de las cuatro fuentes y los hallazgos con su evidencia | `research-notes.md` |
| Qué dice la documentación oficial de los clientes y qué cambió por eso | `research-notes.md` § 6 |
| Qué hay que enviar al grupo y qué se le pide a cada ficha | `tasks.md` — *Workflow follow-up* |
| El mecanismo que esta ficha quiere seleccionar por acción | `internal/harness/config/compose.go` |
| Dónde se congela la composición hoy (Claude y Codex) | `internal/harness/config/install.go:70` (rama `default`) |
| Dónde compone OpenCode | `internal/harness/config/install.go:53` (rama `ConfigDeliveryPrimaryAgent`) |
| Los toggles declarados | `internal/catalog/harnesses.yaml:70` |
| Los fragmentos aditivos | `internal/harness/config/assets/{governance,engram-protocol,strict-tdd}.md` |
| El piso de permisos de Claude (rutas, no valores) | `internal/harness/config/permissions/overlays.go:16-28` |
| Instalación y verificación de skills (frontera de herramientas por tarea) | `internal/harness/skill/` |
