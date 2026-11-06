## Why

Las descripciones largas (`long_description`) que L2 sembró en el catálogo y las
que L1 sembró en las tablas i18n (`mode.*.long`, `tier.*.long`, `strategy.*.long`)
son placeholders de una o dos oraciones: alcanzan para pasar la validación de
paridad de idiomas, pero **no explican qué es cada opción, qué instala/configura
concretamente ni cuándo conviene elegirla**. El panel de detalle lateral de la GUI
(L5) va a mostrar exactamente estos textos: si quedan como están, el panel que
justifica todo el esfuerzo de i18n muestra contenido pobre. Este change redacta el
contenido real, revisado y bilingüe, para que el instalador explique de verdad cada
decisión al usuario.

Es un change de **contenido/copy**, no de código: no toca tipos, parsers,
validadores ni comportamiento. Solo reemplaza el texto de campos ya existentes.

## What Changes

- Reemplazar el `long_description` (ES + EN) de los **13 harnesses del picker
  global** (`!IsStarterOnly()`) en `internal/catalog/harnesses.yaml` por texto real
  de 3-6 oraciones con formato: qué es → qué instala/configura → cuándo conviene.
- Reemplazar el `long_description` (ES + EN) de los **4 starters** (`base`,
  `backend`, `ux-ui`, `active-ia`) en `harnesses.yaml`.
- Reemplazar los textos largos (ES + EN) de **3 modos** (`lite`, `full`, `custom`),
  **3 tiers** (`estricto`, `balanceado`, `bypass`) y **2 estrategias**
  (`targeted`, `restore`) en `internal/i18n/table_es.go` y `table_en.go`.
- Total: **25 pares** de textos ES+EN (17 en el YAML, 8 en las tablas i18n).
- Los **harnesses starter-only** (30 skills materializadas solo vía
  `starter add`) NO reciben `long_description`: sigue siendo opcional por las
  reglas de L2, y su detalle lo cubre el `long_description` del starter que las
  agrupa. Se documenta el razonamiento en `design.md`.
- **GATE de aprobación** (no negociable): la fase apply debe presentar la tabla
  COMPLETA de textos ES+EN propuestos al usuario y obtener aprobación explícita
  ANTES de escribir cualquier archivo. Encodeado como checkpoint en `tasks.md`.
- Sin cambios de contrato: los IDs, claves i18n, schema YAML y firmas de funciones
  quedan idénticos. Solo cambia el valor string de campos ya existentes.

## Capabilities

### New Capabilities
<!-- Ninguna funcional. `openspec validate --strict` exige al menos un delta por
     change; en vez de forzar un spec vacío o inventar comportamiento de runtime
     que no existe, se agrega un delta mínimo y honesto sobre la capability
     `catalog-localization` (ver Modified Capabilities). -->

### Modified Capabilities
- `catalog-localization`: agrega dos requirements de **contenido y proceso** (no
  de código): (1) el `long_description`/`*.long` de harnesses global-picker,
  starters, modos, tiers y estrategias sigue la estructura de tres partes (qué
  es → qué instala/configura → cuándo conviene, en 3-6 oraciones por idioma);
  (2) el contenido redactado se presenta completo (ES+EN) al usuario y se
  aprueba explícitamente ANTES de escribirse en
  `harnesses.yaml`/`table_{es,en}.go`. Declarado `ADDED` en el delta spec
  porque `catalog-localization` (introducida por L2) todavía no está archivada
  a `openspec/specs/` al momento de crear este change — compone igual si L2
  archiva antes o después de L3.

## Impact

- **Archivos editados** (solo valores de string, sin cambios estructurales):
  - `internal/catalog/harnesses.yaml` — 17 bloques `long_description` (13 harnesses
    global-picker + 4 starters).
  - `internal/i18n/table_es.go` y `internal/i18n/table_en.go` — 8 claves `*.long`
    cada uno (3 modos + 3 tiers + 2 estrategias).
- **Sin impacto en**: tipos (`model`), parsers, validadores, firmas de funciones,
  claves i18n, IDs de contrato, `internal/backup`. No es breaking.
- **Verificación**: `go test ./internal/catalog/... ./internal/i18n/...` debe seguir
  verde (parseo + validación bilingüe + paridad de claves). No hay build.
- **Downstream**: L5 (`gui-detail-panel`) consume estos textos en el panel de
  detalle; este change es el que le da valor real a ese panel.
