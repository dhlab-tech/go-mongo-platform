package mongo

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type searcher[T d] interface {
	All(ctx context.Context) (items []T, err error)
}

type processor[T d] interface {
	Create(ctx context.Context, ps T) (id string, err error)
	Update(ctx context.Context, ps T) (T, error)
	Delete(ctx context.Context, id string) (err error)
	PrepareCreate(ctx context.Context, ps T) (prepared T, doc bson.D, err error)
	PrepareUpdate(ctx context.Context, ps T) (prepared T, set bson.D, unset bson.D, err error)
}

type d interface {
	any
	ID() string
	Version() *int64
	SetDeleted(d bool)
}

type cache[T d] interface {
	All(ctx context.Context) (ids []string)
	Get(ctx context.Context, id string) (r T, f bool)
	GetByIndex(ctx context.Context, idx int) (r T, f bool)
}

type creator interface {
	Create(ctx context.Context, doc bson.D) (id primitive.ObjectID, err error)
	C(ctx context.Context, doc interface{}) (id primitive.ObjectID, err error)
}

type updater interface {
	UpdateOne(ctx context.Context, id string, version *int64, set bson.D, unset bson.D) (found bool, err error)
}

type upserter interface {
	UpsertOne(ctx context.Context, id string, doc interface{}) (err error)
	UpsertMany(ctx context.Context, ids []string, set []interface{}) (err error)
}

// StreamListener ...
type StreamListener = interface {
	Listen(ctx context.Context, change []byte) (err error)
}

type handler[T d] interface {
	Add(ctx context.Context, v T)
	Update(ctx context.Context, id primitive.ObjectID, updatedFields T, removedFields []string)
	Delete(ctx context.Context, _id primitive.ObjectID)
}

type remover interface {
	Remove(ctx context.Context, doc interface{}) (deletedCount int, err error)
	RemoveMany(ctx context.Context, doc interface{}) (deletedCount int, err error)
}

// Mongo ...
type Mongo[T d] struct {
	Searcher  searcher[T]
	Processor processor[T]
	Listener  StreamListener
	Creator   creator
	Updater   updater
	Upserter  upserter
	Remover   remover
}

// NewMongo ...
func NewMongo[T d](
	client *mongo.Client,
	db string,
	collection string,
	connectionTimeout time.Duration,
	cache cache[T],
	handler handler[T],
) *Mongo[T] {
	cr := NewCreator(client, db, collection, connectionTimeout)
	up := NewUpdater(client, db, collection, connectionTimeout)
	rm := NewRemover(client, db, collection, connectionTimeout)
	return &Mongo[T]{
		Searcher:  NewSearcher[T](client, db, collection, connectionTimeout),
		Processor: NewProcessor[T](cache, cr, up, rm),
		Listener:  NewListener(collection, handler),
		Creator:   cr,
		Updater:   up,
		Upserter:  NewUpsert(client, db, collection, connectionTimeout),
		Remover:   rm,
	}
}
