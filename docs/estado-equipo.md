# Estado del trabajo en equipo

> **Cada responsable edita solo su propia fila**, en su PR, y completa la fila de la dependencia ya mergeada (estado, PR, SHA) como primer paso de su ficha. El dueño no actualiza esta tabla; solo escribe las «Autorizaciones de excepción».
> Los agentes **leen** este archivo desde `origin/main` y lo contrastan con `git` y `gh`. **La autoridad de «mergeada» es Git/`gh`**; si la tabla va atrasada, se corrige la fila y se sigue.

**Última actualización:** 2026-11-21 · **Roadmap:** [roadmap-implementacion-equipo.md](roadmap-implementacion-equipo.md) · **Guía:** [guia-implementacion-equipo.md](guia-implementacion-equipo.md)

## Objetivo

Que Active Stack permita trabajar en equipo de forma automática en un proyecto fundado (modo equipo) y que la memoria compartida de Engram Cloud distinga el trabajo por rama y autor. Primer alcance: Claude y Codex. Detalle en [plan-trabajo-en-equipo.md](plan-trabajo-en-equipo.md) y [plan-engram-cloud-equipo.md](plan-engram-cloud-equipo.md).

## Equipo

| Persona | Usuario de GitHub | Carril |
|---|---|---|
| Matías | `(TBD)` | Verificación de Engram Cloud, entrega, ClickUp |
| Emiliano | `cachi001` | Roadmap, contrato, motor de memoria, documentación |
| Giuliano | `(TBD)` | Apoyo en R-01, instrucciones y hooks de Engram, verify |
| Gianfranco | `Gianfranco05` | Fundación en modo equipo, permisos |
| Juan Cruz (dueño) | `JuanCruzRobledo` | Revisión, integración, pilotos |

El usuario de GitHub sirve para confirmar la identidad de quien dice «soy X». Hasta completarlo, el agente pregunta y la persona confirma.

## Estado por ficha

Estados: `pendiente` · `en curso` · `PR abierto` · `mergeada` · `bloqueada` · `omitida`.
Una ficha puede empezar solo si **todas** sus dependencias figuran `mergeada` acá y en Git. Las `omitida` no cuentan.

