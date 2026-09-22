package llm

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

// CriticAction defines the decision made by the Shadow Critic.
type CriticAction string

const (
	CriticActionPass   CriticAction = "PASS"
	CriticActionRevise CriticAction = "REVISE"
)

// CriticRequest contains sentence-level data for criticism and verification.
type CriticRequest struct {
	SourceLang       string `json:"source_lang"`
	TargetLang       string `json:"target_lang"`
	SelectiveContext string `json:"selective_context"`
	PrecedingContext string `json:"preceding_context"`
	OriginalSentence string `json:"original_sentence"`
	DraftTranslation string `json:"draft_translation"`
	StyleGuide       string `json:"style_guide,omitempty"`
	ThinkingEffort   string `json:"thinking_effort,omitempty"`
}

// CriticVerdict represents the evaluation result and optional correction from the critic.
type CriticVerdict struct {
	Action          CriticAction `json:"action"`
	Reason          string       `json:"reason"`
	Correction      string       `json:"correction"`
	PromptTokens    int          `json:"prompt_tokens,omitempty"`
	CompTokens      int          `json:"comp_tokens,omitempty"`
	TotalTokens     int          `json:"total_tokens,omitempty"`
	ReasoningTokens int          `json:"reasoning_tokens,omitempty"`
	Thought         string       `json:"thought,omitempty"`
}

// ShadowCritic inspects drafts in real time to catch pronoun drift, glossary errors, and stiffness.
type ShadowCritic struct {
	client      LLMClient
	model       string
	temperature float32
	timeout     time.Duration
}

// NewShadowCritic creates a new ShadowCritic instance.
func NewShadowCritic(client LLMClient, model string, timeout time.Duration) *ShadowCritic {
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	return &ShadowCritic{
		client:      client,
		model:       model,
		temperature: 0.1,
		timeout:     timeout,
	}
}

// EvaluateSentence inspects a draft translation against source text and selective context.
func (c *ShadowCritic) EvaluateSentence(ctx context.Context, req CriticRequest) (*CriticVerdict, error) {
	evalCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	systemPrompt := buildCriticSystemPrompt(req.SourceLang, req.TargetLang, req.SelectiveContext, req.PrecedingContext, req.StyleGuide)
	userPrompt := fmt.Sprintf("Source: %s\nDraft: %s", req.OriginalSentence, req.DraftTranslation)

	messages := []Message{
		{Role: RoleSystem, Content: systemPrompt},
		{Role: RoleUser, Content: userPrompt},
	}

	resp, err := c.client.Generate(evalCtx, CompletionRequest{
		Model:           c.model,
		Messages:        messages,
		Temperature:     c.temperature,
		ReasoningEffort: req.ThinkingEffort,
	})
	if err != nil {
		log.Warn().
			Err(err).
			Str("model", c.model).
			Msg("shadow critic evaluation timed out or failed, failing open with PASS")
		return &CriticVerdict{Action: CriticActionPass}, nil
	}

	verdict := parseCriticResponse(resp.Content, req.DraftTranslation)
	verdict.PromptTokens = resp.PromptTokens
	verdict.CompTokens = resp.CompTokens
	verdict.TotalTokens = resp.TotalTokens
	verdict.ReasoningTokens = resp.ReasoningTokens
	verdict.Thought = resp.Thought

	if verdict.Action == CriticActionRevise {
		log.Info().
			Str("reason", verdict.Reason).
			Str("correction", verdict.Correction).
			Msg("shadow critic proposed revision")
	}

	return verdict, nil
}

