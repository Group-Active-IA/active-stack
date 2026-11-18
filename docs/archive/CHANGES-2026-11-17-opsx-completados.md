# Historial de mejoras OPSX completadas

[Volver al roadmap](../../CHANGES.md)

La ficha 01 se conserva como antecedente de la propuesta. OpenSpec registra [7 de 7 tareas completas](../../openspec/changes/codex-strict-compat/tasks.md). La implementación quedó en el commit `7229a44`.

<a id="change-01"></a>

### 01 · Compatibilidad del tier estricto

**Change** `codex-strict-compat` · **Responsable** Juan Cruz · **Cliente** Codex · **Dependencias** —

**Problema.** El instalador genera `approval_policy = "untrusted"` para el tier estricto. Codex CLI 0.166.0 rechaza ese valor, así que una instalación o actualización puede dejar una configuración que el cliente no acepta.

El caso registrado en el informe se reproduce con `codex -c 'approval_policy="untrusted"' features list`. El cliente devuelve código 1 y señala que el valor dejó de estar soportado. El override es temporal y sirve para comprobar la compatibilidad del parser. La prueba de permisos efectivos requiere casos adicionales en un entorno aislado.

**Qué buscar.** Reconstruí qué escribe cada tier y dónde se respalda el archivo. Reproducí el rechazo solo con un override o un perfil temporal. Trazá si `update` pasa por el instalador de permisos. Si existe una ruta que lo omite, documentá cómo se migraría y quién la ejecutaría. La selección de política debe comprobarse con el cliente real; un chequeo de presencia de claves es insuficiente.

**Qué se implementaría.** Explicá qué garantías debe conservar el tier estricto y compará las políticas compatibles. Diseñá la migración del valor legado en instalación y actualización, con snapshot, preservación de claves y perfiles ajenos, reinstalación idempotente y recuperación ante fallo. Especificá escenarios de carga y permisos efectivos en entornos soportados.

**Qué acordar en grupo.** Qué versiones y entornos se soportan, qué operaciones debe permitir el tier y qué política compatible conserva esas garantías.

**Aceptación futura.** Config aceptada por las versiones soportadas (TBD). Probar instalación limpia, migración, reinstalación y rollback. Verificar permisos efectivos en los entornos soportados, con casos permitidos y denegados.

**Leer antes**

- [internal/harness/config/permissions/install.go](../../internal/harness/config/permissions/install.go)
- [internal/harness/config/permissions/codex_test.go](../../internal/harness/config/permissions/codex_test.go)
- [internal/filemerge/toml.go](../../internal/filemerge/toml.go)
- [internal/verify/harness_checks.go](../../internal/verify/harness_checks.go)
- [cmd/active-stack/update.go](../../cmd/active-stack/update.go)
- [Evidencia y recomendaciones del informe ejecutivo](https://active-stack-executive-site.vercel.app/#evidencia)
