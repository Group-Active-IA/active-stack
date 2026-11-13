# codex-opsx-delivery Specification

## Purpose
Define la activación explícita de OPSX dentro de la sesión actual de Codex, las reglas por rol y acción y los límites comprobables de cada interfaz.

**Alcance de la comprobación de esta entrega.**

Los requisitos siguientes conservan el comportamiento completo objetivo. El ajuste de seguimiento autorizado por Matías separa entrega comprobada, trabajo propio pendiente y deuda externa; no presenta deuda como cumplimiento. Las pruebas de composición/lifecycle no acreditan activación funcional, especialistas o CLI/IDE en celdas no ejecutadas. Los insumos/disposiciones de 02/03/05/06/07/08 y revisores se siguen con IDs en `informes/deudas-aceptadas-05-06-07-2026-11-13.md`, fuera del checklist activo. Matías conserva P2, especialistas y aceptación propia. El mapa literal previo está en `informes/trazabilidad-alcance-y-deuda-04-2026-11-13.md`.

## Requirements

### Requirement: Activación explícita en la sesión actual

El instalador SHALL entregar la skill `sdd-orchestrator` con selección implícita deshabilitada. Una sesión normal SHALL conservar reglas universales sin iniciar coordinación. La invocación explícita SHALL continuar en el hilo actual y consultar el estado de OpenSpec antes de actuar.

#### Scenario: Pedido relacionado sin invocación
- **WHEN** en CLI o extensión IDE declarada se pide proponer un change sin invocar la skill
- **THEN** no se activa implícitamente `sdd-orchestrator` y las reglas universales siguen disponibles

#### Scenario: Invocación manual
- **WHEN** se invoca `$sdd-orchestrator` en cada interfaz declarada
- **THEN** el hilo conserva el contexto previo, consulta OpenSpec y carga los módulos correspondientes a la acción

#### Scenario: Reanudación
- **WHEN** se reanuda una sesión coordinada en una interfaz soportada
- **THEN** se restablecen la acción y los módulos requeridos antes de continuar, sin inferir estado de artefactos

### Requirement: Reglas por rol y acción

La entrega SHALL aplicar la matriz compartida de 03/04 y el corte por acción de 05. El especialista SHALL recibir el encargo, las reglas universales aplicables y los módulos requeridos. El rol coordinador SHALL quedar reservado a la skill invocada.

#### Scenario: Especialista de implementación
- **WHEN** se delega apply con TDD habilitado
- **THEN** el especialista recibe TDD, reglas de autonomía y memoria aplicables, encargo y contrato de evidencia, y no recibe obligaciones exclusivas del coordinador

#### Scenario: Toggle inactivo
- **WHEN** se compone el paquete con un toggle inactivo
- **THEN** se omiten sus módulos y referencias y se informa cualquier acción que no pueda cumplir sus requisitos obligatorios

### Requirement: Evidencia directa y delegada

Una acción que requiere evidencia SHALL cargar el contrato de 06 tanto al ejecutarse directamente como al delegarse. Los retornos delegados SHALL usar el sobre de 07 y su campo de recibos de 06. Un resultado declarado SHALL permanecer sin verificar hasta corroborar ejecución y estado del código y revisar suficiencia.

#### Scenario: Apply directo
- **WHEN** se implementa directamente una tarea que exige evidencia
- **THEN** se generan y revisan sus recibos con los mismos criterios de corroboración y suficiencia que en ejecución delegada

#### Scenario: Retorno autoconsistente sin corroboración
- **WHEN** el worker entrega salida y huella coherentes pero no hay fuente independiente ni reejecución que acredite la ejecución
- **THEN** el resultado queda sin verificar y no acredita aceptación del requisito

### Requirement: Verify con independencia obligatoria

La acción OPSX verify SHALL revisar recibos y cumplimiento. Cuando se requiere independencia, el revisor SHALL ser distinto del constructor y recibir sus reglas y evidencia aplicables.

#### Scenario: Verificación breve
- **WHEN** una verificación breve requiere independencia
- **THEN** su tamaño no elimina el requisito de revisor distinto del constructor

#### Scenario: Delegación desactivada o revisor ausente
- **WHEN** verify requiere independencia y no hay revisor habilitado o falta delegación
- **THEN** se detiene esa verificación, se explica el motivo y se conservan las acciones que sí pueden completarse

### Requirement: Controles y módulos faltantes

La sesión SHALL detener sólo la acción que depende de un módulo o control obligatorio ausente. SHALL explicar el requisito faltante y permitir retomarla con protección equivalente comprobada o en una superficie compatible. La ausencia de un control opcional SHALL informarse como aviso.

#### Scenario: Módulo obligatorio ausente
- **WHEN** falta el módulo de TDD requerido para apply o el contrato de evidencia requerido por la acción
- **THEN** esa acción se detiene indicando qué pieza falta y cómo puede restablecerse

#### Scenario: Protección equivalente comprobada
- **WHEN** una protección equivalente cubre el requisito obligatorio de la acción
- **THEN** la acción puede retomarse registrando la protección efectiva

#### Scenario: Control opcional ausente
- **WHEN** falta un control opcional para una acción cuyos requisitos obligatorios están cubiertos
- **THEN** se continúa con aviso del límite

### Requirement: Soporte y fallback por interfaz

La aceptación SHALL cubrir CLI y extensión IDE en Windows y Linux por separado, con versión y permisos efectivos registrados. Cada fallback SHALL declarar sus capacidades conservadas y faltantes y aplicar las precondiciones de la acción. Una terminal integrada que ejecuta CLI SHALL contabilizarse como CLI.

