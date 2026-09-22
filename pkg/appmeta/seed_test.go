package appmeta

import (
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"novelclaw/pkg/skills"
	"novelclaw/pkg/soul"
)

// fakeDefaults mirrors the shipped tree shape: skills/default/*.skill.md,
// skills/default/SOUL.md and souls/*.soul.md.
func fakeDefaults() fstest.MapFS {
	return fstest.MapFS{
		"skills/default/app_navigation.skill.md": &fstest.MapFile{Data: []byte(`---
id: app_navigation
name: Navigation
category: ui_control
tools:
  - switch_tab
---
Navigate the UI.
`)},
		"skills/default/SOUL.md": &fstest.MapFile{Data: []byte("# Persona\nSystem prompt.")},
		"souls/soul_neko.soul.md": &fstest.MapFile{Data: []byte(`---
id: soul_neko
name: Neko
avatar: "🐱"
is_default: true
---
# Persona: Neko
Body.
`)},
	}
}

func TestSeedDefaultsWritesMissingFiles(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	skillsDir := filepath.Join(home, ".novelclaw", "skills")
	soulsDir := filepath.Join(home, ".novelclaw", "souls")

	SeedDefaults(SeedOptions{
		FS:           fakeDefaults(),
		SkillsRoot:   "skills",
		SoulsRoot:    "souls",
		TargetSkills: skillsDir,
		TargetSouls:  soulsDir,
	})

	// skill playbook + SOUL.md materialised
	for _, want := range []string{
		filepath.Join(skillsDir, "default", "app_navigation.skill.md"),
		filepath.Join(skillsDir, "default", "SOUL.md"),
		filepath.Join(soulsDir, "soul_neko.soul.md"),
	} {
		if _, err := os.Stat(want); err != nil {
			t.Fatalf("expected seeded file %s: %v", want, err)
		}
	}

	// and they are loadable by the real loaders
	loaded, err := skills.NewSkillLoader(filepath.Join(skillsDir, "default")).LoadAll()
	if err != nil {
		t.Fatalf("LoadAll: %v", err)
	}
	if len(loaded) != 1 || loaded[0].Metadata.ID != "app_navigation" {
		t.Fatalf("unexpected skills: %+v", loaded)
	}
	// The soul Manager additionally injects the two built-in personas when
	// they are absent, so assert the SEEDED soul is present rather than an
	// exact count.
	mgr := soul.NewManager(soulsDir)
	if _, ok := mgr.Get("soul_neko"); !ok {
		t.Fatalf("seeded soul not registered; got %d souls", len(mgr.List()))
	}
}

func TestSeedDefaultsNeverOverwritesUserFiles(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	skillsDir := filepath.Join(home, ".novelclaw", "skills")
	soulsDir := filepath.Join(home, ".novelclaw", "souls")
	if err := os.MkdirAll(filepath.Join(skillsDir, "default"), 0o755); err != nil {
		t.Fatal(err)
	}

	edited := filepath.Join(skillsDir, "default", "app_navigation.skill.md")
	userContent := []byte("USER EDITED CONTENT")
	if err := os.WriteFile(edited, userContent, 0o644); err != nil {
		t.Fatal(err)
	}

	SeedDefaults(SeedOptions{
		FS: fakeDefaults(), SkillsRoot: "skills", SoulsRoot: "souls",
		TargetSkills: skillsDir, TargetSouls: soulsDir,
	})

	got, err := os.ReadFile(edited)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(userContent) {
		t.Fatalf("seeder overwrote user file: got %q", got)
	}
	// A file that was missing is still filled in.
	if _, err := os.Stat(filepath.Join(skillsDir, "default", "SOUL.md")); err != nil {
		t.Fatalf("missing file was not seeded: %v", err)
	}
}

func TestSeedDefaultsNilFSIsSafe(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	SeedDefaults(SeedOptions{TargetSkills: filepath.Join(home, "s"), TargetSouls: filepath.Join(home, "o")})
}
