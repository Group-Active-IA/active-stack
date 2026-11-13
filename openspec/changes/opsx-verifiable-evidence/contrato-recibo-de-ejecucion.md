# Contrato del recibo de ejecución (tareas 2.1, 2.2, 2.3 y 2.5)

> Redacción del contrato del change `opsx-verifiable-evidence` (06). Todo lo que depende de otros
> changes se cita y no se redefine. Las decisiones que dependen de Gianfranco son **propuesta de
> 06, pendiente de confirmación**. Los ejemplos válidos e inválidos son los fixtures de
> `eval/evidence/testdata/casos/` (cada ejemplo nombra su archivo) y los verifica
> `go test ./eval/evidence`.

## 1. Qué es (y qué no es)
Un **recibo de ejecución** describe **una ejecución de un comando**. No es el *registro de
instalación* de 04, que describe una instalación. Un resultado que dice "corrí los tests" sin
recibo queda *sin verificar*.

## 2. Esquema (tarea 2.1)

| Campo | Contenido | Quién lo escribe |
|---|---|---|
| `corrida` | identificador de la corrida | la fuente de captura (05) |
| `arbol` | `directorio`, `commit` y, si había cambios sin confirmar, `huella` del estado completo (preparados, sin preparar y archivos nuevos). `HEAD` más "dirty" no alcanza | el que ejecuta; la fuente lo corrobora |
| `comando` | comando completo, tal como se ejecutó | el que ejecuta; la fuente lo corrobora |
| `resultado` | `codigo_de_salida` y `marca_de_tiempo` | declarado; el código lo escribe además quien captura (05 D9) |
| `requisitos` | ids de requisitos o escenarios que dice cubrir | el que ejecuta |
| `salida` | `huella` (una sola, `{algorithm, value}`) de la **copia redactada**, `truncada` y `captura` (registro de captura de 05 D9) | la huella y el registro, quien captura |

- **Huella de la copia:** sha256 de los bytes exactos, sin conversión de fines de línea, con el
  algoritmo nombrado. 05 la escribe `sha256:<hex>`; `{algorithm, value}` (02) es equivalente.
  Medido en `informes/medicion-huella-2026-11-13.md`.
- **Huella del estado probado:** TreeHash sobre los bytes de la copia de trabajo (propuesta de 06,
  pendiente de confirmación; `informes/medicion-huella-2026-11-13.md`).
- **Registro de captura:** es el de 05 D9. 06 lo referencia y no agrega campos. El fixture lo
  llena con valores ilustrativos.
- Nombres de campo en español, como en `design.md` D2. El módulo en inglés los llama `run`, `tree`,
  `command`, `result`, `requirements` y `output` (ver `entrega-evidence-contract.md`).

### Ejemplo válido (`testdata/casos/valido.json`)
```json
{
  "corrida": "r-001",
  "arbol": { "directorio": "/work/active-stack-06", "commit": "7a600895e8f1fe1f3728cffa1fcb57dca5663a59" },
  "comando": "go test ./...",
  "resultado": { "codigo_de_salida": 0, "marca_de_tiempo": "2026-11-13T12:00:00Z" },
  "requisitos": ["R1"],
  "salida": {
    "huella": { "algorithm": "sha256", "value": "<sha256 de los bytes de copias/go-test-ok.txt>" },
    "truncada": false,
    "captura": { "fuente": "arnés", "codigo_de_salida": 0, "cobertura": "demostrada", "...": "campos de 05 D9" }
  }
}
```
Veredicto: *verificado* (integridad aprobada, procedencia corroborada, suficiencia suficiente).

### Ejemplos inválidos (cada uno es un fixture)

| Fixture | Qué tiene de malo | Estado | Control que falla |
|---|---|---|---|
| `incompleto` | falta el registro de captura | sin verificar | ninguno evaluable |
| `inventado-copia-ausente` | la copia referenciada no existe y no venció | rechazado | integridad |
| `inventado-huella-no-coincide` | la copia existe pero sus bytes no dan la huella | rechazado | integridad |
| `copia-crlf-huella-lf` | huella de la copia con LF, archivo guardado con CRLF | rechazado | integridad |
| `inventado-fuente-contradice` | la fuente independiente muestra otro resultado | rechazado | procedencia |
| `inventado-codigo-de-salida` | el worker declara 0 y quien captura registró 1 | rechazado | procedencia |
| `viejo` | el commit del recibo no es el del árbol actual | rechazado | procedencia |
| `mismo-head-otro-diff` | mismo `HEAD`, huella del estado distinta | rechazado | procedencia |
| `directorio-ajeno` | producido fuera del worktree del change | rechazado (propuesta de partida, (TBD) con 05) | procedencia |
| `directorio-ajeno-separadores-windows` | otro directorio, escrito con separadores de Windows | rechazado | procedencia |
| `directorio-separadores-windows` | el mismo directorio con separadores de Windows y barra final | verificado (se normalizan separadores y barra final; mayúsculas, enlaces y 8.3: (TBD)) | — |
| `reejecucion-del-coordinador-sin-captura` | reejecución sin captura independiente | sin verificar | procedencia |
| `reejecucion-con-captura-independiente` | reejecución con captura propia independiente | verificado | — |
| `comando-insuficiente` | el comando no ejercita el requisito | rechazado | suficiencia |
| `tests-verdes-requisito-incumplido` | tests verdes y requisito incumplido | rechazado | suficiencia |
| `truncado-sin-original` | salida truncada que nunca tuvo referencia | rechazado | integridad |
| `cero-recibos` | el requisito no tiene recibos | sin verificar | no aplica |
| `falso-huella-correcta-sin-corroboracion` | recibo fabricado con huella correcta | sin verificar | procedencia |
| `falso-huella-correcta-reejecucion-contradice` | el mismo, contradicho por reejecución | rechazado | procedencia |
| `copia-vencida` | la referencia existía y la copia venció | sin verificar ((TBD), propuesta de partida) | integridad no verificable |
| `detalle-redactado` | verificar exige un detalle que la redacción descartó | sin verificar | suficiencia |

