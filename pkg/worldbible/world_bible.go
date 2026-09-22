package worldbible

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"novelclaw/internal/dtos"
	"novelclaw/internal/gen/sqlc"
	"novelclaw/pkg/storage"
)

// Service provides high-level business logic for the World Bible & Lorebook.
type Service struct {
	store *storage.Storage
}

// NewService creates a new World Bible service.
func NewService(store *storage.Storage) *Service {
	return &Service{
		store: store,
	}
}

// ListCategories returns all categories for a project with entry counts.
func (s *Service) ListCategories(ctx context.Context, projectID string) ([]dtos.WorldCategoryDTO, error) {
	if s.store == nil || projectID == "" {
		return []dtos.WorldCategoryDTO{}, nil
	}

	cats, err := s.store.ListWorldCategories(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("list world categories: %w", err)
	}

	entries, err := s.store.ListWorldEntries(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("list world entries: %w", err)
	}

	counts := make(map[string]int)
	for _, e := range entries {
		counts[e.CategoryID]++
	}

	result := make([]dtos.WorldCategoryDTO, len(cats))
	for i, c := range cats {
		result[i] = dtos.WorldCategoryDTO{
			ID:           c.ID,
			ProjectID:    c.ProjectID,
			Slug:         c.Slug,
			Name:         c.Name,
			Icon:         c.Icon,
			Description:  c.Description,
			DisplayOrder: int(c.DisplayOrder),
			EntryCount:   counts[c.ID],
			CreatedAt:    c.CreatedAt.Format(time.RFC3339),
		}
	}
	return result, nil
}

// ListEntries returns all lorebook entries for a project or filtered by category.
func (s *Service) ListEntries(ctx context.Context, projectID, categoryID string) ([]dtos.WorldEntryDTO, error) {
	if s.store == nil || projectID == "" {
		return []dtos.WorldEntryDTO{}, nil
	}

	var rawEntries []sqlc.WorldEntry
	var err error
	if categoryID != "" {
		rawEntries, err = s.store.ListWorldEntriesByCategory(ctx, projectID, categoryID)
	} else {
		rawEntries, err = s.store.ListWorldEntries(ctx, projectID)
	}
	if err != nil {
		return nil, fmt.Errorf("list world entries: %w", err)
	}

	cats, _ := s.store.ListWorldCategories(ctx, projectID)
	catNames := make(map[string]string)
	catSlugs := make(map[string]string)
	for _, c := range cats {
		catNames[c.ID] = c.Name
		catSlugs[c.ID] = c.Slug
	}

	result := make([]dtos.WorldEntryDTO, len(rawEntries))
	for i, e := range rawEntries {
		var aliases []string
		_ = json.Unmarshal([]byte(e.AliasesJson), &aliases)
		var attrs map[string]string
		_ = json.Unmarshal([]byte(e.AttributesJson), &attrs)
		if attrs == nil {
			attrs = make(map[string]string)
		}

		result[i] = dtos.WorldEntryDTO{
			ID:                 e.ID,
			ProjectID:          e.ProjectID,
			CategoryID:         e.CategoryID,
			CategorySlug:       catSlugs[e.CategoryID],
			CategoryName:       catNames[e.CategoryID],
			Name:               e.Name,
			Aliases:            aliases,
			Summary:            e.Summary,
			FullDescription:    e.FullDescription,
			Attributes:         attrs,
			DiscoveredBy:       e.DiscoveredBy,
			SourceChapterIndex: e.SourceChapterIndex,
			IsVerified:         e.IsVerified == 1,
			CreatedAt:          e.CreatedAt.Format(time.RFC3339),
			UpdatedAt:          e.UpdatedAt.Format(time.RFC3339),
		}
	}
	return result, nil
}