// buildCriticSystemPrompt constructs instructions for the Shadow Critic.
func buildCriticSystemPrompt(sourceLang, targetLang, selectiveContext, precedingContext, styleGuide string) string {
	var sb strings.Builder
	srcNorm := NormalizeLanguage(sourceLang)
	tgtNorm := NormalizeLanguage(targetLang)

	sb.WriteString(fmt.Sprintf("You are the Shadow Critic, an expert literary editor evaluating translations from '%s' to '%s'.\n", srcNorm, tgtNorm))
	sb.WriteString("CRITERIA:\n")
	sb.WriteString("1. Pronoun and Address Consistency: Does the draft strictly follow character address terms from [SELECTIVE CONTEXT]?\n")
	sb.WriteString("2. Mandatory Terminology: Are terms translated according to [SELECTIVE CONTEXT]?\n")
	sb.WriteString("3. Accuracy: Does the draft omit key meaning or hallucinate?\n")
	sb.WriteString("4. Anti-Translationese & Natural Literary Flow (STRICT REVISE IF STIFF, WORD-BY-WORD, OR CALQUED):\n")
	sb.WriteString("   - Does the draft suffer from literal word-by-word translation, idiom calques (e.g. translating idioms literally instead of using native target language idioms), or foreign preposition/clause syntax?\n")
	sb.WriteString(fmt.Sprintf("   - If the draft reads like machine translation or stiff foreign calques rather than natural, fluent literary prose in %s, you MUST issue REVISE with an idiomatic, beautifully polished literary correction!\n", tgtNorm))
	sb.WriteString("5. Capitalization & Orthography (STRICT REVISE IF VIOLATED):\n")
	sb.WriteString("   - Does the draft contain arbitrary capitalization (e.g. mid-sentence capitalized pronouns or common nouns where forbidden by target language grammar, or random Title Case)? If so, issue REVISE and fix to standard lowercase/orthography!\n")
	sb.WriteString("6. Punctuation & Typography Integrity (STRICT REVISE IF VIOLATED):\n")
	sb.WriteString("   - Does any sentence or paragraph terminate with a semicolon (;) or comma (,) instead of a period (.), exclamation mark (!), question mark (?), or ellipsis (...)?\n")
	sb.WriteString("   - Does the draft retain incompatible foreign punctuation marks (such as CJK '。', '，', '、', '；', '「」', '『』' in Western text, or Western semicolons in CJK prose)? If so, issue REVISE and convert to standard target typography!\n")

	isTargetCJK := tgtNorm == "Japanese" || tgtNorm == "Chinese" || tgtNorm == "Korean"
	isSourceCJK := srcNorm == "Japanese" || srcNorm == "Chinese" || srcNorm == "Korean"

	if !isTargetCJK && isSourceCJK {
		sb.WriteString("7. Untranslated Foreign Script Check (STRICT REVISE REQUIREMENT):\n")
		sb.WriteString("   If the draft contains ANY untranslated raw CJK characters (such as Japanese Kanji, Chinese Hanzi, Hiragana, Katakana, or Hangul) in the text, you MUST issue REVISE!\n")
		sb.WriteString("   REASON: Contains untranslated Kanji/CJK characters.\n")
		sb.WriteString(fmt.Sprintf("   CORRECTION: Replace all untranslated Kanji/CJK with proper transliteration or native %s terms according to publishing conventions and provide the complete corrected draft.\n", tgtNorm))
		if tgtNorm == "Vietnamese" {
			sb.WriteString("8. Honorifics Consistency & Stray Romaji Check (STRICT REVISE IF VIOLATED):\n")
			sb.WriteString("   - If relational terms ('tiền bối', 'anh', 'em') are used, verify that NO stray Romaji terms like 'senpai', 'sempai', 'kouhai' are inconsistently mixed into dialogue or prose. If any stray 'senpai'/'sempai' appears when 'tiền bối' is established, issue REVISE and unify it to 'tiền bối'!\n")
			sb.WriteString("   - NEVER retain raw Japanese honorific suffixes ('-chan', '-san', '-kun') attached to names in Romaji (e.g. 'Akari-chan' -> simply 'Akari' or 'em ấy'; 'Nanami-san' -> 'Nanami' or 'bạn Nanami'). If any appear, issue REVISE and remove the suffix!\n")
			sb.WriteString("9. Third-Person Pronoun Consistency with Established Character Relations (STRICT REVISE IF VIOLATED):\n")
			sb.WriteString("   - In 1st-person POV narration, the narrator's self-reference in narrative descriptions and internal thoughts is ALWAYS 'tôi' (I). The self-call pronoun ('anh', 'em', 'tớ') in [CHARACTER RELATIONS] applies to DIRECT SPOKEN DIALOGUE between characters, NOT the general narrative POV 'tôi'. NEVER let the narrator refer to himself as 'anh' in narration or internal thoughts (e.g. 'anh không thể không nảy sinh suy nghĩ' -> MUST BE 'tôi không thể không nảy sinh suy nghĩ')!\n")
			sb.WriteString("   - However, third-person references to other characters in narration must align with the established relationship graph: If the relation is 'anh - em', the narration naturally refers to her as 'em ấy' or by name, rather than aloof 'cô ấy'. If the relation is classmate or distant, use pronouns fitting that relation. Do not invent contradictory or mismatched narrative pronouns!\n")
			sb.WriteString("10. Sino-Vietnamese Convertese & Modern Realism Check (STRICT REVISE IF VIOLATED):\n")
			sb.WriteString("   In modern school/urban settings, strictly ban raw Sino-Vietnamese convertese: e.g. '単位制' MUST be 'hệ thống tín chỉ' / 'học theo tín chỉ' (NEVER 'học chế tín chỉ'); 'とんでもないこと' MUST be rendered naturally (e.g. 'lời lẽ động trời' / 'chuyện không tưởng', NEVER archaic wuxia cliche 'kinh thiên động địa'). If encountered, issue REVISE!\n")
			sb.WriteString("11. Anti-Infantilization Rule (STRICT REVISE IF VIOLATED):\n")
			sb.WriteString("   ABSOLUTELY FORBID prepending 'bé' to teenage or adult character names (e.g. 'bé Akari', 'bé Nanami' are STRICTLY FORBIDDEN and unnatural/cringe). Characters must be called simply by their name ('Akari', 'Nanami') or natural relational pronouns ('em', 'em ấy', 'cô nàng'). If any 'bé + [Name]' appears in draft, issue REVISE and strip 'bé'!\n")
			sb.WriteString("12. Logical Coherence & Modal Negation Inversion Check (STRICT REVISE IF VIOLATED):\n")
			sb.WriteString("   - NEVER invert the logical meaning of negative modal constructions or counterfactuals. For example, translating Japanese '〜なかったはずなのに' as 'lẽ ra phải không... mới đúng chứ' is an ABSURD LOGIC INVERSION (it mistakenly turns certainty of non-occurrence into an obligation to remain silent). It MUST be rendered naturally according to context (e.g. 'rõ ràng là... đâu có hé nửa lời cơ mà' or 'đáng lẽ một chuyện như vậy thì... ít ra cũng phải hé trước nửa lời mới phải chứ')!\n")
			sb.WriteString("   - Check for casual dialogue pronoun leakage into 1st-person narration: casual pronouns ('tao', 'mày') are STRICTLY FORBIDDEN in narration/monologues outside dialogue quotes (narration must strictly use 'tôi' or 'mình').\n")
			sb.WriteString("   - Direct Dialogue Address Consistency: verify that within a continuous conversation, characters maintain their established address terms (e.g. 'anh - em') without randomly switching to 'tôi - em' mid-scene.\n")
		} else {
			sb.WriteString("8. Pronoun & Address Consistency with Character Relations (STRICT REVISE IF VIOLATED):\n")
			sb.WriteString("   All pronouns and address terms in dialogue and narrative MUST strictly align with the established interpersonal relationship in [CHARACTER RELATIONS] / [SELECTIVE CONTEXT]. Do not invent contradictory or mismatched narrative terms.\n")
		}
	} else if isTargetCJK && !isSourceCJK {
		sb.WriteString("7. Untranslated Foreign Script Check (STRICT REVISE REQUIREMENT):\n")
		sb.WriteString("   If the draft contains untranslated raw English/Western sentence fragments or unadapted foreign words that should have been translated or transcribed, you MUST issue REVISE!\n")
		sb.WriteString("   REASON: Contains untranslated foreign words/phrases.\n")
		sb.WriteString(fmt.Sprintf("   CORRECTION: Translate or transcribe all foreign text into standard %s prose and provide the complete corrected draft.\n", tgtNorm))
	}

	sb.WriteString("13. Chatbot Meta-Talk Check (STRICT REVISE IF VIOLATED):\n")
	sb.WriteString("   The translation must contain ZERO chatbot conversational meta-talk (such as 'Here is the translation:', 'Certainly!', 'Sure!'). If the source is only images or SVG, output MUST only be the preserved markup without chat commentary!\n")

	if selectiveContext != "" {
		sb.WriteString(fmt.Sprintf("\n[SELECTIVE CONTEXT (L2/L3)]:\n%s\n", selectiveContext))
	}
	if precedingContext != "" {
		sb.WriteString(fmt.Sprintf("\n[PRECEDING CONTEXT (L0)]:\n%s\n", precedingContext))
	}
	if styleGuide != "" {
		sb.WriteString(fmt.Sprintf("\n[STYLE GUIDE & GUIDELINES]:\n%s\n", styleGuide))
	}

	sb.WriteString("\nOUTPUT FORMAT:\n")
	sb.WriteString("If the draft is acceptable, output ONLY:\nPASS\n\n")
	sb.WriteString("If revision is required, output EXACTLY:\n")
	sb.WriteString("REVISE\n")
	sb.WriteString("REASON: <concise explanation>\n")
	sb.WriteString("CORRECTION: <corrected translation text only>")
	return sb.String()
}

