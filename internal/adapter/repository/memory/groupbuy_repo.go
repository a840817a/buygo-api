package memory

import (
	"context"
	"errors"
	"sort"
	"sync"

	"github.com/hatsubosi/buygo-api/internal/domain/groupbuy"
)

type GroupBuyRepository struct {
	mu             sync.RWMutex
	groupbuys      map[string]*groupbuy.GroupBuy
	orders         map[string]*groupbuy.Order
	categories     map[string]*groupbuy.Category
	priceTemplates map[string]*groupbuy.PriceTemplate
}

func NewGroupBuyRepository() *GroupBuyRepository {
	return &GroupBuyRepository{
		groupbuys:      make(map[string]*groupbuy.GroupBuy),
		orders:         make(map[string]*groupbuy.Order),
		categories:     make(map[string]*groupbuy.Category),
		priceTemplates: make(map[string]*groupbuy.PriceTemplate),
	}
}

// GroupBuy Methods
func (r *GroupBuyRepository) Create(ctx context.Context, gb *groupbuy.GroupBuy) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.groupbuys[gb.ID]; ok {
		return errors.New("group buy already exists")
	}
	r.groupbuys[gb.ID] = gb
	return nil
}

func (r *GroupBuyRepository) GetByID(ctx context.Context, id string) (*groupbuy.GroupBuy, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	gb, ok := r.groupbuys[id]
	if !ok {
		return nil, errors.New("group buy not found")
	}
	return gb, nil
}

func (r *GroupBuyRepository) List(ctx context.Context, limit, offset int, userID string, isSysAdmin bool, manageOnly bool) ([]*groupbuy.GroupBuy, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// 1. Collect all matching items
	var filtered []*groupbuy.GroupBuy
	for _, gb := range r.groupbuys {
		// Filtering Logic
		// 1. SysAdmin sees all
		if isSysAdmin {
			filtered = append(filtered, gb)
			continue
		}

		// 2. Manager Access (if userID provided)
		if userID != "" {
			isManager := gb.CreatorID == userID
			if !isManager {
				for _, mID := range gb.ManagerIDs {
					if mID == userID {
						isManager = true
						break
					}
				}
			}

			if manageOnly {
				if isManager {
					filtered = append(filtered, gb)
				}
				continue
			} else {
				if isManager {
					filtered = append(filtered, gb)
					continue
				}
			}
		}

		// 3. Public Access
		if gb.Status == groupbuy.GroupBuyStatusActive || gb.Status == groupbuy.GroupBuyStatusEnded {
			filtered = append(filtered, gb)
		}
	}

	// 2. Sort by CreatedAt DESC (Newest first)
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].CreatedAt.After(filtered[j].CreatedAt)
	})

	// 3. Apply Pagination
	if offset >= len(filtered) {
		return []*groupbuy.GroupBuy{}, nil
	}
	end := offset + limit
	if end > len(filtered) {
		end = len(filtered)
	}
	return filtered[offset:end], nil
}

func (r *GroupBuyRepository) Update(ctx context.Context, gb *groupbuy.GroupBuy) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.groupbuys[gb.ID]; !ok {
		return errors.New("group buy not found")
	}
	r.groupbuys[gb.ID] = gb
	return nil
}

// Product Methods
func (r *GroupBuyRepository) AddProduct(ctx context.Context, product *groupbuy.Product) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	gb, ok := r.groupbuys[product.GroupBuyID]
	if !ok {
		return errors.New("group buy not found")
	}

	gb.Products = append(gb.Products, product)
	return nil
}

// Order Methods
func (r *GroupBuyRepository) CreateOrder(ctx context.Context, order *groupbuy.Order) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check Stock Limit
	for _, item := range order.Items {
		gb, ok := r.groupbuys[order.GroupBuyID]
		if !ok {
			// Should not happen if service checked, but for safety
			continue
		}

		var maxQty int32 = 0
		// Find product max qty
		for _, prod := range gb.Products {
			if prod.ID == item.ProductID {
				maxQty = prod.MaxQuantity
				break
			}
		}

		if maxQty > 0 {
			// Calculate sold amount
			var sold int64 = 0
			for _, o := range r.orders {
				// Only count valid orders (not cancelled?)
				// For now assume all orders in repo count towards stock
				for _, i := range o.Items {
					if i.ProductID == item.ProductID {
						sold += int64(i.Quantity)
					}
				}
			}

			if sold+int64(item.Quantity) > int64(maxQty) {
				return errors.New("product out of stock")
			}
		}
	}

	if _, ok := r.orders[order.ID]; ok {
		return errors.New("order already exists")
	}
	r.orders[order.ID] = order
	return nil
}

