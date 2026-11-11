# verify-1 · breaker · opsx-baseline (tareas 1.1 y 1.4)

Rol: breaker. Intento 1. Clasificación LOW. Sin edición de código ni artefactos, sin commits, sin cambio de rama.

## Revisión revisada

- HEAD 43eeb7f6fd98e5183d6313e48c6a3dc6998a49b7, rama propose/opsx-baseline, 2026-11-11 10:45:02 -03.
- git status --short: " M openspec/changes/opsx-baseline/tasks.md", "?? auditoria-ramas-opsx-2026-11-10.md", "?? eval/", "?? openspec/changes/opsx-baseline/informes/".
- git hash-object (idénticos al inicio y al final, 10:48:30; la revisión no cambió durante el trabajo):
  - a8b1dfb2 eval/baseline/README.md
  - 122228ff eval/baseline/manifest/load_test.go
  - ba906845 eval/baseline/manifest/manifest.go
  - b6f179cd eval/baseline/manifest/manifest_test.go
  - 02b6ad46 eval/baseline/manifest/plan_test.go
  - ca90e0f8 eval/baseline/review/review.go
  - 21834153 eval/baseline/review/review_test.go
  - d083a60a openspec/changes/opsx-baseline/tasks.md
- Lecciones L-NNN: no existe knowledge-base/ en C:\tools\active-stack ("No such file or directory"). Ninguna lección leída.

## Pruebas ejecutadas por mí (no heredadas)

- go test -count=1 -v ./eval/baseline/... (10:45:54): ambos paquetes ok; 58 "=== RUN" = 20 funciones top-level (load 2, manifest 4, plan 3, review 11) + 38 subtests.
- Tests adversariales temporales eval/baseline/manifest/zz_breaker_test.go (10 funciones) y eval/baseline/review/zz_breaker_test.go (3 funciones): 13 tests propios. Borrados al terminar (10:48:30; git status y hashes verificados después).
- 12 mutantes sobre copias aisladas en C:\Users\espej\AppData\Local\Temp\claude\8a00efe90647aaab\59b3ace0-deab-4305-9f26-437b91b11f36\scratchpad\breaker-02\m (módulo descartable), cada uno corriendo la suite existente del paquete (10:47:19).
- No repetí go test ./...; reutilizo el resultado del constructor (10:40-10:43, misma revisión, archivos sin cambios). No corrí openspec validate, gofmt ni go vet (reutilizados).

## Lente 1 · Correctitud

### F1 (MEDIA) · El emparejamiento depende del orden de la clave y no es uno a uno máximo — EVIDENCIA

Clave A = x.go:10-20 "nil deref", B = x.go:15-25 "nil deref"; f1 = x.go:10-12 (solo A), f2 = x.go:16-18 (A y B).

    go test -count=1 -v -run Breaker ./eval/baseline/review/   (10:46:57)
    clave [A,B]: {TruePositives:1 Duplicates:1 ... Omissions:1} omisiones=[B]
    clave [B,A]: {TruePositives:2 Duplicates:0 ... Omissions:0} omisiones=[]
    --- FAIL: TestBreaker_KeyOrderChangesCounts

Spec (Rúbrica de revisión de un diff sembrado): un hallazgo empareja cuando coinciden archivo, rango solapado y mecanismo, "con emparejamiento uno a uno". f2 cumple las tres condiciones con B y existe la asignación f1-A, f2-B; el código marca f2 duplicado de A y cuenta B como omisión. El design dice que no suman los hallazgos repetidos sobre el mismo defecto; f2 no es repetición sobre B. Los conteos cambian según el orden de la clave. JUICIO (breaker, confianza alta): el supuesto "se ata al primero de la clave" no es neutral; sesga TP y omisiones y rompe la reproducibilidad entre revisores. Sugerencia: emparejamiento bipartito máximo determinista, o probar los demás candidatos libres, con test de invariancia al orden.

### F2 (MEDIA) · Celda sin cliente, interfaz, modalidad o sistema, u observación vacía, queda ejecutable y entra al denominador — EVIDENCIA

    go test -count=1 -v -run Breaker ./eval/baseline/manifest/   (10:46:38)
    interface=desktop -> ejecutable, want incompatible
    interface=empty / client=empty / client=gemini / modality=empty / os=empty -> ejecutable
    celda sin client/os/interface/modality/permisos/observed -> Load ok, Plan={Executable:[a] Incompatible:[]}
    --- FAIL: TestBreaker_UnknownOrMissingDimensions, TestBreaker_LoadMissingCellFields

