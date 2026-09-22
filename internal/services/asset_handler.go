package services

import (
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"novelclaw/pkg/storage"
)

// AssetHandler intercepts reader asset requests and serves local book images.
type AssetHandler struct {
	store    *storage.Storage
	fallback http.Handler
}

// NewAssetHandler creates a new AssetHandler wrapping an existing fallback http.Handler.
func NewAssetHandler(store *storage.Storage, fallback http.Handler) *AssetHandler {
	return &AssetHandler{
		store:    store,
		fallback: fallback,
	}
}

// ServeHTTP handles requests for reader book assets or delegates to the fallback asset server.
func (h *AssetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	const prefix = "/api/v1/reader/"
	if !strings.HasPrefix(r.URL.Path, prefix) {
		h.fallback.ServeHTTP(w, r)
		return
	}

	trimmed := strings.TrimPrefix(r.URL.Path, prefix)
	parts := strings.SplitN(trimmed, "/asset/", 2)
	if len(parts) != 2 {
		h.fallback.ServeHTTP(w, r)
		return
	}

	projectID, err := url.PathUnescape(parts[0])
	if err != nil || projectID == "" {
		http.Error(w, "invalid project id", http.StatusBadRequest)
		return
	}

	assetPath, err := url.PathUnescape(parts[1])
	if err != nil || assetPath == "" {
		http.Error(w, "invalid asset path", http.StatusBadRequest)
		return
	}

	cleanAssetPath := filepath.Clean(filepath.FromSlash(assetPath))
	if strings.HasPrefix(cleanAssetPath, "..") {
		http.Error(w, "forbidden path traversal", http.StatusForbidden)
		return
	}

	assetDir := filepath.Join("data", "assets", projectID)
	fullPath := filepath.Join(assetDir, cleanAssetPath)
	data, readErr := os.ReadFile(fullPath)

	// 1. If fullPath failed, check base name in the specific volume subfolder
	if readErr != nil {
		pathParts := strings.SplitN(cleanAssetPath, string(filepath.Separator), 2)
		if len(pathParts) == 2 {
			volDir := pathParts[0]
			baseInVol := filepath.Join(assetDir, volDir, filepath.Base(pathParts[1]))
			data, readErr = os.ReadFile(baseInVol)
		}
	}

	// 2. If still failed, check root base name (legacy flat format)
	if readErr != nil {
		basePath := filepath.Join(assetDir, filepath.Base(cleanAssetPath))
		data, readErr = os.ReadFile(basePath)
	}

	// 3. If still failed (e.g. legacy chapter requested without volume prefix),
	// search across volume subfolders
	if readErr != nil {
		entries, err := os.ReadDir(assetDir)
		if err == nil {
			for _, entry := range entries {
				if entry.IsDir() {
					candidate := filepath.Join(assetDir, entry.Name(), cleanAssetPath)
					if d, err := os.ReadFile(candidate); err == nil {
						data = d
						readErr = nil
						break
					}
					candidateBase := filepath.Join(assetDir, entry.Name(), filepath.Base(cleanAssetPath))
					if d, err := os.ReadFile(candidateBase); err == nil {
						data = d
						readErr = nil
						break
					}
				}
			}
		}
	}

	if readErr != nil {
		http.NotFound(w, r)
		return
	}

	ext := strings.ToLower(filepath.Ext(cleanAssetPath))
	contentType := mime.TypeByExtension(ext)
	if contentType == "" {
		contentType = http.DetectContentType(data)
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "public, max-age=86400")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}
