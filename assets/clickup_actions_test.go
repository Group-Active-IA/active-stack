package assets

import (
	"io/fs"
	"strings"
	"testing"
)

func TestClickUpActionDiscovery(t *testing.T) {
	for file, expectations := range map[string][]string{
		"skills/active-roadmap/SKILL.md":                      {"references/clickup-actions.md", "clickup_get_list", "lista numerada", "pedido concreto", "Ver detalle"},
		"skills/active-roadmap/references/clickup-actions.md": {"Publicar fichas", "Actualizar una tarea vinculada", "Ver y sincronizar avance", "Configurar ClickUp", "Seguir con el archivo local", "clickup_create_task", "clickup_update_task", "Ejemplo", "Qué cambia"},
	} {
		data, err := fs.ReadFile(SkillsFS, file)
		if err != nil {
			t.Error(err)
			continue
		}
		for _, want := range expectations {
			if !strings.Contains(string(data), want) {
				t.Errorf("%s lacks %q", file, want)
			}
		}
	}
}
