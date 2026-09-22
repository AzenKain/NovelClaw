package services

import (
	"bytes"
	"context"
	"strings"

	"novelclaw/internal/dtos"
	"novelclaw/pkg/glossary"
	"novelclaw/pkg/storage"
)

// GlossaryService exposes terminology and dictionary operations to Wails v3 frontend.
type GlossaryService struct {
	store      *storage.Storage
	hub        *glossary.Hub
	downloader *glossary.RemoteDownloader
}

// NewGlossaryService creates a new GlossaryService.
func NewGlossaryService(store *storage.Storage) *GlossaryService {
	hub := glossary.NewHub(store)
	return &GlossaryService{
		store:      store,
		hub:        hub,
		downloader: glossary.NewRemoteDownloader(hub),
	}
}

// UpsertTerm inserts or updates a glossary term from a DTO request.
func (s *GlossaryService) UpsertTerm(ctx context.Context, req dtos.UpsertGlossaryTermRequest) error {
	return s.store.UpsertGlossaryTerm(ctx, req.ToUpsertParams())
}

// ListTerms lists all terms in a project's glossary as DTOs.
func (s *GlossaryService) ListTerms(ctx context.Context, projectID string) ([]dtos.GlossaryTermDTO, error) {
	terms, err := s.store.ListGlossaryByProject(ctx, projectID)
	if err != nil {
		return nil, err
	}
	dtosList := make([]dtos.GlossaryTermDTO, 0, len(terms))
	for _, t := range terms {
		dtosList = append(dtosList, dtos.ToGlossaryTermDTO(t))
	}
	return dtosList, nil
}

// DeleteTerm deletes a glossary term by ID.
func (s *GlossaryService) DeleteTerm(ctx context.Context, id string) error {
	return s.store.DeleteGlossaryTerm(ctx, id)
}

// ImportTSV imports glossary items from a TSV string.
func (s *GlossaryService) ImportTSV(ctx context.Context, projectID string, tsvContent string) (int, error) {
	return s.hub.ImportTSV(ctx, projectID, strings.NewReader(tsvContent))
}

// ExportTSV exports glossary items as a TSV string.
func (s *GlossaryService) ExportTSV(ctx context.Context, projectID string) (string, error) {
	var buf bytes.Buffer
	if err := s.hub.ExportTSV(ctx, projectID, &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// ImportJSON imports glossary items from a JSON string.
func (s *GlossaryService) ImportJSON(ctx context.Context, projectID string, jsonContent string) (int, error) {
	return s.hub.ImportJSON(ctx, projectID, []byte(jsonContent))
}

// ExportJSON exports glossary items as a JSON string.
func (s *GlossaryService) ExportJSON(ctx context.Context, projectID string) (string, error) {
	data, err := s.hub.ExportJSON(ctx, projectID)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ImportPlainText imports glossary items from standard 'source=target' lines (Vietphrase / QuickTranslator format).
func (s *GlossaryService) ImportPlainText(ctx context.Context, projectID string, content string, defaultCategory string) (int, error) {
	return s.hub.ImportPlainText(ctx, projectID, content, defaultCategory)
}

// ListRemoteSources returns all community online dictionary sources (e.g. GitHub repos) for frontend display.
func (s *GlossaryService) ListRemoteSources(ctx context.Context) ([]dtos.RemoteGlossarySourceDTO, error) {
	defs := glossary.GetCommunityRemoteSources()
	res := make([]dtos.RemoteGlossarySourceDTO, len(defs))
	for i, d := range defs {
		res[i] = dtos.RemoteGlossarySourceDTO{
			ID:              d.ID,
			Title:           d.Title,
			Genre:           d.Genre,
			Description:     d.Description,
			URL:             d.URL,
			Format:          d.Format,
			DefaultCategory: d.DefaultCategory,
		}
	}
	return res, nil
}

// DownloadAndImport fetches dictionary content from a remote URL (GitHub raw, Pastebin, custom link) and imports it into the project.
func (s *GlossaryService) DownloadAndImport(ctx context.Context, req dtos.DownloadGlossaryRequest) (int, error) {
	return s.downloader.DownloadAndImport(ctx, req.ProjectID, req.URL, req.Format, req.DefaultCategory)
}
