package dtos

import (
	"time"

	"novelclaw/internal/gen/sqlc"
)

// GlossaryTermDTO represents a terminology entry for frontend display.
type GlossaryTermDTO struct {
	ID         string    `json:"id"`
	ProjectID  string    `json:"project_id"`
	SourceTerm string    `json:"source_term"`
	TargetTerm string    `json:"target_term"`
	Category   string    `json:"category"`
	Notes      string    `json:"notes"`
	CreatedAt  time.Time `json:"created_at"`
}

// UpsertGlossaryTermRequest contains parameters for creating or updating a glossary term.
type UpsertGlossaryTermRequest struct {
	ID         string `json:"id"`
	ProjectID  string `json:"project_id"`
	SourceTerm string `json:"source_term"`
	TargetTerm string `json:"target_term"`
	Category   string `json:"category"`
}

// ToGlossaryTermDTO converts a sqlc.Glossary row to GlossaryTermDTO.
func ToGlossaryTermDTO(g sqlc.Glossary) GlossaryTermDTO {
	return GlossaryTermDTO{
		ID:         g.ID,
		ProjectID:  g.ProjectID,
		SourceTerm: g.SourceTerm,
		TargetTerm: g.TargetTerm,
		Category:   g.Category,
		Notes:      g.Notes,
		CreatedAt:  g.CreatedAt,
	}
}

// ToUpsertParams converts UpsertGlossaryTermRequest to sqlc.UpsertGlossaryTermParams.
func (r UpsertGlossaryTermRequest) ToUpsertParams() sqlc.UpsertGlossaryTermParams {
	return sqlc.UpsertGlossaryTermParams{
		ID:         r.ID,
		ProjectID:  r.ProjectID,
		SourceTerm: r.SourceTerm,
		TargetTerm: r.TargetTerm,
		Category:   r.Category,
	}
}

// RemoteGlossarySourceDTO represents an online dictionary source (e.g., community repositories on GitHub) for frontend display.
type RemoteGlossarySourceDTO struct {
	ID              string `json:"id"`
	Title           string `json:"title"`
	Genre           string `json:"genre"`
	Description     string `json:"description"`
	URL             string `json:"url"`
	Format          string `json:"format"`
	DefaultCategory string `json:"default_category"`
}

// DownloadGlossaryRequest contains parameters for downloading and importing a remote dictionary file.
type DownloadGlossaryRequest struct {
	ProjectID       string `json:"project_id"`
	URL             string `json:"url"`
	Format          string `json:"format"` // "vietphrase" | "tsv" | "json"
	DefaultCategory string `json:"default_category"`
}

