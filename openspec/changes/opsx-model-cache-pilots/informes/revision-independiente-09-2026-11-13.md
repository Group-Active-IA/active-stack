# Revisión independiente de la preparación offline del cambio 09

13/11/2026. Se revisó el paquete `eval/modelcache` y su CLI local desde un agente distinto de sus implementadores. El resultado acredita contratos de datos, cálculo y preparación offline; no acredita corridas reales, configuración efectiva de clientes, ahorro ni adopción.

## Alcance y referencias

- Rama de trabajo: `propose/opsx-model-cache-pilots`, propuesta original `89f1fa5de22304cc3564c50cf34126cf7e1dd651`, con main `d5e95bd81d3bbab2d54a39d1c07d1256954b5384` incorporado al árbol y la implementación local de 09 pendiente de publicación al revisar.
- Fuentes: proposal, design, spec y tasks de 09; APIs públicas de `eval/baseline/{manifest,usage,summary,fixture,timing,record}`; código y pruebas nuevos de `eval/modelcache`.
- Esta revisión no modificó producción: añadió únicamente `eval/modelcache/independent_report_probe_test.go`, conservado como regresión versionada, y este informe.
- No se ejecutaron clientes, modelos, llamadas pagas, cambios de perfiles/dotfiles, routing, build del producto, commit ni push. El CLI se ejercitó mediante su función `run` en tests.

## Conclusión técnica

No quedan bloqueantes nuevos identificados en el alcance offline inspeccionado. Se encontraron errores de contabilización y comparabilidad durante la revisión, se reprodujeron antes de corregirlos y el implementador los resolvió. Las regresiones y la tanda final están verdes.

| Contrato | Comprobación |
|---|---|
| Independencia de pilotos | A admite modelo o esfuerzo de un rol, manteniendo los demás parámetros y layout; B mantiene modelos/esfuerzos y reordena el mismo conjunto de archivos/bytes. Un brazo combinado no se presenta como contraste aislado. |
| Parámetros solicitados y efectivos | Configuración efectiva vacía/desconocida o fallback respecto de lo solicitado impiden comparabilidad. Esos valores siguen siendo datos declarados: no se probó que un cliente real los aplicara. |
| Layout | Archivos nombrados, bytes, orden y alcance quedan identificados. Traversal, paths repetidos y cambios de bytes fallan. El snapshot no afirma orden real del prompt ni carga por el cliente. |
| D1–D5 pendientes | Decisiones `nil` permanecen pendientes. La única fase admitida es calibración; no se habilita confirmación por completar números de un archivo. |
| Calibración por brazo | La ausencia de estabilidad de 02 o de resultado de 07 no impide preparar/calcular datos separados. Impide concluir adopción: la gate permanece `unavailable`. |
| Caché atribuida | Una sesión nueva, latencia o uso agregado no prueban el estado del prefijo. Se exige señal atribuida al layout materializado declarado para clasificar lecturas. |
| Frío/caliente y datos degradados | El reporte separa condiciones asignadas, conserva observado, contaminación, miss y motivo de desconocimiento. Frío frente a caliente no constituye un par comparable. Misses y lecturas parciales calientes genuinas permanecen en su brazo asignado. |
| Preparación | Cada evento/source ID se imputa una vez; se rechaza reutilizar la fuente como intento y preparación. Un bloque con dos condiciones y preparación sin asignación inequívoca conserva el costo una vez y no produce ratios completos por condición. |
| Fallos, reintentos y denominador | El oráculo 20 de fallo + 30 de reintento + 50 de éxito + 10 de preparación produce 110 para una tarea aceptada. Cero aceptadas conserva totales y deja el cociente sin valor. |
| Agregados padre/hijos | Dinero/cuota con padre inclusivo no vuelven a sumar hijos; padre exclusivo los agrega una vez. Tokens adaptan el contrato público de 02 sin modificarlo. Alcance desconocido y señales ausentes no se convierten en cobertura completa. |
| Modalidades y señales ausentes | API, suscripción, cantidades observadas/estimadas y unidades se separan. La fragmentación de cobertura no oculta intentos fallidos o preparación desconocida en una fila de éxitos. Cuota o tokens no se convierten en dinero inventado. |
| Calidad y cálculo condicional | Fallo obligatorio impide recomendar la variante. Sin objetivos/umbrales explícitos o señales suficientes, el contraste es inconcluso. `candidate_for_review` es aritmética condicional, no inferencia verificada ni adopción. |
| Procedencia y suficiencia | Datos declarados, referencias y hashes no habilitan adopción. El reporte explicita que faltan evaluación independiente de 06 y control/revisión de 07. |
| Reconstrucción | El reporte conserva revisiones 02/07, protocolo, fixture/rúbrica, layouts y source IDs, más una huella del Dataset tipado. El método nombra `Go encoding/json`; no se presenta como hash de bytes del archivo original, TreeHash, recibo o procedencia independiente. |
| CLI | Tres operaciones locales, JSON estricto/EOF, stdin/archivo, flags/argv, root canónico y escape por symlink comprobados. `validate` valida protocolo/layout; `report` además analiza observaciones. No ejecuta procesos de clientes ni cambia configuración. |

