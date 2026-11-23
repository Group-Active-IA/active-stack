# Pruebas pendientes del modo equipo y Engram Cloud

**Frase de inicio** · «Soy <nombre>, quiero hacer las pruebas pendientes» (o «continuar» dentro de este trabajo).

Lo que ya se verificó en una sola máquina, con Claude y en Windows, está en los informes [R-23](equipo/informes/R-23.md) y [R-24](equipo/informes/R-24.md). Esta lista reúne **lo que nadie pudo probar todavía**. Cada prueba dice quién la necesita, qué se espera y dónde registrar la evidencia. El agente toma la primera prueba abierta que corresponda a la persona, la ejecuta con ella y agrega la evidencia al informe indicado. Si una prueba falla, se documenta como límite por agente (DA-12); no se oculta.

Reglas para el agente, todas aprendidas en R-23 y R-24:

- Entorno aislado: `USERPROFILE`, `HOME`, `LOCALAPPDATA` y `APPDATA` apuntan a un temporal, **inline por comando**; nunca se ejecuta `active-stack` ni `claude` con el HOME real.
- Credenciales: copiar un `.credentials.json` a un HOME temporal puede rotar el token y dejar caducado el original. Al terminar se devuelve la copia refrescada, se verifica `claude -p "di ok" --max-turns 1` y se borran todas las copias.
- Nunca `go build` sin pedido; el binario del PATH ya es un build de `main`.
- El proyecto de prueba va fuera del perfil de usuario para que no cargue el `CLAUDE.md` global.

| ID | Prueba | Necesita | Se espera | Evidencia en |
|---|---|---|---|---|
| PP-01 | **Codex real**: R-16 X0–X9 (hooks de lectura, confianza, binario fuera del PATH) | `codex` con login | La tabla X0–X9 completa de la [ficha R-16](equipo/fichas/R-16.md) | informe R-16 |
| PP-02 | **Codex real**: `deny` de push a `main` | `codex` | El push a `main` queda denegado y el remoto no cambia | informe R-18 |
| PP-03 | **Codex real**: proxy, hooks y AUTOSYNC registrados por `install` | `codex` | El bloque `# active-stack:memory-proxy` y los hooks quedan; `uninstall` devuelve el archivo original | informe R-23 |
| PP-04 | **Codex real**: V1–V13 y escenario 6 (traspaso Claude ↔ Codex) | `codex`, Claude | La matriz por agente; el traspaso retoma desde el handoff | informe R-23 |
| PP-05 | **Dos personas reales** con el servidor Engram Cloud de prueba: lo guardado por una lo lee la otra, con etiqueta de rama | Dos personas, servidor | V1, V3 y V13 con dos identidades; aviso 5 de la Fase 9 | informe R-23 |
| PP-06 | **Cloud real y dashboard**: sincronización y proyectos en el dashboard | Servidor de Matías | Los proyectos aparecen con el nombre en minúsculas | informe R-23 |
| PP-07 | **Máquina limpia o VM** (escenario 2 de la Fase 9): `install` Full desde cero | VM Windows | Instalación sin pasos manuales; `verify` sin avisos | informe R-24 |
| PP-08 | **Linux, macOS, WSL y Termux** (V12) | Cada sistema | Mismo resultado que Windows | informe R-23 |
| PP-09 | **OpenCode** (escenario 10) | OpenCode | El modo equipo no escribe nada de la nube; la fundación no falla | informe R-24 |
| PP-10 | **PR real en GitHub**: aviso 4 y PR en borrador del escenario 3 | Repo de prueba en GitHub | El PR en borrador se crea con la plantilla | informe R-24 |
| PP-11 | **Línea base de dos semanas** (VE-8 con trabajo real, no con un saludo) | Un equipo trabajando | Tokens con modo equipo dentro del umbral DA-02 (+35 %) | informe R-24 |
| PP-12 | **VE-6 con un pedido concreto** y «qué me toca» con `gh` autenticado | `gh` con login | Detección en al menos 4 de 5 corridas (DA-03) y «qué me toca» visible | informe R-24 |
| PP-13 | **`verify` visible por CLI**: hoy `install` imprime «0 checks» y no hay comando `verify` suelto | Cambio de código | Los `cloud:*` se ven al instalar | informe R-23 |
| PP-14 | **Aviso de marcador inválido en el proxy** (el hook de lectura ya avisa) | Cambio de código | Aviso por `stderr` sin bloquear el guardado | informe R-23 |
| PP-15 | **Proxy con Engram solo por el plugin de Claude** (sin entrada `engram` en `~/.claude.json`) | Claude con plugin | Hoy se saltea con aviso; decidir si se registra de otra forma | ficha R-26 |

Cuando una prueba pasa, se marca en este archivo (fecha, quién, enlace a la evidencia) y se actualiza el informe correspondiente.
