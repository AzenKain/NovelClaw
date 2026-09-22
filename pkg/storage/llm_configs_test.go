package storage_test

import (
	"context"
	"strings"
	"testing"

	"novelclaw/pkg/storage"
)

// TestLLMConfigEncryptionAndPersistence tests encryption at rest and transparent decryption for LLM configurations.
func TestLLMConfigEncryptionAndPersistence(t *testing.T) {
	store := setupTestDB(t)
	ctx := context.Background()

	secretToken := "sk-dummy-test-token-abcdef1234567890"
	cfg := storage.LLMConfig{
		ID:           "test_provider",
		ProviderName: "openai_compatible",
		ApiURL:       "https://api.example.com",
		Token:        secretToken,
		ModelName:    "test-model",
		IsActive:     true,
		IsDefault:    true,
		RateLimitRPM: 60,
	}

	err := store.SaveLLMConfig(ctx, cfg)
	if err != nil {
		t.Fatalf("SaveLLMConfig failed: %v", err)
	}

	var storedEncryptedToken string
	err = store.DB().QueryRow("SELECT encrypted_token FROM llm_configs WHERE id = ?", "test_provider").Scan(&storedEncryptedToken)
	if err != nil {
		t.Fatalf("Direct DB scan failed: %v", err)
	}

	if storedEncryptedToken == secretToken {
		t.Fatal("SECURITY BREACH: API Token is stored in plaintext!")
	}
	if strings.Contains(storedEncryptedToken, "abcdef123456") {
		t.Fatal("SECURITY BREACH: Plaintext token fragment found in database column!")
	}

	retrieved, err := store.GetDefaultLLMConfig(ctx)
	if err != nil {
		t.Fatalf("GetDefaultLLMConfig failed: %v", err)
	}

	if retrieved.Token != secretToken {
		t.Fatalf("Decrypted token mismatch: expected %s, got %s", secretToken, retrieved.Token)
	}
	if retrieved.ApiURL != "https://api.example.com" {
		t.Errorf("expected api url https://api.example.com, got %s", retrieved.ApiURL)
	}
	if retrieved.ModelName != "test-model" {
		t.Errorf("expected model test-model, got %s", retrieved.ModelName)
	}
	if retrieved.MaxRetries != 3 {
		t.Errorf("expected default max_retries 3, got %d", retrieved.MaxRetries)
	}
	if retrieved.TimeoutSeconds != 120 {
		t.Errorf("expected default timeout_seconds 120, got %d", retrieved.TimeoutSeconds)
	}

	customCfg := storage.LLMConfig{
		ID:             "custom_resilience",
		ProviderName:   "openai_compatible",
		ApiURL:         "https://api.example.com",
		Token:          secretToken,
		ModelName:      "test-model",
		IsActive:       true,
		MaxRetries:     5,
		TimeoutSeconds: 180,
	}
	if err := store.SaveLLMConfig(ctx, customCfg); err != nil {
		t.Fatalf("SaveLLMConfig custom resilience failed: %v", err)
	}
	retrievedCustom, err := store.GetLLMConfigByID(ctx, "custom_resilience")
	if err != nil {
		t.Fatalf("GetLLMConfigByID failed: %v", err)
	}
	if retrievedCustom.MaxRetries != 5 {
		t.Errorf("expected max_retries 5, got %d", retrievedCustom.MaxRetries)
	}
	if retrievedCustom.TimeoutSeconds != 180 {
		t.Errorf("expected timeout_seconds 180, got %d", retrievedCustom.TimeoutSeconds)
	}
}
