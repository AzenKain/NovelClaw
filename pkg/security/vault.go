package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
)

var (
	ErrDecryptionFailed = errors.New("decryption failed: invalid ciphertext or corrupted key")
)

// emptySentinel marks an intentionally empty secret (e.g. a provider that
// needs no API key). Storing a real ciphertext for "" lets custom/Ollama
// providers be saved instead of failing with ErrEmptyPlaintext.
const emptySentinel = "__novelclaw_empty__"

var (
	keysOnce    sync.Once
	primaryKey  []byte
	legacyKeys  [][]byte
	keyRingWarn []string
)

// legacyKeyDirs lists pre-rebrand home directories that may still hold a
// vault key. The project was renamed NekoNovel → NovelClaw, so machines
// upgraded across that rename have tokens encrypted under the old key.
var legacyKeyDirs = []string{".neko-novel", ".neko_novel", ".nekonovel", ".neko"}

// keyFilePath returns the canonical vault key path for this machine.
func keyFilePath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}
	return filepath.Join(homeDir, ".novelclaw", ".vault.key")
}

// legacyKeyPaths returns every legacy vault key path that exists on disk.
func legacyKeyPaths() []string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil
	}
	var out []string
	for _, dir := range legacyKeyDirs {
		out = append(out, filepath.Join(homeDir, dir, ".vault.key"))
	}
	return out
}

// initKeys resolves the primary key and the legacy fallback key ring.
//
// Order of preference:
//  1. NEKO_SECRET_KEY / NOVELCLAW_SECRET_KEY environment override.
//  2. The canonical ~/.novelclaw/.vault.key.
//  3. A legacy pre-rebrand key, promoted to the canonical location.
//  4. A freshly generated key.
//
// Legacy keys that are NOT promoted (because a canonical key already
// exists) are kept in the ring so previously-encrypted tokens still
// decrypt; those tokens are re-encrypted with the primary key on next save.
func initKeys() {
	keysOnce.Do(func() {
		for _, env := range []string{"NOVELCLAW_SECRET_KEY", "NEKO_SECRET_KEY"} {
			if v := os.Getenv(env); v != "" {
				sum := sha256.Sum256([]byte(v))
				primaryKey = sum[:]
				return
			}
		}

		secretFile := keyFilePath()
		_ = os.MkdirAll(filepath.Dir(secretFile), 0700)

		if data, err := os.ReadFile(secretFile); err == nil && len(data) == 32 {
			primaryKey = data
		} else {
			// Promote a legacy key so the machine keeps ONE stable key
			// going forward. This is the critical upgrade path: without it
			// a fresh key is minted and every stored token becomes
			// undecryptable, silently hiding all saved LLM configs.
			for _, legacyPath := range legacyKeyPaths() {
				data, err := os.ReadFile(legacyPath)
				if err != nil || len(data) != 32 {
					continue
				}
				if werr := os.WriteFile(secretFile, data, 0600); werr != nil {
					keyRingWarn = append(keyRingWarn, "could not persist promoted vault key: "+werr.Error())
				}
				primaryKey = data
				break
			}
		}

		if primaryKey == nil {
			newKey := make([]byte, 32)
			if _, err := io.ReadFull(rand.Reader, newKey); err != nil {
				sum := sha256.Sum256([]byte("novelclaw-deterministic-fallback-salt-2026"))
				newKey = sum[:]
			}
			if werr := os.WriteFile(secretFile, newKey, 0600); werr != nil {
				keyRingWarn = append(keyRingWarn, "could not persist new vault key: "+werr.Error())
			}
			primaryKey = newKey
		}

		// Retain any legacy keys that differ from the primary so tokens
		// encrypted before the rebrand still decrypt.
		for _, legacyPath := range legacyKeyPaths() {
			data, err := os.ReadFile(legacyPath)
			if err != nil || len(data) != 32 {
				continue
			}
			if string(data) == string(primaryKey) {
				continue
			}
			legacyKeys = append(legacyKeys, data)
		}
	})
}

// getMasterKey returns the primary key (kept for existing callers/tests).
func getMasterKey() []byte {
	initKeys()
	return primaryKey
}

// Warnings returns non-fatal key-ring problems (e.g. unwritable key file)
// so the application can surface them in logs.
func Warnings() []string {
	initKeys()
	return keyRingWarn
}

// gcmFor builds an AES-256-GCM AEAD for the supplied key.
func gcmFor(key []byte) (cipher.AEAD, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("create cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create gcm: %w", err)
	}
	return gcm, nil
}

// Encrypt encrypts a string using AES-256-GCM under the primary key.
// An empty plaintext is stored as an encrypted sentinel so keyless
// providers (Ollama, custom endpoints) can still be persisted.
func Encrypt(plaintext string) (string, error) {
	initKeys()
	if plaintext == "" {
		plaintext = emptySentinel
	}

	gcm, err := gcmFor(primaryKey)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decryptWith attempts to decrypt using one specific key.
func decryptWith(key []byte, data []byte) (string, error) {
	gcm, err := gcmFor(key)
	if err != nil {
		return "", err
	}
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", ErrDecryptionFailed
	}
	plaintext, err := gcm.Open(nil, data[:nonceSize], data[nonceSize:], nil)
	if err != nil {
		return "", ErrDecryptionFailed
	}
	return string(plaintext), nil
}

// Decrypt decrypts a base64-encoded ciphertext, trying the primary key
// first and then every retained legacy key. Using a legacy key is
// transparent to callers; DecryptWithKeyID reports which key matched so
// the storage layer can re-encrypt stale rows.
func Decrypt(encryptedBase64 string) (string, error) {
	plain, _, err := DecryptWithKeyID(encryptedBase64)
	return plain, err
}

// DecryptWithKeyID decrypts and reports whether a legacy key was used
// (usedLegacy == true means the row should be re-encrypted with the
// primary key on the next write).
func DecryptWithKeyID(encryptedBase64 string) (plaintext string, usedLegacy bool, err error) {
	if encryptedBase64 == "" {
		return "", false, nil // nothing stored → empty token, not an error
	}

	data, err := base64.StdEncoding.DecodeString(encryptedBase64)
	if err != nil {
		return "", false, fmt.Errorf("decode base64: %w", err)
	}

	initKeys()

	if plain, derr := decryptWith(primaryKey, data); derr == nil {
		return unmarkEmpty(plain), false, nil
	}
	for _, legacy := range legacyKeys {
		if plain, derr := decryptWith(legacy, data); derr == nil {
			return unmarkEmpty(plain), true, nil
		}
	}
	return "", false, ErrDecryptionFailed
}

// unmarkEmpty converts the empty-secret sentinel back to "".
func unmarkEmpty(s string) string {
	if s == emptySentinel {
		return ""
	}
	return s
}
