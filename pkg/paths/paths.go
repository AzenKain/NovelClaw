// Package paths centralises every filesystem location NovelClaw uses.
//
// Layout policy (v1):
//   - Windows: everything lives in "./" (next to the executable) so a
//     portable / USB install keeps data colocated with the binary.
//   - macOS & Linux: everything lives under ~/.novelclaw/ so updates and
//     re-installs never touch user data.
//
// Nothing else in the codebase should hardcode "./data", "./souls",
// "./skills" or "~/.novelclaw" — resolve paths only through this package so
// the self-update + storage migration story stays coherent.
package paths

import (
	"os"
	"path/filepath"
	"runtime"
)

// DataDir returns the root directory for all NovelClaw user data.
//
//	Windows: "."                       (cwd — portable install)
//	Darwin : $HOME/.novelclaw
//	Linux  : $HOME/.novelclaw
//
// Falls back to "." if the home directory cannot be resolved.
func DataDir() string {
	if runtime.GOOS == "windows" {
		return "."
	}
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return "."
	}
	return filepath.Join(home, ".novelclaw")
}

// Dir returns a subpath under DataDir. Absolute paths win, relative paths
// are anchored at DataDir, and the result is always cleaned.
func Dir(parts ...string) string {
	if len(parts) == 0 {
		return DataDir()
	}
	joined := filepath.Join(parts...)
	if filepath.IsAbs(joined) {
		return filepath.Clean(joined)
	}
	return filepath.Join(DataDir(), filepath.Clean(joined))
}

// Resolve rewrites a stored / user-supplied path through the DataDir policy.
// Absolute paths pass through unchanged; relative paths are anchored at
// DataDir.
func Resolve(path string) string {
	if path == "" {
		return DataDir()
	}
	if filepath.IsAbs(path) {
		return filepath.Clean(path)
	}
	return filepath.Join(DataDir(), filepath.Clean(path))
}

// DB returns the path of the SQLite database.
func DB() string { return Dir("data", "novelclaw.db") }

// Data returns the root data (books, exports, benchmark) directory.
func Data() string { return Dir("data") }

// Souls returns the user-writable souls directory.
func Souls() string { return Dir("souls") }

// Skills returns the user-writable skills directory.
func Skills() string { return Dir("skills") }

// SkillsDefault returns the shipped (read-only) skill baseline.
func SkillsDefault() string { return Dir("skills", "default") }

// SkillsEvolved returns the per-project evolved-skills parent.
func SkillsEvolved() string { return Dir("skills", "evolved") }

// Ensure creates the canonical data-directory tree and returns its root.
// It never fails on Windows (cwd is assumed writable).
func Ensure() (string, error) {
	root := DataDir()
	if root == "" || root == "." {
		return root, nil
	}
	for _, d := range []string{".", "data", "souls", "skills", filepath.Join("skills", "default"), filepath.Join("skills", "evolved")} {
		if err := os.MkdirAll(filepath.Join(root, d), 0o755); err != nil {
			return "", err
		}
	}
	return root, nil
}
