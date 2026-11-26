# Tiempo y costo del cambio

Registrá el trabajo desde el primer bloque del cambio, aunque la rama todavía
sea provisoria. Incluí exploración, propuesta, implementación, verificación,
revisiones, correcciones y el trabajo de integración hasta el merge. Conservá
por separado los intervalos de actividad y el consumo por modelo.

## Capturar durante el trabajo

1. Al iniciar y detener cada bloque, registrá la hora observada con zona
   horaria. En una terminal POSIX, usá `date -u +%Y-%m-%dT%H:%M:%SZ`. En
   PowerShell, usá `[DateTimeOffset]::UtcNow.ToString("yyyy-MM-dd'T'HH:mm:ss'Z'")`. Volvé a registrar
   la hora al retomar, aunque sea en otro cliente o con otro integrante. El
   cálculo une los minutos concurrentes y mide el tiempo calendario hasta el
   merge por separado. El trabajo realizado mientras GitHub procesa la
   revisión o sus checks lleva su propio intervalo.
2. Antes del primer turno medido, aislá la sesión al cambio o guardá la lectura
   acumulada inicial por modelo y categoría. Volvé a medir al suspender,
   cambiar de modelo o entregar la rama. En una sesión reutilizada, registrá
   los enteros inicial y final de cada contador. Reuní también el consumo del
   revisor, los intentos fallidos facturables y los subagentes.
3. Transcribí únicamente identificadores opacos, fuentes, modelos y contadores
   necesarios. Cargá una fila por delta medido y categoría facturable o una
   por par de lecturas del mismo contador. Conservá IDs estables al copiar
   una fila de un registro a otro. Elegí solicitudes crudas normalizadas o su
   agregado reconciliado, según la cobertura demostrada de cada fuente.

### Dónde leer los tokens

