package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/wailsapp/wails/v3/pkg/application"

	"novelclaw/internal/dtos"
	"novelclaw/internal/gen/sqlc"
	"novelclaw/pkg/auditor"
	"novelclaw/pkg/benchmark"
	"novelclaw/pkg/llm"
	"novelclaw/pkg/skills"
	"novelclaw/pkg/soul"
	"novelclaw/pkg/storage"
	"novelclaw/pkg/voice"
	"novelclaw/pkg/worldbible"
)

// TranslationService exposes multi-mode chapter translation and benchmark execution to Wails v3 frontend.
type TranslationService struct {
	store          *storage.Storage
	client         llm.LLMClient
	soulCtrl       *soul.Controller
	skillReg       *skills.Registry
	wbMatcher      *worldbible.Matcher
	voiceReg       *voice.Registry
	plotAuditor    *auditor.PlotAuditor
	pendingPrompts sync.Map // promptID -> chan string
}

// NewTranslationService creates a new TranslationService.
func NewTranslationService(store *storage.Storage) *TranslationService {
	return &TranslationService{
		store:    store,
		soulCtrl: soul.NewController(),
	}
}

// NewTranslationServiceWithSoul creates a TranslationService bound to a shared SoulController.
func NewTranslationServiceWithSoul(store *storage.Storage, soulCtrl *soul.Controller) *TranslationService {
	if soulCtrl == nil {
		soulCtrl = soul.NewController()
	}
	return &TranslationService{
		store:    store,
		soulCtrl: soulCtrl,
	}
}

// SetSkillRegistry connects the modular skills registry.
func (s *TranslationService) SetSkillRegistry(reg *skills.Registry) {
	s.skillReg = reg
}

// SetWorldBibleMatcher connects the Aho-Corasick World Bible matcher.
func (s *TranslationService) SetWorldBibleMatcher(m *worldbible.Matcher) {
	s.wbMatcher = m
}

// SetVoiceRegistry connects the character voice persona registry.
func (s *TranslationService) SetVoiceRegistry(r *voice.Registry) {
	s.voiceReg = r
}

// SetPlotAuditor connects the plot consistency auditor.
func (s *TranslationService) SetPlotAuditor(a *auditor.PlotAuditor) {
	s.plotAuditor = a
}

// SetTestClient allows test suites to inject mock LLM clients without triggering Wails binding warnings.
func SetTestClient(s *TranslationService, client llm.LLMClient) {
	s.client = client
}

// resolveLLMClient dynamically constructs a resilient LLM client, automatically activating FallbackRouter
// with sequential model failovers if multiple active provider configurations are available.
func (s *TranslationService) resolveLLMClient(ctx context.Context, requestedModel string, timeoutSec int64, maxRetries int) (llm.LLMClient, string, error) {
	if s.client != nil {
		return s.client, "test-model", nil
	}
	if s.store == nil {
		return nil, "", fmt.Errorf("storage not initialized")
	}

	configs, err := s.store.ListActiveLLMConfigs(ctx)
	if err != nil || len(configs) == 0 {
		cfg, err := s.store.GetDefaultLLMConfig(ctx)
		if err != nil {
			return nil, "", fmt.Errorf("load default llm config: %w", err)
		}
		configs = []storage.LLMConfig{*cfg}
	}

	var defaultCfg *storage.LLMConfig
	for _, c := range configs {
		if c.IsDefault {
			copy := c
			defaultCfg = &copy
			break
		}
	}
	if defaultCfg == nil && len(configs) > 0 {
		defaultCfg = &configs[0]
	}

	createClient := func(cfg storage.LLMConfig, tSec int64) llm.LLMClient {
		timeout := time.Duration(tSec) * time.Second
		if timeout <= 0 {
			timeout = 120 * time.Second
		}
		if strings.Contains(cfg.ApiURL, "generativelanguage.googleapis.com") {
			return llm.NewGeminiClient(cfg.Token, timeout)
		}
		return llm.NewOpenAIClient(cfg.ApiURL, cfg.Token, timeout)
	}

	if len(configs) > 1 {
		routes := make([]llm.ModelRoute, 0, len(configs))
		if defaultCfg != nil {
			mName := defaultCfg.ModelName
			if requestedModel != "" {
				mName = requestedModel
			}
			tSec := defaultCfg.TimeoutSeconds
			if timeoutSec > 0 {
				tSec = timeoutSec
			}
			retries := int(defaultCfg.MaxRetries)
			if maxRetries > 0 {
				retries = maxRetries
			}
			if retries <= 0 {
				retries = 3
			}
			routes = append(routes, llm.ModelRoute{
				Name:       fmt.Sprintf("Primary (%s)", defaultCfg.ProviderName),
				Client:     createClient(*defaultCfg, tSec),
				Model:      mName,
				MaxRetries: retries,
			})
		}

		tierIdx := 2
		for _, c := range configs {
			if defaultCfg != nil && c.ID == defaultCfg.ID {
				continue
			}
			retries := int(c.MaxRetries)
			if retries <= 0 {
				retries = 3
			}
			routes = append(routes, llm.ModelRoute{
				Name:       fmt.Sprintf("Tier %d (%s)", tierIdx, c.ProviderName),
				Client:     createClient(c, c.TimeoutSeconds),
				Model:      c.ModelName,
				MaxRetries: retries,
			})
			tierIdx++
		}

		if len(routes) > 1 {
			return llm.NewFallbackRouter(routes), routes[0].Model, nil
		}
	}

	mName := defaultCfg.ModelName
	if requestedModel != "" {
		mName = requestedModel
	}
	tSec := defaultCfg.TimeoutSeconds
	if timeoutSec > 0 {
		tSec = timeoutSec
	}
	return createClient(*defaultCfg, tSec), mName, nil
}

