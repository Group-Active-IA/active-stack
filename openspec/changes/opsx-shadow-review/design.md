# Design: opsx-shadow-review

## Context

El flujo OPSX revisa un change contra sus requisitos y escenarios. 06 agrega recibos con integridad, procedencia y suficiencia; 08 agrega controles nativos con cobertura documentada; 02 (`240e1b2`) define el protocolo de medición y 09 (`89f1fa5`) mide modelo y caché. Esta ficha pregunta si un juez externo, corriendo aparte y sin influir, agrega señal que justifique su costo y su exposición de datos.

"Externo" significa fuera del flujo de aceptación: no es el auditor ni un rol del coordinador, no ve el razonamiento de los otros roles y no decide. Puede ser otro modelo, otro proveedor o un servicio.

**Decisión confirmada (10/11).** La primera ronda usa sólo casos históricos y casos de prueba preparados, con material público revisado y datos ficticios. Quedan fuera los changes vivos y la evidencia real del usuario. Proveedor, paquete de datos, acceso y métricas se acuerdan antes de correr.

## Goals / Non-Goals

**Goals**

- Medir falsos positivos, falsos negativos, señal adicional y costo del juez frente a una referencia.
- Mantener el piloto en sombra: cero influencia sobre la aceptación.
- Exponer el mínimo de datos y dejar constancia de qué salió.
- Terminar con una decisión documentada.

**Non-Goals**

- Integrar un juez al flujo.
- Evaluar Jev para routing (09).
- Mejorar la revisión de referencia: se mide como está.
- Changes vivos o evidencia real en la primera ronda.

## Decisions

### D1 — Qué es la verdad de referencia

| Capa | Qué es | Para qué |
|---|---|---|
| Etiquetas congeladas | defectos conocidos por caso, fijados antes de correr | verdad del análisis principal |
| Revisión de referencia | el flujo actual: requisitos + escenarios + recibos 06 + tabla de cobertura de 08 (qué acciones del caso quedan controladas y cuáles no) + persona | la línea a superar |
| Juez | candidato en sombra | lo que se evalúa |

Referencia y juez se miden **las dos** contra las etiquetas congeladas. La señal adicional es lo que el juez encuentra y la referencia no, confirmado por la etiqueta.

**Alternativa.** Usar la referencia como verdad: cada defecto que el juez encuentre y la referencia no vea contaría como falso positivo, que es justo lo que queremos medir.

### D2 — Muestra

| Fuente | Cómo se arma | Aporta |
|---|---|---|
| Histórica | commits `fix` de `main` (33 en `1b4639b`) mapeados al commit que introdujo el defecto; el caso es el estado **previo** al fix | defectos reales que el flujo dejó pasar |
| Preparada | fixtures de la familia "Revisión de un diff sembrado" de 02 (`opsx-baseline/design.md:30`), con su rúbrica, más casos de rechazo de 06 (recibo falso con hash correcto, mismo HEAD con diff distinto, salida truncada sin original, tests verdes sin cubrir el requisito) y defectos de lógica con ubicación conocida | verdad conocida, cobertura de clases y una sola definición de falso positivo compartida con 02 |
| Control | **controles inspeccionados**: changes revisados manualmente por los etiquetadores y declarados sin defecto conocido, y **fixtures limpios** construidos para ese fin | tasa de falsas alarmas |

Que un change no tenga fixes posteriores conocidos no prueba que esté limpio, así que no alcanza para ser control: los controles históricos pasan primero por inspección. Datos de recibos y salidas en la primera ronda: ficticios.

**Paquetes sin pistas.** Del paquete histórico se excluye el fix posterior y todo lo que revele la respuesta: mensajes de commit del fix, referencias a issues o PRs del fix, texto de specs o tareas escrito después, nombres de rama y etiquetas. El paquete se arma desde el estado previo, no desde el diff del fix.

**Tamaño.** Decisión de grupo (recomendación: ≥30 casos, al menos 10 controles). La muestra se congela con un commit antes de la primera corrida.

**Etiquetado.** Dos personas etiquetan cada caso por separado (severidad bloqueante / mayor / menor, ubicación y requisito afectado); un tercero resuelve desacuerdos. Se reporta el acuerdo. El responsable del piloto no etiqueta solo.

### D3 — Ejecución en sombra y acceso

1. El caso tiene su decisión de referencia registrada (en históricos y preparados ya existe o se fija antes de correr el juez).
2. El juez recibe el paquete de D5 y devuelve hallazgos en el formato de D4.
3. Los hallazgos se guardan fuera del árbol del change, con huella. **La huella prueba integridad, no confidencialidad**: el almacén tiene acceso restringido a un custodio designado por el grupo, que no participa de la referencia ni del etiquetado.
4. **Momento de apertura (único para diseño y spec):** el custodio abre los resultados de un caso sólo cuando ese caso está cerrado definitivamente, es decir, con decisión de referencia y etiquetas congeladas registradas. Antes de eso, nadie del flujo de referencia ni del etiquetado accede.
5. El emparejamiento lo hace el tercer etiquetador.

