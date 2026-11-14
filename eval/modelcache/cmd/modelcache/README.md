# CLI local de modelcache

Tres operaciones offline. Lee JSON y archivos de layout; escribe solo el resultado en stdout y errores en stderr. No ejecuta clientes ni comandos externos, no accede a red, no escribe configuración y no habilita adopción.

Desde la raíz del repositorio, estos comandos son ejemplos para ejecutar manualmente. La implementación se prueba invocando `run` dentro de Go tests, sin generar un binario de producto.

```sh
go run ./eval/modelcache/cmd/modelcache validate --input eval/modelcache/testdata/synthetic/dataset.json --root eval/modelcache/testdata/synthetic
go run ./eval/modelcache/cmd/modelcache report --input eval/modelcache/testdata/synthetic/dataset.json --root eval/modelcache/testdata/synthetic
go run ./eval/modelcache/cmd/modelcache layout --input eval/modelcache/examples/codex-repo-layout.json --root .
```

`--input -` (predeterminado) lee stdin; `--root .` es la raíz predeterminada para archivos de layout. `validate` y `report` reciben un `Dataset` (`protocol`, `runs`, `preparation`). `layout` recibe `{"source_scope":"repo-source","paths":["ruta/relativa"]}` y devuelve un snapshot con hashes. Las rutas usan `/`, son relativas a `--root` y no pueden escapar mediante `..`, rutas absolutas o symlinks.

`validate` comprueba protocolo y bytes del layout local: no acredita ejecución, carga real del cliente, calidad o procedencia. `report` analiza datos declarados; no los captura ni los certifica. La fase admitida es calibración y adopción permanece `unavailable`, incluso si la aritmética condicional muestra una mejora. El fixture incluido es **sintético**: modelos, métricas y referencias no son costos, cuotas o corridas reales; sus decisiones están pendientes.

`layout` conserva el orden de rutas solicitado. El ejemplo de repositorio inventaría fuentes designadas: no representa todo el paquete compuesto por toggles ni el orden real del prompt ensamblado por el cliente. Para comparar brazos deben fijarse explícitamente alcance, archivos y orden; un hash no prueba equivalencia semántica ni carga efectiva.

El JSON es estricto: rechaza campos desconocidos y exige un solo valor, sin otro documento o basura al final. stdout es JSON determinista para las mismas entradas, salvo fallo de escritura. Códigos de salida:

| Código | Significado |
|---|---|
| 0 | Operación local completada; no implica aceptación del piloto. |
| 1 | Error de lectura, layout/protocolo/análisis o escritura de resultado. |
| 2 | Comando/flags/argumentos inválidos o JSON malformado, desconocido o con contenido adicional. |

Pruebas aisladas:

```sh
go test -count=1 ./eval/modelcache/cmd/modelcache
```