| Cliente | Captura utilizable | Límites |
|---|---|---|
| **Codex interactivo** | En clientes que exponen [App Server](https://learn.chatgpt.com/docs/app-server), guardá las lecturas inicial y final de `thread/tokenUsage/updated` del hilo de trabajo y de cada hilo auxiliar que corresponda. En la CLI que guarda historial local, podés seleccionar el JSONL de una sesión conocida y extraer los contadores con el filtro de abajo. Abrí una sesión por modelo si el desglose del historial no identifica cada cambio de modelo. | Los porcentajes de uso y contexto en la interfaz no identifican categorías facturables. El contador acumulado exige modelado de toda la sesión o un delta de un intervalo aislado. Si un agregado del principal puede incluir subagentes, reconciliá con el detalle y usá una sola fuente con cobertura comprobada. [Esquema de historial de Codex](https://github.com/openai/codex/blob/main/codex-rs/app-server/tests/common/rollout.rs) |
| **Claude Code interactivo** | Antes de abrir la terminal de trabajo, habilitá OpenTelemetry con `CLAUDE_CODE_ENABLE_TELEMETRY=1` y `OTEL_LOGS_EXPORTER` según el colector autorizado. Filtrá `claude_code.api_request` por sesión y ventana del cambio. Exportá solo `request_id`, tiempo, `model`, modalidad y contadores facturables. Duplicados de la misma solicitud conservan el mismo ID. | `/usage` sirve para contrastar el alcance por modelo. La vista puede redondear los valores. Conservá los enteros de los eventos. Si el precio distingue TTL de caché, verificá que los contadores por TTL sumen el total escrito. [Monitoreo de Claude Code](https://code.claude.com/docs/en/monitoring-usage) |
| **Ejecución no interactiva** | En Codex `exec --json`, tomá `turn.completed.usage` y el modelo confirmado de esa ejecución. En Claude Code `-p --output-format json`, leé `modelUsage` del resultado final. | En Claude, el agregado `modelUsage` incluye subagentes. En Codex, contrastá si el registro del principal los incluye antes de sumar. Si estos formatos omiten una categoría diferenciada en la tarifa o una ejecución facturable, buscá los datos exactos en la fuente del proveedor. [Medición en Active Stack](https://github.com/Group-Active-IA/active-stack/tree/main/eval/baseline) |
| **API o cliente con exportación propia** | Tomá el modelo de la respuesta y los contadores facturables observados en `usage` o el informe de facturación, filtrado por las solicitudes del cambio. | Si solo hay una cifra total, falta el modelo efectivo o queda uso pendiente de atribuir, dejá el registro incompleto. |

En el JSONL de una sesión local de Codex **ya identificada**, este filtro
muestra solo tiempo y contadores. Aplicalo antes de compartir la salida.
Comprobá los nombres disponibles según la versión del cliente.

```sh
jq -c '
  select(.type == "event_msg" and .payload.type == "token_count")
  | select(.payload.info.total_token_usage != null)
  | {timestamp, totals: (.payload.info.total_token_usage
       | {input_tokens, cached_input_tokens, cache_write_input_tokens, output_tokens})}
' "rollout-de-la-sesion.jsonl"
```

En OpenAI, los tokens leídos y escritos a tarifas de caché forman parte de la
entrada total. Derivá la entrada ordinaria restando las categorías de caché que
correspondan. El razonamiento forma parte de la salida facturable del modelo.
En Claude Code, la entrada de `api_request` ya excluye lecturas y escrituras
de caché. Si una versión o un cliente no expone todas las categorías requeridas,
buscá el desglose de la API o telemetría compatible. [OpenAI Prompt Caching](https://developers.openai.com/api/docs/guides/prompt-caching) · [Claude Messages API](https://platform.claude.com/docs/en/api/typescript/messages)

## Confirmar tarifas y merge

1. Consultá la tarifa del proveedor aplicable cuando ocurrió el uso. Registrá
   una fila por modelo exacto, categoría y nivel tarifario, con USD por millón,
   URL de referencia y fecha de consulta. Si cambia la tarifa, separá el
   consumo antes y después. Una categoría de precio cero requiere la misma
   evidencia que una tarifa positiva. [OpenAI](https://developers.openai.com/api/docs/pricing)
   · [Anthropic](https://platform.claude.com/docs/en/about-claude/pricing)
   · [Gemini](https://ai.google.dev/gemini-api/docs/pricing)
2. Al cierre, consultá el PR por su URL. Usá exactamente los campos de abajo
   como objeto `pr`. Verificá en Git que el SHA de merge es ancestro de la
   rama remota de destino tras actualizarla. [GitHub CLI](https://cli.github.com/manual/gh_pr_view)

```sh
gh pr view <url-del-pr> --json state,headRefName,baseRefName,mergedAt,mergeCommit,url
git fetch <remoto-del-pr> <rama-destino>
git merge-base --is-ancestor <sha-del-merge> FETCH_HEAD
```

## Validar y guardar

Cargá el registro en JSON y corré el validador con Python 3.8 o superior.
La normalización de las fuentes de tokens es manual. Conservá el informe
validado junto al cambio archivado en SDD o junto al handoff en ODD.

```sh
python -B <directorio-de-active-delivery>/usage.py <evidencia.json>
```

El proceso escribe JSON y devuelve código 0 cuando la evidencia tiene el
formato completo. Un error identifica el campo faltante o incoherente. El
valor `usage_status: "complete"` afirma cobertura comprobada de todo el
cambio. Usá `"none"` con `usage` y `rates` vacíos si se verificó que el
cambio se hizo sin IA. Ante uso, tarifa, tiempo o merge sin verificar, dejá
el cierre pendiente hasta recuperar evidencia.

El informe conserva los intervalos, las fuentes, los consumos medidos, las
tarifas y el merge. Cada subtotal es `tokens × tarifa / 1.000.000`, sin
redondear la operación. `api_rate_cost_usd` suma los usos marcados `api`.
`subscription_api_equivalent_usd` suma el equivalente API de las sesiones
marcadas `subscription`. El total combinado expresa ambos a tarifas API.
Para conocer el cargo efectivo, consultá el registro de facturación del
proveedor.

### Ejemplo de entrada con cifras ficticias

```json
{
  "branch": "C-01-integration",
  "pr": {
    "state": "MERGED",
    "headRefName": "C-01-integration",
    "baseRefName": "main",
    "mergedAt": "2026-11-25T14:00:00Z",
    "mergeCommit": {"oid": "0123456789abcdef0123456789abcdef01234567"},
    "url": "https://github.com/example/project/pull/1"
  },
  "intervals": [
    {"start": "2026-11-25T13:00:00Z", "end": "2026-11-25T13:45:00Z", "source": "reloj de la ficha"}
  ],
  "usage_status": "complete",
  "rates": [
    {
      "id": "example-standard-input",
      "provider": "Example",
      "model": "example-1",
      "tier": "standard",
      "category": "input",
      "usd_per_million": "2.50",
      "source": "https://example.com/pricing",
      "checked_on": "2026-11-25"
    },
    {
      "id": "example-standard-output",
      "provider": "Example",
      "model": "example-1",
      "tier": "standard",
      "category": "output",
      "usd_per_million": "8.00",
      "source": "https://example.com/pricing",
      "checked_on": "2026-11-25"
    }
  ],
  "usage": [
    {
      "id": "example-session-1-input",
      "source": "lectura de sesión",
      "billing": "subscription",
      "rate_id": "example-standard-input",
      "tokens": 12000
    },
    {
      "id": "example-session-1-output-initial-block",
      "source": "contador de sesión al inicio y al suspender",
      "billing": "subscription",
      "rate_id": "example-standard-output",
      "counter_id": "example-session-1-output",
      "start": "2026-11-25T13:00:00Z",
      "end": "2026-11-25T13:45:00Z",
      "before": 500,
      "after": 4500
    }
  ]
}
```
