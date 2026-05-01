package pipeline

import (
	"github.com/SirCesarium/refinery/internal/config"
	"github.com/SirCesarium/refinery/internal/engine"
)

type CIProvider interface {
	Name() string
	Generate(cfg *config.Config, eng engine.BuildEngine) ([]byte, error)
	Filename() string
}

type ProviderRegistry struct {
	providers map[string]CIProvider
}

func NewRegistry() *ProviderRegistry {
	return &ProviderRegistry{
		providers: make(map[string]CIProvider),
	}
}

func (r *ProviderRegistry) Register(p CIProvider) {
	r.providers[p.Name()] = p
}

func (r *ProviderRegistry) Get(name string) CIProvider {
	return r.providers[name]
}

type PipelineGenerator struct {
	provider CIProvider
	engine   engine.BuildEngine
}

func NewGenerator(p CIProvider, e engine.BuildEngine) *PipelineGenerator {
	return &PipelineGenerator{provider: p, engine: e}
}

func (g *PipelineGenerator) Generate(cfg *config.Config) ([]byte, error) {
	return g.provider.Generate(cfg, g.engine)
}

func (g *PipelineGenerator) Filename() string {
	return g.provider.Filename()
}
