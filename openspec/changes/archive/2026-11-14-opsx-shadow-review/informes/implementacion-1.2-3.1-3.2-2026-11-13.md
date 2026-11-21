# Implementación de 1.2 (parcial), 3.1 y 3.2 de `opsx-shadow-review` (10)

13 de noviembre de 2026. Responsable: Neyen. Ronda `propose`: el piloto no corre. Esto es preparación
externa al binario, para lo que ya no depende de 08 ni de 09.

## Base

| Qué | Valor |
|---|---|
| Rama | `propose/opsx-shadow-review`, merge de `origin/main` `0005a6f` (`d43f3f7`) |
| 05 (D14, redactor) | en `main` `b27b8e0`: `internal/catalog/taskcontext.yaml` y `eval/taskcontext/redact` |
| 06 (fixtures de recibos) | en `main` `0005a6f`: `eval/evidence/testdata/casos/` |
| 02 (revisión sembrada) | en `main`: `eval/baseline/fixtures/revision/` y `eval/baseline/review` |
| 08 / 09 | PR #15 / PR #14, fuera de `main`: 2.1, 4.5 y costo siguen bloqueados |

## Qué se entregó

`eval/shadow/`: no lo importan `cmd/` ni `internal/`, y no envía nada a ningún proveedor.

- **3.1, `package.go`.** `Build` arma el paquete de un caso con exactamente los archivos de la lista
  de inclusión. Rechaza rutas absolutas, con `..`, repetidas, inexistentes o de directorios, y la
  lista vacía. La huella usa la convención de `TreeHash` de 02 y el test la compara contra
  `fixture.TreeHash`.
- **3.2, `scan.go`.** `Scan` examina todos los archivos del paquete con el redactor de 05 y los
  patrones de D14 leídos del catálogo, sin lista propia. También cancela por nombre los archivos que
  cubre D14 (`.env`, `.env.*`, `*.pem`) y los de configuración de usuario de D5 (`settings.json`,
  `config.toml`). Un positivo cancela el envío.
- **1.2, `sample.go` y `muestra/preparados.json`.** Son 5 casos preparados que **referencian** los fixtures de
  02 y 06 con huella fijada por archivo de origen:

  | Caso | Fuente | Clase de D2 | Fixture |
  |---|---|---|---|
  | `prep-02-revision-catalogo` | 02 | diff sembrado (dos defectos de lógica ubicados) | `eval/baseline/fixtures/revision` |
  | `prep-06-a` | 06 | recibo falso con hash correcto | `falso-huella-correcta-reejecucion-contradice` |
  | `prep-06-b` | 06 | mismo HEAD con diff distinto | `mismo-head-otro-diff` |
  | `prep-06-c` | 06 | salida truncada sin original | `truncado-sin-original` |
  | `prep-06-d` | 06 | tests verdes sin cubrir el requisito | `tests-verdes-requisito-incumplido` |

  La etiqueta se lee de la fuente (`key.json` de 02, `esperado.estado` de 06) y no se reescribe. Los
  ids de los casos son neutros porque el id de origen de 06 revela la respuesta.

## Paquetes sin pistas (D2) en los preparados

Los fixtures de origen traen la respuesta adentro. El material que vería el juez la excluye:

- **06:** sólo `requisito`, `recibos`, `entorno` y las copias redactadas referenciadas. No viajan `id`,
  `descripcion` ni `esperado`.
- **02:** sólo `prompt.md` y `seeded.patch`. No viajan `key.json`, `reference.json`, `defective.json`
  ni `task.json`.

`TestPrepared_MaterialHasNoClues` lo exige: busca esos campos, el id de origen y los mecanismos de
los defectos.

## EVIDENCIA

| Hora | Comando | Resultado |
|---|---|---|
| 20:50:46 | `go test ./eval/shadow/` (con stubs) | RED: 17 fallos de aserción; compila y `go vet` pasa |
| 20:51:40 | ídem, implementado | falla sólo `TestScan_Cases/variable-de-entorno` (ver hallazgo) |
| — | prueba directa con el redactor de 05 | `export SERVICE_PASSWORD=…` → `env-assignment:1`; `+export SERVICE_PASSWORD=…` → sin detección |
| 20:52:19 | `go test ./eval/shadow/` con la normalización de diff | `ok` |
| — | mutaciones (sin vista de diff, sin config de usuario, sin `covers`, sin huellas fijadas, aceptar `..`, `key.json` en el material, caso de 06 entero) | los 7 mutantes mueren, cada uno en su test. El de `..` sobrevivía porque el test usaba un archivo inexistente fuera de la raíz; se corrigió el test con un archivo real afuera |
| — | `git status -- cmd internal` | sin cambios |

## Hallazgo

El patrón `env-assignment` de D14 está anclado al inicio de línea (`^\s*(?:export\s+)?…`), así que no
detecta una asignación agregada en un diff (`+SECRETO=valor`). D5 exige examinar los diffs públicos,
porque un commit público puede contener un secreto filtrado. 10 no redefine D14: para los archivos que
son diffs (`.patch`, `.diff` o contenido `diff --git`), examina además cada línea de contenido sin su
marca. Los encabezados `+++` y `---` no se tocan. **Para informar a 05:** el mismo límite afecta a su
redactor si recibe salidas con formato de diff.

## JUICIO (revisable)

- El material de un caso de 02 no incluye el repositorio en la revisión base (`1b4639b`): el juez
  necesitará leerlo. Cómo se entrega el repo es parte del arnés (4.1).
- `entorno` de 06 incluye lo que observaron la fuente independiente y el revisor (por ejemplo,
  «reejecucion contradice»). Es evidencia que la revisión de referencia también tiene, no la respuesta.
  La respuesta es el estado.
- `settings.json` y `config.toml` se rechazan por nombre de archivo, como dice D5. Otros archivos de
  configuración de usuario (por ejemplo, `settings.local.json`) no están en D5 y no se agregaron.

## Lo que queda

- 1.2: defectos de lógica propios, en la cantidad que fije 0.1, y su etiquetado doble (1.4).
- 1.1: mapear los cinco `fix` nuevos de `0005a6f`.
- 2.1 espera a 08 (PR #15). 4.5 y el costo esperan a 09 (PR #14). Las decisiones de grupo de 0.x
  siguen abiertas.
