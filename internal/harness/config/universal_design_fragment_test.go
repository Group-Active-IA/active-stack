package config

import (
	"strings"
	"testing"
)

func TestDesignMockupsVariantsRequireApprovedProjectFoundation(t *testing.T) {
	for _, name := range []string{"design-mockups.md"} {
		t.Run(name, func(t *testing.T) {
			content, err := loadFragment(name)
			if err != nil {
				t.Fatal(err)
			}
			for _, required := range []string{
				"design-system/README.md", "design-system/brand.md", "design-system/tokens.css",
				"design-system/components.html", "design-system/brand.js", "design-system/assets/", "design-system/previews/",
				"Approval: pending", "(TBD)", "Before creating change mockups",
				"Design system revision:", "affected mockups",
				"visual_design.mode", "system-and-mockups", "system-only", "off", "deferred",
				"do not ask again", "design-system skill",
			} {
				if !strings.Contains(content, required) {
					t.Errorf("missing project-wide contract %q", required)
				}
			}
			if strings.Contains(content, "Artifact Design") || strings.Contains(content, "action \"quickstart\"") {
				t.Error("portable workflow must not direct an agent to Artifact Design")
			}
		})
	}
}
