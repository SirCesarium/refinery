package goengine

import (
	"strings"

	"github.com/SirCesarium/refinery/internal/config"
)

// getBestMatch finds the target config matching OS and arch.
func (e *GoEngine) getBestMatch(art *config.ArtifactConfig, osName, arch, abi string) *config.TargetConfig {
	var bestMatch *config.TargetConfig
	for _, tCfg := range art.Targets {
		if tCfg.OS == osName && e.sliceContains(tCfg.Archs, arch) {
			if abi != "" && e.sliceContains(tCfg.ABIs, abi) {
				targetCopy := tCfg
				bestMatch = &targetCopy
				break
			}
			if abi == "" || (len(tCfg.ABIs) == 1 && tCfg.ABIs[0] == "") || len(tCfg.ABIs) == 0 {
				targetCopy := tCfg
				bestMatch = &targetCopy
				if abi == "" {
					break
				}
			}
		}
	}
	return bestMatch
}

// getExtAndPrefix returns file extension and prefix based on OS and artifact type.
func (e *GoEngine) getExtAndPrefix(osName, artType string) (string, string) {
	var ext, prefix string
	if artType == "lib" {
		prefix = "lib"
		switch osName {
		case "windows":
			ext = "dll"
		case "darwin":
			ext = "dylib"
		default:
			ext = "so"
		}
	} else {
		switch osName {
		case "windows":
			ext = "exe"
		default:
			ext = ""
		}
	}
	return ext, prefix
}

func (e *GoEngine) sliceContains(slice []string, val string) bool {
	for _, s := range slice {
		if s == val {
			return true
		}
	}
	return false
}

// getBuildTags extracts Go build tags from lang_opts.
func (e *GoEngine) getBuildTags(tCfg *config.TargetConfig) string {
	if tCfg == nil {
		return ""
	}
	if tags, ok := tCfg.LangOpts["tags"].(string); ok {
		return tags
	}
	if tags, ok := tCfg.LangOpts["tags"].([]any); ok {
		tagStrs := make([]string, 0, len(tags))
		for _, t := range tags {
			if ts, ok := t.(string); ok {
				tagStrs = append(tagStrs, ts)
			}
		}
		return strings.Join(tagStrs, ",")
	}
	return ""
}

// getLdFlags extracts Go linker flags from lang_opts.
func (e *GoEngine) getLdFlags(tCfg *config.TargetConfig) string {
	if tCfg == nil {
		return ""
	}
	if flags, ok := tCfg.LangOpts["ldflags"].(string); ok {
		return flags
	}
	return ""
}
