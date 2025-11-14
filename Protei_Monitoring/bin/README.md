# bin/ - Application Binaries and Source Code

This directory contains the complete application source code and compiled binaries.

## Contents

### cmd/ - Main Application
```
cmd/protei-monitoring/
└── main.go          # Application entry point
```

**Features:**
- Service initialization
- Configuration loading
- All protocol decoders registration
- AI services integration
- Web server startup

### pkg/ - Application Packages

#### Protocol Decoders
```
pkg/decoder/
├── map/             # MAP (Mobile Application Part)
├── cap/             # CAP (CAMEL Application Part)
├── inap/            # INAP (Intelligent Network Application Part)
├── diameter/        # Diameter protocol
├── gtp/             # GTP (GPRS Tunneling Protocol)
├── pfcp/            # PFCP (Packet Forwarding Control Protocol)
├── http/            # HTTP/2 (5G SBI)
├── ngap/            # NGAP (5G RAN)
├── s1ap/            # S1AP (4G RAN)
└── nas/             # NAS (Non-Access Stratum - 4G/5G)
```

#### AI & Intelligence
```
pkg/
├── knowledge/       # Knowledge Base (18 standards)
├── analysis/        # AI Analysis Engine (7 detection rules)
├── flows/           # Flow Reconstructor (5 procedures)
└── correlation/     # Subscriber Correlator
```

#### Core Services
```
pkg/
├── web/             # Web server and API (35+ endpoints)
├── auth/            # Authentication and authorization
├── database/        # Database access layer
├── storage/         # CDR storage engine
├── capture/         # Packet capture engine
├── analytics/       # KPI and analytics
├── visualization/   # Ladder diagram generation
└── health/          # Health monitoring
```

#### Infrastructure
```
pkg/
├── config/          # Configuration management
├── license/         # License validation
└── dictionary/      # Protocol dictionaries
```

### internal/ - Internal Packages
```
internal/
└── logger/          # Structured logging (zerolog)
```

## Building from Source

### Prerequisites
```bash
# Go 1.21 or higher
go version

# Required dependencies
go mod download
```

### Build Application
```bash
# From repository root
cd /home/user/omar251990

# Build main application
go build -o protei-monitoring ./cmd/protei-monitoring

# Build with version information
go build -ldflags "-X main.version=2.0.0 -X main.buildDate=$(date -u +%Y-%m-%d)" \
  -o protei-monitoring ./cmd/protei-monitoring

# Build for production (optimized)
CGO_ENABLED=0 go build -ldflags "-s -w" -o protei-monitoring ./cmd/protei-monitoring
```

### Run Tests
```bash
# Test all packages
go test ./...

# Test specific package
go test ./pkg/decoder/map
go test ./pkg/analysis
go test ./pkg/knowledge

# Test with coverage
go test -cover ./...

# Test with verbose output
go test -v ./pkg/...
```

## Dependencies

### External Dependencies (go.mod)
```
github.com/rs/zerolog v1.31.0                    # Structured logging
gopkg.in/natefinch/lumberjack.v2 v2.2.1         # Log rotation
github.com/golang-jwt/jwt/v5 v5.2.0             # JWT authentication
golang.org/x/crypto v0.17.0                      # Password hashing (bcrypt)
gopkg.in/yaml.v3 v3.0.1                          # YAML configuration
github.com/lib/pq v1.10.9                        # PostgreSQL driver
github.com/gorilla/websocket v1.5.1              # WebSocket support
```

### Download Dependencies
```bash
# Download all dependencies
go mod download

# Verify dependencies
go mod verify

# Update dependencies (if needed)
go get -u ./...
go mod tidy
```

## Source Code Structure

### Application Flow

1. **Initialization** (main.go)
   - Load configuration from YAML
   - Validate license
   - Initialize database connection
   - Initialize Redis connection
   - Load protocol decoders
   - Initialize AI services
   - Start web server

2. **Packet Processing**
   - Capture packets via libpcap
   - Identify protocol
   - Decode using appropriate decoder
   - Store in database
   - Trigger AI analysis
   - Update subscriber correlation
   - Generate CDRs

3. **Web Interface**
   - Serve static files
   - Handle API requests
   - WebSocket for real-time updates
   - Authentication/Authorization
   - Session management

### Key Files

- **cmd/protei-monitoring/main.go** (735 lines)
  - Application struct
  - Service initialization
  - Graceful shutdown

- **pkg/web/server.go** (800+ lines)
  - 35+ API endpoints
  - Request handlers
  - Authentication middleware

- **pkg/knowledge/knowledge_base.go** (450+ lines)
  - 18 telecom standards
  - 14 error codes
  - 8 procedure references

- **pkg/analysis/analyzer.go** (600+ lines)
  - 7 detection rules
  - Pattern matching
  - Root cause analysis

## Security Notes

⚠️ **Source Code Protection**

In production deployments, this directory should contain:
- **Encrypted binaries** (AES-256-CBC)
- **Obfuscated source code** (if source is distributed)
- **Restricted file permissions** (chmod 750, root:protei)

### Production Security Checklist

```bash
# Set proper ownership
sudo chown -R root:protei bin/
sudo chmod 750 bin/

# Protect source code (optional - delete if only binary distribution)
sudo chmod 640 bin/cmd/ -R
sudo chmod 640 bin/pkg/ -R
sudo chmod 640 bin/internal/ -R

# Encrypt binary (production deployment script)
./scripts/utils/encrypt_binary.sh
```

## Development

### IDE Setup

**VS Code:**
```json
{
  "go.gopath": "/usr/protei/Protei_Monitoring",
  "go.inferGopath": true,
  "go.useLanguageServer": true,
  "go.lintTool": "golangci-lint",
  "go.formatTool": "gofmt"
}
```

**GoLand:**
- Set GOPATH to `/usr/protei/Protei_Monitoring`
- Enable Go modules support
- Set Go version to 1.21+

### Code Style

Follow standard Go conventions:
```bash
# Format code
go fmt ./...

# Run linter
golangci-lint run ./...

# Check for common issues
go vet ./...
```

## Troubleshooting

### Build Errors

**Error: go.sum missing**
```bash
# Solution: Generate go.sum
go mod tidy
```

**Error: Cannot find package**
```bash
# Solution: Download dependencies
go mod download
```

**Error: Embed pattern no matching files**
```bash
# Solution: Create required directories
mkdir -p pkg/web/static pkg/web/templates
echo "placeholder" > pkg/web/static/README.txt
```

### Runtime Errors

**Error: License validation failed**
- Check config/license.cfg
- Verify MAC address binding
- Ensure license hasn't expired

**Error: Database connection failed**
- Check config/db.cfg
- Verify PostgreSQL is running
- Test connection: `psql -h $DB_HOST -U $DB_USER -d $DB_NAME`

## Version Information

- **Go Version Required**: 1.21 or higher
- **Application Version**: 2.0.0
- **Build Type**: Production
- **Architecture Support**: linux/amd64, linux/arm64

---

For more information, see the main [README.md](../README.md) and [Developer Guide](../document/DEVELOPER_GUIDE.md).
