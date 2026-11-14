# Tasks · Pilotos de modelos y caché

Entrega incremental offline. Los parámetros reales D1–D5 siguen `(TBD)`; no hay corridas pagas, adopción ni cambios de routing. Una campaña por Dataset; otras campañas usan archivos separados.

## 1. Preparación común


- [x] 1.1 Implementar protocolo de calibración con pins de 02/07, fixtures/rúbrica, brazos, controles y decisiones opcionales sin defaults; rechazar confirmación sin integración independiente.

- [x] 1.2 Congelar archivos designados por bytes SHA256 y orden explícito, separando fuente de repositorio y prompt materializado; validar existencia e integridad offline.

- [ ] 1.3 Acordar D1–D5 propios: alcance/matriz, variantes, controles y horizonte de caché, métrica/cobertura, muestra/presupuesto/regla de parada y responsables de revisión. Las entregas de 02/06/07 son dependencias externas registradas en el cruce, no tareas de implementación de Matías.

## 2. Piloto A


- [x] 2.1 Validar contrastes aislados de modelo o esfuerzo por rol, parámetros efectivos declarados, fallback y controles constantes con datos sintéticos.

- [ ] 2.2 Preparar perfiles reales, calibrar parámetros efectivos por superficie y ejecutar la matriz acordada, con evidencia independiente.

## 3. Piloto B


- [x] 3.1 Validar reordenamiento del mismo conjunto de archivos/bytes con referencia de revisión semántica; conservar condición asignada, señal observada, contaminación, misses y lecturas parciales.

- [ ] 3.2 Calibrar aislamiento, ventana y atribución reales; revisar equivalencia semántica y ejecutar los bloques acordados.

## 4. Resultados y decisión


- [x] 4.1 Reportar costos de todos los intentos y preparación única, alcances de agregados, cobertura, modalidad/unidades y denominador aceptado; mantener desconocido distinto de cero.

- [x] 4.2 Calcular contrastes condicionales sólo con pares comparables y cobertura completa, sin valores predeterminados; bloquear conclusión favorable por regresión obligatoria o condición de caché no confirmable; adopción siempre no disponible.

- [ ] 4.3 Ejecutar y revisar resultados reales, dispersión, suficiencia/evidencia, umbrales y decisión humana por cohorte. Publicar resultados depurados.

## 5. Entrega utilizable


- [x] 5.1 Entregar CLI offline validate/report/layout, fixture enteramente sintética y guía reproducible.

- [x] 5.2 Completar pruebas, race/vet, validación OpenSpec estricta e informe de implementación/revisión con límites.

## Trazabilidad de las ocho tareas originales

| Original | Entrega offline | Trabajo real preservado |
|---|---|---|
| 1.1 manifiesto y acuerdos | 1.1–1.2 | 1.3 |
| 1.2 registro y controles | 2.1, 3.1, 4.1 | 2.2, 3.2 |
| 2.1 perfiles A | 2.1 (simulación solamente) | 2.2 |
| 2.2 corridas A | controles 2.1 | 2.2 |
| 3.1 prefijo/aislamiento | 3.1 (contrato sintético) | 3.2 |
| 3.2 bloques B | registro 3.1, 4.1 | 3.2 |
| 4.1 reportes | 4.1 | validación con capturas reales 4.3 |
| 4.2 resultados/adopción | 4.2 | 4.3 |

Las dependencias externas están identificadas en [el cruce](informes/estado-cruce-y-dependencias-09-2026-11-13.md). No se consideran implementadas ni aceptadas por el antecedente de deuda del cambio04. Los acuerdos propios D1–D5 permanecen abiertos.

Evidencia de apply/TDD y verificaciones: [informe offline09](informes/apply-offline-09-2026-11-13.md). Revisión independiente documentada por separado.
