# CHANGES — Roadmap de mejoras de Active Stack

> Roadmap activo de Active Stack. Antes de implementar una ficha, investigá su alcance y prepará la propuesta correspondiente. Consultá OpenSpec para el estado técnico de los changes.

[Vista interactiva](https://active-stack-roadmap.vercel.app/) · [Informe ejecutivo y evidencia](https://active-stack-executive-site.vercel.app/) · [Roadmap post-starters archivado](../CHANGES-2026-11-05-post-starters.md)

## Próxima etapa de Active Stack

| Ficha preliminar | Alcance | Depende de |
|---|---|---|
| F1 · Fundación y adaptación de arneses | Definir la ruta directa y la checklist de fundación, y aplicar el contrato en cada arnés de la fundación. Incluye selección, variantes y reanudación por agente y superficie. | — |
| F3 · ODD junto a SDD | Probar cuándo usar el flujo ligero y cuándo recurrir a OpenSpec. Identificar si la decisión vive en Active Orchestrator, el orquestador de entrega o ambos, y qué referencia breve necesita cada archivo de instrucciones. | — |
| F4 · Gestión de tareas y vista previa | **Integrada en main** ([PR #20](https://github.com/Group-Active-IA/active-stack/pull/20), e2ee756). ClickUp opcional por proyecto en Codex, configuración guiada, vista previa y confirmación por filas, reintento y reconciliación. Piloto en ClickUp y 16 pruebas locales aprobadas. **Release pendiente.** | F1 |
| F5 · Visualizaciones transversales | Habilitar diagramas y explicaciones visuales donde aporten, desde discovery hasta arquitectura, roadmap y entrega. | — |

Son cuatro fichas preliminares (F1 incorpora la anterior F2), con responsables y nombres de change (TBD). Al tomar cada ficha se investiga la solución técnica y se prepara su propuesta OpenSpec. El [documento de fundación](research/fundacion-active-stack-2026-11-17.md) conserva las decisiones, alternativas y criterios del piloto. El [historial de mejoras completadas](../CHANGES-2026-11-17-opsx-completados.md) reúne las fichas cerradas.


**Compatibilidad por arnés.** Cada propuesta revisa el catálogo y los adapters para identificar los agentes y superficies afectados. Incluir Codex, Claude, OpenCode y Antigravity en el relevamiento inicial de la fundación, y extender la matriz a los demás agentes según el soporte de cada arnés. Registrar qué funciona, qué alternativa existe y qué requiere ampliar soporte.

La [ronda anterior de OPSX](../CHANGES-2026-11-17-ronda-opsx.md) queda como referencia con sus fichas, acuerdos y evidencias.
