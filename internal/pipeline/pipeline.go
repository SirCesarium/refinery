package pipeline

type CIProvider interface {
	Name() string
	Generate(config any) ([]byte, error)
	Filename() string
}

type PipelineGenerator struct {
	provider CIProvider
}

func NewGenerator(p CIProvider) *PipelineGenerator {
	return &PipelineGenerator{provider: p}
}

func (g *PipelineGenerator) Generate(config any) ([]byte, error) {
	return g.provider.Generate(config)
}

func (g *PipelineGenerator) Filename() string {
	return g.provider.Filename()
}
