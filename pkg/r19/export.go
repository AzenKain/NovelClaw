package r19

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
)

// ExportToTSV writes all configured sensitive term mappings to an io.Writer in TSV format.
func (e *Engine) ExportToTSV(w io.Writer) error {
	e.masker.mu.RLock()
	defer e.masker.mu.RUnlock()

	var keys []string
	for k := range e.masker.terms {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		line := fmt.Sprintf("%s\t%s\n", k, e.masker.terms[k])
		if _, err := io.WriteString(w, line); err != nil {
			return err
		}
	}
	return nil
}

// ExportToJSON writes all configured sensitive term mappings to an io.Writer in formatted JSON.
func (e *Engine) ExportToJSON(w io.Writer) error {
	e.masker.mu.RLock()
	defer e.masker.mu.RUnlock()

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(e.masker.terms)
}
