package model

// PermissionTier defines the level of autonomy granted to the AI agent.
// Three tiers exist; the zero-value normalizes to TierBalanceado (secure-by-default).
type PermissionTier string

const (
	// TierEstricto: defaultMode = "default", no allow-list. Agent must ask for
	// every operation. Highest friction, highest security.
	TierEstricto PermissionTier = "estricto"
	// TierBalanceado (default): defaultMode = "default" with a curated allow-list
	// for safe, repetitive operations (read, edit, go test, go build, git status,
	// git diff). The recommended starting point.
	TierBalanceado PermissionTier = "balanceado"
	// TierBypass: defaultMode = "bypassPermissions" — full autonomy opt-in.
	// The security floor deny-list still applies (C-21).
	TierBypass PermissionTier = "bypass"
)

// IsValid reports whether t is one of the three known permission tiers.
func (t PermissionTier) IsValid() bool {
	switch t {
	case TierEstricto, TierBalanceado, TierBypass:
		return true
	default:
		return false
	}
}

// DefaultPermissionTier returns TierBalanceado, the secure-by-default tier.
// The zero-value ("") should always be normalized to this via Normalize().
func DefaultPermissionTier() PermissionTier {
	return TierBalanceado
}

// Normalize returns the tier itself if valid, or TierBalanceado for the
// zero-value (""). Unknown non-empty values are left unchanged (callers must
// validate separately with IsValid if they need strict rejection).
func (t PermissionTier) Normalize() PermissionTier {
	if t == "" {
		return TierBalanceado
	}
	return t
}
