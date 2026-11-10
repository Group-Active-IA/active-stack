# Ficha 05 · `opsx-task-context` — empezá por acá

**Lectura dirigida y herramientas por tarea** · Emiliano · ambos clientes · P1 · gobernanza **MEDIO**
Rama `propose/opsx-task-context` · base `1b4639b` · alcance de esta ronda: **solo propose**
Depende de **03** y **04**. Habilita **07**. Hermana de **06**: ninguna depende de la otra.

**Diez decisiones tomadas, pendientes de revisión del grupo**, en `design.md`, cada una con en qué
se apoya, qué fichas afecta y qué cambia si se rechaza. Todas están decididas acá; se comparten
porque tocan alcance de otras fichas.

Ficha original: [CHANGES.md § 05](../../../CHANGES.md#change-05).

| Archivo | Qué contesta |
|---|---|
| `proposal.md` | por qué hay que hacerlo, con la medición de partida |
| `specs/` | qué tiene que hacer el sistema — 28 reglas con sus escenarios, en 5 capacidades |
| `design.md` | cómo, y por qué así: 19 decisiones con su alternativa descartada, y los diez puntos para la revisión del grupo |
| `tasks.md` | en qué orden y cómo se verifica cada paso — 62 tareas, todas pendientes |
| `research-notes.md` | de dónde salió todo esto: las cuatro fuentes y los hallazgos |

## Reglas duras de esta ronda

- **NUNCA especificar subcomandos nuevos del binario** (`active-stack run`, `active-stack
  state`) → el resumen de salida se especifica como **contrato de formato**. Los *Límites
  compartidos* del roadmap (`CHANGES.md:49`) dejan un runtime nuevo sujeto a una decisión de
  arquitectura separada y el arnés de evaluación fuera del binario. El informe ejecutivo
  advierte además que un envoltorio de estado no debe crear una segunda autoridad frente a
  OpenSpec, que `ARCHITECTURE.md:165` declara fuente de verdad del estado.
- **NUNCA escribir contenido en `internal/harness/config/assets/*/sdd-orchestrator.md`** →
  ese archivo lo tocan seis fichas (02, 03, 04, 05, 07, 09) y no tiene dueño declarado. Su
  reestructuración en módulos es de 03 y 04. Esta ficha escribe **requisitos** sobre él y
  **redacta el texto de sus módulos** como entrega a 03 y 04, que lo ubican (`design.md` — D17).
- **NUNCA adoptar 1.500 tokens ni 40 líneas como criterio de diseño** → el informe
  ejecutivo los marcó como presupuestos no validados. La frontera núcleo/módulo se define
  **por consecuencia**, no por tamaño (`design.md` — D2). Como hipótesis a medir con 02, sí.
- **NUNCA definir métricas propias de consumo de modelo** → eso es la ficha 02. Los números
  de esta ficha son de **tamaño de texto** y deben etiquetarse como tales.
- **NUNCA tocar archivos rastreados para medir** → el golden de inventario entra como tarea
  (grupo 1), no como artefacto de propose. `git status` debe mostrar cambios solo en esta carpeta.
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
  **cero `(TBD)` propios** y ninguna pregunta abierta: cada supuesto está declarado junto con su
  decisión y con cuándo se recalibra.
- **NUNCA suponer lo que aportan las fichas anteriores** (01, 02, 03, 04) → lo que la 05
  necesita de cada una está pedido explícitamente en `tasks.md` — *Workflow follow-up*. Si algo no
  llegó, se arranca con el estado actual y se declara el punto donde se re-corre.

## Mapa ("Necesito X → Leer Y")

| Necesito… | Leer / mirar |
|---|---|
| La ficha original y sus criterios de cierre | [CHANGES.md § 05](../../../CHANGES.md#change-05) (`CHANGES.md:154-176`) |
| El motivo y la medición de partida | `proposal.md` |
| Las decisiones y por qué se descartó cada alternativa | `design.md` |
| Los requisitos y escenarios verificables | `specs/` |
| Qué hay que hacer y en qué orden | `tasks.md` |
| El análisis de las cuatro fuentes y los hallazgos con su evidencia | `research-notes.md` |
| Qué hay que enviar al grupo y qué se le pide a cada ficha | `tasks.md` — *Workflow follow-up* |
| El mecanismo que esta ficha quiere seleccionar por acción | `internal/harness/config/compose.go` |
| Dónde se congela la composición hoy (Claude y Codex) | `internal/harness/config/install.go:70` (rama `default`) |
| Dónde compone OpenCode | `internal/harness/config/install.go:53` (rama `ConfigDeliveryPrimaryAgent`) |
| Los toggles declarados | `internal/catalog/harnesses.yaml:70` |
| Los fragmentos aditivos | `internal/harness/config/assets/{governance,engram-protocol,strict-tdd}.md` |
| El piso de permisos de Claude (rutas, no valores) | `internal/harness/config/permissions/overlays.go:16-28` |
| Instalación y verificación de skills (frontera de herramientas por tarea) | `internal/harness/skill/` |
