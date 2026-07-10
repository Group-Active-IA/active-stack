// Package i18n — the key-parity anti-drift seatbelt (task 2.1). Asserts
// tableEN and tableES expose the identical key set; any string added to one
// language without its twin in the other fails the build. This file is
// package i18n (not i18n_test) because it reaches into the unexported
// tables map.
package i18n

import (
	"sort"
	"testing"
)

// TestTableKeyParity asserts keys(tableEN) == keys(tableES) as sets, failing
// with the symmetric difference (missing-in-ES and missing-in-EN) named.
func TestTableKeyParity(t *testing.T) {
	enTable := tables[LangEN]
	esTable := tables[LangES]

	var missingInES, missingInEN []string
	for k := range enTable {
		if _, ok := esTable[k]; !ok {
			missingInES = append(missingInES, k)
		}
	}
	for k := range esTable {
		if _, ok := enTable[k]; !ok {
			missingInEN = append(missingInEN, k)
		}
	}
	sort.Strings(missingInES)
	sort.Strings(missingInEN)

	if len(missingInES) > 0 || len(missingInEN) > 0 {
		t.Fatalf("i18n table key parity broken:\nmissing in ES: %v\nmissing in EN: %v", missingInES, missingInEN)
	}

	if len(enTable) == 0 {
		t.Fatal("tableEN must not be empty")
	}
}
