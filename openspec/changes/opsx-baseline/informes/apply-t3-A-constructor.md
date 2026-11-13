# apply-t3-A · constructor · opsx-baseline (2.1, 2.4, 2.5 + DEUDA 2, F3, DEUDA 6)

Rol: constructor, LOW. Rama `propose/opsx-baseline`, base `5dcfc68` + tanda 2 sin commitear (aprobada). Fecha 2026-11-11, 14:40 a 15:15 (-03). Sin commits, sin pushes, sin `go build`, sin procesos en segundo plano al terminar. No toqué `openspec/changes/codex-strict-compat/` ni `archive/`.

## Fuentes
- Lecciones L-NNN: busqué por nombre (`*lecc*`, `*lessons*`) y por `L-0` en `*.md` del repo: no hay base de lecciones numeradas. Ausente; apliqué las del informe del auditor de la tanda 2 (hashes pegados de comando; propiedades con oráculo y permutación; ante una desviación de fixture, dejarla visible).
- `openspec/changes/opsx-baseline/tasks.md`, `design.md` ("Medición del tiempo", D1/D9/D10, tabla de registros), `informes/verify-t2-1-auditor.md` y `verify-t2-1-breaker.md` (F1–F3, S1, S3).
- Código existente: `eval/baseline/fixture/*.go`, `manifest/manifest.go`.
- Contaminación: ninguna de informes ajenos más allá de los dos de la tanda 2 que el brief nombró (los usé solo para la lista de DEUDA). No leí informes del rompedor de esta tanda (no existen todavía).

## EVIDENCIA

### Hashes (salida pegada de `git hash-object`, revisión entregada, 15:09)
```
764a31543a0f7dd173123764d891273bc3ea87ab eval/baseline/fixture/load.go
958f9df51bb7630cb4192564367130efdb7f3758 eval/baseline/fixture/evaluate.go
a1b9227cb24f3138cd1fd38ac69bc4bba632bd7c eval/baseline/fixture/t3_test.go
bb257b40014250f946e1e7672e27f3e9625224f3 eval/baseline/fixture/checkout_test.go
f7ca9c1971764ea615c76a2ac986f8669aab1f7b eval/baseline/fixture/evaluate_test.go
4d7acd999387fe67bbc6cdc9be1c0df80cf48825 eval/baseline/run/run.go
3f0af949f813b01fc35055ee0cd17d31ae9f2031 eval/baseline/run/run_test.go
15935946eb6945b978bd4cbef66dc4afcaa5e623 eval/baseline/trace/trace.go
6c9f10f9a3a3ce3f9bbb3ce0dc9e20386279c280 eval/baseline/trace/trace_test.go
b8ca4c3fa8e728d8339c78d0c7a0738d715f76cd eval/baseline/timing/timing.go
c3f9687bca4f71c3a29b46a683f511ef16ad82f4 eval/baseline/timing/timing_test.go
8e819e46f4f1d49cf15b465d1b762835b3e143ca eval/baseline/timing/recorder_test.go
b7dbfa22ea7dac834197fb560669c9ed5ea1ab28 eval/baseline/README.md
c5b635cc1a92208403880dab81ebb986e2a142cd eval/baseline/fixture/load_test.go   (sin cambios de esta tanda)
```
Corrección de la tanda 2: `eval/baseline/fixtures/` tiene **41** archivos (`find eval/baseline/fixtures -type f | wc -l`), no 49 (los 49 incluían los 8 `.go` de `fixture/`); el hash de `load_test.go` es `c5b635cc…` (salida del comando de arriba), no `c5b635c1`.

### Chequeos sobre la revisión entregada
- Baseline antes de tocar `fixture`: `go test -count=1 ./eval/baseline/fixture/` ok (116 s).
- Suite completa una vez, `go test -count=1 ./...` 15:03–15:07: 33 paquetes `ok`, ninguna línea `FAIL` (`fixture` 111.2 s, `run` 2.5 s, `timing` 0.8 s, `trace` 1.2 s). Resultado también en `tasks.md`.
- `gofmt -l eval` vacío; `go vet ./eval/...` ok; `go list -deps ./cmd/... ./internal/...`: 0 coincidencias con `/eval`; `git diff -- go.mod go.sum` vacío; `openspec validate opsx-baseline --strict`: `Change 'opsx-baseline' is valid`.
- Fuera de mi diff, `git status` muestra `design.md` y `proposal.md` de opsx-baseline como modificados: no los toqué en esta tanda.

