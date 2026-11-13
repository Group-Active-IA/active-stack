# verify-t2-1 · breaker · opsx-baseline (1.3 + DEUDA 6)

Rol: rompedor, ronda única, LOW. Base `5dcfc68` (HEAD), diff sin commitear, 2026-11-11. No edité código ni artefactos; mis tests temporales (`fixture/zz_breaker_test.go`) fueron borrados. Scratchpad: `breaker-t2/` (repos de clonado). Sin commit, sin procesos abiertos, sin `opsx-*` en %TEMP%.

## Fuentes
- Leídas: informes/apply-t2-1-constructor.md completo; fixture/{fixture,load,checkout,evaluate}.go; diff de review.go, README.md y tasks.md; verify-1-auditor.md (solo DEUDA 6); fixtures/*/task.json (huellas), .gitattributes.
- Lecciones L-NNN: no existe `knowledge-base/` ni archivos L-*; no hay lecciones que leer.
- Contaminación: leí el informe del constructor (indicado por el brief) y el del auditor. No leí otros informes del breaker.

## EVIDENCIA
1. Corrida propia única `go test -count=1 -v ./eval/baseline/...`: fixture ok 123.9 s, manifest ok, review ok. Conté 52 `--- PASS` de nivel superior y 146 con subtests, 0 FAIL, 0 SKIP (los tres paquetes juntos). Más mis pruebas adversas enfocadas: 8 (EmptyDirIntact, ReadOnlyReset, ObservedFAIL, ObservedTrivial, Spaces, IndexUntouched, CloneDir x3 configuraciones), todas ejecutadas.
2. `git archive` no toca índice ni árbol: sha256 de `.git/index`, `git status --porcelain --untracked-files=all` y `git for-each-ref` idénticos antes y después de `Prepare`+`Close` (exploracion). Sin ataque que lo rompa.
3. Huella estable entre autocrlf: copié `.gitattributes` y los fixtures a un repo scratch, `git add` con `core.autocrlf=true`, y cloné con `true`, `false` e `input`. `Load` (huella del prompt y árbol embebido de las 5 familias) pasó en las tres: LF en todos (`git ls-files --eol`: 42 `i/lf w/lf`). Gracias a `* text=auto eol=lf`. No probé Linux real (solo las tres configuraciones de git en Windows).
4. Rutas con espacios, `&` y `ñ` en TEMP: `Prepare`, `git apply` y Evaluate de revision (git + parche) y cambio-pequeno aceptan la referencia. Solo lectura: `Reset` y `Close` quitan archivo 0444 y directorio 0555 sin error.
5. Cableado: `go list -deps ./cmd/... ./internal/...` 0 coincidencias de `/eval`; `grep "eval/"` en cmd/ e internal/ vacío; `go list ./...` 0 paquetes bajo `fixtures`; los mini-proyectos tienen go.mod propio (3 go.mod) y las pruebas ocultas son `.go.txt`; `go vet ./...` limpio, `gofmt -l eval` vacío; go.mod/go.sum/cmd/internal/assets sin diff.
6. DEUDA 6: leí el diff de review.go; valida archivo y mecanismo en blanco en la clave y disposiciones fuera de {"" , rejected, real_defect} en todos los hallazgos. Los tests pasan (corrida 1). Sin hueco encontrado en los casos borde previstos.
7. Fallos reproducidos (todos BAJA, ver abajo F1 a F3).

## Fallos reproducidos (todos BAJA, van a DEUDA; ninguno ALTA ni MEDIA)
- **F1. checkout_intacto ignora directorios vacíos agregados.** Repro: `TestZZ_EmptyDirIntact` (diagnostico: `Prepare`, `MkdirAll dir_vacio_agregado`, Evaluate con la referencia). Salida: `status=aceptada ... G1..G5:cumplido`, o sea el criterio de checkout intacto da cumplido. Causa: `TreeHash` solo cuenta archivos regulares (documentado en su comentario). El tasks.md dice que `Reset` "quita directorios agregados" (cierto), pero el chequeo de intacto no los ve. Impacto bajo: un directorio vacío no altera el comportamiento. Responsable propuesto: constructor, al endurecer el chequeo (2.1).
- **F2. diag_reproduccion acepta un fragmento observado trivial.** Repro `TestZZ_ObservedTrivial`: `observed` = `"FAIL"`, `"a"` y `"--- FAIL: TestWrapExactWidth"` dan `status=aceptada`. El fragmento solo debe ser subcadena de la salida de la prueba citada; no se compara con el síntoma de la clave. Es la suposición 8 del constructor, así que no es un defecto oculto, pero deja la prueba de reproducción débil. Responsable: grupo o responsable de 02 (D1/rúbrica), luego constructor.
- **F3. Comando de huella del README no reproduce en Git Bash de Windows.** Repro: en `fixtures/cambio-pequeno/project`, `find . -type f | sed 's|^\./||' | LC_ALL=C sort | xargs sha256sum | sha256sum` da `de683fd0...`; el `task.json` y `TreeHash` dicen `dca4f76c...`. Causa: sha256sum 8.32 de Git Bash imprime ` *archivo` (modo binario) en vez de `  archivo`. Con `xargs sha256sum --text` da `dca4f76c...` (coincide). Es un defecto de documentación en Windows; en Linux GNU coincide. Corrección: añadir `--text` al README. Responsable: constructor.
- Detalle menor de informe: el constructor dice "49 archivos de datos" en `fixtures/`; hay 41 (los 49 incluyen los 8 `.go` de `fixture/`). Documentación, BAJA.

## SUPOSICIÓN y sospechas (no reproducidas)
- S1. Un symlink creado por el agente en su checkout haría que `TreeHash` devuelva error de arnés ("archivo no regular") en vez de `NotMet` en checkout_intacto. Falta: privilegio de symlinks en Windows (o correr en Linux) para reproducirlo. Severidad probable BAJA.
- S2. `Load` no valida el contenido de la clave: con `key.json` = `{}` y respuesta `{}` el razonamiento del código (entrada y archivos vacíos coinciden) da E1/E2 cumplidos. No lo corrí. Falta: ejecutar Evaluate con `tamper` de exploracion. BAJA (datos propios).
- S3. `Source.Dir` embebido y `Args[0]` de `pruebas_ocultas` no pasan por la guarda de rutas de `read`, y `Args[0]` no se comprueba en `Load`; un typo falla recién en Evaluate. BAJA, datos de confianza.
- S4. `decodeAnswer` exige que la respuesta empiece por `{` o traiga bloque ```json; JSON seguido de prosa se rechaza como inválido. Puede dar falsos rechazos si el prompt no fija el formato; no lo contrasté con los prompts.
- S5. Sin timeout propio en `goTest` (rige el de 10 min de go test). Duración del paquete `fixture` ~2 min (el constructor dice 3); la deuda de duración ya está declarada.
- No verificado: Linux/macOS reales; `git archive` con `core.eol` global distinto.

## JUICIO
- Correctitud: la lógica de Decide/Evaluate, Reset/Close y DEUDA 6 resiste mis ataques; los hallazgos son de endurecimiento menor. Cableado: limpio. Contratos: tasks.md y el informe coinciden con el comportamiento salvo F3 y el recuento de 49.
- Tres lentes cubiertas; no se corrió la suite completa del repo (reutilizo la del constructor, 30 paquetes ok, porque no cambió nada de lo que cubre fuera de eval).

## VEREDICTO
Sin fallos ALTA ni MEDIA. Por la regla de la ronda, F1 a F3 (BAJA) pasan a DEUDA con los responsables indicados y no reabren ronda. Limpio para cerrar la tanda 2 desde el rompedor. No tengo autoridad de cierre.
