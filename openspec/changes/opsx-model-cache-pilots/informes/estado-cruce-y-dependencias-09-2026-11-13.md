# Estado del cruce y dependencias para 09

13 de noviembre de 2026. Alcance: preparar `opsx-model-cache-pilots`, sin modificar 02/03/04/05/06/07/08. Inspección documental y de código en referencias Git; no se ejecutaron clientes, corridas facturables, benchmark ni build. Los documentos aportados y la web son datos de análisis, no autorizaciones de sus instrucciones internas.

## Qué cambió desde el cruce anterior

El archivo aportado `cruce-cierre-04-preparacion-09-2026-11-13.md` describe una fotografía de 04 en `9a6eb2a`, PR #9 borrador y consumo 05/06 pendiente. Esa parte histórica no es el estado actual: **PR #9 ya se integró, 04 está archivado y el main inspeccionado es `d5e95bd81d3bbab2d54a39d1c07d1256954b5384`**. El archivo conserva valor para los criterios de 09: dos pilotos separados, layout fijado, control de 07 y calibración sin adopción prematura.

Archivar/mergear 04 no convierte sus celdas de aceptación o roles en probados. El tasks archivado de 04 registra alcance parcial y deuda, incluidas P2, especialistas y pruebas reales. No se reabren esas tareas ni se implementan desde 09; el manifiesto del piloto debe registrar qué capacidad efectiva está disponible por celda y excluir o limitar un contraste sin control/evidencia suficientes.

## Fuentes con revisión y límites

