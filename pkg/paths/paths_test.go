package paths

import (
	"path/filepath"
	"runtime"
	"testing"
)

func TestDataDir(t *testing.T) {
	dir := DataDir()
	if runtime.GOOS == "windows" {
		if dir != "." {
			t.Fatalf("windows DataDir = %q, want %q", dir, ".")
		}
	} else {
		want := filepath.Join(mustHome(t), ".novelclaw")
		if dir != want {
			t.Fatalf("DataDir = %q, want %q", dir, want)
		}
	}
}

func mustHome(t *testing.T) string {
	t.Helper()
	home, err := homeForTest()
	if err != nil {
		t.Skipf("no home dir: %v", err)
	}
	return home
}

func TestResolve(t *testing.T) {
	abs := filepath.Join(t.TempDir(), "x.db")
	if got := Resolve(abs); got != filepath.Clean(abs) {
		t.Fatalf("Resolve(abs) = %q", got)
	}
	rel := Resolve(filepath.Join("data", "novelclaw.db"))
	if want := DB(); rel != want {
		t.Fatalf("Resolve(data/novelclaw.db) = %q, want %q", rel, want)
	}
}

func TestTreeAccessors(t *testing.T) {
	root := DataDir()
	for name, got := range map[string]string{
		"DB":            DB(),
		"Data":          Data(),
		"Souls":         Souls(),
		"Skills":        Skills(),
		"SkillsDefault": SkillsDefault(),
		"SkillsEvolved": SkillsEvolved(),
	} {
		rel, err := filepath.Rel(root, got)
		if err != nil || rel == ".." || len(rel) > 1 && rel[:3] == "../" {
			t.Fatalf("%s = %q escapes DataDir %q", name, got, root)
		}
	}
}

func TestEnsure(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	got, err := Ensure()
	if err != nil {
		t.Fatalf("Ensure: %v", err)
	}
	if runtime.GOOS != "windows" && got == "." {
		t.Fatalf("Ensure on unix must not return '.' when HOME set")
	}
	for _, want := range []string{Data(), Souls(), SkillsDefault(), SkillsEvolved()} {
		if !isDir(want) {
			t.Fatalf("Ensure did not create %q", want)
		}
	}
}