// persistSoulMessage saves a Soul companion utterance into the NovelClaw thread and streams it
// to the UI through the canonical novelclaw:event channel, keeping a single source of truth.
func (s *TranslationService) persistSoulMessage(ctx context.Context, threadID, projectID, content, stepType string) {
	if threadID == "" || s.store == nil || content == "" {
		return
	}
	msg, err := s.store.SaveNovelClawMessage(ctx, sqlc.CreateNovelClawMessageParams{
		ID:                fmt.Sprintf("nc_soul_%d", time.Now().UnixNano()),
		ThreadID:          threadID,
		ProjectID:         projectID,
		Sender:            "novelclaw",
		Role:              "assistant",
		Content:           content,
		StepType:          sql.NullString{String: stepType, Valid: true},
		StepStatus:        sql.NullString{String: "completed", Valid: true},
		IsCollapsed:       0,
		IsArchivedCompact: 0,
		TokenCount:        int64(len(content) / 4),
	})
	if err != nil {
		log.Warn().Err(err).Msg("failed to persist soul message")
		return
	}
	if app := application.Get(); app != nil && app.Event != nil {
		app.Event.Emit("novelclaw:event", map[string]any{
			"type": "message_created",
			"data": map[string]any{
				"thread_id": threadID,
				"message":   dtos.ToNovelClawMessageDTO(msg),
			},
		})
	}
}

