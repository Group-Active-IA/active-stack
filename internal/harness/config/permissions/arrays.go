package permissions

import "encoding/json"

// filemerge.MergeJSONObjects replaces arrays: an overlay `permissions.allow` or
// `permissions.deny` would discard the entries the user already has. R-18 keeps
// filemerge untouched and unions those two arrays here, before the merge: the
// computed union goes into the overlay, which then replaces the base array with
// something that still contains the user's entries.
//
// Ownership rules (no markers exist inside a JSON array, so ownership is by value):
//   - deny: the user's entries are kept; the stack's (floor + push-to-main) are
//     added if missing. They are never removed.
//   - allow: the user's entries are kept; the entries of the current tier are
//     added if missing; entries the stack emits for OTHER tiers are removed, so
//     switching to a stricter tier does not leave stale permissions. Limitation:
//     a user entry identical to a stack entry (for example "Read") is removed
//     when switching to a tier that does not include it.
//
// A base that is not strict JSON, or whose `permissions` is not an object, is left
// to the existing merge (unchanged behavior).

// ownedAllow is every allow entry the stack can emit in any tier.
func ownedAllow() []string { return append([]string(nil), claudeAllowList...) }

// unionClaudeArrays returns overlay with permissions.allow/deny replaced by the
// union described above. It never fails: on any parse problem it returns overlay.
func unionClaudeArrays(base, overlay []byte) []byte {
	var baseObj map[string]any
	if err := json.Unmarshal(base, &baseObj); err != nil {
		return overlay
	}
	basePerms, ok := baseObj["permissions"].(map[string]any)
	if !ok {
		return overlay
	}
	var over map[string]any
	if err := json.Unmarshal(overlay, &over); err != nil {
		return overlay
	}
	overPerms, ok := over["permissions"].(map[string]any)
	if !ok {
		return overlay
	}

	changed := false
	for _, key := range []string{"allow", "deny"} {
		existing, hasExisting := basePerms[key].([]any)
		want := anyStrings(overPerms[key])
		if !hasExisting {
			continue // nothing of the user's to preserve
		}
		var drop map[string]bool
		if key == "allow" {
			drop = subtract(ownedAllow(), want)
		}
		overPerms[key] = unionKeep(existing, want, drop)
		changed = true
	}
	if !changed {
		return overlay
	}
	out, err := json.Marshal(over)
	if err != nil {
		return overlay
	}
	return out
}

// unionKeep returns existing (minus drop) followed by the want entries that are
// not already present. Non-string entries of existing are kept untouched.
func unionKeep(existing []any, want []string, drop map[string]bool) []any {
	out := make([]any, 0, len(existing)+len(want))
	have := map[string]bool{}
	for _, e := range existing {
		if s, ok := e.(string); ok {
			if drop[s] {
				continue
			}
			if have[s] {
				continue // collapse accidental duplicates
			}
			have[s] = true
		}
		out = append(out, e)
	}
	for _, w := range want {
		if !have[w] {
			have[w] = true
			out = append(out, w)
		}
	}
	return out
}

func anyStrings(v any) []string {
	arr, _ := v.([]string)
	if arr != nil {
		return arr
	}
	raw, _ := v.([]any)
	out := make([]string, 0, len(raw))
	for _, e := range raw {
		if s, ok := e.(string); ok {
			out = append(out, s)
		}
	}
	return out
}

// subtract returns the set of owned entries that are not in keep.
func subtract(owned, keep []string) map[string]bool {
	k := make(map[string]bool, len(keep))
	for _, s := range keep {
		k[s] = true
	}
	out := map[string]bool{}
	for _, s := range owned {
		if !k[s] {
			out[s] = true
		}
	}
	return out
}
