// Package app initializes and provides default registries for engines and CI providers.
package app

import (
	"github.com/SirCesarium/refinery/internal/engine"
	"github.com/SirCesarium/refinery/internal/engine/goengine"
	"github.com/SirCesarium/refinery/internal/engine/rust"
	"github.com/SirCesarium/refinery/internal/pipeline"
	"github.com/SirCesarium/refinery/internal/pipeline/github"
)

// NewDefaultEngineRegistry creates a registry with supported language engines.
func NewDefaultEngineRegistry() *engine.EngineRegistry {
	r := engine.NewRegistry()
	r.Register(&rust.RustEngine{})
	r.Register(&goengine.GoEngine{})
	return r
}

// NewDefaultProviderRegistry creates a registry with GitHub Actions provider.
func NewDefaultProviderRegistry() (*pipeline.ProviderRegistry, error) {
	r := pipeline.NewRegistry()

	gh, err := github.NewProvider("refinery-build")
	if err != nil {
		return nil, err
	}
	r.Register(gh)

	return r, nil
}
