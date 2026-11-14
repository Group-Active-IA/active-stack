# selective-delegation

## Purpose

Define cuándo el coordinador OPSX resuelve directo y cuándo delega, qué lleva un encargo, cómo se protege la propiedad de archivos, cuándo se paraleliza y qué devuelve un worker, con el mismo contenido en Claude y en Codex.

## ADDED Requirements

### Requirement: Independencia obligatoria primero

Cuando la acción exige que quien revisa no sea quien construyó, el coordinador SHALL asignar la revisión a un agente distinto del que implementó, y ninguna contraseñal MUST anular esa exigencia. Si no hay módulo de delegación o revisor habilitado, el coordinador MUST detener esa verificación, explicar qué falta y continuar con las tareas que no dependen de ella.

#### Scenario: Verificación breve con separación obligatoria

- **WHEN** la acción es verificar un apply de una sola línea y la acción exige revisión independiente
- **THEN** la verificación la ejecuta un agente distinto del constructor
- **AND** el registro indica que la decisión la cerró la independencia obligatoria, no la contraseñal de tarea breve

#### Scenario: Delegación desactivada con separación obligatoria

- **WHEN** el toggle `delegation` no está activo y una acción exige revisión independiente
- **THEN** el coordinador no ejecuta esa verificación como si fuera independiente
- **AND** informa qué falta y continúa con las demás tareas

#### Scenario: Módulo de delegación ausente

- **WHEN** la acción exige revisión independiente y el módulo de delegación no está disponible
- **THEN** el coordinador detiene esa verificación e informa qué módulo falta

### Requirement: Decisión directo o delegado por señal

Fuera de la independencia obligatoria, el coordinador SHALL serializar o redefinir el alcance ante un conflicto de archivos, resolver directo si aplica una contraseñal (tarea breve, dependencia del contexto vivo, relectura completa del resultado), delegar si aplica una señal (salida verbosa, exploración con conclusión, trabajo paralelizable) y resolver directo por defecto. Cada decisión MUST registrarse con la acción, el paso que la cerró, la señal o contraseñal y la modalidad.

#### Scenario: Tarea breve se resuelve directo

- **WHEN** la tarea es corregir un mensaje y su test en un solo paquete, sin revisión independiente exigida
- **THEN** el coordinador la resuelve sin lanzar un worker
- **AND** el registro contiene la contraseñal "tarea breve" y la modalidad "directo"

#### Scenario: Tarea con salida verbosa se delega

- **WHEN** la tarea requiere correr una suite cuya salida supera el umbral de salida verbosa de 05 (más de 300 líneas o más de 10.000 caracteres)
- **THEN** el coordinador delega y recibe sólo el resumen acotado y los recibos
- **AND** el registro contiene la señal "salida verbosa"

#### Scenario: Decisión sin registro

- **WHEN** un informe de ronda contiene una delegación sin señal registrada
- **THEN** la delegación se reporta como no justificada en la auditoría de la ronda

### Requirement: Módulo de delegación cargado sólo al delegar

La regla de decisión SHALL estar disponible en la skill de invocación explícita del coordinador. El módulo de delegación y coordinación MUST cargarse mediante el mecanismo de carga por acción de 05 únicamente cuando la decisión es delegar.

#### Scenario: Resolución directa no carga el módulo

- **WHEN** el coordinador decide resolver directo una acción sin revisión independiente exigida
- **THEN** el contexto de la sesión no contiene el módulo de delegación y coordinación

### Requirement: Encargo mínimo

Toda delegación SHALL incluir objetivo, acción OPSX, partida, conjunto de escritura, punteros de lectura, límites, criterios de salida, contrato de retorno y tope de intentos. Cuando el worker ejecuta comandos, el encargo MUST incluir el extracto del contrato de evidencia de 06.

#### Scenario: Encargo completo

- **WHEN** el coordinador delega una tarea de apply
- **THEN** el prompt del worker contiene los nueve campos del encargo
- **AND** la partida identifica commit base y huella del estado de trabajo

#### Scenario: Encargo sin conjunto de escritura

- **WHEN** un encargo no declara conjunto de escritura
- **THEN** el worker se trata como de sólo lectura y cualquier cambio atribuible lo deja fuera de alcance

### Requirement: Propiedad contra el punto de partida

El coordinador SHALL atribuir a un worker sólo los cambios entre la partida registrada de su encargo y el estado final de su árbol de trabajo, incluidos commits nuevos, cambios preparados y sin preparar, archivos nuevos, borrados y renombres. Toda ruta atribuible fuera del conjunto de escritura MUST rechazar el retorno con estado `out_of_scope`. Los cambios presentes en la partida MUST NOT atribuirse al worker.

#### Scenario: Archivo ajeno preexistente

- **WHEN** la partida ya contiene un archivo modificado por el usuario fuera del encargo y el worker no lo toca
- **THEN** ese archivo no se atribuye al worker y el retorno no se rechaza por él

#### Scenario: Cambio confirmado fuera de alcance

- **WHEN** el worker confirma en un commit un cambio sobre una ruta que no está en su conjunto de escritura
- **THEN** el retorno se rechaza como `out_of_scope` aunque los tests pasen
- **AND** el cambio no se integra

#### Scenario: Renombre hacia afuera del alcance

- **WHEN** el worker renombra un archivo de su conjunto de escritura a una ruta fuera de él
- **THEN** la ruta destino cuenta como tocada y el retorno se rechaza como `out_of_scope`

#### Scenario: Retorno declara menos archivos que los atribuibles

- **WHEN** el campo `modificados` omite una ruta atribuible
- **THEN** el retorno queda sin verificar y no se acepta por inferencia

