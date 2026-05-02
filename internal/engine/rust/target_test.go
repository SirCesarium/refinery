package rust

import (
	"testing"

	"github.com/SirCesarium/refinery/internal/config"
)

func TestResolveTarget(t *testing.T) {
	e := &RustEngine{}
	tests := []struct {
		os, arch, abi string
		expected      string
	}{
		{"linux", "x86_64", "gnu", "x86_64-unknown-linux-gnu"},
		{"linux", "aarch64", "musl", "aarch64-unknown-linux-musl"},
		{"darwin", "x86_64", "", "x86_64-apple-darwin"},
		{"darwin", "aarch64", "", "aarch64-apple-darwin"},
		{"windows", "x86_64", "gnu", "x86_64-pc-windows-gnu"},
		{"windows", "x86_64", "msvc", "x86_64-pc-windows-msvc"},
		{"wasm", "wasm32", "", "wasm32-unknown-unknown"},
		{"wasi", "wasm32", "", "wasm32-wasip1"},
	}

	for _, tt := range tests {
		res := e.resolveTarget(config.TargetConfig{OS: tt.os}, tt.arch, tt.abi)
		if res != tt.expected {
			t.Errorf("resolveTarget(%s, %s, %s) = %s; want %s", tt.os, tt.arch, tt.abi, res, tt.expected)
		}
	}
}

func TestGetBestMatch(t *testing.T) {
	e := &RustEngine{}
	art := &config.ArtifactConfig{
		Targets: map[string]config.TargetConfig{
			"linux-x64": {OS: "linux", Archs: []string{"x86_64"}, ABIs: []string{"gnu", "musl"}},
			"linux-arm": {OS: "linux", Archs: []string{"aarch64"}, ABIs: []string{"gnu"}},
			"win":       {OS: "windows", Archs: []string{"x86_64"}},
		},
	}

	tests := []struct {
		os, arch, abi string
		found         bool
	}{
		{"linux", "x86_64", "gnu", true},
		{"linux", "x86_64", "musl", true},
		{"linux", "x86_64", "unknown", false},
		{"linux", "aarch64", "gnu", true},
		{"linux", "aarch64", "musl", false},
		{"windows", "x86_64", "", true},
		{"darwin", "x86_64", "", false},
	}

	for _, tt := range tests {
		match := e.getBestMatch(art, tt.os, tt.arch, tt.abi)
		if (match != nil) != tt.found {
			t.Errorf("getBestMatch(%s, %s, %s) found=%v; want %v", tt.os, tt.arch, tt.abi, match != nil, tt.found)
		}
	}
}

func TestValidateTriple(t *testing.T) {
	e := &RustEngine{}
	tests := []struct {
		os, arch, abi string
		wantErr       bool
	}{
		{"darwin", "x86_64", "", false},
		{"darwin", "arm64", "", true}, // expect 'aarch64' not 'arm64'
		{"windows", "aarch64", "gnu", true},
		{"windows", "x86_64", "msvc", false},
		{"wasm", "wasm32", "", false},
		{"wasm", "x86_64", "", true},
	}

	for _, tt := range tests {
		err := e.validateTriple(tt.os, tt.arch, tt.abi)
		if (err != nil) != tt.wantErr {
			t.Errorf("validateTriple(%s, %s, %s) error=%v; wantErr %v", tt.os, tt.arch, tt.abi, err, tt.wantErr)
		}
	}
}
