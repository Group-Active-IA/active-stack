En este módulo, `store.Store.Get` devuelve `""` tanto para una clave inexistente como para una clave guardada con valor vacío, y `api.Describe` no puede distinguirlas.

Cambiá el contrato a `func (s *Store) Get(k string) (string, bool)`, donde el `bool` indica si la clave existe. Actualizá `api.Describe` para que:

- una clave inexistente siga mostrando `clave: (sin valor)`;
- una clave con valor vacío muestre `clave: (vacío)`;
- una clave con valor muestre `clave: valor`, como hoy.

Las llamadas existentes y sus pruebas en `api/api_test.go` no deben cambiar. Podés tocar `store/store.go`, `store/store_test.go` y `api/api.go`. Al terminar, `go test ./...` debe pasar.
