package llm

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	"novelclaw/pkg/graph"
	"novelclaw/pkg/r19"
	"novelclaw/pkg/storage"
	"novelclaw/pkg/websearch"
)

// SkillChecker provides methods to query whether a skill is enabled and retrieve its dynamic prompts.
type SkillChecker interface {
	IsSkillEnabled(ctx context.Context, projectID string, skillID string) bool
	GetActiveSkillPrompts(ctx context.Context, projectID string, category string) string
}

// TranslationOrchestrator manages multi-mode translations and dual-agent workflows.
type TranslationOrchestrator struct {
	client          LLMClient
	model           string
	store           *storage.Storage
	critic          *ShadowCritic
	graph           *graph.ProgressiveGraph
	chunker         *Chunker
	memory          *MemoryManager
	toolReg         *ToolRegistry
	researchEng     *websearch.ResearchEngine
	r19Eng          *r19.Engine
	softStopChecker func() bool
	skillChecker    SkillChecker
}

// NewTranslationOrchestrator creates a new TranslationOrchestrator.
func NewTranslationOrchestrator(client LLMClient, model string, store *storage.Storage, researchEng *websearch.ResearchEngine) *TranslationOrchestrator {
	if researchEng == nil {
		researchEng = websearch.NewResearchEngine(30 * time.Second)
	}
	critic := NewShadowCritic(client, model, 30*time.Second)
	orch := &TranslationOrchestrator{
		client:      client,
		model:       model,
		store:       store,
		critic:      critic,
		graph:       graph.NewProgressiveGraph(store),
		chunker:     NewChunker(2500),
		memory:      NewMemoryManager(store),
		toolReg:     NewToolRegistry(store, researchEng),
		researchEng: researchEng,
		r19Eng:      r19.NewEngine(r19.DefaultConfig(), nil),
	}
	orch.toolReg.SetShadowCritic(critic)
	return orch
}

// SetSoftStopChecker attaches a predicate to check if the user requested a graceful pause between chunks.
func (o *TranslationOrchestrator) SetSoftStopChecker(checker func() bool) {
	o.softStopChecker = checker
}

// makeSoftStopContext wraps a parent context so it is cancelled as soon as
// the soft-stop flag transitions to true.  This ensures that an in-flight
// LLM Generate call is interrupted within ~250ms of the user pressing
// "Soft Stop", instead of waiting until the full chunk completes (which
// can take 30-120 seconds).
func (o *TranslationOrchestrator) makeSoftStopContext(parent context.Context) (context.Context, context.CancelFunc) {
	if o.softStopChecker == nil {
		return parent, func() {}
	}
	ctx, cancel := context.WithCancel(parent)
	go func() {
		ticker := time.NewTicker(250 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if o.softStopChecker() {
					cancel()
					return
				}
			}
		}
	}()
	return ctx, cancel
}

// SetAskHumanHandler assigns a human-in-the-loop callback for interactive clarifications.
func (o *TranslationOrchestrator) SetAskHumanHandler(handler AskHumanCallback) {
	if o.toolReg != nil {
		o.toolReg.SetAskHumanHandler(handler)
	}
}

// GetToolRegistry returns the active tool registry instance.
func (o *TranslationOrchestrator) GetToolRegistry() *ToolRegistry {
	return o.toolReg
}

// SetCriticModel configures a dedicated model or client for the Shadow Critic.
func (o *TranslationOrchestrator) SetCriticModel(client LLMClient, model string) {
	if client == nil {
		client = o.client
	}
	o.critic = NewShadowCritic(client, model, 30*time.Second)
	if o.toolReg != nil {
		o.toolReg.SetShadowCritic(o.critic)
	}
}

// SetSkillChecker attaches a modular skill registry or capability checker.
func (o *TranslationOrchestrator) SetSkillChecker(checker SkillChecker) {
	o.skillChecker = checker
}