El spec exige que el manifiesto identifique cliente, sistema, interfaz real, modalidad y permisos efectivos, deja fuera el escritorio y marca incompatible la celda cuya interfaz o configuración no cumple. Ni Validate ni CheckCell validan los enums: {"version":"1.0","min_version":"1.0"} sin nada más suma al denominador. Además permisos declarados vacíos y observados vacíos (nada observado) dan ejecutable (TestBreaker_EmptyObservedPasses), contra "permisos efectivos comprobados con el cliente real". JUICIO confianza alta.

### F3 (MEDIA) · Los opt-ins de contenido se detectan con una lista cerrada de cuatro nombres exactos — EVIDENCIA

    opt-in "OTEL_LOG_TOOL_CONTENT"=1 -> ejecutable, want incompatible
    opt-in "otel_log_user_prompts"=1 -> ejecutable
    opt-in "otel.log_user_prompt"=1 -> ejecutable
    opt-in "OTEL_LOG_USER_PROMPTS "=1 -> ejecutable
    --- FAIL: TestBreaker_OptInFamily

design.md (Datos sensibles): "las variables OTEL_LOG_* de Claude Code o las opciones equivalentes de [otel] en Codex". El código no trata OTEL_LOG_* por prefijo, distingue mayúsculas en el nombre (Windows no lo hace con variables de entorno) y para Codex solo reconoce la clave plana log_user_prompt, no la forma de tabla otel.*. SUPOSICIÓN no verificada: que OTEL_LOG_TOOL_CONTENT exista en el cliente real; el contrato del diseño es la familia, no la lista. Valores: FALSE y " off " cuentan como apagados; disabled, None, null y 0.0 cuentan como activos (falla cerrada, coherente con el supuesto declarado; lo juzgo aceptable).

### F4 (BAJA) · Load acepta un segundo documento JSON concatenado — EVIDENCIA

TestBreaker_LoadTrailingGarbage: {manifiesto válido} {"protocol_version":"OTRO"} da err == nil. Decode lee solo el primer valor; falta comprobar que el resto sea EOF.

### F5 (BAJA) · Validate acepta task_id duplicado y vacío dentro de una cohorte — EVIDENCIA

TestBreaker_DuplicateTaskIDsAndEmptyTaskID: tareas t(1), t(2), ""(1) dan Validate() == nil. Los pesos por tarea quedan ambiguos para 3.1. Plan() no llama a Validate y con IDs repetidos devuelve Executable:[a a] (TestBreaker_PlanWithoutValidate); solo alcanzable si un consumidor arma el Manifest sin Load (SUPOSICIÓN de uso futuro).

### Observaciones sin defecto

- Versiones: v2.1.310, 2.1.310-beta, " 2.1.310", 2..1 y min_version vacío dan incompatible con motivo (falla cerrada; ver S3). 2.1 frente a 2.1.0 da ejecutable (correcto).
- Review: clave o hallazgos vacíos o nil dan conteos cero sin error; doble espacio interno en el mecanismo no empareja; mayúsculas distintas en la ruta no emparejan; disposición desconocida ("rechazado") queda pendiente_de_revisor sin error; un emparejado con disposición rejected sigue siendo TP; la clave acepta defectos con mecanismo o archivo vacío (el archivo vacío empareja con "." por path.Clean). No los promuevo a defecto: ver sospechas.

## Lente 2 · Cableado

- go list -deps ./cmd/... ./internal/... | grep -c eval da 0; grep de "eval/" en cmd e internal, vacío. eval/ no entra al binario. LIMPIO.
- Denominador: Plan{Executable, Incompatible{CohortID, Reasons}} sirve a 3.1. RunStatus cubre los cinco estados de 2.1; causa, código de salida y eventos no están modelados (alcance de 2.1, no es defecto).
- F6 (BAJA-MEDIA) · El manifiesto no tiene variante — EVIDENCIA: TestBreaker_VariantField (10:48:30): un manifiesto con "variant":"actual" da: manifiesto: json: unknown field "variant". El spec dice que el manifiesto MUST identificar "la variante", y 4.1 empareja actual y modular. Con DisallowUnknownFields un consumidor no puede agregarla sin cambiar el tipo. La tarea 1.1 no nombra la variante; JUICIO confianza media: hueco del contrato de "única fuente" de D10.
- No modelado y declarado por el constructor: plataformas distribuidas, declaradas y probadas (D2) y "señales disponibles" de la comprobación previa (design, paso 1). Alcance abierto aceptable, pero el [x] de 1.1 en tasks.md no lo registra como excepción (solo el informe).

## Lente 3 · Verdad de contratos

