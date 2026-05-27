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

		overlay := buildOverlay(h, adapter)
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
		// Generic merge-into-settings: standard mcpServers key.
		return map[string]any{
			"mcpServers": map[string]any{
				h.ID: map[string]any{
					"url": mcpURL,
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

// readExistingJSON reads a JSON file, returning nil if it doesn't exist.
func readExistingJSON(path string) []byte {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	return data
}
