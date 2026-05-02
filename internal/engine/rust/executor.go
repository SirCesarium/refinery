package rust

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/SirCesarium/refinery/internal/config"
)

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

func (e *RustEngine) addTarget(target string) error {
	cmd := exec.Command("rustup", "target", "add", target)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (e *RustEngine) runCargoBuild(art *config.ArtifactConfig, artifactName, osName, arch, abi, target, profile string) error {
	args := []string{"build", "--target", target}
	if profile == "release" {
		args = append(args, "--release")
	} else if profile != "debug" && profile != "" {
		args = append(args, "--profile", profile)
	}

	bestMatch := e.getBestMatch(art, osName, arch, abi)

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
		args = append(args, "--bin", artifactName)
	}

	cmd := exec.Command("cargo", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()
	return cmd.Run()
}


func (e *RustEngine) runCargoPackager(command string, args []string) error {
	cmd := exec.Command("cargo", append([]string{command}, args...)...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (e *RustEngine) setupEnvironment(art *config.ArtifactConfig, osName, arch, abi, target string) error {
	bestMatch := e.getBestMatch(art, osName, arch, abi)

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

	if linker == "" && osName == "linux" {
		if strings.Contains(arch, "aarch64") {
			linker = "aarch64-linux-gnu-gcc"
		} else if strings.Contains(arch, "i686") {
			linker = "i686-linux-gnu-gcc"
		}
	}

	if linker != "" {
		isArmLinker := strings.Contains(linker, "aarch64") || strings.Contains(linker, "arm")
		isArmTarget := strings.Contains(arch, "aarch64") || strings.Contains(arch, "arm")
		isX64Linker := strings.Contains(linker, "x86_64") || strings.Contains(linker, "x64")
		isX64Target := strings.Contains(arch, "x86_64") || strings.Contains(arch, "x64")

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

	if osName == "darwin" && os.Getenv("MACOSX_DEPLOYMENT_TARGET") == "" {
		os.Setenv("MACOSX_DEPLOYMENT_TARGET", "11.0")
	}
	return nil
}
