# Candidatos a muestra histórica (tarea 1.1, borrador)

> Preparado el 13/11/2026 sobre `main` en `7a60089`, antes de que 08 y 09 existan, como permite el
> roadmap ("los casos de prueba se pueden preparar antes"). **No es la muestra**: no tiene etiquetas,
> no está congelada y no pasó por el doble etiquetado de 1.4. Es la tabla fix → origen → archivos
> que pide 1.1, con la aceptación de cada fila todavía abierta.

## Método

SZZ simple. Para cada commit no-merge de `main` cuyo asunto empieza con `fix` (40 en total):

1. Archivos del fix sin tests, goldens, `.md` ni `openspec/`.
2. Por cada bloque de líneas que el fix borra o cambia, `git blame` en el padre del fix.
3. Origen candidato = los commits más repetidos en ese blame, con la cantidad de líneas entre paréntesis.

Limitaciones conocidas de SZZ: un fix que sólo **agrega** líneas no tiene a quién culpar; un commit
de importación o de rebrand masivo aparece como origen sin haber introducido el defecto; y el blame
apunta a la última línea tocada, que no siempre es la que introdujo el error. Por eso cada fila de
A y B necesita la confirmación manual que pide 1.1 ("descartar los que no se puedan ubicar").

Para reproducirlo (POSIX sh, desde la raíz del repo):

```sh
git log --no-merges --format='%h' main --grep='^fix' -E | while read f; do
  for p in $(git diff --name-only $f^ $f | grep -v -E '_test\.go$|\.golden$|^openspec/|\.md$'); do
    git cat-file -e $f^:"$p" 2>/dev/null || continue
    git diff -U0 $f^ $f -- "$p" | grep '^@@' | sed -E 's/^@@ -([0-9]+)(,([0-9]+))? .*/\1 \3/' |
    while read st n; do [ -z "$n" ] && n=1; [ "$n" = 0 ] && continue
      git blame -l -s -L $st,+$n $f^ -- "$p" | cut -c1-7; done
  done | sort | uniq -c | sort -rn | head -3 | awk -v f=$f '{printf "%s %s(%s)\n", f, $2, $1}'
done
```

## 1. Estrato A: origen acotado y con requisitos escritos

Origen pequeño, reciente y con artefactos OpenSpec (`archive/2026-11-13-claude-native-opsx`), así
que la revisión de referencia puede salir de requisitos y escenarios reales (tarea 2.1).

| Fix | Fecha | Qué corrige | Origen candidato (líneas) | Archivos |
|---|---|---|---|---|
| `fc95a2b` | 13/11 | barras invertidas dobles en mensajes de ruta en Windows | `c19e193`(12) `723f1b2`(8) `8ecd9a9`(4) | 4 |
| `ff7ba15` | 13/11 | uninstall restauraba `settings.json` sin chequear el manifiesto de migración | `c19e193`(4) | 1 |
| `f3fe3c4` | 13/11 | uninstall perdía los modos de `~/.claude` y `CLAUDE.md` | `27a3eea`(7) `c19e193`(6) | 2 |
| `7a9eeb8` | 13/11 | reemplazo de skill instalada sin dir de backup | `97d7f6c`(3) `8ecd9a9`(3) — empate, revisar | 3 |
| `01f944c` | 13/11 | rutas de Windows en avisos de precedencia (fix de `fc95a2b`) | `fc95a2b`(2) | 2 |
| `e0d791b` | 24/07 | uninstall-options no filtraba agentes sin adaptador (fix de `fd01e84`) | `fd01e84`(14) | 3 |

Ojo: los cinco de noviembre vienen del mismo change. Si entran todos, la muestra queda cargada
hacia un solo autor y un solo tipo de defecto (Windows y restore). Conviene tomar dos o tres.

## 2. Estrato B: origen grande, pero blame concentrado

El origen es un `feat` grande. El caso sirve si el paquete se arma con el **diff completo del
origen** (es lo que un revisor habría visto) y el defecto queda localizado. Más caro de etiquetar.

| Fix | Qué corrige | Origen candidato (líneas) | Dominio |
|---|---|---|---|
| `53e22e8` | rollback de backup no contemplaba directorios | `e691901`(17) `6b39617`(15) | backup (ALTO) |
| `7ab8325` | markers viejos nunca se limpiaban | `6a40f17`(16) `e9c69ce`(10) | filemerge (ALTO) |
| `1f127d1` | remover permissions revertía lo que el uninstall ya había sacado | `392984b`(11) | uninstall |
| `ddeb94f` | bloque `permission` de OpenCode no se reemplazaba entero | `253e79a`(10) | permisos |
| `3252036` | backups en Windows nunca encontraban los snapshots | `be27319`(34) | backups |
| `ca5fcde` | sdd-orchestrator no quedaba como primary agent en OpenCode | `5c6bc51`(16) | harness/config |
| `ce348b5` | skills con `SKILL.md` en la raíz no se clonaban | `9d602e3`(6) | harness/skill |
| `fae8bbf` | shape de `command` para MCP locales en OpenCode | `d670c2c`(3) | agents |
| `b13d577` | clave de permisos por agente en el check | `ee26282`(2) | verify |
| `64796ff` | doble clic imprimía usage y cerraba | `8472cf6`(4) | cli |
| `6553fb5` | re-descarga de `engram.exe` con procesos activos | `6db23d1`(9) | external |
| `0d81f65` | tags no-semver al resolver versión de engram | `6db23d1`(8) | external |
| `d670c2c` | engram no operativo en máquina limpia | `7b52941`(14) `fc084a8`(8) | install |
| `7229a44` | compatibilidad del tier estricto de Codex | `9e318ba`(11) `253e79a`(5) `7d6db0e`(4) | permisos — origen disperso, revisar |

## 3. Descartados o en suspenso

| Motivo | Fixes |
|---|---|
| No ubicable: el fix sólo agrega o sólo toca tests/docs | `12ddeb7`, `91bbee6`, `e9c69ce`, `25b8f5a`, `56190a2` |
| Blame de una o dos líneas, origen sin señal | `fd01e84`, `ff66ec4`, `3092573`, `574f5e0`, `ff53e8b`, `fddde4d`, `093ee5d` |
| Instalador gráfico de Windows en C# (`installer/windows/ui`), otro lenguaje y otro revisor | `b9e5f0d`, `c5d580e`, `401f105`, `c7bbf59`, `4585400`, `ce7ac1a`, `b0833c2`, `0b8648d` — **(TBD)**: el grupo decide si es un estrato propio o queda fuera de la primera ronda |
| Origen es un rebrand masivo (`2f8336e`), no introduce defecto | aparece como origen secundario; no se usa |

## 4. Riesgos de esta preparación

- **Fuga de la respuesta.** Este archivo nombra el fix de cada caso y vive en un repo público
  (proposal, "Evidencia"). Un juez con acceso a la red o al historial podría encontrarlo. Antes de
  la primera corrida hay que decidir si la tabla se congela fuera del repo o si el paquete de 1.5
  basta, porque nunca incluye historia posterior al origen. **(TBD)**
- **Sesgo del etiquetador.** Quien armó esta tabla ya sabe la respuesta. No puede ser uno de los dos
  etiquetadores independientes de 1.4; sí puede ser el custodio o el tercero que resuelve.
- **Formato de casos preparados.** Los casos sembrados (1.2) usan el formato de 02 que ya está en
  `main`: `eval/baseline/fixtures/revision/` (`task.json`, `seeded.patch`, `key.json`,
  `reference.json`, `defective.json`) y la regla de emparejamiento de `eval/baseline/review`
  (`review.Match`). No hace falta inventar otro.
