package app

import (
	"testing"

	"github.com/SirCesarium/refinery/internal/engine"
	"github.com/SirCesarium/refinery/internal/pipeline"
)

// TestNewDefaultEngineRegistry checks that Rust engine is registered.
func TestNewDefaultEngineRegistry(t *testing.T) {
	r := NewDefaultEngineRegistry()
	if r == nil {
		t.Fatal("expected non-nil registry")
	}

	// Check if rust engine is registered
	eng := r.Get("rust")
	if eng == nil {
		t.Error("expected rust engine to be registered")
	}
}

// TestNewDefaultProviderRegistry checks that GitHub provider is registered.
func TestNewDefaultProviderRegistry(t *testing.T) {
	r, err := NewDefaultProviderRegistry()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if r == nil {
		t.Fatal("expected non-nil registry")
	}

	// Check if github provider is registered
	p := r.Get("github")
	if p == nil {
		t.Error("expected github provider to be registered")
	}
}

// TestEngineRegistryGet checks nil for non-existent engine.
func TestEngineRegistryGetNil(t *testing.T) {
	r := engine.NewRegistry()
	eng := r.Get("non-existent")
	if eng != nil {
		t.Error("expected nil for non-existent engine")
	}
}

// TestProviderRegistryGet checks nil for non-existent provider.
func TestProviderRegistryGetNil(t *testing.T) {
	r := pipeline.NewRegistry()
	p := r.Get("non-existent")
	if p != nil {
		t.Error("expected nil for non-existent provider")
	}
}
