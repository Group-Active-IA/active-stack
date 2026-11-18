# Configuraciones ilustrativas por agente

Fecha de consulta: **2 de noviembre de 2026**. Anexo práctico para el informe de harnesses. Los bloques muestran esquemas documentados y plantillas propias; **no están instalados ni ejecutados**. Cada comando apunta a un script hipotético que todavía debe implementarse y probarse. No pegar los bloques completos sobre configuraciones existentes: el instalador debe usar backup, adapter y merge idempotente.

## 1. Antes de elegir el formato

Una skill contiene un procedimiento que el modelo interpreta. Un especialista agrega contexto y, cuando el runtime lo permite, herramientas acotadas. Un prehook ejecuta lógica antes de una acción. Un posthook observa el resultado; no revierte automáticamente la acción. Un mensaje entre agentes transporta contexto, pero no establece aislamiento de archivos ni permisos.

Para el stack propongo un núcleo de políticas compartidas con adapters por host. El núcleo devuelve una decisión interna; cada adapter traduce payloads y respuestas al contrato exacto del producto. No compartir directamente un `hooks.json` entre hosts aunque tengan eventos con nombres parecidos.

Ejemplo **de contrato propio, no de una API existente**:

```json
{
  "policy": "user-config-change",
  "effect": "observe | block | defer-to-runtime",
  "reason": "evidencia verificable",
  "context": "instrucción breve opcional"
}
```

No hay `allow-everything`: si la política no tiene una decisión, conserva el flujo normal de permisos. Un clasificador de complejidad no concede autorización. Un hook de verificación tampoco debe ejecutar builds cuando el operador los prohíbe.

## 2. Claude Code

Configuración de proyecto: `.claude/settings.json`. Ejemplo de prehook; `policy.cjs` es un archivo pendiente de implementación:

```json
{
  "hooks": {
    "PreToolUse": [{
      "matcher": "Bash|PowerShell",
      "hooks": [{
        "type": "command",
        "command": "node scripts/agent-hooks/policy.cjs claude",
        "timeout": 5
      }]
    }]
  }
}
```

Respuesta de bloqueo para ese evento:

```json
{
  "hookSpecificOutput": {
    "hookEventName": "PreToolUse",
    "permissionDecision": "deny",
    "permissionDecisionReason": "La acción excede la autorización registrada."
  }
}
```

