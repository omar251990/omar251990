#!/bin/bash
################################################################################
# Protei_Bulk - Quick Development Setup
# Fast setup script for development environment (without system dependencies)
################################################################################

set -e

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DB_USER="protei"
DB_PASSWORD="elephant"
DB_NAME="protei_bulk"
DB_HOST="localhost"
DB_PORT="5432"

echo -e "${BLUE}"
echo "╔════════════════════════════════════════════════════════════════╗"
echo "║                                                                ║"
echo "║          Protei_Bulk Quick Development Setup                  ║"
echo "║                                                                ║"
echo "╚════════════════════════════════════════════════════════════════╝"
echo -e "${NC}"
echo ""
echo "This script will:"
echo "  1. Create PostgreSQL database and user"
echo "  2. Load database schema"
echo "  3. Load seed data"
echo "  4. Set up Python virtual environment"
echo "  5. Install Python dependencies"
echo ""
echo "Prerequisites:"
echo "  • PostgreSQL installed and running"
echo "  • Python 3.8+ installed"
echo "  • Redis installed and running (optional)"
echo ""

read -p "Continue? [y/N] " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Setup cancelled."
    exit 0
fi

echo ""

# Function to check if PostgreSQL is running
check_postgres() {
    if ! pg_isready -h $DB_HOST -p $DB_PORT > /dev/null 2>&1; then
        echo -e "${YELLOW}Warning: PostgreSQL does not appear to be running${NC}"
        echo "Please start PostgreSQL and try again."
        exit 1
    fi
}

# Check PostgreSQL
echo "Checking PostgreSQL..."
check_postgres
echo -e "${GREEN}✓ PostgreSQL is running${NC}"

# Create database user and database
echo ""
echo "Setting up database..."

# Drop and recreate user and database
sudo -u postgres psql <<EOF > /dev/null 2>&1 || true
DROP DATABASE IF EXISTS $DB_NAME;
DROP USER IF EXISTS $DB_USER;
CREATE USER $DB_USER WITH PASSWORD '$DB_PASSWORD';
ALTER USER $DB_USER WITH SUPERUSER;
CREATE DATABASE $DB_NAME OWNER $DB_USER;
EOF

echo -e "${GREEN}✓ Database and user created${NC}"

# Load schema
echo ""
echo "Loading database schema..."
export PGPASSWORD="$DB_PASSWORD"
psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -f "$SCRIPT_DIR/database/schema.sql" > /dev/null 2>&1

if [ $? -eq 0 ]; then
    TABLE_COUNT=$(psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -t -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public';" 2>/dev/null | tr -d ' ')
    echo -e "${GREEN}✓ Schema loaded ($TABLE_COUNT tables created)${NC}"
else
    echo -e "${RED}✗ Failed to load schema${NC}"
    exit 1
fi

# Load seed data
echo ""
echo "Loading seed data..."
psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -f "$SCRIPT_DIR/database/seed_data.sql" > /dev/null 2>&1

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Seed data loaded${NC}"
else
    echo -e "${YELLOW}⚠ Seed data load had warnings (may be OK)${NC}"
fi

unset PGPASSWORD

# Update database configuration
echo ""
echo "Updating configuration files..."
sed -i.bak "s/^password = .*/password = $DB_PASSWORD/" "$SCRIPT_DIR/config/db.conf"
sed -i.bak "s/^username = .*/username = $DB_USER/" "$SCRIPT_DIR/config/db.conf"
sed -i.bak "s/^database = .*/database = $DB_NAME/" "$SCRIPT_DIR/config/db.conf"
echo -e "${GREEN}✓ Configuration updated${NC}"

# Set up Python virtual environment
echo ""
echo "Setting up Python virtual environment..."
VENV_DIR="$SCRIPT_DIR/venv"

if [ -d "$VENV_DIR" ]; then
    echo "Removing existing virtual environment..."
    rm -rf "$VENV_DIR"
fi

python3 -m venv "$VENV_DIR"
source "$VENV_DIR/bin/activate"

echo -e "${GREEN}✓ Virtual environment created${NC}"

# Install Python dependencies
echo ""
echo "Installing Python dependencies (this may take a few minutes)..."
pip install --upgrade pip > /dev/null 2>&1
pip install -r "$SCRIPT_DIR/requirements.txt" > /dev/null 2>&1

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Python dependencies installed${NC}"
else
    echo -e "${YELLOW}⚠ Some Python dependencies may have failed to install${NC}"
fi

deactivate

# Create necessary directories
echo ""
echo "Creating application directories..."
mkdir -p "$SCRIPT_DIR/logs"
mkdir -p "$SCRIPT_DIR/tmp/cache"
mkdir -p "$SCRIPT_DIR/tmp/parser"
mkdir -p "$SCRIPT_DIR/tmp/buffer"
mkdir -p "$SCRIPT_DIR/cdr/smpp"
mkdir -p "$SCRIPT_DIR/cdr/http"
mkdir -p "$SCRIPT_DIR/cdr/internal"
mkdir -p "$SCRIPT_DIR/cdr/archive"
echo -e "${GREEN}✓ Directories created${NC}"

# Summary
echo ""
echo -e "${GREEN}"
echo "╔════════════════════════════════════════════════════════════════╗"
echo "║                                                                ║"
echo "║          ✓ Development Setup Complete!                        ║"
echo "║                                                                ║"
echo "╚════════════════════════════════════════════════════════════════╝"
echo -e "${NC}"
echo ""
echo "Database Connection:"
echo "  Host:     $DB_HOST:$DB_PORT"
echo "  Database: $DB_NAME"
echo "  User:     $DB_USER"
echo "  Password: $DB_PASSWORD"
echo ""
echo "Default Credentials:"
echo "  Username: admin"
echo "  Password: Admin@123"
echo "  (CHANGE ON FIRST LOGIN)"
echo ""
echo "Next Steps:"
echo "  1. Activate virtual environment:"
echo "     source venv/bin/activate"
echo ""
echo "  2. Start the application:"
echo "     ./bin/Protei_Bulk"
echo ""
echo "  3. Or use management scripts:"
echo "     ./scripts/start"
echo "     ./scripts/status"
echo ""
echo "  4. Check logs:"
echo "     tail -f logs/startup.log"
echo ""
