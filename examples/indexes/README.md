# Indexes Example

This example demonstrates all four index types available in `go-mongo-platform`:

- **Inverse Index** — Maps field values to document IDs (one-to-many)
- **Inverse Unique Index** — Maps field values to a single document ID (one-to-one, unique)
- **Sorted Index** — Maintains documents in sorted order, supports intersection
- **Suffix Index** — Provides text search capabilities using trigrams

## What This Example Shows

- How to define indexes using struct tags
- How to access indexes from the cache
- How to use each index type for different query patterns
- Practical examples of each index type

## Prerequisites

- Go 1.22 or later
- Docker and Docker Compose
- MongoDB replica set (provided via Docker Compose)

## Step-by-Step Instructions

### 1. Start MongoDB Replica Set

Use the same docker-compose setup as the CRUD example:

```bash
cd ../crud
docker-compose up -d
```

Wait for the replica set to initialize (about 10-15 seconds). You can check the status:

```bash
docker-compose logs mongodb-setup
```

You should see: `Replica set initialized`

### 2. Run the Example

```bash
go run main.go
```

### 3. Expected Output

The example will demonstrate:

1. **Inverse Index**: Create products with different parent IDs, then find all products by parent ID
2. **Inverse Unique Index**: Create a product with a unique email, then find the product by email
3. **Sorted Index**: Create products with different titles, then demonstrate intersection operations
4. **Suffix Index**: Create products with descriptions, then search for text using trigrams

You should see output like:

```
{"level":"info","time":"2024-01-01T12:00:00Z","message":"Connected to MongoDB"}
{"level":"info","time":"2024-01-01T12:00:01Z","message":"In-memory projection ready"}
{"level":"info","time":"2024-01-01T12:00:01Z","message":"=== Inverse Index Demonstration ==="}
...
```

### 4. Clean Up

```bash
cd ../crud
docker-compose down
```

## Understanding the Code

### Index Definition

Indexes are defined using struct tags:

```go
type Product struct {
    Id      primitive.ObjectID `bson:"_id"`
    V       *int64             `bson:"version"`
    Deleted *bool              `bson:"deleted"`
    
    // Inverse Index: find all products by parent category
    ParentID *string `bson:"parent_id" indexes:"inverse:parent_id:from"`
    
    // Inverse Unique Index: find product by unique email
    Email *string `bson:"email" indexes:"inverse_unique:email_unique:from"`
    
    // Sorted Index: maintain products sorted by title
    Title *string `bson:"title" indexes:"sorted:title:from"`
    
    // Suffix Index: text search on description
    Description *string `bson:"description" indexes:"suffix:description:from"`
}
```

### Index Access

Indexes are automatically created by the library and accessed through the cache:

```go
cache := im.GetCacheWithEventListener()

// Inverse Index
if idx, ok := cache.InverseIndexes["parent_id"]; ok {
    ids := idx.Get(ctx, &parentID)
}

// Inverse Unique Index
if idx, ok := cache.InverseUniqueIndexes["email_unique"]; ok {
    id, found := idx.Get(ctx, email)
}

// Sorted Index
if idx, ok := cache.SortedIndexes["title"]; ok {
    intersected := idx.Intersect(allIDs)
}

// Suffix Index
if idx, ok := cache.SuffixIndexes["description"]; ok {
    results := idx.Search(ctx, "search term")
    results := idx.Find(ctx, "find term")
}
```

## Index Types Explained

### Inverse Index

**Use Case**: Find all documents with a specific field value (one-to-many relationship).

**Example**: Find all products in a specific category.

**Tag Format**: `indexes:"inverse:index_name:from"`

**Access**: `cache.InverseIndexes["index_name"].Get(ctx, &value)`

**Returns**: `[]string` (array of document IDs)

### Inverse Unique Index

**Use Case**: Find a single document by a unique field value (one-to-one relationship).

**Example**: Find a product by unique email address.

**Tag Format**: `indexes:"inverse_unique:index_name:from"`

**Access**: `cache.InverseUniqueIndexes["index_name"].Get(ctx, value)`

**Returns**: `(string, bool)` (document ID and found flag)

### Sorted Index

**Use Case**: Maintain documents in sorted order, support intersection operations.

**Example**: Find products matching multiple sorted criteria.

**Tag Format**: `indexes:"sorted:index_name:from"`

**Access**: `cache.SortedIndexes["index_name"].Intersect([]string)`

**Returns**: `[]string` (intersected document IDs)

### Suffix Index

**Use Case**: Text search using trigrams (three-character sequences).

**Example**: Search for products by description text.

**Tag Format**: `indexes:"suffix:index_name:from"`

**Access**:
- `cache.SuffixIndexes["index_name"].Search(ctx, text)` — Exact matching
- `cache.SuffixIndexes["index_name"].Find(ctx, text)` — Trigram search (sorted by frequency)

**Returns**: `[]string` (array of document IDs)

## Multiple Indexes on Same Field

A field can have multiple indexes:

```go
Title *string `bson:"title" indexes:"sorted:title:from,suffix:title:from"`
```

This creates both a Sorted Index and a Suffix Index on the `Title` field.

## Troubleshooting

### "Index not found"

Ensure the index name in the tag matches the name used to access it:
- Tag: `indexes:"inverse:parent_id:from"` → Access: `cache.InverseIndexes["parent_id"]`
- Tag: `indexes:"inverse_unique:email_unique:from"` → Access: `cache.InverseUniqueIndexes["email_unique"]`

### "Change Streams require a replica set"

Ensure MongoDB replica set is initialized (see Step 1).

### Connection refused

Ensure MongoDB is running:

```bash
cd ../crud
docker-compose ps
```

## Related Documentation

- [docs/production-guide.md](../../docs/production-guide.md) — Production considerations
- [docs/troubleshooting.md](../../docs/troubleshooting.md) — Common issues
- [examples/crud/](../crud/) — Basic CRUD operations
