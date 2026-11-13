# opsx-native-controls — change 08 (archivado con alcance parcial)

Responsable: Gianfranco · Revisor e integrador: Neyén · Gobernanza **ALTO** · Archivado el 2026-11-13 con `openspec archive`.

- Qué se entregó y qué no: [tasks.md](tasks.md) (6 unidades entregadas, 38 en deuda).
- Verificación con comandos y salidas reales: [informes/verificacion-08-2026-11-13.md](informes/verificacion-08-2026-11-13.md).
- Deuda aceptada, con dueño y criterio de cierre: [informes/deuda-aceptada-08-2026-11-13.md](informes/deuda-aceptada-08-2026-11-13.md).

**Es solo propuesta y spec.** No se escribió código ni se corrió el piloto: no hay matriz de cobertura medida, ni pruebas en Windows y Linux, ni escrituras de permisos o hooks.

**08 es ALTO.** Toca `permissions` y, probablemente, `filemerge`. La implementación espera la revisión humana de Neyén; ninguna decisión de este change se implementa antes.

**Aviso sobre la spec.** Al archivar, `specs/native-controls-pilot/spec.md` se sincroniza a `openspec/specs/native-controls-pilot/spec.md`. Las decisiones que lleva (política ante fallo, regla de control obligatorio ausente, preservación de entradas del usuario, archivo ilegible = abortar, reversión con `GuardRollback`, revisión de `filemerge`) son **propuesta de 08, no acuerdos del grupo**; siguen pendientes de confirmación.