#### Scenario: Subagente genérico equivalente
- **WHEN** no hay rol personalizado pero la superficie permite un subagente con reglas, módulos e independencia requeridos
- **THEN** se declara la alternativa, se pasa el encargo completo y se valida el retorno según 07/06

#### Scenario: Skills OpenSpec sin garantías suficientes
- **WHEN** se ofrece una skill OpenSpec y no puede proveer un módulo, control o independencia obligatorio
- **THEN** se describe lo que sí permite y se detiene la acción afectada sin presentarla como coordinación completa

#### Scenario: Superficie fuera del soporte comprobado
- **WHEN** se intenta coordinar en una versión o interfaz sin cobertura declarada
- **THEN** se informa el límite y la superficie compatible disponible, si existe

### Requirement: Verificación estructural y funcional diferenciadas

Verify de instalación SHALL comprobar archivos administrados, política de invocación, referencias y toggles. La aceptación con cliente SHALL comprobar activación y contexto efectivo por interfaz. El reporte SHALL distinguir ambos alcances y preservar el comportamiento de otros agentes y design-mockups.

#### Scenario: Archivos correctos sin aceptación del IDE
- **WHEN** pasan chequeos estáticos pero no se probó la extensión IDE
- **THEN** sólo se acredita la estructura instalada y se deja pendiente la aceptación funcional de esa interfaz

### Requirement: Módulos de contexto con fuente identificada

La entrega Codex SHALL incorporar los cuatro módulos textuales disponibles de 05 con revisión identificada, sin frontmatter de skill y sin toggle propio. SHALL adaptar la tabla a rutas relativas Codex y a los toggles existentes, sin modificar la entrega Claude. Verificar la instalación SHALL rechazar cada módulo ausente, vacío o modificado.

#### Scenario: Entrega y actualización de módulos de 05
- **WHEN** se instala o actualiza el paquete Codex con los módulos disponibles de 05
- **THEN** se inventarían sus cuatro archivos con los bytes de la revisión fijada y la reinstalación idéntica no los reescribe
- **AND** el retiro elimina sólo archivos propios y conserva notas del usuario y la primera copia de migración

#### Scenario: Módulo vacío o sólo espacios
- **WHEN** un módulo de la tabla se elimina o contiene cero bytes o sólo espacios
- **THEN** falla la verificación estructural y no se adopta silenciosamente el contenido alterado
- **AND** la skill instruye detener sólo la acción que requiere el módulo obligatorio

#### Scenario: Módulo opcional ausente
- **WHEN** falta un módulo opcional y los requisitos obligatorios de la acción están satisfechos
- **THEN** se informa el límite y se continúa sin afirmar equivalencia con la capacidad ausente

#### Scenario: Reanudación tras compactación
- **WHEN** se compacta o reanuda una sesión coordinada
- **THEN** se requiere reinvocar la skill y releer los módulos de la acción antes de continuar

### Requirement: Requisitos parciales de evidencia sin contrato ficticio

La entrega SHALL identificar los requisitos normativos consumidos de 06 como parciales mientras falte su contrato final. SHALL mantener el mismo estándar directo y delegado, controles separados y estado exacto probado. No SHALL presentar el texto parcial, una huella coherente o la instalación de módulos como ejecución verificada.

#### Scenario: Norma parcial con delegación desactivada
- **WHEN** se compone Codex con el toggle de delegación desactivado
- **THEN** siguen disponibles las reglas parciales de evidencia en la skill explícita y se declara que el contrato completo aún falta
- **AND** no se crea un módulo final o esquema de recibos en nombre de 06

#### Scenario: Huella válida con procedencia o suficiencia ausente
- **WHEN** un resultado contiene una huella coherente pero no corrobora ejecución y árbol exacto o no cubre el requisito
- **THEN** no se marca verificado y se informan por separado integridad, procedencia y suficiencia

#### Scenario: Bytes y estados distintos
- **WHEN** dos corridas comparten HEAD pero tienen diffs distintos, o sus copias usan bytes LF y CRLF distintos
- **THEN** se exige distinguir los estados exactos y las huellas de bytes sin normalizar fines de línea
- **AND** no se inventa el mecanismo operativo de huella del árbol pendiente de 06

### Requirement: Estado de entrega y deuda sin aceptación ficticia

El seguimiento SHALL distinguir las unidades propias entregadas, las pendientes propias y los insumos o disposiciones externos registrados como deuda. Una unidad parcial completada MUST NOT acreditar el requisito funcional completo ni una aprobación ajena. El reporte SHALL identificar proveedor, disparador, criterio y riesgo de cada deuda externa, conservando el comportamiento de bloqueo ante requisitos obligatorios ausentes.

#### Scenario: Dependencia externa registrada como deuda
- **WHEN** un requisito necesita un contrato, medición o disposición de otro responsable que aún no llegó
- **THEN** se registra fuera del checklist activo de Matías con su ID y criterio de reanudación, sin marcarlo implementado ni verificado
- **AND** el consumo propio se retoma al recibir el insumo y la acción que exige un obligatorio ausente permanece detenida

#### Scenario: Unidad parcial entregada y comportamiento completo no probado
- **WHEN** se completan normas textuales, pruebas estructurales o un paquete de revisión pero faltan especialistas o aceptación real de la interfaz
- **THEN** se acredita solo esa parte y se conservan explícitos los pendientes propios, sin declarar concluido o archivado el change ni cerrar el PR
