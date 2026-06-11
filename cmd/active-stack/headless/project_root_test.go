// Package headless — tests for C-29 ResolveProjectRoot (Task 3.1 RED).
package headless_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Group-Active-IA/active-stack/cmd/active-stack/headless"
)

// TestResolveProjectRoot covers the D2 rules for project root resolution:
//   - A relative path is absolutized via filepath.Abs.
//   - An existing path is returned without error.
//   - A non-existent path is rejected with an error (never created).
//
// RED: fails because ResolveProjectRoot does not exist yet.
func TestResolveProjectRoot(t *testing.T) {
	// Create a real temp dir to use as an "existing" project root.
	existingDir := t.TempDir()

	tests := []struct {
		name    string
		input   string
		wantErr bool
		// wantAbsolute, when true, asserts the returned path is absolute.
		wantAbsolute bool
		// wantPath, when non-empty, asserts the exact returned path.
		wantPath string
	}{
		{
			name:         "existing absolute path is accepted",
			input:        existingDir,
			wantAbsolute: true,
			wantPath:     existingDir,
		},
		{
			name:         "relative path is absolutized",
			input:        ".",
			wantAbsolute: true,
			// We can't assert the exact cwd value, just that it's absolute and exists.
		},
		{
			name:    "non-existent path is rejected",
			input:   filepath.Join(existingDir, "does-not-exist-subfolder"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := headless.ResolveProjectRoot(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ResolveProjectRoot(%q) expected error, got nil", tt.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("ResolveProjectRoot(%q) unexpected error: %v", tt.input, err)
			}
			if tt.wantAbsolute && !filepath.IsAbs(got) {
				t.Errorf("ResolveProjectRoot(%q) = %q, want absolute path", tt.input, got)
			}
			if tt.wantPath != "" && got != tt.wantPath {
				t.Errorf("ResolveProjectRoot(%q) = %q, want %q", tt.input, got, tt.wantPath)
			}
		})
	}
}

// TestResolveProjectRoot_NeverCreatesDir asserts that ResolveProjectRoot does
// NOT create the directory when it does not exist.
func TestResolveProjectRoot_NeverCreatesDir(t *testing.T) {
	base := t.TempDir()
	nonExistent := filepath.Join(base, "new-project")

	_, err := headless.ResolveProjectRoot(nonExistent)
	if err == nil {
		t.Fatal("expected error for non-existent path, got nil")
	}

	// The dir must NOT have been created.
	if _, statErr := os.Stat(nonExistent); !os.IsNotExist(statErr) {
		t.Errorf("ResolveProjectRoot must NOT create the directory; dir exists at %q", nonExistent)
	}
}
