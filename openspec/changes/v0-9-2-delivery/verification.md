# Evidencia de la entrega 0.9.2

- Rama `feat/v0.9.2-delivery` desde `origin/main` en `293deaaf5a664a8a04a7fefb2ae2147c096bd224`.
- Juan Cruz Robledo (`JuanCruzRobledo`) es el revisor independiente confirmado para esta entrega. La 0.9.2 es la última ficha del roadmap al momento de acordar la revisión.
- [PR de Active Orchestrator](https://github.com/Group-Active-IA/active-orchestrator/pull/9) integrado en `main` como `a3d6b0490940a2865beb16d3baf07a5a292eeafb`. Pasaron 77 pruebas sobre ese commit.
- [PR de Roadmap Generator](https://github.com/Group-Active-IA/roadmap-generator/pull/4) integrado en `main` como `ac1c0ef79671e990aff0c50d237e10448c95235b`. Pasaron 63 pruebas sobre ese commit.
- Los dos SHA coinciden con los merges de GitHub y con los tips remotos de `main` consultados. El catálogo los fija juntos en el PR #92. La fundación exige el subcomando `statuses` que incorpora ese mismo PR. La publicación se hace después de integrar el conjunto.
- El comando `roadmap.py statuses` aceptó las seis etiquetas y reportó los nombres faltantes con código 1 para una lista incompleta.
- OpenSpec en modo estricto, 82 pruebas de Active Delivery y 45 de Active Roadmap pasaron.
- Las pruebas de Go de `assets`, `internal/catalog`, `internal/install`, `internal/harness/config` y `eval/taskcontext/inventory` pasaron.
- La repetición de `go test ./...` aprobó 58 paquetes y reprodujo los mismos seis fallos de la base limpia en desinstalación, tema Windows, descarga Homebrew y huella del fixture de evaluación. No se hizo una prueba de escritura en una lista real de ClickUp.
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

## Publicación de v0.9.2 (26 de noviembre de 2026)

- Antes de publicar se corrigió que `update` no reemplazaba el binario en macOS ([PR #93](https://github.com/Group-Active-IA/active-stack/pull/93), merge `c41c387`). Los tests nuevos fallaban en darwin sin el arreglo.
- El tag `v0.9.2` apunta a `c41c387`. El [workflow de release](https://github.com/Group-Active-IA/active-stack/actions/runs/45818874240) terminó correctamente y la [release](https://github.com/Group-Active-IA/active-stack/releases/tag/v0.9.2) quedó como Latest con el instalador de Windows (dos nombres), los binarios de Windows, Linux y macOS (amd64 y arm64) y `SHA256SUMS.txt`.
- El binario v0.9.1 instalado en Windows detectó la versión nueva con `update --check`. No se descargaron ni ejecutaron los assets, y nada se probó en una Mac real (PP-20).
- El cambio sigue sin archivar: faltan las métricas finales de tiempo, tokens y costo.
