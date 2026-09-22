package llm

import "testing"

func TestNormalizeTranslationMode(t *testing.T) {
	cases := map[string]TranslationMode{
		"":                      ModeConcurrentDualAgent,
		"concurrent_dual_agent": ModeConcurrentDualAgent,
		"dual_agent":            ModeConcurrentDualAgent,
		"single_pass":           ModeSinglePass,
		"hierarchical_3pass":    ModeHierarchical3Pass,
		"swarm_arc":             ModeSwarmArcParallel, // legacy frontend value
		"swarm_arc_parallel":    ModeSwarmArcParallel,
		"  SWARM_ARC  ":         ModeSwarmArcParallel,
	}
	for in, want := range cases {
		if got := NormalizeTranslationMode(in); got != want {
			t.Errorf("NormalizeTranslationMode(%q) = %q, want %q", in, got, want)
		}
	}
}
