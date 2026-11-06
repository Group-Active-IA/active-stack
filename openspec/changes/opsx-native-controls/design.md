# Design

## Context

El stack hoy ofrece tres cosas distintas que conviene no mezclar:

1. **Texto** (orquestador, reglas, skills): orienta al agente, no impide nada.
2. **Permisos del cliente**: el harness `permissions` (tipo `config`, no se puede desactivar ni en Custom) escribe, por agente y nivel, un overlay en la configuración del cliente. En Claude se mezcla con `filemerge.MergeJSONObjects` en `~/.claude/settings.json`; en Codex se escriben `approval_policy` y `sandbox_mode` en `~/.codex/config.toml` con `UpsertTOMLTopLevel`. Hay tres niveles: Estricto, Balanceado y Bypass; un nivel vacío se normaliza a Balanceado, nunca a Bypass. Antes de escribir se crea un snapshot en `~/.active-stack/backups/permissions-<agente>`; si el backup falla, no se escribe.
3. **Hook de verificación del instalador** (`install.Options.VerifyHook`, `internal/verify/hook.go`): una función Go que corre al final de la instalación y hace rollback si un check duro falla. **No es un hook del cliente.**

Lo que hoy cubre `permissions`, según el README, es bloquear la lectura de `.env` y pedir confirmación en git destructivo. En Claude, la regla de evaluación es deny → ask → allow, y el piso de seguridad (`deny`) se aplica en todos los niveles, Bypass incluido.

`active-stack update` reinyecta configuraciones con el pipeline de instalación, registra la selección por agente y exige adopción explícita cuando encuentra material sin ownership registrado. Cualquier control nuevo tiene que entrar en ese circuito.

## Goals / Non-Goals

**Goals**

- Medir qué cubren de verdad los controles nativos (por cliente, versión, shell y superficie) y qué pasa cuando fallan.
- Dejar escrito qué queda garantizado por control nativo y qué solo por instrucción.
- Que ninguna escritura sobre la configuración del usuario ocurra sin diff, revisión, backup y reversión, y sin pisar entradas ajenas.

**Non-Goals**

- Prometer protección que no se probó.
- Resolver el tier estricto de Codex (01) o el contrato de evidencia (06).
- Subcomandos nuevos del binario, ledger firmado o runtime nuevo.

## Decisions

### D1 — Un vocabulario sin ambigüedad

En todos los artefactos de este change:

- **Hook del instalador** = `VerifyHook` de la instalación (Go, del binario). Fuera del piloto, salvo para decir qué *no* comprueba.
- **Hook del cliente** = comando que el cliente ejecuta alrededor de una acción (Claude Code en `settings.json`, o el equivalente en Codex si existe).

Cada regla del piloto se clasifica en **una** de tres clases:

| Clase | Qué hace | Ejemplo |
|---|---|---|
| **Orienta** | Le dice al agente qué hacer (texto). No impide nada. | Una instrucción del orquestador |
| **Restringe** | El cliente bloquea o pide confirmación antes de la acción, mediante permisos o rules. | `deny` de lectura de `.env` |
| **Verifica** | Un hook del cliente evalúa un evento y puede validar o frenar. | Un hook posterior contra el recibo de 06 |

### D2 — Controles previos para bloquear, posteriores para validar

- **Previos:** para lo que no debe ocurrir sin autorización (leer `.env`, git destructivo). Impiden el efecto. Si fallan, hay que decidir qué pasa (D5).
- **Posteriores:** para validar el resultado, como el cumplimiento del contrato de 06. No deshacen lo ya hecho, así que no son la única defensa para acciones destructivas.

### D3 — Cobertura declarada, no supuesta

Para cada cliente, versión, shell y superficie se registra qué acciones cubre el control nativo en una **matriz** (cliente × superficie × herramienta × shell). Una acción sin cobertura comprobada se documenta como *no controlada*. No se afirma protección que no se probó.

### D4 — Cómo se entregan los hooks del cliente

Los permisos ya siguen un camino probado: resolver la ruta por adapter, snapshot, merge, escritura atómica y omitir si no hay cambios. Los hooks de Claude viven en el mismo `settings.json`, así que la propuesta es reutilizar ese camino y agregar a los adapters una **capacidad opcional de hooks**, que cada cliente declare si soporta. Para Codex no se asume que exista un equivalente: el inventario lo confirma por versión, y si no existe, el piloto de Codex se limita a permisos.

El JSON no admite comentarios, así que hace falta una forma de reconocer qué entradas son del stack para actualizarlas y desinstalarlas sin tocar las del usuario. Opciones: un identificador distintivo dentro de cada entrada o un registro de ownership junto a los backups. **Decisión del grupo**: toca `filemerge`, nivel ALTO.

### D5 — Los handlers no pueden ser un subcomando del binario

Un hook ejecuta un comando. El roadmap excluye comandos nuevos de runtime, así que el handler no puede ser algo como `active-stack check`. Opciones, de menor a mayor alcance:

