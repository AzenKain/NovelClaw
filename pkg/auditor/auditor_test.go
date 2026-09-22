package auditor

import (
	"context"
	"path/filepath"
	"testing"

	"novelclaw/internal/gen/sqlc"
	"novelclaw/pkg/storage"
)

func TestPlotAuditor_DeceasedReappearance(t *testing.T) {
	ctx := context.Background()
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_auditor.db")

	store, err := storage.OpenStorage(dbPath)
	if err != nil {
		t.Fatalf("OpenStorage failed: %v", err)
	}
	defer store.Close()

	projectID := "proj-audit-test"
	_ = store.CreateProject(ctx, sqlc.CreateProjectParams{
		ID:             projectID,
		Title:          "Audit Novel",
		Author:         "Author",
		SourceLang:     "zh",
		TargetLang:     "vi",
		OriginalFormat: "txt",
	})

	// Add a deceased entity
	deadChar := storage.Entity{
		ID:               "ent-nguyen-ba",
		ProjectID:        projectID,
		Name:             "Nguyên Bá",
		Category:         "character",
		FirstSeenChapter: 1,
		Metadata: map[string]any{
			"status": "đã chết",
			"tier":   "Trúc Cơ",
		},
	}
	if err := store.UpsertEntity(ctx, deadChar); err != nil {
		t.Fatalf("UpsertEntity failed: %v", err)
	}

	auditor := NewPlotAuditor(store)

	// Case 1: Deceased character acts in chapter 10
	badChapterText := "Trong bóng đêm mịt mù, Nguyên Bá bước ra với nụ cười lạnh lùng trên môi."
	warnings, err := auditor.AuditChapter(ctx, projectID, 10, badChapterText)
	if err != nil {
		t.Fatalf("AuditChapter failed: %v", err)
	}
	if len(warnings) != 1 {
		t.Fatalf("expected 1 warning for deceased character reappearance, got %d", len(warnings))
	}
	if warnings[0].Type != AnomalyDeceasedReappearance {
		t.Errorf("expected AnomalyDeceasedReappearance, got %s", warnings[0].Type)
	}

	// Case 2: Only passive mention in flashback or memorial, no active action verb
	passiveText := "Mọi người đều nhớ đến sự hy sinh anh dũng của Nguyên Bá năm xưa."
	warningsPassive, err := auditor.AuditChapter(ctx, projectID, 11, passiveText)
	if err != nil {
		t.Fatalf("AuditChapter failed: %v", err)
	}
	if len(warningsPassive) != 0 {
		t.Errorf("expected 0 warnings for passive mention, got %d", len(warningsPassive))
	}
}

func TestPlotAuditor_RealmRegressionAndFactionMismatch(t *testing.T) {
	ctx := context.Background()
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_auditor_extra.db")

	store, err := storage.OpenStorage(dbPath)
	if err != nil {
		t.Fatalf("OpenStorage failed: %v", err)
	}
	defer store.Close()

	projectID := "proj-audit-extra"
	_ = store.CreateProject(ctx, sqlc.CreateProjectParams{
		ID:             projectID,
		Title:          "Audit Novel 2",
		Author:         "Author",
		SourceLang:     "zh",
		TargetLang:     "vi",
		OriginalFormat: "txt",
	})

	// Add a high-realm character with opposing faction
	grandmaster := storage.Entity{
		ID:               "ent-tieu-viem",
		ProjectID:        projectID,
		Name:             "Tiêu Viêm",
		Category:         "character",
		FirstSeenChapter: 1,
		Metadata: map[string]any{
			"status":          "alive",
			"tier":            "Hóa Thần",
			"opposing_faction": "Hồn Điện",
		},
	}
	if err := store.UpsertEntity(ctx, grandmaster); err != nil {
		t.Fatalf("UpsertEntity failed: %v", err)
	}

	auditor := NewPlotAuditor(store)

	// Test Realm Regression
	regressText := "Mọi người đều khinh bỉ vì Tiêu Viêm chỉ là Trúc Cơ mà thôi."
	warnings, err := auditor.AuditChapter(ctx, projectID, 5, regressText)
	if err != nil {
		t.Fatalf("AuditChapter failed: %v", err)
	}
	if len(warnings) != 1 || warnings[0].Type != AnomalyRealmRegression {
		t.Fatalf("expected 1 AnomalyRealmRegression warning, got %+v", warnings)
	}

	// Test Faction Mismatch
	mismatchText := "Không ai ngờ được rằng Tiêu Viêm gia nhập Hồn Điện từ lâu."
	warnings2, err := auditor.AuditChapter(ctx, projectID, 6, mismatchText)
	if err != nil {
		t.Fatalf("AuditChapter failed: %v", err)
	}
	if len(warnings2) != 1 || warnings2[0].Type != AnomalyFactionMismatch {
		t.Fatalf("expected 1 AnomalyFactionMismatch warning, got %+v", warnings2)
	}
}

