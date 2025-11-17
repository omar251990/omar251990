# Protei_Bulk C++ Edition

High-Performance Enterprise Bulk Messaging Platform - C++20 Implementation

## Overview

Protei_Bulk C++ is a complete rewrite of the Python-based Protei_Bulk messaging platform using modern C++20. This implementation delivers significantly higher performance, lower memory footprint, and better resource utilization for enterprise-scale bulk messaging operations.

## Key Features

### Core Capabilities
- **Multi-Channel Messaging**: SMS/SMPP, WhatsApp, Email, Viber, RCS, Voice, Push Notifications
- **High Performance**: 10,000+ TPS sustained throughput
- **SMPP Protocol**: Full SMPP 3.3, 3.4, and 5.0 support
- **REST API**: Complete RESTful API for all operations
- **Real-time Analytics**: Sub-second dashboards and reporting
- **Connection Pooling**: Optimized PostgreSQL and Redis connection pools
- **Multi-Tenancy**: Full customer isolation and resource management

### Advanced Features
- Dynamic Campaign Data Loader (DCDL)
- Subscriber Profiling with hashing
- Advanced Segmentation Engine
- Multi-SMSC Routing (7 condition types)
- A/B Testing Framework
- Customer Journey Automation
- GDPR & PDPL Compliance

## Architecture

```
Protei_Bulk_CPP/
├── CMakeLists.txt          # Build configuration
├── README.md               # This file
├── LICENSE                 # MIT License
├── src/                    # Source files
│   ├── main.cpp           # Application entry point
│   ├── core/              # Core infrastructure
│   │   ├── config.cpp
│   │   ├── database.cpp
│   │   ├── redis_client.cpp
│   │   └── logger.cpp
│   ├── api/               # HTTP API server
│   │   ├── http_server.cpp
│   │   ├── api_router.cpp
│   │   └── middleware.cpp
│   ├── services/          # Business logic
│   │   ├── routing_service.cpp
│   │   ├── campaign_service.cpp
│   │   ├── profiling_service.cpp
│   │   ├── segmentation_service.cpp
│   │   └── dcdl_service.cpp
│   ├── protocols/         # Protocol implementations
│   │   ├── smpp_server.cpp
│   │   ├── smpp_client.cpp
│   │   ├── smpp_pdu.cpp
│   │   └── whatsapp_client.cpp
│   ├── models/            # Data models
│   │   ├── message.cpp
│   │   ├── campaign.cpp
│   │   └── user.cpp
│   └── utils/             # Utilities
│       ├── crypto.cpp
│       ├── validators.cpp
│       └── string_utils.cpp
├── include/protei/        # Header files (mirror src/)
├── tests/                 # Unit and integration tests
├── config/                # Configuration files
│   ├── app.conf
│   ├── db.conf
│   ├── protocol.conf
│   └── security.conf
├── docker/                # Docker deployment
│   ├── Dockerfile
│   └── docker-compose.yml
└── docs/                  # Documentation
    └── api/
```

## Technology Stack

### Core Libraries
- **C++ Standard**: C++20
- **Build System**: CMake 3.20+
- **HTTP Server**: cpp-httplib 0.14+
- **JSON Processing**: nlohmann/json 3.11+
- **Database**: libpqxx (PostgreSQL C++ client)
- **Cache**: redis-plus-plus (Redis C++ client)
- **Logging**: spdlog 1.12+
- **Async I/O**: Boost.Asio 1.75+
- **Testing**: Google Test 1.14+

### System Requirements
- **OS**: Linux (Ubuntu 20.04+, CentOS 8+, Debian 11+)
- **Compiler**: GCC 11+ or Clang 14+ (C++20 support required)
- **RAM**: Minimum 4GB, Recommended 16GB+
- **CPU**: Minimum 4 cores, Recommended 16+ cores
- **Storage**: 100GB+ for logs and CDR data

## Dependencies Installation

### Ubuntu/Debian
```bash
# Update package list
sudo apt-get update

# Install build tools
sudo apt-get install -y \
    build-essential \
    cmake \
    git \
    pkg-config

# Install C++ dependencies
sudo apt-get install -y \
    libboost-all-dev \
    libpqxx-dev \
    libhiredis-dev \
    libssl-dev \
    zlib1g-dev

# Install redis-plus-plus
git clone https://github.com/sewenew/redis-plus-plus.git
cd redis-plus-plus
mkdir build && cd build
cmake ..
make
sudo make install
cd ../..

# Install PostgreSQL client libraries
sudo apt-get install -y postgresql-client libpq-dev

# Install Redis
sudo apt-get install -y redis-server
```

