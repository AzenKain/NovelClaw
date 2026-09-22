package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"

	"novelclaw/internal/dtos"
	"novelclaw/pkg/auditor"
	"novelclaw/pkg/llm"
	"novelclaw/pkg/skills"
	"novelclaw/pkg/storage"
	"novelclaw/pkg/voice"
	"novelclaw/pkg/worldbible"
)

// WorldBibleService coordinates the Universal World Bible, Autonomous World Builder,
// Reflexion Diff-Learning Distiller, and Character Voice/Plot Auditor for Wails v3.
type WorldBibleService struct {
	store         *storage.Storage
	skillRegistry *skills.Registry
	wbService     *worldbible.Service
	matcher       *worldbible.Matcher
	builder       *worldbible.BuilderAgent
	skillStorage  *skills.StorageManager
	distiller     *skills.Distiller
	plotAuditor   *auditor.PlotAuditor
	voiceRegistry *voice.Registry
	client        llm.LLMClient
}

// NewWorldBibleService creates a new WorldBibleService instance.
func NewWorldBibleService(
	store *storage.Storage,
	skillRegistry *skills.Registry,
	matcher *worldbible.Matcher,
	voiceRegistry *voice.Registry,
	plotAuditor *auditor.PlotAuditor,
) *WorldBibleService {
	if matcher == nil {
		matcher = worldbible.NewMatcher(store)
	}
	if voiceRegistry == nil {
		voiceRegistry = voice.NewRegistry()
	}
	if plotAuditor == nil {
		plotAuditor = auditor.NewPlotAuditor(store)
	}
	wbService := worldbible.NewService(store)
	builder := worldbible.NewBuilderAgent(store)
	skillStorage := skills.NewStorageManager("")
	distiller := skills.NewDistiller()

	return &WorldBibleService{
		store:         store,
		skillRegistry: skillRegistry,
		wbService:     wbService,
		matcher:       matcher,
		builder:       builder,
		skillStorage:  skillStorage,
		distiller:     distiller,
		plotAuditor:   plotAuditor,
		voiceRegistry: voiceRegistry,
	}
}

// SetTestClientWb overrides LLMClient (primarily for testing).
func SetTestClientWb(s *WorldBibleService, c llm.LLMClient) {
	s.client = c
}

// SetSkillStorage overrides skillStorage (primarily for testing with isolated temporary dirs).
func (s *WorldBibleService) SetSkillStorage(sm *skills.StorageManager) {
	if sm != nil {
		s.skillStorage = sm
	}
}

