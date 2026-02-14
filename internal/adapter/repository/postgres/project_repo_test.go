package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/buygo/buygo-api/internal/adapter/repository/postgres/model"
	"github.com/buygo/buygo-api/internal/domain/project"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	assert.NoError(t, err)

	err = db.AutoMigrate(
		&model.User{},
		&model.Project{},
		&model.Product{},
		&model.ProductSpec{},
		&model.Order{},
		&model.OrderItem{},
	)
	assert.NoError(t, err)

	return db
}

func TestProjectRepository_UpdateShippingConfigs(t *testing.T) {
	db := newTestDB(t)
	repo := NewProjectRepository(db)
	ctx := context.Background()

	// 1. Create Initial Project
	creatorID := "user-1"
	db.Create(&model.User{ID: creatorID, Name: "Creator"})

	proj := &project.Project{
		ID:          "proj-1",
		Title:       "Test Project",
		Description: "Desc",
		Status:      project.ProjectStatusActive,
		CreatorID:   creatorID,
		CreatedAt:   time.Now(),
		ShippingConfigs: []*project.ShippingConfig{
			{ID: "sc-1", Name: "Initial Meetup", Type: project.ShippingTypeMeetup, Price: 0},
		},
	}
	err := repo.Create(ctx, proj)
	assert.NoError(t, err)

	// Verify Initial State
	saved, err := repo.GetByID(ctx, "proj-1")
	assert.NoError(t, err)
	assert.Len(t, saved.ShippingConfigs, 1)
	assert.Equal(t, project.ShippingTypeMeetup, saved.ShippingConfigs[0].Type)

	// 2. Update with New Shipping Config (Simulate Type Change / Add)
	// Changing "Initial Meetup" to "Delivery" type (Type 1), adding "Store Pickup" (Type 2)
	updatedConfigs := []*project.ShippingConfig{
		{ID: "sc-1", Name: "Changed to Delivery", Type: project.ShippingTypeDelivery, Price: 100},
		{ID: "sc-2", Name: "New Store Pickup", Type: project.ShippingTypeStorePickup, Price: 60},
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
	var delivery *project.ShippingConfig
	var pickup *project.ShippingConfig

	for _, sc := range final.ShippingConfigs {
		if sc.ID == "sc-1" {
			delivery = sc
		} else if sc.ID == "sc-2" {
			pickup = sc
		}
	}

	assert.NotNil(t, delivery)
	assert.Equal(t, project.ShippingTypeDelivery, delivery.Type, "Type should be Delivery (1)")
	assert.Equal(t, int64(100), delivery.Price)

	assert.NotNil(t, pickup)
	assert.Equal(t, project.ShippingTypeStorePickup, pickup.Type, "Type should be Store Pickup (2)")

	// 4. Test Meetup Again
	proj.ShippingConfigs = []*project.ShippingConfig{
		{ID: "sc-3", Name: "Meetup Only", Type: project.ShippingTypeMeetup, Price: 0},
	}
	err = repo.Update(ctx, proj)
	assert.NoError(t, err)

	final2, err := repo.GetByID(ctx, "proj-1")
	assert.NoError(t, err)
	assert.Len(t, final2.ShippingConfigs, 1)
	assert.Equal(t, project.ShippingTypeMeetup, final2.ShippingConfigs[0].Type, "Type should be Meetup (3)")
}