### CentOS/RHEL
```bash
# Install EPEL repository
sudo yum install -y epel-release

# Install build tools
sudo yum install -y \
    gcc-c++ \
    cmake3 \
    git \
    pkg-config

# Install dependencies
sudo yum install -y \
    boost-devel \
    libpqxx-devel \
    hiredis-devel \
    openssl-devel \
    zlib-devel \
    postgresql-devel

# Build redis-plus-plus from source (see above)
```

## Building from Source

### Quick Build
```bash
# Clone the repository
git clone https://github.com/yourorg/Protei_Bulk_CPP.git
cd Protei_Bulk_CPP

# Create build directory
mkdir build && cd build

# Configure with CMake
cmake .. -DCMAKE_BUILD_TYPE=Release

# Build (use all available cores)
make -j$(nproc)

# Run tests
ctest

# Install
sudo make install
```

### Build Types

#### Release Build (Optimized)
```bash
cmake .. -DCMAKE_BUILD_TYPE=Release
make -j$(nproc)
```

#### Debug Build
```bash
cmake .. -DCMAKE_BUILD_TYPE=Debug
make -j$(nproc)
```

#### With Address Sanitizer (Memory leak detection)
```bash
cmake .. -DCMAKE_BUILD_TYPE=Debug \
         -DCMAKE_CXX_FLAGS="-fsanitize=address"
make -j$(nproc)
```

## Configuration

### Configuration Files

Create configuration directory:
```bash
sudo mkdir -p /opt/protei_bulk/config
sudo mkdir -p /opt/protei_bulk/logs
```

#### app.conf
```ini
[Application]
app_name = Protei_Bulk
version = 1.0.0
environment = production

[Runtime]
max_workers = 10
queue_size = 10000

[Performance]
enable_monitoring = true
```

#### db.conf
```ini
[PostgreSQL]
host = localhost
port = 5432
database = protei_bulk
username = protei
password = elephant
pool_size = 20
max_connections = 50

[Redis]
enabled = true
host = localhost
port = 6379
password =
database = 0
pool_size = 10
```

#### protocol.conf
```ini
[SMPP]
enabled = true
bind_address = 0.0.0.0
bind_port = 2775
system_id = PROTEI_BULK
max_connections = 100
enquire_link_interval = 30

[HTTP]
enabled = true
bind_address = 0.0.0.0
bind_port = 8080
enable_https = false
thread_pool_size = 8
```

## Running the Application

### Standard Execution
```bash
# Run with default config location
./protei_bulk

# Run with custom config
./protei_bulk /path/to/custom/config/app.conf
```

### Using Environment Variables
```bash
# Database settings
export DB_HOST=localhost
export DB_PORT=5432
export DB_NAME=protei_bulk
export DB_USER=protei
export DB_PASSWORD=elephant

# Redis settings
export REDIS_HOST=localhost
export REDIS_PORT=6379
export REDIS_DB=0

# Application settings
export APP_ENV=production
export LOG_LEVEL=info

# Run
./protei_bulk
```

### As a System Service

Create `/etc/systemd/system/protei_bulk.service`:
```ini
[Unit]
Description=Protei_Bulk Messaging Platform
After=network.target postgresql.service redis.service

[Service]
Type=simple
User=protei
WorkingDirectory=/opt/protei_bulk
ExecStart=/usr/local/bin/protei_bulk /opt/protei_bulk/config/app.conf
Restart=always
RestartSec=10

# Environment variables
Environment="DB_HOST=localhost"
Environment="DB_PORT=5432"
Environment="DB_NAME=protei_bulk"
Environment="DB_USER=protei"
Environment="DB_PASSWORD=elephant"
Environment="REDIS_HOST=localhost"
Environment="REDIS_PORT=6379"

[Install]
WantedBy=multi-user.target
```

Enable and start:
```bash
sudo systemctl daemon-reload
sudo systemctl enable protei_bulk
sudo systemctl start protei_bulk
sudo systemctl status protei_bulk
```

## Docker Deployment

### Build Docker Image
```bash
cd docker
docker build -t protei_bulk_cpp:latest .
```

