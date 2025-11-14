#!/bin/bash
#
# Protei Monitoring v2.0 - License Generation Tool
#
# This script generates a license file with MAC-based HMAC signature
#

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Vendor secret (must match the one in license_mac.go)
VENDOR_SECRET="PROTEI_MONITORING_VENDOR_SECRET_KEY_2025"

print_header() {
    echo -e "${BLUE}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "  Protei Monitoring - License Generator"
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

# Function to normalize MAC address
normalize_mac() {
    local mac="$1"
    # Convert to lowercase and replace - or . with :
    mac=$(echo "$mac" | tr '[:upper:]' '[:lower:]' | tr '-' ':' | tr '.' ':')
    echo "$mac"
}

# Function to calculate HMAC-SHA256 signature
calculate_signature() {
    local customer_name="$1"
    local expiry_date="$2"
    local licensed_mac="$3"
    local enable_2g="$4"
    local enable_3g="$5"
    local enable_4g="$6"
    local enable_5g="$7"
    local enable_map="$8"
    local enable_cap="$9"
    local enable_inap="${10}"
    local enable_diameter="${11}"
    local enable_http="${12}"
    local enable_gtp="${13}"
    local max_subscribers="${14}"
    local max_tps="${15}"

    # Normalize MAC
    licensed_mac=$(normalize_mac "$licensed_mac")

    # Build canonical string (fields sorted alphabetically)
    local canonical_string=""
    canonical_string+="customer_name=$customer_name|"
    canonical_string+="enable_2g=$enable_2g|"
    canonical_string+="enable_3g=$enable_3g|"
    canonical_string+="enable_4g=$enable_4g|"
    canonical_string+="enable_5g=$enable_5g|"
    canonical_string+="enable_cap=$enable_cap|"
    canonical_string+="enable_diameter=$enable_diameter|"
    canonical_string+="enable_gtp=$enable_gtp|"
    canonical_string+="enable_http=$enable_http|"
    canonical_string+="enable_inap=$enable_inap|"
    canonical_string+="enable_map=$enable_map|"
    canonical_string+="expiry_date=$expiry_date|"
    canonical_string+="licensed_mac=$licensed_mac|"
    canonical_string+="max_subscribers=$max_subscribers|"
    canonical_string+="max_tps=$max_tps"

    # Calculate HMAC-SHA256
    local signature=$(echo -n "$canonical_string" | openssl dgst -sha256 -hmac "$VENDOR_SECRET" | awk '{print $2}')

    echo "$signature"
}

# Interactive license generation
generate_license_interactive() {
    print_header

    # Collect license information
    echo "Please provide license details:"
    echo ""

    read -p "Customer Name: " customer_name
    read -p "License Expiry Date (YYYY-MM-DD): " expiry_date
    read -p "Server MAC Address (XX:XX:XX:XX:XX:XX): " licensed_mac

    echo ""
    echo "Generation Support (1=enabled, 0=disabled):"
    read -p "  Enable 2G [1]: " enable_2g; enable_2g=${enable_2g:-1}
    read -p "  Enable 3G [1]: " enable_3g; enable_3g=${enable_3g:-1}
    read -p "  Enable 4G [1]: " enable_4g; enable_4g=${enable_4g:-1}
    read -p "  Enable 5G [1]: " enable_5g; enable_5g=${enable_5g:-1}

    echo ""
    echo "Protocol Support (1=enabled, 0=disabled):"
    read -p "  Enable MAP [1]: " enable_map; enable_map=${enable_map:-1}
    read -p "  Enable CAP [1]: " enable_cap; enable_cap=${enable_cap:-1}
    read -p "  Enable INAP [1]: " enable_inap; enable_inap=${enable_inap:-1}
    read -p "  Enable Diameter [1]: " enable_diameter; enable_diameter=${enable_diameter:-1}
    read -p "  Enable HTTP [1]: " enable_http; enable_http=${enable_http:-1}
    read -p "  Enable GTP [1]: " enable_gtp; enable_gtp=${enable_gtp:-1}

    echo ""
    echo "Capacity Limits:"
    read -p "  Max Subscribers [5000000]: " max_subscribers; max_subscribers=${max_subscribers:-5000000}
    read -p "  Max TPS [5000]: " max_tps; max_tps=${max_tps:-5000}

    echo ""
    read -p "Output license file path [./license.cfg]: " output_file
    output_file=${output_file:-./license.cfg}

    # Generate license
    generate_license_file "$customer_name" "$expiry_date" "$licensed_mac" \
        "$enable_2g" "$enable_3g" "$enable_4g" "$enable_5g" \
        "$enable_map" "$enable_cap" "$enable_inap" "$enable_diameter" "$enable_http" "$enable_gtp" \
        "$max_subscribers" "$max_tps" "$output_file"
}

# Generate license file
generate_license_file() {
    local customer_name="$1"
    local expiry_date="$2"
    local licensed_mac="$3"
    local enable_2g="$4"
    local enable_3g="$5"
    local enable_4g="$6"
    local enable_5g="$7"
    local enable_map="$8"
    local enable_cap="$9"
    local enable_inap="${10}"
    local enable_diameter="${11}"
    local enable_http="${12}"
    local enable_gtp="${13}"
    local max_subscribers="${14}"
    local max_tps="${15}"
    local output_file="${16}"

    print_info "Generating license..."

    # Normalize MAC
    licensed_mac=$(normalize_mac "$licensed_mac")

    # Calculate signature
    local signature=$(calculate_signature "$customer_name" "$expiry_date" "$licensed_mac" \
        "$enable_2g" "$enable_3g" "$enable_4g" "$enable_5g" \
        "$enable_map" "$enable_cap" "$enable_inap" "$enable_diameter" "$enable_http" "$enable_gtp" \
        "$max_subscribers" "$max_tps")

    # Create license file
    cat > "$output_file" <<EOF
# ============================================================================
# Protei Monitoring v2.0 - License File
# ============================================================================
# Generated: $(date '+%Y-%m-%d %H:%M:%S')
# Customer: $customer_name
# ============================================================================

[license]
customer_name = $customer_name
expiry_date   = $expiry_date

# Bound MAC address (normalized, lowercase)
licensed_mac  = $licensed_mac

# Generation Support
enable_2g     = $enable_2g
enable_3g     = $enable_3g
enable_4g     = $enable_4g
enable_5g     = $enable_5g

# Protocol Support
enable_map      = $enable_map
enable_cap      = $enable_cap
enable_inap     = $enable_inap
enable_diameter = $enable_diameter
enable_http     = $enable_http
enable_gtp      = $enable_gtp

# Capacity Limits
max_subscribers = $max_subscribers
max_tps         = $max_tps

# HMAC-SHA256 signature (DO NOT MODIFY)
# Calculated over all above fields with vendor secret key
signature = $signature

# ============================================================================
# End of License File
# ============================================================================
EOF

    chmod 600 "$output_file"

    print_success "License file generated: $output_file"
    echo ""

    # Display license info
    print_license_info "$customer_name" "$expiry_date" "$licensed_mac" \
        "$enable_2g" "$enable_3g" "$enable_4g" "$enable_5g" \
        "$enable_map" "$enable_cap" "$enable_inap" "$enable_diameter" "$enable_http" "$enable_gtp" \
        "$max_subscribers" "$max_tps" "$signature"
}

# Print license information
print_license_info() {
    echo -e "${GREEN}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "  License Information"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo -e "${NC}"
    echo "Customer:       $1"
    echo "Expiry Date:    $2"
    echo "Licensed MAC:   $3"
    echo ""
    echo "Generation Support:"
    echo "  2G: $([ "$4" = "1" ] && echo "✅ Enabled" || echo "❌ Disabled")"
    echo "  3G: $([ "$5" = "1" ] && echo "✅ Enabled" || echo "❌ Disabled")"
    echo "  4G: $([ "$6" = "1" ] && echo "✅ Enabled" || echo "❌ Disabled")"
    echo "  5G: $([ "$7" = "1" ] && echo "✅ Enabled" || echo "❌ Disabled")"
    echo ""
    echo "Protocol Support:"
    echo "  MAP:      $([ "$8" = "1" ] && echo "✅ Enabled" || echo "❌ Disabled")"
    echo "  CAP:      $([ "$9" = "1" ] && echo "✅ Enabled" || echo "❌ Disabled")"
    echo "  INAP:     $([ "${10}" = "1" ] && echo "✅ Enabled" || echo "❌ Disabled")"
    echo "  Diameter: $([ "${11}" = "1" ] && echo "✅ Enabled" || echo "❌ Disabled")"
    echo "  HTTP:     $([ "${12}" = "1" ] && echo "✅ Enabled" || echo "❌ Disabled")"
    echo "  GTP:      $([ "${13}" = "1" ] && echo "✅ Enabled" || echo "❌ Disabled")"
    echo ""
    echo "Capacity:"
    echo "  Max Subscribers: ${14}"
    echo "  Max TPS:         ${15}"
    echo ""
    echo "Signature: ${16:0:16}...${16: -16}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
}

# Example license generation (for testing)
generate_example_license() {
    print_info "Generating example license for testing..."

    # Get current system MAC address
    local mac=$(ip link show | grep 'link/ether' | head -1 | awk '{print $2}')

    if [ -z "$mac" ]; then
        print_error "Could not detect system MAC address"
        return 1
    fi

    generate_license_file \
        "Example Customer" \
        "2026-12-31" \
        "$mac" \
        "1" "1" "1" "1" \
        "1" "1" "1" "1" "1" "1" \
        "5000000" \
        "5000" \
        "./license_example.cfg"
}

# Main menu
main() {
    if [ "$1" = "--example" ]; then
        generate_example_license
    elif [ "$1" = "--auto" ] && [ $# -ge 16 ]; then
        # Automated generation with parameters
        generate_license_file "$2" "$3" "$4" "$5" "$6" "$7" "$8" "$9" "${10}" "${11}" "${12}" "${13}" "${14}" "${15}" "${16}" "${17}"
    else
        generate_license_interactive
    fi
}

# Run
main "$@"

exit 0
