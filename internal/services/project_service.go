package services

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"

	"novelclaw/internal/dtos"
	"novelclaw/internal/gen/sqlc"
	"novelclaw/pkg/bookparser"
	"novelclaw/pkg/bookparser/archivebook"
	"novelclaw/pkg/bookparser/audiobook"
	"novelclaw/pkg/bookparser/comic"
	"novelclaw/pkg/bookparser/csv"
	"novelclaw/pkg/bookparser/doc"
	"novelclaw/pkg/bookparser/docx"
	"novelclaw/pkg/bookparser/epub"
	"novelclaw/pkg/bookparser/fb2"
	"novelclaw/pkg/bookparser/htmlfile"
	"novelclaw/pkg/bookparser/mobi"
	"novelclaw/pkg/bookparser/odt"
	"novelclaw/pkg/bookparser/pdf"
	"novelclaw/pkg/bookparser/plain"
	"novelclaw/pkg/bookparser/presentation"
	"novelclaw/pkg/bookparser/rtf"
	"novelclaw/pkg/bookparser/spreadsheet"
	"novelclaw/pkg/bookparser/tex"
	"novelclaw/pkg/ebookconv"
	"novelclaw/pkg/storage"
)

// ProjectService exposes project and chapter operations to the Wails v3 frontend.
type ProjectService struct {
	store    *storage.Storage
	registry bookparser.Registry
}

// NewProjectService creates a new ProjectService with all supported book parsers registered.
func NewProjectService(store *storage.Storage) *ProjectService {
	reg := bookparser.NewRegistry()
	reg.Register(epub.NewParser(), "epub", "kepub.epub")
	reg.Register(plain.NewParser(), "txt", "md", "markdown", "text")
	reg.Register(fb2.NewParser(), "fb2", "fbz")
	reg.Register(mobi.NewParser(), "mobi", "azw", "azw3", "amz", "prc")
	reg.Register(docx.NewParser(), "docx")
	reg.Register(doc.NewParser(), "doc")
	reg.Register(odt.NewParser(), "odt")
	reg.Register(pdf.NewParser(), "pdf")
	reg.Register(rtf.NewParser(), "rtf")
	reg.Register(htmlfile.NewParser(), "html", "htm", "xhtml")
	reg.Register(csv.NewParser(), "csv", "tsv")
	reg.Register(tex.NewParser(), "tex", "latex", "ltx")
	reg.Register(presentation.NewParser(), "pptx", "ppt", "odp")
	reg.Register(spreadsheet.NewParser(), "xlsx", "xls", "ods")
	reg.Register(comic.NewParser("cbz"), "cbz")
	reg.Register(comic.NewParser("cbr"), "cbr")
	reg.Register(comic.NewParser("cbt"), "cbt")
	reg.Register(comic.NewParser("cb7"), "cb7")
	reg.Register(archivebook.NewParser("zip"), "zip")
	reg.Register(archivebook.NewParser("rar"), "rar")
	reg.Register(archivebook.NewParser("7z"), "7z")
	reg.Register(audiobook.New(), "mp3", "m4a", "m4b", "flac", "ogg", "wav", "aac")

	return &ProjectService{
		store:    store,
		registry: reg,
	}
}

// CreateProject creates a new translation project from a DTO request.
func (s *ProjectService) CreateProject(ctx context.Context, req dtos.CreateProjectRequest) error {
	return s.store.CreateProject(ctx, req.ToCreateParams())
}

// GetProject retrieves project details by ID as a DTO.
func (s *ProjectService) GetProject(ctx context.Context, id string) (*dtos.ProjectDTO, error) {
	p, err := s.store.GetProject(ctx, id)
	if err != nil {
		return nil, err
	}
	dto := dtos.ToProjectDTO(p)
	return &dto, nil
}

