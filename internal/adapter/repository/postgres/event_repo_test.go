package postgres

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/hatsubosi/buygo-api/internal/adapter/repository/postgres/model"
	"github.com/hatsubosi/buygo-api/internal/domain/event"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newEventTestDB(t *testing.T) *gorm.DB {
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)

	require.NoError(t, db.AutoMigrate(
		&model.User{},
		&model.Event{},
		&model.EventItem{},
		&model.DiscountRule{},
		&model.Registration{},
		&model.RegistrationItem{},
	))
	return db
}

func TestEventRepository_UpdateManagers(t *testing.T) {
	db := newEventTestDB(t)
	repo := NewEventRepository(db)
	ctx := context.Background()

	creatorID := "event-creator"
	managerA := "event-manager-a"
	managerB := "event-manager-b"

	require.NoError(t, db.Create(&model.User{ID: creatorID, Name: "Creator", Email: "creator@test.com"}).Error)
	require.NoError(t, db.Create(&model.User{ID: managerA, Name: "Manager A", Email: "manager-a@test.com"}).Error)
	require.NoError(t, db.Create(&model.User{ID: managerB, Name: "Manager B", Email: "manager-b@test.com"}).Error)

	e := &event.Event{
		ID:             "event-update-managers",
		Title:          "Event Manager Update",
		Description:    "desc",
		Status:         event.EventStatusActive,
		StartTime:      time.Now(),
		EndTime:        time.Now().Add(2 * time.Hour),
		CreatorID:      creatorID,
		ManagerIDs:     []string{managerA},
		PaymentMethods: []string{"Cash"},
	}
	require.NoError(t, repo.Create(ctx, e))

	e.ManagerIDs = []string{managerB}
	require.NoError(t, repo.Update(ctx, e))

	saved, err := repo.GetByID(ctx, e.ID)
	require.NoError(t, err)
	require.Len(t, saved.ManagerIDs, 1)
	assert.Equal(t, managerB, saved.ManagerIDs[0])
	require.Len(t, saved.Managers, 1)
	assert.Equal(t, managerB, saved.Managers[0].ID)
}
