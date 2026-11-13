# verify-2 · breaker · opsx-baseline (corrección de F1-F8, ronda 2 de 3)

Rol: breaker. Sin edición de código ni artefactos, sin commit, sin cambio de rama. verify-1-breaker.md no se tocó.

## Revisión revisada

- HEAD 43eeb7f, rama propose/opsx-baseline, más los cambios sin commitear. Inicio 2026-11-11 11:02:40 -03, fin 11:05:18.
- git status --short (igual al inicio y al final): " M openspec/changes/opsx-baseline/tasks.md", "?? auditoria-ramas-opsx-2026-11-10.md", "?? eval/", "?? openspec/changes/opsx-baseline/informes/".
- git hash-object (igual al inicio y al final, y coincide con la tabla del informe apply-2):
  - a8b1dfb2 eval/baseline/README.md
  - 3f264d3a manifest/apply2_test.go · d8392c39 manifest/load_test.go · 6a7d92ac manifest/manifest.go · aefebcc5 manifest/manifest_test.go · 765b08ac manifest/plan_test.go
  - 03e88c5a review/apply2_test.go · 84621364 review/review.go · fe24eca9 review/review_test.go
  - 00a28bf9 openspec/changes/opsx-baseline/tasks.md
- L-NNN: sigue sin existir knowledge-base/.

## Pruebas ejecutadas por mí

- go test -count=1 -v ./eval/baseline/... (11:03:06): ambos ok. Conté 92 "=== RUN" = 31 funciones top-level (manifest: apply2 7, load 2, manifest 4, plan 3; review: apply2 4, review 11) + 61 subtests.
- Tests adversariales temporales: manifest/zz_breaker_test.go (4 funciones, con los ataques de verify-1 adaptados a la firma nueva) y review/zz_breaker_test.go (3 funciones, una con 3000 casos aleatorios contra un emparejamiento máximo por fuerza bruta). En total 7 tests propios, todos pasan. Borrados a las 11:05:18 (git status y hashes verificados después).
- 13 mutantes nuevos sobre copias aisladas en C:\Users\espej\AppData\Local\Temp\claude\8a00efe90647aaab\59b3ace0-deab-4305-9f26-437b91b11f36\scratchpad\breaker-02\m (11:04:22).
- Reutilizo go test ./..., gofmt, go vet y openspec validate del constructor (10:59-11:00, misma revisión). No los repetí.

## Estado de F1-F8 (con mi test original re-ejecutado)

| F | Estado | EVIDENCIA |
|---|---|---|
| F1 orden de clave | CERRADO | El caso original da 2 TP, 0 omisiones con clave [A,B], [B,A] y hallazgos permutados (TestBreaker2_F1Original, 11:04:02). En 3000 casos aleatorios (seed 42, hasta 6 defectos y 7 hallazgos, dos mecanismos, disposiciones mezcladas) obtuve 0 casos no máximos frente a la fuerza bruta. Los conteos no cambian al permutar la clave y los hallazgos. La asignación por hallazgo no cambia al permutar la clave. Se cumplen los invariantes: TP+dup+FP+nuevos+pendientes = hallazgos, TP+omisiones = defectos, ningún defecto con dos TP, ningún duplicado sin candidatos. Una cadena de aumento de longitud 3 (g1 sobre D1/D2, g2 sobre D2/D3, g3 solo sobre D1) da 3 TP. |
| F2 dimensiones | CERRADO | desktop, interfaz vacía, cliente vacío o gemini, modalidad vacía, sistema vacío, permisos vacíos y observado vacío dan incompatible. Si el modo observado está vacío y el declarado no, da incompatible por permisos. Queda un residuo en S5. |
| F3 familia de opt-ins | CERRADO, con regresión R1 | OTEL_LOG_TOOL_CONTENT, otel_log_user_prompts, otel.log_user_prompt, "OTEL_LOG_USER_PROMPTS " y Log_User_Prompt dan incompatible. |
| F4 contenido extra | CERRADO | Un segundo objeto, basura o un [] después del manifiesto dan "contenido extra". Los espacios y saltos finales se aceptan. |
| F5 task_id / Plan | CERRADO | task_id duplicado o vacío da error. Plan() de un manifiesto cero devuelve error de Validate. Observación: un task_id " " (solo espacios) se acepta. |
| F6 variante | CERRADO | Sin variant o con variant en blanco, Load falla con "falta variant". Con variant, Load funciona. |
| F7 test vacuo | CERRADO | El test ahora usa un manifiesto válido más api_key y exige "unknown field". Mi Load con api_key da `unknown field "api_key"`. El constructor reporta que M_unknown ya muere; no lo repetí. |
| F8 terminal integrada | CERRADO | ide con cli_in_ide_terminal da ejecutable con Interface=terminal. Plan conserva DeclaredInterface=ide (el mutante N_declared muere). ide sin terminal integrada queda en ide. desktop con terminal integrada sigue incompatible. ide+CLI con versión baja da incompatible con Interface vacía. |

## Consumidores y regresiones

### Cableado

- Plan() (Plan, error) y Executable []Planned: grep de ".Plan()", "[]Planned" y de los imports "baseline/manifest" y "baseline/review" en todo el repo. Solo hay usos en los tests de eval/baseline. No hay consumidores en cmd/ ni internal/ (los "Executable" de cmd son os.Executable). No se rompe nada fuera de eval. LIMPIO.
- tasks.md 1.1 registra ahora las excepciones D2 y "señales disponibles" (línea 21). El README sigue siendo verdadero, aunque no menciona variant (no es un defecto).

