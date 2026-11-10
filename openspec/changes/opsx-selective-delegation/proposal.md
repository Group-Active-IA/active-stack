# Proposal: opsx-selective-delegation

> Change `opsx-selective-delegation` · ficha 07 del roadmap · Responsable: Neyen · Agentes: Claude y Codex · Prioridad P1 · Gobernanza MEDIO
> Base revisada: `1b4639b`. Depende de 05 (`opsx-task-context`) y 06 (`opsx-verifiable-evidence`); usa el protocolo de 02 (`opsx-baseline`, `240e1b2`). Habilita 09 (`89f1fa5`), que toma esta configuración como control.
> Revisión 2 (10/11/2026): incorpora la auditoría técnica del 10/11 y las decisiones confirmadas del grupo; ver `design.md`, "Cambios por la auditoría".
> Esta ronda es solo `propose`: todas las tareas de implementación quedan pendientes.

## Why

Las plantillas del orquestador convierten la delegación en una obligación, no en una decisión. El coordinador "delega el trabajo real", la tabla de umbrales decide por cantidad de archivos y ninguna regla define quién es dueño de qué archivo, cuándo dos workers pueden correr a la vez, cuántas veces se reintenta ni qué devuelve el worker. Delegar puede despejar el contexto principal, pero también suma consumo (cada worker arranca en frío y relee), reintentos y solapamiento entre agentes que editan lo mismo. Hoy nadie midió cuál de las dos cosas pesa más.

### Evidencia

Todo sobre `main` en `1b4639b`.

- **La delegación es incondicional en Claude.** `internal/harness/config/assets/claude/sdd-orchestrator.md:7` ("You are a COORDINATOR, not an executor… delegate real work"), `:14` ("Delegate, don't inflate"), `:98` ("For each action, delegate to a sub-agent") y `:208` ("NEVER do apply or propose work inline — ALWAYS delegate via Agent tool").
- **El umbral es un conteo fijo de archivos.** La tabla `## Delegation Rules` (Claude `:26-46`, Codex `:17-36`) decide por "1-3 archivos inline / 4+ delegar". El informe ejecutivo registra el mismo hallazgo ("La delegación actual usa un umbral fijo de archivos", Codex `:25`).
- **Apagar el toggle no apaga la obligación.** `internal/harness/config/compose.go:57-59` sólo quita la sección marcada `sdd-delegation` cuando `delegation` no está activo. Las líneas `:7`, `:14`, `:98-107` y `:208` de Claude, y `:14` y `:85` de Codex, están fuera de esos markers y sobreviven. Hoy no existe un modo "directo" real.
- **El retorno es prosa libre.** Claude `:154-159` pide "a summary of" cinco puntos en texto. Codex no especifica retorno. Ninguna plantilla define propiedad de archivos, límites de alcance, paralelismo ni tope de reintentos.
- **Se instala siempre.** `internal/catalog/harnesses.yaml:70` declara `toggles: [tdd, engram, model-routing, delegation, governance]` para `lite` y `full`.
- **No hay medición.** El informe ejecutivo dice "Sin benchmark de sesiones ni porcentaje de ahorro" y "Tokens no medidos", y pide "Separar paralelismo de ahorro de consumo" e "incluir todos los agentes" al comparar.

## What Changes

