# R-01 · Organización y herramientas de escritura en Codex

Pruebas del 19 de noviembre de 2026 en macOS con Engram CLI/MCP 3.6.2, el paquete de hooks de Codex 0.1.5 y el servidor compartido Engram Cloud v2.0.0. El proveedor sintético del ensayo de Codex emitió las llamadas; el cliente Codex, el proceso MCP y las escrituras de SQLite fueron reales. Claude queda diferido.

## [EN] Q5 · ¿Se conserva y se puede filtrar por organización?

**Resultado para estas versiones.** La configuración `org` del checkout no llegó a la observación ni a la mutación que envía este cliente. El parámetro MCP tampoco separó resultados. El endpoint de mutaciones del servidor compartido devolvió el marcador de control incluso al recibir un filtro de organización sin coincidencia. El esquema previsto por D1-A (prefijo en la clave) sigue siendo el que se puede ensayar con esta combinación.

| Prueba | Resultado |
|---|---|
| Dos repos locales con el mismo `project_name=r01-org-fixture` y valores `org=alpha` y `org=beta` en sus respectivos `.engram/config.json`. `mem_current_project` en cada repo y un `mem_save` desde cada uno, pasando también su `org` como argumento. | Ambos devolvieron el proyecto correcto. La respuesta de detección omitió `org`. Se crearon dos observaciones, IDs 1 y 2. |
| `mem_search(query="canary", org=...)` con `alpha`, `beta` y `gamma`. | Cada consulta devolvió ambos IDs. `tools/list` no anuncia ese argumento para guardar ni para buscar en 3.6.2. [Sonda del esquema MCP](sonda_engram.py) |
| SQLite de ese proyecto temporal inscrito para sincronización, con autosync desactivado. | `PRAGMA table_info(observations)` no contiene columna `org`. Las dos mutaciones de observación en `sync_mutations` tenían payload sin campo `org`. [Configuración del cliente](https://github.com/Gentleman-Programming/engram/blob/ec2bf0482139360bf61a8eadb76380ffe762de44/internal/project/detect.go#L369-L371) · [Payload de sincronización](https://github.com/Gentleman-Programming/engram/blob/ec2bf0482139360bf61a8eadb76380ffe762de44/internal/store/store.go#L654) |
| Prueba complementaria en Cloud | Las dos consultas recuperaron la observación de control. El filtro incompatible no la excluyó. Las rutas, parámetros operativos y registro del servicio quedan en la evidencia interna. |

La conclusión se circunscribe a las versiones y las rutas medidas. Las versiones posteriores pueden tener otro contrato.

## [EN] Q7 · Qué herramientas afectan el etiquetado

**Seis rutas pueden crear o editar el texto de una observación** en el perfil MCP `all` de 3.6.2. Detectar el nombre es solo el primer paso. El proxy B actual solo implementa reescritura para `mem_save`. La cobertura de esa ruta no se extiende a las otras cinco. [Proxy del piloto](proxy_b.py)

| Herramienta | Efecto observado y consecuencia para una cobertura de rama |
|---|---|
| `mem_save` | Crea o actualiza por clave y permite título, contenido y clave. Un guardado también puede capturar un prompt independiente si está habilitado. Es la ruta cubierta en el piloto. |
| `mem_update` | Editó el ID 1 sin crear otra fila. Los campos son parciales; admite sustituciones `find/replace`. Una edición por ID debe conservar la asociación previa, incluida una edición desde otra rama. La reescritura ciega con la rama actual no verifica esa propiedad. |
| `mem_session_summary` | Creó la fila 5, de tipo `session_summary`, sin clave. El servidor genera el título a partir del proyecto; el cliente entrega `content`. Etiquetar el texto de entrada por sí solo no cubre título y clave. |
| `mem_session_finalize` | Creó la fila 6, sin clave, y cerró la sesión. Repetir sobre la misma sesión no creó otra fila. El argumento de texto es `summary` y la operación es atómica. |
| `mem_capture_passive` | Texto sin bloque reconocible creó cero filas. Dos viñetas bajo `## Key Learnings:` crearon dos filas `passive` sin clave. El etiquetado tendría que preservar cada aprendizaje generado; agregar una cabecera global no demuestra esa cobertura. |
| `mem_consolidate` | La vista previa no escribió. La confirmación con el fingerprint devuelto creó la fila 7 y marcó como eliminadas las dos fuentes. Acepta `new_title`, `new_content` y `topic_key` opcional; debe preservar la coherencia entre vista previa y confirmación. La sonda solo consolidó filas sintéticas. |

Las 21 herramientas restantes se clasificaron por su contrato. La sonda verificó directamente el efecto separado de `mem_save_prompt` y el inicio de sesión. Las demás clasificaciones de esta tabla provienen del código y la referencia de la misma versión. [Registro MCP](https://github.com/Gentleman-Programming/engram/blob/ec2bf0482139360bf61a8eadb76380ffe762de44/internal/mcp/mcp.go) · [Contratos de las herramientas](https://github.com/Gentleman-Programming/engram/blob/ec2bf0482139360bf61a8eadb76380ffe762de44/DOCS.md#mcp-tools-27-tools) · [Consolidación](https://github.com/Gentleman-Programming/engram/blob/ec2bf0482139360bf61a8eadb76380ffe762de44/internal/mcp/consolidate.go) · [Finalización](https://github.com/Gentleman-Programming/engram/blob/ec2bf0482139360bf61a8eadb76380ffe762de44/internal/store/session_finalize.go)

| Entidad/efecto | Herramientas | Tratamiento |
|---|---|---|
| Prompts | `mem_save_prompt` | Pasó de 0 a 1 prompt con el total de observaciones en 2. No corresponde modificar el texto del usuario para etiquetar una observación. |
| Sesiones | `mem_session_start`, `mem_session_end` | Registran y cierran sesiones. El texto opcional del cierre corresponde a la sesión. |
| Estado, borrado y administración | `mem_review`, `mem_pin`, `mem_unpin`, `mem_delete`, `mem_merge_projects` | Revisión o consulta del ciclo, pin local, borrado por ID y migración de identidad de proyecto. No generan prosa nueva para etiquetar por rama. |
| Relaciones | `mem_compare`, `mem_judge`, `mem_judge_many` | Crean o juzgan relaciones y pueden actualizar el estado derivado de observaciones; no redactan nuevo título, contenido ni clave de la observación. |
| Lectura, diagnóstico o sugerencia | `mem_search`, `mem_search_prompts`, `mem_context`, `mem_stats`, `mem_timeline`, `mem_get_observation`, `mem_current_project`, `mem_list_projects`, `mem_doctor`, `mem_suggest_topic_key` | No requieren reescritura. Sugerir una clave no la guarda. |

El matcher nativo de Engram y el etiquetado de rama cumplen funciones distintas. El primero despacha ampliamente y el manejador de 3.6.2 usa una lista interna. El control con el binario reconoció `mem_save`, `mem_update`, `mem_session_summary`, `mem_capture_passive` y `mem_save_prompt` con ambos prefijos. En ese mismo control `mem_session_finalize`, `mem_consolidate`, `mem_judge_many` y `mem_search` devolvieron `{}`. Esto confirma que las nuevas herramientas de escritura no heredan automáticamente el enlace de sesión del manejador. [Clasificación del binario](https://github.com/Gentleman-Programming/engram/blob/ec2bf0482139360bf61a8eadb76380ffe762de44/cmd/engram/hook_claude.go#L18-L44) · [Uso por Codex](https://github.com/Gentleman-Programming/engram/blob/ec2bf0482139360bf61a8eadb76380ffe762de44/cmd/engram/hook_claude.go#L496-L560)

## [EN] Q8 · Identidad de la herramienta por capa

| Capa | Identidad y evidencia |
|---|---|
| Servidor MCP | `tools/list` anuncia `mem_save`; `tools/call` lleva ese nombre lógico. El proxy stdio ve esa superficie y puede identificar al MCP por el registro al que sirve. |
| Codex, registro directo `engram` | Evento de hook real con `tool_name=mcp__engram__mem_save`. En el ensayo padre/hijo, las dos llamadas dejaron su propio `tool_use_id`. [Trazas del observador](codex-diagnostico-observador.json) |
| Codex, registro alternativo explícito `plugin_engram_engram` | Cliente real en perfil aislado y proveedor local sintético: thread `id-local-026`, llamada `llamada-011`. El rollout contiene namespace `mcp__plugin_engram_engram` y nombre `mem_save`. `codex exec --json` registró `server=plugin_engram_engram`, `tool=mem_save`, `status=completed`, y Engram devolvió ID 1. Se configuró el alias expresamente; el resultado no implica que el instalador genere ese alias. Este ensayo midió el despacho; el observador de hook en vivo correspondió al registro directo. |
| Manejador de Engram | El manifiesto Codex 0.1.5 declara literalmente `mcp__engram__mem_*\|mcp__plugin_engram_engram__mem_*`. Con inputs de control para `mcp__engram__mem_save` y `mcp__plugin_engram_engram__mem_save`, el mismo binario clasificó ambos como escritores. Para un matcher acotado a las seis rutas de Q7 en un host con expresiones regulares, el patrón es `^(mcp__engram__\|mcp__plugin_engram_engram__)(mem_save\|mem_update\|mem_session_summary\|mem_session_finalize\|mem_capture_passive\|mem_consolidate)$`. Es una especificación de selección; la cobertura por cada método requiere su adaptador. [Manifiesto Codex](https://github.com/Gentleman-Programming/engram/blob/ec2bf0482139360bf61a8eadb76380ffe762de44/plugin/codex/hooks/hooks.json#L29-L38) |


## Repetición local de Q5 y Q7

Desde el directorio de este documento, con Python 3, Git y Engram 3.6.2 en PATH. La tabla de conteos que imprime tiene el orden observaciones totales (incluye eliminadas lógicamente), prompts y sesiones. La inscripción y la consolidación se ejecutan exclusivamente en una base temporal con autosync desactivado. El helper de stdio se comparte con [la sonda local](sonda_engram.py).

```sh
PYTHONDONTWRITEBYTECODE=1 python3 - <<'PY'
from pathlib import Path
import json, os, sqlite3, subprocess, tempfile
from sonda_engram import llamada-012

PROJECT = "r01-org-fixture"
with tempfile.TemporaryDirectory() as tmp:
    root = Path(tmp)
    env = {"PATH": os.environ["PATH"], "HOME": str(root),
           "ENGRAM_DATA_DIR": str(root / "db"), "ENGRAM_CLOUD_AUTOSYNC": "0",
           "GIT_CONFIG_GLOBAL": os.devnull, "GIT_CONFIG_NOSYSTEM": "1"}
    dbpath = root / "db/engram.db"
    def query(sql):
        with sqlite3.connect(dbpath.as_uri() + "?mode=ro", uri=True) as db:
            return db.execute(sql).fetchall()
    def counts():
        return [query("select count(*) from " + table)[0][0]
                for table in ("observations", "user_prompts", "sessions")]
    def run(name, **args):
        r = llamada-012(repo, env, "tools/call", {"name": name, "arguments": args})
        text = r["content"][0]["text"]
        if r.get("isError"):
            raise RuntimeError(name + " " + text)
        print(name, counts())
        try:
            return json.loads(text)
        except ValueError:
            return text
    for org in ("alpha", "beta"):
        repo = root / org
        subprocess.run(["git", "init", "-q", repo], env=env, check=True)
        (repo / ".engram").mkdir()
        (repo / ".engram/config.json").write_text(json.dumps(
            {"project_name": PROJECT, "org": org}))
        if org == "alpha":
            subprocess.run(["engram", "cloud", "enroll", PROJECT], cwd=repo, env=env, check=True)
        current = run("mem_current_project")
        assert current["project"] == PROJECT and "org" not in current
        run("mem_save", title=org, content=f"canary {org}", org=org,
            type="discovery", topic_key=org, capture_prompt=False)
    for org in ("alpha", "beta", "gamma"):
        result = run("mem_search", query="canary", org=org)
        assert sorted(x["id"] for x in result["results"]) == [1, 2]
    assert "org" not in [r[1] for r in query("pragma table_info(observations)")]
    payloads = query("select payload from sync_mutations where entity='observation'")
    assert len(payloads) == 2 and all("org" not in json.loads(p[0]) for p in payloads)
    print("Q5: 2 outbound observation payloads; no org field")
    inv = llamada-012(repo, env, "tools/list", {})["tools"]
    print("tools:", len(inv))
    run("mem_session_start", id="fixture-session", directory=str(repo))
    run("mem_save_prompt", session_id="fixture-session", content="Synthetic prompt")
    run("mem_update", id=1, expected_project=PROJECT, content="Updated canary alpha")
    run("mem_capture_passive", session_id="fixture-session", content="No list here")
    run("mem_capture_passive", session_id="fixture-session",
        content="## Key Learnings:\n- A synthetic check exercises passive observation capture.\n- A different synthetic learning covers the second generated observation.")
    run("mem_session_summary", session_id="fixture-session", content="Synthetic summary.")
    run("mem_session_finalize", session_id="fixture-session", summary="Synthetic finalization.")
    run("mem_session_finalize", session_id="fixture-session", summary="Synthetic retry.")
    fields = dict(ids_to_delete=[1, 2], new_title="Consolidated observations",
                  new_content="Synthetic consolidated content.", expected_project=PROJECT)
    preview = run("mem_consolidate", **fields)
    fp = preview["plan"]["fingerprint"]
    run("mem_consolidate", **fields, confirm=True, fingerprint=fp)
    print("q7 records:", query("select id,type,topic_key,deleted_at is not null from observations order by id"))
PY
```

## QA macOS de la entrega

**Revalidación del 19/11/2026 sobre `17344a5`.** La rama de R-01 y `origin/main` apuntaban a `huella-001`. Para el candidato se copió el checkout con el paquete local de R-01. Para el control se creó un clon Git limpio y se seleccionó ese mismo SHA en modo detached. Se ejecutó la suite Go completa en ambos entornos.

Las pruebas se ejecutaron con una copia temporal del repositorio y caches de Go. Se verificó la integridad del entorno original antes y después.

| Comando desde la raíz del repositorio | Resultado |
|---|---|
| `PYTHONDONTWRITEBYTECODE=1 python3 -m unittest discover -s docs/research/r01-prototipo` | 14 pruebas, OK. |
| `PYTHONDONTWRITEBYTECODE=1 python3 -m unittest discover -s assets/skills/active-roadmap` | 27 pruebas, OK. |
| `PYTHONDONTWRITEBYTECODE=1 python3 -m unittest discover -s assets/skills/active-delivery` | 55 pruebas del contrato team.py, OK. |
| `python3 docs/research/r01-prototipo/sonda_cloud.py --offline` | Exit 0. Lectura por ID de ambos scopes en el emisor y receptor aislado. |
| `python3 docs/research/r01-prototipo/sonda_engram.py` | Exit 0. En Q4, tras renombrar la segunda observación a la clave ocupada, ambas mostraban `Topic: k`. El siguiente guardado devolvió #1 en la revalidación en 17344a5. La corrida de QA anterior había devuelto #2. |
| `go vet ./...` | Exit 0 con Go 1.27.3 en macOS. |
| `go test ./... -count=1` | Exit 1 en candidato y control. En ambos pasaron 51 paquetes y fallaron los mismos cuatro tests en dos paquetes, con los mismos diagnósticos. |

- En `cmd/active-stack`, `TestWindowsBundleThemeContainsMinimalInstallerPages` intentó abrir la ruta literal `..\..\installer\windows\theme\ActiveStackTheme.xml` y recibió `no such file or directory` en macOS.
- En `eval/baseline/fixture`, el escenario de revisión materializó la huella `huella-002`; el fixture exige `huella-003`. Lo reportaron `TestCheckout_RepetitionsStartIdenticalAndRestore`, `TestEvaluate_ReviewDispositionsAndMechanism` y `TestEvaluate_ReferenceAcceptedDefectiveRejected`.


### [EN] V12 · Reproducción pendiente en Windows

La ejecución de controles locales puede iniciarse con `py -3 -B -m unittest discover -s docs/research/r01-prototipo -v` y `py -3 -B docs/research/r01-prototipo/sonda_engram.py`. La verificación nativa con Codex y Engram continúa pendiente. La sonda de sincronización usa selección de pipes Unix.
