package version

import (
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
)

// InstallMethod identifies how the CLI binary was installed.
// Mirrors Python's _InstallMethod enum in src/specify_cli/_version.py.
type InstallMethod int

const (
	InstallUnsupported InstallMethod = iota
	InstallUVTool
	InstallPipx
	InstallNPM
	InstallUVXEphemeral
	InstallSourceCheckout
)

func (m InstallMethod) String() string {
	switch m {
	case InstallUVTool:
		return "uv-tool"
	case InstallPipx:
		return "pipx"
	case InstallNPM:
		return "npm"
	case InstallUVXEphemeral:
		return "uvx-ephemeral"
	case InstallSourceCheckout:
		return "source-checkout"
	default:
		return "unsupported"
	}
}

// installerPrefixes maps installer prefix substrings to InstallMethod values.
// Matches Python _INSTALLER_PATH_PREFIXES.
var installerPrefixes = []struct {
	substr string
	method InstallMethod
}{
	{"/uv/tools/", InstallUVTool},
	{"/.local/pipx/", InstallPipx},
	{"/pipx/", InstallPipx},
	{"/node_modules/", InstallNPM},
	{"/.local/bin/", InstallUVXEphemeral},
}

// DetectInstallMethod runs 3-tier detection (mirrors Python _detect_install_method).
func DetectInstallMethod() InstallMethod {
	exe, err := os.Executable()
	if err != nil {
		return InstallUnsupported
	}
	resolved, err := filepath.EvalSymlinks(exe)
	if err != nil {
		resolved = exe
	}
	for _, p := range installerPrefixes {
		if strings.Contains(resolved, p.substr) {
			return p.method
		}
	}

	// Tier 2: source checkout via ReadBuildInfo module path.
	info, ok := debug.ReadBuildInfo()
	if ok {
		if strings.HasPrefix(info.Path, "github.com/") || strings.Contains(info.Path, "/specify-go") {
			return InstallSourceCheckout
		}
	}

	return InstallUnsupported
}
