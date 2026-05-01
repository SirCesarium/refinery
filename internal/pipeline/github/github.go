package github

import (
	"errors"
	"fmt"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Workflow struct {
	Name        string            `yaml:"name"`
	On          On                `yaml:"on"`
	Permissions map[string]string `yaml:"permissions,omitempty"`
	Env         map[string]string `yaml:"env,omitempty"`
	Jobs        map[string]Job    `yaml:"jobs"`
}

type On struct {
	Push             *Event `yaml:"push,omitempty"`
	PullRequest      *Event `yaml:"pull_request,omitempty"`
	WorkflowDispatch any    `yaml:"workflow_dispatch,omitempty"`
	Release          any    `yaml:"release,omitempty"`
}

type Event struct {
	Branches []string `yaml:"branches,omitempty"`
	Tags     []string `yaml:"tags,omitempty"`
	Paths    []string `yaml:"paths,omitempty"`
}

type Job struct {
	Name     string            `yaml:"name,omitempty"`
	RunsOn   string            `yaml:"runs-on"`
	Needs    []string          `yaml:"needs,omitempty"`
	If       string            `yaml:"if,omitempty"`
	Strategy *Strategy         `yaml:"strategy,omitempty"`
	Steps    []Step            `yaml:"steps"`
	Env      map[string]string `yaml:"env,omitempty"`
}

type Strategy struct {
	FailFast bool           `yaml:"fail-fast,omitempty"`
	Matrix   map[string]any `yaml:"matrix,omitempty"`
}

type Step struct {
	ID              string            `yaml:"id,omitempty"`
	Name            string            `yaml:"name,omitempty"`
	Uses            string            `yaml:"uses,omitempty"`
	With            map[string]any    `yaml:"with,omitempty"`
	Run             string            `yaml:"run,omitempty"`
	Env             map[string]string `yaml:"env,omitempty"`
	Shell           string            `yaml:"shell,omitempty"`
	ContinueOnError bool              `yaml:"continue-on-error,omitempty"`
}

type GithubProvider struct {
	Config   Workflow
	filename string
}

func NewProvider(cfg Workflow, filename string) (*GithubProvider, error) {
	if filename == "" {
		return nil, errors.New("filename cannot be empty")
	}

	return &GithubProvider{
		Config:   cfg,
		filename: filename,
	}, nil
}

func (p *GithubProvider) Name() string {
	return "github"
}

func (p *GithubProvider) Filename() string {
	return filepath.Join(".github", "workflows", fmt.Sprintf("%s.yml", p.filename))
}

func (p *GithubProvider) Generate(config any) ([]byte, error) {
	if config != nil {
		return yaml.Marshal(config)
	}

	return yaml.Marshal(p.Config)
}
