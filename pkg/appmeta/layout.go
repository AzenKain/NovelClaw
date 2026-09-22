package appmeta

import (
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"

	"novelclaw/pkg/paths"
)

// migrateLegacyData copies legacy cwd-relative data (souls, skills, data)
// into the canonical DataDir exactly once, so an existing portable install
// keeps its work after moving to ~/.novelclaw on macOS/Linux. Windows
// returns immediately because DataDir is "." (nothing to migrate).
//
// Migration runs BEFORE paths.Ensure so the canonical directories are not
// yet scaffolded (Ensure creates them empty, which would make every
// destination look "already populated"). A destination is skipped only when
// it already contains REAL files — never overwrite user data; a pure-empty
// scaffold left by a partial run is discarded and re-filled so the first
// run self-heals.
func migrateLegacyData() {
	root := paths.DataDir()
	if root == "." || root == "" {
		return
	}

	for _, sub := range []string{"souls", "skills", "data"} {
		src := filepath.Join(".", sub)
		info, err := os.Stat(src)
		if err != nil || !info.IsDir() {
			continue
		}
		dst := filepath.Join(root, sub)

		switch sub {
		case "data":
			// Data is treated as a whole: if any real file already exists at
			// the canonical location, do not touch it (the DB must never be
			// overwritten or merged).
			if dirHasFiles(dst) {
				continue
			}
			_ = os.RemoveAll(dst)
			if err := copyDir(src, dst); err != nil {
				log.Warn().Err(err).Str("src", src).Str("dst", dst).Msg("appmeta: legacy data migration failed")
				continue
			}
		default:
			// souls/ and skills/ merge additively: copy only files missing at
			// the destination so user edits and customs survive updates.
			if err := mergeDir(src, dst); err != nil {
				log.Warn().Err(err).Str("src", src).Str("dst", dst).Msg("appmeta: legacy data migration failed")
				continue
			}
		}
		log.Info().Str("src", src).Str("dst", dst).Msg("appmeta: migrated legacy data directory")
	}
}

// dirHasFiles reports whether path contains at least one regular file
// anywhere in its tree (directories alone, e.g. empty scaffolds, do not
// count as content).
func dirHasFiles(path string) bool {
	entries, err := os.ReadDir(path)
	if err != nil {
		return false
	}
	for _, e := range entries {
		if !e.IsDir() {
			return true
		}
		if dirHasFiles(filepath.Join(path, e.Name())) {
			return true
		}
	}
	return false
}

// mergeDir copies entries from src into dst only when dst does not already
// contain them. Existing user files are never overwritten.
func mergeDir(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dst, info.Mode().Perm()); err != nil {
		return err
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, e := range entries {
		s := filepath.Join(src, e.Name())
		d := filepath.Join(dst, e.Name())
		if e.IsDir() {
			if err := mergeDir(s, d); err != nil {
				return err
			}
			continue
		}
		if _, err := os.Stat(d); err == nil {
			continue // keep the user's copy
		}
		data, err := os.ReadFile(s)
		if err != nil {
			return err
		}
		if err := os.WriteFile(d, data, e.Type().Perm()); err != nil {
			return err
		}
	}
	return nil
}

// copyDir recursively copies a directory tree, preserving permissions.
func copyDir(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dst, info.Mode().Perm()); err != nil {
		return err
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, e := range entries {
		s := filepath.Join(src, e.Name())
		d := filepath.Join(dst, e.Name())
		if e.IsDir() {
			if err := copyDir(s, d); err != nil {
				return err
			}
			continue
		}
		data, err := os.ReadFile(s)
		if err != nil {
			return err
		}
		if err := os.WriteFile(d, data, e.Type().Perm()); err != nil {
			return err
		}
	}
	return nil
}

// EnsureDataLayout migrates legacy cwd-relative data first, then creates
// the canonical data tree and logs the version/platform/data report.
func EnsureDataLayout() {
	migrateLegacyData()
	if _, err := paths.Ensure(); err != nil {
		log.Warn().Err(err).Msg("appmeta: failed to create data layout")
	}
	LogVersion()
}

// Report is a small informational struct surfaced via logs (and, in future,
// the UI "About" panel). It describes the running build and where user data
// lives.
type Report struct {
	Version       string `json:"version"`
	UpdateRepo    string `json:"update_repo"`
	Platform      string `json:"platform"`
	DataDir       string `json:"data_dir"`
	DBPath        string `json:"db_path"`
	SoulsDir      string `json:"souls_dir"`
	SkillsDir     string `json:"skills_dir"`
	ChecksumAsset string `json:"checksum_asset"`
}

// VersionReport builds the report for the current process.
func VersionReport() Report {
	return Report{
		Version:       CurrentVersion(),
		UpdateRepo:    UpdateRepo,
		Platform:      PlatformInfo(),
		DataDir:       paths.DataDir(),
		DBPath:        paths.DB(),
		SoulsDir:      paths.Souls(),
		SkillsDir:     paths.Skills(),
		ChecksumAsset: ChecksumAsset,
	}
}

// LogVersion logs the version/platform/data report via zerolog.
func LogVersion() {
	info := VersionReport()
	log.Info().
		Str("version", info.Version).
		Str("update_repo", info.UpdateRepo).
		Str("platform", info.Platform).
		Str("data_dir", info.DataDir).
		Str("db", info.DBPath).
		Msg("NovelClaw starting")
}
