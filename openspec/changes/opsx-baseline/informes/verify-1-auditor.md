<!-- Transcripción literal del informe devuelto por el subagente role-auditor (que no tiene herramienta de escritura); guardado por el orquestador el 2026-11-11 sin modificar el contenido. -->

# verify-1 · auditor · opsx-baseline (tanda 1: tareas 1.1 y 1.4)

**Veredicto: APROBADO para la tanda 1 (tareas 1.1 y 1.4).** No queda ningún bloqueante abierto. Hay 10 puntos de DEUDA y ninguno impide cerrar la tanda. El veredicto no habilita commit, push, merge, archivo del change ni ampliar el alcance. El change sigue en apply y le faltan 1.2, 1.3 y los grupos 2 a 4.

Rol: auditor. No edité código ni artefactos, no hice commits y no cambié de rama. Clasificación LOW. Ventana de trabajo: 2026-11-11, de 11:14:57 a 11:17:41 (-03).

## Revisión auditada
- HEAD `43eeb7f6fd98e5183d6313e48c6a3dc6998a49b7`, rama `propose/opsx-baseline`, con cambios sin commitear.
- `git status`:
  - modificados: `design.md` y `tasks.md` del change;
  - sin seguimiento: `eval/` (11 archivos), `informes/` y `auditoria-ramas-opsx-2026-11-10.md`.
- `git hash-object` a las 11:15:27. Todos coinciden con lo entregado en apply-3 y verify-3:
  - manifest.go 4234ffb3, manifest/apply3_test.go 215aae09, review/apply3_test.go 39702d79, review.go 84621364;
  - README a8b1dfb2, apply2_test 3f264d3a, load_test d8392c39, manifest_test aefebcc5, plan_test 765b08ac, review/apply2_test 03e88c5a, review_test fe24eca9;
  - tasks.md 7163fb94, design.md 8b4f75b3.

## Fuentes
- **Plan:** proposal.md, design.md, tasks.md, specs/opsx-benchmarking/spec.md y revision-2026-11-11.md, completos.
- **Revisión del grupo:** informe unificado en `C:\Users\espej\Downloads\informe-unificado-propuestas-opsx.md`. Solo leí por grep los encabezados y las secciones de 02, 07 y 10.
- **Estado:** informes/LEEME.md.
- **Informes originales, completos:** apply-1, apply-2 y apply-3 del constructor; verify-1, verify-2 y verify-3 del breaker.
- **Código, completo:** manifest.go, review.go, README.md, manifest_test.go y review_test.go. No leí el contenido de apply2_test, apply3_test, load_test ni plan_test; esos tests sí los ejecuté.
- **Diff:** `git diff` de design.md y tasks.md, y `git status`.
- **CLAUDE.md del repo:** §3 y §4, desde el contexto.
- **Lista de `openspec/changes/archive/`:** 25 changes. Ninguno trata de medición ni de arnés, y este change no reescribe ninguno.
- **Lecciones L-NNN:** no hay `knowledge-base/` en el repo, así que no hay ninguna.
- **Engram:** no lo usé.

## Contaminación
Leí todos los informes del constructor y del breaker antes de mis chequeos. Los tres roles son Claude: no hay independencia de proveedor y comparten el contexto del brief del orquestador. Para compensarlo, mis conclusiones se apoyan en ejecuciones propias con otro generador, otra semilla y mis propios mutantes, no en las salidas ajenas. No reproduje los tiempos históricos de RED: son declaraciones del constructor, no evidencia mía.

## EVIDENCIA (ejecutada por mí)
1. **Conteo de tests.** `go test -count=1 -v ./eval/baseline/...` a las 11:15:50:
   - ambos paquetes `ok`, exit 0;
   - conté 100 `=== RUN`: 34 funciones top-level y 66 subtests;
   - 100 `--- PASS`, 0 FAIL, 0 SKIP, sin nombres top-level duplicados.
   - Coincide con el conteo del breaker en verify-3.
