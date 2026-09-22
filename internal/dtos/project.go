package dtos

import (
	"time"

	"novelclaw/internal/gen/sqlc"
)

// ProjectDTO represents a translation project for frontend display.
type ProjectDTO struct {
	ID             string    `json:"id"`
	Title          string    `json:"title"`
	Author         string    `json:"author"`
	SourceLang     string    `json:"source_lang"`
	TargetLang     string    `json:"target_lang"`
	OriginalFormat string    `json:"original_format"`
	TotalChapters  int64     `json:"total_chapters"`
	StyleGuide     string    `json:"style_guide"`
	StyleName      string    `json:"style_name"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// CreateProjectRequest contains parameters for creating a new translation project.
type CreateProjectRequest struct {
	ID             string `json:"id"`
	Title          string `json:"title"`
	Author         string `json:"author"`
	SourceLang     string `json:"source_lang"`
	TargetLang     string `json:"target_lang"`
	OriginalFormat string `json:"original_format"`
	TotalChapters  int64  `json:"total_chapters"`
}

// UpdateProjectLanguagesRequest provides payload for updating source and target languages of a project.
type UpdateProjectLanguagesRequest struct {
	ID         string `json:"id"`
	SourceLang string `json:"source_lang"`
	TargetLang string `json:"target_lang"`
}

// RenameProjectRequest provides payload for renaming a project.
type RenameProjectRequest struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

// UpdateProjectStyleRequest provides payload for updating custom or scouted style guide of a project.
type UpdateProjectStyleRequest struct {
	ID         string `json:"id"`
	StyleName  string `json:"style_name"`
	StyleGuide string `json:"style_guide"`
}

// ToProjectDTO converts a database Project row to ProjectDTO.
func ToProjectDTO(p sqlc.Project) ProjectDTO {
	return ProjectDTO{
		ID:             p.ID,
		Title:          p.Title,
		Author:         p.Author,
		SourceLang:     p.SourceLang,
		TargetLang:     p.TargetLang,
		OriginalFormat: p.OriginalFormat,
		TotalChapters:  p.TotalChapters,
		StyleGuide:     p.StyleGuide,
		StyleName:      p.StyleName,
		CreatedAt:      p.CreatedAt,
		UpdatedAt:      p.UpdatedAt,
	}
}

// ToCreateParams converts CreateProjectRequest to sqlc.CreateProjectParams.
func (r CreateProjectRequest) ToCreateParams() sqlc.CreateProjectParams {
	return sqlc.CreateProjectParams{
		ID:             r.ID,
		Title:          r.Title,
		Author:         r.Author,
		SourceLang:     r.SourceLang,
		TargetLang:     r.TargetLang,
		OriginalFormat: r.OriginalFormat,
		TotalChapters:  r.TotalChapters,
	}
}

// ChapterDTO represents an individual chapter for the workspace UI.
type ChapterDTO struct {
	ID                string    `json:"id"`
	ProjectID         string    `json:"project_id"`
	ChapterIndex      int64     `json:"chapter_index"`
	Title             string    `json:"title"`
	ContentPath       string    `json:"content_path"`
	RawContent        string    `json:"raw_content"`
	TranslatedContent string    `json:"translated_content"`
	Status            string    `json:"status"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// UpdateChapterInlineRequest provides payload for manual in-line human editing.
type UpdateChapterInlineRequest struct {
	ID                string `json:"id"`
	TranslatedContent string `json:"translated_content"`
	Status            string `json:"status,omitempty"`
}

// ImportBookRequest describes a book file to parse and ingest into a new project.
type ImportBookRequest struct {
	FilePath   string `json:"file_path"`
	ProjectID  string `json:"project_id,omitempty"`
	SourceLang string `json:"source_lang,omitempty"`
	TargetLang string `json:"target_lang,omitempty"`
}

// ToChapterDTO converts a database Chapter row to ChapterDTO.
func ToChapterDTO(c sqlc.Chapter) ChapterDTO {
	return ChapterDTO{
		ID:                c.ID,
		ProjectID:         c.ProjectID,
		ChapterIndex:      c.ChapterIndex,
		Title:             c.Title,
		ContentPath:       c.ContentPath,
		RawContent:        c.RawContent,
		TranslatedContent: c.TranslatedContent,
		Status:            c.Status,
		CreatedAt:         c.CreatedAt,
		UpdatedAt:         c.UpdatedAt,
	}
}

// BatchImportBookRequest specifies multiple files to ingest into the studio.
type BatchImportBookRequest struct {
	FilePaths  []string `json:"file_paths"`
	SourceLang string   `json:"source_lang,omitempty"`
	TargetLang string   `json:"target_lang,omitempty"`
	AsSeries   bool     `json:"as_series"`
	SeriesName string   `json:"series_name,omitempty"`
}

// BatchImportResultDTO summarizes the results of batch file ingestion.
type BatchImportResultDTO struct {
	TotalFiles    int          `json:"total_files"`
	SuccessCount  int          `json:"success_count"`
	FailedCount   int          `json:"failed_count"`
	Projects      []ProjectDTO `json:"projects"`
	ErrorMessages []string     `json:"error_messages,omitempty"`
}

// InheritKnowledgeRequest specifies source and target project for copying universe knowledge.
type InheritKnowledgeRequest struct {
	SourceProjectID string `json:"source_project_id"`
	TargetProjectID string `json:"target_project_id"`
}

// InheritKnowledgeResultDTO returns the count of copied items.
type InheritKnowledgeResultDTO struct {
	EntitiesCount  int  `json:"entities_count"`
	RelationsCount int  `json:"relations_count"`
	GlossaryCount  int  `json:"glossary_count"`
	Success        bool `json:"success"`
}

// AppendBookRequest specifies a book file to append as a new volume to an existing project/series.
type AppendBookRequest struct {
	ProjectID string `json:"project_id"`
	FilePath  string `json:"file_path"`
	VolName   string `json:"vol_name,omitempty"`
}
