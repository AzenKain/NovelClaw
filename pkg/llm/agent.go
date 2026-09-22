package llm

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	"novelclaw/pkg/graph"
	"novelclaw/pkg/storage"
	"novelclaw/pkg/websearch"
)

// TranslationEventHandler receives real-time streaming and tool execution events.
type TranslationEventHandler struct {
	OnThought  func(thoughtDelta string)
	OnContent  func(contentDelta string)
	OnStatus   func(status string)
	OnToolCall func(toolName string, input string, output string)
}

// TranslatorAgent coordinates translation workflows with multi-layer memory and Agentic RAG.
type TranslatorAgent struct {
	client          LLMClient
	model           string
	store           *storage.Storage
	chunker         *Chunker
	memory          *MemoryManager
	graph           *graph.ProgressiveGraph
	toolReg         *ToolRegistry
	researchEng     *websearch.ResearchEngine
	timeoutPerChunk time.Duration
}

// NewTranslatorAgent creates a new TranslatorAgent.
func NewTranslatorAgent(client LLMClient, model string, store *storage.Storage, researchEng *websearch.ResearchEngine) *TranslatorAgent {
	if researchEng == nil {
		researchEng = websearch.NewResearchEngine(30 * time.Second)
	}
	return &TranslatorAgent{
		client:          client,
		model:           model,
		store:           store,
		chunker:         NewChunker(2500),
		memory:          NewMemoryManager(store),
		graph:           graph.NewProgressiveGraph(store),
		toolReg:         NewToolRegistry(store, researchEng),
		researchEng:     researchEng,
		timeoutPerChunk: 120 * time.Second,
	}
}

// SetTimeoutPerChunk sets the execution timeout for a chunk translation.
func (a *TranslatorAgent) SetTimeoutPerChunk(d time.Duration) {
	if d <= 0 {
		d = 120 * time.Second
	}
	a.timeoutPerChunk = d
}

