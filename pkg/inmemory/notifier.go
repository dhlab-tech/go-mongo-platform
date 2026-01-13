package inmemory

import (
	"context"
	"sync"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Notify provides a mechanism to wait for cache updates after write operations.
// It is used by Await* methods to ensure read-after-write consistency.
type Notify[T d] interface {
	StreamEventListener[T]
	AddListenerCreate(id string, c func()) string
	AddListenerUpdate(id string, c func()) string
	AddListenerDelete(id string, c func()) string
	DeleteListenerCreate(id, ui string)
	DeleteListenerUpdate(id, ui string)
	DeleteListenerDelete(id, ui string)
}

// Notifier is used to wait for cache updates.
// For example, you need to write to storage and wait for the update in inmemory:
// 1) subscribe to notify
// 2) write to storage
// 3) wait for the update in inmemory
// The wait subscription is automatically removed
type Notifier[T d] struct {
	sync.RWMutex
	listenersCreate map[string]map[string]func()
	listenersUpdate map[string]map[string]func()
	listenersDelete map[string]map[string]func()
}

// Add ...
func (s *Notifier[T]) Add(ctx context.Context, v T) {
	s.Lock()
	defer s.Unlock()
	if c, ok := s.listenersCreate[v.ID()]; ok {
		for _, l := range c {
			l()
		}
	}
	delete(s.listenersCreate, v.ID())
}

// Update ...
func (s *Notifier[T]) Update(ctx context.Context, _id primitive.ObjectID, updatedFields T, removedFields []string) {
	s.Lock()
	defer s.Unlock()
	if c, ok := s.listenersUpdate[_id.Hex()]; ok {
		for _, l := range c {
			l()
		}
	}
	delete(s.listenersUpdate, _id.Hex())
}

// Delete ...
func (s *Notifier[T]) Delete(ctx context.Context, _id primitive.ObjectID) {
	s.Lock()
	defer s.Unlock()
	if c, ok := s.listenersDelete[_id.Hex()]; ok {
		for _, l := range c {
			l()
		}
	}
	delete(s.listenersDelete, _id.Hex())
}

// AddListenerCreate registers a callback to be called when an entity with the given ID is created.
// Returns a unique listener ID that can be used to remove the listener.
func (s *Notifier[T]) AddListenerCreate(id string, c func()) string {
	s.Lock()
	defer s.Unlock()
	ui := uuid.NewString()
	if _, ok := s.listenersCreate[id]; !ok {
		s.listenersCreate[id] = map[string]func(){}
	}
	s.listenersCreate[id][ui] = c
	return ui
}

// AddListenerUpdate registers a callback to be called when an entity with the given ID is updated.
// Returns a unique listener ID that can be used to remove the listener.
func (s *Notifier[T]) AddListenerUpdate(id string, c func()) string {
	s.Lock()
	defer s.Unlock()
	ui := uuid.NewString()
	if _, ok := s.listenersUpdate[id]; !ok {
		s.listenersUpdate[id] = map[string]func(){}
	}
	s.listenersUpdate[id][ui] = c
	return ui
}

// AddListenerDelete registers a callback to be called when an entity with the given ID is deleted.
// Returns a unique listener ID that can be used to remove the listener.
func (s *Notifier[T]) AddListenerDelete(id string, c func()) string {
	s.Lock()
	defer s.Unlock()
	ui := uuid.NewString()
	if _, ok := s.listenersDelete[id]; !ok {
		s.listenersDelete[id] = map[string]func(){}
	}
	s.listenersDelete[id][ui] = c
	return ui
}

// DeleteListenerCreate removes a create listener by its unique ID.
func (s *Notifier[T]) DeleteListenerCreate(id, ui string) {
	s.Lock()
	defer s.Unlock()
	if _, ok := s.listenersCreate[id]; !ok {
		return
	}
	delete(s.listenersCreate[id], ui)
}

// DeleteListenerUpdate removes an update listener by its unique ID.
func (s *Notifier[T]) DeleteListenerUpdate(id, ui string) {
	s.Lock()
	defer s.Unlock()
	if _, ok := s.listenersUpdate[id]; !ok {
		return
	}
	delete(s.listenersUpdate[id], ui)
}

// DeleteListenerDelete removes a delete listener by its unique ID.
func (s *Notifier[T]) DeleteListenerDelete(id, ui string) {
	s.Lock()
	defer s.Unlock()
	if _, ok := s.listenersDelete[id]; !ok {
		return
	}
	delete(s.listenersDelete[id], ui)
}

// NewNotifier creates a new Notifier instance for waiting on cache updates.
// The maps are used to store listeners organized by entity ID and listener ID.
func NewNotifier[T d](
	listenersCreate map[string]map[string]func(),
	listenersUpdate map[string]map[string]func(),
	listenersDelete map[string]map[string]func(),
) *Notifier[T] {
	return &Notifier[T]{
		listenersCreate: listenersCreate,
		listenersUpdate: listenersUpdate,
		listenersDelete: listenersDelete,
	}
}
