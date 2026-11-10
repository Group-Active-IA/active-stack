# Tasks · Pilotos de modelos y caché

Las tareas de implementación quedan pendientes hasta la próxima etapa. Resolver D1–D5 durante la propuesta. Las corridas requieren el protocolo y la referencia aprobados de 02 y el flujo verificable de 07.

## 1. Preparación común

- [ ] 1.1 Fijar en un manifiesto las revisiones de 02 y 07, fixtures, rúbricas, matriz, controles, criterios de calidad, umbrales y límites acordados. Comprobar con un manifiesto incompleto que el arnés identifica la precondición faltante.
- [ ] 1.2 Incorporar al arnés el registro de brazos, pares, configuración efectiva, condiciones de caché y motivos de exclusión. Probar con datos sintéticos que un cambio en los controles invalida el contraste y que los resultados conservan su evidencia depurada.

## 2. Piloto A · Modelo y esfuerzo por rol

- [ ] 2.1 Preparar el control de 07 y los brazos nativos de modelo y esfuerzo en perfiles aislados. Calibrar qué evidencia confirma valores efectivos por rol en cada superficie y probar un fallback o un parámetro no disponible.
- [ ] 2.2 Ejecutar la matriz acordada del piloto A y verificar que los pares mantienen el mismo flujo y estrategia de prefijo. Registrar por separado desvíos de caché, sustituciones y brazos combinados que se hayan acordado.

## 3. Piloto B · Estabilidad de prefijos y caché

- [ ] 3.1 Preparar el control y el prefijo estable con contenido equivalente y condiciones fría y caliente. Calibrar aislamiento, ventana y alcance de la señal. Probar lecturas completas y parciales, misses y telemetría insuficiente con registros de prueba.
- [ ] 3.2 Ejecutar los bloques acordados del piloto B. Verificar que se conserva la asignación de cada corrida, el estado observado, el costo de preparación y el uso de coordinador, workers y reintentos.

## 4. Resultados y decisión

- [ ] 4.1 Generar reportes separados por piloto y cohorte. Probar con fixtures de resultados que el cálculo incluye fallos, reintentos y preparación imputable, evita duplicar agregados de workers y distingue API, cuota y contexto lógico. Cubrir cero aceptaciones y cobertura insuficiente.
- [ ] 4.2 Revisar calidad, dispersión, cobertura, tiempos y mejora frente a los umbrales acordados. Verificar que el reporte permite reconstruir cada recomendación y bloquea adoptar si hay una regresión obligatoria. Publicar los resultados depurados y las decisiones por cohorte.
