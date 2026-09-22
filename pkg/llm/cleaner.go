package llm

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	xmlDeclarationRegex = regexp.MustCompile(`(?i)<\?xml\b[^>]*\?>`)
	docTypeRegex        = regexp.MustCompile(`(?i)<!DOCTYPE\b[^>]*>`)
	headTagRegex        = regexp.MustCompile(`(?is)<head\b[^>]*>.*?<\/head>`)
	styleTagRegex       = regexp.MustCompile(`(?is)<style\b[^>]*>.*?<\/style>`)
	scriptTagRegex      = regexp.MustCompile(`(?is)<script\b[^>]*>.*?<\/script>`)
	bodyWrapperRegex    = regexp.MustCompile(`(?is)<body\b[^>]*>(.*?)<\/body>`)
	htmlWrapperRegex    = regexp.MustCompile(`(?is)<\/?(?:html|body)\b[^>]*>`)
)

// CleanRawContentForTranslation strips XML boilerplate, head, and scripts before chunking.
func CleanRawContentForTranslation(raw string) string {
	if !strings.Contains(raw, "<") {
		return raw
	}
	cleaned := xmlDeclarationRegex.ReplaceAllString(raw, "")
	cleaned = docTypeRegex.ReplaceAllString(cleaned, "")
	cleaned = headTagRegex.ReplaceAllString(cleaned, "")
	cleaned = styleTagRegex.ReplaceAllString(cleaned, "")
	cleaned = scriptTagRegex.ReplaceAllString(cleaned, "")

	if matches := bodyWrapperRegex.FindStringSubmatch(cleaned); len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	cleaned = htmlWrapperRegex.ReplaceAllString(cleaned, "")
	return strings.TrimSpace(cleaned)
}

// ForeignScriptFlag records an unexpected or untranslated foreign character sequence.
type ForeignScriptFlag struct {
	Script string `json:"script"`
	Match  string `json:"match"`
	Pos    int    `json:"pos"`
}

var (
	kanaRegex   = regexp.MustCompile(`[\p{Hiragana}\p{Katakana}ー]{1,}`)
	hanziRegex  = regexp.MustCompile(`[\p{Han}]{2,}`)
	hangulRegex = regexp.MustCompile(`[\p{Hangul}]{1,}`)

	// Semantic noise and author notes regexes
	authorNoteRegex = regexp.MustCompile(`(?im)^\s*(?:【?(?:PS|P\.S|Lời tác giả|Tác giả có lời|Author'?s? Note|T\/N|ND|Lời người dịch|Chương trước|Chương tiếp|Translator Note)[:：】\)].*$)`)
	watermarkRegex  = regexp.MustCompile(`(?im)^\s*(?:Truyện được (?:đăng|dịch|sưu tầm) tại|Nguồn:|Source:|Convert by|Trans by|Website:|Đọc truyện tại).*$`)
)

// DetectForeignResiduals detects unexpected source script residuals in target translation.
func DetectForeignResiduals(targetLang, text string) []ForeignScriptFlag {
	normTarget := strings.ToLower(strings.TrimSpace(targetLang))
	var flags []ForeignScriptFlag

	// When target is Vietnamese or English, Japanese Kana is never expected
	if normTarget == "vi" || normTarget == "en" {
		matches := kanaRegex.FindAllStringIndex(text, -1)
		for _, m := range matches {
			flags = append(flags, ForeignScriptFlag{
				Script: "japanese_kana",
				Match:  text[m[0]:m[1]],
				Pos:    m[0],
			})
		}

		hangulMatches := hangulRegex.FindAllStringIndex(text, -1)
		for _, m := range hangulMatches {
			flags = append(flags, ForeignScriptFlag{
				Script: "korean_hangul",
				Match:  text[m[0]:m[1]],
				Pos:    m[0],
			})
		}
	}

	// When target is Vietnamese or English, long Chinese Hanzi sequences are untranslated residuals
	if normTarget == "vi" || normTarget == "en" {
		matches := hanziRegex.FindAllStringIndex(text, -1)
		for _, m := range matches {
			flags = append(flags, ForeignScriptFlag{
				Script: "chinese_hanzi",
				Match:  text[m[0]:m[1]],
				Pos:    m[0],
			})
		}
	}

	return flags
}

