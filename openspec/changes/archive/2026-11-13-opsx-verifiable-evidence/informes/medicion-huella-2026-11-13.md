# Medición de la huella (tarea 1.3)

Fecha: 2026-11-13. Base: `origin/main` `7a60089`. Script reproducible: `informes/medicion-huella/medir-huella.sh` (crea un repositorio descartable; no toca el repo ni la configuración del usuario).

## Dónde se midió
| Entorno | Detalle |
|---|---|
| Windows | Windows 11, Git Bash (MSYS), Git 2.47.1.windows.2, `sha256sum` GNU 8.32, PowerShell, Go 1.26.2 |
| Linux | Ubuntu en **WSL2** (kernel 6.6.87.2-microsoft-standard-WSL2), Git 2.53.0, `sha256sum` uutils 0.8.0, sistema de archivos ext4 de WSL (`/tmp`) |

Limitaciones: el Linux medido es WSL2, no un Linux nativo ni un runner de CI. macOS y Termux no se midieron. Cada corrida se hizo con `core.autocrlf` en `true`, `false` e `input`, fijado solo en el repo descartable.

## M1. Huella de los bytes (decisión: sha256 de los bytes, algoritmo nombrado)
Mismo texto con LF (6 bytes) y con CRLF (7 bytes):

| Herramienta | LF | CRLF |
|---|---|---|
| `sha256sum --binary` (Windows, Linux) | `73aa9cd9…2b44` | `83a6a5f9…ff2b` |
| `sha256sum --text` (Windows, Linux) | `73aa9cd9…2b44` | `83a6a5f9…ff2b` |
| `Get-FileHash -Algorithm SHA256` (Windows) | `73aa9cd9…2b44` | `83a6a5f9…ff2b` |
| `git hash-object` (autocrlf `true` o `input`) | `3676bd5a…` | `3676bd5a…` (**igual**: Git normaliza) |
| `git hash-object` (autocrlf `false`) | `3676bd5a…` | `2e4dfde4…` |
| `git hash-object --no-filters` | `3676bd5a…` | `2e4dfde4…` |

Resultado: los tres métodos de sha256 coinciden entre sí y entre Windows y Linux; LF y CRLF dan valores distintos. `git hash-object` **no sirve como definición**: no es un sha256 del contenido y depende de `core.autocrlf` (con `true` o `input`, LF y CRLF dan el mismo id). En esta máquina `sha256sum --text` y `--binary` dieron lo mismo; el aviso de `--text` de 02 se mantiene como precaución para Git Bash, sin que esta medición lo haya necesitado.

## M2. Huella del estado probado: TreeHash frente a índice temporal
Escenarios (todos parten de los mismos bytes en la copia de trabajo; `treehash` = sha256 de las líneas `<sha256 archivo>  <ruta>` ordenadas, sobre `git ls-files --cached --others --exclude-standard`; `index` = `git write-tree` sobre una copia del índice tras `git add -A`). Valores abreviados a 16 hex.

| Escenario | treehash (`true`, `false`, `input`, Win y Linux) | index (`true` = `input`) | index (`false`) |
|---|---|---|---|
| S0 limpio | `1642de00a955de70` | `f8b0e07aecfee08c` | `140794fb9b9c5173` |
| S1 f.txt modificado sin preparar | `1e087e8d0a2a92e6` | `41bb4e50186af5df` | `a0296100bd16bf38` |
| S2 mismo HEAD, g.txt modificado y preparado | `632cbba40a79a7a1` | `41e249c67ed055a4` | `8fea447aa59b26da` |
| S3 archivo nuevo sin trackear | `c7631fed78054154` | `d0fcf18f06112a30` | `1adc734ab44365c0` |
| S4 f.txt: mismo texto pero CRLF | `6f46ba11339a6c55` | `f8b0e07aecfee08c` (= S0) | `6eb8ad692a0adad0` |
| S5 g.txt borrado | `e0b0a3701433b552` | `196228cfdfdd8149` | `196228cfdfdd8149` |
| S6 archivo ignorado + `.gitignore` | `abf632b30450a14c` | `34496596abcfecbd` | `d94685b07411f25f` |
| S7 g.txt con LF (el commit lo tenía con CRLF) | `f6cf7032537f4fac` | `f8b0e07aecfee08c` (= S0) | `f8b0e07aecfee08c` (= S0) |

Hallazgos, medidos:
1. **TreeHash no depende de `core.autocrlf` ni de la plataforma**: los ocho valores son idénticos en las tres configuraciones y entre Windows y Linux (WSL2).
2. **El índice temporal depende de la configuración**: el mismo estado (S0) da `f8b0…` con `true` o `input` y `1407…` con `false`.
3. **El índice normaliza fines de línea con `true`/`input`**: S4 y S7 (bytes distintos en el disco) dan el mismo árbol que S0. Un recibo de un estado con CRLF serviría para otro con LF. TreeHash los distingue (S4 ≠ S0, S7 ≠ S0).
4. Los dos distinguen mismo HEAD con dos diffs distintos (S1, S2), archivo nuevo sin trackear (S3) y archivo borrado (S5), y son estables en dos corridas seguidas.
5. **Un archivo ignorado no cambia el estado** en ninguno de los dos (S6 solo cambia por el `.gitignore` nuevo). Es un límite: un archivo ignorado que influya en el resultado de un comando (por ejemplo `.env` o una salida de build) no queda en la huella.
6. TreeHash de Go (`fixture.TreeHash`) sobre una copia del árbol de trabajo S0 (sin `.git`) da `1642de00a955de709d3901a362dc9195c1476411c13628a2714e11134b501921`, **igual** al cálculo con shell sobre el repo. Detalle de implementación: `TreeHash(dir)` recorre todo el directorio, incluido `.git`, y falla ante archivos no regulares; para un repositorio hay que pasarle el conjunto de archivos de Git o una copia sin `.git`.

## Decisión (propuesta de 06 con evidencia; la confirma Gianfranco con Emiliano)
**Huella del estado probado = TreeHash sobre los bytes de la copia de trabajo** (`git ls-files -z --cached --others --exclude-standard`, sin los borrados). Razones: invariante a `core.autocrlf` y a la plataforma, distingue LF de CRLF, recomputable con `sha256sum` y coherente con la huella de bytes y con 02. El índice temporal de Git queda **descartado como definición**: depende de la configuración, normaliza fines de línea y es un id de objeto de Git, no un sha256 del contenido.

## Lo que no se midió
- Linux nativo o CI, macOS, Termux: pendientes. El Linux medido es WSL2.
- Rutas con mayúsculas y minúsculas distintas en sistemas insensibles a mayúsculas, enlaces simbólicos y archivos muy grandes.
- `.gitattributes` con `eol` o `text` por archivo: el TreeHash lee bytes del disco y no debería depender de ellos; no se probó.
- Rendimiento sobre un repositorio grande.
