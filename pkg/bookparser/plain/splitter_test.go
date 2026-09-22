package plain_test

import (
	"testing"

	"novelclaw/pkg/bookparser/plain"
)

func TestDetectChapterMarkers(t *testing.T) {
	sampleText := `Giới thiệu truyện: Đây là cuốn tiểu thuyết chuyển sinh kinh điển.

Chương 1: Khởi đầu mới
Satoru Mikami là một nhân viên công sở bình thường.

Chương 2: Bạo Phong Long
Tại một hang động sâu thẳm, cậu gặp một con rồng.
Nhân vật nói: 「Chương này ta phải đánh bại ngươi!」 nhưng đó chỉ là lời nói suông.

第3章 哥布林村庄
离开洞窟后，利姆鲁遇到了哥布林一族。

第4話 仲間たち
新しい仲間ができました。
`

	markers := plain.DetectChapterMarkers(sampleText)
	if len(markers) < 4 {
		t.Fatalf("expected at least 4 markers, got %d", len(markers))
	}

	// Verify the real chapters
	expectedTitles := []string{
		"Chương 1: Khởi đầu mới",
		"Chương 2: Bạo Phong Long",
		"第3章 哥布林村庄",
		"第4話 仲間たち",
	}

	foundCount := 0
	for _, m := range markers {
		for _, exp := range expectedTitles {
			if m.Title == exp {
				foundCount++
				if m.Confidence < 0.9 {
					t.Errorf("expected high confidence for %s, got %f", exp, m.Confidence)
				}
			}
		}
		// If a dialogue line is matched, it must be flagged as Suspicious
		if m.Title == "「Chương này ta phải đánh bại ngươi!」" {
			if !m.IsSuspicious {
				t.Errorf("dialogue should be marked as suspicious")
			}
		}
	}

	if foundCount != 4 {
		t.Errorf("expected to match all 4 main chapters, matched %d", foundCount)
	}

	// Test SplitByMarkers with the filtered marker list
	var approved []plain.CandidateChapterMarker
	for _, m := range markers {
		if !m.IsSuspicious {
			approved = append(approved, m)
		}
	}

	chapters := plain.SplitByMarkers(sampleText, approved)
	// Should contain 1 preface chapter + 4 main chapters = 5 chapters
	if len(chapters) != 5 {
		t.Fatalf("expected 5 chapters (1 preface + 4 main), got %d", len(chapters))
	}

	if chapters[0].Title != "Preface / Introduction" {
		t.Errorf("expected preface, got %s", chapters[0].Title)
	}
	if chapters[1].Title != "Chương 1: Khởi đầu mới" {
		t.Errorf("expected chapter 1, got %s", chapters[1].Title)
	}
}
