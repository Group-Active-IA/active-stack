# selective-delegation

## Purpose

Define cuándo el coordinador OPSX resuelve directo y cuándo delega, qué lleva un encargo, cómo se protege la propiedad de archivos, cuándo se paraleliza y qué devuelve un worker, con el mismo contenido en Claude y en Codex.

## ADDED Requirements

### Requirement: Decisión directo o delegado por señal

El coordinador SHALL resolver directo por defecto y delegar sólo cuando la tarea presenta al menos una señal declarada (salida verbosa, exploración con conclusión, trabajo paralelizable, separación de rol exigida por la acción) y ninguna contraseñal (tarea breve, dependencia del contexto vivo, conflicto de archivos, relectura completa del resultado). Cada decisión MUST quedar registrada con la acción, la señal o contraseñal y la modalidad elegida.

#### Scenario: Tarea breve se resuelve directo

- **WHEN** la tarea es corregir un mensaje y su test en un solo paquete
- **THEN** el coordinador la resuelve sin lanzar un worker
- **AND** el registro contiene la contraseñal "tarea breve" y la modalidad "directo"

#### Scenario: Tarea con salida verbosa se delega

- **WHEN** la tarea requiere correr una suite cuya salida supera el presupuesto acotado de 05
- **THEN** el coordinador delega y recibe sólo el resumen acotado y los recibos
- **AND** el registro contiene la señal "salida verbosa"

#### Scenario: Decisión sin registro

- **WHEN** un informe de ronda contiene una delegación sin señal registrada
- **THEN** la delegación se reporta como no justificada en la auditoría de la ronda

### Requirement: Módulo de delegación cargado sólo al delegar

La regla de decisión SHALL estar disponible en la definición del coordinador. El módulo de delegación y coordinación MUST cargarse mediante el mecanismo de carga por acción de 05 únicamente cuando la decisión es delegar. Si el módulo no está disponible y la acción exige separación de rol, la acción MUST detenerse e informar la limitación.

#### Scenario: Resolución directa no carga el módulo

- **WHEN** el coordinador decide resolver directo
- **THEN** el contexto de la sesión no contiene el módulo de delegación y coordinación

#### Scenario: Módulo ausente con separación de rol obligatoria

- **WHEN** la acción verify exige un revisor separado y el módulo de delegación no está disponible
- **THEN** el coordinador informa qué módulo falta y no ejecuta la verificación como si fuera independiente

### Requirement: Encargo mínimo

Toda delegación SHALL incluir objetivo, acción OPSX, conjunto de escritura, punteros de lectura, límites, criterios de salida, contrato de retorno y tope de intentos. Cuando el worker ejecuta comandos, el encargo MUST incluir el extracto del contrato de evidencia de 06.

#### Scenario: Encargo completo

- **WHEN** el coordinador delega una tarea de apply
- **THEN** el prompt del worker contiene los ocho campos del encargo
- **AND** el conjunto de escritura enumera rutas o globs explícitos

#### Scenario: Encargo sin conjunto de escritura

- **WHEN** un encargo no declara conjunto de escritura
- **THEN** el worker se trata como de sólo lectura y cualquier modificación lo deja fuera de alcance

### Requirement: Propiedad exclusiva de archivos

Mientras un encargo está abierto, cada ruta de su conjunto de escritura SHALL pertenecer a un único worker. El coordinador MUST comparar los archivos modificados según el diff real, incluidos los nuevos sin trackear, contra el conjunto de escritura y rechazar con estado `out_of_scope` cualquier retorno que modifique rutas fuera de él.

#### Scenario: Worker modifica un archivo ajeno

- **WHEN** el diff del worker incluye una ruta que no está en su conjunto de escritura
- **THEN** el retorno se rechaza como `out_of_scope` aunque los tests pasen
- **AND** el cambio no se integra

#### Scenario: Retorno declara menos archivos que el diff

- **WHEN** el campo `modificados` del retorno omite una ruta presente en el diff
- **THEN** el retorno queda sin verificar y no se acepta por inferencia

