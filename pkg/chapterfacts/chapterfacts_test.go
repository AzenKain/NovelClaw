package chapterfacts

import (
	"context"
	"testing"

	"novelclaw/internal/gen/sqlc"
	"novelclaw/pkg/llm"
	"novelclaw/pkg/storage"
)

type mockFactsClient struct {
	response string
}

func (m *mockFactsClient) Generate(ctx context.Context, req llm.CompletionRequest) (*llm.CompletionResponse, error) {
	return &llm.CompletionResponse{Content: m.response}, nil
}

func (m *mockFactsClient) Stream(ctx context.Context, req llm.CompletionRequest) (<-chan llm.StreamChunk, error) {
	ch := make(chan llm.StreamChunk, 1)
	ch <- llm.StreamChunk{Content: m.response}
	close(ch)
	return ch, nil
}

func (m *mockFactsClient) ProviderName() string {
	return "mock_facts"
}

func TestExtractorSuccess(t *testing.T) {
	jsonPayload := `{
		"title": "Chương 5: Lời Thề Dưới Trăng",
		"summary": "Itsuki và Nayuta chính thức thấu hiểu tâm ý của nhau.",
		"key_events": ["Itsuki giúp Nayuta sửa bản thảo", "Nayuta gọi Itsuki là Darling"],
		"relationship_changes": [
			{"from_char": "Nayuta", "to_char": "Itsuki", "call_as": "Darling", "self_call_as": "Nayuta"}
		],
		"cast_intros": [
			{"name": "Setsuna", "role": "Biên tập viên mới", "gender": "female"}
		]
	}`

	extractor := NewExtractor()
	mock := &mockFactsClient{response: jsonPayload}

	facts, err := extractor.Extract(context.Background(), mock, "Chương 5", "Văn bản chương 5...")
	if err != nil {
		t.Fatalf("expected successful extraction, got err: %v", err)
	}

	if facts.Summary == "" || len(facts.RelationshipChanges) != 1 || len(facts.CastIntros) != 1 {
		t.Errorf("unexpected facts extracted: %+v", facts)
	}

	if facts.RelationshipChanges[0].CallAs != "Darling" {
		t.Errorf("expected call_as Darling, got: %s", facts.RelationshipChanges[0].CallAs)
	}
}

func TestGraphUpdaterApply(t *testing.T) {
	store, err := storage.OpenStorage(":memory:")
	if err != nil {
		t.Fatalf("failed to init storage: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	_ = store.CreateProject(ctx, sqlc.CreateProjectParams{
		ID:         "proj_1",
		Title:      "Test Project",
		SourceLang: "ja",
		TargetLang: "vi",
	})

	updater := NewGraphUpdater(store)
	facts := &ChapterFacts{
		Title:   "Chương 5",
		Summary: "Nayuta đổi cách gọi",
		CastIntros: []CastIntro{
			{Name: "Setsuna", Role: "Editor", Gender: "female"},
		},
		RelationshipChanges: []RelationshipChange{
			{FromChar: "Nayuta", ToChar: "Itsuki", CallAs: "Darling", SelfCallAs: "Nayuta", Tone: "intimate"},
		},
	}

	err = updater.Apply(ctx, "proj_1", 5, facts)
	if err != nil {
		t.Fatalf("expected apply success, got err: %v", err)
	}

	ent, err := store.GetEntityByName(ctx, "proj_1", "Setsuna")
	if err != nil {
		t.Fatalf("expected entity Setsuna to exist: %v", err)
	}
	if ent.Name != "Setsuna" || ent.FirstSeenChapter != 5 {
		t.Errorf("unexpected entity: %+v", ent)
	}

	rel, err := store.GetRelationBetween(context.Background(), "proj_1", "Nayuta", "Itsuki", 5)
	if err != nil {
		t.Fatalf("expected relation to exist: %v", err)
	}
	if rel.CallAs != "Darling" {
		t.Errorf("expected call_as Darling, got: %s", rel.CallAs)
	}
}