Si en una ronda posterior se agregan changes vivos, primero hay que probar que el equipo que decide la aceptación no puede ver hallazgos antes del cierre (escenario de acceso anticipado).

### D4 — Hallazgos, emparejamiento y arbitraje

```
hallazgo: {severidad, ruta, rango de líneas, requisito o escenario, descripción}
```

**Análisis principal (etiquetas congeladas).** La regla de emparejamiento es la de la familia de revisión de 02 (dueña de la rúbrica de revisión); esta ficha no define otra. Mientras 02 no la fije, se propone: un hallazgo empareja si apunta al mismo archivo y rango solapado, o al mismo requisito, y describe el mismo defecto. VP, FP y FN se cuentan sólo contra etiquetas congeladas.

**Arbitraje aparte.** Los hallazgos que no emparejan no se reclasifican dentro del análisis principal. Van a un arbitraje separado, ciego respecto del origen (el árbitro no sabe si el hallazgo vino del juez o de la referencia). Cada uno termina en una de tres categorías:

- **defecto nuevo confirmado** (válido y ausente de las etiquetas);
- **falsa alarma**;
- **indeterminado**.

Se reportan aparte. Un defecto nuevo confirmado no cambia retroactivamente las métricas principales.

**Unidades.**

| Medida | Unidad |
|---|---|
| Señal adicional | cantidad de defectos distintos de severidad bloqueante o mayor, presentes en etiquetas congeladas, que el juez detectó y la referencia no. Es la unidad de decisión. El porcentaje sobre el total de defectos mayores se informa como contexto, no decide. |
| Falsas alarmas | falsas alarmas por caso; y en controles, proporción de controles con al menos una falsa alarma (unidad de decisión) |
| Precisión / exhaustividad | VP/(VP+FP) y VP/(VP+FN), por severidad |
| Costo | tokens (entrada nueva, caché, salida), costo API facturado o estimado o cuota por separado, latencia, por caso |
| Triage | minutos humanos por caso para revisar hallazgos del juez |

### D5 — Minimización de datos

| Clase | Ejemplos | Primera ronda |
|---|---|---|
| Pública revisada | diff del estado previo, proposal/design/specs/tasks del change, código del repo público | sí, tras el examen previo |
| Ficticia | recibos, salidas y configuraciones construidos para los fixtures | sí |
| Derivada real redactada | recibos y salidas reales redactados según 06 | **no** (evaluar en otra ronda) |
| Privada | salidas originales, `settings.json`/`config.toml` del usuario, memoria de Engram, `.env`, transcripciones | nunca |

**Examen previo de todo el paquete.** Aunque el repo sea público, se examina el paquete completo antes de enviarlo, diffs públicos incluidos: un commit público puede contener un secreto filtrado. El examen usa la **lista única de patrones de valor sensibles de 05** (`opsx-task-context/design.md`, D14: `.env` y sus valores, tokens y API keys, credenciales de nube, claves SSH y `*.pem`, valores de variables de entorno, rutas del home, identificadores de cuenta), sin una lista propia. Además rechaza archivos de configuración de usuario (`settings.json`, `config.toml`), que son una clase de archivo y no un patrón de valor. Un positivo cancela el envío del caso.

El paquete se arma por lista de inclusión, se registra su huella y el log de envío guarda caso, huella, candidato, proveedor y fecha.

**Condiciones del proveedor**, verificadas como condiciones **separadas**: (a) retención de datos, (b) uso para entrenamiento. Se acuerdan antes del primer envío. Recomendación: retención cero o mínima declarada **y** no uso para entrenamiento declarado.

### D6 — Candidatos

- **Juez base**: modelo de proveedor distinto del usado por la referencia, con prompt de revisión fijo y versionado.
- **Juez mismo proveedor, prompt distinto** (opcional): separa "otro modelo" de "otra mirada".
- **Jev** (opcional): 09 también lo evalúa (para routing, `opsx-model-cache-pilots/design.md:33`) y ninguna ficha tiene su documentación. Se propone una sola tarea compartida con 09 que obtenga qué es, qué datos recibe y cómo se mide; las dos fichas la usan. Entra sólo si se documenta y cumple los criterios de entrada: acepta el paquete de D5 sin datos adicionales, devuelve hallazgos ubicables (D4), expone costo observable y declara retención y uso para entrenamiento. Si no, se registra el motivo y se descarta sin correr.

**Calibración.** Antes de la muestra, cada candidato corre sobre un conjunto propio de calibración (fuera de la muestra) para fijar prompt, modelo y esfuerzo según el protocolo de 02 y lo que mida 09. Nada de la calibración se reutiliza en la muestra.