Cada Dataset representa una campaña declarada. Reintentos son elementos de `attempts`; no se reetiqueta una repetición duplicada como una corrida nueva cambiando su basis o ID de observación. Campañas diferentes se analizan separadamente. Esto es validación de identidad experimental declarada, no corroboración de identidad física de una ejecución.

## Hallazgos corregidos y ciclo de pruebas

| Hallazgo | RED observado | Corrección del implementador y GREEN |
|---|---|---|
| Tokens padre exclusivo omitían el agregado; overflow y scope sin agregado | Detectados en lectura; el implementador ejecutó sus RED de tokens | Wrapper de 09 adapta el vector público de 02, rechaza overflow y scope contradictorio; no cambia 02. |
| Una fila de éxitos observados ocultaba el intento fallido sin basis/costo; caso equivalente de preparación | `TestIndependentReportMissingFailureBasisCannotMakeSuccessSubsetComplete` y variante de preparación fallaron con ratio completo | Se conservan subtotales y conteos, se identifica fragmentación y se anulan ratios completos de la población incompleta. Cero observado y señal ausente siguen distintos. |
| Dos celdas iguales entre sí pero contradictorias con la cohorte congelada seguían siendo comparables | Cuatro RED: cliente, versión, permisos y herramientas | Se coteja cada celda con la del manifiesto; se normalizan flags, herramientas e interfaz real y se controlan modelo/esfuerzo aparte. |
| B comparaba frío con caliente y descartaba detalles de caché | RED del contraste y de preservación de expected/observed/flags/reason | Condición asignada forma parte de la agrupación y del par; `Caches` conserva evaluación por corrida. |
| Preparación común a dos condiciones se escondía en un ratio mezclado | RED del bloque frío/caliente con una preparación de 10 | Dos condiciones conservadas, costo de preparación una vez y ratios incompletos si la asignación es ambigua. |
| Mejoras aparentes con frío contaminado, prefijo desconocido o aislamiento ausente | RED/GREEN ejecutados por el implementador | Esos casos quedan inconclusos y conservan sus costos. La revisión reejecutó las pruebas. |
| Posible exclusión interesada de miss/lectura parcial caliente | Triangulación positiva, sin nuevo cambio de producción | Miss frente a lectura parcial conserva resultados; con costos sintéticos 50/25 y criterios explícitos produce 50 % aritmético, pero adopción sigue `unavailable`. |

### TDD Cycle Evidence

| Tarea de revisión | Archivo | Safety net | RED | GREEN | Triangulación / refactor |
|---|---|---|---|---|---|
| Cobertura y controles de cohorte | `independent_report_probe_test.go` | Familias existentes verdes del implementador | Ejecutado por esta revisión | Regresiones finales verdes | Basis desconocida, preparación, cuatro mutaciones de celda, cero observado y pérdida de señal |
| Condición de caché y preparación compartida | Mismo archivo | Familias anteriores verdes | Ejecutado antes del fix | Regresiones finales verdes | Frío/caliente, flags/motivos, miss/parcial y costo de preparación una vez |
| Cuota, cero aceptaciones y tiempos | Mismo archivo | Core verde | Solo triangulación, sin producción nueva | Cinco eventos PASS dirigidos | Dinero desconocido/cero/parcial; tokens y tiempo conocido del recorder sintético de 02; cuota sin conversión monetaria |
| Revisión final | Core/CLI y APIs 02 | Ver tabla siguiente | No se atribuyen RED del implementador a esta revisión | 60 eventos PASS finales de 09 | Vet y race verdes; esfuerzo aislado y dos roles reejecutados por separado |

## Verificación independiente final

