# design-system

**Fundación visual de un proyecto** y elección persistente sobre los mockups de cada cambio de frontend.

> Una decisión de no usar mockups se guarda una vez y se respeta en los cambios siguientes.

## ¿Qué hace?

1. Pregunta si el proyecto quiere sistema visual con mockups, solo sistema visual o ninguna de esas etapas.
2. Si corresponde, audita o crea `design-system/` con marca, tokens, componentes y vistas HTML/CSS. Muestra las vistas y espera aprobación.
3. Registra la política y la revisión en `.active-orchestrator-state.json` cuando existe, preservando estados de fundaciones anteriores.

## Instalación

Active Stack incluye el `SKILL.md` en su binario. El modo Full lo instala para Claude, Codex, OpenCode y Antigravity; en Custom se instala al seleccionar `design-mockups`.

```bash
active-stack install --mode full
```

No requiere cuentas externas. Playwright MCP permite capturas automáticas, pero el HTML local puede revisarse manualmente.

## Uso

Durante la fundación, Active Orchestrator invoca la skill después de la KB. En un proyecto fundado anteriormente, el primer cambio visual la invoca si todavía no existe una política visual.

También se puede solicitar directamente: «Definamos el sistema visual de este proyecto» o «Quiero dejar de pedir mockups en cada change».

## Estructura

```text
design-system/
├── SKILL.md
├── README.md
└── LICENSE
```

El instalador entrega `SKILL.md` en el directorio de skills de cada agente. README y licencia documentan el archivo fuente incluido en Active Stack.

## Por qué esta estructura

La lógica es una skill separada para que Active Orchestrator coordine sin generar diseños por sí mismo. El mismo archivo instalado guía a los cuatro agentes y funciona con HTML/CSS local.

## Licencia

Apache-2.0.
