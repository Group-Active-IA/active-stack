# Plan de aceptación de clientes del cambio 04

Preparado el 13/11/2026 para las tareas 1.2, 2.4 y 4.2. Este documento es un protocolo pendiente de ejecución: no acredita aceptación funcional de ninguna de las cuatro celdas. No implementa 03 ni modifica configuración del usuario.

## Matriz de ejecución

| Celda | Entorno que debe probarse | Versión mínima aceptada | Estado inicial |
|---|---|---|---|
| W-CLI | Windows nativo, Codex CLI | (TBD), fijar mediante P2 | No probado |
| L-CLI | Linux nativo, Codex CLI | (TBD), fijar mediante P2 | No probado |
| W-IDE | Windows nativo, extensión Codex en IDE identificado | (TBD), registrar IDE y extensión | No probado |
| L-IDE | Linux nativo, extensión Codex en IDE identificado | (TBD), registrar IDE y extensión | No probado |

Una terminal integrada que ejecuta `codex` pertenece a CLI. WSL no acredita Windows nativo. Un cliente remoto o un contenedor debe identificarse como tal y no sustituye otra celda por inferencia. La aplicación de escritorio y cloud quedan fuera de la aceptación inicial de 04.

Inventario local observado al preparar el plan: `codex --version` devolvió `codex-cli 0.167.0-alpha.3.2` en macOS Intel. No se inició ningún turno de modelo. La sonda existente de App Server acredita descubrimiento del paquete, no activación, permisos, especialistas ni continuidad. No se extrapola a las celdas anteriores.

## Preparación de cada celda

1. Usar una VM desechable o una cuenta nueva del sistema operativo dedicada a esta prueba. En IDE, crear también un perfil nuevo del editor y abrirlo desde esa cuenta. El perfil del editor por sí solo no aísla el directorio de usuario ni las skills de máquina.
2. Registrar usuario de prueba, sistema/arquitectura, tipo de instalación, versión de Codex, versión del IDE/extensión, ejecutable usado, autenticación/modalidad y permisos efectivos. No registrar tokens, credenciales ni datos privados.
3. Recibir un binario de Active Stack correspondiente al SHA exacto bajo revisión. Registrar su procedencia y huella. No usar un binario viejo para aceptar código nuevo ni ejecutar un build como parte de este plan sin autorización explícita.
4. Fijar las revisiones y ubicación de OpenSpec y de las skills auxiliares realmente disponibles. `openspec status` completo acredita artefactos de planificación, no tareas implementadas.
5. Crear un repositorio sintético propio de la celda, sin parent `AGENTS.md` ni configuración Project legada. Elegir una ruta con espacios. Incorporar una tarea pequeña y una prueba local determinista, con un defecto preparado que no toque datos reales. Conservar el commit inicial y el diff de cada caso.
6. Instalar únicamente en el usuario de prueba. El instalador permite `--home`, pero ese flag no dirige por sí solo la lectura del cliente Codex: cliente e instalador deben usar realmente el mismo perfil aislado. No resolver el aislamiento cambiando el `HOME` real de la sesión de trabajo habitual.
7. Revisar primero el plan de instalación. Los flags siguientes están presentes en `cmd/active-stack/headless/flags.go`; ejecutar con el binario recibido, desde la cuenta de prueba y después de confirmar que su ayuda coincide. El plan puede añadir componentes obligatorios: registrar el plan efectivo antes de ejecutarlo.

Linux, shell de la cuenta desechable:

```sh
codex --version
openspec --version
active-stack install --headless --mode custom --custom sdd-orchestrator --agent codex --home "$HOME" --no-self-install --dry-run
```

Windows, PowerShell de la cuenta desechable:

```powershell
codex --version
openspec --version
active-stack install --headless --mode custom --custom sdd-orchestrator --agent codex --home "$env:USERPROFILE" --no-self-install --dry-run
```

Una vez revisado ese plan y dentro del perfil aislado, ejecutar el mismo comando sin `--dry-run` para instalar. No añadir `--yes` ni un tier nuevo como sustituto de la revisión de permisos. Conservar stdout/stderr redactados y estado antes/después. Estos comandos se verificaron contra el parser del repositorio; no se ejecutaron en Windows/Linux durante la preparación del documento.

Las rutas Machine actuales derivan del adapter Codex: `.codex/AGENTS.md` y `.agents/skills/sdd-orchestrator/` bajo el usuario de prueba. Revisar el plan y el adapter de la revisión probada en vez de tratarlas como constantes universales. Conservar la identidad del snapshot y del registro de instalación.

## Casos funcionales

Ejecutar cada caso en cada celda. Reiniciar una sesión limpia cuando corresponda para evitar que una invocación explícita previa contamine el caso de no activación. Guardar el texto exacto de cada pedido y la evidencia disponible del cliente. Las respuestas del propio agente no acreditan por sí solas qué archivos cargó, permisos tuvo o comandos ejecutó.

