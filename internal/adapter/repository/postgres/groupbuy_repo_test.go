package postgres

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/hatsubosi/buygo-api/internal/adapter/repository/postgres/model"
	"github.com/hatsubosi/buygo-api/internal/domain/groupbuy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newTestDB(t *testing.T) *gorm.DB {
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	assert.NoError(t, err)

	err = db.AutoMigrate(
		&model.User{},
		&model.GroupBuy{},
		&model.Product{},
		&model.ProductSpec{},
		&model.Order{},
		&model.OrderItem{},
	)
	assert.NoError(t, err)

	return db
}

func TestGroupBuyRepository_UpdateShippingConfigs(t *testing.T) {
	db := newTestDB(t)
	repo := NewGroupBuyRepository(db)
	ctx := context.Background()

	// 1. Create Initial Project
	creatorID := "user-1"
	db.Create(&model.User{ID: creatorID, Name: "Creator"})

	proj := &groupbuy.GroupBuy{
		ID:          "proj-1",
		Title:       "Test Project",
		Description: "Desc",
		Status:      groupbuy.GroupBuyStatusActive,
		CreatorID:   creatorID,
		CreatedAt:   time.Now(),
		ShippingConfigs: []*groupbuy.ShippingConfig{
			{ID: "sc-1", Name: "Initial Meetup", Type: groupbuy.ShippingTypeMeetup, Price: 0},
		},
	}
	err := repo.Create(ctx, proj)
	assert.NoError(t, err)

	// Verify Initial State
	saved, err := repo.GetByID(ctx, "proj-1")
	assert.NoError(t, err)
	assert.Len(t, saved.ShippingConfigs, 1)
	assert.Equal(t, groupbuy.ShippingTypeMeetup, saved.ShippingConfigs[0].Type)

	// 2. Update with New Shipping Config (Simulate Type Change / Add)
	// Changing "Initial Meetup" to "Delivery" type (Type 1), adding "Store Pickup" (Type 2)
	updatedConfigs := []*groupbuy.ShippingConfig{
		{ID: "sc-1", Name: "Changed to Delivery", Type: groupbuy.ShippingTypeDelivery, Price: 100},
		{ID: "sc-2", Name: "New Store Pickup", Type: groupbuy.ShippingTypeStorePickup, Price: 60},
	}
	proj.ShippingConfigs = updatedConfigs

	err = repo.Update(ctx, proj)
	assert.NoError(t, err)

	// 3. Verify Persistence
	// We need to fetch again to check DB state
	final, err := repo.GetByID(ctx, "proj-1")
	assert.NoError(t, err)
	assert.Len(t, final.ShippingConfigs, 2)

	// Check Types
	var delivery *groupbuy.ShippingConfig
	var pickup *groupbuy.ShippingConfig

	for _, sc := range final.ShippingConfigs {
		if sc.ID == "sc-1" {
			delivery = sc
		} else if sc.ID == "sc-2" {
			pickup = sc
		}
	}

	assert.NotNil(t, delivery)
	assert.Equal(t, groupbuy.ShippingTypeDelivery, delivery.Type, "Type should be Delivery (1)")
	assert.Equal(t, int64(100), delivery.Price)

	assert.NotNil(t, pickup)
	assert.Equal(t, groupbuy.ShippingTypeStorePickup, pickup.Type, "Type should be Store Pickup (2)")

	// 4. Test Meetup Again
	proj.ShippingConfigs = []*groupbuy.ShippingConfig{
		{ID: "sc-3", Name: "Meetup Only", Type: groupbuy.ShippingTypeMeetup, Price: 0},
	}
	err = repo.Update(ctx, proj)
	assert.NoError(t, err)

	final2, err := repo.GetByID(ctx, "proj-1")
	assert.NoError(t, err)
	assert.Len(t, final2.ShippingConfigs, 1)
	assert.Equal(t, groupbuy.ShippingTypeMeetup, final2.ShippingConfigs[0].Type, "Type should be Meetup (3)")
}

// --- Helper ---

