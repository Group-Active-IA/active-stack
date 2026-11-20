# Estado del trabajo en equipo

> **Único escritor: Juan Cruz (dueño).** Nadie más edita este archivo ni en su rama ni en su PR: así no hay conflictos de merge. Las personas informan su avance con el informe de cierre (`docs/equipo/informes/`) y el dueño actualiza esta tabla **al aceptar cada PR**.
> Los agentes **leen** este archivo desde `origin/main` (no desde una copia local desactualizada) y lo contrastan con `git` y `gh`. Si no coinciden, avisan.

**Última actualización:** 2026-11-19 · **Roadmap:** [roadmap-implementacion-equipo.md](roadmap-implementacion-equipo.md) · **Guía:** [guia-implementacion-equipo.md](guia-implementacion-equipo.md)

## Objetivo

Que Active Stack permita trabajar en equipo de forma automática en un proyecto fundado (modo equipo) y que la memoria compartida de Engram Cloud distinga el trabajo por rama y autor. Primer alcance: Claude y Codex. Detalle en [plan-trabajo-en-equipo.md](plan-trabajo-en-equipo.md) y [plan-engram-cloud-equipo.md](plan-engram-cloud-equipo.md).

## Equipo

| Persona | Usuario de GitHub | Carril |
|---|---|---|
| Matías | `(TBD)` | Verificación de Engram Cloud, entrega, ClickUp |
| Emiliano | `cachi001` | Roadmap, contrato, motor de memoria, documentación |
| Giuliano | `(TBD)` | Apoyo en R-01, instrucciones y hooks de Engram, verify |
| Gianfranco | `(TBD)` | Fundación en modo equipo, permisos |
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
| R-05 | Juan Cruz | R-01 | pendiente | | | Puerta G0 (solo Engram). En Windows + Claude Code. [Ficha lista](equipo/fichas/R-05.md) |
| R-06 | Emiliano | — | mergeada | [#30](https://github.com/Group-Active-IA/active-stack/pull/30) | `1cfe767` | Contrato y validador `team.py`; desbloquea R-09, R-10 y R-11 |
| R-07 | Giuliano | R-05 | pendiente | | | |
| R-08 | Emiliano | R-05 | pendiente | | | |
| R-09 | Emiliano | R-06 | pendiente | | | Repo externo |
| R-10 | Gianfranco | R-06 | pendiente | | | Repos externos |
| R-11 | Matías | R-06 | pendiente | | | |
| R-12 | Matías | R-11 | pendiente | | | |
| R-13 | Giuliano | R-07 | pendiente | | | |
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
| DA-10 | Canal de avisos: grupo de WhatsApp del equipo; aprueba cambios de diseño: Juan Cruz | 2026-11-19 |

## Punto de reanudación

**Siguiente acción:** R-06 y R-01 quedaron `mergeada`. Arrancan R-09 (Emiliano), R-10 (Gianfranco) y R-11 (Matías). R-05 (Juan Cruz, en Windows + Claude Code) tiene la ficha lista: decir «Soy Juan Cruz, continuá». Al cerrar R-05 se habilitan R-07 y R-08.
**Bloqueos vigentes:** ninguno.
**Autorizaciones vigentes:** ninguna de excepción.
