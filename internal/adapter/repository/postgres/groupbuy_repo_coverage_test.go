package postgres

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hatsubosi/buygo-api/internal/adapter/repository/postgres/model"
	"github.com/hatsubosi/buygo-api/internal/domain/groupbuy"
	"github.com/hatsubosi/buygo-api/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newGroupBuyTestDB(t *testing.T) *gorm.DB {
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&model.User{},
		&model.GroupBuy{},
		&model.Product{},
		&model.ProductSpec{},
		&model.Order{},
		&model.OrderItem{},
		&model.Category{},
		&model.PriceTemplate{},
	))
	return db
}

// seedUser inserts a User record to satisfy foreign key constraints.
func seedUser(t *testing.T, db *gorm.DB, id, name, email string) {
	require.NoError(t, db.Create(&model.User{ID: id, Name: name, Email: email}).Error)
}

// seedGroupBuy creates a minimal GroupBuy with one product (and one spec) via the repo.
func seedGroupBuy(t *testing.T, repo *GroupBuyRepository, ctx context.Context, gbID, creatorID string) *groupbuy.GroupBuy {
	gb := &groupbuy.GroupBuy{
		ID:          gbID,
		Title:       "GB " + gbID,
		Description: "desc",
		CreatorID:   creatorID,
		Status:      groupbuy.GroupBuyStatusActive,
		Products: []*groupbuy.Product{
			{
				ID:            gbID + "-prod",
				GroupBuyID:    gbID,
				Name:          "Product in " + gbID,
				PriceOriginal: 100,
				Specs: []*groupbuy.ProductSpec{
					{ID: gbID + "-spec", ProductID: gbID + "-prod", Name: "Default"},
				},
			},
		},
	}
	require.NoError(t, repo.Create(ctx, gb))
	return gb
}

// ---------------------------------------------------------------------------
// Test: Order CRUD
// ---------------------------------------------------------------------------

func TestGroupBuyRepository_OrderCRUD(t *testing.T) {
	db := newGroupBuyTestDB(t)
	repo := NewGroupBuyRepository(db)
	ctx := context.Background()

	seedUser(t, db, "u1", "User 1", "u1@test.com")
	seedUser(t, db, "u2", "User 2", "u2@test.com")
	seedGroupBuy(t, repo, ctx, "gb1", "u1")

	// --- CreateOrder ---
	order := &groupbuy.Order{
		ID:            "order-1",
		GroupBuyID:    "gb1",
		UserID:        "u1",
		TotalAmount:   200,
		PaymentStatus: groupbuy.PaymentStatusUnset,
		Items: []*groupbuy.OrderItem{
			{ID: "oi-1", OrderID: "order-1", ProductID: "gb1-prod", SpecID: "gb1-spec", Quantity: 2, Status: groupbuy.OrderItemStatusUnordered, ProductName: "Product in gb1", SpecName: "Default", Price: 100},
		},
	}
	require.NoError(t, repo.CreateOrder(ctx, order))

	// --- GetOrder ---
	got, err := repo.GetOrder(ctx, "order-1")
	require.NoError(t, err)
	assert.Equal(t, "order-1", got.ID)
	assert.Equal(t, "gb1", got.GroupBuyID)
	assert.Equal(t, "u1", got.UserID)
	assert.Equal(t, int64(200), got.TotalAmount)
	assert.Len(t, got.Items, 1)
	assert.Equal(t, "oi-1", got.Items[0].ID)
	assert.Equal(t, 2, got.Items[0].Quantity)

	// --- ListOrders by groupBuyID ---
	orders, err := repo.ListOrders(ctx, "gb1", "")
	require.NoError(t, err)
	assert.Len(t, orders, 1)
	assert.Equal(t, "order-1", orders[0].ID)

	// --- Create a second order for user u2 ---
	order2 := &groupbuy.Order{
		ID:            "order-2",
		GroupBuyID:    "gb1",
		UserID:        "u2",
		TotalAmount:   100,
		PaymentStatus: groupbuy.PaymentStatusUnset,
		Items: []*groupbuy.OrderItem{
			{ID: "oi-2", OrderID: "order-2", ProductID: "gb1-prod", SpecID: "gb1-spec", Quantity: 1, Status: groupbuy.OrderItemStatusUnordered, ProductName: "Product in gb1", SpecName: "Default", Price: 100},
		},
	}
	require.NoError(t, repo.CreateOrder(ctx, order2))

	// --- ListOrders by userID ---
	ordersU2, err := repo.ListOrders(ctx, "", "u2")
	require.NoError(t, err)
	assert.Len(t, ordersU2, 1)
	assert.Equal(t, "order-2", ordersU2[0].ID)

	// --- ListOrders both filters ---
	ordersAll, err := repo.ListOrders(ctx, "gb1", "")
	require.NoError(t, err)
	assert.Len(t, ordersAll, 2)

	// --- UpdateOrder (change items) ---
	got.Items = []*groupbuy.OrderItem{
		{ID: "oi-1-new", OrderID: "order-1", ProductID: "gb1-prod", SpecID: "gb1-spec", Quantity: 5, Status: groupbuy.OrderItemStatusOrdered, ProductName: "Product in gb1", SpecName: "Default", Price: 100},
	}
	got.TotalAmount = 500
	require.NoError(t, repo.UpdateOrder(ctx, got))

	updated, err := repo.GetOrder(ctx, "order-1")
	require.NoError(t, err)
	assert.Equal(t, int64(500), updated.TotalAmount)
	assert.Len(t, updated.Items, 1)
	assert.Equal(t, "oi-1-new", updated.Items[0].ID)
	assert.Equal(t, 5, updated.Items[0].Quantity)

	// --- UpdateOrderPaymentStatus ---
	require.NoError(t, repo.UpdateOrderPaymentStatus(ctx, "order-1", groupbuy.PaymentStatusConfirmed))

	afterPay, err := repo.GetOrder(ctx, "order-1")
	require.NoError(t, err)
	assert.Equal(t, groupbuy.PaymentStatusConfirmed, afterPay.PaymentStatus)
}