func (r *GroupBuyRepository) GetOrder(ctx context.Context, id string) (*groupbuy.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	o, ok := r.orders[id]
	if !ok {
		return nil, errors.New("order not found")
	}
	return o, nil
}

func (r *GroupBuyRepository) ListOrders(ctx context.Context, groupBuyID string, userID string) ([]*groupbuy.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var res []*groupbuy.Order
	for _, o := range r.orders {
		if (groupBuyID == "" || o.GroupBuyID == groupBuyID) && (userID == "" || o.UserID == userID) {
			res = append(res, o)
		}
	}
	return res, nil
}

func (r *GroupBuyRepository) UpdateOrder(ctx context.Context, order *groupbuy.Order) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.orders[order.ID]; !ok {
		return errors.New("order not found")
	}
	r.orders[order.ID] = order
	return nil
}

func (r *GroupBuyRepository) UpdateOrderPaymentStatus(ctx context.Context, orderID string, status groupbuy.PaymentStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	o, ok := r.orders[orderID]
	if !ok {
		return errors.New("order not found")
	}
	o.PaymentStatus = status
	return nil
}

func (r *GroupBuyRepository) BatchUpdateOrderItemStatus(ctx context.Context, groupBuyID string, specID string, fromStatuses []int, toStatus int, limit int) (int64, []string, error) {
	// Not implemented for memory repo in this practice scope
	return 0, nil, nil
}

// Category Methods (Stub for Memory Repo)
// Category Methods
func (r *GroupBuyRepository) CreateCategory(ctx context.Context, c *groupbuy.Category) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.categories[c.ID]; ok {
		return errors.New("category already exists")
	}
	r.categories[c.ID] = c
	return nil
}

func (r *GroupBuyRepository) ListCategories(ctx context.Context) ([]*groupbuy.Category, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var res []*groupbuy.Category
	for _, c := range r.categories {
		res = append(res, c)
	}

	// Sort by name for consistency
	sort.Slice(res, func(i, j int) bool {
		return res[i].Name < res[j].Name
	})

	return res, nil
}

// Price Template Stubs
func (r *GroupBuyRepository) CreatePriceTemplate(ctx context.Context, pt *groupbuy.PriceTemplate) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.priceTemplates[pt.ID]; ok {
		return errors.New("template already exists")
	}
	r.priceTemplates[pt.ID] = pt
	return nil
}

func (r *GroupBuyRepository) ListPriceTemplates(ctx context.Context) ([]*groupbuy.PriceTemplate, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var res []*groupbuy.PriceTemplate
	for _, pt := range r.priceTemplates {
		res = append(res, pt)
	}
	return res, nil
}

func (r *GroupBuyRepository) GetPriceTemplate(ctx context.Context, id string) (*groupbuy.PriceTemplate, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	pt, ok := r.priceTemplates[id]
	if !ok {
		return nil, errors.New("template not found")
	}
	return pt, nil
}

func (r *GroupBuyRepository) UpdatePriceTemplate(ctx context.Context, pt *groupbuy.PriceTemplate) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.priceTemplates[pt.ID]; !ok {
		return errors.New("template not found")
	}
	r.priceTemplates[pt.ID] = pt
	return nil
}

// Product Methods
func (r *GroupBuyRepository) DeleteProduct(ctx context.Context, groupBuyID, productID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 1. Check existing orders
	for _, o := range r.orders {
		for _, item := range o.Items {
			if item.ProductID == productID {
				return errors.New("cannot delete product: existing orders found")
			}
		}
	}

	// 2. Remove from GroupBuy
	gb, ok := r.groupbuys[groupBuyID]
	if !ok {
		return errors.New("group buy not found")
	}

	newProducts := make([]*groupbuy.Product, 0, len(gb.Products))
	found := false
	for _, prod := range gb.Products {
		if prod.ID == productID {
			found = true
			continue
		}
		newProducts = append(newProducts, prod)
	}

	if !found {
		return errors.New("product not found")
	}

	gb.Products = newProducts
	return nil
}

func (r *GroupBuyRepository) DeletePriceTemplate(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.priceTemplates[id]; !ok {
		return errors.New("template not found")
	}
	delete(r.priceTemplates, id)
	return nil
}
