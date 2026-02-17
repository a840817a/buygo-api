package service

import (
	"context"
	"testing"

	"github.com/hatsubosi/buygo-api/internal/adapter/repository/memory"
	dauth "github.com/hatsubosi/buygo-api/internal/domain/auth"
	"github.com/hatsubosi/buygo-api/internal/domain/user"
)

func TestGroupBuyService_ManagerVisibility_ChangesAfterManagerUpdate(t *testing.T) {
	repo := memory.NewGroupBuyRepository()
	svc := NewGroupBuyService(repo)

	creatorCtx := dauth.NewContext(context.Background(), "creator-1", int(user.UserRoleCreator))
	managerCtx := dauth.NewContext(context.Background(), "manager-2", int(user.UserRoleCreator))

	gb, err := svc.CreateGroupBuy(creatorCtx, "GB", "desc", nil, "", nil, nil, nil, 0, nil, "")
	if err != nil {
		t.Fatalf("CreateGroupBuy error: %v", err)
	}

	// Manager should not see it before being assigned.
	before, err := svc.ListManagerGroupBuys(managerCtx, 50, 0)
	if err != nil {
		t.Fatalf("ListManagerGroupBuys before assign error: %v", err)
	}
	if len(before) != 0 {
		t.Fatalf("manager visible count before assign = %d, want 0", len(before))
	}

	// Creator assigns manager-2.
	_, err = svc.UpdateGroupBuy(creatorCtx, gb.ID, "", "", 0, nil, "", nil, nil, []string{"creator-1", "manager-2"}, 0, nil, "")
	if err != nil {
		t.Fatalf("UpdateGroupBuy assign manager error: %v", err)
	}

	afterAssign, err := svc.ListManagerGroupBuys(managerCtx, 50, 0)
	if err != nil {
		t.Fatalf("ListManagerGroupBuys after assign error: %v", err)
	}
	if len(afterAssign) != 1 {
		t.Fatalf("manager visible count after assign = %d, want 1", len(afterAssign))
	}

	// Creator removes manager-2 again.
	_, err = svc.UpdateGroupBuy(creatorCtx, gb.ID, "", "", 0, nil, "", nil, nil, []string{"creator-1"}, 0, nil, "")
	if err != nil {
		t.Fatalf("UpdateGroupBuy remove manager error: %v", err)
	}

	afterRemove, err := svc.ListManagerGroupBuys(managerCtx, 50, 0)
	if err != nil {
		t.Fatalf("ListManagerGroupBuys after remove error: %v", err)
	}
	if len(afterRemove) != 0 {
		t.Fatalf("manager visible count after remove = %d, want 0", len(afterRemove))
	}
}
