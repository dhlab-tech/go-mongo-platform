package inmemory

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// StreamEventListener ...
type StreamEventListener[T d] interface {
	Add(ctx context.Context, v T)
	Update(ctx context.Context, _id primitive.ObjectID, updatedFields T, removedFields []string)
	Delete(ctx context.Context, _id primitive.ObjectID)
}

// Listener ...
type Listener[T d] struct {
	cache           Cache[T]
	listeners       []StreamEventListener[T]
	beforeListeners []StreamEventListener[T]
}

// Add ...
func (c *Listener[T]) Add(ctx context.Context, v T) {
	for _, listener := range c.beforeListeners {
		listener.Add(ctx, v)
	}
	c.cache.Add(ctx, v)
	for _, listener := range c.listeners {
		listener.Add(ctx, v)
	}
}

// Update ...
func (c *Listener[T]) Update(ctx context.Context, _id primitive.ObjectID, updatedFields T, removedFields []string) {
	for _, listener := range c.beforeListeners {
		listener.Update(ctx, _id, updatedFields, removedFields)
	}
	c.cache.Update(ctx, _id, updatedFields, removedFields)
	for _, listener := range c.listeners {
		listener.Update(ctx, _id, updatedFields, removedFields)
	}
}

// Delete ...
func (c *Listener[T]) Delete(ctx context.Context, _id primitive.ObjectID) {
	for _, listener := range c.beforeListeners {
		listener.Delete(ctx, _id)
	}
	c.cache.Delete(ctx, _id)
	for _, listener := range c.listeners {
		listener.Delete(ctx, _id)
	}
}

// AddListener ...
func (c *Listener[T]) AddListener(listener StreamEventListener[T], before bool) (idx int) {
	if before {
		c.beforeListeners = append(c.beforeListeners, listener)
		return
	}
	c.listeners = append(c.listeners, listener)
	return
}

// NewListener ...
func NewListener[T d](cache Cache[T]) *Listener[T] {
	return &Listener[T]{
		cache:           cache,
		listeners:       []StreamEventListener[T]{},
		beforeListeners: []StreamEventListener[T]{},
	}
}
