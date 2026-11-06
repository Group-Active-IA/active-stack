# Proposal

## Why

El harness `design-mockups` ya cubre dibujar, aprobar y programar: hay una etapa de diseño entre `propose` y `apply`, y un gate de aprobación humana. Lo que falta es comprobar que lo construido se parezca a lo aprobado. Hoy `verify` trata los mockups como "criterios visuales de aceptación", pero ninguna instrucción le dice al agente cómo comprobarlos. Por eso puede dar una pantalla por buena sin haberla abierto. Este change cierra el ciclo de diseño con una verificación visual real (navegador automatizado) y, cuando no hay navegador disponible, con una degradación honesta.

## What Changes

- **Harness nuevo `playwright`** (`external`, `method: mcp`, solo `claude`, `install_modes: [full]`). Registra el servidor MCP oficial de Microsoft, Playwright MCP, como servidor stdio `playwright` en la config de MCP de Claude (`~/.claude.json`). El comando es `npx -y @playwright/mcp@<versión pinneada>`, el equivalente a `claude mcp add playwright npx @playwright/mcp@latest`, pero con la versión fija. Tiene descripción y descripción larga bilingües (es/en) y no tiene `depends_on`.
- **Paso de verificación visual dentro de la sección `active-stack:design-mockups`.** Se agrega como paso nuevo del fragmento `design-mockups.md`, con el mismo marker y sin tocar el registro de secciones. Aplica durante `verify` cuando el change tiene `## UI Mockups` con decisión `yes`/`provided` y `Approval: approved`, y además hay herramientas `playwright` disponibles. En ese caso el agente:
  - levanta el servidor de desarrollo;
  - abre la URL local;
  - captura móvil y escritorio;
  - revisa la consola y los requests fallidos;
  - compara contra `mockups/`;
  - reporta las diferencias priorizadas, con pasos de reproducción.

  Regla explícita: no afirmar que algo funciona si no se comprobó. Sin Playwright, el paso se degrada a un checklist manual y lo dice explícitamente. El resultado queda registrado como `Visual verify:` en `## UI Mockups`.
- **Recomendación suave, no dependencia dura.** `design-mockups` no depende de `playwright` en el catálogo. El fragmento recomienda instalar el harness `playwright` cuando falta. En modo Full ambos entran juntos; en Custom se eligen por separado.
- **Fix de alcance por agente en `internal/harness/external`.** `external.Install` saltea los adapters de agentes que el harness no soporta, igual que ya hace `internal/harness/config`. Sin este fix, si se instala `playwright` con `claude` y `opencode` seleccionados, `opencode` recibiría una entrada MCP remota rota (URL vacía). Los harnesses externos actuales no declaran `agents`, así que no cambian de comportamiento.
- **Docs**: README (tabla del catálogo y modo Full), ARCHITECTURE (ejemplo de catálogo), CHANGES.md (marcar `visual-verify` como implementado), y CLAUDE.md/AGENTS.md en espejo (mención del paso visual dentro de `design-mockups`). Se actualizan los conteos de catálogo en los tests (Full/Custom +1).

## Capabilities

### New Capabilities
- `playwright-mcp-harness`: el harness `external` `playwright` (catálogo, registro MCP stdio pinneado solo para Claude, alcance por agente, verificación post-install).

### Modified Capabilities
- `design-mockups-harness`: la sección inyectada suma el paso de verificación visual condicionado a Playwright MCP, con degradación a checklist manual, registro `Visual verify:` y recomendación suave del harness `playwright`. La descripción del catálogo menciona la verificación visual opcional.

## Impact

- **Código**:
  - `internal/catalog/harnesses.yaml`: entrada `playwright` y texto de `design-mockups`.
  - `internal/harness/config/assets/design-mockups.md`: paso 7.
  - `internal/harness/external/install.go`: filtro por `SupportsAgent`.
- **Tests**:
  - `internal/catalog/catalog_test.go`: conteos y resolución.
  - `internal/harness/external`: install de `playwright` y alcance por agente.
  - `internal/harness/config/design_mockups_fragment_test.go`.
  - `internal/verify`: checks de `playwright`.
- **No se tocan** `filemerge`, `backup` ni `pipeline` (governance ALTO): el registro MCP reutiliza el flujo existente de backup, merge JSON y escritura atómica.
- **Dependencia de terceros**: `@playwright/mcp` (Apache-2.0, Node ≥ 18), ejecutado vía `npx` en runtime. El instalador no instala navegadores.
- **Coordinación**: el change paralelo `design-skills-harnesses` también suma harnesses Full. Los conteos se ajustan por delta (ver design D7).
- **Usuarios existentes**: reinstalar `design-mockups` reescribe su sección (contenido nuevo) de forma idempotente. Desinstalarlo sigue siendo limpio.