// ListProjects lists all available projects as DTOs.
func (s *ProjectService) ListProjects(ctx context.Context) ([]dtos.ProjectDTO, error) {
	projects, err := s.store.ListProjects(ctx)
	if err != nil {
		return nil, err
	}
	dtosList := make([]dtos.ProjectDTO, 0, len(projects))
	for _, p := range projects {
		dtosList = append(dtosList, dtos.ToProjectDTO(p))
	}
	return dtosList, nil
}

// DeleteProject deletes a project and cascades related data.
func (s *ProjectService) DeleteProject(ctx context.Context, id string) error {
	_ = os.RemoveAll(filepath.Join("data", "assets", id))
	return s.store.DeleteProject(ctx, id)
}

// sanitizeVolumeDir converts a volume title or tag into a safe, deterministic directory name.
// Examples: "[vol1]" -> "vol1", "Volume 2" -> "volume_2", "" -> "vol_1"
func sanitizeVolumeDir(volName string, fallbackIndex int) string {
	cleaned := strings.TrimSpace(volName)
	cleaned = strings.TrimPrefix(cleaned, "[")
	cleaned = strings.TrimSuffix(cleaned, "]")
	cleaned = strings.TrimSpace(cleaned)

	var sb strings.Builder
	for _, r := range cleaned {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '-' {
			sb.WriteRune(r)
		} else if unicode.IsSpace(r) || r == '.' {
			sb.WriteRune('_')
		}
	}
	res := strings.Trim(sb.String(), "_")
	if res == "" {
		return fmt.Sprintf("vol_%d", fallbackIndex)
	}
	return strings.ToLower(res)
}

var readerAssetSrcRegex = regexp.MustCompile(`(?i)(<(?:img|image)[^>]*\s+(?:src|href|xlink:href)=["'])([^"']+)(["'])`)

func extractAndSaveBookImages(filePath, projectID, volDir string, parser bookparser.Parser) {
	images := ebookconv.CollectImages(parser, filePath)
	if len(images) == 0 {
		return
	}
	baseAssetDir := filepath.Join("data", "assets", projectID)
	var targetDir string
	if volDir != "" {
		targetDir = filepath.Join(baseAssetDir, volDir)
	} else {
		targetDir = baseAssetDir
	}
	_ = os.MkdirAll(targetDir, 0755)
	for _, img := range images {
		cleanSrc := strings.TrimPrefix(filepath.ToSlash(img.Src), "/")
		targetPath := filepath.Join(targetDir, cleanSrc)
		_ = os.MkdirAll(filepath.Dir(targetPath), 0755)
		_ = os.WriteFile(targetPath, img.Data, 0644)
		baseTarget := filepath.Join(targetDir, filepath.Base(cleanSrc))
		_ = os.WriteFile(baseTarget, img.Data, 0644)
	}
}

func rewriteReaderAssetURLs(content, projectID, volDir, contentPath string) string {
	if !strings.Contains(content, "<img") && !strings.Contains(content, "<image") {
		return content
	}
	baseDir := filepath.ToSlash(filepath.Dir(contentPath))
	if baseDir == "." {
		baseDir = ""
	}

	return readerAssetSrcRegex.ReplaceAllStringFunc(content, func(m string) string {
		sub := readerAssetSrcRegex.FindStringSubmatch(m)
		if len(sub) < 4 {
			return m
		}
		val := strings.TrimSpace(sub[2])
		lower := strings.ToLower(val)
		if strings.HasPrefix(val, "#") || strings.HasPrefix(lower, "http:") || strings.HasPrefix(lower, "https:") || strings.HasPrefix(lower, "data:") || strings.HasPrefix(lower, "blob:") || strings.HasPrefix(lower, "/api/") {
			return m
		}
		resolved := filepath.ToSlash(filepath.Clean(filepath.Join(baseDir, val)))
		cleanResolved := strings.TrimPrefix(resolved, "/")

		var assetURL string
		if volDir != "" {
			assetURL = `/api/v1/reader/` + url.PathEscape(projectID) + `/asset/` + url.PathEscape(volDir) + `/` + cleanResolved
		} else {
			assetURL = `/api/v1/reader/` + url.PathEscape(projectID) + `/asset/` + cleanResolved
		}
		return sub[1] + assetURL + sub[3]
	})
}


