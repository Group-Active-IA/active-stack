package memory

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// T-F4-10: the package source imports no network package and reads no
// environment variable (so no ENGRAM_*). Git runs only local subcommands.
func TestPackageIsOfflineAndIgnoresEnvironment(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	forbiddenCalls := []string{"os.Getenv", "os.LookupEnv", "os.Environ", `"ENGRAM_`}
	allowedGit := map[string]bool{"rev-parse": true, "branch": true, "remote": true, "symbolic-ref": true, "config": true, "show-ref": true}
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
			path, _ := strconv.Unquote(imp.Path.Value)
			if path == "net" || strings.HasPrefix(path, "net/") {
				t.Errorf("%s imports %q", file, path)
			}
		}
		for _, call := range forbiddenCalls {
			if strings.Contains(string(src), call) {
				t.Errorf("%s uses %s", file, call)
			}
		}
		for _, sub := range gitSubcommands(string(src)) {
			if !allowedGit[sub] {
				t.Errorf("%s runs git %q, which is not in the local read-only list", file, sub)
			}
		}
	}
}

// gitSubcommands lists the first argument of every g.run(...) call.
func gitSubcommands(src string) []string {
	var subs []string
	for _, part := range strings.Split(src, "g.run(\"")[1:] {
		subs = append(subs, part[:strings.Index(part, "\"")])
	}
	return subs
}