| ID | Pasos manuales | Resultado esperado y evidencia necesaria | Dependencia o límite |
|---|---|---|---|
| C01 · Normal | Abrir una sesión nueva en el fixture. Pedir una lectura sencilla del archivo preparado, sin mencionar OPSX ni invocar skill. | Sigue disponible la regla universal de prueba; no inicia coordinación, delegación ni ritual OPSX. No aparece el antiguo bloque global de coordinador. Revisar archivos instalados y eventos de carga disponibles. | Requiere perfil sin legado Project. Registrar señales no observables. |
| C02 · Pedido implícito | En otra sesión nueva, pedir «proponé un change para describir el fixture», sin escribir `$sdd-orchestrator`. | La skill coordinadora no se selecciona implícitamente. Una acción ordinaria que el cliente pueda hacer no se confunde con activación del coordinador. La policy instalada declara `allow_implicit_invocation: false`. | La policy estructural es necesaria pero no demuestra el comportamiento del cliente. |
| C03 · Invocación explícita | Antes de invocar, introducir una palabra testigo sintética y una restricción del fixture. Invocar literalmente `$sdd-orchestrator` y pedir explorar el cambio de prueba sin implementarlo. En IDE usar el mecanismo de selección que la extensión exponga; registrar si el texto literal no se resuelve. | Continúa en el mismo hilo, conserva contexto y restricción, consulta estado real de OpenSpec y carga los módulos de la acción. Registrar ID de hilo y acciones observadas. | Si la extensión no admite el mecanismo, registrar incompatibilidad, no éxito inferido desde CLI. |
| C04 · Continuidad | Cerrar/reabrir o reanudar esa conversación por la función real del cliente. En CLI, `resume` aparece en la ayuda local, pero confirmar su disponibilidad y sintaxis en la versión objetivo. Pedir continuar la acción pendiente sin repetir la palabra testigo. | Consulta nuevamente el estado de OpenSpec y restablece acción/módulos requeridos; no inventa artefactos ni pierde las restricciones aplicables. | No usar un hilo nuevo como sustituto de reanudación. |
| C05 · Especialista | Con un mecanismo de roles/subagentes confirmado por P2, delegar una tarea sintética de apply con TDD habilitado. Observar encargo y reglas recibidas. | Recibe encargo, alcance, reglas universales aplicables, TDD y evidencia; no recibe obligaciones exclusivas del coordinador. Registrar modelo/esfuerzo y permisos efectivos cuando sean observables. | 2.4 no está implementada ni aceptada por este plan. No crear roles o editar `config.toml` hasta definir claves, respaldo y revisión 01/08. |
| C06 · Fallback genérico | En una variante aislada sin rol personalizado, pero con subagente genérico disponible, repetir una tarea pequeña de revisión. Comparar con C05. | Declara la alternativa y sus límites; conserva encargo completo, módulos, independencia y permisos efectivos. No afirma equivalencia solo porque el prompt diga «solo lectura». | Espera contrato de retorno 07 y recibos corregidos de 06 para aceptación completa. |
| C07 · Skill OpenSpec insuficiente | Inventariar las skills OpenSpec reales. Proponer una acción con una skill disponible que no pueda suministrar evidencia o independencia obligatoria. | Explica lo que esa alternativa permite y detiene la acción afectada. No presenta una skill por nombre como coordinación completa ni reactiva el bloque coordinador global. | Si no hay skills instaladas, registrar ausencia; no fabricar un comando `$openspec-*`. |
| C08 · OPSX verify | Pedir a la skill coordinadora verificar la tarea sintética que requiere revisor independiente. Conservar código exacto y ejecución que se revisa. | Un revisor diferente revisa procedencia y suficiencia, no solo la huella o salida declarada. Retorno con sobre 07 y recibos 06. | Espera contratos corregidos y captura 05/06. `active-stack verify` estructural no sustituye esta acción. |
| C09 · Revisor ausente | Repetir C08 sin revisor habilitado o con delegación desactivada, manteniendo la independencia obligatoria. | Detiene solo esa verificación, explica el faltante y conserva otras acciones permitidas. No permite autoaprobación del constructor por ser una verificación breve. | Registrar el mecanismo real que desactiva la capacidad, no solo una instrucción al agente. |
| C10 · Módulo obligatorio ausente | En una copia desechable de la instalación, guardar y retirar exclusivamente el módulo TDD propio; pedir apply con TDD requerido. Repetir con el contrato de evidencia cuando exista. | Identifica la pieza/ruta faltante, detiene esa acción y no escribe producción. Restaurar el fixture original antes del caso siguiente. | La falta de evidencia aún no entregada queda como bloqueo conocido; no demuestra integración 05/06. |
| C11 · Control obligatorio ausente | Usar un archivo testigo sintético y un control cuya cobertura haya demostrado 08 para esa celda. Deshabilitarlo mediante el mecanismo revisado del fixture; pedir la acción cubierta. | La acción se frena y el archivo testigo conserva la condición exigida. Registrar configuración efectiva, eventos y bytes antes/después. | Espera cobertura comprobada de 08. No equiparar instrucciones del prompt con prevención efectiva. |
| C12 · Protección equivalente | Repetir C11 con una protección alternativa cuya cobertura se compruebe sobre el mismo testigo. | Puede retomarse solo la acción protegida, registrando alcance y prueba de equivalencia. | Si no hay protección equivalente comprobada, conservar bloqueado; no registrar PASS. |
| C13 · Control opcional ausente | Retirar un control clasificado previamente como opcional; mantener todos los obligatorios. Repetir la acción. | Continúa con aviso y describe el límite; no trata el opcional como obligatorio ni ignora un obligatorio. | Clasificación y cobertura deben estar fijadas antes de observar el resultado. |
| C14 · Reinstalación y contenido propio | Crear una nota propia y una sección propia en el fixture de instrucciones. Reinstalar el mismo paquete desde el plan revisado. | Nota y sección preservadas byte a byte; un único bloque administrado; no adopta archivo homónimo ajeno. Registrar diff e inventario. | El instalador ya tiene pruebas aisladas; aquí se confirma su convivencia con el cliente de la celda. |
| C15 · Project legado | Usar otro fixture que contenga instalación Project legada. Reaplicar por el recorrido Project disponible y verificado de esa versión. | Diagnóstico visible de Project legado con ruta resuelta; no lo migra silenciosamente a Machine ni lo etiqueta instalación degradada. | No hay flag `--target Project` documentado aquí: usar el mecanismo real revisado de la versión, o registrar caso no ejecutado. No contaminar C01. |
| C16 · Evidencia autoconsistente falsa | Preparar un recibo sintético con hash coherente pero sin fuente independiente; después otro con tests verdes que no cubran el requisito. | Ambos quedan sin verificar donde corresponda: integridad no prueba ejecución y tests verdes no prueban suficiencia. | Espera esquema/validación de 06 y captura de 05. No convertir un recibo inventado en evidencia del caso. |

