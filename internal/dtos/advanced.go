package dtos

import (
	"novelclaw/pkg/chapterfacts"
	"novelclaw/pkg/r19"
	"novelclaw/pkg/stylestat"
)

// R19ConfigDTO exposes local R19 parameters to the frontend.
type R19ConfigDTO struct {
	Enabled            bool   `json:"enabled"`
	NeutralizeAges     bool   `json:"neutralize_ages"`
	LiteraryDisclaimer bool   `json:"literary_disclaimer"`
	TagFormat          string `json:"tag_format"`
}

// ToR19Config converts DTO to domain r19.Config.
func (d R19ConfigDTO) ToR19Config() r19.Config {
	return r19.Config{
		Enabled:            d.Enabled,
		NeutralizeAges:     d.NeutralizeAges,
		LiteraryDisclaimer: d.LiteraryDisclaimer,
		TagFormat:          d.TagFormat,
	}
}

// RepetitivePhraseDTO represents an overused phrase detected in translated text.
type RepetitivePhraseDTO struct {
	Phrase string `json:"phrase"`
	Count  int    `json:"count"`
}

// StyleReportDTO provides deterministic novel stylistic metrics to the frontend.
type StyleReportDTO struct {
	TotalSentences         int                   `json:"total_sentences"`
	TotalWords             int                   `json:"total_words"`
	DuplicateSentenceCount int                   `json:"duplicate_sentence_count"`
	RepetitivePhrases      []RepetitivePhraseDTO `json:"repetitive_phrases"`
	GlossaryAdherenceRate  float64               `json:"glossary_adherence_rate"`
	QualityScore           float64               `json:"quality_score"`
}

// ToStyleReportDTO converts domain stylestat.StyleReport to StyleReportDTO.
func ToStyleReportDTO(r stylestat.StyleReport) StyleReportDTO {
	var phrases []RepetitivePhraseDTO
	for _, p := range r.RepetitivePhrases {
		phrases = append(phrases, RepetitivePhraseDTO{
			Phrase: p.Phrase,
			Count:  p.Count,
		})
	}
	return StyleReportDTO{
		TotalSentences:         r.TotalSentences,
		TotalWords:             r.TotalWords,
		DuplicateSentenceCount: r.DuplicateSentenceCount,
		RepetitivePhrases:      phrases,
		GlossaryAdherenceRate:  r.GlossaryAdherenceRate,
		QualityScore:           r.QualityScore,
	}
}

// RelationshipChangeDTO represents an episodic address term change.
type RelationshipChangeDTO struct {
	FromChar   string `json:"from_char"`
	ToChar     string `json:"to_char"`
	CallAs     string `json:"call_as"`
	SelfCallAs string `json:"self_call_as"`
	Tone       string `json:"tone"`
}

// StateChangeDTO represents entity state change across chapters.
type StateChangeDTO struct {
	Entity   string `json:"entity"`
	Field    string `json:"field"`
	OldValue string `json:"old_value"`
	NewValue string `json:"new_value"`
	Reason   string `json:"reason"`
}

// CastIntroDTO represents newly introduced characters.
type CastIntroDTO struct {
	Name   string `json:"name"`
	Role   string `json:"role"`
	Gender string `json:"gender"`
}

// ChapterFactsDTO exposes episodic extracted facts to the frontend.
type ChapterFactsDTO struct {
	Title               string                  `json:"title"`
	Summary             string                  `json:"summary"`
	KeyEvents           []string                `json:"key_events"`
	RelationshipChanges []RelationshipChangeDTO `json:"relationship_changes"`
	StateChanges        []StateChangeDTO        `json:"state_changes"`
	CastIntros          []CastIntroDTO          `json:"cast_intros"`
}

// ToChapterFactsDTO converts domain chapterfacts.ChapterFacts to ChapterFactsDTO.
func ToChapterFactsDTO(f *chapterfacts.ChapterFacts) *ChapterFactsDTO {
	if f == nil {
		return nil
	}

	var rels []RelationshipChangeDTO
	for _, r := range f.RelationshipChanges {
		rels = append(rels, RelationshipChangeDTO{
			FromChar:   r.FromChar,
			ToChar:     r.ToChar,
			CallAs:     r.CallAs,
			SelfCallAs: r.SelfCallAs,
			Tone:       r.Tone,
		})
	}

	var states []StateChangeDTO
	for _, s := range f.StateChanges {
		states = append(states, StateChangeDTO{
			Entity:   s.Entity,
			Field:    s.Field,
			OldValue: s.OldValue,
			NewValue: s.NewValue,
			Reason:   s.Reason,
		})
	}

	var casts []CastIntroDTO
	for _, c := range f.CastIntros {
		casts = append(casts, CastIntroDTO{
			Name:   c.Name,
			Role:   c.Role,
			Gender: c.Gender,
		})
	}

	return &ChapterFactsDTO{
		Title:               f.Title,
		Summary:             f.Summary,
		KeyEvents:           f.KeyEvents,
		RelationshipChanges: rels,
		StateChanges:        states,
		CastIntros:          casts,
	}
}