// parseCriticResponse parses LLM output into CriticVerdict.
func parseCriticResponse(rawResponse string, originalDraft string) *CriticVerdict {
	trimmed := strings.TrimSpace(rawResponse)
	if strings.HasPrefix(strings.ToUpper(trimmed), "PASS") {
		return &CriticVerdict{
			Action:     CriticActionPass,
			Correction: originalDraft,
		}
	}

	if strings.Contains(strings.ToUpper(trimmed), "REVISE") {
		lines := strings.Split(trimmed, "\n")
		var reason string
		var correctionLines []string
		isCapturingCorrection := false

		for _, line := range lines {
			trimmedLine := strings.TrimSpace(line)
			if strings.HasPrefix(strings.ToUpper(trimmedLine), "REASON:") {
				isCapturingCorrection = false
				reason = strings.TrimSpace(trimmedLine[7:])
			} else if strings.HasPrefix(strings.ToUpper(trimmedLine), "CORRECTION:") {
				isCapturingCorrection = true
				firstLine := strings.TrimSpace(trimmedLine[11:])
				if firstLine != "" {
					correctionLines = append(correctionLines, firstLine)
				}
			} else if isCapturingCorrection {
				correctionLines = append(correctionLines, line)
			}
		}

		fullCorrection := strings.TrimSpace(strings.Join(correctionLines, "\n"))
		if fullCorrection != "" {
			return &CriticVerdict{
				Action:     CriticActionRevise,
				Reason:     reason,
				Correction: fullCorrection,
			}
		}
	}

	return &CriticVerdict{
		Action:     CriticActionPass,
		Correction: originalDraft,
	}
}

