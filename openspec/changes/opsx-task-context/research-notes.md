# Notas de investigación — ronda de propose, ficha 05

> Análisis previo al propose, hecho el 2026-11-05/06 sobre `97d7f6c`. El roadmap se publicó en
> `CHANGES.md` en `bd63543` y la rama está al día con `1b4639b`; ninguno de los dos tocó código ni
> assets. No es un artefacto de OpenSpec: es el registro de por qué la propuesta tiene la forma
> que tiene, y qué hallazgos hay que llevar al grupo. Referenciado desde el `README.md` de este change.

## 1. Las cuatro fuentes y cómo se encadenan

| Fuente | Fecha | Qué es | Unidad de diseño |
|---|---|---|---|
| PDF `DOC-20261103-WA0079` (34 pp) | 03.11 | Plan de mejora hacia gobernanza verificada | **7 piezas**, 4 niveles, 3 olas, escenarios S-01→S-14 |
| PDF `DOC-20261104-WA0031` (5 pp) | 04.11 | Arquitectura de bajo consumo de tokens y baja alucinación | **6 capas**, 7 reglas anti-alucinación, S-15→S-21 |
| `active-stack-executive-site.vercel.app` | 04.11 | Decision brief: auditoría crítica del PDF del 04.11 | **15 afirmaciones** con veredicto + 12 observaciones de código |
| `active-stack-roadmap.vercel.app` | 05.11 | Plan operativo del equipo | **10 fichas**, 7 responsables, P0–P3 |

La cadena es un embudo deliberado: el 03.11 maximiza ambición → el 04.11 relee lo mismo en
clave de rendimiento → el brief corta lo no probado → el roadmap convierte lo que sobrevive
en fichas con compuerta de medición. **Como proceso es coherente y disciplinado.** Los
problemas son de inventario y de alcance, no de dirección.

## 2. Lo que se propagó bien por la cadena

- **El fallo de compatibilidad.** `approval_policy = "untrusted"` / Codex CLI 0.166.0 aparece
  idéntico como observación #10 del brief (`permissions/install.go:122`) y como P0 en la ficha 01.
- **Medición antes de adopción.** Ausente en el 03.11, introducida en el 04.11 (`bench` se
  construye primero), ascendida por el brief a la única afirmación con *adoptar* limpio,
  convertida por el roadmap en ficha 02 y dependencia de todo. Se fortalece en cada paso.
