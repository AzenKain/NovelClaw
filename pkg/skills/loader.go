package skills

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"novelclaw/pkg/llm"
	"novelclaw/pkg/paths"
)

// SkillMetadata holds frontmatter metadata parsed from a *.skill.md file.
type SkillMetadata struct {
	ID          string   `yaml:"id" json:"id"`
	Name        string   `yaml:"name" json:"name"`
	Category    string   `yaml:"category" json:"category"`
	Description string   `yaml:"description,omitempty" json:"description,omitempty"`
	Tools       []string `yaml:"tools" json:"tools"`
}

// Skill represents a fully parsed skill document.
type Skill struct {
	Metadata SkillMetadata `json:"metadata"`
	Playbook string        `json:"playbook"`
	FilePath string        `json:"file_path"`
}

// SkillLoader loads, caches, and routes skills from disk.
type SkillLoader struct {
	skillsDir string
	skills    []Skill
}

// NewSkillLoader creates a SkillLoader pointing to a skills root directory.
// An empty or default value resolves to the canonical read-only baseline:
// Windows → ./skills/default, mac/Linux → ~/.novelclaw/skills/default.
func NewSkillLoader(skillsDir string) *SkillLoader {
	if skillsDir == "" || skillsDir == "skills/default" || skillsDir == "./skills/default" {
		skillsDir = paths.SkillsDefault()
	}
	return &SkillLoader{
		skillsDir: skillsDir,
	}
}

// LoadAll reads all *.skill.md files in the configured directory.
func (l *SkillLoader) LoadAll() ([]Skill, error) {
	entries, err := os.ReadDir(l.skillsDir)
	if err != nil {
		return nil, fmt.Errorf("read skills dir: %w", err)
	}

	var loaded []Skill
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".skill.md") {
			continue
		}
		path := filepath.Join(l.skillsDir, entry.Name())
		skill, err := ParseSkillFile(path)
		if err != nil {
			return nil, fmt.Errorf("parse skill %s: %w", entry.Name(), err)
		}
		loaded = append(loaded, skill)
	}

	l.skills = loaded
	return loaded, nil
}

// ParseSkillFile parses a single *.skill.md file into frontmatter metadata and body markdown.
func ParseSkillFile(path string) (Skill, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Skill{}, err
	}

	content := string(data)
	if !strings.HasPrefix(content, "---") {
		return Skill{}, fmt.Errorf("invalid skill format: missing frontmatter delimiter '---'")
	}

	parts := strings.SplitN(content[3:], "---", 2)
	if len(parts) < 2 {
		return Skill{}, fmt.Errorf("invalid skill format: unclosed frontmatter delimiter")
	}

	var meta SkillMetadata
	if err := yaml.Unmarshal([]byte(parts[0]), &meta); err != nil {
		return Skill{}, fmt.Errorf("unmarshal frontmatter YAML: %w", err)
	}

	playbook := strings.TrimSpace(parts[1])
	return Skill{
		Metadata: meta,
		Playbook: playbook,
		FilePath: path,
	}, nil
}

