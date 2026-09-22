package graph

import (
	"context"
	"strings"
	"testing"

	"novelclaw/internal/gen/sqlc"
	"novelclaw/pkg/storage"
)

func TestProgressiveGraph_BuildSelectiveContext(t *testing.T) {
	store, err := storage.OpenStorage(":memory:")
	if err != nil {
		t.Fatalf("open storage: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	projectID := "proj_graph_test"

	_ = store.CreateProject(ctx, sqlc.CreateProjectParams{
		ID:         projectID,
		Title:      "A Sister's All You Need (Graph Test)",
		SourceLang: "ja",
		TargetLang: "vi",
	})

	_ = store.UpsertEntity(ctx, storage.Entity{
		ID:               "ent_itsuki",
		ProjectID:        projectID,
		Name:             "羽島伊月",
		Aliases:          []string{"伊月", "Itsuki"},
		Category:         "character",
		Gender:           "male",
		Role:             "Protagonist / Novelist",
		FirstSeenChapter: 1,
	})

	_ = store.UpsertEntity(ctx, storage.Entity{
		ID:               "ent_nayuta",
		ProjectID:        projectID,
		Name:             "可児那由多",
		Aliases:          []string{"那由多", "Nayuta"},
		Category:         "character",
		Gender:           "female",
		Role:             "Genius Novelist",
		FirstSeenChapter: 1,
	})

	_ = store.UpsertEntity(ctx, storage.Entity{
		ID:               "ent_miyako",
		ProjectID:        projectID,
		Name:             "白川京",
		Aliases:          []string{"京", "Miyako"},
		Category:         "character",
		Gender:           "female",
		Role:             "College Friend",
		FirstSeenChapter: 1,
	})

	_ = store.UpsertRelation(ctx, sqlc.UpsertRelationParams{
		ID:           "rel_nayuta_itsuki_v1",
		ProjectID:    projectID,
		FromChar:     "可児那由多",
		ToChar:       "羽島伊月",
		CallAs:       "先輩",
		SelfCallAs:   "わたし",
		SinceChapter: 1,
		Tone:         "adoring",
	})

	_ = store.UpsertRelation(ctx, sqlc.UpsertRelationParams{
		ID:           "rel_nayuta_itsuki_v5",
		ProjectID:    projectID,
		FromChar:     "可児那由多",
		ToChar:       "羽島伊月",
		CallAs:       "ダーリン",
		SelfCallAs:   "妻",
		SinceChapter: 5,
		Tone:         "romantic",
	})

	_ = store.UpsertGlossaryTerm(ctx, sqlc.UpsertGlossaryTermParams{
		ID:         "g_trpg",
		ProjectID:  projectID,
		SourceTerm: "TRPG",
		TargetTerm: "Game nhập vai trên bàn",
		Category:   "game",
	})

	_ = store.UpsertGlossaryTerm(ctx, sqlc.UpsertGlossaryTermParams{
		ID:         "g_magic",
		ProjectID:  projectID,
		SourceTerm: "Excalibur",
		TargetTerm: "Thánh kiếm Excalibur",
		Category:   "weapon",
	})

	pg := NewProgressiveGraph(store)

	scene1 := "伊月は机に向かっていた。すると那由多が「原稿書けたよ！」とTRPGのサイコロを転がしながら飛び込んできた。"

	ctxChap2, err := pg.BuildSelectiveContext(ctx, projectID, 2, scene1)
	if err != nil {
		t.Fatalf("BuildSelectiveContext chap 2: %v", err)
	}

	if len(ctxChap2.ActiveEntities) != 2 {
		t.Errorf("expected 2 active entities (Itsuki, Nayuta), got %d", len(ctxChap2.ActiveEntities))
	}
	if strings.Contains(ctxChap2.FormattedPrompt, "白川京") {
		t.Errorf("unwanted character 白川京 found in prompt: %s", ctxChap2.FormattedPrompt)
	}
	if !strings.Contains(ctxChap2.FormattedPrompt, "先輩") {
		t.Errorf("expected Chapter 2 relation '先輩' not found: %s", ctxChap2.FormattedPrompt)
	}
	if strings.Contains(ctxChap2.FormattedPrompt, "ダーリン") {
		t.Errorf("future Chapter 5 relation 'ダーリン' leaked into Chapter 2: %s", ctxChap2.FormattedPrompt)
	}
	if !strings.Contains(ctxChap2.FormattedPrompt, "Game nhập vai trên bàn") {
		t.Errorf("expected glossary term TRPG not found: %s", ctxChap2.FormattedPrompt)
	}
	if strings.Contains(ctxChap2.FormattedPrompt, "Excalibur") {
		t.Errorf("unrelated glossary term Excalibur leaked into scene: %s", ctxChap2.FormattedPrompt)
	}

	ctxChap6, err := pg.BuildSelectiveContext(ctx, projectID, 6, scene1)
	if err != nil {
		t.Fatalf("BuildSelectiveContext chap 6: %v", err)
	}
	if !strings.Contains(ctxChap6.FormattedPrompt, "ダーリン") {
		t.Errorf("expected Chapter 6 relation 'ダーリン' not found: %s", ctxChap6.FormattedPrompt)
	}
}

func TestProgressiveGraph_MetadataFactsInjection(t *testing.T) {
	store, err := storage.OpenStorage(":memory:")
	if err != nil {
		t.Fatalf("open storage: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	projectID := "proj_facts_test"

	_ = store.CreateProject(ctx, sqlc.CreateProjectParams{
		ID:         projectID,
		Title:      "Facts Injection Test",
		SourceLang: "ja",
		TargetLang: "vi",
	})

	_ = store.UpsertEntity(ctx, storage.Entity{
		ID:               "ent_chihiro",
		ProjectID:        projectID,
		Name:             "羽島千尋",
		Aliases:          []string{"千尋", "Chihiro"},
		Category:         "character",
		Gender:           "female",
		Role:             "Heroine",
		FirstSeenChapter: 1,
		Metadata: map[string]any{
			"description": "Em gái kế của Itsuki",
			"facts":       []string{"Cải trang thành em trai", "Nấu ăn rất giỏi"},
		},
	})

	pg := NewProgressiveGraph(store)
	scene := "千尋がキッチンで朝食を作っていた。"

	sc, err := pg.BuildSelectiveContext(ctx, projectID, 1, scene)
	if err != nil {
		t.Fatalf("BuildSelectiveContext: %v", err)
	}

	if !strings.Contains(sc.FormattedPrompt, "羽島千尋") {
		t.Errorf("expected Chihiro in prompt: %s", sc.FormattedPrompt)
	}
	if !strings.Contains(sc.FormattedPrompt, "Em gái kế của Itsuki") {
		t.Errorf("expected description in prompt: %s", sc.FormattedPrompt)
	}
	if !strings.Contains(sc.FormattedPrompt, "Cải trang thành em trai") || !strings.Contains(sc.FormattedPrompt, "Nấu ăn rất giỏi") {
		t.Errorf("expected facts in prompt: %s", sc.FormattedPrompt)
	}
}