// UpdateProjectLanguages updates translation language direction for a project and returns the updated ProjectDTO.
func (s *ProjectService) UpdateProjectLanguages(ctx context.Context, req dtos.UpdateProjectLanguagesRequest) (*dtos.ProjectDTO, error) {
	if req.ID == "" {
		return nil, fmt.Errorf("project id is required")
	}
	if err := s.store.UpdateProjectLanguages(ctx, req.ID, req.SourceLang, req.TargetLang); err != nil {
		return nil, fmt.Errorf("update project languages: %w", err)
	}
	p, err := s.store.GetProject(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	dto := dtos.ToProjectDTO(p)
	return &dto, nil
}

// RenameProject updates the title of a project and returns the updated ProjectDTO.
func (s *ProjectService) RenameProject(ctx context.Context, req dtos.RenameProjectRequest) (*dtos.ProjectDTO, error) {
	if req.ID == "" {
		return nil, fmt.Errorf("project id is required")
	}
	trimmed := strings.TrimSpace(req.Title)
	if trimmed == "" {
		return nil, fmt.Errorf("project title cannot be empty")
	}
	if err := s.store.UpdateProjectTitle(ctx, req.ID, trimmed); err != nil {
		return nil, fmt.Errorf("rename project: %w", err)
	}
	p, err := s.store.GetProject(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	dto := dtos.ToProjectDTO(p)
	return &dto, nil
}

// UpdateProjectStyle updates the style name and style guide prompt for a project and returns the updated ProjectDTO.
func (s *ProjectService) UpdateProjectStyle(ctx context.Context, req dtos.UpdateProjectStyleRequest) (*dtos.ProjectDTO, error) {
	if req.ID == "" {
		return nil, fmt.Errorf("project id is required")
	}
	if err := s.store.UpdateProjectStyle(ctx, req.ID, req.StyleName, req.StyleGuide); err != nil {
		return nil, fmt.Errorf("update project style: %w", err)
	}
	p, err := s.store.GetProject(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	dto := dtos.ToProjectDTO(p)
	return &dto, nil
}


// ListChapters retrieves all chapters belonging to a project ordered by chapter index.
func (s *ProjectService) ListChapters(ctx context.Context, projectID string) ([]dtos.ChapterDTO, error) {
	chapters, err := s.store.ListChaptersByProject(ctx, projectID)
	if err != nil {
		return nil, err
	}
	result := make([]dtos.ChapterDTO, 0, len(chapters))
	for _, ch := range chapters {
		result = append(result, dtos.ToChapterDTO(ch))
	}
	return result, nil
}

// GetChapter retrieves a specific chapter by its unique ID.
func (s *ProjectService) GetChapter(ctx context.Context, chapterID string) (*dtos.ChapterDTO, error) {
	ch, err := s.store.GetChapter(ctx, chapterID)
	if err != nil {
		return nil, err
	}
	dto := dtos.ToChapterDTO(ch)
	return &dto, nil
}

// UpdateChapterTranslation applies direct user edits to a chapter's translated content.
func (s *ProjectService) UpdateChapterTranslation(ctx context.Context, req dtos.UpdateChapterInlineRequest) error {
	status := req.Status
	if status == "" {
		status = "completed"
	}
	return s.store.UpdateChapterTranslation(ctx, req.ID, req.TranslatedContent, status)
}

// ImportBook parses a novel file (EPUB, TXT, FB2) and ingests it into a new project.
func (s *ProjectService) ImportBook(ctx context.Context, req dtos.ImportBookRequest) (*dtos.ProjectDTO, error) {
	ext := strings.TrimPrefix(filepath.Ext(req.FilePath), ".")
	parser, err := s.registry.Parser(ext, req.FilePath)
	if err != nil {
		return nil, fmt.Errorf("find parser for %s: %w", req.FilePath, err)
	}

	book, err := parser.ParseBook(req.FilePath)
	if err != nil {
		return nil, fmt.Errorf("parse book %s: %w", req.FilePath, err)
	}

	projectID := req.ProjectID
	if projectID == "" {
		projectID = "proj_" + uuid.NewString()[:8]
	}

	title := book.Metadata.Title
	if title == "" {
		title = strings.TrimSuffix(filepath.Base(req.FilePath), filepath.Ext(req.FilePath))
	}

	author := book.Metadata.Author
	if author == "" {
		author = "Unknown Author"
	}

	srcLang := req.SourceLang
	if srcLang == "" {
		srcLang = book.Metadata.Language
	}
	if srcLang == "" {
		srcLang = "ja"
	}

	tgtLang := req.TargetLang
	if tgtLang == "" {
		tgtLang = "en"
	}

	createParams := sqlc.CreateProjectParams{
		ID:             projectID,
		Title:          title,
		Author:         author,
		SourceLang:     srcLang,
		TargetLang:     tgtLang,
		OriginalFormat: ext,
		TotalChapters:  int64(len(book.Chapters)),
	}

	if err := s.store.CreateProject(ctx, createParams); err != nil {
		return nil, fmt.Errorf("create project in storage: %w", err)
	}

	volDir := "vol1"
	extractAndSaveBookImages(req.FilePath, projectID, volDir, parser)

	for i, ch := range book.Chapters {
		chapIndex := int64(i + 1)
		chapID := fmt.Sprintf("%s_ch%d", projectID, chapIndex)
		chTitle := ch.Title
		if chTitle == "" {
			chTitle = fmt.Sprintf("Chapter %d", chapIndex)
		}

		rewrittenContent := rewriteReaderAssetURLs(ch.Content, projectID, volDir, ch.ContentPath)

		err := s.store.SaveChapterAndIndex(ctx, sqlc.CreateChapterParams{
			ID:           chapID,
			ProjectID:    projectID,
			ChapterIndex: chapIndex,
			Title:        chTitle,
			ContentPath:  ch.ContentPath,
			RawContent:   rewrittenContent,
		})
		if err != nil {
			return nil, fmt.Errorf("save chapter %d: %w", chapIndex, err)
		}
	}

	saved, err := s.store.GetProject(ctx, projectID)
	if err != nil {
		return nil, err
	}
	dto := dtos.ToProjectDTO(saved)
	return &dto, nil
}

// ImportBooksBatch imports multiple files, either as separate projects or aggregated as a multi-volume series.
func (s *ProjectService) ImportBooksBatch(ctx context.Context, req dtos.BatchImportBookRequest) (*dtos.BatchImportResultDTO, error) {
	if len(req.FilePaths) == 0 {
		return nil, fmt.Errorf("no file paths provided for batch import")
	}

	result := &dtos.BatchImportResultDTO{
		TotalFiles: len(req.FilePaths),
		Projects:   make([]dtos.ProjectDTO, 0, len(req.FilePaths)),
	}

	if req.AsSeries {
		seriesTitle := strings.TrimSpace(req.SeriesName)
		if seriesTitle == "" && len(req.FilePaths) > 0 {
			firstPath := req.FilePaths[0]
			ext := strings.TrimPrefix(filepath.Ext(firstPath), ".")
			if parser, err := s.registry.Parser(ext, firstPath); err == nil {
				if book, err := parser.ParseBook(firstPath); err == nil && book.Metadata.Title != "" {
					seriesTitle = book.Metadata.Title
				}
			}
			if seriesTitle == "" {
				base := filepath.Base(firstPath)
				seriesTitle = strings.TrimSuffix(base, filepath.Ext(base))
			}
		}
		if seriesTitle == "" {
			seriesTitle = "Series_" + time.Now().Format("20060102_1504")
		}

		projectID := "series_" + uuid.NewString()[:8]
		var totalChaptersCount int64 = 0
		var firstAuthor string
		srcLang := req.SourceLang
		tgtLang := req.TargetLang
		if tgtLang == "" {
			tgtLang = "en"
		}

		type queuedChapter struct {
			index       int64
			title       string
			contentPath string
			rawContent  string
		}
		var queuedChapters []queuedChapter

		for volIdx, path := range req.FilePaths {
			ext := strings.TrimPrefix(filepath.Ext(path), ".")
			parser, err := s.registry.Parser(ext, path)
			if err != nil {
				result.FailedCount++
				result.ErrorMessages = append(result.ErrorMessages, fmt.Sprintf("%s: parser not found: %v", filepath.Base(path), err))
				continue
			}

			book, err := parser.ParseBook(path)
			if err != nil {
				result.FailedCount++
				result.ErrorMessages = append(result.ErrorMessages, fmt.Sprintf("%s: parse failed: %v", filepath.Base(path), err))
				continue
			}

			if firstAuthor == "" && book.Metadata.Author != "" {
				firstAuthor = book.Metadata.Author
			}
			if srcLang == "" && book.Metadata.Language != "" {
				srcLang = book.Metadata.Language
			}

			volName := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
			volDir := sanitizeVolumeDir(volName, volIdx+1)
			extractAndSaveBookImages(path, projectID, volDir, parser)

			for _, ch := range book.Chapters {
				totalChaptersCount++
				chTitle := ch.Title
				if chTitle == "" {
					chTitle = fmt.Sprintf("Chapter %d", totalChaptersCount)
				}
				prefixedTitle := fmt.Sprintf("[%s] %s", volName, chTitle)
				volContentPath := fmt.Sprintf("%s/%s", volDir, ch.ContentPath)
				rewritten := rewriteReaderAssetURLs(ch.Content, projectID, volDir, ch.ContentPath)

				queuedChapters = append(queuedChapters, queuedChapter{
					index:       totalChaptersCount,
					title:       prefixedTitle,
					contentPath: volContentPath,
					rawContent:  rewritten,
				})
			}
			result.SuccessCount++
		}

		if len(queuedChapters) == 0 {
			return result, fmt.Errorf("no chapters could be extracted from provided files")
		}

		if srcLang == "" {
			srcLang = "ja"
		}

		createParams := sqlc.CreateProjectParams{
			ID:             projectID,
			Title:          seriesTitle,
			Author:         firstAuthor,
			SourceLang:     srcLang,
			TargetLang:     tgtLang,
			OriginalFormat: "series",
			TotalChapters:  totalChaptersCount,
		}

		if err := s.store.CreateProject(ctx, createParams); err != nil {
			return nil, fmt.Errorf("create series project: %w", err)
		}

		for _, qCh := range queuedChapters {
			chapID := fmt.Sprintf("%s_ch%d", projectID, qCh.index)
			err := s.store.SaveChapterAndIndex(ctx, sqlc.CreateChapterParams{
				ID:           chapID,
				ProjectID:    projectID,
				ChapterIndex: qCh.index,
				Title:        qCh.title,
				ContentPath:  qCh.contentPath,
				RawContent:   qCh.rawContent,
			})
			if err != nil {
				return nil, fmt.Errorf("save chapter %d: %w", qCh.index, err)
			}
		}

		saved, err := s.store.GetProject(ctx, projectID)
		if err == nil {
			result.Projects = append(result.Projects, dtos.ToProjectDTO(saved))
		}

		return result, nil
	}

	for _, path := range req.FilePaths {
		proj, err := s.ImportBook(ctx, dtos.ImportBookRequest{
			FilePath:   path,
			SourceLang: req.SourceLang,
			TargetLang: req.TargetLang,
		})
		if err != nil {
			result.FailedCount++
			result.ErrorMessages = append(result.ErrorMessages, fmt.Sprintf("%s: %v", filepath.Base(path), err))
			continue
		}
		result.SuccessCount++
		result.Projects = append(result.Projects, *proj)
	}

	return result, nil
}

// InheritKnowledge copies character graph (L2) and glossary (L3) from a source project to a target project.
func (s *ProjectService) InheritKnowledge(ctx context.Context, req dtos.InheritKnowledgeRequest) (*dtos.InheritKnowledgeResultDTO, error) {
	entCount, relCount, glossCount, err := s.store.InheritProjectKnowledge(ctx, req.SourceProjectID, req.TargetProjectID)
	if err != nil {
		return nil, err
	}
	return &dtos.InheritKnowledgeResultDTO{
		EntitiesCount:  entCount,
		RelationsCount: relCount,
		GlossaryCount:  glossCount,
		Success:        true,
	}, nil
}

// RollbackChapter restores the chapter translation and status to the latest saved checkpoint.
func (s *ProjectService) RollbackChapter(ctx context.Context, projectID string, chapterIndex int64) (*dtos.ChapterDTO, error) {
	ch, err := s.store.RollbackChapterToCheckpoint(ctx, projectID, chapterIndex)
	if err != nil {
		return nil, err
	}
	dto := dtos.ToChapterDTO(*ch)
	return &dto, nil
}

// AppendBookToProject parses an additional book/volume and appends its chapters to an existing project/series.
func (s *ProjectService) AppendBookToProject(ctx context.Context, req dtos.AppendBookRequest) (*dtos.ProjectDTO, error) {
	if req.ProjectID == "" {
		return nil, fmt.Errorf("project_id is required")
	}
	if req.FilePath == "" {
		return nil, fmt.Errorf("file_path is required")
	}

	proj, err := s.store.GetProject(ctx, req.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("project not found: %w", err)
	}
	_ = proj

	ext := strings.TrimPrefix(filepath.Ext(req.FilePath), ".")
	parser, err := s.registry.Parser(ext, req.FilePath)
	if err != nil {
		return nil, fmt.Errorf("unsupported file format: %w", err)
	}

	book, err := parser.ParseBook(req.FilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse book: %w", err)
	}

	volName := strings.TrimSpace(req.VolName)
	if volName == "" {
		base := filepath.Base(req.FilePath)
		volName = strings.TrimSuffix(base, filepath.Ext(base))
	}
	volDir := sanitizeVolumeDir(volName, 1)

	extractAndSaveBookImages(req.FilePath, req.ProjectID, volDir, parser)

	chaps, err := s.store.ListChaptersByProject(ctx, req.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list current chapters: %w", err)
	}

	var maxIndex int64 = 0
	for _, ch := range chaps {
		if ch.ChapterIndex > maxIndex {
			maxIndex = ch.ChapterIndex
		}
	}

	for _, ch := range book.Chapters {
		maxIndex++
		chTitle := ch.Title
		if chTitle == "" {
			chTitle = fmt.Sprintf("Chapter %d", maxIndex)
		}
		prefixedTitle := fmt.Sprintf("[%s] %s", volName, chTitle)
		volContentPath := fmt.Sprintf("%s/%s", volDir, ch.ContentPath)
		rewritten := rewriteReaderAssetURLs(ch.Content, req.ProjectID, volDir, ch.ContentPath)

		chapID := fmt.Sprintf("%s_ch%d", req.ProjectID, maxIndex)
		err := s.store.SaveChapterAndIndex(ctx, sqlc.CreateChapterParams{
			ID:           chapID,
			ProjectID:    req.ProjectID,
			ChapterIndex: maxIndex,
			Title:        prefixedTitle,
			ContentPath:  volContentPath,
			RawContent:   rewritten,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to save chapter %d: %w", maxIndex, err)
		}
	}

	if err := s.store.UpdateProjectProgress(ctx, req.ProjectID, maxIndex); err != nil {
		return nil, fmt.Errorf("failed to update project total chapters: %w", err)
	}

	updated, err := s.store.GetProject(ctx, req.ProjectID)
	if err != nil {
		return nil, err
	}
	dto := dtos.ToProjectDTO(updated)
	return &dto, nil
}

