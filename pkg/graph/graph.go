package graph

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"

	"novelclaw/internal/gen/sqlc"
	"novelclaw/pkg/storage"
	"novelclaw/pkg/trie"
)

// SelectiveContext contains filtered relations and terms present in the current text chunk.
type SelectiveContext struct {
	ActiveEntities  []storage.Entity         `json:"active_entities"`
	ActiveRelations []sqlc.CharacterRelation `json:"active_relations"`
	ActiveGlossary  []sqlc.Glossary          `json:"active_glossary"`
	FormattedPrompt string                   `json:"formatted_prompt"`
}

// ProgressiveGraph manages temporal entity graph and selective context injection.
type ProgressiveGraph struct {
	store *storage.Storage
	mu    sync.RWMutex
}

// NewProgressiveGraph creates a new ProgressiveGraph manager.
func NewProgressiveGraph(store *storage.Storage) *ProgressiveGraph {
	return &ProgressiveGraph{
		store: store,
	}
}

// BuildSelectiveContext scans chunk text and injects only relevant characters, relations, and glossary terms.
func (g *ProgressiveGraph) BuildSelectiveContext(ctx context.Context, projectID string, chapterIndex int64, chunkText string) (*SelectiveContext, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	entities, err := g.store.ListActiveEntitiesAtChapter(ctx, projectID, chapterIndex)
	if err != nil {
		return nil, fmt.Errorf("list active entities: %w", err)
	}

	glossaryTerms, err := g.store.ListGlossaryByProject(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("list glossary: %w", err)
	}

	matcher := trie.NewMatcher()
	entityMap := make(map[string]storage.Entity)
	for _, ent := range entities {
		entityMap[ent.Name] = ent
		matcher.AddPattern(ent.Name, ent.Name)
		for _, alias := range ent.Aliases {
			matcher.AddPattern(alias, ent.Name)
		}
	}

	glossaryMap := make(map[string]sqlc.Glossary)
	for _, gt := range glossaryTerms {
		glossaryMap[gt.SourceTerm] = gt
		matcher.AddPattern(gt.SourceTerm, "glossary:"+gt.SourceTerm)
	}

	matcher.Build()
	matches := matcher.FindAll(chunkText)

	matchedEntityNames := make(map[string]bool)
	matchedGlossaryTerms := make(map[string]bool)

	for _, m := range matches {
		val, ok := m.Payload.(string)
		if !ok {
			continue
		}
		if strings.HasPrefix(val, "glossary:") {
			term := strings.TrimPrefix(val, "glossary:")
			matchedGlossaryTerms[term] = true
		} else {
			matchedEntityNames[val] = true
		}
	}

	var activeEntities []storage.Entity
	var activeEntityList []string
	for name := range matchedEntityNames {
		if ent, ok := entityMap[name]; ok {
			activeEntities = append(activeEntities, ent)
			activeEntityList = append(activeEntityList, name)
		}
	}
	sort.Slice(activeEntities, func(i, j int) bool {
		return activeEntities[i].Name < activeEntities[j].Name
	})

	var activeRelations []sqlc.CharacterRelation
	allRels, err := g.store.ListRelationsByProject(ctx, projectID)
	if err == nil && len(allRels) > 0 {
		effectiveRels := make(map[[2]string]sqlc.CharacterRelation)
		for _, r := range allRels {
			if r.SinceChapter <= chapterIndex {
				pair := [2]string{r.FromChar, r.ToChar}
				if prev, ok := effectiveRels[pair]; !ok || r.SinceChapter > prev.SinceChapter {
					effectiveRels[pair] = r
				}
			}
		}

		if len(activeEntityList) >= 2 {
			for i := 0; i < len(activeEntityList); i++ {
				for j := 0; j < len(activeEntityList); j++ {
					if i == j {
						continue
					}
					pair := [2]string{activeEntityList[i], activeEntityList[j]}
					if rel, ok := effectiveRels[pair]; ok && rel.CallAs != "" {
						activeRelations = append(activeRelations, rel)
					}
				}
			}
		} else if len(activeEntityList) == 1 {
			// When only 1 entity is mentioned by name (e.g. speaking to the 1st person POV narrator),
			// inject active relations involving this character so pronouns & honorifics stay consistent.
			charName := activeEntityList[0]
			for pair, rel := range effectiveRels {
				if (pair[0] == charName || pair[1] == charName) && rel.CallAs != "" {
					activeRelations = append(activeRelations, rel)
				}
			}
		}
	}
	sort.Slice(activeRelations, func(i, j int) bool {
		if activeRelations[i].FromChar != activeRelations[j].FromChar {
			return activeRelations[i].FromChar < activeRelations[j].FromChar
		}
		return activeRelations[i].ToChar < activeRelations[j].ToChar
	})

	var activeGlossary []sqlc.Glossary
	for term := range matchedGlossaryTerms {
		if gItem, ok := glossaryMap[term]; ok {
			activeGlossary = append(activeGlossary, gItem)
		}
	}
	sort.Slice(activeGlossary, func(i, j int) bool {
		return activeGlossary[i].SourceTerm < activeGlossary[j].SourceTerm
	})

	formatted := formatPromptBlock(activeEntities, activeRelations, activeGlossary)

	return &SelectiveContext{
		ActiveEntities:  activeEntities,
		ActiveRelations: activeRelations,
		ActiveGlossary:  activeGlossary,
		FormattedPrompt: formatted,
	}, nil
}

