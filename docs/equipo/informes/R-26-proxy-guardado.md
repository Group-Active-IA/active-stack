# Informe de cierre · R-26 · Proxy MCP del guardado etiquetado por rama

> Lo recibe **quien esta ficha desbloquea** (queda en el PR y en `main`). Juan Cruz solo mergea. Menos de una pantalla.

**Quién:** agente, a pedido del dueño · **Fecha:** 2026-11-22
**PR:** ver el PR de la rama · **Repo:** active-stack · **Rama:** `feat/proxy-guardado-memoria`

**Qué se hizo (3 líneas):**
`active-stack memory proxy -- <comando de Engram>` (paquete `internal/memoryproxy`): proxy MCP stdio que etiqueta `mem_save` y las vías narrativas con `memory.PrepareSave`, deja pasar todo lo demás byte a byte y falla abierto.
Biblioteca `internal/harness/proxyreg` (sin conectar): reescribe la entrada `engram` de `~/.claude.json` y de `~/.codex/config.toml` para que arranque el proxy, con backup, idempotente y devolviendo los mismos bytes al quitar.
Fases 5 y 6 del plan reescritas para el proxy (DA-13). Ficha extra: [R-26](../fichas/R-26-proxy-guardado.md); **debe mergearse antes de R-20**.

**Cómo verificarlo:**
Con `USERPROFILE`, `HOME`, `LOCALAPPDATA` y `APPDATA` en un directorio temporal: `go vet ./...` y `go test ./... -count=1` (todo OK); `go test ./internal/memoryproxy ./internal/harness/proxyreg ./cmd/active-stack -count=1`.

**Qué tiene que hacer el dueño (solo si hace falta):**
- [ ] Revisar y mergear el PR (governance ALTO por el registro: no toca `filemerge`, `backup` ni `pipeline`)
- [ ] Decidir cómo registra R-20 el proxy cuando Engram es un plugin de Claude y no hay entrada `engram` de usuario (ficha, `(TBD)`)

**Decisiones que necesito de vos:** una, de producto: con Engram como plugin de Claude, recomiendo que R-20 cree una entrada de usuario `engram` ya proxied (el plugin sigue disponible) y avise; la alternativa es dejar el guardado sin etiquetar con aviso.

**Este informe va para Juan Cruz (ficha R-20)**
**Qué desbloquea:** R-20 de Juan Cruz (conecta `proxyreg.Register` y `Unregister` a install/uninstall)
**Mi siguiente ficha:** ninguna

**No ejecutado / límites:** Engram real detrás del proxy, Codex real (diferido por el dueño al final), registro sobre archivos de configuración reales, Claude Code con el proxy y un subagente, y `go build`. No se editó `docs/estado-equipo.md`.
