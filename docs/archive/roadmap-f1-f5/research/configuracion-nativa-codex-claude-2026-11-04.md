# Configuración nativa de Codex y Claude para Active Stack

Fecha: 4 de noviembre de 2026. Investigación oficial y lectura del instalador.
Estado: propuesta técnica; sin cambios en adapters, catálogo ni configuración personal.
Versiones locales consultadas con `--version`: Codex CLI 0.166.0 y Claude Code 2.1.309.
Se consultó también `--help`; no se iniciaron sesiones de modelo ni pruebas de hooks.

## 1. Conclusión y alcance

Mantener SDD instalado y como flujo predeterminado del stack es compatible con mejorar su ubicación. Instalar el harness, activarlo en la sesión principal y registrar especialistas son operaciones diferentes.

El problema actual es que el instalador entrega el rol de coordinador mediante instrucciones compartidas. La frase «aplicar solo al agente dedicado» no registra un agente ni crea aislamiento de roles.

No debe copiarse un mismo archivo de configuración entre productos. Se necesita un contrato metodológico compartido y renderizadores nativos por cliente y versión. Ninguna propuesta de este documento implica desactivar SDD ni convertirlo en una nueva elección obligatoria del usuario.

## 2. Hallazgos comprobados en el repositorio

| Hallazgo | Evidencia local | Consecuencia |
|---|---|---|
| Claude entrega el orquestador como instrucciones | `internal/agents/claude/adapter.go`, `ConfigDelivery()` | El cuerpo llega al CLAUDE.md general; no registra un agente Claude |
| Codex usa la misma estrategia | `internal/agents/codex/adapter.go`, `ConfigDelivery()` | El cuerpo llega al AGENTS.md general |
| Ambas plantillas dicen que son exclusivas del agente dedicado | `internal/harness/config/assets/{claude,codex}/sdd-orchestrator.md` | Hay una contradicción entre intención y entrega |
| OpenCode ya tiene otra estrategia | `internal/harness/config/install.go` y `primary_agent.go` | Hay un precedente de separación, pero su JSON no sirve para los otros clientes |
| La plantilla Codex anuncia `/opsx:...` | `internal/harness/config/assets/codex/sdd-orchestrator.md` | Debe comprobarse cómo se entregan realmente esas acciones; el adapter deja `CommandsDir()` vacío |
| El modelo explica Claude como archivo plano | `internal/model/harness.go`, comentario de `ConfigDelivery` | La abstracción necesita actualizarse para registrar agentes y activación por separado |
| Los starters MCP modelan comando, argumentos y entorno | `internal/model/starter.go` | No representan un endpoint remoto y su transporte de manera completa |
| La ruta remota Codex concatena `/mcp` | `internal/harness/external/mcp.go`, `writeCodexHarnessMCP()` | Un endpoint completo puede quedar alterado; falta una reproducción con fixture |

Hay integración nativa existente de MCP y rutas de skills. No sería preciso afirmar que Active Stack no usa ninguna configuración propia de estos productos. El déficit concreto es el registro de roles, su activación, hooks y verificación efectiva.

## 3. Claude: separar el agente del contexto compartido

