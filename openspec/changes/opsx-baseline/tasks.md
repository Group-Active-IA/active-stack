# Tasks · Medición de referencia de OPSX

Las tareas que dependen del árbitro de D1, de D3, de D5 o del plazo de borrado de D6 usan los valores que el grupo acuerde antes de la entrega final ([registro de decisiones](design.md#registro-de-decisiones)). Las corridas de referencia de Codex usan el contrato de Estricto vigente de 01 para la serie separada. El protocolo puede construirse y validarse con eventos sintéticos antes de esas corridas.

## 1. Protocolo y fixtures

- [x] 1.1 Versionar el manifiesto y la matriz de cohortes por cliente, versión, sistema, interfaz real, modalidad y permisos efectivos, con tareas y pesos por cohorte. Verificar con manifiestos de prueba que una celda con versión incompatible, permisos distintos de los declarados o un opt-in de contenido activo queda incompatible antes de iniciar, fuera del denominador y listada con su motivo.
  - Código: `eval/baseline/manifest/` (`manifest.go`, `manifest_test.go`, `plan_test.go`, `load_test.go`); `eval/baseline/README.md`. Base `43eeb7f`, cambios sin commitear (ver `git status` / informe `informes/apply-1-constructor.md`).
  - Safety net: paquete nuevo, sin archivos previos que modificar; sin go.mod/go.sum tocados.
  - RED (2026-11-11 10:33, `go test ./eval/baseline/manifest/...`, esqueleto que devuelve siempre `ejecutable`): `FAIL TestCheckCell_Version/menor_que_la_minima: status = "ejecutable", want "incompatible"` (3 subcasos). Mismo patrón por aserción en permisos (3 subcasos, 10:36:26), opt-ins (5), terminal integrada de IDE (1), `Plan` (`Executable = [4 cohortes], want [2]`; 10:36:55), estados (`"aceptada" debería ser válido`) y `Validate`/`Load` (10:37:13).
  - GREEN (`go test ./eval/baseline/manifest/...`): `ok` en 10:36:09, 10:36:36, 10:37:02, 10:37:21. Cobertura 95.3 %.
  - Triangulación: versión igual/mayor/comparación numérica (`2.1.99` < `2.1.310`, `2.1.1000` ok)/no interpretable; flags en otro orden = ejecutable, flag extra/faltante = incompatible; opt-in apagado = ejecutable, valor desconocido = activo; las 4 señales de contenido; plan con todas ejecutables, todas incompatibles y vacío.
  - Mutantes (revertidos, diff idéntico al original verificado): M1 `Plan` deja incompatibles en el denominador → falla `TestManifestPlan_*` (2); M2 ignora flags de permisos → falla `TestCheckCell_Permissions`; M3 opt-in apagado cuenta activo → falla `TestCheckCell_ContentOptIns`; M4 ignora terminal integrada → falla `TestCheckCell_IDEIntegratedTerminal`; M5 `<` por `<=` en versión → fallan 6 tests.
  - Chequeos: `gofmt -l eval` sin salida; `go vet ./eval/...` ok; ningún archivo de `cmd/` ni `internal/` importa `eval/`.
  - **Corrección apply-2 (F2–F8 de `informes/verify-1-breaker.md`; informe `informes/apply-2-constructor.md`).** Test nuevo `manifest/apply2_test.go`; `plan_test.go` adaptado a `Plan() (Plan, error)` y `Executable []Planned`; en `manifest_test.go` la fila de terminal integrada ahora espera `ejecutable` (F8).
    - Safety net 10:53:04: `go test -count=1 ./eval/baseline/...` ok.
    - RED 10:55:37 (F5: `Plan() de un manifiesto vacio debe fallar por Validate`) y 10:56:50 (`go test -count=1 ./eval/baseline/manifest/`): F2 9 subcasos `status = "ejecutable", want incompatible`; F3 7 subcasos y orden de motivos; F8 `CheckCell = {Status:incompatible ...}, want ejecutable en terminal`; F6/F5 4 subcasos `Validate() = <nil>`; F4 `Load() debe rechazar un segundo documento`.
    - GREEN 10:57:23: `ok`. Cobertura 96.4 %.
    - Nacidos verdes, verificados por mutante: `TestLoad_UnknownFieldOnOtherwiseValidManifest` (M_unknown), `TestCheckCell_NegativeVersionComponent` (M_neg), subcaso "apagado con espacios" (M_trim).
    - Mutantes 10:58 (18, todos detectados y revertidos con diff idéntico): MF2a–e (cliente, escritorio, sistema, modalidad, permisos vacíos), MF3a–d (prefijo, mayúsculas, tabla de Codex, orden), M_trim, MF8 (sin reclasificar), MF6 (variante), MF5a–c (task_id vacío, duplicado, Plan sin validar), MF4 (contenido extra), M_unknown, M_neg.
    - Excepciones abiertas (no en el alcance de 1.1): plataformas distribuidas/declaradas/probadas (D2) y "señales disponibles" de la comprobación previa no están modeladas.
  - **Corrección apply-3 (R1 y N_modeTrim de `informes/verify-2-breaker.md`; informe `informes/apply-3-constructor.md`).** Test nuevo `manifest/apply3_test.go`; `isContentOptIn` excluye `otel_log_level` normalizado; nota de una línea en `design.md` (Datos sensibles).
    - Safety net 11:07:44: `go test -count=1 ./eval/baseline/...` ok en ambos paquetes.
    - RED 11:08:03 (`go test -count=1 -run 'OTELLogLevel|BlankPermissionMode' -v ./eval/baseline/manifest/`): 3 subcasos `CheckCell = {Status:incompatible Reasons:[opt-in de contenido activo: OTEL_LOG_LEVEL] ...}, want "ejecutable"` (también `otel_log_level`, `Otel_Log_Level`) y `motivos = [... OTEL_LOG_LEVEL ... OTEL_LOG_USER_PROMPTS], want solo OTEL_LOG_USER_PROMPTS`. `OTEL_LOG_USER_PROMPTS` y `OTEL_LOG_LEVELS` siguen incompatibles (pasan).
    - GREEN 11:08:21: `ok` (focalizado y `./eval/baseline/...`).
    - Nacido verde: `TestCheckCell_BlankPermissionModeIsEmpty` (modo `" "` y `"\t "`, declarado = observado): el código ya trataba el modo en blanco como vacío.
    - Mutantes (revertidos, `git diff --no-index --quiet` idéntico al respaldo): M_R1 11:08:36 (`if false && n == "otel_log_level"`) → falla `TestCheckCell_OTELLogLevelIsNotContentOptIn` (4 aserciones); N_modeTrim 11:08:39 (sin `TrimSpace`) → falla `modo " ": CheckCell = {Status:ejecutable ...}`.
- [ ] 1.2 Definir los perfiles de medición por cliente e interfaz: perfil equivalente para edición y serie separada de Estricto. Verificar con el cliente real en Windows y Linux que el modo y los flags efectivos coinciden con el manifiesto y que una edición dentro del checkout no pide aprobación en el perfil equivalente.
- [ ] 1.3 Preparar los fixtures, prompts y rúbricas de las cinco familias. Ejecutar sus chequeos contra una respuesta o diff de referencia y contra una variante defectuosa. Verificar reinicio desde el fixture inmutable y restauración entre repeticiones.
- [x] 1.4 Implementar la regla de emparejamiento de la revisión de un diff sembrado. Verificar con una clave y hallazgos sintéticos un verdadero positivo, un duplicado que no suma, una omisión, un falso positivo y un hallazgo nuevo derivado a adjudicación.
  - Código: `eval/baseline/review/` (`review.go`, `review_test.go`). Base `43eeb7f`, cambios sin commitear.
  - RED (2026-11-11 10:39:06, `go test ./eval/baseline/review/...`, esqueleto que devuelve `Result{}`): 18 pruebas/subpruebas fallan por aserción (p. ej. `TestMatch_TruePositive`, `TestMatch_UnmatchedFindings`, `TestMatch_InvalidInputErrors/*`).
  - GREEN (10:39:21): `ok`, cobertura 100 %. Tras refactor de comentarios, `ok` otra vez.
  - Casos: verdadero positivo, duplicado que no suma, omisión, falso positivo, hallazgo nuevo (no altera TP/FP/omisiones), pendiente de revisor; bordes: las tres condiciones exigidas por separado, solape inclusivo/adyacente, normalización de ruta, clave vacía, hallazgos vacíos, un hallazgo hacia dos defectos, hallazgo sobre defecto ya emparejado, entradas inválidas.
  - Mutantes (revertidos, diff idéntico verificado; 10:40): M1 duplicado cuenta como TP → falla 2 pruebas; M2 solape exclusivo → falla 2; M3 mecanismo ignorado → falla; M4 hallazgo nuevo cuenta como FP → falla; M5 archivo ignorado → falla.
  - Chequeos: `gofmt -l eval` sin salida; `go vet ./eval/...` ok. Ambigüedades de la rúbrica y lecturas conservadoras elegidas: ver SUPOSICIÓN en el informe.
  - **Corrección apply-2 (F1 y F7/M_emptymech de `informes/verify-1-breaker.md`).** Emparejamiento máximo uno a uno (caminos de aumento) con desempate estable: hallazgos en orden de lista, candidatos por ID, primero un candidato libre. Test nuevo `review/apply2_test.go`; en `review_test.go` el test del defecto ya tomado se renombró a `TestMatch_FindingMatchingTakenDefectTakesFreeSecondDefect` y ahora espera TP de D2 (reemplaza la lectura "se ata al primero").
    - RED 10:53:30 (`go test -count=1 ./eval/baseline/review/`): `clave A,B: counts={TruePositives:1 Duplicates:1 ... Omissions:1}`, `asignacion depende del orden de la clave: map[f1:A f2:A] vs map[f1:A f2:B]`; `TestMatch_MaximumMatchingStillOneToOne` falla por conteos y asignación.
    - GREEN 10:54:13: `ok`. Cobertura 100 %.
    - Nacidos verdes, verificados por mutante: `TestMatch_MaximumMatchingNeedsAugmentingPath` (MF1a) y `TestMatch_EmptyMechanismOnBothSidesDoesNotMatch` (M_emptymech).
    - Mutantes 10:54–10:55 (4, todos detectados y revertidos con diff idéntico): MF1a sin camino de aumento, MF1b sin ordenar por ID, MF1c Kuhn sin preferir libres, M_emptymech. Los primeros MF1a y MF1b no compilaron, así que no cuentan; se repitieron en forma compilable y ambos fallaron.
    - S1–S4 sin cambios: disposiciones intactas.
  - **Corrección apply-3 (N_omOrder de `informes/verify-2-breaker.md`).** Test nuevo `review/apply3_test.go` (`TestMatch_OmissionsFollowKeyOrder`, clave D3,D1,D2,D0 con un hallazgo sobre D1 → omisiones `[D3 D2 D0]`). Sin cambio de producción: nació verde.
    - GREEN 11:08:21: `ok`.
    - Mutante N_omOrder 11:08:51 (bucle de omisiones sobre `byID` en lugar de `key`) → falla `omissions = [D0 D2 D3] (3), want [D3 D2 D0]`; revertido con diff idéntico. Un primer intento 11:08:42 alteró también el bucle de validación y no compiló; no cuenta.

## 2. Arnés y captura

- [ ] 2.1 Implementar el ciclo de preparación, ejecución, intento y evaluación en perfiles descartables. Probar aceptación, rechazo por calidad, bloqueo por permisos, reintento, timeout y fallo del cliente, y verificar que cada caso queda con un estado distinto, la causa, el código de salida bruto y los eventos de error.
- [ ] 2.2 Implementar los adaptadores de captura para Claude Code y Codex en terminal e IDE. Probar con muestras reales de eventos de cada versión fijada, o con muestras sintéticas declarando el riesgo, en casos completos, parciales y duplicados. Incluir una corrida delegada por cliente y verificar que la regla de precedencia reproduce los totales de los ejemplos de reconciliación del diseño.
- [ ] 2.3 Comprobar en Claude la atribución por worker con `agent_id` desde la versión 2.1.310, que `agent.name = "custom"` no se usa para atribuir y que el total del cierre de un subagente no se toma como su consumo total. Comprobar en Codex si el agregado del padre incluye a los hilos hijos. Documentar las fuentes y los límites de atribución de cada versión e interfaz.
- [ ] 2.4 Implementar el registro de lecturas de archivos y llamadas a herramientas por tipo y rol. Verificar con una traza conocida que los conteos coinciden y que una interfaz sin esa señal la marca como no disponible.
- [ ] 2.5 Implementar la medición del tiempo con reloj monotónico, esperas observables aparte y unión de intervalos solapados. Verificar con intervalos sintéticos solapados y con una espera no delimitable que el ajustado figura como no disponible.
- [ ] 2.6 Implementar la sonda de sesión normal y el registro de caché. Comprobar con fixtures que distingue entrada de una solicitud, total del turno y métricas ausentes, y que una sesión nueva con lecturas cacheadas conserva el estado observado.
- [ ] 2.7 Emitir registros por intento y corrida con metadatos de cobertura aplicando la lista sensible y la política de 05 y las huellas de 06. Verificar con una traza sintética que incluya correo, UUID de cuenta, identificador de usuario, rutas del host, detalle de herramientas y cuerpos de solicitudes que ninguno llega al resultado compartido.

## 3. Resultados y referencia

- [ ] 3.1 Generar el resumen por cohorte y documentar su cálculo. Probar con resultados sintéticos que rechazos, bloqueos y reintentos contribuyen al costo total, que las celdas incompatibles quedan fuera del denominador, que API, suscripción e interfaces quedan separadas y que una cohorte sin aceptadas publica el costo total con el cociente sin valor.
- [ ] 3.2 Ejecutar una corrida de calibración por cliente e interfaz después de verificar las precondiciones. Incluir casos directo y delegado para comprobar correlación, permisos, fuentes y reglas de parada. Excluir las calibraciones de la muestra medida.
- [ ] 3.3 Ejecutar una corrida de calibración con una aprobación de demora conocida por cliente e interfaz. Verificar qué métricas de tiempo cambian y que el ajustado descuenta la demora solo cuando los intervalos quedan delimitados.
- [ ] 3.4 Ejecutar las repeticiones de la sonda y del flujo OPSX actual según los valores acordados en D3. Publicar el manifiesto, los resultados depurados y el resumen. Verificar que otra persona puede reconstruir los agregados y rastrear los veredictos desde la evidencia autorizada.

## 4. Comparación modular

- [ ] 4.1 Cuando 03 y 04 entreguen el flujo modular instalable, repetir las celdas comparables con el mismo protocolo, fixture, rúbrica y permisos efectivos y publicar los resultados emparejados por cliente, interfaz y modalidad. Verificar que las interfaces sin variante modular quedan como comparación pendiente.

## Workflow follow-up

- Entregar el protocolo, el manifiesto y los resultados de referencia a 07, 09 y 10 para sus pilotos.

## Evidencia de suite completa (apply-1, tareas 1.1 y 1.4)

- `go test ./...` sobre la revisión entregada (base `43eeb7f` + cambios sin commitear), 2026-11-11 10:40:30–10:43:13: todos los paquetes `ok` (incluidos `eval/baseline/manifest` y `eval/baseline/review`); sin fallos previos. `openspec validate opsx-baseline --strict`: `Change 'opsx-baseline' is valid`.
- apply-2: `go test -count=1 ./...` 2026-11-11 10:59:02–11:00:31 sobre la revisión corregida: todos los paquetes `ok`. `gofmt -l eval` vacío; `go vet ./eval/...` ok; `openspec validate opsx-baseline --strict` válido; go.mod/go.sum sin diff.
- apply-3: `go test -count=1 ./...` 2026-11-11 11:09:06–11:10:57 sobre la revisión entregada: 29 paquetes `ok`, ninguna línea `FAIL`. `gofmt -l eval` vacío; `go vet ./eval/...` ok; `openspec validate opsx-baseline --strict`: `Change 'opsx-baseline' is valid`; go.mod/go.sum sin diff.
