## MODIFIED Requirements

### Requirement: Independencia obligatoria primero

El coordinador SHALL exigir una revisión independiente cuando la gobernanza o los requisitos del change la pidan. Para bajo riesgo SHALL permitir verificación directa con evidencia. Si falta la revisión obligatoria SHALL informar la parte pendiente.

#### Scenario: Verificación directa

- **WHEN** el change es acotado y no exige independencia
- **THEN** el coordinador puede verificar los artefactos y las pruebas directamente

#### Scenario: Revisión independiente

- **WHEN** el change exige independencia
- **THEN** el coordinador requiere otro agente o registra la revisión humana independiente
- **AND** mantiene pendiente la revisión si no hay revisor disponible

#### Scenario: Verificación breve con separación obligatoria

- **WHEN** la acción verifica un cambio de una línea con revisión independiente obligatoria
- **THEN** la brevedad del cambio no dispensa la independencia

#### Scenario: Delegación desactivada con separación obligatoria

- **WHEN** el toggle de delegación está desactivado y se exige independencia
- **THEN** el coordinador registra la revisión humana independiente o informa que sigue pendiente

#### Scenario: Módulo de delegación ausente

- **WHEN** se necesita una revisión delegada y el módulo no está disponible
- **THEN** el coordinador informa que falta el módulo y mantiene pendiente la revisión delegada
