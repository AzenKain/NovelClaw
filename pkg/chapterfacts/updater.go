package chapterfacts

import (
	"context"
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"
	"novelclaw/internal/gen/sqlc"
	"novelclaw/pkg/storage"
)

// GraphUpdater automates updating L2 temporal graph structures from episodic chapter facts.
type GraphUpdater struct {
	store *storage.Storage
}

// NewGraphUpdater creates an initialized GraphUpdater.
func NewGraphUpdater(store *storage.Storage) *GraphUpdater {
	return &GraphUpdater{store: store}
}

// Apply updates project entities and character relations in the dynamic graph.
func (u *GraphUpdater) Apply(ctx context.Context, projectID string, chapterIndex int64, facts *ChapterFacts) error {
	if facts == nil || u.store == nil {
		return nil
	}

	for _, intro := range facts.CastIntros {
		name := strings.TrimSpace(intro.Name)
		if name == "" {
			continue
		}
		ent := storage.Entity{
			ID:               fmt.Sprintf("ent_%s_%s", projectID, name),
			ProjectID:        projectID,
			Name:             name,
			Role:             intro.Role,
			Gender:           intro.Gender,
			FirstSeenChapter: chapterIndex,
			Category:         "character",
		}
		if err := u.store.UpsertEntity(ctx, ent); err != nil {
			log.Warn().
				Err(err).
				Str("character", name).
				Int64("chapter", chapterIndex).
				Msg("failed to auto-upsert newly discovered entity")
		} else {
			log.Info().
				Str("character", name).
				Int64("chapter", chapterIndex).
				Msg("auto-registered new character entity into L2 graph")
		}
	}

	for _, rel := range facts.RelationshipChanges {
		from := strings.TrimSpace(rel.FromChar)
		to := strings.TrimSpace(rel.ToChar)
		callAs := strings.TrimSpace(rel.CallAs)
		if from == "" || to == "" || callAs == "" {
			continue
		}

		relID := fmt.Sprintf("rel_%s_%s_%s_ch%d", projectID, from, to, chapterIndex)
		params := sqlc.UpsertRelationParams{
			ID:           relID,
			ProjectID:    projectID,
			FromChar:     from,
			ToChar:       to,
			CallAs:       callAs,
			SelfCallAs:   strings.TrimSpace(rel.SelfCallAs),
			SinceChapter: chapterIndex,
			Tone:         strings.TrimSpace(rel.Tone),
			IsLocked:     0,
		}

		if err := u.store.UpsertRelation(ctx, params); err != nil {
			log.Warn().
				Err(err).
				Str("from", from).
				Str("to", to).
				Int64("chapter", chapterIndex).
				Msg("failed to update character relation in L2 graph")
		} else {
			log.Info().
				Str("from", from).
				Str("to", to).
				Str("call_as", callAs).
				Int64("chapter", chapterIndex).
				Msg("auto-updated character relationship address term in L2 graph")
		}
	}

	return nil
}
