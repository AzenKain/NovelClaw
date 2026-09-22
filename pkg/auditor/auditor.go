package auditor

import (
	"context"
	"fmt"
	"strings"

	"novelclaw/pkg/storage"
)

// PlotAnomalyType defines the category of detected narrative contradiction.
type PlotAnomalyType string

const (
	AnomalyDeceasedReappearance PlotAnomalyType = "deceased_reappearance" // Dead character acts or speaks
	AnomalyRealmRegression      PlotAnomalyType = "realm_regression"      // Cultivation / power rank regresses without explanation
	AnomalyFactionMismatch      PlotAnomalyType = "faction_mismatch"      // Character acts against established locked alignment
)

// PlotWarning describes a detected continuity flaw or logic anomaly in the narrative.
type PlotWarning struct {
	Type          PlotAnomalyType `json:"type"`
	CharacterName string          `json:"character_name"`
	ChapterIndex  int64           `json:"chapter_index"`
	Message       string          `json:"message"`
	Snippet       string          `json:"snippet"`
	Severity      string          `json:"severity"` // "warning", "critical", "info"
}

// PlotAuditor inspects translated chapters against known entity states in the temporal graph.
type PlotAuditor struct {
	store *storage.Storage
}

// NewPlotAuditor creates a PlotAuditor instance.
func NewPlotAuditor(store *storage.Storage) *PlotAuditor {
	return &PlotAuditor{
		store: store,
	}
}

