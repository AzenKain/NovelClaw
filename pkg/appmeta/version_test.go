package appmeta

import "testing"

func TestVersionIsReleaseShaped(t *testing.T) {
	if Version == "" || Version == "dev" {
		t.Fatalf("Version must be a real release version, got %q", Version)
	}
	if got := CurrentVersion(); got != Version {
		t.Fatalf("CurrentVersion() = %q, want %q", got, Version)
	}
}

func TestPlatformInfo(t *testing.T) {
	if got := PlatformInfo(); got == "" {
		t.Fatal("PlatformInfo must not be empty")
	}
}

func TestVersionReport(t *testing.T) {
	r := VersionReport()
	if r.Version != Version {
		t.Fatalf("report version = %q, want %q", r.Version, Version)
	}
	if r.DataDir == "" || r.DBPath == "" || r.SoulsDir == "" || r.SkillsDir == "" {
		t.Fatalf("report paths incomplete: %+v", r)
	}
	if r.ChecksumAsset == "" {
		t.Fatal("checksum asset must not be empty")
	}
	if r.UpdateRepo == "" {
		t.Fatal("update repo must not be empty")
	}
}
