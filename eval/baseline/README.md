# eval/baseline

Arnés de medición de referencia del change `opsx-baseline` (ficha 02).

- **Externo al binario; no importado por `cmd/` ni `internal/`.** Decisión D6: directorio de evaluación separado, en Go, solo biblioteca estándar, sin tocar `go.mod`.
- **Nombre del directorio: `eval/baseline/` (TBD · renombrable).** No está decidido; renombrarlo solo exige mover la carpeta, porque nada fuera de `eval/` lo importa.

## Paquetes

| Paquete | Tarea | Contenido |
|---|---|---|
| `manifest` | 1.1 | Manifiesto versionado (JSON), matriz de cohortes con tareas y pesos, validación previa por celda (`ejecutable` / `incompatible` con motivo), plan con celdas fuera del denominador y enum de estados de corrida. |
| `review` | 1.4 | Regla de emparejamiento de la revisión de un diff sembrado (verdadero positivo, duplicado, falso positivo, hallazgo nuevo a adjudicación, omisiones). |

## Pruebas

```
go test ./eval/baseline/...
```

Los datos de los tests son sintéticos. Esta tanda no incluye adaptadores de captura ni cálculo de resúmenes (tareas 2.x y 3.1).
