package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"

	"novelclaw/internal/gen/sqlc"
	"novelclaw/pkg/storage"
	"novelclaw/pkg/websearch"
)

// AskHumanCallback is invoked when an agent requests guidance or clarification from a human coworker.
type AskHumanCallback func(ctx context.Context, question string, options []string, contextSnippet string) (string, error)

// ToolRegistry manages and executes tools available to the LLM agent.
type ToolRegistry struct {
	store           *storage.Storage
	researchEng     *websearch.ResearchEngine
	critic          *ShadowCritic
	decisionMode    string // "manual" | "auto"
	styleGuide      string
	AskHumanHandler AskHumanCallback
}

// NewToolRegistry creates a new ToolRegistry.
func NewToolRegistry(store *storage.Storage, researchEng *websearch.ResearchEngine) *ToolRegistry {
	return &ToolRegistry{
		store:        store,
		researchEng:  researchEng,
		decisionMode: "manual",
	}
}

// SetDecisionMode configures whether dilemmas are resolved via human ('manual') or Shadow Critic ('auto').
func (tr *ToolRegistry) SetDecisionMode(mode string) {
	if mode != "" {
		tr.decisionMode = mode
	}
}

// SetShadowCritic attaches a shadow critic for auto dilemma arbitration.
func (tr *ToolRegistry) SetShadowCritic(critic *ShadowCritic) {
	tr.critic = critic
}

// SetStyleGuide sets the style guide to guide shadow critic decisions.
func (tr *ToolRegistry) SetStyleGuide(guide string) {
	tr.styleGuide = guide
}

// ArbitrateDilemma delegates arbitration to the attached Shadow Critic.
func (tr *ToolRegistry) ArbitrateDilemma(ctx context.Context, question string, options []string, contextSnippet string) (string, error) {
	if tr.critic != nil {
		return tr.critic.ArbitrateDilemma(ctx, question, options, contextSnippet, tr.styleGuide)
	}
	if len(options) > 0 {
		return options[0], nil
	}
	return "Default recommendation", nil
}

// SetAskHumanHandler assigns a human-in-the-loop callback for interactive clarifications.
func (tr *ToolRegistry) SetAskHumanHandler(handler AskHumanCallback) {
	tr.AskHumanHandler = handler
}

