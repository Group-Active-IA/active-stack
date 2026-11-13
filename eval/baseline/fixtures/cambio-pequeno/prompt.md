Reporte de un usuario: `Truncate` del paquete `texto` corta mal los textos con letras acentuadas o `ñ` y a veces deja caracteres rotos.

Corregí `Truncate` para que cuente caracteres y no bytes, sin cambiar su firma ni el comportamiento con texto ASCII. Agregá una prueba de regresión que falle con el código actual y pase con tu corrección. Tocá solo `truncate.go` y `truncate_test.go`.

Al terminar, `go test ./...` debe pasar en la raíz del checkout.
