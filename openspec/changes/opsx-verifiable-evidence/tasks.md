# Tasks

> Todas las tareas de implementación quedan **pendientes**: esta ronda es solo de propuesta (corregir la propuesta). Flujo del equipo: **contrato → evidencia → decisión**: se corrige la propuesta, se implementa con `apply`, se verifican los criterios de aceptación y se abre el PR. Cuando llegue la etapa de implementación, TDD estricto: en cada grupo, primero RED (test que falla por la razón correcta), después GREEN mínimo, TRIANGULATE y REFACTOR. Antes de tocar archivos existentes, correr el SAFETY NET del paquete y anotar la línea base. No hacer build ni commitear salvo pedido explícito. No tocar `filemerge`, `backup` ni `pipeline` (ALTO): si surge la necesidad, frenar y reportar. El esquema RED / GREEN / TRIANGULATE / REFACTOR sale de los changes archivados, no de una regla escrita en `CHANGES.md`. Base revisada: `origin/main` `7a60089` (03 mergeada en `main`, citada contra `main`; 04 pendiente, citada contra `origin/propose/codex-native-opsx` `91fbb70`).

## 1. Antes de empezar: acuerdos y confirmaciones (MEDIO)

- [ ] 1.1 Con 03 (Juani) y 04 (Matías): acordar la **propuesta de módulo propio `evidence`** (design D5) con sus ocho cambios (fila 28 de `openspec/matriz-reglas-opsx.md`, tabla acción → módulos con `verify`, excepción a "módulo solo con toggle", D4 y D12 de 03, tarea 2.3 de 03, `design.md:33` y `:43` de 04). No se editan archivos de 03 ni de 04 desde este change: los aplican sus dueños. Si no se acuerda, el contrato queda en `delegation` y se informa el límite.
- [ ] 1.2 Cerrar en el grupo las preguntas abiertas del design (módulo y extracto, mecanismo de captura, quién evalúa la suficiencia, huella del estado, ubicación y versionado, suficiencia de la salida redactada, casos de directorio ajeno y copia vencida).
- [ ] 1.3 Huella y plataformas: (a) fijar en el contrato **sha256 de los bytes con el algoritmo nombrado en el campo** (`{algoritmo, valor}`, como 02 en `eval/baseline/record/record.go:48-51` y 05); no usar `git hash-object` como definición; (b) definir y probar la huella del **estado probado** alineada con `TreeHash(dir)` (`eval/baseline/fixture/load.go:224-243`: sha256 sobre las líneas `<sha256 archivo>  <ruta>` ordenadas), con el conjunto de archivos de la copia de trabajo (preparados, sin preparar y nuevos) y evaluar el árbol de un índice temporal de Git como alternativa (no decidida); recordar `sha256sum --text` en Git Bash; (c) probar en **Windows y Linux** (mínimo decidido), incluso con `core.autocrlf=true`; (d) armar la matriz de plataformas con tres columnas: distribuidas (`.github/workflows/release.yml:71-76`), declaradas (`CLAUDE.md:21`) y probadas. Solo Windows y Linux se exigen como probadas.
- [ ] 1.4 Confirmar que el CLI de OpenSpec tolera la carpeta propuesta para los recibos de ejecución dentro de un change (`openspec status`, `openspec validate`), o definir otra ubicación (TBD).
- [ ] 1.5 Investigar por qué el reporte de `verify-hook` quedó con cero checks en el caso del 02/11 (v0.4.0, Windows, `lite`, `claude`) y reproducirlo sobre `main`. Registrar el resultado y pasarlo a 01, 04 y 08.
- [ ] 1.6 Con 05 (Emiliano): adoptar su formato de referencia, su registro de captura y su redacción al capturar (D9, `e4ec0f5`); ubicación y acceso propuestos por 05, 06 propone aceptar, pendiente de confirmación; vencimiento provisorio (archivo + 7 días, supuesto sin medición de 05); **(TBD)** qué se versiona. El mecanismo de captura independiente por interfaz lo define 05 (matriz D22: solo el arnés es referencia por construcción).
- [ ] 1.7 Con 04 (Matías): confirmar que `verify` en Codex con revisor distinto del constructor, ya cubierto por su spec (`specs/codex-opsx-delivery/spec.md:47-57`) y su `native/SKILL.md` (`:12`, `:16`, `:18`), llega a `main` con su PR. 06 no la agrega. Hasta entonces la suficiencia en Codex queda *sin verificar*.
- [ ] 1.8 Con 07 (Neyen): confirmar que su sobre de retorno (D6, `3fb24ee`) reserva el campo `recibos` para la lista de recibos de ejecución de 06 y que 06 define solo ese contenido.

