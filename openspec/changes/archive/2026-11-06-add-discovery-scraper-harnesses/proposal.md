## Why

Hoy el flujo de fundación (`active-orchestrator`) no tiene ninguna etapa de investigación de mercado/competidores antes de documentar: pasa directo de `openspec init` a `kb-creator`. Ya se construyó, publicó y cableó el contenido real de esa etapa — dos skills nuevas (`discovery-research`, `web-scraper`) publicadas en `Group-Active-IA`, y `active-orchestrator`/`kb-creator` (repos externos) ya editados y pusheados para ofrecer el gate de Discovery y leer su resultado. Falta el último eslabón: que el instalador (`active-stack`) sepa que estas dos skills existen para poder ofrecerlas/instalarlas en modo Full.

## What Changes

- Catalogar `discovery-research` como harness `type: skill` en `internal/catalog/harnesses.yaml`, apuntando a `Group-Active-IA/discovery-research`, `install_modes: [full]`, `depends_on: [web-scraper]` (la invoca como sub-skill cuando hay URLs de competidores).
- Catalogar `web-scraper` como harness `type: skill`, apuntando a `Group-Active-IA/web-scraper`, `install_modes: [full]`.
- Actualizar la `description` bilingüe (es/en) de la entrada existente `active-orchestrator` para reflejar la cadena nueva (`discovery-research` opcional antes de `kb-creator`), ya que su propio `SKILL.md` cambió de flujo.

Ninguno de estos cambios modifica código Go: `catalog.Load()`, `ByID`, `ForMode`, `ForAgent` operan genéricamente a partir del YAML.

## Capabilities

### New Capabilities
- `catalog-discovery-harnesses`: el catálogo embebido incluye las entradas `discovery-research` y `web-scraper`, cargables y resolvibles como cualquier otro harness de tipo `skill`, con el mismo nivel de validación bilingüe que ya exige `catalog-localization` para harnesses no-starter-only.

### Modified Capabilities
(ninguna — no hay spec existente que describa el contenido puntual de `active-orchestrator` como entrada de catálogo; el cambio de su `description` es contenido, no un requirement nuevo)

## Impact

- `internal/catalog/harnesses.yaml` — dos entradas nuevas + un campo `description` actualizado en una entrada existente.
- Sin impacto en código Go, tests existentes de catálogo (deben seguir pasando sin modificación), ni en starters (ninguno de los dos harnesses se agrega a un starter — quedan disponibles solo vía selección global en modo Full).