// GetAvailableTools returns the list of OpenAI-compatible tool schemas.
func (tr *ToolRegistry) GetAvailableTools() []ToolDefinition {
	return []ToolDefinition{
		{
			Type: "function",
			Function: FunctionSchema{
				Name:        "search_book_context",
				Description: "Search passages across all chapters of the current book using FTS5 Hybrid Search. Useful for checking character address terms, past events, or names.",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"query": map[string]any{
							"type":        "string",
							"description": "Keyword or phrase to search within book chapters",
						},
					},
					"required": []string{"query"},
				},
			},
		},
		{
			Type: "function",
			Function: FunctionSchema{
				Name:        "web_lookup",
				Description: "Look up encyclopedic knowledge, cultural terms (Japanese/Chinese), names, memes, or slang via Wikipedia/DuckDuckGo with clean reader extraction.",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"keyword": map[string]any{
							"type":        "string",
							"description": "Term, name, or concept to look up",
						},
						"language": map[string]any{
							"type":        "string",
							"description": "Preferred lookup language code ('ja', 'zh', 'en', 'vi'). Defaults to 'ja'",
						},
					},
					"required": []string{"keyword"},
				},
			},
		},
		{
			Type: "function",
			Function: FunctionSchema{
				Name:        "query_character_history",
				Description: "Query character relations, address terms, and timeline milestones at or before the current chapter from the L2 Progressive Graph.",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"character_name": map[string]any{
							"type":        "string",
							"description": "Name or alias of the character to look up",
						},
						"target_character": map[string]any{
							"type":        "string",
							"description": "Optional other character name to check specific relationship/address term with",
						},
					},
					"required": []string{"character_name"},
				},
			},
		},
		{
			Type: "function",
			Function: FunctionSchema{
				Name:        "lookup_character_relation",
				Description: "Look up the exact relationship and address pronouns between two characters (who calls whom what, how each refers to themselves, and the tone/attitude of their dialogue) from the L2 Knowledge Graph. Call this tool when translating dialogue or when unsure how two characters should address each other.",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"character_a": map[string]any{
							"type":        "string",
							"description": "Name or alias of the first character (e.g. the speaker)",
						},
						"character_b": map[string]any{
							"type":        "string",
							"description": "Name or alias of the second character (e.g. the listener)",
						},
					},
					"required": []string{"character_a", "character_b"},
				},
			},
		},
		{
			Type: "function",
			Function: FunctionSchema{
				Name:        "lookup_world_lore",
				Description: "Look up the world encyclopedia (World Bible / Lorebook), including: factions/corporations, magic or tech tiers, court ranks, cultivation realms, artifacts/weapons, and key locations. Call this tool when encountering setting-specific terminology, factions, or world rules that need clarification.",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"query": map[string]any{
							"type":        "string",
							"description": "Name of the entity, organization, realm, or world-building concept to look up",
						},
						"category": map[string]any{
							"type":        "string",
							"description": "Optional: category slug if known (e.g. factions, magic_ranks, cyberware, court_ranks, cultivation, items, locations)",
						},
					},
					"required": []string{"query"},
				},
			},
		},
		{
			Type: "function",
			Function: FunctionSchema{
				Name:        "ask_human_coworker",
				Description: "Ask the chief editor / human coworker for clarification when encountering ambiguous address terms, name transliteration dilemmas, or stylistic decisions.",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"question": map[string]any{
							"type":        "string",
							"description": "Clear question for the human coworker (e.g. 'Should character A address B as a junior sister or by her given name?')",
						},
						"options": map[string]any{
							"type": "array",
							"items": map[string]any{
								"type": "string",
							},
							"description": "Suggested candidate translations or options",
						},
						"context_snippet": map[string]any{
							"type":        "string",
							"description": "Short excerpt of the surrounding scene for context",
						},
					},
					"required": []string{"question"},
				},
			},
		},
	}
}