// getLLMClient initializes or returns the configured LLM client.
func (s *WorldBibleService) getLLMClient(ctx context.Context) (llm.LLMClient, string, error) {
	if s.client != nil {
		return s.client, "default-model", nil
	}
	cfg, err := s.store.GetDefaultLLMConfig(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("LLM API key is not configured: %w", err)
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

// ListCategories returns all categories defined for a project.
func (s *WorldBibleService) ListCategories(ctx context.Context, projectID string) ([]dtos.WorldCategoryDTO, error) {
	if s.store == nil || projectID == "" {
		return nil, fmt.Errorf("invalid storage or project id")
	}
	return s.wbService.ListCategories(ctx, projectID)
}

// ListEntries returns world entries for a project, optionally filtered by category.
func (s *WorldBibleService) ListEntries(ctx context.Context, projectID, categoryID string) ([]dtos.WorldEntryDTO, error) {
	if s.store == nil || projectID == "" {
		return nil, fmt.Errorf("invalid storage or project id")
	}
	return s.wbService.ListEntries(ctx, projectID, categoryID)
}

// UpsertCategory adds or modifies a dynamic world taxonomy category.
func (s *WorldBibleService) UpsertCategory(ctx context.Context, req dtos.UpsertWorldCategoryRequest) (*dtos.WorldCategoryDTO, error) {
	cat, err := s.wbService.UpsertCategory(ctx, req)
	if err != nil {
		return nil, err
	}

	s.matcher.InvalidateCache(req.ProjectID)

	if app := application.Get(); app != nil && app.Event != nil {
		app.Event.Emit("worldbible:category_updated", map[string]any{
			"project_id": req.ProjectID,
			"category":   cat,
		})
	}

	return cat, nil
}

// DeleteCategory deletes a category and cascades to its entries.
func (s *WorldBibleService) DeleteCategory(ctx context.Context, id string) error {
	cat, err := s.store.GetWorldCategoryByID(ctx, id)
	if err != nil {
		return fmt.Errorf("category not found: %w", err)
	}

	if err := s.wbService.DeleteCategory(ctx, id); err != nil {
		return err
	}

	s.matcher.InvalidateCache(cat.ProjectID)

	if app := application.Get(); app != nil && app.Event != nil {
		app.Event.Emit("worldbible:category_deleted", map[string]any{
			"project_id":  cat.ProjectID,
			"category_id": id,
		})
	}

	return nil
}

// UpsertEntry adds or modifies a Lorebook entry.
func (s *WorldBibleService) UpsertEntry(ctx context.Context, req dtos.UpsertWorldEntryRequest) (*dtos.WorldEntryDTO, error) {
	entry, err := s.wbService.UpsertEntry(ctx, req)
	if err != nil {
		return nil, err
	}

	s.matcher.InvalidateCache(req.ProjectID)

	if app := application.Get(); app != nil && app.Event != nil {
		app.Event.Emit("worldbible:entry_updated", map[string]any{
			"project_id": req.ProjectID,
			"entry":      entry,
		})
	}

	return entry, nil
}

// VerifyEntry sets the verified status (1 = approved by human, 0 = AI suggested).
func (s *WorldBibleService) VerifyEntry(ctx context.Context, id string, verified bool) error {
	entry, err := s.store.GetWorldEntryByID(ctx, id)
	if err != nil {
		return fmt.Errorf("entry not found: %w", err)
	}

	if err := s.wbService.VerifyEntry(ctx, id, verified); err != nil {
		return err
	}

	s.matcher.InvalidateCache(entry.ProjectID)

	if app := application.Get(); app != nil && app.Event != nil {
		app.Event.Emit("worldbible:entry_verified", map[string]any{
			"project_id":  entry.ProjectID,
			"entry_id":    id,
			"is_verified": verified,
		})
	}

	return nil
}

// DeleteEntry removes an entry from the World Bible.
func (s *WorldBibleService) DeleteEntry(ctx context.Context, id string) error {
	entry, err := s.store.GetWorldEntryByID(ctx, id)
	if err != nil {
		return fmt.Errorf("entry not found: %w", err)
	}

	if err := s.wbService.DeleteEntry(ctx, id); err != nil {
		return err
	}

	s.matcher.InvalidateCache(entry.ProjectID)

	if app := application.Get(); app != nil && app.Event != nil {
		app.Event.Emit("worldbible:entry_deleted", map[string]any{
			"project_id": entry.ProjectID,
			"entry_id":   id,
		})
	}

	return nil
}

// ScanAndBuildWorld triggers the autonomous builder agent to scan novel chapters and synthesize the World Bible.
func (s *WorldBibleService) ScanAndBuildWorld(ctx context.Context, projectID string, startChap, endChap int64) (*dtos.ScanWorldResponse, error) {
	client, modelName, err := s.getLLMClient(ctx)
	if err != nil {
		return nil, err
	}

	resp, err := s.builder.ScanAndBuildWorld(ctx, client, modelName, projectID, startChap, endChap)
	if err != nil {
		return nil, err
	}

	s.matcher.InvalidateCache(projectID)

	if app := application.Get(); app != nil && app.Event != nil {
		app.Event.Emit("worldbible:scanned", map[string]any{
			"project_id":       projectID,
			"start_chapter":    startChap,
			"end_chapter":      endChap,
			"inferred_genre":   resp.InferredGenre,
			"categories_added": resp.CategoriesAdded,
			"entries_added":    resp.EntriesAdded,
		})
	}

	return resp, nil
}

// ExportWorldBibleMarkdown generates a complete WORLD_BIBLE.md export.
func (s *WorldBibleService) ExportWorldBibleMarkdown(ctx context.Context, projectID string) (string, error) {
	proj, err := s.store.GetProject(ctx, projectID)
	if err != nil {
		return "", fmt.Errorf("get project: %w", err)
	}

	cats, err := s.wbService.ListCategories(ctx, projectID)
	if err != nil {
		return "", err
	}

	entries, err := s.wbService.ListEntries(ctx, projectID, "")
	if err != nil {
		return "", err
	}

	return worldbible.ExportMarkdown(proj.Title, cats, entries), nil
}

// ExportWorldBibleJSON generates a structured JSON export.
func (s *WorldBibleService) ExportWorldBibleJSON(ctx context.Context, projectID string) (string, error) {
	cats, err := s.wbService.ListCategories(ctx, projectID)
	if err != nil {
		return "", err
	}

	entries, err := s.wbService.ListEntries(ctx, projectID, "")
	if err != nil {
		return "", err
	}

	return worldbible.ExportJSON(projectID, cats, entries)
}

// ImportWorldBibleJSON imports categories and entries from JSON into a project.
func (s *WorldBibleService) ImportWorldBibleJSON(ctx context.Context, projectID, jsonStr string) (int, int, error) {
	catsImported, entriesImported, err := worldbible.ImportJSON(ctx, s.store, projectID, jsonStr)
	if err != nil {
		return 0, 0, err
	}

	s.matcher.InvalidateCache(projectID)

	if app := application.Get(); app != nil && app.Event != nil {
		app.Event.Emit("worldbible:imported", map[string]any{
			"project_id":       projectID,
			"categories_count": catsImported,
			"entries_count":    entriesImported,
		})
	}

	return catsImported, entriesImported, nil
}

// ResetSkillsToDefault provides a 1-click factory reset wiping evolved skill mutations for a project.
func (s *WorldBibleService) ResetSkillsToDefault(ctx context.Context, projectID string) error {
	if s.skillStorage == nil || projectID == "" {
		return fmt.Errorf("invalid storage or project id")
	}

	if err := s.skillStorage.ResetToFactoryDefaults(ctx, projectID); err != nil {
		return err
	}

	if s.skillRegistry != nil {
		for _, skill := range skills.DefaultBuiltinSkills() {
			_ = s.skillRegistry.ResetSkillToDefault(ctx, projectID, skill.ID)
		}
	}

	if app := application.Get(); app != nil && app.Event != nil {
		app.Event.Emit("skill:reset_defaults", map[string]any{
			"project_id": projectID,
		})
	}

	return nil
}

// LearnFromEdits analyzes edits made to a chapter and persists evolved rules via the Distiller.
func (s *WorldBibleService) LearnFromEdits(ctx context.Context, projectID string, chapterIndex int64, notes string) (*dtos.LearnResultDTO, error) {
	ch, err := s.store.GetChapterByIndex(ctx, projectID, chapterIndex)
	if err != nil {
		return nil, fmt.Errorf("get chapter %d: %w", chapterIndex, err)
	}

	if strings.TrimSpace(ch.TranslatedContent) == "" {
		return nil, fmt.Errorf("chapter %d has no translated content to learn from", chapterIndex)
	}

	client, modelName, err := s.getLLMClient(ctx)
	if err != nil {
		return nil, err
	}

	// Look for prior draft in chapter checkpoints
	originalDraft := ""
	if cp, cpErr := s.store.GetLatestChapterCheckpoint(ctx, projectID, chapterIndex); cpErr == nil {
		var cpState map[string]string
		if err := storage.UnpackCheckpointState(cp, &cpState); err == nil && cpState["translated_content"] != "" {
			originalDraft = cpState["translated_content"]
		}
	}
	if originalDraft == "" {
		// If no checkpoint found, use translated content as baseline for note-based distillation
		originalDraft = ch.TranslatedContent
	}

	distillRes, err := s.distiller.DistillDiff(
		ctx,
		client,
		modelName,
		ch.RawContent,
		originalDraft,
		ch.TranslatedContent,
		notes,
	)
	if err != nil {
		return nil, err
	}

	// Persist to evolved skills
	targetSkillID := distillRes.TargetSkillID
	if targetSkillID == "" {
		targetSkillID = "skill_literary_translator"
	}

	if err := s.skillStorage.SaveEvolvedRules(projectID, targetSkillID, distillRes.Rules, distillRes.FewShots); err != nil {
		return nil, fmt.Errorf("save evolved rules: %w", err)
	}

	if app := application.Get(); app != nil && app.Event != nil {
		app.Event.Emit("skill:learned", map[string]any{
			"project_id":        projectID,
			"target_skill_id":   targetSkillID,
			"rules_learned":     distillRes.Rules,
			"few_shots_learned": len(distillRes.FewShots),
			"summary":           distillRes.Summary,
		})
	}

	return &dtos.LearnResultDTO{
		Success:         true,
		RulesLearned:    distillRes.Rules,
		FewShotsLearned: len(distillRes.FewShots),
		Summary:         distillRes.Summary,
		TargetSkillID:   targetSkillID,
	}, nil
}

// ListCharacterVoices returns registered voice profiles for a project.
func (s *WorldBibleService) ListCharacterVoices(projectID string) []dtos.CharacterVoiceDTO {
	if s.voiceRegistry == nil || projectID == "" {
		return nil
	}
	voices := s.voiceRegistry.ListVoices(projectID)
	res := make([]dtos.CharacterVoiceDTO, len(voices))
	for i, v := range voices {
		res[i] = dtos.CharacterVoiceDTO{
			CharacterID:   v.CharacterID,
			Name:          v.Name,
			Aliases:       v.Aliases,
			VoiceTone:     v.VoiceTone,
			DialogueRules: v.DialogueRules,
			Catchphrases:  v.Catchphrases,
		}
	}
	return res
}

// UpsertCharacterVoice registers or updates a character voice profile.
func (s *WorldBibleService) UpsertCharacterVoice(req dtos.UpsertVoiceRequest) error {
	if s.voiceRegistry == nil || req.ProjectID == "" {
		return fmt.Errorf("invalid voice registry or project id")
	}
	s.voiceRegistry.UpsertVoice(req.ProjectID, voice.CharacterVoice{
		CharacterID:   req.CharacterID,
		Name:          req.Name,
		Aliases:       req.Aliases,
		VoiceTone:     req.VoiceTone,
		DialogueRules: req.DialogueRules,
		Catchphrases:  req.Catchphrases,
	})
	return nil
}

// SyncVoicesFromGraph populates voice profiles from knowledge graph character entities.
func (s *WorldBibleService) SyncVoicesFromGraph(ctx context.Context, projectID string) (int, error) {
	if s.store == nil || s.voiceRegistry == nil || projectID == "" {
		return 0, nil
	}
	entities, err := s.store.ListEntitiesByProject(ctx, projectID)
	if err != nil {
		return 0, fmt.Errorf("list entities for voice sync: %w", err)
	}
	count := s.voiceRegistry.SyncFromEntities(projectID, entities)
	return count, nil
}

