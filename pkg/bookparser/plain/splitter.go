package plain

import (
	"bufio"
	"fmt"
	"regexp"
	"strings"

	"novelclaw/pkg/bookparser"
)

// CandidateChapterMarker represents a potential chapter marker detected in raw text.
type CandidateChapterMarker struct {
	Index        int     `json:"index"`
	Title        string  `json:"title"`
	LineNumber   int     `json:"line_number"`
	ByteOffset   int64   `json:"byte_offset"`
	Confidence   float64 `json:"confidence"`
	Reason       string  `json:"reason"`
	IsSuspicious bool    `json:"is_suspicious"` // Suspicious (e.g. dialogue line, too short/too long)
}

var (
	// Regex matching Chinese chapter headings: 第1章, 第一百二十章, 第3回, 卷一...
	zhChapterRegex = regexp.MustCompile(`^\s*第[0-9零一二三四五六七八九十百千万]+[章回节卷部][\s:：\p{Zs}]*(.*)$`)

	// Regex matching Japanese chapter headings: 第1章, 第1話, その1, プロローグ, エピローグ...
	jaChapterRegex = regexp.MustCompile(`^\s*(?:第[0-9０-９一二三四五六七八九十百千万]+[章話節幕]|プロローグ|エピローグ|序章|終章)[\s:：\p{Zs}]*(.*)$`)

	// Regex matching English chapter headings: Chapter 1, CHAPTER I, Volume 2, Prologue, Epilogue...
	enChapterRegex = regexp.MustCompile(`(?i)^\s*(?:Chapter|Book|Volume|Part|Act|Section|Prologue|Epilogue)\s+([0-9IVXLCDM]+.*)$`)

	// Regex matching Vietnamese chapter headings: Chương 1, Hồi 12, Phần 3, Mở đầu, Kết thúc...
	viChapterRegex = regexp.MustCompile(`(?i)^\s*(?:Chương|Hồi|Phần|Tập|Mục|Tiết|Mở đầu|Ngoại truyện|Phiên ngoại)\s*([0-9IVXLCDM]*.*)$`)

	// Regex for a leading line number: 1. Title, 001 Title...
	numberedChapterRegex = regexp.MustCompile(`^\s*([0-9]{1,4})[\.\s、：:](.+)$`)
)

// DetectChapterMarkers quickly scans raw text and detects potential chapter markers.
func DetectChapterMarkers(rawContent string) []CandidateChapterMarker {
	scanner := bufio.NewScanner(strings.NewReader(rawContent))
	var markers []CandidateChapterMarker
	var byteOffset int64 = 0
	lineNum := 0

	for scanner.Scan() {
		line := scanner.Text()
		lineNum++
		lineLen := int64(len(line) + 1) // +1 for the newline character
		trimmed := strings.TrimSpace(line)

		if trimmed == "" {
			byteOffset += lineLen
			continue
		}

		// Skip overly long lines (> 80 chars); chapter headings are rarely paragraph-length.
		if len([]rune(trimmed)) > 80 {
			byteOffset += lineLen
			continue
		}

		// Skip lines that are quoted dialogue (「...」, “...”, "..."); these are usually not headings.
		isDialogue := (strings.HasPrefix(trimmed, "「") && strings.HasSuffix(trimmed, "」")) ||
			(strings.HasPrefix(trimmed, "“") && strings.HasSuffix(trimmed, "”")) ||
			(strings.HasPrefix(trimmed, "\"") && strings.HasSuffix(trimmed, "\""))

		var matchFound bool
		var title string
		var confidence float64
		var reason string

		switch {
		case zhChapterRegex.MatchString(trimmed):
			matchFound = true
			title = trimmed
			confidence = 0.95
			reason = "Matched Chinese heading pattern (第...章/回/节)"
		case jaChapterRegex.MatchString(trimmed):
			matchFound = true
			title = trimmed
			confidence = 0.95
			reason = "Matched Japanese heading pattern (第...話/章/プロローグ)"
		case viChapterRegex.MatchString(trimmed):
			matchFound = true
			title = trimmed
			confidence = 0.95
			reason = "Matched Vietnamese heading pattern (Chương/Hồi/Phần...)"
		case enChapterRegex.MatchString(trimmed):
			matchFound = true
			title = trimmed
			confidence = 0.95
			reason = "Matched English heading pattern (Chapter/Volume/Part...)"
		case numberedChapterRegex.MatchString(trimmed):
			matchFound = true
			title = trimmed
			confidence = 0.75
			reason = "Matched leading line-number pattern"
		}

		if matchFound {
			isSuspicious := false
			if isDialogue {
				isSuspicious = true
				confidence -= 0.4
				reason += " [Warning: this line may be character dialogue]"
			}

			markers = append(markers, CandidateChapterMarker{
				Index:        len(markers) + 1,
				Title:        title,
				LineNumber:   lineNum,
				ByteOffset:   byteOffset,
				Confidence:   confidence,
				Reason:       reason,
				IsSuspicious: isSuspicious,
			})
		}

		byteOffset += lineLen
	}

	return markers
}

// SplitByMarkers splits raw text into a list of ChapterData based on approved chapter markers.
func SplitByMarkers(rawContent string, approvedMarkers []CandidateChapterMarker) []bookparser.ChapterData {
	if len(approvedMarkers) == 0 {
		return []bookparser.ChapterData{
			{
				Title:       "Full Text",
				Content:     rawContent,
				ContentPath: "text/chapter_1.txt",
				Index:       0,
			},
		}
	}

	var chapters []bookparser.ChapterData

	// Leading section before chapter 1 (preface / introduction), if any.
	firstOffset := approvedMarkers[0].ByteOffset
	if firstOffset > 0 {
		prologueText := strings.TrimSpace(rawContent[:firstOffset])
		if len(prologueText) > 0 {
			chapters = append(chapters, bookparser.ChapterData{
				Title:       "Preface / Introduction",
				Content:     prologueText,
				ContentPath: "text/chapter_0.txt",
				Index:       0,
			})
		}
	}

	for i := 0; i < len(approvedMarkers); i++ {
		cur := approvedMarkers[i]
		start := cur.ByteOffset

		var end int64
		if i+1 < len(approvedMarkers) {
			end = approvedMarkers[i+1].ByteOffset
		} else {
			end = int64(len(rawContent))
		}

		if start < end && start < int64(len(rawContent)) {
			if end > int64(len(rawContent)) {
				end = int64(len(rawContent))
			}
			chapterBody := strings.TrimSpace(rawContent[start:end])
			chapters = append(chapters, bookparser.ChapterData{
				Title:       cur.Title,
				Content:     chapterBody,
				ContentPath: fmt.Sprintf("text/chapter_%d.txt", len(chapters)+1),
				Index:       len(chapters),
			})
		}
	}

	return chapters
}
