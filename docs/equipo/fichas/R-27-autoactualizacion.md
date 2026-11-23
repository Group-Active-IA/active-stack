# R-27 · Aviso y actualización del binario `active-stack`

**Estado:** pendiente, sin responsable (decisión del dueño, 2026-11-23). **Gobernanza:** ALTO (descarga y reemplaza un ejecutable).
**Origen:** una instalación con binario viejo o dañado no se detecta: hoy `update` solo reaplica la configuración, y el único reemplazo del binario es el self-install de `install`.

## Decisión del dueño

Al arrancar, **avisar y pedir confirmación**. Nunca actualizar solo.

## Alcance

1. **Chequeo al arrancar:** consulta la última release de `Group-Active-IA/active-stack` como máximo una vez al día (cache en `~/.active-stack/`), con tiempo límite corto; sin red, no falla ni demora.
2. **Aviso:** si hay una versión más nueva, muestra versión actual, versión nueva y el comando `active-stack self-update`. No cambia nada por sí solo.
3. **`active-stack self-update`:** descarga el asset de la plataforma, **verifica el SHA256 publicado** antes de reemplazar, hace backup del binario actual (`internal/backup`) y lo reemplaza con rename-then-replace (Windows seguro). Si algo falla, restaura el anterior.
4. **Binario dañado:** si el ejecutable no responde a `--version` o es un binario de test, `verify` lo reporta y sugiere `self-update`.
5. **Apagado:** `ACTIVE_STACK_NO_UPDATE_CHECK=1` o `--no-update-check`; también apagado en CI.
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
