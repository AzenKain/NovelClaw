package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"novelclaw/internal/dtos"
	"novelclaw/pkg/llm"
	"novelclaw/pkg/storage"
)

// LLMService exposes LLM provider configuration, connection testing, and credential management to Wails v3 frontend.
type LLMService struct {
	store *storage.Storage
}

// NewLLMService creates a new LLMService.
func NewLLMService(store *storage.Storage) *LLMService {
	return &LLMService{store: store}
}

// SaveConfig saves an LLM provider configuration with AES-256-GCM encryption from a DTO request.
func (s *LLMService) SaveConfig(ctx context.Context, req dtos.SaveLLMConfigRequest) error {
	cfg := req.ToDomainLLMConfig()
	if cfg.ID != "" && (cfg.Token == "" || strings.Contains(cfg.Token, "...")) {
		if existing, err := s.store.GetLLMConfigByID(ctx, cfg.ID); err == nil && existing != nil {
			cfg.Token = existing.Token
		}
	}
	return s.store.SaveLLMConfig(ctx, cfg)
}

// SetDefaultConfig marks a specific LLM configuration as the default for translation runs.
func (s *LLMService) SetDefaultConfig(ctx context.Context, id string) error {
	cfg, err := s.store.GetLLMConfigByID(ctx, id)
	if err != nil {
		return fmt.Errorf("find llm config: %w", err)
	}
	cfg.IsDefault = true
	cfg.IsActive = true
	return s.store.SaveLLMConfig(ctx, *cfg)
}

// GetDefaultConfig returns the active default LLM provider configuration with masked token as a DTO.
func (s *LLMService) GetDefaultConfig(ctx context.Context) (*dtos.LLMConfigDTO, error) {
	cfg, err := s.store.GetDefaultLLMConfig(ctx)
	if err != nil {
		return nil, err
	}
	dto := dtos.ToLLMConfigDTO(*cfg)
	return &dto, nil
}

// ListConfigs lists all active LLM configurations with masked tokens as DTOs.
func (s *LLMService) ListConfigs(ctx context.Context) ([]dtos.LLMConfigDTO, error) {
	configs, err := s.store.ListActiveLLMConfigs(ctx)
	if err != nil {
		return nil, err
	}
	dtosList := make([]dtos.LLMConfigDTO, 0, len(configs))
	for _, cfg := range configs {
		dtosList = append(dtosList, dtos.ToLLMConfigDTO(cfg))
	}
	return dtosList, nil
}

// DeleteConfig removes an LLM provider configuration by ID.
func (s *LLMService) DeleteConfig(ctx context.Context, id string) error {
	return s.store.DeleteLLMConfig(ctx, id)
}

// TestConnection sends a lightweight probe to verify connectivity and API key validity with the provider.
func (s *LLMService) TestConnection(ctx context.Context, req dtos.TestConnectionRequest) (*dtos.TestConnectionResponse, error) {
	token := req.Token
	apiURL := req.ApiURL
	modelName := req.ModelName

	if req.ConfigID != "" && (token == "" || strings.Contains(token, "...")) {
		existing, err := s.store.GetLLMConfigByID(ctx, req.ConfigID)
		if err == nil && existing != nil {
			token = existing.Token
			if apiURL == "" {
				apiURL = existing.ApiURL
			}
			if modelName == "" {
				modelName = existing.ModelName
			}
		}
	}

	if apiURL == "" {
		return nil, fmt.Errorf("api url is required")
	}
	if modelName == "" {
		return nil, fmt.Errorf("model name is required, please enter your model name")
	}

	var client llm.LLMClient
	if strings.Contains(apiURL, "generativelanguage.googleapis.com") {
		client = llm.NewGeminiClient(token, 15*time.Second)
	} else {
		client = llm.NewOpenAIClient(apiURL, token, 15*time.Second)
	}
	start := time.Now()

	resp, err := client.Generate(ctx, llm.CompletionRequest{
		Model: modelName,
		Messages: []llm.Message{
			{Role: "user", Content: "Say 'ok'"},
		},
		MaxTokens:   10,
		Temperature: 0.1,
	})
	latency := time.Since(start).Milliseconds()

	if err != nil {
		return &dtos.TestConnectionResponse{
			Success:   false,
			LatencyMs: latency,
			Message:   fmt.Sprintf("Connection failed: %v", err),
		}, nil
	}

	return &dtos.TestConnectionResponse{
		Success:    true,
		LatencyMs:  latency,
		Message:    fmt.Sprintf("Connection successful! Response: %s", strings.TrimSpace(resp.Content)),
		ModelFound: modelName,
	}, nil
}
