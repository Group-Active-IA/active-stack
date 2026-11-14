# Implementación 1.1, 3.3, 3.4, 4.1 a 4.4 · opsx-shadow-review (10)

14 de noviembre de 2026. Rama `propose/opsx-shadow-review`, worktree en Windows. Base: merge de `main`
en `8bfca74` (07, 08 y 09 integradas), commit `dbd3fa9`. No se envió nada a ningún proveedor ni se
corrió ningún juez.

## Qué se entregó

| Tarea | Entrega | Archivo |
|---|---|---|
| 3.3 | Registro de envío append-only (caso, huella, candidato, proveedor, fecha, huella por archivo). Sólo registra con examen previo limpio, condiciones del proveedor documentadas y huella coincidente | `eval/shadow/sendlog.go` |
| 3.4 (parcial) | Compuerta en código: proveedor, retención y entrenamiento se exigen por separado | `eval/shadow/entry.go` (`ProviderTerms`) |
| 4.1 | Verificación de que el diff no toca `cmd/` ni `internal/` | ver Evidencia |
| 4.2 | Almacén del custodio: fuera del repo, escritura única con huella, apertura sólo del custodio con el caso cerrado y registro de todo intento | `eval/shadow/store.go` |
| 4.3 | Prompt fijo `revision-v1` embebido con huella fijada por test; parser estricto del formato de D4 | `eval/shadow/prompt/revision-v1.md`, `eval/shadow/finding.go` |
| 4.4 (parcial) | `CheckEntry` con los criterios de D6 y todos los motivos de descarte | `eval/shadow/entry.go` |
| 1.1 (parcial) | SZZ de los seis `fix` que faltaban; los seis quedan fuera del estrato histórico | `muestra-candidatos.md` |

## Evidencia

EVIDENCIA · RED de aserción, 17:38:08, `go test .` en `eval/shadow` con stubs que compilan: 16 tests
fallan por aserción (`TestStore_*`, `TestNewStore_*`, `TestSendLog_*`, `TestParseFindings_*`,
`TestPrompt_*`, `TestCheckEntry_*`).

EVIDENCIA · GREEN, 17:40:31: `go test -count=1 .` en `eval/shadow` → `ok`. `go vet ./eval/...` y
`go test -count=1 ./eval/...` → todos los paquetes `ok`. `gofmt -l` sin salida.

EVIDENCIA · mutaciones (cada una revertida después): 11 de 11 mueren: custodio, caso cerrado,
huella al abrir, almacén dentro del repo, escritura única (`Mkdir` → `MkdirAll`), examen previo,
condiciones del proveedor, huella del paquete, entrenamiento en `CheckEntry`,
`DisallowUnknownFields` y línea inicial ≥ 1.

EVIDENCIA · 4.1: `git diff --name-only origin/main...HEAD -- cmd internal` → 0 líneas;
`git status --short -- cmd internal` → 0; `grep -rln "eval/shadow" cmd internal --include=*.go` → 0.

EVIDENCIA · 1.1: el script SZZ de `muestra-candidatos.md` sobre `4074694`, `06216ba`, `25ad5c2`,
`41d7cb6`, `d3ab8b3` y `f3a4e89`. Para cada uno, `git merge-base --is-ancestor` contra la historia de
primer padre de `main` muestra que fix y origen entran por el mismo merge (`b27b8e0`, `0005a6f`,
`ad74451`).

NO CORRIDO · `go test -race`: el entorno no tiene cgo (`-race requires cgo`).

## Juicio

JUICIO · El almacén hace cumplir la política y deja rastro, pero no da confidencialidad: un usuario
con acceso al directorio lo lee igual. La confidencialidad depende de que el directorio pertenezca al
custodio, como dice D3 («la huella prueba integridad, no confidencialidad»).

JUICIO · Los fixes de revisión dentro del mismo PR muestran defectos que la referencia sí atrapó.
Podrían servir como casos preparados para medir la señal de la referencia, pero eso lo decide el
grupo (TBD en `muestra-candidatos.md`).

## Suposiciones

SUPOSICIÓN · Los nombres de campo del formato D4 en JSON (`severidad`, `ruta`, `linea_desde`,
`linea_hasta`, `requisito`, `descripcion`) son propuesta de 10; D4 fija los campos, no su escritura.