No usar ese JSON como respuesta universal de `Stop` o `PostToolUse`. Los contratos varían por evento. [Referencia oficial de hooks](https://code.claude.com/docs/en/hooks).

Especialista ilustrativo `.claude/agents/contract-reviewer.md`:

```markdown
---
name: contract-reviewer
description: Revisa contratos y regresiones sin modificar archivos.
tools: Read, Grep, Glob
model: inherit
---
Compará el cambio con su spec. Devolvé evidencia por archivo,
riesgos concretos y pruebas faltantes. No edites ni ejecutes shell.
```

La lista de herramientas restringe mejor que una instrucción sola. `isolation: worktree` es una opción separada para workers de escritura; su interacción con teams debe verificarse. Los subagentes y teams no son lo mismo: el segundo agrega comunicación compartida y continúa experimental. [Subagentes](https://code.claude.com/docs/en/sub-agents), [Teams](https://code.claude.com/docs/en/agent-teams).

## 3. Codex

La documentación actual admite `.codex/hooks.json` y hooks inline en `.codex/config.toml`. Preferir una representación por capa: las fuentes se acumulan, no reemplazan automáticamente las anteriores. Los hooks no administrados requieren revisión y confianza. [Hooks oficiales](https://learn.chatgpt.com/docs/hooks).

Plantilla `.codex/hooks.json`, con handler hipotético:

```json
{
  "hooks": {
    "PreToolUse": [{
      "matcher": "Bash",
      "hooks": [{
        "type": "command",
        "command": "node scripts/agent-hooks/policy.cjs codex"
      }]
    }]
  }
}
```

La selección del matcher debe comprobarse contra las herramientas reales: el nombre expuesto por el host y el del hook no siempre son iguales. No traducir `ask`, `updatedInput` o controles de continuación desde Claude sin revisar qué acepta esa versión de Codex.

Especialista de proyecto `.codex/agents/contract-reviewer.toml`:

```toml
name = "contract_reviewer"
description = "Revisa contratos y pruebas; devuelve hallazgos verificables."
developer_instructions = """
Revisá la spec y los archivos asignados.
No modifiques archivos. Devolvé evidencia y riesgos concretos.
"""
```

Se omite modelo para conservar la política vigente. El archivo no debe asumirse una sandbox de solo lectura: los agentes heredan permisos, proveedor y MCP del padre. Para imponer restricciones hay que usar controles efectivos del runtime. El padre coordina creación, seguimiento y resultados. [Subagentes oficiales](https://learn.chatgpt.com/docs/agent-configuration/subagents).

## 4. OpenCode

En la documentación principal, un especialista puede vivir en `.opencode/agents/contract-reviewer.md`:

```markdown
---
description: Revisa contratos sin editar ni ejecutar comandos.
mode: subagent
permission:
  edit: deny
  bash: deny
---
Revisá los archivos indicados y su spec. Devolvé hallazgos con evidencia.
```

OpenCode V2 documenta otros nombres (`permissions`, `shell`, `subagent`). Esta plantilla corresponde al esquema de la documentación principal consultada; el adapter debe detectar generación y validar schema. [Agentes](https://opencode.ai/docs/agents/), [Permisos V2](https://dev.opencode.ai/v2/docs/permissions/).

Los hooks se materializan con plugins JS/TS, no copiando `hooks.json` de otro producto. Plantilla estructural `.opencode/plugins/stack-policy.ts`:

```typescript
export const StackPolicy = async () => ({
  "tool.execute.before": async (input, output) => {
    // Pendiente: adaptar argumentos, evaluar política y manejar el error.
    // No autorizar automáticamente ni registrar secretos.
  }
})
```

Este bloque tiene deliberadamente un cuerpo vacío: demuestra dónde conectar el adapter y **no implementa protección**. Los eventos `tool.execute.after` y de sesión sirven para evidencia/compactación. [Plugins](https://opencode.ai/docs/plugins/).

## 5. Gemini CLI

Verificar primero si corresponde mantener Gemini CLI o migrar a Antigravity. Esta es una plantilla para el producto Gemini CLI documentado, no un archivo para Antigravity.

En `.gemini/settings.json`:

```json
{
  "hooks": {
    "BeforeTool": [{
      "matcher": "run_shell_command",
      "hooks": [{
        "type": "command",
        "command": "node scripts/agent-hooks/policy.cjs gemini",
        "timeout": 5000
      }]
    }]
  }
}
```

Aquí `timeout` se expresa en **milisegundos**. `BeforeTool` puede devolver `decision: "deny"` y `reason`; no es el JSON de Claude. El handler recibe JSON por stdin y debe reservar stdout para su respuesta JSON. [Referencia de hooks](https://geminicli.com/docs/hooks/reference/).

Especialista `.gemini/agents/contract-reviewer.md`:

```markdown
---
name: contract-reviewer
description: Revisa contratos y pruebas de los archivos asignados.
kind: local
tools:
  - read_file
  - grep_search
max_turns: 10
---
Analizá sin modificar. Reportá evidencia, incertidumbres y pruebas faltantes.
```

La delegación devuelve resultados al principal. No asumir mensajería peer-to-peer por tener subagentes. La documentación además describe A2A remoto, que requiere otra integración y autorización. [Subagentes Gemini](https://geminicli.com/docs/core/subagents/).

## 6. Cursor

Proyecto: `.cursor/hooks.json`. Plantilla con script todavía inexistente:

```json
{
  "version": 1,
  "hooks": {
    "beforeShellExecution": [{
      "command": "node scripts/agent-hooks/policy.cjs cursor"
    }]
  }
}
```

Este evento espera decisiones como `permission: "deny"`; no `decision: "deny"` de Gemini. Los hooks de proyecto usan el directorio raíz del proyecto; los globales tienen otro cwd. Los errores pueden fallar abiertos según el caso: comprobar `failClosed` y comportamiento de JSON inválido en la versión instalada. [Hooks](https://prod.cursor.com/docs/hooks).

Para especialistas, partir del schema de [Subagents](https://prod.cursor.com/docs/subagents) y del editor de personalizaciones de esa versión. `subagentStart` puede controlar la creación y `subagentStop` observar resultados o proponer follow-up acotado. No presentar ese mecanismo como bus general entre compañeros. La configuración de Agent no se extrapola a Tab.

## 7. VS Code con Copilot

La documentación actual distingue el harness seleccionado en VS Code. Para Copilot, `.github/hooks/audit.json` puede contener:

```json
{
  "version": 1,
  "hooks": {
    "preToolUse": [{
      "type": "command",
      "command": "node scripts/agent-hooks/audit.cjs copilot",
      "timeoutSec": 5
    }]
  }
}
```

Es observación propuesta, no autorización. `preToolUse` recibe `toolName/toolArgs`; el alias `PreToolUse` puede cambiar el formato del payload. Los hooks continúan en preview y existen diferencias con el antiguo Local harness. [Hooks VS Code](https://code.visualstudio.com/docs/agent-customization/hooks).

Especialista `.github/agents/contract-reviewer.agent.md`:

```markdown
---
name: Contract Reviewer
description: Revisa contratos con herramientas de lectura.
tools: ['read', 'search']
---
Compará implementación y spec. No edites. Reportá evidencia y riesgos.
```

Handoffs pueden cambiar el agente de la conversación; ejecutar un especialista como subagente es otra operación. Verificar tools y target en el harness seleccionado. [Custom agents](https://code.visualstudio.com/docs/agent-customization/custom-agents).

## 8. Antigravity: App, CLI e IDE

Las tres superficies documentan `.agents/hooks.json` a nivel proyecto y `~/.gemini/config/hooks.json` global. El mapa tiene **nombre del hook por encima del evento**, distinto de Claude:

```json
{
  "stack-policy": {
    "PreToolUse": [{
      "matcher": "run_command",
      "hooks": [{
        "type": "command",
        "command": "node scripts/agent-hooks/policy.cjs antigravity"
      }]
    }]
  }
}
```

La entrada usa `toolCall.name` y `toolCall.args`. No usar `tool_input` sin adaptación. [Hooks Antigravity](https://antigravity.google/docs/hooks).

Para App/CLI, especialista `.agents/agents/contract-reviewer.md`:

```markdown
---
name: contract-reviewer
description: Revisa contratos sin realizar cambios.
tools: [view_file, grep_search]
subagent: true
mainAgent: false
model: inherit
---
Devolvé evidencia y riesgos, sin modificar archivos.
```

Confirmar disponibilidad en cada superficie: no prometer custom subagents del IDE por analogía con App/CLI. La referencia alerta sobre nombres de herramientas inválidos; usar el inventario efectivo. [Custom subagents](https://antigravity.google/docs/subagents).

## 9. Devin y Windsurf

Distinguir Devin CLI, Desktop y cloud: las capacidades no son intercambiables. Devin CLI documenta `.devin/hooks.v1.json` con eventos directamente en la raíz:

```json
{
  "PreToolUse": [{
    "matcher": "exec",
    "hooks": [{
      "type": "command",
      "command": "node scripts/agent-hooks/policy.cjs devin",
      "timeout": 5
    }]
  }]
}
```

`timeout` está en segundos. Para bloquear, la referencia admite salida `decision: "block"` o exit 2; otros errores no bloquean automáticamente. También puede descubrir configuraciones Claude existentes: evitar doble registro. [Hooks Devin CLI](https://docs.devin.ai/cli/extensibility/hooks/overview).

Para especialistas y procedimientos consultar [Subagents CLI](https://docs.devin.ai/cli/subagents) y [Skills](https://docs.devin.ai/cli/extensibility/skills/overview). No inventar que el catálogo actual de Active Stack ya administra `.devin/` o el producto cloud. El identificador histórico `windsurf` del instalador requiere una decisión explícita de migración y compatibilidad.

## 10. Receta propia de especialistas comunicados

La siguiente es una **plantilla metodológica portable**, no sintaxis de configuración. El coordinador mantiene el contrato compartido. Explorar, implementar y revisar tienen brief y retorno diferenciados; el worker no modifica el scope por su cuenta.

```text
COORDINADOR → EXPLORADOR
Objetivo: localizar el contrato y las rutas realmente afectadas.
Entrada: issue/spec + estado CLI + decisiones vigentes.
Permisos: lectura. Salida: archivos/símbolos, evidencia y dudas.

COORDINADOR → IMPLEMENTADOR
Objetivo: comportamiento observable y escenarios de aceptación.
Ownership: rutas concretas; contratos compartidos acordados.
Validación: comandos autorizados, baseline y TDD aplicable.
Retorno: diff, pruebas ejecutadas y limitaciones.

COORDINADOR → REVISOR
Entrada: spec + diff final + evidencia del implementador.
Permisos: lectura. Buscar regresiones y contratos incumplidos.
Retorno: hallazgos con evidencia o resultado sin hallazgos.
```

Comunicación durante el trabajo:

```text
TIPO: bloqueo | cambio-de-contrato | resultado-parcial | final
TAREA: identificador estable
HECHO: evidencia breve, no una suposición
IMPACTO: qué otro trabajo depende de esto
PEDIDO: decisión o dato concreto necesario
```

Cuando el runtime solo soporta spawn/return, el coordinador retransmite mensajes. Donde hay mensajería nativa, usarla sin suponer que entrega instantánea equivale a respuesta o aprobación. Una devolución tardía debe indicar la revisión de archivos/commit sobre la que trabajó: evita aplicar hallazgos de un diff viejo a otro nuevo.

Paralelo solo con ownership disjunto o worktrees. El principal resuelve contratos comunes antes de lanzar modificaciones concurrentes. Un documento de handoff apunta a las fuentes; no copia toda la conversación ni convierte notas externas en instrucciones de mayor autoridad.

## 11. Cómo comprobar la configuración antes de distribuirla

Primero una tarea de lectura inofensiva y logs de nombres de eventos, sin argumentos sensibles. Comprobar carga única, cwd, tiempo de ejecución y que no cambien permisos. Después un fixture desechable para casos admitidos, bloqueados, JSON inválido, timeout y handler ausente. Finalmente probar el especialista con una tarea donde la escritura esté prohibida y verificar sus herramientas reales.

Los prehooks son buenos para reglas determinísticas; no colocar un clasificador caro en cada lectura. Los posthooks deben ejecutar comprobaciones acotadas y permitidas. Los eventos de fin pueden dispararse varias veces: usar idempotencia, límite de continuaciones y distinción entre turno, subagente y sesión.

Ninguna de las plantillas de este anexo sustituye los controles de filesystem, sandbox o autorización del host. La implementación futura requiere tests por adapter y plataforma. Este trabajo solo agregó documentación y ejemplos para revisión.