// TranslateChapter coordinates chapter translation using the selected mode and options.
func (s *TranslationService) TranslateChapter(ctx context.Context, req dtos.StartTranslationRequest) (*dtos.ExecutionTelemetryDTO, error) {
	client := s.client
	modelName := "default-model"

	if client == nil {
		var err error
		client, modelName, err = s.resolveLLMClient(ctx, req.Options.CriticModel, req.Options.TimeoutSeconds, req.Options.MaxRetries)
		if err != nil {
			return nil, fmt.Errorf("load llm client with fallback: %w", err)
		}
	}

	jobCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	jobID := fmt.Sprintf("job_%s_%d", req.ProjectID, req.ChapterIndex)
	if s.soulCtrl != nil {
		s.soulCtrl.RegisterJob(jobID, req.ProjectID, req.ChapterIndex, cancel)
		defer s.soulCtrl.UnregisterJob(jobID)
	}

	var mainThreadID string
	if s.store != nil {
		if mainThread, err := s.store.GetOrCreateMainThread(jobCtx, req.ProjectID); err == nil {
			mainThreadID = mainThread.ID
		}
	}

	orch := llm.NewTranslationOrchestrator(client, modelName, s.store, nil)
	if s.skillReg != nil {
		orch.SetSkillChecker(s.skillReg)
	}
	opts := req.Options.ToDomainOptions()

	var activeSoul soul.Soul
	if s.soulCtrl != nil {
		activeSoul = s.soulCtrl.GetProjectSoul(req.ProjectID)
		if activeSoul.SystemPromptAddon != "" {
			if opts.StyleGuide != "" {
				opts.StyleGuide += "\n\n[ASSISTANT PERSONA & GUIDELINES]:\n" + activeSoul.SystemPromptAddon
			} else {
				opts.StyleGuide = "[ASSISTANT PERSONA & GUIDELINES]:\n" + activeSoul.SystemPromptAddon
			}
		}
		if activeSoul.Greeting != "" {
			s.persistSoulMessage(jobCtx, mainThreadID, req.ProjectID, activeSoul.Greeting, "soul_greeting")
		}
		orch.SetSoftStopChecker(func() bool {
			return s.soulCtrl.IsSoftStopRequested(jobID)
		})
		if patch := s.soulCtrl.GetJobStylePatch(jobID); patch != "" {
			if opts.StyleGuide != "" {
				opts.StyleGuide += "\n[Hot-Steered Steering Override]: " + patch
			} else {
				opts.StyleGuide = "[Hot-Steered Steering Override]: " + patch
			}
		}
	}

	if mainThreadID != "" && s.store != nil {
		startContent := fmt.Sprintf("🚀 **Starting translation of Chapter %d** (Mode: `%s`)\nActivating the Agentic RAG pipeline and connecting knowledge...", req.ChapterIndex, opts.Mode)
		startMsg, _ := s.store.SaveNovelClawMessage(jobCtx, sqlc.CreateNovelClawMessageParams{
			ID:                fmt.Sprintf("nc_start_%d", time.Now().UnixNano()),
			ThreadID:          mainThreadID,
			ProjectID:         req.ProjectID,
			Sender:            "novelclaw",
			Role:              "assistant",
			Content:           startContent,
			StepType:          sql.NullString{String: "step_status", Valid: true},
			StepStatus:        sql.NullString{String: "running", Valid: true},
			IsCollapsed:       0,
			IsArchivedCompact: 0,
			TokenCount:        int64(len(startContent) / 4),
		})
		if app := application.Get(); app != nil && app.Event != nil {
			app.Event.Emit("novelclaw:event", map[string]any{
				"type": "message_created",
				"data": map[string]any{
					"thread_id": mainThreadID,
					"message":   dtos.ToNovelClawMessageDTO(startMsg),
				},
			})
		}
	}

	// Inject World Bible Lorebook context if matched
	if s.wbMatcher != nil {
		ch, err := s.store.GetChapterByIndex(jobCtx, req.ProjectID, req.ChapterIndex)
		if err == nil && ch.RawContent != "" {
			matches := s.wbMatcher.MatchContext(jobCtx, req.ProjectID, ch.RawContent)
			if len(matches) > 0 {
				opts.WorldBibleContext = s.wbMatcher.FormatContextPrompt(matches)
			}
		}
	}

	// Inject Character Voice profiles if matched
	if s.voiceReg != nil {
		// Auto-sync from project entities if registry is empty for this project
		if len(s.voiceReg.ListVoices(req.ProjectID)) == 0 && s.store != nil {
			if ents, err := s.store.ListEntitiesByProject(jobCtx, req.ProjectID); err == nil && len(ents) > 0 {
				s.voiceReg.SyncFromEntities(req.ProjectID, ents)
			}
		}

		ch, err := s.store.GetChapterByIndex(jobCtx, req.ProjectID, req.ChapterIndex)
		if err == nil && ch.RawContent != "" {
			voices := s.voiceReg.MatchVoicesInText(jobCtx, req.ProjectID, ch.RawContent)
			if len(voices) > 0 {
				opts.VoiceContext = voice.FormatVoicePrompt(voices)
			}
		}
	}

	handler := &llm.TranslationEventHandler{
		OnThought: func(thoughtDelta string) {
			if app := application.Get(); app != nil && app.Event != nil {
				app.Event.Emit("translation:thought", map[string]any{
					"project_id":    req.ProjectID,
					"chapter_index": req.ChapterIndex,
					"thought":       thoughtDelta,
				})
				app.Event.Emit("novelclaw:event", map[string]any{
					"type": "thinking",
					"data": map[string]any{
						"thread_id": mainThreadID,
						"content":   thoughtDelta,
					},
				})
			}
		},
		OnToolCall: func(toolName string, input string, output string) {
			if app := application.Get(); app != nil && app.Event != nil {
				app.Event.Emit("translation:tool_call", map[string]any{
					"project_id":    req.ProjectID,
					"chapter_index": req.ChapterIndex,
					"tool_name":     toolName,
					"input":         input,
					"output":        output,
				})
			}
			if mainThreadID != "" && s.store != nil {
				toolJSON, _ := json.Marshal(map[string]any{
					"tool_name": toolName,
					"input":     input,
					"output":    output,
				})
				toolMsg, _ := s.store.SaveNovelClawMessage(jobCtx, sqlc.CreateNovelClawMessageParams{
					ID:                fmt.Sprintf("nc_tool_%d", time.Now().UnixNano()),
					ThreadID:          mainThreadID,
					ProjectID:         req.ProjectID,
					Sender:            "novelclaw",
					Role:              "tool",
					Content:           fmt.Sprintf("Executing tool: %s", toolName),
					StepType:          sql.NullString{String: "tool_call", Valid: true},
					StepStatus:        sql.NullString{String: "completed", Valid: true},
					ActionCallJson:    sql.NullString{String: string(toolJSON), Valid: true},
					IsCollapsed:       1,
					IsArchivedCompact: 0,
					TokenCount:        0,
				})
				if app := application.Get(); app != nil && app.Event != nil {
					app.Event.Emit("novelclaw:event", map[string]any{
						"type": "message_created",
						"data": map[string]any{
							"thread_id": mainThreadID,
							"message":   dtos.ToNovelClawMessageDTO(toolMsg),
						},
					})
				}
			}
		},
		OnStatus: func(status string) {
			if app := application.Get(); app != nil && app.Event != nil {
				app.Event.Emit("translation:status", map[string]any{
					"project_id":    req.ProjectID,
					"chapter_index": req.ChapterIndex,
					"status":        status,
				})
				app.Event.Emit("novelclaw:event", map[string]any{
					"type": "step_badge",
					"data": map[string]any{
						"thread_id": mainThreadID,
						"tool_name": status,
						"status":    "running",
					},
				})
			}
		},
		OnContent: func(contentDelta string) {
			if app := application.Get(); app != nil && app.Event != nil {
				app.Event.Emit("translation:chunk_content", map[string]any{
					"project_id":    req.ProjectID,
					"chapter_index": req.ChapterIndex,
					"content":       contentDelta,
				})
			}
		},
	}

	askHumanCallback := func(promptCtx context.Context, question string, options []string, contextSnippet string) (string, error) {
		promptID := fmt.Sprintf("ask_%s_%d_%d", req.ProjectID, req.ChapterIndex, time.Now().UnixNano())
		ch := make(chan string, 1)
		s.pendingPrompts.Store(promptID, ch)
		defer s.pendingPrompts.Delete(promptID)

		if app := application.Get(); app != nil && app.Event != nil {
			app.Event.Emit("translation:ask_human", map[string]any{
				"job_id":          promptID,
				"project_id":      req.ProjectID,
				"chapter_index":   req.ChapterIndex,
				"question":        question,
				"options":         options,
				"context_snippet": contextSnippet,
				"timeout_seconds": 60,
			})
			if s.soulCtrl != nil && activeSoul.OnConfused != "" {
				s.persistSoulMessage(jobCtx, mainThreadID, req.ProjectID,
					fmt.Sprintf("%s\n\n*(Point of uncertainty: %s)*", activeSoul.OnConfused, question), "soul_confused")
			}
		}

		timer := time.NewTimer(60 * time.Second)
		defer timer.Stop()

		select {
		case answer := <-ch:
			s.autoLearnResolution(promptCtx, req.ProjectID, req.ChapterIndex, question, answer, options)
			if app := application.Get(); app != nil && app.Event != nil {
				app.Event.Emit("translation:ask_human_resolved", map[string]any{
					"job_id": promptID,
					"answer": answer,
				})
			}
			return answer, nil
		case <-timer.C:
			fallback := "Default option"
			if len(options) > 0 {
				fallback = options[0]
			}
			// Transient fallback only - DO NOT call autoLearnResolution because human did not approve!
			if app := application.Get(); app != nil && app.Event != nil {
				app.Event.Emit("translation:ask_human_resolved", map[string]any{
					"job_id":    promptID,
					"answer":    fallback,
					"timed_out": true,
				})
			}
			return fallback, nil
		case <-promptCtx.Done():
			return "", promptCtx.Err()
		case <-jobCtx.Done():
			return "", jobCtx.Err()
		}
	}
	orch.SetAskHumanHandler(askHumanCallback)

	translatedText, telemetry, err := orch.TranslateChapter(jobCtx, req.ProjectID, req.ChapterIndex, opts, handler)
	if err != nil {
		return nil, err
	}

	// Run Plot Consistency Auditor
	if s.plotAuditor != nil && translatedText != "" {
		warnings, _ := s.plotAuditor.AuditChapter(jobCtx, req.ProjectID, req.ChapterIndex, translatedText)
		if len(warnings) > 0 {
			if app := application.Get(); app != nil && app.Event != nil {
				app.Event.Emit("auditor:plot_warning", map[string]any{
					"project_id":    req.ProjectID,
					"chapter_index": req.ChapterIndex,
					"warnings":      warnings,
				})
			}
		}
	}

	if mainThreadID != "" && s.store != nil && telemetry != nil {
		telemetryJSON, _ := json.Marshal(telemetry)
		summaryText := fmt.Sprintf("**Completed Chapter %d Translation**\n- Speed: **%.1f tok/s** (%.1f runes/s)\n- Output volume: %d runes\n- Shadow Critic passes: %d\n- Agentic tools executed: %d\n- Tokens consumed: %d",
			req.ChapterIndex, telemetry.TokensPerSecond, telemetry.RunesPerSecond, telemetry.TotalRunes, telemetry.RevisedCount, telemetry.ToolCallsCount, telemetry.TotalTokens)
		if s.soulCtrl != nil && activeSoul.OnSuccess != "" {
			summaryText += "\n\n" + activeSoul.OnSuccess
		}
		doneMsg, _ := s.store.SaveNovelClawMessage(jobCtx, sqlc.CreateNovelClawMessageParams{
			ID:                fmt.Sprintf("nc_done_%d", time.Now().UnixNano()),
			ThreadID:          mainThreadID,
			ProjectID:         req.ProjectID,
			Sender:            "novelclaw",
			Role:              "assistant",
			Content:           summaryText,
			StepType:          sql.NullString{String: "translation_summary", Valid: true},
			StepStatus:        sql.NullString{String: "completed", Valid: true},
			ActionCallJson:    sql.NullString{String: string(telemetryJSON), Valid: true},
			IsCollapsed:       0,
			IsArchivedCompact: 0,
			TokenCount:        int64(telemetry.TotalTokens),
		})
		if app := application.Get(); app != nil && app.Event != nil {
			app.Event.Emit("novelclaw:event", map[string]any{
				"type": "message_created",
				"data": map[string]any{
					"thread_id": mainThreadID,
					"message":   dtos.ToNovelClawMessageDTO(doneMsg),
				},
			})
		}
	}

	return dtos.ToTelemetryDTO(telemetry), nil
}

