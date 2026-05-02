package cli

import (
	"github.com/SirCesarium/refinery/internal/app"
	"github.com/SirCesarium/refinery/internal/config"
	"github.com/SirCesarium/refinery/internal/engine"
	"github.com/SirCesarium/refinery/internal/ui"
	"github.com/spf13/cobra"
)

var (
	artName string
	osName  string
	arch    string
	abi     string
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build and package an artifact based on configuration",
	Run: func(cmd *cobra.Command, args []string) {
		ui.Section("Initialization")
		cfg, err := config.Load("refinery.toml")
		if err != nil {
			ui.Fatal(err, "Ensure 'refinery.toml' exists and is valid. Run 'refinery init' if you haven't yet.")
		}

		art, ok := cfg.Artifacts[artName]
		if !ok {
			ui.Fatal(nil, "Artifact '"+artName+"' not found in config. Check your refinery.toml.")
		}

		registry := app.NewDefaultEngineRegistry()
		eng := registry.Get(cfg.Project.Lang)
		if eng == nil {
			ui.Fatal(nil, "No engine found for language: "+cfg.Project.Lang)
		}

		opts := engine.BuildOptions{
			ArtifactName: artName,
			OS:           osName,
			Arch:         arch,
			ABI:          abi,
		}

		ui.Section("Building")
		ui.Info("Target: %s-%s-%s", osName, arch, abi)
		if err := eng.Build(cfg, art, opts); err != nil {
			ui.Fatal(err, "The build command failed. Check the logs above for specific toolchain errors (e.g., cargo, gcc).")
		}
		ui.Success("Build completed successfully")

		packageFormats := mergePackages(art.Packages, findTargetPackages(art, osName, arch, abi))
		if len(packageFormats) > 0 {
			ui.Section("Packaging")
			for _, format := range packageFormats {
				ui.Info("Packaging as %s...", format)
				if err := eng.Package(cfg, art, opts, format); err != nil {
					ui.Fatal(err, "Packaging failed for "+format+". Ensure you have the necessary packager tools installed.")
				}
				ui.Success("Packaged as %s", format)
			}
		}

		ui.Section("Finalization")
		ui.Success("All tasks completed successfully for %s", artName)
	},
}

func init() {
	buildCmd.Flags().StringVar(&artName, "artifact", "", "Artifact name to build")
	buildCmd.Flags().StringVar(&osName, "os", "", "Target OS")
	buildCmd.Flags().StringVar(&arch, "arch", "", "Target Architecture")
	buildCmd.Flags().StringVar(&abi, "abi", "", "Target ABI")

	buildCmd.MarkFlagRequired("artifact")
	buildCmd.MarkFlagRequired("os")
	buildCmd.MarkFlagRequired("arch")

	rootCmd.AddCommand(buildCmd)
}

func mergePackages(global, target []string) []string {
	merged := make([]string, 0, len(global)+len(target))
	seen := map[string]bool{}
	for _, p := range global {
		if p == "" || seen[p] {
			continue
		}
		seen[p] = true
		merged = append(merged, p)
	}
	for _, p := range target {
		if p == "" || seen[p] {
			continue
		}
		seen[p] = true
		merged = append(merged, p)
	}
	return merged
}

func findTargetPackages(art *config.ArtifactConfig, osName, arch, abi string) []string {
	for _, tCfg := range art.Targets {
		if tCfg.OS != osName {
			continue
		}
		if !contains(tCfg.Archs, arch) {
			continue
		}
		if abi != "" && len(tCfg.ABIs) > 0 && !contains(tCfg.ABIs, abi) {
			continue
		}
		return tCfg.Packages
	}
	return nil
}

func contains(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}
