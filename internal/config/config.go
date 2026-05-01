package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/pelletier/go-toml/v2"
	"github.com/spf13/viper"
)

type Config struct {
	RefineryVersion string                     `toml:"refinery_version" mapstructure:"refinery_version"`
	Project         Project                    `toml:"project" mapstructure:"project"`
	OutputDir       string                     `toml:"output_dir" mapstructure:"output_dir"`
	Artifacts       map[string]*ArtifactConfig `toml:"artifacts" mapstructure:"artifacts"`
	Naming          NamingConfig               `toml:"naming,omitempty" mapstructure:"naming"`
	Metadata        map[string]string          `toml:"metadata,omitempty" mapstructure:"metadata"`
}

type Project struct {
	Name string `toml:"name" mapstructure:"name"`
	Lang string `toml:"lang" mapstructure:"lang"`
}

type ArtifactConfig struct {
	Type    string                  `toml:"type" mapstructure:"type"`
	Source  string                  `toml:"source" mapstructure:"source"`
	Formats []string                `toml:"formats,omitempty" mapstructure:"formats"`
	Headers bool                    `toml:"headers,omitempty" mapstructure:"headers"`
	Targets map[string]TargetConfig `toml:"targets" mapstructure:"targets"`
	Hooks   Hooks                   `toml:"hooks,omitempty" mapstructure:"hooks"`
}

type TargetConfig struct {
	Archs    []string       `toml:"archs" mapstructure:"archs"`
	ABIs     []string       `toml:"abis,omitempty" mapstructure:"abis"`
	Packages []string       `toml:"packages" mapstructure:"packages"`
	LangOpts map[string]any `toml:"lang_opts,omitempty" mapstructure:"lang_opts"`
}

type Hooks struct {
	PreBuild  []string `toml:"pre_build,omitempty" mapstructure:"pre_build"`
	PostBuild []string `toml:"post_build,omitempty" mapstructure:"post_build"`
}

type NamingConfig struct {
	Binary  string `toml:"binary,omitempty" mapstructure:"binary"`
	Package string `toml:"package,omitempty" mapstructure:"package"`
}

func (c *Config) Validate() error {
	if c.RefineryVersion == "" {
		return fmt.Errorf("refinery_version is required")
	}
	if c.Project.Name == "" || c.Project.Lang == "" {
		return fmt.Errorf("project name and lang are required")
	}
	if c.OutputDir == "" {
		return fmt.Errorf("output_dir is required")
	}
	if c.Naming.Binary == "" || c.Naming.Package == "" {
		return fmt.Errorf("naming binary and package are required")
	}
	if len(c.Artifacts) == 0 {
		return fmt.Errorf("at least one artifact must be defined")
	}
	for name, art := range c.Artifacts {
		if art.Type != "bin" && art.Type != "lib" {
			return fmt.Errorf("artifact %s type must be 'bin' or 'lib'", name)
		}
		if art.Source == "" {
			return fmt.Errorf("artifact %s source is required", name)
		}
		if len(art.Targets) == 0 {
			return fmt.Errorf("artifact %s must have at least one target", name)
		}
		for tName, tCfg := range art.Targets {
			if len(tCfg.Archs) == 0 {
				return fmt.Errorf("target %s in artifact %s must have at least one arch", tName, name)
			}
		}
	}
	return nil
}

func (n NamingConfig) Resolve(template, artifact, osName, arch, version, abi, ext string) string {
	if template == "" {
		return ""
	}

	formattedAbi := abi
	if formattedAbi != "" && !strings.HasPrefix(formattedAbi, "-") {
		formattedAbi = "-" + formattedAbi
	}

	var formattedExt string
	if ext != "" {
		formattedExt = strings.TrimPrefix(ext, ".")
	}

	r := strings.NewReplacer(
		"{artifact}", artifact,
		"{os}", osName,
		"{arch}", arch,
		"{version}", version,
		"{abi}", formattedAbi,
		"{ext}", formattedExt,
	)

	result := r.Replace(template)
	result = strings.Trim(result, "-")

	if formattedExt != "" {
		if !strings.HasSuffix(result, "."+formattedExt) {
			result = strings.TrimSuffix(result, ".")
			result = result + "." + formattedExt
		}
	}

	return result
}

func Load(path string) (*Config, error) {
	v := viper.New()

	v.SetDefault("output_dir", "dist")
	v.SetDefault("naming.binary", "{artifact}-{os}-{arch}{abi}")
	v.SetDefault("naming.package", "{artifact}-{version}-{os}-{arch}{abi}.{ext}")

	if path != "" {
		v.SetConfigFile(path)
		if err := v.ReadInConfig(); err != nil {
			if !os.IsNotExist(err) {
				return nil, err
			}
		}
	}

	cfg := &Config{}

	if err := v.Unmarshal(cfg); err != nil {
		return nil, err
	}

	if cfg.Naming.Binary == "" {
		cfg.Naming.Binary = v.GetString("naming.binary")
	}

	if cfg.Naming.Package == "" {
		cfg.Naming.Package = v.GetString("naming.package")
	}

	if cfg.OutputDir == "" {
		cfg.OutputDir = v.GetString("output_dir")
	}

	for _, art := range cfg.Artifacts {
		for tName, tCfg := range art.Targets {
			if len(tCfg.ABIs) == 0 {
				tCfg.ABIs = []string{""}
				art.Targets[tName] = tCfg
			}
		}
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func Default(name string) *Config {
	return &Config{
		RefineryVersion: "2",
		Project: Project{
			Name: name,
			Lang: "rust",
		},
		OutputDir: "dist",
		Artifacts: make(map[string]*ArtifactConfig),
		Naming: NamingConfig{
			Binary:  "{artifact}-{os}-{arch}{abi}",
			Package: "{artifact}-{version}-{os}-{arch}{abi}.{ext}",
		},
	}
}

func (c *Config) Write(path string) error {
	data, err := toml.Marshal(c)

	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