func createTestGroupBuy(t *testing.T, db *gorm.DB, repo *GroupBuyRepository, id string, status groupbuy.GroupBuyStatus) *groupbuy.GroupBuy {
	t.Helper()
	creatorID := "creator-" + id
	require.NoError(t, db.Create(&model.User{
		ID:    creatorID,
		Name:  "Creator " + id,
		Email: creatorID + "@test.local",
	}).Error)

	gb := &groupbuy.GroupBuy{
		ID:          id,
		Title:       "GroupBuy " + id,
		Description: "Desc",
		Status:      status,
		CreatorID:   creatorID,
		CreatedAt:   time.Now(),
	}
	require.NoError(t, repo.Create(context.Background(), gb))
	return gb
}

// --- Create + GetByID ---

func TestGroupBuyRepository_CreateAndGetByID(t *testing.T) {
	db := newTestDB(t)
	repo := NewGroupBuyRepository(db)
	ctx := context.Background()

	require.NoError(t, db.Create(&model.User{ID: "user-1", Name: "Creator", Email: "user-1@test.local"}).Error)

	gb := &groupbuy.GroupBuy{
		ID:           "gb-1",
		Title:        "Test GB",
		Description:  "A test group buy",
		Status:       groupbuy.GroupBuyStatusDraft,
		ExchangeRate: 0.22,
		CreatorID:    "user-1",
		CreatedAt:    time.Now(),
		Rounding:     &groupbuy.RoundingConfig{Method: groupbuy.RoundingMethodFloor, Digit: 0},
	}

	err := repo.Create(ctx, gb)
	require.NoError(t, err)

	got, err := repo.GetByID(ctx, "gb-1")
	require.NoError(t, err)
	assert.Equal(t, "Test GB", got.Title)
	assert.Equal(t, "A test group buy", got.Description)
	assert.Equal(t, groupbuy.GroupBuyStatusDraft, got.Status)
	assert.Equal(t, 0.22, got.ExchangeRate)
	assert.Equal(t, "user-1", got.CreatorID)
}

func TestGroupBuyRepository_GetByID_NotFound(t *testing.T) {
	db := newTestDB(t)
	repo := NewGroupBuyRepository(db)

	_, err := repo.GetByID(context.Background(), "nonexistent")
	assert.Error(t, err)
}

// --- List Filtering ---

func TestGroupBuyRepository_List_PublicOnly(t *testing.T) {
	db := newTestDB(t)
	repo := NewGroupBuyRepository(db)

	createTestGroupBuy(t, db, repo, "draft-1", groupbuy.GroupBuyStatusDraft)
	createTestGroupBuy(t, db, repo, "active-1", groupbuy.GroupBuyStatusActive)
	createTestGroupBuy(t, db, repo, "ended-1", groupbuy.GroupBuyStatusEnded)

	// Anonymous user (no userID) should only see Active + Ended
	list, err := repo.List(context.Background(), 10, 0, "", false, false)
	require.NoError(t, err)
	assert.Len(t, list, 2)
}

func TestGroupBuyRepository_List_SysAdmin(t *testing.T) {
	db := newTestDB(t)
	repo := NewGroupBuyRepository(db)

	createTestGroupBuy(t, db, repo, "draft-1", groupbuy.GroupBuyStatusDraft)
	createTestGroupBuy(t, db, repo, "active-1", groupbuy.GroupBuyStatusActive)

	// SysAdmin sees all
	list, err := repo.List(context.Background(), 10, 0, "admin", true, false)
	require.NoError(t, err)
	assert.Len(t, list, 2)
}

// --- AddProduct ---

func TestGroupBuyRepository_AddProduct(t *testing.T) {
	db := newTestDB(t)
	repo := NewGroupBuyRepository(db)
	ctx := context.Background()

	gb := createTestGroupBuy(t, db, repo, "gb-1", groupbuy.GroupBuyStatusActive)

	prod := &groupbuy.Product{
		ID:            "prod-1",
		GroupBuyID:    gb.ID,
		Name:          "Widget",
		PriceOriginal: 1000,
		PriceFinal:    220,
		ExchangeRate:  0.22,
		Specs: []*groupbuy.ProductSpec{
			{ID: "spec-1", ProductID: "prod-1", Name: "Red"},
			{ID: "spec-2", ProductID: "prod-1", Name: "Blue"},
		},
	}

	err := repo.AddProduct(ctx, prod)
	require.NoError(t, err)

	got, err := repo.GetByID(ctx, gb.ID)
	require.NoError(t, err)
	require.Len(t, got.Products, 1)
	assert.Equal(t, "Widget", got.Products[0].Name)
	assert.Len(t, got.Products[0].Specs, 2)
}

