package external

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/JuanCruzRobledo/jr-stack/internal/backup"
	"github.com/JuanCruzRobledo/jr-stack/internal/filemerge"
	"github.com/JuanCruzRobledo/jr-stack/internal/model"
)

// snapshotterCreate is replaceable in tests to avoid real filesystem backups.
var snapshotterCreate = func(snapshotDir string, paths []string) error {
	s := backup.NewSnapshotter()
	_, err := s.Create(snapshotDir, paths)
	return err
}

func installMCP(
	ctx context.Context,
	h model.Harness,
	adapters []AgentAdapter,
	homeDir string,
) (Result, error) {
	if len(adapters) == 0 {
		return Result{AlreadyInstalled: true}, nil
	}

	var configFiles []string
	allAlready := true

	for _, adapter := range adapters {
		configPath := adapter.MCPConfigPath(homeDir, h.ID)
		if configPath == "" {
			continue
		}

		// Backup existing file before touching it.
		if _, err := os.Stat(configPath); err == nil {
			snapshotDir := filepath.Join(homeDir, ".jr-stack", "backups",
				fmt.Sprintf("%s-%s", h.ID, adapter.Agent()))
			if err := snapshotterCreate(snapshotDir, []string{configPath}); err != nil {
				return Result{}, fmt.Errorf("backup %q before mcp injection: %w", configPath, err)
			}
		}

		var overlay map[string]any
		if h.External.MCP != nil && adapter.Agent() != model.AgentOpenCode {
			// Claude (and any agent that is not OpenCode): the remote endpoint of
			// this MCP requires authentication → prefer the local stdio server
			// (npx), which needs no auth.
			overlay = buildStdioOverlay(*h.External.MCP, adapter)
		} else {
			overlay = buildOverlay(h, adapter)
		}
		overlayJSON, err := json.Marshal(overlay)
		if err != nil {
			return Result{}, fmt.Errorf("marshal mcp overlay for %s: %w", adapter.Agent(), err)
		}

		baseJSON := readExistingJSON(configPath)

		merged, err := filemerge.MergeJSONObjects(baseJSON, overlayJSON)
		if err != nil {
			return Result{}, fmt.Errorf("merge mcp config for %s: %w", adapter.Agent(), err)
		}

		wr, err := filemerge.WriteFileAtomic(configPath, merged, 0o644)
		if err != nil {
			return Result{}, fmt.Errorf("write mcp config %q: %w", configPath, err)
		}

		if wr.Changed {
			allAlready = false
			configFiles = append(configFiles, configPath)
		}
	}

	return Result{
		ConfigFiles:      configFiles,
		AlreadyInstalled: allAlready && len(configFiles) == 0,
	}, nil
}

// buildOverlay constructs the JSON overlay map from catalog fields without
// any hardcoded server-specific constants.
func buildOverlay(h model.Harness, adapter AgentAdapter) map[string]any {
	serverURL := h.External.URL
	// Append /mcp suffix when the URL doesn't already include a path component.
	mcpURL := strings.TrimRight(serverURL, "/") + "/mcp"

	switch adapter.MCPStrategy() {
	case StrategyMergeIntoSettings:
		if adapter.Agent() == model.AgentOpenCode {
			// OpenCode uses the "mcp" top-level key with remote entry format.
			return map[string]any{
				"mcp": map[string]any{
					h.ID: map[string]any{
						"type":    "remote",
						"url":     mcpURL,
						"enabled": true,
					},
				},
			}
		}
		// Generic merge-into-settings: standard mcpServers key with type discriminator.
		// Claude Code (and any future generic agent) requires "type":"http" to
		// recognise the entry as a remote HTTP MCP server in ~/.claude.json.
		return map[string]any{
			"mcpServers": map[string]any{
				h.ID: map[string]any{
					"type": "http",
					"url":  mcpURL,
				},
			},
		}

	case StrategySeparateFile:
		// Standalone server file: the file IS the server config object.
		return map[string]any{
			"url": mcpURL,
		}

	default:
		return map[string]any{
			"mcpServers": map[string]any{
				h.ID: map[string]any{
					"url": mcpURL,
				},
			},
		}
	}
}

