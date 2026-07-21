package external

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Group-Active-IA/active-stack/internal/model"
	"github.com/Group-Active-IA/active-stack/internal/system"
)

// cmdRunner abstracts os/exec so tests can inject a fake without running real commands.
type cmdRunner func(ctx context.Context, name string, args ...string) ([]byte, error)

// realCmdRunner runs name with args and returns combined stdout+stderr.
func realCmdRunner(ctx context.Context, name string, args ...string) ([]byte, error) {
	return exec.CommandContext(ctx, name, args...).CombinedOutput()
}

// runner is the package-level runner; replaced in tests.
var runner cmdRunner = realCmdRunner

// lookPath is a package-level var so tests can replace it.
var lookPath = exec.LookPath

func installNPM(ctx context.Context, h model.Harness, profile system.PlatformProfile) (Result, error) {
	pkg := h.External.Pkg
	if pkg == "" {
		return Result{}, fmt.Errorf("harness %q: npm method requires External.Pkg", h.ID)
	}

	var (
		name string
		args []string
	)

	// --ignore-scripts: npm packages can ship postinstall hooks that spawn
	// "node" through a lifecycle-script shell npm builds itself; on some
	// Windows setups that shell fails to resolve "node" on PATH even though
	// the interactive shell resolves it fine, turning an optional postinstall
	// script (e.g. openspec's shell-completion tip) into a hard install
	// failure. We never depend on install-time lifecycle scripts.
	if useSudo(profile) {
		name = "sudo"
		args = []string{"npm", "install", "-g", "--ignore-scripts", pkg}
	} else {
		name = "npm"
		args = []string{"install", "-g", "--ignore-scripts", pkg}
	}

	if out, err := runner(ctx, name, args...); err != nil {
		return Result{}, fmt.Errorf("npm install -g %s: %w\n%s", pkg, err, out)
	}

	binaryName := pkgBinaryName(pkg)
	binaryPath, _ := lookPath(binaryName)

	return Result{BinaryPath: binaryPath}, nil
}

// useSudo returns true when the install should be run under sudo.
// Windows never uses sudo; other platforms skip it when npm prefix is user-writable.
func useSudo(profile system.PlatformProfile) bool {
	if profile.OS == "windows" {
		return false
	}
	return !profile.NpmWritable
}

// pkgBinaryName extracts the binary name from an npm package identifier.
// "@scope/name" → "name", "plain" → "plain".
func pkgBinaryName(pkg string) string {
	if strings.HasPrefix(pkg, "@") {
		parts := strings.SplitN(pkg, "/", 2)
		if len(parts) == 2 {
			return filepath.Base(parts[1])
		}
	}
	return filepath.Base(pkg)
}
