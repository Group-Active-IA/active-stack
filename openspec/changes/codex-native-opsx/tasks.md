# Tasks · Codex Native OPSX

Las tareas marcadas están implementadas y probadas en perfiles aislados; continúan pendientes los contratos compartidos y la aceptación real por interfaz. Matías confirmó Machine inicial, snapshot persistente protegido por ruta y avance de tareas independientes con integración de evidencia sujeta a 05/06. El alcance concreto está en design. Cada tarea espera sólo los pendientes que afectan sus rutas y contratos. Las tareas independientes pueden avanzar dentro del alcance aprobado. Matías aprobó el alcance ALTO descrito en design y autorizó iniciar las tareas independientes el 13/11/2026. Antes de editar producción, ejecutar safety net del área y seguir RED → GREEN → triangulación → refactor con evidencia por tarea. Las pruebas se ejecutan en perfiles aislados. El registro de código probado debe identificar también cambios sin commit. No ejecutar build sin pedido explícito.

## 1. Contratos y comprobaciones previas

- [ ] 1.1 Alinear el contrato técnico de entrega con 03 y completar la matriz compartida para Machine, incluida evidencia directa por acción. Verificar coincidencia de filas, targets y rutas entre ambos diseños.
- [ ] 1.2 Resolver P2 con sondas de CLI y extensión IDE en Windows y Linux. Registrar versión, mecanismo de roles, política explícita, contexto, controles y reanudación por celda. Verificar un pedido implícito y uno manual por interfaz.
- [ ] 1.3 Revisar con 01/03/08 el diff de escritura del snapshot persistente confirmado. Verificar que la reversión posterior usa estado esperado persistente y no depende sólo de guardas en memoria.
- [ ] 1.4 Fijar revisiones corregidas de 05/06 y consumir el sobre de 07. Mantener la integración de evidencia pendiente de esas entregas mientras avanzan tareas independientes. Verificar ejecución directa y delegada con los mismos criterios de evidencia y el sobre de retorno común.

## 2. Entrega y composición

- [x] 2.1 Capturar safety net de model, adapters y config. Caracterizar consumidores Claude/OpenCode/Antigravity y design-mockups antes de cambiar despacho. Registrar tests existentes y fallos preexistentes.
- [x] 2.2 Implementar la entrega y las rutas acordadas en P1 mediante el adapter, preservando el valor cero de ConfigDelivery. Probar ruta Machine, error por interfaz opcional ausente y el diagnóstico de Project legado.
- [x] 2.3 Componer skill con agents/openai.yaml, bloque universal y módulos de apoyo. Probar política allow_implicit_invocation falsa, mapa de toggles, ausencia de referencias a módulos omitidos y ausencia de aliases de modelos de Claude.
- [ ] 2.4 Entregar especialistas según la sonda P2 con los módulos y reglas del encargo. Probar TOML cuando corresponda, rutas con espacios y Windows, y fallback genérico con el sobre de 07/06.
- [ ] 2.5 Integrar verify y el contrato de evidencia requerido por acción. Probar apply directo, verificación breve independiente, revisor ausente y evidencia autoconsistente sin corroboración que quede sin verificar.

## 3. Instalación, actualización y retiro

- [x] 3.1 Implementar preflight por ruta y propiedad antes de escribir. Probar skill o rol ajeno, archivo inválido, notas propias dentro de la skill y ausencia de cambios después del rechazo.
- [x] 3.2 Incorporar las rutas al snapshot y plan. Migrar markers actuales y heredados, preservar secciones hermanas y contenido propio y escribir atómicamente. Probar instalación limpia, migración, fallo de backup, coexistencia de markers y reinstalación sin reescritura.
- [x] 3.3 Implementar el punto persistente de reversión confirmado y persistir su identidad y estado esperado. Probar actualización doble, manifiesto recargado y edición posterior conservada. Documentar acceso al respaldo en las interfaces aprobadas.
- [x] 3.4 Probar rollback con fallo de escritura después de AGENTS.md y con fallo duro de verify. Verificar recuperación de rutas sin cambios posteriores y preservación de una edición concurrente con conflicto informado.
- [x] 3.5 Integrar update con y sin registro de instalación. Probar replay, idempotencia y ausencia de adopción silenciosa sin registro. Documentar el término registro de instalación en los mensajes.
- [x] 3.6 Implementar uninstall archivo por archivo para Machine. Probar archivo ajeno dentro de la skill, directorio no vacío, ruta ausente y rollback sin pérdida de datos del usuario.

