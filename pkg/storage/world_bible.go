package storage

import (
	"context"
	"strings"

	"novelclaw/internal/gen/sqlc"
)

// ListWorldCategories returns all categories for a project ordered by display_order.
func (s *Storage) ListWorldCategories(ctx context.Context, projectID string) ([]sqlc.WorldCategory, error) {
	return s.q.ListWorldCategories(ctx, projectID)
}

// GetWorldCategoryBySlug retrieves a category by project and slug.
func (s *Storage) GetWorldCategoryBySlug(ctx context.Context, projectID, slug string) (sqlc.WorldCategory, error) {
	return s.q.GetWorldCategoryBySlug(ctx, sqlc.GetWorldCategoryBySlugParams{
		ProjectID: projectID,
		Slug:      slug,
	})
}

// GetWorldCategoryByID retrieves a category by its primary key ID.
func (s *Storage) GetWorldCategoryByID(ctx context.Context, id string) (sqlc.WorldCategory, error) {
	return s.q.GetWorldCategoryByID(ctx, id)
}

// UpsertWorldCategory inserts or updates a world category.
func (s *Storage) UpsertWorldCategory(ctx context.Context, params sqlc.UpsertWorldCategoryParams) (sqlc.WorldCategory, error) {
	return s.q.UpsertWorldCategory(ctx, params)
}

// DeleteWorldCategory removes a world category and cascades to its entries.
func (s *Storage) DeleteWorldCategory(ctx context.Context, id string) error {
	return s.q.DeleteWorldCategory(ctx, id)
}

// ListWorldEntries retrieves all world entries for a given project.
func (s *Storage) ListWorldEntries(ctx context.Context, projectID string) ([]sqlc.WorldEntry, error) {
	return s.q.ListWorldEntries(ctx, projectID)
}

// ListWorldEntriesByCategory retrieves entries within a specific category.
func (s *Storage) ListWorldEntriesByCategory(ctx context.Context, projectID, categoryID string) ([]sqlc.WorldEntry, error) {
	return s.q.ListWorldEntriesByCategory(ctx, sqlc.ListWorldEntriesByCategoryParams{
		ProjectID:  projectID,
		CategoryID: categoryID,
	})
}

// GetWorldEntryByID retrieves an entry by its ID.
func (s *Storage) GetWorldEntryByID(ctx context.Context, id string) (sqlc.WorldEntry, error) {
	return s.q.GetWorldEntryByID(ctx, id)
}

// GetWorldEntryByName finds an entry by case-insensitive name match.
func (s *Storage) GetWorldEntryByName(ctx context.Context, projectID, name string) (sqlc.WorldEntry, error) {
	return s.q.GetWorldEntryByName(ctx, sqlc.GetWorldEntryByNameParams{
		ProjectID: projectID,
		LowerName: strings.ToLower(name),
	})
}

// UpsertWorldEntry inserts or updates an entry in the Lorebook.
func (s *Storage) UpsertWorldEntry(ctx context.Context, params sqlc.UpsertWorldEntryParams) (sqlc.WorldEntry, error) {
	return s.q.UpsertWorldEntry(ctx, params)
}

// VerifyWorldEntry marks an entry as verified (1) or unverified/AI suggestion (0).
func (s *Storage) VerifyWorldEntry(ctx context.Context, id string, verified bool) error {
	var flag int64
	if verified {
		flag = 1
	}
	return s.q.VerifyWorldEntry(ctx, sqlc.VerifyWorldEntryParams{
		IsVerified: flag,
		ID:         id,
	})
}

// DeleteWorldEntry deletes a specific Lorebook entry.
func (s *Storage) DeleteWorldEntry(ctx context.Context, id string) error {
	return s.q.DeleteWorldEntry(ctx, id)
}

// CountWorldEntries returns total number of entries in a project.
func (s *Storage) CountWorldEntries(ctx context.Context, projectID string) (int64, error) {
	return s.q.CountWorldEntries(ctx, projectID)
}
