package engine

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/SirCesarium/refinery/internal/config"
)

type RustEngine struct{}

func (e *RustEngine) ID() string {
	return "rust"
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
	if art.Type == "lib" {
		args = append(args, "--lib")
	} else {
		args = append(args, "--bin", opts.ArtifactName)
	}

	cmd := exec.Command("cargo", args...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()

	if err := cmd.Run(); err != nil {
		return err
	}

	ext := ""
	if opts.OS == "windows" {
		ext = "exe"
	}

	finalName := cfg.Naming.Resolve(cfg.Naming.Binary, opts.ArtifactName, opts.OS, opts.Arch, "0.0.0", opts.ABI, ext)
	srcBinary := opts.ArtifactName
	if opts.OS == "windows" {
		srcBinary += ".exe"
	}

	srcPath := filepath.Join("target", targetTriple, "release", srcBinary)
	distPath := filepath.Join(cfg.OutputDir, finalName)

	if err := os.MkdirAll(cfg.OutputDir, 0755); err != nil {
		return err
	}

	return os.Rename(srcPath, distPath)
}
