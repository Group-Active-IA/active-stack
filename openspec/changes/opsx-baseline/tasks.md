# Tasks · Medición de referencia de OPSX

La implementación comienza con D1–D6 acordadas y la propuesta revisada. Los valores sugeridos en el diseño son candidatos para el equipo. Las corridas de referencia de Codex requieren la configuración compatible verificada en 01. El protocolo puede construirse y validarse con eventos sintéticos antes de esas corridas.

## 1. Protocolo y fixtures

- [ ] 1.1 Versionar el manifiesto y la matriz acordados. Verificar que el manifiesto registre versiones, permisos, variantes, modalidad, presupuesto, semilla de orden y precondiciones, y que un perfil incompatible produzca un resultado bloqueado con causa.
- [ ] 1.2 Preparar los fixtures, prompts y rúbricas de las cinco familias. Ejecutar sus chequeos contra una respuesta o diff de referencia y contra una variante defectuosa. Verificar reinicio desde el fixture inmutable y restauración entre repeticiones.

## 2. Arnés y captura

- [ ] 2.1 Implementar el ciclo de preparación, ejecución, intento y evaluación en perfiles descartables. Probar éxito, rechazo por evidencia insuficiente, reintento y timeout. Verificar que quedan registrados el tiempo humano y el resultado por criterio.
- [ ] 2.2 Implementar los adaptadores de captura acordados para Claude Code y Codex. Probar con muestras de eventos completas, parciales y duplicadas, incluida una corrida delegada donde un agregado de sesión ya incluya a los workers. Documentar las fuentes y los límites de atribución de cada versión.
- [ ] 2.3 Implementar la sonda de sesión normal y el registro de caché. Comprobar con fixtures que distingue entrada de una solicitud, total del turno y métricas ausentes, y que una sesión nueva con lecturas cacheadas conserva el estado observado.
- [ ] 2.4 Emitir registros por intento y corrida con metadatos de cobertura. Verificar redacción, acceso y retención con una traza sintética que incluya credenciales de prueba, rutas personales e identificadores de cuenta.

## 3. Resultados y referencia

- [ ] 3.1 Generar el resumen por cohorte y documentar su cálculo. Probar con resultados sintéticos que los rechazos y reintentos contribuyen al consumo, API y suscripción quedan separadas y las cohortes sin aceptaciones o con cobertura incompleta muestran la causa de un indicador no calculable.
- [ ] 3.2 Calibrar las celdas acordadas después de verificar las precondiciones, incluida la compatibilidad y la ruta de edición en Codex. Ejecutar casos directo y delegado para comprobar correlación, permisos, fuentes y reglas de parada. Excluir las calibraciones de la muestra medida.
- [ ] 3.3 Ejecutar las repeticiones aprobadas de la sonda y del flujo OPSX actual. Publicar el manifiesto, los resultados depurados y el resumen. Verificar que otra persona puede reconstruir los agregados y rastrear los veredictos desde la evidencia autorizada.

## Workflow follow-up

- Cuando 03 y 04 habiliten el flujo modular, repetir las celdas comparables con el mismo protocolo y reportar los resultados emparejados. Mantener como pendientes las superficies sin variante modular disponible.
- Entregar el protocolo y los resultados de referencia a 07 y 09 para sus pilotos.