// TranslateChapter coordinates the end-to-end translation of a chapter.
func (a *TranslatorAgent) TranslateChapter(ctx context.Context, projectID string, chapterIndex int64, handler *TranslationEventHandler) (string, error) {
	if handler == nil {
		handler = &TranslationEventHandler{}
	}

	project, err := a.store.GetProject(ctx, projectID)
	if err != nil {
		return "", fmt.Errorf("get project: %w", err)
	}

	chapter, err := a.store.GetChapterByIndex(ctx, projectID, chapterIndex)
	if err != nil {
		return "", fmt.Errorf("get chapter: %w", err)
	}

	if handler.OnStatus != nil {
		handler.OnStatus(fmt.Sprintf("Preparing context for: %s", chapter.Title))
	}
	log.Info().
		Str("project_id", projectID).
		Int64("chapter_index", chapterIndex).
		Str("title", chapter.Title).
		Msg("starting chapter translation")

	timelineList, _ := a.store.ListTimelineByProject(ctx, projectID)
	var recentTimeline []string
	startIdx := len(timelineList) - 3
	if startIdx < 0 {
		startIdx = 0
	}
	for i := startIdx; i < len(timelineList); i++ {
		recentTimeline = append(recentTimeline, fmt.Sprintf("- Chapter %d: %s", timelineList[i].ChapterIndex, timelineList[i].SummaryText))
	}
	timelineContext := strings.Join(recentTimeline, "\n")

	cleanRaw := CleanRawContentForTranslation(chapter.RawContent)
	chunks := a.chunker.ChunkText(cleanRaw)
	if len(chunks) == 0 {
		return "", fmt.Errorf("chapter %d raw content is empty", chapterIndex)
	}

	var translatedChunks []string
	var lastChunkTranslation string

	for cIdx, chunk := range chunks {
		chunkCtx, chunkCancel := context.WithTimeout(ctx, a.timeoutPerChunk)

		if handler.OnStatus != nil {
			handler.OnStatus(fmt.Sprintf("Translating chunk %d/%d (%d runes)...", cIdx+1, len(chunks), len([]rune(chunk.Content))))
		}

		var selectiveContext string
		if a.graph != nil {
			selCtx, err := a.graph.BuildSelectiveContext(chunkCtx, projectID, chapterIndex, chunk.Content)
			if err == nil && selCtx != nil {
				selectiveContext = selCtx.FormattedPrompt
			}
		}

		l0Sliding := a.memory.ExtractL0SlidingContext(lastChunkTranslation, 300)
		systemPrompt := buildSystemPrompt(project.SourceLang, project.TargetLang, timelineContext, selectiveContext, l0Sliding)
		userPrompt := fmt.Sprintf("Translate the following segment:\n\n%s", chunk.Content)

		messages := []Message{
			{Role: RoleSystem, Content: systemPrompt},
			{Role: RoleUser, Content: userPrompt},
		}

		var chunkResult string
		maxToolIterations := 3

		for iter := 0; iter < maxToolIterations; iter++ {
			req := CompletionRequest{
				Model:       a.model,
				Messages:    messages,
				Temperature: 0.3,
				Tools:       a.toolReg.GetAvailableTools(),
			}

			resp, err := a.client.Generate(chunkCtx, req)
			if err != nil {
				if strings.Contains(err.Error(), "tool") || strings.Contains(err.Error(), "function") {
					req.Tools = nil
					resp, err = a.client.Generate(chunkCtx, req)
				}
				if err != nil {
					chunkCancel()
					return "", fmt.Errorf("generate chunk %d: %w", cIdx+1, err)
				}
			}

			if resp.Thought != "" && handler.OnThought != nil {
				handler.OnThought(resp.Thought)
			}

			if len(resp.ToolCalls) > 0 {
				messages = append(messages, Message{
					Role:      RoleAssistant,
					Content:   resp.Content,
					ToolCalls: resp.ToolCalls,
				})

				for _, toolCall := range resp.ToolCalls {
					log.Info().
						Str("tool", toolCall.Function.Name).
						Str("args", toolCall.Function.Arguments).
						Msg("agent invoked tool")

					if handler.OnStatus != nil {
						handler.OnStatus(fmt.Sprintf("Agent invoked tool [%s]", toolCall.Function.Name))
					}

					toolResult, execErr := a.toolReg.Execute(ctx, projectID, chapterIndex, toolCall)
					if execErr != nil {
						toolResult = fmt.Sprintf("Tool execution error: %v", execErr)
					}

					if handler.OnToolCall != nil {
						handler.OnToolCall(toolCall.Function.Name, toolCall.Function.Arguments, toolResult)
					}

					messages = append(messages, Message{
						Role:       RoleTool,
						Name:       toolCall.Function.Name,
						ToolCallID: toolCall.ID,
						Content:    toolResult,
					})
				}
				continue
			}

			chunkResult = strings.TrimSpace(resp.Content)
			if handler.OnContent != nil {
				handler.OnContent(chunkResult)
			}
			break
		}

		if chunkResult == "" {
			finalReq := CompletionRequest{
				Model:       a.model,
				Messages:    messages,
				Temperature: 0.3,
				Tools:       nil,
			}
			finalResp, finalErr := a.client.Generate(chunkCtx, finalReq)
			if finalErr == nil && strings.TrimSpace(finalResp.Content) != "" {
				chunkResult = strings.TrimSpace(finalResp.Content)
				if handler.OnContent != nil {
					handler.OnContent(chunkResult)
				}
			} else {
				chunkCancel()
				return "", fmt.Errorf("chunk %d produced empty translation", cIdx+1)
			}
		}

		chunkCancel()
		translatedChunks = append(translatedChunks, chunkResult)
		lastChunkTranslation = chunkResult
	}

	fullTranslation := strings.Join(translatedChunks, "\n\n")

	if handler.OnStatus != nil {
		handler.OnStatus("Saving translated chapter to database...")
	}
	err = a.store.UpdateChapterTranslation(ctx, chapter.ID, fullTranslation, "completed")
	if err != nil {
		return "", fmt.Errorf("update chapter translation: %w", err)
	}

	if handler.OnStatus != nil {
		handler.OnStatus("Summarizing episodic timeline (L1 memory)...")
	}
	_, _ = a.memory.GenerateAndSaveL1Summary(ctx, a.client, a.model, projectID, chapterIndex, chapter.Title, fullTranslation)

	if handler.OnStatus != nil {
		handler.OnStatus(fmt.Sprintf("Chapter %d translation completed successfully!", chapterIndex))
	}
	log.Info().
		Str("project_id", projectID).
		Int64("chapter_index", chapterIndex).
		Int("runes", len([]rune(fullTranslation))).
		Msg("chapter translation completed successfully")

	return fullTranslation, nil
}

