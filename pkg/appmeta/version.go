// Package appmeta holds the application's identity: the version, the storage
// layout report, and the self-update wiring.
package appmeta

import (
	"fmt"
	"runtime"
)

// Version is the single source of truth for the running NovelClaw version.
//
// Bump it by hand when cutting a release — there is no build-time injection,
// no -ldflags, no environment variable. The git tag must match this value:
//
//	Version = "1.2.3"  ↔  tag v1.2.3
//
// The self-updater compares this string against the GitHub release tag, so a
// mismatch means clients will either miss the update or re-install the same
// build forever.
const Version = "1.0.0"

// CurrentVersion returns the version of the running build.
func CurrentVersion() string { return Version }

// FullName is the product name shown to users and in release assets.
func FullName() string { return "NovelClaw" }

// UpdateRepo is the GitHub repository ("owner/repo") that hosts NovelClaw
// release assets.
const UpdateRepo = "AzenKain/NovelClaw"

// ChecksumAsset is the name of the checksum manifest published alongside every
// release. The updater verifies each downloaded asset against it.
const ChecksumAsset = "SHA256SUMS"

// PlatformInfo returns the runtime GOOS/GOARCH pair, e.g. "linux/amd64".
func PlatformInfo() string { return fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH) }
