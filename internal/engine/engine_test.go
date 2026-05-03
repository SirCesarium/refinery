package engine

import (
	"testing"

	"github.com/SirCesarium/refinery/internal/config"
)

// TestNewRegistry checks that registry is properly initialized.
func TestNewRegistry(t *testing.T) {
	r := NewRegistry()
	if r == nil {
		t.Fatal("expected non-nil registry")
	}
	if len(r.engines) != 0 {
		t.Errorf("expected empty registry, got %d engines", len(r.engines))
	}
}

// TestRegisterAndGet checks engine registration and retrieval.
func TestRegisterAndGet(t *testing.T) {
	r := NewRegistry()

	// Create a mock engine
	eng := &mockEngine{id: "test-engine"}
	r.Register(eng)

	// Retrieve the engine
	retrieved := r.Get("test-engine")
	if retrieved == nil {
		t.Fatal("expected non-nil engine")
	}
	if retrieved.ID() != "test-engine" {
		t.Errorf("expected 'test-engine', got '%s'", retrieved.ID())
	}
}

// TestGetNonExistent checks nil for non-existent engine.
func TestGetNonExistent(t *testing.T) {
	r := NewRegistry()
	eng := r.Get("non-existent")
	if eng != nil {
		t.Error("expected nil for non-existent engine")
	}
}

// TestGetAfterRegister checks engine retrieval after registration.
func TestGetAfterRegister(t *testing.T) {
	r := NewRegistry()

	// Register multiple engines
	r.Register(&mockEngine{id: "engine1"})
	r.Register(&mockEngine{id: "engine2"})

	// Check both are retrievable
	if e1 := r.Get("engine1"); e1 == nil {
		t.Error("expected engine1 to be retrievable")
	}
	if e2 := r.Get("engine2"); e2 == nil {
		t.Error("expected engine2 to be retrievable")
	}
}

// mockEngine for testing
type mockEngine struct {
	id string
}

func (m *mockEngine) ID() string {
	return m.id
}
func (m *mockEngine) Prepare(cfg *config.Config) error {
	return nil
}
func (m *mockEngine) Validate(cfg *config.Config) error {
	return nil
}
func (m *mockEngine) Build(cfg *config.Config, art *config.ArtifactConfig, opts BuildOptions) error {
	return nil
}
func (m *mockEngine) GetCIRequirements(cfg *config.Config) []string {
	return nil
}
func (m *mockEngine) Package(cfg *config.Config, art *config.ArtifactConfig, opts BuildOptions, format string) error {
	return nil
}