// buildMCPOverlay constructs the JSON overlay map for writing a local (stdio)
// MCP server entry into a Claude project's .mcp.json file (D4).
//
// The overlay shape is:
//
//	{"mcpServers": {"<MCP.Name>": {"command": ..., "args": [...], "env": {...}}}}
//
// This is the overlay for the single-file Claude project strategy
// (MCPStrategySingleFileMerge). The existing installMCP flow then backs up,
// merges via filemerge.MergeJSONObjects, and writes atomically.
//
// No hardcoded server constants — the overlay key is always mcp.Name.
// The "env" key is omitted when mcp.Env is nil (no spurious empty map).
func buildMCPOverlay(mcp model.MCP) map[string]any {
	entry := map[string]any{
		"command": mcp.Command,
		"args":    mcp.Args,
	}
	if len(mcp.Env) > 0 {
		entry["env"] = mcp.Env
	}
	return map[string]any{
		"mcpServers": map[string]any{
			mcp.Name: entry,
		},
	}
}

// buildStdioOverlay constructs the JSON overlay map for registering a local
// (stdio) MCP server into a machine-target agent config. Unlike buildOverlay
// (which handles remote URL servers) and buildMCPOverlay (which handles the
// Claude project single-file strategy), this function produces the per-strategy
// shape for machine-target registration after a binary install.
//
// Shape by strategy:
//
//   - StrategySeparateFile:
//     {"command":..., "args":[...], "env":{...}}  ← bare server object; the FILE is the server
//
//   - StrategyMergeIntoSettings, agent == OpenCode:
//     {"mcp": {"<name>": {"type":"local","command":...,"args":[...],"enabled":true}}}
//
//   - StrategyMergeIntoSettings (generic, e.g. Claude → ~/.claude.json):
//     {"mcpServers": {"<name>": {"type":"stdio","command":...,"args":[...]}}}
//
// The "env" key is omitted when mcp.Env is nil (no spurious empty map).
// No hardcoded server-name constants — the overlay key is always mcp.Name.
func buildStdioOverlay(mcp model.MCP, adapter AgentAdapter) map[string]any {
	entry := map[string]any{
		"command": mcp.Command,
		"args":    mcp.Args,
	}
	if len(mcp.Env) > 0 {
		entry["env"] = mcp.Env
	}

	switch adapter.MCPStrategy() {
	case StrategyMergeIntoSettings:
		if adapter.Agent() == model.AgentOpenCode {
			// OpenCode local server format: type:"local" + enabled flag.
			localEntry := map[string]any{
				"type":    "local",
				"command": mcp.Command,
				"args":    mcp.Args,
				"enabled": true,
			}
			if len(mcp.Env) > 0 {
				localEntry["env"] = mcp.Env
			}
			// Wrap in __replace__ so any stale remote entry (e.g. {"type":"remote","url":...})
			// is fully discarded on reinstall — a deep merge would leave orphan keys.
			return map[string]any{
				"mcp": map[string]any{
					mcp.Name: map[string]any{"__replace__": localEntry},
				},
			}
		}
		// Generic merge-into-settings (Claude → ~/.claude.json): replace the whole
		// server object atomically so a stale remote entry (e.g. {"type":"http","url":...})
		// is fully discarded on reinstall — a deep merge would leave the orphan "url".
		entry["type"] = "stdio"
		return map[string]any{
			"mcpServers": map[string]any{
				mcp.Name: map[string]any{"__replace__": entry},
			},
		}

	case StrategySeparateFile:
		// Standalone server file: the file IS the server config object.
		return entry

	default:
		return map[string]any{
			"mcpServers": map[string]any{
				mcp.Name: entry,
			},
		}
	}
}

