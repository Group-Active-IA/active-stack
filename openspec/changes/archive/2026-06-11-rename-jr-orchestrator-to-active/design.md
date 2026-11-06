## Context

`active-stack` es un instalador methodology-first (Go 1.26). El operador estableció el principio de marca **"todo se tiene que llamar active"**. Ya se cerraron dos renames: (1) producto `jr-stack` → `active-stack` (module path `github.com/Group-Active-IA/active-stack`), (2) repunte de `source.repo` de los dos repos `jr` → `Group-Active-IA/active-orchestrator` y `Group-Active-IA/active-skills`.

Sobrevive un último identificador con marca `jr` en código activo: el harness **`jr-orchestrator`** (`internal/catalog/harnesses.yaml`, líneas 58-59). Su `source.repo` YA apunta a `Group-Active-IA/active-orchestrator` (rename previo) — hay una **inconsistencia activa**: el harness se llama `jr-orchestrator` pero clona de un repo `active-orchestrator`.

Estado verificado del footprint (excl `openspec/`, `docs/archive/`, `CHANGES.md`):
- **~42 ocurrencias** del string `jr-orchestrator` en código activo, repartidas en: 1 yaml (`harnesses.yaml`), 6 tests `.go`, 4 `.md`, 2 goldens, 1 `.sh`.
- 1 identificador Go CamelCase: `TestForMode_JROrchestratorIsFullOnly` (`internal/catalog/catalog_test.go:302`).

La trampa: el ID `jr-orchestrator` no es un string interno cualquiera — **define el namespace del slash-command** de la skill (`/jr-orchestrator:init`, `:kb`, `:rules`, `:openspec`, `:devops`, `:find-skill`). Ese namespace está declarado en el `SKILL.md` del repo de la **org** (`Group-Active-IA/active-orchestrator`), NO en active-stack. Además, el `~/.claude/CLAUDE.md` GLOBAL del usuario referencia `jr-orchestrator`. Por eso el debranding total es **cross-repo** (3 frentes).

## Goals / Non-Goals

**Goals:**
- Renombrar el identificador del harness `jr-orchestrator` → `active-orchestrator` (`id` + `name`) en `harnesses.yaml`, validado por `catalog.Load()`.
- Propagar el rename a las ~42 refs string en código activo + el identificador Go del test.
- Mantener el baseline sagrado: `go build ./...` + `go test ./...` = 23 paquetes OK / 0 FAIL, idéntico antes y después.
- Dejar documentado el runbook cross-repo (frentes 2 y 3) y el gate de consistencia obligatorio, sin implementarlos desde este repo.

**Non-Goals:**
- NO cambiar comportamiento. Es un rename de identificador; la materialización/instalación/orquestación del harness es idéntica.
- NO crear delta specs (no cambia ningún requirement a nivel spec).
- NO editar el `SKILL.md` del repo de la org desde active-stack (trabajo del operador sobre esa copia — se documenta como runbook).
- NO editar el `~/.claude/CLAUDE.md` global (config del usuario; governance: no pisar sin pedido — solo FLAG).
- NO tocar leftovers legacy `jr-stack` (refs legítimas a lo VIEJO, ya allowlisteadas) ni false positives (`Jrh`, `JrbApNNgaTdGDITg`).
- NO build/release fuera de la red de seguridad. NO commitear (regla dura).

## Decisions

### D1 — Rename mecánico string-for-string `jr-orchestrator` → `active-orchestrator`
Reemplazo literal en las ~42 refs de código activo. Es un identificador kebab-case homogéneo; el reemplazo es seguro y sin ambigüedad.
- **Alternativa descartada**: alias/compatibilidad (mantener `jr-orchestrator` como sinónimo). Rechazada: contradice el principio "todo active" y deja marca `jr` viva; un instalador no necesita compatibilidad hacia atrás de un identificador interno.

### D2 — Renombrar el identificador Go del test
`TestForMode_JROrchestratorIsFullOnly` → `TestForMode_ActiveOrchestratorIsFullOnly`. Por consistencia con el rename del dominio. Las **aserciones internas** de ese test (líneas ~309-321) comparan `h.ID == "jr-orchestrator"` — deben actualizarse a `"active-orchestrator"` o el test falla (verifican que el harness es full-only y NO está en lite).
- **Alternativa descartada**: dejar el nombre del test con `JR`. Rechazada: queda marca `jr` en un símbolo Go exportado del test, inconsistente.

### D3 — Goldens y `sdd-orchestrator.md` se editan + se validan por test
Los dos goldens (`compose_all_toggles_{claude,opencode}.golden`) contienen el bloque compuesto del orquestador; `go test ./internal/harness/config/...` los compara. Se editan a mano (o se regeneran con el flag de update del paquete si existe) y el test valida. Los `sdd-orchestrator.md` bundleados cambian la instrucción "invoke `jr-orchestrator`" → "invoke `active-orchestrator`".

