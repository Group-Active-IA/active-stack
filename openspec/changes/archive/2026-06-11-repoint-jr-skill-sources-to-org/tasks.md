## 1. Precondición — Gate de existencia en la org (operativo, FUERA de active-stack)

> El runbook completo de mirror-then-strip vive en design.md §Migration Plan. Estas tareas
> son la PRECONDICIÓN del repunte: ninguna referencia se repunta sin su gate verde.

- [x] 1.1 Confirmar con el operador los nombres destino en `Group-Active-IA` (Open Question 1): `jr-orchestrator` → `active-orchestrator`, `jr-skills` → `active-skills`. Confirmado por el coordinador.
- [x] 1.2 Verificar gate del repo `jr-orchestrator` destino: `git ls-remote https://github.com/Group-Active-IA/active-orchestrator` exit 0 (existe y clona). Gate verde (coordinador: ref_count=2, HEAD + refs/heads/main, SHA 3debfea).
- [x] 1.3 Verificar gate del repo `jr-skills` destino: `git ls-remote https://github.com/Group-Active-IA/active-skills` exit 0 (existe y clona). Gate verde (coordinador: ref_count=2, HEAD + refs/heads/main, SHA 782565f).
- [x] 1.4 Capturar baseline de la red de seguridad ANTES de tocar nada: `go build ./...` ok + `go test ./...` → 23 paquetes ok / 0 FAIL. ✓

## 2. Repunte del catálogo — `internal/catalog/harnesses.yaml`

> SOLO ejecutar cada repunte tras el gate verde correspondiente (tareas 1.2 / 1.3).

- [x] 2.1 Repuntar `source.repo` del harness `jr-orchestrator` (línea ~62): `JuanCruzRobledo/jr-orchestrator` → `Group-Active-IA/active-orchestrator`. `ref: latest`, `method: clone` intactos. ✓
- [x] 2.2 Repuntar los 14 `source.repo` de las skills curadas propias/fork (líneas ~140–245): `JuanCruzRobledo/jr-skills` → `Group-Active-IA/active-skills`, preservando `path: skills/<...>` de cada una. 14 líneas actualizadas. ✓
- [x] 2.3 Verificar por grep que NO queda ninguna referencia `JuanCruzRobledo/jr-orchestrator` ni `JuanCruzRobledo/jr-skills` en `harnesses.yaml` (incluidos comentarios). Grep limpio en el YAML + comentario de sección actualizado. ✓
- [x] 2.4 Confirmar que NO se tocó ningún otro `source.repo` (kb-creator, roadmap-generator, agent-instruction, skill-registry, find-skill, skill-creator, terceros quedan intactos). Verificado — solo los 2 repos con marca `jr` fueron modificados. ✓

## 3. Aserciones de test (D4)

- [x] 3.1 [OBLIGATORIO] Actualizar `internal/catalog/c30_curation_test.go`: `const wantRepo = "JuanCruzRobledo/jr-skills"` → `"Group-Active-IA/active-skills"`. También actualizado el comentario de la función. ✓
- [x] 3.2 [consistencia] Actualizar fixture en `internal/install/best_effort_test.go` (~línea 326): `Source{Repo:"JuanCruzRobledo/jr-orchestrator"}` → `Group-Active-IA/active-orchestrator`. ✓
- [x] 3.3 [consistencia] Actualizar referencias a `jr-skills` en `e2e/starter/starter_network_test.go` (comentario de función) y `e2e/starter/starter_e2e_helpers.go` (comentario de arquitectura). ✓

## 4. Validación y red de seguridad

- [x] 4.1 Validar el catálogo: `go test ./internal/catalog/...` verde — todos los tests PASS (57 tests). ✓
- [x] 4.2 Red de seguridad completa: `go build ./...` ok + `go test ./...` verde → 23 ok / 0 FAIL == baseline. ✓
- [x] 4.3 Confirmar que el diff toca SOLO `harnesses.yaml` + archivos de test de §3. No se introdujo ningún script de mirror/`gh`/migración. ✓

## 5. Cierre

- [x] 5.1 Resolución de Open Questions: OQ1 → `active-orchestrator` / `active-skills`; OQ2 → fixture actualizado; OQ3 → `ref: latest` mantenido (status quo). ✓
- [x] 5.2 NO commitear sin pedido explícito del operador (regla dura). Cambio listo para review. ✓
