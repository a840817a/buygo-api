package memory

import (
	"context"
	"errors"
	"sort"
	"sync"

	"github.com/hatsubosi/buygo-api/internal/domain/event"
)

type EventRepository struct {
	mu            sync.RWMutex
	events        map[string]*event.Event
	registrations map[string]*event.Registration
}

func NewEventRepository() *EventRepository {
	return &EventRepository{
		events:        make(map[string]*event.Event),
		registrations: make(map[string]*event.Registration),
	}
}

// Event Methods
func (r *EventRepository) Create(ctx context.Context, e *event.Event) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.events[e.ID]; ok {
		return errors.New("event already exists")
	}
	r.events[e.ID] = e
	return nil
}

func (r *EventRepository) GetByID(ctx context.Context, id string) (*event.Event, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	e, ok := r.events[id]
	if !ok {
		return nil, errors.New("event not found")
	}
	return e, nil
}

func (r *EventRepository) List(ctx context.Context, limit, offset int, userID string, isSysAdmin bool, manageOnly bool) ([]*event.Event, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// 1. Collect all matching items
	var filtered []*event.Event
	for _, e := range r.events {
		// Filtering Logic
		if isSysAdmin {
			filtered = append(filtered, e)
			continue
		}

		if userID != "" {
			isManager := e.CreatorID == userID
			if !isManager {
				for _, mID := range e.ManagerIDs {
					if mID == userID {
						isManager = true
						break
					}
				}
			}

			if manageOnly {
				// Strict Manager View
				if isManager {
					filtered = append(filtered, e)
				}
				continue
			} else {
				// Public + Manager View
				if isManager {
					filtered = append(filtered, e)
					continue
				}
				// Fallthrough to check status if not manager
			}
		}

		// Public
		if e.Status == event.EventStatusActive || e.Status == event.EventStatusEnded {
			filtered = append(filtered, e)
		}
	}

	// 2. Sort by CreatedAt DESC (Newest first)
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].CreatedAt.After(filtered[j].CreatedAt)
	})

	// 3. Apply Pagination
	if offset >= len(filtered) {
		return []*event.Event{}, nil
	}
	end := offset + limit
	if end > len(filtered) {
		end = len(filtered)
	}
	return filtered[offset:end], nil
}

func (r *EventRepository) Update(ctx context.Context, e *event.Event) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.events[e.ID]; !ok {
		return errors.New("event not found")
	}
	r.events[e.ID] = e
	return nil
}

// Registration Methods
func (r *EventRepository) Register(ctx context.Context, reg *event.Registration) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Capacity Check
	e, ok := r.events[reg.EventID]
	if !ok {
		// Should not happen if service checked, but good for safety
		return errors.New("event not found")
	}

	for _, newItem := range reg.SelectedItems {
		var maxQty int32 = 0
		// Find limit
		for _, ei := range e.Items {
			if ei.ID == newItem.EventItemID {
				maxQty = ei.MaxParticipants
				break
			}
		}

		if maxQty > 0 {
			var sold int64 = 0
			for _, existReg := range r.registrations {
				if existReg.Status == event.RegistrationStatusCancelled {
					continue
				}
				for _, existItem := range existReg.SelectedItems {
					if existItem.EventItemID == newItem.EventItemID {
						sold += int64(existItem.Quantity)
					}
				}
			}

			if sold+int64(newItem.Quantity) > int64(maxQty) {
				return errors.New("registration limit exceeded")
			}
		}
	}

	r.registrations[reg.ID] = reg
	return nil
}

func (r *EventRepository) GetRegistration(ctx context.Context, id string) (*event.Registration, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	reg, ok := r.registrations[id]
	if !ok {
		return nil, errors.New("registration not found")
	}
	return reg, nil
}

func (r *EventRepository) ListRegistrations(ctx context.Context, eventID string, userID string) ([]*event.Registration, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var res []*event.Registration
	for _, reg := range r.registrations {
		if (eventID == "" || reg.EventID == eventID) && (userID == "" || reg.UserID == userID) {
			res = append(res, reg)
		}
	}
	return res, nil
}

func (r *EventRepository) UpdateRegistration(ctx context.Context, reg *event.Registration) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.registrations[reg.ID]; !ok {
		return errors.New("registration not found")
	}
	r.registrations[reg.ID] = reg
	return nil
}
