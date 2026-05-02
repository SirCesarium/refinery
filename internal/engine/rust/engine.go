package rust

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/SirCesarium/refinery/internal/config"
	"github.com/SirCesarium/refinery/internal/engine"
)

type RustEngine struct{}

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

		for tID, tCfg := range art.Targets {
			for _, arch := range tCfg.Archs {
				abis := tCfg.ABIs
				if len(abis) == 0 {
					abis = []string{""}
				}
				for _, abi := range abis {
					if err := e.validateTriple(tCfg.OS, arch, abi); err != nil {
						return fmt.Errorf("invalid target config '%s' in artifact '%s': %w", tID, name, err)
					}
				}
			}
		}
	}
	return nil
}

func (e *RustEngine) GetCIRequirements(cfg *config.Config) []string {
	reqs := []string{"rust"}
	for _, art := range cfg.Artifacts {
		for _, tCfg := range art.Targets {
			if tCfg.OS == "linux" && e.sliceContains(tCfg.Archs, "aarch64") {
				reqs = append(reqs, "cross-linker:linux-aarch64")
			}
			if e.sliceContains(tCfg.ABIs, "musl") {
				reqs = append(reqs, "pkg:musl-tools")
			}
		}
		for _, p := range art.Packages {
			switch p {
			case "deb":
				reqs = append(reqs, "pkg:cargo-deb")
			case "rpm":
				reqs = append(reqs, "pkg:cargo-generate-rpm")
			case "msi":
				reqs = append(reqs, "pkg:cargo-wix")
			}
		}
	}
	return reqs
}

func (e *RustEngine) Build(cfg *config.Config, art *config.ArtifactConfig, opts engine.BuildOptions) error {
	return e.build(cfg, art, opts)
}

func (e *RustEngine) Package(cfg *config.Config, art *config.ArtifactConfig, opts engine.BuildOptions, format string) error {
	return e.pkg(cfg, art, opts.ArtifactName, opts.OS, opts.Arch, opts.ABI, format)
}

