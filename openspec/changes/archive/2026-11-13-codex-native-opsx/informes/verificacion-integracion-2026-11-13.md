# Verificación de integración · Change 04

13 de noviembre de 2026. Alcance Machine aprobado por Matías. Código probado en macOS Intel con Go 1.27.3, desde un árbol con main 35a89f8 integrado y cambios locales del 04.

## Resultado

El adapter Codex Machine instala la entrega administrada. Project conserva la entrega anterior. Los tests usan perfiles temporales. Se implementaron migración, primer snapshot protegido persistente, actualización por el replay existente, verificación estructural y retiro de archivos propios.

| Comprobación | Resultado |
|---|---|
| Suite completa con una exclusión previa de plataforma | 1914 casos aprobados, 6 eventos de fallo, 2 omitidos; 36 paquetes aprobados y 2 con fallos |
| Comparación de fallos contra origin/main 35a89f8 | Los mismos fallos y mensajes se reproducen en ambos árboles |
| go vet ./... | Aprobado |
| Race de backup, config, filemerge, install, verify y uninstall | Seis paquetes aprobados |
| Descubrimiento real del paquete instalado | Aprobado con Codex CLI 0.167.0-alpha.3.2 en macOS Intel |
| Snapshot inicial tras dos actualizaciones | Prueba de aceptación dirigida aprobada tras ampliar el caso inicial |

## Comandos

```bash
go test -json ./... -skip '^TestWindowsBundleThemeContainsMinimalInstallerPages$'
go vet ./...
go test -race ./internal/backup ./internal/harness/config ./internal/filemerge ./internal/install ./internal/verify ./internal/uninstall -count=1
ACTIVE_STACK_CODEX_PROBE=1 go test ./internal/harness/config -run TestManagedSkillCodexDiscovery -count=1 -v
go test ./internal/harness/config -run '^TestManagedSkillPreservesFirstSnapshot$' -count=1
openspec validate codex-native-opsx --type change --strict
git diff --check
```

## Fallos anteriores al cambio

El test Windows del theme se excluyó expresamente porque usa separadores Windows y falla en macOS antes de editar código. Las dos omisiones de la pasada global corresponden a la sonda opt-in y a la aceptación de binario Windows.

Los otros fallos se comprobaron ejecutando los mismos casos en un worktree limpio de origin/main 35a89f8 y en el árbol del 04. Ambos dieron los mismos mensajes: el descargador esperaba un error HTTP 404 y recibió éxito, y los fixtures de revisión del 02 difieren en su huella congelada. La suite completa conserva resultado fallido.

Eventos fallidos registrados:
- `github.com/Group-Active-IA/active-stack/eval/baseline/fixture/TestCheckout_RepetitionsStartIdenticalAndRestore`
- `github.com/Group-Active-IA/active-stack/eval/baseline/fixture/TestCheckout_RepetitionsStartIdenticalAndRestore/revision`
- `github.com/Group-Active-IA/active-stack/eval/baseline/fixture/TestEvaluate_ReferenceAcceptedDefectiveRejected`
- `github.com/Group-Active-IA/active-stack/eval/baseline/fixture/TestEvaluate_ReferenceAcceptedDefectiveRejected/revision`
- `github.com/Group-Active-IA/active-stack/eval/baseline/fixture/TestEvaluate_ReviewDispositionsAndMechanism`
- `github.com/Group-Active-IA/active-stack/internal/harness/external/TestDownloadBinary_APIError`

## Evidencia del ciclo TDD

| Tareas | Test File | Layer | Safety Net | RED | GREEN | TRIANGULATE | REFACTOR |
|---|---|---|---|---|---|---|---|
| 2.1, 2.3 | codex_package_test.go | Unit | 298 casos | API y módulos ausentes | Composición aprobada | 32 combinaciones, YAML, límites y toggles inválidos | gofmt y vet |
| 3.1–3.3 | managed_skill_test.go | Integración FS | 285 casos del área | API ausente, lock, snapshot perdido y padres no registrados | Instalación y respaldo aprobados | Colisiones, actualización doble, errores, reintentos, permisos y symlinks | Suites y vet |
| 3.4 | managed_skill_lifecycle_test.go | Pipeline | Referencia install | Rollback borraba nota o destino externo | Datos ajenos preservados | Fallo tardío, reintento y sustitución de padre | Race y vet |
| 3.5 | update tests y managed lifecycle | Integración | Update dirigido verde | Cambio de recorrido cubierto por dispatcher nativo | Replay vigente y reaplicación nativa aprobados | Registro presente/ausente y módulos retirados | Vet |
| 3.6, 4.1 | managed_skill_test.go en verify/uninstall | Integración FS | 90 casos | Enum ausente y archivo recreado sobrescrito | Retiro y estructura aprobados | Policy, módulos, ownership, notas, CRLF y rollback | Race y vet |
| Preservación de bytes y permisos | remove_sections_test.go y writer_preserving_parent_test.go | Unit/FS | Filemerge previo verde | APIs ausentes y casos de markers | Helpers aprobados | LF/CRLF, Unicode, directorios privados e idempotencia | Suite de 70 casos y vet |

Los informes de composición, backup y lifecycle contienen la secuencia y comandos de cada tanda.

## Cobertura real y dependencias

La sonda instala mediante el adapter real, comprueba el YAML y consulta App Server para descubrir la skill. Los módulos permanecen como archivos de apoyo. No ejecuta una inferencia. Activación manual/implícita, herencia y continuidad deben comprobarse por versión e interfaz.

- 03 mantiene la revisión común del contrato de entrega y matriz.
- 05 debe entregar módulos de contexto y captura independiente demostrada.
- 06 debe entregar recibos corregidos con procedencia y estado exacto del código.
- Windows/Linux, CLI/IDE y comparación del 02 siguen pendientes. Una terminal integrada pertenece a CLI.

El PR se presenta como borrador. El traspaso adjunto indica lo utilizable una vez integrado y las condiciones que conserva.

## Registros locales de comprobación

Los logs permanecen fuera del repositorio. Las huellas identifican los archivos usados en esta revisión.
- `/tmp/opsx04-full-final.jsonl` · SHA-256 `3916b29decb283c46e4c4ae0a0717f153726041b92e5da7a9bc69cdd7620bcd1`
- `/tmp/opsx04-vet-final.log` · SHA-256 `255936402ef104470caff9c821fd5e8cb5b5af94117595b5ca9706abf7fa00bf`
- `/tmp/opsx04-race-final.log` · SHA-256 `b69a83953d4f2b2063a1ea019a383a6626c6aaa9ddcbced308120507a6d802ea`
- `/tmp/opsx04-codex-probe-final.log` · SHA-256 `ed4dcc764937c6ef7163a50beba44b4f2a53b7990e10728fe8ece5d2432173a0`
- `/tmp/opsx04-main-external.jsonl` · SHA-256 `d937e063765c47b962981c0ad9bec9546e3afb802e1589715a25c63f39d305f9`
- `/tmp/opsx04-ours-external.jsonl` · SHA-256 `7cb3f42274e0d241afc613f5579b394d7d3f4702690c6fba523b0fd8192970ca`
- `/tmp/opsx04-main-fixture.jsonl` · SHA-256 `8096dbf12ec5c9017d5a98b7362be94a6eef386457c15e67342b65ba121d3408`
- `/tmp/opsx04-ours-fixture.jsonl` · SHA-256 `62dcccbf1e1ebaa20058b824e5d2e4f5788f2c90340d29ab11028d1eeffecbf9`
