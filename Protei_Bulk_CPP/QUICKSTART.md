# Protei_Bulk C++ - Quick Start Guide

Get up and running with Protei_Bulk C++ in 5 minutes!

## Prerequisites

```bash
# Ubuntu/Debian
sudo apt-get update
sudo apt-get install -y build-essential cmake git \
    libboost-all-dev libpqxx-dev libhiredis-dev \
    libssl-dev zlib1g-dev postgresql-client libpq-dev
```

## Install redis-plus-plus

```bash
git clone https://github.com/sewenew/redis-plus-plus.git
cd redis-plus-plus
mkdir build && cd build
cmake ..
make -j$(nproc)
sudo make install
sudo ldconfig
cd ../..
```

## Build the Application

```bash
# Clone the repository
git clone <repository-url>
cd Protei_Bulk_CPP

# Build (Release mode, optimized)
./build.sh

# Or build with tests
./build.sh --test

# Or debug build
./build.sh --debug
```

## Quick Run (Local Development)

### Option 1: Using Docker (Recommended)

```bash
# Start PostgreSQL and Redis
cd docker
docker-compose up -d postgres redis

# Run the application locally
cd ../build
./bin/protei_bulk
```

### Option 2: Full Docker Stack

```bash
# Build and run everything in Docker
cd docker
docker-compose build
docker-compose up -d

# View logs
docker logs -f protei_bulk_cpp_app

# Check health
curl http://localhost:8081/api/v1/health
```

### Option 3: System Services

```bash
# Install PostgreSQL and Redis
sudo apt-get install -y postgresql redis-server

# Start services
sudo systemctl start postgresql redis-server

# Initialize database
sudo -u postgres createuser -s protei
sudo -u postgres createdb -O protei protei_bulk
sudo -u postgres psql -U protei -d protei_bulk < ../Protei_Bulk/database/schema.sql

# Run application
cd build
./bin/protei_bulk
```

## Verify Installation

```bash
# Health check
curl http://localhost:8080/api/v1/health

# Expected response:
# {
#   "status": "healthy",
#   "version": "1.0.0",
#   "timestamp": 1705747200
# }

# Test authentication
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"Admin@123"}'

# Send test message
curl -X POST http://localhost:8080/api/v1/messages/send \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "msisdn": "966500000000",
    "message": "Hello from Protei_Bulk C++!"
  }'
```

## Performance Testing

```bash
# Simple load test with curl
for i in {1..1000}; do
    curl -s http://localhost:8080/api/v1/health > /dev/null &
done
wait

# Using Apache Bench
ab -n 10000 -c 100 http://localhost:8080/api/v1/health

# Using wrk (if installed)
wrk -t4 -c100 -d30s http://localhost:8080/api/v1/health
```

## Configuration

Edit configuration files in `config/`:

```bash
# Main application config
vim config/app.conf

# Database config
vim config/db.conf

# Protocol config (HTTP, SMPP)
vim config/protocol.conf
```

## Environment Variables

Override config with environment variables:

```bash
export DB_HOST=localhost
export DB_PORT=5432
export DB_NAME=protei_bulk
export DB_USER=protei
export DB_PASSWORD=elephant
export REDIS_HOST=localhost
export REDIS_PORT=6379

./build/bin/protei_bulk
```

## Monitoring

```bash
# View logs
tail -f /opt/protei_bulk/logs/protei_bulk.log

# Docker logs
docker logs -f protei_bulk_cpp_app

# Check process
ps aux | grep protei_bulk

# Monitor connections
netstat -an | grep -E "8080|2775"

# System resources
htop
```

## Troubleshooting

### Build Errors

```bash
# Check C++ compiler version (needs C++20)
g++ --version  # Should be 11+

# Clean build
./build.sh --clean

# Verbose build
cd build
make VERBOSE=1
```

### Connection Errors

```bash
# Test PostgreSQL
psql -h localhost -U protei -d protei_bulk

# Test Redis
redis-cli ping

# Check if ports are available
sudo netstat -tlnp | grep -E "8080|2775|5432|6379"
```

### Performance Issues

```bash
# Check CPU usage
top -p $(pgrep protei_bulk)

# Check memory
ps aux | grep protei_bulk

# Check database connections
SELECT count(*) FROM pg_stat_activity WHERE datname='protei_bulk';

# Check Redis memory
redis-cli info memory
```

## Next Steps

1. **Read the full [README.md](README.md)** for comprehensive documentation
2. **Configure your environment** according to your requirements
3. **Set up monitoring** and alerting
4. **Load test** your configuration
5. **Deploy to production** using Docker or systemd

## Support

- Documentation: See [README.md](README.md)
- Issues: Report bugs in the issue tracker
- Performance: See [README.md](README.md#performance-benchmarks)

## Quick Reference

| Service | Port | Protocol |
|---------|------|----------|
| HTTP API | 8080 | HTTP/REST |
| SMPP | 2775 | SMPP 3.4 |
| PostgreSQL | 5432 | PostgreSQL |
| Redis | 6379 | Redis |

| Endpoint | Method | Description |
|----------|--------|-------------|
| /api/v1/health | GET | Health check |
| /api/v1/auth/login | POST | Authentication |
| /api/v1/messages/send | POST | Send message |
| /api/v1/campaigns | GET | List campaigns |

---

**Ready to scale? Check the full documentation for advanced features!**
