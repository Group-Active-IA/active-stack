<!-- Transcripción literal del informe devuelto por el subagente role-auditor (Opus), que no tiene herramienta de escritura; guardado por el orquestador el 2026-11-11 sin modificar el contenido. -->

# verify-t2-1 · auditor · opsx-baseline (1.3 + DEUDA 6)

Rol: auditor, LOW. Rama `propose/opsx-baseline`, base `5dcfc68`, diff sin commitear. Revisé la revisión que identifican los hashes de la sección 1. Fecha: 2026-11-11, de 14:43 a 14:50 (-03). No edité nada y no hice commits. Los mutantes que corrí son míos y los hice sobre una copia en el scratchpad (`.../scratchpad/aud`), nunca en el árbol del repo.

## VEREDICTO: **APROBADO**
No hay bloqueantes. No hay fallos ALTA ni MEDIA. La DEUDA queda abajo, con un responsable para cada punto.

## Fuentes
- **Plan:** proposal.md, design.md (por grep: D1, D10 y riesgos), tasks.md completo y spec.md (por grep: requisitos y escenarios de rúbrica, fixture inmutable y repeticiones).
- **Estado:** informes/LEEME.md completo.
- **Informes originales completos:** apply-t2-1-constructor.md y verify-t2-1-breaker.md. De verify-1-auditor.md leí la sección de DEUDA, las lecciones, las limitaciones y la derivación.
- **Código:** el diff de `eval/baseline/review` contra `5dcfc68`, el comando de huella del README y `fixtures/cambio-pequeno/task.json`. El código de `fixture/*.go` no lo leí completo: me apoyo en sus tests y en el informe del breaker. Lo declaro como limitación.
- **Lecciones:** no hay `knowledge-base/` ni archivos L-NNN.
- **Engram:** no lo usé.
- **Archive:** no lo consulté. El de codex-strict-compat es ajeno al change.
- **Contaminación:** leí los informes del constructor y del breaker antes de mis chequeos. Agente, constructor y breaker son todos Claude; el breaker corrió en Sonnet y yo en Opus. No hay independencia de proveedor.

## EVIDENCIA (propia, de esta sesión)
1. **Hashes (`git hash-object`, 14:43).**
   - Coinciden con el informe del constructor: README, review.go, apply2_test.go, t2_test.go, los cuatro `.go` de producción de fixture, checkout_test, decide_test, evaluate_test, los cinco task.json y seeded.patch.
   - Excepción: `fixture/load_test.go` da `c5b635cc` y el informe dice `c5b635c1`.
   - El mtime de ese archivo (14:31:07) es anterior a la suite del constructor (14:31:35). Todos los archivos de `fixture/` son anteriores a 14:31:35. Por eso lo leo como un error de transcripción y no como un cambio posterior, pero esto es SUPOSICIÓN.
   - Conclusión: la suite del constructor y la corrida del breaker siguen vigentes para esta revisión.
2. **Conteo de fixtures.** `git ls-files -o eval/baseline/fixtures` da 41. Confirma la observación del breaker: el "49" del constructor es un error.
3. **Aislamiento y dependencias.**
   - `go list -deps ./cmd/... ./internal/...`: 0 coincidencias de `/eval`.
   - `go list ./...`: 0 paquetes bajo `fixtures`.
   - `git diff 5dcfc68 -- go.mod go.sum cmd internal assets`: vacío.
   - `gofmt -l eval`: vacío.
   - No se tocó código crítico (backup, filemerge, pipeline) ni nada fuera de `eval/` y del change.
4. **Tests propios.**
   - `go test -count=1 -v` de review y manifest: ok, 107 `--- PASS` contando subtests.
   - Corrida enfocada de fixture (`TestDecide|TestLoad|TestEvaluate_Exploration|TestEvaluate_Review`): ok en 32.5 s, 8 tests de primer nivel, todos PASS. Incluye la propiedad exhaustiva de `Decide`.
   - No repetí el paquete fixture completo ni `./...`: reutilizo las corridas del constructor (30 paquetes) y del breaker (146), que siguen vigentes por el punto 1.
5. **Mutantes propios sobre la DEUDA 6** (en la copia del scratchpad):
   - aceptar `"pending"` como disposición → `--- FAIL: TestMatch_UnknownDispositionIsError`;
   - desactivar la guarda de archivo vacío en la clave → `--- FAIL: TestMatch_KeyDefectWithBlankFileOrMechanismIsError`.
   - Sin mutar, la copia da ok.
6. **El diff de DEUDA 6 se limita a dos guardas en `Match`** (archivo o mecanismo en blanco en la clave; disposición fuera de {"", rejected, real_defect}) y a la adaptación del test F7. No reescribe la lógica de emparejamiento.
7. **F3, por lectura.** El README, línea 29, usa `xargs sha256sum` sin `--text`, lo que coincide con lo que reprodujo el breaker. No ejecuté su repro.

## JUICIO
- **Cobertura de 1.3.**
  - Hay cinco familias con prompt (huella sha256), rúbrica de criterios obligatorios, referencia y variante defectuosa. Cada familia acepta la referencia y rechaza la defectuosa con el criterio esperado.
  - Hay tests de reinicio y restauración entre repeticiones.
  - El escenario del spec "Rechazo por calidad" (la prueba pasa pero un obligatorio queda sin evidencia) está cubierto por cambio-pequeno: S1 se cumple y S3 no.
  - El escenario "Repeticiones del mismo caso" está cubierto por las huellas idénticas entre repeticiones.
  - El RED por comportamiento está declarado con hora y mensaje de aserción. Donde el RED fue por setup (algunos subcasos de Load), se acreditó por mutante, y así quedó declarado.
  - Que dos mutantes sobrevivieran y se endurecieran las aserciones está bien documentado. Lo acepto.
