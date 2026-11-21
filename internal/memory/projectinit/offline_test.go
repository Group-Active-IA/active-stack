package projectinit

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// T-F7-8, T-F7-11: the package source runs only local read-only Git
// subcommands (never add, commit, push, fetch...), imports no network package
// and reads no environment variable, so it cannot read ENGRAM_* or tokens.
func TestPackageOnlyReadsGitAndTouchesNoEnvironmentOrNetwork(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	allowed := map[string]bool{"rev-parse": true, "config": true, "remote": true, "status": true, "ls-files": true, "check-ignore": true}
	call := regexp.MustCompile(`git(?:Out|OK)\(\s*[A-Za-z_.]+\s*,\s*"([^"]+)"`)
	forbidden := []string{"os.Getenv", "os.LookupEnv", `"ENGRAM_`, `"add"`, `"commit"`, `"push"`, `"fetch"`, `"pull"`, `"reset"`, `"checkout"`, `"stash"`}
	sawGit := 0
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		src, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		parsed, err := parser.ParseFile(token.NewFileSet(), file, src, parser.ImportsOnly)
		if err != nil {
			t.Fatal(err)
		}
		for _, imp := range parsed.Imports {
			p, _ := strconv.Unquote(imp.Path.Value)
			if p == "net" || strings.HasPrefix(p, "net/") && p != "net/url" {
				t.Errorf("%s imports %q", file, p)
			}
		}
		for _, f := range forbidden {
			if strings.Contains(string(src), f) {
				t.Errorf("%s contains %s", file, f)
			}
		}
		for _, m := range call.FindAllStringSubmatch(string(src), -1) {
			sawGit++
			if !allowed[m[1]] {
				t.Errorf("%s runs git %q, which is not in the local read-only list", file, m[1])
			}
		}
	}
	if sawGit == 0 {
		t.Fatal("the scan found no git calls: the pattern is stale")
	}
}