| Ficha | Responsable | Depende de | Estado | PR / repo | SHA de merge | Nota |
|---|---|---|---|---|---|---|
| R-01 | Matías | — | mergeada | [#32](https://github.com/Group-Active-IA/active-stack/pull/32) | `88b05cb` | Aceptada con límites: solo Codex/macOS; Claude y Windows pasan a R-05 |
| R-02 | Emiliano | — | omitida | | | Absorbida: VE-1 pasa al primer paso de R-06 |
| R-03 | Giuliano | — | omitida | | | Absorbida en R-11, R-12 y el piloto R-24 |
| R-04 | Gianfranco | — | omitida | | | Absorbida en R-10 y R-18 |
| R-05 | Juan Cruz | R-01 | mergeada | [#35](https://github.com/Group-Active-IA/active-stack/pull/35) | `ebcbb17` | Puerta G0 cerrada para Claude + Windows. B (proxy) con vías narrativas; DA-02/05/09/12 registradas. Aviso: plugin de Claude desactualizado, ver [informe](equipo/informes/R-05.md) |
| R-06 | Emiliano | — | mergeada | [#30](https://github.com/Group-Active-IA/active-stack/pull/30) | `1cfe767` | Contrato y validador `team.py`; desbloquea R-09, R-10 y R-11 |
| R-07 | Giuliano | R-05 | mergeada | [#40](https://github.com/Group-Active-IA/active-stack/pull/40) | `1cf9e36` | Catálogo `engram-cloud` + guarda no-op en el instalador y en verify. [Informe](equipo/informes/R-07.md) |
| R-08 | Emiliano | R-05 | mergeada | [#44](https://github.com/Group-Active-IA/active-stack/pull/44) | `5b2f0b3` | `internal/memory` y `active-stack memory branch`, T-F4-1 a T-F4-12. Colisión de slugs `(TBD)` hasta que decida el dueño; desbloquea R-15 y, con R-16, R-17 |
| R-09 | Emiliano | R-06 | mergeada | [roadmap-generator #3](https://github.com/Group-Active-IA/roadmap-generator/pull/3) · [stack #48](https://github.com/Group-Active-IA/active-stack/pull/48) | rg `462de73` · stack `08e48cd` | Repo externo, desde el fork `cachi001`. El bump del `ref` queda para R-20. Sin validador de campos en OpenCode/Antigravity (avisa); decisión del dueño en el informe |
| R-10 | Gianfranco | R-06 | mergeada | [orchestrator #7](https://github.com/Group-Active-IA/active-orchestrator/pull/7) · [agent-instruction #2](https://github.com/JuanCruzRobledo/agent-instruction/pull/2) · [stack #34](https://github.com/Group-Active-IA/active-stack/pull/34) | orchestrator `32142a9` · agent-instruction `c9f371c` · stack `17a276d` | Repos externos mergeados desde forks. El bump del `ref` en `harnesses.yaml` queda para R-20. Sin validador en OpenCode/Antigravity (la fase valida con reglas escritas) |
| R-11 | Matías | R-06 | mergeada | [#46](https://github.com/Group-Active-IA/active-stack/pull/46) | `f4efb7f` | Entrega en modo equipo en `active-delivery`. [Informe](equipo/informes/R-11.md) |
| R-12 | Matías | R-11 | mergeada | [#49](https://github.com/Group-Active-IA/active-stack/pull/49) | `45243dd` | Seguimiento de fichas con ClickUp; T-F5-10 cerrada con lectura real en Codex y Claude. [Informe](equipo/informes/R-12.md) |
| R-13 | Giuliano | R-07 | mergeada | [#45](https://github.com/Group-Active-IA/active-stack/pull/45) | `f8d677e` | Sección `engram-cloud` + golden de `Compose` de las ocho variantes. [Informe](equipo/informes/R-13.md) |
| R-14 | Giuliano | R-13 | mergeada | [#47](https://github.com/Group-Active-IA/active-stack/pull/47) | `20c0c79` | Pruebas de coherencia T-F3-1 a T-F3-4; corrige la clave `opsx` de R-13; DA-11 verificado. [Informe](equipo/informes/R-14.md) |
| R-15 | Emiliano | R-08, R-01 | mergeada | [#50](https://github.com/Group-Active-IA/active-stack/pull/50) | `03e6e1d` | Núcleo de etiquetado (sirve a hook y proxy) y hooks `session-start`/`pre-read`. Sin `pre-save`: el hook volvió a fallar en Claude (5/12 contra 12/12 del proxy); espera la decisión pendiente 1 del dueño. [Informe](equipo/informes/R-15.md) |
| R-16 | Giuliano | R-15, R-14 | mergeada | [#54](https://github.com/Group-Active-IA/active-stack/pull/54) | `03fa64c` | ALTO. Instalador de los hooks de lectura (Claude y Codex), sin hook de guardado y sin conectar a `install` (R-20). Verificado con Claude real; **falta que alguien con Codex pruebe los pasos X0–X9 de la ficha**. [Informe](equipo/informes/R-16.md) **Pendiente: verificación con Codex real (X0–X9 de la ficha)**; ver «Decisiones pendientes del dueño». |
| R-17 | Giuliano | R-08, R-16 | mergeada | [#57](https://github.com/Group-Active-IA/active-stack/pull/57) | `c0b10ec` | ALTO. Propuesta aprobada (R17-A a R17-I) y implementada. R17-D queda como decisión abierta del dueño (Engram baja el nombre a minúsculas), no frena el merge. `active-stack memory init` con `internal/memory/projectinit`; [ficha](equipo/fichas/R-17.md) · [informe](equipo/informes/R-17.md) |
| R-18 | Gianfranco | R-11, R-01, R-16 | mergeada | [#56](https://github.com/Group-Active-IA/active-stack/pull/56) | `ffdc541` | ALTO. `deny` de push a `main` (Claude, con equivalentes PowerShell) y lecturas de `gh` en `balanceado`; une los arrays del usuario sin tocar `filemerge`. Codex `(TBD)`: se prueba con Codex real al final, junto con R-16. [Informe](equipo/informes/R-18.md) |
| R-19 | Gianfranco | R-10, R-17 | mergeada | [orchestrator #8](https://github.com/Group-Active-IA/active-orchestrator/pull/8) · [agent-instruction #3](https://github.com/JuanCruzRobledo/agent-instruction/pull/3) · [stack #61](https://github.com/Group-Active-IA/active-stack/pull/61) | orchestrator `5d758fc` · agent-instruction `1e8ca73` · stack `185d58d` | Fase `memory-policy` y sección que no duplica `engram-cloud`. DA-07: política solo en `.active-orchestrator-state.json`. El bump del `ref` en `harnesses.yaml` queda para R-20. [Informe](equipo/informes/R-19.md) |
| R-20 | Juan Cruz | R-09, R-10, R-12, R-14, R-16, R-17, R-18, R-19, R-26 | mergeada | [#65](https://github.com/Group-Active-IA/active-stack/pull/65) | `2a4b439` | Bump de los tres `ref` por SHA; hooks, AUTOSYNC y proxy de guardado conectados a install/uninstall (Claude y Codex). Codex real pendiente. [Informe](equipo/informes/R-20.md) |
| R-26 | Juan Cruz | R-15, R-08 | mergeada | [#64](https://github.com/Group-Active-IA/active-stack/pull/64) | `6d62242` | Proxy de guardado (DA-13): `active-stack memory proxy` y registro en Claude y Codex; reescribe las Fases 5 y 6 del plan. [Informe](equipo/informes/R-26-proxy-guardado.md) |
| R-21 | Giuliano | R-20 | mergeada | [#67](https://github.com/Group-Active-IA/active-stack/pull/67) | `e23919c` | Checks de `verify` de `engram-cloud` (marcador, hooks, `PATH` como aviso, proxy, AUTOSYNC, archivos del proyecto). Codex real pendiente. [Informe](equipo/informes/R-21.md) |
| R-22 | Emiliano | R-20, R-21 | mergeada | [#69](https://github.com/Group-Active-IA/active-stack/pull/69) | `4068f6d` | `ARCHITECTURE.md`, guías de equipo y de Engram Cloud, láminas de la guía externa (fuera del repo). Espejos `CLAUDE.md`/`AGENTS.md` sin tocar. [Informe](equipo/informes/R-22.md) |
| R-23 | Matías | R-22 | parcial | | [PR #74](https://github.com/Group-Active-IA/active-stack/pull/74), [PR #76](https://github.com/Group-Active-IA/active-stack/pull/76) | Claude/Windows aislado con agente real: V1-V7, V10, V11, V13 pasan; V8 con límite (`verify` no se ve por CLI); V9 corregido en hook (falta binario nuevo y el proxy no avisa); V12 solo Windows. Faltan Codex, dos personas reales y cloud real |
| R-24 | Juan Cruz | R-22 | parcial | | [PR #73](https://github.com/Group-Active-IA/active-stack/pull/73), [PR #77](https://github.com/Group-Active-IA/active-stack/pull/77) | Piloto local en Claude: VE-6 detecta 5/5, VE-8 +30 % de tokens en un saludo; DA-02/DA-03 sin umbral. Faltan Codex, dos personas reales, ClickUp (OK del dueño), OpenCode |
| R-25 | Juan Cruz | R-23, R-24 | cerrada con límites | 2026-11-23 | | Cierre documental con las pruebas que faltan registradas en [pruebas-pendientes.md](pruebas-pendientes.md). La versión sale con el release (no generado) |

## Quién sigue ahora (2026-11-21)

| Persona | Ficha | Qué hacer |
|---|---|---|
| Matías | R-23 (con autorización) | E2E de Engram y matriz V1–V13 en Claude y Codex |
| Giuliano | — | Sin ficha habilitada: ya terminó todas las suyas |
| Emiliano | — | Sin ficha habilitada: ya terminó todas las suyas |
| Gianfranco | — | Sin ficha habilitada: ya terminó R-10, R-18 y R-19 |

## Decisiones pendientes del dueño

- **Verificación de R-16 con Codex real (X0–X9, [ficha](equipo/fichas/R-16.md)).** R-16 se mergeó sin ella (el instalador no está conectado a `install`, así que no cambia la configuración de nadie). Claude real sí se verificó (Claude Code 2.1.330). **Se deja para el final (decisión del dueño, 2026-11-21), junto con el `deny` de Codex de R-18; debe estar hecha antes de que R-20 conecte el instalador a Codex y, a más tardar, antes del cierre R-25.** Quien la haga pega la tabla completa en la ficha y cierra T-F6-10; si algo falla, se avisa al grupo (guía, sección 8) y Giuliano lo corrige.
- **Cómo se etiqueta el guardado de memoria (hook o proxy):** decidido el 2026-11-21 (proxy). Ver «Decisiones tomadas» y «Trabajo derivado de las decisiones».
- **Colisión de slugs:** decidido el 2026-11-21. Ver «Decisiones tomadas».

## Autorizaciones de excepción

Aquí el dueño anota, si hace falta, una autorización explícita para adelantar una ficha cuya dependencia no está mergeada (por ejemplo, trabajar sobre una rama sin mergear). **Sin una línea acá, el agente no adelanta nada.**

| Ficha | Autorizada por | Fecha | Condición |
|---|---|---|---|
| — | — | — | — |

## Decisiones tomadas

Se registran acá **solo** las decisiones del dueño de la sección 3 del roadmap (DA-xx). Las que toma cada implementador quedan en el traspaso de su ficha (`docs/equipo/fichas/R-xx.md`).

| ID | Decisión | Fecha |
|---|---|---|
| DA-01 | Opción (a): se versionan `team.json` y los vínculos de fichas con ClickUp; observaciones de cada sincronización, `team.local.json` y `me.local.json` quedan locales | 2026-11-19 |
| DA-02 (Engram, Claude) | Tope de sobrecosto de B: entrada sin caché ≤ +15 % de la línea base y ≤ 3.000 tokens por sesión de 3 guardados (medido: mediana +4,3 %, máx +15,0 % / +2.411). Codex: se re-mide en R-08 `(TBD)`. Decidida por delegación del dueño | 2026-11-19 |
| DA-05 | Vía **B (proxy)** para Claude (R-05, Windows + Claude Code 2.1.327, Engram 3.7.2, plugin 0.1.22). Con el plugin real, hook y proxy pasan en secuencia (6/6), pero bajo carga el **hook falla 3 de 12 (9/12) y B 0 de 12 (12/12)**: el hook se ejecuta bien y su reescritura se descarta porque corre en paralelo con el `PreToolUse` del plugin (orden no especificado). Un hook de Claude **no puede ser el mecanismo de entrega**; B sí. B cubre también resumen, finalize, prompt y captura pasiva. Límite: si el proxy termina, el siguiente guardado falla (visto en Codex; en Claude no ejecutado). Evidencia: [matriz](research/r05-windows-claude/matriz-claude-windows.md) | 2026-11-19 |
| DA-09 | No se hace la Fase 8: la lectura de rama no pide confirmación (Claude en 3 modos; Codex en 2). Se reabre si R-14/R-16 mide un pedido en el cierre completo | 2026-11-19 |
| DA-12 | Garantía de etiquetado por agente: Claude + Windows pasa V1–V5, V7, V9, V12, V13 y vías narrativas; V8 interactivo y V9 con proxy terminado no ejecutados; Codex parcial. Nada se activa solo | 2026-11-19 |
| DA-04 | La fundación ofrece proteger `main` solo con confirmación; si el plan o los permisos no lo permiten, lo dice y deja la convención escrita (lo que R-10 ya implementa). La cuenta del dueño es GitHub Free | 2026-11-19 |
| DA-10 | Canal de avisos: grupo de WhatsApp del equipo; aprueba cambios de diseño: Juan Cruz | 2026-11-19 |
| R-16 D-A | El código de los hooks va en un paquete nuevo `internal/harness/hooks`, sin tocar `filemerge` ni `backup` (se reutiliza `backup` sin modificarlo) | 2026-11-21 |
| R-16 D-B | R-16 entrega solo el instalador (Install/Remove/Status con pruebas); R-20 lo conecta a install/uninstall/verify | 2026-11-21 |
| R-16 D-C | Si `active-stack` no está en el PATH, los hooks se instalan igual con aviso, y `verify` lo marca | 2026-11-21 |
| R-16 verificación | Giuliano implementa Claude y Codex con archivos y documentación y deja la rama lista; la prueba con Codex real y Claude real la hace el dueño antes de mergear | 2026-11-21 |
| DA-13 | Guardado de memoria etiquetado por **proxy** en Claude y Codex, sin hook de guardado en ninguno (los hooks de lectura `session-start`/`pre-read` se mantienen). Se reescriben las Fases 5 y 6 del plan [EN] | 2026-11-21 |
| R17-D | `memory init` escribe `project_name` en minúsculas, igual que Engram | 2026-11-21 |
| Slugs | Se agrega el sufijo de 6 hex también cuando el slug pierde información respecto de la rama (`/`, `_`, mayúsculas), no solo por largo (recomendación de R-08) | 2026-11-21 |
| Hecho el desarrollo pendiente | El dueño pidió terminar el desarrollo pendiente con el agente; lo que requiere Codex real queda para el final y el piloto R-24 requiere su OK explícito | 2026-11-21 |
| DA-02 | Tope de tokens del modo equipo: hasta +35 % medido en un pedido real, no en un saludo (hoy +30 % en un saludo) | 2026-11-23 |
| DA-03 | Detección del modo equipo: al menos 4 de 5 corridas (hoy 5/5) | 2026-11-23 |
| Escenario 8 | ClickUp probado con una lista descartable: asignado, dependencia y estado verificados; tareas borradas. La lista vacía `ZZ-PILOTO-active-stack (borrar)` queda para borrar a mano (no hay herramienta para borrar listas) | 2026-11-23 |
| Release v0.9.0 | Publicada el 2026-11-23 con modo equipo, memoria de equipo, proxy de guardado, aviso y `self-update` (R-27) y los arreglos de `hooks` y `filemerge`. Instalada en la PC del dueño. Guía y landing republicadas | 2026-11-23 |
| Release v0.9.1 | Publicada el 2026-11-24: las reglas de entrega ya no piden ODD o SDD en las fases de la fundación ([PR #88](https://github.com/Group-Active-IA/active-stack/pull/88)); binarios de macOS subidos a mano. `self-update` probado contra la release real (v0.9.0 a v0.9.1, con backup) | 2026-11-24 |
| macOS en release y self-update | El workflow de release compila `darwin_amd64` y `darwin_arm64` y `self-update` los reconoce ([PR #90](https://github.com/Group-Active-IA/active-stack/pull/90)); llega con la próxima release. Falta probarlo en una Mac real (PP-20) | 2026-11-24 |
| Release v0.9.2 | Publicada el 2026-11-26 desde `c41c387` por el workflow, con los binarios de macOS incluidos. Antes se arregló `update` en macOS ([PR #93](https://github.com/Group-Active-IA/active-stack/pull/93)). Falta probar en una Mac real (PP-20) y archivar el cambio `v0-9-2-delivery` | 2026-11-26 |
| PR #16 cerrada | `propose/opsx-shadow-review` se cerró sin mergear al cerrar la sesión del 2026-11-24; la rama queda intacta en el remoto y la PR se puede reabrir | 2026-11-24 |
| Ramas de rescate | `wip/clickup-actions-rescate` y `wip/f5-pilots-rescate` guardan cambios locales de worktrees viejos; sin mergear, a revisar por el dueño | 2026-11-24 |

## Punto de reanudación

**Siguiente acción:** R-05 quedó `mergeada`: **R-07 (Giuliano) y R-08 (Emiliano) están habilitadas** y tienen ficha. R-10 quedó `mergeada` (R-19 espera a R-17 de Giuliano). R-09 y R-11 siguen en curso.
**Decisiones pendientes del dueño (no bloquean R-07 ni R-08, sí a R-13, R-14 y R-16):**
1. **Forma de entrega del etiquetado en Claude.** El plan [EN] (Fases 5–6) da por hecho un hook; R-05 demostró que un hook no es fiable en Claude y se eligió el proxy B (DA-05). Hay que decidir si el producto entrega B (proxy dentro de `active-stack`, registrado como MCP) y reescribir las Fases 5–6, antes de que Giuliano empiece R-13. Codex sigue sin re-medirse (R-08 lo re-mide `(TBD)`).
2. **Colisión de slugs** (`feat/una` vs `feat-una`): ver [R-08](equipo/fichas/R-08.md).
**Riesgos detectados que ninguna ficha cubre todavía:** el stack no instala ni verifica el plugin de Engram para Claude (0.1.0 en vez de 0.1.22+); un `engram serve` huérfano o de otra base hace que el plugin bloquee los guardados con «identidad del servidor local sin confirmar»; falta `jq` en la máquina del dueño y el plugin lo usa. Candidatas: `verify` (Fase 10) y la instalación (Fase 9). Detalle en la [matriz de R-05](research/r05-windows-claude/matriz-claude-windows.md).
**Bloqueos vigentes:** ninguno.
**Autorizaciones vigentes:** ninguna de excepción.

## Trabajo derivado de las decisiones

- ~~Proxy de guardado~~: hecho (R-26, #64) y conectado (R-20, #65).
- ~~R-08: sufijo hex en slugs~~: hecho (#62).
- ~~R-17: nombre en minúsculas~~: hecho (#62).
- Pruebas que faltan (Codex real, dos personas, cloud real, otros sistemas): [pruebas-pendientes.md](pruebas-pendientes.md).

### Hallazgos de R-23 (sin arreglar)

- `install` headless Machine imprime «0 checks»: los `cloud:*` de `verify` no corren y no hay comando `verify` suelto.
- V9: con marcador inválido el guardado no se bloquea, pero no sale aviso en `stderr` (el plan lo pide).
- Login de `claude` caducado en la máquina del dueño (2026-11-22): hay que hacer `/login` antes de repetir R-23 y R-24.
- Hallazgos de R-24: los agentes de ficha dicen «no existe `CHANGES.md`» cuando sí existe; un agente pisó la sección `team-mode` del `.gitignore`; los commits de ficha arrastran `__pycache__` si no se ignora.
- Pendiente del dueño: umbrales DA-02 (tope de tokens) y DA-03; borrar `E:23c-tmp` y `C:24b`.
