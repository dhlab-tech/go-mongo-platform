# Anti-Patterns

**Status:** Public Documentation  
**Purpose:** Common mistakes and how to avoid them

This document describes anti-patterns when using `go-mongo-platform` and how to avoid them.

---

## 1. Redis Inside the Project

**Smell:** Using Redis alongside go-mongo-platform for the same data.

**Why Bad:**
- Creates dual-write complexity
- Introduces race conditions
- Causes stale data issues
- Violates single source of truth principle

**What to Do Instead:**
- Use go-mongo-platform as the only in-memory layer for MongoDB data
- If you need shared state across services, use Redis for that specific use case (outside go-mongo-platform)
- Keep MongoDB as the single source of truth

---

## 2. Multi-Instance Assumptions

**Smell:** Assuming shared state or consistency across service instances.

**Why Bad:**
- go-mongo-platform is process-local
- Each instance has its own independent projection
- No cross-instance synchronization is provided

**What to Do Instead:**
- Design for process-local state
- If you need shared state, use external coordination (Redis, etcd)
- Accept that each instance may have slightly different views during updates

---

## 3. Snapshots or WAL

**Smell:** Attempting to persist in-memory state to disk (snapshots, write-ahead logs).

**Why Bad:**
- In-memory state is ephemeral by design
- Persistence adds complexity and failure modes
- The projection rebuilds from MongoDB on startup

**What to Do Instead:**
- Accept that in-memory state is lost on restart
- Rely on MongoDB as the source of truth
- Let the projection rebuild on startup (it queries MongoDB)

---

## 4. Default Eviction Policies

**Smell:** Expecting automatic TTL or LRU eviction.

**Why Bad:**
- No automatic eviction is provided
- Memory management is the application's responsibility
- TTL-based eviction may violate domain invariants

**What to Do Instead:**
- Monitor memory usage
- Implement application-level memory limits if needed
- Consider data partitioning or archiving old data in MongoDB
- If you need TTL, use Redis for that specific use case

---

## 5. Hiding Eventual Consistency Behind Await

**Smell:** Using `Await*` operations but assuming distributed consistency.

**Why Bad:**
- `Await*` guarantees consistency only within a single process
- Does not provide distributed consistency across instances
- Misleading expectations about consistency scope

**What to Do Instead:**
- Understand that `Await*` provides process-local consistency
- Accept eventual consistency across instances
- Design for process-local guarantees

---

## 6. Global Mutex for Coordination

**Smell:** Using global mutexes or locks for cross-instance coordination.

**Why Bad:**
- go-mongo-platform is process-local
- Global mutexes don't work across instances
- Adds unnecessary complexity

**What to Do Instead:**
- Use external coordination primitives (Redis, etcd) if needed
- Design for process-local operations
- Accept independent operation of each instance

---

## 7. Retries That Reorder Events

**Smell:** Implementing retry logic that may reorder Change Stream events.

**Why Bad:**
- Change Streams provide ordered events
- Reordering breaks consistency guarantees
- May cause incorrect state

**What to Do Instead:**
- Preserve event ordering in retry logic
- Handle failures without reordering
- Let Change Streams handle ordering

---

## 8. Second Source of Truth

**Smell:** Maintaining a second source of truth alongside MongoDB.

**Why Bad:**
- MongoDB is the only source of truth
- Dual sources create consistency issues
- Violates architectural contract

**What to Do Instead:**
- Use MongoDB as the single source of truth
- In-memory projection is a read-optimized view
- All writes go to MongoDB only

---

## 9. Assuming Exactly-Once Across Restarts

**Smell:** Expecting exactly-once delivery of events across process restarts.

**Why Bad:**
- Duplicate events may occur after restart
- Missed events are possible under extreme failures
- Not guaranteed by the library

**What to Do Instead:**
- Implement idempotency at the application level if needed
- Use versioning for deduplication
- Handle duplicate events gracefully

---

## 10. Using for Distributed Coordination

**Smell:** Using go-mongo-platform for distributed locking or coordination.

**Why Bad:**
- No distributed coordination primitives provided
- Process-local semantics don't support cross-instance coordination
- Wrong tool for the job

**What to Do Instead:**
- Use dedicated coordination primitives (etcd, Consul, Redis)
- Keep go-mongo-platform for in-memory projection only
- Separate concerns: projection vs. coordination

---

## Related Documentation

- [ARCHITECTURE_CONTRACT.md](../ARCHITECTURE_CONTRACT.md) — Architectural constraints
- [NON_GOALS.md](../NON_GOALS.md) — Explicit non-goals
- [docs/decision-tree.md](decision-tree.md) — Decision guide
- [docs/production-guide.md](production-guide.md) — Production considerations

---

**End of Anti-Patterns Document**

