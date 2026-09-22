package storage

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"

	"novelclaw/internal/gen/sqlc"
	"novelclaw/pkg/jsonx"
	"novelclaw/pkg/security"
)

// LLMConfig represents LLM provider configuration with an encrypted token at rest.
type LLMConfig struct {
	ID              string            `json:"id"`
	ProviderName    string            `json:"provider_name"`
	ApiURL          string            `json:"api_url"`
	Token           string            `json:"token"`
	ModelName       string            `json:"model_name"`
	IsActive        bool              `json:"is_active"`
	IsDefault       bool              `json:"is_default"`
	RateLimitRPM    int64             `json:"rate_limit_rpm"`
	CustomHeaders   map[string]string `json:"custom_headers"`
	MaxRetries      int64             `json:"max_retries"`
	TimeoutSeconds  int64             `json:"timeout_seconds"`
	ReasoningEffort string            `json:"reasoning_effort"`
}

// SaveLLMConfig saves LLM configuration and encrypts the token using AES-256-GCM.
func (s *Storage) SaveLLMConfig(ctx context.Context, cfg LLMConfig) error {
	if cfg.ID == "" {
		cfg.ID = "default"
	}
	if cfg.RateLimitRPM <= 0 {
		cfg.RateLimitRPM = 60
	}
	if cfg.MaxRetries <= 0 {
		cfg.MaxRetries = 3
	}
	if cfg.TimeoutSeconds <= 0 {
		cfg.TimeoutSeconds = 120
	}
	if cfg.ReasoningEffort == "" {
		cfg.ReasoningEffort = "off"
	}

	encryptedToken, err := security.Encrypt(cfg.Token)
	if err != nil {
		return fmt.Errorf("encrypt api token: %w", err)
	}

	headersJSON := "{}"
	if cfg.CustomHeaders != nil {
		b, err := jsonx.Marshal(cfg.CustomHeaders)
		if err != nil {
			return fmt.Errorf("marshal custom headers: %w", err)
		}
		headersJSON = string(b)
	}

	var isActive int64 = 0
	if cfg.IsActive {
		isActive = 1
	}

	var isDefault int64 = 0
	if cfg.IsDefault {
		isDefault = 1
		if _, err := s.db.ExecContext(ctx, "UPDATE llm_configs SET is_default = 0 WHERE id != ?", cfg.ID); err != nil {
			return fmt.Errorf("reset previous default configs: %w", err)
		}
	}

	return s.q.UpsertLLMConfig(ctx, sqlc.UpsertLLMConfigParams{
		ID:                cfg.ID,
		ProviderName:      cfg.ProviderName,
		ApiUrl:            cfg.ApiURL,
		EncryptedToken:    encryptedToken,
		ModelName:         cfg.ModelName,
		IsActive:          isActive,
		IsDefault:         isDefault,
		RateLimitRpm:      cfg.RateLimitRPM,
		CustomHeadersJson: headersJSON,
		MaxRetries:        cfg.MaxRetries,
		TimeoutSeconds:    cfg.TimeoutSeconds,
		ReasoningEffort:   cfg.ReasoningEffort,
	})
}

// GetDefaultLLMConfig returns the default LLM configuration with decrypted token.
func (s *Storage) GetDefaultLLMConfig(ctx context.Context) (*LLMConfig, error) {
	row, err := s.q.GetDefaultLLMConfig(ctx)
	if err != nil {
		return nil, err
	}
	cfg, legacy, err := toDomainLLMConfigWithKeyID(row)
	if err != nil {
		return nil, err
	}
	if legacy {
		_ = s.reencryptLLMConfig(ctx, *cfg)
	}
	return cfg, nil
}

// GetLLMConfigByID returns an LLM configuration by ID with decrypted token.
func (s *Storage) GetLLMConfigByID(ctx context.Context, id string) (*LLMConfig, error) {
	row, err := s.q.GetLLMConfigByID(ctx, id)
	if err != nil {
		return nil, err
	}
	cfg, legacy, err := toDomainLLMConfigWithKeyID(row)
	if err != nil {
		return nil, err
	}
	if legacy {
		_ = s.reencryptLLMConfig(ctx, *cfg)
	}
	return cfg, nil
}

