package main

import (
	"context"
	"log"
	"os"
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
	// Note: Change Streams require a replica set
	db := client.Database("example_db")
	changeStream, err := db.Watch(ctx, mongodb.Pipeline{}, options.ChangeStream().SetFullDocument(options.UpdateLookup))
	if err != nil {
		log.Fatalf("Failed to create Change Stream: %v", err)
	}

	// Create Stream instance
	stream := mongo.NewStream(changeStream, make(map[string]map[string]mongo.StreamListener))

	// Create in-memory projection
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

	// Wait a moment for initial load
	time.Sleep(2 * time.Second)
	logger.Info().Msg("In-memory projection ready")

	// Demonstrate AwaitCreate
	logger.Info().Msg("=== Creating a user ===")
	name := "Alice"
	email := "alice@example.com"
	user := &User{
		Name:  &name,
		Email: &email,
	}

	id, err := im.AwaitCreate(ctx, user)
	if err != nil {
		log.Fatalf("Failed to create user: %v", err)
	}
	logger.Info().Str("id", id).Msg("User created")

	// Read from cache immediately (demonstrates read-after-write consistency)
	cache := im.GetCacheWithEventListener()
	readUser, found := cache.Cache.Get(ctx, id)
	if !found {
		log.Fatalf("User not found in cache after creation")
	}
	logger.Info().
		Str("id", readUser.ID()).
		Str("name", *readUser.Name).
		Str("email", *readUser.Email).
		Msg("User read from cache (read-after-write consistency)")

	// Demonstrate AwaitUpdate
	logger.Info().Msg("=== Updating the user ===")
	updatedEmail := "alice.updated@example.com"
	readUser.Email = &updatedEmail

	updatedUser, err := im.AwaitUpdate(ctx, readUser)
	if err != nil {
		log.Fatalf("Failed to update user: %v", err)
	}
	logger.Info().
		Str("id", updatedUser.ID()).
		Str("email", *updatedUser.Email).
		Msg("User updated")

	// Verify update in cache
	cacheUser, found := cache.Cache.Get(ctx, id)
	if !found {
		log.Fatalf("User not found in cache after update")
	}
	if *cacheUser.Email != updatedEmail {
		log.Fatalf("Cache does not reflect update: expected %s, got %s", updatedEmail, *cacheUser.Email)
	}
	logger.Info().
		Str("email", *cacheUser.Email).
		Msg("Update verified in cache (read-after-write consistency)")

	// Demonstrate AwaitDelete
	logger.Info().Msg("=== Deleting the user ===")
	err = im.AwaitDelete(ctx, cacheUser)
	if err != nil {
		log.Fatalf("Failed to delete user: %v", err)
	}
	logger.Info().Str("id", id).Msg("User deleted")

	// Verify deletion in cache
	_, found = cache.Cache.Get(ctx, id)
	if found {
		log.Fatalf("User still found in cache after deletion")
	}
	logger.Info().Msg("Deletion verified in cache (read-after-write consistency)")

	logger.Info().Msg("=== Example completed successfully ===")
	logger.Info().Msg("This demonstrates deterministic read-after-write behavior:")
	logger.Info().Msg("1. AwaitCreate blocks until change is observed in memory")
	logger.Info().Msg("2. AwaitUpdate blocks until change is observed in memory")
	logger.Info().Msg("3. AwaitDelete blocks until change is observed in memory")
	logger.Info().Msg("4. All subsequent reads from cache reflect the writes")
}
