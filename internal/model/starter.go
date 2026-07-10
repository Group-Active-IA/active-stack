package model

import "fmt"

// MCP represents a project-scope MCP entry in a starter.
//
// The shape covers the local (stdio) server entry needed to emit a Claude
// .mcp.json record: Name (the server key), Command, Args, Env.
// Name is first for C-26 back-compat (YAML key order + zero-value compat).
//
// Decisions made in C-28:
//   - Name is the server key under mcpServers in .mcp.json.
//   - Command, Args, Env carry the stdio entry (fully typed, self-documenting).
//   - MCPStrategy enum lives here in internal/model (not in external) to avoid
//     the model↔external import cycle (D1a). external.MCPStrategy is kept for
//     the legacy machine-target adapter.MCPStrategy() method.
//
// (TBD) Remote/HTTP/SSE transport shape (type:"http", url, etc.) is deferred.
// A future Transport/URL field can be added additively without breaking this struct.
type MCP struct {
	Name    string            `yaml:"name"`
	Command string            `yaml:"command,omitempty"`
	Args    []string          `yaml:"args,omitempty"`
	Env     map[string]string `yaml:"env,omitempty"`
}

// Starter is a curated, composable bundle of harnesses and project-scope MCPs.
// It is a distinct domain concept from Harness: while a Harness is an
// installable module, a Starter is a curation layer (think Spring Boot starters)
// that groups harnesses for a specific project domain.
//
// Fields:
//   - ID:          unique identifier within the catalog (e.g. "active-ia").
//   - Name:        human-readable display name.
//   - Description: optional longer description.
//   - Harnesses:   list of harness IDs from the catalog that this starter directly bundles.
//   - Includes:    list of other starter IDs this starter composes (composable via includes).
//   - MCPs:        project-scope MCP entries. Each MCP carries a local (stdio) server entry
//                  written to the correct per-agent, per-target path (C-28).
//                  (TBD) Remote/HTTP/SSE transport shape deferred — see MCP comment.
type Starter struct {
	ID          string        `yaml:"id"`
	Name        string        `yaml:"name"`
	Description LocalizedText `yaml:"description,omitempty"`
	// LongDescription is mandatory (both languages) for every starter (D2,
	// design.md catalog-localized-descriptions).
	LongDescription LocalizedText `yaml:"long_description,omitempty"`
	Harnesses       []string      `yaml:"harnesses,omitempty"`
	Includes        []string      `yaml:"includes,omitempty"`
	MCPs            []MCP         `yaml:"mcps,omitempty"`
}

// Validate performs field-level validation on this MCP's own fields.
// Name and Command are required for a local (stdio) entry.
// It does NOT validate that the Command binary exists on disk (that is the
// installer's responsibility). Returns an error naming the offending field.
func (m MCP) Validate() error {
	if m.Name == "" {
		return fmt.Errorf("mcp: missing required field 'name'")
	}
	if m.Command == "" {
		return fmt.Errorf("mcp %q: missing required field 'command'", m.Name)
	}
	return nil
}

// Validate performs field-level validation on this Starter's own fields.
// It does NOT validate references to harnesses or other starters in the catalog
// (that is the catalog's responsibility). Returns an error naming the offending
// field when validation fails.
func (s Starter) Validate() error {
	if s.ID == "" {
		return fmt.Errorf("starter: missing required field 'id'")
	}
	if s.Name == "" {
		return fmt.Errorf("starter %q: missing required field 'name'", s.ID)
	}
	return nil
}
