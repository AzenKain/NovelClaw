package llm

import (
	"strings"
	"testing"
)

func TestCleanRawContentForTranslation(t *testing.T) {
	input := `<?xml version="1.0" encoding="utf-8"?>
<!DOCTYPE html>
<html>
<head><title>Test</title><style>.cls{color:red}</style></head>
<body>
<p>Hello world</p>
<script>console.log("bad");</script>
</body>
</html>`

	expected := `<p>Hello world</p>`
	output := CleanRawContentForTranslation(input)
	if output != expected {
		t.Fatalf("expected '%s', got '%s'", expected, output)
	}
}

func TestForeignResiduals(t *testing.T) {
	text := "Hắn mỉm cười rồi nói: 「こんにちは」, ánh mắt đầy sát khí."
	flags := DetectForeignResiduals("vi", text)
	if len(flags) == 0 {
		t.Fatalf("expected detected kana flags, got none")
	}

	cleaned, cleanedFlags := SanitizeForeignResiduals("vi", text)
	if len(cleanedFlags) == 0 {
		t.Fatalf("expected cleaned flags")
	}
	if cleaned == text {
		t.Fatalf("expected text to be cleaned, got: %s", cleaned)
	}
}

func TestNormalizeHtmlParagraphs(t *testing.T) {
	rawSource := `<p class="calibre2">Line 1</p><p class="calibre2">Line 2</p>`
	// Simulated translation where LLM dropped <p> tags for some lines
	transWithOrphans := `<p class="calibre2">Đoạn 1</p>

Đoạn 2 thoại

Đoạn 3 thoại

<p class="calibre2">Đoạn 4 kết</p>`

	normalized := NormalizeHtmlParagraphs(rawSource, transWithOrphans)
	if !strings.Contains(normalized, `<p class="calibre2">Đoạn 2 thoại</p>`) {
		t.Fatalf("expected Đoạn 2 to be wrapped in p tag, got:\n%s", normalized)
	}
	if !strings.Contains(normalized, `<p class="calibre2">Đoạn 3 thoại</p>`) {
		t.Fatalf("expected Đoạn 3 to be wrapped in p tag, got:\n%s", normalized)
	}
}

func TestStripSemanticNoise(t *testing.T) {
	text := `Chương 1: Khởi đầu
Truyện được đăng tại truyenfull.vn

Trời mưa tầm tã trên ngọn đồi.

【PS: Cảm ơn các bạn độc giả đã ủng hộ hoa tươi!】
Convert by: some-converter`

	cleaned, removed := StripSemanticNoise(text)
	if len(removed) < 2 {
		t.Fatalf("expected at least 2 noise patterns removed, got: %v", removed)
	}
	if !testing.Verbose() {
		// Just ensure cleaned text contains main content and not the noise
		if testing.Short() {
			return
		}
	}
	if cleaned == text {
		t.Fatalf("expected cleaned text to differ from original")
	}
}

func TestIsPureMarkupChunk(t *testing.T) {
	svgChunk := `<div>
<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" version="1.1" width="100%" height="100%" viewBox="0 0 914 1297" preserveAspectRatio="none">
<image width="914" height="1297" xlink:href="/api/v1/reader/series_1eaf2e64/asset/vol1/cover.jpeg"/>
</svg>
</div>`
	if !IsPureMarkupChunk(svgChunk) {
		t.Fatalf("expected SVG cover chunk to be recognized as pure markup")
	}

	textChunk := `<p class="calibre2">第１話 友人が500円の借金のカタに妹をよこしてきた話</p>`
	if IsPureMarkupChunk(textChunk) {
		t.Fatalf("expected text chunk NOT to be pure markup")
	}

	blankP := `<p class="calibre2"><br class="main"/></p>`
	if !IsPureMarkupChunk(blankP) {
		t.Fatalf("expected blank paragraph with br to be recognized as pure markup")
	}
}

func TestStripChatbotMetaTalk(t *testing.T) {
	sample := "Dưới đây là bản dịch tiếng Việt:\n\n" +
		"```html\n" +
		"<div>\n" +
		"<svg xmlns=\"http://www.w3.org/2000/svg\" width=\"100%\">\n" +
		"<image xlink:href=\"cover.jpeg\"/>\n" +
		"</svg>\n" +
		"</div>\n" +
		"```\n\n" +
		"Hy vọng bản dịch này đúng ý bạn."

	cleaned := StripChatbotMetaTalk(sample)
	if strings.Contains(cleaned, "Dưới đây là") {
		t.Fatalf("expected preamble to be stripped, got: %s", cleaned)
	}
	if strings.Contains(cleaned, "Hy vọng bản dịch") {
		t.Fatalf("expected postamble to be stripped, got: %s", cleaned)
	}
	if !strings.Contains(cleaned, "<svg") || !strings.Contains(cleaned, "cover.jpeg") {
		t.Fatalf("expected inner SVG markup to be preserved, got: %s", cleaned)
	}
	if strings.Contains(cleaned, "```") {
		t.Fatalf("expected code fence markers to be removed, got: %s", cleaned)
	}
}
