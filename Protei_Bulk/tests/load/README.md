# Load Testing Guide

Load testing framework for Protei_Bulk to validate 10,000+ TPS performance.

## Prerequisites

```bash
pip install locust
```

## Running Load Tests

### Basic Load Test

```bash
# Run with web UI
locust -f locustfile.py --host=http://localhost:8080

# Open browser at http://localhost:8089
# Configure number of users and spawn rate
```

### Headless Mode (No UI)

```bash
# Run test directly without web UI
locust -f locustfile.py \
    --host=http://localhost:8080 \
    --users 5000 \
    --spawn-rate 100 \
    --run-time 10m \
    --headless

# With HTML report
locust -f locustfile.py \
    --host=http://localhost:8080 \
    --users 5000 \
    --spawn-rate 100 \
    --run-time 10m \
    --headless \
    --html report.html
```

### Gradual Ramp-Up Test

```bash
# Uses GradualRampUp shape class
locust -f locustfile.py \
    --host=http://localhost:8080 \
    --shape GradualRampUp
```

## Load Test Scenarios

### 1. Message Sending Test (10K TPS Target)

```bash
locust -f locustfile.py \
    --host=http://localhost:8080 \
    --users 10000 \
    --spawn-rate 500 \
    --run-time 30m
```

**Expected Results:**
- Target: 10,000+ requests/second
- Response time p95: <100ms
- Response time p99: <500ms
- Error rate: <1%

### 2. Sustained Load Test

```bash
locust -f locustfile.py \
    --host=http://localhost:8080 \
    --users 5000 \
    --spawn-rate 100 \
    --run-time 2h
```

### 3. Spike Test

```bash
# Sudden traffic spike
locust -f locustfile.py \
    --host=http://localhost:8080 \
    --users 15000 \
    --spawn-rate 1000 \
    --run-time 10m
```

## Test Configuration

### User Classes

1. **MessagingUser** - API-based message sending
   - Send single messages (weight: 10)
   - Send bulk messages (weight: 5)
   - Query message status (weight: 3)
   - List messages (weight: 2)
   - Create campaigns (weight: 1)

2. **SMPPUser** - SMPP protocol testing
   - SMPP submit operations

### Custom Metrics

The load tests track:
- Total requests
- Failures
- Response times (avg, median, p95, p99)
- Requests per second
- Error rates by endpoint

## Performance Targets

| Metric | Target | Acceptable |
|--------|--------|------------|
| TPS | 10,000+ | 5,000+ |
| Avg Response Time | <50ms | <100ms |
| P95 Response Time | <100ms | <200ms |
| P99 Response Time | <500ms | <1000ms |
| Error Rate | <0.1% | <1% |
| CPU Usage | <70% | <85% |
| Memory Usage | <80% | <90% |

## Distributed Load Testing

For testing beyond single machine capacity:

### Master Node

```bash
locust -f locustfile.py \
    --host=http://localhost:8080 \
    --master \
    --expect-workers 4
```

### Worker Nodes

```bash
# Run on each worker machine
locust -f locustfile.py \
    --host=http://localhost:8080 \
    --worker \
    --master-host=<master-ip>
```

## Monitoring During Tests

### System Metrics

```bash
# Monitor system resources
htop

# Monitor network
iftop

# Monitor PostgreSQL
pg_top

# Monitor Redis
redis-cli --stat
```

### Application Metrics

```bash
# View logs
tail -f ../logs/system.log

# Monitor API
watch -n 1 'curl -s http://localhost:8080/api/v1/health'
```

## Results Analysis

### Key Metrics to Check

1. **Throughput**: Requests/second achieved
2. **Response Time**: Check p50, p95, p99 percentiles
3. **Error Rate**: Should be <1% under load
4. **Resource Usage**: CPU, memory, disk I/O
5. **Database Performance**: Query times, connection pool
6. **Queue Depth**: Redis queue size

### Sample Report

```
============================================================
Protei_Bulk Load Test Complete
============================================================
Total Requests: 1,200,000
Total Failures: 50
Average Response Time: 45.32ms
Median Response Time: 38.21ms
95th Percentile: 98.44ms
99th Percentile: 287.31ms
Requests/sec: 10,523.45
============================================================
```

## Optimization Tips

If performance targets not met:

1. **Scale Horizontally**: Add more application nodes
2. **Database Tuning**: Optimize queries, increase connection pool
3. **Redis Tuning**: Increase max connections
4. **Queue Optimization**: Use Celery for async processing
5. **Connection Pooling**: Tune database connection limits
6. **Caching**: Add caching layer for frequent queries

## Troubleshooting

### High Error Rate

- Check database connection pool
- Verify SMSC connections
- Check disk I/O (especially for CDR writes)

### High Response Time

- Enable database query logging
- Check slow queries
- Verify network latency
- Check for lock contention

### Memory Issues

- Monitor for memory leaks
- Check garbage collection
- Verify connection cleanup

## References

- Locust Documentation: https://docs.locust.io/
- Performance Testing Best Practices
- Protei_Bulk Architecture Guide
