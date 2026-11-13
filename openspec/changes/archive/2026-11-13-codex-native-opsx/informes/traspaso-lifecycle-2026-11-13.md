# Traspaso del lifecycle de Codex · 13 de noviembre de 2026

Este traspaso acompaña al PR del change 04. Su disponibilidad en main depende de la revisión y el merge del PR.

## Lo que el siguiente responsable podrá usar al integrar

- Composición de skill explícita, política y módulos por toggle mediante `ComposeCodexPackage`.
- Entrega Machine administrada con inventario, preflight y primer respaldo de migración.
- Verificación estructural y retiro por archivo propio mediante los helpers de configuración.
- Estado esperado persistente en los respaldos protegidos y ausencias esperadas en rollback de operaciones.
- Remoción de markers actuales y heredados conservando los bytes externos.

## Coordinación pendiente

- **03**. Alinear el contrato compartido de entrega y la matriz final, especialmente la carga de evidencia en ejecución directa. Project conserva la entrega anterior. Los roles personalizados esperan las sondas por versión e interfaz.
- **05**. Entregar módulos de lectura, herramientas y captura con fuente independiente y cobertura demostrada antes del truncamiento.
- **06**. Entregar recibos corregidos que identifiquen la ejecución y el estado exacto probado y separen integridad, procedencia y suficiencia.
- **02**. Reutilizar el arnés para comparar sesiones normal y coordinada por cliente e interfaz.
- **Aceptación real**. Completar CLI y extensión IDE en Windows/Linux. La sonda local de descubrimiento en macOS no acredita esas celdas.

Las acciones que necesitan un módulo o control obligatorio ausente quedan detenidas con el motivo. El relato de un worker conserva estado sin verificar hasta su corroboración.