### D4 — Validación por capas (orden del apply)
1. **Catálogo** (regla dura): editar `harnesses.yaml` (`id`+`name`) → `go test ./internal/catalog/...` (incluye `catalog.Load()` + las aserciones del test renombrado).
2. **Tests `.go` restantes**: actualizar refs en los 5 tests no-catalog → `go test` de cada paquete.
3. **Assets `.md` + goldens**: editar `sdd-orchestrator.md` + goldens → `go test ./internal/harness/config/...`.
4. **Docs + e2e**: `README.md`, `ARCHITECTURE.md`, `e2e/e2e_test.sh` (docs y script, no afectan build).
5. **Red de seguridad final**: `go build ./...` + `go test ./...` = 23 OK / 0 FAIL.

### D5 — Scope cross-repo explícito (3 frentes)
1. **active-stack (este repo)**: harness `id`/`name` + ~42 refs. ← lo que implementa este change.
2. **Repo `Group-Active-IA/active-orchestrator` (org)**: su `SKILL.md` (`name:` y namespace del slash-command `jr-orchestrator:` → `active-orchestrator:`). Trabajo del operador sobre la copia de la org (parte del "strip jr" ya previsto). **Runbook**, no se implementa desde active-stack.
3. **`~/.claude/CLAUDE.md` global**: refs a `jr-orchestrator`. Fuera del repo — solo **FLAG**, no auto-editar (governance: config del usuario, no pisar sin pedido).

### D6 — Gate de consistencia cross-repo (OBLIGATORIO)
El `id`/`name` del harness en active-stack DEBE matchear el `name:` del `SKILL.md` del repo de la org. Si active-stack dice `active-orchestrator` pero el `SKILL.md` de la org sigue diciendo `jr-orchestrator`, la skill instalada no matchea lo que el flujo de fundación invoca → **roto en runtime**. El apply debe verificar/coordinar este alineamiento (o, si la org aún no migró, dejarlo flaggeado como bloqueante de release).

## Risks / Trade-offs

- **[Desalineación cross-repo]** active-stack renombra pero el `SKILL.md` de la org sigue en `jr-orchestrator` → la skill instalada no matchea el namespace que el flujo de fundación invoca. → **Mitigación**: gate D6 obligatorio en el apply; el rename de active-stack no se considera "listo para release" hasta que el `SKILL.md` de la org esté alineado (runbook frente 2).
- **[Muscle-memory / refs globales]** cambiar el namespace `jr-orchestrator:` → `active-orchestrator:` rompe los slash-commands que el usuario ya tiene memorizados y las refs en `~/.claude/CLAUDE.md` global. → **Mitigación**: coordinación explícita con el operador (Open Question 1 + 3); el global es solo FLAG, no auto-edición.
- **[Goldens stale]** editar `sdd-orchestrator.md` sin regenerar los goldens (o viceversa) → `go test ./internal/harness/config/...` falla. → **Mitigación**: capa 3 del apply trata ambos juntos y corre el test del paquete antes de avanzar.
- **[Aserciones del catálogo]** olvidar las aserciones internas de `catalog_test.go` (no solo el nombre del test) → falso verde aparente pero el test compara el ID viejo y falla. → **Mitigación**: D2 lo explicita; capa 1 corre `go test ./internal/catalog/...`.
- **[False positives]** un reemplazo global ciego tocaría `jr-stack`, `Jrh`, `JrbApNNgaTdGDITg`. → **Mitigación**: el patrón de reemplazo es el string exacto `jr-orchestrator` (+ los dos símbolos CamelCase nombrados), no un `jr` suelto.

## Migration Plan

- **Deploy**: rename in-place en una rama del change; sin migración de datos (es un instalador, no hay estado persistente del identificador).
- **Orden**: capas D4 (1→5), corriendo el test del paquete tocado en cada capa antes de avanzar.
- **Rollback**: `git restore` de los archivos tocados; el change no commitea hasta pedido explícito del operador.
- **Coordinación de release**: NO marcar el debranding como cerrado hasta que el frente 2 (SKILL.md de la org) esté alineado (gate D6).

## Open Questions

1. ¿El namespace del slash-command cambia `jr-orchestrator:` → `active-orchestrator:`? Rompe muscle-memory del usuario y refs en `~/.claude/CLAUDE.md` global, pero es coherente con "todo active". **Recomendación: sí**, pero requiere editar el `SKILL.md` del repo de la org en coordinación (frente 2). — Decisión del operador.
2. ¿Entran los `JRStackInvocation` (4 nombres de funciones Go del rename de producto, en `assets/commands_test.go`) por consistencia, o quedan como cosméticos del rename de producto previo? Son cosméticos (no afectan comportamiento). **Recomendación: sí por consistencia "todo active"**, pero es decisión del operador y NO es el foco de este change.
3. ¿Quién/cómo edita el `SKILL.md` del repo `Group-Active-IA/active-orchestrator` (org) y el `~/.claude/CLAUDE.md` global del usuario? Coordinación fuera de active-stack (frentes 2 y 3). — Decisión del operador.
