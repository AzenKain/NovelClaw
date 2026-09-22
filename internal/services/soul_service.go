package services

import (
	"fmt"

	"github.com/wailsapp/wails/v3/pkg/application"

	"novelclaw/internal/dtos"
	"novelclaw/pkg/soul"
)

// SoulService provides coworker personality interactions, soul profile CRUD, and hot-steering controls to Wails v3.
type SoulService struct {
	controller *soul.Controller
}

// NewSoulService creates a new SoulService instance.
func NewSoulService(controller *soul.Controller) *SoulService {
	if controller == nil {
		controller = soul.NewController()
	}
	return &SoulService{
		controller: controller,
	}
}

// Controller exposes the underlying soul controller for engine-level coordination.
func (s *SoulService) Controller() *soul.Controller {
	return s.controller
}

func soulToDTO(sData soul.Soul) dtos.SoulDTO {
	return dtos.SoulDTO{
		ID:                sData.ID,
		Name:              sData.Name,
		Avatar:            sData.Avatar,
		Title:             sData.Title,
		Archetype:         sData.Archetype,
		Description:       sData.Description,
		Greeting:          sData.Greeting,
		OnConfused:        sData.OnConfused,
		OnSuccess:         sData.OnSuccess,
		SystemTone:        sData.SystemTone,
		CurrentMood:       string(sData.CurrentMood),
		SystemPromptAddon: sData.SystemPromptAddon,
		IsDefault:         sData.IsDefault,
	}
}

func dtoToSoul(dto dtos.SoulDTO) soul.Soul {
	return soul.Soul{
		ID:                dto.ID,
		Name:              dto.Name,
		Avatar:            dto.Avatar,
		Title:             dto.Title,
		Archetype:         dto.Archetype,
		Description:       dto.Description,
		Greeting:          dto.Greeting,
		OnConfused:        dto.OnConfused,
		OnSuccess:         dto.OnSuccess,
		SystemTone:        dto.SystemTone,
		CurrentMood:       soul.EmotionState(dto.CurrentMood),
		SystemPromptAddon: dto.SystemPromptAddon,
		IsDefault:         dto.IsDefault,
	}
}

// GetSoul returns the active Soul companion metadata and current mood.
func (s *SoulService) GetSoul() dtos.SoulDTO {
	return soulToDTO(s.controller.GetSoul())
}

// ListAvailableSouls returns all registered soul profiles.
func (s *SoulService) ListAvailableSouls() []dtos.SoulDTO {
	souls := s.controller.ListSouls()
	dtosList := make([]dtos.SoulDTO, len(souls))
	for i, sl := range souls {
		dtosList[i] = soulToDTO(sl)
	}
	return dtosList
}

// GetActiveSoul returns the active soul for a project or the global active soul.
func (s *SoulService) GetActiveSoul(projectID string) dtos.SoulDTO {
	return soulToDTO(s.controller.GetProjectSoul(projectID))
}

// SetActiveSoul changes the active soul for a project and emits a soul:changed event to frontend.
func (s *SoulService) SetActiveSoul(projectID string, soulID string) error {
	if err := s.controller.SetProjectSoul(projectID, soulID); err != nil {
		return err
	}
	active := s.controller.GetSoul()
	dto := soulToDTO(active)

	if app := application.Get(); app != nil && app.Event != nil {
		app.Event.Emit("soul:changed", map[string]any{
			"project_id": projectID,
			"soul":       dto,
		})
	}
	return nil
}

// SaveCustomSoul persists a custom soul profile to disk and memory.
func (s *SoulService) SaveCustomSoul(soulDTO dtos.SoulDTO) error {
	if !soul.IsValidSoulID(soulDTO.ID) {
		return fmt.Errorf("soul ID '%s' is invalid: must contain only alphanumeric characters, underscores, and dashes", soulDTO.ID)
	}
	sl := dtoToSoul(soulDTO)
	if err := s.controller.SaveSoul(sl); err != nil {
		return err
	}
	if app := application.Get(); app != nil && app.Event != nil {
		app.Event.Emit("soul:saved", soulToDTO(sl))
	}
	return nil
}

