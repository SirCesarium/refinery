package pipeline

import "github.com/SirCesarium/refinery/internal/config"

type Step struct {
	Name string
	Run  string
	If   string
}

type CIProvider interface {
	Name() string
	Generate(cfg *config.Config) ([]byte, error)
	Filename() string
	GetSetupSteps(lang string) []Step
}

type PipelineGenerator struct {
	provider CIProvider
}

func NewGenerator(p CIProvider) *PipelineGenerator {
	return &PipelineGenerator{provider: p}
}

func (g *PipelineGenerator) Generate(cfg *config.Config) ([]byte, error) {
	return g.provider.Generate(cfg)
}

func (g *PipelineGenerator) Filename() string {
	return g.provider.Filename()
}