- **Recibos → evidencia.** La mejor corrección de la cadena. El 04.11 afirma que el esquema JSON
  elimina la alucinación de éxito; el brief lo rechaza como garantía ("validación de estructura
  ≠ verdad semántica"); la ficha 06 reescribe el problema en los términos exactos del brief.
- **Jev se desescala de forma monótona.** Pieza 7 completa → capa 6 en sombra que solo endurece
  → "sin prueba local" → en el roadmap sobrevive solo como candidato opcional de comparación
  en 09 y 10, P2/P3. Nunca gana autoridad.

## 3. Lo que quedó huérfano (llevar al grupo)

### 3.1 Las siete piezas del plan del 03.11 no están en el roadmap

Cero menciones a *policy-as-code*, *explain*, *lockfile*, *manifiesto*, *conformance* en la
página del roadmap. Y el 04.11 declara pararse sobre ellas: *"Las piezas 1, 3, 4 y 6 de ese plan
son la capa 1 de esta; la Pieza 7 es la capa 6; conformance es cómo se prueban las siete
reglas."* El brief lo detectó (afirmación #13, *dependencia no verificada*), y el roadmap lo
resolvió **dejando caer la dependencia en silencio**, sin decisión escrita.

Consecuencias concretas:
- La ficha 08 pide *"armá una matriz por cliente, superficie, herramienta y shell"*. Eso **es**
  policy-as-code (clases de operación × tabla de traducción) re-derivado con otro nombre.
- Los **21 escenarios de conformance** (S-01→S-14 del Anexo B + S-15→S-21 del 04.11) son un
  artefacto numerado y diseñado, sin casa en el roadmap. El criterio de cierre de la 08
  ("permitido, denegado, timeout, JSON inválido, handler ausente, rutas alternativas, shells")
  es un subconjunto reinventado de esa suite. **S-01 a S-14 están listos para pegarse ahí.**

### 3.2 Las diez adopciones de Gentle-AI se evaporaron

El 03.11 §7 compromete 10 adopciones con 9 changes listados: piso de deny ampliado + S-13/S-14,
versiones centralizadas con Renovate, firma Minisign, AI_POLICY.md, **servidor de modelo
guionado**, ledger por hash, `doctor`, journal de mutaciones, guía de estilo de skills con
presupuesto. Ninguna aparece en las 10 fichas.

La más dolorosa es el **servidor de modelo guionado**, que el propio 03.11 llama "la adopción
más valiosa": agente real, binario real, FS real, y el modelo como servidor HTTP local con
respuestas guionadas. Corre en CI sin API keys. La ficha 08 necesita exactamente eso y dice
"shells reales" sin especificar cómo se maneja el modelo.

La **guía de estilo de skills con presupuesto** (estructura obligatoria activación / reglas
duras / gates / pasos / contrato de salida; cuerpo 180–450 tokens, máx. 1.000) cae de lleno en
el alcance de la ficha 05. Esta ficha reclama solo la medición: publica el tamaño de cada cuerpo
de skill y declara el tope fuera de su alcance (`design.md` — D12).

### 3.3 El alcance de agentes cambió y se llevó la tesis del producto

| Fuente | Agentes |
|---|---|
| 03.11 | diagnostica **9**, soporta 4, elige 3 para conformance (Claude, Codex, OpenCode) |
| 04.11 | Claude Code y OpenCode |
| brief + roadmap | **Codex y Claude, nada más** |

OpenCode desapareció sin declaración. El hallazgo central del 03.11 —*"la misma elección del
usuario produce seis realidades distintas según el agente, y ninguna parte del sistema lo
informa"*— **no se puede probar con dos clientes**, y sobre él se apoyan las frases de
posicionamiento ("la única herramienta del mercado", "ningún competidor puede copiar esto en un
fin de semana"). Un roadmap de dos clientes cambia la tesis de "gobernanza verificada entre
agentes" a "dos clientes bien configurados". Es un cambio de producto, no de cronograma.

### 3.4 Las siete reglas anti-alucinación perdieron su mecanismo

El brief desarmó tres de las siete y el roadmap conserva la intención en 06 y 08, pero las
reglas ya no están numeradas ni son rastreables:

| Regla | Mecanismo declarado en el 04.11 | Veredicto del brief |
|---|---|---|
| S-16 sin recibo no hay éxito | PostToolUse compara salida vs recibos | **matizar** — observa *después* de ejecutar: no previene |
| S-20 sin carga no hay cita | el binario mantiene la lista de lo cargado | **no demostrado** — requiere instrumentación universal |
| S-18 sin ownership no hay escritura | hook de gobernanza + ledger | **matizar** |

Falta una tabla de 7 filas × {mecanismo pretendido, veredicto, ficha que lo toma o descartado}.
Es el arreglo más barato para que la cadena informe→roadmap quede auditable.

### 3.5 Dos veredictos del brief son injustos, por una sola causa

- #5 *"Zhang et al. demuestra que las reglas negativas son las más efectivas"* → quedó **no
  verificable** por falta de datos de cita. Pero el Anexo C del 03.11 trae la referencia
  completa: *Guardrails Beat Guidance*, arXiv abril 2026, 679 archivos, 25.532 reglas, +5.000
  ejecuciones.
- #13 → **dependencia no verificada** porque "el plan previo no viene adjunto".

**Las dos fallan porque el 04.11 se revisó sin el 03.11 adjunto.** Una sola acción las arregla:
adjuntar el plan del 03.11 con su Anexo C y reemitir esos dos veredictos. Importa porque el
diseño del kernel de reglas negativas se apoya en esa cita.

## 4. Análisis interno del roadmap

### 4.1 Lo que está bien hecho

Los problemas están escritos como fallos observables, no como soluciones disfrazadas (la 08:
*"que un hook esté instalado no prueba que se ejecute"*). La separación *qué planificás solo* /
*qué se cierra en grupo* existe en las 10. Nada se adopta sin medición (*"el ahorro queda sujeto
a medición"*). Los *Límites compartidos* (`CHANGES.md:49`) cortan el scope creep de antemano: un runtime nuevo
requiere una decisión de arquitectura separada y el arnés de evaluación queda fuera del binario. El P0 es un hecho
reproducido y versionado. El nivel de riesgo tiene consecuencia procedimental.

### 4.2 Defectos estructurales, por daño

**D1 · Dependencia circular 02 ↔ 03/04.** Las fichas 03 y 04 declaran depender de 02. Pero el
criterio de cierre de 02 dice *"comparar ejecuciones equivalentes del flujo OPSX anterior y del
**flujo modular**"*, y el flujo modular no existe hasta que 03/04 estén implementadas. En
propose no explota; en implementación sí. Y al revés: 03/04 no necesitan nada de 02 para
planificarse. **Arreglo:** partir 02 en *baseline del flujo actual* (sin dependencias) y
*comparación modular* (depende de 03/04).

**D2 · Falta la arista 01 → 08.** La 08 declara depender de 06, pero 06 es confianza en el
recibo y 08 es enforcement: ejes distintos. El acoplamiento real es con **01**: ambas tocan
`permissions/install.go`, y 01 decide la política de tier que 08 debe verificar. Está declarado
solo como "coordiná con 01". Riesgo: 08 se planifica contra una política que 01 va a cambiar.

**D3 · 01 → 02 es blanda presentada como dura.** 02 depende de 01 solo para el tier estricto de
Codex; el baseline se puede medir en Claude o en Codex balanceado. Tal como está, **01 es la
única puerta de las 10 fichas**: P0, riesgo alto, un responsable, y con un TBD abierto en su
propio criterio de cierre.

**D4 · Los tres dobles-responsables tienen fichas encadenadas entre sí.**

| Responsable | Fichas | Relación |
|---|---|---|
| Gianfranco | 06, 08 | 08 depende de 06 → serie para una persona |
| Neyen | 07, 10 | 07 → (09, de Matías) → 10 → serie con intermedio ajeno |
| Matías | 02, 09 | los dos extremos del camino crítico |

Ninguno es paralelizable. Repartir 06+09 y 07+08 rompe las cadenas.

**D5 · La mitad de los criterios de cierre son de ejecución, no de propose.** La guía es
taxativa (*"solo propose"*, *"las tareas quedan pendientes"*), pero 02, 05, 07, 09 y 10 cierran
con criterios de correr el experimento. **Solo la 02 resuelve la ambigüedad internamente**: *"la
propuesta entrega un protocolo y una estructura de resultados; las corridas se planifican como
tareas"*. Copiar esa frase a 05, 07, 09 y 10 es el arreglo más barato y más útil del roadmap.
(En esta ficha ya está aplicada, en el encabezado de `tasks.md` y en `proposal.md` — What Changes.)

**D6 · La regla de TBD se contradice con dos fichas.** La guía dice que un TBD que cambie
alcance, diseño o aceptación se resuelve antes de revisión. Pero el TBD **es** la aceptación en
01 (*"config aceptada por las versiones soportadas (TBD)"*) y en 09 (*"acordar umbrales y
muestra (TBD)"*). Y 01 es el P0.

### 4.3 Mapa de contención de archivos (el roadmap no lo muestra)

| Artefacto | Fichas que lo tocan | Dueño declarado |
|---|---|---|
| `assets/{claude,codex}/sdd-orchestrator.md` | **02, 03, 04, 05, 07, 09** | ninguno |
| `internal/catalog/harnesses.yaml` | 03, 04, 05, 07 | ninguno |
| `ARCHITECTURE.md` | 02, 06, 09, 10 | ninguno |
| `internal/verify/report.go` | 02, 06, 10 | ninguno |
| `permissions/install.go` | 01, 08 | ninguno |

Seis fichas y cinco personas sobre `sdd-orchestrator.md`: 03/04 lo reestructuran, 05 define qué
se carga por tarea, 07 define las reglas de delegación de adentro, 09 lee el routing, 02 mide su
tamaño. **Propuesta para el grupo:** dueño = 03+04 en conjunto; las demás escriben requisitos.

### 4.4 Huecos de cobertura del roadmap

1. **Nadie es dueño de la migración combinada.** 01, 03 y 04 traen cada una su snapshot y
   rollback, pero el usuario con el bloque global instalado va a ser tocado por las tres.
2. **No hay ficha de documentación ni comunicación del cambio**, y el roadmap cambia cómo se
   activa el orquestador para todo usuario existente.
3. **Las "superficies" nunca se enumeran.** La aceptación transversal dice *"en cada superficie
   declarada"*, pero la lista la produce una ficha (03/04). Hoy no se sabe contra qué se acepta.
4. **Los agentes que no son Claude ni Codex quedan en silencio.** Si el instalador sigue
   escribiendo para OpenCode o Antigravity, ninguna ficha pregunta qué les pasa al retirar el
   bloque global.
5. **Etiqueta:** la ruta dice `05 06 · Contratos`, pero la 05 no es un contrato (es lectura
   dirigida); la del contrato es la 06. Y la ruta termina en 07, omitiendo la etapa 04.

## 5. Decisiones que estos hallazgos produjeron en esta ficha

| Hallazgo | Qué cambió en el propose |
|---|---|
| `CHANGES.md:49` deja un runtime nuevo sujeto a decisión de arquitectura separada + obs. #9 del brief | el resumen de salida se especifica como **contrato de formato**, no como subcomando; el productor es el agente bajo instrucción (`design.md` — D8) y quién ejecuta cada regla lo fija D17 |
| 1.500 tokens / 40 líneas = presupuestos no validados (brief) | la frontera núcleo/módulo se define **por consecuencia** (D2 de `design.md`), no por tamaño |
| `sdd-orchestrator.md` sin dueño, 6 fichas | esta ficha escribe **requisitos** sobre el archivo, no contenido |
| D5 (criterios de ejecución en ronda de propose) | la frase de alcance de la 02 aplicada al encabezado de `tasks.md` |
| Guía de estilo de skills huérfana (§3.2) | inventario de cuerpos de skill en tareas 1.5–1.7; el tope queda fuera de alcance (`design.md` — D12) |
| La 02 mide tamaño de texto, no consumo | todos los números de esta ficha etiquetados como **tamaño de texto** |

## 6. Contraste con la documentación oficial de los clientes (10/11/2026)

Fuente: documentación de Claude Code (`https://code.claude.com/docs/en/`) y de Codex (`https://learn.chatgpt.com/docs/`), leída el 10/11/2026 sobre `6a45233`. Algunas páginas superan el límite de lectura y se leyeron en parte (sub-agents, skills, env-vars, mcp, config-reference); lo que cae fuera de lo leído queda **no verificable**. El código fuente de Codex y el issue `openai/codex#23454` son evidencia, no documentación, y se marcan así.

| Hallazgo | Fuente | Qué cambió en el propose |
|---|---|---|
| El cliente recorta la salida antes de que el modelo la vea: Claude, un resultado fallido de más de ~10.000 caracteres llega como cabeza y cola *"with no file path"*; Codex recorta el medio | https://code.claude.com/docs/en/tools-reference; https://learn.chatgpt.com/docs/config-file/config-reference (`tool_output_token_limit`) | captura a archivo antes de leer (D8); umbral de salidas en dos unidades, aparte del de lectura (D3, D19); resumen "parcial — sin verificar" sin captura; tareas 2.15–2.17 y 2.19 |
| Los clientes guardan su propia copia de la salida | https://code.claude.com/docs/en/tools-reference | la garantía de redacción se acota a la copia de Active Stack; el crudo de la captura se borra (D9) |
| Stderr llega mezclado con stdout en Claude; en Codex no está documentado | https://code.claude.com/docs/en/monitoring-usage | fixture de fallo solo por stderr, corrido en los dos clientes (D4, tarea 2.15) |
| Los tres niveles de "no habilitada" existen, con mecanismos distintos por cliente; en Codex el archivo de un rol no fija `mcp_servers`; `allowed-tools` no restringe | https://code.claude.com/docs/en/permissions; https://code.claude.com/docs/en/skills; https://learn.chatgpt.com/docs/agent-configuration/subagents | nivel de control declarado por cliente (D18, tarea 5.9) |
| La invocación explícita difiere: Claude `/nombre`; Codex `$nombre` en CLI e IDE, `@` en ChatGPT | https://code.claude.com/docs/en/skills; https://learn.chatgpt.com/docs/build-skills | referencia de invocación por cliente en el núcleo (D21, tarea 3.11) |
| Explore y Plan no cargan `CLAUDE.md`; tampoco un agente con `omitClaudeMd` | https://code.claude.com/docs/en/sub-agents | las reglas obligatorias viajan en el encargo de esos especialistas (D21, tarea 3.7) |
| Codex tiene un piso nativo de lectura (`permissions.<name>.filesystem` con `"deny"`) | https://learn.chatgpt.com/docs/config-file/config-reference | se cita como mecanismo concreto en el pedido a 01/08 (D14) |
| Codex no documenta subagentes en la nube | https://learn.chatgpt.com/docs/agent-configuration/subagents | `verify` queda detenida en una superficie sin subagentes, como límite declarado (D19) |
| Una lectura de Claude admite hasta 25.000 tokens y 256 KB; Read y Grep aceptan rangos | https://code.claude.com/docs/en/tools-reference | el umbral de 300 líneas de lectura se mantiene como valor inicial medido; se agrega la excepción justificada (D11, D15) |

**Queda no verificable con la documentación** (y por eso se declara, no se decide): cómo combina Codex stdout y stderr; el valor por defecto de `tool_output_token_limit`; dónde persiste Codex las salidas; si un subagente de Codex lee `AGENTS.md`; si `allow_implicit_invocation: false` saca la descripción del contexto; el estado beta o estable de los perfiles de permisos de Codex; cuánto contexto pesa una herramienta MCP diferida en Claude.
