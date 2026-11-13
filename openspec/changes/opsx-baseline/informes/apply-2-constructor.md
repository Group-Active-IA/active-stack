# apply-2 · constructor · corrección de F1–F8 (verify-1-breaker)

Change `opsx-baseline`. Rama `propose/opsx-baseline`, base `43eeb7f`, cambios sin commitear. Sin commit, sin push, sin build. Clasificación LOW. Fecha 2026-11-11. Alcance: solo F1–F8; S1–S4 sin tocar.

## Revisión entregada (git hash-object)
- `eval/baseline/manifest/apply2_test.go` 3f264d3a · `load_test.go` d8392c39 · `manifest.go` 6a7d92ac · `manifest_test.go` aefebcc5 · `plan_test.go` 765b08ac
- `eval/baseline/review/apply2_test.go` 03e88c5a · `review.go` 84621364 · `review_test.go` fe24eca9
- `openspec/changes/opsx-baseline/tasks.md`: evidencia de apply-2 agregada bajo 1.1, 1.4 y en la sección de suite completa.

## Tabla fallo → test RED → corrección → mutante

| Fallo | Test que reproduce (RED observado) | Corrección | Mutante que lo mata |
|---|---|---|---|
| F1 orden de clave / no máximo | `TestMatch_MaximumMatchingIndependentOfKeyOrder`, `TestMatch_MaximumMatchingStillOneToOne` (10:53:30: clave A,B daba TP 1/omisión B; asignación distinta según orden) | `review.go`: defectos ordenados por ID; emparejamiento máximo con caminos de aumento (`maxMatching`). Desempate estable: hallazgos en orden de lista; cada uno toma primero el candidato libre de menor ID y solo si no hay libre desplaza por un camino de aumento. Un hallazgo con candidatos sin par queda como duplicado del primer candidato por ID. Las omisiones se listan en el orden de la clave. | MF1a sin aumento → falla `..._NeedsAugmentingPath`; MF1b sin orden por ID → falla `..._StillOneToOne`; MF1c Kuhn sin preferir libres → falla `..._StillOneToOne` |
| F2 dimensiones / permisos vacíos | `TestCheckCell_UnknownOrMissingDimensions` (10:56:50, 9 subcasos `ejecutable, want incompatible`) | `checkDimensions` en `CheckCell`: cliente ∉ {claude, codex}, interfaz ∉ {terminal, ide} (`desktop` con motivo "escritorio fuera de la primera comparación"), sistema ∉ {windows, linux}, modalidad ∉ {api, subscription} o vacíos dan **incompatible con motivo**. Modo de permisos declarado vacío también da incompatible. Se resuelve en `CheckCell` y no en `Validate` porque el spec trata como incompatible la celda que no cumple. | MF2a–e (uno por dimensión y permisos) → fallan |
| F3 familia de opt-ins | `TestCheckCell_OptInFamily` (10:56:50, 7 subcasos y orden de motivos) | `isContentOptIn`: nombre normalizado (minúsculas, sin espacios en los extremos) con prefijo `otel_log_`, o `log_user_prompt`, o `otel.log_user_prompt`. Los valores de apagado no cambian. Los motivos salen ordenados por nombre. | MF3a prefijo, MF3b mayúsculas, MF3c tabla de Codex, MF3d orden → fallan |
| F4 contenido extra | `TestLoad_RejectsTrailingContent` (10:56:50) | `Load` exige `io.EOF` tras el primer objeto; los espacios finales se aceptan. | MF4 → falla |
| F5 task_id / Plan sin validar | `TestManifestValidate_VariantAndTaskIDs` (task_id vacío o duplicado) y `TestManifestPlan_AllExecutableAndAllIncompatible` (10:55:37, `Plan() de un manifiesto vacio debe fallar`) | `Validate` rechaza task_id vacío o duplicado dentro de una cohorte (entre cohortes se permite). `Plan()` pasa a devolver `(Plan, error)` y valida antes de planificar. | MF5a, MF5b, MF5c → fallan |
| F6 variante | `TestManifestValidate_VariantAndTaskIDs` (sin variante / en blanco) | Campo `Manifest.Variant` (`json:"variant"`), obligatorio no vacío. Constantes `VariantCurrent="actual"`, `VariantModular="modular"`; se acepta cualquier otro identificador. | MF6 → falla |
| F7 test vacuo + sobrevivientes | `TestLoad_UnknownFieldOnOtherwiseValidManifest` (manifiesto válido + `api_key`, exige "unknown field"); `TestCheckCell_NegativeVersionComponent` (`3.-1.0`); subcaso "apagado con espacios" (`" off "`); `TestMatch_EmptyMechanismOnBothSidesDoesNotMatch` | Sin cambio de producción: aserciones nuevas. Nacieron verdes y su sensibilidad se acreditó por mutante. | M_unknown, M_neg, M_trim, M_emptymech → fallan |
| F8 terminal integrada | `TestPlan_IDEIntegratedTerminalReclassifiedToTerminal` y la fila de `TestCheckCell_IDEIntegratedTerminal` ahora espera ejecutable (10:56:50, ambos fallaban con incompatible) | Reclasificación según el spec: `CheckCell` devuelve ejecutable con `Interface = terminal`. `Plan.Executable` pasa a `[]Planned{CohortID, Interface (real), DeclaredInterface}` y preserva la declarada. | MF8 → falla |

