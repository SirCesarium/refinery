package engine

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/SirCesarium/refinery/internal/config"
)

type RustEngine struct{}

func (e *RustEngine) ID() string {
	return "rust"
}

func (e *RustEngine) Build(cfg *config.Config, art *config.ArtifactConfig, opts BuildOptions) error {
	targetTriple := fmt.Sprintf("%s-unknown-%s", opts.Arch, opts.OS)
	if opts.ABI != "" {
		targetTriple = fmt.Sprintf("%s-%s", targetTriple, opts.ABI)
	}

	if runtime.GOOS == "linux" && opts.OS == "linux" {
		if opts.Arch == "aarch64" {
			exec.Command("sudo", "apt-get", "update").Run()
			exec.Command("sudo", "apt-get", "install", "-y", "gcc-aarch64-linux-gnu").Run()
			os.Setenv("CARGO_TARGET_AARCH64_UNKNOWN_LINUX_GNU_LINKER", "aarch64-linux-gnu-gcc")
			os.Setenv("CARGO_TARGET_AARCH64_UNKNOWN_LINUX_MUSL_LINKER", "aarch64-linux-gnu-gcc")
		}
	}

	if runtime.GOOS == "darwin" && opts.OS == "darwin" {
		if opts.Arch == "aarch64" || opts.Arch == "x86_64" {
			os.Setenv("SDKROOT", "/Library/Developer/CommandLineTools/SDKs/MacOSX.sdk")
			os.Setenv("MACOSX_DEPLOYMENT_TARGET", "11.0")
		}
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
	return cmd.Run()
}
