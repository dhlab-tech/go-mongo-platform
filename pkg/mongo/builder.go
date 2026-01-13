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
	Find(ctx context.Context, filter map[string]interface{}) (items []T, err error)
	FindOne(ctx context.Context, filter map[string]interface{}) (item T, found bool, err error)
	FindWithFilter(ctx context.Context, filter bson.M) (items []T, err error)
	FindOneWithFilter(ctx context.Context, filter bson.M) (item T, found bool, err error)
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

// StreamListener handles MongoDB Change Stream events.
// It processes change events from MongoDB Change Streams and applies them
// to the in-memory projection layer.
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

// Mongo provides MongoDB operations and Change Streams integration for a typed entity.
// It aggregates all MongoDB-related operations: searching, creating, updating, deleting,
// and listening to Change Streams. Each field provides a specific set of operations
// for interacting with MongoDB.
type Mongo[T d] struct {
	Searcher  searcher[T]    // Provides search and query operations
	Processor processor[T]   // Handles create, update, and delete operations
	Listener  StreamListener // Processes Change Stream events
	Creator   creator        // Creates documents in MongoDB
	Updater   updater        // Updates documents in MongoDB
	Upserter  upserter       // Upserts documents in MongoDB
	Remover   remover        // Removes documents from MongoDB
}

// NewMongo creates a new Mongo instance for a typed entity.
// It initializes all MongoDB operations (searcher, processor, creator, updater, etc.)
// and sets up Change Streams listener for the specified collection.
//
// Parameters:
//   - client: MongoDB client connection
//   - db: Database name
//   - collection: Collection name
//   - connectionTimeout: Timeout for MongoDB operations
//   - cache: In-memory cache for the entity type
//   - handler: Handler for Change Stream events (Add, Update, Delete)
//
// Returns a fully configured Mongo instance ready for use.
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
