# Production Guide

**Status:** Public Documentation  
**Purpose:** Production deployment considerations and operational guidance

---

## Change Streams Configuration

### Replica Set Requirement

MongoDB Change Streams require a replica set. Single-node MongoDB instances cannot use Change Streams.

**Minimum Configuration:**
- MongoDB replica set with at least one node
- Replica set name configured
- Change Streams enabled (default in MongoDB 3.6+)

### Connection String

Use a MongoDB connection string that includes all replica set members:

```
mongodb://host1:27017,host2:27017,host3:27017/dbname?replicaSet=rs0
```

### Change Stream Options

The library uses MongoDB Change Streams with default options. For production:

- **Resume tokens** are used automatically when available
- **Full document** is retrieved for update operations
- **Ordering** is guaranteed by MongoDB Change Streams

---

## Reconnection and Resilience Patterns

### Automatic Reconnection

The library attempts automatic reconnection when Change Streams disconnect:

1. **Resume token available** — Stream resumes from last processed event
2. **Resume token unavailable** — Full resync by querying all documents from MongoDB

### Reconnection Behavior

**On Stream Disconnect:**
- Automatic reconnect attempts
- Resume token used when possible
- Full resync if resume fails

**On Process Restart:**
- In-memory state is lost (expected for in-process storage)
- Full rebuild by querying all documents from MongoDB
- Change Streams started after rebuild completes

### Handling Reconnection Failures

If reconnection fails repeatedly:

1. Check MongoDB connectivity
2. Verify replica set status
3. Check network connectivity
4. Review MongoDB logs for errors
5. Consider implementing exponential backoff (application-level)

---

## Monitoring and Observability

### Key Metrics to Monitor

**MongoDB Metrics:**
- Replica set health
- Change Stream lag (if available)
- Connection pool usage
- Write latency

**Application Metrics:**
- In-memory cache size (number of documents)
- Memory usage
- Await operation latency
- Change Stream reconnection events

**Error Metrics:**
- Change Stream disconnection count
- Reconnection failures
- Write failures
- Cache rebuild duration

### Logging

The library uses `zerolog` for structured logging. Key log events:

- Change Stream disconnections
- Reconnection attempts
- Cache rebuild operations
- Write failures

### Health Checks

Implement health checks that verify:

1. MongoDB connectivity
2. Change Stream status
3. Cache readiness (after initial rebuild)
4. Memory usage (application-level)

---

## Performance Tuning Guidelines

### Memory Management

**Memory Considerations:**
- The projection grows with the number of documents in MongoDB
- No automatic eviction is provided
- Applications must manage memory based on data size

**Recommendations:**
- Monitor memory usage per instance
- Consider data partitioning if memory becomes a concern
- Implement application-level memory limits if needed

### Startup Performance

**Initial Load:**
- On startup, all documents are queried and loaded into memory
- Large collections may take time to rebuild
- Consider implementing readiness checks that wait for rebuild completion

**Optimization Strategies:**
- Use indexes on MongoDB for faster queries
- Consider pagination for very large collections (if supported)
- Monitor rebuild duration

### Read Performance

**In-Memory Reads:**
- Reads from the projection are in-process (minimal latency)
- No network overhead for cache hits
- Index lookups are fast (in-memory)

**MongoDB Reads:**
- Direct MongoDB queries (via Searcher) have network latency
- Use projection reads when possible for better performance

### Write Performance

**Write Latency:**
- Writes go to MongoDB (network latency)
- Await operations block until Change Streams propagate
- Change Stream latency depends on MongoDB replica set configuration

**Optimization Strategies:**
- Use appropriate MongoDB write concern
- Monitor Change Stream lag
- Consider write batching if applicable

---

## Deployment Considerations

### Horizontal Scaling

**Process-Local State:**
- Each service instance has its own in-memory projection
- No shared state across instances
- Each instance rebuilds its projection on startup

**Scaling Implications:**
- Memory usage scales with number of instances
- Each instance independently syncs with MongoDB
- No coordination overhead between instances

### High Availability

**Single Instance Failure:**
- Other instances continue operating independently
- Failed instance rebuilds on restart
- No impact on other instances

**MongoDB Failure:**
- Writes fail immediately
- In-memory projection stops updating
- Projection remains in last known state (may be stale)
- Automatic resync when MongoDB recovers

### Deployment Strategy

**Rolling Updates:**
- Each instance rebuilds its projection on restart
- Consider staging deployments to avoid simultaneous rebuilds
- Monitor memory usage during deployments

**Blue-Green Deployments:**
- New instances rebuild projections independently
- Old instances continue serving until new instances are ready
- No shared state to migrate

---

## Operational Best Practices

### Startup Sequence

1. Connect to MongoDB
2. Initialize in-memory projection (queries all documents)
3. Start Change Streams
4. Mark service as ready (after rebuild completes)

### Graceful Shutdown

1. Stop accepting new requests
2. Allow in-flight requests to complete
3. Close Change Streams
4. Shutdown MongoDB connections

### Error Handling

**Write Failures:**
- Handle MongoDB write errors appropriately
- Await operations will fail if writes fail
- Implement retry logic if needed (application-level)

**Change Stream Failures:**
- Library attempts automatic reconnection
- Monitor reconnection events
- Alert on repeated failures

### Data Consistency

**Read-After-Write:**
- Use Await operations for write-then-read patterns
- Do not rely on eventual consistency for critical paths
- Understand the consistency model (see Product Contract)

---

## Limitations and Constraints

### Memory Constraints

- No automatic eviction
- Memory grows with data size
- Applications must manage memory

### Process-Local State

- No shared state across instances
- Each instance has its own projection
- Cannot be used for cross-service coordination

### MongoDB Dependency

- Requires MongoDB replica set
- Dependent on MongoDB availability
- Change Streams must be available

### No TTL Support

- No automatic expiration
- No TTL-based eviction
- Applications must handle data lifecycle

---

## Troubleshooting

For common issues and solutions, see [troubleshooting.md](troubleshooting.md).

---

## Related Documentation

- [positioning.md](positioning.md) — Canonical positioning
- [decision-tree.md](decision-tree.md) — Decision guide
- [troubleshooting.md](troubleshooting.md) — Common issues

