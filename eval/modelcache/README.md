# Calibración offline · pilotos09

Este paquete consume APIs públicas del baseline02. Valida y analiza **datos declarados** de una campaña; no ejecuta clientes, modelos ni llamadas pagas, no cambia routing y no acredita ejecución ni procedencia independiente. `adoption.status` siempre es `unavailable`.

Desde la raíz del repositorio, los comandos siguientes usan `go run` para ejecutar únicamente esta utilidad offline. No son un build del instalador:

```sh
go run ./eval/modelcache/cmd/modelcache validate --input eval/modelcache/testdata/synthetic/dataset.json --root eval/modelcache/testdata/synthetic
go run ./eval/modelcache/cmd/modelcache report --input eval/modelcache/testdata/synthetic/dataset.json --root eval/modelcache/testdata/synthetic
go run ./eval/modelcache/cmd/modelcache layout --input eval/modelcache/examples/codex-repo-layout.json --root .
```

La fixture es enteramente sintética: A cambia un modelo ficticio; B reordena dos archivos iguales y conserva una lectura parcial caliente. Sus cantidades, IDs y referencias no son resultados de los clientes ni elecciones del equipo. Decisiones `null` preservan D1–D5 pendientes y generan contrastes inconclusos.

El ejemplo de layout Codex enumera seis fuentes reales de main `d5e95bd`: SKILL, tabla de acciones y cuatro módulos de contexto. El orden es el designado para ese snapshot, **no** el orden de carga del cliente ni la composición completa de toggles. Para evaluar un prompt materializado hay que capturar sus archivos reales y declarar `materialized-prompt`; esa declaración tampoco prueba que el cliente los haya leído. Los hashes se calculan sobre bytes, por lo que LF y CRLF difieren. Ningún hash de este paquete es un TreeHash06 ni un recibo.

Cada Dataset corresponde a una campaña. Los modelos/esfuerzos efectivos, celdas, condiciones, recibos de señales y revisión semántica se declaran en los datos y requieren corroboración externa para experimentos reales. Los reintentos son elementos de `attempts`; una repetición duplicada no es un nuevo intento. `preparation: null` significa captura desconocida; `[]` declara que no hubo preparación imputable. Cada evento de costo tiene `source_id` único y cada preparación un ID y bloque únicos; no se reutiliza como costo de otro intento.

Las medidas ausentes son desconocidas, nunca cero. Los costos observados y estimados, API y suscripción y unidades de cuota se mantienen separados. Si esa separación fragmenta la cobertura de una misma población, los subtotales disponibles permanecen visibles pero no aparecen ratios completos de una cohorte parcial. Los agregados que incluyen workers no vuelven a sumarlos. Los tiempos son sumas de duraciones declaradas, no tiempo calendario ni inferencias de costo.

En B sólo se admite reordenar el mismo conjunto de archivos y bytes, con referencia declarada de revisión semántica. Se compara frío con frío y caliente con caliente. Una sesión nueva o los tokens cached globales no prueban atribución al prefijo; contaminación fría, aislamiento ausente o señal desconocida impiden atribuir una mejora. Los misses y las lecturas parciales calientes genuinas permanecen en el análisis asignado. La ventana caliente efectiva exige calibración externa.

`candidate_for_review` expresa aritmética condicional con parámetros declarados: no es una recomendación verificada. No hay modelos, muestra, presupuesto, objetivo ni umbrales predeterminados. Faltan acuerdos D1–D5, referencia estable02, A0/resultados07, captura y evaluación independientes06, y revisión humana antes de confirmar o adoptar. No se importan implementaciones de otras ramas.

Las funciones `CompareArms` y `ObserveCache` son comprobaciones locales; el flujo de reporte ejecuta primero `Protocol.Validate`. El CLI valida además los archivos congelados bajo `--root`, rechaza JSON desconocido/trailing y escapes de raíz. Ver [uso del CLI](cmd/modelcache/README.md) y [el informe de apply](../../openspec/changes/opsx-model-cache-pilots/informes/apply-offline-09-2026-11-13.md).
