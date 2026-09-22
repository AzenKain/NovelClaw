package services

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"

	"novelclaw/internal/dtos"
	"novelclaw/pkg/llm"
	"novelclaw/pkg/novelclaw"
	"novelclaw/pkg/paths"
	"novelclaw/pkg/soul"
	"novelclaw/pkg/storage"
)

// wailsNovelClawEmitter dispatches events to Wails v3 event bus.
type wailsNovelClawEmitter struct{}

func (e *wailsNovelClawEmitter) EmitStreamEvent(eventType string, data map[string]any) {
	if app := application.Get(); app != nil && app.Event != nil {
		app.Event.Emit(eventType, data)
		app.Event.Emit("novelclaw:event", map[string]any{
			"type": eventType,
			"data": data,
		})
	}
}

// NovelClawService exposes NovelClaw Agentic chat, threads, and auto-compact to Wails v3 frontend.
type NovelClawService struct {
	store    *storage.Storage
	engine   *novelclaw.Engine
	client   llm.LLMClient
	soulCtrl *soul.Controller
}

// NewNovelClawService initializes a new NovelClawService.
func NewNovelClawService(store *storage.Storage, soulCtrl *soul.Controller, wbSvc *WorldBibleService) *NovelClawService {
	emitter := &wailsNovelClawEmitter{}
	hooks := novelclaw.AppToolsHooks{
		OnPauseTranslation: func(ctx context.Context, projectID string) error {
			if soulCtrl != nil {
				soulCtrl.StopJobsByProject(projectID)
			}
			return nil
		},
		OnSoftStop: func(ctx context.Context, projectID string) error {
			if soulCtrl != nil {
				soulCtrl.StopJobsByProject(projectID)
			}
			return nil
		},
		OnAbort: func(ctx context.Context, projectID string) error {
			if soulCtrl != nil {
				soulCtrl.AbortJobsByProject(projectID)
			}
			return nil
		},
		OnRollback: func(ctx context.Context, projectID string, chapterIndex int64, checkpointType string) error {
			if store != nil {
				_, err := store.RollbackChapterToCheckpoint(ctx, projectID, chapterIndex)
				return err
			}
			return nil
		},
		OnSetSoul: func(ctx context.Context, projectID, soulID string) error {
			if soulCtrl != nil {
				return soulCtrl.SetProjectSoul(projectID, soulID)
			}
			return nil
		},
		OnScanWorld: func(ctx context.Context, projectID string, startChap, endChap int64) error {
			if wbSvc != nil {
				_, err := wbSvc.ScanAndBuildWorld(ctx, projectID, startChap, endChap)
				return err
			}
			return nil
		},
		OnTriggerLearn: func(ctx context.Context, projectID string, notes string) error {
			if wbSvc != nil {
				_, err := wbSvc.LearnFromEdits(ctx, projectID, 0, notes)
				return err
			}
			return nil
		},
	}
	cfg := novelclaw.Config{
		Store:        store,
		SkillsDir:    paths.SkillsDefault(),
		SoulFilePath: filepath.Join(paths.SkillsDefault(), "SOUL.md"),
		Emitter:      emitter,
		ActionHooks:  hooks,
		CompactLimit: novelclaw.DefaultAutoCompactThreshold,
	}

	svc := &NovelClawService{
		store:    store,
		soulCtrl: soulCtrl,
		engine:   novelclaw.NewEngine(cfg),
	}

	return svc
}

// SetNovelClawTestClient allows test suites to inject mock LLM clients without triggering Wails warnings.
func SetNovelClawTestClient(s *NovelClawService, client llm.LLMClient) {
	s.client = client
	if s.engine != nil {
		s.engine.SetClient(client, "test-model")
	}
}

func (s *NovelClawService) resolveClient(ctx context.Context) (llm.LLMClient, string, error) {
	if s.client != nil {
		return s.client, "default-model", nil
	}
	cfg, err := s.store.GetDefaultLLMConfig(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("API key is not configured: %w", err)
	}
	timeout := time.Duration(cfg.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 120 * time.Second
	}
	var client llm.LLMClient
	if strings.Contains(cfg.ApiURL, "generativelanguage.googleapis.com") {
		client = llm.NewGeminiClient(cfg.Token, timeout)
	} else {
		client = llm.NewOpenAIClient(cfg.ApiURL, cfg.Token, timeout)
	}
	return client, cfg.ModelName, nil
}

// GetOrCreateMainThread ensures a main thread exists for the project and returns it.
func (s *NovelClawService) GetOrCreateMainThread(ctx context.Context, projectID string) (*dtos.NovelClawThreadDTO, error) {
	thread, err := s.store.GetOrCreateMainThread(ctx, projectID)
	if err != nil {
		return nil, err
	}
	dto := dtos.ToNovelClawThreadDTO(thread)
	tokens, _ := s.store.SumActiveThreadTokens(ctx, thread.ID)
	dto.TokenCount = tokens
	return &dto, nil
}