// TranslateChapter coordinates chapter translation using the specified translation mode.
func (o *TranslationOrchestrator) TranslateChapter(
	ctx context.Context,
	projectID string,
	chapterIndex int64,
	opts TranslationOptions,
	handler *TranslationEventHandler,
) (string, *ExecutionTelemetry, error) {
	if handler == nil {
		handler = &TranslationEventHandler{}
	}

	startTime := time.Now()
	telemetry := &ExecutionTelemetry{
		Mode: opts.Mode,
	}

	project, err := o.store.GetProject(ctx, projectID)
	if err != nil {
		return "", nil, fmt.Errorf("get project: %w", err)
	}

	chapter, err := o.store.GetChapterByIndex(ctx, projectID, chapterIndex)
	if err != nil {
		return "", nil, fmt.Errorf("get chapter: %w", err)
	}

	if handler.OnStatus != nil {
		handler.OnStatus(fmt.Sprintf("Starting %s for Chapter %d: %s", opts.Mode, chapterIndex, chapter.Title))
	}
	log.Info().
		Str("mode", string(opts.Mode)).
		Str("project_id", projectID).
		Int64("chapter_index", chapterIndex).
		Str("title", chapter.Title).
		Msg("orchestrating chapter translation")

	// Save pre-translation checkpoint for rollback support
	_ = o.store.SaveCheckpoint(ctx, projectID, chapterIndex, "pre_translation", map[string]string{
		"chapter_id":         chapter.ID,
		"translated_content": chapter.TranslatedContent,
		"status":             chapter.Status,
	})

	rawContent := CleanRawContentForTranslation(chapter.RawContent)
	rawContent, _ = StripSemanticNoise(rawContent)

	var translatedText string
	switch opts.Mode {
	case ModeHierarchical3Pass:
		translatedText, err = o.translateHierarchical3Pass(ctx, projectID, chapterIndex, project.SourceLang, project.TargetLang, rawContent, opts, telemetry, handler)
	case ModeConcurrentDualAgent:
		translatedText, err = o.translateConcurrentDualAgent(ctx, projectID, chapterIndex, project.SourceLang, project.TargetLang, rawContent, opts, telemetry, handler)
	default:
		translatedText, err = o.translateSinglePass(ctx, projectID, chapterIndex, project.SourceLang, project.TargetLang, rawContent, opts, telemetry, handler)
	}

	if err != nil {
		return "", nil, err
	}

	// Sanitize any untranslated foreign script residuals
	translatedText, _ = SanitizeForeignResiduals(project.TargetLang, translatedText)

	// Strip any chatbot conversational meta-talk or markdown code fences
	translatedText = StripChatbotMetaTalk(translatedText)

	// Ensure HTML paragraph tags are intact and normalize structure if source had HTML
	translatedText = NormalizeHtmlParagraphs(chapter.RawContent, translatedText)

	duration := time.Since(startTime)
	totalRunes := len([]rune(translatedText))
	runesPerSec := 0.0
	tokensPerSec := 0.0
	if duration.Seconds() > 0 {
		runesPerSec = float64(totalRunes) / duration.Seconds()
		tokensPerSec = float64(telemetry.TotalTokens) / duration.Seconds()
	}

	telemetry.Duration = duration
	telemetry.TotalRunes = totalRunes
	telemetry.RunesPerSecond = runesPerSec
	telemetry.TokensPerSecond = tokensPerSec
	telemetry.EstimatedCostUSD = EstimateTokenCostUSD(o.model, telemetry.PromptTokens, telemetry.CompletionTokens)

	// Check if translation was partially paused via soft stop
	totalExpectedChunks := len(o.chunker.ChunkText(rawContent))
	isPaused := o.softStopChecker != nil && o.softStopChecker()

	targetStatus := "completed"
	if isPaused {
		targetStatus = "paused"
	}

	if handler.OnStatus != nil {
		if isPaused {
			handler.OnStatus(fmt.Sprintf("Translation paused gracefully. Progress saved (%d runes).", totalRunes))
		} else {
			handler.OnStatus("Saving translated chapter to database...")
		}
	}

	_ = o.store.UpdateChapterTranslation(ctx, chapter.ID, translatedText, targetStatus)
	if !isPaused {
		_, _ = o.memory.GenerateAndSaveL1Summary(ctx, o.client, o.model, projectID, chapterIndex, chapter.Title, translatedText)
	}

	// Save post-translation checkpoint
	_ = o.store.SaveCheckpoint(ctx, projectID, chapterIndex, "post_translation", map[string]string{
		"chapter_id":         chapter.ID,
		"translated_content": translatedText,
		"status":             targetStatus,
	})

	if isPaused {
		return translatedText, telemetry, fmt.Errorf("translation paused by user: chapter %d saved (%d runes, %d total chunks)", chapterIndex, totalRunes, totalExpectedChunks)
	}

	if handler.OnStatus != nil {
		handler.OnStatus(fmt.Sprintf("Completed Chapter %d (%d runes, %.1f runes/sec)", chapterIndex, totalRunes, runesPerSec))
	}

	return translatedText, telemetry, nil
}

