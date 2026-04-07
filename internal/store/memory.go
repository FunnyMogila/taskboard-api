package store

import (
	"errors"
	"slices"
	"strings"
	"sync"
	"time"

	"example.com/go-master-web-sample/internal/models"
)

var (
	ErrItemNotFound     = errors.New("item not found")
	ErrInvalidItemState = errors.New("status must be one of: draft, published, archived")
)

var allowedStatuses = map[string]struct{}{
	"draft":     {},
	"published": {},
	"archived":  {},
}

type MemoryStore struct {
	mu         sync.RWMutex
	items      map[int]models.Item
	nextItemID int
	users      map[string]models.User
}

func NewMemoryStore() *MemoryStore {
	now := time.Now().UTC()

	return &MemoryStore{
		items: map[int]models.Item{
			1: {
				ID:          1,
				Title:       "Learn chi routing",
				Description: "Walk through route groups, params, and middleware.",
				Status:      "published",
				Tags:        []string{"chi", "routing"},
				CreatedAt:   now,
				UpdatedAt:   now,
			},
			2: {
				ID:          2,
				Title:       "Protect endpoints with JWT",
				Description: "Issue and validate bearer tokens for private routes.",
				Status:      "draft",
				Tags:        []string{"auth", "jwt"},
				CreatedAt:   now,
				UpdatedAt:   now,
			},
		},
		nextItemID: 3,
		users: map[string]models.User{
			"alice": {ID: 1, Username: "alice", Password: "wonderland", Role: "admin"},
			"bob":   {ID: 2, Username: "bob", Password: "builder", Role: "user"},
		},
	}
}

func (s *MemoryStore) Authenticate(username, password string) (models.User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, ok := s.users[username]
	if !ok || user.Password != password {
		return models.User{}, false
	}

	return user, true
}

func (s *MemoryStore) ListItems(q, status string, page, pageSize int) ([]models.Item, int) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := make([]models.Item, 0, len(s.items))
	for _, item := range s.items {
		if status != "" && item.Status != status {
			continue
		}
		if q != "" {
			text := strings.ToLower(item.Title + " " + item.Description + " " + strings.Join(item.Tags, " "))
			if !strings.Contains(text, strings.ToLower(q)) {
				continue
			}
		}
		items = append(items, item)
	}

	slices.SortFunc(items, func(a, b models.Item) int {
		switch {
		case a.ID < b.ID:
			return -1
		case a.ID > b.ID:
			return 1
		default:
			return 0
		}
	})

	total := len(items)
	start := (page - 1) * pageSize
	if start >= total {
		return []models.Item{}, total
	}

	end := min(start+pageSize, total)
	return items[start:end], total
}

func (s *MemoryStore) GetItem(id int) (models.Item, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	item, ok := s.items[id]
	return item, ok
}

func (s *MemoryStore) CreateItem(req models.CreateItemRequest) (models.Item, error) {
	if err := validateStatus(req.Status); err != nil {
		return models.Item{}, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()
	item := models.Item{
		ID:          s.nextItemID,
		Title:       strings.TrimSpace(req.Title),
		Description: strings.TrimSpace(req.Description),
		Status:      req.Status,
		Tags:        sanitizeTags(req.Tags),
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	s.items[item.ID] = item
	s.nextItemID++

	return item, nil
}

func (s *MemoryStore) ReplaceItem(id int, req models.ReplaceItemRequest) (models.Item, error) {
	if err := validateStatus(req.Status); err != nil {
		return models.Item{}, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	item, ok := s.items[id]
	if !ok {
		return models.Item{}, ErrItemNotFound
	}

	item.Title = strings.TrimSpace(req.Title)
	item.Description = strings.TrimSpace(req.Description)
	item.Status = req.Status
	item.Tags = sanitizeTags(req.Tags)
	item.UpdatedAt = time.Now().UTC()
	s.items[id] = item

	return item, nil
}

func (s *MemoryStore) PatchItem(id int, req models.PatchItemRequest) (models.Item, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	item, ok := s.items[id]
	if !ok {
		return models.Item{}, ErrItemNotFound
	}

	if req.Title != nil {
		item.Title = strings.TrimSpace(*req.Title)
	}
	if req.Description != nil {
		item.Description = strings.TrimSpace(*req.Description)
	}
	if req.Status != nil {
		if err := validateStatus(*req.Status); err != nil {
			return models.Item{}, err
		}
		item.Status = *req.Status
	}
	if req.Tags != nil {
		item.Tags = sanitizeTags(*req.Tags)
	}

	item.UpdatedAt = time.Now().UTC()
	s.items[id] = item
	return item, nil
}

func (s *MemoryStore) DeleteItem(id int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.items[id]; !ok {
		return false
	}

	delete(s.items, id)
	return true
}

func validateStatus(status string) error {
	if _, ok := allowedStatuses[status]; !ok {
		return ErrInvalidItemState
	}
	return nil
}

func sanitizeTags(tags []string) []string {
	clean := make([]string, 0, len(tags))
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag != "" {
			clean = append(clean, tag)
		}
	}
	return clean
}
