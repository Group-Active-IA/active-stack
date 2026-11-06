## Context

Este es L3 del plan aprobado de idioma + descripciones largas del instalador GUI
(ver `C:\Users\Juan Cruz Robledo\.claude\plans\me-gusto-pero-el-dazzling-music.md`,
sección L3). L1 (`i18n-engine-locales`) y L2 (`catalog-localized-descriptions`) ya
landearon (sin commitear, todo verde): la infraestructura bilingüe existe y sus
tests de paridad de claves y validación de catálogo la protegen. Lo que falta es el
**contenido real**: hoy los `long_description` son placeholders de 1-2 oraciones que
pasan la validación pero no explican nada.

Este change no es de código. No hay tipos nuevos, ni parsers, ni validadores, ni
firmas que cambien. Es **redacción de copy** dentro de campos string que ya existen.
El riesgo técnico es mínimo (romper el parseo YAML o la paridad i18n); el riesgo real
es de **calidad editorial**, y por eso el proceso central es un gate de aprobación
humana antes de escribir.

### Estado verificado del source (contado, no estimado)

**A. `internal/catalog/harnesses.yaml` — 13 harnesses global-picker** (tienen
`long_description` hoy, `!IsStarterOnly()`):

| # | id | name | type |
|---|----|------|------|
| 1 | `openspec` | OpenSpec CLI | external |
| 2 | `engram` | Engram | external |
| 3 | `context7` | Context7 | external |
| 4 | `sdd-orchestrator` | SDD Orchestrator | config |
| 5 | `permissions` | Permissions (security-first) | config |
| 6 | `active-orchestrator` | active-orchestrator | skill |
| 7 | `kb-creator` | kb-creator | skill |
| 8 | `roadmap-generator` | roadmap-generator | skill |
| 9 | `agent-instruction` | agent-instruction | skill |
| 10 | `skill-registry` | skill-registry | skill |
| 11 | `starter-add-command` | Starter Add Command | command |
| 12 | `find-skill` | find-skill | skill (3rd-party) |
| 13 | `skill-creator` | skill-creator | skill (3rd-party) |

**B. `internal/catalog/harnesses.yaml` — 4 starters** (todos con `long_description`):
`base`, `backend`, `ux-ui`, `active-ia`.

**C. `internal/i18n/table_{es,en}.go` — 8 claves `*.long`:**
- Modos (3): `mode.lite.long`, `mode.full.long`, `mode.custom.long`
- Tiers (3): `tier.estricto.long`, `tier.balanceado.long`, `tier.bypass.long`
- Estrategias (2): `strategy.targeted.long`, `strategy.restore.long`

**Total: 25 items × 2 idiomas = 50 textos (25 pares ES+EN).** (El plan estimó
~28-35 pares; el conteo real sobre el source es 25.)

## Goals / Non-Goals

**Goals:**
- Reemplazar los 25 pares de placeholders por contenido real de 3-6 oraciones cada
  uno, en ES (voseo, tono del catálogo actual) y EN.
- Que cada texto siga el formato: **qué es → qué instala/configura concretamente
  (binarios, archivos, MCPs, dónde) → cuándo conviene elegirlo (y cuándo no)**.
- Presentar la tabla completa al usuario para aprobación explícita ANTES de escribir.
- Mantener `catalog.Load()` y los tests de i18n verdes tras cada edición.

**Non-Goals:**
- NO agregar `long_description` a los harnesses starter-only (ver Decisión D2).
- NO tocar `name`, IDs, claves i18n, schema YAML, tipos, parsers ni `internal/backup`.
- NO tocar los `description` cortos (ya bilingües desde L2 y fuera de scope).
- NO hacer build ni commit (regla del operador; el commit lo autoriza el usuario).
- NO agregar spec deltas (ver Decisión D3).

## Decisions