2. **Aislamiento del binario.**
   - `go list -deps ./cmd/... ./internal/... | grep -c /eval` da 0;
   - grep de `active-stack/eval` en cmd/ e internal/: vacío;
   - imports de eval: solo stdlib (manifest: encoding/json, fmt, io, sort, strconv, strings; review: fmt, path, sort, strings);
   - `go list ./...` lista 29 paquetes, lo que coincide con "29 ok" del constructor.
3. **Archivos fuera del alcance.**
   - `git diff --quiet HEAD -- go.mod go.sum`: sin diff;
   - `git diff --stat HEAD -- cmd internal assets`: vacío. No se tocó backup, filemerge, pipeline ni el instalador.
4. **Formato y validación.**
   - `gofmt -l eval`: vacío;
   - `go vet ./eval/...`: ok;
   - `openspec validate opsx-baseline --strict`: `Change 'opsx-baseline' is valid` (11:16:04).
5. **Casillas de tasks.md.** Solo 1.1 y 1.4 pasan de `[ ]` a `[x]`; el archivo tiene 2 `[x]`. El resto del diff son subviñetas de evidencia y la sección final de suite. design.md cambia en una sola oración (Datos sensibles, `OTEL_LOG_LEVEL`). En la copia de trabajo los dos archivos tienen todas sus líneas en CRLF; con `core.autocrlf=true` es normal y el diff lógico es de una línea.
6. **Test de propiedades propio (review, 1.4).** Corrí `TestAuditor_PropertyMaxAndInvariants` en una copia del módulo en el scratchpad, con semilla 20261111 y 5000 casos:
   - hasta 5 defectos y 6 hallazgos, dos archivos con variantes de ruta (`./`, `\`), mecanismos con variantes de mayúsculas y espacios, y disposiciones mezcladas.
   - Comprobé que la cantidad de TP es igual al máximo calculado por fuerza bruta.
   - Comprobé que cada TP cumple `matches` y que ningún defecto tiene dos TP.
   - Comprobé que los duplicados tienen candidatos y que FP, nuevos y pendientes no tienen candidatos.
   - Comprobé que TP+dup+FP+nuevos+pendientes es igual a la cantidad de hallazgos, y que TP+omisiones es igual a la cantidad de defectos.
   - Comprobé que los conteos no cambian al permutar la clave y los hallazgos.
   - Resultado: PASS (11:16:50).
   - Para probar que el test es sensible, corrí un mutante propio sin camino de aumento (`if !taken || (false && try(prev, seen))`): FAIL, `it 2107: TP=1, max=2`.
7. **Escenarios del spec por la vía JSON (1.1).** `TestAuditor_LoadPlanSpecScenarios` hace Load y luego Plan sobre un manifiesto de 5 cohortes:
   - `Executable` = [ok, ideterm]; ideterm sale con Interface=terminal y DeclaredInterface=ide;
   - `Incompatible` = ver ("menor que la mínima"), perm ("permisos efectivos distintos") y optin (un solo motivo, OTEL_LOG_USER_PROMPTS; `OTEL_LOG_LEVEL=debug` en la misma celda no agrega motivo);
   - resultado: PASS.
8. **Mutantes propios contra la suite entregada**, en la copia, todos revertidos y verificados con `cmp` contra el repo:
   - A1 Plan usa la interfaz declarada: muere (`TestPlan_IDEIntegratedTerminalReclassifiedToTerminal`);
   - A5 acepta peso 0: muere;
   - A6 sin chequeo de fixture_revision: muere;
   - A7 el valor "1" cuenta como apagado: muere (`TestCheckCell_OptInFamily`);
   - A9 un hallazgo sin decisión cuenta como FP: muere (`TestMatch_UnmatchedFindings`).
   - Mis tests temporales quedaron solo en el scratchpad y ya los borré; el repo no se tocó.
9. **El manifiesto no exige todo lo que el spec pide.** El manifiesto de 7, sin `models`, `prompt`, `profile` ni `active_stack_revision`, pasó Load y Plan como ejecutable. Ver DEUDA 1.

## Contraste
- **Cumplimiento del plan.**
  - 1.1 cubre los tres casos de su texto: versión incompatible, permisos distintos y opt-in activo. Esas celdas quedan incompatibles antes de iniciar, fuera del denominador (`Plan.Executable`) y listadas con su motivo (`Plan.Incompatible`).
  - Además cubre los escenarios "Celda incompatible antes de iniciar", "Opt-in de contenido activo" y "CLI en la terminal integrada del IDE"; este último se resuelve reclasificando, según el spec (F8).
  - 1.4 cubre TP con evidencia, duplicado que no suma, omisión, FP rechazado y hallazgo nuevo que no altera los conteos.
  - El arnés queda fuera del binario, en Go, sin dependencias nuevas, tal como decidió D6. No hubo build.
- **Frescura.** La última edición de producción es de las 11:08:21 y la suite completa del constructor corrió de 11:09:06 a 11:10:57. Los hashes no cambiaron desde entonces: verify-3 los vio de 11:12 a 11:13 y yo a las 11:15. No repetí `go test ./...`: el punto 2 muestra que ningún paquete fuera de eval importa eval y go.mod no cambió.
- **RED antes de producción.** El constructor declara RED por aserción en cada comportamiento: 1.1 y 1.4 en apply-1, F1 a F6 y F8 en apply-2, R1 en apply-3. Las aserciones que nacieron verdes (F7, M_trim, M_neg, M_emptymech, N_modeTrim, N_omOrder) no cambiaron producción y se acreditaron por mutante. Lo acepto como evidencia declarada y lo corroboré con mis mutantes A1 a A9.
- **Supuestos y sospechas.** Se clasifican abajo; ninguna se promueve a fallo demostrado.

## JUICIO
- Autor: auditor. Comparé el plan, el spec y las decisiones con mis ejecuciones de los puntos 1 a 9.
- Confianza alta en la corrección de la lógica pura de 1.1 y 1.4 y en que el binario queda aislado.
- Confianza media en cómo se comportará con datos reales: versiones, flags y claves `[otel]` dependen de 1.2 con clientes reales, y Codex no está instalado.

## BLOQUEANTES
Ninguno.

## DEUDA
1. **Manifiesto: campos que el spec exige y Validate no comprueba.** El spec dice que el manifiesto MUST identificar el modelo, la configuración efectiva y la revisión. Hoy no se exigen `models`, `profile.config_fingerprint`, `active_stack_revision` ni `prompt` (EVIDENCIA 9). El texto de la tarea 1.1 no los pide, así que no bloquea esta tanda.
   - Responsable: constructor, con el planificador si hace falta mover el criterio a 1.2 o 3.4.
   - Plazo: antes de publicar un manifiesto real (3.4) y antes del archivo.
2. **Rúbrica de 1.4 pendiente de revisión por 10.** revision-2026-11-11 dice que el texto de la regla es nuevo y que 10 debe revisarlo. La implementación fija varias interpretaciones:
   - emparejamiento máximo con desempate por orden de hallazgo y por ID;
   - clase `pendiente_de_revisor`;
   - evidencia del TP que se propaga pero no se exige;
   - un hallazgo emparejado con disposición rejected cuenta como TP (S1).
   - Responsable: el responsable de 02 lo lleva a 10 y al grupo.
3. **Valores de versión (S3).** Una versión con prefijo `v` o con prerelease hoy da celda incompatible. Hay que confirmarlo con la salida real de `--version` de los clientes fijados. Responsable: constructor, en la tarea 1.2.
4. **Celda sin observación de settings (S5).** Hoy `settings` ausente da ejecutable, igual que "observé y no hay ninguno". Está ligado a la excepción de "señales disponibles". Responsable: decisión del grupo y luego constructor en 1.2.
5. **Claves `[otel]` de Codex (supuesto 5 de apply-2).** Solo se reconocen `log_user_prompt` y `otel.log_user_prompt`. Si existe otra clave de contenido, la celda pasaría como ejecutable en un control de privacidad. SUPOSICIÓN: no lo verifiqué contra la configuración real de Codex. Responsable: constructor en 1.2 o 2.2.
6. **Validación de la clave y de las disposiciones (S2 y S4).** La clave acepta defectos con archivo o mecanismo vacío, que serían una omisión garantizada. Una disposición desconocida queda pendiente sin error, así que un typo del revisor sale de los conteos en silencio. Responsable: constructor en 1.3, cuando se congelen las claves.
7. **Cosméticos.**
   - S6: dos grafías de la misma variable dan dos motivos.
   - "Desktop" con mayúscula recibe el motivo genérico.
   - El mensaje de error de `TestMatch_FindingPointingAtTwoDefects` todavía dice "primero de la clave", aunque el criterio vigente es el menor ID; en ese caso coinciden.
   - Responsable: constructor, sin urgencia.
8. **Decisiones de nombres y alcance para el grupo.**
   - Modalidad `subscription` o `suscripcion`.
   - Solo el modo vacío cuenta como "permisos vacíos".
   - La excepción de `OTEL_LOG_LEVEL` está en design.md pero sin fila, fuente ni fecha en el registro de decisiones.
   - Plataformas distribuidas, declaradas y probadas (D2) y "señales disponibles": no están modeladas; figuran como excepción en tasks.md 1.1.
   - Responsable: el responsable de 02 o el planificador, que registra con fecha.
9. **LEEME.md desactualizado.**
   - "Check compartido" sigue citando la corrida de 10:40 a 10:43.
   - "Siguiente acción" sigue diciendo "breaker sobre el diff".
   - "Revisión / diff" dice 1083 líneas, pero eval/ hoy tiene 1611 líneas contando el README.
   - Falta registrar este informe del auditor.
   - Responsable: orquestador, antes de la próxima tanda.
10. **Archivo ajeno sin seguimiento.** `auditoria-ramas-opsx-2026-11-10.md` en la raíz no es de este change. Si alguna vez se commitea, no debe entrar en el commit de 02. Responsable: orquestador o responsable.

## LECCIONES (candidatas; no hay L-NNN previas)
- Cuando un breaker sugiere ampliar una lista de seguridad por prefijo (F3, `OTEL_LOG_*`), el mismo pedido de corrección debería incluir un test negativo de miembros benignos conocidos (como `OTEL_LOG_LEVEL`). Así se evita una regresión R1 y una ronda extra.
- Para reglas de emparejamiento o asignación, un test de propiedades contra fuerza bruta con permutación de entradas detecta lo que los casos de ejemplo no detectan (F1). Conviene exigirlo desde RED en la tarea que reutilice 10.

## Limitaciones y chequeos no ejecutados
- No reejecuté `go test ./...`; lo justifico por el aislamiento verificado en EVIDENCIA 2.
- No reproduje los RED históricos ni los mutantes MF*, M* y N* ajenos. Los míos son otros (A1 a A9 y el de aumento).
- No hay clientes reales: todos los datos son sintéticos.
- No leí completos el informe unificado del grupo ni los archivos de test apply2, apply3, load y plan; esos tests sí los ejecuté.
- No hubo proveedor de identidad externo ni aprobación humana HIGH o CRITICAL, así que nada de eso aplica.

## Derivación e intentos
- Hay un solo intento de auditoría para esta tanda y no hay correcciones pendientes que invaliden verificaciones aguas abajo.
- Responsables de la DEUDA:
  - constructor: 1, 3, 5, 6 y 7;
  - grupo o responsable de 02: 2, 4 y 8;
  - orquestador: 9 y 10.
