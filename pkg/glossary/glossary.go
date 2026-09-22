package glossary

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"strings"

	"novelclaw/internal/gen/sqlc"
	"novelclaw/pkg/jsonx"
	"novelclaw/pkg/storage"
)

// GlossaryItem represents a terminology entry for import/export.
type GlossaryItem struct {
	ID         string `json:"id"`
	ProjectID  string `json:"project_id"`
	SourceTerm string `json:"source_term"`
	TargetTerm string `json:"target_term"`
	Category   string `json:"category"`
	Notes      string `json:"notes"`
}

// Hub manages multi-language glossary import, export, and category lookups.
type Hub struct {
	store *storage.Storage
}

// NewHub creates a new Glossary Hub.
func NewHub(store *storage.Storage) *Hub {
	return &Hub{store: store}
}

// ImportTSV imports glossary items from a Tab-Separated Values (TSV) stream.
func (h *Hub) ImportTSV(ctx context.Context, projectID string, r io.Reader) (int, error) {
	reader := csv.NewReader(r)
	reader.Comma = '\t'
	reader.FieldsPerRecord = -1

	imported := 0
	lineNum := 0
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return imported, fmt.Errorf("read tsv line %d: %w", lineNum+1, err)
		}
		lineNum++

		if len(record) == 0 || strings.TrimSpace(record[0]) == "" {
			continue
		}

		if lineNum == 1 && (strings.EqualFold(record[0], "source") || strings.EqualFold(record[0], "source_term")) {
			continue
		}

		sourceTerm := strings.TrimSpace(record[0])
		targetTerm := ""
		if len(record) > 1 {
			targetTerm = strings.TrimSpace(record[1])
		}
		category := "general"
		if len(record) > 2 && strings.TrimSpace(record[2]) != "" {
			category = strings.TrimSpace(record[2])
		}
		notes := ""
		if len(record) > 3 {
			notes = strings.TrimSpace(record[3])
		}

		itemID := fmt.Sprintf("glo_%s_%s", projectID, sourceTerm)
		err = h.store.UpsertGlossaryTerm(ctx, sqlc.UpsertGlossaryTermParams{
			ID:         itemID,
			ProjectID:  projectID,
			SourceTerm: sourceTerm,
			TargetTerm: targetTerm,
			Category:   category,
			Notes:      notes,
		})
		if err != nil {
			return imported, fmt.Errorf("upsert term %s: %w", sourceTerm, err)
		}
		imported++
	}

	return imported, nil
}

// ExportTSV exports project glossary items to a TSV stream.
func (h *Hub) ExportTSV(ctx context.Context, projectID string, w io.Writer) error {
	items, err := h.store.ListGlossaryByProject(ctx, projectID)
	if err != nil {
		return fmt.Errorf("list glossary: %w", err)
	}

	writer := csv.NewWriter(w)
	writer.Comma = '\t'

	_ = writer.Write([]string{"source_term", "target_term", "category", "notes"})
	for _, item := range items {
		err := writer.Write([]string{item.SourceTerm, item.TargetTerm, item.Category, item.Notes})
		if err != nil {
			return fmt.Errorf("write tsv row: %w", err)
		}
	}

	writer.Flush()
	return writer.Error()
}

// ImportJSON imports glossary items from JSON bytes.
func (h *Hub) ImportJSON(ctx context.Context, projectID string, data []byte) (int, error) {
	var items []GlossaryItem
	if err := jsonx.Unmarshal(data, &items); err != nil {
		return 0, fmt.Errorf("unmarshal glossary json: %w", err)
	}

	count := 0
	for _, it := range items {
		cleanSource := strings.TrimSpace(it.SourceTerm)
		if cleanSource == "" {
			continue
		}
		category := it.Category
		if category == "" {
			category = "general"
		}
		id := it.ID
		if id == "" {
			id = fmt.Sprintf("glo_%s_%s", projectID, cleanSource)
		}
		err := h.store.UpsertGlossaryTerm(ctx, sqlc.UpsertGlossaryTermParams{
			ID:         id,
			ProjectID:  projectID,
			SourceTerm: cleanSource,
			TargetTerm: strings.TrimSpace(it.TargetTerm),
			Category:   category,
			Notes:      it.Notes,
		})
		if err != nil {
			return count, fmt.Errorf("upsert json item %s: %w", cleanSource, err)
		}
		count++
	}
	return count, nil
}

// ExportJSON exports project glossary items as JSON bytes.
func (h *Hub) ExportJSON(ctx context.Context, projectID string) ([]byte, error) {
	items, err := h.store.ListGlossaryByProject(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("list glossary: %w", err)
	}

	out := make([]GlossaryItem, 0, len(items))
	for _, it := range items {
		out = append(out, GlossaryItem{
			ID:         it.ID,
			ProjectID:  it.ProjectID,
			SourceTerm: it.SourceTerm,
			TargetTerm: it.TargetTerm,
			Category:   it.Category,
			Notes:      it.Notes,
		})
	}

	return jsonx.Marshal(out)
}

// ImportPlainText imports glossary items from standard 'source=target' or 'source\ttarget' lines (Vietphrase / QuickTranslator / Names.txt format).
func (h *Hub) ImportPlainText(ctx context.Context, projectID string, content string, defaultCategory string) (int, error) {
	if defaultCategory == "" {
		defaultCategory = "other"
	}

	lines := strings.Split(content, "\n")
	imported := 0

	for _, rawLine := range lines {
		line := strings.TrimSpace(rawLine)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
			continue
		}

		var sourceTerm, targetTerm string
		if strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			sourceTerm = strings.TrimSpace(parts[0])
			targetTerm = strings.TrimSpace(parts[1])
		} else if strings.Contains(line, "\t") {
			parts := strings.SplitN(line, "\t", 2)
			sourceTerm = strings.TrimSpace(parts[0])
			targetTerm = strings.TrimSpace(parts[1])
		} else if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			sourceTerm = strings.TrimSpace(parts[0])
			targetTerm = strings.TrimSpace(parts[1])
		}

		if sourceTerm == "" || targetTerm == "" {
			continue
		}

		itemID := fmt.Sprintf("glo_%s_%s", projectID, sourceTerm)
		err := h.store.UpsertGlossaryTerm(ctx, sqlc.UpsertGlossaryTermParams{
			ID:         itemID,
			ProjectID:  projectID,
			SourceTerm: sourceTerm,
			TargetTerm: targetTerm,
			Category:   defaultCategory,
			Notes:      "Imported via PlainText/Vietphrase format",
		})
		if err != nil {
			return imported, fmt.Errorf("upsert plain-text term %s: %w", sourceTerm, err)
		}
		imported++
	}

	return imported, nil
}

