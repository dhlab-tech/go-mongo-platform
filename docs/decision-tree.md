# Decision Tree: Redis vs go-mongo-platform

**Status:** Public Documentation  
**Purpose:** Guide for choosing between Redis, distributed caches, and go-mongo-platform

---

## Quick Decision Tree

```
Do you need shared state across multiple services?
├─ YES → Use Redis or distributed cache
└─ NO → Continue

Do you need TTL-based eviction?
├─ YES → Use Redis or distributed cache
└─ NO → Continue

Do you need cross-language compatibility?
├─ YES → Use Redis or distributed cache
└─ NO → Continue

Are you building a Go service with MongoDB?
├─ NO → Use Redis or distributed cache
└─ YES → Continue

Do you need strong read-after-write consistency?
├─ NO → Consider Redis (if other requirements fit)
└─ YES → Use go-mongo-platform

Do you want to eliminate dual-write complexity?
├─ NO → Consider Redis (if other requirements fit)
└─ YES → Use go-mongo-platform
```

---

## Detailed Comparison

### Use go-mongo-platform When

**Architecture Requirements:**
- ✅ Single Go service (or multiple independent Go services)
- ✅ MongoDB as primary database
- ✅ Process-local in-memory state is acceptable
- ✅ No need for shared state across services

**Consistency Requirements:**
- ✅ Strong read-after-write consistency needed
- ✅ Deterministic behavior after writes
- ✅ No tolerance for stale reads
- ✅ Need to eliminate race conditions

**Operational Requirements:**
- ✅ Want to eliminate dual-write complexity
- ✅ Prefer MongoDB as single source of truth
- ✅ Want minimal operational overhead
- ✅ Can manage memory based on data size

**Performance Requirements:**
- ✅ Read-heavy workloads
- ✅ Need minimal latency for reads
- ✅ Complex in-memory filtering/sorting/indexing
- ✅ Network overhead for reads is unacceptable

### Use Redis When

**Architecture Requirements:**
- ✅ Need shared state across multiple services
- ✅ Multi-language microservice architecture
- ✅ Need pub/sub or messaging patterns
- ✅ Require distributed coordination primitives

**Consistency Requirements:**
- ✅ Eventual consistency is acceptable
- ✅ Can tolerate occasional stale reads
- ✅ Have robust cache invalidation logic

**Operational Requirements:**
- ✅ Need TTL-based eviction
- ✅ Require LRU or other eviction policies
- ✅ Need generic key-value storage
- ✅ Have Redis operational expertise

**Performance Requirements:**
- ✅ Network latency for reads is acceptable
- ✅ Need distributed caching across regions
- ✅ Require high-throughput key-value operations

### Use Distributed Cache When

**Architecture Requirements:**
- ✅ Need shared state across multiple services
- ✅ Require cross-region replication
- ✅ Need distributed coordination
- ✅ Multi-language support required

**Consistency Requirements:**
- ✅ Eventual consistency is acceptable
- ✅ Can tolerate cache coherence delays
- ✅ Have distributed consistency strategies

**Operational Requirements:**
- ✅ Have distributed systems expertise
- ✅ Can manage cache clusters
- ✅ Need high availability across regions
- ✅ Require sophisticated eviction policies

---

## Migration Considerations

### Migrating from Redis to go-mongo-platform

**When Migration Makes Sense:**
- You're experiencing stale data issues with Redis
- Dual-write complexity is causing bugs
- You want stronger consistency guarantees
- Your workload is primarily read-heavy
- You're using Go + MongoDB

**Migration Steps:**
1. Identify all Redis write paths
2. Remove Redis writes (keep only MongoDB writes)
3. Replace Redis reads with go-mongo-platform cache reads
4. Use Await operations for write-then-read patterns
5. Remove Redis invalidation logic
6. Test thoroughly for consistency

**Challenges:**
- Redis may be used for other purposes (pub/sub, rate limiting)
- Need to ensure all writes go through MongoDB
- May need to handle existing Redis data migration

### Migrating from go-mongo-platform to Redis

**When Migration Makes Sense:**
- You need shared state across services
- You require TTL-based eviction
- You need cross-language compatibility
- You're moving to a polyglot architecture

**Migration Steps:**
1. Implement dual-write pattern (MongoDB + Redis)
2. Replace cache reads with Redis reads
3. Implement cache invalidation logic
4. Add Redis operational infrastructure
5. Handle consistency trade-offs

**Challenges:**
- Lose strong consistency guarantees
- Need to implement invalidation logic
- Add operational complexity
- May introduce race conditions

---

## Trade-Off Analysis

### Consistency

| Solution | Consistency Model | Read-After-Write | Guarantees |
|----------|------------------|------------------|------------|
| go-mongo-platform | Strong (Await) | Guaranteed | Hard guarantee |
| Redis | Eventual | Not guaranteed | Best effort |
| Distributed Cache | Eventual | Not guaranteed | Best effort |

### Operational Complexity

| Solution | Setup Complexity | Maintenance | Infrastructure |
|----------|-----------------|-------------|----------------|
| go-mongo-platform | Low | Low | MongoDB only |
| Redis | Medium | Medium | Redis cluster |
| Distributed Cache | High | High | Cache cluster + coordination |

### Performance

| Solution | Read Latency | Write Latency | Network Overhead |
|----------|-------------|--------------|------------------|
| go-mongo-platform | Minimal (in-process) | Network (MongoDB) | None for reads |
| Redis | Network | Network | Yes |
| Distributed Cache | Network | Network | Yes |

### Use Case Fit

| Use Case | go-mongo-platform | Redis | Distributed Cache |
|----------|-------------------|-------|-------------------|
| Single Go service | ✅ Excellent | ⚠️ Overkill | ❌ Not needed |
| Multi-service shared state | ❌ Not supported | ✅ Good | ✅ Good |
| TTL-based eviction | ❌ Not supported | ✅ Good | ✅ Good |
| Strong consistency | ✅ Excellent | ❌ Not supported | ❌ Not supported |
| Cross-language | ❌ Not supported | ✅ Good | ✅ Good |

---

## Common Misconceptions

### "go-mongo-platform is a Redis replacement"

**False.** go-mongo-platform is not a Redis replacement. It serves a different purpose:
- Redis: Distributed cache with shared state
- go-mongo-platform: In-process projection with process-local state

### "I can use go-mongo-platform for shared state across services"

**False.** go-mongo-platform is process-local. Each service instance has its own projection. For shared state, use Redis or a distributed cache.

### "go-mongo-platform provides TTL-based eviction"

**False.** go-mongo-platform does not provide automatic eviction. Memory management is the application's responsibility.

### "go-mongo-platform is faster than Redis"

**Misleading comparison.** go-mongo-platform has lower read latency (in-process) but serves a different purpose. For shared state or TTL eviction, Redis is the appropriate choice.

---

## Decision Checklist

Before choosing go-mongo-platform, verify:

- [ ] I'm building a Go service
- [ ] I'm using MongoDB as the primary database
- [ ] I need strong read-after-write consistency
- [ ] I can accept process-local state (no shared state)
- [ ] I don't need TTL-based eviction
- [ ] I want to eliminate dual-write complexity
- [ ] I understand the operational model (in-process, resync on restart)

If all items are checked, go-mongo-platform may be a good fit.

---

## Related Documentation

- [positioning.md](positioning.md) — Canonical positioning statement
- [production-guide.md](production-guide.md) — Production deployment considerations
- [troubleshooting.md](troubleshooting.md) — Common issues and solutions

