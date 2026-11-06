## 1. Verificar el estado actual del source (no asumir conteos)

- [x] 1.1 Releer `internal/catalog/harnesses.yaml` completo y confirmar la lista exacta
      de los 13 harnesses global-picker (`!IsStarterOnly()`) que hoy tienen
      `long_description` placeholder (ver tabla verificada en `design.md`).
- [x] 1.2 Confirmar los 4 starters (`base`, `backend`, `ux-ui`, `active-ia`) y sus
      `long_description` placeholder actuales.
- [x] 1.3 Releer `internal/i18n/table_es.go` y `table_en.go` y confirmar las 8 claves
      `*.long` (3 modos, 3 tiers, 2 estrategias) y su contenido placeholder actual.
- [x] 1.4 Si el conteo real difiere de lo documentado en `design.md` (25 items / 25
      pares), actualizar `design.md` antes de continuar. — conteo confirmado (25/25), sin cambios.

## 2. Redactar el contenido — harnesses del picker global (13)

- [x] 2.1 Redactar `long_description` ES+EN (3-6 oraciones, formato qué es → qué
      instala/configura → cuándo conviene) para: `openspec`, `engram`, `context7`.
- [x] 2.2 Redactar ídem para: `sdd-orchestrator`, `permissions`.
- [x] 2.3 Redactar ídem para: `active-orchestrator`, `kb-creator`, `roadmap-generator`,
      `agent-instruction`, `skill-registry`.
- [x] 2.4 Redactar ídem para: `starter-add-command`, `find-skill`, `skill-creator`.
- [x] 2.5 Releer los `description` cortos existentes en el YAML para calibrar tono ES
      (voseo) antes de dar por cerrado este grupo.

## 3. Redactar el contenido — starters (4)

- [x] 3.1 Redactar `long_description` ES+EN para `base`.
- [x] 3.2 Redactar `long_description` ES+EN para `backend`.
- [x] 3.3 Redactar `long_description` ES+EN para `ux-ui`.
- [x] 3.4 Redactar `long_description` ES+EN para `active-ia`.

## 4. Redactar el contenido — modos, tiers y estrategias (i18n, 8)

- [x] 4.1 Redactar `mode.lite.long`, `mode.full.long`, `mode.custom.long` (ES+EN).
- [x] 4.2 Redactar `tier.estricto.long`, `tier.balanceado.long`, `tier.bypass.long`
      (ES+EN) — mantener consistencia con `tier.bypass.warning` existente (no tocarlo).
- [x] 4.3 Redactar `strategy.targeted.long`, `strategy.restore.long` (ES+EN).

## 5. GATE — Aprobación del usuario (BLOQUEANTE, no escribir archivos antes de esto)

- [x] 5.1 Compilar UNA tabla completa con los 25 items × 2 idiomas (id/clave | ES
      propuesto | EN propuesto), agrupada por sección (harnesses / starters / i18n).
- [x] 5.2 Presentar la tabla completa al usuario y pedir aprobación EXPLÍCITA.
      **No avanzar a la sección 6 sin un OK explícito del usuario.** Si el usuario
      pide cambios, iterar sobre 2-4 y volver a presentar antes de continuar. — usuario
      aprobó la tabla completa AS-IS, sin cambios.

## 6. Escribir los archivos (solo tras aprobación explícita)

- [x] 6.1 Escribir los 13 `long_description` de harnesses global-picker aprobados en
      `internal/catalog/harnesses.yaml`.
- [x] 6.2 Correr `go test ./internal/catalog/...` — debe seguir verde. — verde.
- [x] 6.3 Escribir los 4 `long_description` de starters aprobados en el mismo YAML.
- [x] 6.4 Correr `go test ./internal/catalog/...` de nuevo — debe seguir verde. — verde.
- [x] 6.5 Escribir las 8 claves `*.long` aprobadas en `internal/i18n/table_es.go` y
      `internal/i18n/table_en.go`.
- [x] 6.6 Correr `go test ./internal/i18n/...` — debe seguir verde (incluye el test de
      paridad de claves ES/EN). — verde.

## 7. Verificación final

- [x] 7.1 Correr `go test ./internal/catalog/... ./internal/i18n/...` una vez más de
      punta a punta y confirmar 0 fallos. — 0 fallos; además `go test ./...` de punta a
      punta también verde (todos los paquetes del repo).
- [x] 7.2 Confirmar que ningún harness starter-only recibió `long_description` (D2 de
      `design.md`) — diff review del YAML. — confirmado: 17 ocurrencias del campo
      `long_description` en el YAML (13 harnesses + 4 starters), ninguna en bloques
      `scope: starter-only`.
- [x] 7.3 Confirmar que no se tocó ningún ID, clave i18n, tipo, parser ni
      `internal/backup` — diff review acotado a valores de string. — confirmado: únicos
      cambios son valores de `long_description` (YAML) y de las 8 claves `*.long` (i18n).
- [x] 7.4 Reportar al usuario: archivos tocados, resumen del contenido escrito, y que
      NO se hizo commit ni build (pendiente de pedido explícito del operador).
