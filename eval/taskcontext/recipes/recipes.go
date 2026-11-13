// Package recipes fija las recetas de captura declarada por el agente (change
// opsx-task-context, D8). El módulo bounded-command-output las lleva byte a byte
// y el arnés las prueba en cada shell disponible. Vive fuera del binario
// (CHANGES.md:49).
package recipes

// Bash es la receta de bash y sh: sin tubería, `$?` es el del comando.
const Bash = `cmd > "$f" 2>&1; echo "exit=$?" > "$f.exit"`

// PowerShell es la receta de PowerShell 5.1 y 7. `*>` en 5.1 escribe UTF-16 LE
// con BOM y envuelve stderr como NativeCommandError; `cmd /c` deja los bytes del
// comando.
const PowerShell = `cmd /c "<command> > ""$f"" 2>&1"; "exit=$LASTEXITCODE" | Set-Content -Encoding ascii "$f.exit"`

// BashScript instancia la receta de bash: command es el comando y f la ruta del
// archivo de captura.
func BashScript(command, f string) string {
	return "f='" + f + "'; " + command + ` > "$f" 2>&1; echo "exit=$?" > "$f.exit"`
}

// PowerShellScript instancia la receta de PowerShell. command va dentro de
// cmd /c "...": las comillas del comando se duplican.
func PowerShellScript(command, f string) string {
	return "$f='" + f + `'; cmd /c "` + command + ` > ""$f"" 2>&1"; "exit=$LASTEXITCODE" | Set-Content -Encoding ascii "$f.exit"`
}
