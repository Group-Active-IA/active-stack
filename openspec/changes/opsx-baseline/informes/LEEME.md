# opsx-baseline — estado de orquestación e índice de informes

Change 02 · Responsable: Giuliano · Rama: `propose/opsx-baseline` (apply en la misma rama, sin push, decisión del responsable 2026-11-11) · Partida de apply: `43eeb7f`.

## Clasificación
**LOW.** Gobernanza Baja en CHANGES.md; el arnés es externo al binario y no toca el instalador, la configuración del usuario ni rollback. TDD con RED→GREEN por comportamiento; breaker y auditor corren como agentes separados. Planificación y revisión del plan: la propuesta ya fue revisada por tres revisiones del grupo (informe unificado, 2026-11-10); el plan-reviewer formal se fusiona en el orquestador para LOW y se registra acá.

## Estado actual
| Campo | Valor |
|---|---|
| Fase | apply |
| Rol activo | ninguno (tanda 4 construida y verificada por conformidad) |
| Intento | tanda 3, intento 1 |
| Revisión / diff | base `5dcfc68` (tanda 1) + tanda 2 sin commitear (eval/baseline/fixture/, fixtures/ 41 archivos, review DEUDA 6), hashes en apply-t2-1-constructor.md (load_test.go real `c5b635cc`) |
| Check compartido | tanda 2: `go test ./...` 30 paquetes ok (14:31–14:34); breaker 146 ejecuciones ok; auditor 107 PASS review+manifest y fixture enfocado ok |
| Alcance de esta tanda | tanda 3: 2.1, 2.4, 2.5 (parte A) y 2.6, 2.7, 3.1 (parte B), un solo breaker y auditor |
| Bloqueos | Codex CLI no instalado en esta máquina: 1.2, 2.2, 2.3 y 3.x necesitan clientes reales y quedan para otra tanda. D1 árbitro, D3, D5, plazo D6: (TBD) |
| Siguiente acción | constructor parte A, luego parte B, luego breaker y auditor de la tanda 3 |

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

Tanda 1 commiteada en `5dcfc68` (sin push).

| 8 | apply-t2-1 | constructor | `role-constructor` (Claude, nuevo) | `informes/apply-t2-1-constructor.md` |

apply-t2-1: 1.3 + DEUDA 6 hechas; `go test ./...` 30 paquetes ok 14:31–14:34; validate válido. Cambios ajenos en el árbol (move de codex-strict-compat a archive/, borrado de revision-2026-11-11.md restaurado por el orquestador) no son del constructor.

| 9 | verify-t2-1 | breaker | `role-breaker` (Sonnet, nuevo; medida de consumo) | `informes/verify-t2-1-breaker.md` |

verify-t2-1 breaker: sin ALTA/MEDIA; F1–F3 BAJA a DEUDA (directorio vacío ignorado por TreeHash; fragmento de reproducción trivial; comando de huella del README en Git Bash). 146 ejecuciones de tests.

| 10 | verify-t2-1 | auditor | `role-auditor` (Opus) | `informes/verify-t2-1-auditor.md` |

auditor tanda 2: **APROBADO** (`informes/verify-t2-1-auditor.md`, transcripto literal). DEUDA 1: ratificar fixtures sintéticos en 3 familias frente a D1 (responsable de 02). DEUDA 2–3, 6: constructor en 2.1 (directorios vacíos/symlinks/guarda de rutas, `--text` en README, duración). DEUDA 7: R3 estricto con la rúbrica de 10.

Aclaración DEUDA 4 (orquestador): la regla BAJA→DEUDA fue autorizada por el responsable el 2026-11-11 antes de despachar el breaker de la tanda 2; el constructor de la tanda 2 ya corría, pero la ronda de breaker y la auditoría se rigieron por la regla nueva.

| 11 | apply-t3-A | constructor | `role-constructor` (Sonnet) | `informes/apply-t3-A-constructor.md` |

apply-t3-A: 2.1, 2.4, 2.5 + DEUDA 2, F3 hechas; DEUDA 6 a medias (fixture ~104–111 s); `go test ./...` 33 paquetes ok 15:03–15:07. Las ediciones en design.md/proposal.md son del orquestador (D1 ratificado, D3, D5 con el responsable, 2026-11-11).