// ResolveAskHuman sends the human decision back to the waiting translation worker.
func (s *TranslationService) ResolveAskHuman(jobID string, answer string) bool {
	val, ok := s.pendingPrompts.Load(jobID)
	if !ok {
		return false
	}
	ch, ok := val.(chan string)
	if !ok {
		return false
	}
	select {
	case ch <- strings.TrimSpace(answer):
		s.pendingPrompts.Delete(jobID)
		return true
	default:
		return false
	}
}

// autoLearnResolution automatically records resolved dilemmas into Glossary (L3) or Entity Graph (L2).
func (s *TranslationService) autoLearnResolution(ctx context.Context, projectID string, chapterIndex int64, question string, answer string, options []string) {
	if s.store == nil || projectID == "" || strings.TrimSpace(answer) == "" {
		return
	}

	trimmedAnswer := llm.CleanAnswer(answer)
	if trimmedAnswer == "" {
		return
	}

	// 1. Extract quoted terms from question
	reQuote := regexp.MustCompile(`["“'‘]([^"“'‘]+)["”'’]`)
	matches := reQuote.FindAllStringSubmatch(question, -1)

	// Filter matches: ignore any quoted string that is one of the options or equals the answer
	var candidates []string
	for _, m := range matches {
		term := strings.TrimSpace(m[1])
		if term == "" {
			continue
		}
		isOption := false
		for _, opt := range options {
			if strings.EqualFold(term, strings.TrimSpace(opt)) {
				isOption = true
				break
			}
		}
		if isOption || strings.EqualFold(term, trimmedAnswer) {
			continue
		}
		candidates = append(candidates, term)
	}

	lowerQ := strings.ToLower(question)
	// NOTE: the Vietnamese keywords below are intentional input-matching data
	// (they detect relationship dilemmas phrased in Vietnamese), not display text.
	isRelationQuestion := strings.Contains(lowerQ, "gọi") ||
		strings.Contains(lowerQ, "xưng hô") ||
		strings.Contains(lowerQ, "quan hệ") ||
		strings.Contains(lowerQ, "address") ||
		strings.Contains(lowerQ, "call")

	if isRelationQuestion && len(candidates) >= 2 {
		// Two entities dilemma -> character relation L2
		charA := candidates[0]
		charB := candidates[1]
		if charA != "" && charB != "" {
			_ = s.store.UpsertRelation(ctx, sqlc.UpsertRelationParams{
				ID:           "hitl_rel_" + strings.ReplaceAll(uuid.New().String(), "-", "")[:12],
				ProjectID:    projectID,
				FromChar:     charA,
				ToChar:       charB,
				CallAs:       trimmedAnswer,
				SelfCallAs:   "",
				SinceChapter: chapterIndex,
				Tone:         "respectful",
			})
		}
	} else if len(candidates) >= 1 {
		// Single term (or first non-option candidate) -> Glossary L3
		sourceTerm := candidates[0]
		if sourceTerm != "" && !strings.EqualFold(sourceTerm, trimmedAnswer) {
			_ = s.store.UpsertGlossaryTerm(ctx, sqlc.UpsertGlossaryTermParams{
				ID:         "hitl_" + strings.ReplaceAll(uuid.New().String(), "-", "")[:12],
				ProjectID:  projectID,
				SourceTerm: sourceTerm,
				TargetTerm: trimmedAnswer,
				Category:   "hitl_decision",
				Notes:      fmt.Sprintf("Auto-learned from HITL decision at chapter %d: %s", chapterIndex, question),
			})
		}
	}
}

