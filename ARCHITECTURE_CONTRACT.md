# Architecture Contract (Normative)

**Status:** Public Normative Document  
**Version:** 1.0  
**Purpose:** Explicit architectural constraints and behavioral guarantees

> **If any documentation conflicts with this contract, this contract wins.**

---

## 1. Core Architectural Constraints

### 1.1 Single-Process Semantics

**MUST:** The in-memory projection is process-local and instance-scoped.

- Each service instance maintains its own independent in-memory projection
- There is no shared or distributed in-memory state
- Cross-process consistency is **NOT** provided

**MUST NOT:** Assume shared state across service instances.

---

### 1.2 MongoDB as Source of Truth

**MUST:** MongoDB is the **only** source of truth.

- All writes go to MongoDB only
- No data is ever written directly to in-memory structures
- All mutations originate in MongoDB

**MUST NOT:** Write to in-memory structures directly or maintain a second source of truth.

---

### 1.3 No Distributed Coordination

**MUST NOT:** Provide distributed coordination primitives.

- No cross-service synchronization
- No distributed locking
- No shared cache state
- No cross-instance coordination

**SHOULD NOT:** Use this library for distributed coordination use cases.

---

### 1.4 No Persistence of In-Memory State

**MUST:** In-memory state is ephemeral and process-local.

- In-memory state is lost on process restart
- No snapshots or write-ahead logs (WAL) are maintained
- No persistence of in-memory projection to disk

**MUST NOT:** Rely on in-memory state surviving process restarts.

**SHOULD:** Rebuild the projection from MongoDB on startup.

---

### 1.5 No Default Eviction

**MUST NOT:** Provide automatic eviction policies by default.

- No TTL-based eviction
- No LRU eviction
- No automatic memory management

**SHOULD:** Manage memory based on application requirements and data size.

---

### 1.6 No Exactly-Once Across Restarts

**MUST NOT:** Guarantee exactly-once delivery across process restarts.

- Duplicate events may occur after restart
- Missed events are possible under extreme failures
- Deduplication is handled by versioning logic, not by the library

**SHOULD:** Handle idempotency at the application level if required.

---

## 2. Consistency Guarantees

### 2.1 Read-After-Write Consistency (Guaranteed)

**MUST:** Provide read-after-write consistency after `Await*` operations.

- After `AwaitCreate` returns, the entity is available in the in-memory cache
- After `AwaitUpdate` returns, the update is reflected in the in-memory cache
- After `AwaitDelete` returns, the entity is removed from the in-memory cache

This is a **hard guarantee** within a single process.

---

### 2.2 Change Streams Synchronization

**MUST:** Synchronize the in-memory projection with MongoDB via Change Streams.

- Changes are processed in the order they occur in MongoDB
- Resume tokens are used when available for reconnection
- The projection automatically rebuilds on startup by querying MongoDB

**SHOULD:** Handle Change Stream disconnections gracefully (automatic reconnection).

---

### 2.3 Ordering Consistency

**MUST:** Maintain ordering consistent with MongoDB Change Streams.

- Events are processed in the order they occur in MongoDB
- No reordering of events within a single Change Stream

---

## 3. Explicitly NOT Guaranteed

### 3.1 Distributed Consistency

**MUST NOT:** Guarantee consistency across service instances.

- Each instance has its own projection
- No cross-instance synchronization
- No distributed consistency guarantees

---

### 3.2 Exactly-Once Delivery

**MUST NOT:** Guarantee exactly-once delivery across crashes.

- Duplicate events may occur
- Missed events are possible under extreme failures
- Application-level deduplication may be required

---

### 3.3 Zero-Latency Change Streams

**MUST NOT:** Guarantee zero-latency Change Streams under network partitions.

- Change Streams may be delayed during network issues
- The projection may become stale during MongoDB unavailability

---

### 3.4 Cross-Service Cache Coherence

**MUST NOT:** Provide cross-service cache coherence.

- No shared state across services
- No cache invalidation across instances
- Each service instance operates independently

---

### 3.5 TTL-Based Semantics

**MUST NOT:** Provide TTL-based eviction or expiration policies.

- No automatic expiration
- No TTL semantics
- Memory management is the application's responsibility

---

## 4. Responsibility Boundaries

| Concern | Responsibility |
|---------|---------------|
| Data correctness | MongoDB |
| Ordering | MongoDB Change Streams |
| In-memory rebuild | go-mongo-platform |
| Memory limits | Application |
| Horizontal coordination | Application / infrastructure |
| Distributed consistency | Application / infrastructure |

---

## 5. Contract Stability

This contract is considered **stable for v0.x**.

Breaking changes to architectural constraints **MUST** be:
- Documented explicitly
- Versioned appropriately
- Communicated in release notes

---

## Related Documentation

- [NON_GOALS.md](NON_GOALS.md) — Explicit non-goals
- [docs/anti-patterns.md](docs/anti-patterns.md) — Common anti-patterns to avoid
- [docs/production-guide.md](docs/production-guide.md) — Production deployment considerations

---

**End of Architecture Contract**