### Requirement: Comprobación del resultado combinado

Antes de integrar cambios de más de un encargo, el coordinador SHALL volver a comprobar el alcance sobre el resultado combinado y MUST re-ejecutar los escenarios que dependan de combinar esos cambios. Esa re-ejecución SHALL correr con captura propia de una fuente independiente (06 D3) para que sus recibos corroboren; sin esa captura MUST NOT contar como corroboración.

#### Scenario: Dos encargos correctos que rompen al combinarse

- **WHEN** dos retornos aceptados por separado hacen fallar un escenario al combinarse
- **THEN** el resultado combinado no se integra y el fallo se registra contra ambos encargos

### Requirement: Paralelismo condicionado

Hasta dos workers con escritura SHALL ejecutarse en paralelo, cada uno en su propio árbol de trabajo, sólo si sus conjuntos de escritura son disjuntos, ninguno depende de lo que el otro escribe y cada uno tiene pruebas propias. Los encargos que no cumplan esas condiciones MUST serializarse.

#### Scenario: Cambios en paquetes disjuntos

- **WHEN** dos subtareas escriben en paquetes distintos sin archivos calientes compartidos
- **THEN** el coordinador puede lanzarlas en paralelo, cada una en su árbol de trabajo

#### Scenario: Tercer worker con escritura

- **WHEN** ya hay dos workers con escritura en curso y aparece una tercera subtarea con escritura
- **THEN** la tercera espera a que termine alguno de los dos

### Requirement: Archivos calientes nunca concurrentes

Los archivos calientes declarados (`go.mod`, `go.sum`, `internal/catalog/harnesses.yaml`, `internal/harness/config/assets/*/sdd-orchestrator.md`, `CLAUDE.md`, `AGENTS.md`) MUST NOT asignarse a dos workers concurrentes.

#### Scenario: Cambios que compiten por el mismo archivo

- **WHEN** dos subtareas necesitan escribir `internal/catalog/harnesses.yaml`
- **THEN** el coordinador las serializa o redefine su alcance
- **AND** registra la contraseñal "conflicto de archivos"

### Requirement: Retorno estructurado

Todo worker SHALL devolver estado (`completed`, `blocked`, `out_of_scope` o `failed`), partida, archivos modificados con su tipo de cambio, hallazgos acotados, recibos de ejecución según 06 y pendiente. El coordinador MUST validar en orden la partida, la propiedad, los recibos y los criterios de salida. Un campo faltante o no comprobable MUST dejar el retorno sin verificar.

#### Scenario: Retorno en prosa libre

- **WHEN** un worker devuelve sólo un resumen en texto sin los campos del contrato
- **THEN** el retorno queda sin verificar y el coordinador pide el contrato o reintenta

#### Scenario: Partida distinta de la registrada

- **WHEN** la partida declarada por el worker no coincide con la registrada por el coordinador
- **THEN** el retorno queda sin verificar

### Requirement: Declaración del worker sin corroborar

Lo que el worker declara SHALL conservarse, pero MUST quedar sin verificar hasta que una fuente independiente o una re-ejecución con captura propia de una fuente independiente corrobore la ejecución y el estado del código y la evidencia alcance para el requisito. El mismo contrato y la misma revisión de recibos SHALL aplicarse cuando la acción se resuelve directo y requiere evidencia.

#### Scenario: Tests verdes declarados sin corroboración

- **WHEN** el retorno declara que los tests pasan con recibos completos pero ninguna fuente independiente ni re-ejecución los corrobora
- **THEN** los requisitos que dice cubrir quedan sin verificar

#### Scenario: Re-ejecución del coordinador sin captura independiente

- **WHEN** el coordinador re-ejecuta los tests de un retorno sin captura propia de una fuente independiente y obtiene el mismo resultado
- **THEN** esa re-ejecución no corrobora ni contradice y los requisitos siguen sin verificar

#### Scenario: Re-ejecución con captura independiente

- **WHEN** la re-ejecución de integración corre con la captura del arnés sobre el mismo estado del árbol y confirma comando, código de salida y estado
- **THEN** la procedencia de esos recibos queda corroborada y el requisito puede quedar verificado si la integridad y la suficiencia también se cumplen

#### Scenario: Ejecución directa que requiere evidencia

- **WHEN** el coordinador resuelve directo una acción que requiere evidencia
- **THEN** produce recibos con el mismo contrato y pasa por la misma revisión de recibos

### Requirement: Reintentos acotados

Cada encargo SHALL tener un tope de 3 intentos en total salvo acuerdo distinto del grupo. Cada reintento MUST recibir el motivo del rechazo anterior y partir de una nueva partida registrada. Agotado el tope, el coordinador MUST escalar a una persona con el historial.

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
- **THEN** el resultado contiene la regla de decisión con independencia obligatoria primero y no contiene el umbral "1-3 / 4+ archivos"

### Requirement: Paridad entre Claude y Codex

Los assets de Claude y de Codex SHALL entregar la misma regla de decisión, el mismo encargo mínimo y el mismo contrato de retorno, y MUST producir el mismo veredicto sobre el mismo conjunto de retornos de prueba. Una diferencia de capacidad del cliente MUST quedar declarada como límite.

#### Scenario: Mismos veredictos

- **WHEN** se evalúa la misma batería de retornos (completo, en prosa, fuera de alcance confirmado, renombre fuera de alcance, ajeno en la partida, sin corroboración, partida distinta) con el flujo de Claude y con el de Codex
- **THEN** ambos devuelven el mismo veredicto para cada retorno
