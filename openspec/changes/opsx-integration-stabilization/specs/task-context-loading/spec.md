## ADDED Requirements

### Requirement: Carga de delegación según modalidad de verify

Durante verify el módulo de delegación SHALL ser obligatorio para una revisión delegada. La verificación directa y la revisión humana documentada SHALL poder operar en superficies sin subagentes.

#### Scenario: Superficie sin subagentes

- **WHEN** se verifica directamente un change que admite revisión directa
- **THEN** la ausencia de subagentes no detiene verify

#### Scenario: Revisión delegada sin módulo

- **WHEN** verify intenta una revisión delegada y falta el módulo
- **THEN** el coordinador informa la revisión pendiente
