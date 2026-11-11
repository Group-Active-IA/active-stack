# Tasks

## 1. Configuración y migración

- [x] 1.1 Confirmar con el equipo el contrato de Estricto y la matriz de versiones y entornos. Registrar el acuerdo antes de editar rutas que escriben configuración de usuario.
- [x] 1.2 Aplicar `read-only` y `on-request` para Codex Estricto mediante el merge existente. Verificar con pruebas de instalación limpia, preferencias y perfiles preexistentes, respaldo fallido y contenido idéntico tras reinstalar.
- [x] 1.3 Aplicar las descripciones corta y larga de la propuesta y comprobar ambas en español e inglés en los selectores afectados.
- [x] 1.4 Cubrir el replay de la selección registrada y la ruta sin registro. Verificar con fixtures que se conserva el tier registrado, que sin registro se indica la reinstalación y que un fallo que exige rollback restaura el archivo anterior.

## 2. Aceptación integrada

- [x] 2.1 Ejecutar en perfiles aislados de los entornos acordados una prueba de carga con el TOML generado y una sesión que lea un archivo permitido y pruebe una escritura con escalación denegada. Registrar versión, entorno, salida y estado final del archivo.
- [x] 2.2 Completar la matriz con escritura sin escalación, escalación aprobada y superficie sin solicitud de escalación. Registrar política efectiva, quién aprueba y estado final. Entregar el resultado a `opsx-baseline` para medir Estricto por separado.
- [x] 2.3 Verificar diagnóstico de valor legado en otra capa, archivo inválido intacto y rollback con ediciones posteriores.