// --- Order CRUD ---

func TestGroupBuyRepository_OrderCRUD(t *testing.T) {
	db := newTestDB(t)
	repo := NewGroupBuyRepository(db)
	ctx := context.Background()

	gb := createTestGroupBuy(t, db, repo, "gb-1", groupbuy.GroupBuyStatusActive)

	// Create user for order
	require.NoError(t, db.Create(&model.User{ID: "buyer-1", Name: "Buyer", Email: "buyer-1@test.local"}).Error)

	order := &groupbuy.Order{
		ID:              "order-1",
		GroupBuyID:      gb.ID,
		UserID:          "buyer-1",
		TotalAmount:     500,
		PaymentStatus:   groupbuy.PaymentStatusUnset,
		ContactInfo:     "contact@test.com",
		ShippingAddress: "123 Main St",
		Items: []*groupbuy.OrderItem{
			{ID: "item-1", OrderID: "order-1", ProductID: "prod-1", SpecID: "spec-1", Quantity: 2, Price: 250},
		},
		CreatedAt: time.Now(),
	}

	// Create
	err := repo.CreateOrder(ctx, order)
	require.NoError(t, err)

	// Get
	got, err := repo.GetOrder(ctx, "order-1")
	require.NoError(t, err)
	assert.Equal(t, "buyer-1", got.UserID)
	assert.Equal(t, int64(500), got.TotalAmount)
	require.Len(t, got.Items, 1)
	assert.Equal(t, 2, got.Items[0].Quantity)

	// List by group buy
	orders, err := repo.ListOrders(ctx, gb.ID, "")
	require.NoError(t, err)
	assert.Len(t, orders, 1)

	// List by user
	orders, err = repo.ListOrders(ctx, "", "buyer-1")
	require.NoError(t, err)
	assert.Len(t, orders, 1)

	// Update payment status
	err = repo.UpdateOrderPaymentStatus(ctx, "order-1", groupbuy.PaymentStatusConfirmed)
	require.NoError(t, err)

	got, err = repo.GetOrder(ctx, "order-1")
	require.NoError(t, err)
	assert.Equal(t, groupbuy.PaymentStatusConfirmed, got.PaymentStatus)
}

func TestGroupBuyRepository_ListOrders_SortedByCreatedAtDesc(t *testing.T) {
	db := newTestDB(t)
	repo := NewGroupBuyRepository(db)
	ctx := context.Background()

	gb := createTestGroupBuy(t, db, repo, "gb-sort", groupbuy.GroupBuyStatusActive)
	require.NoError(t, db.Create(&model.User{ID: "buyer-sort", Name: "Buyer", Email: "buyer-sort@test.local"}).Error)

	oldOrder := &groupbuy.Order{
		ID:              "order-old",
		GroupBuyID:      gb.ID,
		UserID:          "buyer-sort",
		TotalAmount:     100,
		PaymentStatus:   groupbuy.PaymentStatusUnset,
		ContactInfo:     "buyer",
		ShippingAddress: "addr",
		CreatedAt:       time.Now().Add(-1 * time.Hour),
	}
	newOrder := &groupbuy.Order{
		ID:              "order-new",
		GroupBuyID:      gb.ID,
		UserID:          "buyer-sort",
		TotalAmount:     100,
		PaymentStatus:   groupbuy.PaymentStatusUnset,
		ContactInfo:     "buyer",
		ShippingAddress: "addr",
		CreatedAt:       time.Now(),
	}

	require.NoError(t, repo.CreateOrder(ctx, oldOrder))
	require.NoError(t, repo.CreateOrder(ctx, newOrder))

	orders, err := repo.ListOrders(ctx, gb.ID, "buyer-sort")
	require.NoError(t, err)
	require.Len(t, orders, 2)
	assert.Equal(t, "order-new", orders[0].ID)
	assert.Equal(t, "order-old", orders[1].ID)
}

func TestGroupBuyRepository_GetOrder_NotFound(t *testing.T) {
	db := newTestDB(t)
	repo := NewGroupBuyRepository(db)

	_, err := repo.GetOrder(context.Background(), "nonexistent")
	assert.Error(t, err)
}
