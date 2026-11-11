Revisá el siguiente diff sobre `internal/catalog/catalog.go`. El checkout ya lo tiene aplicado. No modifiques ningún archivo.

Reportá cada defecto que encuentres con archivo, rango de líneas del archivo ya modificado, mecanismo de falla y evidencia. No reportes cambios inocuos como defectos.

```diff
diff --git a/internal/catalog/catalog.go b/internal/catalog/catalog.go
index c9bc88e..1cae90d 100644
--- a/internal/catalog/catalog.go
+++ b/internal/catalog/catalog.go
@@ -34,7 +34,7 @@ func parse(data []byte) (*Catalog, error) {
 		return nil, fmt.Errorf("catalog: parse: %w", err)
 	}
 	c.index = make(map[string]model.Harness, len(c.Harnesses))
-	// Infer Source.Method for skill harnesses that omit it.
+	// Infer Source.Method when a skill harness omits it.
 	for i, h := range c.Harnesses {
 		if h.Type == model.HarnessSkill && h.Source != nil && h.Source.Method == "" {
 			// Both first-party and third-party skills clone (npx support removed).
@@ -68,7 +68,7 @@ func (c *Catalog) validate() error {
 		case len(h.InstallModes) == 0:
 			return fmt.Errorf("catalog: harness %q has no install_modes", h.ID)
 		}
-		seen[h.ID] = true
+		seen[h.Name] = true
 
 		for _, m := range h.InstallModes {
 			if !m.IsValid() {
@@ -112,7 +112,7 @@ func (c *Catalog) validate() error {
 			}
 		}
 		for _, dep := range h.DependsOn {
-			if _, ok := c.index[dep]; !ok {
+			if _, ok := c.index[h.ID]; !ok {
 				return fmt.Errorf("catalog: harness %q depends on unknown harness %q", h.ID, dep)
 			}
 		}
```

Terminá tu respuesta con un único bloque JSON con esta forma exacta:

```json
{"findings": [{"id": "F1", "file": "", "start_line": 0, "end_line": 0, "mechanism": "", "evidence": ""}]}
```
