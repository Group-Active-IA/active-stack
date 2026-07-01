package filemerge

import (
	"fmt"
	"sort"
	"strings"

	toml "github.com/pelletier/go-toml/v2"
)

// ValidateTOML rejects malformed input before any caller writes it.
func ValidateTOML(content string) error {
	if strings.TrimSpace(content) == "" {
		return nil
	}
	var value map[string]any
	if err := toml.Unmarshal([]byte(content), &value); err != nil {
		return fmt.Errorf("invalid TOML: %w", err)
	}
	return nil
}

// UpsertTOMLTopLevel replaces selected top-level keys without touching
// same-named keys inside tables. Values are valid TOML expressions.
func UpsertTOMLTopLevel(content string, values map[string]string) (string, error) {
	if err := ValidateTOML(content); err != nil {
		return "", err
	}
	content = strings.ReplaceAll(content, "\r\n", "\n")
	lines := strings.Split(content, "\n")
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var kept []string
	inTable := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "[") {
			inTable = true
		}
		remove := false
		if !inTable {
			for _, key := range keys {
				if strings.HasPrefix(trimmed, key+" ") || strings.HasPrefix(trimmed, key+"=") {
					remove = true
					break
				}
			}
		}
		if !remove {
			kept = append(kept, line)
		}
	}

	insertAt := len(kept)
	for i, line := range kept {
		if strings.HasPrefix(strings.TrimSpace(line), "[") {
			insertAt = i
			break
		}
	}
	added := make([]string, 0, len(keys))
	for _, key := range keys {
		added = append(added, key+" = "+values[key])
	}
	out := append([]string{}, kept[:insertAt]...)
	out = append(out, added...)
	out = append(out, kept[insertAt:]...)
	result := strings.TrimSpace(strings.Join(out, "\n")) + "\n"
	if err := ValidateTOML(result); err != nil {
		return "", err
	}
	return result, nil
}

// UpsertTOMLSection replaces one table and all of its nested tables.
func UpsertTOMLSection(content, table, body string) (string, error) {
	without, err := RemoveTOMLSection(content, table)
	if err != nil {
		return "", err
	}
	block := "[" + table + "]\n" + strings.TrimSpace(body)
	result := strings.TrimSpace(without)
	if result != "" {
		result += "\n\n"
	}
	result += block + "\n"
	if err := ValidateTOML(result); err != nil {
		return "", err
	}
	return result, nil
}

// RemoveTOMLSection removes one table and its nested tables while preserving
// every unrelated table and top-level setting.
func RemoveTOMLSection(content, table string) (string, error) {
	if err := ValidateTOML(content); err != nil {
		return "", err
	}
	content = strings.ReplaceAll(content, "\r\n", "\n")
	lines := strings.Split(content, "\n")
	var kept []string
	skipping := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
			name := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(trimmed, "["), "]"))
			skipping = name == table || strings.HasPrefix(name, table+".")
		}
		if !skipping {
			kept = append(kept, line)
		}
	}
	result := strings.TrimSpace(strings.Join(kept, "\n"))
	if result != "" {
		result += "\n"
	}
	return result, nil
}

// UpsertTOMLTableValues replaces keys directly inside one table while
// preserving other keys and nested/unrelated tables.
func UpsertTOMLTableValues(content, table string, values map[string]string) (string, error) {
	if err := ValidateTOML(content); err != nil {
		return "", err
	}
	content = strings.ReplaceAll(content, "\r\n", "\n")
	lines := strings.Split(content, "\n")
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var out []string
	inTarget := false
	found := false
	flush := func() {
		for _, key := range keys {
			out = append(out, key+" = "+values[key])
		}
	}
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
			name := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(trimmed, "["), "]"))
			if inTarget {
				flush()
			}
			inTarget = name == table
			if inTarget {
				found = true
			}
			out = append(out, line)
			continue
		}
		if inTarget {
			remove := false
			for _, key := range keys {
				if strings.HasPrefix(trimmed, key+" ") || strings.HasPrefix(trimmed, key+"=") {
					remove = true
					break
				}
			}
			if remove {
				continue
			}
		}
		out = append(out, line)
	}
	if inTarget {
		flush()
	}
	if !found {
		if strings.TrimSpace(strings.Join(out, "\n")) != "" {
			out = append(out, "")
		}
		out = append(out, "["+table+"]")
		flush()
	}
	result := strings.TrimSpace(strings.Join(out, "\n")) + "\n"
	if err := ValidateTOML(result); err != nil {
		return "", err
	}
	return result, nil
}

// UpsertCodexEngramBlock removes any existing [mcp_servers.engram] block from
// the given TOML content and appends a fresh block with the canonical engram
// MCP entry (including --tools=agent). All other sections are preserved.
//
// engramCmd is the command string to use (e.g. an absolute path like
// "/usr/local/bin/engram"). If engramCmd is empty, it falls back to "engram".
func UpsertCodexEngramBlock(content, engramCmd string) string {
	if engramCmd == "" {
		engramCmd = "engram"
	}
	// Escape backslashes for TOML double-quoted strings (Windows paths).
	// e.g. C:\Users\foo → C:\\Users\\foo — prevents TOML unicode escape errors (\U).
	escapedCmd := strings.ReplaceAll(engramCmd, `\`, `\\`)
	codexEngramBlock := "[mcp_servers.engram]\ncommand = \"" + escapedCmd + "\"\nargs = [\"mcp\", \"--tools=agent\"]"
	content = strings.ReplaceAll(content, "\r\n", "\n")
	lines := strings.Split(content, "\n")

	var kept []string
	for i := 0; i < len(lines); {
		trimmed := strings.TrimSpace(lines[i])
		if trimmed == "[mcp_servers.engram]" {
			// Skip the old block header and all its key-value lines.
			i++
			for i < len(lines) {
				next := strings.TrimSpace(lines[i])
				if strings.HasPrefix(next, "[") && strings.HasSuffix(next, "]") {
					break
				}
				i++
			}
			continue
		}

		kept = append(kept, lines[i])
		i++
	}

	base := strings.TrimSpace(strings.Join(kept, "\n"))
	if base == "" {
		return codexEngramBlock + "\n"
	}

	return base + "\n\n" + codexEngramBlock + "\n"
}

// UpsertTopLevelTOMLString inserts or replaces a top-level key = "value" pair
// in TOML content. The key is placed before the first [section] header so it
// remains a top-level (non-table) setting. Existing occurrences of the key are
// removed before inserting the new value (idempotent).
func UpsertTopLevelTOMLString(content, key, value string) string {
	content = strings.ReplaceAll(content, "\r\n", "\n")
	lines := strings.Split(content, "\n")
	lineValue := fmt.Sprintf("%s = %q", key, value)

	// Remove all existing occurrences of the key.
	var cleaned []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, key+" ") || strings.HasPrefix(trimmed, key+"=") {
			continue
		}
		cleaned = append(cleaned, line)
	}

	// Find insertion point: before the first [section] header.
	insertAt := len(cleaned)
	for i, line := range cleaned {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
			insertAt = i
			break
		}
	}

	var out []string
	out = append(out, cleaned[:insertAt]...)
	out = append(out, lineValue)
	out = append(out, cleaned[insertAt:]...)

	return strings.TrimSpace(strings.Join(out, "\n")) + "\n"
}