### R1 (BAJA, reproducido) · Falso positivo del prefijo otel_log_ con OTEL_LOG_LEVEL

    go test -count=1 -v -run Breaker2 ./eval/baseline/manifest/   (11:03:33)
    no-contenido "OTEL_LOG_LEVEL"="info" -> incompatible [opt-in de contenido activo: OTEL_LOG_LEVEL]
    no-contenido "OTEL_LOGS_EXPORTER"/"OTEL_METRICS_EXPORTER"/"otel.exporter"/"otel.environment" -> ejecutable

OTEL_LOG_LEVEL es la variable estándar del SDK de OpenTelemetry para la verbosidad del propio SDK; no registra contenido. Con cualquier valor que no sea de apagado (info, debug, warn), la celda queda excluida y el motivo afirma algo falso ("opt-in de contenido activo"). La falla es cerrada y el motivo queda visible, así que no se contamina el denominador. Pero la celda se pierde y el motivo miente.

JUICIO (breaker, confianza media): el design dice literalmente "variables OTEL_LOG_*" y el código cumple esa letra. El spec, en cambio, habla de "opt-in de registro de contenido". Mi sugerencia de verify-1 (tratarlo por prefijo) contribuyó a esto. Hay dos salidas:
- excluir OTEL_LOG_LEVEL de la familia, con un test;
- documentarlo como decisión de falla cerrada.

SUPOSICIÓN: que el entorno de medición pueda tener OTEL_LOG_LEVEL definido.

### Mutantes sobrevivientes nuevos (BAJA)

- N_modeTrim: quitar el TrimSpace del chequeo de modo vacío sobrevive. Ningún test cubre un modo declarado que sea solo espacios. Con observado " " igual, la celda pasaría.
- N_omOrder: listar las omisiones por ID en lugar de en el orden de la clave sobrevive. apply-2-constructor.md afirma "Las omisiones se listan en el orden de la clave" y ningún test lo fija.
- N_reasonTrim: el texto del motivo con o sin trim sobrevive. Es cosmético.
- N_seenFree: quitar "!seen[d]" del primer bucle de maxMatching sobrevive. Probablemente es equivalente, porque un defecto libre no puede estar visto.
- Muertos: N_varTrim, N_desktop, N_declared, N_planValidate, N_osCheck, N_clientCheck, N_flatCodex, N_offValues, N_dupFirst.

### checkDimensions con mayúsculas y espacios

"Terminal", " terminal", "IDE", "Claude", "Subscription" y "suscripcion" dan incompatible con un motivo que cita el valor entre comillas. La falla es cerrada y explícita; lo juzgo aceptable porque los valores JSON son sensibles a mayúsculas y el motivo muestra el valor exacto. Observación cosmética: "Desktop" con mayúscula recibe el motivo genérico, no "escritorio fuera de la primera comparación".

### Supuestos nuevos del constructor

- 3 (subscription en lugar de suscripcion): JUICIO: aceptable. El valor del contrato JSON es estable y suscripcion da incompatible con un motivo explícito, sin pasar en silencio. Decisión de nombre para el grupo.
- 4 (solo el modo vacío cuenta como "permisos vacíos"): verificado. Flags declarados nil con un flag observado extra dan incompatible. [] frente a nil da ejecutable (correcto). JUICIO: aceptable. Ver N_modeTrim para el modo con solo espacios.
- 5 (solo dos claves otel.* de Codex): verificado. otel.exporter, otel.environment y otel.log_user_prompt_redacted no se marcan. No pude verificarlo contra la configuración real de Codex (limitación, igual que el constructor).
- 1 y 2 (F8 y F1): coherentes con el spec según las pruebas de arriba. La agregación por cohorte queda para 3.1.

## S1-S4 (fuera de alcance): sin cambios, confirmado (11:04:02)

- S1: un emparejado con disposición rejected sigue siendo TP.
- S2: la clave con archivo y mecanismo vacíos se acepta (err nil).
- S3: el código de parseVersion no cambió (la v y el prerelease siguen incompatibles).
- S4: una disposición "rechazado" sigue como pendiente_de_revisor.

## Sospechas nuevas (no promovidas)

- S5: Observed.Settings ausente (nil, opt-ins no observados) da ejecutable. No se distingue "no se comprobó" de "se comprobó y no hay ninguno". Para confirmarlo como defecto falta que el grupo decida si la comprobación previa debe registrar explícitamente que observó la configuración.
- S6: una misma variable con dos grafías (OTEL_LOG_USER_PROMPTS y otel_log_user_prompts) da dos motivos. Es cosmético.

## Contaminación

Leí apply-2-constructor.md completo (tabla, supuestos y hashes) y las líneas de tasks.md citadas. Mis estados F1-F8 se apoyan en mis re-ejecuciones y en los mutantes propios. Tomé del constructor, sin repetirlos, los mutantes MF* y la suite completa.

## Fuentes leídas

eval/baseline/manifest/manifest.go y review/review.go (completos); grep de apply2_test.go, review_test.go y tasks.md; informes/apply-2-constructor.md; verify-1-breaker.md (mío).

## Veredicto

FALLOS REPRODUCIDOS, solo de severidad BAJA. F1-F8 están cerrados. Queda abierta R1 (OTEL_LOG_LEVEL como falso positivo, regresión de F3), que se resuelve con un test más una exclusión, o con una decisión documentada de falla cerrada. Además hay dos huecos de prueba BAJA (N_modeTrim, N_omOrder). Si el orquestador manda corregir R1, corresponde la ronda 3 (la última) solo sobre checkContentOptIns. Sin autoridad de cierre.