// UpsertCategory adds or modifies a category.
func (s *Service) UpsertCategory(ctx context.Context, req dtos.UpsertWorldCategoryRequest) (*dtos.WorldCategoryDTO, error) {
	if s.store == nil || req.ProjectID == "" {
		return nil, fmt.Errorf("invalid storage or project id")
	}

	id := req.ID
	if strings.TrimSpace(id) == "" {
		id = fmt.Sprintf("cat_%s", strings.ReplaceAll(uuid.New().String(), "-", "")[:12])
	}
	slug := strings.TrimSpace(strings.ToLower(req.Slug))
	if slug == "" {
		slug = strings.TrimSpace(strings.ToLower(req.Name))
		slug = strings.ReplaceAll(slug, " ", "_")
	}
	icon := req.Icon
	if icon == "" {
		icon = "Boxes"
	}

	cat, err := s.store.UpsertWorldCategory(ctx, sqlc.UpsertWorldCategoryParams{
		ID:           id,
		ProjectID:    req.ProjectID,
		Slug:         slug,
		Name:         req.Name,
		Icon:         icon,
		Description:  req.Description,
		DisplayOrder: int64(req.DisplayOrder),
	})
	if err != nil {
		return nil, fmt.Errorf("upsert category: %w", err)
	}

	return &dtos.WorldCategoryDTO{
		ID:           cat.ID,
		ProjectID:    cat.ProjectID,
		Slug:         cat.Slug,
		Name:         cat.Name,
		Icon:         cat.Icon,
		Description:  cat.Description,
		DisplayOrder: int(cat.DisplayOrder),
		CreatedAt:    cat.CreatedAt.Format(time.RFC3339),
	}, nil
}

// DeleteCategory deletes a category and its entries.
func (s *Service) DeleteCategory(ctx context.Context, id string) error {
	if s.store == nil {
		return nil
	}
	return s.store.DeleteWorldCategory(ctx, id)
}

// UpsertEntry creates or updates an entry.
func (s *Service) UpsertEntry(ctx context.Context, req dtos.UpsertWorldEntryRequest) (*dtos.WorldEntryDTO, error) {
	if s.store == nil || req.ProjectID == "" {
		return nil, fmt.Errorf("invalid storage or project id")
	}

	id := req.ID
	var isVerified int64 = 0
	if req.IsVerified {
		isVerified = 1
	}

	if strings.TrimSpace(id) == "" {
		if existing, err := s.store.GetWorldEntryByName(ctx, req.ProjectID, strings.ToLower(req.Name)); err == nil && existing.ID != "" {
			id = existing.ID
			if existing.IsVerified == 1 && !req.IsVerified {
				isVerified = 1
			}
		} else {
			id = fmt.Sprintf("entry_%s", strings.ReplaceAll(uuid.New().String(), "-", "")[:12])
		}
	}

	aliasesJSON, _ := json.Marshal(req.Aliases)
	if len(req.Aliases) == 0 {
		aliasesJSON = []byte("[]")
	}

	attrsJSON, _ := json.Marshal(req.Attributes)
	if req.Attributes == nil {
		attrsJSON = []byte("{}")
	}



	entry, err := s.store.UpsertWorldEntry(ctx, sqlc.UpsertWorldEntryParams{
		ID:                 id,
		ProjectID:          req.ProjectID,
		CategoryID:         req.CategoryID,
		Name:               req.Name,
		AliasesJson:        string(aliasesJSON),
		Summary:            req.Summary,
		FullDescription:    req.FullDescription,
		AttributesJson:     string(attrsJSON),
		DiscoveredBy:       "manual",
		SourceChapterIndex: req.SourceChapterIndex,
		IsVerified:         isVerified,
	})
	if err != nil {
		return nil, fmt.Errorf("upsert entry: %w", err)
	}

	cat, _ := s.store.GetWorldCategoryByID(ctx, entry.CategoryID)

	return &dtos.WorldEntryDTO{
		ID:                 entry.ID,
		ProjectID:          entry.ProjectID,
		CategoryID:         entry.CategoryID,
		CategorySlug:       cat.Slug,
		CategoryName:       cat.Name,
		Name:               entry.Name,
		Aliases:            req.Aliases,
		Summary:            entry.Summary,
		FullDescription:    entry.FullDescription,
		Attributes:         req.Attributes,
		DiscoveredBy:       entry.DiscoveredBy,
		SourceChapterIndex: entry.SourceChapterIndex,
		IsVerified:         entry.IsVerified == 1,
		CreatedAt:          entry.CreatedAt.Format(time.RFC3339),
		UpdatedAt:          entry.UpdatedAt.Format(time.RFC3339),
	}, nil
}

// VerifyEntry updates the verification status of an entry.
func (s *Service) VerifyEntry(ctx context.Context, id string, verified bool) error {
	if s.store == nil {
		return nil
	}
	return s.store.VerifyWorldEntry(ctx, id, verified)
}

// DeleteEntry deletes an entry by ID.
func (s *Service) DeleteEntry(ctx context.Context, id string) error {
	if s.store == nil {
		return nil
	}
	return s.store.DeleteWorldEntry(ctx, id)
}
