# Apply incremental offline del cambio09 · 2026-11-13

## Alcance entregado

`eval/modelcache` prepara y analiza datasets declarados de calibración de los dos pilotos, consumiendo APIs públicas de `eval/baseline` sin reescribir el 02. Se incorporó main `d5e95bd81d3bbab2d54a39d1c07d1256954b5384` con merge sin conflictos; el coordinador realizará commit/push. Entorno: Go1.27.3 darwin/amd64. No se ejecutó build del producto, llamadas pagas, routing ni escrituras de perfiles.

Se validan manifiesto, decisiones opcionales, revisión/fixture/rúbrica, layouts de archivos por hash y orden, modelo/esfuerzo aislados, prefijo equivalente por reordenamiento, condiciones asignadas/observadas y costos completos de intentos y preparación única. Hay reconciliación explícita de dinero, cuota y tokens; un agregado que incluye workers no se vuelve a sumar. Las medidas faltantes conservan el intento y otras métricas. Costos estimados/observados, modalidades y unidades se separan; las fragmentaciones no aparentan cobertura completa.

El reporte preserva las fuentes declaradas y la huella JSON tipada del Dataset; no acredita ejecución ni independencia. Un Dataset es una campaña. Las repeticiones duplicadas se rechazan antes de fragmentar grupos; los retries son eventos dentro de cada corrida. Los contrastes exigen pares completos del conjunto iniciado, celdas congeladas, condiciones comparables y decisiones explícitas. La contaminación fría/atribución desconocida/aislamiento faltante impiden atribuir mejoras; misses y lecturas parciales calientes genuinos permanecen incluidos. Una regresión obligatoria invalida una mejora. `candidate_for_review` es aritmética condicional; adopción siempre `unavailable`.

## TDD Cycle Evidence

| Familia | Archivo de test | Capa | Safety net | RED | GREEN | Triangulación/refactor |
|---|---|---|---|---|---|---|
| Reconciliación | amount_test.go | unit | APIs02 verdes | amount/tokens/parent/validation-red | amount-final | parent-only, includes-children, unknown, missing, cero, overflow; código mínimo |
| Layout | layout_test.go | unit | paquete nuevo | layout-red | layout-green | orden, scope, paths, ausencia, LF/CRLF |
| Protocolo | protocol_test.go | unit | paquete nuevo | protocol-red, protocol-factor-red | protocol-green | parámetros null, factores cruzados, revisión ausente |
| Comparación | compare_test.go | unit | previo verde | compare-red | compare-green | un factor por rol, fallback, reorder y bytes/ref |
| Cache | cache_test.go | unit | previo verde | cache-red | cache-green | miss, parcial, completo, agregado, contaminación |
| Reporte | report_test.go | unit/integración pura02 | 02 manifest/record/summary/usage verdes | report-red, trial-red, source-red | report-green, source-green | fallos/retries/prep, duplicados, huella de fuente |
| Regresiones independientes | independent_report_probe_test.go | unit | paquete previo verde | independent-report-red[-triangulation], independent-cache-red, independent-cache-preparation-red | independent-report-green-review, independent-cache-green | fragmentación basis/unidad, celdas, condiciones, prep ambigua |
| Contrastes | contrast_test.go | unit | previo verde | contrast-red y contrast-cache-red | contrast-cache-green | criterios faltantes, calidad, contaminación/aislamiento/atribución; warm miss/parcial triangulado por revisor |
| CLI | cmd/modelcache/main_test.go | integración in-process | core verde | `/tmp/modelcache-cli-red.log` (run ausente) | `/tmp/modelcache-cli-green.log` | 12 eventos, archivo/stdin, JSON estricto/EOF, paths/symlinks y errores de salida; vet PASS |
| Ejemplo | example_test.go | integración pura | core verde | deriva de contratos RED anteriores | example-green | fixture A+B, layouts en disco, decisiones null y adopción no disponible |

Excepto los paths explícitos del CLI, los nombres de logs de la tabla tienen prefijo `/tmp/opsx09-` y sufijo `.log`. Los RED de API nueva fueron fallos de compilación por símbolos ausentes; los RED de regresión fueron aserciones sobre comportamiento real. El refuerzo de warm miss/parcial y el fixture son triangulación, no se presentan como RED nuevos. Revisión independiente en su informe separado.

## Verificación registrada

Baseline anterior al merge: model/catalog PASS (`/tmp/opsx09-before-merge-baseline.log`). Tras merge: APIs02 manifest/record/summary/usage PASS (`/tmp/opsx09-after-merge-baseline.log`). `go vet ./eval/modelcache/...` PASS. `openspec validate opsx-model-cache-pilots --strict` PASS.

