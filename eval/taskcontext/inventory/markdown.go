package inventory

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Group-Active-IA/active-stack/internal/catalog"
)

// thousands formatea con punto de millares, como el resto de las cifras del change.
func thousands(n int) string {
	s := strconv.Itoa(n)
	if len(s) <= 3 {
		return s
	}
	var parts []string
	for len(s) > 3 {
		parts = append([]string{s[len(s)-3:]}, parts...)
		s = s[:len(s)-3]
	}
	return s + "." + strings.Join(parts, ".")
}

func bytesLines(r ComposeRow) string {
	return fmt.Sprintf("%s B · %d L", thousands(r.Bytes), r.Lines)
}

// percent devuelve el porcentaje con coma decimal y un decimal.
func percent(part, whole int) string {
	p := float64(part) * 1000 / float64(whole)
	return strings.Replace(fmt.Sprintf("%.1f", float64(int(p+0.5))/10), ".", ",", 1) + " %"
}

// cumulativeLabels es el orden de las seis combinaciones acumulativas.
var cumulativeLabels = []string{"none", "+delegation", "+model-routing", "+governance", "+engram", "+tdd"}

var fragmentOf = map[string]string{
	"+delegation": "delegation", "+model-routing": "model-routing",
	"+governance": "governance", "+engram": "engram", "+tdd": "tdd",
}

