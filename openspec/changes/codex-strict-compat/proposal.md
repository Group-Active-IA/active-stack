# Proposal

## Why

Active Stack escribe `approval_policy = "untrusted"` al elegir Estricto para Codex. Codex CLI 0.166.0 en Windows rechazó esa configuración con código 1 en una prueba por override temporal. Con `sandbox_mode = "read-only"`, el mismo comando devolvió 0 con `on-request` y con `never`.

## What Changes

- Configurar Estricto para Codex como solo lectura con solicitudes de aprobación habilitadas para escalaciones admitidas, siguiendo la [guía oficial de migración](https://learn.chatgpt.com/docs/agent-approvals-security#migrate-from-the-retired-untrusted-approval-policy).
- Reaplicar la configuración mediante el instalador existente al actualizar una selección registrada de Codex en Estricto. Guiar la reinstalación explícita para instalaciones sin registro.
- Ajustar la descripción de Estricto para Codex y verificar carga, permisos efectivos, preservación de preferencias y recuperación.

**Descripción corta** · «Solo lectura por defecto».

**Descripción larga** · «Codex trabaja en solo lectura. Las acciones que requieren permisos adicionales necesitan aprobación explícita cuando la interfaz admite solicitarla. Si la interfaz carece de ese mecanismo, la acción permanece bloqueada».

## Capabilities

### New Capabilities

- `codex-strict-permissions` cubre la configuración y migración del tier Estricto de Codex.

### Modified Capabilities

## Impact

Instalador de permisos de Codex, texto del selector y verificación. El alcance excluye Balanceado, Bypass, otros clientes y el flujo OPSX. La revisión humana precede a la implementación que escribe configuración del usuario.
