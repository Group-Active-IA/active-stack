package external

import (
	"context"
	"fmt"

	"github.com/JuanCruzRobledo/jr-stack/internal/model"
	"github.com/JuanCruzRobledo/jr-stack/internal/system"
)

// Install installs a harness of type external by dispatching to the correct
// method based on h.External.Method. It returns a Result describing what was
// installed or merged.
func Install(
	ctx context.Context,
	h model.Harness,
	profile system.PlatformProfile,
	adapters []AgentAdapter,
	homeDir string,
) (Result, error) {
	if h.External == nil {
		return Result{}, fmt.Errorf("harness %q has no External config", h.ID)
	}

	switch h.External.Method {
	case "npm":
		return installNPM(ctx, h, profile)
	case "homebrew":
		return installHomebrew(ctx, h, profile)
	case "mcp":
		return installMCP(ctx, h, adapters, homeDir)
	default:
		return Result{}, fmt.Errorf("harness %q: unsupported install method %q (supported: npm, homebrew, mcp)", h.ID, h.External.Method)
	}
}
