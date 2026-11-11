# apply-t2-1 · constructor · opsx-baseline (tarea 1.3 + DEUDA 6)

Rol: constructor. Clasificación LOW. Rama `propose/opsx-baseline`, base `5dcfc68`, cambios sin commitear. Sin commit, push ni build. El trabajo se cortó por límite de uso después del RED de `Decide` (13:34:25) y se retomó a las 14:09 a pedido del orquestador.

## Revisión entregada (`git hash-object`, 14:35)
- Modificados: `eval/baseline/README.md` 31650d71, `review/review.go` 233e3865, `review/apply2_test.go` 6134db9f.
- Nuevos en review: `review/t2_test.go` 6c6ee3d7.
- Nuevos en fixture: `fixture/fixture.go` 647e2a33, `fixture/load.go` 2b7d6999, `fixture/checkout.go` 596fbc5c, `fixture/evaluate.go` eacc0e90, `fixture/decide_test.go` 8d1a9058, `fixture/load_test.go` c5b635c1, `fixture/checkout_test.go` 0cf7ae59, `fixture/evaluate_test.go` 3f6cdec5.
- `eval/baseline/fixtures/` (49 archivos de datos). Selección de hashes:
  - task.json: exploracion 9e79e406, cambio-pequeno 24547ac7, cambio-multiarchivo b9752218, diagnostico 07b7c10d, revision 98d0f4bd;
  - seeded.patch bb3dccd2;
  - los demás hashes salen del listado de `git hash-object` de la sesión y se recalculan con `git ls-files -o eval/baseline/fixtures | xargs git hash-object`.
- Además: `openspec/changes/opsx-baseline/tasks.md` (1.3 marcada `[x]` con evidencia, más la sección de suite) y este informe.

## Fuentes
- **Artefactos del change:** proposal, design (completo), tasks y spec (completos); informes/LEEME.md y verify-1-auditor.md.
- **Código:** review.go y review/apply2_test.go completos. De manifest.go, solo por grep las constantes de estado.
- **Repositorio:** `git show` de `catalog.go` y `.gitattributes` en `1b4639b`.
- **Lecciones L-NNN:** no existe `knowledge-base/` en el repo, así que no hay ninguna.
- **Engram:** no lo usé.

## Contaminación
- Leí el informe del auditor (DEUDA y LECCIONES) por indicación del brief y apliqué sus dos lecciones: un test de propiedades desde RED y un negativo de miembros benignos al ampliar la lista de disposiciones.
- No leí informes del breaker de esta tanda; no existen.

## EVIDENCIA
1. **Baseline.** 11:25:42, `go test -count=1 ./eval/baseline/...`: ok. Al retomar, 14:09:16: ok en fixture, manifest y review.
2. **DEUDA 6.**
   - RED 11:26:00 por aserción: 4 subcasos de clave con archivo o mecanismo vacío o en blanco y 12 de disposiciones desconocidas (`rejectd`, `Rejected`, `REAL_DEFECT`, `real defect`, ` rejected` y `pending`, cada una emparejada y sin emparejar).
   - A las 11:26:14 falló el test viejo F7, que fijaba la semántica anterior; lo adapté.
   - GREEN 11:26:31.
3. **Tarea 1.3, RED y GREEN por comportamiento** (horas en tasks.md):
   - Decide: 13:34:25 → 13:34:58.
   - Load: 14:10:08 → 14:13:15.
   - Checkout: 14:13:46 → 14:14:22.
   - Evaluate: 14:17:03 → 14:20:07.
4. **Huellas.** Las calculé primero por shell (`sha256sum` de bytes crudos y de las líneas ordenadas del árbol) y `TreeHash` en Go coincidió en las cinco familias.
5. **Mutantes.** Aplicados con `scratchpad/mut.sh`, que respalda el archivo, aplica la sustitución, corre el test enfocado, restaura y verifica con `cmp`. Todos quedaron revertidos e idénticos.

