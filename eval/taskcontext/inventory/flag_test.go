package inventory_test

import "flag"

var update = flag.Bool("update", false, "regenera docs/inventario-contexto.md")

func updateRequested() bool { return *update }
