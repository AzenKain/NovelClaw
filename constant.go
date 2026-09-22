// Package main bridge for legacy path/version helpers.
//
// constant.go is kept as a thin compatibility shim so existing callers
// (elevate_*.go, main.go, cmd/ tools) do not need to be rewritten in one
// step. The canonical storage-layout and version policy lives in
// pkg/paths and pkg/appmeta — new code should use those packages directly.
package main

import (
	"novelclaw/pkg/appmeta"
	"novelclaw/pkg/paths"
)

// CurrentVersion is the current NovelClaw version. Prefer
// appmeta.CurrentVersion() — this wrapper exists for legacy callers.
const CurrentVersion = appmeta.Version

// GetDataDir returns the canonical NovelClaw data directory:
//
//	Windows: "."
//	macOS  : $HOME/.novelclaw
//	Linux  : $HOME/.novelclaw
//
// Delegates to paths.DataDir().
func GetDataDir() string { return paths.DataDir() }

// ResolvePath anchors a relative path in the data directory; absolute paths
// pass through unchanged. Delegates to paths.Resolve().
func ResolvePath(path string) string { return paths.Resolve(path) }