// SanitizeForeignResiduals removes or marks residual foreign characters if needed.
func SanitizeForeignResiduals(targetLang, text string) (string, []ForeignScriptFlag) {
	flags := DetectForeignResiduals(targetLang, text)
	if len(flags) == 0 {
		return text, nil
	}

	normTarget := strings.ToLower(strings.TrimSpace(targetLang))
	cleaned := text
	if normTarget == "vi" || normTarget == "en" {
		// Clean out untranslated kana characters that leak through
		cleaned = kanaRegex.ReplaceAllString(cleaned, "")
	}

	return strings.TrimSpace(cleaned), flags
}

// StripSemanticNoise strips unwanted watermarks, author notes, and conversion artifacts.
func StripSemanticNoise(text string) (string, []string) {
	var removed []string

	authorMatches := authorNoteRegex.FindAllString(text, -1)
	for _, m := range authorMatches {
		trimmed := strings.TrimSpace(m)
		if trimmed != "" {
			removed = append(removed, trimmed)
		}
	}

	watermarkMatches := watermarkRegex.FindAllString(text, -1)
	for _, m := range watermarkMatches {
		trimmed := strings.TrimSpace(m)
		if trimmed != "" {
			removed = append(removed, trimmed)
		}
	}

	cleaned := authorNoteRegex.ReplaceAllString(text, "")
	cleaned = watermarkRegex.ReplaceAllString(cleaned, "")

	// Collapse 3+ consecutive newlines into 2
	multiNewline := regexp.MustCompile(`\n{3,}`)
	cleaned = multiNewline.ReplaceAllString(cleaned, "\n\n")

	return strings.TrimSpace(cleaned), removed
}

var (
	pTagRegex = regexp.MustCompile(`(?is)<p\b[^>]*>.*?<\/p>`)
)

func extractDefaultPClass(raw string) string {
	m := regexp.MustCompile(`(?i)<p\s+class=["']([^"']+)["']`).FindStringSubmatch(raw)
	if len(m) > 1 {
		return m[1]
	}
	return "calibre2"
}

// NormalizeHtmlParagraphs ensures all orphan text outside <p> tags is cleanly wrapped
// in matching <p> tags, preserving the original novel paragraph structure and preventing whitespace collapse.
func NormalizeHtmlParagraphs(rawSource, translatedText string) string {
	if !strings.Contains(rawSource, "<p") && !strings.Contains(rawSource, "<P") {
		return translatedText
	}

	pClass := extractDefaultPClass(rawSource)
	text := strings.TrimSpace(translatedText)
	if text == "" {
		return ""
	}

	// If translatedText contains NO <p tags at all, wrap every non-empty line
	if !strings.Contains(text, "<p") && !strings.Contains(text, "<P") {
		paras := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
		var result []string
		for _, p := range paras {
			trimmed := strings.TrimSpace(p)
			if trimmed == "" {
				continue
			}
			if strings.HasPrefix(trimmed, "<") && strings.HasSuffix(trimmed, ">") {
				result = append(result, trimmed)
			} else {
				result = append(result, fmt.Sprintf(`<p class="%s">%s</p>`, pClass, trimmed))
			}
		}
		return strings.Join(result, "\n\n")
	}

	// If translatedText has mixed content: some <p> tags and some un-tagged lines outside <p> tags
	var sb strings.Builder
	lastIdx := 0
	matches := pTagRegex.FindAllStringIndex(text, -1)
	for _, loc := range matches {
		start, end := loc[0], loc[1]
		if start > lastIdx {
			orphan := text[lastIdx:start]
			sb.WriteString(wrapOrphanLines(orphan, pClass))
		}
		sb.WriteString(text[start:end])
		lastIdx = end
	}
	if lastIdx < len(text) {
		orphan := text[lastIdx:]
		sb.WriteString(wrapOrphanLines(orphan, pClass))
	}

	return strings.TrimSpace(sb.String())
}