| Comando | Resultado |
|---|---|
| `go test -count=1 -json ./eval/modelcache/... ./eval/baseline/manifest ./eval/baseline/usage ./eval/baseline/summary` | Corte inicial: 151 eventos PASS; 43 del core 09 y 108 de APIs 02 (75 manifest, 17 usage, 16 summary). Precede a la adición de cinco eventos de triangulación y del CLI. |
| `go test -count=1 -json ./eval/modelcache/...` | Corte final: **60 eventos PASS**, sin fallos ni skips: core 48 y CLI 12. Incluye fixture sintético y regresiones independientes. |
| `go vet ./eval/modelcache/...` | Exit 0 sobre core y CLI finales. |
| `go test -race -count=1 ./eval/modelcache/...` | Exit 0 sobre core y CLI finales. |
| `go test -count=1 ./eval/modelcache -run '^TestCompareArmsSeparatesModelAndEffort$'` | PASS tras la triangulación adicional de esfuerzo aislado y cambio de dos roles; no cambió producción. |
| `openspec validate opsx-model-cache-pilots --strict` | Válido. |
| `git diff --check` y `git diff origin/main --check` | Exit 0 en el corte revisado. |

No se suman las tandas repetidas para presentar una cantidad de casos únicos. La suite Go global es evidencia separada del implementador, no una ejecución de esta revisión. El diff propio de 09 respecto de main no modifica `internal/`, `cmd/active-stack/` ni `eval/baseline/`.

## Logs locales

Son archivos locales de reproducción. Sus hashes acreditan integridad de esos archivos; no crean un contrato de recibos de 06.

| Archivo | SHA-256 |
|---|---|
| `/tmp/opsx09-independent-report-red.log` | `4a71e3d616a3e5d32769760570d21e99c0802900eee8653ea6f4d3b75aecf516` |
| `/tmp/opsx09-independent-report-red-triangulation.log` | `180da82c9473f56fe94efb9a42a97240edde0606587b1e66860cc41796a3ffeb` |
| `/tmp/opsx09-independent-cache-red.log` | `2e91ea152adc8a9f3a5a1222e8bfb070b93662892d6a7b0dd01f06c433bf11a0` |
| `/tmp/opsx09-independent-cache-preparation-red.log` | `112b2a2c61d056a256c28a35121a1605dfc00349c44728e703bbd6ce3bd42c8b` |
| `/tmp/opsx09-independent-final-package-cli.jsonl` | `44a0f0d5e6ed6f6569760c419df1f9402f4f6cf23ebe496558238abc6d103718` |
| `/tmp/opsx09-independent-final-package-cli-vet.log` | `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` |
| `/tmp/opsx09-independent-final-package-cli-race.log` | `8b93c2518415cf600a08a7f28e8b5de40af45ae09c8e7900fbd19b2457a6b65a` |

### Identidad del código revisado

| Ruta bajo `eval/modelcache/` | SHA-256 |
|---|---|
| `amount.go` | `b9710b744d64fc6a50be33607e9869a142c8834db673553e5b176395e95573aa` |
| `layout.go` | `4f27b050a52404ae86e49fe4434f328d3c98d4e372c4a6c562f034aaf1e51417` |
| `protocol.go` | `685af401c4c1bc30f1af5278e025253bd489a91f5fedefa95035d326735fc3d0` |
| `compare.go` | `a987a75ec7a6d723b60b2ff57a55163136ba200cae29e81c870df5aab5ec8b66` |
| `cache.go` | `a2f9c0e79d7d2320f6ea92f835827be58518d926d4bd0518efddad78e5ec5699` |
| `report.go` | `4bc6665925c64827906f246f81bc9060477a12aeb74947875ef67039790154b2` |
| `contrast.go` | `0c4130277f544772060c319544ecf5e5bab9b9703f08fdfc6acbc174d39c4653` |
| `cmd/modelcache/main.go` | `85b85086a26ae6ea20a488401fa22cab62e6315906986727953d3e8d11b8b0d0` |

## Qué no acredita esta entrega

No se ejecutaron los pilotos reales ni se adoptaron modelos/esfuerzos, estrategias de prefijo o routing. D1–D5 siguen sin decisiones inventadas. La aritmética no establece estabilidad estadística, equivalencia real de controles, procedencia de recibos ni permisos efectivos de un cliente. Una referencia semántica o un valor efectivo declarado no es corroboración.

Antes de confirmar o adoptar faltan el protocolo/muestra/umbrales/presupuesto acordados, la referencia de 07, las señales reales calibradas por cliente/interfaz y su evidencia independiente suficiente. La preparación offline puede revisarse y publicarse por separado sin presentar esos pendientes como realizados ni archivar el cambio completo.
