# verify tanda 4 · conformidad (`/opsx:verify`) · opsx-baseline

Fecha 2026-11-11. Ejecutor: orquestador (Claude). Sin breaker ni auditor por decisión del responsable; informe de conformidad, no aprobación de cierre.

## EVIDENCIA
- Rama `propose/opsx-baseline`, base `3f8afd2`, diff sin commitear (profile/, capture/, fixtures/samples/, manifest VersionAtLeast, README, tasks.md).
- `openspec validate opsx-baseline --strict` válido; `gofmt -l eval` vacío; `go vet ./eval/...` ok; `cmd/`/`internal/` no importan `eval/`.
- Suite del constructor: `go test -count=1 -skip 'TestAddToUserPath' ./...` todo ok (19:36–19:38).
- Tareas: 10/16. Tanda 4: 2.2 [x]; 1.2 y 2.3 abiertas.

## Conformidad
| Tarea | Estado | Motivo |
|---|---|---|
| 1.2 | Parcial | 8 perfiles documentados y comparación; falta verificación con clientes reales en Windows y Linux (16 ítems en `profile.Pending()`). |
| 2.2 | Hecha con riesgo declarado | Adaptadores Claude (OTel, `-p`, stream-json) y Codex (`exec --json`) sobre muestras derivadas de la doc. IDE queda "no comprobada": la doc no describe fuentes del IDE. **Lectura a ratificar por el responsable.** |
| 2.3 | Parcial | Claude cubierto; Codex padre/hijos no documentado → `no_comprobado`, requiere corrida real. |

## WARNING
- Contradicción con design.md: en Claude, `usage` del resultado excluye subagentes; solo `modelUsage`/`total_cost_usd` los incluyen. El ejemplo de 60.000 del design vale para `modelUsage`. Corregir design.md (planificación).
- DEUDA de seguridad: `manifest.isContentOptIn` no cubre `otel.log_agent_responses` ni `otel.log_guardian_assessments` de Codex (opt-ins de contenido según la doc).
- Citas de documentación obtenidas vía resumidor de WebFetch: confianza media, no textuales.
- `TestAddToUserPath` (internal/system) escribe en el PATH real de Windows; fuera de 02.

## Evaluación
Conformidad evaluada para el alcance documental. No habilita archivo.