// MatchSkillsByIntent inspects a user query and returns relevant skills to include in the prompt.
// If query is broad or empty, it returns all loaded skills.
func (l *SkillLoader) MatchSkillsByIntent(query string) []Skill {
	if len(l.skills) == 0 {
		_, _ = l.LoadAll()
	}
	if len(l.skills) == 0 {
		return nil
	}

	q := strings.ToLower(query)
	if strings.TrimSpace(q) == "" {
		return l.skills
	}

	var matched []Skill
	for _, s := range l.skills {
		match := false
		switch s.Metadata.Category {
		// NOTE: the Vietnamese keywords below are intentional input-matching data so that
		// users may phrase requests in either English or Vietnamese.
		case "ui_control":
			match = strings.Contains(q, "tab") || strings.Contains(q, "switch") || strings.Contains(q, "open") || strings.Contains(q, "scroll") || strings.Contains(q, "chapter") || strings.Contains(q, "volume") || strings.Contains(q, "vol") || strings.Contains(q, "chuyển") || strings.Contains(q, "mở") || strings.Contains(q, "cuộn") || strings.Contains(q, "chương") || strings.Contains(q, "tập")
		case "character_relations":
			match = strings.Contains(q, "character") || strings.Contains(q, "relation") || strings.Contains(q, "address") || strings.Contains(q, "master") || strings.Contains(q, "nhân vật") || strings.Contains(q, "quan hệ") || strings.Contains(q, "xưng hô") || strings.Contains(q, "sư phụ") || strings.Contains(q, "huynh") || strings.Contains(q, "muội") || strings.Contains(q, "gọi là")
		case "world_bible":
			match = strings.Contains(q, "world") || strings.Contains(q, "lore") || strings.Contains(q, "faction") || strings.Contains(q, "realm") || strings.Contains(q, "magic") || strings.Contains(q, "scan") || strings.Contains(q, "thế giới") || strings.Contains(q, "tập đoàn") || strings.Contains(q, "tông môn") || strings.Contains(q, "cảnh giới") || strings.Contains(q, "ma pháp") || strings.Contains(q, "quét")
		case "terminology":
			match = strings.Contains(q, "glossary") || strings.Contains(q, "term") || strings.Contains(q, "translate as") || strings.Contains(q, "=") || strings.Contains(q, "từ điển") || strings.Contains(q, "thuật ngữ") || strings.Contains(q, "dịch là")
		case "pipeline_config":
			match = strings.Contains(q, "model") || strings.Contains(q, "settings") || strings.Contains(q, "mode") || strings.Contains(q, "dual-agent") || strings.Contains(q, "hot patch") || strings.Contains(q, "r19") || strings.Contains(q, "soul") || strings.Contains(q, "cài đặt") || strings.Contains(q, "chế độ")
		case "execution":
			match = strings.Contains(q, "translate") || strings.Contains(q, "start") || strings.Contains(q, "stop") || strings.Contains(q, "pause") || strings.Contains(q, "resume") || strings.Contains(q, "abort") || strings.Contains(q, "rollback") || strings.Contains(q, "undo") || strings.Contains(q, "dịch") || strings.Contains(q, "bắt đầu") || strings.Contains(q, "dừng") || strings.Contains(q, "tạm dừng") || strings.Contains(q, "tiếp tục") || strings.Contains(q, "hủy") || strings.Contains(q, "hoàn tác")
		case "publishing":
			match = strings.Contains(q, "export") || strings.Contains(q, "epub") || strings.Contains(q, "pdf") || strings.Contains(q, "learn") || strings.Contains(q, "benchmark") || strings.Contains(q, "xuất") || strings.Contains(q, "học")
		}

		if match {
			matched = append(matched, s)
		}
	}

	// Fallback to all skills if no specific keyword matched
	if len(matched) == 0 {
		return l.skills
	}
	return matched
}

// BuildSkillsPrompt returns formatted playbooks of the matched skills to inject into the system prompt.
func BuildSkillsPrompt(skills []Skill) string {
	if len(skills) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString("[ACTIVE AGENT SKILLS & CAPABILITY PLAYBOOKS]:\n\n")
	for _, s := range skills {
		sb.WriteString(fmt.Sprintf("### Skill: %s (%s)\n", s.Metadata.Name, s.Metadata.ID))
		sb.WriteString(fmt.Sprintf("Tools Available: %s\n\n", strings.Join(s.Metadata.Tools, ", ")))
		sb.WriteString(s.Playbook)
		sb.WriteString("\n\n---\n\n")
	}
	return sb.String()
}

// FilterToolsBySkills filters available tools matching only the tools declared by the active skills.
func FilterToolsBySkills(activeSkills []Skill, allTools []llm.ToolDefinition) []llm.ToolDefinition {
	if len(activeSkills) == 0 {
		return allTools
	}

	allowedNames := make(map[string]bool)
	for _, s := range activeSkills {
		for _, t := range s.Metadata.Tools {
			allowedNames[t] = true
		}
	}

	var filtered []llm.ToolDefinition
	for _, tool := range allTools {
		if allowedNames[tool.Function.Name] {
			filtered = append(filtered, tool)
		}
	}
	return filtered
}
