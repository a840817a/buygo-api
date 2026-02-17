package memory

import (
	"context"
	"strings"
	"sync"

	"github.com/hatsubosi/buygo-api/internal/domain/user"
)

type UserRepository struct {
	mu    sync.RWMutex
	users map[string]*user.User
}

func NewUserRepository() *UserRepository {
	return &UserRepository{
		users: make(map[string]*user.User),
	}
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*user.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	u, ok := r.users[id]
	if !ok {
		return nil, user.ErrNotFound
	}
	return u, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, u := range r.users {
		if strings.EqualFold(u.Email, email) {
			return u, nil
		}
	}
	return nil, user.ErrNotFound
}

func (r *UserRepository) Create(ctx context.Context, u *user.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.users[u.ID]; ok {
		return nil // Should be error already exists? But for upsert logic...
	}
	r.users[u.ID] = u
	return nil
}

func (r *UserRepository) Update(ctx context.Context, u *user.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.users[u.ID]; !ok {
		return user.ErrNotFound
	}
	r.users[u.ID] = u
	return nil
}

func (r *UserRepository) List(ctx context.Context, limit, offset int) ([]*user.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	users := make([]*user.User, 0, len(r.users))
	for _, u := range r.users {
		users = append(users, u)
	}

	// Simple pagination on unsorted list (map iteration is random but sufficient for basic mock)
	if offset >= len(users) {
		return []*user.User{}, nil
	}
	end := offset + limit
	if end > len(users) {
		end = len(users)
	}
	return users[offset:end], nil
}
