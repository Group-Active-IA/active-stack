package external

import (
	"context"
	"fmt"

	"github.com/Group-Active-IA/active-stack/internal/model"
	"github.com/Group-Active-IA/active-stack/internal/system"
)

// Install installs a harness of type external by dispatching to the correct
// method based on h.External.Method. It returns a Result describing what was
// installed or merged.
//
// For binary-install methods (npm, homebrew): if h.External.MCP is set, after
// the binary step succeeds the function also registers the MCP stdio server
// entry into each adapter's config (same backup+merge+atomic-write flow as
// installMCP). The returned Result merges BinaryPath from the binary step with
// ConfigFiles from the MCP registration step.
func Install(
	ctx context.Context,
	h model.Harness,
	profile system.PlatformProfile,
	adapters []AgentAdapter,
	homeDir string,
	downloadFn DownloadEventFunc,
) (Result, error) {
	if h.External == nil {
		return Result{}, fmt.Errorf("harness %q has no External config", h.ID)
	}

	switch h.External.Method {
	case "npm":
		result, err := installNPM(ctx, h, profile)
		if err != nil {
			return Result{}, err
		}
		return maybeRegisterStdioMCP(result, h, adapters, homeDir)
	case "homebrew":
		result, err := installHomebrew(ctx, h, profile, downloadFn)
		if err != nil {
			return Result{}, err
		}
		return maybeRegisterStdioMCP(result, h, adapters, homeDir)
	case "mcp":
		// Remote MCP method: handled by installMCP (URL-based). External.MCP
		// is intentionally not used here — remote MCPs have no local stdio spec.
		return installMCP(ctx, h, adapters, homeDir)
	default:
		return Result{}, fmt.Errorf("harness %q: unsupported install method %q (supported: npm, homebrew, mcp)", h.ID, h.External.Method)
	}
}

// maybeRegisterStdioMCP checks whether h.External.MCP is set. If so, it calls
// registerStdioMCP and merges the result (ConfigFiles) into the binary result
// (BinaryPath). If not set, the binary result is returned unchanged.
func maybeRegisterStdioMCP(binaryResult Result, h model.Harness, adapters []AgentAdapter, homeDir string) (Result, error) {
	if h.External.MCP == nil {
		return binaryResult, nil
	}
	mcpResult, err := registerStdioMCP(*h.External.MCP, adapters, homeDir, h.ID)
	if err != nil {
		return Result{}, fmt.Errorf("register stdio mcp for %q: %w", h.ID, err)
	}
	return Result{
		BinaryPath:       binaryResult.BinaryPath,
		ConfigFiles:      mcpResult.ConfigFiles,
		AlreadyInstalled: binaryResult.AlreadyInstalled && mcpResult.AlreadyInstalled,
	}, nil
}
