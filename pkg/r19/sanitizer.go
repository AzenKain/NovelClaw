package r19

import (
	"sort"
	"strings"

	"novelclaw/pkg/trie"
)

// Sanitizer removes or camouflages sensitive terms in historical context to prevent safety filter leakage.
type Sanitizer struct {
	masker *Masker
}

// NewSanitizer creates an initialized Sanitizer backed by a Masker.
func NewSanitizer(masker *Masker) *Sanitizer {
	return &Sanitizer{masker: masker}
}

// SanitizeContext replaces sensitive terms in historical sliding windows with safe euphemisms.
func (s *Sanitizer) SanitizeContext(text string) string {
	if text == "" || s.masker == nil {
		return text
	}

	s.masker.mu.RLock()
	defer s.masker.mu.RUnlock()

	matches := s.masker.matcher.FindAll(text)
	if len(matches) == 0 {
		return text
	}

	sort.Slice(matches, func(i, j int) bool {
		if matches[i].Start != matches[j].Start {
			return matches[i].Start < matches[j].Start
		}
		return matches[i].End > matches[j].End
	})

	var nonOverlapping []trie.MatchResult
	lastEnd := 0
	for _, match := range matches {
		if match.Start >= lastEnd {
			nonOverlapping = append(nonOverlapping, match)
			lastEnd = match.End
		}
	}

	var sb strings.Builder
	cursor := 0
	for _, match := range nonOverlapping {
		sb.WriteString(text[cursor:match.Start])
		sb.WriteString("[...]")
		cursor = match.End
	}
	sb.WriteString(text[cursor:])

	return sb.String()
}

// StripTerms completely removes sensitive terms from text.
func (s *Sanitizer) StripTerms(text string) string {
	if text == "" || s.masker == nil {
		return text
	}

	s.masker.mu.RLock()
	defer s.masker.mu.RUnlock()

	matches := s.masker.matcher.FindAll(text)
	if len(matches) == 0 {
		return text
	}

	sort.Slice(matches, func(i, j int) bool {
		if matches[i].Start != matches[j].Start {
			return matches[i].Start < matches[j].Start
		}
		return matches[i].End > matches[j].End
	})

	var nonOverlapping []trie.MatchResult
	lastEnd := 0
	for _, match := range matches {
		if match.Start >= lastEnd {
			nonOverlapping = append(nonOverlapping, match)
			lastEnd = match.End
		}
	}

	var sb strings.Builder
	cursor := 0
	for _, match := range nonOverlapping {
		sb.WriteString(text[cursor:match.Start])
		cursor = match.End
	}
	sb.WriteString(text[cursor:])

	return sb.String()
}
