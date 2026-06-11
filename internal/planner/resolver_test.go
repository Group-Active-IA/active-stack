package planner

import (
	"reflect"
	"strings"
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/model"
)

// catalog builds a map[string]model.Harness from a slice for test convenience.
func catalog(harnesses []model.Harness) map[string]model.Harness {
	m := make(map[string]model.Harness, len(harnesses))
	for _, h := range harnesses {
		m[h.ID] = h
	}
	return m
}

// h is a helper to create a minimal Harness for testing.
func h(id string, dependsOn ...string) model.Harness {
	return model.Harness{
		ID:           id,
		Name:         id,
		Type:         model.HarnessConfig,
		InstallModes: []model.InstallMode{model.ModeFull},
		DependsOn:    dependsOn,
	}
}

func TestResolverSingleHarnessNoDeps(t *testing.T) {
	resolver := NewResolver()

	engram := h("engram")
	all := catalog([]model.Harness{engram})

	plan, err := resolver.Resolve([]model.Harness{engram}, all)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	if !reflect.DeepEqual(plan.OrderedIDs, []string{"engram"}) {
		t.Fatalf("OrderedIDs = %v, want [engram]", plan.OrderedIDs)
	}
	if len(plan.AddedIDs) != 0 {
		t.Fatalf("AddedIDs = %v, want empty", plan.AddedIDs)
	}
}

func TestResolverTransitiveDependenciesAddedInOrder(t *testing.T) {
	// skills → sdd → engram; user selects only skills.
	resolver := NewResolver()

	engram := h("engram")
	sdd := h("sdd", "engram")
	skills := h("skills", "sdd")

	all := catalog([]model.Harness{engram, sdd, skills})

	plan, err := resolver.Resolve([]model.Harness{skills}, all)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	want := []string{"engram", "sdd", "skills"}
	if !reflect.DeepEqual(plan.OrderedIDs, want) {
		t.Fatalf("OrderedIDs = %v, want %v", plan.OrderedIDs, want)
	}

	wantAdded := []string{"engram", "sdd"}
	if !reflect.DeepEqual(plan.AddedIDs, wantAdded) {
		t.Fatalf("AddedIDs = %v, want %v", plan.AddedIDs, wantAdded)
	}
}

func TestResolverExplicitSelectionNotInAddedIDs(t *testing.T) {
	// User selects sdd explicitly; engram pulled in as dep.
	resolver := NewResolver()

	engram := h("engram")
	sdd := h("sdd", "engram")

	all := catalog([]model.Harness{engram, sdd})

	plan, err := resolver.Resolve([]model.Harness{engram, sdd}, all)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	if !reflect.DeepEqual(plan.OrderedIDs, []string{"engram", "sdd"}) {
		t.Fatalf("OrderedIDs = %v, want [engram sdd]", plan.OrderedIDs)
	}
	if len(plan.AddedIDs) != 0 {
		t.Fatalf("AddedIDs = %v, want empty (both explicitly selected)", plan.AddedIDs)
	}
}

func TestResolverUnknownHarnessInSelectionReturnsError(t *testing.T) {
	resolver := NewResolver()

	ghost := h("ghost")
	all := catalog([]model.Harness{}) // ghost not in catalog

	_, err := resolver.Resolve([]model.Harness{ghost}, all)
	if err == nil {
		t.Fatal("Resolve() expected error for unknown harness, got nil")
	}
}

func TestResolverUnknownDependencyReturnsError(t *testing.T) {
	resolver := NewResolver()

	// sdd claims it depends on "engram" but engram is not in all.
	sdd := h("sdd", "engram")
	all := catalog([]model.Harness{sdd}) // engram missing from catalog

	_, err := resolver.Resolve([]model.Harness{sdd}, all)
	if err == nil {
		t.Fatal("Resolve() expected error for unknown dependency, got nil")
	}
	if !strings.Contains(err.Error(), "engram") {
		t.Fatalf("error should mention missing dep 'engram', got: %v", err)
	}
}

func TestResolverCycleReturnsError(t *testing.T) {
	resolver := NewResolver()

	a := h("a", "b")
	b := h("b", "a")
	all := catalog([]model.Harness{a, b})

	_, err := resolver.Resolve([]model.Harness{a}, all)
	if err == nil {
		t.Fatal("Resolve() expected cycle error, got nil")
	}
}

func TestResolverMultipleIndependentHarnessesAlphabeticalOrder(t *testing.T) {
	resolver := NewResolver()

	ctx7 := h("context7")
	engram := h("engram")
	persona := h("persona")

	all := catalog([]model.Harness{ctx7, engram, persona})

	plan, err := resolver.Resolve([]model.Harness{persona, ctx7, engram}, all)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	want := []string{"context7", "engram", "persona"}
	if !reflect.DeepEqual(plan.OrderedIDs, want) {
		t.Fatalf("OrderedIDs = %v, want %v (alphabetical for no-dep nodes)", plan.OrderedIDs, want)
	}
	if len(plan.AddedIDs) != 0 {
		t.Fatalf("AddedIDs = %v, want empty", plan.AddedIDs)
	}
}
