# Tasks

> Todas las tareas de implementación quedan **pendientes**: esta ronda es solo de propuesta. Este change es de gobernanza **ALTO**: antes de implementar hace falta la revisión humana que exige `CLAUDE.md`. Cuando llegue la etapa de implementación, TDD estricto, con fakes para el snapshotter y la lectura de archivos, y **nunca** el entorno real del usuario: RED, GREEN, TRIANGULATE y REFACTOR (esquema tomado de los changes archivados, no de una regla escrita en `CHANGES.md`). Antes de tocar archivos existentes, correr el SAFETY NET del paquete y anotar la línea base. No hacer build ni commitear salvo pedido explícito. Si surge la necesidad de tocar `filemerge`, `backup` o `pipeline`, frenar y reportar para revisión.

## 1. Antes de empezar: acuerdos y reproducciones (ALTO)

- [ ] 1.1 Confirmar que 06 definió el contrato de recibo que validarán los controles posteriores.
- [ ] 1.2 Alinear con 01 el método de prueba de permisos efectivos y la secuencia para Codex.
- [ ] 1.3 Cerrar en el grupo las preguntas abiertas del design y pasar por la revisión humana que exige el nivel ALTO.
- [ ] 1.4 Reproducir en un `HOME` temporal, con el binario y un `settings.json` de prueba, el caso del 02/11: `defaultMode` en `"auto"` y un permiso propio en `allow`. Registrar qué cambia al instalar con cada tier (sin tocar el entorno real). El resultado decide si D9 pasa a requisito de implementación y qué se arregla en `filemerge`.

## 2. Inventario (BAJO, solo lectura)

- [ ] 2.1 Listar los mecanismos de permisos y hooks de cada cliente y versión soportada, incluyendo si Codex tiene hooks.
- [ ] 2.2 Listar las superficies declaradas de cada cliente y los shells reales por sistema operativo.
- [ ] 2.3 Clasificar las acciones en "debe bloquearse" y "se valida después", partiendo de lo que hoy cubre `permissions` (lectura de `.env` y git destructivo), y clasificar cada regla en orienta / restringe / verifica.
- [ ] 2.4 Armar la matriz cliente × superficie × herramienta × shell con el estado de cobertura de cada celda (cubierta, no cubierta, no probada).

## 3. Diseño del piloto (ALTO)

- [ ] 3.1 Definir la política de fallo por tipo de acción.
- [ ] 3.2 Definir el contrato de entrada y salida de los handlers y la opción de handler a pilotear (D5).
- [ ] 3.3 Definir la capacidad opcional de hooks en los adapters y el mecanismo de ownership.
- [ ] 3.4 Definir cómo se integran los controles a `update` y `uninstall`, incluida la remoción de claves JSON.
- [ ] 3.5 Definir cómo se mezclan listas y escalares del usuario (D9) y quién revisa el cambio en `filemerge`.
- [ ] 3.6 Extender la vista de `install --dry-run` con el diff de permisos y hooks.

## 4. Pruebas previstas (ALTO)

- [ ] 4.1 Acción permitida y acción denegada en cada cliente.
- [ ] 4.2 Timeout, JSON inválido y handler ausente, con el comportamiento real de cada cliente registrado.
- [ ] 4.3 Rutas alternativas en shells reales por sistema operativo.
- [ ] 4.4 Escritura con revisión, backup, fallo de backup, preservación de entradas ajenas (lista `allow` del usuario, `defaultMode` fijado por el usuario, hooks del usuario), reinstalación idempotente, `update`, `uninstall` y reversión.
- [ ] 4.5 Pruebas por superficie, por separado.
- [ ] 4.6 Suite Go completa en verde y revisión de qué paquetes ALTO cambiaron (`git diff --stat`).

## 5. Documentación

- [ ] 5.1 Tabla de cobertura y límites por cliente, versión, shell y superficie.
- [ ] 5.2 Guía breve de qué queda garantizado por control nativo y qué solo por instrucción.
- [ ] 5.3 Glosario corto: hook del instalador vs. hook del cliente.

## 6. Integración con otras propuestas

- [ ] 6.1 Con 06: qué campos del recibo valida un control posterior.
- [ ] 6.2 Con 01: casos de permisos efectivos compartidos.
- [ ] 6.3 Con 10: línea de base de cobertura para comparar con un juez en sombra.

## 7. Cierre de la ronda de propuesta

- [ ] 7.1 Comprobar los artefactos con `openspec status --change opsx-native-controls`.
- [ ] 7.2 Compartir en el grupo el ID, el responsable, el enlace a la rama `propose/opsx-native-controls`, la revisión de partida, los acuerdos principales y las excepciones `(TBD)`.
- [ ] 7.3 Registrar el resultado de la revisión del grupo y compartir la versión revisada.
