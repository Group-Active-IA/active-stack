# Estado del trabajo en equipo

> **Cada responsable edita solo su propia fila**, en su PR, y completa la fila de la dependencia ya mergeada (estado, PR, SHA) como primer paso de su ficha. El dueño no actualiza esta tabla; solo escribe las «Autorizaciones de excepción».
> Los agentes **leen** este archivo desde `origin/main` y lo contrastan con `git` y `gh`. **La autoridad de «mergeada» es Git/`gh`**; si la tabla va atrasada, se corrige la fila y se sigue.

**Última actualización:** 2026-11-19 · **Roadmap:** [roadmap-implementacion-equipo.md](roadmap-implementacion-equipo.md) · **Guía:** [guia-implementacion-equipo.md](guia-implementacion-equipo.md)

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
| R-08 | Emiliano | R-05 | pendiente | | | Habilitada. [Ficha lista](equipo/fichas/R-08.md); T-F4-4 con el caso de colisión `(TBD)` hasta que decida el dueño |
| R-09 | Emiliano | R-06 | pendiente | | | Repo externo |
| R-10 | Gianfranco | R-06 | mergeada | [orchestrator #7](https://github.com/Group-Active-IA/active-orchestrator/pull/7) · [agent-instruction #2](https://github.com/JuanCruzRobledo/agent-instruction/pull/2) · [stack #34](https://github.com/Group-Active-IA/active-stack/pull/34) | orchestrator `32142a9` · agent-instruction `c9f371c` · stack `17a276d` | Repos externos mergeados desde forks. El bump del `ref` en `harnesses.yaml` queda para R-20. Sin validador en OpenCode/Antigravity (la fase valida con reglas escritas) |
| R-11 | Matías | R-06 | pendiente | | | |
| R-12 | Matías | R-11 | pendiente | | | |
| R-13 | Giuliano | R-07 | PR abierto | [#45](https://github.com/Group-Active-IA/active-stack/pull/45) | | Sección `engram-cloud` + golden de `Compose` de las ocho variantes. Dos decisiones de producto con recomendación en el informe. [Informe](equipo/informes/R-13.md) |
| R-14 | Giuliano | R-13 | pendiente | | | |
| R-15 | Emiliano | R-08, R-01 | pendiente | | | |
| R-16 | Giuliano | R-15, R-14 | pendiente | | | ALTO |
| R-17 | Giuliano | R-08, R-16 | pendiente | | | ALTO |
| R-18 | Gianfranco | R-11, R-01, R-16 | pendiente | | | ALTO |
| R-19 | Gianfranco | R-10, R-17 | pendiente | | | Repos externos |
| R-20 | Juan Cruz | R-09, R-10, R-12, R-14, R-16, R-17, R-18, R-19 | pendiente | | | Integración |
| R-21 | Giuliano | R-20 | pendiente | | | |
| R-22 | Emiliano | R-20, R-21 | pendiente | | | |
| R-23 | Matías | R-22 | pendiente | | | Con autorización |
| R-24 | Juan Cruz | R-22 | pendiente | | | Con autorización |
| R-25 | Juan Cruz | R-23, R-24 | pendiente | | | Cierre |

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

## Punto de reanudación

**Siguiente acción:** R-05 quedó `mergeada`: **R-07 (Giuliano) y R-08 (Emiliano) están habilitadas** y tienen ficha. R-10 quedó `mergeada` (R-19 espera a R-17 de Giuliano). R-09 y R-11 siguen en curso.
**Decisiones pendientes del dueño (no bloquean R-07 ni R-08, sí a R-13, R-14 y R-16):**
1. **Forma de entrega del etiquetado en Claude.** El plan [EN] (Fases 5–6) da por hecho un hook; R-05 demostró que un hook no es fiable en Claude y se eligió el proxy B (DA-05). Hay que decidir si el producto entrega B (proxy dentro de `active-stack`, registrado como MCP) y reescribir las Fases 5–6, antes de que Giuliano empiece R-13. Codex sigue sin re-medirse (R-08 lo re-mide `(TBD)`).
2. **Colisión de slugs** (`feat/una` vs `feat-una`): ver [R-08](equipo/fichas/R-08.md).
**Riesgos detectados que ninguna ficha cubre todavía:** el stack no instala ni verifica el plugin de Engram para Claude (0.1.0 en vez de 0.1.22+); un `engram serve` huérfano o de otra base hace que el plugin bloquee los guardados con «identidad del servidor local sin confirmar»; falta `jq` en la máquina del dueño y el plugin lo usa. Candidatas: `verify` (Fase 10) y la instalación (Fase 9). Detalle en la [matriz de R-05](research/r05-windows-claude/matriz-claude-windows.md).
**Bloqueos vigentes:** ninguno.
**Autorizaciones vigentes:** ninguna de excepción.
