# Pruebas de contexto efectivo con Claude real (tareas 1.x y 8.1)

Fecha: 2026-11-13. Cliente: Claude Code 2.1.318, Linux (Pop!_OS), terminal, modo `claude -p`,
modelo haiku. Proyecto temporal con target Project:

- `CLAUDE.md`: `TESTIGO-UNIV-U19` (testigo universal).
- `.claude/skills/sdd-orchestrator/SKILL.md`: frontmatter con `disable-model-invocation: true`,
  `active-stack-managed: true`, descripción con `testigo-desc-ZQ7`; cuerpo con `TESTIGO-COORD-K41`
  y referencia a `${CLAUDE_SKILL_DIR}/modules/tdd.md`.
- `.claude/skills/sdd-orchestrator/modules/tdd.md`: `TESTIGO-MOD-TDD-P88`.
- `.claude/skills/control/SKILL.md`: skill de control sin el flag, descripción `testigo-desc-CTRL-55`.

| # | Prueba | Oráculo | Resultado |
|---|---|---|---|
| 1 | Sesión normal sin invocar (1.10) | mensaje `init` de `stream-json` y respuesta del modelo | `init.skills` lista `control` y `sdd-orchestrator` (lista del harness); el modelo ve `testigo-desc-CTRL-55` (control positivo) y `TESTIGO-UNIV-U19`, y NO ve `testigo-desc-ZQ7` ni `TESTIGO-COORD-K41`. PASA |
| 2 | La clave `active-stack-managed` en el frontmatter | la skill carga y se invoca | sin error. PASA |
| 3 | `/sdd-orchestrator` activa el coordinador en el hilo | respuesta | devuelve `TESTIGO-COORD-K41`. PASA |
| 4 | `${CLAUDE_SKILL_DIR}` resuelve en terminal (1.8) | lectura del módulo | devuelve `TESTIGO-MOD-TDD-P88`. PASA |
| 5 | Subagente general lanzado por el coordinador (1.7) | respuesta textual del subagente | contiene `TESTIGO-UNIV-U19` y no `TESTIGO-COORD-K41`. PASA |
| 6 | Reanudación con `--resume` sin compactar (1.14) | respuesta | el cuerpo de la skill sigue en contexto. PASA |
| 7 | `/compact` y continuación (D10) | transcripción: `compact_boundary` seguido sólo del resumen, sin reenganche del cuerpo; el modelo declara no tener las instrucciones | NO se reengancha en 2.1.318 `-p`. Se aplica la rama "el sistema indica cómo recuperarlo": el bloque universal pide volver a invocar `/sdd-orchestrator` |

Versión mínima (1.4), según el CHANGELOG oficial (`anthropics/claude-code`): `${CLAUDE_SKILL_DIR}`
en 2.1.69; arreglo de `disable-model-invocation` al invocar `/<skill>` en 2.1.110. Mínima: 2.1.110.

No disponible, con motivo:
- IDE y Windows: no hay entorno en esta máquina.
- Precedencia Machine sobre Project (1.5): requiere escribir en `~/.claude/skills` real; queda la
  documentación oficial ("personal over project") y la advertencia del instalador (5.3).
- Precarga en el campo `skills` de un subagente (1.9): queda la documentación oficial.
- Explore y Plan sin `CLAUDE.md` (1.2): queda la documentación oficial.

## Instalador de punta a punta (binario de 4341660, home temporal con `--home`)

`CLAUDE.md` previo en 0600 con texto del usuario antes y después de un bloque legado
`active-stack:sdd-orchestrator`, y una skill propia `skills/mia/`.

| Paso | Resultado |
|---|---|
| `install --agent claude --mode custom --custom sdd-orchestrator` | bloque legado purgado; texto del usuario idéntico; bloque `universal-rules` una vez; skill con `SKILL.md` (`disable-model-invocation: true`, marca) y cinco módulos; `CLAUDE.md` sigue en 0600; aviso BREAKING en es; registro en `~/.active-stack/claude-migration.json` con hashes |
| segunda corrida de install | `diff -r` sin diferencias (idempotente) |
| `uninstall` targeted | borra la skill con marca, deja `skills/mia/`, quita el bloque universal, texto del usuario idéntico |
| `uninstall --strategy restore` con `CLAUDE.md` cambiado | no escribe, guarda el contenido actual en un snapshot, exit 1, nombra `--overwrite-edited` |
| `restore --overwrite-edited` | devuelve el `CLAUDE.md` previo con el bloque legado, después de guardar el contenido actual |

Observaciones fuera de esta ficha: en `--mode custom` el resumen de verify dice 0 chequeos; con el
binario de `35a89f8` sale igual (preexistente). El aviso BREAKING se muestra con la etiqueta
"degraded, best-effort" porque viaja por el canal de eventos degradados del pipeline.

## Alcance de la prueba

La prueba real se hizo con `claude -p`, modelo haiku y un `SKILL.md` escrito a mano con testigos,
no con el generado por el instalador. La prueba de punta a punta del instalador (sección anterior)
verificó los archivos escritos, no el contexto efectivo de Claude sobre ellos. `--continue` no se
probó: sólo `--resume`.

