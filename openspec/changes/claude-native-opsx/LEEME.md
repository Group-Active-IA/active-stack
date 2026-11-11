# claude-native-opsx: índice de informes

Propuesta 03 del roadmap (Claude Code, riesgo ALTO pendiente de confirmar por la persona). Rama `claude-native-opsx` con `main` `35a89f8` integrado. Fase: propose, revisada el 2026-11-11 según el informe unificado (`revision-2026-11-11.md`): el coordinador pasa a ser la skill `/sdd-orchestrator`. Falta la revisión del plan sobre esta versión, las decisiones abiertas de design.md y la aprobación humana.

| Informe | Rol | Ejecutor declarado | Intento | Veredicto |
|---|---|---|---|---|
| informes/propose-1-planner.md | planner | subagente nativo role-planner (Claude) | 1 | plan inicial |
| informes/propose-1-plan-reviewer.md | plan-reviewer | subagente nativo role-plan-reviewer (Claude) | 1 | VUELVE AL PLANNER (1 B, 9 M, 6 m) |
| informes/propose-2-planner.md | planner | mismo ejecutor, reanudado | 2 | correcciones |
| informes/propose-2-plan-reviewer.md | plan-reviewer | mismo ejecutor, reanudado | 2 | VUELVE AL PLANNER (3 M, 5 m) |
| informes/propose-3-planner.md | planner | mismo ejecutor, reanudado | 3 | correcciones N1-N8 |
| informes/propose-3-plan-reviewer.md | plan-reviewer | mismo ejecutor, reanudado | 3 | LISTO PARA COMPARTIR (O1 menor) |
| revision-2026-11-11.md | orquestador | Claude, con Juani | revisión de grupo | aplicado el informe unificado; pendientes listados |

Los informes `propose-*` corresponden al diseño anterior (agente de sesión con `claude --agent`, base 97d7f6c). Familia compartida: planner y plan-reviewer son la misma familia de modelo (Claude); la independencia es de rol y contexto, no de proveedor.
