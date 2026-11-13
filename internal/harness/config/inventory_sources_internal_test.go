package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/agents/claude"
	"github.com/Group-Active-IA/active-stack/internal/agents/codex"
	"github.com/Group-Active-IA/active-stack/internal/agents/opencode"
	"github.com/Group-Active-IA/active-stack/internal/catalog"
	"github.com/Group-Active-IA/active-stack/internal/model"
)

// Tareas 1.4 y 1.6 (opsx-task-context): qué ruta de código produce el bloque
// instalado hoy y cuántos bytes de instrucciones llegan realmente al proyecto.

func catalogHarness(t *testing.T, id string) model.Harness {
	t.Helper()
	c, err := catalog.Load()
	if err != nil {
		t.Fatal(err)
	}
	h, ok := c.ByID(id)
	if !ok {
		t.Fatalf("el catálogo no declara %q", id)
	}
	return h
}

// Tarea 1.4. Con la 03 y la 04 en main, Claude y Codex entregan por
// ConfigDeliveryManagedSkill (bloque universal en CLAUDE.md o AGENTS.md más la
// skill coordinadora) y OpenCode por ConfigDeliveryPrimaryAgent.
// Este test documenta la ruta vigente de cada cliente.
func TestInstall_DeliveryRouteTodayOnMain(t *testing.T) {
	h := catalogHarness(t, "sdd-orchestrator")
	if got := strings.Join(h.Toggles, ","); got != "tdd,engram,model-routing,delegation,governance" {
		t.Fatalf("los toggles salen literales del catálogo: %q", got)
	}

	tests := []struct {
		name         string
		adapter      AgentAdapter
		wantDelivery model.ConfigDelivery
	}{
		{"claude", claude.NewAdapter(), model.ConfigDeliveryManagedSkill},
		{"codex", codex.NewAdapter(), model.ConfigDeliveryManagedSkill},
		{"opencode", opencode.NewAdapter(), model.ConfigDeliveryPrimaryAgent},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.adapter.ConfigDelivery(); got != tt.wantDelivery {
				t.Fatalf("ConfigDelivery() = %v, want %v", got, tt.wantDelivery)
			}
			home := t.TempDir()
			if _, err := installDeterministic(h, tt.adapter, home); err != nil {
				t.Fatalf("Install: %v", err)
			}
			full, err := Compose(h.Toggles, tt.adapter.VariantKey())
			if err != nil {
				t.Fatal(err)
			}
			instructions := readOrEmpty(tt.adapter.InstructionsPath(home))

			switch tt.wantDelivery {
			case model.ConfigDeliveryManagedSkill:
				// Capas: CLAUDE.md lleva solo el bloque universal (no el compuesto de cinco
				// toggles) y el coordinador vive en la skill gestionada.
				if strings.Contains(instructions, full) {
					t.Errorf("el archivo de instrucciones de %s no debe llevar el bloque completo de %d B: va en capas", tt.name, len(full))
				}
				// Cada cliente compone su propio paquete: Claude con ComposeLayers (03),
				// Codex con ComposeCodexPackage (04).
				var universal string
				if tt.adapter.Agent() == model.AgentCodex {
					pkg, err := ComposeCodexPackage(h.Toggles)
					if err != nil {
						t.Fatal(err)
					}
					universal = pkg.Universal
				} else {
					layers, err := ComposeLayers(h.Toggles)
					if err != nil {
						t.Fatal(err)
					}
					universal = layers.Universal
				}
				if !strings.Contains(instructions, universal) {
					t.Errorf("el archivo de instrucciones de %s no contiene el bloque universal", tt.name)
				}
				native, ok := tt.adapter.(interface{ SkillsDir(string) string })
				if !ok {
					t.Fatalf("%s entrega por skill pero su adapter no expone SkillsDir", tt.name)
				}
				skill := filepath.Join(native.SkillsDir(home), "sdd-orchestrator", "SKILL.md")
				if _, err := os.Stat(skill); err != nil {
					t.Errorf("falta la skill coordinadora %s: %v", skill, err)
				}
			case model.ConfigDeliveryInstructions:
				// Rama default: el bloque completo (cinco toggles) va al archivo de instrucciones.
				if !strings.Contains(instructions, full) {
					t.Errorf("el archivo de instrucciones de %s no contiene el bloque compuesto con los cinco toggles (%d B)", tt.name, len(full))
				}
			case model.ConfigDeliveryPrimaryAgent:
				// Rama primary-agent: se registra en el JSON de settings, no en las instrucciones.
				settings := readOrEmpty(tt.adapter.SettingsPath(home))
				if !strings.Contains(settings, "sdd-orchestrator") {
					t.Errorf("OpenCode no registró sdd-orchestrator como agente primario en %s", tt.adapter.SettingsPath(home))
				}
				if strings.Contains(instructions, full) {
					t.Error("OpenCode no debe inyectar el bloque en el archivo de instrucciones")
				}
			}
		})
	}
}

