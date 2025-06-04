package inmemory

import (
	"context"

	"github.com/dhlab-tech/go-mongo-platform/pkg/mongo"
)

type stream interface {
	AddListener(ctx context.Context, db, col string, listener streamListener)
}

type streamListener = interface {
	Listen(ctx context.Context, change []byte) (err error)
}

type InMemory[T d] interface {
	GetCacheWithEventListener() *CacheWithEventListener[T]
	GetMongo() *mongo.Mongo[T]
	AwaitCreate(ctx context.Context, ps T) (id string, err error)
	AwaitUpdate(ctx context.Context, ps T) (res T, err error)
	AwaitDelete(ctx context.Context, ps T) (err error)
}

type inMemory[T d] struct {
	CacheWithEventListener *CacheWithEventListener[T]
	Mongo                  *mongo.Mongo[T]
}

func (im *inMemory[T]) GetCacheWithEventListener() *CacheWithEventListener[T] {
	return im.CacheWithEventListener
}

func (im *inMemory[T]) GetMongo() *mongo.Mongo[T] {
	return im.Mongo
}

// AwaitCreate ...
func (p *inMemory[T]) AwaitCreate(ctx context.Context, ps T) (id string, err error) {
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
	return
}

// AwaitUpdate ...
func (p *inMemory[T]) AwaitUpdate(ctx context.Context, ps T) (res T, err error) {
	ch := make(chan struct{})
	defer close(ch)
	ui := p.CacheWithEventListener.AwaitNotify.AddListenerUpdate(ps.ID(), func() {
		ch <- struct{}{}
	})
	res, err = p.Mongo.Processor.Update(ctx, ps)
	if err != nil {
		p.CacheWithEventListener.AwaitNotify.DeleteListenerUpdate(ps.ID(), ui)
		return
	}
	<-ch
	return
}

// AwaitDelete ...
func (p *inMemory[T]) AwaitDelete(ctx context.Context, ps T) (err error) {
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

// NewInMemory ...
func NewInMemory[T d](ctx context.Context, stream stream, deps MongoDeps, entityDeps Entity[T]) (InMemory[T], error) {
	if entityDeps.Collection == "" {
		return nil, nil
	}
	im := NewCacheWithEventListener[T](
		entityDeps.BeforeListeners,
		entityDeps.AfterListeners,
		entityDeps.Notify,
	)
	m := mongo.NewMongo[T](
		deps.Client,
		deps.Db,
		entityDeps.Collection,
		deps.ConnectionTimeout,
		im.Cache,
		im.EventListener,
	)
	stream.AddListener(ctx, deps.Db, entityDeps.Collection, m.Listener)
	if entityDeps.Option != nil {
		entityDeps.Option(im, m)
	}
	its, err := m.Searcher.All(ctx)
	if err != nil {
		return nil, err
	}
	for _, it := range its {
		im.EventListener.Add(ctx, it)
	}
	return &inMemory[T]{
		CacheWithEventListener: im,
		Mongo:                  m,
	}, nil
}
