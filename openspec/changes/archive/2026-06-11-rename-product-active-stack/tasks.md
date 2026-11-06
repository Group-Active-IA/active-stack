# Tasks — rename-product-active-stack

> Rename mecánico. NO se fuerzan ciclos RED-GREEN: la red de seguridad es la suite existente.
> Regla de oro: distinguir SIEMPRE "este módulo" (`jr-stack`) de los repos-fuente externos
> (`JuanCruzRobledo/jr-orchestrator`, `JuanCruzRobledo/skill-registry`, starters), que se PRESERVAN.

## 1. Baseline (safety-net) — antes de tocar nada

- [x] 1.1 `go build ./...` verde. Capturar que compila. → **BUILD OK** (sin errores de compilación).
- [x] 1.2 `go test ./...` verde. **Anotar el conteo de tests/paquetes** (este número debe ser idéntico al final). → **BASELINE: 23 paquetes `ok`, 0 FAIL.**
- [x] 1.3 Confirmar working tree limpio == `origin/main` (`git status` sin cambios pendientes ajenos a esta tarea). → **LIMPIO**, HEAD `af83cd0` == upstream `af83cd0`.

## 2. Decisión bloqueante del marker (governance ALTO — filemerge)

- [x] 2.1 Resolver Open Question D3: ¿prefijo de marker `jr-stack:` → `active-stack:` (opción A, recomendada para repo nuevo) o se conserva como identificador técnico interno (opción B)? Confirmar con el operador antes de la capa 8. Registrar la decisión. → **DECISIÓN: OPCIÓN A** — renombrar prefijo de marker a `active-stack:` de forma atómica en `section.go` + 8 `sdd-orchestrator.md` + 2 golden. Confirmado por el operador (2026-06-10). Repo nuevo, sin instalaciones previas en producción → sin riesgo de migración.

## 3. Capa 1 — Module path + imports (`go build` después)

- [x] 3.1 `go.mod`: `module github.com/JuanCruzRobledo/jr-stack` → `github.com/Group-Active-IA/active-stack`. → **HECHO**.
- [x] 3.2 Reemplazar en TODOS los `.go` el import path `github.com/JuanCruzRobledo/jr-stack` → `github.com/Group-Active-IA/active-stack`. **Anclar el reemplazo al token `jr-stack`**; NO tocar otros `JuanCruzRobledo/...`. → **HECHO** (170 archivos actualizados; `JuanCruzRobledo/jr-orchestrator` intacto).
- [x] 3.3 `go build ./...` verde. → **BUILD OK** post-capa-1.

## 4. Capa 2 — Directorio del entrypoint (`go build` después)

- [x] 4.1 Renombrar `cmd/jr-stack/` → `cmd/active-stack/` (incluye el subpaquete `cmd/jr-stack/headless/`). Usar rename que preserve historia de git si es posible (`git mv`). → **HECHO** (`git mv`).
- [x] 4.2 Ajustar cualquier import o referencia interna al path `cmd/jr-stack` (p. ej. en tests de wiring) → `cmd/active-stack`. → **HECHO** (imports + comentarios actualizados).
- [x] 4.3 `go build ./...` verde. → **BUILD OK** post-capa-2.

## 5. Capa 3 — Build scripts

- [x] 5.1 `build.bat`: nombre de binario `jr-stack` → `active-stack` y path del entrypoint `cmd/jr-stack` → `cmd/active-stack`. → **HECHO**.
- [x] 5.2 `build.sh`: idem 5.1. → **HECHO**.
- [x] 5.3 Ejecutar ambos scripts (o el `go build -o` equivalente) y confirmar que producen el binario `active-stack`. → **HECHO** (`go build -o active-stack_linux_amd64 ./cmd/active-stack` OK).

## 6. Capa 4 — Marca en prosa / docs / banner

- [x] 6.1 Reemplazar "JR Stack" → "Active Stack" en `README.md`, `CLAUDE.md`, `AGENTS.md`, `ARCHITECTURE.md`, `CHANGES.md`. → **HECHO**.
- [x] 6.2 Banner: `docs/img/banner.svg` → "Active Stack". → **HECHO** (wordmark, terminal prompt, label).
- [x] 6.3 En `CHANGES.md` y `docs/archive/CHANGES-*.md`: renombrar SOLO menciones a este producto; **preservar** menciones a `jr-orchestrator` / `skill-registry` (son repos externos). → **HECHO**. `jr-orchestrator`/`skill-registry` intactos.

## 7. Capa 5 — Comandos + assets (catálogo, slash-commands, orquestador)

