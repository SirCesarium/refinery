package rust

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/SirCesarium/refinery/internal/config"
)

func (e *RustEngine) pkg(cfg *config.Config, art *config.ArtifactConfig, artifactName, osName, arch, abi, format string) error {
	manifest, err := e.loadManifest()
	if err != nil {
		return err
	}

	switch format {
	case "deb":
		if osName != "linux" || abi == "musl" || arch == "wasm32" {
			return nil
		}
		if _, err := exec.LookPath("cargo-deb"); err != nil {
			return nil
		}
		bestMatch := e.getBestMatch(art, osName, arch, abi)
		target := e.resolveTarget(*bestMatch, arch, abi)
		return e.runCargoPackager("deb", []string{"--target", target})
	case "rpm":
		if osName != "linux" || abi == "musl" || arch == "wasm32" {
			return nil
		}
		if _, err := exec.LookPath("cargo-generate-rpm"); err != nil {
			return nil
		}
		bestMatch := e.getBestMatch(art, osName, arch, abi)
		target := e.resolveTarget(*bestMatch, arch, abi)
		return e.runCargoPackager("generate-rpm", []string{"--target", target})
	case "msi":
		if osName != "windows" {
			return nil
		}
		if _, err := exec.LookPath("candle"); err != nil {
			return nil
		}
		bestMatch := e.getBestMatch(art, osName, arch, abi)
		target := e.resolveTarget(*bestMatch, arch, abi)
		return e.runCargoPackager("wix", []string{"--target", target})
	case "tar.gz", "targz":
		return e.createTarGz(cfg, art, artifactName, osName, arch, abi, manifest)
	case "zip":
		return e.createZip(cfg, art, artifactName, osName, arch, abi, manifest)
	}
	return nil
}

func (e *RustEngine) createTarGz(cfg *config.Config, art *config.ArtifactConfig, artifactName, osName, arch, abi string, manifest *cargoManifest) error {
	packageName := cfg.Naming.Resolve(cfg.Naming.Package, artifactName, osName, arch, manifest.Package.Version, abi, "tar.gz")
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

	return e.archiveArtifactFiles(tw, nil, cfg, art, artifactName, osName, arch, abi, manifest)
}

func (e *RustEngine) createZip(cfg *config.Config, art *config.ArtifactConfig, artifactName, osName, arch, abi string, manifest *cargoManifest) error {
	packageName := cfg.Naming.Resolve(cfg.Naming.Package, artifactName, osName, arch, manifest.Package.Version, abi, "zip")
	outPath := filepath.Join(cfg.OutputDir, packageName)

	f, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer f.Close()

	zw := zip.NewWriter(f)
	defer zw.Close()

	return e.archiveArtifactFiles(nil, zw, cfg, art, artifactName, osName, arch, abi, manifest)
}

