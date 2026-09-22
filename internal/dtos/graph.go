package dtos

import (
	"time"

	"novelclaw/internal/gen/sqlc"
	"novelclaw/pkg/graph"
	"novelclaw/pkg/storage"
)

// EntityDTO represents a character or entity node in the story graph.
type EntityDTO struct {
	ID               string                 `json:"id"`
	ProjectID        string                 `json:"project_id"`
	Name             string                 `json:"name"`
	Aliases          []string               `json:"aliases"`
	Category         string                 `json:"category"`
	Gender           string                 `json:"gender"`
	Role             string                 `json:"role"`
	FirstSeenChapter int64                  `json:"first_seen_chapter"`
	Metadata         map[string]interface{} `json:"metadata"`
}

// UpsertEntityRequest contains parameters for creating or updating an entity node.
type UpsertEntityRequest struct {
	ID               string                 `json:"id"`
	ProjectID        string                 `json:"project_id"`
	Name             string                 `json:"name"`
	Aliases          []string               `json:"aliases"`
	Category         string                 `json:"category"`
	Gender           string                 `json:"gender"`
	Role             string                 `json:"role"`
	FirstSeenChapter int64                  `json:"first_seen_chapter"`
	Metadata         map[string]interface{} `json:"metadata"`
}

// RelationDTO represents an address term relation between two characters.
type RelationDTO struct {
	ID           string    `json:"id"`
	ProjectID    string    `json:"project_id"`
	FromChar     string    `json:"from_char"`
	ToChar       string    `json:"to_char"`
	CallAs       string    `json:"call_as"`
	SelfCallAs   string    `json:"self_call_as"`
	SinceChapter int64     `json:"since_chapter"`
	Tone         string    `json:"tone"`
	IsLocked     bool      `json:"is_locked"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// UpsertRelationRequest contains parameters for creating or updating a character relation.
type UpsertRelationRequest struct {
	ID           string `json:"id"`
	ProjectID    string `json:"project_id"`
	FromChar     string `json:"from_char"`
	ToChar       string `json:"to_char"`
	CallAs       string `json:"call_as"`
	SelfCallAs   string `json:"self_call_as"`
	SinceChapter int64  `json:"since_chapter"`
	Tone         string `json:"tone"`
	IsLocked     bool   `json:"is_locked,omitempty"`
}

// SelectiveContextDTO contains the filtered characters, relations, and glossary for a chunk.
type SelectiveContextDTO struct {
	ActiveEntities  []EntityDTO       `json:"active_entities"`
	ActiveRelations []RelationDTO     `json:"active_relations"`
	ActiveGlossary  []GlossaryTermDTO `json:"active_glossary"`
	FormattedPrompt string            `json:"formatted_prompt"`
}

// ToEntityDTO converts storage.Entity to EntityDTO.
func ToEntityDTO(e storage.Entity) EntityDTO {
	return EntityDTO{
		ID:               e.ID,
		ProjectID:        e.ProjectID,
		Name:             e.Name,
		Aliases:          e.Aliases,
		Category:         e.Category,
		Gender:           e.Gender,
		Role:             e.Role,
		FirstSeenChapter: e.FirstSeenChapter,
		Metadata:         e.Metadata,
	}
}

// ToDomainEntity converts UpsertEntityRequest to storage.Entity.
func (r UpsertEntityRequest) ToDomainEntity() storage.Entity {
	return storage.Entity{
		ID:               r.ID,
		ProjectID:        r.ProjectID,
		Name:             r.Name,
		Aliases:          r.Aliases,
		Category:         r.Category,
		Gender:           r.Gender,
		Role:             r.Role,
		FirstSeenChapter: r.FirstSeenChapter,
		Metadata:         r.Metadata,
	}
}

// ToRelationDTO converts sqlc.CharacterRelation to RelationDTO.
func ToRelationDTO(r sqlc.CharacterRelation) RelationDTO {
	return RelationDTO{
		ID:           r.ID,
		ProjectID:    r.ProjectID,
		FromChar:     r.FromChar,
		ToChar:       r.ToChar,
		CallAs:       r.CallAs,
		SelfCallAs:   r.SelfCallAs,
		SinceChapter: r.SinceChapter,
		Tone:         r.Tone,
		IsLocked:     r.IsLocked == 1,
		CreatedAt:    r.CreatedAt,
		UpdatedAt:    r.UpdatedAt,
	}
}

// ToUpsertParams converts UpsertRelationRequest to sqlc.UpsertRelationParams.
func (r UpsertRelationRequest) ToUpsertParams() sqlc.UpsertRelationParams {
	isLocked := int64(0)
	if r.IsLocked {
		isLocked = 1
	}
	return sqlc.UpsertRelationParams{
		ID:           r.ID,
		ProjectID:    r.ProjectID,
		FromChar:     r.FromChar,
		ToChar:       r.ToChar,
		CallAs:       r.CallAs,
		SelfCallAs:   r.SelfCallAs,
		SinceChapter: r.SinceChapter,
		Tone:         r.Tone,
		IsLocked:     isLocked,
	}
}

// ToSelectiveContextDTO converts graph.SelectiveContext to SelectiveContextDTO.
func ToSelectiveContextDTO(sc *graph.SelectiveContext) *SelectiveContextDTO {
	if sc == nil {
		return nil
	}
	entities := make([]EntityDTO, 0, len(sc.ActiveEntities))
	for _, e := range sc.ActiveEntities {
		entities = append(entities, ToEntityDTO(e))
	}
	relations := make([]RelationDTO, 0, len(sc.ActiveRelations))
	for _, r := range sc.ActiveRelations {
		relations = append(relations, ToRelationDTO(r))
	}
	glossaries := make([]GlossaryTermDTO, 0, len(sc.ActiveGlossary))
	for _, g := range sc.ActiveGlossary {
		glossaries = append(glossaries, ToGlossaryTermDTO(g))
	}
	return &SelectiveContextDTO{
		ActiveEntities:  entities,
		ActiveRelations: relations,
		ActiveGlossary:  glossaries,
		FormattedPrompt: sc.FormattedPrompt,
	}
}

// AutoScanRequest specifies parameters for AI auto-extracting characters and relations.
type AutoScanRequest struct {
	ProjectID    string `json:"project_id"`
	ChapterIndex int64  `json:"chapter_index,omitempty"`
	ScanMode     string `json:"scan_mode"` // "volume" (default), "pre_scan", "current_chapter", "all"
	Volume       string `json:"volume,omitempty"`
}

// AutoScanResultDTO contains the results of AI auto-scanning entities and relations.
type AutoScanResultDTO struct {
	Success             bool          `json:"success"`
	Message             string        `json:"message"`
	EntitiesFound       int           `json:"entities_found"`
	RelationsFound      int           `json:"relations_found"`
	DurationMs          int64         `json:"duration_ms"`
	VolumeScanned       string        `json:"volume_scanned,omitempty"`
	EffectiveChapter    int64         `json:"effective_chapter"`
	DiscoveredEntities  []EntityDTO   `json:"discovered_entities"`
	DiscoveredRelations []RelationDTO `json:"discovered_relations"`
}

// GraphReadinessDTO conveys whether the graph is ready and up-to-date for a specific chapter/volume.
type GraphReadinessDTO struct {
	IsReady               bool   `json:"is_ready"`
	ProjectID             string `json:"project_id"`
	ChapterIndex          int64  `json:"chapter_index"`
	VolumeTag             string `json:"volume_tag,omitempty"`
	VolumeStartChapter    int64  `json:"volume_start_chapter,omitempty"`
	VolumeEndChapter      int64  `json:"volume_end_chapter,omitempty"`
	HasVolumeScan         bool   `json:"has_volume_scan"`
	VolumeRelationsCount  int    `json:"volume_relations_count"`
	TotalProjectRelations int    `json:"total_project_relations"`
	TotalProjectEntities  int    `json:"total_project_entities"`
	Message               string `json:"message"`
}


