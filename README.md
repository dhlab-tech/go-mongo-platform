# go-mongo-platform

A MongoDB-native, strongly consistent in-memory projection layer for Go services. This library provides deterministic read-after-write consistency through MongoDB Change Streams, eliminating dual-write complexity and stale data issues common with external caches.

**What this is:** An in-process, strongly consistent projection layer that synchronizes with MongoDB via Change Streams. It provides typed domain models with rich in-memory indexes and Await semantics for deterministic read-after-write behavior.

**What this is NOT:** This is not a generic cache, not a Redis replacement, and not a distributed cache solution. It does not provide TTL-based eviction, cross-service synchronization, or shared state across instances.

---

## When to Use / When NOT to Use

### When to Use

- Backend service written in Go
- MongoDB used as the primary database
- Read-heavy workloads with complex in-memory filtering, sorting, or indexing
- Strict read-after-write consistency requirements
- Redis or external cache causes stale reads, invalidation bugs, or dual-write complexity
- You want MongoDB to remain the single source of truth

### When NOT to Use

This project is **not intended for**:

- **Beginners learning Go** — This library requires understanding of MongoDB Change Streams, Go generics, and consistency models
- **Teams looking for a Redis replacement** — This is not a distributed cache or key-value store
- **Systems requiring shared cache state** — The in-memory layer is process-local and instance-scoped
- **Data engineering / ETL pipelines** — This is designed for application-level consistency, not analytical workloads
- **Polyglot microservice architectures** — This is Go-specific
- **TTL-driven business logic** — No automatic eviction or expiration policies
- **Rate limiting or locking** — Not designed for coordination primitives
- **Multi-language environments** — Go-only SDK

If you need distributed cache semantics, cross-service synchronization, or TTL-based eviction, this library is not the right fit.

---

## Why Not Redis / Distributed Cache?

### The Problem with Redis

Traditional MongoDB + Redis architectures suffer from:

- **Dual-write complexity** — Applications must write to both MongoDB and Redis, increasing failure modes
- **Race conditions** — Writes to MongoDB and Redis can occur in different orders, causing inconsistencies
- **Stale data** — Cache invalidation logic is complex and error-prone, leading to stale reads
- **Unpredictable visibility** — There's no guarantee that a write will be immediately visible in the cache

### Why go-mongo-platform is Different

- **Single source of truth** — Only MongoDB is written to; the in-memory layer is a projection
- **Change Streams synchronization** — MongoDB Change Streams provide ordered, reliable updates
- **Await semantics** — Operations block until changes are observed in memory, ensuring deterministic read-after-write behavior
- **No invalidation logic** — The projection automatically stays in sync with MongoDB
- **Process-local** — Minimal latency with no network overhead for reads

### When to Choose Each Solution

**Choose go-mongo-platform when:**
- You need strong consistency guarantees
- Your workload is read-heavy with complex queries
- You want to eliminate dual-write complexity
- You're building a Go service with MongoDB

**Choose Redis or distributed cache when:**
- You need shared state across multiple services
- You require TTL-based eviction policies
- You need cross-language compatibility
- You're building a polyglot microservice architecture

For a detailed decision guide, see [docs/decision-tree.md](docs/decision-tree.md).

---

## Consistency Guarantees (Await Semantics)

### Read-After-Write Consistency

The library provides **Await semantics** through `AwaitCreate`, `AwaitUpdate`, and `AwaitDelete` operations. These operations:

1. Write to MongoDB
2. Block until the corresponding change is observed in the in-memory projection via Change Streams
3. Return only after the in-memory state reflects the write

This ensures that after an `Await*` operation returns, subsequent reads from the in-memory cache will reflect the write. This is a **hard guarantee**.

### Change Streams Synchronization

The in-memory projection is synchronized with MongoDB through Change Streams:

- Changes are processed in the order they occur in MongoDB
- Resume tokens are used when available for reconnection
- The projection automatically rebuilds on startup by querying MongoDB

### What is Guaranteed

- Read-after-write consistency after `Await*` calls
- Ordering consistent with MongoDB Change Streams
- Deterministic rebuild on restart
- No dual-write scenarios

### What is NOT Guaranteed

- Distributed consistency across service instances (each instance has its own projection)
- Exactly-once delivery across crashes (duplicate events may occur)
- Zero-latency Change Streams under network partitions
- Cross-service cache coherence
- TTL-based semantics or eviction policies

