# CRUD Example

This example demonstrates basic CRUD operations with `go-mongo-platform`, showing deterministic read-after-write consistency through Await semantics.

## What This Example Shows

- Setting up a MongoDB replica set with Docker Compose
- Creating an in-memory projection synchronized via Change Streams
- Using `AwaitCreate` for deterministic create operations
- Using `AwaitUpdate` for deterministic update operations
- Using `AwaitDelete` for deterministic delete operations
- Verifying read-after-write consistency

## Prerequisites

- Go 1.22 or later
- Docker and Docker Compose
- MongoDB replica set (provided via Docker Compose)

## Step-by-Step Instructions

### 1. Start MongoDB Replica Set

```bash
docker-compose up -d
```

Wait for the replica set to initialize (about 10-15 seconds). You can check the status:

```bash
docker-compose logs mongodb-setup
```

You should see: `Replica set initialized`

### 2. Verify MongoDB is Running

```bash
docker-compose ps
```

Both `mongodb-replica` and `mongodb-setup` should be running.

### 3. Run the Example

```bash
go run main.go
```

### 4. Expected Output

The example will:

1. Connect to MongoDB
2. Create a Change Stream
3. Initialize the in-memory projection
4. Create a user with `AwaitCreate`
5. Read the user from cache (demonstrating read-after-write consistency)
6. Update the user with `AwaitUpdate`
7. Verify the update in cache
8. Delete the user with `AwaitDelete`
9. Verify the deletion in cache

You should see output like:

```
{"level":"info","time":"2024-01-01T12:00:00Z","message":"Connected to MongoDB"}
{"level":"info","time":"2024-01-01T12:00:01Z","message":"In-memory projection ready"}
{"level":"info","time":"2024-01-01T12:00:01Z","message":"=== Creating a user ==="}
{"level":"info","time":"2024-01-01T12:00:02Z","id":"...","message":"User created"}
{"level":"info","time":"2024-01-01T12:00:02Z","id":"...","name":"Alice","email":"alice@example.com","message":"User read from cache (read-after-write consistency)"}
...
```

### 5. Clean Up

```bash
docker-compose down
```

## Understanding the Code

### Entity Definition

The `User` struct implements the required interface:

```go
type User struct {
    ID      primitive.ObjectID `bson:"_id"`
    Version *int64             `bson:"version"`
    Deleted *bool              `bson:"deleted"`
    Name    *string            `bson:"name"`
    Email   *string            `bson:"email"`
}
```

It implements:
- `ID() string` - Returns the document ID
- `Version() *int64` - Returns the version for optimistic locking
- `SetDeleted(bool)` - Sets the deleted flag

### Await Semantics

The `Await*` operations provide deterministic read-after-write consistency:

- `AwaitCreate` - Writes to MongoDB and blocks until the change is observed in memory
- `AwaitUpdate` - Updates MongoDB and blocks until the change is observed in memory
- `AwaitDelete` - Deletes from MongoDB and blocks until the change is observed in memory

After an `Await*` operation returns, subsequent reads from the in-memory cache will reflect the write. This is a **hard guarantee**.

### Change Streams

The example uses MongoDB Change Streams to synchronize the in-memory projection:

1. A Change Stream is created for the database
2. The stream listens for changes to all collections
3. Changes are processed and applied to the in-memory projection
4. The projection stays in sync with MongoDB automatically

## Troubleshooting

### "Change Streams require a replica set"

Ensure the replica set is initialized:

```bash
docker-compose logs mongodb-setup
```

If initialization failed, restart:

```bash
docker-compose restart mongodb-setup
```

### Connection refused

Ensure MongoDB is running:

```bash
docker-compose ps
```

If not running, start it:

```bash
docker-compose up -d
```

### "Failed to create Change Stream"

This usually means the replica set is not ready. Wait a few seconds and try again.

## Next Steps

- See [examples/indexes/](../indexes/) for index usage examples
- See [examples/listeners/](../listeners/) for event listener examples
- Read [docs/production-guide.md](../../docs/production-guide.md) for production considerations