// installDeterministic instala con la versión de Claude Code fijada como
// compatible, para que el resultado no dependa del binario `claude` de la máquina.
func installDeterministic(h model.Harness, ad AgentAdapter, home string) (Result, error) {
	return InstallWith(h, []AgentAdapter{ad}, home, InstallOptions{
		Home: home,
		DetectVersion: func() claude.VersionCheck {
			return claude.VersionCheck{State: claude.VersionCompatible, Version: "2.1.318", Min: claude.MinVersion}
		},
	})
}

func readOrEmpty(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(data)
}

// Tarea 1.6. Fuentes de instrucción que el instalador escribe en el archivo
// de instrucciones del cliente: el bloque del orquestador y las secciones
// propias (design-mockups). La suma por cliente coincide con el total realmente
// inyectado: cada sección se mide instalándola sola en un home limpio y el total
// se mide instalando ambas; la diferencia es el separador entre secciones.
func TestInventory_InjectedInstructionSources(t *testing.T) {
	orch := catalogHarness(t, "sdd-orchestrator")
	mock := catalogHarness(t, "design-mockups")

	type row struct {
		client string
		ad     AgentAdapter
	}
	rows := []row{{"claude", claude.NewAdapter()}, {"codex", codex.NewAdapter()}}

	var report strings.Builder
	report.WriteString("# Fuentes de instrucción inyectadas en el archivo de instrucciones (tarea 1.6)\n")
	report.WriteString("# client\torchestrator_section\tdesign_mockups_section\tseparator\ttotal_injected\n")
	for _, r := range rows {
		sizeOf := func(harnesses ...model.Harness) int {
			home := t.TempDir()
			for _, h := range harnesses {
				if _, err := installDeterministic(h, r.ad, home); err != nil {
					t.Fatalf("%s: Install(%s): %v", r.client, h.ID, err)
				}
			}
			data, err := os.ReadFile(r.ad.InstructionsPath(home))
			if err != nil {
				t.Fatal(err)
			}
			return len(data)
		}
		a := sizeOf(orch)
		b := sizeOf(mock)
		total := sizeOf(orch, mock)
		sep := total - a - b
		if sep < 0 || sep > 2 {
			t.Fatalf("%s: el separador entre secciones debe ser de 0 a 2 bytes, es %d (a=%d b=%d total=%d)", r.client, sep, a, b, total)
		}
		if a+b+sep != total {
			t.Errorf("%s: la suma de las fuentes (%d) no coincide con el total inyectado (%d)", r.client, a+b+sep, total)
		}
		// La sección propia de design-mockups pesa lo que el fragmento más sus marcadores.
		frag, err := loadFragment("design-mockups.md")
		if err != nil {
			t.Fatal(err)
		}
		if b < len(frag) || b > len(frag)+200 {
			t.Errorf("%s: la sección design-mockups (%d B) no corresponde al fragmento (%d B) más sus marcadores", r.client, b, len(frag))
		}
		fmt.Fprintf(&report, "%s\t%d\t%d\t%d\t%d\n", r.client, a, b, sep, total)
	}
	checkInventoryGolden(t, "inventory_sources", report.String())
}

// El fragmento design-mockups mide 8.865 B (valor de partida de la tarea 1.6).
func TestInventory_DesignMockupsFragmentSize(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("assets", "design-mockups.md"))
	if err != nil {
		t.Fatal(err)
	}
	if len(data) != 8865 {
		t.Errorf("design-mockups.md mide %d B, el inventario de partida registra 8.865 B: si cambió a propósito, actualizá el valor y el documento", len(data))
	}
}
