package model

import "testing"

// TestPermissionTierValues verifies exactly three valid tiers exist and that an
// unknown value is rejected (spec: "Existen exactamente tres tiers válidos",
// "Un tier desconocido se rechaza").
func TestPermissionTierValues(t *testing.T) {
	valid := []PermissionTier{TierEstricto, TierBalanceado, TierBypass}
	if len(valid) != 3 {
		t.Fatalf("expected exactly 3 tiers, got %d", len(valid))
	}

	for _, tier := range valid {
		if !tier.IsValid() {
			t.Errorf("IsValid() = false for %q, expected true", tier)
		}
	}

	// Unknown values must be rejected.
	unknown := []PermissionTier{"unknown", "", "BYPASS", "Balanceado"}
	for _, tier := range unknown {
		if tier.IsValid() {
			t.Errorf("IsValid() = true for %q, expected false", tier)
		}
	}
}

// TestDefaultPermissionTier verifies the default tier is balanceado
// (spec: "El tier por defecto es balanceado").
func TestDefaultPermissionTier(t *testing.T) {
	got := DefaultPermissionTier()
	if got != TierBalanceado {
		t.Errorf("DefaultPermissionTier() = %q, want %q", got, TierBalanceado)
	}
}

// TestPermissionTierNormalize verifies zero-value normalization to balanceado
// and that known/unknown values pass through (spec: "El zero-value se normaliza a balanceado").
func TestPermissionTierNormalize(t *testing.T) {
	tests := []struct {
		name  string
		input PermissionTier
		want  PermissionTier
	}{
		// Zero-value MUST normalize to balanceado — never bypass.
		{"empty string normalizes to balanceado", "", TierBalanceado},
		// Valid tiers pass through unchanged.
		{"balanceado stays balanceado", TierBalanceado, TierBalanceado},
		{"bypass stays bypass (no secret normalization)", TierBypass, TierBypass},
		{"estricto stays estricto", TierEstricto, TierEstricto},
		// Unknown non-empty values pass through (caller validates with IsValid).
		{"unknown passes through", PermissionTier("unknown"), PermissionTier("unknown")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.input.Normalize()
			if got != tt.want {
				t.Errorf("Normalize() = %q, want %q", got, tt.want)
			}
		})
	}
}
