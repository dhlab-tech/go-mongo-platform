# Non-Goals (Normative)

**Status:** Public Normative Document  
**Version:** 1.0  
**Purpose:** Explicitly prevent misalignment and misuse

> **If any documentation conflicts with this document, this document wins.**

---

## Non-Goals

The following are explicitly **out of scope** for this project:

- **Distributed cache semantics** — No shared state across services
- **Cross-service synchronization** — No coordination across instances
- **TTL-based eviction** — No automatic expiration policies
- **Automatic memory management** — No LRU or other eviction policies
- **Generic key-value storage** — Not a general-purpose cache
- **Cross-language SDKs** — Go-only implementation
- **Beginner-friendly abstractions** — Requires understanding of MongoDB Change Streams and Go generics
- **Snapshots or WAL (Write-Ahead Log)** — No persistence of in-memory state
- **Exactly-once delivery across restarts** — Duplicate events may occur
- **Distributed coordination primitives** — No locking, no shared state
- **Second source of truth** — MongoDB is the only source of truth
- **Default eviction policies** — Memory management is application responsibility

---

## Anti-Audience

This project is **NOT intended for**:

- **Beginners learning Go** — Requires understanding of MongoDB Change Streams, Go generics, and consistency models
- **Teams looking for a Redis replacement** — This is not a distributed cache or key-value store
- **Systems requiring shared cache state** — The in-memory layer is process-local and instance-scoped
- **Data engineering / ETL pipelines** — Designed for application-level consistency, not analytical workloads
- **Polyglot microservice architectures** — Go-specific implementation
- **Use cases requiring TTL-based eviction** — No automatic expiration provided
- **Systems requiring distributed coordination** — No cross-service synchronization

---

## Expected User Profile

The expected user:

- Understands MongoDB Change Streams
- Is comfortable with Go generics
- Thinks in terms of consistency models
- Is responsible for production systems
- Needs strong read-after-write consistency within a single process
- Can accept process-local state (no shared state across services)

---

## When to Use Alternatives

**If you need shared state across services** — Use Redis or a distributed cache outside this project.

**If you need TTL-based eviction** — Use Redis or a cache with TTL support.

**If you need cross-language compatibility** — Use Redis or a language-agnostic solution.

**If you need distributed coordination** — Use dedicated coordination primitives (e.g., etcd, Consul).

**If you need exactly-once delivery across restarts** — Implement application-level deduplication.

---

## Why This Matters

Explicitly excluding use cases and users:

- Reduces incorrect expectations
- Increases trust among senior engineers
- Lowers support and issue noise
- Prevents architectural misalignment

---

## Related Documentation

- [ARCHITECTURE_CONTRACT.md](ARCHITECTURE_CONTRACT.md) — Architectural constraints
- [docs/anti-patterns.md](docs/anti-patterns.md) — Common anti-patterns to avoid
- [docs/decision-tree.md](docs/decision-tree.md) — Decision guide

---

**End of Non-Goals Document**

