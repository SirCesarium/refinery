package goengine

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/SirCesarium/refinery/internal/config"
)

// pkg dispatches packaging to the correct format handler.
func (e *GoEngine) pkg(cfg *config.Config, art *config.ArtifactConfig, artifactName, osName, arch, abi, format string) error {
	switch format {
	case "tar.gz", "targz":
		return e.createTarGz(cfg, art, artifactName, osName, arch, abi)
	case "zip":
		return e.createZip(cfg, art, artifactName, osName, arch, abi)
	default:
		return fmt.Errorf("unsupported package format: %s", format)
	}
}

// createTarGz creates a .tar.gz package from built artifacts.
func (e *GoEngine) createTarGz(cfg *config.Config, art *config.ArtifactConfig, artifactName, osName, arch, abi string) error {
	packageName := cfg.Naming.Resolve(cfg.Naming.Package, artifactName, osName, arch, "0.0.0", abi, "tar.gz")
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

	return e.archiveArtifactFiles(tw, nil, cfg, art, artifactName, osName, arch, abi)
}

// createZip creates a .zip package from built artifacts.
func (e *GoEngine) createZip(cfg *config.Config, art *config.ArtifactConfig, artifactName, osName, arch, abi string) error {
	packageName := cfg.Naming.Resolve(cfg.Naming.Package, artifactName, osName, arch, "0.0.0", abi, "zip")
	outPath := filepath.Join(cfg.OutputDir, packageName)

	f, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer f.Close()

	zw := zip.NewWriter(f)
	defer zw.Close()

	return e.archiveArtifactFiles(nil, zw, cfg, art, artifactName, osName, arch, abi)
}

// archiveArtifactFiles adds built artifacts to a tar or zip archive.
func (e *GoEngine) archiveArtifactFiles(tw *tar.Writer, zw *zip.Writer, cfg *config.Config, art *config.ArtifactConfig, artifactName, osName, arch, abi string) error {
	ext, _ := e.getExtAndPrefix(osName, art.Type)
	finalName := cfg.Naming.Resolve(cfg.Naming.Binary, artifactName, osName, arch, "0.0.0", abi, ext)

	filePath := filepath.Join(cfg.OutputDir, finalName)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("artifact not found for archiving: %s", filePath)
	}

	if tw != nil {
		if err := e.addFileToTar(tw, filePath, finalName); err != nil {
			return err
		}
	} else {
		if err := e.addFileToZip(zw, filePath, finalName); err != nil {
			return err
		}
	}

	if art.Headers {
		headers, err := filepath.Glob("*.go")
		if err != nil {
			return fmt.Errorf("failed to search for .go headers: %w", err)
		}
		for _, h := range headers {
			if tw != nil {
				if err := e.addFileToTar(tw, h, h); err != nil {
					return fmt.Errorf("failed to add header %s to tar: %w", h, err)
				}
			} else {
				if err := e.addFileToZip(zw, h, h); err != nil {
					return fmt.Errorf("failed to add header %s to zip: %w", h, err)
				}
			}
		}
	}

	return nil
}

// addFileToTar adds a single file to a tar archive.
func (e *GoEngine) addFileToTar(tw *tar.Writer, path, name string) error {
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

// addFileToZip adds a single file to a zip archive.
func (e *GoEngine) addFileToZip(zw *zip.Writer, path, name string) error {
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
