# Prueba real de estabilización OPSX

La ejecución interactiva de Claude queda para la validación final del operador. Si Claude Code informa que venció la sesión, autenticar con `claude auth login` antes de abrir el proyecto descartable. En Codex CLI, confirmar primero que la sesión puede ejecutar `openspec list --json` con su herramienta de terminal.

Ejecutar en un proyecto descartable de cada cliente, con la versión candidata y una sesión autenticada. Registrar versión, sistema y superficie. Repetir una vez con instalación limpia y otra tras actualizar una instalación anterior con una nota propia en el archivo de instrucciones.

1. **Migración.** Confirmar que la nota sigue intacta, que hay un único bloque gestionado y que el módulo de evidencia está instalado. Abrir una conversación nueva en el proyecto.
2. **Tarea cotidiana.** Pedir que corrija un typo en un archivo concreto. Esperar la corrección y la comprobación correspondiente. La sesión no debe recomendar OPSX ni crear un change.
3. **Cambio complejo.** Pedir un cambio con una decisión de diseño abierta, impacto en varias partes y criterios de aceptación. Esperar una recomendación de propuesta con sus artefactos y una invocación válida para el cliente instalado. Invocar explícitamente el coordinador: `/sdd-orchestrator` en Claude y `$sdd-orchestrator` en Codex.
4. **Verify directo.** En un change pequeño que admite revisión directa, ejecutar `verify` con delegación desactivada. Esperar trazabilidad de requisito, estado probado, comando y resultado, sin abrir un subagente. Una captura sin corroboración debe quedar sin verificar.
5. **Verify independiente.** Marcar un requisito de revisión independiente. Esperar una revisión por otra persona o agente. Si no está disponible, esperar que el reporte deje pendiente esa revisión. Verificar que Codex no use aliases de Claude.
6. **Reanudación.** Compactar durante una acción OPSX y volver a invocar el coordinador. Confirmar que retoma el mismo change desde OpenSpec y relee los módulos de la acción. Repetir una tarea cotidiana en una sesión nueva.
7. **Control negativo.** Modificar un archivo después de obtener un recibo y verificar de nuevo. Esperar que el recibo anterior quede rechazado para el nuevo estado. Una copia redactada con datos insuficientes debe quedar sin verificar.

Guardar versión, superficie, acción, resultado observado y referencias a evidencia redactada. Omitir credenciales y salidas sin redactar.

## Preparación de la versión candidata

Usar el ejecutable construido desde este worktree y comprobar que `active-stack evidence-capture --help` muestre la sintaxis. La última release pública puede entregar otro paquete. Mantener la prueba en un proyecto descartable y verificar las rutas efectivas de las skills que carga cada cliente.

El catálogo candidato instala las revisiones fijadas de Active Orchestrator (d862278), Roadmap Generator (dd45758) y Agent Instruction (a873714). Confirmar los nombres y la invocación de cada skill desde la sesión del cliente.

## Lectura de los resultados

- En Codex, dos versiones reconocieron la skill generada y fallaron al iniciar la herramienta de terminal dentro del runtime de escritorio. Para la prueba de verify usar una sesión capaz de ejecutar `openspec list --json`.
- La prueba de symlinks requiere una cuenta Windows que pueda crearlos. Si el proceso devuelve el error 1314, registrar que no se probó ese escenario.
- `go test ./...` y la validación estricta cubren el repositorio. La tabla de resultados de esta guía queda para las pruebas interactivas por cliente y superficie.

## Resultado de la sesión real de Codex

- Windows, Codex CLI 0.166.0. La terminal ejecutó OpenSpec y el Active Stack candidato. El aviso del hook no impidió los comandos.
- Verify directo: el recibo ebb7783e6373c246105078b2f5b83e67 registró salida 0. Coincidieron el árbol y la huella de la copia. Codex corrigió un veredicto inicial prematuro y dejó R1 sin verificar por falta de corroboración independiente.
- Reanudación: después de /compact recuperó el mismo change, recibo y veredicto, y volvió a leer los módulos.
- Verify delegado: se activó el módulo de delegación en el proyecto descartable y se capturó el recibo 2b84df45743080d4830d6ce72aebcbb6. El subagente review_r1 terminó con status completed y modified vacío. Confirmó integridad y suficiencia. R1 quedó sin verificar por falta de corroboración independiente de la ejecución original.
- La prueba interactiva se limitó a Codex por indicación del usuario.
- Reparación de Engram en Windows: el instalador comprobó el reemplazo y la reutilización con una prueba de regresión. Engram 3.7.2 ejecutó los hooks nativos de Codex y Claude con salida JSON y código 0 en una carpeta temporal. Queda pendiente confirmar la desaparición del aviso en una sesión nueva después de reinstalar la copia global.
