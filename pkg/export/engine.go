package export

import (
	"archive/zip"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	"novelclaw/pkg/bookparser"
	"novelclaw/pkg/bookparser/defaultcover"
	"novelclaw/pkg/bookparser/epub"
	"novelclaw/pkg/ebookconv"
	"novelclaw/internal/gen/sqlc"
	"novelclaw/pkg/storage"
)

// SupportedExportFormats lists all export formats supported by the synthesis engine.
var SupportedExportFormats = []string{
	"epub",
	"kepub.epub",
	"mobi",
	"azw",
	"pdf",
	"docx",
	"txt",
	"fb2",
	"neko",
}

// SynthesisOptions configures the ebook generation process.
type SynthesisOptions struct {
	ProjectID        string
	TargetFormat     string
	OutputPath       string
	CustomTitle      string
	CustomAuthor     string
	SourceFilePath   string
	CoverImagePath   string
	UseTranslatedOnly bool
	Volume           string
}

// SynthesisResult reports the outcome of an ebook export operation.
type SynthesisResult struct {
	OutputPath    string
	Format        string
	FileSizeBytes int64
	ChapterCount  int
	Duration      time.Duration
}

// NekoBundleManifest defines the top-level metadata stored inside a .neko project archive.
type NekoBundleManifest struct {
	Version      string    `json:"version"`
	ProjectID    string    `json:"project_id"`
	Title        string    `json:"title"`
	Author       string    `json:"author"`
	SourceLang   string    `json:"source_lang"`
	TargetLang   string    `json:"target_lang"`
	TotalChapters int      `json:"total_chapters"`
	ExportedAt   time.Time `json:"exported_at"`
}

// Engine coordinates ebook synthesis and project bundle packaging.
type Engine struct {
	store    *storage.Storage
	registry bookparser.Registry
}

// NewEngine creates a new export Engine backed by SQLite storage.
func NewEngine(store *storage.Storage) *Engine {
	reg := bookparser.NewRegistry()
	reg.Register(epub.NewParser(), "epub")
	return &Engine{
		store:    store,
		registry: reg,
	}
}