// CreateSubThread creates a specialized sub-chat thread for a project, volume, or chapter.
func (s *NovelClawService) CreateSubThread(ctx context.Context, req dtos.CreateSubThreadRequest) (*dtos.NovelClawThreadDTO, error) {
	thread, err := s.store.CreateSubThread(ctx, req.ProjectID, req.Title, req.VolumeIndex, req.ChapterIndex)
	if err != nil {
		return nil, err
	}
	dto := dtos.ToNovelClawThreadDTO(thread)
	return &dto, nil
}

// ListThreads returns all chat threads for a project with token counts.
func (s *NovelClawService) ListThreads(ctx context.Context, projectID string) ([]dtos.NovelClawThreadDTO, error) {
	threads, err := s.store.ListNovelClawThreads(ctx, projectID)
	if err != nil {
		return nil, err
	}
	dtosList := make([]dtos.NovelClawThreadDTO, 0, len(threads))
	for _, t := range threads {
		dto := dtos.ToNovelClawThreadDTO(t)
		tokens, _ := s.store.SumActiveThreadTokens(ctx, t.ID)
		dto.TokenCount = tokens
		dtosList = append(dtosList, dto)
	}
	return dtosList, nil
}

// UpdateThreadTitle changes the title of a sub-chat thread.
func (s *NovelClawService) UpdateThreadTitle(ctx context.Context, threadID, title string) error {
	return s.store.UpdateNovelClawThreadTitle(ctx, threadID, title)
}

// DeleteThread deletes a sub-chat thread and its messages.
func (s *NovelClawService) DeleteThread(ctx context.Context, threadID string) error {
	return s.store.DeleteNovelClawThread(ctx, threadID)
}

// ListMessages returns non-archived messages in chronological order for a thread.
func (s *NovelClawService) ListMessages(ctx context.Context, threadID string) ([]dtos.NovelClawMessageDTO, error) {
	msgs, err := s.store.ListActiveNovelClawMessages(ctx, threadID)
	if err != nil {
		return nil, err
	}
	dtosList := make([]dtos.NovelClawMessageDTO, 0, len(msgs))
	for _, m := range msgs {
		dtosList = append(dtosList, dtos.ToNovelClawMessageDTO(m))
	}
	return dtosList, nil
}

// SendMessage sends a user message to NovelClaw and executes tools / returns reply.
func (s *NovelClawService) SendMessage(ctx context.Context, req dtos.SendNovelClawMessageRequest) (*dtos.NovelClawMessageDTO, error) {
	// Dynamically resolve LLM client
	client, modelName, err := s.resolveClient(ctx)
	if err == nil && client != nil {
		s.engine.SetClient(client, modelName)
	}

	// Dynamically sync active Soul Persona into NovelClaw engine
	if s.soulCtrl != nil {
		activeSoul := s.soulCtrl.GetProjectSoul(req.ProjectID)
		if activeSoul.SystemPromptAddon != "" {
			personaPrompt := fmt.Sprintf("# ACTIVE ASSISTANT PERSONA: %s (%s)\n%s\n\n[COMMUNICATION GUIDELINES]:\n- Greeting: %s\n- When confused: %s\n- When success: %s",
				activeSoul.Name, activeSoul.Archetype, activeSoul.SystemPromptAddon, activeSoul.Greeting, activeSoul.OnConfused, activeSoul.OnSuccess)
			s.engine.SetSoulPrompt(personaPrompt)
		}
	}

	return s.engine.SendMessage(ctx, req)
}

// ClearThread deletes all messages in a thread.
func (s *NovelClawService) ClearThread(ctx context.Context, threadID string) error {
	return s.store.ClearNovelClawThreadMessages(ctx, threadID)
}

// CompactThread forces an immediate compression of older messages in the thread.
func (s *NovelClawService) CompactThread(ctx context.Context, threadID string) (*dtos.CompactThreadResponse, error) {
	client, modelName, _ := s.resolveClient(ctx)
	compactor := novelclaw.NewCompactor(s.store, client, modelName, s.engine.GetAutoCompactThreshold())
	return compactor.CompactThread(ctx, threadID, true)
}

// SetAutoCompactThreshold updates the threshold for auto-compacting context.
func (s *NovelClawService) SetAutoCompactThreshold(tokens int64) error {
	if tokens < 10000 {
		return fmt.Errorf("minimum compaction threshold is 10,000 tokens")
	}
	s.engine.SetAutoCompactThreshold(tokens)
	return nil
}

// GetAutoCompactThreshold returns current threshold.
func (s *NovelClawService) GetAutoCompactThreshold() int64 {
	return s.engine.GetAutoCompactThreshold()
}
