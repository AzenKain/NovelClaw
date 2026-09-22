package llm

import (
	"strings"
	"unicode/utf8"
)

// Chunk represents a discrete text segment for translation.
type Chunk struct {
	Index      int    `json:"index"`
	Content    string `json:"content"`
	TokenCount int    `json:"token_count"`
}

// Chunker segments raw text into quote-aware chunks.
type Chunker struct {
	MaxRunesPerChunk int
}

// NewChunker creates a new Chunker.
func NewChunker(maxRunesPerChunk int) *Chunker {
	if maxRunesPerChunk <= 0 {
		maxRunesPerChunk = 2500
	}
	return &Chunker{MaxRunesPerChunk: maxRunesPerChunk}
}

// SplitIntoParagraphs splits raw text or HTML by newlines into non-empty paragraphs.
func (c *Chunker) SplitIntoParagraphs(text string) []string {
	normalized := strings.ReplaceAll(text, "\r\n", "\n")
	rawParas := strings.Split(normalized, "\n")

	var paragraphs []string
	for _, p := range rawParas {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			paragraphs = append(paragraphs, trimmed)
		}
	}
	return paragraphs
}

// hasUnclosedQuotes detects unclosed dialogue quotes in multiple languages.
func hasUnclosedQuotes(text string) bool {
	quotePairs := []struct{ open, close rune }{
		{'"', '"'},
		{'“', '”'},
		{'「', '」'},
		{'『', '』'},
		{'《', '》'},
		{'（', '）'},
		{'(', ')'},
	}

	for _, qp := range quotePairs {
		openCount := 0
		if qp.open == qp.close {
			for _, r := range text {
				if r == qp.open {
					openCount++
				}
			}
			if openCount%2 != 0 {
				return true
			}
		} else {
			for _, r := range text {
				if r == qp.open {
					openCount++
				} else if r == qp.close {
					openCount--
				}
			}
			if openCount > 0 {
				return true
			}
		}
	}
	return false
}

// ChunkText splits text into chunks without breaking open quotes.
func (c *Chunker) ChunkText(text string) []Chunk {
	paragraphs := c.SplitIntoParagraphs(text)
	if len(paragraphs) == 0 {
		return nil
	}

	var chunks []Chunk
	var currentBuffer strings.Builder
	currentRunes := 0

	for i, p := range paragraphs {
		pRunes := utf8.RuneCountInString(p)

		forceSplit := currentRunes >= c.MaxRunesPerChunk*2
		if currentRunes > 0 && ((currentRunes+pRunes > c.MaxRunesPerChunk && !hasUnclosedQuotes(currentBuffer.String())) || forceSplit) {
			bufferedText := currentBuffer.String()
			chunks = append(chunks, Chunk{
				Index:      len(chunks) + 1,
				Content:    strings.TrimSpace(bufferedText),
				TokenCount: currentRunes / 2,
			})
			currentBuffer.Reset()
			currentRunes = 0
		}

		if currentBuffer.Len() > 0 {
			currentBuffer.WriteString("\n\n")
		}
		currentBuffer.WriteString(p)
		currentRunes += pRunes

		if i == len(paragraphs)-1 && currentBuffer.Len() > 0 {
			chunks = append(chunks, Chunk{
				Index:      len(chunks) + 1,
				Content:    strings.TrimSpace(currentBuffer.String()),
				TokenCount: currentRunes / 2,
			})
		}
	}

	return chunks
}
