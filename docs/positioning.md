# Positioning: go-mongo-platform

**Status:** Public Documentation  
**Purpose:** Canonical positioning statement and differentiation

---

## Core Positioning

`go-mongo-platform` is a **MongoDB-native, strongly consistent in-memory projection layer for Go services**. It is not a generic cache, not a Redis replacement, and not a distributed cache solution.

---

## What This Is

### Architectural Pattern

This library implements an **in-process projection pattern**:

- MongoDB is the **single source of truth**
- All writes go to MongoDB only
- The in-memory layer is a **projection** that synchronizes with MongoDB via Change Streams
- Each service instance maintains its own process-local projection
- The projection provides deterministic read-after-write consistency through Await semantics

### Key Characteristics

- **MongoDB-native** — Designed specifically for MongoDB Change Streams
- **Strongly consistent** — Read-after-write guarantees after Await operations
- **In-process** — No network overhead for reads, minimal latency
- **Typed domain models** — Go generics provide type safety
- **Rich in-memory indexes** — Inverse, Unique, Sorted, and Suffix indexes
- **Await semantics** — Operations block until changes are observed in memory

---

## What This Is NOT

### Not a Generic Cache

This is not a general-purpose caching solution. It does not provide:

- TTL-based eviction
- LRU or other eviction policies
- Generic key-value storage
- Cache warming strategies

### Not a Redis Replacement

This is fundamentally different from Redis:

- **No distributed state** — Each instance has its own projection
- **No shared cache** — Cannot be used for cross-service coordination
- **No pub/sub** — Not designed for messaging patterns
- **No persistence** — In-memory state is lost on restart (rebuilds from MongoDB)

### Not a Distributed Cache

This does not provide:

- Cross-service synchronization
- Shared state across instances
- Distributed coordination primitives
- Multi-region replication

### Not a CDC Pipeline

While it uses Change Streams (similar to CDC), it is not:

- A data pipeline tool (like Debezium/Kafka)
- An ETL solution
- An analytical data warehouse sync tool
- A multi-database replication system

---

## Target Use Cases

### Primary Scenario

Backend services that:

- Use Go as the primary language
- Use MongoDB as the primary database
- Have read-heavy workloads
- Require complex in-memory filtering, sorting, or indexing
- Need strict read-after-write consistency
- Have experienced issues with Redis (stale reads, invalidation bugs, dual-write complexity)

### Typical Problem Statement

> "We write to MongoDB, but reads must be fast and consistent. Redis caused subtle race conditions and stale data bugs. We want MongoDB to remain the source of truth."

---

## Why Existing Solutions Fail Here

### Redis

- Introduces dual-write complexity (must write to both MongoDB and Redis)
- Creates race conditions (writes can occur in different orders)
- Causes stale data (cache invalidation logic is error-prone)
- Provides unpredictable visibility (no guarantee of immediate cache updates)

### TTL-Based Caches

- TTL-based invalidation is incompatible with domain invariants
- Cannot guarantee consistency for time-sensitive operations
- Requires complex invalidation logic

### CDC Pipelines (Kafka/Debezium)

- Too heavy for application-level consistency needs
- Adds operational complexity (Kafka clusters, connectors)
- Designed for data pipelines, not application consistency

### Distributed Caches

- Add operational and conceptual complexity
- Require coordination across services
- Introduce network latency for reads
- Create consistency challenges across regions

---

## Why go-mongo-platform Fits

### In-Process Memory

- Minimal latency (no network overhead for reads)
- Process-local (no coordination overhead)
- Type-safe (Go generics)

### Change Streams

- Ordered updates (consistent with MongoDB ordering)
- Reliable synchronization (resume tokens for reconnection)
- Native MongoDB feature (no additional infrastructure)

### Await Semantics

- Deterministic behavior (operations block until observed)
- Read-after-write guarantees (hard guarantee)
- Eliminates race conditions

### Typed Domain Model

- Fewer logic errors (compile-time type checking)
- Rich indexing (Inverse, Unique, Sorted, Suffix)
- Domain-driven design support

---

## Differentiation Summary

| Feature | go-mongo-platform | Redis | Distributed Cache | CDC Pipeline |
|---------|-------------------|-------|-------------------|-------------|
| Source of Truth | MongoDB only | Dual-write | Dual-write | Source DB |
| Consistency | Strong (Await) | Eventual | Eventual | Eventual |
| State Sharing | No (process-local) | Yes | Yes | No |
| Network Overhead | None (in-process) | Yes | Yes | Yes |
| Operational Complexity | Low | Medium | High | High |
| Use Case | App consistency | General cache | Shared cache | Data pipeline |

---

## When to Choose go-mongo-platform

Choose this library when:

- ✅ You need strong consistency guarantees
- ✅ Your workload is read-heavy with complex queries
- ✅ You want to eliminate dual-write complexity
- ✅ You're building a Go service with MongoDB
- ✅ You need deterministic read-after-write behavior
- ✅ You want MongoDB to remain the single source of truth

---

## When NOT to Choose go-mongo-platform

Do not choose this library when:

- ❌ You need shared state across multiple services
- ❌ You require TTL-based eviction policies
- ❌ You need cross-language compatibility
- ❌ You're building a polyglot microservice architecture
- ❌ You need distributed coordination primitives
- ❌ You require cross-region replication

For a detailed decision guide, see [decision-tree.md](decision-tree.md).

---

## Related Documentation

- [decision-tree.md](decision-tree.md) — Detailed decision guide
- [production-guide.md](production-guide.md) — Production deployment considerations
- [troubleshooting.md](troubleshooting.md) — Common issues and solutions

