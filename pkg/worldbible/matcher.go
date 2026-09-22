package worldbible

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"novelclaw/internal/dtos"
	"novelclaw/pkg/storage"
	"novelclaw/pkg/trie"
)

// Matcher manages Trie-based keyword matching on World Bible entries to inject into LLM prompts.
type Matcher struct {
	store    *storage.Storage
	mu       sync.RWMutex
	cache    map[string]*trie.Matcher     // projectID -> trie
	entryMap map[string]map[string]dtos.WorldEntryDTO // projectID -> (name/alias -> entry)
}

// NewMatcher creates a new World Bible context matcher.
func NewMatcher(store *storage.Storage) *Matcher {
	return &Matcher{
		store:    store,
		cache:    make(map[string]*trie.Matcher),
		entryMap: make(map[string]map[string]dtos.WorldEntryDTO),
	}
}

// Invalidate clears the cache for a project when entries are added, modified or deleted.
func (m *Matcher) Invalidate(projectID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.cache, projectID)
	delete(m.entryMap, projectID)
}

// InvalidateCache is an alias for Invalidate.
func (m *Matcher) InvalidateCache(projectID string) {
	m.Invalidate(projectID)
}

// ensureMatcher loads entries and builds the Aho-Corasick automaton.
func (m *Matcher) ensureMatcher(ctx context.Context, projectID string) (*trie.Matcher, map[string]dtos.WorldEntryDTO, error) {
	m.mu.RLock()
	if t, ok := m.cache[projectID]; ok {
		em := m.entryMap[projectID]
		m.mu.RUnlock()
		return t, em, nil
	}
	m.mu.RUnlock()

	m.mu.Lock()
	defer m.mu.Unlock()

	// Double check
	if t, ok := m.cache[projectID]; ok {
		return t, m.entryMap[projectID], nil
	}

	rawEntries, err := m.store.ListWorldEntries(ctx, projectID)
	if err != nil {
		return nil, nil, fmt.Errorf("list world entries for matcher: %w", err)
	}

	cats, _ := m.store.ListWorldCategories(ctx, projectID)
	catNames := make(map[string]string)
	for _, c := range cats {
		catNames[c.ID] = c.Name
	}

	t := trie.NewMatcher()
	em := make(map[string]dtos.WorldEntryDTO)

	for _, e := range rawEntries {
		var aliases []string
		_ = json.Unmarshal([]byte(e.AliasesJson), &aliases)
		var attrs map[string]string
		_ = json.Unmarshal([]byte(e.AttributesJson), &attrs)

		entryDTO := dtos.WorldEntryDTO{
			ID:                 e.ID,
			ProjectID:          e.ProjectID,
			CategoryID:         e.CategoryID,
			CategoryName:       catNames[e.CategoryID],
			Name:               e.Name,
			Aliases:            aliases,
			Summary:            e.Summary,
			FullDescription:    e.FullDescription,
			Attributes:         attrs,
			DiscoveredBy:       e.DiscoveredBy,
			SourceChapterIndex: e.SourceChapterIndex,
			IsVerified:         e.IsVerified == 1,
		}

		cleanName := strings.TrimSpace(e.Name)
		if cleanName != "" {
			lowerName := strings.ToLower(cleanName)
			t.AddPattern(lowerName, e.ID)
			em[lowerName] = entryDTO
		}

		for _, alias := range aliases {
			cleanAlias := strings.TrimSpace(alias)
			if cleanAlias != "" {
				lowerAlias := strings.ToLower(cleanAlias)
				t.AddPattern(lowerAlias, e.ID)
				em[lowerAlias] = entryDTO
			}
		}
	}

	t.Build()
	m.cache[projectID] = t
	m.entryMap[projectID] = em
	return t, em, nil
}

// MatchContext finds all world entries referenced in the given text segment (case-insensitive).
func (m *Matcher) MatchContext(ctx context.Context, projectID, text string) []dtos.WorldEntryDTO {
	if m.store == nil || projectID == "" || strings.TrimSpace(text) == "" {
		return nil
	}

	t, em, err := m.ensureMatcher(ctx, projectID)
	if err != nil || t == nil {
		return nil
	}

	// Match case-insensitively using lowercase text
	matches := t.FindAll(strings.ToLower(text))
	if len(matches) == 0 {
		return nil
	}

	seen := make(map[string]bool)
	var matchedEntries []dtos.WorldEntryDTO

	for _, match := range matches {
		entry, ok := em[match.Pattern]
		if ok && !seen[entry.ID] {
			seen[entry.ID] = true
			matchedEntries = append(matchedEntries, entry)
		}
	}
	return matchedEntries
}

// FormatContextPrompt formats matched entries into a compact prompt injection block.
func (m *Matcher) FormatContextPrompt(entries []dtos.WorldEntryDTO) string {
	if len(entries) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("[WORLD BIBLE & LOREBOOK CONTEXT]:\n")
	for _, e := range entries {
		catLabel := ""
		if e.CategoryName != "" {
			catLabel = fmt.Sprintf("[%s] ", e.CategoryName)
		}
		sb.WriteString(fmt.Sprintf("- %s**%s**: %s", catLabel, e.Name, e.Summary))

		if len(e.Attributes) > 0 {
			var attrPairs []string
			for k, v := range e.Attributes {
				attrPairs = append(attrPairs, fmt.Sprintf("%s: %s", k, v))
			}
			sb.WriteString(fmt.Sprintf(" (%s)", strings.Join(attrPairs, ", ")))
		}
		sb.WriteString("\n")
	}
	return strings.TrimSpace(sb.String())
}
