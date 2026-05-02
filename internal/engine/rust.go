package engine

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
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
		Name        string   `toml:"name"`
		Version     string   `toml:"version"`
		Description string   `toml:"description"`
		Authors     []string `toml:"authors"`
		License     string   `toml:"license"`
		Homepage    string   `toml:"homepage"`
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

func (e *RustEngine) validateTriple(osName, arch, abi string) error {
	switch osName {
	case "darwin":
		if arch != "x86_64" && arch != "aarch64" {
			return fmt.Errorf("unsupported architecture for macOS: %s (only x86_64 and aarch64 are supported)", arch)
		}
	case "windows":
		if arch == "aarch64" && abi == "gnu" {
			return fmt.Errorf("windows-aarch64-gnu is a Tier 3 target and not supported by standard toolchains (use msvc instead)")
		}
		if arch != "x86_64" && arch != "i686" && arch != "aarch64" {
			return fmt.Errorf("unsupported architecture for Windows: %s", arch)
		}
	case "wasm", "wasi":
		if arch != "wasm32" {
			return fmt.Errorf("unsupported architecture for web: %s (must be wasm32)", arch)
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

	profile := "release"
	bestMatch := e.getBestMatch(art, opts)
	if bestMatch != nil {
		if p, ok := bestMatch.LangOpts["profile"].(string); ok {
			profile = p
		}
	}

	if err := e.setupEnvironment(art, opts, targetTriple); err != nil {
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

	if err := e.runCargoBuild(art, opts, targetTriple, profile); err != nil {
		return err
	}

	if err := e.moveArtifacts(cfg, art, opts, targetTriple, version, profile, manifest); err != nil {
		return err
	}

	for _, hook := range resolvedHooks.PostBuild {
		if err := e.runHook(hook); err != nil {
			return fmt.Errorf("post-build hook failed: %w", err)
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

func (e *RustEngine) Package(cfg *config.Config, art *config.ArtifactConfig, opts BuildOptions, format string) error {
	manifest, err := e.loadManifest()
	if err != nil {
		return err
	}

	switch format {
	case "deb":
		return e.runCargoPackager("deb", []string{"--target", e.resolveTarget(opts)})
	case "rpm":
		return e.runCargoPackager("generate-rpm", []string{"--target", e.resolveTarget(opts)})
	case "msi":
		return e.runCargoPackager("wix", []string{"--target", e.resolveTarget(opts)})
	case "tar.gz", "targz":
		return e.createTarGz(cfg, art, opts, manifest)
	case "zip":
		return e.createZip(cfg, art, opts, manifest)
	}
	return nil
}

func (e *RustEngine) runCargoPackager(command string, args []string) error {
	cmd := exec.Command("cargo", append([]string{command}, args...)...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (e *RustEngine) createTarGz(cfg *config.Config, art *config.ArtifactConfig, opts BuildOptions, manifest *cargoManifest) error {
	ext, _ := e.getExtAndPrefix(opts.OS, art.Type, "cdylib")
	finalBinaryName := cfg.Naming.Resolve(cfg.Naming.Binary, opts.ArtifactName, opts.OS, opts.Arch, manifest.Package.Version, opts.ABI, ext)

	packageName := cfg.Naming.Resolve(cfg.Naming.Package, opts.ArtifactName, opts.OS, opts.Arch, manifest.Package.Version, opts.ABI, "tar.gz")
	outPath := filepath.Join(cfg.OutputDir, packageName)

	f, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer f.Close()

	gw := gzip.NewWriter(f)
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()

	binaryPath := filepath.Join(cfg.OutputDir, finalBinaryName)
	return e.addFileToTar(tw, binaryPath, finalBinaryName)
}

func (e *RustEngine) createZip(cfg *config.Config, art *config.ArtifactConfig, opts BuildOptions, manifest *cargoManifest) error {
	ext, _ := e.getExtAndPrefix(opts.OS, art.Type, "cdylib")
	finalBinaryName := cfg.Naming.Resolve(cfg.Naming.Binary, opts.ArtifactName, opts.OS, opts.Arch, manifest.Package.Version, opts.ABI, ext)

	packageName := cfg.Naming.Resolve(cfg.Naming.Package, opts.ArtifactName, opts.OS, opts.Arch, manifest.Package.Version, opts.ABI, "zip")
	outPath := filepath.Join(cfg.OutputDir, packageName)

	f, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer f.Close()

	zw := zip.NewWriter(f)
	defer zw.Close()

	binaryPath := filepath.Join(cfg.OutputDir, finalBinaryName)
	return e.addFileToZip(zw, binaryPath, finalBinaryName)
}

func (e *RustEngine) addFileToTar(tw *tar.Writer, path, name string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return err
	}

	header, err := tar.FileInfoHeader(stat, "")
	if err != nil {
		return err
	}
	header.Name = name

	if err := tw.WriteHeader(header); err != nil {
		return err
	}

	_, err = io.Copy(tw, file)
	return err
}

func (e *RustEngine) addFileToZip(zw *zip.Writer, path, name string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	w, err := zw.Create(name)
	if err != nil {
		return err
	}

	_, err = io.Copy(w, file)
	return err
}

func (e *RustEngine) runHook(hook string) error {
	parts := strings.Fields(hook)
	if len(parts) == 0 {
		return nil
	}
	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
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

func (e *RustEngine) getBestMatch(art *config.ArtifactConfig, opts BuildOptions) *config.TargetConfig {
	var bestMatch *config.TargetConfig
	for _, tCfg := range art.Targets {
		if tCfg.OS == opts.OS && e.sliceContains(tCfg.Archs, opts.Arch) {
			if opts.ABI != "" && e.sliceContains(tCfg.ABIs, opts.ABI) {
				targetCopy := tCfg
				bestMatch = &targetCopy
				break
			}
			if opts.ABI == "" {
				targetCopy := tCfg
				bestMatch = &targetCopy
			}
		}
	}
	return bestMatch
}

func (e *RustEngine) setupEnvironment(art *config.ArtifactConfig, opts BuildOptions, target string) error {
	bestMatch := e.getBestMatch(art, opts)

	linker := ""
	if bestMatch != nil {
		if l, ok := bestMatch.LangOpts["linker"].(string); ok {
			linker = l
		}

		if depTarget, ok := bestMatch.LangOpts["deployment_target"].(string); ok {
			os.Setenv("MACOSX_DEPLOYMENT_TARGET", depTarget)
		}
		if sdk, ok := bestMatch.LangOpts["sdk_root"].(string); ok {
			os.Setenv("SDKROOT", sdk)
		}
	}

	if linker == "" && opts.OS == "linux" {
		if strings.Contains(opts.Arch, "aarch64") {
			linker = "aarch64-linux-gnu-gcc"
		} else if strings.Contains(opts.Arch, "i686") {
			linker = "i686-linux-gnu-gcc"
		}
	}

	if linker != "" {
		isArmLinker := strings.Contains(linker, "aarch64") || strings.Contains(linker, "arm")
		isArmTarget := strings.Contains(opts.Arch, "aarch64") || strings.Contains(opts.Arch, "arm")
		isX64Linker := strings.Contains(linker, "x86_64") || strings.Contains(linker, "x64")
		isX64Target := strings.Contains(opts.Arch, "x86_64") || strings.Contains(opts.Arch, "x64")

		shouldApply := true
		if (isArmLinker && !isArmTarget) || (isX64Linker && !isX64Target) {
			shouldApply = false
		}

		if shouldApply {
			envKey := fmt.Sprintf("CARGO_TARGET_%s_LINKER",
				strings.ReplaceAll(strings.ReplaceAll(strings.ToUpper(target), "-", "_"), ".", "_"))
			os.Setenv(envKey, linker)
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

func (e *RustEngine) runCargoBuild(art *config.ArtifactConfig, opts BuildOptions, target, profile string) error {
	args := []string{"build", "--target", target}
	if profile == "release" {
		args = append(args, "--release")
	} else if profile != "debug" && profile != "" {
		args = append(args, "--profile", profile)
	}

	bestMatch := e.getBestMatch(art, opts)

	if bestMatch != nil {
		if features, ok := bestMatch.LangOpts["features"].([]any); ok {
			for _, f := range features {
				if fs, ok := f.(string); ok {
					args = append(args, "--features", fs)
				}
			}
		}
		if features, ok := bestMatch.LangOpts["features"].(string); ok {
			args = append(args, "--features", features)
		}

		if tags, ok := bestMatch.LangOpts["tags"].([]any); ok {
			for _, t := range tags {
				if ts, ok := t.(string); ok {
					args = append(args, "--features", ts)
				}
			}
		}
		if tags, ok := bestMatch.LangOpts["tags"].(string); ok {
			args = append(args, "--features", tags)
		}

		if all, ok := bestMatch.LangOpts["all-features"].(bool); ok && all {
			args = append(args, "--all-features")
		}
		if noDefault, ok := bestMatch.LangOpts["no-default-features"].(bool); ok && noDefault {
			args = append(args, "--no-default-features")
		}
	}

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

func (e *RustEngine) moveArtifacts(cfg *config.Config, art *config.ArtifactConfig, opts BuildOptions, target, version, profile string, manifest *cargoManifest) error {
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
	if profile == "debug" {
		cargoProfileDir = "debug"
	} else if profile == "release" {
		cargoProfileDir = "release"
	}

	for _, bt := range buildTypes {
		ext, prefix := e.getExtAndPrefix(opts.OS, art.Type, bt)
		finalName := cfg.Naming.Resolve(cfg.Naming.Binary, opts.ArtifactName, opts.OS, opts.Arch, version, opts.ABI, ext)

		realSrcName := opts.ArtifactName
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
		return fmt.Errorf("no artifacts found for %s in %s", opts.ArtifactName, target)
	}
	return nil
}

func (e *RustEngine) getExtAndPrefix(osName, artType, format string) (string, string) {
	var ext, prefix string
	if artType == "lib" {
		prefix = "lib"
		switch osName {
		case "windows":
			prefix = ""
			if format == "staticlib" {
				ext = "lib"
			} else {
				ext = "dll"
			}
		case "darwin":
			if format == "staticlib" {
				ext = "a"
			} else {
				ext = "dylib"
			}
		case "wasm", "wasi":
			prefix = ""
			ext = "wasm"
		default:
			if format == "staticlib" {
				ext = "a"
			} else {
				ext = "so"
			}
		}
	} else {
		switch osName {
		case "windows":
			ext = "exe"
		case "wasm", "wasi":
			ext = "wasm"
		}
	}
	return ext, prefix
}

func (e *RustEngine) sliceContains(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}
