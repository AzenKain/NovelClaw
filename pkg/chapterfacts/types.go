package chapterfacts

// RelationshipChange records a change in addressing pronouns or interpersonal dynamic.
type RelationshipChange struct {
	FromChar   string `json:"from_char"`
	ToChar     string `json:"to_char"`
	CallAs     string `json:"call_as"`
	SelfCallAs string `json:"self_call_as"`
	Tone       string `json:"tone"`
}

// StateChange captures mutations in character or entity attributes across chapter progression.
type StateChange struct {
	Entity   string `json:"entity"`
	Field    string `json:"field"`
	OldValue string `json:"old_value"`
	NewValue string `json:"new_value"`
	Reason   string `json:"reason"`
}

// CastIntro represents a newly introduced character or entity in the chapter.
type CastIntro struct {
	Name   string `json:"name"`
	Role   string `json:"role"`
	Gender string `json:"gender"`
}

// ChapterFacts contains structured episodic extractions from a completed chapter.
type ChapterFacts struct {
	Title               string               `json:"title"`
	Summary             string               `json:"summary"`
	KeyEvents           []string             `json:"key_events"`
	RelationshipChanges []RelationshipChange `json:"relationship_changes"`
	StateChanges        []StateChange        `json:"state_changes"`
	CastIntros          []CastIntro          `json:"cast_intros"`
}