- **Independencia obligatoria primero.** Si la acción exige que quien revisa no sea quien construyó, la revisión va a otro agente y ninguna contraseñal lo anula. Si falta el módulo o un revisor habilitado, se detiene esa verificación y se explica.
- **Decisión explícita directo/delegado.** Fuera de ese caso, el coordinador serializa ante conflicto de archivos, resuelve directo ante una contraseñal (tarea breve, contexto vivo, relectura completa) y delega ante una señal (salida verbosa, exploración con conclusión, trabajo paralelizable). Por defecto, directo. La regla reemplaza al conteo de archivos y cada decisión queda registrada con el paso que la cerró.
- **La regla vive en la skill del coordinador; el módulo se carga sólo si se delega.** La regla de decisión (breve) va en la skill de invocación explícita acordada para 03/04. El módulo de delegación y coordinación se carga con el mecanismo de 05 únicamente cuando la decisión es delegar.
- **Encargo mínimo.** Toda delegación lleva objetivo, acción OPSX, conjunto de escritura, punteros de lectura, límites, criterios de salida, contrato de retorno y tope de intentos.
- **Propiedad contra un punto de partida.** Cada worker con escritura trabaja en su árbol, desde una partida registrada (commit base y huella del estado de trabajo), con un conjunto exclusivo de archivos. Se le atribuye sólo lo que cambió desde esa partida —commits, preparados, sin preparar, nuevos, borrados y renombres— y se rechaza lo que caiga fuera. Antes de integrar se comprueba el resultado combinado.
- **Paralelismo (decisión confirmada).** Hasta dos workers con escritura a la vez, cada uno en su árbol, sólo con alcances disjuntos, sin dependencias y con pruebas propias. Si no, se serializan.
- **Retorno estructurado.** Estado, partida, archivos modificados con su tipo de cambio, hallazgos, recibos de 06 y próximo paso. Lo declarado queda *sin verificar* hasta que una fuente independiente o una re-ejecución lo corrobore (decisión confirmada para 05/06).
- **Reintentos acotados.** Cada encargo tiene un tope; agotado, se escala a una persona. Los reintentos se cuentan en el costo.
- **Toggle `delegation` coherente.** Sin el toggle, el bloque compuesto no conserva ninguna obligación de delegar y el coordinador declara que trabaja sin subagentes.
- **Piloto controlado antes de adoptar.** El arnés de 02 compara ejecución directa, delegación selectiva y el flujo actual con la misma tarea, rúbrica, criterios y controles, por cliente e interfaz. Costo por tarea aceptada sobre el total de la cohorte; cuota, tokens y latencia separados; ahorro de costo y reducción de espera separados. Primero calibración; los parámetros de adopción se fijan antes de la prueba que decide.

## Capabilities

### New Capabilities

- `selective-delegation`: regla de decisión directo/delegado, encargo mínimo, propiedad de archivos, paralelismo, retorno estructurado, reintentos, semántica del toggle `delegation` y paridad entre Claude y Codex.
- `delegation-pilot`: protocolo del piloto que compara las modalidades, clases de tarea, métricas, contabilidad de principal, workers y reintentos, y criterio de adopción.

### Modified Capabilities

(ninguna: no existe un spec archivado de delegación en `openspec/specs/`)

## Alcance

- Contenido de la regla de decisión y de los módulos de delegación y de coordinación de subagentes.
- Encargo mínimo, propiedad de archivos, paralelismo, retorno y reintentos.
- Semántica del toggle `delegation` en la composición.
- Paridad de contenido y veredictos entre Claude y Codex.
- Diseño del piloto, sus fixtures y su criterio de adopción.

## Exclusiones

- **Mecanismo de carga de módulos por acción** → 05 (`task-context-loading`). Esta ficha lo consume.
- **Ubicación física de módulos y roles nativos** → 03 y 04. Esta ficha escribe el contenido, no el layout.
- **Esquema del recibo, procedencia y suficiencia** → 06 (`execution-receipts`). Esta ficha lo exige en el retorno.
- **Formato de salida acotada y referencia al original** → 05.
- **Modelo y esfuerzo por rol, caché de prefijos** → 09, que usa el flujo de esta ficha.
- **Controles ejecutables que bloqueen escrituras fuera de alcance** → 08. Acá la comprobación es posterior y la hace el coordinador contra el diff.
- OpenCode y Antigravity, igual que 05, 06 y 08, salvo acuerdo del grupo.
- Ledger firmado, runtime nuevo o subcomando en el binario: el arnés del piloto vive fuera del binario.
- Implementación: `apply`, ejecución del piloto, `verify` y `archive` son de una etapa posterior.

## Impact

- **Assets previstos** (nada se toca en esta ronda): `internal/harness/config/assets/{claude,codex}/sdd-orchestrator.md` o los módulos que resulten de 03/04; `internal/harness/config/compose.go` y `compose_test.go` (semántica del toggle y golden de paridad).
- **Catálogo**: `internal/catalog/harnesses.yaml` sólo si el grupo decide cambiar la descripción del toggle o sacarlo de `lite` (BAJO, validado por `catalog.Load()`).
- **No se tocan** `backup`, `filemerge` ni `pipeline` (ALTO). Si el diseño final obliga a tocarlos, la propuesta vuelve a revisión.
- **Gobernanza**: `harness/config` es MEDIO, implementación con checkpoints.
- **Otras fichas**: fija qué carga 05 cuando la acción delega, qué exige del recibo de 06 y qué flujo mide 09. Comparte con 03/04 el archivo `sdd-orchestrator.md`, que no tiene dueño declarado (ver design, decisiones de grupo).
- **Espejo obligatorio**: si se modifica la constitución del repo, `CLAUDE.md` y `AGENTS.md` se actualizan juntos.