- **Decisión 1 (fixtures sintéticos frente a D1).**
  - D1 registra "un fixture por familia sobre una revisión fija del repositorio de Active Stack" y "los fixtures concretos se nombran al implementar la tarea 1.3, congelados antes de las corridas".
  - Exploración y revisión cumplen D1 literalmente (`1b4639b`). Cambio pequeño, multiarchivo y diagnóstico usan mini-proyectos versionados dentro del repo de Active Stack. Cuando se commiteen, quedan en una revisión fija de ese repo, congelados además por la huella del árbol.
  - Clasificación: **interpretación válida, pero no literal, que necesita ratificación**. No contradice un acuerdo cerrado: D1 está "Parcial" y deja abiertos los fixtures concretos hasta antes de las corridas. Además, el registro de riesgos del design pide "fixtures sintéticos". Sí se aparta de la lectura natural de D1, que es que la tarea se hace sobre código de Active Stack, y afecta la representatividad de las familias de edición.
  - No bloquea porque todavía no empezaron corridas. Queda como DEUDA 1.
- **Decisiones 2 a 6 del constructor.**
  - 3 (el rechazo domina) y 4 (disposiciones exactas) son conservadoras y coherentes con la DEUDA 6. Las acepto.
  - 5 (el mecanismo del revisor reemplaza al del agente) y 6 (un hallazgo inválido del agente incumple sin error; un revisor inválido es error) son razonables y quedan visibles.
  - 2 (R3 exige cero FP y nada sin decidir) es más estricta que la rúbrica de conteos. Toca la rúbrica de 1.4, que tiene revisión pendiente por 10. Va con la DEUDA 2 de la tanda 1.
  - Ninguna contradice una decisión fechada.
- **DEUDA 6 de la tanda 1:** cerrada. Hay RED con hora, mutantes del constructor y mutantes míos.
- **F1 a F3 del breaker.** Coincido en que son BAJA: un directorio vacío no altera el comportamiento; el fragmento trivial ya figuraba como suposición 8; F3 es documentación solo en Windows. Las sospechas S1 a S5 quedan como sospechas; no las reproduje.
- **Regla BAJA → DEUDA.** El LEEME registra la decisión del responsable como "vigente desde la tanda 3" y dice que "la tanda 2 ya estaba en curso con la configuración anterior". El brief la aplica a la tanda 2. No encontré un registro fechado que la extienda a esta tanda. No bloqueo, porque los tres fallos son BAJA y en la tanda 1 también se aceptó deuda baja en la auditoría. Aun así, ver DEUDA 4.
- **Cambios ajenos.** Hay un move de `codex-strict-compat` a `archive/` (sin fecha en el nombre del directorio). No afectan al código de 02, pero no deben entrar en su commit.

## DEUDA (no bloquea)
1. **Ratificar la decisión 1 frente a D1.** El responsable de 02 o el grupo confirma, con fecha en el registro de decisiones, si se aceptan familias de edición sintéticas o si se exige un código de Active Stack. Hay que hacerlo antes de 3.2 o 3.4. Responsable: responsable de 02 o planificador.
2. **F1, S1 y S3 del breaker** (directorios vacíos y symlinks en checkout intacto; rutas sin guarda). Responsable: constructor, en 2.1.
3. **F2: fragmento de reproducción trivial.** Primero la rúbrica (responsable de 02), luego el constructor. **F3:** agregar `--text` al README. **Corregir en el informe del constructor** "49" por 41 y el hash `c5b635c1` por `c5b635cc`. Responsable: constructor, en la próxima tanda que toque esos archivos.
4. **Cobertura de la regla BAJA → DEUDA en la tanda 2.** El orquestador confirma con el responsable y registra en el LEEME que la regla cubre la tanda 2, o lo aclara antes del commit. También tiene que actualizar la tabla "Estado actual" del LEEME, que sigue describiendo la tanda 1 (alcance 1.1/1.4, 1611 líneas, check de 11:09). Responsable: orquestador.
5. **Commit de 02.** Excluir los deletes y el move de `codex-strict-compat`. Responsable: orquestador.
6. **Duración del paquete fixture** (2 a 3 min). Ya está declarada por el constructor. Responsable: 2.1.
7. **Decisión 2 (R3 estricto).** Se suma a la DEUDA 2 de la tanda 1, la revisión de la rúbrica por 10.

## LECCIONES (candidatas)
- Cuando un informe lista hashes como evidencia de revisión, conviene que sean completos o que salgan de un comando pegado tal cual, no transcriptos a mano. Un error de un carácter obliga a inferir vigencia por mtime.
- Cuando una tarea "nombra fixtures" bajo una decisión que está Parcial, conviene que el constructor deje la desviación como pregunta en el registro de decisiones (fila o nota fechada), y no solo como suposición en su informe.

## Limitaciones y chequeos no ejecutados
- No reejecuté el paquete fixture completo, `./...` ni `openspec validate`. Tampoco el repro de F1 a F3 ni las pruebas de autocrlf o índice del breaker.
- No leí completo el código de `fixture/*.go`.
- No hay clientes reales, Linux ni macOS.
- No hubo proveedor de identidad externo ni aprobación HIGH o CRITICAL, así que nada de eso aplica.

## Derivación e intentos
- Es el intento 1 de la auditoría de la tanda 2. No hay correcciones pendientes que invaliden verificaciones aguas abajo.
- Responsables de la DEUDA:
  - responsable de 02 o planificador: 1, 3 (rúbrica de F2) y 7;
  - constructor: 2, 3 y 6;
  - orquestador: 4 y 5.
- APROBADO habilita el cierre documental de la tanda. No autoriza commit, push ni merge sin pedido explícito del responsable.