| 12 | apply-t3-B | constructor | `role-constructor` (Sonnet) | `informes/apply-t3-B-constructor.md` |

apply-t3-B: 2.6, 2.7, 3.1 + cancelación y pendiente_de_revisor; `go test ./...` ok 15:24–15:26. verify tanda 3: `informes/verify-t3-conformidad.md` (solo conformidad, por decisión del responsable).

## Informes faltantes
Tanda 3: breaker y auditor no ejecutados por decisión del responsable (2026-11-11). Tandas 1 y 2: ninguno. Tanda 1: ninguno. Tareas 1.2, 1.3 y grupos 2–4 sin empezar.

## Decisión del responsable · medidas de consumo (2026-11-11)
Autorizado por Giuliano en la conversación ("aplica esas medidas"), vigente desde la tanda 3:
1. **Tandas más grandes:** se agrupan varias tareas sintéticas por pasada de constructor → breaker → auditor (p.ej. 2.1 + 2.4 + 2.5).
2. **Severidad BAJA → DEUDA:** un fallo reproducido de severidad BAJA no reabre ronda; se registra como DEUDA con responsable y se corrige en la tanda siguiente que toque ese archivo. ALTA y MEDIA siguen reabriendo ronda. El auditor sigue pudiendo bloquear.
3. **Modelos:** constructor y breaker en Sonnet; auditor en Opus.
4. **Contexto:** se recomienda retomar en sesión nueva desde este LEEME y los commits.
La tanda 2 (1.3) ya estaba en curso con la configuración anterior.

## Decisión del responsable · sin breaker (2026-11-11)
Giuliano decidió en la conversación: "no uses el breaker". Desde la tanda 3 no se despacha el rol breaker; tras el constructor va directo el auditor, que recibe esta decisión y verifica por su cuenta (ataques propios incluidos). Consecuencia declarada: se pierde la ronda de ataque independiente previa a la auditoría; el informe verify-t3-breaker queda ausente por decisión humana, no por omisión. Tandas 1 y 2 conservan sus informes de breaker.

## Decisión del responsable · solo apply y verify (2026-11-11)
Giuliano: "contructor, verify y apply" / "no uses nada que se salga de eso". Desde la tanda 3 el flujo es apply (rol constructor) → verify (`/opsx:verify`). No se despachan breaker ni auditor. Consecuencia declarada: la tanda 3 no tiene veredicto APROBADO del auditor; el cierre de la tanda se apoya en la evidencia del constructor y en el resultado de `/opsx:verify`.


## Tanda 4 (2026-11-11)
Pedido del responsable: "usa la documentacion oficial de claude y codex para realizar los cambios restante". Alcance: 1.2, 2.2, 2.3 con muestras derivadas de la documentación oficial (riesgo declarado, como permite 2.2), sin corridas reales que gasten cuota. 3.2/3.3/3.4 quedan para ejecución real; 4.1 espera 03/04. Ejecutor: agente general-purpose (Sonnet) con contrato de constructor en el brief, porque los agentes `role-*` no están disponibles en esta sesión. Flujo: apply → verify (`/opsx:verify`), sin breaker ni auditor (decisión del responsable). El repo local fue reclonado desde cero antes de esta tanda; se retomó desde `origin/propose/opsx-baseline` en `3f8afd2`.
Informe: `informes/apply-t4-constructor.md`.

apply-t4: 2.2 [x]; 1.2 y 2.3 parciales. verify: `informes/verify-t4-conformidad.md`. Tareas 10/16.

## Decisión del responsable · 2.2 IDE (2026-11-11)
Giuliano eligió la opción (a): 2.2 queda cerrada con la captura del IDE como "no comprobada / no disponible" (D2), porque la documentación oficial no describe una fuente de eventos propia de las extensiones de IDE. La prueba con la extensión real se hace en 3.2. También se corrigió en design.md el ejemplo de Claude: `modelUsage`/`total_cost_usd` incluyen subagentes, `usage` solo el hilo principal.
