# Inventario del contexto de partida

> Generado por `go test ./eval/taskcontext/inventory -update` desde los goldens de
> `internal/harness/config/testdata/` (`inventory_compose.golden`, `inventory_sources.golden`), el catálogo
> y `eval/taskcontext/inventory/skills-terceros.json`. No se edita a mano: un test falla si se desvía.
> Change `opsx-task-context`, tareas 1.1 a 1.7. Las líneas son `wc -l` (cantidad de saltos de línea).

## 1. Bloque compuesto por combinación de toggles

El bloque instalado hoy es el último renglón: Claude entrega en capas por `ConfigDeliveryManagedSkill` (a CLAUDE.md llega solo el bloque universal; el coordinador vive en una skill aparte, que este inventario aún no suma), Codex compone por la rama `default` de `internal/harness/config/install.go` y OpenCode por `ConfigDeliveryPrimaryAgent`, y `h.Toggles` sale literal del catálogo con los cinco toggles activos.

| Toggles activos | Claude | Codex |
|---|---|---|
| ninguno | 7.116 B · 188 L | 5.543 B · 147 L |
| +delegation | 8.121 B · 210 L | 6.394 B · 168 L |
| +model-routing | 8.899 B · 226 L | 6.844 B · 181 L |
| +governance | 10.175 B · 246 L | 8.120 B · 201 L |
| +engram | 16.268 B · 365 L | 14.213 B · 320 L |
| +tdd (**lo que se instala hoy**) | **20.065 B · 448 L** | **18.010 B · 403 L** |

## 2. Atribución de bytes por fragmento

Cada fragmento es la diferencia entre dos combinaciones consecutivas; la base es el bloque sin toggles.

| Fragmento | Claude | % | Codex | % |
|---|---|---|---|---|
| base del variante | 7.116 B | 35,5 % | 5.543 B | 30,8 % |
| `delegation` | 1.005 B | 5,0 % | 851 B | 4,7 % |
| `model-routing` | 778 B | 3,9 % | 450 B | 2,5 % |
| `governance` | 1.276 B | 6,4 % | 1.276 B | 7,1 % |
| `engram` | 6.093 B | 30,4 % | 6.093 B | 33,8 % |
| `tdd` | 3.797 B | 18,9 % | 3.797 B | 21,1 % |

`engram` y `tdd` juntos son 49,3 % del bloque de Claude y 54,9 % del de Codex.

## 3. Otras variantes (referencia, fuera del alcance de paridad)

| Variante | Base | Con los cinco toggles |
|---|---|---|
| antigravity | 4.923 B · 134 L | 18.102 B · 397 L |
| cursor | 5.579 B · 147 L | 18.556 B · 411 L |
| gemini | 5.525 B · 147 L | 17.992 B · 403 L |
| generic | 5.461 B · 154 L | 17.847 B · 408 L |
| opencode | 7.371 B · 187 L | 20.320 B · 447 L |
| windsurf | 5.794 B · 149 L | 18.677 B · 418 L |

## 4. Fuentes inyectadas en el archivo de instrucciones

Cada sección se mide instalándola sola en un home limpio; el total, instalando ambas. La suma por cliente coincide con el total realmente inyectado (el separador entre secciones es de un byte).

| Cliente | Bloque del orquestador | Sección design-mockups | Separador | Total inyectado |
|---|---|---|---|---|
| Claude | 4.382 B | 8.940 B | 1 B | 13.323 B |
| Codex | 3.661 B | 8.940 B | 1 B | 12.602 B |

El fragmento `design-mockups` mide 8.865 B; con los marcadores de sección de la inyección son 8.940 B.

## 5. Todas las fuentes, por tamaño descendente

Las cinco más grandes están en negrita: Skill tup-campus-navigator (60.699 B), Skill active-orchestrator (38.606 B), Skill rubrica-builder (35.164 B), Skill skill-creator (33.168 B), Skill emil-design-eng (27.123 B).