// CancelTranslation requests an immediate emergency abort of an in-flight translation job.
func (s *TranslationService) CancelTranslation(jobID string) bool {
	if s.soulCtrl == nil {
		return false
	}
	return s.soulCtrl.RequestHardAbort(jobID)
}

// SoftStopTranslation requests an active translation job to stop safely after the current segment.
func (s *TranslationService) SoftStopTranslation(jobID string) bool {
	if s.soulCtrl == nil {
		return false
	}
	return s.soulCtrl.RequestSoftStop(jobID)
}

// RunBenchmark executes a comparative benchmark across selected modes and returns a report.
func (s *TranslationService) RunBenchmark(ctx context.Context, req dtos.StartBenchmarkRequest) (*dtos.BenchmarkReportDTO, error) {
	client := s.client
	modelName := "default-model"

	if client == nil {
		var err error
		client, modelName, err = s.resolveLLMClient(ctx, "", 120, 3)
		if err != nil {
			return nil, fmt.Errorf("load llm client for benchmark: %w", err)
		}
	}

	runner := benchmark.NewRunner(s.store, client, modelName)
	var modes []llm.TranslationMode
	for _, m := range req.Modes {
		modes = append(modes, llm.TranslationMode(m))
	}

	report, err := runner.RunBenchmark(ctx, req.ProjectID, req.ChapterIndices, modes)
	if err != nil {
		return nil, err
	}

	return dtos.ToBenchmarkReportDTO(report), nil
}