| Comportamiento | Test (RED → GREEN) | Mutante → resultado |
|---|---|---|
| Un obligatorio incumplido da rechazo | `TestDecide_*` (13:34:25 → 13:34:58) | M_decide_reject (`> 1`) → FAIL de la propiedad |
| Un pendiente deja el veredicto pendiente | `TestDecide_*` y `TestEvaluate_ReviewerCriteriaPending…` | M_decide_pending → FAIL |
| Un no obligatorio no decide | `TestDecide_Examples` | M_decide_nonmand → FAIL |
| Huella del prompt | `TestLoad_RejectsInvalid/prompt_*` (14:10:08 → 14:13:15) | M_prompt_hash → FAIL en 2 subcasos |
| Huella del árbol embebido | `…/fuente_embebida_alterada` | M_embedded_tree → FAIL |
| Hay al menos un obligatorio | `…/sin_obligatorios` | M_no_mandatory → FAIL |
| La revisión es un commit completo | `…/revision_corta`, `…/revision_no_hex` | M_short_rev y ML_hex → FAIL |
| Chequeo conocido, id único, herramientas, límites, campos estrictos | `TestLoad_RejectsInvalid` | ML_check, ML_dup, ML_tools, ML_limits y ML_unknown → FAIL |
| Familia desconocida y clave obligatoria | ídem, con el texto esperado endurecido | ML_family y ML_key: sobrevivieron con el texto genérico; tras endurecerlo → FAIL |
| Sin criterios | ídem | ML_nocrit: **equivalente**. La guarda "sin criterios obligatorios" ya da el error. |
| Restauración entre repeticiones | `TestCheckout_RepetitionsStartIdenticalAndRestore` (14:13:46 → 14:14:22) | M_reset_noop → FAIL (`tras Reset = 051d…`, `Reset dejo un directorio agregado`) |
| Reinicio solo desde el fixture inmutable | `TestCheckout_PrepareRejectsSourceChangedAfterLoad` | M_prepare_nohash → FAIL |
| Prueba de regresión que falla sobre el original | `TestEvaluate_SmallChange…` (14:17:03 → 14:20:07) | M_regression_inv: sobrevivió; con la aserción nueva → FAIL |
| Falso positivo rechazado incumple R3 | `TestEvaluate_ReviewDispositionsAndMechanism` | M_fp_ignored: sobrevivió; con la aserción nueva → FAIL |
| Checkout intacto | `TestEvaluate_ExplorationAnswer…` | M_intact_always → FAIL |
| Un archivo borrado cuenta como tocado | `TestEvaluate_TouchedFiles…` | M_touched_nodel → FAIL |
| DEUDA 6: disposición desconocida | `TestMatch_UnknownDispositionIsError` (11:26:00 → 11:26:31) | M_d6_disp → FAIL; M_d6_eval por la vía de Evaluate → FAIL |
| DEUDA 6: archivo o mecanismo vacío | `TestMatch_KeyDefectWithBlank…` | M_d6_file y M_d6_mech → FAIL |

6. **Suite completa**, una vez, 14:31:35–14:34:55: 30 paquetes `ok` (antes 29; el nuevo es `fixture`).
7. **Chequeos estáticos:**
   - `gofmt -l eval`: vacío;
   - `go vet ./eval/...`: ok;
   - `go list -deps ./cmd/... ./internal/...`: 0 coincidencias de `/eval`;
   - go.mod y go.sum sin diff;
   - imports de `fixture`: solo stdlib y `eval/baseline/review`;
   - `openspec validate opsx-baseline --strict`: válido;
   - sin cambios en cmd/, internal/ ni assets/;
   - sin directorios `opsx-*` filtrados en %TEMP%.

