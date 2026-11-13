# Trazabilidad del alcance de 04 antes y después de separar deuda externa

13 de noviembre de 2026. Fuente previa: `cd8d8c5c7639dc26cbe26370f712a1acf0e04a42:openspec/changes/codex-native-opsx/tasks.md`.

Antes del ajuste: **21 tareas originales, 11 completas y 10 pendientes**, varias mixtas. Matías amplió su instrucción a registrar como deuda lo que no depende de él, incluidos insumos de 02 y disposiciones de otros responsables. No autorizó declarar hechas pruebas propias no ejecutadas.

Después: **20 unidades activas de Matías, 17 entregadas y 3 propias pendientes**. Las unidades sufijadas `a` son subalcances explícitos: completar `2.5a` no significa completar el `2.5` original. El cambio de denominador responde a dividir responsabilidades y retirar del checklist activo la medición dependiente de 02; no significa que se implementaron 17 de las 21 tareas originales. Los requisitos externos y la futura integración siguen en el registro de deuda, con criterio de cierre.

## Mapa de las diez tareas originalmente abiertas

| Original | Unidad activa de Matías | Parte externa en deuda |
|---|---|---|
| 1.1 | 1.1a: conciliación técnica Codex, matriz y paquete de revisión entregados | D04-0301: acuerdo del responsable de 03 y fila común de evidencia, más D04-0601 |
| 1.2 | 1.2: sondas propias P2, pendiente | Solo insumos externos de captura/recibos D04-0501/0601; no se difiere su autoría |
| 1.3 | 1.3a: evidencia de snapshot persistente y paquete de revisión entregados | D04-010308: disposición de responsables 01/03/08 |
| 1.4 | 1.4a: fijación de fuentes y consumo textual disponible entregados | D04-0501, D04-0601–0604, D04-0701–0702: operación e integración condicionada a entrega |
| 2.4 | 2.4: especialistas/rutas/encargo propios, pendiente | Insumos de retorno/recibos D04-0601/0701; no se difiere implementación independiente |
| 2.5 | 2.5a: normas parciales y pruebas de contenido/bloqueo distribuidas, entregadas | D04-0501, D04-0601–0604, D04-0701: verify funcional condicionado, no implementado |
| 4.2 | 4.2: aceptación propia con clientes reales, pendiente | Solo fuentes externas concretas anteriores; falta de ejecución propia no es deuda de otro |
| 4.3 | Sin unidad activa hasta insumo del proveedor | D04-0201: protocolo/referencia y medición conjunta con 02; Matías retoma como consumidor |
| 4.4 | 4.4a: comprobaciones y revisiones técnicas ejecutadas, entregadas | D04-BASE01: fallos preexistentes de main y su disposición por responsables |
| 4.6 | 4.6a: paquete de revisión y pruebas técnicas comunes entregados | D04-REV01: disposición humana conjunta; D04-0301 cuando aplique |

Las otras once tareas originales conservan su descripción y comprobación. El alcance funcional completo no se declara terminado ni archivado. El PR de GitHub se mantiene abierto hasta su disposición y merge; no es el estado OpenSpec.

## Checklist original conservado literalmente

Fotografía histórica; no es el checklist activo. No se interpreta como la lista actual de pendientes de Matías.