// ExportEbook builds an ebook in the specified format from project chapters and metadata.
func (e *Engine) ExportEbook(ctx context.Context, opts SynthesisOptions) (*SynthesisResult, error) {
	startTime := time.Now()

	project, err := e.store.GetProject(ctx, opts.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("load project %s: %w", opts.ProjectID, err)
	}

	chapters, err := e.store.ListChaptersByProject(ctx, opts.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("list chapters for project %s: %w", opts.ProjectID, err)
	}
	if len(chapters) == 0 {
		return nil, fmt.Errorf("project %s has no chapters to export", opts.ProjectID)
	}

	// Filter chapters by volume if specified (e.g. "vol1")
	if opts.Volume != "" && opts.Volume != "all" {
		volTag := fmt.Sprintf("[%s]", strings.ToLower(opts.Volume))
		var filtered []sqlc.Chapter
		for _, ch := range chapters {
			if strings.HasPrefix(strings.ToLower(ch.Title), volTag) {
				filtered = append(filtered, ch)
			}
		}
		if len(filtered) > 0 {
			chapters = filtered
		}
	}

	title := project.Title
	if strings.TrimSpace(opts.CustomTitle) != "" {
		title = strings.TrimSpace(opts.CustomTitle)
	}

	author := project.Author
	if strings.TrimSpace(opts.CustomAuthor) != "" {
		author = strings.TrimSpace(opts.CustomAuthor)
	}

	targetLang := project.TargetLang
	if targetLang == "" {
		targetLang = "en"
	}

	var bookChapters []bookparser.ChapterData
	for i, ch := range chapters {
		content := ch.TranslatedContent
		if strings.TrimSpace(content) == "" {
			if opts.UseTranslatedOnly {
				continue
			}
			content = ch.RawContent
		}

		if !strings.Contains(content, "<p>") && !strings.Contains(content, "<html") {
			content = formatTextToHTML(ch.Title, content)
		}

		cleanTitle := ch.Title
		if opts.Volume != "" && opts.Volume != "all" {
			if strings.HasPrefix(cleanTitle, "[") {
				if idx := strings.Index(cleanTitle, "]"); idx != -1 {
					cleanTitle = strings.TrimSpace(cleanTitle[idx+1:])
				}
			}
		}

		bookChapters = append(bookChapters, bookparser.ChapterData{
			Title:       cleanTitle,
			Content:     content,
			ContentPath: fmt.Sprintf("Text/chapter_%d.xhtml", ch.ChapterIndex),
			Index:       i + 1,
		})
	}

	if len(bookChapters) == 0 {
		return nil, fmt.Errorf("no chapters available to export (all chapters were skipped because they are not translated yet)")
	}

	var images []ebookconv.Image
	var coverData []byte
	var coverType string

	if opts.CoverImagePath != "" {
		if data, err := os.ReadFile(opts.CoverImagePath); err == nil && len(data) > 0 {
			coverData = data
			coverType = "image/png"
			if strings.HasSuffix(strings.ToLower(opts.CoverImagePath), ".jpg") || strings.HasSuffix(strings.ToLower(opts.CoverImagePath), ".jpeg") {
				coverType = "image/jpeg"
			}
			images = append(images, ebookconv.Image{
				Src:  opts.CoverImagePath,
				Name: "cover.png",
				Data: coverData,
			})
		}
	}

	// Check project asset directories for illustrations and cover images
	assetDirs := []string{
		filepath.Join("data", "assets", opts.ProjectID),
	}
	if opts.Volume != "" && opts.Volume != "all" {
		assetDirs = append([]string{filepath.Join("data", "assets", opts.ProjectID, opts.Volume)}, assetDirs...)
	}

	for _, aDir := range assetDirs {
		if fi, err := os.Stat(aDir); err == nil && fi.IsDir() {
			_ = filepath.Walk(aDir, func(p string, info os.FileInfo, err error) error {
				if err != nil || info.IsDir() {
					return nil
				}
				ext := strings.ToLower(filepath.Ext(p))
				if ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".webp" || ext == ".gif" {
					if data, err := os.ReadFile(p); err == nil {
						rel, _ := filepath.Rel(aDir, p)
						images = append(images, ebookconv.Image{
							Src:  rel,
							Name: filepath.Base(p),
							Data: data,
						})
						if coverData == nil && (strings.Contains(strings.ToLower(p), "cover") || strings.Contains(strings.ToLower(p), "titlepage") || strings.Contains(strings.ToLower(p), "frontcover") || strings.Contains(strings.ToLower(p), "bìa")) {
							coverData = data
							coverType = "image/jpeg"
							if ext == ".png" {
								coverType = "image/png"
							}
						}
					}
				}
				return nil
			})
		}
	}

	if opts.SourceFilePath != "" && e.registry != nil {
		if parser, err := e.registry.Parser("", opts.SourceFilePath); err == nil {
			collected := ebookconv.CollectImages(parser, opts.SourceFilePath)
			for _, img := range collected {
				images = append(images, img)
				if coverData == nil && (strings.Contains(strings.ToLower(img.Src), "cover") || strings.Contains(strings.ToLower(img.Name), "cover")) {
					coverData = img.Data
					coverType = "image/png"
				}
			}
		}
	}

	if coverData == nil {
		coverData = defaultcover.GenerateSVG(title, author)
		coverType = "image/svg+xml"
	}

	book := &bookparser.BookData{
		Metadata: bookparser.BookMetadata{
			Title:       title,
			Author:      author,
			Language:    targetLang,
			Date:        time.Now().Format("2006-01-02"),
			CoverData:   coverData,
			CoverType:   coverType,
			Publisher:   "NovelClaw",
			Subjects:    []string{"Light Novel", "Translated"},
			Description: fmt.Sprintf("Translated with NovelClaw - %s", title),
		},
		Chapters: bookChapters,
	}

	normFormat := bookparser.NormalizeFormat(opts.TargetFormat)
	if normFormat == "kepub" {
		normFormat = "kepub.epub"
	}

	data, err := ebookconv.WriteEbook(book, images, normFormat)
	if err != nil {
		return nil, fmt.Errorf("synthesize ebook (%s): %w", normFormat, err)
	}

	outPath := opts.OutputPath
	if outPath == "" {
		ext := normFormat
		if ext == "kepub.epub" {
			ext = "kepub.epub"
		}
		safeTitle := sanitizeFilename(title)
		outPath = fmt.Sprintf("%s.%s", safeTitle, ext)
	}

	if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil && filepath.Dir(outPath) != "." {
		return nil, fmt.Errorf("create output directory: %w", err)
	}

	if err := os.WriteFile(outPath, data, 0644); err != nil {
		return nil, fmt.Errorf("write ebook file %s: %w", outPath, err)
	}

	log.Info().
		Str("project_id", opts.ProjectID).
		Str("format", normFormat).
		Str("path", outPath).
		Int64("size_bytes", int64(len(data))).
		Msg("successfully exported ebook")

	return &SynthesisResult{
		OutputPath:    outPath,
		Format:        normFormat,
		FileSizeBytes: int64(len(data)),
		ChapterCount:  len(bookChapters),
		Duration:      time.Since(startTime),
	}, nil
}

