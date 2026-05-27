package external

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/JuanCruzRobledo/jr-stack/internal/model"
)

// Verify runs a post-install health check appropriate to the harness method.
// For npm/homebrew it checks that the binary is in PATH.
// For mcp with HTTPS URLs it skips the check (remote server, not locally verifiable).
func Verify(ctx context.Context, h model.Harness, r Result) error {
	if h.External == nil {
		return fmt.Errorf("harness %q has no External config", h.ID)
	}

	switch h.External.Method {
	case "npm", "homebrew":
		return verifyBinary(h)
	case "mcp":
		return verifyMCP(h)
	default:
		return fmt.Errorf("verify: unsupported method %q", h.External.Method)
	}
}

func verifyBinary(h model.Harness) error {
	binaryName := filepath.Base(h.External.Pkg)
	// Strip scope prefix for npm packages like "@scope/name".
	if strings.HasPrefix(h.External.Pkg, "@") {
		parts := strings.SplitN(h.External.Pkg, "/", 2)
		if len(parts) == 2 {
			binaryName = filepath.Base(parts[1])
		}
	}

	if _, err := lookPath(binaryName); err != nil {
		return fmt.Errorf("binary %q not found in PATH after install: %w", binaryName, err)
	}
	return nil
}

func verifyMCP(h model.Harness) error {
	if strings.HasPrefix(h.External.URL, "https://") {
		// Remote HTTPS MCP server — cannot verify locally.
		return nil
	}
	// For local/HTTP MCP servers we could do a health check, but that is
	// out of scope for C-07 (no local MCP servers in the current catalog).
	return nil
}
