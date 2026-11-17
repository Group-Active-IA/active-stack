package config

import (
	"strings"
	"testing"
)

func TestCodexActiveDeliveryEntrypoint(t *testing.T) {
	machine, err := ComposeCodexPackage(nil)
	if err != nil {
		t.Fatal(err)
	}
	project, err := Compose(nil, "codex")
	if err != nil {
		t.Fatal(err)
	}
	for name, instructions := range map[string]string{"machine": machine.Universal, "project": project} {
		t.Run(name, func(t *testing.T) {
			if !strings.Contains(instructions, "$active-delivery") {
				t.Error("instructions lack the installed delivery entrypoint")
			}
		})
	}
}
