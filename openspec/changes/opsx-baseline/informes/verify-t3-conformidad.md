# verify tanda 3 · conformidad (`/opsx:verify`) · opsx-baseline

Fecha 2026-11-11. Ejecutor: orquestador (Claude). Por decisión del responsable ("constructor, verify y apply; no uses nada que se salga de eso") no se despacharon breaker ni auditor: este informe es solo de conformidad y no constituye aprobación de cierre.

## EVIDENCIA
- `openspec validate opsx-baseline --strict` → válido. `openspec status` → planificación y artefactos completos.
- Tareas: 9 / 16 hechas (1.1, 1.3, 1.4, 2.1, 2.4, 2.5, 2.6, 2.7, 3.1). Pendientes: 1.2, 2.2, 2.3, 3.2, 3.3 (clientes reales), 3.4 (D3 topes, D5 cuentas), 4.1 (03/04).
- `gofmt -l eval` vacío; `go vet ./eval/...` ok; `go list -deps ./cmd/... ./internal/...` sin `/eval`. Suite completa del constructor 3B: `go test -count=1 ./...` ok (15:24–15:26).

## Mapeo requisito → implementación
| Requisito | Estado | Dónde |
|---|---|---|
| Corridas reproducibles y aisladas | Cubierto con datos sintéticos | `manifest`, `fixture` (Prepare/Reset/huella) |
| Cohortes por interfaz | Cubierto | `manifest` (terminal integrada → terminal), `summary` (filas por interfaz) |
| Permisos de medición | Parcial | estado `bloqueada_por_permisos` y chequeo de permisos efectivos; perfiles reales pendientes de 1.2 |
| Evidencia de aceptación por tarea | Cubierto | `fixture` Evaluate/Decide, 5 familias |
| Rúbrica de revisión de un diff sembrado | Cubierto | `review` (emparejamiento máximo) |
| Contexto de sesión normal y variantes | Parcial | `usage` (sonda sintética); variante modular pendiente de 4.1 |
| Registro de consumo y cobertura | Parcial | `usage`, `record`; atribución real pendiente de 2.2/2.3 |
| Medición del tiempo | Cubierto (sintético) | `timing`; validación con demora real pendiente de 3.3 |
| Comparación por tarea aceptada | Cubierto (sintético) | `summary` |
| Evidencia apta para compartir | Parcial | `record` con lista sensible mínima; lista dueña 05 pendiente |

## CRITICAL (antes de archivar)
- 7 tareas sin completar (listadas arriba). No se puede archivar.

## WARNING
- Sin breaker ni auditor en la tanda 3: no hay verificación independiente de 2.1, 2.4–2.7, 3.1. Riesgo aceptado por el responsable.
- `summary` se probó con planes armados a mano, sin `Plan()` real (deuda del constructor → 2.2).
- Lista sensible heurística hasta que 05 publique la suya.

## Evaluación
Conformidad evaluada para el alcance sintético. No habilita archivo ni cierre del change.
