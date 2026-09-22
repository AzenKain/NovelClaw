package services

import (
	"context"
	"fmt"

	"novelclaw/internal/dtos"
	"novelclaw/pkg/export"
	"novelclaw/pkg/storage"
)

// ExportService coordinates ebook creation, packaging, and format inquiries for Wails v3.
type ExportService struct {
	engine *export.Engine
}

// NewExportService creates an ExportService instance backed by storage.
func NewExportService(store *storage.Storage) *ExportService {
	return &ExportService{
		engine: export.NewEngine(store),
	}
}

// ExportEbook synthesizes an ebook file in the requested format.
func (s *ExportService) ExportEbook(ctx context.Context, req dtos.ExportEbookRequest) (*dtos.ExportResultDTO, error) {
	opts := export.SynthesisOptions{
		ProjectID:         req.ProjectID,
		TargetFormat:      req.TargetFormat,
		OutputPath:        req.OutputPath,
		CustomTitle:       req.CustomTitle,
		CustomAuthor:      req.CustomAuthor,
		CoverImagePath:    req.CoverImagePath,
		UseTranslatedOnly: req.UseTranslatedOnly,
		Volume:            req.Volume,
	}

	res, err := s.engine.ExportEbook(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("export ebook: %w", err)
	}

	return &dtos.ExportResultDTO{
		OutputPath:    res.OutputPath,
		Format:        res.Format,
		FileSizeBytes: res.FileSizeBytes,
		ChapterCount:  res.ChapterCount,
		DurationMs:    res.Duration.Milliseconds(),
	}, nil
}

// ExportProjectBundle packages all project database records and assets into a portable .neko file.
func (s *ExportService) ExportProjectBundle(ctx context.Context, req dtos.ExportBundleRequest) (*dtos.ExportResultDTO, error) {
	res, err := s.engine.ExportNekoBundle(ctx, req.ProjectID, req.OutputPath)
	if err != nil {
		return nil, fmt.Errorf("export project bundle: %w", err)
	}

	return &dtos.ExportResultDTO{
		OutputPath:    res.OutputPath,
		Format:        res.Format,
		FileSizeBytes: res.FileSizeBytes,
		ChapterCount:  res.ChapterCount,
		DurationMs:    res.Duration.Milliseconds(),
	}, nil
}

// ListSupportedFormats returns all available ebook export formats with metadata.
func (s *ExportService) ListSupportedFormats() []dtos.FormatDescriptorDTO {
	return []dtos.FormatDescriptorDTO{
		{
			Format:         "epub",
			Name:           "EPUB 3",
			Extension:      ".epub",
			Description:    "Standard modern electronic publication with reflowable text and full image support.",
			RecommendedFor: "Apple Books, Google Play Books, Thorium, generic e-readers",
		},
		{
			Format:         "kepub.epub",
			Name:           "Kobo EPUB (KEPUB)",
			Extension:      ".kepub.epub",
			Description:    "Optimized EPUB format specifically designed for Kobo e-readers with fast paging.",
			RecommendedFor: "Kobo Clara, Kobo Libra, Kobo Sage",
		},
		{
			Format:         "mobi",
			Name:           "Mobipocket / Kindle",
			Extension:      ".mobi",
			Description:    "Legacy Kindle format compatible with all Amazon Kindle devices and apps.",
			RecommendedFor: "Kindle Paperwhite, Kindle Oasis, Kindle Basic",
		},
		{
			Format:         "pdf",
			Name:           "Portable Document Format (PDF)",
			Extension:      ".pdf",
			Description:    "Fixed-layout book format suitable for printing or tablet reading.",
			RecommendedFor: "Desktop reading, tablets, printing",
		},
		{
			Format:         "docx",
			Name:           "Microsoft Word Document",
			Extension:      ".docx",
			Description:    "Editable document format for manual editing, proofreading, and typesetting.",
			RecommendedFor: "Microsoft Word, LibreOffice, Google Docs",
		},
		{
			Format:         "txt",
			Name:           "Plain Text",
			Extension:      ".txt",
			Description:    "Clean raw UTF-8 text representation of the novel chapters.",
			RecommendedFor: "Text editors, lightweight portable viewing",
		},
		{
			Format:         "fb2",
			Name:           "FictionBook 2.0",
			Extension:      ".fb2",
			Description:    "Structured XML-based ebook format popular in Eastern Europe.",
			RecommendedFor: "FBReader, PocketBook, Moon+ Reader",
		},
		{
			Format:         "neko",
			Name:           "NovelClaw Project Bundle",
			Extension:      ".neko",
			Description:    "Complete project package containing translation database, character graph, and glossary.",
			RecommendedFor: "NovelClaw backup and sharing",
		},
	}
}
