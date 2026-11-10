# Tasks: opsx-shadow-review

> Ronda `propose`: todas las tareas quedan pendientes. El piloto no toca el binario.
> Prerrequisitos: 06 y 08 aprobadas; protocolo de 02 (`240e1b2`) y pilotos de 09 (`89f1fa5`) disponibles para los grupos 4 y 5.
> Primera ronda (decisión confirmada): sólo históricos y preparados, material público revisado y datos ficticios.

## 0. Decisiones previas (grupo)

- [ ] 0.1 Registrar proveedor, paquete de datos, acceso y métricas acordados, más tamaño de muestra, U1–U3, acuerdo mínimo de etiquetado, custodio y etiquetadores.
- [ ] 0.2 Junto con 09, en una sola tarea compartida: obtener documentación de Jev (qué es, entrada, salida, costo, retención y entrenamiento) o registrar que no se evalúa en ninguna de las dos fichas.

## 1. Muestra

- [ ] 1.1 Mapear los commits `fix` de `main` al commit que introdujo cada defecto; descartar los que no se puedan ubicar. Aceptación: tabla fix → origen → archivos → requisito afectado.
- [ ] 1.2 Construir casos preparados reutilizando la familia "Revisión de un diff sembrado" de 02 (fixtures, rúbrica y regla de emparejamiento) y sumando los rechazos de 06 (recibo falso con hash correcto, mismo HEAD con diff distinto, salida truncada sin original, tests verdes sin cubrir el requisito) y defectos de lógica con ubicación conocida.
- [ ] 1.3 Armar controles: inspeccionar manualmente los changes candidatos y construir fixtures limpios. Un change sin fixes posteriores no entra como control sin inspección.
- [ ] 1.4 Doble etiquetado independiente + resolución por tercero; calcular acuerdo. Congelar la muestra en un commit.
- [ ] 1.5 Armar cada paquete histórico desde el estado previo al fix y verificar que no contiene el fix, sus mensajes, referencias, texto posterior, ramas ni etiquetas.

## 2. Referencia

- [ ] 2.1 Definir la revisión de referencia por caso: requisitos y escenarios, recibos ficticios de 06, la tabla de cobertura de 08 aplicada a las acciones del caso y decisión humana.
- [ ] 2.2 Medir la referencia contra las etiquetas congeladas (VP, FP, FN por severidad).

## 3. Datos

- [ ] 3.1 Lista de inclusión y armado del paquete con huella.
- [ ] 3.2 RED: fixtures con secreto sembrado en un diff público, ruta del home, variable de entorno y `settings.json` real deben cancelar el envío. GREEN: examen previo sobre todo el paquete con la lista de patrones de 05 (D14), leída de su declaración en el catálogo y no copiada.
- [ ] 3.3 Registro de envío por caso, candidato y proveedor.
- [ ] 3.4 Verificar retención y uso para entrenamiento de cada proveedor como condiciones separadas antes del primer envío.

## 4. Arnés en sombra

- [ ] 4.1 Arnés externo al binario; verificar que el diff no toca `cmd/` ni `internal/`.
- [ ] 4.2 Almacén fuera del árbol, con huella y acceso restringido al custodio; apertura sólo con caso cerrado definitivamente. RED: intento de lectura anticipada desde el rol de referencia o etiquetado debe denegarse y registrarse.
- [ ] 4.3 Prompt de revisión fijo y versionado; formato de hallazgo de D4.
- [ ] 4.4 Verificar criterios de entrada por candidato; registrar descartes.
- [ ] 4.5 Calibrar cada candidato con un conjunto propio fuera de la muestra (prompt, modelo y esfuerzo según 02 y 09).

## 5. Corridas y métricas

- [ ] 5.1 Correr cada candidato sobre la muestra congelada.
- [ ] 5.2 Emparejar hallazgos contra etiquetas congeladas (tercer etiquetador).
- [ ] 5.3 Arbitraje aparte, ciego al origen, de los hallazgos sin emparejar: defecto nuevo confirmado, falsa alarma, indeterminado.
- [ ] 5.4 Calcular precisión, exhaustividad, señal adicional (cantidad), proporción de controles con falsa alarma, costo por caso y minutos de triage.

## 6. Cierre

- [ ] 6.1 Informe con datos crudos, muestra, acuerdo, arbitraje aparte, log de envío y salida (seguir explorando, repetir, descartar).
- [ ] 6.2 Si la salida es seguir explorando: abrir la decisión de arquitectura separada y evaluar si una ronda posterior incluye evidencia real redactada o changes vivos; no integrar en esta ficha.
