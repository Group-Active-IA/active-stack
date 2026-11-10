# Tasks: opsx-shadow-review

> Ronda `propose`: todas las tareas quedan pendientes. El piloto no toca el binario.
> Prerrequisitos: 06 y 08 aprobadas; 09 (y 02) con protocolo de costo disponible para el grupo 4.

## 0. Decisiones previas (grupo)

- [ ] 0.1 Cerrar las decisiones de `design.md`: datos compartibles, condiciones del proveedor, etiquetadores, tamaño de muestra, umbrales U1–U3 y tratamiento de Jev.
- [ ] 0.2 Obtener documentación de Jev (qué es, entrada, salida, costo, política de datos) o registrar que no se evalúa.

## 1. Muestra

- [ ] 1.1 Mapear los commits `fix` de `main` al cambio que introdujo cada defecto; descartar los que no se puedan ubicar. Aceptación: tabla fix → origen → archivos → requisito afectado.
- [ ] 1.2 Construir fixtures sembrados con los casos de rechazo de 06 y defectos de lógica con ubicación conocida.
- [ ] 1.3 Elegir casos de control entre changes archivados sin fixes posteriores conocidos.
- [ ] 1.4 Doble etiquetado independiente + resolución por tercero; calcular acuerdo. Congelar la muestra en un commit.

## 2. Referencia

- [ ] 2.1 Definir la revisión de referencia por caso: requisitos y escenarios, recibos de 06, cobertura de 08 y decisión humana.
- [ ] 2.2 Medir la referencia contra las etiquetas (VP, FP, FN por severidad).

## 3. Datos

- [ ] 3.1 Lista de inclusión y armado del paquete con huellas antes y después de redactar.
- [ ] 3.2 Chequeo previo al envío de secretos, variables, rutas del home y configuración del usuario. RED: fixtures con cada clase prohibida deben cancelar el envío.
- [ ] 3.3 Registro de envío por caso, candidato y proveedor.

## 4. Arnés en sombra

- [ ] 4.1 Arnés externo al binario; verificar que el diff no toca `cmd/` ni `internal/`.
- [ ] 4.2 Almacén sellado con huella; apertura sólo con decisión registrada.
- [ ] 4.3 Prompt de revisión fijo y versionado; formato de hallazgo de D4.
- [ ] 4.4 Verificar criterios de entrada por candidato; registrar descartes.

## 5. Corridas y métricas

- [ ] 5.1 Correr cada candidato sobre la muestra congelada con modelo y esfuerzo fijados según 09.
- [ ] 5.2 Emparejar hallazgos con etiquetas (tercero que no decidió aceptación).
- [ ] 5.3 Calcular precisión, exhaustividad, señal adicional, FP en controles, costo por caso y tiempo de triage.

## 6. Cierre

- [ ] 6.1 Informe con datos crudos, muestra, acuerdo, log de envío y salida (seguir explorando, repetir, descartar).
- [ ] 6.2 Si la salida es seguir explorando: abrir la decisión de arquitectura separada; no integrar en esta ficha.