// Execute executes the requested tool call respecting temporal horizon and returns the formatted result.
func (tr *ToolRegistry) Execute(ctx context.Context, projectID string, currentChapter int64, call ToolCall) (string, error) {
	switch call.Function.Name {
	case "search_book_context":
		var args struct {
			Query string `json:"query"`
		}
		if err := json.Unmarshal([]byte(call.Function.Arguments), &args); err != nil {
			return "", fmt.Errorf("unmarshal search_book_context arguments: %w", err)
		}

		if strings.TrimSpace(args.Query) == "" {
			return "No search query provided", nil
		}

		if tr.store == nil {
			return "The book storage database has not been initialized.", nil
		}

		results, err := tr.store.SearchBookContextTemporal(ctx, projectID, args.Query, currentChapter, 3)
		if err != nil {
			return fmt.Sprintf("Internal search error: %v", err), nil
		}

		if len(results) == 0 {
			return fmt.Sprintf("No passages found containing '%s' in book.", args.Query), nil
		}

		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("Book context search results for '%s':\n", args.Query))
		for i, res := range results {
			contentPreview := res.RawContent
			if res.TranslatedContent != "" {
				contentPreview = res.TranslatedContent
			}
			if len([]rune(contentPreview)) > 200 {
				contentPreview = string([]rune(contentPreview)[:200]) + "..."
			}
			sb.WriteString(fmt.Sprintf("[%d] Chapter %v: %s\n", i+1, res.ChapterIndex, contentPreview))
		}
		return sb.String(), nil

	case "web_lookup":
		var args struct {
			Keyword  string `json:"keyword"`
			Language string `json:"language"`
		}
		if err := json.Unmarshal([]byte(call.Function.Arguments), &args); err != nil {
			return "", fmt.Errorf("unmarshal web_lookup arguments: %w", err)
		}

		if strings.TrimSpace(args.Keyword) == "" {
			return "No keyword provided for web lookup", nil
		}

		if tr.researchEng == nil {
			return "Web research engine is not configured", nil
		}

		lang := args.Language
		if lang == "" {
			lang = "ja"
		}

		brief, err := tr.researchEng.Research(ctx, args.Keyword, lang)
		if err != nil {
			return fmt.Sprintf("Web research error: %v", err), nil
		}

		return brief.FormatForPrompt(), nil

	case "query_character_history":
		var args struct {
			CharacterName   string `json:"character_name"`
			TargetCharacter string `json:"target_character"`
		}
		if err := json.Unmarshal([]byte(call.Function.Arguments), &args); err != nil {
			return "", fmt.Errorf("unmarshal query_character_history arguments: %w", err)
		}

		if strings.TrimSpace(args.CharacterName) == "" {
			return "No character name provided", nil
		}

		if tr.store == nil {
			return "The character database has not been initialized.", nil
		}

		entities, err := tr.store.ListEntitiesByProject(ctx, projectID)
		if err != nil {
			return fmt.Sprintf("Failed to list entities: %v", err), nil
		}

		var matchedEntity *storage.Entity
		for i := range entities {
			e := &entities[i]
			if strings.EqualFold(e.Name, args.CharacterName) {
				matchedEntity = e
				break
			}
			for _, alias := range e.Aliases {
				if strings.EqualFold(alias, args.CharacterName) {
					matchedEntity = e
					break
				}
			}
			if matchedEntity != nil {
				break
			}
		}

		relations, _ := tr.store.ListRelationsByProject(ctx, projectID)
		var sb strings.Builder
		if matchedEntity != nil {
			sb.WriteString(fmt.Sprintf("Character: %s (Gender: %s, Role: %s, First seen: Ch.%d)\n",
				matchedEntity.Name, matchedEntity.Gender, matchedEntity.Role, matchedEntity.FirstSeenChapter))
			if len(matchedEntity.Aliases) > 0 {
				sb.WriteString(fmt.Sprintf("Aliases: %s\n", strings.Join(matchedEntity.Aliases, ", ")))
			}
		} else {
			sb.WriteString(fmt.Sprintf("Character '%s' is not yet formally indexed in L2 Graph.\n", args.CharacterName))
		}

		sb.WriteString("Active relations (up to current chapter):\n")
		foundRel := false
		for _, r := range relations {
			if r.SinceChapter > currentChapter {
				continue
			}
			isRelevant := strings.EqualFold(r.FromChar, args.CharacterName) || strings.EqualFold(r.ToChar, args.CharacterName)
			if args.TargetCharacter != "" {
				isRelevant = isRelevant && (strings.EqualFold(r.FromChar, args.TargetCharacter) || strings.EqualFold(r.ToChar, args.TargetCharacter))
			}
			if isRelevant {
				foundRel = true
				sb.WriteString(fmt.Sprintf("- %s calls %s as '%s' (self: '%s', tone: %s)\n",
					r.FromChar, r.ToChar, r.CallAs, r.SelfCallAs, r.Tone))
			}
		}
		if !foundRel {
			sb.WriteString("- No specific historical relations recorded for this character.\n")
		}
		return sb.String(), nil

	case "lookup_character_relation":
		var args struct {
			CharacterA string `json:"character_a"`
			CharacterB string `json:"character_b"`
		}
		if err := json.Unmarshal([]byte(call.Function.Arguments), &args); err != nil {
			return "", fmt.Errorf("unmarshal lookup_character_relation arguments: %w", err)
		}

		charA := strings.TrimSpace(args.CharacterA)
		charB := strings.TrimSpace(args.CharacterB)
		if charA == "" || charB == "" {
			return "Both characters (character_a and character_b) are required to look up their relationship and address terms.", nil
		}

		if tr.store == nil {
			return "The character graph database has not been initialized.", nil
		}

		// First resolve canonical entity profiles for gender, role, and alias resolution
		entities, _ := tr.store.ListEntitiesByProject(ctx, projectID)
		var entA, entB *storage.Entity
		for i := range entities {
			e := &entities[i]
			if strings.EqualFold(e.Name, charA) || hasAlias(e.Aliases, charA) {
				entA = e
			}
			if strings.EqualFold(e.Name, charB) || hasAlias(e.Aliases, charB) {
				entB = e
			}
		}

		matchesA := func(name string) bool {
			if strings.EqualFold(name, charA) {
				return true
			}
			if entA != nil {
				return strings.EqualFold(name, entA.Name) || hasAlias(entA.Aliases, name)
			}
			return false
		}
		matchesB := func(name string) bool {
			if strings.EqualFold(name, charB) {
				return true
			}
			if entB != nil {
				return strings.EqualFold(name, entB.Name) || hasAlias(entB.Aliases, name)
			}
			return false
		}

		relations, err := tr.store.ListRelationsByProject(ctx, projectID)
		if err != nil {
			return fmt.Sprintf("Relationship query error: %v", err), nil
		}

		var matchingRels []sqlc.CharacterRelation
		for _, r := range relations {
			if r.SinceChapter > currentChapter {
				continue
			}
			isAtoB := matchesA(r.FromChar) && matchesB(r.ToChar)
			isBtoA := matchesB(r.FromChar) && matchesA(r.ToChar)
			if isAtoB || isBtoA {
				matchingRels = append(matchingRels, r)
			}
		}

		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("[RELATIONSHIP & ADDRESS TERMS]: %s ↔ %s (up to Chapter %d)\n", charA, charB, currentChapter))
		if entA != nil {
			sb.WriteString(fmt.Sprintf("- %s: Gender: %s, Role: %s\n", entA.Name, entA.Gender, entA.Role))
		}
		if entB != nil {
			sb.WriteString(fmt.Sprintf("- %s: Gender: %s, Role: %s\n", entB.Name, entB.Gender, entB.Role))
		}

		if len(matchingRels) > 0 {
			sb.WriteString("Address conventions already established in this work:\n")
			for _, r := range matchingRels {
				sb.WriteString(fmt.Sprintf("  * %s calls %s: '%s' (self-reference: '%s', tone: %s, since ch.%d)\n",
					r.FromChar, r.ToChar, r.CallAs, r.SelfCallAs, r.Tone, r.SinceChapter))
			}
		} else {
			sb.WriteString("No direct address record exists between these two characters yet. ")
			if entA != nil && entB != nil {
				sb.WriteString(fmt.Sprintf("Role-based suggestion: %s (%s) toward %s (%s) should use pronouns appropriate to their social relationship.",
					entA.Name, entA.Role, entB.Name, entB.Role))
			}
		}
		return sb.String(), nil

	case "lookup_world_lore":
		var args struct {
			Query    string `json:"query"`
			Category string `json:"category"`
		}
		if err := json.Unmarshal([]byte(call.Function.Arguments), &args); err != nil {
			return "", fmt.Errorf("unmarshal lookup_world_lore arguments: %w", err)
		}

		q := strings.TrimSpace(args.Query)
		if q == "" {
			return "A search keyword for the world encyclopedia is required.", nil
		}

		if tr.store == nil {
			return "The world-bible database has not been initialized.", nil
		}

		// 1. Exact match by name
		exact, err := tr.store.GetWorldEntryByName(ctx, projectID, strings.ToLower(q))
		var matchedEntries []sqlc.WorldEntry
		if err == nil && exact.ID != "" {
			matchedEntries = append(matchedEntries, exact)
		} else {
			// Search all entries by name substring, alias, or summary
			allEntries, err := tr.store.ListWorldEntries(ctx, projectID)
			if err == nil {
				lowerQ := strings.ToLower(q)
				for _, e := range allEntries {
					if args.Category != "" {
						cat, _ := tr.store.GetWorldCategoryByID(ctx, e.CategoryID)
						if cat.Slug != args.Category && !strings.EqualFold(cat.Name, args.Category) {
							continue
						}
					}

					if strings.Contains(strings.ToLower(e.Name), lowerQ) || strings.Contains(strings.ToLower(e.Summary), lowerQ) {
						matchedEntries = append(matchedEntries, e)
						if len(matchedEntries) >= 3 {
							break
						}
						continue
					}

					var aliases []string
					if json.Unmarshal([]byte(e.AliasesJson), &aliases) == nil {
						for _, alias := range aliases {
							if strings.Contains(strings.ToLower(alias), lowerQ) {
								matchedEntries = append(matchedEntries, e)
								break
							}
						}
					}
					if len(matchedEntries) >= 3 {
						break
					}
				}
			}
		}

		if len(matchedEntries) == 0 {
			return fmt.Sprintf("No World Bible entry matched the keyword '%s'. Translate using the available context, or ask your coworker if needed.", q), nil
		}

		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("[WORLD BIBLE ENCYCLOPEDIA] Results for '%s':\n", q))
		for i, e := range matchedEntries {
			cat, _ := tr.store.GetWorldCategoryByID(ctx, e.CategoryID)
			catName := "General"
			if cat.Name != "" {
				catName = cat.Name
			}
			sb.WriteString(fmt.Sprintf("[%d] %s (Category: %s)\n", i+1, e.Name, catName))
			if e.Summary != "" {
				sb.WriteString(fmt.Sprintf("  - Summary: %s\n", e.Summary))
			}
			if e.AttributesJson != "" && e.AttributesJson != "{}" {
				sb.WriteString(fmt.Sprintf("  - Attributes: %s\n", e.AttributesJson))
			}
			if e.FullDescription != "" {
				desc := e.FullDescription
				if len([]rune(desc)) > 300 {
					desc = string([]rune(desc)[:300]) + "..."
				}
				sb.WriteString(fmt.Sprintf("  - Details: %s\n", desc))
			}
		}
		return sb.String(), nil

	case "ask_human_coworker":
		var args struct {
			Question       string   `json:"question"`
			Options        []string `json:"options"`
			ContextSnippet string   `json:"context_snippet"`
		}
		if err := json.Unmarshal([]byte(call.Function.Arguments), &args); err != nil {
			return "", fmt.Errorf("unmarshal ask_human_coworker arguments: %w", err)
		}

		// Auto-Pilot Mode: Arbitrate via Shadow Critic based on Style Guide
		if strings.EqualFold(tr.decisionMode, "auto") {
			answer, _ := tr.ArbitrateDilemma(ctx, args.Question, args.Options, args.ContextSnippet)
			tr.AutoLearnResolution(ctx, projectID, currentChapter, args.Question, answer, args.Options)
			return fmt.Sprintf("Auto-Pilot Decision: Shadow Critic selected '%s' for: %s", answer, args.Question), nil
		}

		// Manual Coworker Mode: Call AskHumanHandler if registered
		if tr.AskHumanHandler != nil {
			answer, err := tr.AskHumanHandler(ctx, args.Question, args.Options, args.ContextSnippet)
			if err != nil {
				if ctx.Err() != nil {
					return "", ctx.Err()
				}
				return fmt.Sprintf("Chief Editor communication error: %v", err), nil
			}
			if strings.TrimSpace(answer) != "" {
				return fmt.Sprintf("Human Chief Editor responded: %s", answer), nil
			}
		}

		// Fallback when human is unattended or no handler attached
		if len(args.Options) > 0 {
			return fmt.Sprintf("Chief Editor is away. Default recommended option chosen: '%s'. Question was: %s",
				args.Options[0], args.Question), nil
		}
		return fmt.Sprintf("Chief Editor is away. Proceeding with best linguistic context for: %s", args.Question), nil

	default:
		return "", fmt.Errorf("unrecognized tool '%s'", call.Function.Name)
	}
}

