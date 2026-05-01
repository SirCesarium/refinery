package engine

import (
	"github.com/SirCesarium/refinery/internal/config"
)

type BuildOptions struct {
	ArtifactName string
	OS           string
	Arch         string
	ABI          string
}

type CIStep struct {
	Name string
	Run  string
	Uses string
	With map[string]any
}

type BuildEngine interface {
	ID() string
	Prepare(cfg *config.Config) error
	Validate(cfg *config.Config) error
	Build(cfg *config.Config, art *config.ArtifactConfig, opts BuildOptions) error
	GetCIRequirements() []string
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
