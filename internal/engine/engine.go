// Package engine defines the core interfaces and registry for build engines.
package engine

import (
	"github.com/SirCesarium/refinery/internal/config"
)

// BuildOptions defines the target for a build operation.
type BuildOptions struct {
	ArtifactName string
	OS           string
	Arch         string
	ABI          string
}

// BuildEngine defines the interface for language-specific build engines.
type BuildEngine interface {
	ID() string
	Prepare(cfg *config.Config) error
	Validate(cfg *config.Config) error
	Build(cfg *config.Config, art *config.ArtifactConfig, opts BuildOptions) error
	GetCIRequirements(cfg *config.Config) []string
	Package(cfg *config.Config, art *config.ArtifactConfig, opts BuildOptions, format string) error
}

type EngineRegistry struct {
	engines map[string]BuildEngine
}

func NewRegistry() *EngineRegistry {
	return &EngineRegistry{
		engines: make(map[string]BuildEngine),
	}
}

func (r *EngineRegistry) Register(e BuildEngine) {
	r.engines[e.ID()] = e
}

func (r *EngineRegistry) Get(id string) BuildEngine {
	return r.engines[id]
}
