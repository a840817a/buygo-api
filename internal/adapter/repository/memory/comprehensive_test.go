package memory

import (
	"context"
	"testing"
	"time"

	"github.com/hatsubosi/buygo-api/internal/domain/groupbuy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGroupBuyRepository_List_MixedAccess(t *testing.T) {
	repo := NewGroupBuyRepository()
	ctx := context.Background()
	now := time.Now()

	// 1. My draft
	_ = repo.Create(ctx, &groupbuy.GroupBuy{
		ID:        "my-draft",
		Status:    groupbuy.GroupBuyStatusDraft,
		CreatorID: "me",
		CreatedAt: now,
	})
	// 2. Public active
	_ = repo.Create(ctx, &groupbuy.GroupBuy{
		ID:        "public-active",
		Status:    groupbuy.GroupBuyStatusActive,
		CreatorID: "other",
		CreatedAt: now.Add(-time.Hour),
	})
	// 3. Other's draft (should not see)
	_ = repo.Create(ctx, &groupbuy.GroupBuy{
		ID:        "other-draft",
		Status:    groupbuy.GroupBuyStatusDraft,
		CreatorID: "other",
		CreatedAt: now.Add(-2 * time.Hour),
	})

	t.Run("Me_NotManageOnly", func(t *testing.T) {
		// Should see "my-draft" and "public-active"
		res, err := repo.List(ctx, 10, 0, "me", false, false)
		require.NoError(t, err)
		assert.Len(t, res, 2)
		ids := []string{res[0].ID, res[1].ID}
		assert.Contains(t, ids, "my-draft")
		assert.Contains(t, ids, "public-active")
		assert.NotContains(t, ids, "other-draft")
	})
}

func TestGroupBuyRepository_UpdateOrder(t *testing.T) {
	repo := NewGroupBuyRepository()
	ctx := context.Background()

	o := &groupbuy.Order{
		ID:    "order-1",
		Note:  "Original",
		Items: []*groupbuy.OrderItem{{ProductID: "p1", Quantity: 1}},
	}
	repo.orders[o.ID] = o

	t.Run("Success", func(t *testing.T) {
		o.Note = "Updated"
		err := repo.UpdateOrder(ctx, o)
		require.NoError(t, err)
		assert.Equal(t, "Updated", repo.orders["order-1"].Note)
	})

	t.Run("NotFound", func(t *testing.T) {
		err := repo.UpdateOrder(ctx, &groupbuy.Order{ID: "nonexistent"})
		assert.Error(t, err)
	})
}

func TestGroupBuyRepository_BatchUpdateOrderItemStatus(t *testing.T) {
	repo := NewGroupBuyRepository()
	ctx := context.Background()

	// Setup orders with items
	repo.orders["o1"] = &groupbuy.Order{
		ID:         "o1",
		GroupBuyID: "gb1",
		Items: []*groupbuy.OrderItem{
			{ID: "i1", ProductID: "p1", SpecID: "s1", Status: groupbuy.OrderItemStatusUnordered},
		},
	}
	repo.orders["o2"] = &groupbuy.Order{
		ID:         "o2",
		GroupBuyID: "gb1",
		Items: []*groupbuy.OrderItem{
			{ID: "i2", ProductID: "p1", SpecID: "s1", Status: groupbuy.OrderItemStatusUnordered},
		},
	}

	t.Run("PartialUpdate", func(t *testing.T) {
		n, ids, err := repo.BatchUpdateOrderItemStatus(ctx, "gb1", "s1", []int{int(groupbuy.OrderItemStatusUnordered)}, int(groupbuy.OrderItemStatusOrdered), 1)
		require.NoError(t, err)
		assert.Equal(t, int64(1), n)
		assert.Len(t, ids, 1)

		// Check one is updated, one is not
		i1 := repo.orders["o1"].Items[0].Status
		i2 := repo.orders["o2"].Items[0].Status

		// Order in map is random, but one should be Ordered, one Unordered
		if i1 == groupbuy.OrderItemStatusOrdered {
			assert.Equal(t, groupbuy.OrderItemStatusUnordered, i2)
		} else {
			assert.Equal(t, groupbuy.OrderItemStatusOrdered, i2)
			assert.Equal(t, groupbuy.OrderItemStatusUnordered, i1)
		}
	})
}
