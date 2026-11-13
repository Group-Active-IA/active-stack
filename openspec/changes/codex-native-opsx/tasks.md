# Tasks · Codex Native OPSX

## Alcance activo de Matías y deuda externa

Esta lista sigue exclusivamente trabajo propio de 04. Matías pidió separar como **deuda** los cambios, insumos y disposiciones que no dependen de él; no dejarlos como tareas activas abiertas suyas. Las tareas mixtas se dividieron: el sufijo `a` identifica solo la parte entregada, sin afirmar completado el requisito original completo.

**20 unidades activas: 17 entregadas y 3 propias pendientes (1.2, 2.4, 4.2).** Antes del ajuste había 21 tareas originales, 11 completas y 10 pendientes mixtas. El checklist original literal y la correspondencia por responsabilidad están en [trazabilidad del alcance](informes/trazabilidad-alcance-y-deuda-04-2026-11-13.md). Este cambio de conteo no equivale a 17/21 tareas originales implementadas.

La deuda externa se gestiona fuera de este checklist en [registro de deuda](informes/deudas-aceptadas-05-06-07-2026-11-13.md): D04-0501/0601–0604/0701–0702, D04-0201, D04-0301, D04-010308, D04-BASE01 y D04-REV01. Cada fila conserva proveedor, origen, disparador y criterio de cierre; Matías integra lo que corresponda cuando llegue el insumo. No se marca como implementado lo que no se hizo.

Alcance Machine y escritura ALTA ya aprobados. Las pruebas usan perfiles aislados; aplicar safety net y TDD antes de producción. No ejecutar build sin pedido explícito. La entrega funcional completa aún no está acreditada: las acciones con requisitos obligatorios ausentes se detienen y quedan sin verificar. No se archiva automáticamente ni se cierra el PR antes de su disposición y merge.

## Trabajo propio del alcance vigente

- [x] 1.1a Entregar conciliación técnica del despacho Codex con main, columna Codex de matriz y paquete de revisión compartida; no incluye la disposición humana de 03 ni la fila de evidencia final. Ver D04-0301/D04-0601.
- [ ] 1.2 Ejecutar las sondas propias P2 de Codex CLI y extensión IDE en Windows/Linux: versión, roles, invocación manual/implícita, contexto, controles y reanudación. Los insumos externos concretos se retoman desde deuda; esta comprobación de Matías sigue pendiente.
- [x] 1.3a Entregar pruebas de reversión con snapshot persistente, estado esperado y manifiesto recargado, junto con el diff y evidencia para revisión 01/03/08. La disposición de esos responsables se registra en D04-010308.
- [x] 1.4a Fijar revisiones 05/06/07 y consumir en Codex los cuatro módulos disponibles de 05 y las normas parciales D2–D4 de 06, con pruebas de composición y lifecycle. Operación final de captura/recibos/retorno queda en D04-0501/0601–0604/0701–0702.
- [x] 2.1 Capturar safety net de model, adapters y config. Caracterizar consumidores Claude/OpenCode/Antigravity y design-mockups antes de cambiar despacho. Registrar tests existentes y fallos preexistentes.
- [x] 2.2 Implementar la entrega y las rutas acordadas en P1 mediante el adapter, preservando el valor cero de ConfigDelivery. Probar ruta Machine, error por interfaz opcional ausente y el diagnóstico de Project legado.
- [x] 2.3 Componer skill con agents/openai.yaml, bloque universal y módulos de apoyo. Probar política allow_implicit_invocation falsa, mapa de toggles, ausencia de referencias a módulos omitidos y ausencia de aliases de modelos de Claude.
- [ ] 2.4 Implementar los especialistas y el encargo propios de Codex según P2, con rutas por adapter, espacios/Windows y alternativa genérica en lo independiente. La integración con contrato final de recibos/retorno espera D04-0601/D04-0701; no reemplaza el trabajo propio pendiente.
- [x] 2.5a Distribuir y probar las normas parciales de evidencia e independencia y los mensajes de bloqueo ante contrato/captura/revisor ausentes. No incluye verify funcional ni generación/validación operativa de recibos; esos tramos condicionados están en D04-0501/0601–0604/0701.
- [x] 3.1 Implementar preflight por ruta y propiedad antes de escribir. Probar skill o rol ajeno, archivo inválido, notas propias dentro de la skill y ausencia de cambios después del rechazo.
- [x] 3.2 Incorporar las rutas al snapshot y plan. Migrar markers actuales y heredados, preservar secciones hermanas y contenido propio y escribir atómicamente. Probar instalación limpia, migración, fallo de backup, coexistencia de markers y reinstalación sin reescritura.
- [x] 3.3 Implementar el punto persistente de reversión confirmado y persistir su identidad y estado esperado. Probar actualización doble, manifiesto recargado y edición posterior conservada. Documentar acceso al respaldo en las interfaces aprobadas.
- [x] 3.4 Probar rollback con fallo de escritura después de AGENTS.md y con fallo duro de verify. Verificar recuperación de rutas sin cambios posteriores y preservación de una edición concurrente con conflicto informado.
- [x] 3.5 Integrar update con y sin registro de instalación. Probar replay, idempotencia y ausencia de adopción silenciosa sin registro. Documentar el término registro de instalación en los mensajes.
- [x] 3.6 Implementar uninstall archivo por archivo para Machine. Probar archivo ajeno dentro de la skill, directorio no vacío, ruta ausente y rollback sin pérdida de datos del usuario.
- [x] 4.1 Extender verify de instalación para propiedad, skill, policy, módulos y referencias. Probar piezas ausentes y distinguir resultado estructural de aceptación del cliente real.
- [ ] 4.2 Ejecutar aceptación propia con Codex real en CLI y extensión IDE Windows/Linux: normal, explícito, especialista, verify en lo habilitado, módulo/control ausente y alternativas. Identificar los casos detenidos por deuda externa; no acreditar celdas no ejecutadas.
- [x] 4.4a Ejecutar y documentar regresiones, gofmt, go vet, suite Go y revisión técnica independiente de preservación/ownership/rollback sobre la entrega publicada. Registrar suite fallida y comparación exacta con main; resolución/disposición de fallos ajenos queda en D04-BASE01.
- [x] 4.5 Actualizar README, arquitectura y AGENTS/CLAUDE en espejo para invocación, matriz, fallback, límites y reversión. Si se toca el catálogo validar catalog.Load(). Comprobar que los comandos documentados existen.
- [x] 4.6a Entregar paquete revisable de matriz, compatibilidad y escenarios técnicos comunes 03/04 con resultados y límites. La disposición humana conjunta y revisión de responsables siguen en D04-REV01/D04-0301.

