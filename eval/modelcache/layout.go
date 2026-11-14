package modelcache

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"strings"

	"github.com/Group-Active-IA/active-stack/eval/baseline/record"
)

func validFingerprint(f record.Fingerprint) bool {
	b, err := hex.DecodeString(f.Value)
	return f.Algorithm == "sha256" && err == nil && len(b) == 32
}

type FrozenFile struct {
	Path        string             `json:"path"`
	Fingerprint record.Fingerprint `json:"fingerprint"`
}

// Layout identifies ordered file bytes, not a reconstructed client prompt.
type Layout struct {
	SourceScope        string             `json:"source_scope"`
	Files              []FrozenFile       `json:"files"`
	Fingerprint        record.Fingerprint `json:"fingerprint"`
	ObservedClientLoad bool               `json:"observed_client_load"`
}

// FreezeLayout reads only the named logical paths. Input order is intentional;
// it must not be inferred from map iteration or treated as observed prompt order.
func FreezeLayout(files fs.FS, scope string, paths []string) (Layout, error) {
	if scope != "repo-source" && scope != "materialized-prompt" {
		return Layout{}, fmt.Errorf("layout scope must be explicit")
	}
	if len(paths) == 0 {
		return Layout{}, fmt.Errorf("layout has no files")
	}
	l := Layout{SourceScope: scope}
	seen := map[string]bool{}
	for _, path := range paths {
		if !fs.ValidPath(path) || strings.Contains(path, `\`) || strings.Contains(path, ":") || seen[path] {
			return Layout{}, fmt.Errorf("invalid or repeated logical layout path %s", path)
		}
		seen[path] = true
		raw, err := fs.ReadFile(files, path)
		if err != nil {
			return Layout{}, fmt.Errorf("read layout %s: %w", path, err)
		}
		l.Files = append(l.Files, FrozenFile{Path: path, Fingerprint: record.SHA256(raw)})
	}
	identity, _ := json.Marshal(struct {
		Scope string
		Files []FrozenFile
	}{scope, l.Files})
	l.Fingerprint = record.SHA256(identity)
	return l, nil
}

func (l Layout) Validate() error {
	if l.ObservedClientLoad {
		return fmt.Errorf("offline layout cannot attest client loading")
	}
	if l.SourceScope != "repo-source" && l.SourceScope != "materialized-prompt" {
		return fmt.Errorf("layout source scope missing")
	}
	if len(l.Files) == 0 {
		return fmt.Errorf("layout files missing")
	}
	seen := map[string]bool{}
	for _, f := range l.Files {
		if !fs.ValidPath(f.Path) || strings.ContainsAny(f.Path, `\:`) || seen[f.Path] || !validFingerprint(f.Fingerprint) {
			return fmt.Errorf("invalid layout file %s", f.Path)
		}
		seen[f.Path] = true
	}
	identity, _ := json.Marshal(struct {
		Scope string
		Files []FrozenFile
	}{l.SourceScope, l.Files})
	if record.SHA256(identity) != l.Fingerprint {
		return fmt.Errorf("layout identity differs from its ordered files")
	}
	return nil
}
