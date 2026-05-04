package rust

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/SirCesarium/refinery/internal/config"
	"github.com/SirCesarium/refinery/internal/engine"
	"github.com/SirCesarium/refinery/internal/ui"
)

// build orchestrates the full build process: setup, compile, and move artifacts.
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

	profile := e.getProfile(*bestMatch)
	if err := e.setupEnvironment(art, opts.OS, opts.Arch, opts.ABI, targetTriple); err != nil {
		return err
	}

	version := opts.Version
	if version == "" || version == "0.0.0" {
		version = manifest.Package.Version
	}

	_, binaryPath := e.resolveBinaryInfo(cfg, art, opts, manifest, targetTriple, profile, version)
	if err := e.runHooks(art, opts, binaryPath, version, "PreBuild"); err != nil {
		return err
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

	return e.runHooks(art, opts, binaryPath, version, "PostBuild")
}

// getProfile extracts the build profile from target config.
func (e *RustEngine) getProfile(tCfg config.TargetConfig) string {
	profile := "release"
	if p, ok := tCfg.LangOpts["profile"].(string); ok {
		profile = p
	}
	return profile
}

// resolveBinaryInfo returns the binary name and full path based on naming config.
func (e *RustEngine) resolveBinaryInfo(cfg *config.Config, art *config.ArtifactConfig, opts engine.BuildOptions, manifest *cargoManifest, targetTriple, profile, version string) (string, string) {
	ext := e.getBinaryExt(art, opts.OS, opts.ABI)
	binaryName := cfg.Naming.Resolve(cfg.Naming.Binary, opts.ArtifactName, opts.OS, opts.Arch, version, opts.ABI, ext)
	return binaryName, filepath.Join(cfg.OutputDir, binaryName)
}

// getBinaryExt determines the file extension for the binary based on OS and type.
func (e *RustEngine) getBinaryExt(art *config.ArtifactConfig, osName, abi string) string {
	ext := ""
	if art.Type == "bin" {
		ext, _ = e.getExtAndPrefix(osName, abi, art.Type, "bin")
	} else if len(art.LibraryTypes) > 0 {
		ext, _ = e.getExtAndPrefix(osName, abi, art.Type, art.LibraryTypes[0])
	} else {
		ext, _ = e.getExtAndPrefix(osName, abi, art.Type, "cdylib")
	}
	return ext
}

// runHooks executes either pre-build or post-build hooks.
func (e *RustEngine) runHooks(art *config.ArtifactConfig, opts engine.BuildOptions, binaryPath, version, hookType string) error {
	resolvedHooks := art.Hooks.ResolveAll(opts.ArtifactName, opts.OS, opts.Arch, version, opts.ABI, binaryPath)

	var hooks []string
	if hookType == "PreBuild" {
		hooks = resolvedHooks.PreBuild
	} else {
		hooks = resolvedHooks.PostBuild
	}

	for _, hook := range hooks {
		if err := e.runHook(hook); err != nil {
			return fmt.Errorf("%s hook failed: %w", strings.ToLower(hookType), err)
		}
	}
	return nil
}

// moveArtifacts copies built files from cargo target dir to the output directory.
// It also handles header files if configured.
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
		ext, prefix := e.getExtAndPrefix(osName, abi, art.Type, bt)
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
			ui.Warn("Expected artifact not found at %s (build type: %s). Skipping...", srcPath, bt)
			continue
		}

		if err := moveFile(srcPath, distPath); err != nil {
			return err
		}
		movedCount++
	}

	if art.Headers {
		headers, err := filepath.Glob("*.h")
		if err != nil {
			return fmt.Errorf("failed to search for .h headers: %w", err)
		}

		headers2, err := filepath.Glob("*.hpp")
		if err != nil {
			return fmt.Errorf("failed to search for .hpp headers: %w", err)
		}

		headers = append(headers, headers2...)

		for _, h := range headers {
			dest := filepath.Join(cfg.OutputDir, h)
			if err := copyFile(h, dest); err != nil {
				return fmt.Errorf("failed to copy header %s to %s: %w", h, dest, err)
			}
		}
	}

	if movedCount == 0 {
		return fmt.Errorf("no artifacts found for %s in target %s (searched for build types: %v)", artifactName, target, buildTypes)
	}
	return nil
}

func moveFile(src, dst string) error {
	if err := os.Rename(src, dst); err == nil {
		return nil
	}

	if err := copyFile(src, dst); err != nil {
		return err
	}

	return os.Remove(src)
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}

	defer in.Close()

	stat, err := in.Stat()
	if err != nil {
		return err
	}

	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, stat.Mode())
	if err != nil {
		return err
	}

	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