func wrapOrphanLines(orphan, pClass string) string {
	lines := strings.Split(strings.ReplaceAll(orphan, "\r\n", "\n"), "\n")
	var result strings.Builder
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		// Skip block level tags
		if strings.HasPrefix(trimmed, "<div") || strings.HasPrefix(trimmed, "</div>") ||
			strings.HasPrefix(trimmed, "<body") || strings.HasPrefix(trimmed, "</body>") ||
			strings.HasPrefix(trimmed, "<html") || strings.HasPrefix(trimmed, "</html>") ||
			strings.HasPrefix(trimmed, "<svg") || strings.HasPrefix(trimmed, "</svg>") ||
			strings.HasPrefix(trimmed, "<image") || strings.HasPrefix(trimmed, "</image>") ||
			strings.HasPrefix(trimmed, "<img") || strings.HasPrefix(trimmed, "</img>") {
			result.WriteString(trimmed)
			result.WriteString("\n")
			continue
		}
		result.WriteString(fmt.Sprintf("\n<p class=\"%s\">%s</p>\n", pClass, trimmed))
	}
	return result.String()
}

var (
	htmlContentTagRegex      = regexp.MustCompile(`(?s)<[^>]+>`)
	xmlCommentTagRegex       = regexp.MustCompile(`(?s)<!--.*?-->`)
	hasTranslatableTextRegex = regexp.MustCompile(`[\p{Hiragana}\p{Katakana}\p{Han}\p{Hangul}]|\b[a-zA-ZÀ-ỹ0-9]{3,}\b`)
	markdownCodeFenceRegex   = regexp.MustCompile("(?s)```(?:html|xml|xhtml)?\\s*([\\s\\S]*?)\\s*```")
	chatPreambleLineRegex    = regexp.MustCompile(`(?im)^\s*(?:dưới đây là (?:bản dịch|phần dịch|nội dung)|sau đây là (?:bản dịch|phần dịch)|tôi xin gửi|bản dịch (?:tiếng việt)?[:：]|here is the translation|certainly!|sure,|the provided text|as an ai).*$`)
	chatPostambleLineRegex   = regexp.MustCompile(`(?im)^\s*(?:nếu bạn có (?:bất kỳ|thắc mắc)|chúc bạn đọc|hy vọng bản dịch|nếu cần (?:chỉnh sửa|thay đổi)|vui lòng cung cấp).*$`)
)

// IsPureMarkupChunk returns true if the chunk content contains only HTML/SVG/XML markup,
// comments, image references, or whitespace with no natural language text to translate.
func IsPureMarkupChunk(chunk string) bool {
	stripped := xmlCommentTagRegex.ReplaceAllString(chunk, "")
	stripped = htmlContentTagRegex.ReplaceAllString(stripped, "")
	stripped = strings.TrimSpace(stripped)
	if stripped == "" {
		return true
	}
	return !hasTranslatableTextRegex.MatchString(stripped)
}

// StripChatbotMetaTalk removes markdown code block fences and conversational chatbot preambles/postambles.
func StripChatbotMetaTalk(text string) string {
	cleaned := strings.TrimSpace(text)
	if cleaned == "" {
		return ""
	}

	// 1. Unwrap markdown code fence if present
	if matches := markdownCodeFenceRegex.FindStringSubmatch(cleaned); len(matches) > 1 {
		inner := strings.TrimSpace(matches[1])
		if strings.Contains(inner, "<") && strings.Contains(inner, ">") {
			cleaned = inner
		} else if strings.Count(cleaned, "```") >= 2 {
			cleaned = inner
		}
	}

	// 2. Strip line-by-line conversational chatter
	lines := strings.Split(strings.ReplaceAll(cleaned, "\r\n", "\n"), "\n")
	var keptLines []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			keptLines = append(keptLines, "")
			continue
		}
		if chatPreambleLineRegex.MatchString(trimmed) {
			continue
		}
		if chatPostambleLineRegex.MatchString(trimmed) {
			continue
		}
		keptLines = append(keptLines, line)
	}

	cleaned = strings.TrimSpace(strings.Join(keptLines, "\n"))
	multiNewline := regexp.MustCompile(`\n{3,}`)
	cleaned = multiNewline.ReplaceAllString(cleaned, "\n\n")

	return cleaned
}