// PreviewStyle translates a raw excerpt with the specified style instructions using the configured LLM.
func (s *TranslationService) PreviewStyle(ctx context.Context, req dtos.PreviewStyleRequest) (*dtos.PreviewStyleResponse, error) {
	if strings.TrimSpace(req.RawContent) == "" {
		return &dtos.PreviewStyleResponse{
			Success: false,
			Error:   "Please provide source text to preview the translation.",
		}, nil
	}

	client := s.client
	modelName := "default-model"

	if client == nil {
		var err error
		client, modelName, err = s.resolveLLMClient(ctx, "", 60, 2)
		if err != nil {
			return &dtos.PreviewStyleResponse{
				Success: false,
				Error:   "No model configured. Please open Settings to configure your API key: " + err.Error(),
			}, nil
		}
	}

	sourceLang := req.SourceLang
	if sourceLang == "" {
		sourceLang = "Japanese"
	}
	targetLang := req.TargetLang
	if targetLang == "" {
		targetLang = "Vietnamese"
	}

	styleInstructions := req.StylePrompt
	if styleInstructions == "" {
		styleInstructions = "Natural, expressive, and precise prose in the spirit of Light Novel."
	}

	srcName := llm.NormalizeLanguage(sourceLang)
	tgtName := llm.NormalizeLanguage(targetLang)
	scriptRule := llm.GetLanguageScriptRule(sourceLang, targetLang)

	systemPrompt := fmt.Sprintf(
		"You are an elite literary translator specializing in Asian light novels and web novels (%s to %s).\nStyle Requirements: %s\n\n%s\n\nTranslate the user's excerpt into highly expressive literary %s following the style guide and script rules.\nOutput ONLY the final translated %s text, with no explanations, notes, or conversational filler.",
		srcName,
		tgtName,
		styleInstructions,
		scriptRule,
		tgtName,
		tgtName,
	)

	start := time.Now()
	resp, err := client.Generate(ctx, llm.CompletionRequest{
		Model: modelName,
		Messages: []llm.Message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: req.RawContent},
		},
		Temperature: 0.7,
		MaxTokens:   1500,
	})
	latency := time.Since(start).Milliseconds()

	if err != nil {
		return &dtos.PreviewStyleResponse{
			Success:   false,
			LatencyMs: latency,
			ModelUsed: modelName,
			Error:     fmt.Sprintf("LLM error: %v", err),
		}, nil
	}

	return &dtos.PreviewStyleResponse{
		Success:           true,
		TranslatedContent: strings.TrimSpace(resp.Content),
		ModelUsed:         modelName,
		LatencyMs:         latency,
	}, nil
}