## Evidencia de las unidades entregadas

| Unidades | Evidencia |
|---|---|
| 1.1a, 4.6a | [Integración de main](informes/integracion-main-pr8-2026-11-13.md), [revisión de integración](informes/revision-integracion-pr8-04-2026-11-13.md). No sustituyen revisión humana de otros responsables. |
| 1.3a, 3.1–3.6 | [Lifecycle](informes/apply-lifecycle-2026-11-13.md), [backup](informes/apply-backup-2026-11-13.md), [revisión independiente](informes/revision-independiente-04-2026-11-13.md). |
| 1.4a, 2.3, 2.5a | [Consumo implementado de módulos](informes/apply-modulos-05-06-2026-11-13.md), [revisión de módulos](informes/revision-modulos-05-06-2026-11-13.md). Son normas y entrega estructural, no aceptación de verify funcional. |
| 2.1, 2.2 | [Composición](informes/apply-composition-2026-11-13.md), [diagnóstico Project](informes/avance-paralelo-04-2026-11-13.md). |
| 4.1, 4.4a, 4.5 | Informes de integración, módulos, revisiones independientes y [correcciones de revisión](informes/correcciones-revision-pr9-2026-11-13.md). Suite final: 2081 eventos aprobados, 6 fallidos conocidos y 3 omitidos; vet/race aprobados. Suite global no verde. |

## Trabajo propio que sigue pendiente

- **1.2:** obtener y registrar las sondas P2; el [plan de clientes](informes/plan-aceptacion-clientes-04.md) está preparado, no ejecutado por completo.
- **2.4:** entregar especialistas y sus rutas/encargos; no hay prueba de roles funcionales que permita darla por hecha.
- **4.2:** ejecutar la matriz con clientes reales sobre el paquete final. La sonda macOS solo acredita descubrimiento; Linux de Juani acredita instalador. No sustituyen CLI/IDE Windows/Linux.

No se asignó ese trabajo propio a terceros sin acuerdo ni se lo ocultó como deuda externa. Las partes realmente condicionadas a otros changes se retoman desde sus IDs de deuda, no como duplicados abiertos del checklist.
