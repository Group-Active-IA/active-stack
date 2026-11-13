# Tasks

> Todas las tareas de implementación quedan **pendientes**: esta ronda es solo de propuesta (corregir la propuesta) y solo toca `openspec/changes/opsx-native-controls/`. Flujo del equipo: **contrato → evidencia → decisión**: se corrige la propuesta, se implementa con `apply`, se verifican los criterios de aceptación y se abre el PR. Este change es de gobernanza **ALTO**: antes de implementar hace falta la revisión humana que exige `CLAUDE.md` (revisa e integra Neyén). Cuando llegue la etapa de implementación, TDD estricto, con fakes para el snapshotter y la lectura de archivos, y **nunca** el entorno real del usuario: RED, GREEN, TRIANGULATE y REFACTOR (esquema tomado de los changes archivados, no de una regla escrita en `CHANGES.md`). Antes de tocar archivos existentes, correr el SAFETY NET del paquete y anotar la línea base. No hacer build ni commitear salvo pedido explícito. Si surge la necesidad de tocar `filemerge`, `backup` o `pipeline`, frenar y reportar para revisión. Base revisada: `origin/main` `7a60089` (01, 02 y 03 integrados; 04 pendiente, citada contra `origin/propose/codex-native-opsx` `9a6eb2a`; 06 contra la rama `opsx-verifiable-evidence`, local y pendiente de push: D2, D3 y D9).

## 1. Antes de empezar: acuerdos y reproducciones (ALTO)

- [ ] 1.1 Leer el contrato de 06 (rama `opsx-verifiable-evidence`, local, pendiente de push: D2 campos del recibo de ejecución, D3 tres controles, D9 hooks) y confirmar con su revisor (Juan Cruz) que no cambia antes de integrarlo en 6.1. No se redefine nada de 06.
- [ ] 1.2 Registrar como hechos de `main` la política de Estricto de Codex (`sandbox_mode = "read-only"`, `approval_policy = "on-request"`, `install.go:120-126`) y el método de aceptación de 01 (perfil aislado, cliente real, lectura permitida y escritura con escalación denegada, Windows CLI 0.166.0 primero), y registrar los resultados una sola vez con 01 (D11).
- [ ] 1.3 Cerrar en el grupo las preguntas abiertas del design que dependan solo de 08 y pasar por la revisión humana que exige el nivel ALTO (Neyén). Las que dependen de 03, 04, 05, 06 o 07 siguen (TBD).
- [ ] 1.4 Reproducir en un `HOME` temporal, con el binario y un `settings.json` de prueba, el caso del 02/11: `defaultMode` en `"auto"` y un permiso propio en `allow`. Registrar qué cambia al instalar con cada tier (sin tocar el entorno real). El resultado decide si D9 pasa a requisito de implementación y qué se arregla en `filemerge`.
- [ ] 1.5 Con 03 (Juani): confirmar que D12 sigue vigente (activación por `/sdd-orchestrator`, `agent` sin escribir) y acordar con el revisor ALTO la integración serializada de los arreglos de `WriteFileAtomic` (03 G6, alcance chico, sin cambios en `filemerge` global) y de `MergeJSONObjects` (D9).

## 2. Inventario (BAJO, solo lectura)

- [ ] 2.1 Listar los mecanismos de permisos y hooks de cada cliente y versión soportada. Claude: `onFailure: "block"` desde v2.1.295 para hooks `command`/`http`. Codex: confirmar desde qué versión hay hooks y en qué interfaces (la documentación no lo dice, TBD), dónde se configuran (`hooks.json` o `[hooks]`) y el modelo de confianza por hash.
- [ ] 2.2 Listar las superficies declaradas de cada cliente y los shells reales por sistema operativo.
- [ ] 2.3 Clasificar las acciones en "debe bloquearse" y "se valida después", partiendo de lo que hoy cubre `permissions` (el piso `deny`, `overlays.go:16-28`), y clasificar cada regla en orienta / restringe / verifica.
- [ ] 2.4 Armar la matriz cliente × superficie × herramienta × shell con el estado de cada celda (cubierta, no cubierta, no probada) en el formato de la tabla de cobertura del design D13.
- [ ] 2.5 Armar el inventario de escrituras de D13 (archivo, función de `filemerge`, nivel) y marcar las ALTO.
- [ ] 2.6 Verificar con el cliente real lo que la documentación no afirma y D6 deja (TBD): si un `PreToolUse` de Claude bloquea bajo `bypassPermissions`, el efecto de `onFailure` en `mcp_tool`/`prompt`/`agent`, el error y el timeout de un hook local de Codex (incluido `PreToolUse`; la documentación solo afirma la salida no soportada o inválida), el comportamiento de los demás eventos de Codex ante fallo o timeout, la tabla de códigos de salida por evento de Claude, y releer textual `disableAllHooks` y `allowManagedHooksOnly`.

