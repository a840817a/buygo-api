package memory

import (
	"context"
	"errors"
	"sort"
	"sync"

	"github.com/buygo/buygo-api/internal/domain/project"
)

type ProjectRepository struct {
	mu         sync.RWMutex
	projects   map[string]*project.Project
	orders     map[string]*project.Order
	categories map[string]*project.Category
}

func NewProjectRepository() *ProjectRepository {
	return &ProjectRepository{
		projects:   make(map[string]*project.Project),
		orders:     make(map[string]*project.Order),
		categories: make(map[string]*project.Category),
	}
}

// Project Methods
func (r *ProjectRepository) Create(ctx context.Context, p *project.Project) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.projects[p.ID]; ok {
		return errors.New("project already exists")
	}
	r.projects[p.ID] = p
	return nil
}

func (r *ProjectRepository) GetByID(ctx context.Context, id string) (*project.Project, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	p, ok := r.projects[id]
	if !ok {
		return nil, errors.New("project not found")
	}
	return p, nil
}

func (r *ProjectRepository) List(ctx context.Context, limit, offset int, userID string, isSysAdmin bool, manageOnly bool) ([]*project.Project, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// 1. Collect all matching items
	var filtered []*project.Project
	for _, p := range r.projects {
		// Filtering Logic
		// 1. SysAdmin sees all
		if isSysAdmin {
			filtered = append(filtered, p)
			continue
		}

		// 2. Manager Access (if userID provided)
		if userID != "" {
			isManager := p.CreatorID == userID
			if !isManager {
				for _, mID := range p.ManagerIDs {
					if mID == userID {
						isManager = true
						break
					}
				}
			}

			if manageOnly {
				if isManager {
					filtered = append(filtered, p)
				}
				continue
			} else {
				if isManager {
					filtered = append(filtered, p)
					continue
				}
			}
		}

		// 3. Public Access
		if p.Status == project.ProjectStatusActive || p.Status == project.ProjectStatusEnded {
			filtered = append(filtered, p)
		}
	}

	// 2. Sort by CreatedAt DESC (Newest first)
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].CreatedAt.After(filtered[j].CreatedAt)
	})

	// 3. Apply Pagination
	if offset >= len(filtered) {
		return []*project.Project{}, nil
	}
	end := offset + limit
	if end > len(filtered) {
		end = len(filtered)
	}
	return filtered[offset:end], nil
}

func (r *ProjectRepository) Update(ctx context.Context, p *project.Project) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.projects[p.ID]; !ok {
		return errors.New("project not found")
	}
	r.projects[p.ID] = p
	return nil
}

// Product Methods
func (r *ProjectRepository) AddProduct(ctx context.Context, product *project.Product) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	p, ok := r.projects[product.ProjectID]
	if !ok {
		return errors.New("project not found")
	}

	p.Products = append(p.Products, product)
	return nil
}

// Order Methods
func (r *ProjectRepository) CreateOrder(ctx context.Context, order *project.Order) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check Stock Limit
	for _, item := range order.Items {
		p, ok := r.projects[order.ProjectID]
		if !ok {
			// Should not happen if service checked, but for safety
			continue
		}
		
		var maxQty int32 = 0
		// Find product max qty
		for _, prod := range p.Products {
			if prod.ID == item.ProductID {
				maxQty = prod.MaxQuantity
				break
			}
		}

		if maxQty > 0 {
			// Calculate sold amount
			var sold int32 = 0
			for _, o := range r.orders {
				// Only count valid orders (not cancelled?) 
				// For now assume all orders in repo count towards stock
				for _, i := range o.Items {
					if i.ProductID == item.ProductID {
						sold += int32(i.Quantity)
					}
				}
			}
			
			if sold + int32(item.Quantity) > maxQty {
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

func (r *ProjectRepository) GetOrder(ctx context.Context, id string) (*project.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	o, ok := r.orders[id]
	if !ok {
		return nil, errors.New("order not found")
	}
	return o, nil
}

func (r *ProjectRepository) ListOrders(ctx context.Context, projectID string, userID string) ([]*project.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var res []*project.Order
	for _, o := range r.orders {
		if (projectID == "" || o.ProjectID == projectID) && (userID == "" || o.UserID == userID) {
			res = append(res, o)
		}
	}
	return res, nil
}

func (r *ProjectRepository) UpdateOrder(ctx context.Context, order *project.Order) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.orders[order.ID]; !ok {
		return errors.New("order not found")
	}
	r.orders[order.ID] = order
	return nil
}

func (r *ProjectRepository) UpdateOrderPaymentStatus(ctx context.Context, orderID string, status int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	o, ok := r.orders[orderID]
	if !ok {
		return errors.New("order not found")
	}
	o.PaymentStatus = status
	return nil
}

func (r *ProjectRepository) BatchUpdateOrderItemStatus(ctx context.Context, projectID string, specID string, fromStatuses []int, toStatus int, limit int) (int64, []string, error) {
	// Not implemented for memory repo in this practice scope
	return 0, nil, nil
}

// Category Methods (Stub for Memory Repo)
// Category Methods
func (r *ProjectRepository) CreateCategory(ctx context.Context, c *project.Category) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.categories[c.ID]; ok {
		return errors.New("category already exists")
	}
	r.categories[c.ID] = c
	return nil
}

func (r *ProjectRepository) ListCategories(ctx context.Context) ([]*project.Category, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var res []*project.Category
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
func (r *ProjectRepository) CreatePriceTemplate(ctx context.Context, pt *project.PriceTemplate) error {
	return nil
}

func (r *ProjectRepository) ListPriceTemplates(ctx context.Context) ([]*project.PriceTemplate, error) {
	return nil, nil
}

func (r *ProjectRepository) GetPriceTemplate(ctx context.Context, id string) (*project.PriceTemplate, error) {
	return nil, nil
}

func (r *ProjectRepository) UpdatePriceTemplate(ctx context.Context, pt *project.PriceTemplate) error {
	return nil
}

// Product Methods
func (r *ProjectRepository) DeleteProduct(ctx context.Context, projectID, productID string) error {
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

	// 2. Remove from Project
	p, ok := r.projects[projectID]
	if !ok {
		return errors.New("project not found")
	}

	newProducts := make([]*project.Product, 0, len(p.Products))
	found := false
	for _, prod := range p.Products {
		if prod.ID == productID {
			found = true
			continue
		}
		newProducts = append(newProducts, prod)
	}

	if !found {
		return errors.New("product not found")
	}

	p.Products = newProducts
	return nil
}

func (r *ProjectRepository) DeletePriceTemplate(ctx context.Context, id string) error {
	return nil
}
