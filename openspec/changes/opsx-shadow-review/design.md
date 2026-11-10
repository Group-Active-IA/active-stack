# Design: opsx-shadow-review

## Context

El flujo OPSX revisa un change contra sus requisitos y escenarios. 06 agrega recibos con procedencia y suficiencia; 08 agrega controles nativos con cobertura documentada; 09 mide modelo y caché por rol. Esta ficha pregunta si un juez externo, corriendo en paralelo y sin influir, agrega señal que justifique su costo y su exposición de datos.

"Externo" significa fuera del flujo de aceptación: no es el auditor ni un rol del coordinador, no ve el razonamiento de los otros roles y no decide. Puede ser otro modelo, otro proveedor o un servicio.

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

## Decisions

### D1 — Qué es la verdad de referencia

Tres capas distintas, que no hay que confundir:

| Capa | Qué es | Para qué |
|---|---|---|
| Etiquetas | defectos conocidos por change, fijados antes de correr | verdad contra la que se mide todo |
| Revisión de referencia | el flujo actual: requisitos + escenarios + recibos 06 + controles 08 + persona | la línea a superar |
| Juez | candidato en sombra | lo que se evalúa |

Se miden **las dos** (referencia y juez) contra las etiquetas. La "señal adicional" es lo que el juez encuentra y la referencia no, confirmado por la etiqueta.

**Alternativa.** Usar la referencia como verdad: más barato, pero cada defecto que el juez encuentra y la referencia no se contaría como falso positivo, que es exactamente lo que queremos medir.

### D2 — Muestra

| Fuente | Cómo se arma | Aporta |
|---|---|---|
| Histórica | commits `fix` de `main` (33 en `1b4639b`) mapeados al change o commit que introdujo el defecto | defectos reales que el flujo dejó pasar |
| Sembrada | fixtures con defectos inyectados: los casos de rechazo de 06 (recibo viejo, inventado, comando insuficiente, salida truncada sin original, test verde con requisito incumplido) y defectos de lógica | casos con verdad conocida y cobertura de clases |
| Control | changes archivados sin fixes posteriores conocidos | tasa de falsos positivos |

Tamaño: decisión de grupo (recomendación: ≥30 casos, al menos 10 de control). La muestra se congela con un commit antes de la primera corrida.

**Etiquetado.** Dos personas etiquetan cada caso por separado con severidad (bloqueante / mayor / menor) y ubicación; un tercero resuelve desacuerdos. Se reporta el acuerdo entre etiquetadores. Quién etiqueta: decisión de grupo; el responsable del piloto no etiqueta solo.

### D3 — Ejecución en sombra

1. El caso entra al flujo de referencia y se registra su decisión.
2. En paralelo o después, el juez recibe el paquete de D5 y devuelve hallazgos en el formato de D4.
3. Los hallazgos se guardan en un almacén sellado (directorio fuera del árbol del change, con huella) que nadie del flujo consulta.
4. El emparejamiento contra etiquetas se hace al final, por alguien que no tomó la decisión de aceptación.

Para casos históricos la aceptación ya ocurrió, así que la sombra es trivial. Para casos nuevos, si el piloto corre sobre changes vivos, los resultados se abren recién cuando el change está archivado o rechazado.

### D4 — Formato y emparejamiento de hallazgos

```
hallazgo: {severidad, ruta, rango de líneas, requisito o escenario, descripción}
```

Un hallazgo empareja con una etiqueta si apunta al mismo archivo y a un rango solapado o al mismo requisito, y describe el mismo defecto. Los casos dudosos los decide el tercer etiquetador. Métricas por candidato:

- precisión = VP / (VP + FP); exhaustividad = VP / (VP + FN), por severidad;
- señal adicional = VP del juez que la referencia no detectó;
- tasa de FP en controles;
- costo por caso: tokens de entrada (con y sin caché), salida, dinero o cuota (09/02), latencia;
- tiempo humano de triage por hallazgo.

### D5 — Minimización de datos

| Clase | Ejemplos | Al juez |
|---|---|---|
| Pública | diff, proposal/design/specs/tasks del change, código del repo público | sí |
| Derivada redactada | recibos de 06 y salidas acotadas, redactadas según 06 (secretos, valores de variables, rutas del home), con huella antes y después | sí, si el caso la necesita |
| Privada | salidas originales sin redactar, `settings.json`/`config.toml` del usuario, memoria de Engram, `.env`, transcripciones de sesión | nunca |

