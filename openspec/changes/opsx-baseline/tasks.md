# Tasks · Medición de referencia de OPSX

Las tareas que dependen de D1, D3, D5 o D6 usan los valores que el grupo acuerde antes de la entrega final ([registro de decisiones](design.md#registro-de-decisiones)). Las corridas de referencia de Codex usan el contrato de Estricto vigente de 01 para la serie separada. El protocolo puede construirse y validarse con eventos sintéticos antes de esas corridas.

## 1. Protocolo y fixtures

- [ ] 1.1 Versionar el manifiesto y la matriz de cohortes por cliente, versión, sistema, interfaz real, modalidad y permisos efectivos, con tareas y pesos por cohorte. Verificar con manifiestos de prueba que una celda con versión incompatible, permisos distintos de los declarados o un opt-in de contenido activo queda incompatible antes de iniciar, fuera del denominador y listada con su motivo.
- [ ] 1.2 Definir los perfiles de medición por cliente e interfaz: perfil equivalente para edición y serie separada de Estricto. Verificar con el cliente real en Windows y Linux que el modo y los flags efectivos coinciden con el manifiesto y que una edición dentro del checkout no pide aprobación en el perfil equivalente.
- [ ] 1.3 Preparar los fixtures, prompts y rúbricas de las cinco familias. Ejecutar sus chequeos contra una respuesta o diff de referencia y contra una variante defectuosa. Verificar reinicio desde el fixture inmutable y restauración entre repeticiones.
- [ ] 1.4 Implementar la regla de emparejamiento de la revisión de un diff sembrado. Verificar con una clave y hallazgos sintéticos un verdadero positivo, un duplicado que no suma, una omisión, un falso positivo y un hallazgo nuevo derivado a adjudicación.

## 2. Arnés y captura

- [ ] 2.1 Implementar el ciclo de preparación, ejecución, intento y evaluación en perfiles descartables. Probar aceptación, rechazo por calidad, bloqueo por permisos, reintento, timeout y fallo del cliente, y verificar que cada caso queda con un estado distinto, la causa, el código de salida bruto y los eventos de error.
- [ ] 2.2 Implementar los adaptadores de captura para Claude Code y Codex en terminal e IDE. Probar con muestras reales de eventos de cada versión fijada, o con muestras sintéticas declarando el riesgo, en casos completos, parciales y duplicados. Incluir una corrida delegada por cliente y verificar que la regla de precedencia reproduce los totales de los ejemplos de reconciliación del diseño.
- [ ] 2.3 Comprobar en Claude la atribución por worker con `agent_id` desde la versión 2.1.310, que `agent.name = "custom"` no se usa para atribuir y que el total del cierre de un subagente no se toma como su consumo total. Comprobar en Codex si el agregado del padre incluye a los hilos hijos. Documentar las fuentes y los límites de atribución de cada versión e interfaz.
- [ ] 2.4 Implementar el registro de lecturas de archivos y llamadas a herramientas por tipo y rol. Verificar con una traza conocida que los conteos coinciden y que una interfaz sin esa señal la marca como no disponible.
- [ ] 2.5 Implementar la medición del tiempo con reloj monotónico, esperas observables aparte y unión de intervalos solapados. Verificar con intervalos sintéticos solapados y con una espera no delimitable que el ajustado figura como no disponible.
- [ ] 2.6 Implementar la sonda de sesión normal y el registro de caché. Comprobar con fixtures que distingue entrada de una solicitud, total del turno y métricas ausentes, y que una sesión nueva con lecturas cacheadas conserva el estado observado.
- [ ] 2.7 Emitir registros por intento y corrida con metadatos de cobertura aplicando la lista sensible y la política de 05 y las huellas de 06. Verificar con una traza sintética que incluya correo, UUID de cuenta, identificador de usuario, rutas del host, detalle de herramientas y cuerpos de solicitudes que ninguno llega al resultado compartido.

## 3. Resultados y referencia

- [ ] 3.1 Generar el resumen por cohorte y documentar su cálculo. Probar con resultados sintéticos que rechazos, bloqueos y reintentos contribuyen al costo total, que las celdas incompatibles quedan fuera del denominador, que API, suscripción e interfaces quedan separadas y que una cohorte sin aceptadas publica el costo total con el cociente sin valor.
- [ ] 3.2 Ejecutar una corrida de calibración por cliente e interfaz después de verificar las precondiciones. Incluir casos directo y delegado para comprobar correlación, permisos, fuentes y reglas de parada. Excluir las calibraciones de la muestra medida.
- [ ] 3.3 Ejecutar una corrida de calibración con una aprobación de demora conocida por cliente e interfaz. Verificar qué métricas de tiempo cambian y que el ajustado descuenta la demora solo cuando los intervalos quedan delimitados.
- [ ] 3.4 Ejecutar las repeticiones de la sonda y del flujo OPSX actual según los valores acordados en D3. Publicar el manifiesto, los resultados depurados y el resumen. Verificar que otra persona puede reconstruir los agregados y rastrear los veredictos desde la evidencia autorizada.

## 4. Comparación modular

- [ ] 4.1 Cuando 03 y 04 entreguen el flujo modular instalable, repetir las celdas comparables con el mismo protocolo, fixture, rúbrica y permisos efectivos y publicar los resultados emparejados por cliente, interfaz y modalidad. Verificar que las interfaces sin variante modular quedan como comparación pendiente.

## Workflow follow-up

- Entregar el protocolo, el manifiesto y los resultados de referencia a 07, 09 y 10 para sus pilotos.
