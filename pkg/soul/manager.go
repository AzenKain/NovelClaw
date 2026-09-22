package soul

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/rs/zerolog/log"

	"novelclaw/pkg/paths"
)

var validIDPattern = regexp.MustCompile(`^[a-zA-Z0-9_\-]+$`)

// IsValidSoulID checks if an ID is valid and safe from path traversal.
func IsValidSoulID(id string) bool {
	return validIDPattern.MatchString(id) && !strings.Contains(id, "..") && !strings.ContainsAny(id, `/\`)
}

// Manager handles filesystem loading, caching, and persistence of Soul profiles in the souls/ directory.
type Manager struct {
	dir         string
	souls       map[string]Soul
	projectSoul map[string]string // projectID -> soulID
	mu          sync.RWMutex
}

// NewManager creates and initializes a Soul Manager for the specified directory.
func NewManager(dir string) *Manager {
	if dir == "" {
		dir = paths.Souls()
	}
	m := &Manager{
		dir:         dir,
		souls:       make(map[string]Soul),
		projectSoul: make(map[string]string),
	}
	m.Initialize()
	return m
}

// Initialize creates the souls directory if missing and populates it with default profiles.
func (m *Manager) Initialize() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err := os.MkdirAll(m.dir, 0755); err != nil {
		log.Warn().Err(err).Str("dir", m.dir).Msg("failed to create souls directory")
	}

	// Scan directory for *.soul.md files
	files, err := os.ReadDir(m.dir)
	if err == nil {
		for _, f := range files {
			if f.IsDir() || !strings.HasSuffix(f.Name(), ".soul.md") {
				continue
			}
			filePath := filepath.Join(m.dir, f.Name())
			content, readErr := os.ReadFile(filePath)
			if readErr != nil {
				log.Warn().Err(readErr).Str("file", filePath).Msg("read soul file failed")
				continue
			}
			s, parseErr := ParseSoulMarkdown(string(content))
			if parseErr != nil {
				log.Warn().Err(parseErr).Str("file", filePath).Msg("parse soul file failed")
				continue
			}
			if s.ID != "" && IsValidSoulID(s.ID) {
				m.souls[s.ID] = *s
			}
		}
	}

	// Ensure built-in default souls always exist
	defaultNeko := DefaultSoul()
	defaultMai := MaiSoul()

	if _, ok := m.souls[defaultNeko.ID]; !ok {
		m.souls[defaultNeko.ID] = defaultNeko
		if content, err := FormatSoulMarkdown(defaultNeko); err == nil {
			_ = os.WriteFile(filepath.Join(m.dir, fmt.Sprintf("%s.soul.md", defaultNeko.ID)), []byte(content), 0644)
		}
	}

	if _, ok := m.souls[defaultMai.ID]; !ok {
		m.souls[defaultMai.ID] = defaultMai
		if content, err := FormatSoulMarkdown(defaultMai); err == nil {
			_ = os.WriteFile(filepath.Join(m.dir, fmt.Sprintf("%s.soul.md", defaultMai.ID)), []byte(content), 0644)
		}
	}

}

// List returns all registered Souls sorted with default profiles first.
func (m *Manager) List() []Soul {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []Soul
	for _, s := range m.souls {
		result = append(result, s)
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].IsDefault && !result[j].IsDefault {
			return true
		}
		if !result[i].IsDefault && result[j].IsDefault {
			return false
		}
		return result[i].Name < result[j].Name
	})

	return result
}

// Get retrieves a soul by its unique identifier.
func (m *Manager) Get(id string) (Soul, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, ok := m.souls[id]
	return s, ok
}

// Save persists a soul to disk and registers it in memory.
func (m *Manager) Save(s Soul) error {
	s.ID = strings.TrimSpace(s.ID)
	s.Name = strings.TrimSpace(s.Name)

	if !IsValidSoulID(s.ID) {
		return fmt.Errorf("invalid soul ID '%s': must contain only alphanumeric characters, underscores, and dashes", s.ID)
	}
	if s.Name == "" {
		return fmt.Errorf("soul Name cannot be empty")
	}

	// If saving a built-in soul, enforce IsDefault = true
	if s.ID == "soul_neko_assistant" || s.ID == "soul_tieu_mai_wuxia" {
		s.IsDefault = true
	}

	content, err := FormatSoulMarkdown(s)
	if err != nil {
		return fmt.Errorf("format soul markdown: %w", err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	filePath := filepath.Clean(filepath.Join(m.dir, fmt.Sprintf("%s.soul.md", s.ID)))
	cleanDir := filepath.Clean(m.dir)
	if !strings.HasPrefix(filePath, cleanDir) {
		return fmt.Errorf("path traversal detected for soul ID '%s'", s.ID)
	}

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("write soul file: %w", err)
	}

	m.souls[s.ID] = s
	return nil
}

// Delete removes a custom soul from memory and disk.
func (m *Manager) Delete(id string) error {
	id = strings.TrimSpace(id)
	if !IsValidSoulID(id) {
		return fmt.Errorf("invalid soul ID '%s'", id)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	s, ok := m.souls[id]
	if !ok {
		return fmt.Errorf("soul '%s' not found", id)
	}
	if s.IsDefault || id == "soul_neko_assistant" || id == "soul_tieu_mai_wuxia" {
		return fmt.Errorf("cannot delete default built-in soul '%s'", id)
	}

	delete(m.souls, id)
	filePath := filepath.Clean(filepath.Join(m.dir, fmt.Sprintf("%s.soul.md", id)))
	if strings.HasPrefix(filePath, filepath.Clean(m.dir)) {
		_ = os.Remove(filePath)
	}
	return nil
}

// RestoreDefaults resets built-in default souls (Neko and Mai) to their original factory values.
func (m *Manager) RestoreDefaults() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	defaultNeko := DefaultSoul()
	defaultMai := MaiSoul()

	m.souls[defaultNeko.ID] = defaultNeko
	m.souls[defaultMai.ID] = defaultMai

	if content, err := FormatSoulMarkdown(defaultNeko); err == nil {
		_ = os.WriteFile(filepath.Join(m.dir, fmt.Sprintf("%s.soul.md", defaultNeko.ID)), []byte(content), 0644)
	}
	if content, err := FormatSoulMarkdown(defaultMai); err == nil {
		_ = os.WriteFile(filepath.Join(m.dir, fmt.Sprintf("%s.soul.md", defaultMai.ID)), []byte(content), 0644)
	}
	return nil
}

// SetProjectSoul associates a specific soul with a project.
func (m *Manager) SetProjectSoul(projectID, soulID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.souls[soulID]; !ok {
		return fmt.Errorf("soul '%s' does not exist", soulID)
	}
	m.projectSoul[projectID] = soulID
	return nil
}

// GetProjectSoul returns the active soul for a project or the default soul.
func (m *Manager) GetProjectSoul(projectID string) Soul {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if soulID, ok := m.projectSoul[projectID]; ok {
		if s, exists := m.souls[soulID]; exists {
			return s
		}
	}
	if defaultSoul, ok := m.souls["soul_neko_assistant"]; ok {
		return defaultSoul
	}
	return DefaultSoul()
}
