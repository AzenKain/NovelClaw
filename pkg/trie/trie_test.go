package trie

import (
	"strings"
	"testing"
)

func TestMatcher_FindAll(t *testing.T) {
	m := NewMatcher()
	m.AddPattern("羽島伊月", "char_itsuki")
	m.AddPattern("伊月", "char_itsuki_short")
	m.AddPattern("可児那由多", "char_nayuta")
	m.AddPattern("那由多", "char_nayuta_short")
	m.AddPattern("白木屋", "location_shirakiya")
	m.AddPattern("TRPG", "term_trpg")
	m.AddPattern("Bia hơi", "term_beer")

	text := "昨夜、羽島伊月と那由多は白木屋に行ってTRPGについて語りながらBia hơiを飲んだ。"

	matches := m.FindAll(text)
	if len(matches) < 5 {
		t.Fatalf("expected at least 5 matches, got %d", len(matches))
	}

	foundMap := make(map[string]bool)
	for _, match := range matches {
		foundMap[match.Pattern] = true
	}

	expected := []string{"羽島伊月", "那由多", "白木屋", "TRPG", "Bia hơi"}
	for _, exp := range expected {
		if !foundMap[exp] {
			t.Errorf("expected match for %q not found", exp)
		}
	}
}

func BenchmarkMatcher_FindAll(b *testing.B) {
	m := NewMatcher()
	for i := 0; i < 500; i++ {
		m.AddPattern(strings.Repeat("char", 2)+string(rune('A'+i%26)), i)
	}
	m.AddPattern("可児那由多", "target_1")
	m.AddPattern("羽島伊月", "target_2")
	m.Build()

	sampleText := "这是一个长文本测试。羽島伊月在房间里写作，可児那由多走了进来。" + strings.Repeat("一些填充文本，用于模拟小说的一个完整段落。", 20)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m.FindAll(sampleText)
	}
}
