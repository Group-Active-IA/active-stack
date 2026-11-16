## Context

Claude y Codex ya componen una capa universal y una skill explícita. El catálogo genera la tabla de módulos.

## Decisions

- Compartir un texto universal breve con referencias específicas del cliente. Sugerir OPSX cuando propuesta, especificación, diseño y tareas sirvan para acordar una solución compleja. Resolver trabajo cotidiano directamente.
- Condicionar la revisión independiente a gobernanza o requisitos del change. Permitir revisión humana documentada y verificación directa de alcance acotado. Mantener evidencia por requisito y separar la independencia del revisor de la corroboración de la ejecución.
- Probar los cambios con instalaciones aisladas. Reproducir Windows antes de modificar componentes de gobernanza alta.

### Línea base de Windows

Sobre 8bfca74, `go test ./internal/harness/config ./internal/catalog` pasó en catálogo y falló en configuración antes de editar código. `TestManagedOperationSnapshotSymlinkIsRejected` no pudo crear un enlace por falta de privilegios. `TestManagedInstallRemovePreserveParentPermissions` observó modo 777 en el padre. Queda pendiente distinguir el comportamiento de Windows de una regresión en permisos y repetir en un entorno representativo.

### Revisión de skills de fundación

- `active-orchestrator` en `63c5131` reconoce capacidades equivalentes por cliente y admite invocación directa de sus sub-skills cuando falta delegación. Su salida recomienda `/opsx:propose`; esa referencia debe resolverse según el cliente en Codex.
- `roadmap-generator` en `deaabe9` también imprime `/opsx:propose` en la plantilla y al cerrar. La adaptación debe conservar la recomendación para arrancar un change del roadmap y presentar la invocación que el cliente tenga instalada.
- El repositorio del harness `agent-instruction` en `1f7922f` declara `name: agents-md-generator`. El prompt depende de rutas globales y locales de Claude incluso al generar AGENTS.md. La corrección debe alinear el nombre que invoca fundación con el ID instalado y consultar las instrucciones y el registro del cliente activo.
- Los prompts principales de `kb-creator` (`0f73e9b`) y `discovery-research` (`dd7efb7`) no contienen referencias a SDD, OPSX, aliases de Claude ni a las rutas de instrucciones buscadas. La prueba real todavía debe comprobar su invocación en ambos clientes.

Los repositorios fuente publicaron las correcciones en ramas de revisión. El catálogo candidato fija d862278 (Active Orchestrator), dd45758 (Roadmap Generator) y a873714 (Agent Instruction). El instalador real de skills clonó las tres revisiones e instaló cada SKILL.md bajo las rutas de Claude y Codex en perfiles temporales.

### Pruebas de distribución y clientes

- Codex 0.166.0 en Windows: `TestManagedSkillCodexDiscovery` pasó contra el app-server real en un perfil aislado. Comprueba descubrimiento, no ejecución de modelo.
- Codex: `TestCodexPre05PackageUpgrade` pasó con una instalación sembrada por el instalador de 91fbb70. Claude: `TestInstallManaged_MigratesLegacyBlock` pasó, incluida la presencia del módulo de evidencia y preservación del texto del usuario.
- Claude Code 2.1.309 en Windows: el perfil creado con `CLAUDE_CONFIG_DIR` temporal informa `loggedIn: false`. Queda pendiente probar invocación y compactación con un perfil de prueba autenticado.

### Resultados de estabilización

- `go test ./...` pasó en Windows con el CLI de captura y ambos paquetes instalables. Los dos escenarios que requieren crear symlinks se omitieron por el error 1314 de Windows. Las pruebas de permisos compararon el modo observado antes y después, no las ACL.
- La prueba real de Codex respondió directo a la tarea cotidiana y sugirió una propuesta OPSX para una migración transversal. La invocación explícita reconoció la skill generada. El proceso shell de la sesión falló al iniciarse con el ejecutor habitual y con `unified_exec` desactivado, ambos en modo de solo lectura.
- OpenSpec 1.16.0 devuelve «Artifact verify not found» para `openspec instructions verify`. La integración usa status y los artefactos para verify; el perfil predeterminado de OpenSpec no instala su skill de verify.
