# apply-t3-B · constructor · opsx-baseline (2.6, 2.7, 3.1 + cancelación y `pendiente_de_revisor`)

Rol: constructor, LOW. Rama `propose/opsx-baseline`, base `5dcfc68` + tandas 2 y 3A sin commitear. Fecha 2026-11-11, 15:10 a 15:30 (-03). Sin commits, sin pushes, sin `go build`, sin procesos en segundo plano al terminar. No toqué `openspec/changes/codex-strict-compat/`, `archive/`, `design.md` ni `proposal.md`.

## Fuentes
- Lecciones L-NNN: busqué por `L-0` en `*.md` del repo (y directorios `knowledge`/`kb`): no hay base de lecciones numeradas. Ausente. Apliqué las lecciones del auditor citadas en el brief (negativo de benignos al redactar; propiedad de permutación en agregados; hashes pegados).
- `openspec/changes/opsx-baseline/tasks.md`, `design.md` (Señales y fuentes, Denominadores, Reconciliación, API y cuota, Datos sensibles, Estructura de resultados, D5), `informes/apply-t3-A-constructor.md`.
- Código reutilizado: `run`, `trace`, `timing`, `manifest`.
- Contaminación: ninguna de informes ajenos más allá de `apply-t3-A-constructor.md`, que el brief nombró.

## EVIDENCIA

### Hashes (salida pegada de `git hash-object`, revisión entregada, 15:29)
```
f19982a55735cce45160c14e8151801665166231 eval/baseline/run/run.go
15a02189179fcedec9d0e77155a109be5a3fdeb5 eval/baseline/run/run_test.go
7472706f8628d432479929b8804cf43ccb65b4fe eval/baseline/usage/usage.go
2673410de17d6d8c7ba778d9a6228c0a3e14e8d0 eval/baseline/usage/usage_test.go
db84e1c5f025171ca68b64637dcf1d521e37dcd5 eval/baseline/record/record.go
c3ae3172e3271029f9157f571d923a3536792329 eval/baseline/record/record_test.go
46400948cfef06334b362d0e3bd1abe8941f95be eval/baseline/summary/summary.go
77073de19261f82297b916d9053b0681be803b7e eval/baseline/summary/summary_test.go
1be1ecdd9253f3daf8cffd6c6e29f5b6f5bde56a eval/baseline/README.md
2a516a592857fb30616d120f0d2fc41d73b37236 openspec/changes/opsx-baseline/tasks.md
```

### Chequeos sobre la revisión entregada
- Suite completa, una vez, `go test -count=1 ./...` 15:24:21–15:26:17: ninguna línea que no sea `ok` o `no test files`. Una segunda pasada (15:26–15:28, lanzada por error, terminó sola; sin procesos al cerrar) dio 36 paquetes `ok` (`fixture` 111.8 s).
- `gofmt -l eval` vacío; `go vet ./eval/...` ok; `go list -deps ./cmd/... ./internal/...`: 0 coincidencias con `/eval`; `git diff -- go.mod go.sum` vacío; `openspec validate opsx-baseline --strict`: `Change 'opsx-baseline' is valid` (tras corregir un error mío de ubicación de 3.1 en tasks.md que el validador marcó como duplicado).
- Safety net: `run` ok antes de tocarlo; paquetes nuevos sin archivos previos. No corrí `fixture` en los ciclos.

### Tabla comportamiento → RED → GREEN → mutante
| Comportamiento | RED (aserción) | GREEN | Mutante que falla |
|---|---|---|---|
| Cancelación: conserva el intento en curso con estado y causa | `Run = {... Attempts:[]}, want 1 intento conservado con estado "interrumpida_por_llamador"` (15:11:10) | `run` ok 15:11:17 | `return nil, parent.Err()` |
| Entrada lógica de Claude = entrada + lectura + creación | `FromClaude = {Input:0...}, want {Input:5400 Uncached:100 ...}` | usage ok 15:12:59, 100 % | entrada sin caché |
| Codex: cacheada dentro de la entrada, sin caché por diferencia | `FromCodex = {...}, want {Input:1000 Uncached:100 CacheRead:900 ...}` | ok | sumar la cacheada |
| Entrada de solicitud ≠ total del turno; métrica ausente = no disponible | `entrada de la solicitud = {Available:false}, want 5400 base solicitud`; `sin señales ... want ambas no disponibles con motivo` | ok | total del turno como entrada sin comprobar |
| Sesión nueva con lecturas cacheadas conserva el estado observado | `caché = {Expected: Observed: Diverges:false}, want observada con_lectura, prevista fría` | ok | observada reescrita con la prevista; sin divergencia |
| Reconciliación: no suma padre e hijos (60.000 / 48.000), sin atribuir, parcial | `Claude = {Total:0 ...}, want total 60000, principal 40000, worker 20000` | ok | sumar agregado + detalle; ignorar alcance solo-padre; perder sin atribuir; confiar en alcance no comprobado |
| Redacción de cada campo sensible (18 filas) | `<fila>: la salida conserva "SENTINEL"` / `categorías = [], want incluir "email"` | record ok 15:16:38, 98.2 % | un mutante por categoría y por clave (ver tasks.md, 29 distintos) |
| Miembros benignos intactos | comparación estructural (nace verde) | ok | sobre-redactar cadenas o arreglos |
| JSON inválido o con dos documentos no filtra | traza sintética filtraba `usr-SECRET-77`, `rm -rf SECRET-TOOL`, `SECRET PROMPT BODY` (15:16:06) | ok | conservar el crudo; aceptar segundo documento |
| Huella sha256 con algoritmo nombrado; evidencia conservada o ausencia | `sha256` vacío; `NewEvidence` vacío | ok | otro algoritmo; ruta absoluta conservada |
| Cobertura por intento y corrida (la más débil), parcial, null | `consumo ausente = ...` | ok | invertir la más débil; parcial ignorado |
| Costo total suma rechazos, bloqueos, timeouts y reintentos | `Compute = {Cohorts:[]...}` (15:20:03) | summary ok 15:20:42 | solo la aceptada cuenta; sin reintentos |
| Incompatibles fuera del denominador y listadas | `cohortes = []`, `incompatibles = []` | ok | iterar el manifiesto en vez del plan ejecutable (la guarda explícita es equivalente) |
| API, suscripción, interfaces separadas; interfaz real | `totales mezclados`/vacío | ok | modalidad perdida; interfaz declarada |
| Cero aceptadas: total publicado, cociente sin valor | `cociente = {...}, want sin valor y motivo` | ok | `accepted == 0` desactivado |
| Cuota no disponible: no calculable con motivo (D5) | `cuota = {...}, want no calculable con motivo` | ok | motivo perdido; no disponible como disponible |
| Cobertura incompleta/parcial: cociente sin valor, cobertura publicada | `cobertura = {...}` | ok | cociente con cobertura incompleta; parcial ignorado |
| `pendiente_de_revisor` aparte, provisional | `conteos = ...` | ok | contada como aceptada; sin provisional |
| Permutar corridas, cohortes, plan e intentos no cambia el resumen | nace verde (propiedad, 100 casos) | ok | sumar sin ordenar valores; sin ordenar cohortes |