// executeChunkWithTools executes completion for a chunk, engaging in a multi-turn tool calling loop
// if Agentic RAG is enabled and the model requests external tool executions.
func (o *TranslationOrchestrator) executeChunkWithTools(
	ctx context.Context,
	projectID string,
	chapterIndex int64,
	initialMessages []Message,
	opts TranslationOptions,
	telemetry *ExecutionTelemetry,
	handler *TranslationEventHandler,
) (string, error) {
	messages := make([]Message, len(initialMessages))
	copy(messages, initialMessages)

	maxIterations := 1
	var tools []ToolDefinition
	if opts.EnableAgenticRAG && o.toolReg != nil {
		isResearcherEnabled := true
		if o.skillChecker != nil && !o.skillChecker.IsSkillEnabled(ctx, projectID, "skill_agentic_researcher") {
			isResearcherEnabled = false
		}
		if isResearcherEnabled {
			o.toolReg.SetDecisionMode(opts.DecisionMode)
			o.toolReg.SetStyleGuide(opts.StyleGuide)
			tools = o.toolReg.GetAvailableTools()
			maxIterations = opts.MaxToolIterations
			if maxIterations <= 0 {
				maxIterations = 2
			}

			if len(tools) > 0 && len(messages) > 0 && messages[0].Role == RoleSystem {
				messages[0].Content += "\n\n[DYNAMIC AGENTIC TOOLS GUIDANCE]:" +
					"\nYou have access to dynamic lookup tools. USE THEM PROACTIVELY:" +
					"\n- Call 'lookup_character_relation' whenever dialogue occurs and pronoun/address terms between characters are unclear or unestablished." +
					"\n- Call 'lookup_world_lore' whenever encountering unfamiliar factions, ranks, magic tiers, technology, or world concepts." +
					"\n- Call 'search_book_context' to verify past events or foreshadowing from earlier chapters." +
					"\n- Call 'web_lookup' for obscure cultural idioms or slang." +
					"\n- Call 'ask_human_coworker' for critical translation dilemmas." +
					"\nAfter receiving the tool response, incorporate the retrieved facts and complete the natural Vietnamese translation."
			}
		}
	}

	var chunkResult string

	for iter := 0; iter < maxIterations; iter++ {
		if ctx.Err() != nil {
			return "", ctx.Err()
		}
		if o.softStopChecker != nil && o.softStopChecker() {
			return "", fmt.Errorf("soft stop requested")
		}
		req := CompletionRequest{
			Model:           o.model,
			Messages:        messages,
			Temperature:     0.3,
			Tools:           tools,
			ReasoningEffort: opts.ThinkingEffort,
		}

		resp, err := o.client.Generate(ctx, req)
		if err != nil {
			// Fallback without tools if endpoint does not support function declarations
			if len(tools) > 0 && (strings.Contains(err.Error(), "tool") || strings.Contains(err.Error(), "function")) {
				log.Warn().Err(err).Msg("model does not support tool calling, falling back without tools")
				req.Tools = nil
				tools = nil
				resp, err = o.client.Generate(ctx, req)
			}
			if err != nil {
				return "", err
			}
		}

		if telemetry != nil {
			telemetry.PromptTokens += resp.PromptTokens
			telemetry.CompletionTokens += resp.CompTokens
			telemetry.TotalTokens += resp.TotalTokens
			telemetry.ReasoningTokens += resp.ReasoningTokens
		}

		if resp.Thought != "" && handler != nil && handler.OnThought != nil {
			handler.OnThought(resp.Thought)
		}

		if len(resp.ToolCalls) > 0 && o.toolReg != nil {
			if telemetry != nil {
				telemetry.ToolCallsCount += len(resp.ToolCalls)
			}

			messages = append(messages, Message{
				Role:      RoleAssistant,
				Content:   resp.Content,
				ToolCalls: resp.ToolCalls,
			})

			for _, toolCall := range resp.ToolCalls {
				toolName := toolCall.Function.Name
				toolArgs := toolCall.Function.Arguments

				log.Info().
					Str("tool", toolName).
					Str("args", toolArgs).
					Int("iter", iter+1).
					Msg("agentic RAG invoked tool during translation")

				if handler != nil && handler.OnStatus != nil {
					handler.OnStatus(fmt.Sprintf("Agentic RAG: executing [%s]...", toolName))
				}

				toolResult, execErr := o.toolReg.Execute(ctx, projectID, chapterIndex, toolCall)
				if execErr != nil {
					log.Warn().Err(execErr).Str("tool", toolName).Msg("tool execution error, feeding error message to LLM")
					toolResult = fmt.Sprintf("Tool execution error: %v", execErr)
				}

				if handler != nil && handler.OnToolCall != nil {
					handler.OnToolCall(toolName, toolArgs, toolResult)
				}

				messages = append(messages, Message{
					Role:       RoleTool,
					Name:       toolName,
					ToolCallID: toolCall.ID,
					Content:    toolResult,
				})
			}
			continue
		}

		chunkResult = strings.TrimSpace(resp.Content)
		break
	}

	if chunkResult == "" && len(messages) > 0 {
		finalReq := CompletionRequest{
			Model:           o.model,
			Messages:        messages,
			Temperature:     0.3,
			ReasoningEffort: opts.ThinkingEffort,
		}
		finalResp, err := o.client.Generate(ctx, finalReq)
		if err != nil {
			return "", fmt.Errorf("final translation synthesis: %w", err)
		}
		chunkResult = strings.TrimSpace(finalResp.Content)
		if telemetry != nil {
			telemetry.PromptTokens += finalResp.PromptTokens
			telemetry.CompletionTokens += finalResp.CompTokens
			telemetry.TotalTokens += finalResp.TotalTokens
			telemetry.ReasoningTokens += finalResp.ReasoningTokens
		}
	}

	if chunkResult == "" {
		return "", fmt.Errorf("chunk produced empty translation after tool calling iterations")
	}

	return chunkResult, nil
}

