#!/bin/bash
#
# Protei Monitoring v2.0 - Source Code Encryption Tool
#
# This script encrypts Go source code files using AES-256-CBC
# for IP protection in production deployments
#

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Encryption settings
CIPHER="aes-256-cbc"
PBKDF2_ITER=100000

print_header() {
    echo -e "${BLUE}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "  Protei Monitoring - Source Code Encryption"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo -e "${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_error() {
    echo -e "${RED}❌ ERROR: $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  WARNING: $1${NC}"
}

# Encrypt a single file
encrypt_file() {
    local input_file="$1"
    local output_file="$2"
    local password="$3"

    if [ ! -f "$input_file" ]; then
        print_error "Input file not found: $input_file"
        return 1
    fi

    # Encrypt using OpenSSL with AES-256-CBC
    if openssl enc -$CIPHER -salt -pbkdf2 -iter $PBKDF2_ITER -in "$input_file" -out "$output_file" -k "$password" 2>/dev/null; then
        local input_size=$(stat -f%z "$input_file" 2>/dev/null || stat -c%s "$input_file" 2>/dev/null)
        local output_size=$(stat -f%z "$output_file" 2>/dev/null || stat -c%s "$output_file" 2>/dev/null)

        print_success "Encrypted: $input_file → $output_file ($input_size → $output_size bytes)"
        return 0
    else
        print_error "Failed to encrypt: $input_file"
        return 1
    fi
}

# Decrypt a single file
decrypt_file() {
    local input_file="$1"
    local output_file="$2"
    local password="$3"

    if [ ! -f "$input_file" ]; then
        print_error "Input file not found: $input_file"
        return 1
    fi

    # Decrypt using OpenSSL
    if openssl enc -$CIPHER -d -pbkdf2 -iter $PBKDF2_ITER -in "$input_file" -out "$output_file" -k "$password" 2>/dev/null; then
        print_success "Decrypted: $input_file → $output_file"
        return 0
    else
        print_error "Failed to decrypt: $input_file (wrong password?)"
        return 1
    fi
}

# Encrypt entire directory recursively
encrypt_directory() {
    local source_dir="$1"
    local dest_dir="$2"
    local password="$3"
    local extensions="${4:-.go}" # Default to .go files

    if [ ! -d "$source_dir" ]; then
        print_error "Source directory not found: $source_dir"
        return 1
    fi

    mkdir -p "$dest_dir"

    local count=0
    local failed=0

    print_info "Encrypting files in: $source_dir"
    print_info "Output directory: $dest_dir"
    print_info "File extensions: $extensions"
    echo ""

    # Find all matching files
    while IFS= read -r -d '' file; do
        # Get relative path
        local rel_path="${file#$source_dir/}"
        local dest_file="$dest_dir/$rel_path.enc"

        # Create destination directory
        mkdir -p "$(dirname "$dest_file")"

        # Encrypt file
        if encrypt_file "$file" "$dest_file" "$password"; then
            ((count++))
        else
            ((failed++))
        fi
    done < <(find "$source_dir" -type f \( -name "*$extensions" \) -print0)

    echo ""
    print_success "Encryption complete: $count files encrypted, $failed failed"

    if [ $failed -gt 0 ]; then
        return 1
    fi
    return 0
}

# Decrypt entire directory recursively
decrypt_directory() {
    local source_dir="$1"
    local dest_dir="$2"
    local password="$3"

    if [ ! -d "$source_dir" ]; then
        print_error "Source directory not found: $source_dir"
        return 1
    fi

    mkdir -p "$dest_dir"

    local count=0
    local failed=0

    print_info "Decrypting files in: $source_dir"
    print_info "Output directory: $dest_dir"
    echo ""

    # Find all .enc files
    while IFS= read -r -d '' file; do
        # Get relative path and remove .enc extension
        local rel_path="${file#$source_dir/}"
        local dest_file="$dest_dir/${rel_path%.enc}"

        # Create destination directory
        mkdir -p "$(dirname "$dest_file")"

        # Decrypt file
        if decrypt_file "$file" "$dest_file" "$password"; then
            ((count++))
        else
            ((failed++))
        fi
    done < <(find "$source_dir" -type f -name "*.enc" -print0)

    echo ""
    print_success "Decryption complete: $count files decrypted, $failed failed"

    if [ $failed -gt 0 ]; then
        return 1
    fi
    return 0
}

# Generate random password
generate_password() {
    local length=${1:-32}

    # Generate random password using /dev/urandom
    local password=$(openssl rand -base64 $length | tr -d "=+/" | cut -c1-$length)

    echo "$password"
}

# Interactive encryption
interactive_encrypt() {
    print_header

    echo "Source Code Encryption"
    echo ""

    read -p "Source directory: " source_dir
    read -p "Destination directory: " dest_dir

    # Check if password should be generated
    read -p "Generate random password? (y/n) [y]: " gen_pwd
    gen_pwd=${gen_pwd:-y}

    if [ "$gen_pwd" = "y" ] || [ "$gen_pwd" = "Y" ]; then
        password=$(generate_password 32)
        print_success "Generated password: $password"
        echo ""
        print_warning "SAVE THIS PASSWORD! You'll need it to decrypt the source code."
        echo ""
        read -p "Press Enter to continue..."
    else
        read -s -p "Enter encryption password: " password
        echo ""
        read -s -p "Confirm password: " password2
        echo ""

        if [ "$password" != "$password2" ]; then
            print_error "Passwords do not match!"
            return 1
        fi
    fi

    read -p "File extensions to encrypt [.go]: " extensions
    extensions=${extensions:-.go}

    echo ""
    read -p "Proceed with encryption? (y/n): " confirm
    if [ "$confirm" != "y" ] && [ "$confirm" != "Y" ]; then
        print_info "Cancelled."
        return 0
    fi

    # Perform encryption
    if encrypt_directory "$source_dir" "$dest_dir" "$password" "$extensions"; then
        echo ""
        print_success "Encryption successful!"

        if [ "$gen_pwd" = "y" ] || [ "$gen_pwd" = "Y" ]; then
            # Save password to file
            local pwd_file="$dest_dir/.encryption_key"
            echo "$password" > "$pwd_file"
            chmod 600 "$pwd_file"
            print_warning "Password saved to: $pwd_file (protect this file!)"
        fi
    else
        print_error "Encryption failed!"
        return 1
    fi
}

# Interactive decryption
interactive_decrypt() {
    print_header

    echo "Source Code Decryption"
    echo ""

    read -p "Encrypted directory: " source_dir
    read -p "Destination directory: " dest_dir

    # Check for saved password
    local pwd_file="$source_dir/.encryption_key"
    if [ -f "$pwd_file" ]; then
        read -p "Use saved password from $pwd_file? (y/n) [y]: " use_saved
        use_saved=${use_saved:-y}

        if [ "$use_saved" = "y" ] || [ "$use_saved" = "Y" ]; then
            password=$(cat "$pwd_file")
            print_info "Using saved password"
        else
            read -s -p "Enter decryption password: " password
            echo ""
        fi
    else
        read -s -p "Enter decryption password: " password
        echo ""
    fi

    echo ""
    read -p "Proceed with decryption? (y/n): " confirm
    if [ "$confirm" != "y" ] && [ "$confirm" != "Y" ]; then
        print_info "Cancelled."
        return 0
    fi

    # Perform decryption
    if decrypt_directory "$source_dir" "$dest_dir" "$password"; then
        echo ""
        print_success "Decryption successful!"
    else
        print_error "Decryption failed! Check password and try again."
        return 1
    fi
}

# Usage information
usage() {
    cat <<EOF
Usage: $0 <command> [options]

Commands:
    encrypt         - Interactive encryption
    decrypt         - Interactive decryption
    encrypt-dir     - Encrypt directory (automated)
    decrypt-dir     - Decrypt directory (automated)
    generate-pwd    - Generate random password

Examples:
    # Interactive encryption
    $0 encrypt

    # Interactive decryption
    $0 decrypt

    # Automated encryption
    $0 encrypt-dir /path/to/source /path/to/encrypted "mypassword" ".go"

    # Automated decryption
    $0 decrypt-dir /path/to/encrypted /path/to/source "mypassword"

    # Generate password
    $0 generate-pwd 32

Encryption details:
    Cipher: AES-256-CBC
    Key derivation: PBKDF2 with 100,000 iterations
    Salt: Random (OpenSSL default)

EOF
}

# Main
main() {
    local command="${1:-}"

    case "$command" in
        encrypt)
            interactive_encrypt
            ;;
        decrypt)
            interactive_decrypt
            ;;
        encrypt-dir)
            if [ $# -lt 4 ]; then
                usage
                exit 1
            fi
            encrypt_directory "$2" "$3" "$4" "${5:-.go}"
            ;;
        decrypt-dir)
            if [ $# -lt 4 ]; then
                usage
                exit 1
            fi
            decrypt_directory "$2" "$3" "$4"
            ;;
        generate-pwd)
            local len="${2:-32}"
            pwd=$(generate_password $len)
            echo "Generated password: $pwd"
            ;;
        -h|--help|help)
            usage
            ;;
        *)
            print_error "Unknown command: $command"
            echo ""
            usage
            exit 1
            ;;
    esac
}

# Run
main "$@"

exit 0