Pruebas que nacieron verdes y se acreditaron por mutante: la propiedad de `Reconcile` (stub y resultado vacío coinciden; mutante de orden `=` en vez de `+=` la hace fallar), la de permutación de `summary`, la de benignos de `record` (mutante de sobre-redacción) y `TestCancelBetweenAttemptsKeepsPriorAttempts`.

## JUICIO
- 2.6, 2.7 y 3.1 están completos contra eventos sintéticos y sin clientes reales; quedan las muestras reales para 2.2 y 2.3.
- Endurecí la redacción durante GREEN: un fallback que redactaba solo texto de un JSON inválido dejaba pasar claves sensibles, así que un evento no parseable se reduce a su tipo. Es más conservador que la letra del brief y es lo correcto ante datos reales truncados.
- Interpretación de "UUID de cuenta": en texto libre cualquier UUID se redacta (no se puede distinguir una cuenta de otro UUID); los UUID de sesión y agente viajan por campos estructurados y se conservan.
- `Summary.Compute` valida que las corridas pertenezcan al plan ejecutable; una corrida en una celda incompatible es error y no se descarta en silencio.

## SUPOSICIÓN
1. Codex: el razonamiento está incluido en la salida (`Total = Input + Output`); confirmar con la versión fijada.
2. Alcance `ScopeUnknown` solo se confía con coincidencia exacta; con diferencia, parcial con ambas cifras (la regla 4 del diseño no dice nada del alcance no comprobado).
3. Sin agregado, el total es la suma del detalle y se considera cobertura completa (verdadero solo si el detalle cubre todos los hilos; el adaptador de 2.2 debe garantizarlo).
4. Una atribución parcial (detalle menor que el total) deja el total confiable y no cuenta como cifra parcial en el resumen; solo detalle mayor o alcance no comprobado lo hace.
5. Las claves sensibles son un conjunto mínimo mío (`user_email`, `account_uuid`, `user_id`, `tool_input`, `arguments`, `request_body`, `messages`, `prompt`, entre otras, e `id` bajo `user`); la lista dueña es de 05 y la reemplaza.
6. Las listas de rutas POSIX (`/home`, `/Users`, `/root`, `/tmp`, `/var`, `/etc`, `/opt`, `/mnt`, `/usr`, `/private`, `/Volumes`) son heurísticas: una ruta absoluta en otro lugar en texto libre no se redacta (sí en `NewEvidence`).
7. Los pesos de las tareas del manifiesto no entran en el cociente de 3.1 (el diseño define el cociente sin ponderar).
8. Los planes de los tests se arman a mano con `manifest.Plan` y `Cohort`, no vía `Plan()`; la integración con `Plan()` real queda sin prueba específica.

## Limitaciones y chequeos no ejecutados
- Sin clientes reales, Linux ni macOS; todos los eventos son sintéticos.
- No hice 1.2, 2.2, 2.3, 3.2–3.4 ni 4.1. Del resumen faltan dispersión por tarea, comparación emparejada y agregados entre cohortes (no pedidos en 3.1).
- Cobertura: `usage` 100 %, `record` 98.2 %, `summary` 96.1 %.
- Un error de procedimiento mío: lancé la suite completa dos veces seguidas; la segunda terminó sola antes de cerrar.

## DEUDA (nueva, no bloquea)
1. Integrar `record`/`summary` con `Plan()` real cuando 2.2 traiga un manifiesto de ejemplo: asignado a 2.2.
2. Lista sensible de 05 y plazo de borrado de originales (D6, TBD): reemplazar `keyCategories` y los patrones de ruta cuando 05 los defina; asignado a 05.
3. Alcance de `ScopeUnknown` y sin agregado (supuestos 2 y 3): confirmar con muestras reales: asignado a 2.3.