### Run with Docker Compose
```bash
docker-compose up -d
```

### Docker Environment
```yaml
version: '3.8'
services:
  protei_bulk_cpp:
    image: protei_bulk_cpp:latest
    ports:
      - "8080:8080"
      - "2775:2775"
    environment:
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_NAME=protei_bulk
      - DB_USER=protei
      - DB_PASSWORD=elephant
      - REDIS_HOST=redis
      - REDIS_PORT=6379
    depends_on:
      - postgres
      - redis
```

## API Endpoints

### Health Check
```bash
curl http://localhost:8080/api/v1/health
```

### Authentication
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"Admin@123"}'
```

### Send SMS
```bash
curl -X POST http://localhost:8080/api/v1/messages/send \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "msisdn": "966500000000",
    "message": "Hello from Protei_Bulk C++",
    "sender_id": "ProteiApp"
  }'
```

### Create Campaign
```bash
curl -X POST http://localhost:8080/api/v1/campaigns \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Marketing Campaign",
    "message": "Special offer!",
    "schedule_time": "2025-01-20T10:00:00Z",
    "contact_list_id": 1
  }'
```

## Performance Benchmarks

### Target Specifications
- **Throughput**: 10,000 TPS sustained
- **Latency**: <5ms average API response time
- **Messages/sec**: 5,000+ delivered messages
- **Memory**: <2GB RAM for 10K TPS
- **CPU**: 80% efficiency on multi-core systems

### Load Testing
```bash
# Using Apache Bench
ab -n 100000 -c 1000 http://localhost:8080/api/v1/health

# Using wrk
wrk -t12 -c400 -d30s http://localhost:8080/api/v1/health
```

## Testing

### Run All Tests
```bash
cd build
ctest --output-on-failure
```

### Run Specific Test Suite
```bash
./tests/protei_tests --gtest_filter=DatabaseTest.*
```

### Coverage Report
```bash
cmake .. -DCMAKE_BUILD_TYPE=Debug -DENABLE_COVERAGE=ON
make -j$(nproc)
make coverage
```

## Monitoring & Logging

### Log Files
- Main log: `/opt/protei_bulk/logs/protei_bulk.log`
- Error log: `/opt/protei_bulk/logs/protei_bulk_error.log`
- Access log: `/opt/protei_bulk/logs/access.log`

### Log Levels
- TRACE: Detailed debugging
- DEBUG: Debug information
- INFO: General information
- WARN: Warning messages
- ERROR: Error messages
- CRITICAL: Critical errors

### Performance Metrics
Monitor these metrics in production:
- Active connections (HTTP + SMPP)
- Messages per second
- Database pool utilization
- Redis cache hit rate
- CPU and memory usage
- Queue depths

## Troubleshooting

### Common Issues

#### Build Fails
```bash
# Verify C++20 support
g++ --version  # Should be 11+

# Check CMake version
cmake --version  # Should be 3.20+

# Clean build
rm -rf build && mkdir build && cd build
cmake .. && make clean && make -j$(nproc)
```

#### Connection Issues
```bash
# Test PostgreSQL
psql -h localhost -U protei -d protei_bulk

# Test Redis
redis-cli ping

# Check ports
sudo netstat -tlnp | grep -E "8080|2775"
```

#### Performance Issues
```bash
# Check CPU affinity
taskset -c -p <pid>

# Monitor in real-time
htop
iotop

# Check database performance
EXPLAIN ANALYZE SELECT ...;
```

## Contributing

### Code Style
- C++20 standard features
- Google C++ Style Guide
- Use `clang-format` for formatting
- Document all public APIs

### Pull Request Process
1. Fork the repository
2. Create feature branch
3. Write tests
4. Ensure all tests pass
5. Submit pull request

## License

MIT License - see LICENSE file for details

## Support

- Documentation: https://docs.protei-bulk.com
- Issues: https://github.com/yourorg/Protei_Bulk_CPP/issues
- Email: support@protei-bulk.com

## Roadmap

### Version 1.1
- WebSocket support for real-time updates
- Advanced caching strategies
- Enhanced security features
- Kubernetes deployment

### Version 1.2
- Machine learning integration
- Predictive analytics
- Auto-scaling capabilities
- Cloud-native deployment options

---

**Built with ❤️ using Modern C++20**