// buildSystemPrompt constructs the system instructions for professional translation.
func buildSystemPrompt(sourceLang, targetLang, timelineContext, selectiveContext, l0Sliding string) string {
	var sb strings.Builder
	srcNorm := NormalizeLanguage(sourceLang)
	tgtNorm := NormalizeLanguage(targetLang)

	sb.WriteString(fmt.Sprintf("You are a professional literary translator specializing in light novels and web fiction, translating from '%s' (%s) to '%s' (%s).\n\n", srcNorm, sourceLang, tgtNorm, targetLang))
	sb.WriteString("CORE TRANSLATION PRINCIPLES:\n")
	sb.WriteString("1. Maintain natural, fluent, and highly expressive literary prose appropriate for the genre.\n")
	sb.WriteString("2. Anti-Translationese (Strictly No Word-by-Word): Never translate word-by-word. Deconstruct foreign grammatical structures and reconstruct them into native, idiomatic target syntax. Unroll unnatural passives, avoid literal preposition calques, and render idioms, metaphors, and slang by their true cultural and contextual meaning (e.g. 'in her birthday suit' means completely naked; 'triumphant smile' means a smug/victorious grin; 'for a second there' means almost/briefly thought).\n")
	sb.WriteString("3. Living Dialogue & Authentic Voice: Dialogue must sound like real living speech between characters, complete with natural tone, humor, and emotional nuance, never stiff or robotic.\n")
	sb.WriteString("4. Relational & Pronoun Coherence: Maintain strict honorific and address hierarchy throughout the scene.\n")
	sb.WriteString("5. STRICT HTML TAG & PARAGRAPH INTEGRITY (CRITICAL):\n")
	sb.WriteString("   - If the source text contains HTML markup (such as <p class=\"...\">...</p>), EVERY paragraph in your translation MUST be enclosed in matching <p class=\"...\">...</p> tags!\n")
	sb.WriteString("   - ABSOLUTELY NEVER drop <p> tags, NEVER output un-tagged plain text blocks, and NEVER merge multiple paragraphs into one. Every dialogue line and description paragraph in the source must have its own corresponding paragraph in the translation.\n")
	sb.WriteString("6. YOU ARE EQUIPPED WITH TOOLS: Use 'search_book_context' for past lore/character relations, and 'web_lookup' for unfamiliar slang, cultural allusions, or real-world names.\n")
	sb.WriteString("7. PURE MARKUP / IMAGES / COVER / NO-TEXT SEGMENTS (CRITICAL):\n")
	sb.WriteString("   - If the input segment contains only images or HTML/SVG markup (such as cover pages, illustrations, <img>, <svg>, <image>), OUTPUT THE EXACT PRESERVED MARKUP AS-IS.\n")
	sb.WriteString("   - ABSOLUTELY NEVER output conversational chatbot excuses, explanations, greetings, or meta-commentary (such as 'Here is the translation:', 'Sure!', 'As an AI language model...', 'Certainly!').\n\n")

	sb.WriteString(GetLanguageScriptRule(sourceLang, targetLang))
	sb.WriteString("\n")

	if selectiveContext != "" {
		sb.WriteString(fmt.Sprintf("\n[SELECTIVE CONTEXT INJECTION (L2/L3 - Dynamic Character Relations & Glossary)]:\n%s\n", selectiveContext))
	}
	if timelineContext != "" {
		sb.WriteString(fmt.Sprintf("\n[PREVIOUS CHAPTERS TIMELINE (L1)]:\n%s\n", timelineContext))
	}
	if l0Sliding != "" {
		sb.WriteString(fmt.Sprintf("\n[PRECEDING CONTEXT (L0)]:\n%s\n", l0Sliding))
	}

	sb.WriteString("\nOUTPUT ONLY THE TRANSLATED TEXT WITHOUT META COMMENTARY.")
	return sb.String()
}
