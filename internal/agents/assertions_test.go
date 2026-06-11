// Package agents_test contains compile-time assertions that verify each
// concrete adapter satisfies every harness installer's local interface.
// These catch interface drift without requiring runtime execution.
package agents_test

import (
	claudeadapter "github.com/Group-Active-IA/active-stack/internal/agents/claude"
	opencodeadapter "github.com/Group-Active-IA/active-stack/internal/agents/opencode"
	"github.com/Group-Active-IA/active-stack/internal/agents"
	"github.com/Group-Active-IA/active-stack/internal/harness/config"
	"github.com/Group-Active-IA/active-stack/internal/harness/config/permissions"
	"github.com/Group-Active-IA/active-stack/internal/harness/external"
	"github.com/Group-Active-IA/active-stack/internal/harness/skill"
)

// Task 5.1 — claude adapter satisfies all four installer interfaces.
var _ skill.AgentAdapter = (*claudeadapter.Adapter)(nil)
var _ config.AgentAdapter = (*claudeadapter.Adapter)(nil)
var _ permissions.PermissionsAdapter = (*claudeadapter.Adapter)(nil)
var _ external.AgentAdapter = (*claudeadapter.Adapter)(nil)

// Task 5.2 — opencode adapter satisfies all four installer interfaces.
var _ skill.AgentAdapter = (*opencodeadapter.Adapter)(nil)
var _ config.AgentAdapter = (*opencodeadapter.Adapter)(nil)
var _ permissions.PermissionsAdapter = (*opencodeadapter.Adapter)(nil)
var _ external.AgentAdapter = (*opencodeadapter.Adapter)(nil)

// C-31: both adapters satisfy the full agents.Adapter interface, including the
// new CommandsDir method added in C-31 (D1). RED: fails if CommandsDir is
// missing from either adapter or from the interface.
var _ agents.Adapter = (*claudeadapter.Adapter)(nil)
var _ agents.Adapter = (*opencodeadapter.Adapter)(nil)
