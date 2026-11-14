package modelcache

import (
	"testing"
	"testing/fstest"
)

func TestFreezeLayoutPinsBytesOrderAndScope(t *testing.T) {
	files := fstest.MapFS{"one.md": {Data: []byte("one\n")}, "two.md": {Data: []byte("two\n")}}
	a, err := FreezeLayout(files, "repo-source", []string{"one.md", "two.md"})
	if err != nil {
		t.Fatal(err)
	}
	b, err := FreezeLayout(files, "repo-source", []string{"two.md", "one.md"})
	if err != nil || a.Fingerprint == b.Fingerprint {
		t.Fatal("order did not affect layout identity")
	}
	if a.ObservedClientLoad {
		t.Fatal("repository layout falsely proves client load")
	}
}

func TestFreezeLayoutMissingTraversalDuplicateAndByteChanges(t *testing.T) {
	files := fstest.MapFS{"one.md": {Data: []byte("one\n")}}
	for _, paths := range [][]string{{"missing.md"}, {"../outside"}, {"one.md", "one.md"}} {
		if _, err := FreezeLayout(files, "repo-source", paths); err == nil {
			t.Fatalf("invalid layout accepted %v", paths)
		}
	}
	a, _ := FreezeLayout(files, "repo-source", []string{"one.md"})
	files["one.md"].Data = []byte("one\r\n")
	b, _ := FreezeLayout(files, "repo-source", []string{"one.md"})
	if a.Fingerprint == b.Fingerprint {
		t.Fatal("LF/CRLF bytes were normalized")
	}
	if err := b.Validate(); err != nil {
		t.Fatal(err)
	}
	if _, err := FreezeLayout(files, "", []string{"one.md"}); err == nil {
		t.Fatal("unspecified scope accepted")
	}
}
