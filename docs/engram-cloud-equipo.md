# Memoria de equipo con Engram Cloud

> Guía del harness `engram-cloud`. Describe solo lo que está construido y mergeado en `main`. Lo que depende de una prueba pendiente figura como `(TBD)`.
> Diseño: [plan Engram Cloud](plan-engram-cloud-equipo.md) · Arquitectura: [ARCHITECTURE.md §5.1](../ARCHITECTURE.md) · Equipo: [equipo.md](equipo.md)

## Qué es y qué no es

`engram-cloud` prepara a Claude y a Codex para que la memoria compartida de un equipo en Engram Cloud distinga el trabajo **por rama y autor**. Es un harness `config` (Full lo instala; Custom permite elegirlo) que depende de `engram` y `sdd-orchestrator`. **Nada se activa solo**: un proyecto entra en el modo cuando tiene el marcador `.engram/active-stack.json` con `{"memory_policy": "cloud-branch"}`. Sin marcador, ningún guardado se reescribe y la sección del agente no aplica.

**No instala ni administra el servidor de Engram Cloud, ni el plugin de Engram de Claude, ni las credenciales de nadie.**

## Esquema de etiquetas y autor

El etiquetado lo hace código del stack (`active-stack memory proxy`, un proxy MCP entre el agente y Engram), no el modelo: no cuesta tokens y no depende de que el agente se acuerde.

| Campo | En la rama principal | En otra rama |
|---|---|---|
| `topic_key` | `K` sin cambios | `branch/<slug>/K` (si el agente no dio `K`, queda vacío) |
| `title` | `T` sin cambios | `[<slug>] T` |
| Primera línea de `content` | `Author: <nombre>` | `Branch: <rama> · Author: <nombre>` |
| Resto de los campos | sin cambios | sin cambios |

- **Alcance:** `mem_save` y las vías narrativas (`mem_session_summary`, `mem_session_finalize`, `mem_capture_passive`, `mem_save_prompt`), que reciben el encabezado pero conservan clave y título. El resto de las herramientas pasa sin tocar.
- **Autor:** `git config user.name` (si falta, el usuario del sistema). Nunca el correo. Engram no guarda autor por observación; como hace *upsert* por clave, el autor visible es el último que editó.
- **Rama principal:** la de `origin/HEAD`; si no existe, `main` o `master`. En *detached HEAD* el slug es `detached-<sha7>`. Sin repositorio o sin remoto no se etiqueta.

### Regla del slug (`active-stack memory branch`)

1. Se pasa a **minúsculas**; todo carácter fuera de `[a-z0-9]` pasa a `-`; los guiones se colapsan y se recortan.
2. Máximo 40 caracteres. Si el slug **pierde información respecto de la rama** (por ejemplo `/`, `_` o mayúsculas) o supera el máximo, se recorta a 33 y se agrega `-` más los **6 primeros hex del SHA-256 del nombre completo de la rama**. Así `feat/una` y `feat-una` no comparten etiqueta. Una rama ya escrita como slug se devuelve igual.
3. El nombre del proyecto en `.engram/config.json` también va en minúsculas, igual que Engram.

`active-stack memory branch --json` devuelve `{"mode","tagged","branch","slug"}` para comprobarlo.

### Lectura

Los hooks de lectura (`session-start` y `pre-read`) agregan al contexto del agente la **rama actual**, solo en proyectos con marcador, y nunca bloquean: ante cualquier error salen con 0 y avisan por `stderr`. Las reglas de la sección del agente: sin etiqueta = verdad del equipo; etiqueta de mi rama = válido para esta rama; etiqueta de otra rama = **no verificado**; ante contradicción gana la sin etiqueta hasta el merge.

## La garantía y sus condiciones

El etiquetado se declara **por agente**, no en general. Vale cuando se cumplen todas:

1. El proxy está registrado en la entrada MCP `engram` del agente (`~/.claude.json` o `~/.codex/config.toml`).
2. `active-stack` está en el `PATH` (si no, el proxy no arranca y el guardado sale **sin etiqueta**).
3. El proyecto tiene el marcador.
4. En Codex, el hook de lectura fue revisado y confiado (`/hooks`); si el hook cambia, se omite hasta revisarlo de nuevo.
5. En Claude, el plugin de Engram está instalado y actualizado (el stack no lo instala ni lo verifica).

**Fuera de esas condiciones rige solo el respaldo por protocolo, que es una convención y no una garantía.** El proxy falla abierto: ante un error o más de 3 s reenvía el guardado original, así que un fallo nunca bloquea al agente, pero ese guardado queda sin etiqueta.

