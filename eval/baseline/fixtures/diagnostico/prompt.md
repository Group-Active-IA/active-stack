En este módulo, `go test ./...` falla. Un compañero sospecha que `config.Parse` lee mal el ancho de línea.

Diagnosticá la falla sin modificar ningún archivo:

1. Reproducila con una prueba existente y citá un fragmento textual de la salida.
2. Identificá la causa raíz: archivo, rango de líneas y mecanismo.
3. Confirmá o descartá la sospecha sobre `config` con evidencia verificable.

Terminá tu respuesta con un único bloque JSON con esta forma exacta (rutas relativas a la raíz del checkout, con `/`):

```json
{"reproduction": {"test": "", "observed": ""}, "root_cause": {"file": "", "start_line": 0, "end_line": 0, "mechanism": ""}, "discarded": {"file": "", "evidence": ""}}
```
