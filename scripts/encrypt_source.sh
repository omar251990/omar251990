#!/bin/bash
#
# Source Code Encryption Script for Protei_Monitoring
# This script encrypts all source code files to protect intellectual property
#
# Usage: ./encrypt_source.sh [encryption_key]
#

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
SOURCE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ENCRYPTED_DIR="${SOURCE_DIR}/encrypted_source"
KEY_FILE="${SOURCE_DIR}/.encryption_key"

# File extensions to encrypt
EXTENSIONS=("go" "yaml" "yml" "json" "sh" "html" "css" "js")

echo -e "${GREEN}=== Protei_Monitoring Source Code Encryption ===${NC}\n"

# Get encryption key
if [ -n "$1" ]; then
    ENCRYPTION_KEY="$1"
elif [ -f "$KEY_FILE" ]; then
    ENCRYPTION_KEY=$(cat "$KEY_FILE")
    echo -e "${YELLOW}Using stored encryption key${NC}"
else
    echo -e "${YELLOW}Enter encryption key (will be saved for decryption):${NC}"
    read -s ENCRYPTION_KEY
    echo
    echo -e "${YELLOW}Confirm encryption key:${NC}"
    read -s ENCRYPTION_KEY_CONFIRM
    echo

    if [ "$ENCRYPTION_KEY" != "$ENCRYPTION_KEY_CONFIRM" ]; then
        echo -e "${RED}Error: Keys do not match${NC}"
        exit 1
    fi

    # Save key for decryption
    echo "$ENCRYPTION_KEY" > "$KEY_FILE"
    chmod 600 "$KEY_FILE"
    echo -e "${GREEN}Encryption key saved to $KEY_FILE${NC}\n"
fi

# Create encrypted directory
rm -rf "$ENCRYPTED_DIR"
mkdir -p "$ENCRYPTED_DIR"

# Create metadata file
METADATA_FILE="${ENCRYPTED_DIR}/.encryption_metadata"
echo "encrypted_at=$(date -u +%Y-%m-%dT%H:%M:%SZ)" > "$METADATA_FILE"
echo "version=2.0.0" >> "$METADATA_FILE"
echo "algorithm=aes-256-cbc" >> "$METADATA_FILE"

echo -e "${GREEN}Encrypting source files...${NC}\n"

TOTAL_FILES=0
ENCRYPTED_FILES=0

# Function to encrypt a file
encrypt_file() {
    local src_file="$1"
    local rel_path="${src_file#$SOURCE_DIR/}"
    local dst_file="${ENCRYPTED_DIR}/${rel_path}.enc"

    # Create directory structure
    mkdir -p "$(dirname "$dst_file")"

    # Encrypt file using OpenSSL with AES-256-CBC
    if openssl enc -aes-256-cbc -salt -pbkdf2 -in "$src_file" -out "$dst_file" -k "$ENCRYPTION_KEY" 2>/dev/null; then
        echo "  ✓ $rel_path"
        ((ENCRYPTED_FILES++))
    else
        echo -e "  ${RED}✗ Failed: $rel_path${NC}"
    fi

    ((TOTAL_FILES++))
}

# Find and encrypt all source files
for ext in "${EXTENSIONS[@]}"; do
    while IFS= read -r -d '' file; do
        # Skip encrypted directory and vendor directories
        if [[ "$file" == *"/encrypted_source/"* ]] || \
           [[ "$file" == *"/vendor/"* ]] || \
           [[ "$file" == *"/.git/"* ]] || \
           [[ "$file" == *"/bin/"* ]] || \
           [[ "$file" == *"/out/"* ]]; then
            continue
        fi

        encrypt_file "$file"
    done < <(find "$SOURCE_DIR" -type f -name "*.${ext}" -print0)
done

# Create archive of encrypted files
ARCHIVE_NAME="protei_monitoring_encrypted_$(date +%Y%m%d_%H%M%S).tar.gz"
echo -e "\n${GREEN}Creating encrypted archive...${NC}"
cd "$ENCRYPTED_DIR"
tar -czf "../${ARCHIVE_NAME}" .
cd - > /dev/null

# Generate SHA256 checksum
CHECKSUM=$(sha256sum "${SOURCE_DIR}/${ARCHIVE_NAME}" | awk '{print $1}')
echo "$CHECKSUM" > "${SOURCE_DIR}/${ARCHIVE_NAME}.sha256"

echo -e "\n${GREEN}=== Encryption Summary ===${NC}"
echo -e "Total files processed:    $TOTAL_FILES"
echo -e "Successfully encrypted:   $ENCRYPTED_FILES"
echo -e "Encrypted directory:      $ENCRYPTED_DIR"
echo -e "Archive created:          ${ARCHIVE_NAME}"
echo -e "Checksum:                 ${CHECKSUM:0:16}..."
echo -e "\n${YELLOW}IMPORTANT:${NC}"
echo -e "1. Keep the encryption key ($KEY_FILE) secure and backed up"
echo -e "2. The encrypted archive is located at: ${SOURCE_DIR}/${ARCHIVE_NAME}"
echo -e "3. To decrypt, use: ./decrypt_source.sh [encryption_key]"
echo -e "\n${GREEN}Encryption completed successfully!${NC}"