**Por qué proxy y no hook (DA-13).** Un hook de guardado corre en paralelo con el del plugin de Engram y su reescritura se descartó bajo carga en Claude (9/12 contra 12/12 con proxy). No se instala ningún hook de guardado; los de lectura se mantienen.

**Estado de la prueba (DA-12).** Claude + Windows pasó V1–V5, V7, V9, V12, V13 y las vías narrativas; V8 interactivo y V9 con el proxy terminado no se ejecutaron; en Codex la verificación es parcial. La matriz V1–V13 completa en Claude y Codex, con evidencia, es la ficha R-23 `(TBD)`. Linux y macOS: no ejecutado.

**Codex real en Windows, 2026-11-23:** [PP-01 a PP-04](equipo/informes/PP-codex.md) verificaron lectura con hooks confiados, guardados etiquetados, cambio de rama, worktrees y subagente (sin `--ephemeral`). **Límites DA-12:** el binario probado borra la confianza al reinstalar; hay corrección con TDD pendiente de revalidación con binario actualizado. El plugin Engram 0.1.5 bloquea el guardado por sesión Codex no registrada, incluso en una sesión nueva no efímera; causa raíz `(TBD)`. El deny de push de Codex no está entregado y un push al bare local pasó. Full instala proxy/hooks/AUTOSYNC, pero su retiro falla por markers de hooks desparejos y no devuelve los bytes originales. No se declara la garantía completa por agente. Cloud real y otros sistemas siguen `(TBD)`.

`active-stack verify` lo comprueba sin escribir nada (solo con el harness instalado y solo en Claude y Codex): marcador una vez (`config:engram-cloud:<agente>`), hooks (`cloud:hooks`), `active-stack` en el `PATH` (`cloud:path`, **aviso**, no revierte), proxy registrado (`cloud:proxy`), `ENGRAM_CLOUD_AUTOSYNC` (`cloud:autosync`, aviso), confianza del hook en Codex (`cloud:trust`, aviso), OpenCode/Antigravity presentes (`cloud:unsupported`, aviso) y los archivos de `.engram/` (`cloud:project-files`). Ningún check lee ni imprime tokens. En una instalación Project no se exigen hooks, proxy ni autosync (son de nivel usuario).

## Cierre de ficha: promover lo durable

Una memoria etiquetada tiene otra clave (`branch/<slug>/K`), así que Engram no la fusiona con `K`. Sin promoción, lo que entró a `main` quedaría «no verificado» para siempre. Al cerrar la ficha, por protocolo:

1. Solo después de verificar con git que la rama se mergeó, y solo memorias propias.
2. **Solo lo durable** (decisión, arquitectura, patrón, bug con causa raíz reutilizable, gotcha): si la clave destino `K` está libre, `mem_update` quita la etiqueta; si existe, se actualiza la existente.
3. **Lo efímero no se toca ni se borra** (resúmenes, `opsx/*`, próximos pasos): queda etiquetado y la regla de lectura lo ignora.

Este paso lo hace el agente por protocolo (no es código del stack) y es el único con costo en tokens.

## Limitación: OpenCode y Antigravity

No están cubiertos. Lo que guarden desde una rama irá **sin etiqueta** y entrará a la bolsa común. Está declarado en el catálogo y `verify` avisa si están entre los agentes y el proyecto usa el modo. **En ramas, usar Claude o Codex.**

## Qué es automático y qué no

| Paso | Estado |
|---|---|
| Etiquetar y poner autor al guardar | Automático (proxy) |
| Hooks de lectura y `ENGRAM_CLOUD_AUTOSYNC=1` en el MCP | Los instala `active-stack install` (solo Machine, Claude y Codex) |
| Subir y bajar entre personas | Automático con `ENGRAM_CLOUD_AUTOSYNC=1` |
| `.engram/config.json` y `.engram/active-stack.json` | `active-stack memory init` (fase `memory-policy` de la fundación); **solo propone** el commit |
| Promoción al cerrar la ficha | Por protocolo, solo lo durable |
| Alta de cada persona (token, servidor, enrolar) | **Manual, una sola vez**: las credenciales no las escribe el stack |
| Confianza del hook en Codex | **Manual, una vez** (`/hooks`) |

`active-stack uninstall` quita proxy, hooks y autosync sin tocar hooks ajenos ni `.engram/`; quedan solo `.engram/config.json`, `.engram/active-stack.json` y las memorias.