// translateSinglePass executes direct single-pass chunk translation with selective context injection.
func (o *TranslationOrchestrator) translateSinglePass(
	ctx context.Context,
	projectID string,
	chapterIndex int64,
	sourceLang, targetLang, rawContent string,
	opts TranslationOptions,
	telemetry *ExecutionTelemetry,
	handler *TranslationEventHandler,
) (string, error) {
	chunks := o.chunker.ChunkText(rawContent)
	if len(chunks) == 0 {
		return "", fmt.Errorf("chapter %d raw content is empty", chapterIndex)
	}

	var translatedChunks []string
	var lastChunkTranslation string

	if opts.EnableR19 && o.r19Eng != nil {
		r19Cfg := o.r19Eng.Config()
		r19Cfg.TargetLang = targetLang
		o.r19Eng.SetConfig(r19Cfg)
	}

	for cIdx, chunk := range chunks {
		timeout := opts.TimeoutPerChunk
		if timeout <= 0 {
			timeout = 120 * time.Second
		}
		chunkCtx, cancel := context.WithTimeout(ctx, timeout)
		// Wrap with soft-stop-aware context so in-flight LLM calls are cancelled
		// within ~250ms of the user pressing "Soft Stop"
		chunkCtx, softCancel := o.makeSoftStopContext(chunkCtx)
		cancelBoth := func() { softCancel(); cancel() }

		if IsPureMarkupChunk(chunk.Content) {
			cancelBoth()
			cleanMarkup := StripChatbotMetaTalk(chunk.Content)
			if handler != nil && handler.OnStatus != nil {
				handler.OnStatus(fmt.Sprintf("Preserving pure markup/image segment (Chunk %d/%d)...", cIdx+1, len(chunks)))
			}
			translatedChunks = append(translatedChunks, cleanMarkup)
			lastChunkTranslation = cleanMarkup
			if handler != nil && handler.OnContent != nil {
				handler.OnContent(cleanMarkup)
			}
			continue
		}

		if handler.OnStatus != nil {
			handler.OnStatus(fmt.Sprintf("[Single-Pass] Chunk %d/%d (%d runes)...", cIdx+1, len(chunks), len([]rune(chunk.Content))))
		}

		var selectiveCtx string
		if o.graph != nil {
			if sc, err := o.graph.BuildSelectiveContext(chunkCtx, projectID, chapterIndex, chunk.Content); err == nil && sc != nil {
				selectiveCtx = sc.FormattedPrompt
				if len(sc.ActiveGlossary) > 0 && handler != nil && handler.OnToolCall != nil {
					var termsSummary []string
					for _, g := range sc.ActiveGlossary {
						termsSummary = append(termsSummary, fmt.Sprintf("• %s -> %s (%s)", g.SourceTerm, g.TargetTerm, g.Category))
					}
					handler.OnToolCall("match_glossary_terms", fmt.Sprintf("Segment %d/%d (matched %d glossary terms)", cIdx+1, len(chunks), len(sc.ActiveGlossary)), strings.Join(termsSummary, "\n"))
				}
			}
		}

		l0Sliding := o.memory.ExtractL0SlidingContext(lastChunkTranslation, 300)
		if opts.EnableR19 && o.r19Eng != nil {
			l0Sliding = o.r19Eng.SanitizeContext(l0Sliding)
		}
		sysPrompt := buildSystemPrompt(sourceLang, targetLang, "", selectiveCtx, l0Sliding)
		chunkText := chunk.Content

		var r19Res r19.MaskResult
		if opts.EnableR19 && o.r19Eng != nil {
			var notice string
			r19Res, notice = o.r19Eng.Prepare(chunk.Content)
			chunkText = r19Res.MaskedText
			if notice != "" {
				sysPrompt = sysPrompt + "\n\n" + notice
			}
		}
		if opts.StyleGuide != "" {
			sysPrompt += "\n\n[STYLE GUIDE & GUIDELINES]:\n" + opts.StyleGuide
		}
		if opts.WorldBibleContext != "" {
			sysPrompt += "\n\n" + opts.WorldBibleContext
		}
		if opts.VoiceContext != "" {
			sysPrompt += "\n\n" + opts.VoiceContext
		}
		if o.skillChecker != nil {
			if prompts := o.skillChecker.GetActiveSkillPrompts(chunkCtx, projectID, ""); prompts != "" {
				sysPrompt += "\n\n[MODULAR ACTIVE SKILL RULES]:\n" + prompts
			}
		}

		userPrompt := fmt.Sprintf("Translate the following segment:\n\n%s", chunkText)
		messages := []Message{
			{Role: RoleSystem, Content: sysPrompt},
			{Role: RoleUser, Content: userPrompt},
		}

		result, err := o.executeChunkWithTools(chunkCtx, projectID, chapterIndex, messages, opts, telemetry, handler)
		cancelBoth()

		if err != nil {
			if o.softStopChecker != nil && o.softStopChecker() {
				if handler.OnStatus != nil {
					handler.OnStatus(fmt.Sprintf("Soft-stop: pausing translation gracefully after chunk %d/%d.", cIdx, len(chunks)))
				}
				break
			}
			return "", fmt.Errorf("single-pass chunk %d: %w", cIdx+1, err)
		}

		result = StripChatbotMetaTalk(result)

		if opts.EnableR19 && o.r19Eng != nil {
			result = o.r19Eng.Restore(result, r19Res, false)
		}

		if handler.OnContent != nil {
			handler.OnContent(result)
		}

		translatedChunks = append(translatedChunks, result)
		lastChunkTranslation = result

		if o.softStopChecker != nil && o.softStopChecker() {
			if handler.OnStatus != nil {
				handler.OnStatus(fmt.Sprintf("Soft-stop: pausing translation gracefully after chunk %d/%d.", cIdx+1, len(chunks)))
			}
			break
		}
	}

	return strings.Join(translatedChunks, "\n\n"), nil
}

