# Continuidad de Active Stack · F1 y F3

Trabajá primero F1 y después F3. Usá este flujo experimental sin SDD ni artefactos de OpenSpec. El roadmap define qué funcionalidad explorar y el chat documenta las decisiones de implementación. F3 puede investigar ODD y SDD aunque se desarrolle con este flujo.

## Ciclo por funcionalidad

1. **Leé e investigá la ficha.** Consultá AGENTS.md, CHANGES.md y research/fundacion-active-stack-2026-11-17.md (dentro de docs/). Revisá Git y recuperá decisiones pertinentes de Engram. Delegá el relevamiento a subagentes con encargos acotados. Identificá los componentes, arneses, agentes y superficies afectados.
2. **Definila conmigo en el chat.** Haceme las preguntas necesarias sobre comportamiento, alcance, decisiones técnicas y criterios de aceptación. Presentá opciones y consecuencias cuando haya una decisión real. Esperá mis respuestas y acordá el resultado verificable antes de empezar a programar.
3. **Fijá un objetivo por ficha.** Cuando quede definida, usá la función nativa Objetivo o Goal del cliente para completar esa funcionalidad, con el alcance y los criterios acordados. Implementá con subagentes y TDD. Coordiná dependencias, aplicá las reglas del repo y verificá el resultado integrado.
4. **Cerrá y compactá.** Cuando la funcionalidad esté implementada y verificada, cerrá el objetivo y guardá en Engram un resumen breve de las decisiones, archivos, pruebas y pendientes. Compactá el contexto si el cliente ofrece esa acción. Si requiere intervención manual, dejame el resumen y pausá para que pueda compactar.
5. **Pasá a la siguiente ficha.** Tras compactar F1, recuperá el resumen y repetí el ciclo desde la investigación y las preguntas de F3. F4 y F5 quedan para otra sesión.

## Alcance inicial

- **F1** reúne el contrato y la adaptación de los arneses de fundación. Investigá la ruta directa, el mínimo obligatorio, la checklist previa, las variantes y la reanudación. Revisá compatibilidad por arnés, agente y superficie, empezando por Codex, Claude, OpenCode y Antigravity.
- **F3** investiga cuándo usar ODD o SDD, dónde vive esa decisión y cómo conservar decisiones y evidencia al cambiar de flujo. Revisá Active Orchestrator, el flujo de entrega y las referencias breves en instrucciones de agentes.
