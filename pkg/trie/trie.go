package trie

import (
	"strings"
	"sync"
	"unicode/utf8"
)

// MatchResult represents a single pattern match in text.
type MatchResult struct {
	Pattern string
	Payload any
	Start   int
	End     int
}

// Node represents a node in the Aho-Corasick automaton.
type Node struct {
	children map[rune]*Node
	fail     *Node
	output   []*outputEntry
}

type outputEntry struct {
	pattern string
	payload any
}

// Matcher provides multi-pattern string matching using Aho-Corasick.
type Matcher struct {
	mu    sync.RWMutex
	root  *Node
	built bool
}

// NewMatcher creates an uncompiled Matcher.
func NewMatcher() *Matcher {
	return &Matcher{
		root: &Node{
			children: make(map[rune]*Node),
		},
	}
}

// AddPattern adds a pattern and its associated payload to the Matcher.
func (m *Matcher) AddPattern(pattern string, payload any) {
	m.mu.Lock()
	defer m.mu.Unlock()

	clean := strings.TrimSpace(pattern)
	if clean == "" {
		return
	}

	m.built = false
	curr := m.root
	for _, r := range clean {
		next, ok := curr.children[r]
		if !ok {
			next = &Node{children: make(map[rune]*Node)}
			curr.children[r] = next
		}
		curr = next
	}

	curr.output = append(curr.output, &outputEntry{
		pattern: clean,
		payload: payload,
	})
}

// Build compiles failure transitions for linear-time Aho-Corasick searching.
func (m *Matcher) Build() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.built {
		return
	}

	var queue []*Node
	for _, child := range m.root.children {
		child.fail = m.root
		queue = append(queue, child)
	}

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]

		for r, child := range curr.children {
			f := curr.fail
			for f != nil && f.children[r] == nil {
				f = f.fail
			}

			if f != nil {
				child.fail = f.children[r]
			} else {
				child.fail = m.root
			}

			if child.fail != nil && len(child.fail.output) > 0 {
				child.output = append(child.output, child.fail.output...)
			}

			queue = append(queue, child)
		}
	}

	m.built = true
}

// FindAll scans text and returns all unique matches found.
func (m *Matcher) FindAll(text string) []MatchResult {
	m.mu.RLock()
	if !m.built {
		m.mu.RUnlock()
		m.Build()
		m.mu.RLock()
	}
	defer m.mu.RUnlock()

	var results []MatchResult
	seen := make(map[string]bool)

	curr := m.root
	byteOffset := 0

	for i, r := range text {
		for curr != nil && curr.children[r] == nil {
			curr = curr.fail
		}

		if curr == nil {
			curr = m.root
			continue
		}

		curr = curr.children[r]
		if len(curr.output) > 0 {
			byteOffset = i + utf8.RuneLen(r)
			for _, out := range curr.output {
				if !seen[out.pattern] {
					seen[out.pattern] = true
					results = append(results, MatchResult{
						Pattern: out.pattern,
						Payload: out.payload,
						Start:   byteOffset - len(out.pattern),
						End:     byteOffset,
					})
				}
			}
		}
	}

	return results
}