## JUICIO (constructor)
- 1.3 cumple su texto. Hay cinco familias con prompt, rúbrica y chequeos. Cada familia acepta la referencia y rechaza la variante defectuosa con el criterio esperado. El reinicio y la restauración quedan verificados por huella, sin tocar el fixture ni el árbol del repo.
- El escenario "Rechazo por calidad" del spec queda cubierto por cambio-pequeno con la variante defectuosa: S1 se cumple y S3 no.
- Confianza alta en la lógica. Confianza media en la duración: el paquete `fixture` tarda unos 2–3 min en Windows, sobre todo por extraer unos 643 archivos con `git archive` en cada `Reset`.

## SUPOSICIÓN (para el grupo o el revisor de 02)
1. **D1 y fixtures sintéticos.** Solo exploración y revisión usan la revisión fija del repo. Cambio, multiarchivo y diagnóstico son mini-proyectos sintéticos autocontenidos, congelados por huella de árbol en vez de commit, para no depender del repo completo ni de la red. Cuando 1.3 se commitee, el commit pasa a ser también su revisión. Conviene confirmarlo con el grupo.
2. **R3 obligatorio en revisión.** Ningún falso positivo y ningún hallazgo sin decidir. Es más estricto que la rúbrica, que solo define los conteos. Los hallazgos nuevos no lo afectan. Pendiente de revisión por 10 (relacionado con la DEUDA 2).
3. **Rechazo dominante.** Un obligatorio incumplido da rechazo aunque haya pendientes de revisor. Sin obligatorios, `Decide` devuelve error.
4. **Disposiciones exactas.** `Rejected` o ` rejected` son error, no se normalizan. Es una elección conservadora: el error es explícito.
5. **Mecanismo en revisión.** Si el revisor fija el mecanismo en `Output.Review`, ese reemplaza al del agente. Si no, se usa el del agente, comparado sin distinguir mayúsculas.
6. **Errores y respuestas inválidas.** Un hallazgo inválido del agente (rango, id vacío o duplicado) incumple R1–R3 sin error. Una disposición o una decisión inválida del revisor sí es error.
7. **Prueba de regresión en un paquete externo.** Una prueba que no compila sobre el original cuenta como prueba que falla.
8. **Reproducción en diagnóstico.** Se acepta cualquier prueba que falle con `--- FAIL: <nombre>` y cuya salida contenga el fragmento citado; no se exige el nombre de la clave.
9. **Repo para fixtures copiados.** Si un fixture se copia fuera del repo, `git archive` usa el repo del directorio actual.
10. **Límites.** Quedan `(TBD · D3)`.

## Limitaciones y chequeos no ejecutados
- No hubo corridas con clientes reales: las respuestas y los diffs son la referencia y las variantes del propio fixture.
- Los criterios E3 y G4 necesitan juicio humano.
- Lo que se cubrió por mutante en lugar de un RED propio:
  - los subcasos de `Load` que fallaron por setup;
  - las aserciones agregadas después del GREEN: prueba nueva que pasa con el original, FP rechazado y hallazgo nuevo.
- `ML_nocrit` es un mutante equivalente.
- No probé Linux, ni con `core.autocrlf` distinto: dependo de `* text=auto eol=lf` del repo y de `core.autocrlf=false` en `git archive`.
- Incidencias de herramienta:
  - un `python -` vacío (alias de WindowsApps) quedó colgado. Maté solo ese proceso (PID 13900, creado 14:10:20); los otros procesos python son de un MCP ajeno y no los toqué;
  - el stub temporal de `Prepare` dejó 4 directorios vacíos `stub*` en %TEMP%; los borré.
- No moví ni borré nada fuera de `eval/baseline/` y `openspec/changes/opsx-baseline/{tasks.md,informes/}`. El movimiento de `codex-strict-compat` y el borrado de `revision-2026-11-11.md` no son míos. El LEEME.md ya estaba modificado y no lo toqué.

## DEUDA nueva (no bloquea)
- Duración de `fixture` (unos 3 min). Opción: cachear una copia prístina por fixture y que `Reset` copie desde ella, o marcar las familias git con `testing.Short`. Responsable: el primer change que corra la suite muchas veces (2.1).