// isNarrativeChapter filters out non-story chapters such as covers, table of contents, inserts, illustrations,
// colophon, copyright pages, and image-only chapters.
func isNarrativeChapter(title, rawContent string) (bool, string) {
	lowerTitle := strings.ToLower(title)
	skipPatterns := []string{
		"cover", "bìa", "insert", "title page", "copyright", "colophon",
		"table of contents", "contents", "content", "toc", "mục lục", "muc luc",
		"目次", "目录", "封面", "插图", "插畫", "版权", "制作人员",
		"illustration", "illustrations", "gallery", "pinup", "color insert",
		"newsletter", "bản quyền", "afterword", "hậu ký", "hau ky", "lời bạt", "loi bat",
		"あとがき", "奥付", "profile", "character", "giới thiệu nhân vật", "nhân vật",
	}
	for _, p := range skipPatterns {
		if strings.Contains(lowerTitle, p) {
			return false, ""
		}
	}

	cleaned := llm.CleanRawContentForTranslation(rawContent)
	text := storage.StripHTML(cleaned)
	text = strings.TrimSpace(text)
	runes := []rune(text)

	// Image tags check: if rawContent has image markers and clean text is small (< 350 runes)
	hasImages := strings.Contains(rawContent, "<img") || strings.Contains(rawContent, "<image") || strings.Contains(rawContent, "<svg")
	if hasImages && len(runes) < 350 {
		return false, ""
	}

	// Substantive narrative prose check (> 300 runes)
	if len(runes) < 300 {
		return false, ""
	}

	// Content-level TOC check (e.g. text starts with CONTENTS, 目次, or has table of contents structure)
	headLen := 100
	if len(runes) < headLen {
		headLen = len(runes)
	}
	head := strings.ToLower(string(runes[:headLen]))
	if strings.Contains(head, "contents") || strings.Contains(head, "目次") || strings.Contains(head, "mục lục") || strings.Contains(head, "table of contents") {
		return false, ""
	}

	return true, text
}

