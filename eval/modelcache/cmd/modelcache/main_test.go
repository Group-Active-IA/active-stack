package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) { return 0, errors.New("sink unavailable") }

func fixture(t *testing.T) ([]byte, string) {
	t.Helper()
	root, err := filepath.Abs("../../testdata/synthetic")
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(root, "dataset.json"))
	if err != nil {
		t.Fatal(err)
	}
	return b, root
}

func invoke(args []string, input []byte) (int, string, string) {
	var out, errs bytes.Buffer
	code := run(args, bytes.NewReader(input), &out, &errs)
	return code, out.String(), errs.String()
}

func TestValidateAndReportFromStdin(t *testing.T) {
	b, root := fixture(t)
	for _, command := range []string{"validate", "report"} {
		t.Run(command, func(t *testing.T) {
			code, out, errs := invoke([]string{command, "--input", "-", "--root", root}, b)
			if code != 0 || errs != "" || !json.Valid([]byte(out)) {
				t.Fatalf("code=%d out=%s err=%s", code, out, errs)
			}
			if command == "report" && !strings.Contains(out, `"unavailable"`) {
				t.Fatal("report must not authorize adoption")
			}
		})
	}
}

func TestLayoutFileAndByteChange(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "p")
	if err := os.WriteFile(path, []byte("first"), 0600); err != nil {
		t.Fatal(err)
	}
	input := filepath.Join(t.TempDir(), "layout.json")
	if err := os.WriteFile(input, []byte(`{"source_scope":"materialized-prompt","paths":["p"]}`), 0600); err != nil {
		t.Fatal(err)
	}
	code, first, errs := invoke([]string{"layout", "--input", input, "--root", root}, nil)
	if code != 0 || errs != "" || !json.Valid([]byte(first)) {
		t.Fatalf("%d %s %s", code, first, errs)
	}
	if err := os.WriteFile(path, []byte("second"), 0600); err != nil {
		t.Fatal(err)
	}
	code, second, errs := invoke([]string{"layout", "--input", input, "--root", root}, nil)
	if code != 0 || first == second {
		t.Fatalf("hash failed to track changed bytes: %d %s", code, errs)
	}
}

func TestStrictJSONAndFlags(t *testing.T) {
	b, root := fixture(t)
	for name, data := range map[string][]byte{"unknown": []byte(`{"surprise":true}`), "trailing": append(append([]byte{}, b...), []byte(" {}")...), "malformed": []byte("{")} {
		t.Run(name, func(t *testing.T) {
			code, out, errs := invoke([]string{"validate", "--root", root}, data)
			if code != 2 || out != "" || errs == "" {
				t.Fatalf("%d %s %s", code, out, errs)
			}
		})
	}
	for _, args := range [][]string{{}, {"unknown"}, {"validate", "--bad"}, {"validate", "positional"}, {"layout", "--input"}} {
		code, out, errs := invoke(args, b)
		if code != 2 || out != "" || errs == "" {
			t.Fatalf("args=%v code=%d %s %s", args, code, out, errs)
		}
	}
}

func TestLayoutsRejectChangedFilesAndEscapingPaths(t *testing.T) {
	b, root := fixture(t)
	code, out, errs := invoke([]string{"validate", "--root", t.TempDir()}, b)
	if code != 1 || out != "" || errs == "" {
		t.Fatalf("missing layout: %d %s %s", code, out, errs)
	}
	for _, p := range []string{"../outside", "/absolute"} {
		data, _ := json.Marshal(map[string]any{"source_scope": "materialized-prompt", "paths": []string{p}})
		code, out, errs := invoke([]string{"layout", "--root", root}, data)
		if code != 1 || out != "" || errs == "" {
			t.Fatalf("traversal: %d %s %s", code, out, errs)
		}
	}
	isolated := t.TempDir()
	outside := filepath.Join(t.TempDir(), "private")
	if err := os.WriteFile(outside, []byte("secret"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(isolated, "link")); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	code, out, errs = invoke([]string{"layout", "--root", isolated}, []byte(`{"source_scope":"materialized-prompt","paths":["link"]}`))
	if code != 1 || out != "" || strings.Contains(errs, "secret") {
		t.Fatalf("escape: %d %s %s", code, out, errs)
	}
}

func TestDatasetFileDeterministicReportAndReadFailures(t *testing.T) {
	b, root := fixture(t)
	input := filepath.Join(t.TempDir(), "dataset.json")
	if err := os.WriteFile(input, b, 0600); err != nil {
		t.Fatal(err)
	}
	args := []string{"report", "--input", input, "--root", root}
	code, first, errs := invoke(args, nil)
	if code != 0 || errs != "" {
		t.Fatalf("%d %s", code, errs)
	}
	_, second, _ := invoke(args, nil)
	if first != second {
		t.Fatal("same data produced different JSON")
	}
	code, out, errs := invoke([]string{"validate", "--input", input + ".missing", "--root", root}, nil)
	if code != 1 || out != "" || errs == "" {
		t.Fatalf("read failure: %d %s %s", code, out, errs)
	}
	var errOut bytes.Buffer
	if code := run([]string{"validate", "--root", root}, bytes.NewReader(b), failingWriter{}, &errOut); code != 1 || !strings.Contains(errOut.String(), "sink unavailable") {
		t.Fatalf("write failure: %d %s", code, &errOut)
	}
}

func TestNestedUnknownAndInvalidProtocol(t *testing.T) {
	b, root := fixture(t)
	var data map[string]any
	if err := json.Unmarshal(b, &data); err != nil {
		t.Fatal(err)
	}
	protocol := data["protocol"].(map[string]any)
	protocol["unrecognized"] = true
	encoded, _ := json.Marshal(data)
	code, out, errs := invoke([]string{"validate", "--root", root}, encoded)
	if code != 2 || out != "" || !strings.Contains(errs, "unknown field") {
		t.Fatalf("%d %s %s", code, out, errs)
	}
	delete(protocol, "unrecognized")
	protocol["phase"] = "adoption"
	encoded, _ = json.Marshal(data)
	code, out, errs = invoke([]string{"report", "--root", root}, encoded)
	if code != 1 || out != "" || errs == "" {
		t.Fatalf("adoption accepted: %d %s %s", code, out, errs)
	}
}

func TestRootedFilesRejectsDirectorySymlinkEscape(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	if err := os.WriteFile(filepath.Join(outside, "p"), []byte("outside"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(root, "dir")); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	code, out, errs := invoke([]string{"layout", "--root", root}, []byte(`{"source_scope":"repo-source","paths":["dir/p"]}`))
	if code != 1 || out != "" || errs == "" {
		t.Fatalf("directory escape: %d %s %s", code, out, errs)
	}
	for _, path := range []string{"../p", "p\\file", "C:/outside"} {
		files, err := rootedFiles(root)
		if err != nil {
			t.Fatal(err)
		}
		if f, err := files.Open(path); err == nil {
			f.Close()
			t.Fatalf("unsafe path opened: %s", path)
		}
	}
}
