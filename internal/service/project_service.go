package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	"buygo/internal/domain/project"
)

type ProjectService struct {
	repo project.Repository
}

func NewProjectService(repo project.Repository) *ProjectService {
	return &ProjectService{
		repo: repo,
	}
}

func (s *ProjectService) CreateProject(ctx context.Context, userID string, title string) (*project.Project, error) {
	// TODO: Validate user exists (via User Service or similar)
	// For now, assume user exists.

	now := time.Now()
	p := &project.Project{
		ID:        uuid.New().String(),
		Title:     title,
		CreatorID: userID,
		IsActive:  true, // Default to true or draft?
		CreatedAt: now,
		// Managers: []string{userID}, // Creator is usually a manager
	}

	if err := s.repo.Create(ctx, p); err != nil {
		return nil, err
	}

	return p, nil
}

func (s *ProjectService) GetProject(ctx context.Context, id string) (*project.Project, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *ProjectService) ListProjects(ctx context.Context, limit, offset int) ([]*project.Project, error) {
	return s.repo.List(ctx, limit, offset)
}