## 3. Diseño del piloto (ALTO)

- [ ] 3.1 Definir la política de fallo por tipo de acción, partiendo de D6 (TBD del grupo).
- [ ] 3.2 Definir el contrato de entrada y salida de los handlers y la opción de handler a pilotear (D5).
- [ ] 3.3 Definir la capacidad opcional de hooks en los adapters y el mecanismo de ownership (D4).
- [ ] 3.4 Definir cómo se integran los controles a `update` y `uninstall`: remoción acotada del overlay de permisos y de las claves de hooks (hoy el paso de `permissions` de `uninstall` no hace nada, `internal/uninstall/steps.go:333`), y registro de cada archivo escrito con `Manifest.GuardRollback` (reutilizar, no reimplementar; D8). Definir la actualización desatendida (D14).
- [ ] 3.5 Definir cómo se mezclan listas y escalares del usuario y cómo se aborta ante un archivo ilegible (D9), y quién revisa el cambio en `filemerge` (D15).
- [ ] 3.6 Extender la vista de `install --dry-run` con el diff de permisos y hooks.
- [ ] 3.7 Definir el contenido del mensaje de control ausente y qué cuenta como protección equivalente verificada, alineado con 03 D11/D12 y con la regla de 04 (`specs/codex-opsx-delivery/spec.md:59-65`) sin redefinirlas, y la lista de controles obligatorios y opcionales (TBD, grupo).
- [ ] 3.8 Completar la tabla D14 (dueño, reversión, precedencia, Bypass, actualización desatendida, restauración) para cada entrada que se instale.

## 4. Pruebas previstas (ALTO)

- [ ] 4.1 Acción permitida y acción denegada en cada cliente. En Codex, con el método de 01 (D11).
- [ ] 4.2 Claude: instalar y actualizar permisos sobre un `settings.json` de prueba con la skill `/sdd-orchestrator` de 03 instalada y un `agent` fijado por el usuario; la clave `agent` y los hooks del usuario quedan idénticos (D12).
- [ ] 4.3 Timeout, JSON inválido y handler ausente, con el comportamiento real de cada cliente registrado y contrastado con la documentación de D6; en Codex el error y el timeout de un hook local son (TBD) y se miden, sin asumir que fallan abierto.
- [ ] 4.4 Rutas alternativas en shells reales por sistema operativo.
- [ ] 4.5 **Archivo ilegible:** con un `settings.json` malformado en un `HOME` temporal, la escritura aborta, el archivo queda byte-idéntico y se informa ruta y motivo (hoy `MergeJSONObjects` lo trata como `{}`, `json_merge.go:10-15`: la prueba debe fallar primero por esa razón). Lo mismo con un `config.toml` inválido.
- [ ] 4.6 Escritura con revisión, backup, fallo de backup, preservación de entradas ajenas (lista `allow` del usuario, `defaultMode` fijado por el usuario, hooks del usuario), reinstalación idempotente, `update`, `uninstall`, reversión y reversión con ediciones posteriores del usuario (el archivo actual se conserva).
- [ ] 4.7 **Control obligatorio:** (a) obligatorio presente, la acción procede; (b) obligatorio ausente, se detiene solo esa acción y el mensaje nombra la protección y cómo restablecerla; (c) opcional ausente, advertencia y la acción continúa.
- [ ] 4.8 **CRLF y modos:** medir en Windows y Linux la conversión de CRLF a LF de `UpsertTOMLTopLevel` (`toml.go:29`), el `chmod 0755` del directorio padre y el `0644` del archivo de `WriteFileAtomic` (`writer.go:40,62`) y la pérdida de comentarios de `MergeJSONObjects`.
- [ ] 4.9 Matriz de plataformas con Windows y Linux como mínimo: distribuidas (`.github/workflows/release.yml:71-76`), declaradas (`CLAUDE.md:21`) y probadas (lo que se corrió).
- [ ] 4.10 Pruebas por superficie, por separado.
- [ ] 4.11 Suite Go completa en verde y revisión de qué paquetes ALTO cambiaron (`git diff --stat`).

