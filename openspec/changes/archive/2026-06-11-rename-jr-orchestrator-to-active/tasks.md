## 1. Red de seguridad (baseline)

- [x] 1.1 Correr `go build ./...` + `go test ./...` y capturar el baseline: 23 paquetes OK / 0 FAIL. Si algo ya falla, STOP y reportar como falla pre-existente (no arreglar acá).

## 2. Catálogo (regla dura — gate `catalog.Load()`)

- [x] 2.1 En `internal/catalog/harnesses.yaml` (líneas ~58-59), renombrar `id: jr-orchestrator` → `id: active-orchestrator` y `name: jr-orchestrator` → `name: active-orchestrator`. (El `source.repo` ya apunta a `Group-Active-IA/active-orchestrator` — no tocar.)
- [x] 2.2 En `internal/catalog/catalog_test.go`: renombrar el test `TestForMode_JROrchestratorIsFullOnly` → `TestForMode_ActiveOrchestratorIsFullOnly` y actualizar TODAS las aserciones internas (líneas ~309-321) que comparan `h.ID == "jr-orchestrator"` → `"active-orchestrator"` y sus mensajes de error.
- [x] 2.3 Correr `go test ./internal/catalog/...` → debe pasar (valida `catalog.Load()` + el test renombrado).

## 3. Tests `.go` restantes

- [x] 3.1 Actualizar refs `jr-orchestrator` → `active-orchestrator` en `internal/install/best_effort_test.go`.
- [x] 3.2 Actualizar refs `jr-orchestrator` → `active-orchestrator` en `internal/verify/harness_checks_test.go`.
- [x] 3.3 Actualizar refs `jr-orchestrator` → `active-orchestrator` en `internal/verify/hook_test.go`.
- [x] 3.4 Actualizar refs `jr-orchestrator` → `active-orchestrator` en `internal/backup/restore_test.go`.
- [x] 3.5 Actualizar refs `jr-orchestrator` → `active-orchestrator` en `e2e/starter/starter_hermetic_test.go`.
- [x] 3.6 Correr `go test` de cada paquete tocado (`./internal/install/...`, `./internal/verify/...`, `./internal/backup/...`, `./e2e/starter/...`) → todos verdes.

## 4. Assets bundleados + goldens

- [x] 4.1 En `internal/harness/config/assets/claude/sdd-orchestrator.md` y `internal/harness/config/assets/opencode/sdd-orchestrator.md`: reemplazar la instrucción "invoke `jr-orchestrator`" y demás refs `jr-orchestrator` → `active-orchestrator`.
- [x] 4.2 Actualizar/regenerar los goldens `internal/harness/config/testdata/compose_all_toggles_claude.golden` y `compose_all_toggles_opencode.golden` para reflejar el nuevo identificador.
- [x] 4.3 Correr `go test ./internal/harness/config/...` → goldens y compose verdes.

## 5. Docs + e2e (sin impacto en build)

- [x] 5.1 Actualizar refs `jr-orchestrator` → `active-orchestrator` en `README.md` (4 refs).
- [x] 5.2 Actualizar refs `jr-orchestrator` → `active-orchestrator` en `ARCHITECTURE.md` (6 refs).
- [x] 5.3 Actualizar refs `jr-orchestrator` → `active-orchestrator` en `e2e/e2e_test.sh` (5 refs).

## 6. Verificación de footprint (sin marca `jr` residual en scope)

- [x] 6.1 Confirmar 0 ocurrencias de `jr-orchestrator` en código activo (excl `openspec/`, `docs/archive/`, `CHANGES.md`).
- [x] 6.2 Confirmar que NO se tocaron los out-of-scope: `jr-stack` (legacy allowlisteado), `Jrh`/`JrbApNNgaTdGDITg` (false positives), y `JRStackInvocation` salvo decisión explícita del operador (Open Question 2).

## 7. Red de seguridad final (baseline sagrado)

- [x] 7.1 Correr `go build ./...` + `go test ./...` → 23 paquetes OK / 0 FAIL, idéntico al baseline de 1.1.

## 8. Coordinación cross-repo (documentar — NO implementar desde active-stack)

- [x] 8.1 Verificar/flaggear el gate de consistencia (D6): el `id`/`name` del harness en active-stack DEBE matchear el `name:` del `SKILL.md` del repo `Group-Active-IA/active-orchestrator` (org). Si la org aún dice `jr-orchestrator`, dejar el rename flaggeado como bloqueante de release.
- [x] 8.2 Documentar runbook frente 2: editar el `SKILL.md` de la org (`name:` + namespace del slash-command `jr-orchestrator:` → `active-orchestrator:`) — trabajo del operador sobre la copia de la org. NO desde active-stack.
- [x] 8.3 FLAG frente 3: refs a `jr-orchestrator` en `~/.claude/CLAUDE.md` global del usuario — solo señalar al operador, NO auto-editar (config del usuario, governance: no pisar sin pedido).