// Markdown genera docs/inventario-contexto.md (tareas 1.3 y 1.7).
func Markdown(root string, cat *catalog.Catalog, tp ThirdParty) (string, error) {
	compose, err := LoadCompose(root)
	if err != nil {
		return "", err
	}
	idx := map[string]ComposeRow{}
	for _, r := range compose {
		idx[r.Variant+"|"+r.Scope] = r
	}
	get := func(variant, scope string) (ComposeRow, error) {
		r, ok := idx[variant+"|"+scope]
		if !ok {
			return r, fmt.Errorf("inventory: el golden de Compose no tiene %s/%s", variant, scope)
		}
		return r, nil
	}
	injected, err := LoadInjected(root)
	if err != nil {
		return "", err
	}
	srcs, err := Build(root, cat, tp)
	if err != nil {
		return "", err
	}
	unmeasured, err := Unmeasured(root, cat, tp)
	if err != nil {
		return "", err
	}

	var b strings.Builder
	b.WriteString("# Inventario del contexto de partida\n\n")
	b.WriteString("> Generado por `go test ./eval/taskcontext/inventory -update` desde los goldens de\n")
	b.WriteString("> `internal/harness/config/testdata/` (`inventory_compose.golden`, `inventory_sources.golden`), el catálogo\n")
	b.WriteString("> y `eval/taskcontext/inventory/skills-terceros.json`. No se edita a mano: un test falla si se desvía.\n")
	b.WriteString("> Change `opsx-task-context`, tareas 1.1 a 1.7. Las líneas son `wc -l` (cantidad de saltos de línea).\n\n")

	// 1. Combinaciones acumulativas.
	b.WriteString("## 1. Bloque compuesto por combinación de toggles\n\n")
	b.WriteString("El bloque instalado hoy es el último renglón: Claude entrega en capas por `ConfigDeliveryManagedSkill` (a CLAUDE.md llega solo el bloque universal; el coordinador vive en una skill aparte, que este inventario aún no suma), Codex compone por la rama `default` de `internal/harness/config/install.go` y OpenCode por `ConfigDeliveryPrimaryAgent`, y `h.Toggles` sale literal del catálogo con los cinco toggles activos.\n\n")
	b.WriteString("| Toggles activos | Claude | Codex |\n|---|---|---|\n")
	for _, l := range cumulativeLabels {
		c, err := get("claude", l)
		if err != nil {
			return "", err
		}
		x, err := get("codex", l)
		if err != nil {
			return "", err
		}
		label := l
		if l == "none" {
			label = "ninguno"
		}
		if l == "+tdd" {
			fmt.Fprintf(&b, "| %s (**lo que se instala hoy**) | **%s** | **%s** |\n", label, bytesLines(c), bytesLines(x))
		} else {
			fmt.Fprintf(&b, "| %s | %s | %s |\n", label, bytesLines(c), bytesLines(x))
		}
	}

	// 2. Atribución por fragmento.
	b.WriteString("\n## 2. Atribución de bytes por fragmento\n\n")
	b.WriteString("Cada fragmento es la diferencia entre dos combinaciones consecutivas; la base es el bloque sin toggles.\n\n")
	b.WriteString("| Fragmento | Claude | % | Codex | % |\n|---|---|---|---|---|\n")
	cFull, _ := get("claude", "+tdd")
	xFull, _ := get("codex", "+tdd")
	cBase, _ := get("claude", "none")
	xBase, _ := get("codex", "none")
	fmt.Fprintf(&b, "| base del variante | %s B | %s | %s B | %s |\n",
		thousands(cBase.Bytes), percent(cBase.Bytes, cFull.Bytes), thousands(xBase.Bytes), percent(xBase.Bytes, xFull.Bytes))
	for i := 1; i < len(cumulativeLabels); i++ {
		l := cumulativeLabels[i]
		cp, _ := get("claude", cumulativeLabels[i-1])
		cc, _ := get("claude", l)
		xp, _ := get("codex", cumulativeLabels[i-1])
		xc, _ := get("codex", l)
		cd, xd := cc.Bytes-cp.Bytes, xc.Bytes-xp.Bytes
		fmt.Fprintf(&b, "| `%s` | %s B | %s | %s B | %s |\n",
			fragmentOf[l], thousands(cd), percent(cd, cFull.Bytes), thousands(xd), percent(xd, xFull.Bytes))
	}
	eng1, _ := get("claude", "+governance")
	eng2, _ := get("claude", "+tdd")
	xeng1, _ := get("codex", "+governance")
	xeng2, _ := get("codex", "+tdd")
	both := eng2.Bytes - eng1.Bytes
	xboth := xeng2.Bytes - xeng1.Bytes
	fmt.Fprintf(&b, "\n`engram` y `tdd` juntos son %s del bloque de Claude y %s del de Codex.\n", percent(both, cFull.Bytes), percent(xboth, xFull.Bytes))

	// 3. Otras variantes.
	b.WriteString("\n## 3. Otras variantes (referencia, fuera del alcance de paridad)\n\n")
	b.WriteString("| Variante | Base | Con los cinco toggles |\n|---|---|---|\n")
	for _, v := range []string{"antigravity", "cursor", "gemini", "generic", "opencode", "windsurf"} {
		n, err := get(v, "none")
		if err != nil {
			return "", err
		}
		a, err := get(v, "all")
		if err != nil {
			return "", err
		}
		fmt.Fprintf(&b, "| %s | %s | %s |\n", v, bytesLines(n), bytesLines(a))
	}

	// 4. Fuentes inyectadas.
	b.WriteString("\n## 4. Fuentes inyectadas en el archivo de instrucciones\n\n")
	b.WriteString("Cada sección se mide instalándola sola en un home limpio; el total, instalando ambas. La suma por cliente coincide con el total realmente inyectado (el separador entre secciones es de un byte).\n\n")
	b.WriteString("| Cliente | Bloque del orquestador | Sección design-mockups | Separador | Total inyectado |\n|---|---|---|---|---|\n")
	for _, r := range injected {
		fmt.Fprintf(&b, "| %s | %s B | %s B | %d B | %s B |\n", clientTitle(r.Client),
			thousands(r.Orchestrator), thousands(r.DesignMock), r.Separator, thousands(r.Total))
	}
	b.WriteString("\nEl fragmento `design-mockups` mide 8.865 B; con los marcadores de sección de la inyección son 8.940 B.\n")

	// 5. Tabla única.
	b.WriteString("\n## 5. Todas las fuentes, por tamaño descendente\n\n")
	top := Top(srcs, 5)
	b.WriteString("Las cinco más grandes están en negrita: ")
	for i, s := range top {
		if i > 0 {
			b.WriteString(", ")
		}
		fmt.Fprintf(&b, "%s (%s B)", s.Name, thousands(s.Bytes))
	}
	b.WriteString(".\n\n")
	b.WriteString("| # | Fuente | Tipo | Bytes | Nota |\n|---|---|---|---|---|\n")
	isTop := map[string]bool{}
	for _, s := range top {
		isTop[s.Name] = true
	}
	for i, s := range srcs {
		name := s.Name
		if isTop[name] {
			name = "**" + name + "**"
		}
		fmt.Fprintf(&b, "| %d | %s | %s | %s | %s |\n", i+1, name, s.Kind, thousands(s.Bytes), s.Note)
	}

	if len(unmeasured) > 0 {
		b.WriteString("\n### Skills del catálogo sin medición\n\n")
		b.WriteString("Ninguna queda sin entrada: estas dicen por qué no tienen tamaño.\n\n")
		b.WriteString("| Skill | Motivo |\n|---|---|\n")
		for _, s := range unmeasured {
			fmt.Fprintf(&b, "| %s | %s |\n", s.ID, s.Status)
		}
	}
	fmt.Fprintf(&b, "\nLas skills de terceros se midieron una vez (%s) contra el commit que resolvió cada `ref: latest`; no son reproducibles bit a bit hasta que el catálogo las pinee por SHA.\n", tp.MeasuredOn)
	return b.String(), nil
}
