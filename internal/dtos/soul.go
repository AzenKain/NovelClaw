package dtos

// SoulDTO represents the AI coworker companion persona and emotional state.
type SoulDTO struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	Avatar            string `json:"avatar"`
	Title             string `json:"title"`
	Archetype         string `json:"archetype"`
	Description       string `json:"description"`
	Greeting          string `json:"greeting"`
	OnConfused        string `json:"on_confused"`
	OnSuccess         string `json:"on_success"`
	SystemTone        string `json:"system_tone"`
	CurrentMood       string `json:"current_mood"`
	SystemPromptAddon string `json:"system_prompt_addon"`
	IsDefault         bool   `json:"is_default"`
}

// SteerActionRequest conveys emergency control or dynamic style overrides from the UI.
type SteerActionRequest struct {
	JobID      string `json:"job_id"`
	ProjectID  string `json:"project_id,omitempty"`
	Action     string `json:"action"` // "soft_stop", "hard_abort", "style_patch"
	PatchValue string `json:"patch_value,omitempty"`
}

// SteerActionResultDTO provides immediate feedback for a steering or abort command.
type SteerActionResultDTO struct {
	Success     bool   `json:"success"`
	Message     string `json:"message"`
	CurrentMood string `json:"current_mood"`
}