// ExportNekoBundle packages the complete project database tables and assets into a portable .neko ZIP archive.
func (e *Engine) ExportNekoBundle(ctx context.Context, projectID string, outputPath string) (*SynthesisResult, error) {
	startTime := time.Now()

	project, err := e.store.GetProject(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("load project %s: %w", projectID, err)
	}

	chapters, err := e.store.ListChaptersByProject(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("list chapters for project %s: %w", projectID, err)
	}

	relations, err := e.store.ListRelationsByProject(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("list relations for project %s: %w", projectID, err)
	}

	glossaries, err := e.store.ListGlossaryByProject(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("list glossary for project %s: %w", projectID, err)
	}

	entities, err := e.store.ListEntitiesByProject(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("list entities for project %s: %w", projectID, err)
	}

	if outputPath == "" {
		safeTitle := sanitizeFilename(project.Title)
		outputPath = fmt.Sprintf("%s.neko", safeTitle)
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil && filepath.Dir(outputPath) != "." {
		return nil, fmt.Errorf("create bundle directory: %w", err)
	}

	bundleFile, err := os.Create(outputPath)
	if err != nil {
		return nil, fmt.Errorf("create bundle file %s: %w", outputPath, err)
	}
	defer bundleFile.Close()

	zw := zip.NewWriter(bundleFile)
	defer zw.Close()

	manifest := NekoBundleManifest{
		Version:      "1.0.0",
		ProjectID:    project.ID,
		Title:        project.Title,
		Author:       project.Author,
		SourceLang:   project.SourceLang,
		TargetLang:   project.TargetLang,
		TotalChapters: len(chapters),
		ExportedAt:   time.Now(),
	}

	if err := writeJSONToZip(zw, "manifest.json", manifest); err != nil {
		return nil, err
	}
	if err := writeJSONToZip(zw, "project.json", project); err != nil {
		return nil, err
	}
	if err := writeJSONToZip(zw, "chapters.json", chapters); err != nil {
		return nil, err
	}
	if err := writeJSONToZip(zw, "relations.json", relations); err != nil {
		return nil, err
	}
	if err := writeJSONToZip(zw, "glossary.json", glossaries); err != nil {
		return nil, err
	}
	if err := writeJSONToZip(zw, "entities.json", entities); err != nil {
		return nil, err
	}

	if err := zw.Close(); err != nil {
		return nil, fmt.Errorf("close zip bundle: %w", err)
	}

	fi, err := bundleFile.Stat()
	if err != nil {
		return nil, fmt.Errorf("stat bundle file: %w", err)
	}

	log.Info().
		Str("project_id", projectID).
		Str("path", outputPath).
		Int64("size_bytes", fi.Size()).
		Msg("successfully packaged .neko project bundle")

	return &SynthesisResult{
		OutputPath:    outputPath,
		Format:        "neko",
		FileSizeBytes: fi.Size(),
		ChapterCount:  len(chapters),
		Duration:      time.Since(startTime),
	}, nil
}

func writeJSONToZip(zw *zip.Writer, filename string, v any) error {
	w, err := zw.Create(filename)
	if err != nil {
		return fmt.Errorf("create zip entry %s: %w", filename, err)
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		return fmt.Errorf("encode json for %s: %w", filename, err)
	}
	return nil
}

func formatTextToHTML(title string, text string) string {
	var sb strings.Builder
	sb.WriteString("<?xml version=\"1.0\" encoding=\"utf-8\"?>\n")
	sb.WriteString("<!DOCTYPE html>\n")
	sb.WriteString("<html xmlns=\"http://www.w3.org/1999/xhtml\">\n<head>\n")
	sb.WriteString(fmt.Sprintf("<title>%s</title>\n", title))
	sb.WriteString("<link rel=\"stylesheet\" type=\"text/css\" href=\"../Styles/style.css\"/>\n")
	sb.WriteString("</head>\n<body>\n")
	sb.WriteString(fmt.Sprintf("<h2>%s</h2>\n", title))

	lines := strings.Split(text, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		sb.WriteString(fmt.Sprintf("<p>%s</p>\n", trimmed))
	}
	sb.WriteString("</body>\n</html>")
	return sb.String()
}

func sanitizeFilename(name string) string {
	replacer := strings.NewReplacer(
		"/", "_",
		"\\", "_",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",
		">", "_",
		"|", "_",
		" ", "_",
	)
	clean := replacer.Replace(name)
	if clean == "" {
		return "novel_export"
	}
	return clean
}
