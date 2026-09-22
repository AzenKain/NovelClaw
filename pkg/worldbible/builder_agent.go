package worldbible

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/wailsapp/wails/v3/pkg/application"

	"novelclaw/internal/dtos"
	"novelclaw/internal/gen/sqlc"
	"novelclaw/pkg/jsonx"
	"novelclaw/pkg/llm"
	"novelclaw/pkg/storage"
)

// BuilderAgent performs autonomous reading, genre inference, and Lorebook taxonomy synthesis.
type BuilderAgent struct {
	store *storage.Storage
}

// NewBuilderAgent creates an Autonomous World Builder Agent.
func NewBuilderAgent(store *storage.Storage) *BuilderAgent {
	return &BuilderAgent{store: store}
}

// LLMWorldScanResult defines the JSON schema returned by the LLM.
type LLMWorldScanResult struct {
	InferredGenre string `json:"inferred_genre"`
	SettingSummary string `json:"setting_summary"`
	Categories    []struct {
		Slug        string `json:"slug"`
		Name        string `json:"name"`
		Icon        string `json:"icon"`
		Description string `json:"description"`
	} `json:"categories"`
	Entries []struct {
		CategorySlug    string            `json:"category_slug"`
		Name            string            `json:"name"`
		Aliases         []string          `json:"aliases"`
		Summary         string            `json:"summary"`
		FullDescription string            `json:"full_description"`
		Attributes      map[string]string `json:"attributes"`
	} `json:"entries"`
}

