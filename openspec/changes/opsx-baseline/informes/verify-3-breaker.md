# verify-3 · breaker · opsx-baseline (ronda 3 de 3, última)

Rol: breaker. Sin edición de código ni artefactos, sin commit, sin cambio de rama. verify-1 y verify-2 no se tocaron. Alcance: R1, N_modeTrim, N_omOrder, la nota de design.md y la regresión de F1-F8.

## Revisión revisada

- HEAD 43eeb7f más cambios sin commitear. Inicio 2026-11-11 11:12:28 -03, fin 11:13:40.
- git status --short (igual al inicio y al final): " M openspec/changes/opsx-baseline/design.md", " M openspec/changes/opsx-baseline/tasks.md", "?? auditoria-ramas-opsx-2026-11-10.md", "?? eval/", "?? openspec/changes/opsx-baseline/informes/".
- git hash-object (igual al inicio y al final, y coincide con lo entregado en apply-3):
  - manifest/manifest.go 4234ffb3, manifest/apply3_test.go 215aae09, review/apply3_test.go 39702d79, review/review.go 84621364 (sin cambio).
  - Resto de eval/ igual que en apply-2: README a8b1dfb2, apply2_test 3f264d3a, load_test d8392c39, manifest_test aefebcc5, plan_test 765b08ac, review/apply2_test 03e88c5a, review_test fe24eca9.
  - tasks.md 7163fb94 y design.md 8b4f75b3.
- L-NNN: sigue sin existir knowledge-base/.

## Pruebas ejecutadas por mí

- go test -count=1 -v ./eval/baseline/..., ya sin mis archivos temporales (11:13:40): los dos paquetes dan ok.
  - Conté 100 "=== RUN": 34 funciones top-level y 66 subtests.
  - Son 3 funciones nuevas frente a apply-2: 2 en manifest/apply3_test.go y 1 en review/apply3_test.go.
- Tests adversariales temporales: 3 funciones (TestBreaker3_R1, TestBreaker3_Regression, TestBreaker3_F1Regression), todas PASS a las 11:13:02. Los borré a las 11:13:40 (git status y hashes verificados después).
- 6 mutantes sobre copias aisladas en C:\Users\espej\AppData\Local\Temp\claude\8a00efe90647aaab\59b3ace0-deab-4305-9f26-437b91b11f36\scratchpad\breaker-02\m (11:13:18).
- Reutilizo del constructor la suite completa (29 paquetes ok, 11:09:06-11:10:57), gofmt, vet y openspec validate. No los repetí.
- Limitación: la llamada de Bash de las 11:13:02 terminó con exit 1 por un artefacto del shell ("line 177: /c/Users/: Is a directory") después de imprimir los resultados. Los tests reportaron PASS/ok. No afecta a la evidencia.

## R1 · OTEL_LOG_LEVEL — CERRADO

EVIDENCIA (TestBreaker3_R1, 11:13:02):
- Dan ejecutable con valor "debug": OTEL_LOG_LEVEL, otel_log_level, " OTEL_LOG_LEVEL " y Otel_Log_Level. La normalización de mayúsculas y espacios funciona.
- Siguen dando incompatible, sin falsos negativos: OTEL_LOG_LEVELS, OTEL_LOG_LEVEL_X, OTEL_LOG_LEVEL2, OTEL_LOG__LEVEL, "otel_log_level\t_x", otel_log_, OTEL_LOG_USER_PROMPTS, OTEL_LOG_TOOL_DETAILS, OTEL_LOG_RAW_API_BODIES, OTEL_LOG_TOOL_CONTENT, otel.log_user_prompt y LOG_USER_PROMPT.
- Con OTEL_LOG_LEVEL y OTEL_LOG_USER_PROMPTS juntas, la celda queda incompatible con un solo motivo (OTEL_LOG_USER_PROMPTS).
- otel.log_level (forma de tabla) da ejecutable, coherente con el supuesto 2 del constructor.

Mutantes, todos muertos:
- R_noExcl: sin la exclusión.
- R_prefixExcl: excluye por prefijo otel_log_level, lo que daría un falso negativo con LEVELS.
- R_rawName: compara el nombre sin normalizar.
- R_broadExcl: excluye cualquier nombre que contenga "level".

## N_modeTrim y N_omOrder — CERRADOS

Mis mutantes de la ronda 2, re-ejecutados a las 11:13:18 con la suite nueva:
- N_modeTrim (quitar TrimSpace del chequeo de modo vacío): FAIL, muere.
- N_omOrder (las omisiones recorren byID en lugar de key): FAIL, muere.

Además, en 3000 casos aleatorios con la clave permutada, las omisiones siguen el orden de la clave.

## Nota en design.md

El diff es una sola oración agregada en "Datos sensibles": "OTEL_LOG_LEVEL no cuenta como opt-in de contenido: solo fija la verbosidad del SDK." Coincide con el código: es una excepción exacta por nombre normalizado, y el resto de la familia OTEL_LOG_* sigue fallando cerrada. No contradice el spec, que habla de "opt-in de registro de contenido". El texto no dice que la comparación ignore mayúsculas; es una omisión menor, no un defecto, porque el comentario de isContentOptIn sí lo dice.

## Regresión de F1-F8 — sin regresión

EVIDENCIA (TestBreaker3_Regression y TestBreaker3_F1Regression, 11:13:02):

- F1:
  - El caso original da 2 TP y 0 omisiones en las cuatro combinaciones de orden de clave y de hallazgos.
  - En 3000 casos aleatorios (seed 7), el resultado es máximo frente a la fuerza bruta y los conteos no cambian al permutar la clave.
- F2: quedan incompatibles desktop, interfaz vacía, cliente vacío o gemini, modalidad vacía, sistema vacío, permisos vacíos, observado vacío y modo con solo espacios.
- F3: OTEL_LOG_USER_PROMPTS=true da incompatible y " off " da ejecutable.
- F4: un manifiesto seguido de otro objeto se rechaza.
- F5: task_id duplicado o vacío, peso 0 e id de cohorte duplicado se rechazan. Plan() de un manifiesto cero devuelve error.
- F6: un manifiesto sin variant se rechaza.
- F7: un campo desconocido (api_key) sobre un manifiesto válido se rechaza.
- F8: Load y Plan con cli_in_ide_terminal devuelven Interface=terminal y DeclaredInterface=ide.
- También se mantienen como en las rondas 1 y 2: el flag extra y la versión baja dan incompatible, y la v de prefijo da incompatible (S3).

## S1-S6: sin cambios, confirmado

- S1: un emparejado con disposición rejected sigue siendo TP.
- S2: la clave con archivo y mecanismo vacíos se acepta (err nil).
- S3: la v de prefijo sigue incompatible.
- S4: una disposición "rechazado" sigue como pendiente_de_revisor.
- S5 y S6: no los ataqué. review.go no cambió, y en manifest.go solo cambió isContentOptIn.

## Contaminación

Leí las secciones de tabla, evidencia, suposición y contaminación de apply-3-constructor.md. Mis estados se apoyan en mis propias ejecuciones y en mis mutantes.

## Fuentes leídas

git diff de design.md, isContentOptIn en manifest.go (líneas 405-440) y apply-3-constructor.md (por grep).

## JUICIO

Autor: breaker. Comparé lo que pedía el brief de la ronda 3 con mis ejecuciones. Confianza alta.

## Veredicto

LIMPIO. R1, N_modeTrim y N_omOrder quedan cerrados, y F1-F8 no tienen regresión. Las sospechas S1-S6 siguen abiertas como estaban. Siguen fuera de alcance y quedan para el grupo y para el auditor. Sin autoridad de cierre.
