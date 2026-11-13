# Respuesta a los bloqueantes de la revisión del PR #9

13 de noviembre de 2026. Alcance: exclusivamente cambio 04, `codex-native-opsx`. Este documento responde al [informe aportado de Juani](</Users/matiastorres/Downloads/revision-pr9-codex-native-opsx.md>) (archivo local `revision-pr9-codex-native-opsx.md`); no publica una respuesta en GitHub, no acepta deuda por el equipo y no implementa 03.

## Estado y referencias exactas

- 04 inspeccionado: `cd4aa3748cbd4642814be99becf965a58242f2ca`, rama `propose/codex-native-opsx`, PR [#9](https://github.com/Group-Active-IA/active-stack/pull/9), borrador contra main `35a89f8409eaacc3f5176de55ea96ba6564d3536`.
- 03 inspeccionado en lectura: `55f66de2566a89462de1eed4050c7e84b099e20a`, referencia local `origin/claude-native-opsx`, PR [#8](https://github.com/Group-Active-IA/active-stack/pull/8). Esta revisión es posterior a `ed37113`, citado por informes anteriores. No se hizo merge ni cherry-pick.
- Revisión externa aportada: Juani, Linux Pop!_OS, checkout aislado de `cd4aa37`, 13/11. Sus resultados se atribuyen a ese informe y no a una ejecución nueva del presente análisis.
- [Tasks de 04](../tasks.md): **11/21 completas y 10 pendientes**. Los escenarios pendientes siguen vigentes. No se modifican sus checks ni su alcance con esta respuesta.

## B1 · Circuito de cierre y tareas pendientes

**Respuesta: el bloqueo para declarar terminado 04 es válido. No se archiva un change con 11/21 tareas como si estuviera terminado. Tampoco se convierte lo pendiente en deuda aceptada por la sola existencia de este informe, por la aprobación de tareas independientes ni por el archivo que haya realizado otro change.**

La revisión pide implementar y archivar antes del PR; el procedimiento comunicado por Matías describe propuesta, apply, verificación y PR con evidencia, seguido de traspaso tras merge. Esa diferencia de orden debe aclararse con el equipo si afecta su circuito. Ninguno de ambos textos justifica marcar aceptación no ejecutada ni archivar anticipadamente el 04.

El PR permanece en borrador mientras se completan los requisitos. Si el equipo decide integrar una entrega parcial, necesita una decisión humana explícita que delimite lo entregado, qué cambia en la aceptación, responsables y seguimiento de lo restante. Hasta que exista, **no hay aceptación de deuda ni aprobación de reducción del alcance**. El objetivo vigente sigue siendo terminar el cambio 04 completo.

### Las diez tareas pendientes y su comprobación de cierre

| Tarea | Tipo de pendiente | Evidencia que ya existe | Necesidad verificable para cerrarla | Coordinación |
|---|---|---|---|---|
| 1.1 · Contrato y matriz | Acuerdo técnico + integración | Columna Codex completada; Juani la considera alineada sin cambios de capa. | Registrar acuerdo sobre IDs, rutas, targets y evidencia directa por acción; conciliar contratos de código y probar ambos clientes. La revisión favorable de columna no demuestra despacho conjunto. | Matías + Juani; 06 para fila de evidencia |
| 1.2 · Sondas P2 | Aceptación con clientes reales | Descubrimiento del paquete Codex en macOS; protocolo preparado. | CLI y extensión real en Windows/Linux, con versiones, mecanismo de roles, política explícita, contexto, controles y reanudación. Pedido implícito y manual por celda. | Matías + equipo con entornos |
| 1.3 · Snapshot y reversión | Revisión humana compartida | Tests de manifiesto recargado y estado esperado persistente; revisión independiente técnica; Juani prueba preservación del instalador en Linux. | Revisar con 01/03/08 diff y reversión posterior protegida y registrar disposición humana. El install/uninstall del informe no cubre por sí solo toda esa revisión. | Matías + 01/03/08 |
| 1.4 · Contratos 05/06/07 | Entrega externa + integración | Referencias normativas y sobre de retorno identificado. | Fijar revisiones corregidas, captura independiente y recibos con estado exacto del código; mismo estándar de evidencia directo/delegado. | Matías; Emiliano 05, Gianfranco 06, Neyen 07 |
| 2.4 · Especialistas | Implementación propia de 04, condicionada a P2 | Diseño de candidatos y fallback. | Entregar lo demostrado por P2; probar TOML si corresponde, rutas con espacios/Windows y fallback genérico con sobre 07/06. Cambios nuevos de configuración se revisan antes según gobernanza. | Matías + contratos externos |
| 2.5 · OPSX verify y evidencia | Implementación e integración de 04 | Verify estructural del instalador implementado; reglas de independencia documentadas. | Apply directo, revisión breve independiente, revisor ausente y recibo coherente sin corroboración que permanezca sin verificar. No sustituirlo por `active-stack verify`. | Matías + 05/06/07 |
| 4.2 · Matriz funcional | Aceptación de la entrega final | Pruebas Go y recorrido del binario en Linux; no Codex real. | Repetir escenarios normales, explícitos, especialista, verify, módulo/control obligatorio ausente, protección equivalente y control opcional ausente en CLI/IDE Windows/Linux. | Matías + entornos |
| 4.3 · Comparación con 02 | Medición | Protocolo y arnés parciales, aún sin comparación acreditada de 04. | Corrida normal/coordinada por interfaz con permisos equivalentes, aprobaciones y métricas ausentes registradas. Terminal integrada ejecutando CLI se clasifica como CLI. | Matías + Giuliano 02 |
| 4.4 · Regresiones y revisión | Verificación de revisión final + disposición | Tests afectados, vet y race documentados; Juani informa mismos cinco tests principales fallidos que main en Linux y ningún fallo nuevo. | Repetir checks pertinentes tras cambios finales/integración, registrar RED/GREEN y revisión de preservación. Mantener suite fallida como tal; disponer los fallos preexistentes con sus responsables sin atribuirlos a 04. | Matías + revisores; dueños de fallos ajenos |
| 4.6 · Revisión conjunta 03/04 | Coordinación humana y aceptación común | Revisión de Juani sobre 04 y matriz; identifica B2. | Concluir conciliación, ejecutar escenarios comunes en resultado combinado y registrar disposición humana de migración y pendientes. No está cerrada mientras B2 siga abierto. | Matías + Juani y revisores |

La [revisión independiente de 04](revision-independiente-04-2026-11-13.md), el [informe de integración](verificacion-integracion-2026-11-13.md) y el [plan de aceptación](plan-aceptacion-clientes-04.md) aportan evidencia y protocolo; no sustituyen los criterios que siguen abiertos. El [cruce del informe unificado y roadmap](cruce-cierre-04-preparacion-09-2026-11-13.md) explica la secuencia hacia el 09, que no se inicia como adopción mientras falte su referencia de 07.

## B2 · Conciliación con PR #8

**Respuesta: el bloqueo de integración está identificado y no se presenta como resuelto.** Juani reporta un merge de prueba abortado con once conflictos. En esta respuesta se inspeccionaron ambas ramas en lectura; no se repitió ni lanzó un merge. Un PR limpio contra main, por separado, no prueba compatibilidad entre ambos PR.

Los once archivos informados son:

| Archivo | Contrato o contenido que debe conciliarse |
|---|---|
| `internal/model/harness.go` | Un solo `ConfigDeliveryManagedSkill`, valor 2; preservar `Instructions=0` y `PrimaryAgent=1`. |
| `internal/harness/config/types.go` | Capacidad `SkillsDir`, avisos y modos de 03, y estados de escritura/rollback de 04. |
| `internal/harness/config/install.go` | Entrada compatible y selección explícita por agente y target. |
| `internal/install/plan.go` | Rutas resueltas por adapters, target, home de respaldo, preflight y conjunto exacto de escrituras. |
| `internal/install/steps.go` | Opciones, avisos y resultados parciales antes de devolver error; guardas de rollback. |
| `internal/uninstall/steps.go` | Retiro según cliente, propiedad y target; conservar estrategia protegida de 04. |
| `internal/verify/harness_checks.go` | Verificación según cliente/target, policy y marker propios. |
| `openspec/matriz-reglas-opsx.md` | Una matriz común; conservar aporte de cada responsable y resolver evidencia directa. |
| `AGENTS.md` | Descripción compartida de ambos recorridos sin borrar reglas ajenas. |
| `CLAUDE.md` | Mantener espejo con AGENTS, preservando la descripción de ambas entregas. |
| `ARCHITECTURE.md` | Documentar la combinación final y sus límites, sin declarar aceptación no ejecutada. |

### Contrato propuesto para revisar con Juani

Esta tabla es una **propuesta de conciliación**, todavía sin acuerdo registrado ni implementación. La coincidencia numérica del enum no hace intercambiables las dos estrategias de entrega.

| Área | Estado observado en 03 `55f66de` y 04 `cd4aa37` | Contrato que debe preservar la integración |
|---|---|---|
| Enum y discriminación | Ambos agregan `ManagedSkill=2`. 03 dirige ese modo a `installManaged` de Claude; 04 lo dirige a `InstallManagedSkill` de Codex. | Compartir enum, pero discriminar por `Agent()` y target. No llamar detección/composición Claude para Codex ni ownership/policy Codex para Claude. Un cliente no soportado no debe tomar silenciosamente la estrategia del otro. |
| Entrada de instalación | 03 mantiene `Install` como wrapper a `InstallWith`; opciones `Home`, `Target`, `ShadowSkillsDir`, `DetectVersion` y `Layers`. 04 usa `Install` y capacidad opcional `ManagedSkillAdapter`. | Conservar wrapper compatible y opciones de 03, con comportamiento Codex aislado. `DetectVersion` y `Layers` no gobiernan Codex. Las rutas vienen del adapter y del target, sin paths hardcodeados. |
| AgentAdapter y target | 03 requiere `SkillsDir` en el contrato local. 04 exige esa capacidad solo al tomar entrega nativa y usa trait `InstallTarget()` para aviso Project. | Decidir contrato final explícitamente y adaptar dobles/tests. Conservar el error de capacidad requerida ausente y Project Codex legado con diagnóstico; Claude Project conserva su estrategia de 03 y su chequeo de shadowing. |
| Resultado | 03: `Files`, `AllAlready`, `Warnings`, `Modes`. 04: esos campos salvo `Modes`, más `ManagedWrites`. | Unión de datos, sin perder avisos ni modos y sin usar un modo Claude para inferir entrega Codex. Preservar `ManagedWrites` de todas las escrituras realizadas aun si falla un adapter posterior; `AllAlready` no debe anunciar éxito ante error. |
| Pipeline y rollback | 04 captura padres y aplica guardas con existencia/ausencia esperadas antes de retornar el error. 03 retorna error antes de reportar resultado. | Aplicar guardas de 04 también ante resultados parciales en la interfaz combinada. Mantener protección de ediciones concurrentes, padres sustituidos, archivos ajenos y primer snapshot persistente. No degradar ni suprimir una guarda al incorporar `InstallWith`. |
| Avisos | 03 usa eventos `StepStatusDegraded` para notices; 04 escribe el aviso soportado de Project a stderr sin degradar. | Definir cómo se muestran límites no fatales y conservar su visibilidad. La conciliación no debe volver a calificar Project Codex soportado como fallo best-effort. Una mejora del canal común se acuerda sin alterar unilateralmente 03. |
| Verificación | 03 comprueba versión Claude, `disable-model-invocation: true`, módulos esperados y `universal-rules`. 04 comprueba ownership, inventario y `agents/openai.yaml` con `allow_implicit_invocation: false`. | Selección por cliente/target: Codex Machine usa comprobación Codex, Codex Project conserva legado; Claude usa su rama correspondiente. No buscar marker/policy Claude en Codex ni viceversa. Se preservan checks de otros harnesses. |
| Markers | Codex universal conserva ID `sdd-orchestrator`; Claude universal usa `universal-rules`, con limpieza del legado `sdd-orchestrator`. | Cada verify y uninstall reconoce los markers de su propia entrega y mantiene contenido propio y secciones hermanas como design-mockups. Mismo ID conceptual de harness no significa mismo marker universal. |
| Desinstalación y permisos | 03 tiene retiro por `removeLayered`, marca de dueño y respaldo local de instrucciones. 04 inventaría archivos y usa rollback protegido por estados esperados. | Dispatch por cliente/target; no aplicar removeLayered a Codex ni adoptar inventario Codex para Claude. En 04 conservar retiro archivo por archivo, directorios solo vacíos y notas ajenas. Preservar modos y rutas; el helper aditivo `WriteFileAtomicPreservingParent` puede evaluarse como mecanismo común sin implementar 03 dentro de 04. |
| Pruebas de combinación | No hay resultado combinado acreditado en este documento. | Instalar con Claude y Codex seleccionados en homes temporales; probar ambos órdenes y fallo del adapter posterior, reinstalar, actualizar, verify y retirar cada cliente sin afectar el otro. Afirmar compatibilidad solo con evidencia de la revisión integrada. |

### Orden propuesto, no acordado

1. Se consultó a Matías por el orden; al redactar esta respuesta no hay respuesta ni acuerdo registrado. La recomendación es que, si PR #8 recibe aceptación antes, se integre primero 03 por el circuito del equipo. Matías no implementa ni mergea ese trabajo en nombre de Juani.
2. Después de su merge, actualizar main y rebasar la rama de 04 sobre ese main, conciliando únicamente la integración necesaria para que 04 coexista con la entrega ya integrada. Preservar el trabajo de 03; cualquier corrección propia de Claude vuelve a su responsable.
3. Matías y Juani revisan los once archivos, el contrato anterior y el diff ALTO. Resolver el formato no es evidencia de comportamiento correcto.
4. Ejecutar pruebas aisladas de ambos clientes seleccionados y los recorridos de error/rollback, más regresiones de los demás clientes. Repetir OpenSpec estricto sobre 04 y registrar revisión exacta. Juani confirma preservación del recorrido de 03.
5. Mantener PR #9 en borrador hasta resolver B1 y B2 y sus criterios vigentes. La propuesta «entra primero quien se apruebe antes» permite revisar el orden si 04 llegara primero; no hay orden definitivo aprobado en este documento.

No se ejecutó rebase, merge ni resolución de conflictos para redactar esta propuesta. El paso siguiente de conciliación depende del acuerdo del equipo y del orden real de integración, no de una implementación paralela de 03 por Matías.

## Qué acredita la evidencia nueva de Linux

El informe de Juani sobre `cd4aa37` agrega validación independiente del **instalador** en Linux Pop!_OS: compiló la rama en su entorno y probó install, reinstalación, uninstall y colisión sobre homes temporales. Confirma texto propio y modo 0600 conservados, skill ajena intacta, rechazo de colisión y ausencia de fallos nuevos respecto de main. La modificación cosmética de líneas vacías de `config.toml` se reprodujo en main; no se atribuye a 04.

Juani declara expresamente **«No se ejecutó Codex real»**. Su prueba no cierra activación manual/implícita, contexto, especialistas, reanudación ni las celdas CLI/IDE de tareas 1.2 y 4.2. Las sondas reales de Claude documentadas por 03 son otra evidencia: Claude Code 2.1.318, `claude -p`, haiku y una skill testigo escrita a mano. No demuestran comportamiento de Codex ni contexto efectivo del paquete Codex generado por el instalador. Tampoco se extrapolan esas sondas a Windows o IDE.

## Disposición de observaciones adicionales

- **Rutas Windows:** revisar mensajes con `%q` antes de declarar Windows. La tarea requiere evaluar legibilidad de separadores sin confundirla con escape correcto de datos; su resultado y pruebas se registran aparte. Este documento no declara corregido ese punto.
- **Líneas vacías tras uninstall:** observación cosmética, opcional según la revisión. No retirar contenido propio para conseguir una limpieza estética ni ampliar alcance automáticamente.
- **Helper aditivo de permisos:** Juani confirma que no modifica `WriteFileAtomic`. Evaluar reutilización en integración; no usarla como permiso para corregir la desinstalación de 03 desde esta rama.

La revisión externa apoya la calidad de lo implementado en 04 y confirma dos pendientes de cierre. La respuesta conserva ambos abiertos con condiciones verificables; no archiva, no acepta deuda ni declara conciliados los PR.
