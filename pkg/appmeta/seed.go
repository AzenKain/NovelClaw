package appmeta

import (
	"io/fs"
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"

	"novelclaw/pkg/paths"
)

type SeedOptions struct {
	FS           fs.FS
	SkillsRoot   string
	SoulsRoot    string
	TargetSkills string
	TargetSouls  string
}

func SeedDefaults(opts SeedOptions) {
	if opts.FS == nil {
		return
	}
	targetSkills := opts.TargetSkills
	if targetSkills == "" {
		targetSkills = paths.Skills()
	}
	targetSouls := opts.TargetSouls
	if targetSouls == "" {
		targetSouls = paths.Souls()
	}

	if n := seedTree(opts.FS, opts.SkillsRoot, targetSkills, "skills"); n > 0 {
		log.Info().Int("files", n).Str("dst", targetSkills).Msg("appmeta: seeded default skills")
	}
	if n := seedTree(opts.FS, opts.SoulsRoot, targetSouls, "souls"); n > 0 {
		log.Info().Int("files", n).Str("dst", targetSouls).Msg("appmeta: seeded default souls")
	}
}

func seedTree(fsys fs.FS, src, dst, label string) int {
	if src == "" {
		return 0
	}
	if _, err := fs.Stat(fsys, src); err != nil {
		log.Debug().Err(err).Str("src", src).Msg("appmeta: embedded defaults unavailable")
		return 0
	}

	written := 0
	walkErr := fs.WalkDir(fsys, src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, relErr := filepath.Rel(src, path)
		if relErr != nil {
			return relErr
		}
		target := filepath.Join(dst, rel)

		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		// Never clobber an existing file — it may hold user edits.
		if _, statErr := os.Stat(target); statErr == nil {
			return nil
		}
		if !d.Type().IsRegular() {
			return nil
		}

		data, readErr := fs.ReadFile(fsys, path)
		if readErr != nil {
			log.Warn().Err(readErr).Str("src", path).Msg("appmeta: read embedded default failed")
			return nil
		}
		if mkErr := os.MkdirAll(filepath.Dir(target), 0o755); mkErr != nil {
			log.Warn().Err(mkErr).Str("dst", target).Msg("appmeta: create seed dir failed")
			return nil
		}
		if writeErr := os.WriteFile(target, data, 0o644); writeErr != nil {
			log.Warn().Err(writeErr).Str("dst", target).Msg("appmeta: write seeded default failed")
			return nil
		}
		written++
		return nil
	})
	if walkErr != nil {
		log.Warn().Err(walkErr).Str("tree", label).Msg("appmeta: seeding walk failed")
	}
	return written
}

func SeedSkillBaselineFromDefaults(fsys fs.FS) {
	SeedDefaults(SeedOptions{
		FS:         fsys,
		SkillsRoot: "skills",
		SoulsRoot:  "souls",
	})
}
