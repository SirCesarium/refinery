package rust

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/SirCesarium/refinery/internal/config"
	"github.com/SirCesarium/refinery/internal/engine"
)

func (e *RustEngine) build(cfg *config.Config, art *config.ArtifactConfig, opts engine.BuildOptions) error {
	bestMatch := e.getBestMatch(art, opts.OS, opts.Arch, opts.ABI)
	if bestMatch == nil {
		return fmt.Errorf("no matching target found for %s-%s-%s", opts.OS, opts.Arch, opts.ABI)
	}

	targetTriple := e.resolveTarget(*bestMatch, opts.Arch, opts.ABI)
	manifest, err := e.loadManifest()
	if err != nil {
		return err
	}

	profile := "release"
	if p, ok := bestMatch.LangOpts["profile"].(string); ok {
		profile = p
	}

	if err := e.setupEnvironment(art, opts.OS, opts.Arch, opts.ABI, targetTriple); err != nil {
		return err
	}

	version := manifest.Package.Version
	resolvedHooks := art.Hooks.ResolveAll(opts.ArtifactName, opts.OS, opts.Arch, version, opts.ABI, "")
	for _, hook := range resolvedHooks.PreBuild {
		if err := e.runHook(hook); err != nil {
			return fmt.Errorf("pre-build hook failed: %w", err)
		}
	}

	if err := e.addTarget(targetTriple); err != nil {
		return err
	}

	if err := e.runCargoBuild(art, opts.ArtifactName, opts.OS, opts.Arch, opts.ABI, targetTriple, profile); err != nil {
		return err
	}

	if err := e.moveArtifacts(cfg, art, opts.ArtifactName, opts.OS, opts.Arch, opts.ABI, targetTriple, version, profile, manifest); err != nil {
		return err
	}

	for _, hook := range resolvedHooks.PostBuild {
		if err := e.runHook(hook); err != nil {
			return fmt.Errorf("post-build hook failed: %w", err)
		}
	}

	return nil
}

func (e *RustEngine) moveArtifacts(cfg *config.Config, art *config.ArtifactConfig, artifactName, osName, arch, abi, target, version, profile string, manifest *cargoManifest) error {
	var buildTypes []string
	if art.Type == "bin" {
		buildTypes = []string{"bin"}
	} else {
		buildTypes = art.LibraryTypes
		if len(buildTypes) == 0 {
			buildTypes = []string{"cdylib"}
		}
	}

	movedCount := 0
	cargoProfileDir := profile
	if profile == "debug" || profile == "dev" {
		cargoProfileDir = "debug"
	}

	for _, bt := range buildTypes {
		ext, prefix := e.getExtAndPrefix(osName, art.Type, bt)
		finalName := cfg.Naming.Resolve(cfg.Naming.Binary, artifactName, osName, arch, version, abi, ext)

		realSrcName := artifactName
		if art.Type == "lib" && manifest.Lib != nil && manifest.Lib.Name != "" {
			realSrcName = manifest.Lib.Name
		}

		if prefix != "" && !strings.HasPrefix(realSrcName, prefix) {
			realSrcName = prefix + realSrcName
		}
		if ext != "" {
			realSrcName += "." + ext
		}

		srcPath := filepath.Join("target", target, cargoProfileDir, realSrcName)
		distPath := filepath.Join(cfg.OutputDir, finalName)

		if err := os.MkdirAll(cfg.OutputDir, 0755); err != nil {
			return err
		}

		if _, err := os.Stat(srcPath); os.IsNotExist(err) {
			continue
		}

		if err := os.Rename(srcPath, distPath); err != nil {
			return err
		}
		movedCount++
	}

	if movedCount == 0 {
		return fmt.Errorf("no artifacts found for %s in %s", artifactName, target)
	}
	return nil
}

