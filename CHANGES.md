# CHANGES — Roadmap de mejoras de Active Stack

> Leé este documento antes de abrir una rama. Es el plan canónico de mejoras de OPSX para Codex y Claude. El estado de cada change se consulta con OpenSpec. Esta ronda cubre **solo `propose`**, incluida la planificación de la implementación futura.

[Vista interactiva](https://active-stack-roadmap.vercel.app/) · [Informe ejecutivo y evidencia](https://active-stack-executive-site.vercel.app/) · [Roadmap post-starters archivado](docs/archive/CHANGES-2026-11-05-post-starters.md)

## Cómo trabajar en esta ronda

1. **Ubicá tu ficha y leé sus referencias.** Revisá [AGENTS.md](AGENTS.md) o [CLAUDE.md](CLAUDE.md), el alcance y las dependencias. Confirmá qué código existe, qué comportamiento se observó y qué falta investigar. Los identificadores 01–10 son para coordinar. El nombre en kebab-case identifica el change en OpenSpec.
2. **Creá una rama por change en Active Stack.** Con la rama principal actualizada, usá `git switch -c propose/<nombre-del-change>`. Si redactás dos propuestas en paralelo, usá un worktree por rama. El worktree mantiene separados los borradores de cada rama.
3. **Prepará los artefactos.** Consultá `openspec list --json` y usá `/opsx:propose` con el nombre de la ficha. Si el change existe, consultá `openspec status --change <nombre>` para continuar. Si el checkout todavía requiere inicialización de OpenSpec, coordiná la configuración con el equipo y usá el CLI. Completá:
   - `proposal.md` con problema, evidencia, alcance, exclusiones y capacidades afectadas.
   - `design.md` con el recorrido de la implementación, archivos o componentes, alternativas, decisiones, compatibilidad y riesgos.
   - `specs/` con requisitos y escenarios verificables de éxito y de fallo.
   - `tasks.md` con tareas futuras ordenadas, dependencias, pruebas y condiciones de aceptación. Dejá todas las tareas de implementación pendientes.
4. **Cerrá las decisiones mientras redactás.** Llevá al grupo de comunicación cualquier cambio de stack, diseño o implementación con el ID, las opciones y los changes afectados. Incorporá lo acordado a los artefactos. Reservá `(TBD)` para excepciones no bloqueantes acordadas por el grupo. Las decisiones que cambian alcance, diseño o aceptación se resuelven antes de compartir la propuesta final.
5. **Compartí la propuesta final cuando esté lista.** Confirmá los artefactos con `openspec status --change <nombre>`. Subí los cuatro artefactos de `openspec/changes/<nombre>/` a tu rama y compartí el enlace con acceso para los siete, junto con ID, revisión de partida y acuerdos principales. La revisión de los siete empieza con la entrega final de cada change y avanza independientemente de las demás.
6. **Cerrá la revisión.** El responsable integra los comentarios, comparte la versión revisada y registra el resultado del grupo. La implementación, la ejecución de su plan de pruebas y las acciones OPSX posteriores quedan para otra ronda.

## Responsables y dependencias

El reparto está propuesto por afinidad entre cambios. Las dependencias ordenan la **implementación futura**. La redacción se puede preparar en paralelo y los responsables acuerdan juntos los puntos compartidos.

| ID | Change | Responsable | Prioridad | Gobernanza | Depende de |
|---|---|---|---|---|---|
| 01 | [Compatibilidad del tier estricto](#change-01) | Juan Cruz | P0 | Alto | — |
| 02 | [Medición de referencia](#change-02) | Matías | P1 | Bajo | [01](#change-01) |
| 03 | [Roles nativos en Claude](#change-03) | Juani | P1 | Alto | [02](#change-02) |
| 04 | [Roles nativos en Codex](#change-04) | Giuliano | P1 | Alto | [02](#change-02) |
| 05 | [Lectura dirigida y herramientas por tarea](#change-05) | Emiliano | P1 | Medio | [03](#change-03), [04](#change-04) |
| 06 | [Evidencia ligada a la ejecución](#change-06) | Gianfranco | P1 | Medio | [03](#change-03), [04](#change-04) |
| 07 | [Delegación selectiva con contrato](#change-07) | Neyen | P1 | Medio | [05](#change-05), [06](#change-06) |
| 08 | [Controles efectivos por cliente](#change-08) | Gianfranco | P2 | Alto | [06](#change-06) |
| 09 | [Pilotos de modelos y caché](#change-09) | Matías | P2 | Medio | [07](#change-07) |
| 10 | [Juez externo en sombra](#change-10) | Neyen | P3 | Medio | [08](#change-08), [09](#change-09) |

**Ruta al primer piloto** · 01 → 02 → (03 y 04 en paralelo) → (05 y 06 en paralelo) → 07. Después de 06 puede prepararse la verificación de 08. La implementación de 09 usa los resultados de 07, y 10 queda condicionado a 08 y 09.

## Comportamiento objetivo

- **Activación explícita.** Active Stack instala y registra `sdd-orchestrator` para Claude y Codex. En una sesión normal, el orquestador está disponible para invocarlo y permanece inactivo. Una invocación explícita inicia el rol especializado. OpenSpec mantiene el estado de los changes.
- **Contexto global liviano.** Los `CLAUDE.md` y `AGENTS.md` globales contienen las reglas generales aplicables y, si hace falta, una referencia breve al orquestador. La migración retira el bloque global de SDD administrado por Active Stack, conserva las instrucciones y configuraciones del usuario y admite rollback.
- **Módulos según rol y acción.** TDD, delegación y coordinación de subagentes se entregan como módulos independientes. El orquestador carga los que necesita para la acción activa. Los especialistas reciben el encargo, las reglas aplicables y el contrato de retorno. Las reglas generales obligatorias siguen disponibles en todos los roles a los que apliquen.
- **Configuración nativa y feedback verificable.** Cada cliente usa sus mecanismos de configuración, rules, permisos y hooks según la función de cada control. El texto orienta al agente y los controles ejecutables verifican o hacen cumplir aquello que el cliente puede controlar. Los resultados de coordinación llevan un contrato estructurado de estado, hallazgos y evidencia. El resumen puede enlazar la salida original con acceso y retención definidos.
- **Disponibilidad comprobada por superficie.** Las propuestas 03 y 04 definen las superficies soportadas y cómo se invoca el rol en cada una. Una superficie sin soporte para el rol debe informar el límite y ofrecer un camino explícito si existe. La configuración resultante mantiene la sesión normal libre del bloque global de SDD.

**Aceptación transversal.** En cada superficie declarada, comparar una sesión normal con otra que invoque el orquestador. Verificar por separado un especialista y una acción que requiera TDD. Comprobar qué instrucciones y módulos recibe cada sesión, tanto en instalación limpia como después de actualizar una instalación anterior. Reinstalar y revertir sobre una configuración con contenido del usuario. Medir contexto y consumo por tarea aceptada con el protocolo de 02. El ahorro queda sujeto a medición. Acordar una matriz de entrega por cliente, modo de instalación y alcance de configuración. Si falta un módulo obligatorio, informar la limitación y detener la acción que lo requiere. Acordar una revisión conjunta de 03 y 04 con los mismos escenarios y criterios antes de promover la migración.

**Límites compartidos.** La ronda actual define propuestas. Cada responsable investiga cómo cumplir este comportamiento. Las propuestas de riesgo alto deben explicitar la revisión humana previa a cambios sobre configuración del usuario. Los pilotos comparan costo o cuota por tarea aceptada con calidad equivalente o superior. El arnés de evaluación queda fuera del binario. Un ledger firmado, un runtime nuevo o una integración permanente del juez requieren una decisión de arquitectura separada.

## Fichas de change

<a id="change-01"></a>

### 01 · Compatibilidad del tier estricto

**Change** `codex-strict-compat` · **Responsable** Juan Cruz · **Cliente** Codex · **Dependencias** —

**Problema.** El instalador genera `approval_policy = "untrusted"` para el tier estricto. Codex CLI 0.166.0 rechaza ese valor, así que una instalación o actualización puede dejar una configuración que el cliente no acepta.

El caso registrado en el informe se reproduce con `codex -c 'approval_policy="untrusted"' features list`. El cliente devuelve código 1 y señala que el valor dejó de estar soportado. El override es temporal y sirve para comprobar la compatibilidad del parser. La prueba de permisos efectivos requiere casos adicionales en un entorno aislado.

**Qué buscar.** Reconstruí qué escribe cada tier y dónde se respalda el archivo. Reproducí el rechazo solo con un override o un perfil temporal. Trazá si `update` pasa por el instalador de permisos. Si existe una ruta que lo omite, documentá cómo se migraría y quién la ejecutaría. La selección de política debe comprobarse con el cliente real; un chequeo de presencia de claves es insuficiente.

**Qué se implementaría.** Explicá qué garantías debe conservar el tier estricto y compará las políticas compatibles. Diseñá la migración del valor legado en instalación y actualización, con snapshot, preservación de claves y perfiles ajenos, reinstalación idempotente y recuperación ante fallo. Especificá escenarios de carga y permisos efectivos en entornos soportados.

**Qué acordar en grupo.** Qué versiones y entornos se soportan, qué operaciones debe permitir el tier y qué política compatible conserva esas garantías.

**Aceptación futura.** Config aceptada por las versiones soportadas (TBD). Probar instalación limpia, migración, reinstalación y rollback. Verificar permisos efectivos en los entornos soportados, con casos permitidos y denegados.

**Leer antes**

- [internal/harness/config/permissions/install.go](internal/harness/config/permissions/install.go)
- [internal/harness/config/permissions/codex_test.go](internal/harness/config/permissions/codex_test.go)
- [internal/filemerge/toml.go](internal/filemerge/toml.go)
- [internal/verify/harness_checks.go](internal/verify/harness_checks.go)
- [cmd/active-stack/update.go](cmd/active-stack/update.go)
- [Evidencia y recomendaciones del informe ejecutivo](https://active-stack-executive-site.vercel.app/#evidencia)

<a id="change-02"></a>

### 02 · Medición de referencia

**Change** `opsx-baseline` · **Responsable** Matías · **Cliente** Ambos · **Dependencias** [01](#change-01)

**Problema.** El informe mide el tamaño del texto, pero falta una referencia de consumo y calidad por tarea. Sin esa referencia, las mejoras posteriores no se pueden comparar.

**Qué buscar.** Revisá qué datos emiten los clientes por sesión y qué puede observar un arnés externo. Diferenciá telemetría del instalador de consumo del modelo. Si una métrica no está disponible, acordá cómo reportar esa limitación. La propuesta entrega un protocolo y una estructura de resultados; las corridas se planifican como tareas.

**Qué se implementaría.** Definí un arnés externo, tareas reproducibles, fixtures y criterios de aceptación. Diseñá el registro de versiones, entradas con y sin caché, salidas, workers, reintentos y tiempo humano. Prepará el plan de corridas sobre la configuración compatible de 01 y explicá cómo tratar métricas ausentes.

**Qué acordar en grupo.** Qué tareas, repeticiones, versiones, presupuesto y métricas usar. Acordá cómo distinguir consumo de API y cuota.

**Aceptación futura.** Repetir tareas comparables de exploración, cambio pequeño, cambio multiarchivo, diagnóstico y revisión. Registrar contexto, entrada fría/cacheada, salida, reintentos y tiempo humano por coordinador y worker. Separar API de cuota. Medir cuánto contexto carga una sesión normal y comparar ejecuciones equivalentes del flujo OPSX anterior y del flujo modular.

**Leer antes**

- [ARCHITECTURE.md](ARCHITECTURE.md)
- [internal/verify/report.go](internal/verify/report.go)
- [internal/harness/config/assets/claude/sdd-orchestrator.md](internal/harness/config/assets/claude/sdd-orchestrator.md)
- [internal/harness/config/assets/codex/sdd-orchestrator.md](internal/harness/config/assets/codex/sdd-orchestrator.md)
- [Evidencia y recomendaciones del informe ejecutivo](https://active-stack-executive-site.vercel.app/#plan)

<a id="change-03"></a>

### 03 · Roles nativos en Claude

**Change** `claude-native-opsx` · **Responsable** Juani · **Cliente** Claude · **Dependencias** [02](#change-02)

**Problema.** Las instrucciones compartidas de Claude mezclan reglas universales con el rol de coordinador. Eso dificulta saber qué contexto recibe cada especialista.

**Qué buscar.** Seguí `InstallSDD` o el punto equivalente del adapter hasta los archivos que ve Claude. Identificá alcance, precedencia y contenido efectivo en sesión y en especialistas. Investigá qué nombres y ubicaciones soporta la versión objetivo. Documentá la matriz de reglas compartidas con 04.

**Qué se implementaría.** Registrá el orquestador nativo para invocación explícita y separá las reglas universales, la definición del rol y los módulos por acción. Planificá la migración del bloque administrado en `CLAUDE.md` y de cualquier activación automática administrada por el stack. Conservá las preferencias del usuario y el rollback.

**Qué acordar en grupo.** Dónde se instala cada pieza, cómo se invoca el coordinador y qué reglas necesitan los especialistas. Coordiná con 04 el inventario de módulos, quién los carga y cuándo se aplican.

**Aceptación futura.** Una sesión normal conserva sus reglas generales. La invocación explícita activa el coordinador y sus módulos según la acción. El especialista conserva las reglas que le aplican. Probar precedencia, actualización, reinstalación y rollback con contenido del usuario.

**Leer antes**

- [internal/agents/claude/adapter.go](internal/agents/claude/adapter.go)
- [internal/harness/config/assets/claude/sdd-orchestrator.md](internal/harness/config/assets/claude/sdd-orchestrator.md)
- [internal/catalog/harnesses.yaml](internal/catalog/harnesses.yaml)
- [internal/install](internal/install)
- [internal/verify](internal/verify)
- [Evidencia y recomendaciones del informe ejecutivo](https://active-stack-executive-site.vercel.app/#evidencia)

<a id="change-04"></a>

### 04 · Roles nativos en Codex

**Change** `codex-native-opsx` · **Responsable** Giuliano · **Cliente** Codex · **Dependencias** [02](#change-02)

**Problema.** La plantilla de Codex presupone un coordinador dedicado, pero la entrega actual como instrucciones compartidas no registra ese rol ni demuestra cómo se activa en cada superficie.

**Qué buscar.** Seguí el adapter y la plantilla hasta la configuración efectiva. Contrastá sesión principal, roles nativos e invocación de skills en cada superficie declarada. Verificá la diferencia entre tener archivos instalados y activar el flujo en un proyecto real. Documentá la matriz de reglas compartidas con 03.

**Qué se implementaría.** Separá el arranque normal, el orquestador de invocación explícita, los especialistas y los módulos por acción. Planificá retirar el bloque administrado de `AGENTS.md` global preservando reglas del usuario. Documentá por superficie la invocación, el contexto por rol, la alternativa explícita y la migración reversible.

**Qué acordar en grupo.** Qué superficies se van a soportar, cómo arranca el coordinador en cada una y qué fallback mantiene OPSX disponible. Alineá reglas compartidas con 03.

**Aceptación futura.** Verificar sesión normal, invocación explícita y carga de módulos y reglas por rol. Probar migración idempotente y rollback con contenido del usuario. Validar CLI y cada superficie declarada por separado.

**Leer antes**

- [internal/agents/codex/adapter.go](internal/agents/codex/adapter.go)
- [internal/harness/config/assets/codex/sdd-orchestrator.md](internal/harness/config/assets/codex/sdd-orchestrator.md)
- [internal/catalog/harnesses.yaml](internal/catalog/harnesses.yaml)
- [internal/install](internal/install)
- [internal/verify](internal/verify)
- [Evidencia y recomendaciones del informe ejecutivo](https://active-stack-executive-site.vercel.app/#evidencia)

<a id="change-05"></a>

### 05 · Lectura dirigida y herramientas por tarea

**Change** `opsx-task-context` · **Responsable** Emiliano · **Cliente** Ambos · **Dependencias** [03](#change-03), [04](#change-04)

**Problema.** Una lectura amplia, una salida extensa o una integración innecesaria pueden consumir contexto y ocultar fallos importantes.

**Qué buscar.** Inventariá qué se carga por defecto y qué puede resolverse por tarea. Localizá dónde aparecen instrucciones largas o resultados verbosos. Proponé fixtures que coloquen fallos al inicio, medio y final de una salida grande y comprueben la recuperación del original.

**Qué se implementaría.** Definí reglas de lectura por ruta o símbolo y un formato de salida breve que retenga los fallos relevantes. Mapeá con 03, 04 y 07 cuándo se cargan TDD, delegación y coordinación de subagentes. Diseñá cómo recuperar la salida original y qué skills o herramientas habilita cada clase de tarea. Planificá fixtures con ruido y una comparación con y sin integraciones opcionales.

**Qué acordar en grupo.** Qué se conserva en un resumen, dónde queda el original y qué capacidad es necesaria para cada tarea. Coordiná la referencia al original con 06.

**Aceptación futura.** Fixtures con errores entre miles de líneas recuperan cada fallo relevante y su salida completa. Comparar contexto efectivo y aceptación con y sin cada herramienta opcional.

**Leer antes**

- [internal/harness/config/assets/claude/sdd-orchestrator.md](internal/harness/config/assets/claude/sdd-orchestrator.md)
- [internal/harness/config/assets/codex/sdd-orchestrator.md](internal/harness/config/assets/codex/sdd-orchestrator.md)
- [internal/catalog/harnesses.yaml](internal/catalog/harnesses.yaml)
- [internal/harness/skill](internal/harness/skill)
- [Evidencia y recomendaciones del informe ejecutivo](https://active-stack-executive-site.vercel.app/#prioridades)

<a id="change-06"></a>

### 06 · Evidencia ligada a la ejecución

**Change** `opsx-verifiable-evidence` · **Responsable** Gianfranco · **Cliente** Ambos · **Dependencias** [03](#change-03), [04](#change-04)

**Problema.** Un resultado bien formado puede corresponder a otra revisión, a un comando insuficiente o a una ejecución inventada. Un test verde tampoco prueba por sí solo el cumplimiento de un requisito.

**Qué buscar.** Revisá qué contiene hoy el reporte y qué señales corresponden a la verificación de instalación. Determiná si el nuevo contrato puede aprovechar esas señales o corresponde a un artefacto externo. Separá la estructura del recibo, su procedencia y la revisión de requisitos.

**Qué se implementaría.** Definí un contrato de retorno con estado, hallazgos, requisitos cubiertos y evidencia vinculada a revisión, ruta, comando y salida. Especificá quién comprueba cada campo, qué queda sin verificar y cómo se revisa la suficiencia. Incluí evidencia vieja, inventada, truncada o insuficiente y acordá acceso, retención y redacción de datos sensibles en salidas originales.

**Qué acordar en grupo.** Qué datos se consideran confiables, quién verifica su procedencia y qué evidencia mínima requiere cada aceptación. Coordiná el acceso a la salida original con 05.

**Aceptación futura.** Rechazar recibos inventados o viejos, comandos insuficientes, salidas truncadas sin original y requisitos incumplidos con tests verdes. Un recibo incompleto queda sin verificar.

**Leer antes**

- [internal/verify/report.go](internal/verify/report.go)
- [internal/verify/checks.go](internal/verify/checks.go)
- [internal/verify/scenario_test.go](internal/verify/scenario_test.go)
- [ARCHITECTURE.md](ARCHITECTURE.md)
- [Evidencia y recomendaciones del informe ejecutivo](https://active-stack-executive-site.vercel.app/#evidencia)

<a id="change-07"></a>

### 07 · Delegación selectiva con contrato

**Change** `opsx-selective-delegation` · **Responsable** Neyen · **Cliente** Ambos · **Dependencias** [05](#change-05), [06](#change-06)

**Problema.** Delegar puede despejar el contexto principal, pero también sumar consumo, reintentos y solapamiento entre agentes.

**Qué buscar.** Identificá qué delegación exigen hoy las plantillas y dónde se definen ownership y retorno. Diseñá casos de tarea breve, tarea verbosa y cambios que compiten por los mismos archivos. Usá una comparación controlada antes de convertir la delegación en política predeterminada.

**Qué se implementaría.** Definí cuándo resolver directo y cuándo cargar el módulo de delegación y coordinación. Especificá el encargo mínimo, límites de alcance, propiedad de archivos, condiciones de paralelismo y el retorno estructurado de 06. Diseñá un piloto que compare ambas modalidades sobre las mismas tareas, usando 02, 05 y 06.

**Qué acordar en grupo.** Qué señales justifican delegar, qué cambios se pueden paralelizar y qué métrica justifica adoptar el flujo.

**Aceptación futura.** Comparar ejecución directa y delegada con misma tarea y aceptación. Incluir principal, workers y reintentos. Adoptar si mejora la métrica objetivo con calidad equivalente o superior.

**Leer antes**

- [internal/harness/config/assets/codex/sdd-orchestrator.md](internal/harness/config/assets/codex/sdd-orchestrator.md)
- [internal/harness/config/assets/claude/sdd-orchestrator.md](internal/harness/config/assets/claude/sdd-orchestrator.md)
- [internal/catalog/harnesses.yaml](internal/catalog/harnesses.yaml)
- [Evidencia y recomendaciones del informe ejecutivo](https://active-stack-executive-site.vercel.app/#prioridades)

<a id="change-08"></a>

### 08 · Controles efectivos por cliente

**Change** `opsx-native-controls` · **Responsable** Gianfranco · **Cliente** Ambos · **Dependencias** [06](#change-06)

**Problema.** Que un hook esté instalado no prueba que se ejecute, que cubra todas las rutas ni que pueda bloquear una acción antes de que ocurra.

**Qué buscar.** Inventariá permisos, controles y capacidades nativas por cliente. Separá los hooks del cliente del hook de verificación interno del instalador. Rastreá qué shells y herramientas pueden evitar cada control y acordá el alcance verificable antes de diseñar la configuración.

**Qué se implementaría.** Armá una matriz por cliente, superficie, herramienta y shell. Para cada regla, indicá si orienta al agente, restringe una acción mediante permisos o rules del cliente, o verifica un evento mediante un hook. Separá prevención de verificación posterior. Diseñá casos permitido, denegado, timeout, JSON inválido, handler ausente y caminos alternativos, con revisión previa de cualquier cambio sobre permisos del usuario.

**Qué acordar en grupo.** Qué acciones deben bloquearse, cómo se responde si falla el control y qué superficies quedan cubiertas. Coordiná con 01 y 06.

**Aceptación futura.** Probar permitido, denegado, timeout, JSON inválido, handler ausente, rutas alternativas y shells reales. Documentar límites por superficie. Revisar cambios antes de escribir configuración de usuario.

**Leer antes**

- [internal/harness/config/permissions/install.go](internal/harness/config/permissions/install.go)
- [internal/harness/config/permissions/overlays.go](internal/harness/config/permissions/overlays.go)
- [internal/verify/hook.go](internal/verify/hook.go)
- [internal/verify/hook_test.go](internal/verify/hook_test.go)
- [Evidencia y recomendaciones del informe ejecutivo](https://active-stack-executive-site.vercel.app/#prioridades)

<a id="change-09"></a>

### 09 · Pilotos de modelos y caché

**Change** `opsx-model-cache-pilots` · **Responsable** Matías · **Cliente** Ambos · **Dependencias** [07](#change-07)

**Problema.** El beneficio de elegir modelo por rol y reutilizar prefijos depende de la tarea, del proveedor y de la caché observada. Todavía falta medirlo.

**Qué buscar.** Localizá las políticas de enrutamiento ya expresadas en las plantillas y documentá qué knobs están disponibles en cada cliente objetivo. Tratá la selección de modelos y la estabilidad de prefijos como variables independientes. Considerá caché como reutilización de cómputo y medí por separado el contexto lógico.

**Qué se implementaría.** Diseñá dos experimentos separados, uno para modelo y esfuerzo por rol y otro para prefijos y caché. Usá la referencia de 02 y el flujo de 07. Definí variantes, corridas frías y calientes, condiciones de aceptación y cómo contabilizar workers y reintentos.

**Qué acordar en grupo.** Qué variantes tienen sentido, qué métricas expone cada cliente y qué resultado justificaría adoptar cada cambio. Si se propone Jev para routing, comparalo como experimento opcional frente al routing nativo y separalo del piloto de revisión de 10.

**Aceptación futura.** Avanzar solo si el costo o consumo por tarea aceptada mejora con calidad comparable. Acordar umbrales y muestra (TBD). Reportar tiempo y cuota por separado.

**Leer antes**

- [internal/harness/config/assets/claude/sdd-orchestrator.md](internal/harness/config/assets/claude/sdd-orchestrator.md)
- [internal/harness/config/assets/codex/sdd-orchestrator.md](internal/harness/config/assets/codex/sdd-orchestrator.md)
- [ARCHITECTURE.md](ARCHITECTURE.md)
- [Evidencia y recomendaciones del informe ejecutivo](https://active-stack-executive-site.vercel.app/#plan)

<a id="change-10"></a>

### 10 · Juez externo en sombra

**Change** `opsx-shadow-review` · **Responsable** Neyen · **Cliente** Ambos · **Dependencias** [08](#change-08), [09](#change-09)

**Problema.** Hace falta saber si un juez externo detecta problemas adicionales frente a la revisión con criterios y cuánto cuesta incorporar esa señal.

**Qué buscar.** Tomá los requisitos y la evidencia acordados en 06 como referencia de evaluación. Identificá qué información podría salir del entorno y cómo minimizarla. Definí si el piloto agrega señal frente a reglas y revisión humana, y cómo terminarlo con una decisión documentada.

**Qué se implementaría.** Diseñá un piloto en sombra con una muestra etiquetada, una revisión de referencia y métricas de falsos positivos, falsos negativos y costo. Evaluá Jev como candidato opcional si soporta el caso de revisión. Documentá los datos que recibiría, cómo minimizar su exposición y el criterio para cerrar, repetir o descartar el piloto. Sus resultados se analizan aparte de la aceptación del cambio.

**Qué acordar en grupo.** Qué datos se pueden compartir, quién etiqueta la muestra y qué umbral justificaría seguir explorando.

**Aceptación futura.** Medir falsos positivos, falsos negativos y costo adicional frente a revisión de referencia. Mantener los resultados fuera de la decisión de aceptación durante el piloto.

**Leer antes**

- [internal/verify/report.go](internal/verify/report.go)
- [internal/verify/scenario_test.go](internal/verify/scenario_test.go)
- [ARCHITECTURE.md](ARCHITECTURE.md)
- [Evidencia y recomendaciones del informe ejecutivo](https://active-stack-executive-site.vercel.app/#plan)

## Seguimiento de iniciativas anteriores

El [roadmap post-starters archivado](docs/archive/CHANGES-2026-11-05-post-starters.md) conserva las fichas y decisiones de la etapa anterior. Su archivo es documental. El estado de los changes sigue en OpenSpec.

- **Verificación visual pendiente.** [`codex-design-mockups`](openspec/changes/archive/2026-11-06-codex-design-mockups/tasks.md) se archivó con 9/11 tareas. El roadmap anterior también registra validación con agentes reales pendiente para `universal-design-system`, aunque sus tareas locales figuran 6/6. El commit `1f493f1` incorpora la base visual persistente; la validación real se sigue por separado.
- **Verificación del instalador pendiente.** [`gui-installer-dependency-detection`](openspec/changes/archive/2026-11-06-gui-installer-dependency-detection/tasks.md) se archivó con 29/30 tareas y [`complete-page-finish-button-fix`](openspec/changes/archive/2026-11-06-complete-page-finish-button-fix/tasks.md) con 13/14. Son seguimientos de la etapa anterior.
- **Pendientes históricos a conciliar.** El hub TUI ya existe en el código y conviene reconciliar su ficha histórica. La investigación `claude-agent-switch-research` aporta contexto para 03 y `tui-configure-models` sigue fuera del alcance de esta ronda. `stitch-design` y `figma-mcp` quedan como ideas históricas sujetas a priorización. Consultá sus fichas archivadas antes de retomarlas.
