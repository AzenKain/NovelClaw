package novelclaw

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"novelclaw/internal/dtos"
	"novelclaw/internal/gen/sqlc"
	"novelclaw/pkg/llm"
	"novelclaw/pkg/skills"
	"novelclaw/pkg/storage"
)

// NovelClawEventEmitter sends live events to the frontend via Wails.
type NovelClawEventEmitter interface {
	EmitStreamEvent(eventType string, data map[string]any)
}

// Engine coordinates NovelClaw agentic chat, context compilation, tool execution, and activity streams.
type Engine struct {
	store        *storage.Storage
	client       llm.LLMClient
	modelName    string
	skillsLoader *skills.SkillLoader
	toolsExec    *AppToolsExecutor
	compactor    *Compactor
	emitter      NovelClawEventEmitter
	soulPrompt   string
}

// Config provides configuration options for NovelClaw Engine.
type Config struct {
	Store        *storage.Storage
	Client       llm.LLMClient
	ModelName    string
	SkillsDir    string
	SoulFilePath string
	Emitter      NovelClawEventEmitter
	ActionHooks  AppToolsHooks
	CompactLimit int64
}

// NewEngine initializes a NovelClaw Engine instance.
func NewEngine(cfg Config) *Engine {
	skillsLoader := skills.NewSkillLoader(cfg.SkillsDir)
	_ , _ = skillsLoader.LoadAll()

	var actionEmitter AppActionEmitter
	if cfg.Emitter != nil {
		actionEmitter = &wailsActionAdapter{emitter: cfg.Emitter}
	} else {
		actionEmitter = DefaultNoopEmitter{}
	}

	toolsExec := NewAppToolsExecutor(cfg.Store, actionEmitter, cfg.ActionHooks)
	compactor := NewCompactor(cfg.Store, cfg.Client, cfg.ModelName, cfg.CompactLimit)

	soulContent := ""
	if cfg.SoulFilePath != "" {
		if data, err := os.ReadFile(cfg.SoulFilePath); err == nil {
			soulContent = string(data)
		}
	}
	if soulContent == "" {
		soulContent = `You are NovelClaw, the autonomous Lead AI Editor and Studio Companion for NovelClaw.
You assist authors and translators with literary translation, character relationship tracking, glossary management, and studio automation.
Always execute appropriate function tools when asked to manipulate the workspace or perform actions.`
	}

	return &Engine{
		store:        cfg.Store,
		client:       cfg.Client,
		modelName:    cfg.ModelName,
		skillsLoader: skillsLoader,
		toolsExec:    toolsExec,
		compactor:    compactor,
		emitter:      cfg.Emitter,
		soulPrompt:   soulContent,
	}
}

// SetClient updates the active LLM client and model name (for dynamic model switching).
func (e *Engine) SetClient(client llm.LLMClient, modelName string) {
	e.client = client
	e.modelName = modelName
	if e.compactor != nil {
		e.compactor.client = client
		e.compactor.modelName = modelName
	}
}

// SetSoulPrompt dynamically updates the persona system prompt for NovelClaw.
func (e *Engine) SetSoulPrompt(prompt string) {
	e.soulPrompt = prompt
}

// SetAutoCompactThreshold updates the token threshold for Auto-Compact.
func (e *Engine) SetAutoCompactThreshold(tokens int64) {
	if e.compactor != nil {
		e.compactor.SetThreshold(tokens)
	}
}

// GetAutoCompactThreshold returns the current threshold.
func (e *Engine) GetAutoCompactThreshold() int64 {
	if e.compactor != nil {
		return e.compactor.GetThreshold()
	}
	return DefaultAutoCompactThreshold
}

