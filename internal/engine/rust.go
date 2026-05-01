package engine

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/SirCesarium/refinery/internal/config"
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

func (e *RustEngine) Build(cfg *config.Config, art *config.ArtifactConfig, opts BuildOptions) error {
	var targetTriple string

	switch opts.OS {
	case "darwin":
		targetTriple = fmt.Sprintf("%s-apple-darwin", opts.Arch)
	case "windows":
		targetTriple = fmt.Sprintf("%s-pc-windows", opts.Arch)
		if opts.ABI != "" {
			targetTriple = fmt.Sprintf("%s-%s", targetTriple, opts.ABI)
		}
	case "wasm":
		targetTriple = "wasm32-unknown-unknown"
	case "wasi":
		targetTriple = "wasm32-wasip1"
	default:
		targetTriple = fmt.Sprintf("%s-unknown-%s", opts.Arch, opts.OS)
		if opts.ABI != "" {
			targetTriple = fmt.Sprintf("%s-%s", targetTriple, opts.ABI)
		}
	}

	if runtime.GOOS == "linux" && opts.OS == "linux" && opts.Arch == "aarch64" {
		if _, err := exec.LookPath("aarch64-linux-gnu-gcc"); err != nil {
			exec.Command("sudo", "apt-get", "update").Run()
			exec.Command("sudo", "apt-get", "install", "-y", "gcc-aarch64-linux-gnu").Run()
		}

		os.Setenv("CARGO_TARGET_AARCH64_UNKNOWN_LINUX_GNU_LINKER", "aarch64-linux-gnu-gcc")
		os.Setenv("CARGO_TARGET_AARCH64_UNKNOWN_LINUX_MUSL_LINKER", "aarch64-linux-gnu-gcc")
	}

	if runtime.GOOS == "darwin" && opts.OS == "darwin" {
		if _, err := os.Stat("/Library/Developer/CommandLineTools/SDKs/MacOSX.sdk"); err == nil {
			os.Setenv("SDKROOT", "/Library/Developer/CommandLineTools/SDKs/MacOSX.sdk")
		}

		os.Setenv("MACOSX_DEPLOYMENT_TARGET", "11.0")
	}

	setupCmd := exec.Command("rustup", "target", "add", targetTriple)

	setupCmd.Stdout = os.Stdout
	setupCmd.Stderr = os.Stderr

	if err := setupCmd.Run(); err != nil {
		return fmt.Errorf("failed to add target %s: %w", targetTriple, err)
	}

	args := []string{"build", "--release", "--target", targetTriple}

	var srcFileName string

	if art.Type == "lib" {
		args = append(args, "--lib")
		srcFileName = strings.ReplaceAll(cfg.Project.Name, "-", "_")
	} else {
		args = append(args, "--bin", opts.ArtifactName)
		srcFileName = opts.ArtifactName
	}

	cmd := exec.Command("cargo", args...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()

	if err := cmd.Run(); err != nil {
		return err
	}

	ext := ""
	prefix := ""
	switch opts.OS {
	case "windows":
		if art.Type == "bin" {
			ext = "exe"
		} else {
			ext = "dll"
		}
	case "wasm", "wasi":
		ext = "wasm"
	case "linux", "darwin":
		if art.Type == "lib" {
			prefix = "lib"
			if opts.OS == "linux" {
				ext = "so"
			} else {
				ext = "dylib"
			}
		}
	}

	finalName := cfg.Naming.Resolve(cfg.Naming.Binary, opts.ArtifactName, opts.OS, opts.Arch, "0.0.0", opts.ABI, ext)

	realSrcName := srcFileName
	if prefix != "" {
		realSrcName = prefix + realSrcName
	}
	if ext != "" {
		realSrcName += "." + ext
	}

	srcPath := filepath.Join("target", targetTriple, "release", realSrcName)
	distPath := filepath.Join(cfg.OutputDir, finalName)

	fmt.Printf("[DEBUG] Artifact Name: %s\n", opts.ArtifactName)
	fmt.Printf("[DEBUG] Project Name: %s\n", cfg.Project.Name)
	fmt.Printf("[DEBUG] Source File Name: %s\n", realSrcName)
	fmt.Printf("[DEBUG] Expected Source Path: %s\n", srcPath)
	fmt.Printf("[DEBUG] Destination Path: %s\n", distPath)

	if err := os.MkdirAll(cfg.OutputDir, 0755); err != nil {
		return err
	}

	if _, err := os.Stat(srcPath); os.IsNotExist(err) {
		fmt.Printf("[DEBUG] Source path not found. Contents of %s:\n", filepath.Dir(srcPath))
		files, _ := os.ReadDir(filepath.Dir(srcPath))
		for _, f := range files {
			fmt.Printf(" - %s\n", f.Name())
		}
		return fmt.Errorf("artifact not found at %s", srcPath)
	}

	return os.Rename(srcPath, distPath)
}
