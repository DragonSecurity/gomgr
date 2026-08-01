package version

import (
	"runtime/debug"
)

// stamped is set at link time by the release build with
// -ldflags "-X github.com/DragonSecurity/gomgr/internal/version.stamped=vX.Y.Z"
// so a released binary reports its tag. Nothing else can supply it: the module
// version recorded in the binary is only meaningful for `go install
// module@version`, and the VCS revision is a commit hash, not a tag.
var stamped string

// BuildInfo contains version and build information
type BuildInfo struct {
	Version    string
	Revision   string
	Modified   bool
	CommitTime string
}

// GetBuildInfo returns version information from runtime/debug
func GetBuildInfo() BuildInfo {
	info := BuildInfo{
		Version: "dev",
	}

	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		return info
	}

	// Extract VCS information from build settings
	for _, setting := range buildInfo.Settings {
		switch setting.Key {
		case "vcs.revision":
			info.Revision = setting.Value
			if len(info.Revision) > 12 {
				info.Revision = info.Revision[:12]
			}
		case "vcs.time":
			info.CommitTime = setting.Value
		case "vcs.modified":
			info.Modified = setting.Value == "true"
		}
	}

	// A link-time stamp is the only source that carries the release tag, so it
	// wins over the revision and module fallbacks below.
	if stamped != "" {
		info.Version = stamped
		return info
	}

	// Use VCS revision as version if available, otherwise check for version in main module
	if info.Revision != "" {
		if info.Modified {
			info.Version = info.Revision + "-dirty"
		} else {
			info.Version = info.Revision
		}
	} else if buildInfo.Main.Version != "" && buildInfo.Main.Version != "(devel)" {
		info.Version = buildInfo.Main.Version
	}

	return info
}