### Requirement: Paralelismo condicionado

Dos workers SHALL ejecutarse en paralelo sólo si sus conjuntos de escritura son disjuntos, ninguno lee lo que el otro escribe en ese encargo y cada uno tiene verificación propia. Los workers con escritura en paralelo MUST trabajar en árboles de trabajo separados, y la concurrencia máxima SHALL ser un valor declarado.

#### Scenario: Cambios en paquetes disjuntos

- **WHEN** dos subtareas escriben en paquetes distintos sin archivos calientes compartidos
- **THEN** el coordinador puede lanzarlas en paralelo, cada una en su árbol de trabajo

### Requirement: Archivos calientes nunca concurrentes

Los archivos calientes declarados (`go.mod`, `go.sum`, `internal/catalog/harnesses.yaml`, `assets/*/sdd-orchestrator.md`, `CLAUDE.md`, `AGENTS.md`) MUST NOT asignarse a dos workers concurrentes.

#### Scenario: Cambios que compiten por el mismo archivo

- **WHEN** dos subtareas necesitan escribir `internal/catalog/harnesses.yaml`
- **THEN** el coordinador las serializa
- **AND** registra la contraseñal "conflicto de archivos"

### Requirement: Retorno estructurado

Todo worker SHALL devolver estado (`completed`, `blocked`, `out_of_scope` o `failed`), archivos modificados, hallazgos acotados, recibos de ejecución según 06 y pendiente. El coordinador MUST validar primero la propiedad de archivos, después los recibos y por último los criterios de salida. Un campo faltante o no comprobable MUST dejar el retorno sin verificar.

#### Scenario: Retorno en prosa libre

- **WHEN** un worker devuelve sólo un resumen en texto sin los campos del contrato
- **THEN** el retorno queda sin verificar y el coordinador pide el contrato o reintenta

#### Scenario: Retorno que afirma tests verdes sin recibo

- **WHEN** el retorno declara que los tests pasan y no incluye recibos
- **THEN** los requisitos que dice cubrir quedan sin verificar según 06

### Requirement: Reintentos acotados

Cada encargo SHALL tener un tope de intentos, de 3 en total salvo acuerdo distinto del grupo. Cada reintento MUST recibir el motivo del rechazo anterior. Agotado el tope, el coordinador MUST escalar a una persona con el historial de intentos. Los consumos de todos los intentos MUST sumarse al costo de la tarea.

#### Scenario: Tope agotado

- **WHEN** un encargo se rechaza por tercera vez
- **THEN** el coordinador no lanza un cuarto intento y escala con los tres motivos de rechazo

### Requirement: Toggle delegation sin obligaciones residuales

Cuando el toggle `delegation` no está activo, el bloque compuesto SHALL NOT contener ninguna obligación de delegar, y el coordinador MUST declarar que trabaja sin subagentes. Cuando está activo, la regla de decisión y el módulo SHALL estar presentes.

#### Scenario: Composición sin delegation

- **WHEN** se compone el orquestador de Claude o de Codex con todos los toggles salvo `delegation`
- **THEN** el resultado no contiene `ALWAYS delegate`, `Delegate, don't inflate` ni la tabla de umbrales
- **AND** contiene la declaración de trabajo sin subagentes

#### Scenario: Composición con delegation

- **WHEN** se compone con `delegation` activo
- **THEN** el resultado contiene la regla de decisión por señal y no contiene el umbral "1-3 / 4+ archivos"

### Requirement: Paridad entre Claude y Codex

Los assets de Claude y de Codex SHALL entregar la misma regla de decisión, el mismo encargo mínimo y el mismo contrato de retorno, y MUST producir el mismo veredicto sobre el mismo conjunto de retornos de prueba. Una diferencia de capacidad del cliente MUST quedar declarada como límite.

#### Scenario: Mismos veredictos

- **WHEN** se evalúa la misma batería de retornos (completo, en prosa, fuera de alcance, sin recibos, con `modificados` incompleto) con el flujo de Claude y con el de Codex
- **THEN** ambos devuelven el mismo veredicto para cada retorno
