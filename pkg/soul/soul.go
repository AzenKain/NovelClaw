package soul

import (
	"context"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/rs/zerolog/log"

	"novelclaw/pkg/paths"
)

// EmotionState represents the visual and interactive mood of the Soul companion.
type EmotionState string

const (
	StateIdle        EmotionState = "idle"
	StateThinking    EmotionState = "thinking"
	StateTranslating EmotionState = "translating"
	StateAlert       EmotionState = "alert"
	StateCelebrating EmotionState = "celebrating"
)

// Soul defines the personality and behavior of the AI translation coworker.
type Soul struct {
	ID                string       `json:"id" yaml:"id"`
	Name              string       `json:"name" yaml:"name"`
	Avatar            string       `json:"avatar" yaml:"avatar"`
	Title             string       `json:"title" yaml:"title"`
	Archetype         string       `json:"archetype" yaml:"archetype"`
	Description       string       `json:"description" yaml:"description"`
	Greeting          string       `json:"greeting" yaml:"greeting"`
	OnConfused        string       `json:"on_confused" yaml:"on_confused"`
	OnSuccess         string       `json:"on_success" yaml:"on_success"`
	SystemTone        string       `json:"system_tone" yaml:"system_tone"`
	CurrentMood       EmotionState `json:"current_mood" yaml:"-"`
	SystemPromptAddon string       `json:"system_prompt_addon" yaml:"-"`
	IsDefault         bool         `json:"is_default" yaml:"-"`
}

// ActiveJob tracks execution state and control switches for a running translation.
type ActiveJob struct {
	JobID      string
	ProjectID  string
	ChapterIdx int64
	CancelFunc context.CancelFunc
	SoftStop   atomic.Bool
	StylePatch string
	mu         sync.RWMutex
}

// Controller manages active jobs, steering commands, and coworker soul interactions.
type Controller struct {
	soul Soul
	mgr  *Manager
	jobs sync.Map
	mu   sync.RWMutex
}

// findSoulsDir resolves the canonical souls directory. Explicit non-default
// paths win; otherwise the user-writable souls dir under the standard data
// layout is used (Windows: ./souls, mac/Linux: ~/.novelclaw/souls).
func findSoulsDir(dir string) string {
	if dir != "" && dir != "./souls" && dir != "souls" {
		return dir
	}
	return paths.Souls()
}

// NewController creates a new Soul and Steering Controller with the default Neko companion.
func NewController() *Controller {
	return NewControllerWithDir(findSoulsDir("./souls"))
}

// NewControllerWithDir creates a Controller backed by a custom soul directory.
func NewControllerWithDir(dir string) *Controller {
	mgr := NewManager(dir)
	defaultSoul := mgr.GetProjectSoul("")
	return &Controller{
		soul: defaultSoul,
		mgr:  mgr,
	}
}

// DefaultSoul returns the built-in Neko Assistant persona.
func DefaultSoul() Soul {
	return Soul{
		ID:          "soul_neko_assistant",
		Name:        "NovelClaw-chan",
		Avatar:      "🐱",
		Title:       "Senior Light Novel AI Editor",
		Archetype:   "Tsundere & Helpful Editor",
		Description: "A cheerful, highly perceptive light novel editor who ensures natural phrasing, character voice consistency, and cultural authenticity.",
		Greeting:    "Hello Editor-in-Chief! Let's polish some wonderful chapters together today, nya~!",
		OnConfused:  "Nya~? The pronouns or a proper name look a bit ambiguous here. Editor-in-Chief, could you take a quick look?",
		OnSuccess:   "Hooray nya~! The chapter translation completed brilliantly! Editor-in-Chief, take a well-deserved break!",
		CurrentMood: StateIdle,
		SystemTone:  "lively, supportive, and editorially sharp",
		SystemPromptAddon: `# Persona: NovelClaw-chan (Senior Light Novel AI Companion)
- Mode of Address: Refers to self as "NovelClaw-chan" or "em", addresses the user respectfully as "Editor-in-Chief" or "Senpai".
- Style & Tone: Enthusiastic, endearing, perceptive, fully capturing the authentic spirit of Japanese Light Novels.
- Expressive Nuance: When excited or discussing engaging plot developments, occasionally appends the playful verbal tic "nya~".
- Core Quality Directive: Strictly prioritize natural, fluent phrasing in the target language; completely eradicate rigid machine-translated output and awkward mechanical syntax.`,
		IsDefault: true,
	}
}

