package postgres

import (
	"context"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/buygo/buygo-api/internal/adapter/repository/postgres/model"
	"github.com/buygo/buygo-api/internal/domain/groupbuy"
	"github.com/google/uuid"
)

type ProjectRepository struct {
	db *gorm.DB
}

func NewProjectRepository(db *gorm.DB) *ProjectRepository {
	return &ProjectRepository{db: db}
}

// Project Core
func (r *ProjectRepository) Create(ctx context.Context, p *project.Project) error {
	return r.CreateWithTx(ctx, p)
}

func (r *ProjectRepository) CreateWithTx(ctx context.Context, p *project.Project) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		m := model.FromDomainProject(p)

		// 1. Create Project (and Products by cascade, excluding Users)
		if err := tx.Omit("Creator", "Managers").Create(m).Error; err != nil {
			return err
		}

		// 2. Associate Managers
		if len(m.Managers) > 0 {
			if err := tx.Model(m).Association("Managers").Replace(m.Managers); err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *ProjectRepository) GetByID(ctx context.Context, id string) (*project.Project, error) {
	var m model.Project
	if err := r.db.WithContext(ctx).
		Preload("Creator").
		Preload("Managers").
		Preload("Products.Specs").
		First(&m, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("project not found")
		}
		return nil, err
	}
	return m.ToDomain(), nil
}

func (r *ProjectRepository) List(ctx context.Context, limit, offset int, userID string, isSysAdmin bool, manageOnly bool) ([]*project.Project, error) {
	var models []*model.Project

	query := r.db.WithContext(ctx).
		Limit(limit).Offset(offset).
		Preload("Creator")

	// Filter Logic:
	// If SysAdmin: Show ALL
	// Else If userID provided:
	//   If manageOnly: Show (Creator=userID OR Manager=userID)
	//   Else: Show (Active/Ended) OR (Creator=userID OR Manager=userID)
	// Else: Show (Active/Ended) ONLY

	if isSysAdmin {
		// No filter
	} else if userID != "" {
		if manageOnly {
			// Strict Manager View
			query = query.Where(
				r.db.Where("creator_id = ?", userID).
					Or("id IN (?)", r.db.Table("project_managers").Select("group_buy_id").Where("user_id = ?", userID)),
			)
		} else {
			// Public + My Items View
			query = query.Where(
				r.db.Where("status IN ?", []int{2, 3}).
					Or("creator_id = ?", userID).
					Or("id IN (?)", r.db.Table("project_managers").Select("group_buy_id").Where("user_id = ?", userID)),
			)
		}
	} else {
		// Public Only
		query = query.Where("status IN ?", []int{2, 3})
	}

	if err := query.Find(&models).Error; err != nil {
		return nil, err
	}

	var res []*project.Project
	for _, m := range models {
		res = append(res, m.ToDomain())
	}
	return res, nil
}

func (r *ProjectRepository) Update(ctx context.Context, p *project.Project) error {
	m := model.FromDomainProject(p)

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Update basic fields
		// Use Select + Updates(struct) to ensure GORM serializers (JSON) run and zero values are updated.
		if err := tx.Model(&model.Project{ID: m.ID}).
			Select("Title", "Description", "Status", "CoverImage", "Deadline", "PaymentMethods", "ShippingConfigs", "ExchangeRate", "RoundingMethod", "RoundingDigit", "SourceCurrency").
			Updates(m).Error; err != nil {
			return err
		}

		// 2. Replace Products
		if len(m.Products) > 0 {
			// Using Association.Replace works for relationships, but for updating fields on the child models
			// (like ExchangeRate on an existing Product), it might depend on GORM configuration (FullSaveAssociations).
			// To be safe and explicit:
			// 1. Replace the association (handles FKs and insertions/deletions from the set)
			if err := tx.Model(&model.Project{ID: m.ID}).Association("Products").Replace(m.Products); err != nil {
				return err
			}

			// 2. Explicitly Save products to ensure all fields are updated (Upsert)
			// This addresses the issue where "Replace" might just manage the relationship and not sending UPDATE for changed fields of existing records.
			// Batched Upsert constraint on ID.
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "id"}},
				UpdateAll: true,
			}).Create(m.Products).Error; err != nil {
				return err
			}
		} else {
			// ... (clear logic)
			if err := tx.Model(&model.Project{ID: m.ID}).Association("Products").Clear(); err != nil {
				return err
			}
		}

		return nil
	})
}

// Product Methods
func (r *ProjectRepository) AddProduct(ctx context.Context, product *project.Product) error {
	m := model.FromDomainProduct(product)
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *ProjectRepository) DeleteProduct(ctx context.Context, projectID, productID string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Check existing orders
		var count int64
		if err := tx.Model(&model.OrderItem{}).
			Joins("JOIN orders ON orders.id = order_items.order_id").
			Where("orders.group_buy_id = ? AND order_items.product_id = ?", projectID, productID).
			Count(&count).Error; err != nil {
			return err
		}

		if count > 0 {
			return errors.New("cannot delete product: existing orders found")
		}

		// 2. Delete product (cascade will handle specs)
		return tx.Where("id = ? AND group_buy_id = ?", productID, projectID).Delete(&model.Product{}).Error
	})
}