// ListActiveLLMConfigs returns all active LLM configurations.
//
// A single undecryptable row must never hide every other saved provider:
// each row is decoded independently, failures are logged and skipped, and
// rows that still decrypt under a legacy (pre-rebrand) vault key are
// transparently re-encrypted with the primary key so the list self-heals.
func (s *Storage) ListActiveLLMConfigs(ctx context.Context) ([]LLMConfig, error) {
	rows, err := s.q.ListActiveLLMConfigs(ctx)
	if err != nil {
		return nil, err
	}

	out := make([]LLMConfig, 0, len(rows))
	for _, r := range rows {
		cfg, legacy, err := toDomainLLMConfigWithKeyID(r)
		if err != nil {
			log.Warn().
				Err(err).
				Str("config_id", r.ID).
				Str("provider", r.ProviderName).
				Msg("skipping LLM config with undecryptable token")
			continue
		}
		if legacy {
			// Re-encrypt under the primary key so future runs need no fallback.
			if healed := s.reencryptLLMConfig(ctx, *cfg); healed != nil {
				log.Info().
					Str("config_id", r.ID).
					Msg("re-encrypted LLM token with primary vault key")
			} else {
				log.Warn().Str("config_id", r.ID).Msg("could not re-encrypt legacy LLM token")
			}
		}
		out = append(out, *cfg)
	}
	return out, nil
}

// reencryptLLMConfig rewrites a config row with a freshly encrypted token.
func (s *Storage) reencryptLLMConfig(ctx context.Context, cfg LLMConfig) error {
	if err := s.SaveLLMConfig(ctx, cfg); err != nil {
		return err
	}
	return nil
}

// DeleteLLMConfig deletes an LLM configuration by ID.
func (s *Storage) DeleteLLMConfig(ctx context.Context, id string) error {
	return s.q.DeleteLLMConfig(ctx, id)
}

// toDomainLLMConfig converts a database row to domain LLMConfig.
func toDomainLLMConfig(row sqlc.LlmConfig) (*LLMConfig, error) {
	cfg, _, err := toDomainLLMConfigWithKeyID(row)
	return cfg, err
}

// toDomainLLMConfigWithKeyID converts a row and reports whether the token
// had to be decrypted with a legacy (pre-rebrand) vault key.
func toDomainLLMConfigWithKeyID(row sqlc.LlmConfig) (*LLMConfig, bool, error) {
	token, legacy, err := security.DecryptWithKeyID(row.EncryptedToken)
	if err != nil {
		return nil, false, fmt.Errorf("decrypt token for provider %s: %w", row.ID, err)
	}

	var headers map[string]string
	if row.CustomHeadersJson != "" && row.CustomHeadersJson != "{}" {
		_ = jsonx.Unmarshal([]byte(row.CustomHeadersJson), &headers)
	}

	maxRetries := row.MaxRetries
	if maxRetries <= 0 {
		maxRetries = 3
	}

	timeoutSeconds := row.TimeoutSeconds
	if timeoutSeconds <= 0 {
		timeoutSeconds = 120
	}

	effort := row.ReasoningEffort
	if effort == "" {
		effort = "off"
	}

	return &LLMConfig{
		ID:              row.ID,
		ProviderName:    row.ProviderName,
		ApiURL:          row.ApiUrl,
		Token:           token,
		ModelName:       row.ModelName,
		IsActive:        row.IsActive == 1,
		IsDefault:       row.IsDefault == 1,
		RateLimitRPM:    row.RateLimitRpm,
		CustomHeaders:   headers,
		MaxRetries:      maxRetries,
		TimeoutSeconds:  timeoutSeconds,
		ReasoningEffort: effort,
	}, legacy, nil
}
