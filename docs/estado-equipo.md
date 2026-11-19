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
| Emiliano | `(TBD)` | Roadmap, contrato, motor de memoria, documentación |
| Giuliano | `(TBD)` | Instrucciones y hooks de Engram, verify |
| Gianfranco | `(TBD)` | Fundación en modo equipo, permisos |
| Juan Cruz (dueño) | `(TBD)` | Revisión, integración, pilotos |

El usuario de GitHub sirve para confirmar la identidad de quien dice «soy X». Hasta completarlo, el agente pregunta y la persona confirma.

## Estado por ficha

Estados: `pendiente` · `en curso` · `PR abierto` · `mergeada` · `bloqueada`.
Una ficha puede empezar solo si **todas** sus dependencias figuran `mergeada` acá y en Git.

| Ficha | Responsable | Depende de | Estado | PR / repo | SHA de merge | Nota |
|---|---|---|---|---|---|---|
| R-01 | Matías | — | pendiente | | | |
| R-02 | Emiliano | — | pendiente | | | |
| R-03 | Giuliano | — | pendiente | | | |
| R-04 | Gianfranco | — | pendiente | | | |
| R-05 | Juan Cruz | R-01, R-02, R-03, R-04 | pendiente | | | Puerta G0 |
| R-06 | Emiliano | R-05 | pendiente | | | |
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
| R-18 | Gianfranco | R-11, R-04, R-01, R-16 | pendiente | | | ALTO |
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
| DA-10 | Canal de avisos: grupo de WhatsApp del equipo; aprueba cambios de diseño: Juan Cruz | 2026-11-19 |

## Punto de reanudación

**Siguiente acción:** el dueño abre la Ola 0 avisando al grupo para que Matías, Emiliano, Giuliano y Gianfranco tomen R-01 a R-04 en paralelo. Cuando las cuatro estén `mergeada`, el dueño resuelve R-05.
**Bloqueos vigentes:** ninguno.
**Autorizaciones vigentes:** ninguna de excepción.