- [x] 7.1 `internal/catalog/harnesses.yaml`: descripciones de marca y IDs de comando `jr:starter-add`/`jr-starter-add` → `active:starter-add`/`active-starter-add` y `jr-stack starter add` → `active-stack starter add`. **NO tocar** `source: { repo: JuanCruzRobledo/jr-orchestrator ... }` (línea ~62). → **HECHO**. Repos externos intactos.
- [x] 7.2 Renombrar dir de asset Claude `assets/commands/claude/jr/` → `assets/commands/claude/active/` y archivo opencode `assets/commands/opencode/jr-starter-add.md` → `assets/commands/opencode/active-starter-add.md` (preferir `git mv`). Actualizar contenido de marca dentro de esos archivos. → **HECHO** (`git mv` preserva historia).
- [x] 7.3 Ajustar `assets/assets.go`, `assets/commands_test.go`, `internal/harness/command/*` (installer.go, command_test.go) y `internal/install/command_integration_test.go` que referencian los paths/IDs viejos (`jr:starter-add`, `jr-starter-add`, `cmd/jr-stack`). → **HECHO**.
- [x] 7.4 Marca del producto en los `internal/harness/config/assets/<agente>/sdd-orchestrator.md` (claude, opencode, gemini, codex, cursor, antigravity, windsurf, generic): "JR Stack" → "Active Stack". → **HECHO** (también `governance.md`).

## 8. Capa 6 — Prefijo de marker (según decisión 2.1; review ALTO)

- [x] 8.1 SI decisión = opción A: actualizar el prefijo de marker `jr-stack:` → `active-stack:` de forma atómica en `internal/filemerge/section.go` Y en los 8 `sdd-orchestrator.md` bundleados. → **HECHO** (atómico). También actualizado en todos los tests/archivos de filemerge, uninstall, harness/config, verify, e2e/lib.sh.
- [x] 8.2 Verificar que los tests de `internal/filemerge/...` y `internal/uninstall/...` (marker_purge, marker_step) reflejan el prefijo decidido y pasan. → **VERDE** (`ok internal/filemerge`, `ok internal/uninstall`).

## 9. Golden files

- [x] 9.1 Regenerar (preferido, si existe flag `-update`) o editar a mano `internal/harness/config/testdata/compose_all_toggles_claude.golden` y `..._opencode.golden`: marca "JR Stack" → "Active Stack" + prefijo de marker según 8.1. → **HECHO** (editado a mano; marcadores + marca actualizados).
- [x] 9.2 `go test ./internal/harness/config/...` verde contra los golden regenerados. → **VERDE** (`ok internal/harness/config`, `ok internal/harness/config/permissions`).

## 10. Validación del catálogo (regla dura)

- [x] 10.1 Correr el test que ejercita `catalog.Load()` (`go test ./internal/catalog/...`) para validar que el `harnesses.yaml` editado sigue siendo válido. Un catálogo inválido rompe el release: NO avanzar si falla. → **VERDE** (`ok internal/catalog`).

## 11. Verificación final (red de seguridad + cero leftovers)

- [x] 11.1 `go build ./...` verde y `go test ./...` verde con el **mismo conteo** que el baseline (1.2). → **BUILD OK**. **23 paquetes `ok`, 0 FAIL** == baseline.
- [x] 11.2 Grep final case-insensitive `jr-stack|jrstack|jr_stack` sobre todo el repo → **cero leftovers legítimos**. Allowlist documentada de leftovers intencionales:
  - `ARCHITECTURE.md`: `JuanCruzRobledo/jr-stack` → mención al repo GitHub fuente original archivado como legacy (histórico, no es una dependencia activa).
  - `CHANGES.md`: `jr-stack-legacy` → ruta local al folder legacy en disco (`E:\…\jr-stack-legacy`), folder de porting, no el producto.
  - `assets/commands_test.go`: `TestCommandsFS_ClaudeVariant_ResolvesAndContainsJRStackInvocation`, `TestCommandsFS_OpenCodeVariant_ResolvesAndContainsJRStackInvocation` → nombres de funciones Go (identificadores camelCase `JRStack`), no prosa del producto.
  - `CLAUDE.md` / `AGENTS.md`: path `E:\ESCRITORIO\programar\2026\jr-stack` (puntero "repo viejo" al codebase legacy de porting) → **legacy/externo, preservado** (mismo criterio que D2). Nota: el apply lo había renombrado por error a `...\Framework\active-stack` (auto-referencia sin sentido); el coordinador lo restauró a su valor original.
  - `internal/catalog/harnesses.yaml`, `internal/install/best_effort_test.go`, `docs/archive/`: `JuanCruzRobledo/jr-orchestrator` / `JuanCruzRobledo/skill-registry` → repos-fuente externos del instalador (D2), preservados intactos (ver 11.3).
- [x] 11.3 Confirmar que `git grep "JuanCruzRobledo/jr-orchestrator"` y `git grep "JuanCruzRobledo/skill-registry"` siguen devolviendo sus ocurrencias INTACTAS. → **CONFIRMADO**: `jr-orchestrator` en `harnesses.yaml`, `best_effort_test.go`, `CHANGES.md`, `docs/archive/...`; `skill-registry` en `harnesses.yaml`, `docs/archive/...`.