// ---------------------------------------------------------------------------
// Test: Category CRUD
// ---------------------------------------------------------------------------

func TestGroupBuyRepository_CategoryCRUD(t *testing.T) {
	db := newGroupBuyTestDB(t)
	repo := NewGroupBuyRepository(db)
	ctx := context.Background()

	// --- Empty list ---
	cats, err := repo.ListCategories(ctx)
	require.NoError(t, err)
	assert.Empty(t, cats)

	// --- CreateCategory ---
	cat1 := &groupbuy.Category{ID: "cat-1", Name: "Snacks"}
	require.NoError(t, repo.CreateCategory(ctx, cat1))

	cat2 := &groupbuy.Category{ID: "cat-2", Name: "Drinks"}
	require.NoError(t, repo.CreateCategory(ctx, cat2))

	// --- ListCategories ---
	cats, err = repo.ListCategories(ctx)
	require.NoError(t, err)
	assert.Len(t, cats, 2)

	names := []string{cats[0].Name, cats[1].Name}
	assert.Contains(t, names, "Snacks")
	assert.Contains(t, names, "Drinks")
}

// ---------------------------------------------------------------------------
// Test: PriceTemplate CRUD
// ---------------------------------------------------------------------------

func TestGroupBuyRepository_PriceTemplateCRUD(t *testing.T) {
	db := newGroupBuyTestDB(t)
	repo := NewGroupBuyRepository(db)
	ctx := context.Background()

	pt := &groupbuy.PriceTemplate{
		ID:             "pt-1",
		Name:           "JPY Template",
		SourceCurrency: "JPY",
		ExchangeRate:   0.22,
		Rounding:       &groupbuy.RoundingConfig{Method: groupbuy.RoundingMethodRound, Digit: 0},
	}

	// --- Create ---
	require.NoError(t, repo.CreatePriceTemplate(ctx, pt))

	// --- Get ---
	got, err := repo.GetPriceTemplate(ctx, "pt-1")
	require.NoError(t, err)
	assert.Equal(t, "JPY Template", got.Name)
	assert.Equal(t, "JPY", got.SourceCurrency)
	assert.InDelta(t, 0.22, got.ExchangeRate, 0.001)
	assert.Equal(t, groupbuy.RoundingMethodRound, got.Rounding.Method)

	// --- List ---
	pts, err := repo.ListPriceTemplates(ctx)
	require.NoError(t, err)
	assert.Len(t, pts, 1)

	// --- Update ---
	pt.Name = "Updated JPY"
	pt.ExchangeRate = 0.23
	require.NoError(t, repo.UpdatePriceTemplate(ctx, pt))

	got2, err := repo.GetPriceTemplate(ctx, "pt-1")
	require.NoError(t, err)
	assert.Equal(t, "Updated JPY", got2.Name)
	assert.InDelta(t, 0.23, got2.ExchangeRate, 0.001)

	// --- Delete ---
	require.NoError(t, repo.DeletePriceTemplate(ctx, "pt-1"))

	// --- Get after delete -> error ---
	_, err = repo.GetPriceTemplate(ctx, "pt-1")
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// Test: AddProduct / DeleteProduct
// ---------------------------------------------------------------------------

func TestGroupBuyRepository_ProductManagement(t *testing.T) {
	db := newGroupBuyTestDB(t)
	repo := NewGroupBuyRepository(db)
	ctx := context.Background()

	seedUser(t, db, "u1", "User 1", "u1@test.com")
	seedGroupBuy(t, repo, ctx, "gb1", "u1")

	// --- AddProduct ---
	newProd := &groupbuy.Product{
		ID:            "new-prod",
		GroupBuyID:    "gb1",
		Name:          "New Product",
		PriceOriginal: 250,
		Specs: []*groupbuy.ProductSpec{
			{ID: "new-spec", ProductID: "new-prod", Name: "Large"},
		},
	}
	require.NoError(t, repo.AddProduct(ctx, newProd))

	// Verify the product was added by fetching the group buy.
	fetched, err := repo.GetByID(ctx, "gb1")
	require.NoError(t, err)
	assert.Len(t, fetched.Products, 2)

	// --- DeleteProduct (no orders) -> success ---
	require.NoError(t, repo.DeleteProduct(ctx, "gb1", "new-prod"))

	fetched2, err := repo.GetByID(ctx, "gb1")
	require.NoError(t, err)
	assert.Len(t, fetched2.Products, 1)

	// --- DeleteProduct when orders exist -> should error ---
	order := &groupbuy.Order{
		ID:         "ord-block",
		GroupBuyID: "gb1",
		UserID:     "u1",
		Items: []*groupbuy.OrderItem{
			{ID: "oi-block", OrderID: "ord-block", ProductID: "gb1-prod", SpecID: "gb1-spec", Quantity: 1, Price: 100},
		},
	}
	require.NoError(t, repo.CreateOrder(ctx, order))

	err = repo.DeleteProduct(ctx, "gb1", "gb1-prod")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot delete product: existing orders found")
}

// ---------------------------------------------------------------------------
// Test: Pagination
// ---------------------------------------------------------------------------

func TestGroupBuyRepository_Pagination(t *testing.T) {
	db := newGroupBuyTestDB(t)
	repo := NewGroupBuyRepository(db)
	ctx := context.Background()

	seedUser(t, db, "u1", "User 1", "u1@test.com")

	// Create 5 active groupbuys.
	for i := 1; i <= 5; i++ {
		id := fmt.Sprintf("pg-gb-%d", i)
		gb := &groupbuy.GroupBuy{
			ID:        id,
			Title:     fmt.Sprintf("GB %d", i),
			CreatorID: "u1",
			Status:    groupbuy.GroupBuyStatusActive,
		}
		require.NoError(t, repo.Create(ctx, gb))
	}

	// limit=2, offset=0 -> 2 results
	page1, err := repo.List(ctx, 2, 0, "", false, false)
	require.NoError(t, err)
	assert.Len(t, page1, 2)

	// limit=2, offset=2 -> 2 results
	page2, err := repo.List(ctx, 2, 2, "", false, false)
	require.NoError(t, err)
	assert.Len(t, page2, 2)

	// limit=10, offset=10 -> 0 results
	pageEmpty, err := repo.List(ctx, 10, 10, "", false, false)
	require.NoError(t, err)
	assert.Empty(t, pageEmpty)
}

// ---------------------------------------------------------------------------
// Test: NotFound
// ---------------------------------------------------------------------------

func TestGroupBuyRepository_NotFound(t *testing.T) {
	db := newGroupBuyTestDB(t)
	repo := NewGroupBuyRepository(db)
	ctx := context.Background()

	// GetByID non-existent
	_, err := repo.GetByID(ctx, "does-not-exist")
	assert.ErrorIs(t, err, service.ErrNotFound)

	// GetOrder non-existent
	_, err = repo.GetOrder(ctx, "no-such-order")
	assert.ErrorIs(t, err, service.ErrNotFound)
}
