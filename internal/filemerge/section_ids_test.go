package filemerge

import (
	"reflect"
	"testing"
)

func TestMarkedSectionIDs(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    []string
	}{
		{
			name:    "empty content",
			content: "",
			want:    nil,
		},
		{
			name:    "no markers",
			content: "# Plain config\n\nNothing to see here.\n",
			want:    nil,
		},
		{
			name:    "single well-formed section",
			content: "<!-- active-stack:persona -->\nbody\n<!-- /active-stack:persona -->\n",
			want:    []string{"persona"},
		},
		{
			name: "multiple sections in document order",
			content: "<!-- active-stack:persona -->\na\n<!-- /active-stack:persona -->\n" +
				"<!-- active-stack:engram-protocol -->\nb\n<!-- /active-stack:engram-protocol -->\n" +
				"<!-- active-stack:sdd-orchestrator -->\nc\n<!-- /active-stack:sdd-orchestrator -->\n",
			want: []string{"persona", "engram-protocol", "sdd-orchestrator"},
		},
		{
			name: "nested owned children reported alongside parent",
			content: "<!-- active-stack:sdd-orchestrator -->\n" +
				"intro\n" +
				"<!-- active-stack:sdd-delegation -->\nx\n<!-- /active-stack:sdd-delegation -->\n" +
				"<!-- active-stack:sdd-model-assignments -->\ny\n<!-- /active-stack:sdd-model-assignments -->\n" +
				"<!-- /active-stack:sdd-orchestrator -->\n",
			want: []string{"sdd-orchestrator", "sdd-delegation", "sdd-model-assignments"},
		},
		{
			name:    "open marker without close is ignored",
			content: "<!-- active-stack:orphan -->\nbody but no close marker\n",
			want:    nil,
		},
		{
			name:    "close marker alone is not treated as an open",
			content: "leftover close\n<!-- /active-stack:ghost -->\n",
			want:    nil,
		},
		{
			name: "duplicate id reported once",
			content: "<!-- active-stack:dup -->\nfirst\n<!-- /active-stack:dup -->\n" +
				"<!-- active-stack:dup -->\nsecond\n<!-- /active-stack:dup -->\n",
			want: []string{"dup"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MarkedSectionIDs(tt.content)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MarkedSectionIDs() = %#v, want %#v", got, tt.want)
			}
		})
	}
}