### D7 — Criterio de cierre

Valores a fijar antes de correr (recomendación entre paréntesis):

- U1 señal adicional mínima: cantidad de defectos bloqueantes o mayores (≥2).
- U2 falsas alarmas máximas: proporción de controles con al menos una falsa alarma (≤20 %).
- U3 presupuesto por caso: dinero o cuota y minutos de triage (a fijar con 02/09).
- Acuerdo mínimo entre etiquetadores (a fijar).

| Salida | Condición |
|---|---|
| Seguir explorando | señal adicional ≥ U1, falsas alarmas ≤ U2, costo ≤ U3 |
| Repetir | acuerdo de etiquetado bajo el mínimo, o resultado dentro de la incertidumbre |
| Descartar | señal adicional < U1, o falsas alarmas > U2 |

Si el costo supera U3 con señal suficiente, la salida es repetir con otro candidato o configuración, no seguir. "Seguir explorando" no es integrar: abre la decisión de arquitectura separada que exige el roadmap. Toda salida produce un informe con datos crudos, muestra, acuerdo, arbitraje aparte, log de envío y decisión. La aceptación de los changes sigue su flujo habitual.

## Risks / Trade-offs

- **Fuga de sombra.** Mitigación: custodio con acceso restringido, apertura sólo al cierre definitivo del caso, emparejamiento por un tercero.
- **Pistas en el paquete histórico.** Mitigación: paquete desde el estado previo, exclusión del fix y metadatos; revisión del paquete por el custodio.
- **Muestra histórica sesgada.** Los fixes que conocemos son los que alguien notó. Mitigación: casos preparados y reporte por fuente.
- **Controles mal clasificados.** Mitigación: inspección previa y fixtures limpios.
- **Etiquetas incompletas.** El juez puede encontrar defectos reales no etiquetados. Mitigación: arbitraje aparte con categoría de defecto nuevo confirmado.
- **Costo del triage.** Un juez ruidoso consume tiempo humano; se mide.
- **Jev indefinido.** Si nadie lo documenta, el piloto sigue con el juez base.

## Migration Plan

No aplica: el piloto no cambia el instalador ni la configuración del usuario.

## UI Mockups

Decision: no (no UI) — 2026-11-10.

## Decisiones

### Confirmadas por el grupo (10/11)

| Tema | Decisión |
|---|---|
| Alcance de la primera ronda | Casos históricos y preparados; material público revisado y datos ficticios; sin changes vivos ni evidencia real del usuario. |
| Antes de correr | Acordar proveedor, paquete de datos, acceso y métricas. |

### Pendientes

| Tema | Opciones | Recomendación | Afecta a |
|---|---|---|---|
| Condiciones del proveedor | — | retención cero o mínima y no entrenamiento, verificadas por separado | — |
| Custodio y etiquetadores | — | dos etiquetadores + tercero; custodio distinto de todos ellos | — |
| Tamaño de muestra | — | ≥30 casos, ≥10 controles | 02 |
| U1, U2, U3 y acuerdo mínimo | — | ver D7 | 02, 09 |
| Jev | incluir / excluir / condicionado | condicionado a criterios de entrada | 09 |

## Cambios por la auditoría del 10/11

| Observación | Dónde se incorporó |
|---|---|
| "Sin fixes posteriores" no acredita limpio | D2 controles inspeccionados y fixtures limpios; tarea 1.3 |
| Etiquetas congeladas y arbitraje aparte ciego | D4; requisito "Arbitraje aparte" |
| Defecto nuevo, falsa alarma, indeterminado | D4 |
| Excluir el fix y pistas del paquete | D2 "Paquetes sin pistas"; requisito y tarea 1.5 |
| Huella ≠ confidencialidad; quién abre | D3 custodio y acceso restringido |
| Unificar momento de apertura | D3 paso 4; requisito "Resultados en sombra" |
| Primera ronda sin vivos | Context, D3, D5 |
| Examinar todo el paquete, diffs públicos incluidos | D5; escenario de secreto en diff público |
| Retención y entrenamiento separados | D5; requisito de proveedor |
| 02 y 09 tienen rama; calibrar candidatos | Context, D6 |
| Unidades de señal adicional y falsas alarmas; presupuesto por caso | D4 unidades, D7 |

## Cambios por la auditoría complementaria del 10/11

| Observación | Dónde se incorporó |
|---|---|
| Clases propias en el examen de datos | D5: lista única de 05 D14; requisito "Examen previo"; tarea 3.2 |
| Dos definiciones de falso positivo (02 y 10) | D2 casos preparados desde la familia de revisión de 02; D4 regla de emparejamiento de 02; tarea 1.2 |
| Cobertura de 08 implícita en la referencia | D1 tabla; tarea 2.1 |
| Jev sin dueño en 09 y 10 | D6; tarea 0.2 compartida con 09 |