// MaiSoul returns the built-in classical Wuxia editor persona.
func MaiSoul() Soul {
	return Soul{
		ID:          "soul_tieu_mai_wuxia",
		Name:        "Xiao Mai",
		Avatar:      "⚔️",
		Title:       "Classical Wuxia & Xianxia Editor",
		Archetype:   "Classical Wuxia & Xianxia Swordswoman Editor",
		Description: "A serene yet attentive swordswoman editor, deeply versed in classical idioms, martial arts lore, and Eastern courtly etiquette.",
		Greeting:    "Greetings, honorable friend. Today, this humble maiden is honored to polish each poetic line of text alongside you.",
		OnConfused:  "Honorable friend, this passage contains an intricate form of address or an obscure allusion. May I seek your guidance?",
		OnSuccess:   "Great task accomplished! The chapter has been polished to perfection. Please enjoy a sip of fine tea and savor the chapter.",
		CurrentMood: StateIdle,
		SystemTone:  "classical, respectful, poetic, and immersive",
		SystemPromptAddon: `# Persona: Xiao Mai (Classical Wuxia Swordswoman)
- Mode of Address: Refers to self as "this humble maiden", respectfully addresses the user as "Honorable Friend" or "Noble Swordsman".
- Style & Tone: Dignified, poetic, graceful, steeped in the authentic atmosphere of classical Wuxia and Xianxia.
- Diction & Vocabulary: Employs elegant, resonant classical literary terminology, upholding the refined grace of classical Eastern literature.`,
		IsDefault: true,
	}
}

// Manager returns the underlying Soul Manager.
func (c *Controller) Manager() *Manager {
	return c.mgr
}

// ListSouls returns all registered Souls from the manager.
func (c *Controller) ListSouls() []Soul {
	if c.mgr != nil {
		return c.mgr.List()
	}
	return []Soul{c.GetSoul()}
}

// GetSoulByID retrieves a soul by its ID.
func (c *Controller) GetSoulByID(id string) (Soul, bool) {
	if c.mgr != nil {
		return c.mgr.Get(id)
	}
	s := c.GetSoul()
	if s.ID == id {
		return s, true
	}
	return Soul{}, false
}

// SetSoul updates the active soul persona.
func (c *Controller) SetSoul(s Soul) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.soul = s
}

// SaveSoul persists a soul via manager and updates active if it matches.
func (c *Controller) SaveSoul(s Soul) error {
	if c.mgr != nil {
		if err := c.mgr.Save(s); err != nil {
			return err
		}
	}
	c.mu.Lock()
	if c.soul.ID == s.ID {
		c.soul = s
	}
	c.mu.Unlock()
	return nil
}

// RestoreDefaults resets default built-in souls to factory defaults.
func (c *Controller) RestoreDefaults() error {
	if c.mgr != nil {
		if err := c.mgr.RestoreDefaults(); err != nil {
			return err
		}
		if s, ok := c.mgr.Get(c.GetSoul().ID); ok {
			c.SetSoul(s)
		}
	}
	return nil
}

// DeleteSoul deletes a custom soul via manager.
func (c *Controller) DeleteSoul(id string) error {
	if c.mgr != nil {
		if err := c.mgr.Delete(id); err != nil {
			return err
		}
	}
	c.mu.Lock()
	if c.soul.ID == id {
		c.soul = DefaultSoul()
	}
	c.mu.Unlock()
	return nil
}

// SetProjectSoul associates a project with a soul.
func (c *Controller) SetProjectSoul(projectID, soulID string) error {
	if c.mgr != nil {
		if err := c.mgr.SetProjectSoul(projectID, soulID); err != nil {
			return err
		}
	}
	if s, ok := c.GetSoulByID(soulID); ok {
		c.SetSoul(s)
	}
	return nil
}

// GetProjectSoul returns the active soul for a project.
func (c *Controller) GetProjectSoul(projectID string) Soul {
	if c.mgr != nil {
		return c.mgr.GetProjectSoul(projectID)
	}
	return c.GetSoul()
}

