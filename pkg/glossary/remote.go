package glossary

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// RemoteSourceDefinition represents a curated public online dictionary source.
type RemoteSourceDefinition struct {
	ID              string `json:"id"`
	Title           string `json:"title"`
	Genre           string `json:"genre"`
	Description     string `json:"description"`
	URL             string `json:"url"`
	Format          string `json:"format"` // "vietphrase" | "tsv" | "json"
	DefaultCategory string `json:"default_category"`
}

// GetCommunityRemoteSources returns curated public dictionary sources hosted on GitHub and community repositories.
// Returns an empty slice if no verified sources are configured.
func GetCommunityRemoteSources() []RemoteSourceDefinition {
	return []RemoteSourceDefinition{}
}

// RemoteDownloader handles fetching and importing dictionaries from remote HTTP/HTTPS endpoints.
type RemoteDownloader struct {
	hub        *Hub
	httpClient *http.Client
}

// NewRemoteDownloader creates a RemoteDownloader with a default HTTP client.
func NewRemoteDownloader(hub *Hub) *RemoteDownloader {
	return &RemoteDownloader{
		hub: hub,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// SetHTTPClient overrides the internal HTTP client (useful for custom timeouts or testing).
func (d *RemoteDownloader) SetHTTPClient(client *http.Client) {
	if client != nil {
		d.httpClient = client
	}
}

// DownloadAndImport fetches dictionary content from a remote URL and imports it into the project's SQLite glossary.
func (d *RemoteDownloader) DownloadAndImport(ctx context.Context, projectID string, sourceURL string, format string, defaultCategory string) (int, error) {
	cleanURL := strings.TrimSpace(sourceURL)
	if cleanURL == "" {
		return 0, fmt.Errorf("source URL is empty")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, cleanURL, nil)
	if err != nil {
		return 0, fmt.Errorf("create http request: %w", err)
	}
	req.Header.Set("User-Agent", "NekoNovel-Studio/1.0 (Glossary Downloader)")

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("fetch remote dictionary: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("remote server returned status: %d %s", resp.StatusCode, resp.Status)
	}

	// Limit maximum download size to 25MB to prevent memory exhaustion
	limitR := io.LimitReader(resp.Body, 25*1024*1024)

	formatClean := strings.ToLower(strings.TrimSpace(format))
	if formatClean == "" {
		if strings.HasSuffix(cleanURL, ".json") {
			formatClean = "json"
		} else if strings.HasSuffix(cleanURL, ".tsv") {
			formatClean = "tsv"
		} else {
			formatClean = "vietphrase"
		}
	}

	switch formatClean {
	case "tsv":
		return d.hub.ImportTSV(ctx, projectID, limitR)
	case "json":
		data, err := io.ReadAll(limitR)
		if err != nil {
			return 0, fmt.Errorf("read remote json: %w", err)
		}
		return d.hub.ImportJSON(ctx, projectID, data)
	default: // "vietphrase" or plain text
		data, err := io.ReadAll(limitR)
		if err != nil {
			return 0, fmt.Errorf("read remote text: %w", err)
		}
		return d.hub.ImportPlainText(ctx, projectID, string(data), defaultCategory)
	}
}
