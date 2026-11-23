# R-27 · Aviso y actualización del binario `active-stack`

**Estado:** implementada en la rama `feat/R-27-autoactualizacion` y en revisión del dueño (nivel ALTO: no se mergea sin su OK). Faltan los puntos marcados «pendiente» abajo. **Gobernanza:** ALTO (descarga y reemplaza un ejecutable).
**Origen:** una instalación con binario viejo o dañado no se detecta: hoy `update` solo reaplica la configuración, y el único reemplazo del binario es el self-install de `install`.

## Decisión del dueño

Al arrancar, **avisar y pedir confirmación**. Nunca actualizar solo.

## Alcance

1. **Chequeo al arrancar:** consulta la última release de `Group-Active-IA/active-stack` como máximo una vez al día (cache en `~/.active-stack/`), con tiempo límite corto; sin red, no falla ni demora.
2. **Aviso:** si hay una versión más nueva, muestra versión actual, versión nueva y el comando `active-stack self-update`. No cambia nada por sí solo.
3. **`active-stack self-update`:** descarga el asset de la plataforma, **verifica el SHA256 publicado** antes de reemplazar, hace backup del binario actual (`internal/backup`) y lo reemplaza con rename-then-replace (Windows seguro). Si algo falla, restaura el anterior.
4. **Binario dañado (pendiente):** que `verify` revise que el `active-stack` del PATH responde a `--version` y sugiera `self-update`. Hoy solo se evita que `install` copie un binario de test (#72).
5. **Apagado:** `ACTIVE_STACK_NO_UPDATE_CHECK=1`; también apagado cuando `CI` está definida. El flag `--no-update-check` **no se implementó** (los parsers de `install` y `update` rechazarían un flag desconocido): decisión a confirmar.
6. **Fuera de alcance:** actualizar solo y en silencio; actualizar el catálogo o los harnesses (eso lo hace `update`).

## Criterios de aceptación (tests)

- Sin red: arranca igual y no imprime errores.
- Dentro de las 24 h no vuelve a consultar.
- Versión igual o menor: no avisa.
- SHA256 que no coincide: no reemplaza y deja el binario actual intacto.
- Fallo a mitad de reemplazo: queda el binario anterior.
- No toca configuración de agentes.

## Verificación

Con HOME aislado y un servidor de releases falso local. Nunca contra el binario real del PATH sin pedido del dueño.

## Qué quedó hecho

- `internal/selfupdate`: `Check` (una consulta por día, 2 s de límite, silencioso sin red o con build `dev`) y `Apply` (SHA256 contra `SHA256SUMS.txt`, backup con `internal/backup`, reemplazo rename-then-replace con restauración).
- `cmd/active-stack`: aviso a `stderr` solo en el arranque sin argumentos, `install` y `update` (nunca en hooks ni proxy) y el comando `self-update [--yes]` con confirmación.
- Plataformas con asset publicado: Windows y Linux amd64; en otras `self-update` lo dice y no cambia nada.
- Verificado con tests y HOME aislado; **no se probó contra una release real** (PP-16).
