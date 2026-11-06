# Tasks

> Todas las tareas de implementación quedan **pendientes**: esta ronda es solo de propuesta. Cuando llegue la etapa de implementación, TDD estricto: en cada grupo, primero RED (test que falla por la razón correcta), después GREEN mínimo, TRIANGULATE y REFACTOR. Antes de tocar archivos existentes, correr el SAFETY NET del paquete y anotar la línea base. No hacer build ni commitear salvo pedido explícito. No tocar `filemerge`, `backup` ni `pipeline` (ALTO): si surge la necesidad, frenar y reportar. El esquema RED / GREEN / TRIANGULATE / REFACTOR sale de los changes archivados, no de una regla escrita en `CHANGES.md`.

## 1. Antes de empezar: acuerdos y confirmaciones (MEDIO)

- [ ] 1.1 Confirmar con 03 y 04 el inventario de módulos y el formato de retorno del coordinador y de los especialistas; fijar el texto exacto del extracto `## Evidence Contract`.
- [ ] 1.2 Cerrar en el grupo las preguntas abiertas del design (D5 ubicación del contrato, quién evalúa la suficiencia, criterio de recibo viejo, `verify` en Codex, agentes incluidos, política de salidas).
- [ ] 1.3 Confirmar que `git rev-parse`, `git status --porcelain` y `git hash-object --no-filters` se comportan igual en Windows, macOS, Linux, WSL y Termux (incluso con `core.autocrlf=true`).
- [ ] 1.4 Confirmar que el CLI de OpenSpec tolera la carpeta `evidence/` dentro de un change (`openspec status`, `openspec validate`).
- [ ] 1.5 Investigar por qué el reporte de `verify-hook` quedó con cero checks en el caso del 02/11 (v0.4.0, Windows, `lite`, `claude`) y reproducirlo sobre `main`. Registrar el resultado y pasarlo a 01, 04 y 08.
- [ ] 1.6 Coordinar con 05 cómo se conserva y se redacta la salida original al acotar salidas.

## 2. Contrato (MEDIO)

- [ ] 2.1 Redactar el esquema del recibo con un ejemplo válido y ejemplos inválidos.
- [ ] 2.2 Definir los tres estados y sus transiciones, incluida la regla "cero recibos no es verificado".
- [ ] 2.3 Redactar la regla de datos sensibles (D6): qué se versiona, qué se redacta, huella original y huella redactada, retención.
- [ ] 2.4 Redactar el extracto `## Evidence Contract` para el prompt de delegación.

## 3. Assets de Claude y de Codex (MEDIO, checkpoint)

- [ ] 3.1 SAFETY NET: correr `go test ./internal/harness/config/...` y anotar la línea base.
- [ ] 3.2 RED: golden de composición de Codex con todos los toggles (hoy solo existe `TestCompose_VariantCodex`, que busca la palabra "Codex"). Verificar que falla por la razón correcta.
- [ ] 3.3 RED: prueba de que el prompt de delegación incluye el contrato en Claude y en Codex.
- [ ] 3.4 GREEN: agregar el contrato al módulo de delegación y coordinación de `assets/claude/sdd-orchestrator.md` y de `assets/codex/sdd-orchestrator.md`.
- [ ] 3.5 GREEN: incluir el paso de verificación de recibos en el flujo de ambos, resolviendo la ausencia de `verify` en el asset de Codex junto con 04.
- [ ] 3.6 TRIANGULATE: actualizar los golden de Claude y de OpenCode si la composición los afecta, y comprobar que `generic` y los demás agentes no cambian.
- [ ] 3.7 REFACTOR y volver a correr la suite de `internal/harness/config` en verde.

## 4. Fixtures y pruebas de aceptación (MEDIO)

- [ ] 4.1 Fixtures de recibos: válido, incompleto, inventado, viejo, de directorio ajeno.
- [ ] 4.2 Fixtures: comando insuficiente, tests verdes con requisito incumplido, salida truncada sin original, cero recibos.
- [ ] 4.3 Caso de huella: mismo archivo con fin de línea LF y CRLF, debe dar la misma huella con `--no-filters`.
- [ ] 4.4 Caso de datos sensibles: una salida con un token y una ruta del home; comprobar que no se versiona y que se redacta antes de compartir.
- [ ] 4.5 Correr la misma batería con el flujo de Claude y con el de Codex y comparar los veredictos.
- [ ] 4.6 Suite Go completa en verde. Confirmar con `git diff --stat` que no se tocaron `backup`, `filemerge` ni `pipeline`.

## 5. Integración con otras propuestas

- [ ] 5.1 Con 05: cómo se conserva la salida original al acotar salidas.
- [ ] 5.2 Con 07: qué evidencia esperada puede exigir a un worker.
- [ ] 5.3 Con 08: qué campos del recibo puede validar un control posterior a la acción.

## 6. Cierre de la ronda de propuesta

- [ ] 6.1 Comprobar los artefactos con `openspec status --change opsx-verifiable-evidence`.
- [ ] 6.2 Compartir en el grupo el ID, el responsable, el enlace a la rama `propose/opsx-verifiable-evidence`, la revisión de partida, los acuerdos principales y las excepciones `(TBD)`.
- [ ] 6.3 Registrar el resultado de la revisión del grupo y compartir la versión revisada.
