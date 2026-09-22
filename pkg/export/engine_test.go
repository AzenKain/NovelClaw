package export_test

import (
	"archive/zip"
	"context"
	"os"
	"path/filepath"
	"testing"

	"novelclaw/internal/gen/sqlc"
	"novelclaw/pkg/export"
	"novelclaw/pkg/storage"
)

func setupTestStorage(t *testing.T) (*storage.Storage, string, func()) {
	t.Helper()
	tempDir, err := os.MkdirTemp("", "export_test_*")
	if err != nil {
		t.Fatalf("create temp dir: %v", err)
	}

	dbPath := filepath.Join(tempDir, "test.db")
	store, err := storage.OpenStorage(dbPath)
	if err != nil {
		t.Fatalf("open storage: %v", err)
	}

	cleanup := func() {
		_ = store.Close()
		_ = os.RemoveAll(tempDir)
	}

	return store, tempDir, cleanup
}

func seedTestData(t *testing.T, store *storage.Storage, projectID string) {
	t.Helper()
	ctx := context.Background()

	err := store.CreateProject(ctx, sqlc.CreateProjectParams{
		ID:             projectID,
		Title:          "Chàng Kiếm Sĩ Trọng Sinh",
		Author:         "Akira",
		SourceLang:     "ja",
		TargetLang:     "vi",
		OriginalFormat: "epub",
		TotalChapters:  2,
	})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}

	ch1 := sqlc.CreateChapterParams{
		ID:           projectID + "_c1",
		ProjectID:    projectID,
		ChapterIndex: 1,
		Title:        "Chương 1: Khởi đầu mới",
		ContentPath:  "Text/c1.xhtml",
		RawContent:   "転生した剣士は新しい世界で目覚めた。",
	}
	if err := store.CreateChapter(ctx, ch1); err != nil {
		t.Fatalf("create ch1: %v", err)
	}
	if err := store.UpdateChapterTranslation(ctx, ch1.ID, "Kiếm sĩ chuyển sinh tỉnh lại ở thế giới mới với thanh kiếm rực sáng.", "completed"); err != nil {
		t.Fatalf("update ch1: %v", err)
	}

	ch2 := sqlc.CreateChapterParams{
		ID:           projectID + "_c2",
		ProjectID:    projectID,
		ChapterIndex: 2,
		Title:        "Chương 2: Gặp gỡ đồng đội",
		ContentPath:  "Text/c2.xhtml",
		RawContent:   "森の中で少女と出会った。",
	}
	if err := store.CreateChapter(ctx, ch2); err != nil {
		t.Fatalf("create ch2: %v", err)
	}
	if err := store.UpdateChapterTranslation(ctx, ch2.ID, "Trong khu rừng sâu thẳm, chàng gặp gỡ một thiếu nữ pháp sư.", "completed"); err != nil {
		t.Fatalf("update ch2: %v", err)
	}

	_ = store.UpsertEntity(ctx, storage.Entity{
		ID:               projectID + "_e1",
		ProjectID:        projectID,
		Name:             "Kiếm Sĩ",
		Aliases:          []string{"Akira", "Kenshi"},
		Category:         "character",
		Gender:           "male",
		Role:             "protagonist",
		FirstSeenChapter: 1,
	})

	_ = store.UpsertRelation(ctx, sqlc.UpsertRelationParams{
		ID:           projectID + "_r1",
		ProjectID:    projectID,
		FromChar:     "Pháp Sư",
		ToChar:       "Kiếm Sĩ",
		CallAs:       "Tiền bối",
		SelfCallAs:   "Em",
		SinceChapter: 2,
		Tone:         "kính trọng",
		IsLocked:     1,
	})

	_ = store.UpsertGlossaryTerm(ctx, sqlc.UpsertGlossaryTermParams{
		ID:         projectID + "_g1",
		ProjectID:  projectID,
		SourceTerm: "魔力",
		TargetTerm: "Ma lực",
		Category:   "magic",
		Notes:      "Năng lượng phép thuật",
	})
}

func TestEngine_ExportEbook(t *testing.T) {
	store, tempDir, cleanup := setupTestStorage(t)
	defer cleanup()

	projectID := "proj_export_1"
	seedTestData(t, store, projectID)

	engine := export.NewEngine(store)
	ctx := context.Background()

	formats := []string{"epub", "kepub.epub", "txt", "docx", "mobi", "fb2"}

	for _, fmtName := range formats {
		t.Run("Format_"+fmtName, func(t *testing.T) {
			outPath := filepath.Join(tempDir, "output."+fmtName)
			res, err := engine.ExportEbook(ctx, export.SynthesisOptions{
				ProjectID:    projectID,
				TargetFormat: fmtName,
				OutputPath:   outPath,
			})
			if err != nil {
				t.Fatalf("ExportEbook failed for %s: %v", fmtName, err)
			}
			if res.FileSizeBytes <= 0 {
				t.Errorf("expected positive file size, got %d", res.FileSizeBytes)
			}
			if res.ChapterCount != 2 {
				t.Errorf("expected 2 chapters, got %d", res.ChapterCount)
			}
			if _, err := os.Stat(outPath); err != nil {
				t.Fatalf("file does not exist on disk: %v", err)
			}
		})
	}
}

func TestEngine_ExportNekoBundle(t *testing.T) {
	store, tempDir, cleanup := setupTestStorage(t)
	defer cleanup()

	projectID := "proj_bundle_1"
	seedTestData(t, store, projectID)

	engine := export.NewEngine(store)
	ctx := context.Background()

	bundlePath := filepath.Join(tempDir, "project.neko")
	res, err := engine.ExportNekoBundle(ctx, projectID, bundlePath)
	if err != nil {
		t.Fatalf("ExportNekoBundle failed: %v", err)
	}

	if res.FileSizeBytes <= 0 {
		t.Errorf("expected positive bundle size, got %d", res.FileSizeBytes)
	}

	zr, err := zip.OpenReader(bundlePath)
	if err != nil {
		t.Fatalf("open zip reader: %v", err)
	}
	defer zr.Close()

	expectedEntries := map[string]bool{
		"manifest.json":  false,
		"project.json":   false,
		"chapters.json":  false,
		"relations.json": false,
		"glossary.json":  false,
		"entities.json":  false,
	}

	for _, f := range zr.File {
		if _, ok := expectedEntries[f.Name]; ok {
			expectedEntries[f.Name] = true
		}
	}

	for name, found := range expectedEntries {
		if !found {
			t.Errorf("missing bundle entry: %s", name)
		}
	}
}