GREEN: review 10:54:13, manifest 10:57:23. Cobertura: manifest 96.4 %, review 100 %. En total corrieron 22 mutantes (18 de manifest y 4 de review), todos detectados y revertidos con `git diff --no-index --quiet` idéntico al respaldo. MF1a y MF1b no compilaron en el primer intento y no cuentan; se repitieron en forma compilable y fallaron por aserción.

## EVIDENCIA
- Safety net 10:53:04: `go test -count=1 ./eval/baseline/...` ok en ambos paquetes.
- RED, GREEN y mutantes: comandos `go test -count=1 ./eval/baseline/<pkg>/` con las horas de la tabla; detalle en `tasks.md`.
- Final: `gofmt -l eval` vacío; `go vet ./eval/...` ok; `go test -count=1 ./...` 10:59:02–11:00:31 todos `ok`; `openspec validate opsx-baseline --strict` válido; go.mod/go.sum sin diff. `grep "active-stack/eval" cmd internal` vacío.
- L-NNN: sigue sin existir `knowledge-base/`.

## Tests existentes cambiados (y por qué)
- `review_test.go`: `TestMatch_FindingMatchingTakenDefectIsDuplicateNotSecondDefect` se renombró a `..._TakesFreeSecondDefect`. Ahora espera TP de D2 y 2 TP, porque codificaba la lectura que F1 invalida.
- `manifest_test.go`: la fila de terminal integrada espera ejecutable (F8).
- `plan_test.go`: adaptado a `Plan() (Plan, error)` y a `[]Planned` (helpers `mustPlan`, `plannedIDs`), con manifiestos que pasan `Validate`. El caso de manifiesto vacío ahora espera error.
- `load_test.go`: `validManifest()` incluye `Variant: "actual"`.

## SUPOSICIÓN
1. **F8 reemplaza el supuesto 1 de apply-1.** Una celda declarada `ide` cuya CLI corre en la terminal integrada se agrega a la cohorte terminal y conserva la interfaz declarada en `Planned.DeclaredInterface`. Agregar realmente por cohorte en el resumen es trabajo de 3.1.
2. **F1 reemplaza el supuesto 5.2 de apply-1** ("se ata al primero de la clave"). Los conteos ya no dependen del orden de la clave. Los IDs asignados dependen del orden de los hallazgos y de los IDs de defecto, por el desempate documentado.
3. El brief menciona la modalidad `suscripcion`. Conservé el valor existente `subscription` (`ModalitySubscription`) para no cambiar el contrato JSON en esta ronda. Si el grupo prefiere el valor en castellano, es un cambio de una constante.
4. Solo el modo vacío cuenta como "permisos declarados vacíos". Un modo con flags vacíos es válido.
5. En la familia de Codex solo se reconocen `log_user_prompt` y `otel.log_user_prompt`, según la lista del brief. Otras claves `otel.*` no se tratan como contenido. No verificado contra la configuración real de Codex.
6. S1–S4 sin cambios: un emparejado con disposición rejected sigue contando como TP; una disposición desconocida sigue pendiente; la clave sigue aceptando archivo o mecanismo vacío; las versiones con prefijo `v` o prerelease siguen incompatibles.

## Limitaciones
- Datos sintéticos; sin clientes reales.
- Siguen sin modelarse las plataformas distribuidas, declaradas y probadas (D2) y las "señales disponibles". Quedan anotadas como excepción en `tasks.md` 1.1.

## Fuentes leídas
`informes/verify-1-breaker.md` (completo), `eval/baseline/manifest/manifest.go`, `plan_test.go`, `tasks.md`. El resto ya lo conocía de apply-1.

## Contaminación
Este trabajo parte del informe del breaker, que era el encargo. Reproduje cada fallo con tests propios antes de corregir, y mis veredictos se apoyan en esas ejecuciones.

## JUICIO
Autor: constructor. Comparé cada F con su test RED y su mutante. Confianza alta en F1–F7. En F8 es alta respecto del texto del spec; la agregación efectiva por cohorte se completa en 3.1.
