package app

import (
	"github.com/SirCesarium/refinery/internal/engine"
	"github.com/SirCesarium/refinery/internal/engine/rust"
	"github.com/SirCesarium/refinery/internal/pipeline"
	"github.com/SirCesarium/refinery/internal/pipeline/github"
)

func NewDefaultEngineRegistry() *engine.EngineRegistry {
	r := engine.NewRegistry()
	r.Register(&rust.RustEngine{})
	return r
}

func NewDefaultProviderRegistry() (*pipeline.ProviderRegistry, error) {
	r := pipeline.NewRegistry()

	gh, err := github.NewProvider("refinery-build")
	if err != nil {
		return nil, err
	}
	r.Register(gh)

	return r, nil
}
