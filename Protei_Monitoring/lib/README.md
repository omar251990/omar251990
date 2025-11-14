# lib/ - External Libraries and Dependencies

This directory contains all external libraries and dependencies required by Protei Monitoring.

## Contents

When the application is built, Go dependencies are compiled into the binary.
This directory may contain:

- Go module cache (when building)
- Third-party libraries
- Shared objects (.so files)
- Native libraries (libpcap, etc.)

## Go Dependencies

See `bin/go.mod` for the complete list of dependencies:

- **github.com/rs/zerolog** - Structured logging
- **gopkg.in/natefinch/lumberjack.v2** - Log rotation
- **github.com/golang-jwt/jwt/v5** - JWT authentication
- **golang.org/x/crypto** - Cryptographic functions
- **gopkg.in/yaml.v3** - YAML configuration parsing
- **github.com/lib/pq** - PostgreSQL driver
- **github.com/gorilla/websocket** - WebSocket support

## System Libraries

Required system libraries:

- **libpcap** - Packet capture library
- **PostgreSQL client libraries** - Database connectivity
- **OpenSSL** - TLS/SSL support

Install on RHEL/CentOS:
```bash
sudo yum install libpcap libpq openssl
```

Install on Ubuntu/Debian:
```bash
sudo apt install libpcap0.8 libpq5 openssl
```

## License Information

All dependencies are used in compliance with their respective licenses.
See individual package documentation for license details.

