package filemerge

import "testing"

func TestRemoveMarkdownSectionsPreservesOutsideBytes(t *testing.T) {
	before := "\r\n  Usuario 日本語\r\n\r\n"
	after := "\r\n\r\n  Más reglas\r\n"
	section := "<!-- active-stack:sdd-orchestrator -->\r\nmanaged\r\n<!-- /active-stack:sdd-orchestrator -->"
	got := RemoveMarkdownSectionsPreservingContent(before+section+after, "sdd-orchestrator")
	if got != before+after {
		t.Fatalf("outside bytes changed: %q, want %q", got, before+after)
	}
}

func TestRemoveMarkdownSectionsAllRecognizedOccurrences(t *testing.T) {
	section := func(prefix string) string {
		return "<!-- " + prefix + ":sdd-orchestrator -->\nmanaged\n<!-- /" + prefix + ":sdd-orchestrator -->"
	}
	unrelated := "<!-- active-stack:design-mockups -->\nkeep\n<!-- /active-stack:design-mockups -->"
	input := "A\n" + section("active-stack") + "\n\nB\r\n" + section("jr-stack") + "\r\n" + unrelated + "\n" + section("active-stack") + "\nC"
	want := "A\n\n\nB\r\n\r\n" + unrelated + "\n\nC"
	if got := RemoveMarkdownSectionsPreservingContent(input, "sdd-orchestrator"); got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestRemoveMarkdownSectionsKeepsMalformedMarkers(t *testing.T) {
	for _, input := range []string{
		"before\n<!-- active-stack:sdd-orchestrator -->\nunfinished",
		"before\n<!-- /active-stack:sdd-orchestrator -->\nafter",
		"<!-- active-stack:sdd-orchestrator -->\nwrong prefix\n<!-- /jr-stack:sdd-orchestrator -->",
		"<!-- active-stack:sdd orchestrator -->x<!-- /active-stack:sdd orchestrator -->",
	} {
		if got := RemoveMarkdownSectionsPreservingContent(input, "sdd-orchestrator"); got != input {
			t.Fatalf("malformed content changed: got %q, want %q", got, input)
		}
	}
}

func TestRemoveMarkdownSectionsAfterStrayClosingMarker(t *testing.T) {
	before := "<!-- /active-stack:sdd-orchestrator -->\nnotes\n"
	section := "<!-- active-stack:sdd-orchestrator -->x<!-- /active-stack:sdd-orchestrator -->"
	if got := RemoveMarkdownSectionsPreservingContent(before+section+"\nend", "sdd-orchestrator"); got != before+"\nend" {
		t.Fatalf("got %q, want stray close preserved and valid later section removed", got)
	}
}

func TestRemoveMarkdownSectionsKeepsAmbiguousNestedMarkers(t *testing.T) {
	input := "before\n<!-- active-stack:sdd-orchestrator -->outer\n<!-- active-stack:sdd-orchestrator -->inner<!-- /active-stack:sdd-orchestrator -->\n<!-- /active-stack:sdd-orchestrator -->after"
	if got := RemoveMarkdownSectionsPreservingContent(input, "sdd-orchestrator"); got != input {
		t.Fatalf("ambiguous nested markers changed: %q", got)
	}
}
