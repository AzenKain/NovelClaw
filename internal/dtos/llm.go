package dtos

import (
	"novelclaw/pkg/storage"
)

// LLMConfigDTO represents an LLM provider configuration with masked sensitive tokens.
type LLMConfigDTO struct {
	ID             string            `json:"id"`
	ProviderName   string            `json:"provider_name"`
	ApiURL         string            `json:"api_url"`
	MaskedToken    string            `json:"masked_token"`
	HasToken       bool              `json:"has_token"`
	ModelName      string            `json:"model_name"`
	IsActive       bool              `json:"is_active"`
	IsDefault      bool              `json:"is_default"`
	RateLimitRPM   int64             `json:"rate_limit_rpm"`
	CustomHeaders  map[string]string `json:"custom_headers"`
	MaxRetries     int64             `json:"max_retries"`
	TimeoutSeconds int64             `json:"timeout_seconds"`
	ReasoningEffort string           `json:"reasoning_effort"`
}

// SaveLLMConfigRequest represents parameters for saving or updating an LLM configuration.
type SaveLLMConfigRequest struct {
	ID             string            `json:"id"`
	ProviderName   string            `json:"provider_name"`
	ApiURL         string            `json:"api_url"`
	Token          string            `json:"token"`
	ModelName      string            `json:"model_name"`
	IsActive       bool              `json:"is_active"`
	IsDefault      bool              `json:"is_default"`
	RateLimitRPM   int64             `json:"rate_limit_rpm"`
	CustomHeaders  map[string]string `json:"custom_headers"`
	MaxRetries     int64             `json:"max_retries"`
	TimeoutSeconds int64             `json:"timeout_seconds"`
	ReasoningEffort string           `json:"reasoning_effort"`
}

// MaskToken masks an API key token to prevent leaking secret keys over Wails IPC.
func MaskToken(token string) string {
	if token == "" {
		return ""
	}
	if len(token) <= 8 {
		return "********"
	}
	return token[:3] + "..." + token[len(token)-4:]
}

// ToLLMConfigDTO converts domain LLMConfig to LLMConfigDTO.
func ToLLMConfigDTO(cfg storage.LLMConfig) LLMConfigDTO {
	effort := cfg.ReasoningEffort
	if effort == "" {
		effort = "off"
	}
	return LLMConfigDTO{
		ID:             cfg.ID,
		ProviderName:   cfg.ProviderName,
		ApiURL:         cfg.ApiURL,
		MaskedToken:    MaskToken(cfg.Token),
		HasToken:       cfg.Token != "",
		ModelName:      cfg.ModelName,
		IsActive:       cfg.IsActive,
		IsDefault:      cfg.IsDefault,
		RateLimitRPM:   cfg.RateLimitRPM,
		CustomHeaders:  cfg.CustomHeaders,
		MaxRetries:     cfg.MaxRetries,
		TimeoutSeconds: cfg.TimeoutSeconds,
		ReasoningEffort: effort,
	}
}

// ToDomainLLMConfig converts SaveLLMConfigRequest to domain LLMConfig.
func (r SaveLLMConfigRequest) ToDomainLLMConfig() storage.LLMConfig {
	effort := r.ReasoningEffort
	if effort == "" {
		effort = "off"
	}
	return storage.LLMConfig{
		ID:             r.ID,
		ProviderName:   r.ProviderName,
		ApiURL:         r.ApiURL,
		Token:          r.Token,
		ModelName:      r.ModelName,
		IsActive:       r.IsActive,
		IsDefault:      r.IsDefault,
		RateLimitRPM:   r.RateLimitRPM,
		CustomHeaders:  r.CustomHeaders,
		MaxRetries:     r.MaxRetries,
		TimeoutSeconds: r.TimeoutSeconds,
		ReasoningEffort: effort,
	}
}

// TestConnectionRequest conveys credentials to test with an LLM provider endpoint.
type TestConnectionRequest struct {
	ApiURL    string `json:"api_url"`
	Token     string `json:"token"`
	ModelName string `json:"model_name"`
	ConfigID  string `json:"config_id,omitempty"`
}

// TestConnectionResponse reports the result and latency of the ping check.
type TestConnectionResponse struct {
	Success    bool   `json:"success"`
	LatencyMs  int64  `json:"latency_ms"`
	Message    string `json:"message"`
	ModelFound string `json:"model_found,omitempty"`
}

// PreviewStyleRequest carries text and style instructions to test LLM translation.
type PreviewStyleRequest struct {
	ProjectID   string `json:"project_id"`
	RawContent  string `json:"raw_content"`
	StylePrompt string `json:"style_prompt"`
	SourceLang  string `json:"source_lang"`
	TargetLang  string `json:"target_lang"`
}

// PreviewStyleResponse returns the real-time LLM translated preview snippet.
type PreviewStyleResponse struct {
	Success           bool   `json:"success"`
	TranslatedContent string `json:"translated_content"`
	ModelUsed         string `json:"model_used"`
	LatencyMs         int64  `json:"latency_ms"`
	Error             string `json:"error,omitempty"`
}

// AutoScoutStyleRequest requests autonomous literary style analysis across opening novel chapters.
type AutoScoutStyleRequest struct {
	ProjectID  string `json:"project_id"`
	SourceLang string `json:"source_lang"`
	TargetLang string `json:"target_lang"`
}

// AutoScoutStyleResponse returns the LLM's tailored literary style profile.
type AutoScoutStyleResponse struct {
	Success          bool   `json:"success"`
	StyleName        string `json:"style_name"`
	StyleDescription string `json:"style_description"`
	StylePrompt      string `json:"style_prompt"`
	AnalysisNotes    string `json:"analysis_notes"`
	SampleExcerpt    string `json:"sample_excerpt"`
	SampleTranslated string `json:"sample_translated"`
	ModelUsed        string `json:"model_used"`
	LatencyMs        int64  `json:"latency_ms"`
	ChaptersSampled  int    `json:"chapters_sampled"`
	Error            string `json:"error,omitempty"`
}