// AuditChapter scans chapter text against known entities and relations in the project knowledge graph.
func (a *PlotAuditor) AuditChapter(
	ctx context.Context,
	projectID string,
	chapterIndex int64,
	chapterText string,
) ([]PlotWarning, error) {
	if a.store == nil || projectID == "" || strings.TrimSpace(chapterText) == "" {
		return nil, nil
	}

	// 1. Retrieve all entities in this project
	entities, err := a.store.ListEntitiesByProject(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("list entities for plot audit: %w", err)
	}

	var warnings []PlotWarning
	lowerText := strings.ToLower(chapterText)

	// Action indicator verbs suggesting proactive presence/speaking (English & input-matching variants)
	actionVerbs := []string{
		"said:", "replied:", "spoke:", "laughed:", "shouted:", "stepped forward", "appeared",
		"nói:", "nói :", "bảo:", "lên tiếng", "cười lạnh", "quát to", "vung kiếm",
		"tiến lên", "bước ra", "cười nói", "đáp lời", "gầm lên", "xuất hiện",
	}

	for _, ent := range entities {
		meta := ent.Metadata
		if meta == nil {
			continue
		}

		// Check status: deceased / dead
		status, _ := meta["status"].(string)
		statusLower := strings.ToLower(strings.TrimSpace(status))
		isDead := statusLower == "deceased" || statusLower == "dead" || statusLower == "killed" || statusLower == "đã chết" || statusLower == "tử trận" || statusLower == "bị giết"

		if isDead {
			// Check if character appears and performs actions in this chapter
			nameLower := strings.ToLower(ent.Name)
			if strings.Contains(lowerText, nameLower) {
				// Check if there is an active action verb near the name
				hasActiveAction := false
				var foundSnippet string

				for _, verb := range actionVerbs {
					pattern := nameLower + " " + verb
					if idx := strings.Index(lowerText, pattern); idx != -1 {
						hasActiveAction = true
						start := idx - 30
						if start < 0 {
							start = 0
						}
						end := idx + len(pattern) + 50
						if end > len(chapterText) {
							end = len(chapterText)
						}
						foundSnippet = chapterText[start:end]
						break
					}
				}

				if hasActiveAction {
					warnings = append(warnings, PlotWarning{
						Type:          AnomalyDeceasedReappearance,
						CharacterName: ent.Name,
						ChapterIndex:  chapterIndex,
						Severity:      "critical",
						Message: fmt.Sprintf(
							"PLOT CONTRADICTION WARNING: Character '%s' is recorded as deceased (status='%s') but appears with actions/dialogue in chapter %d.",
							ent.Name, status, chapterIndex,
						),
						Snippet: strings.TrimSpace(foundSnippet),
					})
				}
			}
		}

		nameLower := strings.ToLower(ent.Name)
		if !strings.Contains(lowerText, nameLower) {
			continue
		}

		// 2. Check realm regression (an implausible drop in cultivation/power realm).
		cultivationRanks := map[string]int{
			// English ranks
			"qi condensation": 1, "foundation establishment": 2, "core formation": 3,
			"nascent soul": 4, "soul formation": 5, "void refinement": 6,
			"body integration": 7, "great vehicle": 8, "tribulation transcendence": 9, "true immortal": 10,
			// Sino-Vietnamese / Pinyin equivalents
			"luyện khí": 1, "trúc cơ": 2, "kim đan": 3, "nguyên anh": 4, "hóa thần": 5,
			"luyện hư": 6, "hợp thể": 7, "đại thừa": 8, "độ kiếp": 9, "chân tiên": 10,
		}
		rawTier, _ := meta["tier"].(string)
		if rawTier == "" {
			rawTier, _ = meta["realm"].(string)
		}
		currentRankIdx := cultivationRanks[strings.ToLower(strings.TrimSpace(rawTier))]

		if currentRankIdx > 1 {
			// Scan if text mentions this character having a lower rank
			for lowerRank, lowerIdx := range cultivationRanks {
				if lowerIdx >= currentRankIdx {
					continue
				}
				patterns := []string{
					nameLower + " is only a " + lowerRank,
					nameLower + " is merely a " + lowerRank,
					nameLower + " was only a " + lowerRank,
					nameLower + " chỉ là " + lowerRank,
					nameLower + " bất quá là " + lowerRank,
					nameLower + " là " + lowerRank + " tu sĩ",
					nameLower + " vốn là " + lowerRank,
				}
				for _, p := range patterns {
					if idx := strings.Index(lowerText, p); idx != -1 {
						snippetStart := idx - 20
						if snippetStart < 0 {
							snippetStart = 0
						}
						snippetEnd := idx + len(p) + 30
						if snippetEnd > len(chapterText) {
							snippetEnd = len(chapterText)
						}
						warnings = append(warnings, PlotWarning{
							Type:          AnomalyRealmRegression,
							CharacterName: ent.Name,
							ChapterIndex:  chapterIndex,
							Severity:      "warning",
							Message: fmt.Sprintf(
								"REALM REGRESSION WARNING: Character '%s' was at realm '%s' (tier %d) but is described as '%s' (tier %d) in chapter %d.",
								ent.Name, rawTier, currentRankIdx, lowerRank, lowerIdx, chapterIndex,
							),
							Snippet: strings.TrimSpace(chapterText[snippetStart:snippetEnd]),
						})
						break
					}
				}
			}
		}

		// 3. Check faction mismatch / antagonistic alignment conflict.
		opposingFaction, _ := meta["opposing_faction"].(string)
		if opposingFaction == "" {
			opposingFaction, _ = meta["enemy_faction"].(string)
		}
		if opposingFaction != "" {
			oppLower := strings.ToLower(strings.TrimSpace(opposingFaction))
			mismatchPatterns := []string{
				nameLower + " joined " + oppLower,
				nameLower + " pledged to " + oppLower,
				nameLower + " is a member of " + oppLower,
				nameLower + " is a disciple of " + oppLower,
				nameLower + " surrendered to " + oppLower,
				nameLower + " quy thuận " + oppLower,
				nameLower + " gia nhập " + oppLower,
				nameLower + " là người của " + oppLower,
				nameLower + " là đệ tử của " + oppLower,
			}
			for _, p := range mismatchPatterns {
				if idx := strings.Index(lowerText, p); idx != -1 {
					snippetStart := idx - 20
					if snippetStart < 0 {
						snippetStart = 0
					}
					snippetEnd := idx + len(p) + 30
					if snippetEnd > len(chapterText) {
						snippetEnd = len(chapterText)
					}
					warnings = append(warnings, PlotWarning{
						Type:          AnomalyFactionMismatch,
						CharacterName: ent.Name,
						ChapterIndex:  chapterIndex,
						Severity:      "critical",
						Message: fmt.Sprintf(
							"FACTION CONFLICT WARNING: Character '%s' has '%s' as an opposing faction but is described as aligned with/pledging to it in chapter %d.",
							ent.Name, opposingFaction, chapterIndex,
						),
						Snippet: strings.TrimSpace(chapterText[snippetStart:snippetEnd]),
					})
					break
				}
			}
		}
	}

	return warnings, nil
}