For detailed consistency and failure model documentation, see [docs/production-guide.md](docs/production-guide.md) and [docs/troubleshooting.md](docs/troubleshooting.md).

---

## Failure Modes (High-Level)

### Change Stream Disconnect

If the Change Stream disconnects:

- Automatic reconnect attempts are made
- Resume tokens are used when possible to continue from the last processed event
- If resume fails, a full resync is performed by querying MongoDB

### Process Crash

On process restart:

- In-memory state is lost (as expected for in-process storage)
- On startup, the projection is rebuilt by querying all documents from MongoDB
- Change Streams are resumed if possible, otherwise a new stream is started

### MongoDB Unavailability

If MongoDB becomes unavailable:

- Writes fail immediately (no speculative updates)
- In-memory updates pause until MongoDB is available
- The projection remains in its last known state (may be stale)

### Network Partitions

During network partitions:

- Writes to MongoDB will fail
- The in-memory projection will not update until connectivity is restored
- No incorrect state is introduced (the projection does not speculate)

For detailed troubleshooting and operational guidance, see [docs/troubleshooting.md](docs/troubleshooting.md).

---

## Operational Model (In-Process, Resync on Restart)

### In-Process Architecture

The in-memory projection is **process-local** and **instance-scoped**:

- Each service instance maintains its own in-memory projection
- There is no shared or distributed in-memory state
- Reads are served from local memory (minimal latency)
- Writes go to MongoDB and propagate via Change Streams

### Resync on Restart

On service restart:

1. The service connects to MongoDB
2. All documents are queried and loaded into the in-memory projection
3. Change Streams are started to receive ongoing updates
4. The service is ready to serve reads from the projection

This ensures that the projection always reflects the current state of MongoDB, even after restarts.

### Memory Management

- Memory limits are the responsibility of the application
- The projection grows with the number of documents in MongoDB
- No automatic eviction or TTL policies are provided
- Applications must manage memory based on their data size

For production deployment considerations, see [docs/production-guide.md](docs/production-guide.md).

---

## Quick Start

### Prerequisites

- Go 1.22 or later
- MongoDB replica set (required for Change Streams)
- Docker and Docker Compose (for local development)

### Minimal Example

See [examples/crud/](examples/crud/) for a complete working example with step-by-step instructions.

The example demonstrates:
- Setting up a MongoDB replica set with Docker Compose
- Creating a typed entity
- Using `AwaitCreate`, `AwaitUpdate`, and `AwaitDelete` for deterministic read-after-write behavior
- Reading from the in-memory cache

Run the example:

```bash
cd examples/crud
docker-compose up -d
go run main.go
```

For more examples, see:
- [examples/indexes/](examples/indexes/) — Demonstrates all four index types
- [examples/listeners/](examples/listeners/) — Shows Before/After listeners and Notify/AwaitNotify patterns

---

## Documentation

### Normative Docs / Project Constraints

These documents define architectural constraints and guardrails:

- [ARCHITECTURE_CONTRACT.md](ARCHITECTURE_CONTRACT.md) — Architectural constraints and behavioral guarantees (normative)
- [NON_GOALS.md](NON_GOALS.md) — Explicit non-goals and anti-audience (normative)
- [docs/anti-patterns.md](docs/anti-patterns.md) — Common anti-patterns to avoid
- [docs/decision-tree.md](docs/decision-tree.md) — Redis vs go-mongo-platform decision guide
- [docs/production-guide.md](docs/production-guide.md) — Production deployment considerations
- [API_STABILITY.md](API_STABILITY.md) — Public API boundaries and stability policy

### Additional Documentation

- [docs/positioning.md](docs/positioning.md) — Canonical positioning and differentiation
- [docs/troubleshooting.md](docs/troubleshooting.md) — Common issues and operational guidance

---

## Support & Commercial Help

### Community Support

- **GitHub Issues** — Report bugs and request features
- **GitHub Discussions** — Ask questions and discuss usage patterns
- Best-effort support, no SLA

### Commercial Support

We provide professional services for teams running `go-mongo-platform` in production:

- Architecture reviews
- Production readiness audits
- Performance tuning
- Integration assistance
- Long-term support agreements

📩 Contact: **support@digital-heroes.tech**

For more details, see [SUPPORT.md](SUPPORT.md).

---

## License

Apache-2.0 — See [LICENSE](LICENSE) for details.
