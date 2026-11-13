# apply-1 · constructor · tareas 1.1 y 1.4

Change `opsx-baseline` (ficha 02). Rama `propose/opsx-baseline`. Revisión de partida `43eeb7f`; cambios sin commitear (sin commit, sin push, sin build). Clasificación LOW. Fecha 2026-11-11.

## Archivos creados / modificados
- `eval/baseline/README.md` (nombre del directorio `(TBD · renombrable)`; "externo al binario; no importado por cmd/ ni internal/")
- `eval/baseline/manifest/{manifest.go,manifest_test.go,plan_test.go,load_test.go}` (1.1)
- `eval/baseline/review/{review.go,review_test.go}` (1.4)
- `openspec/changes/opsx-baseline/tasks.md`: 1.1 y 1.4 marcadas `[x]` con evidencia; sección final con la suite completa. Sin tocar otras tareas.
- `go.mod` / `go.sum`: sin cambios (stdlib solamente).

## EVIDENCIA
- Baseline: paquetes nuevos, no hay suite previa que cubra los archivos tocados. La suite completa se corrió una vez al final.
- RED por aserción observado antes de cada producción (esqueleto compilable con valor neutro): 1.1 en 10:33:32 (versión), 10:36:26 (permisos, opt-ins, terminal IDE), 10:36:55 (Plan, estados), 10:37:13 (Validate/Load); 1.4 en 10:39:06 (18 pruebas/subpruebas fallan). Salidas en `tasks.md`.
- GREEN: manifest `ok` 10:36:09, 10:36:36, 10:37:02, 10:37:21 (cobertura 95.3 %); review `ok` 10:39:21 y tras refactor (cobertura 100 %).
- Pruebas: manifest 4 funciones de `CheckCell`/`Plan`/estados + `Validate` + `Load` (con subcasos); review 11 funciones con subcasos.
- Mutantes (todos hicieron fallar al menos una prueba; revertidos con comprobación de diff idéntico, nada entregado): manifest M1 denominador, M2 flags, M3 opt-in apagado, M4 terminal integrada, M5 `<`→`<=` en versión; review M1 duplicado→TP, M2 solape exclusivo, M3 mecanismo ignorado, M4 nuevo→FP, M5 archivo ignorado. Incidente: la primera tanda de mutantes de review se corrió sobre un archivo roto por una edición mía (borré una línea equivocada) y no probó nada; se reparó y se repitió completa sobre el archivo válido (resultado arriba).
- `gofmt -l eval`: vacío. `go vet ./eval/...`: ok. `grep "active-stack/eval" cmd internal`: sin coincidencias.
- `go test ./...` (10:40:30–10:43:13) sobre la revisión entregada: todos `ok`, sin fallos previos. `openspec validate opsx-baseline --strict`: válido.
- Lecciones L-NNN: no existe `knowledge-base/` en el repo; ninguna leída.

## Decisiones de modelado (1.1)
- Celda = configuración declarada + `Observed` (lo comprobado con el cliente real: permisos, settings, si la CLI corre en terminal integrada). `CheckCell` devuelve `ejecutable` o `incompatible` con todos los motivos, en orden fijo (versión, permisos, opt-ins, terminal).
- `Plan()` separa `Executable` (denominador) de `Incompatible` (listadas con motivo). `RunStatus` define solo el enum (aceptada, rechazada_por_calidad, bloqueada_por_permisos, timeout, fallo_del_cliente); sin cálculo de resumen (3.1).

## SUPOSICIÓN (no verificado; lectura elegida)
1. **Terminal integrada de IDE (1.1).** El brief la lista entre las causas de `incompatible`. Interpretación: una celda declarada `ide` cuya CLI corre en la terminal integrada queda incompatible con el motivo "la cohorte es terminal"; si se declara `terminal` es ejecutable. Alternativa no elegida: reclasificar automáticamente a cohorte terminal. Confianza media: el spec dice que "el manifiesto registra la interfaz real como terminal", lo que también admite reclasificar.
2. **Opt-ins.** Solo un valor explícitamente apagado (`""`, `0`, `false`, `no`, `off`) los desactiva; cualquier otro valor cuenta como activo (conservador). El spec no define los valores.
3. **Versión.** Comparación numérica por componentes `X.Y.Z`; un sufijo (p. ej. `-beta`) no se interpreta y la celda queda incompatible. No verificado contra los formatos reales de versión de Claude Code y Codex.
4. **Permisos.** Mismo modo y mismo multiconjunto de flags, sin importar el orden. La semántica de flags equivalentes de cada cliente no se verificó (depende de 1.2 con clientes reales).
5. **Rúbrica 1.4, ambigüedades de design.md** (lectura conservadora en cada una):
   - "El mismo mecanismo de falla" lo decide el revisor: se modela como igualdad del texto de mecanismo normalizado (minúsculas, trim); un mecanismo vacío no empareja. Cómo se normaliza lo que el revisor registre queda a definir en 1.3.
   - Un hallazgo que coincide con dos defectos: se ata al primero de la clave. Si ese ya está emparejado es duplicado, no pasa al segundo, que queda como omisión; `Candidates` conserva ambos para el revisor. Resultado dependiente del orden de la clave; la alternativa (emparejamiento máximo) puede contar más verdaderos positivos.
   - Entre varios hallazgos del mismo defecto, el verdadero positivo es el primero de la lista.
   - Un hallazgo sin emparejar sin decisión del revisor no entra en falso positivo ni en hallazgo nuevo: clase extra `pendiente_de_revisor`, fuera de los conteos de TP/FP/omisiones.
   - La evidencia del verdadero positivo se propaga pero no se exige.
   - Solape de rangos inclusivo en los extremos; rutas comparadas tras normalizar `\` y `./`, sensibles a mayúsculas. Entradas inválidas (IDs vacíos o repetidos, rangos invertidos) devuelven error en vez de ignorarse.
6. D2 pide distinguir plataformas distribuidas, declaradas y probadas; el campo no está en el manifiesto porque la tarea 1.1 no lo lista. Pendiente para quien arme la matriz real.
7. Límites (`Limits`) con cero = no fijado: los valores dependen de D3 (TBD).

## Limitaciones
- Todo con datos sintéticos; no hay cliente real (Codex no instalado). 1.2, 1.3, grupos 2, 3 y 4 no tocados.
- No se ejecutó `go build` ni `go install`; no se commiteó.
- Los `go: downloading` de la suite completa son de dependencias de prueba ya declaradas en go.mod; `go.mod`/`go.sum` sin diff.

## Fuentes leídas
`openspec/changes/opsx-baseline/{tasks,design}.md`, `specs/opsx-benchmarking/spec.md`, `informes/LEEME.md`, `CLAUDE.md` del repo (vía contexto), tests existentes de `internal/catalog` para estilo (paquete interno, `t.Run`, tablas).

## Contaminación por informes ajenos
Ninguna: no leí informes de breaker, auditor ni `auditoria-ramas-opsx-2026-11-10.md`. Solo `LEEME.md`, que es el estado de orquestación.

## JUICIO
Autor: constructor (este informe). Comparación: cada comportamiento del spec de "Corridas reproducibles", "Cohortes por interfaz" y "Rúbrica de revisión" mapeado a una prueba que falló por aserción antes del código y a un mutante que la rompe. Confianza: alta en la lógica pura y su sensibilidad; media en las lecturas 1 y 5 (ambigüedades de diseño); baja en el formato real de versión y flags hasta que 1.2 use clientes reales.
