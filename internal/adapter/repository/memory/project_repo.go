package memory

import (
	"context"
	"errors"
	"sync"

	"buygo/internal/domain/project"
)

type ProjectRepository struct {
	mu       sync.RWMutex
	projects map[string]*project.Project
}

func NewProjectRepository() *ProjectRepository {
	return &ProjectRepository{
		projects: make(map[string]*project.Project),
	}
}

func (r *ProjectRepository) Create(ctx context.Context, p *project.Project) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.projects[p.ID]; exists {
		return errors.New("project already exists")
	}
	r.projects[p.ID] = p
	return nil
}

func (r *ProjectRepository) GetByID(ctx context.Context, id string) (*project.Project, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, exists := r.projects[id]
	if !exists {
		return nil, errors.New("project not found")
	}
	return p, nil
}

func (r *ProjectRepository) List(ctx context.Context, limit int, offset int) ([]*project.Project, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var res []*project.Project
	count := 0
	// Note: Map iteration order is random, real impl needs sorting
	for _, p := range r.projects {
		if count >= offset && len(res) < limit {
			res = append(res, p)
		}
		count++
		if len(res) >= limit {
			break
		}
	}
	return res, nil
}

func (r *ProjectRepository) Update(ctx context.Context, p *project.Project) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.projects[p.ID]; !exists {
		return errors.New("project not found")
	}
	r.projects[p.ID] = p
	return nil
}