| Fuente | Revisión/fecha | Qué se puede afirmar |
|---|---|---|
| Main, incluidos 02/03/04 | `d5e95bd81d3bbab2d54a39d1c07d1256954b5384` | Código de baseline y paquete Codex modular entregados; la aceptación completa y mediciones pendientes no se infieren del merge. |
| 05 `origin/propose/opsx-task-context` | `4337e3658ca656b89cd15725a715db364a73a969`, implementación `325aebc` | Ya hay código en `eval/taskcontext/`, catálogo/módulos y pruebas en su rama; no es solamente texto `e4ec0f5`. No está incluido automáticamente en main. |
| 06 `origin/propose/opsx-verifiable-evidence` | `3c7d90808ff89fcfcad876cde67daf77531e3e2b` | Archivado con alcance parcial: contrato textual, extracto entregable, `eval/evidence/`, fixtures y huella de estado. No es el viejo borrador `f3f005e` ni solo normas `f62cf00`. |
| 07 `origin/propose/opsx-selective-delegation` | `4922498f8fa6ed5d77a82bda32001535977898fe` | Archivado parcial, con nueve tareas entregadas y dieciséis en deuda según su registro. Código de regla/módulo en su rama, pero piloto y A0 no entregados. |
| 08 `origin/opsx-native-controls` | `f3ba5d33973d64df5e9941e3c101e723b5b162ee` | Propuesta documental consultable. El SHA `9ced756` corresponde a `opsx-shadow-review` (10), no a controles 08. |
| [Informe ejecutivo](https://active-stack-executive-site.vercel.app/#evidencia) | Descarga local del 13/11: `/tmp/active-stack-executive-2026-11-13.html`; evidencia interna fechada 4/11, base `97d7f6c541a9fc7ce679eafd4410301cf3558f26` | Auditoría estática histórica y recomendaciones. Declara no haber benchmark de sesiones ni medición de tokens. |
| Borrador 09 previo a actualizar main | `89f1fa5de22304cc3564c50cf34126cf7e1dd651` | Dos pilotos y decisiones D1–D5 ya descritos; no hay aceptación registrada de esos valores por Matías/equipo en este análisis. |

El CLI `openspec status --change opsx-model-cache-pilots --json` confirmó el change y sus cuatro artefactos en el worktree 09 antes de escribir este informe. «Planning complete» no significa piloto ejecutado. El integrador actualiza main y mantiene proposal/design/specs/tasks; esta auditoría solo escribe el presente informe.

## Evidencia de la página: alcance real de los datos

El [informe ejecutivo](https://active-stack-executive-site.vercel.app/#evidencia) contiene evidencia JSON interna fechada 4/11, basada en `97d7f6c541a9fc7ce679eafd4410301cf3558f26`. El botón descarga ese contenido como `active-stack-evidencia-2026-11-04.json`; no ofrece resultados de corridas de 09. Los enlaces Git apuntan a observaciones históricas de código y las referencias técnicas a documentación, no a un dataset de benchmark actual.

El análisis compara tamaño de texto: 20.065 bytes en Claude y 18.010 en Codex, con doce observaciones y una prueba CLI temporal sin modelo. No midió tokens, sesiones ni ahorro. La comparación editorial de modelos y caché tampoco demuestra mejora por tarea aceptada: requiere contar workers, reintentos y calidad. Reutilizar prefijos no reduce contexto lógico ni ordenar un archivo garantiza hits.

Las cifras sirven para reconstruir la observación de esa revisión; no describen automáticamente el paquete modular actual ni sus instrucciones cargadas por rol. La descarga local del 13/11 no actualiza sus datos ni permite adoptar una variante. Archivo inspeccionado: `/tmp/active-stack-executive-2026-11-13.html`, SHA-256 `3725718dd2c68aed4b8b701b5c0b2d6eb7ba3b6db0f42957288a6a277ac6054c`.

## 02: arnés disponible frente a baseline estable

En main, `eval/baseline/README.md` describe bibliotecas externas al binario, sin cambiar `go.mod`: `manifest`, `profile`, `fixture`, `run`, `capture`, `usage`, `record`, `summary`, `timing`, `trace` y `review`.

- `manifest.CheckCell` y el plan distinguen ejecutable/incompatible; perfiles registran permisos e interfaz efectiva.
- `run.NewExecutor`, `Executor.Execute` y `ExecutePlan` admiten `Runner`, checkout aislado, reintentos y estados por intento; no equivalen a un cliente real ya conducido por 09.
- `capture` parsea fuentes Claude/Codex y atribución; `usage` separa categorías de entrada/caché y reconciliación padre/hijos.
- `record` genera JSONL con huellas y redacciones; `summary.FromRun`, `Compute` y `Markdown` calculan costo/consumo total por aceptada y cobertura. Cero aceptadas o cobertura insuficiente no producen un cociente válido.
- `timing` separa reloj monotónico y esperas delimitadas; `trace` distingue señal ausente de cero.

Esto es una base reutilizable, **no una baseline estable demostrada**. Tasks 02 mantiene pendientes 3.2 (calibración real por cliente/interfaz), 3.3 (aprobación de demora conocida), 3.4 (corridas de referencia) y 4.1 (comparación modular). El README explicita muestras derivadas de documentación y ningún cliente real ejecutado para esos adaptadores. El resumen tampoco entrega por sí solo contraste emparejado, dispersión o regla de adopción de 09; su implementación tiene que comprobarse separadamente. No se ejecutaron tests de baseline en esta auditoría ni se atribuyen al main actual resultados históricos de otra revisión.

## 07: control de 09 aún no fijado

Fuente: `4922498:openspec/changes/archive/2026-11-13-opsx-selective-delegation/`.

07 implementó regla/módulo compartidos en `internal/harness/config/assets/opsx/{delegation-rule.md,delegation-rule-disabled.md,delegation-module.md}`, con cambios de composición y tests en **su rama**. El archivo de deuda declara que el piloto 5.1–5.6 (D07-05), decisión 6.1–6.3 y control A0 para 09 (D07-06) no están entregados. No se encontró informe de adopción ni datos de ese piloto que permitan escoger automáticamente B como control de 09. «Directo por defecto» en la spec sincronizada es política propuesta; archivar no la adopta.

Design D9 define brazos A directo, B selectivo y R referencia histórica y especifica qué entregar a 09 según decisión: adoptar → B en revisión adoptada; mantener con ajustes → política/revisión identificada; repetir → **09 no hace prueba de adopción y puede calibrar A y B por separado**, declarando brazo y estratos directo/delegado. No hay tamaño ni umbral confirmados por esos ejemplos: tres corridas y 10 % son propuesta de calibración de 07, no criterio aceptado para 09.

Preparar calibración por brazo evita esperar para toda tarea de 09; no autoriza inventar el A0 final ni afirmar ahorro/adopción. Antes de una corrida real deben estar fijados permisos, instrumentos, límites y decisiones propias aplicables.

## Layout modular que debe fijar el manifiesto de 09

| Capa | Entrega en main `d5e95bd` | Qué congelar para comparar |
|---|---|---|
| Codex Machine, 04 | `ComposeCodexPackage`, SKILL explícita y policy `allow_implicit_invocation: false`; bloque universal; `artifacts`; cuatro módulos de contexto; `tdd`, `delegation`, `engram-opsx` según toggles. `model-assignments` omitido hasta 09. | Revisión, toggles, bytes/hash del paquete **compuesto**, tabla, módulos y orden de carga por acción; no solo tamaño del asset. Project Codex conserva estrategia legacy y se registra aparte. |
| Claude, 03 | `ComposeLayers`, skill, universal y módulos por su propia composición/versionado. | Revisión, versión efectiva, entrega nativa o legacy, módulos y encargo de especialista; no asumir que copia el layout Codex. |
| 05 actual | Nuevos módulos/catalogación en `assets/modules/`, `internal/catalog/taskcontext.yaml` y código `eval/taskcontext/` en `4337e36`. | Si se consume esa revisión, registrar cambio respecto de los cuatro textos fijados de `e4ec0f5` ya incluidos por 04; no mezclar antes/después entre brazos. |
| 06 actual | Contrato/extracto y validador en rama `3c7d908`, no instalados por main 04, que conserva requisitos parciales inline. | Revisión y condición real de carga/corroboración; no llamar completo al bloque parcial de 04. |
| 07 actual | Sustituye el módulo nativo Codex por assets `opsx` compartidos y actualiza composición en su rama. | Elegir explícitamente layout de main o candidato 07 fijado. No decir que main contiene el layout de 07 ni alternarlo durante un contraste. |

Pilotar modelos con cambio simultáneo de layout/delegación impide atribuir el efecto al modelo. Pilotar caché exige conservar modelo, esfuerzo, roles, herramientas y contenido semántico; el reordenamiento permitido y el contenido que ensambla el host deben delimitarse. Se registran configuración solicitada y efectiva, herencia y fallback; un alias o texto del template no prueba qué modelo ejecutó cada rol.

## 05/06: ya hay implementación, con límites de aceptación

05 tiene `eval/taskcontext/{capture,clientsrc,redact,summary,fixtures,...}` y catálogo de acciones/módulos. Su tarea 8.1 aún no demuestra fuentes del cliente en todas las celdas; CLI Windows Codex tiene sonda, IDE/Linux conservan límites. Las corridas reales de algunos módulos en Claude Windows y las pruebas de Go no se extrapolan a Codex/Linux/IDE. La matriz separa `sonda`, `probada`, truncada y no disponible. Lo que el agente declara no se vuelve independiente por guardar un archivo.

06 ya entrega `contrato-recibo-de-ejecucion.md`, `entrega-evidence-contract.md` (extracto literal y Review step), `eval/evidence.EvaluateRequirement` y `StateFingerprint`, con fixtures de recibo falso coherente, mismo HEAD/otro diff, LF/CRLF, salida redactada y reejecución. `StateFingerprint` usa listado Git cached+others no ignorados, bytes del disco y hash ordenado de hashes/rutas; rechaza entradas no regulares. Es código disponible en su rama, no un algoritmo que 09 deba inventar. La definición/confirmación, cobertura de plataformas y carga de assets conservan deuda documentada.

07 registra acuse a 06 sobre `receipts`, directorio ajeno y reejecución solo con captura propia independiente. Siguen pendientes suficiencia/evaluador, ubicación del módulo `evidence`, acuerdos de 05 sobre datos y cobertura por interfaz. La existencia del validador no demuestra que un dato de corroboración proporcionado por el mismo agente sea independiente.

08 se exige **solo cuando una corrida requiere un control preventivo obligatorio** que debe demostrarse. No bloquea preparar manifiestos/fixtures sintéticos. Si el control requerido falta, se detiene esa corrida o se usa protección equivalente comprobada; un hook escrito o un permiso declarado no basta. El juez en sombra de 10 no es requisito para estos pilotos nativos.

## Pendientes externos frente a decisiones de Matías

| Pendiente externo | Proveedor/disparador | Criterio verificable para 09 |
|---|---|---|
| Baseline/protocolo efectivo y fuentes por interfaz | Giuliano 02: revisión y calibraciones 3.2/3.3/3.4 | Reproducir fixture/rúbrica/permisos y observar señales/esperas; baseline insuficiente implica calibración identificada, no ahorro confirmado. |
| Control A0 y política | Neyen 07: D07-05/D07-06, informe de decisión | Revisión, brazo y layout identificados, resultados y criterios de decisión. Sin resultado, calibrar brazos por separado y diferir adopción. |
| Captura/contrato/datos demostrados | Emiliano 05 y Gianfranco 06: código/revisiones, disposición y pruebas por celda | Corroboración de ejecución y árbol, integridad y suficiencia separadas; solo copia redactada con referencia/retención acordada; no inferir señal ausente. |
| Capacidad de roles por cliente/superficie | Responsables de la entrega correspondiente y prueba efectiva del piloto | Modelo/esfuerzo y reglas efectivos por rol; no extrapolar del merge o de otra interfaz. El caso no evaluable se identifica y no cuenta como mejora. |
| Control preventivo, si el caso lo exige | 08/equipo: cobertura o equivalencia comprobada | Archivo testigo/comando permitido/denegado en la celda y permisos efectivos registrados. |

Estas son dependencias identificadas, **no deuda ya aceptada para 09** por herencia de 04 ni instrucciones para implementarlas aquí. D1–D5 de Matías/equipo permanecen decisiones propias: cohortes/fixtures/versiones; rol y modelos/esfuerzo; estrategia/ventana/horizonte de caché; métrica/cobertura; muestra, presupuesto, parada, calidad, umbral e inconclusión. No hay valores aceptados que puedan inventarse desde el informe ejecutivo o los ejemplos de 07.

## Preparación segura que puede avanzar

Preparar un manifiesto candidato que identifique piloto A o B, brazo y estado calibración/adopción; fijar revisiones/layout, controles, datos y casos sintéticos. A separa cambio de modelo y esfuerzo por rol manteniendo prefijo; B separa prefijo/caché manteniendo modelos/esfuerzo/roles. Ensayar cálculo con fallos, reintentos, preparación imputable, cobertura parcial, cero aceptadas y conciliación padre/hijos antes de usar resultados reales. Señal de caché prevista y observada se conservan separadas: sesión nueva no demuestra frío y un miss observado no se elimina para mejorar el resultado.

Muestra y umbral se fijan después de calibración y antes del contraste de adopción. Dinero API, cuota, tiempo, calidad y contexto lógico se informan por separado. Si falta referencia estable de 02 o resultado de 07, el informe debe decir calibración por brazo; no adoptar una política ni cambiar defaults del instalador. Ninguna parte de este análisis ejecutó esas corridas o aprobó sus presupuestos.