// translateHierarchical3Pass implements a 3-pass workflow: Draft -> Critique -> Polish.
func (o *TranslationOrchestrator) translateHierarchical3Pass(
	ctx context.Context,
	projectID string,
	chapterIndex int64,
	sourceLang, targetLang, rawContent string,
	opts TranslationOptions,
	telemetry *ExecutionTelemetry,
	handler *TranslationEventHandler,
) (string, error) {
	chunks := o.chunker.ChunkText(rawContent)
	if len(chunks) == 0 {
		return "", fmt.Errorf("chapter %d raw content is empty", chapterIndex)
	}

	var translatedChunks []string
	var lastChunkTranslation string

	if opts.EnableR19 && o.r19Eng != nil {
		r19Cfg := o.r19Eng.Config()
		r19Cfg.TargetLang = targetLang
		o.r19Eng.SetConfig(r19Cfg)
	}

	for cIdx, chunk := range chunks {
		timeout := opts.TimeoutPerChunk
		if timeout <= 0 {
			timeout = 180 * time.Second
		}
		chunkCtx, cancel := context.WithTimeout(ctx, timeout)
		// Wrap with soft-stop-aware context so in-flight LLM calls are cancelled
		// within ~250ms of the user pressing "Soft Stop"
		chunkCtx, softCancel := o.makeSoftStopContext(chunkCtx)
		cancelBoth := func() { softCancel(); cancel() }

		if IsPureMarkupChunk(chunk.Content) {
			cancelBoth()
			cleanMarkup := StripChatbotMetaTalk(chunk.Content)
			if handler != nil && handler.OnStatus != nil {
				handler.OnStatus(fmt.Sprintf("Preserving pure markup/image segment (Chunk %d/%d)...", cIdx+1, len(chunks)))
			}
			translatedChunks = append(translatedChunks, cleanMarkup)
			lastChunkTranslation = cleanMarkup
			if handler != nil && handler.OnContent != nil {
				handler.OnContent(cleanMarkup)
			}
			continue
		}

		var selectiveCtx string
		if o.graph != nil {
			if sc, err := o.graph.BuildSelectiveContext(chunkCtx, projectID, chapterIndex, chunk.Content); err == nil && sc != nil {
				selectiveCtx = sc.FormattedPrompt
				if len(sc.ActiveGlossary) > 0 && handler != nil && handler.OnToolCall != nil {
					var termsSummary []string
					for _, g := range sc.ActiveGlossary {
						termsSummary = append(termsSummary, fmt.Sprintf("• %s -> %s (%s)", g.SourceTerm, g.TargetTerm, g.Category))
					}
					handler.OnToolCall("match_glossary_terms", fmt.Sprintf("Segment %d/%d (matched %d glossary terms)", cIdx+1, len(chunks), len(sc.ActiveGlossary)), strings.Join(termsSummary, "\n"))
				}
			}
		}

		if handler.OnStatus != nil {
			handler.OnStatus(fmt.Sprintf("[Pass 1/3: Draft] Chunk %d/%d...", cIdx+1, len(chunks)))
		}

		l0Sliding := o.memory.ExtractL0SlidingContext(lastChunkTranslation, 300)
		if opts.EnableR19 && o.r19Eng != nil {
			l0Sliding = o.r19Eng.SanitizeContext(l0Sliding)
		}
		draftSysPrompt := buildSystemPrompt(sourceLang, targetLang, "", selectiveCtx, l0Sliding)
		chunkText := chunk.Content

		var r19Res r19.MaskResult
		if opts.EnableR19 && o.r19Eng != nil {
			var notice string
			r19Res, notice = o.r19Eng.Prepare(chunk.Content)
			chunkText = r19Res.MaskedText
			if notice != "" {
				draftSysPrompt = draftSysPrompt + "\n\n" + notice
			}
		}
		if opts.StyleGuide != "" {
			draftSysPrompt += "\n\n[STYLE GUIDE & GUIDELINES]:\n" + opts.StyleGuide
		}
		if opts.WorldBibleContext != "" {
			draftSysPrompt += "\n\n" + opts.WorldBibleContext
		}
		if opts.VoiceContext != "" {
			draftSysPrompt += "\n\n" + opts.VoiceContext
		}
		if o.skillChecker != nil {
			if prompts := o.skillChecker.GetActiveSkillPrompts(chunkCtx, projectID, ""); prompts != "" {
				draftSysPrompt += "\n\n[MODULAR ACTIVE SKILL RULES]:\n" + prompts
			}
		}

		draftText, err := o.executeChunkWithTools(chunkCtx, projectID, chapterIndex, []Message{
			{Role: RoleSystem, Content: draftSysPrompt},
			{Role: RoleUser, Content: fmt.Sprintf("Translate the following segment:\n\n%s", chunkText)},
		}, opts, telemetry, handler)
		if err != nil {
			cancelBoth()
			if o.softStopChecker != nil && o.softStopChecker() {
				if handler.OnStatus != nil {
					handler.OnStatus(fmt.Sprintf("Soft-stop: pausing translation gracefully after chunk %d/%d.", cIdx, len(chunks)))
				}
				break
			}
			return "", fmt.Errorf("pass 1 draft chunk %d: %w", cIdx+1, err)
		}

		// Pass 2: Shadow Critic & Quality Gate (Bypassed if skill_shadow_critic is disabled)
		var critiqueNotes string
		isCriticEnabled := true
		if o.skillChecker != nil && !o.skillChecker.IsSkillEnabled(chunkCtx, projectID, "skill_shadow_critic") {
			isCriticEnabled = false
		}

		if !isCriticEnabled {
			if handler.OnStatus != nil {
				handler.OnStatus(fmt.Sprintf("[Pass 2/3: Critique] Skipped (Skill disabled for Token Eco mode) for Chunk %d/%d", cIdx+1, len(chunks)))
			}
			critiqueNotes = "PASS"
		} else {
			if handler.OnStatus != nil {
				handler.OnStatus(fmt.Sprintf("[Pass 2/3: Critique] Inspecting Chunk %d/%d...", cIdx+1, len(chunks)))
			}

			criticSysPrompt := fmt.Sprintf("You are the Shadow Critic. Analyze this draft translation from '%s' to '%s' against the source text and selective context.\nInspect specifically for: (1) literal word-by-word translation, foreign idiom calques, or unnatural syntax that sounds like machine translation; (2) character address and pronoun errors; (3) glossary or world lore violations; (4) arbitrary mid-sentence capitalization; (5) illegal punctuation (e.g. ending sentences with semicolons ';'); (6) untranslated foreign/CJK characters.\nIf the draft is already natural, fluent, idiomatically sound, and obeys target orthography, reply with 'PASS'. Otherwise, provide detailed critique and the required idiomatic literary correction.\n\n[SELECTIVE CONTEXT]:\n%s", sourceLang, targetLang, selectiveCtx)
			if opts.StyleGuide != "" {
				criticSysPrompt += "\n\n[STYLE GUIDE & GUIDELINES]:\n" + opts.StyleGuide
			}
			if opts.WorldBibleContext != "" {
				criticSysPrompt += "\n\n" + opts.WorldBibleContext
			}
			if opts.VoiceContext != "" {
				criticSysPrompt += "\n\n" + opts.VoiceContext
			}
			if o.skillChecker != nil {
				if prompts := o.skillChecker.GetActiveSkillPrompts(chunkCtx, projectID, ""); prompts != "" {
					criticSysPrompt += "\n\n[MODULAR ACTIVE SKILL RULES]:\n" + prompts
				}
			}
			criticUserPrompt := fmt.Sprintf("Source:\n%s\n\nDraft Translation:\n%s", chunkText, draftText)
			criticResp, err := o.critic.client.Generate(chunkCtx, CompletionRequest{
				Model:       o.critic.model,
				Messages: []Message{
					{Role: RoleSystem, Content: criticSysPrompt},
					{Role: RoleUser, Content: criticUserPrompt},
				},
				Temperature:     0.1,
				ReasoningEffort: opts.ThinkingEffort,
			})
			if err != nil {
				log.Warn().Err(err).Msg("pass 2 critique failed, continuing to polish pass")
			} else {
				telemetry.PromptTokens += criticResp.PromptTokens
				telemetry.CompletionTokens += criticResp.CompTokens
				telemetry.TotalTokens += criticResp.TotalTokens
				telemetry.ReasoningTokens += criticResp.ReasoningTokens
			}
			if criticResp != nil {
				critiqueNotes = strings.TrimSpace(criticResp.Content)
				if criticResp.Thought != "" && handler != nil && handler.OnThought != nil {
					handler.OnThought(criticResp.Thought)
				}
				if critiqueNotes != "" && !strings.HasPrefix(strings.ToUpper(critiqueNotes), "PASS") && handler != nil && handler.OnToolCall != nil {
					handler.OnToolCall("shadow_critic_review", fmt.Sprintf("Segment %d/%d", cIdx+1, len(chunks)), critiqueNotes)
				}
			}
		}

		if strings.HasPrefix(strings.ToUpper(critiqueNotes), "PASS") {
			cancelBoth()
			finalDraft := StripChatbotMetaTalk(draftText)
			if opts.EnableR19 && o.r19Eng != nil {
				finalDraft = o.r19Eng.Restore(finalDraft, r19Res, false)
			}
			translatedChunks = append(translatedChunks, finalDraft)
			lastChunkTranslation = finalDraft
			if handler.OnContent != nil {
				handler.OnContent(finalDraft)
			}
			continue
		}

		telemetry.RevisedCount++
		if handler.OnStatus != nil {
			handler.OnStatus(fmt.Sprintf("[Pass 3/3: Polish] Polishing Chunk %d/%d...", cIdx+1, len(chunks)))
		}

		polishSysPrompt := fmt.Sprintf("You are an expert publishing copyeditor. Your task is to produce the final, polished translation from '%s' to '%s'.\nIncorporate the editorial critique to eliminate any word-by-word translationese, fix foreign idiom/syntax calques, and correct address terms, arbitrary capitalization, illegal punctuation (such as trailing semicolons), or terminology issues.\nEnsure 100%% of the output is in natural, expressive, idiomatic %s obeying strict orthography and clean typography without any raw source script or meta-commentary.", sourceLang, targetLang, NormalizeLanguage(targetLang))
		if opts.StyleGuide != "" {
			polishSysPrompt += "\n\n[STYLE GUIDE & GUIDELINES]:\n" + opts.StyleGuide
		}
		if opts.WorldBibleContext != "" {
			polishSysPrompt += "\n\n" + opts.WorldBibleContext
		}
		if opts.VoiceContext != "" {
			polishSysPrompt += "\n\n" + opts.VoiceContext
		}
		if o.skillChecker != nil {
			if prompts := o.skillChecker.GetActiveSkillPrompts(chunkCtx, projectID, ""); prompts != "" {
				polishSysPrompt += "\n\n[MODULAR ACTIVE SKILL RULES]:\n" + prompts
			}
		}
		polishUserPrompt := fmt.Sprintf("Source Text:\n%s\n\nDraft Translation:\n%s\n\nEditorial Critique:\n%s\n\nContext:\n%s", chunkText, draftText, critiqueNotes, selectiveCtx)
		polishResp, err := o.client.Generate(chunkCtx, CompletionRequest{
			Model:       o.model,
			Messages: []Message{
				{Role: RoleSystem, Content: polishSysPrompt},
				{Role: RoleUser, Content: polishUserPrompt},
			},
			Temperature:     0.2,
			ReasoningEffort: opts.ThinkingEffort,
		})
		cancelBoth()

		if err != nil {
			finalDraft := draftText
			if opts.EnableR19 && o.r19Eng != nil {
				finalDraft = o.r19Eng.Restore(finalDraft, r19Res, false)
			}
			translatedChunks = append(translatedChunks, finalDraft)
			lastChunkTranslation = finalDraft
			if o.softStopChecker != nil && o.softStopChecker() {
				if handler.OnStatus != nil {
					handler.OnStatus(fmt.Sprintf("Soft-stop: pausing translation gracefully after chunk %d/%d.", cIdx+1, len(chunks)))
				}
				break
			}
			continue
		}

		telemetry.PromptTokens += polishResp.PromptTokens
		telemetry.CompletionTokens += polishResp.CompTokens
		telemetry.TotalTokens += polishResp.TotalTokens
		telemetry.ReasoningTokens += polishResp.ReasoningTokens

		if polishResp.Thought != "" && handler != nil && handler.OnThought != nil {
			handler.OnThought(polishResp.Thought)
		}

		polishedResult := StripChatbotMetaTalk(strings.TrimSpace(polishResp.Content))
		if opts.EnableR19 && o.r19Eng != nil {
			polishedResult = o.r19Eng.Restore(polishedResult, r19Res, false)
		}
		if handler.OnContent != nil {
			handler.OnContent(polishedResult)
		}

		translatedChunks = append(translatedChunks, polishedResult)
		lastChunkTranslation = polishedResult

		if o.softStopChecker != nil && o.softStopChecker() {
			if handler.OnStatus != nil {
				handler.OnStatus(fmt.Sprintf("Soft-stop: pausing translation gracefully after chunk %d/%d.", cIdx+1, len(chunks)))
			}
			break
		}
	}

	return strings.Join(translatedChunks, "\n\n"), nil
}

