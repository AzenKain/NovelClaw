package soul

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// soulFrontmatter defines the YAML frontmatter serialized at the top of *.soul.md files.
type soulFrontmatter struct {
	ID          string `yaml:"id"`
	Name        string `yaml:"name"`
	Avatar      string `yaml:"avatar"`
	Title       string `yaml:"title"`
	Archetype   string `yaml:"archetype"`
	Description string `yaml:"description"`
	Greeting    string `yaml:"greeting"`
	OnConfused  string `yaml:"on_confused"`
	OnSuccess   string `yaml:"on_success"`
	SystemTone  string `yaml:"system_tone"`
	IsDefault   bool   `yaml:"is_default"`
}

// ParseSoulMarkdown parses a *.soul.md document containing YAML frontmatter and Markdown body.
func ParseSoulMarkdown(content string) (*Soul, error) {
	trimmed := strings.TrimSpace(content)
	if !strings.HasPrefix(trimmed, "---") {
		return nil, fmt.Errorf("soul markdown must begin with YAML frontmatter delimiter '---'")
	}

	rest := strings.TrimPrefix(trimmed, "---")
	parts := strings.SplitN(rest, "---", 2)
	if len(parts) < 2 {
		return nil, fmt.Errorf("soul markdown missing closing YAML frontmatter delimiter '---'")
	}

	yamlPart := strings.TrimSpace(parts[0])
	bodyPart := strings.TrimSpace(parts[1])

	var fm soulFrontmatter
	if err := yaml.Unmarshal([]byte(yamlPart), &fm); err != nil {
		return nil, fmt.Errorf("unmarshal soul yaml frontmatter: %w", err)
	}

	s := &Soul{
		ID:                fm.ID,
		Name:              fm.Name,
		Avatar:            fm.Avatar,
		Title:             fm.Title,
		Archetype:         fm.Archetype,
		Description:       fm.Description,
		Greeting:          fm.Greeting,
		OnConfused:        fm.OnConfused,
		OnSuccess:         fm.OnSuccess,
		SystemTone:        fm.SystemTone,
		CurrentMood:       StateIdle,
		SystemPromptAddon: bodyPart,
		IsDefault:         fm.IsDefault,
	}

	return s, nil
}

// FormatSoulMarkdown serializes a Soul struct into a standard *.soul.md document.
func FormatSoulMarkdown(s Soul) (string, error) {
	fm := soulFrontmatter{
		ID:          s.ID,
		Name:        s.Name,
		Avatar:      s.Avatar,
		Title:       s.Title,
		Archetype:   s.Archetype,
		Description: s.Description,
		Greeting:    s.Greeting,
		OnConfused:  s.OnConfused,
		OnSuccess:   s.OnSuccess,
		SystemTone:  s.SystemTone,
		IsDefault:   s.IsDefault,
	}

	frontmatterBytes, err := yaml.Marshal(fm)
	if err != nil {
		return "", fmt.Errorf("marshal soul frontmatter: %w", err)
	}

	body := strings.TrimSpace(s.SystemPromptAddon)
	if body == "" {
		body = fmt.Sprintf("# Persona: %s\n\n%s", s.Name, s.Description)
	}

	return fmt.Sprintf("---\n%s---\n\n%s\n", strings.TrimSpace(string(frontmatterBytes)), body), nil
}