El paquete se arma con una lista de inclusión (no de exclusión), se registra su huella y el log de envío guarda qué casos y qué huellas salieron, a qué proveedor y cuándo. Condiciones del proveedor (retención, uso para entrenamiento, región): decisión de grupo antes de la primera corrida; recomendación: sólo proveedores con retención cero o no uso para entrenamiento declarados por contrato o configuración.

**Alternativa.** Mandar el repo completo: más contexto, más señal posible, pero rompe la minimización y mezcla el efecto del contexto con el del juez.

### D6 — Candidatos

- **Juez base**: un modelo de proveedor distinto del que usó el flujo de referencia, con un prompt de revisión fijo versionado. Da un piso barato y reproducible.
- **Juez mismo proveedor, prompt distinto** (opcional): separa el efecto "otro modelo" del efecto "otra mirada".
- **Jev** (opcional): entra sólo si se documenta qué es y cumple los criterios de entrada: acepta el paquete de D5 sin datos adicionales, devuelve hallazgos ubicables (D4), expone costo observable y declara su política de datos. Si no los cumple, se registra el motivo y se descarta sin correr.

Modelo y esfuerzo de cada candidato se fijan con lo que mida 09.

### D7 — Criterio de cierre

| Salida | Condición (umbrales a acordar) | Recomendación de valores |
|---|---|---|
| Seguir explorando | señal adicional en bloqueantes/mayores ≥ U1, FP en controles ≤ U2, costo por caso ≤ U3 | U1 ≥ 2 casos o ≥ 10 % de los defectos mayores; U2 ≤ 20 %; U3 a fijar con 09 |
| Repetir | métricas dentro de la incertidumbre o acuerdo entre etiquetadores bajo | — |
| Descartar | sin señal adicional o FP sobre U2 | — |

"Seguir explorando" no es integrar: abre la decisión de arquitectura separada que exige el roadmap. Toda salida produce un informe con datos crudos, muestra, acuerdo de etiquetado, log de envío y decisión.

## Risks / Trade-offs

- **Fuga de sombra.** Si alguien ve los hallazgos antes de decidir, la referencia se contamina. Mitigación: almacén sellado, emparejamiento por un tercero, apertura sólo con la decisión registrada.
- **Muestra histórica sesgada.** Los fixes que conocemos son los que alguien notó. Mitigación: complementar con defectos sembrados y reportar por fuente.
- **Etiquetas inconsistentes.** Mitigación: doble etiquetado y acuerdo reportado; con acuerdo bajo la salida es "repetir".
- **Costo del triage.** Un juez ruidoso consume tiempo humano. Se mide explícitamente.
- **Dependencias sin rama.** 09 (y 02) no tienen propuesta al 10/11. Las tareas de costo y elección de modelo quedan bloqueadas por ellas.
- **Jev indefinido.** Si nadie lo documenta, el piloto sigue con el juez base.

## Migration Plan

No aplica: el piloto no cambia el instalador ni la configuración del usuario.

## UI Mockups

Decision: no (no UI) — 2026-11-10.

## Decisiones que el grupo tiene que cerrar

| Tema | Opciones | Recomendación | Afecta a |
|---|---|---|---|
| Qué datos se comparten | sólo público / público + derivado redactado / todo | público + derivado redactado con lista de inclusión | 05, 06 |
| Condiciones del proveedor | cualquiera / retención cero o sin entrenamiento | retención cero o sin entrenamiento declarado | — |
| Quién etiqueta | responsable / dos del grupo + tercero | dos del grupo + tercero; el responsable no etiqueta solo | — |
| Tamaño de muestra | — | ≥30 casos, ≥10 de control | 02 |
| Umbrales de cierre | — | ver D7 | 09 |
| Jev | incluir / excluir / condicionado | condicionado a criterios de entrada | 09 |

## Open Questions

- ¿Qué es Jev y dónde está su documentación? Necesario para evaluar los criterios de entrada.
- ¿El piloto corre sólo sobre casos históricos y sembrados, o también sobre changes vivos de la próxima ronda?
