# opsx-baseline — estado de orquestación e índice de informes

Change 02 · Responsable: Giuliano · Rama: `propose/opsx-baseline` (apply en la misma rama, sin push, decisión del responsable 2026-11-11) · Partida de apply: `43eeb7f`.

## Clasificación
**LOW.** Gobernanza Baja en CHANGES.md; el arnés es externo al binario y no toca el instalador, la configuración del usuario ni rollback. TDD con RED→GREEN por comportamiento; breaker y auditor corren como agentes separados. Planificación y revisión del plan: la propuesta ya fue revisada por tres revisiones del grupo (informe unificado, 2026-11-10); el plan-reviewer formal se fusiona en el orquestador para LOW y se registra acá.

## Estado actual
| Campo | Valor |
|---|---|
| Fase | apply |
| Rol activo | ninguno (tanda 1 cerrada con APROBADO del auditor) |
| Intento | 3 de 3 (corrección de verify-2-breaker: R1 + huecos N_modeTrim, N_omOrder). Autorizado por el responsable 2026-11-11 |
| Revisión / diff | partida `43eeb7f` + diff sin commitear: `eval/baseline/` (11 archivos, 1611 líneas con README; hashes en verify-1-auditor.md), tasks.md 1.1/1.4, design.md (nota OTEL_LOG_LEVEL) |
| Check compartido | `go test -count=1 ./...` 29 paquetes ok (constructor apply-3, 11:09–11:10); `go test -v ./eval/baseline/...` 100 ejecuciones ok (auditor, 11:15:50); gofmt/vet limpios; validate --strict válido (11:16:04) |
| Alcance de esta tanda | tareas 1.1 y 1.4 (lógica pura, datos sintéticos) |
| Bloqueos | Codex CLI no instalado en esta máquina: 1.2, 2.2, 2.3 y 3.x necesitan clientes reales y quedan para otra tanda. D1 árbitro, D3, D5, plazo D6: (TBD) |
| Siguiente acción | decidir commit de la tanda 1 con el responsable; siguiente tanda: 1.3 (fixtures) o 1.2 (requiere clientes reales) |

## Despachos
| # | Fase-intento | Rol | Ejecutor | Informe |
|---|---|---|---|---|
| 1 | apply-1 | constructor | subagente nativo `role-constructor` (Claude; el proyecto no declara tabla Role executors) | `informes/apply-1-constructor.md` |
| 2 | verify-1 | breaker | subagente nativo `role-breaker` (Claude, misma familia que el constructor: sin independencia de proveedor) | `informes/verify-1-breaker.md` |

| 3 | apply-2 | constructor | `role-constructor` (Claude, reanudado) | `informes/apply-2-constructor.md` |

## Disposición verify-1
Breaker: FALLOS REPRODUCIDOS (F1, F2, F3, F8 media; F6 baja-media; F4, F5, F7 baja). F8 se resuelve siguiendo el spec (terminal integrada → cohorte terminal), no requiere decisión nueva. Sospechas S1–S4 quedan visibles sin reproducir.

| 4 | verify-2 | breaker | `role-breaker` (Claude, reanudado) | `informes/verify-2-breaker.md` |

apply-2: F1–F8 corregidos con RED/GREEN y 22 mutantes; `go test ./...` ok 10:59–11:00; validate --strict válido.

verify-2: F1–F8 cerrados; R1 (baja, `OTEL_LOG_LEVEL` tratado como opt-in de contenido) y dos huecos de prueba. Sospechas S1–S6 visibles. Constructor apply-2 detenido a las 11:06 tras terminar (quedó esperando un proceso propio); archivos verificados por hash.

| 5 | apply-3 | constructor | `role-constructor` (Claude, nuevo agente) | `informes/apply-3-constructor.md` |

Si el intento 3 falla: detener y ofrecer al responsable nuevo alcance, deuda aceptada o parar. No hay cuarto intento automático.

apply-3: R1, N_modeTrim y N_omOrder cerrados; `go test ./...` 29 paquetes ok 11:09–11:10; validate --strict válido.

| 6 | verify-3 | breaker | `role-breaker` (Claude, reanudado) | `informes/verify-3-breaker.md` |

verify-3: LIMPIO. R1, N_modeTrim, N_omOrder cerrados; F1–F8 sin regresión; 100 ejecuciones de tests contadas por el breaker. S1–S6 quedan visibles para el grupo.

| 7 | verify-1 (tanda 1) | auditor | `role-auditor` (Claude, misma familia que constructor y breaker) | `informes/verify-1-auditor.md` |

auditor tanda 1: **APROBADO**, sin bloqueantes, 10 puntos de DEUDA (ver `informes/verify-1-auditor.md`, transcripto literal por el orquestador porque el rol auditor no tiene herramienta de escritura).

## DEUDA abierta (del auditor, tanda 1)
constructor: 1 (campos obligatorios del manifiesto), 3 (formato de versión), 5 (claves otel de Codex), 6 (validación de clave y disposiciones), 7 (cosméticos). Grupo/responsable de 02: 2 (rúbrica de 1.4 a revisión de 10, incluye S1), 4 (S5 settings no observados), 8 (nombres y registro de OTEL_LOG_LEVEL, D2 plataformas). Orquestador: 9 (este LEEME, resuelto 2026-11-11), 10 (no commitear auditoria-ramas en 02).

## LECCIONES candidatas
- Al ampliar una lista de seguridad por prefijo, pedir en la misma corrección un test negativo de miembros benignos.
- Para reglas de emparejamiento, test de propiedades contra fuerza bruta con permutación desde RED.
Destino pendiente: no hay knowledge-base/ en el repo.

## Informes faltantes
Ninguno para la tanda 1. Tareas 1.2, 1.3 y grupos 2–4 sin empezar.
