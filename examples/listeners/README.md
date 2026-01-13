# Event Listeners Example

This example demonstrates event listeners in `go-mongo-platform`:

- **Before Listeners** — Execute before cache updates (validation, pre-processing)
- **After Listeners** — Execute after cache updates (side effects, notifications)
- **Notify/AwaitNotify** — Wait for specific cache update events (used internally by `Await*` operations)

## What This Example Shows

- How to create Before and After listeners using callback listeners
- How to register listeners with the in-memory projection
- The execution order of listeners (before → cache update → after)
- How Notify/AwaitNotify patterns ensure read-after-write consistency

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

1. **Before Listeners**: Execute before cache updates (e.g., validation, pre-processing)
2. **After Listeners**: Execute after cache updates (e.g., side effects, notifications)
3. **Notify/AwaitNotify**: Used internally by `Await*` operations to ensure read-after-write consistency

You should see output showing the execution order:

```
{"level":"info","time":"2024-01-01T12:00:00Z","message":"Connected to MongoDB"}
{"level":"info","time":"2024-01-01T12:00:01Z","message":"In-memory projection ready"}
{"level":"info","time":"2024-01-01T12:00:01Z","message":"=== Before/After Listeners Demonstration ==="}
{"level":"info","time":"2024-01-01T12:00:01Z","id":"...","name":"Alice","message":"Before Add Listener: Validating user"}
{"level":"info","time":"2024-01-01T12:00:02Z","id":"...","message":"After Add Listener: User added to cache"}
{"level":"info","time":"2024-01-01T12:00:02Z","execution_order":["before:add","after:add"],"message":"Listener execution order for Create"}
...
```

### 4. Clean Up

```bash
cd ../crud
docker-compose down
```

## Understanding the Code

### Listener Types

#### Before Listeners

Execute **before** the cache is updated. Useful for:
- Validation
- Pre-processing
- Logging

**Example**:

```go
beforeAddListener := inmemory.NewAddCallbackListener[*User](func(ctx context.Context, u *User) {
    logger.Info().Str("id", u.ID()).Msg("Before Add Listener: Validating user")
})
```

#### After Listeners

Execute **after** the cache is updated. Useful for:
- Side effects
- Notifications
- Post-processing

**Example**:

```go
afterAddListener := inmemory.NewAddCallbackListener[*User](func(ctx context.Context, u *User) {
    logger.Info().Str("id", u.ID()).Msg("After Add Listener: User added to cache")
})
```

### Registering Listeners

Listeners are registered when creating the in-memory projection:

```go
im, err := inmemory.NewInMemory[*User](
    ctx,
    stream,
    inmemory.MongoDeps{...},
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
```

### Execution Order

Listeners execute in a specific order:

1. **Before Listeners** (all before listeners in order)
2. **Cache Update** (cache is updated)
3. **After Listeners** (all after listeners in order)

For example, when creating a user:
1. `before:add` listener executes
2. Cache is updated with the new user
3. `after:add` listener executes

### Notify/AwaitNotify Pattern

The `Notify/AwaitNotify` pattern is used internally by `Await*` operations to ensure read-after-write consistency:

1. **Subscribe to notify** — `AwaitNotify.AddListenerCreate(id, callback)`
2. **Write to MongoDB** — `Processor.Create()`
3. **Wait for Change Stream** — Change Stream propagates the change
4. **Notify triggers** — `Notify.Add()` is called, callback executes
5. **Await* returns** — Operation completes, subscription is automatically removed

This ensures that after `AwaitCreate` returns, the entity is immediately available in the cache.

**Example**:

```go
id, err := im.AwaitCreate(ctx, user)
// After AwaitCreate returns, user is guaranteed to be in cache
cachedUser, found := cache.Cache.Get(ctx, id)
// found == true (guaranteed)
```

### Callback Listeners

The library provides callback listeners for convenience:

- `NewAddCallbackListener` — Called on Add operations
- `NewUpdateCallbackListener` — Called on Update operations
- `NewDeleteCallbackListener` — Called on Delete operations

**Add Callback Listener**:

```go
listener := inmemory.NewAddCallbackListener[*User](func(ctx context.Context, u *User) {
    // Handle Add event
})
```

**Update Callback Listener**:

```go
listener := inmemory.NewUpdateCallbackListener[*User](func(ctx context.Context, id string, u *User, removedFields []string) {
    // Handle Update event
})
```

**Delete Callback Listener**:

```go
listener := inmemory.NewDeleteCallbackListener[*User](func(ctx context.Context, id string) {
    // Handle Delete event
})
```

## Use Cases

### Before Listeners

- **Validation**: Validate data before it's added to cache
- **Pre-processing**: Transform or enrich data before caching
- **Logging**: Log operations before they complete

### After Listeners

- **Notifications**: Send notifications when data changes
- **Side Effects**: Trigger external systems when data changes
- **Analytics**: Track cache updates for analytics

### Notify/AwaitNotify

- **Read-after-write consistency**: Ensure data is available immediately after write
- **Synchronous operations**: Wait for cache updates before proceeding
- **Deterministic behavior**: Guarantee cache state after operations

## Troubleshooting

### "Listeners not executing"

Ensure listeners are registered correctly in `Entity[T]` when creating `NewInMemory`.

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