// SendMessage handles an incoming user chat message in a thread and returns the assistant's response.
func (e *Engine) SendMessage(ctx context.Context, req dtos.SendNovelClawMessageRequest) (*dtos.NovelClawMessageDTO, error) {
	projectID := req.ProjectID
	threadID := req.ThreadID
	userText := strings.TrimSpace(req.Content)

	if threadID == "" {
		return nil, fmt.Errorf("missing thread_id")
	}
	if userText == "" {
		return nil, fmt.Errorf("message content is empty")
	}

	// 1. Check auto-compact threshold before processing
	if e.compactor != nil {
		if needed, _, _ := e.compactor.NeedsCompact(ctx, threadID); needed {
			_, _ = e.compactor.CompactThread(ctx, threadID, false)
		}
	}

	// 2. Check for slash command shortcut execution
	slashRes, err := e.ParseAndExecuteSlashCommand(ctx, projectID, threadID, userText)
	if err != nil {
		return nil, err
	}
	if slashRes != nil && slashRes.Handled {
		if strings.HasPrefix(strings.ToLower(userText), "/clear") {
			dto := dtos.NovelClawMessageDTO{
				ThreadID:  threadID,
				ProjectID: projectID,
				Sender:    "novelclaw",
				Role:      "assistant",
				Content:   slashRes.Content,
			}
			e.emitEvent("thread_cleared", map[string]any{
				"thread_id": threadID,
			})
			return &dto, nil
		}

		// Save user message
		userMsgID := "nc_msg_" + strings.ReplaceAll(uuid.New().String(), "-", "")[:16]
		userTokens := int64(len(userText) / 4)
		_, _ = e.store.SaveNovelClawMessage(ctx, sqlc.CreateNovelClawMessageParams{
			ID:                userMsgID,
			ThreadID:          threadID,
			ProjectID:         projectID,
			Sender:            "user",
			Role:              "user",
			Content:           userText,
			ThinkingContent:   sql.NullString{Valid: false},
			StepType:          sql.NullString{Valid: false},
			StepStatus:        sql.NullString{String: "completed", Valid: true},
			ActionCallJson:    sql.NullString{Valid: false},
			IsCollapsed:       0,
			IsArchivedCompact: 0,
			TokenCount:        userTokens,
		})

		// Save response message
		respMsgID := "nc_msg_" + strings.ReplaceAll(uuid.New().String(), "-", "")[:16]
		respTokens := int64(len(slashRes.Content) / 4)
		var actionJSON sql.NullString
		if slashRes.ActionPayload != nil {
			if b, err := json.Marshal(slashRes.ActionPayload); err == nil {
				actionJSON = sql.NullString{String: string(b), Valid: true}
			}
		}

		respMsg, err := e.store.SaveNovelClawMessage(ctx, sqlc.CreateNovelClawMessageParams{
			ID:                respMsgID,
			ThreadID:          threadID,
			ProjectID:         projectID,
			Sender:            "novelclaw",
			Role:              "assistant",
			Content:           slashRes.Content,
			ThinkingContent:   sql.NullString{Valid: false},
			StepType:          sql.NullString{String: "slash_command", Valid: true},
			StepStatus:        sql.NullString{String: "completed", Valid: true},
			ActionCallJson:    actionJSON,
			IsCollapsed:       0,
			IsArchivedCompact: 0,
			TokenCount:        respTokens,
		})
		if err != nil {
			return nil, fmt.Errorf("save slash command message: %w", err)
		}

		dto := dtos.ToNovelClawMessageDTO(respMsg)
		e.emitEvent("message_created", map[string]any{
			"thread_id": threadID,
			"message":   dto,
		})
		return &dto, nil
	}

	// 3. Save User Message into thread
	userMsgID := "nc_msg_" + strings.ReplaceAll(uuid.New().String(), "-", "")[:16]
	userTokens := int64(len(userText) / 4)
	_, err = e.store.SaveNovelClawMessage(ctx, sqlc.CreateNovelClawMessageParams{
		ID:                userMsgID,
		ThreadID:          threadID,
		ProjectID:         projectID,
		Sender:            "user",
		Role:              "user",
		Content:           userText,
		ThinkingContent:   sql.NullString{Valid: false},
		StepType:          sql.NullString{Valid: false},
		StepStatus:        sql.NullString{String: "completed", Valid: true},
		ActionCallJson:    sql.NullString{Valid: false},
		IsCollapsed:       0,
		IsArchivedCompact: 0,
		TokenCount:        userTokens,
	})
	if err != nil {
		return nil, fmt.Errorf("save user message: %w", err)
	}

	if e.client == nil {
		// Fallback if no LLM client configured
		fallbackText := "Hello! I am NovelClaw. Please configure your API key in Settings so I can reason, execute tools, and assist your translation workflow."
		fallbackID := "nc_msg_" + strings.ReplaceAll(uuid.New().String(), "-", "")[:16]
		msg, _ := e.store.SaveNovelClawMessage(ctx, sqlc.CreateNovelClawMessageParams{
			ID:                fallbackID,
			ThreadID:          threadID,
			ProjectID:         projectID,
			Sender:            "novelclaw",
			Role:              "assistant",
			Content:           fallbackText,
			ThinkingContent:   sql.NullString{Valid: false},
			StepType:          sql.NullString{Valid: false},
			StepStatus:        sql.NullString{String: "completed", Valid: true},
			ActionCallJson:    sql.NullString{Valid: false},
			IsCollapsed:       0,
			IsArchivedCompact: 0,
			TokenCount:        int64(len(fallbackText) / 4),
		})
		dto := dtos.ToNovelClawMessageDTO(msg)
		return &dto, nil
	}

	// 4. Compile Multilayer Context
	systemPrompt := e.buildSystemPrompt(ctx, projectID, userText)

	// Match skills and filter tools
	matchedSkills := e.skillsLoader.MatchSkillsByIntent(userText)
	tools := skills.FilterToolsBySkills(matchedSkills, GetOmniAppToolDefinitions())

	// Load active conversation history
	activeMsgs, err := e.store.ListActiveNovelClawMessages(ctx, threadID)
	if err != nil {
		return nil, fmt.Errorf("load chat history: %w", err)
	}

	var messages []llm.Message
	messages = append(messages, llm.Message{
		Role:    llm.RoleSystem,
		Content: systemPrompt,
	})

	for _, m := range activeMsgs {
		role := llm.Role(m.Role)
		if role == "" {
			role = llm.RoleUser
		}
		messages = append(messages, llm.Message{
			Role:    role,
			Content: m.Content,
		})
	}

	// 5. ReAct Multi-turn Tool Calling Execution Loop
	maxIterations := 5
	var finalContent strings.Builder
	var allThinking strings.Builder
	var dispatchedActions []dtos.AppActionPayload

	e.emitEvent("thinking_start", map[string]any{
		"thread_id": threadID,
		"status":    "Analyzing request and planning actions...",
	})

	defaultEffort := "off"
	if defCfg, err := e.store.GetDefaultLLMConfig(ctx); err == nil && defCfg != nil && defCfg.ReasoningEffort != "" {
		defaultEffort = defCfg.ReasoningEffort
	}

	for iter := 0; iter < maxIterations; iter++ {
		req := llm.CompletionRequest{
			Model:           e.modelName,
			Messages:        messages,
			Tools:           tools,
			Temperature:     0.3,
			ReasoningEffort: defaultEffort,
		}

		resp, err := e.client.Generate(ctx, req)
		if err != nil {
			log.Error().Err(err).Msg("NovelClaw generation failed")
			e.emitEvent("error", map[string]any{"thread_id": threadID, "error": err.Error()})
			return nil, fmt.Errorf("generate LLM response: %w", err)
		}

		// Check if thought/reasoning was emitted
		if resp.Thought != "" {
			allThinking.WriteString(resp.Thought)
			e.emitEvent("thinking", map[string]any{
				"thread_id": threadID,
				"content":   resp.Thought,
			})
		}

		// Tool calling requested
		if len(resp.ToolCalls) > 0 {
			for _, tc := range resp.ToolCalls {
				e.emitEvent("step_badge", map[string]any{
					"thread_id":  threadID,
					"tool_name":  tc.Function.Name,
					"step_state": "running",
				})

				resText, payload, execErr := e.toolsExec.Execute(ctx, projectID, tc)
				if execErr != nil {
					resText = fmt.Sprintf("Tool error: %v", execErr)
				}

				if payload != nil {
					dispatchedActions = append(dispatchedActions, *payload)
				}

				// Append assistant tool call and tool result to messages
				messages = append(messages, llm.Message{
					Role:      llm.RoleAssistant,
					ToolCalls: []llm.ToolCall{tc},
				})
				messages = append(messages, llm.Message{
					Role:       llm.RoleTool,
					ToolCallID: tc.ID,
					Content:    resText,
				})
			}
			// Continue loop to let LLM respond based on tool results
			continue
		}

		// No more tool calls: final assistant reply received
		finalContent.WriteString(resp.Content)
		break
	}

	replyText := finalContent.String()
	if replyText == "" {
		replyText = "Completed the requested actions successfully!"
	}

	// 6. Save Assistant Response Message
	lastThinking := allThinking.String()
	respMsgID := "nc_msg_" + strings.ReplaceAll(uuid.New().String(), "-", "")[:16]
	respTokens := int64(len(replyText)/4) + int64(len(lastThinking)/4)

	var actionCallJSON sql.NullString
	if len(dispatchedActions) > 0 {
		if b, err := json.Marshal(dispatchedActions); err == nil {
			actionCallJSON = sql.NullString{String: string(b), Valid: true}
		}
	}

	var thinkingNull sql.NullString
	if lastThinking != "" {
		thinkingNull = sql.NullString{String: lastThinking, Valid: true}
	}

	stepType := "chat"
	if len(dispatchedActions) > 0 {
		stepType = "action_dispatched"
	}

	savedMsg, err := e.store.SaveNovelClawMessage(ctx, sqlc.CreateNovelClawMessageParams{
		ID:                respMsgID,
		ThreadID:          threadID,
		ProjectID:         projectID,
		Sender:            "novelclaw",
		Role:              "assistant",
		Content:           replyText,
		ThinkingContent:   thinkingNull,
		StepType:          sql.NullString{String: stepType, Valid: true},
		StepStatus:        sql.NullString{String: "completed", Valid: true},
		ActionCallJson:    actionCallJSON,
		IsCollapsed:       0,
		IsArchivedCompact: 0,
		TokenCount:        respTokens,
	})
	if err != nil {
		return nil, fmt.Errorf("save reply message: %w", err)
	}

	dto := dtos.ToNovelClawMessageDTO(savedMsg)
	e.emitEvent("message_created", map[string]any{
		"thread_id": threadID,
		"message":   dto,
	})

	return &dto, nil
}

