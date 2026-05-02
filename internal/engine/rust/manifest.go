package rust

import (
	"fmt"
	"os"

	"github.com/pelletier/go-toml/v2"
)

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
