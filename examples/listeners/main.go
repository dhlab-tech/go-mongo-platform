package main

import (
	"context"
	"log"
	"os"
	"sync"
	"time"

	"github.com/dhlab-tech/go-mongo-platform/pkg/inmemory"
	"github.com/dhlab-tech/go-mongo-platform/pkg/mongo"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/bson/primitive"
	mongodb "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// User represents a simple user entity
type User struct {
	Id      primitive.ObjectID `bson:"_id"`
	V       *int64             `bson:"version"`
	Deleted *bool              `bson:"deleted"`
	Name    *string            `bson:"name"`
	Email   *string            `bson:"email"`
}

// ID implements the d interface
func (u *User) ID() string {
	if u.Id.IsZero() {
		u.Id = primitive.NewObjectID()
	}
	return u.Id.Hex()
}

// Version implements the d interface
func (u *User) Version() *int64 {
	return u.V
}

// SetDeleted implements the d interface
func (u *User) SetDeleted(d bool) {
	_deleted := d
	u.Deleted = &_deleted
}

func main() {
	// Initialize logger
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	ctx := logger.WithContext(context.Background())

	// MongoDB connection string
	uri := "mongodb://localhost:27017/?replicaSet=rs0"
	if uriEnv := os.Getenv("MONGODB_URI"); uriEnv != "" {
		uri = uriEnv
	}

	// Connect to MongoDB
	client, err := mongodb.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer client.Disconnect(ctx)

	// Verify connection
	if err := client.Ping(ctx, nil); err != nil {
		log.Fatalf("Failed to ping MongoDB: %v", err)
	}
	logger.Info().Msg("Connected to MongoDB")

	// Create Change Stream on the database
	db := client.Database("example_db")
	changeStream, err := db.Watch(ctx, mongodb.Pipeline{}, options.ChangeStream().SetFullDocument(options.UpdateLookup))
	if err != nil {
		log.Fatalf("Failed to create Change Stream: %v", err)
	}

	// Create Stream instance
	stream := mongo.NewStream(changeStream, make(map[string]map[string]mongo.StreamListener))

	// Track listener execution order for demonstration
	var mu sync.Mutex
	var executionOrder []string

	// Create Before Listeners (DP-029)
	beforeAddListener := inmemory.NewAddCallbackListener[*User](func(ctx context.Context, u *User) {
		mu.Lock()
		defer mu.Unlock()
		executionOrder = append(executionOrder, "before:add")
		logger.Info().Str("id", u.ID()).Str("name", *u.Name).Msg("Before Add Listener: Validating user")
	})

	beforeUpdateListener := inmemory.NewUpdateCallbackListener[*User](func(ctx context.Context, id string, u *User, removedFields []string) {
		mu.Lock()
		defer mu.Unlock()
		executionOrder = append(executionOrder, "before:update")
		logger.Info().Str("id", id).Msg("Before Update Listener: Pre-processing update")
	})

	beforeDeleteListener := inmemory.NewDeleteCallbackListener[*User](func(ctx context.Context, id string) {
		mu.Lock()
		defer mu.Unlock()
		executionOrder = append(executionOrder, "before:delete")
		logger.Info().Str("id", id).Msg("Before Delete Listener: Pre-processing delete")
	})

	// Create After Listeners (DP-029)
	afterAddListener := inmemory.NewAddCallbackListener[*User](func(ctx context.Context, u *User) {
		mu.Lock()
		defer mu.Unlock()
		executionOrder = append(executionOrder, "after:add")
		logger.Info().Str("id", u.ID()).Msg("After Add Listener: User added to cache")
	})

	afterUpdateListener := inmemory.NewUpdateCallbackListener[*User](func(ctx context.Context, id string, u *User, removedFields []string) {
		mu.Lock()
		defer mu.Unlock()
		executionOrder = append(executionOrder, "after:update")
		logger.Info().Str("id", id).Msg("After Update Listener: User updated in cache")
	})

	afterDeleteListener := inmemory.NewDeleteCallbackListener[*User](func(ctx context.Context, id string) {
		mu.Lock()
		defer mu.Unlock()
		executionOrder = append(executionOrder, "after:delete")
		logger.Info().Str("id", id).Msg("After Delete Listener: User deleted from cache")
	})

	// Create in-memory projection with Before and After listeners
	dbName := "example_db"
	collectionName := "users"
	connectionTimeout := 10 * time.Second

	im, err := inmemory.NewInMemory[*User](
		ctx,
		stream,
		inmemory.MongoDeps{
			Client:            client,
			Db:                dbName,
			ConnectionTimeout: connectionTimeout,
		},
		inmemory.Entity[*User]{
			Collection: collectionName,
			BeforeListeners: []inmemory.StreamEventListener[*User]{
				beforeAddListener,
				beforeUpdateListener,
				beforeDeleteListener,
			},
			AfterListeners: []inmemory.StreamEventListener[*User]{
				afterAddListener,
				afterUpdateListener,
				afterDeleteListener,
			},
		},
	)
	if err != nil {
		log.Fatalf("Failed to create in-memory projection: %v", err)
	}

	// Start Change Stream listener in background
	go func() {
		if err := stream.Listen(ctx); err != nil {
			logger.Error().Err(err).Msg("Change Stream listener error")
		}
	}()

	// Wait for initial load
	time.Sleep(2 * time.Second)
	logger.Info().Msg("In-memory projection ready")

	// Demonstrate Before/After Listeners (DP-029)
	logger.Info().Msg("=== Before/After Listeners Demonstration ===")
	name := "Alice"
	email := "alice@example.com"
	user := &User{
		Name:  &name,
		Email: &email,
	}

	// Clear execution order
	mu.Lock()
	executionOrder = []string{}
	mu.Unlock()

	// Create user - should trigger before:add, then after:add
	id, err := im.AwaitCreate(ctx, user)
	if err != nil {
		log.Fatalf("Failed to create user: %v", err)
	}
	logger.Info().Str("id", id).Msg("User created")

	// Wait a moment for listeners to execute
	time.Sleep(500 * time.Millisecond)

	// Show execution order
	mu.Lock()
	logger.Info().Strs("execution_order", executionOrder).Msg("Listener execution order for Create")
	mu.Unlock()

	// Update user - should trigger before:update, then after:update
	updatedEmail := "alice.updated@example.com"
	readUser, found := im.GetCacheWithEventListener().Cache.Get(ctx, id)
	if !found {
		log.Fatalf("User not found in cache")
	}
	readUser.Email = &updatedEmail

	// Clear execution order
	mu.Lock()
	executionOrder = []string{}
	mu.Unlock()

	_, err = im.AwaitUpdate(ctx, readUser)
	if err != nil {
		log.Fatalf("Failed to update user: %v", err)
	}
	logger.Info().Str("id", id).Msg("User updated")

	// Wait a moment for listeners to execute
	time.Sleep(500 * time.Millisecond)

	// Show execution order
	mu.Lock()
	logger.Info().Strs("execution_order", executionOrder).Msg("Listener execution order for Update")
	mu.Unlock()

	// Delete user - should trigger before:delete, then after:delete
	cacheUser, found := im.GetCacheWithEventListener().Cache.Get(ctx, id)
	if !found {
		log.Fatalf("User not found in cache")
	}

	// Clear execution order
	mu.Lock()
	executionOrder = []string{}
	mu.Unlock()

	err = im.AwaitDelete(ctx, cacheUser)
	if err != nil {
		log.Fatalf("Failed to delete user: %v", err)
	}
	logger.Info().Str("id", id).Msg("User deleted")

	// Wait a moment for listeners to execute
	time.Sleep(500 * time.Millisecond)

	// Show execution order
	mu.Lock()
	logger.Info().Strs("execution_order", executionOrder).Msg("Listener execution order for Delete")
	mu.Unlock()

	// Demonstrate Notify/AwaitNotify (DP-030)
	logger.Info().Msg("=== Notify/AwaitNotify Demonstration ===")
	logger.Info().Msg("AwaitCreate, AwaitUpdate, and AwaitDelete use Notify/AwaitNotify internally")
	logger.Info().Msg("The Await* operations:")
	logger.Info().Msg("1. Subscribe to notify (via AwaitNotify)")
	logger.Info().Msg("2. Write to MongoDB")
	logger.Info().Msg("3. Wait for Change Stream to propagate the change")
	logger.Info().Msg("4. Notify listener triggers, Await* operation returns")
	logger.Info().Msg("5. Subscription is automatically removed")

	// Create another user to demonstrate
	name2 := "Bob"
	email2 := "bob@example.com"
	user2 := &User{
		Name:  &name2,
		Email: &email2,
	}

	id2, err := im.AwaitCreate(ctx, user2)
	if err != nil {
		log.Fatalf("Failed to create user2: %v", err)
	}
	logger.Info().
		Str("id", id2).
		Msg("AwaitCreate completed - Notify/AwaitNotify ensured read-after-write consistency")

	// Verify user is immediately available in cache (demonstrates Notify/AwaitNotify guarantee)
	_, found = im.GetCacheWithEventListener().Cache.Get(ctx, id2)
	if !found {
		log.Fatalf("User not found in cache after AwaitCreate - Notify/AwaitNotify failed")
	}
	logger.Info().Str("id", id2).Msg("User immediately available in cache (Notify/AwaitNotify guarantee)")

	logger.Info().Msg("=== All Listener Types Demonstrated Successfully ===")
	logger.Info().Msg("1. Before Listeners: Execute before cache updates (validation, pre-processing)")
	logger.Info().Msg("2. After Listeners: Execute after cache updates (side effects, notifications)")
	logger.Info().Msg("3. Notify/AwaitNotify: Used internally by Await* operations for read-after-write consistency")
}

