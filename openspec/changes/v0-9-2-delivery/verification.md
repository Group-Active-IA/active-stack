# Evidencia de la entrega 0.9.2

- Rama `feat/v0.9.2-delivery` desde `origin/main` en `293deaaf5a664a8a04a7fefb2ae2147c096bd224`.
- Juan Cruz Robledo (`JuanCruzRobledo`) es el revisor independiente confirmado para esta entrega. La 0.9.2 es la última ficha del roadmap al momento de acordar la revisión.
- [PR de Active Orchestrator](https://github.com/Group-Active-IA/active-orchestrator/pull/9) en `2b5ab6500b2666216f26ee3e7f03a3f6056bb71f` con 77 pruebas locales.
- [PR de Roadmap Generator](https://github.com/Group-Active-IA/roadmap-generator/pull/4) en `6c26d6262f85e0404ac7c26e25927692c84537c7` con 63 pruebas locales.
- Los PR externos requieren integración por los mantenedores. Después, verificar ambos merges en sus ramas remotas, fijar sus SHAs en el catálogo y ejecutar las pruebas de instalación antes de solicitar revisión final.
- OpenSpec en modo estricto, 82 pruebas de Active Delivery y 45 de Active Roadmap pasaron.
- Las pruebas de Go de `assets`, `internal/catalog`, `internal/install`, `internal/harness/config` y `eval/taskcontext/inventory` pasaron.
- Las seis pruebas generales que fallaron por desinstalación, tema Windows, descarga Homebrew y huella del fixture de evaluación también fallaron en un checkout limpio de la base. No se hizo una prueba de escritura en una lista real de ClickUp.
- Tras integrar el PR, conservar la evidencia del merge y las métricas finales con el cambio, y archivar con OpenSpec. Como el registro completo exige el merge, el archivo se publica con un PR posterior.

## Medición parcial hasta el 25 de noviembre de 2026

Lectura local de los contadores acumulados del cliente durante el desarrollo.
La fila principal llega hasta las 10:13:32 de Mendoza. Las otras tres son los
últimos contadores visibles de cada ejecución delegada. La conciliación entre
hilos y el tiempo activo final quedan pendientes hasta que cierre el PR.

| Ejecución | Modelo | Entrada sin caché | Lectura de caché | Salida |
| --- | --- | ---: | ---: | ---: |
| Principal | gpt-6.2-sol | 145.429 | 6.184.704 | 17.212 |
| Revisión por secuencia | gpt-6.2-sol | 48.700 | 2.267.136 | 8.769 |
| Contabilidad | gpt-6.2-sol | 59.614 | 2.494.848 | 12.200 |
| Fundación y generador | gpt-6.2-sol | 48.930 | 2.657.152 | 10.029 |

Para referencia, la tarifa API estándar publicada para el modelo es USD 2 por
millón de entrada, USD 0,10 por millón de lecturas de caché y USD 10 por millón
de salida. La escritura de caché observada fue cero en las cuatro ejecuciones.
[Tarifas de OpenAI, consultadas el 25 de noviembre de 2026](https://developers.openai.com/api/docs/pricing)