- F7 (BAJA) · Test vacuo para "Rechaza campos desconocidos" — EVIDENCIA: el mutante M_unknown (borra dec.DisallowUnknownFields()) sobrevive con "ok". En load_test.go:67 el JSON {"protocol_version":"1","api_key":"secreto"} falla igual por Validate (falta rubric_version); la aserción no distingue el rechazo por campo desconocido. El comentario de Load y el nombre TestLoad_RoundTripAndUnknownFields prometen algo que la suite no prueba.
- Otros sobrevivientes (BAJA): M_trim (quitar TrimSpace del valor de opt-in); M_neg (quitar n < 0, probablemente casi equivalente); M_emptymech (quitar el guard de mecanismo vacío; observable porque la clave admite mecanismo vacío, S2; el comentario "un mecanismo vacío no empareja" queda sin prueba que lo distinga).
- Mutantes muertos: M_weight, M_dupcohort, M_flagsorder, M_fproute, M_frange, M_lower, M_dupfid, M_evidence.
- F8 (MEDIA, divergencia de contrato) · Terminal integrada del IDE — EVIDENCIA: TestBreaker_IDETerminalSpecScenario: celda interface=ide con cli_in_ide_terminal=true da incompatible ("la CLI corre en la terminal integrada de un IDE: la cohorte es terminal, no ide"). El escenario del spec dice THEN "el manifiesto registra la interfaz real como terminal y la corrida se agrega a la cohorte de terminal". El código descarta la celda en lugar de agregarla a terminal. Supuesto declarado; JUICIO confianza media: defendible como precondición (declarada distinta de real), pero el [x] de 1.1 no dice que el escenario se cumpla por otra vía. Hay que llevarlo al grupo o corregirlo.
- tasks.md 1.1 y 1.4: los casos y mutantes listados existen en los tests (comprobado por nombre); los tiempos RED y GREEN son históricos y no los puedo reproducir (limitación). El [x] de 1.4 cubre los cinco escenarios pedidos, pero F1 muestra que el uno a uno falla con varios candidatos. El [x] de 1.1 cubre versión, permisos y opt-in, pero F2 y F3 muestran huecos en interfaz y en la familia de opt-ins.
- README: "solo biblioteca estándar, sin tocar go.mod", "no importado por cmd/ ni internal/" y "externo al binario": verdaderos.

## Sospechas (no promovidas)

- S1: un hallazgo emparejado con disposición rejected cuenta como TP. Falta que el grupo diga si el rechazo prevalece.
- S2: la clave acepta defectos con archivo o mecanismo vacío (solo valida ID y rango); un defecto así es omisión garantizada.
- S3: se rechazan versiones con prefijo v o prerelease. Si las versiones fijadas se reportan así, toda celda quedaría incompatible. Falta la salida real de --version de los clientes fijados.
- S4: una disposición desconocida queda pendiente en lugar de error; un typo del revisor sale de los conteos en silencio.

## Contaminación

Leí el informe del constructor (informes/apply-1-constructor.md: RED y la existencia de SUPOSICIÓN) y la evidencia de tasks.md. Mis veredictos se apoyan en mis ejecuciones; la lista de supuestos la tomé del brief del orquestador.

## Fuentes leídas

eval/baseline/README.md, manifest/manifest.go, manifest/load_test.go (53-79), review/review.go, grep de review_test.go; openspec/changes/opsx-baseline/tasks.md; design.md (17-34, 47-57, 128, 145, 150-153); specs/opsx-benchmarking/spec.md (11-100).

## Reproducción (tests ya borrados del repo)

    review: dA={A, x.go, 10-20, nil deref}  dB={B, x.go, 15-25, nil deref}
            f1={f1, x.go, 10-12, nil deref}  f2={f2, x.go, 16-18, nil deref}
      Match([dA,dB], [f1,f2])  -> TP 1, Duplicates 1, omision B
      Match([dB,dA], [f1,f2])  -> TP 2, sin omisiones                       (F1)
    manifest, celda base claude/linux/terminal/api 2.1.310 con permisos iguales:
      Interface = desktop                                -> ejecutable       (F2)
      Observed.Settings OTEL_LOG_TOOL_CONTENT = 1        -> ejecutable       (F3)
      Interface = ide con CLIInIDETerminal               -> incompatible     (F8)
      Load de manifiesto valido seguido de otro objeto   -> err nil          (F4)
      Load de manifiesto con campo variant               -> unknown field    (F6)
    mutante: borrar la linea dec.DisallowUnknownFields() -> suite ok         (F7)

## Veredicto

FALLOS REPRODUCIDOS. Al constructor vía orquestador para corrección TDD: F1, F2, F3, F8 (MEDIA; F8 puede cerrarse con una decisión de grupo documentada), F6 (BAJA-MEDIA), F4, F5, F7 (BAJA). Tras la corrección corresponde la ronda 2 de breaker (de 3) sobre los archivos afectados. Sin autoridad de cierre.
