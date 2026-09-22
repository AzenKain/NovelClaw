package glossary

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"novelclaw/internal/gen/sqlc"
	"novelclaw/pkg/storage"
)

func TestHub_TSVImportExport(t *testing.T) {
	store, err := storage.OpenStorage(":memory:")
	if err != nil {
		t.Fatalf("open storage: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	projectID := "proj_glossary_test"

	_ = store.CreateProject(ctx, sqlc.CreateProjectParams{
		ID:         projectID,
		Title:      "Glossary Test",
		SourceLang: "ja",
		TargetLang: "vi",
	})

	hub := NewHub(store)

	tsvData := `source_term	target_term	category	notes
白木屋	Shirakiya	location	Chuỗi quán nhậu
TRPG	Tabletop RPG	game	Trò chơi nhập vai
妹バカ	Cuồng em gái	slang	Tính từ chỉ Itsuki`

	count, err := hub.ImportTSV(ctx, projectID, strings.NewReader(tsvData))
	if err != nil {
		t.Fatalf("ImportTSV failed: %v", err)
	}
	if count != 3 {
		t.Fatalf("expected 3 items imported, got %d", count)
	}

	var buf bytes.Buffer
	err = hub.ExportTSV(ctx, projectID, &buf)
	if err != nil {
		t.Fatalf("ExportTSV failed: %v", err)
	}

	exportedStr := buf.String()
	if !strings.Contains(exportedStr, "Shirakiya") || !strings.Contains(exportedStr, "TRPG") {
		t.Errorf("ExportTSV missing expected items: %s", exportedStr)
	}
}

func TestHub_JSONImportExport(t *testing.T) {
	store, err := storage.OpenStorage(":memory:")
	if err != nil {
		t.Fatalf("open storage: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	projectID := "proj_json_test"

	_ = store.CreateProject(ctx, sqlc.CreateProjectParams{
		ID:         projectID,
		Title:      "Glossary JSON Test",
		SourceLang: "ja",
		TargetLang: "vi",
	})

	hub := NewHub(store)

	jsonData := []byte(`[
		{"source_term": "可児 那由多", "target_term": "Kani Nayuta", "category": "character", "notes": "Tác giả thiên tài"},
		{"source_term": "羽島 千尋", "target_term": "Hashima Chihiro", "category": "character", "notes": "Em trai/gái"}
	]`)

	count, err := hub.ImportJSON(ctx, projectID, jsonData)
	if err != nil {
		t.Fatalf("ImportJSON failed: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected 2 items imported, got %d", count)
	}

	outBytes, err := hub.ExportJSON(ctx, projectID)
	if err != nil {
		t.Fatalf("ExportJSON failed: %v", err)
	}

	outStr := string(outBytes)
	if !strings.Contains(outStr, "Kani Nayuta") || !strings.Contains(outStr, "Hashima Chihiro") {
		t.Errorf("ExportJSON missing items: %s", outStr)
	}
}

func TestHub_ImportPlainText_Vietphrase(t *testing.T) {
	store, err := storage.OpenStorage(":memory:")
	if err != nil {
		t.Fatalf("open storage: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	projectID := "proj_vietphrase_test"

	_ = store.CreateProject(ctx, sqlc.CreateProjectParams{
		ID:         projectID,
		Title:      "Vietphrase Test",
		SourceLang: "zh",
		TargetLang: "vi",
	})

	hub := NewHub(store)

	plainText := `
# Comment line
// Another comment
筑基=Trúc Cơ
金丹=Kim Đan
Arasaka	Tập đoàn Arasaka
Excalibur: Thánh Kiếm Excalibur
`

	count, err := hub.ImportPlainText(ctx, projectID, plainText, "rank")
	if err != nil {
		t.Fatalf("ImportPlainText failed: %v", err)
	}
	if count != 4 {
		t.Fatalf("expected 4 items imported, got %d", count)
	}

	terms, err := store.ListGlossaryByProject(ctx, projectID)
	if err != nil {
		t.Fatalf("list glossary: %v", err)
	}
	if len(terms) != 4 {
		t.Fatalf("expected 4 terms in DB, got %d", len(terms))
	}
}

func TestGetCommunityRemoteSources(t *testing.T) {
	sources := GetCommunityRemoteSources()
	for _, s := range sources {
		if s.ID == "" || s.Title == "" || s.URL == "" {
			t.Errorf("invalid source definition: %+v", s)
		}
		if !strings.HasPrefix(s.URL, "https://") {
			t.Errorf("source URL must start with https://: %s", s.URL)
		}
	}
}

func TestRemoteDownloader_DownloadAndImport(t *testing.T) {
	store, err := storage.OpenStorage(":memory:")
	if err != nil {
		t.Fatalf("open storage: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	projectID := "proj_remote_dl_test"

	_ = store.CreateProject(ctx, sqlc.CreateProjectParams{
		ID:         projectID,
		Title:      "Remote Download Test",
		SourceLang: "zh",
		TargetLang: "vi",
	})

	hub := NewHub(store)
	downloader := NewRemoteDownloader(hub)

	// Spin up a mock HTTP server serving different dictionary formats
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/names.txt":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("萧炎=Tiêu Viêm\n药老=Dược Lão\n美杜莎=Mỹ Đỗ Toa\n"))
		case "/terms.tsv":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("source_term\ttarget_term\tcategory\tnotes\nSandevistan\tSandevistan\titem\tCyberware\n"))
		case "/terms.json":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`[{"source_term":"Elf","target_term":"Tinh Linh","category":"faction"}]`))
		case "/404":
			w.WriteHeader(http.StatusNotFound)
		default:
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
	defer server.Close()

	downloader.SetHTTPClient(server.Client())

	// Test 1: Download and import Vietphrase plain text
	count, err := downloader.DownloadAndImport(ctx, projectID, server.URL+"/names.txt", "vietphrase", "proper_name")
	if err != nil {
		t.Fatalf("DownloadAndImport plain text failed: %v", err)
	}
	if count != 3 {
		t.Fatalf("expected 3 items, got %d", count)
	}

	// Test 2: Download and import TSV
	countTSV, err := downloader.DownloadAndImport(ctx, projectID, server.URL+"/terms.tsv", "tsv", "item")
	if err != nil {
		t.Fatalf("DownloadAndImport tsv failed: %v", err)
	}
	if countTSV != 1 {
		t.Fatalf("expected 1 TSV item, got %d", countTSV)
	}

	// Test 3: Download and import JSON
	countJSON, err := downloader.DownloadAndImport(ctx, projectID, server.URL+"/terms.json", "json", "faction")
	if err != nil {
		t.Fatalf("DownloadAndImport json failed: %v", err)
	}
	if countJSON != 1 {
		t.Fatalf("expected 1 JSON item, got %d", countJSON)
	}

	// Test 4: Error handling on HTTP 404
	_, err = downloader.DownloadAndImport(ctx, projectID, server.URL+"/404", "vietphrase", "proper_name")
	if err == nil {
		t.Fatal("expected error on HTTP 404, got nil")
	}

	// Verify DB state
	terms, err := store.ListGlossaryByProject(ctx, projectID)
	if err != nil {
		t.Fatalf("list glossary: %v", err)
	}
	if len(terms) != 5 {
		t.Fatalf("expected 5 total terms in DB, got %d", len(terms))
	}
}
