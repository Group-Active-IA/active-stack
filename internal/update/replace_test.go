package update_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/update"
)

func TestReplaceBinaryBacksUpAndReplaces(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "active-stack.exe")
	source := filepath.Join(dir, "download")
	if err := os.WriteFile(target, []byte("old"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(source, []byte("new"), 0o755); err != nil {
		t.Fatal(err)
	}
	changed, err := update.ReplaceBinary(source, target, filepath.Join(dir, "backups"))
	if err != nil || !changed {
		t.Fatalf("changed=%v, err=%v", changed, err)
	}
	got, err := os.ReadFile(target)
	if err != nil || string(got) != "new" {
		t.Fatalf("target=%q, err=%v", got, err)
	}
	entries, err := os.ReadDir(filepath.Join(dir, "backups"))
	if err != nil || len(entries) != 1 {
		t.Fatalf("backups=%v, err=%v", entries, err)
	}
	got, err = os.ReadFile(filepath.Join(dir, "backups", entries[0].Name()))
	if err != nil || string(got) != "old" {
		t.Fatalf("backup=%q, err=%v", got, err)
	}
	changed, err = update.ReplaceBinary(source, target, filepath.Join(dir, "backups"))
	if err != nil || changed {
		t.Fatalf("second changed=%v, err=%v", changed, err)
	}
}

func TestBinaryMatchesComparesContent(t *testing.T) {
	dir := t.TempDir()
	a, b := filepath.Join(dir, "a"), filepath.Join(dir, "b")
	if err := os.WriteFile(a, []byte("same"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(b, []byte("same"), 0o755); err != nil {
		t.Fatal(err)
	}
	match, err := update.BinaryMatches(a, b)
	if err != nil || !match {
		t.Fatalf("match=%v err=%v", match, err)
	}
	if err := os.WriteFile(b, []byte("different"), 0o755); err != nil {
		t.Fatal(err)
	}
	match, err = update.BinaryMatches(a, b)
	if err != nil || match {
		t.Fatalf("match=%v err=%v", match, err)
	}
}