## 2. Contrato (MEDIO)

- [ ] 2.1 Redactar el esquema del recibo de ejecución (identificador de corrida, árbol y estado exacto, comando, resultado, requisitos, referencia a la copia redactada con una huella `{algoritmo, valor}` y registro de captura) con un ejemplo válido y ejemplos inválidos.
- [ ] 2.2 Definir los controles de integridad, procedencia y suficiencia, los tres estados con el texto de la ficha para *rechazado* (`CHANGES.md:192`) y sus transiciones, incluidas las reglas "una ejecución sin recibos nunca queda verificada" y "sin corroboración independiente o reejecución es sin verificar".
- [ ] 2.3 Redactar la regla de la salida en el recibo de ejecución: referencia a la copia redactada, una sola huella, registro de captura. Ubicación, acceso y vencimiento se remiten a 05 (propuesta de 05, pendiente de confirmación); la redacción no se redacta aquí.
- [ ] 2.4 Redactar el extracto `## Evidence Contract` y el paso de revisión de recibos de ejecución, único para ejecución directa y delegada, incluido el contenido del campo `recibos`.
- [ ] 2.5 Redactar la nota de que los hooks complementan la evidencia y no son un registro exhaustivo.
- [ ] 2.6 **Entregar a 03 (Juani) el texto del `## Evidence Contract`**, que su tarea 2.3 espera (`openspec/changes/archive/2026-11-13-claude-native-opsx/tasks.md:59`: "recibir el `## Evidence Contract` y el sobre de retorno"). Se entrega solo el texto del contrato y del paso de revisión; **no se redefine nada de lo que 03 ya definió** (capas y matriz, mecanismo de módulos G3, módulo obligatorio ausente D11, contratos que consume D12). La ubicación del archivo la decide 03 según el acuerdo de 1.1.

## 3. Assets de Claude y de Codex (MEDIO, checkpoint)

- [ ] 3.1 SAFETY NET: correr `go test ./internal/harness/config/...` sobre la rama que integre 03 y 04, y anotar la línea base. Reconfirmar dónde viven hoy las pruebas de composición (`ComposeLayers`, `ComposeCodexPackage`).
- [ ] 3.2 RED: golden de composición de Codex con el contrato de evidencia (en `main` ya existe `TestCompose_AllToggles_codex` para el asset legado, `compose_test.go:274`, además de `TestCompose_VariantCodex`, que busca la palabra "Codex"). Verificar que falla por la razón correcta.
- [ ] 3.3 RED: prueba de que el contrato se carga cuando la acción requiere evidencia en ejecución directa y delegada (también con el toggle `delegation` apagado), de que su extracto va en el encargo delegado, y de que no aparece como regla universal en `CLAUDE.md` ni `AGENTS.md`, en Claude y en Codex.
- [ ] 3.4 GREEN: agregar el contrato y el paso de revisión de recibos de ejecución al lugar acordado en 1.1, para Claude y para Codex.
- [ ] 3.5 GREEN: conectar el paso de revisión de Codex con la acción `verify` de 04 (1.7). Si el PR de 04 no está en `main`, la parte dependiente queda marcada y *sin verificar*.
- [ ] 3.6 TRIANGULATE: actualizar los golden de Claude y de OpenCode si la composición los afecta, y comprobar que `generic` y los demás agentes no cambian.
- [ ] 3.7 REFACTOR y volver a correr la suite de `internal/harness/config` en verde.

## 4. Fixtures y pruebas de aceptación (MEDIO)

