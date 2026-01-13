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

// Product represents a product entity with all four index types
type Product struct {
	Id      primitive.ObjectID `bson:"_id"`
	V       *int64             `bson:"version"`
	Deleted *bool              `bson:"deleted"`
	// Inverse Index: find all products by parent category
	ParentID *string `bson:"parent_id" indexes:"inverse:parent_id:from"`
	// Inverse Unique Index: find product by unique email (e.g., for notifications)
	Email *string `bson:"email" indexes:"inverse_unique:email_unique:from"`
	// Sorted Index: maintain products sorted by title
	// Suffix Index: text search on title (multiple indexes on same field)
	Title *string `bson:"title" indexes:"sorted:title:from,suffix:title:from"`
	// Suffix Index: text search on description
	Description *string `bson:"description" indexes:"suffix:description:from"`
}

// ID implements the d interface
func (p *Product) ID() string {
	if p.Id.IsZero() {
		p.Id = primitive.NewObjectID()
	}
	return p.Id.Hex()
}

// Version implements the d interface
func (p *Product) Version() *int64 {
	return p.V
}

// SetDeleted implements the d interface
func (p *Product) SetDeleted(d bool) {
	_deleted := d
	p.Deleted = &_deleted
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

	// Create in-memory projection with indexes
	dbName := "example_db"
	collectionName := "products"
	connectionTimeout := 10 * time.Second

	im, err := inmemory.NewInMemory[*Product](
		ctx,
		stream,
		inmemory.MongoDeps{
			Client:            client,
			Db:                dbName,
			ConnectionTimeout: connectionTimeout,
		},
		inmemory.Entity[*Product]{
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

	// Wait for initial load
	time.Sleep(2 * time.Second)
	logger.Info().Msg("In-memory projection ready")

	// Get cache with indexes
	cache := im.GetCacheWithEventListener()

	// Demonstrate Inverse Index (DP-023)
	logger.Info().Msg("=== Inverse Index Demonstration ===")
	parentID1 := "category-electronics"
	parentID2 := "category-books"

	product1 := &Product{
		ParentID: &parentID1,
		Email:    stringPtr("product1@example.com"),
		Title:    stringPtr("Laptop"),
	}
	id1, err := im.AwaitCreate(ctx, product1)
	if err != nil {
		log.Fatalf("Failed to create product1: %v", err)
	}
	logger.Info().Str("id", id1).Str("parent_id", parentID1).Msg("Created product1")

	product2 := &Product{
		ParentID: &parentID1,
		Email:    stringPtr("product2@example.com"),
		Title:    stringPtr("Phone"),
	}
	id2, err := im.AwaitCreate(ctx, product2)
	if err != nil {
		log.Fatalf("Failed to create product2: %v", err)
	}
	logger.Info().Str("id", id2).Str("parent_id", parentID1).Msg("Created product2")

	product3 := &Product{
		ParentID: &parentID2,
		Email:    stringPtr("product3@example.com"),
		Title:    stringPtr("Go Programming Book"),
	}
	id3, err := im.AwaitCreate(ctx, product3)
	if err != nil {
		log.Fatalf("Failed to create product3: %v", err)
	}
	logger.Info().Str("id", id3).Str("parent_id", parentID2).Msg("Created product3")

	// Use Inverse Index to find all products in category-electronics
	if idx, ok := cache.InverseIndexes["parent_id"]; ok {
		ids := idx.Get(ctx, &parentID1)
		logger.Info().
			Strs("product_ids", ids).
			Str("parent_id", parentID1).
			Msg("Inverse Index: Found products by parent_id")
		if len(ids) != 2 {
			log.Fatalf("Expected 2 products, got %d", len(ids))
		}
	}

	// Demonstrate Inverse Unique Index (DP-024)
	logger.Info().Msg("=== Inverse Unique Index Demonstration ===")
	uniqueEmail := "unique@example.com"
	product4 := &Product{
		ParentID: &parentID1,
		Email:    &uniqueEmail,
		Title:    stringPtr("Unique Product"),
	}
	id4, err := im.AwaitCreate(ctx, product4)
	if err != nil {
		log.Fatalf("Failed to create product4: %v", err)
	}
	logger.Info().Str("id", id4).Str("email", uniqueEmail).Msg("Created product4 with unique email")

	// Use Inverse Unique Index to find product by email
	if idx, ok := cache.InverseUniqueIndexes["email_unique"]; ok {
		foundID, found := idx.Get(ctx, uniqueEmail)
		if found {
			logger.Info().
				Str("product_id", foundID).
				Str("email", uniqueEmail).
				Msg("Inverse Unique Index: Found product by email")
			if foundID != id4 {
				log.Fatalf("Expected product ID %s, got %s", id4, foundID)
			}
		}
	}

	// Demonstrate Sorted Index (DP-025)
	logger.Info().Msg("=== Sorted Index Demonstration ===")
	// Create products with different titles for sorting
	title1 := "Alpha Product"
	title2 := "Beta Product"
	title3 := "Gamma Product"

	product5 := &Product{
		ParentID: &parentID1,
		Email:    stringPtr("sorted1@example.com"),
		Title:    &title2,
	}
	id5, err := im.AwaitCreate(ctx, product5)
	if err != nil {
		log.Fatalf("Failed to create product5: %v", err)
	}

	product6 := &Product{
		ParentID: &parentID1,
		Email:    stringPtr("sorted2@example.com"),
		Title:    &title1,
	}
	id6, err := im.AwaitCreate(ctx, product6)
	if err != nil {
		log.Fatalf("Failed to create product6: %v", err)
	}

	product7 := &Product{
		ParentID: &parentID1,
		Email:    stringPtr("sorted3@example.com"),
		Title:    &title3,
	}
	id7, err := im.AwaitCreate(ctx, product7)
	if err != nil {
		log.Fatalf("Failed to create product7: %v", err)
	}

	// Sorted Index maintains products in sorted order
	// Intersect can be used to find products matching multiple criteria
	if idx, ok := cache.SortedIndexes["title"]; ok {
		allIDs := []string{id5, id6, id7}
		intersected := idx.Intersect(allIDs)
		logger.Info().
			Strs("all_ids", allIDs).
			Strs("intersected", intersected).
			Msg("Sorted Index: Intersect operation")
	}

	// Demonstrate Suffix Index (DP-026)
	logger.Info().Msg("=== Suffix Index Demonstration ===")
	desc1 := "High-performance laptop for developers"
	desc2 := "Modern smartphone with advanced features"
	desc3 := "Book about Go programming language"

	product8 := &Product{
		ParentID:    &parentID1,
		Email:       stringPtr("search1@example.com"),
		Title:       stringPtr("Dev Laptop"),
		Description: &desc1,
	}
	_, err = im.AwaitCreate(ctx, product8)
	if err != nil {
		log.Fatalf("Failed to create product8: %v", err)
	}

	product9 := &Product{
		ParentID:    &parentID1,
		Email:       stringPtr("search2@example.com"),
		Title:       stringPtr("Smart Phone"),
		Description: &desc2,
	}
	_, err = im.AwaitCreate(ctx, product9)
	if err != nil {
		log.Fatalf("Failed to create product9: %v", err)
	}

	product10 := &Product{
		ParentID:    &parentID2,
		Email:       stringPtr("search3@example.com"),
		Title:       stringPtr("Go Book"),
		Description: &desc3,
	}
	id10, err := im.AwaitCreate(ctx, product10)
	if err != nil {
		log.Fatalf("Failed to create product10: %v", err)
	}

	// Use Suffix Index for text search
	if idx, ok := cache.SuffixIndexes["description"]; ok {
		// Search for "programming" - should find product10
		results := idx.Search(ctx, "programming")
		logger.Info().
			Strs("product_ids", results).
			Str("search_term", "programming").
			Msg("Suffix Index: Search results")
		if len(results) > 0 {
			found := false
			for _, id := range results {
				if id == id10 {
					found = true
					break
				}
			}
			if !found {
				log.Fatalf("Expected to find product10 in search results")
			}
		}

		// Find (trigram search)
		findResults := idx.Find(ctx, "laptop")
		logger.Info().
			Strs("product_ids", findResults).
			Str("find_term", "laptop").
			Msg("Suffix Index (description): Find results (trigram search)")
	}

	// Use Suffix Index for text search on title (multiple indexes on same field)
	if idx, ok := cache.SuffixIndexes["title"]; ok {
		// Search for "Book" - should find product10
		results := idx.Search(ctx, "Book")
		logger.Info().
			Strs("product_ids", results).
			Str("search_term", "Book").
			Msg("Suffix Index (title): Search results")
		if len(results) > 0 {
			found := false
			for _, id := range results {
				if id == id10 {
					found = true
					break
				}
			}
			if found {
				logger.Info().Msg("Suffix Index (title): Found product10 by title search")
			}
		}
	}

	logger.Info().Msg("=== All Index Types Demonstrated Successfully ===")
	logger.Info().Msg("1. Inverse Index: Find all products by parent_id")
	logger.Info().Msg("2. Inverse Unique Index: Find product by unique email")
	logger.Info().Msg("3. Sorted Index: Maintain sorted order, support intersection")
	logger.Info().Msg("4. Suffix Index: Text search using trigrams")
}

func stringPtr(s string) *string {
	return &s
}
