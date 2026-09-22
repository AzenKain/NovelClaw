package dtos

// WorldCategoryDTO represents a dynamic classification group for world lore.
type WorldCategoryDTO struct {
	ID           string `json:"id"`
	ProjectID    string `json:"project_id"`
	Slug         string `json:"slug"`
	Name         string `json:"name"`
	Icon         string `json:"icon"`
	Description  string `json:"description"`
	DisplayOrder int    `json:"display_order"`
	EntryCount   int    `json:"entry_count"`
	CreatedAt    string `json:"created_at"`
}

// WorldEntryDTO represents a specific Lorebook entity in the World Bible.
type WorldEntryDTO struct {
	ID                 string            `json:"id"`
	ProjectID          string            `json:"project_id"`
	CategoryID         string            `json:"category_id"`
	CategorySlug       string            `json:"category_slug"`
	CategoryName       string            `json:"category_name"`
	Name               string            `json:"name"`
	Aliases            []string          `json:"aliases"`
	Summary            string            `json:"summary"`
	FullDescription    string            `json:"full_description"`
	Attributes         map[string]string `json:"attributes"`
	DiscoveredBy       string            `json:"discovered_by"` // 'ai_scan' | 'manual' | 'chat_command'
	SourceChapterIndex int64             `json:"source_chapter_index"`
	IsVerified         bool              `json:"is_verified"`
	CreatedAt          string            `json:"created_at"`
	UpdatedAt          string            `json:"updated_at"`
}

// UpsertWorldCategoryRequest holds inputs to create or update a category.
type UpsertWorldCategoryRequest struct {
	ID           string `json:"id,omitempty"`
	ProjectID    string `json:"project_id"`
	Slug         string `json:"slug"`
	Name         string `json:"name"`
	Icon         string `json:"icon"`
	Description  string `json:"description"`
	DisplayOrder int    `json:"display_order"`
}

// UpsertWorldEntryRequest holds inputs to create or update an entry.
type UpsertWorldEntryRequest struct {
	ID                 string            `json:"id,omitempty"`
	ProjectID          string            `json:"project_id"`
	CategoryID         string            `json:"category_id"`
	Name               string            `json:"name"`
	Aliases            []string          `json:"aliases"`
	Summary            string            `json:"summary"`
	FullDescription    string            `json:"full_description"`
	Attributes         map[string]string `json:"attributes"`
	SourceChapterIndex int64             `json:"source_chapter_index"`
	IsVerified         bool              `json:"is_verified"`
}

// ScanWorldRequest requests the Autonomous World Builder Agent to analyze chapters.
type ScanWorldRequest struct {
	ProjectID          string `json:"project_id"`
	StartChapterIndex  int64  `json:"start_chapter_index"`
	EndChapterIndex    int64  `json:"end_chapter_index"`
	IncludeWorldBible  bool   `json:"include_world_bible"`
}

// ScanWorldResponse returns the outcome of an autonomous world scan.
type ScanWorldResponse struct {
	Success          bool     `json:"success"`
	InferredGenre    string   `json:"inferred_genre"`
	CategoriesAdded  int      `json:"categories_added"`
	EntriesAdded     int      `json:"entries_added"`
	SummaryReport    string   `json:"summary_report"`
	Errors           []string `json:"errors,omitempty"`
}

// ExportWorldBibleResponse conveys exported Markdown or JSON.
type ExportWorldBibleResponse struct {
	Format  string `json:"format"`
	Content string `json:"content"`
}

// LearnResultDTO conveys rules and few-shots inferred from user edits via /learn.
type LearnResultDTO struct {
	Success         bool     `json:"success"`
	RulesLearned    []string `json:"rules_learned"`
	FewShotsLearned int      `json:"few_shots_learned"`
	Summary         string   `json:"summary"`
	TargetSkillID   string   `json:"target_skill_id"`
}

// CharacterVoiceDTO conveys persona and dialogue style rules for character voice prompting.
type CharacterVoiceDTO struct {
	CharacterID   string   `json:"character_id"`
	Name          string   `json:"name"`
	Aliases       []string `json:"aliases"`
	VoiceTone     string   `json:"voice_tone"`
	DialogueRules []string `json:"dialogue_rules"`
	Catchphrases  []string `json:"catchphrases"`
}

// UpsertVoiceRequest registers or updates a character voice profile.
type UpsertVoiceRequest struct {
	ProjectID     string   `json:"project_id"`
	CharacterID   string   `json:"character_id"`
	Name          string   `json:"name"`
	Aliases       []string `json:"aliases"`
	VoiceTone     string   `json:"voice_tone"`
	DialogueRules []string `json:"dialogue_rules"`
	Catchphrases  []string `json:"catchphrases"`
}


