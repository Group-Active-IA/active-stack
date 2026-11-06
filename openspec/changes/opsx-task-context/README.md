# Ficha 05 · `opsx-task-context` — empezá por acá

**Lectura dirigida y herramientas por tarea** · Emiliano · ambos clientes · P1 · gobernanza **MEDIO**
Rama `propose/opsx-task-context` · base `1b4639b` · alcance de esta ronda: **solo propose**
Depende de **03** y **04**. Habilita **07**. Hermana de **06**: ninguna depende de la otra.

**Seis puntos que el grupo tiene que cerrar**, en `design.md`, cada uno con opciones,
recomendación y en qué se apoya. Dos no son de esta ficha: el inventario de módulos (lo produce
la 04) y los agentes incluidos (lo define el roadmap).

Ficha original: [CHANGES.md § 05](../../../CHANGES.md#change-05).

| Archivo | Qué contesta |
|---|---|
| `proposal.md` | por qué hay que hacerlo, con la medición de partida |
| `specs/` | qué tiene que hacer el sistema — 26 reglas con sus escenarios, en 5 capacidades |
| `design.md` | cómo, y por qué así: 16 decisiones con su alternativa descartada, y los seis puntos que el grupo tiene que cerrar |
| `tasks.md` | en qué orden y cómo se verifica cada paso — 58 tareas, todas pendientes |
| `research-notes.md` | de dónde salió todo esto: las cuatro fuentes y los hallazgos |

## Reglas duras de esta ronda

- **NUNCA especificar subcomandos nuevos del binario** (`active-stack run`, `active-stack
  state`) → el resumen de salida se especifica como **contrato de formato**. *Alcance del
  instalador* del roadmap deja los comandos de runtime nuevos sujetos a decisión de
  arquitectura separada (TBD), y `ARCHITECTURE.md:165` advierte que un envoltorio de estado
  no debe crear una segunda autoridad frente a OpenSpec.
- **NUNCA escribir contenido en `internal/harness/config/assets/*/sdd-orchestrator.md`** →
  ese archivo lo tocan seis fichas (02, 03, 04, 05, 07, 09) y no tiene dueño declarado. Su
  reestructuración en módulos es de 03 y 04. Esta ficha escribe **requisitos** sobre él.
- **NUNCA adoptar 1.500 tokens ni 40 líneas como criterio de diseño** → el informe
  ejecutivo los marcó como presupuestos no validados. La frontera núcleo/módulo se define
  **por consecuencia**, no por tamaño (`design.md` — D2). Como hipótesis a medir con 02, sí.
- **NUNCA definir métricas propias de consumo de modelo** → eso es la ficha 02. Los números
  de esta ficha son de **tamaño de texto** y deben etiquetarse como tales.
- **NUNCA tocar archivos rastreados para medir** → el test de inventario entra como tarea
  (grupo 1), no como artefacto de propose. `git status` debe quedar limpio en esta rama.
- **NUNCA cerrar la propuesta con un `(TBD)` que cambie alcance, diseño o aceptación** → se
  resuelve en el grupo antes de enviarla a revisión (guía del roadmap, paso 4). Si marcás
  `(TBD)`, documentá su impacto.
- **SIEMPRE que una decisión afecte otra ficha** → compartí en el grupo el ID, el cambio
  propuesto, las opciones y las fichas afectadas, y acordá antes de cerrar.
- **NUNCA cerrar de palabra lo que pasó por el grupo** → toda duda, modificación, pedido o
  respuesta se asienta en `tasks.md` — *Workflow follow-up* con fecha,
  con quién se acordó y en qué artefacto quedó incorporada.
- **NO confundir consultar con dejar abierto** → consultar es amplio: va al grupo *toda duda o
  modificación de stack o de implementación*. Lo **excepcional** es dejar algo abierto como
  `(TBD)`, y eso exige acuerdo del grupo más su impacto documentado. Hoy esta ficha tiene
  **cero `(TBD)` propios**.
- **NUNCA suponer lo que aportan las fichas anteriores** (01, 02, 03, 04) → lo que la 05
  necesita de cada una está pedido explícitamente en la sección B de ese registro. Si algo no
  llegó, se arranca con el estado actual y se declara el punto donde se re-corre.

## Mapa ("Necesito X → Leer Y")

| Necesito… | Leer / mirar |
|---|---|
| La ficha original y sus criterios de cierre | [CHANGES.md § 05](../../../CHANGES.md#change-05) |
| El motivo y la medición de partida | `proposal.md` |
| Las decisiones y por qué se descartó cada alternativa | `design.md` |
| Los requisitos y escenarios verificables | `specs/` |
| Qué hay que hacer y en qué orden | `tasks.md` |
| El análisis de las cuatro fuentes y los hallazgos con su evidencia | `research-notes.md` |
| Qué pasó por el grupo, qué respondió y qué se le pide a 01–04 | `tasks.md` — *Workflow follow-up* |
| El mecanismo que esta ficha quiere seleccionar por acción | `internal/harness/config/compose.go` |
| Dónde se congela la composición hoy | `internal/harness/config/install.go:53` |
| Los toggles declarados | `internal/catalog/harnesses.yaml:70` |
| Los fragmentos aditivos | `internal/harness/config/assets/{governance,engram-protocol,strict-tdd}.md` |
| Instalación y verificación de skills (frontera de herramientas por tarea) | `internal/harness/skill/` |