## 5. Documentación

- [ ] 5.1 Tabla de cobertura y límites por cliente, versión, shell y superficie, en el formato de D13.
- [ ] 5.2 Guía breve de qué queda garantizado por control nativo y qué solo por instrucción.
- [ ] 5.3 Glosario corto: hook del instalador vs. hook del cliente, y recibo de ejecución (06) vs. registro de instalación (04).

## 6. Integración con otras propuestas

- [ ] 6.1 Con 06: apuntar al contrato de 06 sin redefinirlo. Un control posterior valida de forma mecánica los campos del recibo de ejecución de 06 (D2: `corrida`, `arbol`, `comando`, `resultado`, `requisitos`, `salida`) y la integridad de la copia redactada; la procedencia (fuente independiente o reejecución) y la suficiencia son de 06 (D3). Los hooks complementan la evidencia pero no son un registro exhaustivo, y la ausencia de un rastro no verifica ni rechaza un recibo (06 D9 y requisito "Hooks complement evidence"). La contraparte en 06 es su tarea 5.3.
- [ ] 6.2 Con 01: casos de permisos efectivos compartidos y registro único de resultados (D11).
- [ ] 6.3 Con 03: punto de contacto en `WriteFileAtomic` (G6) y clave `agent` (D12); el mensaje de control ausente que 03 D12 le asigna a 08.
- [ ] 6.4 Con 04 (Matías): que su regla de controles y módulos faltantes (`specs/codex-opsx-delivery/spec.md:59-65`) use el mensaje de D7, y qué hooks de Codex entran al piloto.
- [ ] 6.5 Con 10: línea de base de cobertura para comparar con un juez en sombra, con la tabla de D13.

## 7. Cierre del change (flujo: entrega, comprobación, pendientes, traspaso, PR)

- [ ] 7.1 **Qué entrega.** Listar lo entregado (documentación, matriz, tabla de cobertura, código y pruebas si los hubo) con rutas. **Declarar alcance parcial** si algo depende de 06 o de 04 sin mergear, o de una decisión (TBD) abierta.
- [ ] 7.2 **Cómo se comprobó.** Para cada criterio de aceptación de la ficha (permitido, denegado, timeout, JSON inválido, handler ausente, rutas alternativas, shells reales, límites por superficie, revisión antes de escribir), registrar comando, resultado y plataforma (Windows y Linux). Comprobar los artefactos con `openspec status --change opsx-native-controls` y `openspec validate opsx-native-controls`.
- [ ] 7.3 **Qué queda pendiente.** Listar los `(TBD)` abiertos (política ante fallo, handlers, ownership, Bypass, hooks de Codex por versión, controles obligatorios, actualización desatendida, remoción acotada en `uninstall`, mínimo de cobertura) y las dependencias de 03, 04, 06 y 10.
- [ ] 7.4 **Informe de traspaso.** Escribir `openspec/changes/opsx-native-controls/informes/traspaso-<fecha>.md` con el modelo de `openspec/changes/codex-native-opsx/informes/traspaso-lifecycle-2026-11-13.md` (`origin/propose/codex-native-opsx`): qué podrá usar el siguiente responsable y coordinación pendiente por change (03, 04, 06, 10).
- [ ] 7.5 **Archivar y abrir el PR.** El PR se abre **después de archivar el change, con todas las tareas completas y verificadas** (o con el alcance parcial declarado en 7.1). El revisor e integrador es **Neyén**, que comprueba **"contrato y evidencia sobre el último commit"**; los hallazgos se resuelven en la misma rama. La rama real es `opsx-native-controls`; renombrarla a `propose/opsx-native-controls` se decide con Juan Cruz.