```markdown
- [ ] 1.1 Alinear el contrato técnico de entrega con 03 y completar la matriz compartida para Machine, incluida evidencia directa por acción. Verificar coincidencia de filas, targets y rutas entre ambos diseños.
- [ ] 1.2 Resolver P2 con sondas de CLI y extensión IDE en Windows y Linux. Registrar versión, mecanismo de roles, política explícita, contexto, controles y reanudación por celda. Verificar un pedido implícito y uno manual por interfaz.
- [ ] 1.3 Revisar con 01/03/08 el diff de escritura del snapshot persistente confirmado. Verificar que la reversión posterior usa estado esperado persistente y no depende sólo de guardas en memoria.
- [ ] 1.4 Fijar revisiones corregidas de 05/06 y consumir el sobre de 07. Mantener la integración de evidencia pendiente de esas entregas mientras avanzan tareas independientes. Verificar ejecución directa y delegada con los mismos criterios de evidencia y el sobre de retorno común.
- [x] 2.1 Capturar safety net de model, adapters y config. Caracterizar consumidores Claude/OpenCode/Antigravity y design-mockups antes de cambiar despacho. Registrar tests existentes y fallos preexistentes.
- [x] 2.2 Implementar la entrega y las rutas acordadas en P1 mediante el adapter, preservando el valor cero de ConfigDelivery. Probar ruta Machine, error por interfaz opcional ausente y el diagnóstico de Project legado.
- [x] 2.3 Componer skill con agents/openai.yaml, bloque universal y módulos de apoyo. Probar política allow_implicit_invocation falsa, mapa de toggles, ausencia de referencias a módulos omitidos y ausencia de aliases de modelos de Claude.
- [ ] 2.4 Entregar especialistas según la sonda P2 con los módulos y reglas del encargo. Probar TOML cuando corresponda, rutas con espacios y Windows, y fallback genérico con el sobre de 07/06.
- [ ] 2.5 Integrar verify y el contrato de evidencia requerido por acción. Probar apply directo, verificación breve independiente, revisor ausente y evidencia autoconsistente sin corroboración que quede sin verificar.
- [x] 3.1 Implementar preflight por ruta y propiedad antes de escribir. Probar skill o rol ajeno, archivo inválido, notas propias dentro de la skill y ausencia de cambios después del rechazo.
- [x] 3.2 Incorporar las rutas al snapshot y plan. Migrar markers actuales y heredados, preservar secciones hermanas y contenido propio y escribir atómicamente. Probar instalación limpia, migración, fallo de backup, coexistencia de markers y reinstalación sin reescritura.
- [x] 3.3 Implementar el punto persistente de reversión confirmado y persistir su identidad y estado esperado. Probar actualización doble, manifiesto recargado y edición posterior conservada. Documentar acceso al respaldo en las interfaces aprobadas.
- [x] 3.4 Probar rollback con fallo de escritura después de AGENTS.md y con fallo duro de verify. Verificar recuperación de rutas sin cambios posteriores y preservación de una edición concurrente con conflicto informado.
- [x] 3.5 Integrar update con y sin registro de instalación. Probar replay, idempotencia y ausencia de adopción silenciosa sin registro. Documentar el término registro de instalación en los mensajes.
- [x] 3.6 Implementar uninstall archivo por archivo para Machine. Probar archivo ajeno dentro de la skill, directorio no vacío, ruta ausente y rollback sin pérdida de datos del usuario.
- [x] 4.1 Extender verify de instalación para propiedad, skill, policy, módulos y referencias. Probar piezas ausentes y distinguir resultado estructural de aceptación del cliente real.
- [ ] 4.2 Repetir los mismos escenarios funcionales en CLI y extensión IDE en Windows y Linux: normal, explícito, especialista, verify independiente, módulo/control obligatorio ausente, protección equivalente y control opcional ausente. Guardar evidencia por celda y declarar las que no pudieron probarse.
- [ ] 4.3 Ejecutar con 02 la comparación de sesión normal y coordinada, incluidas aprobaciones y métricas ausentes por interfaz. Verificar que la baseline pendiente no se presenta como concluida y que terminal integrada se clasifica como CLI.
- [ ] 4.4 Verificar regresiones de consumidores, pruebas de áreas afectadas, gofmt y go vet, y suite Go completa al entregar. Registrar RED/GREEN, casos de borde y revisión independiente de preservación, ownership y rollback.
- [x] 4.5 Actualizar README, arquitectura y AGENTS/CLAUDE en espejo para invocación, matriz, fallback, límites y reversión. Si se toca el catálogo validar catalog.Load(). Comprobar que los comandos documentados existen.
- [ ] 4.6 Realizar revisión conjunta 03/04 sobre la matriz y escenarios comunes antes de promover la migración. Registrar disposición humana del alcance ALTO y pendientes de las dependencias.
```
