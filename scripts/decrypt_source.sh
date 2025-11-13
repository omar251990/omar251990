#!/bin/bash
#
# Source Code Decryption Script for Protei_Monitoring
# This script decrypts encrypted source code files
#
# Usage: ./decrypt_source.sh [encryption_key] [encrypted_archive]
#

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
SOURCE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DECRYPTED_DIR="${SOURCE_DIR}/decrypted_source"
KEY_FILE="${SOURCE_DIR}/.encryption_key"

echo -e "${GREEN}=== Protei_Monitoring Source Code Decryption ===${NC}\n"

# Get encryption key
if [ -n "$1" ]; then
    ENCRYPTION_KEY="$1"
elif [ -f "$KEY_FILE" ]; then
    ENCRYPTION_KEY=$(cat "$KEY_FILE")
    echo -e "${YELLOW}Using stored encryption key${NC}"
else
    echo -e "${YELLOW}Enter decryption key:${NC}"
    read -s ENCRYPTION_KEY
    echo
fi

# Find encrypted archive
if [ -n "$2" ]; then
    ARCHIVE_FILE="$2"
elif [ -f "${SOURCE_DIR}/encrypted_source.tar.gz" ]; then
    ARCHIVE_FILE="${SOURCE_DIR}/encrypted_source.tar.gz"
else
    # Find most recent encrypted archive
    ARCHIVE_FILE=$(find "$SOURCE_DIR" -maxdepth 1 -name "protei_monitoring_encrypted_*.tar.gz" -type f -printf '%T@ %p\n' | sort -rn | head -1 | cut -d' ' -f2-)

    if [ -z "$ARCHIVE_FILE" ]; then
        echo -e "${RED}Error: No encrypted archive found${NC}"
        echo "Usage: $0 [encryption_key] [encrypted_archive]"
        exit 1
    fi
fi

echo -e "${GREEN}Using archive: $(basename "$ARCHIVE_FILE")${NC}"

# Verify checksum if available
CHECKSUM_FILE="${ARCHIVE_FILE}.sha256"
if [ -f "$CHECKSUM_FILE" ]; then
    echo -e "${YELLOW}Verifying archive integrity...${NC}"
    EXPECTED_CHECKSUM=$(cat "$CHECKSUM_FILE")
    ACTUAL_CHECKSUM=$(sha256sum "$ARCHIVE_FILE" | awk '{print $1}')

    if [ "$EXPECTED_CHECKSUM" = "$ACTUAL_CHECKSUM" ]; then
        echo -e "${GREEN}✓ Checksum verified${NC}\n"
    else
        echo -e "${RED}✗ Checksum mismatch! Archive may be corrupted.${NC}"
        echo -e "${YELLOW}Continue anyway? (yes/no)${NC}"
        read -r CONTINUE
        if [ "$CONTINUE" != "yes" ]; then
            exit 1
        fi
    fi
fi

# Create decrypted directory
rm -rf "$DECRYPTED_DIR"
mkdir -p "$DECRYPTED_DIR"

# Extract archive
echo -e "${GREEN}Extracting encrypted files...${NC}"
tar -xzf "$ARCHIVE_FILE" -C "$DECRYPTED_DIR"

# Read metadata
METADATA_FILE="${DECRYPTED_DIR}/.encryption_metadata"
if [ -f "$METADATA_FILE" ]; then
    echo -e "\n${YELLOW}Archive Information:${NC}"
    cat "$METADATA_FILE" | sed 's/^/  /'
    echo
fi

echo -e "${GREEN}Decrypting source files...${NC}\n"

TOTAL_FILES=0
DECRYPTED_FILES=0
FAILED_FILES=0

# Function to decrypt a file
decrypt_file() {
    local enc_file="$1"
    local rel_path="${enc_file#$DECRYPTED_DIR/}"
    local dst_file="${SOURCE_DIR}/${rel_path%.enc}"

    # Skip metadata
    if [[ "$enc_file" == *".encryption_metadata" ]]; then
        return
    fi

    # Create directory structure
    mkdir -p "$(dirname "$dst_file")"

    # Decrypt file using OpenSSL
    if openssl enc -aes-256-cbc -d -pbkdf2 -in "$enc_file" -out "$dst_file" -k "$ENCRYPTION_KEY" 2>/dev/null; then
        echo "  ✓ ${rel_path%.enc}"
        ((DECRYPTED_FILES++))

        # Restore executable permissions for scripts
        if [[ "$dst_file" == *.sh ]]; then
            chmod +x "$dst_file"
        fi
    else
        echo -e "  ${RED}✗ Failed: ${rel_path%.enc}${NC}"
        ((FAILED_FILES++))
    fi

    ((TOTAL_FILES++))
}

# Find and decrypt all encrypted files
while IFS= read -r -d '' file; do
    decrypt_file "$file"
done < <(find "$DECRYPTED_DIR" -type f -name "*.enc" -print0)

# Clean up decrypted directory
rm -rf "$DECRYPTED_DIR"

echo -e "\n${GREEN}=== Decryption Summary ===${NC}"
echo -e "Total files processed:    $TOTAL_FILES"
echo -e "Successfully decrypted:   $DECRYPTED_FILES"
echo -e "Failed:                   $FAILED_FILES"

if [ $FAILED_FILES -gt 0 ]; then
    echo -e "\n${RED}WARNING: Some files failed to decrypt. This may indicate:${NC}"
    echo -e "  - Incorrect decryption key"
    echo -e "  - Corrupted archive"
    echo -e "  - Modified encrypted files"
    exit 1
else
    echo -e "\n${GREEN}Decryption completed successfully!${NC}"
    echo -e "${YELLOW}Source code has been restored to: $SOURCE_DIR${NC}"
fi