## 3. Los tres controles (tarea 2.2)
Se informan **por separado**; aprobar uno no implica los otros.

1. **Integridad.** La copia redactada existe y su huella recomputada coincide con la del recibo.
   Es mecánico y **no prueba que la ejecución ocurrió**: quien fabrica un recibo fabrica la huella
   (`falso-huella-correcta-sin-corroboracion`).
2. **Procedencia.** La ejecución es real y corresponde al estado del código: una **fuente
   independiente** o una **reejecución** corrobora el comando, el código de salida y el estado
   exacto del árbol. El código de salida lo escribe quien captura (05 D9) y se compara con el
   declarado. Estado actual de la fuente por interfaz (05 D22, citado, no decidido por 06): solo
   el arnés es referencia por construcción; sin fuente demostrada, un resultado declarado queda
   *sin verificar* salvo reejecución por otra parte con captura (05 D20). **Una reejecución solo cuenta como fuente independiente si tiene captura propia de una fuente independiente; la reejecución del coordinador sin esa captura no corrobora** (propuesta de 06, pendiente de confirmación; fixtures `reejecucion-con-captura-independiente` y `reejecucion-del-coordinador-sin-captura`).
3. **Suficiencia.** El comando ejercita lo que el requisito pide y el resultado lo respalda.
   Cada requisito nombra qué evidencia lo satisface. Quién la evalúa: **abierta**; propuesta de
   partida de 06, pendiente de confirmación: el revisor distinto del constructor de 04 (D8).
   Si verificar exige un detalle que la redacción descartó, queda *sin verificar* (requisito de 05
   "Un detalle redactado no alcanza para verificar", referenciado y no redefinido).

## 4. Los tres estados y sus transiciones (tarea 2.2)
Texto de la ficha de 06 para *rechazado* (`CHANGES.md:194`, "Aceptación futura"), literal:

> Rechazar recibos inventados o viejos, comandos insuficientes, salidas truncadas sin original y
> requisitos incumplidos con tests verdes. Un recibo incompleto queda sin verificar.

| Estado | Cuándo |
|---|---|
| **verificado** | integridad, procedencia (corroborada) y suficiencia aprobadas para el requisito, y existe al menos un recibo. Una ejecución sin recibos **nunca** queda verificada |
| **rechazado** | los cinco casos de la ficha: (1) inventado, (2) viejo, (3) comando insuficiente, (4) salida truncada sin original, (5) requisito incumplido con tests verdes |
| **sin verificar** | por defecto: recibo incompleto, sin corroboración, sin copia por vencimiento, no evaluable, o sin recibos. No cuenta como aprobado |

Orden de precedencia (implementado en `eval/evidence/judge.go`): salida truncada sin original →
rechazado; recibo incompleto → sin verificar; después se reúnen los motivos de rechazo y de duda, y
**gana el rechazo** sobre la duda. Con varios recibos para un requisito manda el peor estado; esa
regla de agregación es **(TBD)** y no tiene fixture.

Casos que la ficha no menciona (**(TBD) con 05**): *directorio ajeno* (propuesta de partida:
rechazado) y *copia vencida* (propuesta de partida: sin verificar).

## 5. Regla de la salida (tarea 2.3)
- Se persiste **solo la copia redactada**, con **una** huella y el registro de captura. No existe
  una copia con secretos de la que el recibo dependa.
- La redacción, el formato de la referencia, el registro de captura y la lista sensible son de 05;
  06 no define una segunda política.
- Ubicación, acceso y vencimiento: propuestos por 05 (D9); 06 propone aceptar, **pendiente de
  confirmación**. La referencia informa el vencimiento. Si la copia venció, el recibo no queda
  verificado (`copia-vencida`).
- Una salida truncada debe tener referencia a la copia completa; sin ella se rechaza
  (`truncado-sin-original`).
- Tras una corrida, ninguna salida completa aparece en `git status`: pendiente de la captura de 05
  (tarea 4.9).

## 6. Hooks (tarea 2.5)
Los hooks del cliente (08) pueden dejar rastro de ciertas ejecuciones y sirven para corroborar, pero
**complementan la evidencia y no son un registro exhaustivo**: no cubren todas las interfaces ni
todos los comandos, y su silencio no prueba que algo no se ejecutó. La ausencia de un rastro no
verifica ni rechaza un recibo. Un control posterior a la acción (08) puede validar de forma
mecánica los campos del recibo y la integridad de la copia; la procedencia y la suficiencia siguen
siendo de este contrato.

## 7. Lo que este contrato no define
El formato de la referencia, el registro de captura y la redacción (05); el sobre de retorno (07,
`recibos` es el campo cuyo contenido define 06); la acción `verify` y el revisor distinto del
constructor (04); las capas y los módulos (03); los controles nativos (08).
