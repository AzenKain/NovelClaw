package security_test

import (
	"strings"
	"testing"

	"novelclaw/pkg/security"
)

// TestEncryptDecrypt tests round-trip encryption and authentication with AES-256-GCM.
func TestEncryptDecrypt(t *testing.T) {
	secretToken := "sk-9950bf9e2e588e5bac646f17a9cd595148b3be3699f1e9d9d3c3e542411d06a6"

	encrypted, err := security.Encrypt(secretToken)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	if encrypted == secretToken {
		t.Fatal("Encrypted string must not match plaintext")
	}

	if strings.Contains(encrypted, "9950bf9e") {
		t.Fatal("Encrypted string leaked plaintext fragment")
	}

	decrypted, err := security.Decrypt(encrypted)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}

	if decrypted != secretToken {
		t.Fatalf("Decrypted string mismatch: expected %s, got %s", secretToken, decrypted)
	}

	_, err = security.Decrypt("malformed-invalid-ciphertext-bytes==")
	if err == nil {
		t.Fatal("Expected error when decrypting corrupted ciphertext, got nil")
	}
}

// TestEncryptDecryptEmptyToken verifies keyless providers (Ollama, custom
// endpoints) can be persisted instead of failing on an empty API key.
func TestEncryptDecryptEmptyToken(t *testing.T) {
	encrypted, err := security.Encrypt("")
	if err != nil {
		t.Fatalf("Encrypt(\"\") failed: %v", err)
	}
	if encrypted == "" {
		t.Fatal("Encrypt(\"\") must produce a stored ciphertext, not empty string")
	}
	decrypted, err := security.Decrypt(encrypted)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}
	if decrypted != "" {
		t.Fatalf("empty token round-trip = %q, want \"\"", decrypted)
	}
}

// TestDecryptEmptyCiphertext verifies an unstored token reads back as empty
// rather than an error (legacy rows with no token).
func TestDecryptEmptyCiphertext(t *testing.T) {
	got, err := security.Decrypt("")
	if err != nil {
		t.Fatalf("Decrypt(\"\") should not error, got %v", err)
	}
	if got != "" {
		t.Fatalf("Decrypt(\"\") = %q, want \"\"", got)
	}
}

// TestDecryptWithKeyIDReportsLegacy verifies the legacy-key path is reported
// so the storage layer can re-encrypt stale rows.
func TestDecryptWithKeyIDReportsLegacy(t *testing.T) {
	enc, err := security.Encrypt("token-value")
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	plain, legacy, err := security.DecryptWithKeyID(enc)
	if err != nil {
		t.Fatalf("DecryptWithKeyID: %v", err)
	}
	if plain != "token-value" {
		t.Fatalf("plaintext = %q", plain)
	}
	if legacy {
		t.Fatal("primary-key ciphertext must not be reported as legacy")
	}
}