- [ ] 4.1 Fixtures de recibos de ejecución: válido, incompleto, inventado, viejo (estado del árbol distinto), de directorio ajeno.
- [ ] 4.2 Fixtures: comando insuficiente, tests verdes que no cubren el requisito, salida truncada sin original, cero recibos.
- [ ] 4.3 Caso de recibo de ejecución falso con huella correcta: pasa integridad, queda *sin verificar* sin corroboración y *rechazado* si la reejecución o la fuente independiente lo contradicen.
- [ ] 4.4 Caso de mismo `HEAD` con dos diffs distintos (incluido un archivo nuevo sin trackear): las huellas del estado exacto difieren y un recibo de ejecución de uno no sirve para el otro.
- [ ] 4.5 Caso de huella entre plataformas y fines de línea: el mismo archivo (mismos bytes) da el mismo sha256 en Windows y en Linux, con y sin `core.autocrlf`; dos archivos que difieren solo en los fines de línea (LF frente a CRLF) tienen **bytes distintos y por lo tanto dan sha256 distinto**. Comprobar también con `sha256sum --text` / `Get-FileHash -Algorithm SHA256` que el valor del recibo se recomputa por shell.
- [ ] 4.6 Caso de salida redactada con referencia válida: una salida con un token y una ruta del home, capturada con la redacción de 05; el recibo de ejecución guarda una sola huella (la de la copia redactada) y su registro de captura, y la integridad pasa sobre ese archivo.
- [ ] 4.7 Caso de copia vencida: la copia ya no está disponible; el recibo nunca queda *verificado* y su estado es el que se acuerde en 1.2 (propuesta: *sin verificar*).
- [ ] 4.8 Caso de `apply` directo: el coordinador ejecuta las pruebas sin delegar (con `delegation` apagado), se producen recibos de ejecución y pasan por el mismo paso de revisión que en la delegación.
- [ ] 4.9 Caso de `git status`: tras una corrida con salida voluminosa, ninguna salida completa aparece en `git status` (ni sin trackear).
- [ ] 4.10 Correr la misma batería con el flujo de Claude y con el de Codex, en ejecución directa y delegada, y comparar los veredictos.
- [ ] 4.11 Suite Go completa en verde. Confirmar con `git diff --stat` que no se tocaron `backup`, `filemerge` ni `pipeline`.
- [ ] 4.12 Caso de detalle descartado por la redacción: verificar un requisito exige un detalle que la redacción de 05 descartó; el recibo queda *sin verificar* y el fixture comprueba que no se persiste el crudo para evitarlo (requisito de 05 "Un detalle redactado no alcanza para verificar").

## 5. Integración con otras propuestas

- [ ] 5.1 Con 05: confirmar que su formato de referencia contiene lo que el recibo de ejecución necesita (una huella sha256 y registro de captura) y, con un fixture, que la redacción conserva la línea y su posición para comprobar el código de salida.
- [ ] 5.2 Con 07: qué evidencia esperada puede exigir a un worker y cómo se rellena `recibos` en su sobre.
- [ ] 5.3 Con 08: qué campos del recibo de ejecución puede validar un control posterior a la acción y qué rastro de hooks sirve para corroborar (sin tomarlo como registro exhaustivo).

## 6. Cierre del change (flujo: entrega, comprobación, pendientes, traspaso, PR)

- [ ] 6.1 **Qué entrega.** Listar lo entregado (contrato, assets, pruebas) con rutas. Si algún criterio queda sin cumplir o depende de un PR aún sin merge (por ejemplo `verify` de 04), **declarar alcance parcial** de forma explícita.
- [ ] 6.2 **Cómo se comprobó.** Para cada criterio de aceptación de la ficha (recibos inventados o viejos, comandos insuficientes, salidas truncadas sin original, requisitos incumplidos con tests verdes, recibo incompleto sin verificar), registrar comando, resultado y plataforma (Windows y Linux). Comprobar los artefactos con `openspec status --change opsx-verifiable-evidence` y `openspec validate opsx-verifiable-evidence`.
- [ ] 6.3 **Qué queda pendiente.** Listar los `(TBD)` que siguen abiertos (captura independiente por interfaz, ubicación/acceso/vencimiento de la copia redactada, módulo `evidence` si no se acordó, dependencias de 03, 04, 05, 07 y 08).
- [ ] 6.4 **Informe de traspaso.** Escribir `openspec/changes/opsx-verifiable-evidence/informes/traspaso-<fecha>.md` con el modelo de `openspec/changes/codex-native-opsx/informes/traspaso-lifecycle-2026-11-13.md` (`origin/propose/codex-native-opsx`): qué podrá usar el siguiente responsable y coordinación pendiente por change (03, 04, 05, 07, 08).
- [ ] 6.5 **Archivar y abrir el PR.** El PR de 06 se abre **después de archivar el change, con todas las tareas completas y verificadas** (o con el alcance parcial declarado en 6.1). El revisor e integrador es **Juan Cruz**, que comprueba **"contrato y evidencia sobre el último commit"**. Los hallazgos se resuelven en la misma rama (`propose/opsx-verifiable-evidence`, convención de `CHANGES.md:10`).
