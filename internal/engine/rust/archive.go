package rust

import (
	"archive/tar"
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/SirCesarium/refinery/internal/config"
)

func (e *RustEngine) archiveArtifactFiles(tw *tar.Writer, zw *zip.Writer, cfg *config.Config, art *config.ArtifactConfig, artifactName, osName, arch, abi string, manifest *cargoManifest) error {
	var buildTypes []string
	if art.Type == "bin" {
		buildTypes = []string{"bin"}
	} else {
		buildTypes = art.LibraryTypes
		if len(buildTypes) == 0 {
			buildTypes = []string{"cdylib"}
		}
	}

	for _, bt := range buildTypes {
		ext, _ := e.getExtAndPrefix(osName, art.Type, bt)
		finalName := cfg.Naming.Resolve(cfg.Naming.Binary, artifactName, osName, arch, manifest.Package.Version, abi, ext)
		filePath := filepath.Join(cfg.OutputDir, finalName)

		if _, err := os.Stat(filePath); err == nil {
			if tw != nil {
				if err := e.addFileToTar(tw, filePath, finalName); err != nil {
					return err
				}
			} else {
				if err := e.addFileToZip(zw, filePath, finalName); err != nil {
					return err
				}
			}
		}
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
