package services

import (
	"context"
	"path/filepath"
	"testing"

	"novelclaw/internal/gen/sqlc"
	"novelclaw/pkg/soul"
	"novelclaw/pkg/storage"
)

func TestTranslationService_ResolveAskHuman(t *testing.T) {
	svc := NewTranslationServiceWithSoul(nil, soul.NewControllerWithDir(t.TempDir()))

	// 1. Resolve on non-existent job ID should return false
	if ok := svc.ResolveAskHuman("non_existent_job", "some_answer"); ok {
		t.Error("expected ResolveAskHuman on non-existent job to return false")
	}

	// 2. Set up a pending prompt channel
	jobID := "ask_test_job_123"
	ch := make(chan string, 1)
	svc.pendingPrompts.Store(jobID, ch)

	// 3. Resolve with answer
	expectedAnswer := "Arthur"
	if ok := svc.ResolveAskHuman(jobID, expectedAnswer); !ok {
		t.Fatal("expected ResolveAskHuman to succeed")
	}

	// 4. Verify channel received the answer
	select {
	case ans := <-ch:
		if ans != expectedAnswer {
			t.Errorf("expected received answer '%s', got '%s'", expectedAnswer, ans)
		}
	default:
		t.Error("channel did not receive answer")
	}

	// 5. Verify pendingPrompt was deleted
	if _, ok := svc.pendingPrompts.Load(jobID); ok {
		t.Error("expected pending prompt to be deleted after resolve")
	}

	// 6. Calling again should return false (already resolved/deleted)
	if ok := svc.ResolveAskHuman(jobID, "another_answer"); ok {
		t.Error("expected second resolve on same job to return false")
	}
}

func TestTranslationService_AutoLearnResolution(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "auto_learn_test.db")

	store, err := storage.OpenStorage(dbPath)
	if err != nil {
		t.Fatalf("OpenStorage failed: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	projectID := "proj_hitl_learn"

	err = store.CreateProject(ctx, sqlc.CreateProjectParams{
		ID:         projectID,
		Title:      "Auto Learn Novel",
		SourceLang: "ja",
		TargetLang: "vi",
	})
	if err != nil {
		t.Fatalf("CreateProject failed: %v", err)
	}

	svc := NewTranslationServiceWithSoul(store, soul.NewControllerWithDir(t.TempDir()))

	// Test 1: Single term quoted -> should create Glossary term
	termQuestion := `Nên dịch tên "アーサー" thành gì?`
	termAnswer := "Arthur"
	svc.autoLearnResolution(ctx, projectID, 3, termQuestion, termAnswer, []string{"Arthur", "Asa"})

	terms, err := store.ListGlossaryByProject(ctx, projectID)
	if err != nil {
		t.Fatalf("ListGlossaryByProject failed: %v", err)
	}
	if len(terms) == 0 {
		t.Fatalf("expected at least 1 glossary term learned, got 0")
	}

	var foundTerm *sqlc.Glossary
	for i := range terms {
		if terms[i].SourceTerm == "アーサー" {
			foundTerm = &terms[i]
			break
		}
	}
	if foundTerm == nil {
		t.Fatalf("expected glossary term for 'アーサー' not found")
	}
	if foundTerm.TargetTerm != "Arthur" {
		t.Errorf("expected TargetTerm 'Arthur', got '%s'", foundTerm.TargetTerm)
	}
	if foundTerm.Category != "hitl_decision" {
		t.Errorf("expected Category 'hitl_decision', got '%s'", foundTerm.Category)
	}

	// Test 2: Two terms quoted -> should create Character Relation
	relQuestion := `"Tiêu Viêm" gọi "Dược Lão" là gì trong phân cảnh này?`
	relAnswer := "Sư phụ"
	svc.autoLearnResolution(ctx, projectID, 4, relQuestion, relAnswer, []string{"Sư phụ", "Lão đầu"})

	relations, err := store.ListRelationsByProject(ctx, projectID)
	if err != nil {
		t.Fatalf("ListRelationsByProject failed: %v", err)
	}
	if len(relations) == 0 {
		t.Fatalf("expected at least 1 relation learned, got 0")
	}

	var foundRel *sqlc.CharacterRelation
	for i := range relations {
		if relations[i].FromChar == "Tiêu Viêm" && relations[i].ToChar == "Dược Lão" {
			foundRel = &relations[i]
			break
		}
	}
	if foundRel == nil {
		t.Fatalf("expected relation from 'Tiêu Viêm' to 'Dược Lão' not found")
	}
	if foundRel.CallAs != "Sư phụ" {
		t.Errorf("expected CallAs 'Sư phụ', got '%s'", foundRel.CallAs)
	}
	if foundRel.SinceChapter != 4 {
		t.Errorf("expected SinceChapter 4, got %d", foundRel.SinceChapter)
	}
}

func TestTranslationService_SoftStop(t *testing.T) {
	soulCtrl := soul.NewControllerWithDir(t.TempDir())
	svc := NewTranslationServiceWithSoul(nil, soulCtrl)

	jobID := "job_test_softstop"
	cancelled := false
	cancelFn := func() { cancelled = true }
	soulCtrl.RegisterJob(jobID, "proj_1", 1, cancelFn)

	if soulCtrl.IsSoftStopRequested(jobID) {
		t.Error("expected soft stop to initially be false")
	}

	ok := svc.SoftStopTranslation(jobID)
	if !ok {
		t.Fatal("expected SoftStopTranslation to succeed")
	}

	if !soulCtrl.IsSoftStopRequested(jobID) {
		t.Error("expected IsSoftStopRequested to be true after SoftStopTranslation")
	}

	if cancelled {
		t.Error("expected hard cancel NOT to be called on soft stop")
	}
}
