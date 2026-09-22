package worldbible

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"novelclaw/internal/dtos"
	"novelclaw/internal/gen/sqlc"
	"novelclaw/pkg/storage"
)

// ExportBundle represents the full structured backup of a project's Lorebook.
type ExportBundle struct {
	ProjectID  string                  `json:"project_id"`
	ExportedAt string                  `json:"exported_at"`
	Categories []dtos.WorldCategoryDTO `json:"categories"`
	Entries    []dtos.WorldEntryDTO    `json:"entries"`
}

// ExportMarkdown renders a structured Markdown World Bible.
func ExportMarkdown(projectTitle string, categories []dtos.WorldCategoryDTO, entries []dtos.WorldEntryDTO) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# WORLD BIBLE ENCYCLOPEDIA: %s\n\n", projectTitle))
	sb.WriteString("> *Comprehensive lorebook documenting universe settings, factions, hierarchies, and power systems.*\n\n---\n\n")

	entriesByCat := make(map[string][]dtos.WorldEntryDTO)
	for _, e := range entries {
		entriesByCat[e.CategoryID] = append(entriesByCat[e.CategoryID], e)
	}

	for _, c := range categories {
		sb.WriteString(fmt.Sprintf("## %s (%s)\n", c.Name, c.Slug))
		if c.Description != "" {
			sb.WriteString(fmt.Sprintf("*%s*\n\n", c.Description))
		} else {
			sb.WriteString("\n")
		}

		catEntries := entriesByCat[c.ID]
		if len(catEntries) == 0 {
			sb.WriteString("*(No entries in this category)*\n\n")
			continue
		}

		for _, e := range catEntries {
			statusTag := "[AI Suggested]"
			if e.IsVerified {
				statusTag = "[Verified]"
			}

			sb.WriteString(fmt.Sprintf("### %s %s\n", e.Name, statusTag))
			if len(e.Aliases) > 0 {
				sb.WriteString(fmt.Sprintf("- **Aliases**: %s\n", strings.Join(e.Aliases, ", ")))
			}
			if e.Summary != "" {
				sb.WriteString(fmt.Sprintf("- **Summary**: %s\n", e.Summary))
			}
			if len(e.Attributes) > 0 {
				sb.WriteString("- **Attributes**:\n")
				for k, v := range e.Attributes {
					sb.WriteString(fmt.Sprintf("  - *%s*: %s\n", k, v))
				}
			}
			if e.FullDescription != "" {
				sb.WriteString(fmt.Sprintf("\n%s\n", e.FullDescription))
			}
			sb.WriteString("\n---\n\n")
		}
	}

	return sb.String()
}

// ExportJSON serializes all categories and entries to JSON.
func ExportJSON(projectID string, categories []dtos.WorldCategoryDTO, entries []dtos.WorldEntryDTO) (string, error) {
	bundle := ExportBundle{
		ProjectID:  projectID,
		Categories: categories,
		Entries:    entries,
	}
	b, err := json.MarshalIndent(bundle, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// ImportJSON restores categories and entries from JSON.
func ImportJSON(ctx context.Context, store *storage.Storage, projectID, content string) (int, int, error) {
	if store == nil || projectID == "" {
		return 0, 0, fmt.Errorf("invalid storage or project id")
	}

	var bundle ExportBundle
	if err := json.Unmarshal([]byte(content), &bundle); err != nil {
		return 0, 0, fmt.Errorf("parse json bundle: %w", err)
	}

	slugToID := make(map[string]string)
	catsCount := 0
	for _, c := range bundle.Categories {
		id := c.ID
		if id == "" {
			id = fmt.Sprintf("cat_%s", strings.ReplaceAll(uuid.New().String(), "-", "")[:12])
		}
		cat, err := store.UpsertWorldCategory(ctx, sqlc.UpsertWorldCategoryParams{
			ID:           id,
			ProjectID:    projectID,
			Slug:         c.Slug,
			Name:         c.Name,
			Icon:         c.Icon,
			Description:  c.Description,
			DisplayOrder: int64(c.DisplayOrder),
		})
		if err == nil {
			slugToID[c.Slug] = cat.ID
			catsCount++
		}
	}

	entriesCount := 0
	for _, e := range bundle.Entries {
		catID := e.CategoryID
		if mapped, ok := slugToID[e.CategorySlug]; ok && mapped != "" {
			catID = mapped
		}
		if catID == "" {
			continue
		}

		aliasesJSON, _ := json.Marshal(e.Aliases)
		attrsJSON, _ := json.Marshal(e.Attributes)

		id := e.ID
		if id == "" {
			id = fmt.Sprintf("entry_%s", strings.ReplaceAll(uuid.New().String(), "-", "")[:12])
		}

		var isVer int64 = 0
		if e.IsVerified {
			isVer = 1
		}

		_, err := store.UpsertWorldEntry(ctx, sqlc.UpsertWorldEntryParams{
			ID:                 id,
			ProjectID:          projectID,
			CategoryID:         catID,
			Name:               e.Name,
			AliasesJson:        string(aliasesJSON),
			Summary:            e.Summary,
			FullDescription:    e.FullDescription,
			AttributesJson:     string(attrsJSON),
			DiscoveredBy:       e.DiscoveredBy,
			SourceChapterIndex: e.SourceChapterIndex,
			IsVerified:         isVer,
		})
		if err == nil {
			entriesCount++
		}
	}

	return catsCount, entriesCount, nil
}