// ScanAndBuildWorld executes deep ingestion scan over chapters to construct the World Bible.
func (b *BuilderAgent) ScanAndBuildWorld(
	ctx context.Context,
	client llm.LLMClient,
	modelName string,
	projectID string,
	startChap, endChap int64,
) (*dtos.ScanWorldResponse, error) {
	if b.store == nil || projectID == "" {
		return nil, fmt.Errorf("invalid storage or project id")
	}
	if client == nil {
		return nil, fmt.Errorf("llm client is not configured")
	}

	// 1. Fetch chapter samples
	chapters, err := b.store.ListChaptersByProject(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("list chapters for world building: %w", err)
	}

	var textBuilder strings.Builder
	count := 0
	for _, ch := range chapters {
		if ch.ChapterIndex >= startChap && ch.ChapterIndex <= endChap {
			textBuilder.WriteString(fmt.Sprintf("\n--- CHAPTER %d: %s ---\n", ch.ChapterIndex, ch.Title))
			// Take up to 2500 runes per chapter to balance context window safely across multi-byte UTF-8
			content := ch.RawContent
			runes := []rune(content)
			if len(runes) > 2500 {
				content = string(runes[:2500]) + "..."
			}
			textBuilder.WriteString(content)
			textBuilder.WriteString("\n")
			count++
			if count >= 8 {
				break
			}
		}
	}

	sampleText := textBuilder.String()
	if strings.TrimSpace(sampleText) == "" {
		return &dtos.ScanWorldResponse{
			Success:       false,
			SummaryReport: "No chapter content found in the selected range to synthesize the World Bible.",
		}, nil
	}

	sysPrompt := `You are the Autonomous World Builder Agent.
Your task is to analyze literature text and reconstruct the underlying World Bible / Lorebook.
You must be completely genre-agnostic:
- Cyberpunk/Sci-Fi: Megacorps, Cyberware, Netrunner ranks, Space colonies, AI networks.
- High Fantasy/LitRPG: Magic circles/systems, Races (Elf, Dwarf, Beastmen), Guilds, Dungeons, Ancient Artifacts.
- Historical/Court: Imperial harem ranks, Nine-rank court bureaucracy, Aristocratic titles, Clan lineage, Etiquette.
- Urban/Thriller: Crime syndicates, Intelligence agencies, City sectors, Case timelines.
- Xianxia/Wuxia: Cultivation realms, Sects/Schools, Pills, Spiritual treasures, Meridians.

INSTRUCTIONS:
1. Infer the primary Genre and Setting.
2. Create 3 to 6 dynamic taxonomy categories suited for this world. Assign each an appropriate Lucide icon name: 'Cpu', 'Crown', 'Wand2', 'Building2', 'Sword', 'Shield', 'Skull', 'Globe', 'BookOpen', 'Key', 'Layers', 'Boxes'.
3. Extract up to 15 key world lore entries (factions, hierarchies, locations, power systems, technologies, artifacts) with concise 1-2 sentence summaries, aliases, and key attributes (key-value strings).

OUTPUT FORMAT:
Respond with STRICT VALID JSON only, no markdown markdown ticks outside the JSON object:
{
  "inferred_genre": "Cyberpunk / High Fantasy / etc.",
  "setting_summary": "1-2 sentence summary of the universe",
  "categories": [
    { "slug": "factions", "name": "Tech Megacorporations", "icon": "Building2", "description": "Conglomerates dominating society and economy" }
  ],
  "entries": [
    {
      "category_slug": "factions",
      "name": "Arasaka",
      "aliases": ["Arasaka Corp", "Arasaka Conglomerate"],
      "summary": "Foremost security and weaponry conglomerate.",
      "full_description": "Comprehensive encyclopedia details...",
      "attributes": { "tier": "S-Rank", "leader": "Saburo Arasaka" }
    }
  ]
}`

	userPrompt := fmt.Sprintf("Analyze the following novel chapters and synthesize the World Bible:\n\n%s", sampleText)

	resp, err := client.Generate(ctx, llm.CompletionRequest{
		Model: modelName,
		Messages: []llm.Message{
			{Role: llm.RoleSystem, Content: sysPrompt},
			{Role: llm.RoleUser, Content: userPrompt},
		},
		Temperature: 0.2,
	})
	if err != nil {
		return nil, fmt.Errorf("llm world scan generation: %w", err)
	}

	cleanedJSON := jsonx.ExtractJSONObject(resp.Content)

	var scanResult LLMWorldScanResult
	if err := json.Unmarshal([]byte(cleanedJSON), &scanResult); err != nil {
		log.Warn().Err(err).Str("raw", cleanedJSON).Msg("failed to parse world builder JSON")
		return &dtos.ScanWorldResponse{
			Success:       false,
			SummaryReport: fmt.Sprintf("Model response did not follow valid JSON schema: %s", err.Error()),
		}, nil
	}

	// 2. Persist Categories
	slugToCatID := make(map[string]string)
	catsAdded := 0
	for idx, c := range scanResult.Categories {
		slug := strings.TrimSpace(strings.ToLower(c.Slug))
		if slug == "" {
			slug = fmt.Sprintf("cat_%d", idx+1)
		}
		catID := fmt.Sprintf("cat_%s", strings.ReplaceAll(uuid.New().String(), "-", "")[:12])
		icon := c.Icon
		if icon == "" {
			icon = "Boxes"
		}

		created, err := b.store.UpsertWorldCategory(ctx, sqlc.UpsertWorldCategoryParams{
			ID:           catID,
			ProjectID:    projectID,
			Slug:         slug,
			Name:         c.Name,
			Icon:         icon,
			Description:  c.Description,
			DisplayOrder: int64(idx + 1),
		})
		if err == nil {
			slugToCatID[slug] = created.ID
			catsAdded++
		} else {
			// If conflict, find existing
			if existing, err := b.store.GetWorldCategoryBySlug(ctx, projectID, slug); err == nil {
				slugToCatID[slug] = existing.ID
			}
		}
	}

	// 3. Persist Entries
	entriesAdded := 0
	for _, e := range scanResult.Entries {
		catID, ok := slugToCatID[strings.ToLower(e.CategorySlug)]
		if !ok || catID == "" {
			// Fallback to first created category
			for _, id := range slugToCatID {
				catID = id
				break
			}
		}
		if catID == "" {
			continue
		}

		entryID := fmt.Sprintf("entry_%s", strings.ReplaceAll(uuid.New().String(), "-", "")[:12])
		var isVerified int64 = 0
		if existing, err := b.store.GetWorldEntryByName(ctx, projectID, strings.ToLower(e.Name)); err == nil && existing.ID != "" {
			entryID = existing.ID
			isVerified = existing.IsVerified // Preserve verification if previously verified
		}

		aliasesJSON, _ := json.Marshal(e.Aliases)
		if len(e.Aliases) == 0 {
			aliasesJSON = []byte("[]")
		}
		attrsJSON, _ := json.Marshal(e.Attributes)
		if e.Attributes == nil {
			attrsJSON = []byte("{}")
		}

		_, err := b.store.UpsertWorldEntry(ctx, sqlc.UpsertWorldEntryParams{
			ID:                 entryID,
			ProjectID:          projectID,
			CategoryID:         catID,
			Name:               e.Name,
			AliasesJson:        string(aliasesJSON),
			Summary:            e.Summary,
			FullDescription:    e.FullDescription,
			AttributesJson:     string(attrsJSON),
			DiscoveredBy:       "ai_scan",
			SourceChapterIndex: startChap,
			IsVerified:         isVerified, // Preserve user verification if already verified, else 0 (AI suggestion)
		})
		if err == nil {
			entriesAdded++
		}
	}

	report := fmt.Sprintf(
		"Detected genre: **%s**.\nSynthesized **%d world taxonomy categories** and **%d lorebook entries** from chapters %d to %d.",
		scanResult.InferredGenre, catsAdded, entriesAdded, startChap, endChap,
	)

	// Emit event to Wails frontend
	if app := application.Get(); app != nil && app.Event != nil {
		app.Event.Emit("world:scanned", map[string]any{
			"project_id":       projectID,
			"inferred_genre":   scanResult.InferredGenre,
			"categories_added": catsAdded,
			"entries_added":    entriesAdded,
		})
	}

	return &dtos.ScanWorldResponse{
		Success:         true,
		InferredGenre:   scanResult.InferredGenre,
		CategoriesAdded: catsAdded,
		EntriesAdded:    entriesAdded,
		SummaryReport:   report,
	}, nil
}