// registerStdioMCP registers a local (stdio) MCP server into each adapter's
// machine-target config path. It mirrors the installMCP flow (backup + JSON
// merge + atomic write) but uses buildStdioOverlay instead of buildOverlay so
// the resulting overlay is a stdio server entry rather than a remote URL entry.
//
// harnessID is used only to build a per-agent backup directory name.
//
// Returns a Result with ConfigFiles listing paths that were written or changed,
// and AlreadyInstalled=true when every adapter's config was already up-to-date.
//
// Governance ALTO: backup is taken before touching any existing file. Reuses
// snapshotterCreate (replaceable in tests) and filemerge.MergeJSONObjects
// + filemerge.WriteFileAtomic for idempotent merge, exactly like installMCP.
func registerStdioMCP(
	mcp model.MCP,
	adapters []AgentAdapter,
	homeDir string,
	harnessID string,
) (Result, error) {
	if len(adapters) == 0 {
		return Result{AlreadyInstalled: true}, nil
	}

	var configFiles []string
	allAlready := true

	for _, adapter := range adapters {
		configPath := adapter.MCPConfigPath(homeDir, mcp.Name)
		if configPath == "" {
			continue
		}

		// Backup existing file before touching it (governance ALTO).
		if _, err := os.Stat(configPath); err == nil {
			snapshotDir := filepath.Join(homeDir, ".jr-stack", "backups",
				fmt.Sprintf("%s-%s-mcp", harnessID, adapter.Agent()))
			if err := snapshotterCreate(snapshotDir, []string{configPath}); err != nil {
				return Result{}, fmt.Errorf("backup %q before stdio mcp injection: %w", configPath, err)
			}
		}

		overlay := buildStdioOverlay(mcp, adapter)
		overlayJSON, err := json.Marshal(overlay)
		if err != nil {
			return Result{}, fmt.Errorf("marshal stdio mcp overlay for %s: %w", adapter.Agent(), err)
		}

		baseJSON := readExistingJSON(configPath)

		merged, err := filemerge.MergeJSONObjects(baseJSON, overlayJSON)
		if err != nil {
			return Result{}, fmt.Errorf("merge stdio mcp config for %s: %w", adapter.Agent(), err)
		}

		wr, err := filemerge.WriteFileAtomic(configPath, merged, 0o644)
		if err != nil {
			return Result{}, fmt.Errorf("write stdio mcp config %q: %w", configPath, err)
		}

		if wr.Changed {
			allAlready = false
			configFiles = append(configFiles, configPath)
		}
	}

	return Result{
		ConfigFiles:      configFiles,
		AlreadyInstalled: allAlready && len(configFiles) == 0,
	}, nil
}

// WriteMCPProjectEntry writes a local (stdio) MCP server entry into the given
// config file path using the resolved project strategy. It backs up the file
// if it already exists, then merges the new entry idempotently and writes
// atomically.
//
// This is the write-path for the Claude project single-file strategy (D4, D5).
// It reuses the same backup + MergeJSONObjects + WriteFileAtomic flow as
// the legacy installMCP, so governance constraints are automatically satisfied.
//
// The strategy parameter currently only supports MCPStrategySingleFileMerge
// (the Claude project case). Other strategies are not yet wired here.
//
// Returns (changed bool, err error). changed is true when the file was written
// or updated, false when the entry was already present (idempotent re-run).
func WriteMCPProjectEntry(
	mcp model.MCP,
	configPath string,
	snapshotDir string,
) (bool, error) {
	// Backup existing file before touching it (governance ALTO).
	if _, err := os.Stat(configPath); err == nil {
		if err := snapshotterCreate(snapshotDir, []string{configPath}); err != nil {
			return false, fmt.Errorf("backup %q before mcp write: %w", configPath, err)
		}
	}

	overlay := buildMCPOverlay(mcp)
	overlayJSON, err := json.Marshal(overlay)
	if err != nil {
		return false, fmt.Errorf("marshal mcp overlay for %q: %w", mcp.Name, err)
	}

	base := readExistingJSON(configPath)
	merged, err := filemerge.MergeJSONObjects(base, overlayJSON)
	if err != nil {
		return false, fmt.Errorf("merge mcp config for %q: %w", mcp.Name, err)
	}

	wr, err := filemerge.WriteFileAtomic(configPath, merged, 0o644)
	if err != nil {
		return false, fmt.Errorf("write mcp config %q: %w", configPath, err)
	}
	return wr.Changed, nil
}

// readExistingJSON reads a JSON file, returning nil if it doesn't exist.
func readExistingJSON(path string) []byte {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	return data
}