// translateConcurrentDualAgent translates with live Actor-Critic streaming and real-time hot-patching.
func (o *TranslationOrchestrator) translateConcurrentDualAgent(
	ctx context.Context,
	projectID string,
	chapterIndex int64,
	sourceLang, targetLang, rawContent string,
	opts TranslationOptions,
	telemetry *ExecutionTelemetry,
	handler *TranslationEventHandler,
) (string, error) {
	chunks := o.chunker.ChunkText(rawContent)
	if len(chunks) == 0 {
		return "", fmt.Errorf("chapter %d raw content is empty", chapterIndex)
	}

	bus := NewDualStreamBus(32)
	defer bus.Close()

	var translatedChunks []string
	var lastChunkTranslation string

	if opts.EnableR19 && o.r19Eng != nil {
		r19Cfg := o.r19Eng.Config()
		r19Cfg.TargetLang = targetLang
		o.r19Eng.SetConfig(r19Cfg)
	}

	for cIdx, chunk := range chunks {
		timeout := opts.TimeoutPerChunk
		if timeout <= 0 {
			timeout = 120 * time.Second
		}
		chunkCtx, cancel := context.WithTimeout(ctx, timeout)
		// Wrap with soft-stop-aware context so in-flight LLM calls are cancelled
		// within ~250ms of the user pressing "Soft Stop"
		chunkCtx, softCancel := o.makeSoftStopContext(chunkCtx)
		cancelBoth := func() { softCancel(); cancel() }

		if IsPureMarkupChunk(chunk.Content) {
			cancelBoth()
			cleanMarkup := StripChatbotMetaTalk(chunk.Content)
			if handler != nil && handler.OnStatus != nil {
				handler.OnStatus(fmt.Sprintf("Preserving pure markup/image segment (Chunk %d/%d)...", cIdx+1, len(chunks)))
			}
			translatedChunks = append(translatedChunks, cleanMarkup)
			lastChunkTranslation = cleanMarkup
			if handler != nil && handler.OnContent != nil {
				handler.OnContent(cleanMarkup)
			}
			continue
		}

		if handler.OnStatus != nil {
			handler.OnStatus(fmt.Sprintf("[Dual-Agent] Chunk %d/%d (%d runes)...", cIdx+1, len(chunks), len([]rune(chunk.Content))))
		}

		var selectiveCtx string
		if o.graph != nil {
			if sc, err := o.graph.BuildSelectiveContext(chunkCtx, projectID, chapterIndex, chunk.Content); err == nil && sc != nil {
				selectiveCtx = sc.FormattedPrompt
				if len(sc.ActiveGlossary) > 0 && handler != nil && handler.OnToolCall != nil {
					var termsSummary []string
					for _, g := range sc.ActiveGlossary {
						termsSummary = append(termsSummary, fmt.Sprintf("• %s -> %s (%s)", g.SourceTerm, g.TargetTerm, g.Category))
					}
					handler.OnToolCall("match_glossary_terms", fmt.Sprintf("Segment %d/%d (matched %d glossary terms)", cIdx+1, len(chunks), len(sc.ActiveGlossary)), strings.Join(termsSummary, "\n"))
				}
			}
		}

		l0Sliding := o.memory.ExtractL0SlidingContext(lastChunkTranslation, 300)
		if opts.EnableR19 && o.r19Eng != nil {
			l0Sliding = o.r19Eng.SanitizeContext(l0Sliding)
		}
		sysPrompt := buildSystemPrompt(sourceLang, targetLang, "", selectiveCtx, l0Sliding)
		chunkText := chunk.Content

		var r19Res r19.MaskResult
		if opts.EnableR19 && o.r19Eng != nil {
			var notice string
			r19Res, notice = o.r19Eng.Prepare(chunk.Content)
			chunkText = r19Res.MaskedText
			if notice != "" {
				sysPrompt = sysPrompt + "\n\n" + notice
			}
		}
		if opts.StyleGuide != "" {
			sysPrompt += "\n\n[STYLE GUIDE & GUIDELINES]:\n" + opts.StyleGuide
		}
		if opts.WorldBibleContext != "" {
			sysPrompt += "\n\n" + opts.WorldBibleContext
		}
		if opts.VoiceContext != "" {
			sysPrompt += "\n\n" + opts.VoiceContext
		}
		if o.skillChecker != nil {
			if prompts := o.skillChecker.GetActiveSkillPrompts(chunkCtx, projectID, ""); prompts != "" {
				sysPrompt += "\n\n[MODULAR ACTIVE SKILL RULES]:\n" + prompts
			}
		}

		userPrompt := fmt.Sprintf("Translate the following segment:\n\n%s", chunkText)

		draftChunk, err := o.executeChunkWithTools(chunkCtx, projectID, chapterIndex, []Message{
			{Role: RoleSystem, Content: sysPrompt},
			{Role: RoleUser, Content: userPrompt},
		}, opts, telemetry, handler)
		if err != nil {
			cancelBoth()
			if o.softStopChecker != nil && o.softStopChecker() {
				if handler.OnStatus != nil {
					handler.OnStatus(fmt.Sprintf("Soft-stop: pausing translation gracefully after chunk %d/%d.", cIdx, len(chunks)))
				}
				break
			}
			return "", fmt.Errorf("dual-agent draft chunk %d: %w", cIdx+1, err)
		}

		bus.Publish(chunkCtx, StreamEvent{
			Index:  cIdx,
			Source: chunk.Content,
			Draft:  draftChunk,
			Kind:   EventDraftEmitted,
		})

		isCriticEnabled := true
		if o.skillChecker != nil && !o.skillChecker.IsSkillEnabled(chunkCtx, projectID, "skill_shadow_critic") {
			isCriticEnabled = false
		}

		approvedChunk := draftChunk
		if isCriticEnabled {
			verdict, err := o.critic.EvaluateSentence(chunkCtx, CriticRequest{
				SourceLang:       sourceLang,
				TargetLang:       targetLang,
				SelectiveContext: selectiveCtx,
				PrecedingContext: l0Sliding,
				OriginalSentence: chunkText,
				DraftTranslation: draftChunk,
				StyleGuide:       opts.StyleGuide,
				ThinkingEffort:    opts.ThinkingEffort,
			})
			cancelBoth()

			if verdict != nil {
				telemetry.PromptTokens += verdict.PromptTokens
				telemetry.CompletionTokens += verdict.CompTokens
				telemetry.TotalTokens += verdict.TotalTokens
				telemetry.ReasoningTokens += verdict.ReasoningTokens
				if verdict.Thought != "" && handler != nil && handler.OnThought != nil {
					handler.OnThought(verdict.Thought)
				}
			}

			if err == nil && verdict != nil && verdict.Action == CriticActionRevise && verdict.Correction != "" {
				approvedChunk = verdict.Correction
				bus.HotPatch(cIdx, verdict.Correction, verdict.Reason)
				telemetry.RevisedCount++

				log.Info().
					Int("chunk", cIdx+1).
					Str("reason", verdict.Reason).
					Msg("hot-patched chunk translation in memory")

				if handler.OnStatus != nil {
					handler.OnStatus(fmt.Sprintf("Shadow Critic hot-patched Chunk %d: %s", cIdx+1, verdict.Reason))
				}
				if handler.OnToolCall != nil {
					handler.OnToolCall("shadow_critic_hotpatch", fmt.Sprintf("Segment %d: %s", cIdx+1, verdict.Reason), verdict.Correction)
				}
			}
		} else {
			cancelBoth()
		}

		approvedChunk = StripChatbotMetaTalk(approvedChunk)

		if opts.EnableR19 && o.r19Eng != nil {
			approvedChunk = o.r19Eng.Restore(approvedChunk, r19Res, false)
		}

		if handler.OnContent != nil {
			handler.OnContent(approvedChunk)
		}

		translatedChunks = append(translatedChunks, approvedChunk)
		lastChunkTranslation = approvedChunk

		if o.softStopChecker != nil && o.softStopChecker() {
			if handler.OnStatus != nil {
				handler.OnStatus(fmt.Sprintf("Soft-stop: pausing translation gracefully after chunk %d/%d.", cIdx+1, len(chunks)))
			}
			break
		}
	}

	return strings.Join(translatedChunks, "\n\n"), nil
}
