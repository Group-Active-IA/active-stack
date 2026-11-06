## 1. Catálogo

- [x] 1.1 Agregar la entrada `web-scraper` en `internal/catalog/harnesses.yaml` (sección "Skills propias"): `type: skill`, `source: {repo: Group-Active-IA/web-scraper, ref: latest, method: clone}`, `description`/`long_description` bilingües, `install_modes: [full]`. Verificar: la clave `id: web-scraper` existe en el YAML.
- [x] 1.2 Agregar la entrada `discovery-research` en la misma sección, con `depends_on: [web-scraper]`, `source: {repo: Group-Active-IA/discovery-research, ref: latest, method: clone}`, `description`/`long_description` bilingües, `install_modes: [full]`. Verificar: la clave `id: discovery-research` existe y `depends_on` lista `web-scraper`.
- [x] 1.3 Actualizar la `description` bilingüe (es/en) de la entrada existente `active-orchestrator` para mencionar la fase opcional `discovery-research` antes de `kb-creator`. Verificar: el texto es/en menciona ambas fases en orden.

## 2. Validación

- [x] 2.1 Correr `go build ./...` y confirmar que compila sin errores. Compiló limpio.
- [x] 2.2 Correr los tests de catálogo existentes (`go test ./internal/catalog/...`). Dos tests hardcodeaban el conteo de harnesses "foundation-global" en 13 (`TestForMode_CustomReturnsOnlyGlobal`, `TestForMode_ExcludesStarterOnlyHarnesses`) — bump legítimo a 15 (dos harnesses Full-only nuevas, no starter-only), actualizados con comentario explicando el bump. `go test ./...` completo (toda la suite, no solo catalog) pasa.
- [x] 2.3 Verificado vía la suite de tests existente (no hizo falta un test ad-hoc: `TestForMode_ExcludesStarterOnlyHarnesses` ya cubre `ForMode(full)` vs `ForMode(lite)` de forma genérica sobre todo el catálogo, y ahora pasa con el conteo actualizado).

## 3. OpenSpec

- [x] 3.1 Marcar todas las tasks de este change como completas y correr `openspec validate --strict` antes de archivar.