// AutoScoutStyle analyzes opening narrative chapters of a novel and generates a tailored style guide.
func (s *TranslationService) AutoScoutStyle(ctx context.Context, req dtos.AutoScoutStyleRequest) (*dtos.AutoScoutStyleResponse, error) {
	if req.ProjectID == "" {
		return &dtos.AutoScoutStyleResponse{
			Success: false,
			Error:   "Project ID is required.",
		}, nil
	}

	chapters, err := s.store.ListChaptersByProject(ctx, req.ProjectID)
	if err != nil {
		return &dtos.AutoScoutStyleResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to load chapter list: %v", err),
		}, nil
	}

	// Filter and sample the first 3 substantive narrative chapters (skip covers, TOCs, inserts, image-only)
	var sampleSnippets []string
	var firstNarrativeExcerpt string
	chaptersSampled := 0

	for _, ch := range chapters {
		isNarrative, text := isNarrativeChapter(ch.Title, ch.RawContent)
		if !isNarrative {
			continue
		}

		sampleLen := 1200
		runes := []rune(text)
		if len(runes) < sampleLen {
			sampleLen = len(runes)
		}
		excerpt := string(runes[:sampleLen])

		if firstNarrativeExcerpt == "" {
			firstNarrativeExcerpt = excerpt
		}

		chaptersSampled++
		sampleSnippets = append(sampleSnippets, fmt.Sprintf("[Chapter %d: %s]\n%s", ch.ChapterIndex, ch.Title, excerpt))

		if chaptersSampled >= 3 {
			break
		}
	}

	if chaptersSampled == 0 {
		return &dtos.AutoScoutStyleResponse{
			Success: false,
			Error:   "No narrative chapter with literary content found to analyze the style.",
		}, nil
	}

	client := s.client
	modelName := "default-model"

	if client == nil {
		var err error
		client, modelName, err = s.resolveLLMClient(ctx, "", 60, 2)
		if err != nil {
			return &dtos.AutoScoutStyleResponse{
				Success: false,
				Error:   "No model configured. Please open Settings to configure your API key: " + err.Error(),
			}, nil
		}
	}

	sourceLang := req.SourceLang
	if sourceLang == "" {
		sourceLang = "Japanese"
	}
	targetLang := req.TargetLang
	if targetLang == "" {
		targetLang = "Vietnamese"
	}

	srcName := llm.NormalizeLanguage(sourceLang)
	tgtName := llm.NormalizeLanguage(targetLang)

	systemPrompt := fmt.Sprintf(`You are an elite Chief Literary Scout and Master Translation Director specializing in Asian Light Novels and Web Novels (%s to %s).
Your mission is to read excerpts from the opening narrative chapters of this novel, deeply analyze the author's unique voice, narrative cadence, genre flavor, and relationship tone, and synthesize a specialized, custom Translation Style Profile tailored specifically for this book.

Examine:
1. Narrative Persona & Tone: Is it comedic/sarcastic, introspective, dark/gritty, whimsical, or classical/solemn?
2. Sentence Cadence & Pacing: Short/punchy, highly descriptive, rapid-fire dialogue, or poetic?
3. Dialogue & Relational Dynamics: Honorifics, banter, formal vs colloquial hierarchy, slang/character mannerisms.
4. Recommended Target-Language Directives: Formulate precise guidelines for the AI translation engine (pronoun system, idiom adaptation, register and colloquial-expression balance for the target language).

Respond ONLY with a valid JSON object matching this schema:
{
  "style_name": "<Short, evocative title for this style in the target language>",
  "style_description": "<1-2 sentence overview of the novel's core tone and voice, in the target language>",
  "style_prompt": "<Comprehensive directive for the translation AI engine, in the target language, specifying tone, pronoun etiquette, pacing, and vocabulary choices>",
  "analysis_notes": "<2-3 bullet points analyzing the author's unique voice discovered from the sample chapters, in the target language>"
}`, srcName, tgtName)

	userPrompt := fmt.Sprintf("Here are excerpts from the opening narrative chapters of the book:\n\n%s\n\nAnalyze and return the specialized Style Profile in JSON format.", strings.Join(sampleSnippets, "\n\n---\n\n"))

	start := time.Now()
	resp, err := client.Generate(ctx, llm.CompletionRequest{
		Model: modelName,
		Messages: []llm.Message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		Temperature: 0.7,
		MaxTokens:   2000,
	})
	latency := time.Since(start).Milliseconds()

	if err != nil {
		return &dtos.AutoScoutStyleResponse{
			Success:   false,
			LatencyMs: latency,
			ModelUsed: modelName,
			Error:     fmt.Sprintf("LLM error: %v", err),
		}, nil
	}

	type styleJSON struct {
		StyleName        string `json:"style_name"`
		StyleDescription string `json:"style_description"`
		StylePrompt      string `json:"style_prompt"`
		AnalysisNotes    string `json:"analysis_notes"`
	}

	cleanJSON := strings.TrimSpace(resp.Content)
	cleanJSON = strings.TrimPrefix(cleanJSON, "```json")
	cleanJSON = strings.TrimPrefix(cleanJSON, "```")
	cleanJSON = strings.TrimSuffix(cleanJSON, "```")
	cleanJSON = strings.TrimSpace(cleanJSON)

	var parsed styleJSON
	if err := json.Unmarshal([]byte(cleanJSON), &parsed); err != nil {
		parsed.StyleName = "AI-Defined Style"
		parsed.StyleDescription = "Style automatically extracted and refined by AI from the opening chapters."
		parsed.StylePrompt = resp.Content
		parsed.AnalysisNotes = "Automated analysis based on the first 3 chapters of the work."
	}

	// Quick preview translation for the first narrative excerpt
	previewExcerpt := firstNarrativeExcerpt
	if len(previewExcerpt) > 400 {
		previewExcerpt = previewExcerpt[:400]
	}

	previewResp, _ := s.PreviewStyle(ctx, dtos.PreviewStyleRequest{
		ProjectID:   req.ProjectID,
		RawContent:  previewExcerpt,
		StylePrompt: parsed.StylePrompt,
		SourceLang:  sourceLang,
		TargetLang:  targetLang,
	})

	var sampleTrans string
	if previewResp != nil && previewResp.Success {
		sampleTrans = previewResp.TranslatedContent
	}

	// Persist the scouted style to the project in SQLite database so it survives across sessions
	_ = s.store.UpdateProjectStyle(ctx, req.ProjectID, parsed.StyleName, parsed.StylePrompt)

	return &dtos.AutoScoutStyleResponse{
		Success:          true,
		StyleName:        parsed.StyleName,
		StyleDescription: parsed.StyleDescription,
		StylePrompt:      parsed.StylePrompt,
		AnalysisNotes:    parsed.AnalysisNotes,
		SampleExcerpt:    previewExcerpt,
		SampleTranslated: sampleTrans,
		ModelUsed:        modelName,
		LatencyMs:        latency,
		ChaptersSampled:  chaptersSampled,
	}, nil
}
