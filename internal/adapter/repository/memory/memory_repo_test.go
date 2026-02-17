package memory

import (
	"context"
	"testing"

	"github.com/hatsubosi/buygo-api/internal/domain/user"
)

func TestUserRepository(t *testing.T) {
	r := NewUserRepository()
	ctx := context.Background()

	u := &user.User{ID: "u1", Email: "test@example.com", Name: "Test"}
	_ = r.Create(ctx, u)

	// GetByID
	got, _ := r.GetByID(ctx, "u1")
	if got.ID != "u1" {
		t.Errorf("GetByID failed")
	}

	// GetByEmail
	got, _ = r.GetByEmail(ctx, "test@example.com")
	if got.ID != "u1" {
		t.Errorf("GetByEmail failed")
	}

	// Update
	u.Name = "Updated"
	_ = r.Update(ctx, u)
	got, _ = r.GetByID(ctx, "u1")
	if got.Name != "Updated" {
		t.Errorf("Update failed")
	}

	// List
	list, _ := r.List(ctx, 10, 0)
	if len(list) != 1 {
		t.Errorf("List failed")
	}
}

func TestUserRepository_NotFound(t *testing.T) {
	r := NewUserRepository()
	ctx := context.Background()

	_, err := r.GetByID(ctx, "none")
	if err != user.ErrNotFound {
		t.Errorf("expected ErrNotFound")
	}

	_, err = r.GetByEmail(ctx, "none@example.com")
	if err != user.ErrNotFound {
		t.Errorf("expected ErrNotFound")
	}
}
