package catalog

import (
	"fmt"
	"slices"
	"strings"
)

// Paridad por cliente y superficie (D19, tarea 6.8). Las superficies son las
// que declara la 04 (codex-native-opsx/design.md:59-63). Un módulo que un
// cliente no puede recibir queda registrado como límite con su motivo.

// Ámbitos de un límite además de las superficies.
const (
	// ScopeAll: toda superficie del cliente.
	ScopeAll = "all"
	// ScopeNoSubagents: superficie sin subagentes.
	ScopeNoSubagents = "no-subagents"
	// ScopeLegacyMode: modo legado de Claude (anterior a 2.1.110).
	ScopeLegacyMode = "legacy-mode"
)

// Parity es la paridad declarada.
type Parity struct {
	Surfaces []ParitySurface `yaml:"surfaces"`
	Limits   []ParityLimit   `yaml:"limits"`
	Pending  []ParityPending `yaml:"pending"`
}

// ParitySurface es una superficie declarada por la 04.
type ParitySurface struct {
	ID   string `yaml:"id"`
	Name string `yaml:"name"`
}

// ParityLimit: los módulos no se entregan en ese cliente y ámbito.
type ParityLimit struct {
	Modules []string `yaml:"modules"`
	Client  string   `yaml:"client"`
	Scope   string   `yaml:"scope"`
	Reason  string   `yaml:"reason"`
}

// ParityPending es una verificación que espera a otra ficha.
type ParityPending struct {
	Client string `yaml:"client"`
	Task   string `yaml:"task"`
	What   string `yaml:"what"`
}

// DeliveryResult dice si un módulo se entrega y, si no, el límite declarado.
type DeliveryResult struct {
	Delivered bool
	Limit     string
}

// Delivery resuelve si un cliente recibe un módulo en un ámbito (id de
// superficie, ScopeNoSubagents o ScopeLegacyMode).
func (tc *TaskContext) Delivery(client, scope, module string) DeliveryResult {
	for _, l := range tc.Parity.Limits {
		if l.Client == client && (l.Scope == ScopeAll || l.Scope == scope) && slices.Contains(l.Modules, module) {
			return DeliveryResult{Limit: l.Reason}
		}
	}
	return DeliveryResult{Delivered: true}
}

func (tc *TaskContext) validateParity() error {
	p := tc.Parity
	if len(p.Limits) == 0 {
		return fmt.Errorf("catalog: parity declares no limits")
	}
	scopes := []string{ScopeAll, ScopeNoSubagents, ScopeLegacyMode}
	seen := map[string]bool{}
	for _, s := range p.Surfaces {
		if s.ID == "" || s.Name == "" || seen[s.ID] {
			return fmt.Errorf("catalog: parity surface %q is empty or repeated", s.ID)
		}
		seen[s.ID] = true
		scopes = append(scopes, s.ID)
	}
	ids := map[string]bool{}
	for _, m := range tc.Modules {
		ids[m.ID] = true
	}
	for _, l := range p.Limits {
		if len(l.Modules) == 0 || l.Reason == "" {
			return fmt.Errorf("catalog: parity limit for %q needs modules and a reason", l.Client)
		}
		for _, m := range l.Modules {
			if !ids[m] {
				return fmt.Errorf("catalog: parity limit names unknown module %q", m)
			}
		}
		if !slices.Contains(toolClients, l.Client) {
			return fmt.Errorf("catalog: parity limit has unknown client %q", l.Client)
		}
		if !slices.Contains(scopes, l.Scope) {
			return fmt.Errorf("catalog: parity limit has unknown scope %q (want one of %v)", l.Scope, scopes)
		}
	}
	for _, pe := range p.Pending {
		if !slices.Contains(toolClients, pe.Client) || pe.Task == "" || pe.What == "" {
			return fmt.Errorf("catalog: parity pending entry %+v is incomplete", pe)
		}
	}
	return nil
}

// renderParityMarkdown es la sección de paridad de docs/carga-de-modulos-por-accion.md.
func (tc *TaskContext) renderParityMarkdown() string {
	var b strings.Builder
	b.WriteString("\n## Paridad por cliente y superficie\n\n")
	b.WriteString("Superficies declaradas por la 04 (`codex-native-opsx/design.md:59-63`): ")
	var names []string
	for _, s := range tc.Parity.Surfaces {
		names = append(names, s.Name+" (`"+s.ID+"`)")
	}
	b.WriteString(strings.Join(names, "; ") + ".\n\n")
	b.WriteString("Un módulo que un cliente no puede recibir figura como **límite**, no como entrega:\n\n")
	b.WriteString("| Cliente | Ámbito | Módulos | Motivo |\n|---|---|---|---|\n")
	for _, l := range tc.Parity.Limits {
		fmt.Fprintf(&b, "| %s | `%s` | `%s` | %s |\n", clientTitle(l.Client), l.Scope, strings.Join(l.Modules, "`, `"), l.Reason)
	}
	b.WriteString("\nPendiente de otra ficha:\n\n")
	for _, p := range tc.Parity.Pending {
		fmt.Fprintf(&b, "- %s: %s — %s.\n", clientTitle(p.Client), p.Task, p.What)
	}
	return b.String()
}