## 4. Verificación y documentación de entrega

- [x] 4.1 Extender verify de instalación para propiedad, skill, policy, módulos y referencias. Probar piezas ausentes y distinguir resultado estructural de aceptación del cliente real.
- [ ] 4.2 Repetir los mismos escenarios funcionales en CLI y extensión IDE en Windows y Linux: normal, explícito, especialista, verify independiente, módulo/control obligatorio ausente, protección equivalente y control opcional ausente. Guardar evidencia por celda y declarar las que no pudieron probarse.
- [ ] 4.3 Ejecutar con 02 la comparación de sesión normal y coordinada, incluidas aprobaciones y métricas ausentes por interfaz. Verificar que la baseline pendiente no se presenta como concluida y que terminal integrada se clasifica como CLI.
- [ ] 4.4 Verificar regresiones de consumidores, pruebas de áreas afectadas, gofmt y go vet, y suite Go completa al entregar. Registrar RED/GREEN, casos de borde y revisión independiente de preservación, ownership y rollback.
- [x] 4.5 Actualizar README, arquitectura y AGENTS/CLAUDE en espejo para invocación, matriz, fallback, límites y reversión. Si se toca el catálogo validar catalog.Load(). Comprobar que los comandos documentados existen.
- [ ] 4.6 Realizar revisión conjunta 03/04 sobre la matriz y escenarios comunes antes de promover la migración. Registrar disposición humana del alcance ALTO y pendientes de las dependencias.

## Workflow follow-up

- Archivar después de verificar los escenarios y resolver las limitaciones de aceptación registradas.
- Los informes propose anteriores conservan evidencia histórica del diseño sustituido.

## Avance del 13 de noviembre

- 2.1 y 2.3 completadas para composición independiente. Ver `informes/apply-composition-2026-11-13.md`.
- La entrega Machine ya está conectada al instalador, verify y retiro. Project conserva su recorrido legado.
- 3.3 completada: primer snapshot persistente, dos actualizaciones, publicación atómica, listado existente y restauración protegida desde disco. Ver informe de lifecycle y verificación final.
- 03 mantiene pendiente el contrato común de entrega y matriz final. 05/06 mantienen pendiente la integración de captura y recibos corroborados.
- 1.2 y 4.2 requieren clientes CLI/IDE Windows/Linux. Ver sonda local en `informes/apply-2026-11-13-sonda-local.md`.
- El protocolo pendiente de ejecución para esas celdas está en [plan de aceptación de clientes](informes/plan-aceptacion-clientes-04.md); su preparación no marca tareas de aceptación como completas.


## Integración del 13 de noviembre

La evidencia por tarea está en `informes/apply-lifecycle-2026-11-13.md` y `informes/verificacion-integracion-2026-11-13.md`.

- 3.1–3.4 prueban propiedad, migración, snapshot persistente, reintentos, fallos de escritura y rollback tardío con notas o directorios sustituidos.
- 3.5 reutiliza el replay de update por install. Las pruebas dirigidas de update preservan la configuración sin registro; el recorrido nativo de instalación prueba reaplicación y actualización de módulos.
- 3.6 y 4.1 prueban retiro por archivo y estructura, ownership y policy. No equivalen a aceptación del comportamiento del modelo.
- 4.5 actualiza README, arquitectura y el bloque correspondiente de AGENTS/CLAUDE en espejo.
- 2.2 completada en el instalador: entrega Machine, error de capacidad opcional ausente y diagnóstico stderr de Project legado, sin cambiar su comportamiento ni etiquetarlo degradado. La alineación compartida con 03 sigue en 1.1/4.6; la aceptación de clientes reales, en 1.2/4.2. Ver `informes/avance-paralelo-04-2026-11-13.md`. 2.4 espera sondas de roles. 2.5 espera contratos/captura corregidos de 05/06.
- 4.4 sigue pendiente: la suite completa tiene fallos reproducidos en main; el detalle y los comandos están en el informe final. CLI/IDE Windows/Linux y comparación modular siguen pendientes en 1.2, 4.2 y 4.3.
- Revisión independiente de preservación y diagnóstico, con corrección del retorno parcial tras fallo de un adapter posterior y regresión de rollback: `informes/revision-independiente-04-2026-11-13.md`. La tanda final conserva los mismos fallos globales conocidos; el conteo de tareas permanece 11/21.
