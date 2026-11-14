// Command modelcache checks and summarizes local pilot data. It never runs a client.
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/Group-Active-IA/active-stack/eval/modelcache"
)

func main() { os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr)) }

func run(args []string, in io.Reader, out, errOut io.Writer) int {
	fail := func(code int, err error) int { fmt.Fprintln(errOut, err); return code }
	if len(args) == 0 || (args[0] != "validate" && args[0] != "report" && args[0] != "layout") {
		return fail(2, errors.New("usage: modelcache validate|report|layout [--input FILE|-] [--root DIRECTORY]"))
	}
	flags := flag.NewFlagSet(args[0], flag.ContinueOnError)
	flags.SetOutput(errOut)
	inputPath := flags.String("input", "-", "local JSON file, or - for stdin")
	root := flags.String("root", ".", "read-only root for layout files")
	if err := flags.Parse(args[1:]); err != nil {
		return 2
	}
	if flags.NArg() != 0 {
		return fail(2, errors.New("unexpected positional arguments"))
	}
	if *inputPath != "-" {
		file, err := os.Open(*inputPath)
		if err != nil {
			return fail(1, fmt.Errorf("read input: %w", err))
		}
		defer file.Close()
		in = file
	}
	files, err := rootedFiles(*root)
	if err != nil {
		return fail(1, err)
	}
	var result any
	if args[0] == "layout" {
		var request struct {
			SourceScope string   `json:"source_scope"`
			Paths       []string `json:"paths"`
		}
		if err := decode(in, &request); err != nil {
			return fail(2, err)
		}
		result, err = modelcache.FreezeLayout(files, request.SourceScope, request.Paths)
	} else {
		var dataset modelcache.Dataset
		if err := decode(in, &dataset); err != nil {
			return fail(2, err)
		}
		if err := dataset.Protocol.Validate(); err != nil {
			return fail(1, err)
		}
		if err := modelcache.ValidateLayouts(dataset.Protocol, files); err != nil {
			return fail(1, err)
		}
		if args[0] == "report" {
			result, err = modelcache.Analyze(dataset)
		} else {
			result = struct {
				Valid bool   `json:"valid"`
				Scope string `json:"scope"`
			}{true, "offline protocol and layout validation; no execution or provenance verified"}
		}
	}
	if err != nil {
		return fail(1, err)
	}
	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(result); err != nil {
		return fail(1, fmt.Errorf("write result: %w", err))
	}
	return 0
}

func decode(in io.Reader, target any) error {
	d := json.NewDecoder(in)
	d.DisallowUnknownFields()
	if err := d.Decode(target); err != nil {
		return fmt.Errorf("decode input: %w", err)
	}
	var extra any
	if err := d.Decode(&extra); err != io.EOF {
		if err == nil {
			return errors.New("input must contain exactly one JSON value")
		}
		return fmt.Errorf("trailing input: %w", err)
	}
	return nil
}

// localFiles opens only resolved files inside its canonical root; it has no write API.
type localFiles struct{ root string }

func rootedFiles(root string) (localFiles, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return localFiles{}, err
	}
	resolved, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return localFiles{}, fmt.Errorf("resolve root: %w", err)
	}
	info, err := os.Stat(resolved)
	if err != nil {
		return localFiles{}, err
	}
	if !info.IsDir() {
		return localFiles{}, errors.New("layout root must be a directory")
	}
	return localFiles{resolved}, nil
}

func (l localFiles) Open(name string) (fs.File, error) {
	if !fs.ValidPath(name) || strings.Contains(name, "\\") || strings.Contains(name, ":") {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrInvalid}
	}
	resolved, err := filepath.EvalSymlinks(filepath.Join(l.root, filepath.FromSlash(name)))
	if err != nil {
		return nil, err
	}
	rel, err := filepath.Rel(l.root, resolved)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrPermission}
	}
	return os.Open(resolved)
}