// Order Methods
func (r *ProjectRepository) CreateOrder(ctx context.Context, order *project.Order) error {
	m := model.FromDomainOrder(order)
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *ProjectRepository) GetOrder(ctx context.Context, id string) (*project.Order, error) {
	var m model.Order
	if err := r.db.WithContext(ctx).Preload("Items").First(&m, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("order not found")
		}
		return nil, err
	}
	return m.ToDomain(), nil
}

func (r *ProjectRepository) ListOrders(ctx context.Context, projectID string, userID string) ([]*project.Order, error) {
	var models []*model.Order
	query := r.db.WithContext(ctx).Preload("Items")

	if projectID != "" {
		query = query.Where("group_buy_id = ?", projectID)
	}
	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}

	if err := query.Find(&models).Error; err != nil {
		return nil, err
	}

	var res []*project.Order
	for _, m := range models {
		res = append(res, m.ToDomain())
	}
	return res, nil
}

func (r *ProjectRepository) UpdateOrder(ctx context.Context, order *project.Order) error {
	m := model.FromDomainOrder(order)
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(m).Error; err != nil {
			return err
		}
		// Replace Items
		return tx.Model(m).Association("Items").Replace(m.Items)
	})
}

func (r *ProjectRepository) UpdateOrderPaymentStatus(ctx context.Context, orderID string, status int) error {
	return r.db.WithContext(ctx).Model(&model.Order{}).Where("id = ?", orderID).Update("payment_status", status).Error
}

func (r *ProjectRepository) BatchUpdateOrderItemStatus(ctx context.Context, projectID string, specID string, fromStatuses []int, toStatus int, limit int) (int64, []string, error) {
	var items []model.OrderItem
	var uniqueOrderIDs []string
	var movedCount int64

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Find candidates (FIFO: Join Orders order by CreatedAt)
		// We use Limit(limit) as an optimization, assuming worst case 1 qty per row.
		// If rows have larger quantity, we might fetch 1 row for a large limit, which is fine.
		if err := tx.Model(&model.OrderItem{}).
			Joins("JOIN orders ON orders.id = order_items.order_id").
			Where("orders.group_buy_id = ? AND order_items.spec_id = ? AND order_items.status IN ?", projectID, specID, fromStatuses).
			Order("orders.created_at ASC").
			Limit(limit).
			Find(&items).Error; err != nil {
			return err
		}

		if len(items) == 0 {
			return nil
		}

		// 2. Iterate and Allocate
		var idsToFullUpdate []string
		var itemsToUpdateQty []model.OrderItem
		var itemsToCreate []model.OrderItem

		remaining := limit
		affectedOrderIDs := make(map[string]bool)

		for _, item := range items {
			if remaining <= 0 {
				break
			}

			affectedOrderIDs[item.OrderID] = true

			if item.Quantity <= remaining {
				// FULL MOVE
				idsToFullUpdate = append(idsToFullUpdate, item.ID)
				remaining -= item.Quantity
			} else {
				// PARTIAL MOVE (SPLIT)
				moveQty := remaining

				// A. Update Original (Reduce Qty, keep old status)
				item.Quantity -= moveQty
				itemsToUpdateQty = append(itemsToUpdateQty, item)

				// B. Create New (The moved part, new status)
				newItem := item // Copy struct
				newItem.ID = uuid.New().String()
				newItem.Quantity = moveQty
				newItem.Status = toStatus

				itemsToCreate = append(itemsToCreate, newItem)

				remaining = 0
			}
		}

		// 3. Execute Updates

		// A. Full Updates
		if len(idsToFullUpdate) > 0 {
			if err := tx.Model(&model.OrderItem{}).
				Where("id IN ?", idsToFullUpdate).
				Update("status", toStatus).Error; err != nil {
				return err
			}
		}

		// B. Update Quantities (for split items)
		for _, item := range itemsToUpdateQty {
			if err := tx.Model(&model.OrderItem{}).Where("id = ?", item.ID).Update("quantity", item.Quantity).Error; err != nil {
				return err
			}
		}

		// C. Create New Items (for split moved parts)
		if len(itemsToCreate) > 0 {
			if err := tx.Create(&itemsToCreate).Error; err != nil {
				return err
			}
		}

		// D. Capture Output
		for oid := range affectedOrderIDs {
			uniqueOrderIDs = append(uniqueOrderIDs, oid)
		}
		movedCount = int64(limit - remaining)

		return nil
	})

	if err != nil {
		return 0, nil, err
	}

	return movedCount, uniqueOrderIDs, nil
}

// Category Methods

func (r *ProjectRepository) CreateCategory(ctx context.Context, c *project.Category) error {
	m := model.FromDomainCategory(c)
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *ProjectRepository) ListCategories(ctx context.Context) ([]*project.Category, error) {
	var models []*model.Category
	if err := r.db.WithContext(ctx).Find(&models).Error; err != nil {
		return nil, err
	}

	var res []*project.Category
	for _, m := range models {
		res = append(res, m.ToDomain())
	}
	return res, nil
}