// formatPromptBlock creates a compact markdown block for the LLM prompt.
func formatPromptBlock(entities []storage.Entity, relations []sqlc.CharacterRelation, glossary []sqlc.Glossary) string {
	if len(entities) == 0 && len(relations) == 0 && len(glossary) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("\n[SELECTIVE CONTEXT INJECTION (L2/L3)]\n")

	if len(entities) > 0 {
		sb.WriteString("- Active Characters in Scene:\n")
		for _, e := range entities {
			roleStr := ""
			if e.Role != "" {
				roleStr = fmt.Sprintf(" (%s)", e.Role)
			}
			metaDetails := ""
			if desc, ok := e.Metadata["description"].(string); ok && desc != "" {
				metaDetails += fmt.Sprintf(" - %s", desc)
			}
			var factsList []string
			if fSlice, ok := e.Metadata["facts"].([]any); ok {
				for _, item := range fSlice {
					if s, ok := item.(string); ok && s != "" {
						factsList = append(factsList, s)
					}
				}
			} else if fSliceStr, ok := e.Metadata["facts"].([]string); ok {
				for _, s := range fSliceStr {
					if s != "" {
						factsList = append(factsList, s)
					}
				}
			}
			if len(factsList) > 0 {
				metaDetails += fmt.Sprintf(" [Facts: %s]", strings.Join(factsList, "; "))
			}
			sb.WriteString(fmt.Sprintf("  * %s%s%s\n", e.Name, roleStr, metaDetails))
		}
	}

	if len(relations) > 0 {
		sb.WriteString("- Verified Address Terms for Current Scene:\n")
		for _, r := range relations {
			toneStr := ""
			if r.Tone != "" {
				toneStr = fmt.Sprintf(" [%s]", r.Tone)
			}
			sb.WriteString(fmt.Sprintf("  * %s calls %s: '%s' | Self: '%s'%s\n", r.FromChar, r.ToChar, r.CallAs, r.SelfCallAs, toneStr))
		}
	}

	if len(glossary) > 0 {
		sb.WriteString("- Scene Terminology:\n")
		for _, g := range glossary {
			sb.WriteString(fmt.Sprintf("  * %s -> %s (%s)\n", g.SourceTerm, g.TargetTerm, g.Category))
		}
	}
	sb.WriteString("[END SELECTIVE CONTEXT]\n")

	return sb.String()
}
