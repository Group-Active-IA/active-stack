# Proposal: opsx-verifiable-evidence

> Change `opsx-verifiable-evidence` · ficha 06 del roadmap · Responsable: Gianfranco · Agentes: Claude y Codex · Prioridad P1 · Gobernanza MEDIO
> Base revisada: `7a60089` (`origin/main`, con 02 y 03 mergeados; PR #8). Depende de 03 (`claude-native-opsx`, **mergeada en `main`**, citada contra `main`) y 04 (`codex-native-opsx`, **pendiente**, citada contra `origin/propose/codex-native-opsx` `9a6eb2a`). Habilita 07 y 08.
> Esta ronda es solo `propose`: todas las tareas de implementación quedan pendientes.

## Why

El orquestador delega a sub-agentes el trabajo que ensucia el contexto, incluidas las pruebas y los builds. Lo que el sub-agente devuelve llega como texto libre. En los assets revisados (`internal/harness/config/assets/claude|codex/sdd-orchestrator.md`, `origin/main` `7a60089`) no hay un contrato de retorno que permita comprobar que un "los tests pasan" corresponde a una ejecución real, sobre el estado actual del código y con un comando que cubra lo pedido. Un resultado bien formado puede venir de otro estado del árbol, de un comando insuficiente o de una ejecución inventada. Y un test verde tampoco prueba por sí solo que se cumpla un requisito.

Las propuestas 07 (delegar con contrato) y 08 (controles nativos) dependen de ese contrato. Sin él, delegar equivale a confiar en el texto del worker.

### Evidencia

- **El único "reporte" que existe hoy es el de salud de la instalación y no sirve como contrato.** En `internal/verify/report.go` (`origin/main`, `7a60089`), `BuildReport` marca `Ready = (Failed == 0)`. Los checks `skipped` y los `warning` no bloquean. Un reporte con cero checks queda `Ready` y emite "Installation verified. You're ready to start building.".
- **Caso observado el 02/11/2026** en una instalación de v0.4.0 (Windows, modo `lite`, agente `claude`): el paso `verify-hook` terminó con ✓ y el resumen dijo "0 passed, 0 failed". No se ejecutó ninguna comprobación y `~/.claude/settings.json` no tenía ningún hook. No se investigó por qué la lista de checks quedó vacía (queda como tarea en 1.5).
- **La ficha 06 del roadmap** fija los casos de rechazo: recibos inventados o viejos, comandos insuficientes, salidas truncadas sin original y requisitos incumplidos con tests verdes.
- **El asset de Codex** declara en su tercera línea "Do NOT apply it to executor agents" y su flujo lista `explore, propose, apply, archive` sin `verify`. El asset de Claude sí incluye `openspec-verify`. Un contrato escrito solo ahí no llega a los ejecutores de Codex.
- **`openspec/` ya se versiona en Git** (desde `1b4639b`, vigente en `7a60089`). Todo lo que se guarde dentro de un change viaja al repositorio del equipo, salidas originales incluidas.

## What Changes

- Se define el **recibo de ejecución** (siempre con este nombre, distinto del **registro de instalación** de 04): un registro, uno por ejecución de un comando, con identificador de corrida, árbol y estado exacto probado, comando, resultado, requisitos cubiertos y referencia a la copia redactada de la salida, con **una** huella y un registro de captura (design D2).
- Se separan **tres controles** con resultados independientes: **integridad** (el archivo conservado no cambió), **procedencia** (la ejecución es real y corresponde al estado del código) y **suficiencia** (lo ejecutado alcanza para el requisito). La integridad por sí sola nunca basta (design D3).
- **Acuerdo del 05/06:** un resultado declarado por el worker queda *sin verificar* hasta que una fuente independiente o una reejecución corrobora la ejecución y el estado del código. La verificación puramente mecánica (huella más relato) deja de ser criterio suficiente. El mecanismo de captura (arnés o fuente del cliente, antes del truncamiento) queda (TBD) hasta que 05 demuestre qué fuente cubre cada interfaz.
- Tres estados: **verificado**, **rechazado** y **sin verificar**. *Rechazado* es el texto de la ficha del roadmap (`CHANGES.md:192`): recibos inventados o viejos, comandos insuficientes, salidas truncadas sin original y requisitos incumplidos con tests verdes. Un recibo de ejecución incompleto o sin corroboración queda *sin verificar*; una ejecución sin recibos nunca queda *verificada*. Esto invierte a propósito la semántica de `skipped` de los health checks.
- El contrato se carga **cuando la acción requiere evidencia**, tanto en ejecución directa como delegada, con **el mismo paso de revisión de recibos** en ambas. En delegación, un extracto (`## Evidence Contract`) viaja además en el encargo. No es una regla universal de `CLAUDE.md` (todo lo que queda ahí llega también a los especialistas, 03 D0). **Propuesta (pendiente de acuerdo con 03 y 04):** un módulo propio `evidence`, cargado en `apply` y `verify` y no atado al toggle `delegation`, porque hoy la matriz de 03 (fila 28) lo ubica en `delegation`, que solo existe con su toggle. El design (D5) lista los cambios exactos que implicaría; esta ronda no edita archivos de 03 ni de 04. Claude y Codex reciben el mismo contenido y la misma aceptación.
- **Sobre de retorno:** 07 es dueña del sobre (`opsx-selective-delegation` D6). 06 define solo el campo `recibos` (una lista de recibos de ejecución, uno por comando).
- Se persiste solo la **copia redactada** de la salida. Su ubicación, acceso y vencimiento son (TBD) a acordar con 05; esta propuesta no fija una ruta. La política de redacción y retención es de 05 (design D6).
- **Huella:** sha256 de los bytes, con el algoritmo nombrado en el campo (como 02 y 05); no `git hash-object`. El estado probado con cambios sin confirmar se alinea con `TreeHash(dir)` de `eval/baseline/fixture/load.go`; el índice temporal de Git queda como alternativa a evaluar (design D2).
- Plataformas: Windows y Linux son el mínimo. La matriz distingue plataformas **distribuidas** (`release.yml:71-76`), **declaradas** (`CLAUDE.md:21`) y **probadas** (design D2).
- Los hooks complementan la evidencia pero no son un registro exhaustivo de ejecuciones (design D9).
- `verify` en Codex está **cubierto por 04** (spec `codex-opsx-delivery` y `native/SKILL.md` en `origin/propose/codex-native-opsx`; PR pendiente, design D8). Ya no es una pregunta abierta de 06.

## Capabilities

### New Capabilities

- `execution-receipts`: contrato de recibo de ejecución (esquema, estados, integridad, procedencia, suficiencia, copia redactada de la salida, carga del contrato en ejecución directa y delegada, y paridad entre Claude y Codex).

### Modified Capabilities

(ninguna: no existe un spec previo de delegación o de evidencia; el texto de los assets se modifica sin que haya un requisito archivado que lo describa)

## Alcance

- Esquema del recibo y reglas de validación.
- Reglas de integridad, procedencia y suficiencia, y los tres estados.
- Casos de cierre obligatorios: los cinco de la ficha, el reporte de cero recibos, el recibo falso con huella correcta, el mismo HEAD con dos diffs distintos, la salida redactada con referencia válida, el original vencido, un `apply` directo que produce recibos y pasa por la misma revisión, y que ninguna salida completa aparezca en `git status` tras una corrida.
- Qué se versiona (recibos de ejecución; ubicación (TBD)).
- Cambios previstos en los assets de Claude y de Codex y su prueba de paridad.
- Qué guarda el recibo de la salida: referencia a la copia redactada con una huella y registro de captura. La redacción, ubicación, acceso y vencimiento se acuerdan con 05.

## Exclusiones

- Ledger firmado, firma de recibos, subcomandos nuevos de runtime en el binario o un juez integrado. El roadmap los deja para una decisión de arquitectura separada.
- Reutilizar o modificar `internal/verify` (ver design D1).
- Lectura dirigida, acotado, redacción y retención de salidas (propuesta 05, design D9). Aquí solo se exige que el recibo de ejecución referencie la copia redactada con una huella y un registro de captura.
- Cuándo delegar y cuándo ejecutar directo (07), y el sobre de retorno (07; 06 solo define `recibos`). Controles previos y posteriores en el cliente (08).
- Agregar la acción `verify` al flujo de Codex: es de 04 (cubierto, PR pendiente).
- OpenCode y Antigravity: el alcance a Claude y Codex lo fija el roadmap (`CHANGES.md:3` y `:182`, `origin/main` `7a60089`), no esta propuesta.
- Implementación: `apply`, pruebas de implementación, `verify` y `archive` son de una etapa posterior.

## Impact

- **Código y assets previstos** (nada se toca en esta ronda):
  - El módulo `evidence` (propuesta, design D5) en los assets de Claude y de Codex, según el mecanismo de módulos de 03 y 04; el asset legado `internal/harness/config/assets/claude|codex/sdd-orchestrator.md` solo si el grupo no adopta la separación.
  - `internal/harness/config/compose_test.go` (golden de composición; hoy existen `TestCompose_AllToggles` para Claude y `_opencode`, `TestCompose_ClaudeDiagramIncludesVerify`, los golden `_codex` y `_antigravity`, y `TestCompose_VariantCodex`, que comprueba que aparezca la palabra "Codex").
  - Eventualmente `internal/catalog/harnesses.yaml` si el grupo elige un harness propio (BAJO).
- **No se tocan** `backup`, `filemerge` ni `pipeline` (ALTO). Si el diseño final obliga a tocarlos, la propuesta vuelve a revisión.
- **Gobernanza:** `harness/config` es MEDIO (implementar con checkpoints).
- **Otras propuestas:** modifica lo que reciben los roles de 03 y 04 y fija lo que 07 puede exigir y 08 puede validar.
- **Espejo obligatorio:** si se modifica la constitución del repo, `CLAUDE.md` y `AGENTS.md` se actualizan juntos.
