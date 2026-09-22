package voice

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"novelclaw/pkg/storage"
)

// CharacterVoice defines the linguistic cadence, honorific rules, and persona voice of a character.
type CharacterVoice struct {
	CharacterID   string   `json:"character_id"`
	Name          string   `json:"name"`
	Aliases       []string `json:"aliases"`
	VoiceTone     string   `json:"voice_tone"`     // e.g., "Solemn and classical", "Cold and imperious", "Playful and humorous"
	DialogueRules []string `json:"dialogue_rules"` // Specific speech constraints
	Catchphrases  []string `json:"catchphrases"`   // Signature recurring expressions
}

// Registry stores and matches character voices per project.
type Registry struct {
	mu     sync.RWMutex
	voices map[string]map[string]CharacterVoice // projectID -> characterID -> CharacterVoice
}

// NewRegistry creates a new in-memory voice registry.
func NewRegistry() *Registry {
	return &Registry{
		voices: make(map[string]map[string]CharacterVoice),
	}
}

// SyncEntityVoice converts an entity into a CharacterVoice profile.
func SyncEntityVoice(ent storage.Entity) CharacterVoice {
	v := CharacterVoice{
		CharacterID: ent.ID,
		Name:        ent.Name,
		Aliases:     ent.Aliases,
	}

	if ent.Metadata != nil {
		if tone, ok := ent.Metadata["voice_tone"].(string); ok && tone != "" {
			v.VoiceTone = tone
		}
		if rules, ok := ent.Metadata["dialogue_rules"].([]string); ok {
			v.DialogueRules = rules
		}
		if cp, ok := ent.Metadata["catchphrases"].([]string); ok {
			v.Catchphrases = cp
		}
	}

	// Heuristic default if voice tone is empty
	if v.VoiceTone == "" {
		switch strings.ToLower(ent.Role) {
		case "protagonist", "lead", "main_character", "nam chính", "nữ chính":
			v.VoiceTone = "Resolute, decisive, sharp and confident presence"
		case "mentor", "elder", "master", "sư phụ", "trưởng lão":
			v.VoiceTone = "Solemn, profound, venerable and authoritative"
		case "villain", "antagonist", "phản diện":
			v.VoiceTone = "Cold, imperious, imposing and sharp"
		case "sidekick", "comic_relief", "companion", "bạn đồng hành":
			v.VoiceTone = "Warm, lively, humorous and expressive"
		default:
			if ent.Gender == "female" {
				v.VoiceTone = "Graceful, poised, composed"
			}
		}
	}

	return v
}

// SyncFromEntities populates the voice registry from knowledge graph character entities.
func (r *Registry) SyncFromEntities(projectID string, entities []storage.Entity) int {
	count := 0
	for _, ent := range entities {
		if ent.Category == "character" || ent.Category == "" {
			v := SyncEntityVoice(ent)
			r.UpsertVoice(projectID, v)
			count++
		}
	}
	return count
}

// UpsertVoice registers or updates a character voice profile.
func (r *Registry) UpsertVoice(projectID string, voice CharacterVoice) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.voices[projectID]; !exists {
		r.voices[projectID] = make(map[string]CharacterVoice)
	}
	r.voices[projectID][voice.CharacterID] = voice
}

// ListVoices returns all character voice profiles for a project.
func (r *Registry) ListVoices(projectID string) []CharacterVoice {
	r.mu.RLock()
	defer r.mu.RUnlock()

	pVoices, exists := r.voices[projectID]
	if !exists {
		return nil
	}

	result := make([]CharacterVoice, 0, len(pVoices))
	for _, v := range pVoices {
		result = append(result, v)
	}
	return result
}

// MatchVoicesInText scans text for character names/aliases and returns matching voice profiles.
func (r *Registry) MatchVoicesInText(ctx context.Context, projectID string, text string) []CharacterVoice {
	r.mu.RLock()
	defer r.mu.RUnlock()

	pVoices, exists := r.voices[projectID]
	if !exists || len(pVoices) == 0 {
		return nil
	}

	lowerText := strings.ToLower(text)
	var matched []CharacterVoice

	for _, v := range pVoices {
		found := false
		if strings.Contains(lowerText, strings.ToLower(v.Name)) {
			found = true
		} else {
			for _, alias := range v.Aliases {
				if strings.TrimSpace(alias) != "" && strings.Contains(lowerText, strings.ToLower(alias)) {
					found = true
					break
				}
			}
		}

		if found {
			matched = append(matched, v)
		}
	}

	return matched
}

// FormatVoicePrompt constructs an LLM prompt block guiding character dialogue tone.
func FormatVoicePrompt(voices []CharacterVoice) string {
	if len(voices) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("[CHARACTER VOICE & DIALOGUE STYLE GUIDE]\n")
	sb.WriteString("When translating direct dialogue (in quotes or dialogue dashes), you MUST adhere to the distinct voice style of each character below:\n")

	for _, v := range voices {
		sb.WriteString(fmt.Sprintf("- Character: %s", v.Name))
		if len(v.Aliases) > 0 {
			sb.WriteString(fmt.Sprintf(" (Aliases: %s)", strings.Join(v.Aliases, ", ")))
		}
		sb.WriteString("\n")

		if v.VoiceTone != "" {
			sb.WriteString(fmt.Sprintf("  * Voice Tone / Cadence: %s\n", v.VoiceTone))
		}
		if len(v.DialogueRules) > 0 {
			sb.WriteString(fmt.Sprintf("  * Dialogue & Address Rules: %s\n", strings.Join(v.DialogueRules, "; ")))
		}
		if len(v.Catchphrases) > 0 {
			sb.WriteString(fmt.Sprintf("  * Catchphrases / Quirks: %s\n", strings.Join(v.Catchphrases, ", ")))
		}
	}

	return sb.String()
}