| # | Fuente | Tipo | Bytes | Nota |
|---|---|---|---|---|
| 1 | **Skill tup-campus-navigator** | skill | 60.699 | terceros @2530ac78 |
| 2 | **Skill active-orchestrator** | skill | 38.606 | terceros @63c5131c |
| 3 | **Skill rubrica-builder** | skill | 35.164 | terceros @253c454f |
| 4 | **Skill skill-creator** | skill | 33.168 | terceros @2c5958ec |
| 5 | **Skill emil-design-eng** | skill | 27.123 | terceros @e8a175de |
| 6 | Skill pwa-development | skill | 25.315 | terceros @55aabe24 |
| 7 | Skill design-taste-frontend-v1 | skill | 21.195 | terceros @b482f7a9 |
| 8 | Skill typescript-advanced-types | skill | 16.806 | terceros @782565f7 |
| 9 | Skill postgresql-table-design | skill | 16.082 | terceros @782565f7 |
| 10 | Skill tailwind-design-system | skill | 15.730 | terceros @782565f7 |
| 11 | Skill interface-design | skill | 14.925 | terceros @782565f7 |
| 12 | Skill dashboard-crud-page | skill | 12.759 | terceros @782565f7 |
| 13 | Skill fastapi-domain-service | skill | 12.746 | terceros @782565f7 |
| 14 | Skill roadmap-generator | skill | 12.384 | terceros @deaabe9d |
| 15 | Skill postgresql-optimization | skill | 11.368 | terceros @222eba62 |
| 16 | Skill api-security-best-practices | skill | 11.267 | terceros @560b762c |
| 17 | Skill monorepo-scaffold | skill | 10.079 | terceros @782565f7 |
| 18 | Skill ws-frontend-subscription | skill | 9.677 | terceros @782565f7 |
| 19 | Skill test-driven-development | skill | 9.578 | terceros @22f0b463 |
| 20 | Skill kb-creator | skill | 9.575 | terceros @0f73e9b8 |
| 21 | Skill systematic-debugging | skill | 9.530 | terceros @22f0b463 |
| 22 | Skill zustand-store-pattern | skill | 9.024 | terceros @782565f7 |
| 23 | Sección design-mockups inyectada — Claude | instrucción inyectada | 8.940 | fragmento de 8.865 B más marcadores; solo con el harness design-mockups |
| 24 | Sección design-mockups inyectada — Codex | instrucción inyectada | 8.940 | fragmento de 8.865 B más marcadores; solo con el harness design-mockups |
| 25 | Skill react19-form-pattern | skill | 8.734 | terceros @782565f7 |
| 26 | Skill redis-best-practices | skill | 8.526 | terceros @7682ca77 |
| 27 | Skill agent-instruction | skill | 8.243 | terceros @1f7922f9 |
| 28 | Skill help-system-content | skill | 8.039 | terceros @782565f7 |
| 29 | Skill skill-registry | skill | 7.702 | terceros @78434d8f |
| 30 | Skill clean-architecture | skill | 7.397 | terceros @782565f7 |
| 31 | Skill design-system | skill | 6.948 | embebida |
| 32 | Skill websocket-engineer | skill | 6.586 | terceros @078a25cc |
| 33 | Skill receiving-code-review | skill | 6.203 | terceros @22f0b463 |
| 34 | Skill discovery-research | skill | 6.148 | terceros @dd7efb7d |
| 35 | Skill alembic-migrations | skill | 5.577 | terceros @782565f7 |
| 36 | Skill find-skill | skill | 5.472 | terceros @b716f9dc |
| 37 | Skill web-scraper | skill | 5.079 | terceros @1ac73c20 |
| 38 | Bloque sdd-orchestrator inyectado — Claude | instrucción inyectada | 4.382 | bloque universal de la entrega en capas; sección con marcadores |
| 39 | Skill fastapi-code-review | skill | 3.686 | terceros @782565f7 |
| 40 | Bloque sdd-orchestrator inyectado — Codex | instrucción inyectada | 3.661 | bloque universal de la entrega en capas; sección con marcadores |
| 41 | Skill requesting-code-review | skill | 2.974 | terceros @22f0b463 |
| 42 | Skill multi-stage-dockerfile | skill | 2.314 | terceros @222eba62 |
| 43 | Comando claude/active/starter-add.md | comando | 885 |  |
| 44 | Comando opencode/active-starter-add.md | comando | 790 |  |
| 45 | Skill openspec-archive | skill | 179 | embebida sin harness en el catálogo |
| 46 | Skill openspec-explore | skill | 179 | embebida sin harness en el catálogo |
| 47 | Skill openspec-onboard | skill | 179 | embebida sin harness en el catálogo |
| 48 | Skill openspec-propose | skill | 179 | embebida sin harness en el catálogo |
| 49 | Skill openspec-design | skill | 177 | embebida sin harness en el catálogo |
| 50 | Skill openspec-verify | skill | 177 | embebida sin harness en el catálogo |
| 51 | Skill openspec-apply | skill | 175 | embebida sin harness en el catálogo |
| 52 | Skill openspec-tasks | skill | 175 | embebida sin harness en el catálogo |
| 53 | Skill openspec-init | skill | 173 | embebida sin harness en el catálogo |
| 54 | Skill openspec-spec | skill | 173 | embebida sin harness en el catálogo |
| 55 | Skill judgment-day | skill | 171 | embebida sin harness en el catálogo |

### Skills del catálogo sin medición

Ninguna queda sin entrada: estas dicen por qué no tienen tamaño.

| Skill | Motivo |
|---|---|
| agile-product-owner | sin SKILL.md en skills/agile-product-owner del commit 08cd8a9f (el layout del repo cambió: sin SKILL.md en la ruta) |
| code-review-excellence | sin SKILL.md en skills/code-review-excellence del commit 9c61ab24 (el layout del repo cambió: sin SKILL.md en la ruta) |
| fastapi-templates | sin SKILL.md en skills/fastapi-templates del commit 9c61ab24 (el layout del repo cambió: sin SKILL.md en la ruta) |
| playwright-best-practices | sin SKILL.md en skills/playwright-best-practices del commit 32a1b130 (el layout del repo cambió: sin SKILL.md en la ruta) |
| python-testing-patterns | sin SKILL.md en skills/python-testing-patterns del commit 9c61ab24 (el layout del repo cambió: sin SKILL.md en la ruta) |
| vercel-react-best-practices | sin SKILL.md en skills/vercel-react-best-practices del commit 1a2c79d4 (el layout del repo cambió: sin SKILL.md en la ruta) |

Las skills de terceros se midieron una vez (2026-11-13) contra el commit que resolvió cada `ref: latest`; no son reproducibles bit a bit hasta que el catálogo las pinee por SHA.