1. Solo reglas declarativas de permisos, sin hooks.
2. Hooks con comandos de shell en línea, simples y revisables en el diff.
3. Scripts instalados por el stack: exigen una variante por sistema (shell POSIX y PowerShell) y su propio ciclo de `update` y `uninstall`.
4. Subcomando del binario: requiere la decisión de arquitectura separada del roadmap.

Se propone pilotear las dos primeras y evaluar la tercera según los resultados. El stack se distribuye para Windows, macOS, Linux, WSL y Termux, así que cualquier comando se prueba en los shells reales de cada uno.

### D6 — Comportamiento ante fallo del control

Cuando el control no responde a tiempo, devuelve JSON inválido o su handler no existe, hay dos políticas: **fallar cerrado** (se bloquea la acción) o **fallar abierto** (se permite y se avisa). Propuesta de partida: cerrado para acciones de bloqueo y abierto con aviso para validaciones posteriores. El roadmap marca el punto como pendiente de acuerdo. Además, el piloto **registra lo que cada cliente hace realmente** en esos tres casos; no se asume una semántica que no se midió.

### D7 — Rutas alternativas y shells reales

Un control que mira un solo camino se esquiva con otro: un comando equivalente, otro intérprete, un script intermedio, una redirección. Las pruebas se hacen con shells reales (por ejemplo PowerShell y cmd en Windows, Git Bash, WSL, bash y zsh), no con simulaciones, y los caminos sin cobertura se registran como límites.

### D8 — Revisión antes de escribir configuración del usuario

El instalador ya ofrece `install --dry-run` y el planner genera un review payload. El piloto agrega a esa vista el **diff exacto** de permisos y hooks. Ninguna escritura ocurre sin backup previo; se preservan claves y perfiles ajenos y la reversión restaura el snapshot. Esta es la revisión humana que el roadmap exige a las propuestas de riesgo alto.

### D9 — Preservar las entradas del usuario

`MergeJSONObjects` reemplaza escalares y listas del usuario con los del overlay (solo mezcla objetos). Para `permissions` eso significa que `defaultMode`, `deny` y `allow` del usuario pueden perderse. Antes de sumar hooks al mismo archivo, el piloto tiene que resolver cómo se mezclan las **listas** (unión con ownership, o reemplazo solo de lo propio) y cómo se trata un valor escalar que el usuario fijó a propósito (por ejemplo `defaultMode`). Arreglarlo toca `filemerge` (ALTO) y se decide en el grupo. Se parte de reproducir el caso del 02/11 de forma aislada (tarea 1.4).

### D10 — Relación con los niveles de permisos

Hay que definir si los hooks dependen del nivel (Estricto, Balanceado, Bypass) o son independientes. En Claude, el piso `deny` se aplica en todos los niveles; en Bypass el resto se auto-aprueba. Si Bypass desactiva los controles de bloqueo, no puede usarse como base de la validación.

## Risks / Trade-offs

- **Falsa sensación de seguridad.** Un control parcial hace creer que todo está protegido. Mitigación: límites documentados y pruebas de rutas alternativas.
- **Configuración equivocada en la máquina del usuario** (ya observada el 02/11). Mitigación: diff previo, backup obligatorio, preservación de entradas ajenas, reversión.
- **Hooks propios mezclados con los del usuario.** Mitigación: ownership explícito y pruebas de preservación.
- **Handler roto que detiene el trabajo al fallar cerrado.** Mitigación: pruebas de handler ausente y timeout, y mensajes claros.
- **Diferencias entre versiones, shells y superficies.** Mitigación: matriz por combinación.
- **Superposición con 01.** Ambas verifican permisos efectivos de Codex. Mitigación: método de prueba compartido con Juan Cruz.
- **Deuda de desinstalación.** Quitar claves JSON es el punto débil conocido. Mitigación: tratarlo como requisito del diseño, no como mejora posterior.
- **Tocar `filemerge`** puede afectar a todos los harnesses. Mitigación: cambio mínimo, suite completa y revisión humana (ALTO).

## Migration Plan

No hay migración en esta ronda: no se escribe nada. Para la etapa de implementación, los controles entran con snapshot previo, son idempotentes al reinstalar, forman parte de `update` y se quitan por completo con `uninstall`. Una instalación existente sin controles sigue funcionando igual hasta que el usuario acepte el diff.

## Open Questions

Para acordar en el grupo antes de enviar a revisión:

1. Política ante fallo del control (punto pendiente del roadmap).
2. Alcance de handlers: hasta qué opción de D5 se pilotea.
3. Ownership de hooks: identificador dentro de la entrada o registro aparte (requiere review de quien mantiene `filemerge`).
4. Secuencia con 01: si el piloto de Codex espera la política compatible con el tier estricto.
5. Interacción con Bypass (D10).
6. Soporte de hooks en Codex: se confirma en el inventario y define si Codex entra al piloto solo con permisos.
7. Cómo se mezclan listas y escalares del usuario (D9) y quién arregla `filemerge`.
8. Agentes incluidos: confirmar que OpenCode y Antigravity quedan fuera.
9. Mínimo de cobertura para considerar adoptable el piloto.