**Excepción observada en Codex/Windows:** el retiro Full de PP-03 falló; la afirmación anterior no está garantizada para ese caso. Conservar el backup y revisar la [evidencia de PP-03](equipo/informes/PP-codex.md#pp-03--install-y-uninstall) antes de dar por restaurada la configuración.

## Servidor (infraestructura del equipo)

No lo administra el stack. Recomendado en el plan (Apéndice B), a cargo del administrador del equipo:

- `engram cloud serve` con Postgres, imagen con **versión fija**, detrás de **HTTPS** (el sync comprime pero no cifra).
- `ENGRAM_CLOUD_ALLOWED_PROJECTS=<repo>` (nunca `*`); un token por persona con `--issue-token` y `--grant-project <repo>`; `ENGRAM_CLOUD_TOKEN_PEPPER` distinto de `ENGRAM_JWT_SECRET`; **nunca** `ENGRAM_CLOUD_INSECURE_NO_AUTH`.
- Revisión por dashboard y registro de auditoría; para contener, pausar el sync del proyecto. Backups de Postgres: el Production Checklist de la documentación de Engram Cloud.

Limitaciones heredadas de Cloud (según su documentación): no hay token de solo lectura; el dashboard solo lee; un `topic_key` existente se sobrescribe sin versiones; `scope: personal` no es privacidad; no hay autor por observación.

## Alta de cada persona (una sola vez)

1. `active-stack install` en Full, o Custom con `engram`, `sdd-orchestrator` y `engram-cloud` (Claude o Codex). Requiere Engram 3.7.x y, en Claude, el plugin de Engram actualizado.
2. En Codex: abrir `/hooks` y **confiar** el hook de `active-stack`.
3. Comprobar que el repo trae `.engram/config.json` y `.engram/active-stack.json` (los crea `active-stack memory init`).
4. Pedir al administrador su token personal (se muestra una vez) y definir `ENGRAM_CLOUD_TOKEN` en su entorno. **Nunca en el repo.**
5. `engram cloud config --server https://<host>` y `engram cloud enroll <repo>`.
6. Verificar `ENGRAM_CLOUD_AUTOSYNC=1` (lo agrega el stack; `active-stack verify` lo revisa).
7. Prueba de humo: guardar algo desde una rama y comprobar en el dashboard que aparece **etiquetado y con autor**.

## Documentación oficial consultada

Antes de tocar esta integración, leer la documentación vigente y anotar las versiones de Engram, Claude Code y Codex usadas.

| Recurso | Enlace |
|---|---|
| Repositorio de Engram | https://github.com/Gentleman-Programming/engram |
| README | https://github.com/Gentleman-Programming/engram/blob/main/README.md |
| Referencia completa (esquema, CLI, sync, MCP tools) | https://github.com/Gentleman-Programming/engram/blob/main/DOCS.md |
| Uso en equipo | https://github.com/Gentleman-Programming/engram/blob/main/docs/TEAM-USAGE.md |
| Arquitectura (topic keys, higiene de memoria) | https://github.com/Gentleman-Programming/engram/blob/main/docs/ARCHITECTURE.md |
| Setup de agentes (Claude Code, Codex, hooks, autosync, conflictos) | https://github.com/Gentleman-Programming/engram/blob/main/docs/AGENT-SETUP.md |
| Engram Cloud | https://github.com/Gentleman-Programming/engram/blob/main/docs/engram-cloud/README.md |
| Issues del repo (para proponer mejoras) | https://github.com/Gentleman-Programming/engram/issues |
| Hooks de Claude Code | https://code.claude.com/docs/en/hooks |
| Hooks de Codex | https://developers.openai.com/codex/hooks (redirige a https://learn.chatgpt.com/docs/hooks) |

## Pendiente `(TBD)`

- Matriz V1–V13 completa en Claude y Codex, con evidencia: R-23.
- Codex real: confianza inicial y subagentes verificados en PP-codex; falta cerrar X6 con binario actualizado y resolver los límites de la matriz, especialmente la coexistencia del plugin.
- Engram como plugin de Claude (su `.mcp.json` vive en la caché del plugin): qué hace el registro del proxy si no hay entrada `engram` en `~/.claude.json` no está decidido.
- Idioma de los mensajes de `verify`: sale en inglés por defecto hasta que la CLI exponga el idioma.
- Detectar el proyecto actual en una instalación Machine con el directorio de trabajo.
- Si el `PreToolUse` propio de Engram pisa algo, abrir un issue en el repositorio de Engram con la evidencia.
