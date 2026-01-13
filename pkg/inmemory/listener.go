package inmemory

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// StreamEventListener handles Change Stream events for typed entities.
// Implementations receive notifications when entities are added, updated, or deleted.
type StreamEventListener[T d] interface {
	Add(ctx context.Context, v T)
	Update(ctx context.Context, _id primitive.ObjectID, updatedFields T, removedFields []string)
	Delete(ctx context.Context, _id primitive.ObjectID)
}

// Listener coordinates multiple StreamEventListeners and manages the execution order.
// BeforeListeners are called before cache operations, regular listeners are called after.
type Listener[T d] struct {
	cache           Cache[T]
	listeners       []StreamEventListener[T]
	beforeListeners []StreamEventListener[T]
}

// Add processes an Add event by calling before listeners, updating the cache, then calling after listeners.
func (c *Listener[T]) Add(ctx context.Context, v T) {
	for _, listener := range c.beforeListeners {
		listener.Add(ctx, v)
	}
	c.cache.Add(ctx, v)
	for _, listener := range c.listeners {
		listener.Add(ctx, v)
	}
}

// Update processes an Update event by calling before listeners, updating the cache, then calling after listeners.
func (c *Listener[T]) Update(ctx context.Context, _id primitive.ObjectID, updatedFields T, removedFields []string) {
	for _, listener := range c.beforeListeners {
		listener.Update(ctx, _id, updatedFields, removedFields)
	}
	c.cache.Update(ctx, _id, updatedFields, removedFields)
	for _, listener := range c.listeners {
		listener.Update(ctx, _id, updatedFields, removedFields)
	}
}

// Delete processes a Delete event by calling before listeners, deleting from the cache, then calling after listeners.
func (c *Listener[T]) Delete(ctx context.Context, _id primitive.ObjectID) {
	for _, listener := range c.beforeListeners {
		listener.Delete(ctx, _id)
	}
	c.cache.Delete(ctx, _id)
	for _, listener := range c.listeners {
		listener.Delete(ctx, _id)
	}
}

// AddListener registers a new StreamEventListener.
// If before is true, the listener is called before cache operations; otherwise, it's called after.
func (c *Listener[T]) AddListener(listener StreamEventListener[T], before bool) (idx int) {
	if before {
		c.beforeListeners = append(c.beforeListeners, listener)
		return
	}
	c.listeners = append(c.listeners, listener)
	return
}

// NewListener creates a new Listener that coordinates cache operations and event listeners.
func NewListener[T d](cache Cache[T]) *Listener[T] {
	return &Listener[T]{
		cache:           cache,
		listeners:       []StreamEventListener[T]{},
		beforeListeners: []StreamEventListener[T]{},
	}
}

// AddCallbackListener is a StreamEventListener that calls a callback function only for Add events.
type AddCallbackListener[T d] struct {
	callback func(ctx context.Context, v T)
}

// NewAddCallbackListener creates a new AddCallbackListener with the specified callback.
func NewAddCallbackListener[T d](callback func(ctx context.Context, v T)) *AddCallbackListener[T] {
	return &AddCallbackListener[T]{
		callback: callback,
	}
}

func (s *AddCallbackListener[T]) Add(ctx context.Context, v T) {
	s.callback(ctx, v)
}

func (s *AddCallbackListener[T]) Update(ctx context.Context, _id primitive.ObjectID, updatedFields T, removedFields []string) {
}

func (s *AddCallbackListener[T]) Delete(ctx context.Context, _id primitive.ObjectID) {

}

// UpdateCallbackListener is a StreamEventListener that calls a callback function only for Update events.
type UpdateCallbackListener[T d] struct {
	callback func(ctx context.Context, id string, v T, removedFields []string)
}

// NewUpdateCallbackListener creates a new UpdateCallbackListener with the specified callback.
func NewUpdateCallbackListener[T d](callback func(ctx context.Context, id string, v T, removedFields []string)) *UpdateCallbackListener[T] {
	return &UpdateCallbackListener[T]{
		callback: callback,
	}
}

func (s *UpdateCallbackListener[T]) Add(ctx context.Context, v T) {
}

func (s *UpdateCallbackListener[T]) Update(ctx context.Context, _id primitive.ObjectID, updatedFields T, removedFields []string) {
	s.callback(ctx, _id.Hex(), updatedFields, removedFields)
}

func (s *UpdateCallbackListener[T]) Delete(ctx context.Context, _id primitive.ObjectID) {
}

// DeleteCallbackListener is a StreamEventListener that calls a callback function only for Delete events.
type DeleteCallbackListener[T d] struct {
	callback func(ctx context.Context, id string)
}

// NewDeleteCallbackListener creates a new DeleteCallbackListener with the specified callback.
func NewDeleteCallbackListener[T d](callback func(ctx context.Context, id string)) *DeleteCallbackListener[T] {
	return &DeleteCallbackListener[T]{
		callback: callback,
	}
}

func (s *DeleteCallbackListener[T]) Add(ctx context.Context, v T) {
}

func (s *DeleteCallbackListener[T]) Update(ctx context.Context, _id primitive.ObjectID, updatedFields T, removedFields []string) {
}

func (s *DeleteCallbackListener[T]) Delete(ctx context.Context, _id primitive.ObjectID) {
	s.callback(ctx, _id.Hex())
}