// CleanAnswer extracts the core answer text, stripping any chat wrapper prefixes or surrounding quotes.
func CleanAnswer(rawAnswer string) string {
	cleaned := strings.TrimSpace(rawAnswer)
	// NOTE: the Vietnamese prefixes are intentional input-matching data kept for
	// backward compatibility with previously stored HITL answers.
	prefixes := []string{
		"[HITL] I choose: ",
		"I choose option: ",
		"I choose: ",
		"Custom instruction: ",
		"[HITL] Tôi chọn: ",
		"Tôi chọn phương án: ",
		"Tôi chọn: ",
		"Chỉ thị tùy chỉnh: ",
		"[HITL] ",
	}
	for _, p := range prefixes {
		if strings.HasPrefix(cleaned, p) {
			cleaned = strings.TrimPrefix(cleaned, p)
		}
	}
	cleaned = strings.Trim(cleaned, `"'“”‘’`)
	return strings.TrimSpace(cleaned)
}

// AutoLearnResolution automatically records resolved dilemmas into Glossary (L3) or Entity Graph (L2).
func (tr *ToolRegistry) AutoLearnResolution(ctx context.Context, projectID string, chapterIndex int64, question string, answer string, options []string) {
	if tr.store == nil || projectID == "" || strings.TrimSpace(answer) == "" {
		return
	}

	trimmedAnswer := CleanAnswer(answer)
	if trimmedAnswer == "" {
		return
	}

	// 1. Extract quoted terms from question
	reQuote := regexp.MustCompile(`["“'‘]([^"“'‘]+)["”'’]`)
	matches := reQuote.FindAllStringSubmatch(question, -1)

	// Filter matches: ignore any quoted string that is one of the options or equals the answer
	var candidates []string
	for _, m := range matches {
		term := strings.TrimSpace(m[1])
		if term == "" {
			continue
		}
		isOption := false
		for _, opt := range options {
			if strings.EqualFold(term, strings.TrimSpace(opt)) {
				isOption = true
				break
			}
		}
		if isOption || strings.EqualFold(term, trimmedAnswer) {
			continue
		}
		candidates = append(candidates, term)
	}

	lowerQ := strings.ToLower(question)
	isRelationQuestion := strings.Contains(lowerQ, "gọi") ||
		strings.Contains(lowerQ, "xưng hô") ||
		strings.Contains(lowerQ, "quan hệ") ||
		strings.Contains(lowerQ, "address") ||
		strings.Contains(lowerQ, "call")

	if isRelationQuestion && len(candidates) >= 2 {
		// Two entities dilemma -> character relation L2
		charA := candidates[0]
		charB := candidates[1]
		if charA != "" && charB != "" {
			_ = tr.store.UpsertRelation(ctx, sqlc.UpsertRelationParams{
				ID:           "hitl_rel_" + strings.ReplaceAll(uuid.New().String(), "-", "")[:12],
				ProjectID:    projectID,
				FromChar:     charA,
				ToChar:       charB,
				CallAs:       trimmedAnswer,
				SelfCallAs:   "",
				SinceChapter: chapterIndex,
				Tone:         "respectful",
			})
		}
	} else if len(candidates) >= 1 {
		// Single term (or first non-option candidate) -> Glossary L3
		sourceTerm := candidates[0]
		if sourceTerm != "" && !strings.EqualFold(sourceTerm, trimmedAnswer) {
			_ = tr.store.UpsertGlossaryTerm(ctx, sqlc.UpsertGlossaryTermParams{
				ID:         "hitl_" + strings.ReplaceAll(uuid.New().String(), "-", "")[:12],
				ProjectID:  projectID,
				SourceTerm: sourceTerm,
				TargetTerm: trimmedAnswer,
				Category:   "hitl_decision",
				Notes:      fmt.Sprintf("Auto-learned from HITL decision at chapter %d: %s", chapterIndex, question),
			})
		}
	}
}

func hasAlias(aliases []string, name string) bool {
	for _, a := range aliases {
		if strings.EqualFold(a, name) {
			return true
		}
	}
	return false
}