Tanda global previa a incorporar CLI: `go test -json -count=1 ./... -skip '^TestWindowsBundleThemeContainsMinimalInstallerPages$'`, 2124 eventos pass, 6 fail y 3 skip. Los seis fallos corresponden al conjunto histórico: `TestDownloadBinary_APIError` y tests/subtests del fixture review congelado del 02; no se modificaron esas áreas. Log `/tmp/opsx09-full-final.jsonl`, SHA256 `06190349e2e120b08d7e35cd4c76d814014a69050fb3da61d8159da3aacd88d2`. No constituye una suite global verde.

## Pendientes reales

D1–D5, tamaños, presupuesto, objetivo y umbrales reales siguen `(TBD)`. Faltan calibraciones reales por cliente/superficie, 02 estable, A0/resultados de 07, evidencia corroborada y suficiencia 06, ventana/aislamiento efectivo de caché y revisión humana de resultados. Las dependencias se asignan en [el cruce actualizado](estado-cruce-y-dependencias-09-2026-11-13.md). La deuda histórica del 04 no acepta automáticamente pendientes del 09. No se archiva ni se declara terminado el experimento. Las tareas originales se conservan mediante una tabla de trazabilidad.


## Cierre de verificación del árbol integrado

Tras congelar core y CLI, la tanda global final arrojó **2141 pass / 6 fail / 3 skip** (eventos de tests y subtests), con 38 paquetes pass y 2 paquetes fail: `/tmp/opsx09-full-postcli.jsonl`, SHA256 `cf7438e811bba73b32b53a8f47af78ff011e9d4be8ac9bc3afd9fcc3a76d77b3`. Se ejecutó el mismo comando global indicado arriba. Los seis eventos fallidos son el mismo conjunto histórico enumerado; no se declara una suite global verde ni se corrigen otros cambios. Los tres skips se conservan explícitos.

`go test -race -count=1 ./eval/modelcache/...` PASS (`/tmp/opsx09-race-postcli.log`); `go vet ./eval/modelcache/...` PASS (`/tmp/opsx09-vet-postcli.log`); `openspec validate opsx-model-cache-pilots --strict` PASS (`/tmp/opsx09-openspec-strict.log`). Las pruebas CLI corren en proceso: no se invocó go run ni build del producto. Los tests del ejemplo leen archivos ya generados y no escriben en el repositorio.

Se trianguló además esfuerzo aislado positivo y cambios de dos roles negativos, sin cambio de producción, en `TestCompareArmsSeparatesModelAndEffort` (`/tmp/opsx09-compare-final-triangulation.log`). El revisor añadió cero aceptados/money faltante que conserva tokens/tiempo y suscripción con cuota sin conversión ficticia; son refuerzos GREEN, no nuevos RED.

El inventario de seis fuentes Codex en `examples/codex-repo-layout.json` permite congelar un snapshot reproducible con el CLI. Es una selección de fuentes con orden designado, no el prompt completo de toggles ni prueba de lectura del cliente. El paquete se limita a `eval/modelcache`; los archivos de runtime/plantillas de 02/03/04/Claude no se editan respecto de main incorporado.

Entrega: ocho tareas incrementales offline terminadas de doce, con trazabilidad a las ocho originales. Las cuatro pendientes son acuerdos propios (1.3), preparación/corridas reales A (2.2), B (3.2) y revisión/publicación de resultados reales (4.3); las entregas de proveedores externos quedan identificadas por responsable en el cruce. No se marcan hechas por traslado a deuda.


`go vet ./...` también PASS (`/tmp/opsx09-vet-global-final.log`).

## Registro condicionado de dependencias externas

Estos IDs asignan seguimiento al proveedor: no son tareas de implementación pendientes de Matías ni aceptación automática de deuda/valores para09. La acción consumidora se habilita al recibir y verificar la entrega correspondiente.

| ID | Proveedor/fuente | Disparador de consumo por09 |
|---|---|---|
| DEP09-02 | Giuliano02, tareas3.2/3.3/3.4/4.1 en main | Pin de baseline/calibración y celda reproducible; actualizar Dataset para corrida real, sin cambiar02 |
| DEP09-05 | Emiliano05, revisión4337e3658ca656b89cd15725a715db364a73a969 | Captura disponible por celda y contrato de datos acordado; consumir sólo señales demostradas |
| DEP09-06 | Gianfranco06, revisión3c7d90808ff89fcfcad876cde67daf77531e3e2b | Contrato/evaluador disponible en base aprobada, corroboración independiente y suficiencia; integrar frontera sin confiar en booleanos del Dataset |
| DEP09-07 | Neyen07, revisión4922498f8fa6ed5d77a82bda32001535977898fe, D07-05/06 | Resultado/política/control A0 pinneado; hasta entonces calibrar brazos identificados sin adopción |
| DEP09-ROLES | Responsables de interfaces cliente | Prueba del modelo/esfuerzo efectivo por rol y superficie antes del contraste correspondiente |
| DEP09-08 | Responsable08/equipo, sólo casos que lo exigen | Control preventivo obligatorio demostrado o equivalente comprobado antes de esa corrida |

Los acuerdos propios D1–D5, presupuesto y parámetros reales no se trasladan a estos IDs ni se consideran aprobados.
