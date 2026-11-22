# R-26 · Proxy MCP del guardado etiquetado por rama

> Traspaso único de la ficha. Lo crea y lo mantiene quien la implementa. No crear otros archivos de handoff para esta ficha.

**Responsable:** agente a pedido del dueño (Juan Cruz) · **Rama:** `feat/proxy-guardado-memoria` · **Repo del PR:** active-stack
**Plan y fase:** [EN] Fases 5 y 6 (reescritas) · **Ficha extra** creada por la decisión **DA-13** del dueño (2026-11-21). **Debe mergearse antes de R-20** (R-20 conecta el registro a install/uninstall). **Depende de:** R-08 y R-15 (núcleo `internal/memory`, `PrepareSave`), R-16 (patrón de `internal/harness/hooks`), R-05 (prototipo B medido, DA-05)

## Objetivo para el goal

Entregar el etiquetado de guardados por rama como **proxy MCP stdio** (`active-stack memory proxy`) entre el agente y Engram, en Claude y Codex, sin hook de guardado; más la biblioteca que registra el proxy en la entrada MCP de Engram de cada agente, sin conectarla a install/uninstall.

## Criterios de aceptación

- [x] Proxy: pasa todo lo demás sin cambios, conserva ids y fin de línea, falla abierto, maneja la salida del hijo (`go test ./internal/memoryproxy/... -count=1`, 21 casos)
- [x] Subcomando `active-stack memory proxy [--project] -- <cmd> [args]` con prueba extremo a extremo y de uso (`go test ./cmd/active-stack -count=1`)
- [x] Registro en Claude y Codex: marcas, backup, idempotente, mismos bytes al quitar, sin tocar otras entradas (`go test ./internal/harness/proxyreg/... -count=1`, 23 casos)
- [x] Fases 5 y 6 del plan reescritas con el proxy (ids conservados)
- [x] `go vet ./...` y `go test ./... -count=1` verdes con home aislado
- [ ] Codex real y Engram real: **no ejecutado** (ver Evidencia)

## Decisiones

| ID | Decisión | Quién decidió | Fecha |
|---|---|---|---|
| DA-13 | Etiquetado por proxy en Claude y Codex, sin hook de guardado; hooks de lectura se mantienen | Juan Cruz (dueño) | 2026-11-21 |
| R26-1 | El proxy copia el stdout del hijo directo al agente (como el prototipo) y solo transforma agente→hijo; reenvía los bytes originales si nada cambia o si el etiquetado falla o tarda más de 3 s | agente | 2026-11-22 |
| R26-2 | Herramientas etiquetadas: `mem_save`, `mem_session_summary`, `mem_session_finalize`, `mem_capture_passive`, `mem_save_prompt` (las del prototipo); lotes JSON-RPC (`[...]`) pasan sin etiquetar | agente | 2026-11-22 |
| R26-3 | Registro: se reescribe la entrada **existente** `engram` (no se crea una entrada nueva, porque cambiaría el nombre de herramienta). Claude: el comando original viaja crudo dentro de `args` (JSON sin comentarios) para devolver los mismos bytes. Codex: tabla original comentada con `#orig ` dentro de marcas | agente | 2026-11-22 |
| R26-4 | No se modificaron `filemerge`, `backup`, `pipeline` ni `jsonedit`; solo se los llama | agente | 2026-11-22 |

## `(TBD)` para el dueño o R-20

- **Engram como plugin de Claude:** su `.mcp.json` está en la caché del plugin (nombre `mcp__plugin_engram_engram__*`), no en `~/.claude.json`. La biblioteca solo edita un archivo con `mcpServers.engram`; qué hace R-20 si no existe (crear una entrada de usuario `engram` con el comando de Engram, o dejar el guardado sin etiquetar con aviso) **no está decidido**.
- `~/.claude.json` lo reescribe también Claude Code mientras corre: el registro debe hacerse con Claude cerrado o reintentarse (no probado).
- Una tabla `[mcp_servers.engram]` al final del archivo sin salto de línea se maneja (`#orig-noeol`), pero un Codex real no lo probó.
- Umbral de sobrecosto (DA-02): el proxy no agrega tokens propios; Codex se re-mide en R-08 `(TBD)`.

## Documentación consultada

| Recurso | Versión | Enlace |
|---|---|---|
| Ficha R-05 y matriz Claude/Windows (prototipo B) | repo, 2026-11-19 | [R-05](R-05.md) |
| Prototipo `proxy_b.py` | repo | `docs/research/r01-prototipo/proxy_b.py` |
| Plan, Fases 5 y 6 | repo, reescrito 2026-11-22 | [plan](../../plan-engram-cloud-equipo.md) |
| Informes R-15 y R-16 | repo | [R-15](../informes/R-15.md), [R-16](../informes/R-16.md) |

## Evidencia

Entorno: Windows 11, Go 1.26, `USERPROFILE`/`HOME`/`LOCALAPPDATA`/`APPDATA` en un directorio temporal.

- `go vet ./...`: sin hallazgos. `go test ./... -count=1`: todos los paquetes OK.
- `internal/memoryproxy`: 21 casos (hijo falso = el binario de prueba por pipes reales: tuberías, CRLF/LF, ids grandes, 5 herramientas, error y tiempo excedido, salida del hijo con y sin cliente abierto, hijo inexistente, línea de 3 MiB).
- `internal/harness/proxyreg`: 23 casos (round-trip byte a byte en 7 formas de JSON y 8 de TOML, idempotencia, backup antes y fallo de backup, archivos inválidos, otros agentes sin tocar).
- `cmd/active-stack`: extremo a extremo con repo Git real (rama `feature/x`) y sin marcador.
- **No ejecutada:** Engram real detrás del proxy en Claude; Codex real (diferido por el dueño al final); `go build`/`install` reales; registro sobre un `~/.claude.json` o `config.toml` reales; Claude Code con el proxy y un subagente.

## Punto de reanudación (se sobrescribe, no se acumula)

**Hecho:** proxy, subcomando, biblioteca de registro, pruebas, Fases 5-6, ficha e informe.
**Abierto:** PR sin mergear; verificación con Engram y Codex reales; los `(TBD)` de arriba.
**Decisiones pendientes:** dueño (plugin de Claude, ver `(TBD)`).
**Siguiente acción:** revisar y mergear antes de R-20; R-20 conecta `proxyreg.Register/Unregister`.
**Subagentes en curso:** ninguno.
