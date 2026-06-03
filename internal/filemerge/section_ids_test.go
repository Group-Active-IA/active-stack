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
			content: "<!-- jr-stack:persona -->\nbody\n<!-- /jr-stack:persona -->\n",
			want:    []string{"persona"},
		},
		{
			name: "multiple sections in document order",
			content: "<!-- jr-stack:persona -->\na\n<!-- /jr-stack:persona -->\n" +
				"<!-- jr-stack:engram-protocol -->\nb\n<!-- /jr-stack:engram-protocol -->\n" +
				"<!-- jr-stack:sdd-orchestrator -->\nc\n<!-- /jr-stack:sdd-orchestrator -->\n",
			want: []string{"persona", "engram-protocol", "sdd-orchestrator"},
		},
		{
			name: "nested owned children reported alongside parent",
			content: "<!-- jr-stack:sdd-orchestrator -->\n" +
				"intro\n" +
				"<!-- jr-stack:sdd-delegation -->\nx\n<!-- /jr-stack:sdd-delegation -->\n" +
				"<!-- jr-stack:sdd-model-assignments -->\ny\n<!-- /jr-stack:sdd-model-assignments -->\n" +
				"<!-- /jr-stack:sdd-orchestrator -->\n",
			want: []string{"sdd-orchestrator", "sdd-delegation", "sdd-model-assignments"},
		},
		{
			name:    "open marker without close is ignored",
			content: "<!-- jr-stack:orphan -->\nbody but no close marker\n",
			want:    nil,
		},
		{
			name:    "close marker alone is not treated as an open",
			content: "leftover close\n<!-- /jr-stack:ghost -->\n",
			want:    nil,
		},
		{
			name: "duplicate id reported once",
			content: "<!-- jr-stack:dup -->\nfirst\n<!-- /jr-stack:dup -->\n" +
				"<!-- jr-stack:dup -->\nsecond\n<!-- /jr-stack:dup -->\n",
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