// buildSystemPrompt aggregates Soul Persona, Project Context, and Skill Playbooks.
func (e *Engine) buildSystemPrompt(ctx context.Context, projectID, userQuery string) string {
	var sb strings.Builder

	// 1. Persona & Identity
	sb.WriteString(e.soulPrompt)
	sb.WriteString("\n\n")

	// 2. Project Context
	if projectID != "" && e.store != nil {
		proj, err := e.store.GetProject(ctx, projectID)
		if err == nil {
			sb.WriteString(fmt.Sprintf("[PROJECT CONTEXT]:\n- Title: %s\n- Source: %s -> Target: %s\n",
				proj.Title, proj.SourceLang, proj.TargetLang))
		}

		// Summary of character relations count
		rels, _ := e.store.ListRelationsByProject(ctx, projectID)
		if len(rels) > 0 {
			sb.WriteString(fmt.Sprintf("- L2 Character Graph: %d established relations across timeline.\n", len(rels)))
		}

		// Summary of glossary count
		terms, _ := e.store.ListGlossaryByProject(ctx, projectID)
		if len(terms) > 0 {
			sb.WriteString(fmt.Sprintf("- L3 Domain Glossary: %d specialized terms defined.\n", len(terms)))
		}
		sb.WriteString("\n")
	}

	// 3. Matched Skills Playbooks
	matchedSkills := e.skillsLoader.MatchSkillsByIntent(userQuery)
	skillsPrompt := skills.BuildSkillsPrompt(matchedSkills)
	if skillsPrompt != "" {
		sb.WriteString(skillsPrompt)
	}

	// 4. Strict Autonomous Tool Calling Directive
	sb.WriteString(`[DIRECTIVE - ACTION EXECUTION]:
When the user requests an action on the studio app (e.g. switching tabs, selecting volumes/chapters, creating or updating character address relations, updating World Bible entries, managing glossary terms, modifying translation settings, starting/pausing/stopping/aborting translation, exporting books...), YOU MUST INVOKE THE CORRESPONDING FUNCTION TOOL IMMEDIATELY.
NEVER merely claim in text that you have performed an action without executing the function call. Invoke the tool first, then report the concise result to the user in their preferred language.`)

	return sb.String()
}

func (e *Engine) emitEvent(eventType string, data map[string]any) {
	if e.emitter != nil {
		e.emitter.EmitStreamEvent(eventType, data)
	}
}

// wailsActionAdapter adapts NovelClawEventEmitter to AppActionEmitter.
type wailsActionAdapter struct {
	emitter NovelClawEventEmitter
}

func (a *wailsActionAdapter) EmitAction(payload dtos.AppActionPayload) {
	if a.emitter != nil {
		data := map[string]any{
			"action":      payload.Action,
			"category":    payload.Category,
			"description": payload.Description,
			"data":        payload.Data,
			"timestamp":   time.Now(),
		}
		a.emitter.EmitStreamEvent("app:action_dispatched", data)
	}
}
