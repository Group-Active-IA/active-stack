# Investigación (tarea 1.5): `verify-hook` con cero checks

Fecha: 2026-11-13. Base: `origin/main` `7a60089`. Entorno: Windows 11, Go 1.26.2, `HOME` temporal (no se tocó el entorno real).

## Reproducción
Binario compilado de `main` en un directorio temporal y corrido con un `--home` temporal:

```
as.exe install --mode lite --agent claude --home <tmp> --yes --no-self-install
...
  → verify-hook running...
  ✓ verify-hook
Verification checks: 0 passed, 0 failed, 0 warnings, 0 skipped
Installation verified. You're ready to start building.
```

Se reproduce el caso del 02/11 (modo `lite`, agente `claude`, Windows): el paso `verify-hook` aparece como correcto y el reporte dice cero checks.

## Causa
El reporte impreso **no** viene de los checks que corrió el hook.

- `cmd/active-stack/headless/executor.go:158-159` construye el reporte final con `verify.BuildReport(nil)` y lo imprime. Con una lista vacía, `Failed == 0` y `Ready == true`, así que siempre dice "0 checks ... Installation verified". El comentario de `:153-157` lo reconoce ("the real checks ran inside the verify-hook step").
- Los checks reales corren dentro del paso `verify-hook` (`internal/verify/hook.go`, `BuildHook`), pero su resultado se descarta cuando pasan: `BuildHook` devuelve `nil` y no devuelve ni imprime el reporte. Solo cuando falla se renderiza, dentro del error.
- La línea viene del commit `d34bb8be` (2026-05-29, repo anterior).

## Qué no es la causa (medido)
En un test descartable (borrado, no commiteado) que armó lo mismo que `main.go`: `collectSelectedHarnesses` devolvió 6 harnesses, `resolveVerifyAdaptersForTarget` devolvió 1 adaptador y `ChecksForHarness` produjo **6 checks** (openspec 1, engram 1, context7 2, sdd-orchestrator 1, permissions 1, starter-add-command 0). El hook de `lite`/`claude` no está vacío; lo vacío es lo que se imprime.

Caminos que sí construyen un hook vacío a propósito: `starter add` y el despacho de starters (`BuildHook(nil, nil, ...)`, `starter_add.go:59`, `starter_dispatch.go:81`). No son el caso del 02/11.

## Consecuencia para 06
El texto "Verification checks: 0 passed ... Installation verified" **no es evidencia de que se verificó algo**. Un reporte con cero checks debe tratarse como "sin verificar", coherente con la regla de 06 de que cero recibos nunca es verificado (D4). Un recibo de ejecución no puede apoyarse en ese reporte.

## Para 01, 04 y 08
- Corrección posible: que `BuildHook` guarde el reporte y `executor.go` imprima ese. Está en `cmd/active-stack/headless` (BAJO/MEDIO) y `internal/verify`. **No se tocó**: queda como propuesta, fuera del change 06.
- Pendiente: la parte Windows del flujo `windows install` (`runWindowsPipeline`) no se corrió; solo se reprodujo la ruta `install` headless.
