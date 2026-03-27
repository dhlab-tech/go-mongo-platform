package inmemory

import (
	"context"
	"errors"
	"reflect"

	"github.com/dhlab-tech/go-mongo-platform/pkg/mongo"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// noOpHandler is a no-op implementation of mongo handler interface
type noOpHandler[T d] struct{}

func (n *noOpHandler[T]) Add(ctx context.Context, v T) {}

func (n *noOpHandler[T]) Update(ctx context.Context, id primitive.ObjectID, updatedFields T, removedFields []string) {
}

func (n *noOpHandler[T]) Delete(ctx context.Context, _id primitive.ObjectID) {}

// isStreamValid checks if stream interface is not nil and contains a valid value
func isStreamValid(s stream) bool {
	if s == nil {
		return false
	}
	v := reflect.ValueOf(s)
	if !v.IsValid() {
		return false
	}
	// Check if the value inside the interface is nil (e.g., (*SomeType)(nil))
	if v.Kind() == reflect.Ptr && v.IsNil() {
		return false
	}
	return true
}

type stream interface {
	AddListener(ctx context.Context, db, col string, listener streamListener)
}

type streamListener = interface {
	Listen(ctx context.Context, change []byte) (err error)
}

// InMemory provides the main interface for working with typed entities in MongoDB
// with an in-memory projection layer. It combines MongoDB operations with Change Streams
// synchronization to maintain a strongly consistent in-memory cache.
type InMemory[T d] interface {
	GetCacheWithEventListener() *CacheWithEventListener[T]
	GetMongo() *mongo.Mongo[T]
	Spawn(ctx context.Context) T
	AwaitCreate(ctx context.Context, ps T) (id string, err error)
	AwaitUpdate(ctx context.Context, ps T) (res T, err error)
	AwaitUpdateDoc(ctx context.Context, id string, set, unset bson.D) (found bool, err error)
	AwaitDelete(ctx context.Context, ps T) (err error)
}

type inMemory[T d] struct {
	CacheWithEventListener *CacheWithEventListener[T]
	Mongo                  *mongo.Mongo[T]
}

// Spawn creates a new instance of the entity type T.
// For pointer types, it creates a new pointer to a zero value.
// For value types, it creates a zero value.
func (im *inMemory[T]) Spawn(ctx context.Context) (instance T) {
	_t := reflect.TypeOf(instance)
	if _t.Kind() == reflect.Ptr {
		instance = reflect.New(_t.Elem()).Interface().(T)
		instance.ID()
		return
	}
	instance = reflect.New(_t).Elem().Interface().(T)
	instance.ID()
	return
}

// GetCacheWithEventListener returns the cache with event listeners and indexes.
func (im *inMemory[T]) GetCacheWithEventListener() *CacheWithEventListener[T] {
	return im.CacheWithEventListener
}

// GetMongo returns the MongoDB operations instance.
func (im *inMemory[T]) GetMongo() *mongo.Mongo[T] {
	return im.Mongo
}

// AwaitCreate creates an entity in MongoDB and waits until the change is reflected in the in-memory cache.
// This provides read-after-write consistency: after AwaitCreate returns, subsequent reads from the cache
// will see the newly created entity. Returns the ID of the created entity.
func (p *inMemory[T]) AwaitCreate(ctx context.Context, ps T) (id string, err error) {
	if p.CacheWithEventListener == nil {
		return "", errors.New("cache is not initialized, AwaitCreate requires cache")
	}
	ch := make(chan struct{})
	defer close(ch)
	ui := p.CacheWithEventListener.AwaitNotify.AddListenerCreate(ps.ID(), func() {
		ch <- struct{}{}
	})
	_, err = p.Mongo.Processor.Create(ctx, ps)
	if err != nil {
		p.CacheWithEventListener.AwaitNotify.DeleteListenerCreate(ps.ID(), ui)
		return
	}
	<-ch
	id = ps.ID()
	return
}

// AwaitUpdate updates an entity in MongoDB and waits until the change is reflected in the in-memory cache.
// This provides read-after-write consistency: after AwaitUpdate returns, subsequent reads from the cache
// will see the updated entity. Returns the updated entity.
func (p *inMemory[T]) AwaitUpdate(ctx context.Context, ps T) (res T, err error) {
	if p.CacheWithEventListener == nil {
		return res, errors.New("cache is not initialized, AwaitUpdate requires cache")
	}
	ch := make(chan struct{})
	defer close(ch)
	ui := p.CacheWithEventListener.AwaitNotify.AddListenerUpdate(ps.ID(), func() {
		ch <- struct{}{}
	})
	res, err = p.Mongo.Processor.Update(ctx, ps)
	if err != nil {
		p.CacheWithEventListener.AwaitNotify.DeleteListenerUpdate(ps.ID(), ui)
		if errors.Is(err, mongo.ErrNothingToUpdate) {
			err = nil
		}
		return
	}
	<-ch
	return
}

// AwaitUpdateDoc updates a document directly using BSON update operations and waits until
// the change is reflected in the in-memory cache. This provides read-after-write consistency.
// Returns a boolean indicating if the document was found and updated.
func (p *inMemory[T]) AwaitUpdateDoc(ctx context.Context, id string, set, unset bson.D) (found bool, err error) {
	if p.CacheWithEventListener == nil {
		return false, errors.New("cache is not initialized, AwaitUpdateDoc requires cache")
	}
	ch := make(chan struct{})
	defer close(ch)
	ui := p.CacheWithEventListener.AwaitNotify.AddListenerUpdate(id, func() {
		ch <- struct{}{}
	})
	found, err = p.Mongo.Updater.UpdateOne(ctx, id, nil, set, unset)
	if err != nil {
		p.CacheWithEventListener.AwaitNotify.DeleteListenerUpdate(id, ui)
		if errors.Is(err, mongo.ErrNothingToUpdate) {
			err = nil
		}
		return
	}
	<-ch
	return
}

// AwaitDelete deletes an entity from MongoDB and waits until the change is reflected in the in-memory cache.
// This provides read-after-write consistency: after AwaitDelete returns, subsequent reads from the cache
// will not find the deleted entity.
func (p *inMemory[T]) AwaitDelete(ctx context.Context, ps T) (err error) {
	if p.CacheWithEventListener == nil {
		return errors.New("cache is not initialized, AwaitDelete requires cache")
	}
	ch := make(chan struct{})
	defer close(ch)
	ui := p.CacheWithEventListener.AwaitNotify.AddListenerDelete(ps.ID(), func() {
		ch <- struct{}{}
	})
	err = p.Mongo.Processor.Delete(ctx, ps.ID())
	if err != nil {
		p.CacheWithEventListener.AwaitNotify.DeleteListenerDelete(ps.ID(), ui)
		return
	}
	<-ch
	return
}

// NewInMemory creates a new InMemory instance for a typed entity.
// It sets up MongoDB operations, Change Streams listener, and in-memory cache with indexes.
// On initialization, it loads all existing documents from MongoDB into the cache.
// Returns nil if the collection name is empty (no-op mode).
func NewInMemory[T d](ctx context.Context, stream stream, deps MongoDeps, entityDeps Entity[T]) (InMemory[T], error) {
	if entityDeps.Collection == "" {
		return nil, nil
	}
	var im *CacheWithEventListener[T]
	var cache Cache[T]
	var handler interface {
		Add(ctx context.Context, v T)
		Update(ctx context.Context, id primitive.ObjectID, updatedFields T, removedFields []string)
		Delete(ctx context.Context, _id primitive.ObjectID)
	}
	if isStreamValid(stream) {
		im = NewCacheWithEventListener[T](
			entityDeps.BeforeListeners,
			entityDeps.AfterListeners,
			entityDeps.Notify,
		)
		cache = im.Cache
		handler = im.EventListener
	} else {
		handler = &noOpHandler[T]{}
	}
	m := mongo.NewMongo[T](
		deps.Client,
		deps.Db,
		entityDeps.Collection,
		deps.ConnectionTimeout,
		cache,
		handler,
	)
	if isStreamValid(stream) {
		stream.AddListener(ctx, deps.Db, entityDeps.Collection, m.Listener)
	}
	i := inMemory[T]{
		CacheWithEventListener: im,
		Mongo:                  m,
	}
	if entityDeps.Option != nil {
		entityDeps.Option(&i)
	}
	zerolog.Ctx(ctx).Debug().Str("collection", entityDeps.Collection).Any("im", im).Msg("in-memory initialized")
	if im != nil {
		var (
			its []T
			err error
		)
		if entityDeps.WarmupFilter != nil {
			its, err = m.Searcher.FindWithFilter(ctx, *entityDeps.WarmupFilter)
		} else {
			its, err = m.Searcher.All(ctx)
		}
		if err != nil {
			return nil, err
		}
		for _, it := range its {
			im.EventListener.Add(ctx, it)
		}
	}
	return &i, nil
}