### D1 — Formato y tono de cada texto
Cada `long_description`/`*.long` tiene 3-6 oraciones con esta estructura:
1. **Qué es** — una frase que define la opción.
2. **Qué instala/configura concretamente** — binario, paquete, archivo, MCP, dónde
   aterriza (p.ej. "se clona al dir de skills del agente", "inyecta el bloque en
   CLAUDE.md/AGENTS.md", "registra un MCP `engram`").
3. **Cuándo conviene elegirlo (y cuándo no)** — el criterio de decisión para el
   usuario.

**Tono ES = voseo**, consistente con los `description` cortos ya presentes en el
YAML ("Elegí exactamente qué instalar", "Usalo cuando ya conocés tu stack",
"podés agregar más después"). **EN** neutro y paralelo al ES en contenido (mismo
qué/qué/cuándo), no traducción literal palabra por palabra. `name` nunca se traduce
(son nombres de producto).

Alternativa descartada: textos más largos tipo documentación. Se descarta porque el
panel de detalle lateral tiene ancho acotado (MinWidth≈230, ScrollViewer) y 3-6
oraciones es lo que se lee cómodo sin scroll excesivo.

### D2 — Los harnesses starter-only NO reciben `long_description`
Los 30 harnesses con `scope: starter-only` (las skills curadas del starter pack) se
materializan solo vía `active-stack starter add`, nunca aparecen individualmente en
el picker de la GUI. Por las reglas de validación de L2, `long_description` es
**opcional** para ellos (obligatorio solo para `!IsStarterOnly()` y para starters).
Su detalle explicativo lo cubre el `long_description` del **starter** que los agrupa
(`base`/`backend`/`ux-ui`/`active-ia`), que sí se redacta acá. Agregarles texto
individual sería trabajo sin superficie de UI que lo muestre. **Confirmado: este
change no les agrega `long_description`.**

Alternativa descartada: redactar los 30 igual "por completitud". Se descarta porque
infla el scope y el gate de revisión sin beneficio de UX (no hay panel que los
muestre solos).

### D3 — Un delta spec mínimo y honesto (revisado tras verificar con la CLI)
La intención original era omitir specs por completo: este change no cambia el
contrato de código (schema YAML, claves i18n, obligatoriedad bilingüe y
validación ya están especificados y testeados por L1 y L2). Sin embargo,
`openspec validate catalog-long-desc-content --strict` se probó en vivo y
**rechaza changes sin al menos un delta** ("Change must have at least one
delta. No deltas found"), sin importar el schema (`spec-driven` exige
`specs` como dependencia dura de `tasks`, ver `schema.yaml`). El precedente
del repo de changes sin `specs/` (p.ej. `archive/2026-06-11-rename-jr-orchestrator-to-active`)
no aplica: esos archived items ya no se validan como *change* con esta regla
(o fueron creados con una versión anterior de la CLI que no la exigía).

En vez de forzar un spec vacío/artificial que declare comportamiento de
runtime falso ("el texto ahora es mejor" no es una aserción de contrato), se
agregó `specs/catalog-localization/spec.md` con **dos requirements honestos y
sí verificables como proceso**: (1) el contenido sigue la estructura de tres
partes (D1) para cada tipo de item en scope; (2) el contenido se presenta
completo y se aprueba explícitamente antes de escribirse (formaliza D4 como
requirement, no solo como tarea). Se declara con header `## ADDED
Requirements` porque `catalog-localization` — introducida por L2 — todavía no
está archivada a `openspec/specs/` al momento de crear este change; compone
igual sin importar el orden real de archivado entre L2 y L3.

### D4 — El gate de aprobación es una tarea con checkpoint, no un paso opcional
El flujo de apply es estrictamente: **(1) redactar los 25 pares → (2) presentar la
tabla COMPLETA ES+EN al usuario → (3) esperar aprobación explícita → (4) recién
entonces escribir los 3 archivos → (5) verificar tests verdes**. Escribir antes de
la aprobación viola el gate. Esto se encoda en `tasks.md` con el checkpoint entre la
redacción y la escritura. Governance: `catalog`/`model` es BAJO, pero este contenido
es cara visible al usuario final → se trata con revisión humana obligatoria.

### D5 — Orden de edición y verificación incremental
Al escribir (post-aprobación): editar en grupos (harnesses → starters → i18n) y
correr `go test ./internal/catalog/... ./internal/i18n/...` tras cada grupo, para
que un error de sintaxis YAML o una clave rota se detecte acotado. Los tests de
paridad de claves y validación bilingüe de L1/L2 son la red de seguridad: si un
texto queda en un solo idioma, el test falla ruidoso.

## Risks / Trade-offs

- **[Romper el parseo YAML al editar bloques `long_description` multilínea]** →
  Mitigación: usar el mismo estilo de bloque inline `{ es: "...", en: "..." }` que ya
  usa el YAML; correr `go test ./internal/catalog/...` tras cada grupo editado.
- **[Dejar un texto en un solo idioma / romper paridad de claves i18n]** →
  Mitigación: los tests de paridad (L1) y validación bilingüe (L2) fallan ruidoso;
  se corren en el paso de verificación antes de dar por cerrado.
- **[Contenido editorialmente pobre o inconsistente en tono]** → Mitigación: el gate
  de aprobación humana (D4) es exactamente para esto; la tabla completa se revisa
  antes de tocar archivos.
- **[Comillas/escapes en los strings Go de `table_*.go`]** → Mitigación: evitar
  comillas dobles sin escapar dentro de los literales; correr `go test ./internal/i18n/...`.
- **Trade-off asumido**: no se redactan los 30 starter-only (D2). Aceptado: no hay
  superficie de UI que los muestre individualmente.
