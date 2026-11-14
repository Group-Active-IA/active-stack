# Verificación de 08 antes de archivar

13 de noviembre de 2026. Rama local `opsx-native-controls`. Commit sobre el que se verificó: `021fe0bc416cbfe35b3795ecdefe9245341e984b` (merge de `origin/main` `0005a6f45f45062939facb492e6919c8260fa445`, sin conflictos), más los cambios de esta ronda, que se commitean después y no pueden citar su propio hash (el hash final del commit de cierre está en el cuerpo del PR). Entorno: Windows 11, Git Bash, OpenSpec 2.0.0.

Salidas completas, con exit code, en `C:\Proyectos\informes-opsx\08-opsx-native-controls\` (archivos `v1-*.txt`).

**Qué se verificó y qué no.** Se verificó **la propuesta**: sus artefactos validan, las citas coinciden con `main` y no se tocó código. **No se corrió el piloto**: no hay matriz de cobertura medida, ni pruebas de permisos o hooks en Windows o Linux, ni reproducción del caso del 02/11, ni escrituras de configuración. No se corrió `go build`, `go vet` ni `go test` porque esta ronda no cambia código.

## 1. Comandos y resultados

| Comando | Resultado real | Salida |
|---|---|---|
| `git fetch origin '+refs/heads/*:refs/remotes/origin/*'` y `git ls-remote --heads origin main` | `main` = `0005a6f45f45062939facb492e6919c8260fa445` | `v1-fetch-merge-diff.txt` |
| `git merge origin/main` | integrado sin conflictos (merge, sin rebase) | `v1-fetch-merge-diff.txt` |
| `git diff --stat d5e95bd origin/main -- internal cmd CLAUDE.md AGENTS.md CHANGES.md .github` | 27 archivos, +3371 −1, todos en `internal/catalog/` y `internal/harness/config/`. Nada en `cmd`, `CLAUDE.md`, `AGENTS.md`, `CHANGES.md` ni `.github` | `v1-fetch-merge-diff.txt` |
| `openspec validate opsx-native-controls --type change --strict` (antes de archivar) | «Change 'opsx-native-controls' is valid», exit 0 | `v1-validate-change-strict.txt` |
| `openspec status --change opsx-native-controls` | 4/4 artefactos completos, exit 0 | `v1-validate-change-strict.txt` |
| `openspec archive opsx-native-controls --yes` | archivado como `2026-11-13-opsx-native-controls`; spec `native-controls-pilot` creada (+17 requisitos); avisó 38 tareas incompletas (esperado, son la deuda) y dos advertencias no bloqueantes de `proposal.md` | `v1-archive.txt` |
| `openspec validate native-controls-pilot --type spec --strict` | «Specification 'native-controls-pilot' is valid», exit 0 | `v1-validate-spec-strict.txt` |
| `diff` del delta archivado contra `openspec/specs/native-controls-pilot/spec.md` | solo difieren los encabezados (`# Spec Delta` por `# native-controls-pilot Specification`, `## ADDED Requirements` por `## Requirements`) y una línea en blanco; el contenido de los requisitos y escenarios es idéntico | `v1-diff-delta-vs-spec.txt` |
| `git diff --name-only origin/main` más archivos nuevos | solo `openspec/changes/archive/2026-11-13-opsx-native-controls/` y `openspec/specs/native-controls-pilot/`; ningún `.go`, ningún `filemerge`, `backup` ni `pipeline` | `v1-git-diff.txt` |

El delta archivado es byte a byte el que validó en strict antes de archivar. Para pasar strict se acortaron dos requisitos de la spec a menos de 500 caracteres (`Behavior when a control fails`, de 835 a menos de 500, y `Missing mandatory or optional control`, de 530 a menos de 500), con todos los escenarios intactos. El detalle de lo que cada cliente documenta ante un hook que falla sigue en design D6.

## 2. Diferencia con lo anunciado sobre `d5e95bd..0005a6f`

Se esperaba que solo cambiaran `internal/catalog` y `internal/harness/config/assets/modules`. El `git diff --stat` mostró además archivos de prueba y fixtures en `internal/harness/config/` (`inventory_internal_test.go`, `inventory_sources_internal_test.go`, `modules_internal_test.go`, `testdata/inventory_*.golden`), y `internal/catalog/catalog.go` tuvo 20 líneas cambiadas (una eliminada). Ninguno de esos archivos es uno de los citados por 08, y ningún archivo no-test fuera de `internal/catalog/` y `internal/harness/config/assets/modules/` cambió (`v1-citas.txt`).

## 3. Citas re-verificadas contra `0005a6f`

Se releyeron las líneas citadas con `sed`; ninguna se desplazó (`v1-citas.txt`):

| Cita | Resultado |
|---|---|
| `internal/filemerge/json_merge.go:10-15` (JSON malformado = `{}`), `:49-52` (comentarios) | coincide |
| `internal/filemerge/toml.go:26-28` (validación), `:29` (CRLF a LF) | coincide |
| `internal/filemerge/writer.go:54-58` (`chmod 0755`), `:78` | coincide |
| `internal/harness/config/permissions/install.go:120-126`, `:143`, `:185`, `:144-148` | coincide |
| `internal/harness/config/permissions/overlays.go:16-28`, `:79-81` | coincide |
| `internal/backup/restore.go:22-29`, `:31-37`, `:39-47`, `:64-66` | coincide |
| `internal/install/steps.go:223-230`, `:371-378` | coincide |
| `internal/uninstall/steps.go:362-384` | coincide |
| `cmd/active-stack/headless/executor.go:158-159` | coincide |
| `CLAUDE.md:13`, `CLAUDE.md:122-130`, `CHANGES.md:10` | coincide |
| `.github/workflows/release.yml:71-76` | coincide |
| `openspec/specs/codex-opsx-delivery/spec.md:64-78` | coincide |
| `eval/evidence/judge.go:246-249` (cita nueva: fuente independiente de 06) | coincide |

`managed_skill.go`, `layered_install.go` y `permissions/` no cambiaron entre `d5e95bd` y `0005a6f`. Base cambiada a `0005a6f` en `proposal.md`, `design.md` (Base y fuentes) y `tasks.md`; 05 y 06 se citan desde `main` (PR #10 y PR #12) y 07, como rama `4922498` por sección.

## 4. Ajuste sobre 06 (`d3ab8b3`)

Con el fix de Juan Cruz en `main` (`eval/evidence/judge.go:246-249`), una fuente de corroboración vacía o `agente` no cuenta como independiente, y el mismo arnés puede capturar y corroborar. 08 no define la independencia; lo dice en design D2 y en la tarea 6.1, y la tarea 1.1 pasa a hecha con el PR #12 mergeado.

## 5. Qué NO está verificado

Todo lo de `deuda-aceptada-08-2026-11-13.md`: el piloto, las pruebas en Windows y Linux, las escrituras de permisos y hooks, la reproducción del caso del 02/11, y las decisiones del grupo. La suite Go no se corrió.