### Tabla comportamiento → RED → GREEN → mutante
| Comportamiento | RED (aserción) | GREEN | Mutante que falla |
|---|---|---|---|
| 5 estados distintos por caso, causa, código bruto, eventos `turn.failed`/`error` intactos | `aceptacion: Status = "" con 0 intentos, want "aceptada"` | run ok | timeout, bloqueo, rechazo cambiados; código descartado; eventos descartados |
| Reintento desde el fixture reiniciado, causa del reintento, intentos conservados | `"" con 0 intentos, want aceptada con 2` | ok | sin `Reset`; causa genérica; `max` corrido |
| Bordes de política (sin política, máx. 1, agotado, estado fuera de RetryOn, calidad no reintenta) | `"" con 0 intentos, want "timeout" con 2` | ok | reintento sobre cualquier estado |
| Timeout por plazo real, error del cliente, cancelación del llamador | `Run = {...}, want timeout por el plazo` | ok | `DeadlineExceeded` ignorado; M_parent (sobrevivió, se agregó prueba) |
| `ExecutePlan` solo cohortes ejecutables y repeticiones | `corridas = [], want [ok-1#1 ok-1#2 ok-2#1 ok-2#2]` | ok | plan con una incompatible; 0 repeticiones aceptadas |
| Conteos 2.4 exactos, no disponible ≠ cero, errores de registro | `KnownTraceExactCounts` (0 vs 5/4) | trace ok, 100 % | total ×2, tipo/rol cruzados, cero en vez de no disponible, sin validar |
| Unión de intervalos (ejemplos + propiedad con oráculo y permutaciones) | `uno: Union = [], want [{1ns 4ns}]` | timing ok, 100 % | sin fusionar contiguos, sin extender, sin ordenar, suma ingenua |
| Ajustado no disponible si una espera no se delimita | `Stop sin Start = nil, want error`; `m = {Elapsed:0s…}, want elapsed 10…` | ok | abiertas/duplicadas/huérfanas ignoradas, `waitsObservable` ignorado |
| DEUDA 2: directorio vacío agregado | `directorio vacio: E4 = "cumplido", want "incumplido"` (3) | ok | sin `emptyDirs` |
| DEUDA 2: entradas no regulares = incumplido, no error | `symlink: err = <nil>, want errNonRegular` (3) | ok | `case false` en `hashVerdict` (sobrevivía hasta `TestHashVerdict`) |
| DEUDA 2: guarda de rutas `Source.Dir`/`Args[0]` | 7 subcasos `Load = <nil>` / mensaje distinto | ok | sin `safeName` en dir, args, barra, contrabarra, volumen; sin `Stat` |
| F3 README `--text` | n/a (documentación; reproducido el hash `dca4f76c…` en Git Bash) | | |

Pruebas que nacieron verdes y se acreditaron por mutante: `TestExecute_CancelDuringRun…` (M_parent), `TestSystemClock_OriginKeepsMonotonicReading` (M_strip), los casos de `/etc`, `\x` y `C:x` de `TestLoad_RejectsUnsafeSourceDirAndHiddenArgs` (la guarda ya cubría el caso base) y `TestHashVerdict`. El RED de `TestSystemClock_*` fue un pánico por puntero nulo del esqueleto, no una aserción; no cuenta como RED y la sensibilidad se acreditó por M_strip.

## JUICIO
- 2.1 está completo salvo lo que depende de un cliente real (2.2): el `Runner` es una interfaz y todo se probó con runners falsos deterministas.
- DEUDA 6 queda a medias: la reducción es de ~10 %, no de orden de magnitud. El costo es de E/S en Windows (apertura de ~650 archivos en cada `Reset`), no de lógica; sin recortar la verificación de la huella no hay más margen. Lo documenté en README y `tasks.md`.
- El control de symlinks no se ejercitó con un symlink real en este host (sin privilegio): el riesgo residual es que el recorrido real no entregue `ModeSymlink` como supuse; `WalkDir` no sigue symlinks y reporta `d.Type()`, así que es bajo.

## SUPOSICIÓN
1. Los estados de corrida salen de `Result.Termination` (lo fija el adaptador de 2.2) y del veredicto; los eventos `turn.failed`/`error` se copian a `ErrorEvents` sin influir en el estado. Verificable cuando 2.2 traiga muestras reales.
2. `pendiente_de_revisor` se modela como estado no final de la corrida (`StatusPendingReview`, no válido en `manifest.RunStatus.Valid`); 3.1 debe decidir cómo lo trata.
3. Política de reintento: `Policy{MaxAttempts, Timeout, RetryOn}` desde `manifest.Limits`; cero es no fijado y no reintenta (valores reales: D3, TBD). Calidad y bloqueo por permisos no se reintentan salvo que se listen en `RetryOn`.
4. Una lectura de archivo (2.4) no cuenta además como llamada a herramienta; el adaptador registra ambas si la interfaz las expone por separado. Los conteos de lecturas llevan también rutas únicas.
5. Un directorio vacío en un checkout es siempre una adición del agente, porque git no versiona directorios vacíos y los fixtures embebidos viven en git.
6. Una espera con resolución sin solicitud, solicitud duplicada o fin anterior al inicio hace el ajustado no disponible (no es error de uso): son datos que puede traer un adaptador real.

## Limitaciones y chequeos no ejecutados
- Sin clientes reales, Linux ni macOS; los casos de ruta `\x` y `C:x` solo corren en Windows.
- Una corrida interrumpida por el llamador (contexto cancelado durante la corrida) devuelve error y la corrida parcial sin el intento en curso; conservar ese intento con su consumo parcial queda para 2.2/2.7.
- No hay cobertura de `Execute` por encima de 91.7 % (ramas de error de `Prepare`/`Reset` no forzadas).
- No hice 2.2, 2.3, 2.6, 2.7, 3.1, 1.2, 3.2–3.4 ni 4.1.

## DEUDA (nueva, no bloquea)
1. Tope de duración del paquete `fixture` (≈110 s): calcular la huella al escribir en `Reset` en lugar de releer; decide el responsable de 02. Asignado a 2.2 (primer change que ejecuta muchas corridas).
2. Conservar el intento parcial ante cancelación del llamador: asignado a 2.7.
3. Tratamiento de `pendiente_de_revisor` en el resumen: asignado a 3.1.
