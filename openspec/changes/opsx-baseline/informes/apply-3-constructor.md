# apply-3 · constructor · corrección de R1, N_modeTrim y N_omOrder (verify-2-breaker)

Change `opsx-baseline`. Rama `propose/opsx-baseline`, base `43eeb7f` más los cambios sin commitear de apply-1 y apply-2, que no se revirtieron. No hubo commit, push ni build. Clasificación LOW. Fecha 2026-11-11, de 11:07:44 a 11:11 (-03). Intento 3 de 3.

Alcance: solo R1, N_modeTrim y N_omOrder. No toqué S1–S6, N_reasonTrim ni N_seenFree.

L-NNN: `knowledge-base/` no existe, así que no se leyó ninguna lección.

## Revisión entregada (git hash-object)

| Archivo | Hash | Estado |
|---|---|---|
| `eval/baseline/manifest/manifest.go` | 4234ffb3 | modificado (antes 6a7d92ac) |
| `eval/baseline/manifest/apply3_test.go` | 215aae09 | nuevo |
| `eval/baseline/review/apply3_test.go` | 39702d79 | nuevo |
| `eval/baseline/review/review.go` | 84621364 | sin cambio |
| resto de `eval/` (README a8b1dfb2, manifest/apply2_test 3f264d3a, load_test d8392c39, manifest_test aefebcc5, plan_test 765b08ac, review/apply2_test 03e88c5a, review_test fe24eca9) | — | sin cambio |
| `openspec/changes/opsx-baseline/design.md` | 8b4f75b3 | una oración agregada en Datos sensibles |
| `openspec/changes/opsx-baseline/tasks.md` | 7163fb94 | evidencia de apply-3 bajo 1.1, 1.4 y suite completa |

## Tabla: hallazgo → test RED → corrección → mutante

| Hallazgo | Test (RED observado) | Corrección | Mutante que lo mata |
|---|---|---|---|
| R1: `OTEL_LOG_LEVEL` daba un falso opt-in | `TestCheckCell_OTELLogLevelIsNotContentOptIn`. RED a las 11:08:03 en 3 subcasos (`OTEL_LOG_LEVEL`, `otel_log_level`, ` Otel_Log_Level `), todos con `incompatible ... want "ejecutable"`, y en la aserción de motivos `want solo OTEL_LOG_USER_PROMPTS`. `OTEL_LOG_USER_PROMPTS=1` y `OTEL_LOG_LEVELS=1` siguen dando incompatible. | `isContentOptIn` devuelve false si el nombre normalizado (minúsculas y sin espacios en los extremos) es exactamente `otel_log_level`. Un comentario documenta la excepción. Una línea en `design.md` aclara que `OTEL_LOG_LEVEL` no cuenta. | M_R1 (`if false && n == "otel_log_level"`), a las 11:08:36: falla en 4 aserciones |
| N_modeTrim | `TestCheckCell_BlankPermissionModeIsEmpty`, con modo `" "` y `"\t "` en el declarado y en el observado. Nació verde, porque el código ya aplicaba TrimSpace. | Ninguna | N_modeTrim (se quita `TrimSpace`), a las 11:08:39: falla con `modo " ": CheckCell = {Status:ejecutable ...}` |
| N_omOrder | `TestMatch_OmissionsFollowKeyOrder`: clave D3,D1,D2,D0 y un hallazgo sobre D1. Espera las omisiones `[D3 D2 D0]`. Nació verde, porque el código ya recorre `key`. | Ninguna | N_omOrder (el bucle de omisiones recorre `byID`), a las 11:08:51: falla con `omissions = [D0 D2 D3] (3), want [D3 D2 D0]`. Un primer intento a las 11:08:42 alteró también el bucle de validación y no compiló, así que no cuenta. |

Todos los mutantes se revirtieron desde el respaldo guardado en el scratchpad (`.../scratchpad/apply3/`). Comprobé con `git diff --no-index --quiet` que cada archivo quedó idéntico. El hash final de `review.go` coincide con el de apply-2.

## EVIDENCIA

- Safety net a las 11:07:44: `go test -count=1 ./eval/baseline/...` dio ok en ambos paquetes.
- RED a las 11:08:03 y GREEN a las 11:08:21, con `go test -count=1 -run ... -v ./eval/baseline/...` y luego con `go test -count=1 ./eval/baseline/...`. Los dos paquetes dieron ok. Después de los mutantes, a las 11:08:55, volvió a dar ok.
- Suite completa: `go test -count=1 ./...` corrió de 11:09:06 a 11:10:57. Dio 29 paquetes `ok` y ninguna línea fuera de `ok`. Resultado registrado en `tasks.md`.
- `gofmt -l eval` no devolvió nada. `go vet ./eval/...` dio ok. `openspec validate opsx-baseline --strict` devolvió `Change 'opsx-baseline' is valid`, antes y después de editar `tasks.md`. `go.mod` y `go.sum` no tienen diff.
- No dejé procesos en segundo plano.

## JUICIO

- Autor: constructor. Confianza alta en que se cerró R1 como pidió el brief: la excepción es exacta, `OTEL_LOG_LEVELS` y cualquier otro `OTEL_LOG_*` siguen fallando cerrados, y el mutante prueba que la exclusión es la que hace pasar el test.
- N_modeTrim y N_omOrder no ocultaban defectos de comportamiento: eran huecos de prueba, y ya quedaron fijados.

## SUPOSICIÓN

1. Solo excluí el nombre exacto `OTEL_LOG_LEVEL`. No excluí otras variables de nivel o de exportador que también empiezan con `OTEL_LOG_` (por ejemplo, un hipotético `OTEL_LOG_LEVEL_SDK`). Esas siguen contando como opt-in, según la regla de fallo cerrado del brief.
2. La clave de Codex `otel.log_level` no empieza con `otel_log_` ni figura en la tabla de Codex, así que ya no se marcaba. No se tocó.

## Limitaciones

- Datos sintéticos. No verifiqué contra la configuración real del SDK de OpenTelemetry, ni de Claude Code ni de Codex.
- No ejecuté ningún build, por la regla del proyecto.

## Fuentes leídas

`informes/verify-2-breaker.md` (completo), `informes/apply-2-constructor.md` (completo), `eval/baseline/manifest/manifest.go` (completo), `manifest/apply2_test.go`, `review/review.go` (líneas 95–184), `review/apply2_test.go`, `baseCell` en `manifest_test.go`, la línea 128 de `design.md` y las secciones 1.1, 1.4 y de suite de `tasks.md`.

## Contaminación

El encargo parte del informe del breaker y define los mutantes N_modeTrim y N_omOrder, que reproduje tal como están descritos. Tomé del informe apply-2 la afirmación de que las omisiones van en el orden de la clave y la verifiqué con un test propio. Mis veredictos se apoyan en mis propias ejecuciones.