// ArbitrateDilemma automatically arbitrates ambiguous translation choices via LLM critique.
func (c *ShadowCritic) ArbitrateDilemma(ctx context.Context, question string, options []string, contextSnippet string, styleGuide string) (string, error) {
	if len(options) == 0 {
		return "Default recommendation", nil
	}
	if c == nil || c.client == nil {
		return options[0], nil
	}

	evalCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	systemPrompt := `You are the Lead Editorial Arbitrator (Shadow Critic) for a professional translation studio.
A dilemma or ambiguity has arisen during literary translation.
Your task is to select the single best option based on literary context, historical consistency, and style guidelines.

Format your response strictly as:
SELECTED: <exact option>
RATIONALE: <one-sentence reasoning>`

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Dilemma Question: %s\n", question))
	sb.WriteString(fmt.Sprintf("Context Snippet: %s\n", contextSnippet))
	if styleGuide != "" {
		sb.WriteString(fmt.Sprintf("Style Guide: %s\n", styleGuide))
	}
	sb.WriteString("Options:\n")
	for idx, opt := range options {
		sb.WriteString(fmt.Sprintf("%d. %s\n", idx+1, opt))
	}

	resp, err := c.client.Generate(evalCtx, CompletionRequest{
		Model: c.model,
		Messages: []Message{
			{Role: RoleSystem, Content: systemPrompt},
			{Role: RoleUser, Content: sb.String()},
		},
		Temperature: c.temperature,
		MaxTokens:   300,
	})
	if err != nil {
		log.Warn().Err(err).Msg("Shadow Critic ArbitrateDilemma failed, falling back to option 0")
		return options[0], nil
	}

	content := strings.TrimSpace(resp.Content)
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "SELECTED:") {
			candidate := strings.TrimSpace(strings.TrimPrefix(trimmed, "SELECTED:"))
			for _, opt := range options {
				if strings.EqualFold(candidate, opt) || strings.Contains(strings.ToLower(candidate), strings.ToLower(opt)) {
					return opt, nil
				}
			}
			if candidate != "" {
				return candidate, nil
			}
		}
	}

	// Match any option mentioned in the response
	for _, opt := range options {
		if strings.Contains(content, opt) {
			return opt, nil
		}
	}

	return options[0], nil
}
