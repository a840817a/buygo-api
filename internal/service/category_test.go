package service

import (
	"context"
	"errors"
	"sort"
	"testing"

	"github.com/hatsubosi/buygo-api/internal/adapter/repository/memory"
	"github.com/hatsubosi/buygo-api/internal/domain/auth"
	"github.com/hatsubosi/buygo-api/internal/domain/user"
)

func TestGroupBuyService_CategoryAccessControl(t *testing.T) {
	repo := memory.NewGroupBuyRepository()
	svc := NewGroupBuyService(repo)

	// Contexts
	adminCtx := auth.NewContext(context.Background(), "admin", int(user.UserRoleSysAdmin))
	creatorCtx := auth.NewContext(context.Background(), "creator-1", int(user.UserRoleCreator))
	userCtx := auth.NewContext(context.Background(), "user-1", int(user.UserRoleUser))
	anonCtx := context.Background()

	// 1. Create Category
	// Anon -> Fail
	_, err := svc.CreateCategory(anonCtx, "Cat1", []string{"Color"})
	if !errors.Is(err, ErrPermissionDenied) {
		t.Errorf("Anon should not create category, got %v", err)
	}

	// User -> Fail
	_, err = svc.CreateCategory(userCtx, "Cat1", []string{"Color"})
	if !errors.Is(err, ErrPermissionDenied) {
		t.Errorf("Regular User should not create category, got %v", err)
	}

	// Creator -> Fail
	_, err = svc.CreateCategory(creatorCtx, "Cat1", []string{"Color"})
	if !errors.Is(err, ErrPermissionDenied) {
		t.Errorf("Creator should not create category, got %v", err)
	}

	// SysAdmin -> Success
	_, err = svc.CreateCategory(adminCtx, "Cat1", []string{"Color"})
	if err != nil {
		t.Errorf("SysAdmin should create category, got %v", err)
	}

	// 2. List Categories (Public Access)
	// Create another for sorting test
	_, err = svc.CreateCategory(adminCtx, "AliceCat", []string{"Size"})
	if err != nil {
		t.Fatal(err)
	}

	// Anon -> Fail (Service requires Auth)
	_, err = svc.ListCategories(anonCtx)
	if !errors.Is(err, ErrPermissionDenied) {
		t.Errorf("Anon should not list categories (requires auth), got %v", err)
	}

	// User -> Success
	cats, err := svc.ListCategories(userCtx)
	if err != nil {
		t.Errorf("User should list categories, got %v", err)
	}
	if len(cats) != 2 {
		t.Errorf("Expected 2 categories, got %d", len(cats))
	}

	// Verify Sorting (AliceCat should be first if sorted by name)
	// Memory repo implementation sorts by name
	sort.Slice(cats, func(i, j int) bool {
		return cats[i].Name < cats[j].Name
	})

	if cats[0].Name != "AliceCat" {
		t.Errorf("Expected AliceCat first, got %s", cats[0].Name)
	}
	if cats[1].Name != "Cat1" {
		t.Errorf("Expected Cat1 second, got %s", cats[1].Name)
	}

	// Verify Specs
	if len(cats[1].SpecNames) != 1 || cats[1].SpecNames[0] != "Color" {
		t.Errorf("Expected Spec Color, got %v", cats[1].SpecNames)
	}
}
