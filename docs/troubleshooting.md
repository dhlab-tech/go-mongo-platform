# Troubleshooting Guide

**Status:** Public Documentation  
**Purpose:** Common issues, solutions, and operational guidance

---

## Failure Scenarios

### Change Stream Disconnect

**Symptoms:**
- In-memory projection stops updating
- Writes succeed but reads show stale data
- Logs show Change Stream disconnection errors

**Causes:**
- Network connectivity issues
- MongoDB replica set member failure
- Change Stream timeout
- MongoDB server restart

**Solutions:**
1. Check MongoDB connectivity
2. Verify replica set status (`rs.status()`)
3. Check network connectivity between application and MongoDB
4. Review MongoDB logs for errors
5. The library will attempt automatic reconnection
6. If reconnection fails, restart the service (triggers full resync)

**Prevention:**
- Monitor MongoDB replica set health
- Use connection pooling appropriately
- Implement health checks for Change Stream status

---

### Process Crash / Restart

**Symptoms:**
- In-memory state is lost on restart
- Service takes time to become ready
- Initial queries may be slow

**Expected Behavior:**
- On restart, the projection is rebuilt by querying all documents from MongoDB
- This is expected behavior for in-process storage
- Service should not accept requests until rebuild completes

**Solutions:**
1. Implement readiness checks that wait for rebuild completion
2. Monitor rebuild duration
3. Consider data partitioning if rebuild takes too long
4. Use MongoDB indexes to speed up initial queries

**Prevention:**
- Design for eventual consistency during startup
- Implement proper health checks
- Monitor startup time

---

### MongoDB Unavailability

**Symptoms:**
- Writes fail immediately
- In-memory projection stops updating
- Error logs show MongoDB connection failures

**Expected Behavior:**
- Writes fail (no speculative updates)
- In-memory projection remains in last known state (may be stale)
- Projection does not update until MongoDB is available

**Solutions:**
1. Check MongoDB server status
2. Verify network connectivity
3. Check MongoDB replica set health
4. Review MongoDB logs
5. Wait for MongoDB to recover (automatic resync will occur)

**Prevention:**
- Monitor MongoDB health
- Use MongoDB replica sets for high availability
- Implement appropriate retry logic (application-level)

---

### Stale Data in Projection

**Symptoms:**
- Reads return outdated data
- Changes written to MongoDB not reflected in reads

**Causes:**
- Change Stream disconnect (projection not updating)
- Service restart (rebuild in progress)
- MongoDB unavailability (projection paused)

**Solutions:**
1. Check Change Stream status
2. Verify MongoDB connectivity
3. Check if service is in rebuild phase
4. Review logs for Change Stream errors
5. If persistent, restart service (triggers full resync)

**Prevention:**
- Monitor Change Stream health
- Implement health checks
- Use Await operations for write-then-read patterns

---

## Common Issues

### "Change Streams require a replica set"

**Error:**
```
Change Streams are only available with replica sets
```

**Solution:**
- Configure MongoDB as a replica set (minimum one node)
- Use connection string with `replicaSet` parameter
- See MongoDB documentation for replica set setup

### "Cache is not initialized"

**Error:**
```
cache is not initialized, AwaitCreate requires cache
```

**Cause:**
- In-memory projection not initialized
- Change Streams not configured
- Stream listener not set up

**Solution:**
1. Ensure MongoDB connection is established
2. Verify Change Streams are configured
3. Check that stream listener is properly initialized
4. Review initialization code

### High Memory Usage

**Symptoms:**
- Application memory usage grows over time
- OOM (Out of Memory) errors

**Causes:**
- Large number of documents in MongoDB
- No automatic eviction
- Memory leaks in application code

**Solutions:**
1. Monitor memory usage per instance
2. Consider data partitioning
3. Implement application-level memory limits
4. Review application code for memory leaks
5. Consider archiving old data in MongoDB

**Prevention:**
- Design data model with memory constraints in mind
- Monitor memory usage
- Implement data lifecycle management

### Slow Startup

**Symptoms:**
- Service takes long time to become ready
- Initial queries are slow

**Causes:**
- Large number of documents to rebuild
- Slow MongoDB queries
- Missing indexes

**Solutions:**
1. Add indexes to MongoDB collections
2. Monitor rebuild duration
3. Consider data partitioning
4. Optimize MongoDB queries
5. Implement pagination if supported

**Prevention:**
- Use appropriate MongoDB indexes
- Monitor collection sizes
- Design for startup performance

---

## Operational Guidance

### Health Checks

Implement health checks that verify:

1. **MongoDB Connectivity**
   - Test MongoDB connection
   - Verify replica set status

2. **Change Stream Status**
   - Check if Change Streams are active
   - Monitor for disconnection events

3. **Cache Readiness**
   - Verify projection is initialized
   - Check if rebuild is complete

4. **Memory Usage**
   - Monitor memory consumption
   - Alert on high memory usage

### Monitoring

**Key Metrics:**
- Change Stream disconnection count
- Reconnection attempts
- Cache rebuild duration
- Memory usage
- Await operation latency

**Logging:**
- Change Stream events
- Reconnection attempts
- Cache rebuild operations
- Error conditions

### Debugging

**Enable Debug Logging:**
- Use structured logging (zerolog)
- Enable MongoDB driver logging if needed
- Review Change Stream events

**Common Debugging Steps:**
1. Check MongoDB replica set status
2. Verify Change Stream configuration
3. Review application logs
4. Check network connectivity
5. Monitor memory usage

---

## FAQ

### Q: Why is my projection not updating?

**A:** Check Change Stream status. The projection updates via Change Streams. If Change Streams are disconnected, the projection will not update until reconnection.

### Q: Can I use this with a single-node MongoDB?

**A:** No. Change Streams require a replica set. Configure MongoDB as a replica set (minimum one node).

### Q: What happens if MongoDB goes down?

**A:** Writes fail immediately. The in-memory projection stops updating and remains in its last known state. When MongoDB recovers, the projection will resync.

### Q: How do I handle memory constraints?

**A:** Memory management is the application's responsibility. Consider data partitioning, archiving old data, or implementing application-level memory limits.

### Q: Can I share state across service instances?

**A:** No. The projection is process-local. Each instance has its own projection. For shared state, use Redis or a distributed cache.

### Q: How do I know when the projection is ready?

**A:** Implement a readiness check that verifies the projection is initialized and Change Streams are active. Do not accept requests until ready.

### Q: What is the consistency model?

**A:** Read-after-write consistency is guaranteed after Await operations. See [Product Contract](../.github/internal-docs/PRODUCT_CONTRACT.md) for details.

### Q: Can I use TTL-based eviction?

**A:** No. The library does not provide TTL-based eviction. Memory management is the application's responsibility.

---

## Getting Help

### Community Support

- **GitHub Issues** — Report bugs and request features
- **GitHub Discussions** — Ask questions and discuss usage patterns

### Commercial Support

For production support, architecture reviews, and integration assistance:

📩 Contact: **support@digital-heroes.tech**

See [SUPPORT.md](../SUPPORT.md) for details.

---

## Related Documentation

- [positioning.md](positioning.md) — Canonical positioning
- [decision-tree.md](decision-tree.md) — Decision guide
- [production-guide.md](production-guide.md) — Production deployment considerations