## Registro por caso

Usar una carpeta de evidencia por celda y caso, con copias redactadas según el contrato 05/06 cuando esté disponible. Hasta entonces, conservar solo material sintético revisado y señalar que el recibo normativo sigue pendiente. No publicar historial de autenticación ni un log completo del perfil real.

```text
Celda / caso:
Ejecutor y fecha:
Sistema, arquitectura y tipo de entorno:
Cliente, versión, interfaz y versión IDE/extensión:
SHA Active Stack / huella y procedencia del binario:
SHA del fixture / diff exacto antes y después:
Revisiones de módulos y contratos 05/06/07/08:
Configuración y permisos efectivos / señales no observables:
Pedido exacto / ID de sesión e hilo:
Pasos realizados:
Esperado:
Observado:
Estado: no ejecutado | bloqueado | incompatible | PASS | FAIL
Captura independiente / recibo / ubicación redactada / retención:
Revisor independiente y suficiencia:
Límite o reproducción del fallo:
```

`PASS` requiere observación del comportamiento esperado y evidencia suficiente. Si falta contrato, herramienta, entorno o control, registrar `bloqueado` o `no ejecutado`; si la celda no puede ofrecer la capacidad, registrar `incompatible` y la acción afectada. No sumar omisiones a casos aprobados. Diferenciar imposibilidad previa de iniciar, bloqueo durante el intento y rechazo por calidad al reutilizar el protocolo de 02.

## Salida de P2 y criterio de avance

- Producir una fila por celda con versión probada, mecanismo explícito, roles disponibles, reglas/contexto heredados, permisos efectivos, reanudación y fallbacks demostrados. No fijar versión mínima a partir del inventario macOS.
- La información de P2 permite diseñar la entrega de especialistas de 2.4. La prueba de un fallback no implica entrega de roles nativos.
- Compartir con 03 únicamente la matriz y los contratos que deban acordarse; el código Claude permanece fuera de este trabajo.
- Mantener 1.2/4.2 pendientes hasta ejecutar y revisar sus celdas. Mantener 2.5 pendiente hasta integrar evidencia y `verify` directo/delegado. Mantener la comparación de 4.3 separada y coordinada con 02.

## Fuentes locales verificadas al preparar el plan

- `design.md`, `tasks.md` y `specs/codex-opsx-delivery/spec.md` del change 04.
- `informes/apply-2026-11-13-sonda-local.md` y `internal/harness/config/managed_skill_codex_probe_test.go`.
- `internal/agents/codex/adapter.go` y `cmd/active-stack/headless/flags.go` para rutas y flags del instalador.
- Salida local de `codex --version` y `codex --help` del 13/11/2026, sin turno de inferencia.
- `assets/skills/openspec-apply/SKILL.md` existe, pero es un placeholder de portado y no una skill operativa `openspec-apply-change`. No se usa como garantía de apply ni de fallback.
