package dtos

// ExportEbookRequest defines the parameters for synthesizing an ebook.
type ExportEbookRequest struct {
	ProjectID         string `json:"project_id"`
	TargetFormat      string `json:"target_format"`
	OutputPath        string `json:"output_path"`
	CustomTitle       string `json:"custom_title,omitempty"`
	CustomAuthor      string `json:"custom_author,omitempty"`
	CoverImagePath    string `json:"cover_image_path,omitempty"`
	UseTranslatedOnly bool   `json:"use_translated_only"`
	Volume            string `json:"volume,omitempty"`
}

// ExportResultDTO represents the outcome of an export or bundle packaging operation.
type ExportResultDTO struct {
	OutputPath    string `json:"output_path"`
	Format        string `json:"format"`
	FileSizeBytes int64  `json:"file_size_bytes"`
	ChapterCount  int    `json:"chapter_count"`
	DurationMs    int64  `json:"duration_ms"`
}

// ExportBundleRequest specifies the project ID and destination path for a .neko bundle.
type ExportBundleRequest struct {
	ProjectID  string `json:"project_id"`
	OutputPath string `json:"output_path"`
}

// FormatDescriptorDTO provides user-friendly metadata for each supported export format.
type FormatDescriptorDTO struct {
	Format         string `json:"format"`
	Name           string `json:"name"`
	Extension      string `json:"extension"`
	Description    string `json:"description"`
	RecommendedFor string `json:"recommended_for"`
}
