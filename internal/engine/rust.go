package engine

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/SirCesarium/refinery/internal/config"
	"github.com/pelletier/go-toml/v2"
)

type RustEngine struct{}

type cargoManifest struct {
	Package struct {
		Name    string `toml:"name"`
		Version string `toml:"version"`
	} `toml:"package"`
	Lib *struct {
		Name string `toml:"name"`
	} `toml:"lib"`
	Bin []struct {
		Name string `toml:"name"`
	} `toml:"bin"`
}

func (e *RustEngine) ID() string {
	return "rust"
}

func (e *RustEngine) Prepare(cfg *config.Config) error {
	if _, err := exec.LookPath("rustup"); err != nil {
		return fmt.Errorf("rustup not found in PATH")
	}
	return nil
}

func (e *RustEngine) Validate(cfg *config.Config) error {
	manifest, err := e.loadManifest()
	if err != nil {
		return err
	}

	for name, art := range cfg.Artifacts {
		found := false
		if art.Type == "lib" {
			if manifest.Lib != nil && manifest.Lib.Name == name {
				found = true
			}
			if !found && strings.ReplaceAll(manifest.Package.Name, "-", "_") == name {
				found = true
			}
		} else {
			for _, b := range manifest.Bin {
				if b.Name == name {
					found = true
					break
				}
			}
			if !found && manifest.Package.Name == name {
				found = true
			}
		}

		if !found {
			return fmt.Errorf("artifact mismatch: '%s' (type %s) defined in refinery.toml not found in Cargo.toml", name, art.Type)
		}
	}
	return nil
}

func (e *RustEngine) Build(cfg *config.Config, art *config.ArtifactConfig, opts BuildOptions) error {
	targetTriple := e.resolveTarget(opts)
	manifest, err := e.loadManifest()
	if err != nil {
		return err
	}

	if err := e.setupEnvironment(art, opts, targetTriple); err != nil {
		return err
	}

	if err := e.addTarget(targetTriple); err != nil {
		return err
	}

	if err := e.runCargoBuild(cfg, art, opts, targetTriple); err != nil {
		return err
	}

	return e.moveArtifact(cfg, art, opts, targetTriple, manifest.Package.Version)
}

func (e *RustEngine) GetCIRequirements() []string {
	return []string{"rust"}
}

func (e *RustEngine) loadManifest() (*cargoManifest, error) {
	cargoBytes, err := os.ReadFile("Cargo.toml")
	if err != nil {
		return nil, fmt.Errorf("could not read Cargo.toml: %w", err)
	}

	var manifest cargoManifest
	if err := toml.Unmarshal(cargoBytes, &manifest); err != nil {
		return nil, fmt.Errorf("error parsing Cargo.toml: %w", err)
	}
	return &manifest, nil
}

func (e *RustEngine) resolveTarget(opts BuildOptions) string {
	switch opts.OS {
	case "darwin":
		return fmt.Sprintf("%s-apple-darwin", opts.Arch)
	case "windows":
		triple := fmt.Sprintf("%s-pc-windows", opts.Arch)
		if opts.ABI != "" {
			triple = fmt.Sprintf("%s-%s", triple, opts.ABI)
		}
		return triple
	case "wasm":
		return "wasm32-unknown-unknown"
	case "wasi":
		return "wasm32-wasip1"
	default:
		triple := fmt.Sprintf("%s-unknown-%s", opts.Arch, opts.OS)
		if opts.ABI != "" {
			triple = fmt.Sprintf("%s-%s", triple, opts.ABI)
		}
		return triple
	}
}

func (e *RustEngine) setupEnvironment(art *config.ArtifactConfig, opts BuildOptions, target string) error {
	for tName, tCfg := range art.Targets {
		if tName == opts.OS {
			if linker, ok := tCfg.LangOpts["linker"].(string); ok {
				envKey := fmt.Sprintf("CARGO_TARGET_%s_LINKER",
					strings.ReplaceAll(strings.ReplaceAll(strings.ToUpper(target), "-", "_"), ".", "_"))
				os.Setenv(envKey, linker)
			}
			if depTarget, ok := tCfg.LangOpts["deployment_target"].(string); ok {
				os.Setenv("MACOSX_DEPLOYMENT_TARGET", depTarget)
			}
			if sdk, ok := tCfg.LangOpts["sdk_root"].(string); ok {
				os.Setenv("SDKROOT", sdk)
			}
		}
	}

	if opts.OS == "darwin" && os.Getenv("MACOSX_DEPLOYMENT_TARGET") == "" {
		os.Setenv("MACOSX_DEPLOYMENT_TARGET", "11.0")
	}
	return nil
}

func (e *RustEngine) addTarget(target string) error {
	cmd := exec.Command("rustup", "target", "add", target)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (e *RustEngine) runCargoBuild(cfg *config.Config, art *config.ArtifactConfig, opts BuildOptions, target string) error {
	profile := "release"
	args := []string{"build", "--" + profile, "--target", target}
	
	if art.Type == "lib" {
		args = append(args, "--lib")
	} else {
		args = append(args, "--bin", opts.ArtifactName)
	}

	cmd := exec.Command("cargo", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()
	return cmd.Run()
}

func (e *RustEngine) moveArtifact(cfg *config.Config, art *config.ArtifactConfig, opts BuildOptions, target, version string) error {
	ext, prefix := e.getArtifactAttributes(opts, art.Type)
	finalName := cfg.Naming.Resolve(cfg.Naming.Binary, opts.ArtifactName, opts.OS, opts.Arch, version, opts.ABI, ext)

	realSrcName := opts.ArtifactName
	if prefix != "" && !strings.HasPrefix(realSrcName, prefix) {
		realSrcName = prefix + realSrcName
	}
	if ext != "" {
		realSrcName += "." + ext
	}

	srcPath := filepath.Join("target", target, "release", realSrcName)
	distPath := filepath.Join(cfg.OutputDir, finalName)

	if err := os.MkdirAll(cfg.OutputDir, 0755); err != nil {
		return err
	}

	if _, err := os.Stat(srcPath); os.IsNotExist(err) {
		return fmt.Errorf("artifact not found at %s", srcPath)
	}

	return os.Rename(srcPath, distPath)
}

func (e *RustEngine) getArtifactAttributes(opts BuildOptions, artType string) (string, string) {
	var ext, prefix string
	switch opts.OS {
	case "windows":
		if artType == "bin" {
			ext = "exe"
		} else {
			ext = "dll"
		}
	case "wasm", "wasi":
		ext = "wasm"
	case "linux", "darwin":
		if artType == "lib" {
			prefix = "lib"
			ext = "so"
			if opts.OS == "darwin" {
				ext = "dylib"
			}
		}
	}
	return ext, prefix
}