Claude admite agentes Markdown en `.claude/agents/` y `~/.claude/agents/`. El frontmatter configura capacidades y el cuerpo define el prompt del agente. Campos relevantes: `tools`, `disallowedTools`, `model`, `permissionMode`, `skills`, `hooks`, `maxTurns` e `isolation`. El coordinador principal puede limitar los tipos delegables con `Agent(nombre,...)`. Los campos desconocidos pueden ignorarse: escribir un YAML válido no demuestra que se aplique. [Agentes oficiales](https://code.claude.com/docs/en/sub-agents).

Ejemplo de estructura propuesta, no desplegada:

```text
.claude/
  agents/
    sdd-orchestrator.md
    sdd-explorer.md
    sdd-implementer.md
    sdd-reviewer.md
  skills/
    <acciones OPSX y procedimientos>/SKILL.md
  settings.json
```

Esqueleto del coordinador:

```yaml
---
name: sdd-orchestrator
description: Coordina el trabajo OPSX y delega tareas con aceptación y evidencia.
tools: Agent(sdd-explorer,sdd-implementer,sdd-reviewer), Read, Glob, Grep, Bash
---
```

Después del frontmatter iría el cuerpo compuesto del orquestador, revisado para las herramientas reales. No se propone fijar modelos ni copiar indiscriminadamente todos los toggles a los workers.

Para ejecutar ese agente como sesión principal existe `claude --agent sdd-orchestrator`. Para conservarlo como predeterminado se puede registrar `"agent": "sdd-orchestrator"` en el settings del alcance elegido. `--agent` tiene precedencia por sesión. Esto permite mantener el default del stack sin pegar el rol completo en CLAUDE.md. [CLI](https://code.claude.com/docs/en/cli-reference), [setting agent](https://code.claude.com/docs/en/settings-reference#agent).

Propuesta: CLAUDE.md conserva reglas del proyecto; el archivo del agente contiene coordinación; cada especialista recibe sus responsabilidades. No usar `omitClaudeMd` como solución general: puede omitir también reglas de proyecto necesarias. Probar que el worker implementa y que el coordinador coordina.

Los settings tienen alcances de usuario, proyecto, local y administrado. El instalador debe respetar el alcance elegido y la precedencia; una política administrada no se reemplaza con preferencias de usuario. [Settings y precedencia](https://code.claude.com/docs/en/settings).

## 4. Codex: distinguir sesión principal de subagentes

Los especialistas se definen en `.codex/agents/*.toml` o `~/.codex/agents/*.toml`. Requieren `name`, `description` y `developer_instructions`; pueden declarar modelo y esfuerzo. Heredan permisos, proveedor y MCP del padre. No se puede crear un sandbox read-only poniendo `sandbox_mode` en ese archivo. Hay reducciones de capacidades y desactivación de skills admitidas. [Subagentes oficiales](https://learn.chatgpt.com/docs/agent-configuration/subagents).

Esqueleto ilustrativo de un especialista:

```toml
name = "sdd_explorer"
description = "Localiza contratos, símbolos y evidencias antes de implementar."
developer_instructions = """
Explorá el alcance delegado y devolvé rutas, símbolos y dudas con evidencia.
No modifiques archivos ni ejecutes operaciones de escritura.
"""
```

Esa instrucción de lectura expresa una responsabilidad; no constituye enforcement técnico.

Registrar `sdd-orchestrator.toml` no demuestra que el chat principal adopte ese rol. No se encontró un equivalente documentado de `claude --agent` en las páginas revisadas ni en el `--help` local de Codex. No inventar una clave `agent = "sdd-orchestrator"` para conseguirlo.

Codex ofrece perfiles CLI como capas de configuración. Desde 0.134.0 se seleccionan con `codex --profile nombre` y se almacenan en `$CODEX_HOME/nombre.config.toml`; el formato anterior `[profiles.nombre]` y el selector superior `profile` dejaron de funcionar. La configuración de proyecto está sujeta a confianza y no admite seleccionar perfiles. [Configuración avanzada](https://learn.chatgpt.com/docs/config-file/config-advanced).

La referencia distingue `developer_instructions`, que agrega instrucciones, de `model_instructions_file`, que reemplaza las instrucciones base. La clave `instructions` está reservada. No usar el reemplazo del prompt base para instalar SDD. [Referencia de configuración](https://learn.chatgpt.com/docs/config-file/config-reference).

Propuesta para el default actual de Codex:

1. Mantener en AGENTS.md un bootstrap breve: la sesión principal coordina OPSX; los ejecutores siguen el contrato delegado.
2. Mover el procedimiento largo a skills y registrar especialistas TOML.
3. Para CLI, evaluar un perfil SDD explícito con instrucciones adicionales; no asumir que cambia automáticamente el comportamiento de la app o del IDE.
4. Probar herencia con un worker real: un perfil es una capa de sesión, no una garantía de que sus instrucciones no lleguen a los hijos.

AGENTS.md es guía descubierta por directorios; los overrides y el alcance global/proyecto importan. Escribir un encabezado de rol no crea un selector de agentes. [Descubrimiento de instrucciones](https://learn.chatgpt.com/docs/agent-configuration/agents-md).

## 5. Skills y comandos: entregar la acción que el cliente reconoce

Claude unificó comandos personalizados y skills. `.claude/commands/` sigue funcionando, pero para trabajo nuevo recomienda skills con archivos de apoyo. Permite `disable-model-invocation: true`, `allowed-tools` y `context: fork` con selección de agente. Un fork de skill no recibe la conversación completa. [Skills Claude](https://code.claude.com/docs/en/skills).

Codex usa `.agents/skills/<nombre>/SKILL.md`, tanto en repositorio como en el directorio personal. La metadata opcional `agents/openai.yaml` admite `policy.allow_implicit_invocation: false` para exigir invocación explícita. [Skills Codex](https://learn.chatgpt.com/docs/build-skills).

Los custom prompts de Codex están deprecados a favor de skills. No proponerlos como integración nueva de comandos. [Aviso oficial](https://learn.chatgpt.com/docs/custom-prompts).

Propuesta: mantener identificadores funcionales comunes de acciones OPSX y documentar la invocación real por cliente. No anunciar `/opsx:apply` en Codex por analogía con Claude. Primero comprobar lo que generó la versión instalada de OpenSpec y lo que descubre el host.

## 6. Prehooks y posthooks: contratos propios

| Momento | Uso propuesto en Active Stack | Restricción de diseño |
|---|---|---|
| `SessionStart` | Cargar un contexto breve y referencias | Evitar volcar toda la memoria |
| `PreToolUse` | Comprobación determinista antes de una operación | Probar denegación y fallo del handler |
| `PostToolUse` | Formato o chequeo limitado a archivos propios | No revierte lo que ya ocurrió |
| `SubagentStart` | Añadir contrato de rol si falta | No duplicar todo el prompt |
| `SubagentStop` | Recoger estado y evidencia | Distinguir retorno de subagente y fin de sesión |
| `Stop` | Señalar verificaciones pendientes | Límite de continuaciones y salida ante bloqueo externo |

Claude registra hooks en settings y admite también hooks asociados a agentes/skills. Sus command hooks pueden usar `command` + `args` sin shell; en Windows los shims `.cmd`/`.bat` requieren otro tratamiento. `onFailure: "block"`, desde 2.1.295, permite bloquear ante fallos en eventos compatibles. No copiar ese campo a Codex sin soporte documentado. [Referencia de hooks Claude](https://code.claude.com/docs/en/hooks).

Codex admite `hooks.json` junto a las capas activas o tablas `[hooks]`; conviene una representación por capa. Los hooks no administrados requieren confianza sobre su definición. Soporta handlers `command` y `mcp_tool`; `prompt` y `agent` se parsean pero se omiten. Hay `commandWindows`. Un error de MCP hook no bloquea por sí mismo; un posthook no deshace efectos. Hay diferencias entre ejecución local y orquestación cloud. [Hooks Codex](https://learn.chatgpt.com/docs/hooks).

Propuesta: scripts pequeños compartidos solo si la entrada/salida se adapta por host. Evitar un clasificador LLM en cada lectura y una suite completa tras cada edición. Las comprobaciones de calidad deben poder ejecutarse fuera del hook. No introduce GGA ni automatización de commits.

## 7. Configuración propia que conviene aprovechar

| Necesidad | Claude | Codex | Cambio propuesto |
|---|---|---|---|
| Rol principal SDD | Archivo de agente + setting `agent` | Bootstrap principal; perfil CLI como alternativa a validar | Separar registro de activación |
| Especialistas | Frontmatter y cuerpo Markdown | TOML nativo | Generar perfiles distintos desde contratos comunes |
| Modelo/esfuerzo | Campos del agente/settings | Campos del agente/config | Configurar donde el runtime lo interpreta |
| Herramientas del rol | `tools` / `disallowedTools` | Capacidades soportadas; límites heredados | No traducir permisos literalmente |
| Procedimientos | Skills | Skills | Carga bajo demanda |
| Normas por dominio | `.claude/rules/` con `paths` | AGENTS.md por directorio | Mantener reglas cercanas a su alcance |
| Eventos | Settings de hooks | Hooks JSON o TOML | Implementaciones por host |
| MCP | Alcance usuario/proyecto ya modelado por adapter | `mcp_servers` ya usado | Completar transporte/endpoint y comprobar conexión |

Las reglas Claude con `paths` permiten cargar normas según los archivos afectados. Son una mejor ubicación para convenciones de dominio que un coordinador global cada vez más largo. [Reglas por rutas](https://code.claude.com/docs/en/memory#path-specific-rules).

No instalar todos los mecanismos por el solo hecho de existir. Memoria, permisos, skills y hooks de terceros deben tener un responsable; evitar duplicar captura y registros con Engram.

## 8. Cambios y aceptación recomendados

Orden propuesto, sin changes creados:

1. **Registro y activación nativos:** ampliar adapter/modelo para resolver agentes, hooks, bootstrap y activación; conservar SDD predeterminado. Reemplazar el comentario obsoleto sobre Claude.
2. **Migración segura:** backup; generar destino nativo; verificar descubrimiento; retirar únicamente secciones propias antiguas. Nunca purgar instrucciones del usuario ni migrar a medias silenciosamente.
3. **Especialistas y acciones OPSX:** registrar contratos; comprobar invocación real, roles y herramientas. Evitar instrucciones de coordinación en implementadores.
4. **Hooks:** agregar pocos eventos con probes de carga, confianza, denegación, timeout, JSON inválido y handler ausente.
5. **MCP:** modelar endpoints completos, transportes y autenticación gestionada por el cliente; no añadir `/mcp` a ciegas.

La aceptación debe comprobar comportamiento: descubrimiento del agente, activación del principal, ausencia de rol coordinador en el worker, preservación de configuración ajena, reinstalación idempotente, rollback y retorno de evidencia. Registrar estado `instalado`, `pendiente de confianza`, `cargado` y `verificado` por separado.

Primero validar Windows nativo, relevante para el entorno observado; después la matriz de plataformas soportadas. No afirmar paridad app/CLI/IDE/cloud solo porque comparten parte de la configuración. Las versiones mínimas de cada integración nueva quedan (TBD) hasta probarlas.

Esta investigación no ejecutó builds, tests de integración, hooks ni cambios en dotfiles. Los esqueletos son propuestas para revisión; necesitan fixtures y pruebas de comportamiento antes de incorporarse al instalador.