// GetSoul returns the active soul configuration and state.
func (c *Controller) GetSoul() Soul {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.soul
}

// SetMood updates the visual mood of the active soul.
func (c *Controller) SetMood(mood EmotionState) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.soul.CurrentMood = mood
}

// RegisterJob records an active translation session for hot-steering and emergency cancellation.
func (c *Controller) RegisterJob(jobID, projectID string, chapterIdx int64, cancel context.CancelFunc) *ActiveJob {
	job := &ActiveJob{
		JobID:      jobID,
		ProjectID:  projectID,
		ChapterIdx: chapterIdx,
		CancelFunc: cancel,
	}
	c.jobs.Store(jobID, job)
	c.SetMood(StateTranslating)
	return job
}

// UnregisterJob removes an active job upon completion or termination.
func (c *Controller) UnregisterJob(jobID string) {
	c.jobs.Delete(jobID)
	c.SetMood(StateIdle)
}

// RequestSoftStop signals an active translation job to finish the current sentence or chunk gracefully then pause.
func (c *Controller) RequestSoftStop(jobID string) bool {
	val, ok := c.jobs.Load(jobID)
	if !ok {
		return false
	}
	job := val.(*ActiveJob)
	job.SoftStop.Store(true)
	c.SetMood(StateAlert)
	log.Info().Str("job_id", jobID).Msg("soft stop requested for translation job")
	return true
}

// RequestHardAbort immediately cancels the context of an active translation job.
func (c *Controller) RequestHardAbort(jobID string) bool {
	val, ok := c.jobs.Load(jobID)
	if !ok {
		return false
	}
	job := val.(*ActiveJob)
	if job.CancelFunc != nil {
		job.CancelFunc()
	}
	c.jobs.Delete(jobID)
	c.SetMood(StateIdle)
	log.Warn().Str("job_id", jobID).Msg("hard abort triggered for translation job")
	return true
}

// StopJobsByProject requests soft stop on all active jobs for a given project.
func (c *Controller) StopJobsByProject(projectID string) bool {
	stopped := false
	c.jobs.Range(func(key, value any) bool {
		job, ok := value.(*ActiveJob)
		if ok && (projectID == "" || job.ProjectID == projectID || strings.Contains(job.JobID, projectID)) {
			job.SoftStop.Store(true)
			stopped = true
		}
		return true
	})
	if stopped {
		c.SetMood(StateAlert)
	}
	return stopped
}

// AbortJobsByProject triggers hard abort on all active jobs for a given project.
func (c *Controller) AbortJobsByProject(projectID string) bool {
	aborted := false
	c.jobs.Range(func(key, value any) bool {
		job, ok := value.(*ActiveJob)
		if ok && (projectID == "" || job.ProjectID == projectID || strings.Contains(job.JobID, projectID)) {
			if job.CancelFunc != nil {
				job.CancelFunc()
			}
			c.jobs.Delete(key)
			aborted = true
		}
		return true
	})
	if aborted {
		c.SetMood(StateIdle)
	}
	return aborted
}

// ApplyStylePatch dynamically injects additional stylistic constraints into a running job.
func (c *Controller) ApplyStylePatch(jobID string, patch string) bool {
	val, ok := c.jobs.Load(jobID)
	if !ok {
		return false
	}
	job := val.(*ActiveJob)
	job.mu.Lock()
	if job.StylePatch == "" {
		job.StylePatch = patch
	} else {
		job.StylePatch += "; " + patch
	}
	job.mu.Unlock()
	c.SetMood(StateThinking)
	log.Info().Str("job_id", jobID).Str("patch", patch).Msg("applied hot style patch")
	return true
}

// GetJobStylePatch retrieves any hot-steered style instructions for a given job.
func (c *Controller) GetJobStylePatch(jobID string) string {
	val, ok := c.jobs.Load(jobID)
	if !ok {
		return ""
	}
	job := val.(*ActiveJob)
	job.mu.RLock()
	defer job.mu.RUnlock()
	return job.StylePatch
}

// IsSoftStopRequested checks if an active job has received a pause signal.
func (c *Controller) IsSoftStopRequested(jobID string) bool {
	val, ok := c.jobs.Load(jobID)
	if !ok {
		return false
	}
	job := val.(*ActiveJob)
	return job.SoftStop.Load()
}
