package agents_test

import (
	"errors"
	"testing"

	"github.com/JuanCruzRobledo/jr-stack/internal/agents"
	"github.com/JuanCruzRobledo/jr-stack/internal/harness/external"
	"github.com/JuanCruzRobledo/jr-stack/internal/model"
)

// fakeAdapter is a minimal test double that satisfies agents.Adapter.
type fakeAdapter struct {
	agent model.Agent
}

func (f fakeAdapter) Agent() model.Agent                { return f.agent }
func (f fakeAdapter) InstructionsPath(_ string) string  { return "" }
func (f fakeAdapter) SkillsDir(_ string) string         { return "" }
func (f fakeAdapter) CommandsDir(_ string) string       { return "" }
func (f fakeAdapter) SettingsPath(_ string) string      { return "" }
func (f fakeAdapter) MCPConfigPath(_, _ string) string  { return "" }
func (f fakeAdapter) MCPStrategy() external.MCPStrategy { return external.StrategySeparateFile }
func (f fakeAdapter) VariantKey() string                { return string(f.agent) }
func (f fakeAdapter) PathsFor(_ string, _ model.InstallTarget) model.AgentPaths {
	return model.AgentPaths{}
}
func (f fakeAdapter) ConfigDelivery() model.ConfigDelivery { return model.ConfigDeliveryInstructions }

// Compile-time check: fakeAdapter satisfies agents.Adapter.
var _ agents.Adapter = fakeAdapter{}

func TestRegistry_LookupHit(t *testing.T) {
	r := agents.NewRegistry()
	_ = r.Register(fakeAdapter{agent: model.AgentClaude})

	got, ok := r.Get(model.AgentClaude)
	if !ok {
		t.Fatal("Get(claude) returned ok=false, want true")
	}
	if got.Agent() != model.AgentClaude {
		t.Errorf("Get(claude).Agent() = %q, want %q", got.Agent(), model.AgentClaude)
	}
}

func TestRegistry_LookupMiss(t *testing.T) {
	r := agents.NewRegistry()
	_, ok := r.Get(model.AgentGemini)
	if ok {
		t.Fatal("Get(gemini) returned ok=true on empty registry, want false")
	}
}

func TestRegistry_DuplicateRegistration(t *testing.T) {
	r := agents.NewRegistry()
	_ = r.Register(fakeAdapter{agent: model.AgentClaude})
	err := r.Register(fakeAdapter{agent: model.AgentClaude})

	if err == nil {
		t.Fatal("Register duplicate expected error, got nil")
	}
	if !errors.Is(err, agents.ErrDuplicateAdapter) {
		t.Errorf("error = %v, want ErrDuplicateAdapter", err)
	}
}

func TestRegistry_SupportedAgents_ExactlyP0(t *testing.T) {
	r, err := agents.NewDefaultRegistry()
	if err != nil {
		t.Fatalf("NewDefaultRegistry() error: %v", err)
	}

	got := r.SupportedAgents()
	want := []model.Agent{model.AgentClaude, model.AgentOpenCode}

	if len(got) != len(want) {
		t.Fatalf("SupportedAgents() len = %d, want %d; got %v", len(got), len(want), got)
	}
	for i, a := range want {
		if got[i] != a {
			t.Errorf("SupportedAgents()[%d] = %q, want %q", i, got[i], a)
		}
	}
}

func TestFactory_UnsupportedAgent_TypedError(t *testing.T) {
	_, err := agents.NewAdapter(model.AgentGemini)
	if err == nil {
		t.Fatal("NewAdapter(gemini) expected error, got nil")
	}
	if !errors.Is(err, agents.ErrAgentNotSupported) {
		t.Errorf("error = %v, want ErrAgentNotSupported", err)
	}

	var notSupported agents.AgentNotSupportedError
	if !errors.As(err, &notSupported) {
		t.Errorf("error does not wrap AgentNotSupportedError")
	}
	if notSupported.Agent != model.AgentGemini {
		t.Errorf("AgentNotSupportedError.Agent = %q, want %q", notSupported.Agent, model.AgentGemini)
	}
}

func TestDefaultRegistry_DoesNotContainRemainingAgents(t *testing.T) {
	r, err := agents.NewDefaultRegistry()
	if err != nil {
		t.Fatalf("NewDefaultRegistry() error: %v", err)
	}

	remaining := []model.Agent{
		model.AgentGemini,
		model.AgentCodex,
		model.AgentCursor,
		model.AgentVSCode,
		model.AgentWindsurf,
		model.AgentAntigravity,
	}

	for _, a := range remaining {
		if _, ok := r.Get(a); ok {
			t.Errorf("default registry unexpectedly contains %q", a)
		}
	}
}