// DeleteCustomSoul deletes a custom soul profile from disk and memory.
func (s *SoulService) DeleteCustomSoul(soulID string) error {
	if !soul.IsValidSoulID(soulID) {
		return fmt.Errorf("soul ID '%s' is invalid", soulID)
	}
	if err := s.controller.DeleteSoul(soulID); err != nil {
		return err
	}
	if app := application.Get(); app != nil && app.Event != nil {
		app.Event.Emit("soul:deleted", map[string]string{"soul_id": soulID})
	}
	return nil
}

// RestoreDefaultSouls resets built-in default souls to their factory values.
func (s *SoulService) RestoreDefaultSouls() error {
	if err := s.controller.RestoreDefaults(); err != nil {
		return err
	}
	if app := application.Get(); app != nil && app.Event != nil {
		active := s.controller.GetSoul()
		app.Event.Emit("soul:changed", map[string]any{
			"soul_id": active.ID,
			"soul":    soulToDTO(active),
		})
	}
	return nil
}

// ExportSoulMarkdown exports a soul profile formatted as *.soul.md text.
func (s *SoulService) ExportSoulMarkdown(soulID string) (string, error) {
	sl, ok := s.controller.GetSoulByID(soulID)
	if !ok {
		return "", fmt.Errorf("soul '%s' not found", soulID)
	}
	return soul.FormatSoulMarkdown(sl)
}

// ImportSoulMarkdown imports and validates a soul profile from *.soul.md text.
func (s *SoulService) ImportSoulMarkdown(content string) (dtos.SoulDTO, error) {
	sl, err := soul.ParseSoulMarkdown(content)
	if err != nil {
		return dtos.SoulDTO{}, fmt.Errorf("parse soul markdown: %w", err)
	}
	if !soul.IsValidSoulID(sl.ID) {
		return dtos.SoulDTO{}, fmt.Errorf("imported soul ID '%s' is invalid: must contain only alphanumeric characters, underscores, and dashes", sl.ID)
	}
	if err := s.controller.SaveSoul(*sl); err != nil {
		return dtos.SoulDTO{}, fmt.Errorf("save imported soul: %w", err)
	}
	dto := soulToDTO(*sl)
	if app := application.Get(); app != nil && app.Event != nil {
		app.Event.Emit("soul:saved", dto)
	}
	return dto, nil
}

// TriggerSteerAction directly issues emergency stops or hot style patches without chatting.
func (s *SoulService) TriggerSteerAction(req dtos.SteerActionRequest) dtos.SteerActionResultDTO {
	var success bool
	var msg string

	switch req.Action {
	case "soft_stop":
		if req.JobID != "" {
			success = s.controller.RequestSoftStop(req.JobID)
		}
		if !success && req.ProjectID != "" {
			success = s.controller.StopJobsByProject(req.ProjectID)
		}
		if !success && req.JobID == "" && req.ProjectID == "" {
			success = s.controller.StopJobsByProject("")
		}
		msg = "Safe soft stop (Soft Stop) has been triggered."
	case "hard_abort":
		if req.JobID != "" {
			success = s.controller.RequestHardAbort(req.JobID)
		}
		if req.ProjectID != "" {
			projAborted := s.controller.AbortJobsByProject(req.ProjectID)
			success = success || projAborted
		}
		if !success && req.JobID == "" && req.ProjectID == "" {
			success = s.controller.AbortJobsByProject("")
		}
		msg = "Emergency abort (Hard Abort) succeeded."
	case "style_patch":
		success = s.controller.ApplyStylePatch(req.JobID, req.PatchValue)
		msg = fmt.Sprintf("Style updated: %s", req.PatchValue)
	default:
		success = false
		msg = "Invalid steering action."
	}

	return dtos.SteerActionResultDTO{
		Success:     success,
		Message:     msg,
		CurrentMood: string(s.controller.GetSoul().CurrentMood),
	}
}
