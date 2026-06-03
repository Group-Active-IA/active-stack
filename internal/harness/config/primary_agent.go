package config

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/JuanCruzRobledo/jr-stack/internal/filemerge"
)

// primaryAgentDescription is the human-facing label shown for the orchestrator
// in the host TUI's agent picker. Kept constant so re-installs stay idempotent.
const primaryAgentDescription = "JR Stack SDD Orchestrator — coordina sub-agentes; nunca ejecuta trabajo inline."

// primaryAgentEntry is the fixed schema written under agent.<id> in the
// settings JSON for a primary-agent config harness. mode:primary is what makes
// the host TUI expose it as a tab-able agent.
type primaryAgentEntry struct {
	Mode        string          `json:"mode"`
	Description string          `json:"description"`
	Prompt      string          `json:"prompt"`
	Tools       map[string]bool `json:"tools"`
}

// InjectPrimaryAgent registers the composed orchestrator block as a primary
// agent inside the settings JSON at settingsPath, and purges any stale copy of
// the same section from the instructions file at instrPath.
//
// Why both files: agents like OpenCode only treat a block as a tab-able primary
// agent when it lives under agent.<id> in opencode.json. Injecting it into the
// shared AGENTS.md instead leaks the orchestrator into every agent (plan/build)
// and registers no new tab. So this path WRITES to the settings JSON and CLEANS
// the instructions file of any orchestrator section a previous (buggy) install
// may have left behind.
//
// Flow:
//  1. Empty settingsPath → skip (no-op, no error).
//  2. Purge the orchestrator + stale sections from the instructions file
//     (backup-first; no-op if the file does not exist).
//  3. Backup the settings file BEFORE any write (if it exists).
//  4. Deep-merge the primary-agent overlay into the settings JSON, preserving
//     all existing user config, and write atomically (skips if byte-identical).
func InjectPrimaryAgent(agentID, composed, instrPath, settingsPath, snapshotDir string) (InjectResult, error) {
	if settingsPath == "" {
		return InjectResult{}, nil
	}

	// Step 1-2: clean the instructions file so the orchestrator no longer leaks
	// into every agent. Best-effort against a file that may not exist.
	if err := purgeInstructionsSection(instrPath, agentID, snapshotDir); err != nil {
		return InjectResult{}, err
	}

	// Step 3: build the overlay { "agent": { "<id>": { … } } }. json.Marshal
	// handles all escaping of the composed markdown into a JSON string.
	entry := primaryAgentEntry{
		Mode:        "primary",
		Description: primaryAgentDescription,
		Prompt:      composed,
		Tools:       map[string]bool{"read": true, "write": true, "edit": true, "bash": true},
	}
	overlay, err := json.Marshal(map[string]any{
		"agent": map[string]any{agentID: entry},
	})
	if err != nil {
		return InjectResult{}, fmt.Errorf("marshal primary agent overlay: %w", err)
	}

	// Step 4a: backup settings before any write.
	if _, err := os.Stat(settingsPath); err == nil {
		if err := SnapshotterCreate(snapshotDir, []string{settingsPath}); err != nil {
			return InjectResult{}, fmt.Errorf("backup %q before injection: %w", settingsPath, err)
		}
	}

	// Step 4b: deep-merge into existing settings (preserves user config).
	base := readFileBytesOrNil(settingsPath)
	merged, err := filemerge.MergeJSONObjects(base, overlay)
	if err != nil {
		return InjectResult{}, fmt.Errorf("merge primary agent into %q: %w", settingsPath, err)
	}

	wr, err := filemerge.WriteFileAtomic(settingsPath, merged, 0o644)
	if err != nil {
		return InjectResult{}, fmt.Errorf("write %q: %w", settingsPath, err)
	}

	return InjectResult{Changed: wr.Changed, Created: wr.Created}, nil
}

// purgeInstructionsSection removes the named orchestrator section (and any
// stale jr-stack sections from older layouts) from the instructions file, so a
// primary-agent install leaves no orchestrator content leaking through the
// shared instructions. It backs up the file first and is a no-op when the file
// does not exist or already has no jr-stack sections.
func purgeInstructionsSection(instrPath, sectionID, snapshotDir string) error {
	if instrPath == "" {
		return nil
	}
	existing, err := os.ReadFile(instrPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read instructions %q: %w", instrPath, err)
	}

	updated := filemerge.InjectMarkdownSection(string(existing), sectionID, "")
	updated = PurgeStaleSections(updated)
	if updated == string(existing) {
		return nil // nothing to clean
	}

	// Back up before rewriting (the section being removed is captured here).
	if err := SnapshotterCreate(snapshotDir, []string{instrPath}); err != nil {
		return fmt.Errorf("backup %q before purge: %w", instrPath, err)
	}
	if _, err := filemerge.WriteFileAtomic(instrPath, []byte(updated), 0o644); err != nil {
		return fmt.Errorf("write %q: %w", instrPath, err)
	}
	return nil
}

// readFileBytesOrNil reads a file's bytes, returning nil if it does not exist.
func readFileBytesOrNil(path string) []byte {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	return data
}
