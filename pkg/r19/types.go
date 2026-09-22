package r19

// Category identifies the type of sensitive content for masking.
type Category string

const (
	CategorySexual    Category = "sexual"
	CategoryAnatomy   Category = "anatomy"
	CategoryTaboo     Category = "taboo"
	CategoryUnderage  Category = "underage_age"
	CategoryEuphemism Category = "euphemism"
)

// Entry records an original sensitive term and its masked representation.
type Entry struct {
	ID          int      `json:"id"`
	Source      string   `json:"source"`
	Translation string   `json:"translation"`
	Tag         string   `json:"tag"`
	Category    Category `json:"category"`
}

// MaskResult contains the masked text and metadata required for restoration.
type MaskResult struct {
	MaskedText          string  `json:"masked_text"`
	Entries             []Entry `json:"entries"`
	TermsMaskedCount    int     `json:"terms_masked_count"`
	AgeNeutralizedCount int     `json:"age_neutralized_count"`
}

// Config controls the behavior of the local R19 masking and restoration engine.
type Config struct {
	Enabled            bool   `json:"enabled"`
	NeutralizeAges     bool   `json:"neutralize_ages"`
	LiteraryDisclaimer bool   `json:"literary_disclaimer"`
	TagFormat          string `json:"tag_format"`
	TargetLang         string `json:"target_lang"`
}

// DefaultConfig returns recommended production defaults for local R19 processing.
func DefaultConfig() Config {
	return Config{
		Enabled:            true,
		NeutralizeAges:     true,
		LiteraryDisclaimer: true,
		TagFormat:          "tag",
		TargetLang:         "Vietnamese",
	}
}
