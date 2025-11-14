#!/bin/bash
#
# Protei Monitoring v2.0 - CDR Management Utility
#
# This script manages CDR files:
# - List CDR files by protocol and date
# - Compress uncompressed CDR files
# - Cleanup old CDR files
# - Generate CDR statistics
# - Export CDR data
#

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# CDR base directory
CDR_BASE_DIR="/usr/protei/Protei_Monitoring/cdr"

# Database configuration
DB_CONFIG="/usr/protei/Protei_Monitoring/config/db.cfg"

print_header() {
    echo -e "${BLUE}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "  Protei Monitoring - CDR Management"
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

# List CDR files for a protocol
list_cdr_files() {
    local protocol="${1:-}"
    local date_filter="${2:-}"

    print_header
    echo "CDR File Listing"
    echo ""

    if [ -z "$protocol" ]; then
        # List all protocols
        protocols=("MAP" "CAP" "INAP" "Diameter" "GTP" "PFCP" "HTTP2" "NGAP" "S1AP" "NAS")
    else
        protocols=("$protocol")
    fi

    for proto in "${protocols[@]}"; do
        local proto_dir="$CDR_BASE_DIR/$proto"

        if [ ! -d "$proto_dir" ]; then
            print_warning "Directory not found: $proto_dir"
            continue
        fi

        echo -e "${BLUE}Protocol: $proto${NC}"
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

        # Find files with optional date filter
        if [ -n "$date_filter" ]; then
            files=$(find "$proto_dir" -type f -name "*${date_filter}*" 2>/dev/null || true)
        else
            files=$(find "$proto_dir" -type f 2>/dev/null || true)
        fi

        if [ -z "$files" ]; then
            echo "  No CDR files found"
        else
            local total_size=0
            local file_count=0
            local compressed_count=0

            echo -e "${YELLOW}Filename${NC}\t\t\t\t\t${YELLOW}Size${NC}\t${YELLOW}Date${NC}"
            echo "────────────────────────────────────────────────────────────"

            while IFS= read -r file; do
                if [ -f "$file" ]; then
                    local filename=$(basename "$file")
                    local size=$(du -h "$file" | cut -f1)
                    local date=$(stat -c %y "$file" 2>/dev/null | cut -d' ' -f1)

                    echo -e "$filename\t$size\t$date"

                    # Count statistics
                    ((file_count++))
                    total_size=$((total_size + $(stat -c %s "$file" 2>/dev/null || echo 0)))

                    if [[ "$filename" == *.gz ]]; then
                        ((compressed_count++))
                    fi
                fi
            done <<< "$files"

            echo ""
            echo "Total files: $file_count"
            echo "Compressed: $compressed_count"
            echo "Total size: $(numfmt --to=iec-i --suffix=B $total_size)"
        fi
        echo ""
    done
}

# Compress CDR files older than specified days
compress_old_files() {
    local days="${1:-1}"
    local protocol="${2:-}"

    print_header
    echo "Compressing CDR files older than $days day(s)"
    echo ""

    if [ -z "$protocol" ]; then
        protocols=("MAP" "CAP" "INAP" "Diameter" "GTP" "PFCP" "HTTP2" "NGAP" "S1AP" "NAS")
    else
        protocols=("$protocol")
    fi

    local total_compressed=0
    local total_saved=0

    for proto in "${protocols[@]}"; do
        local proto_dir="$CDR_BASE_DIR/$proto"

        if [ ! -d "$proto_dir" ]; then
            continue
        fi

        print_info "Processing $proto..."

        # Find uncompressed files older than N days
        find "$proto_dir" -type f ! -name "*.gz" -mtime +$days 2>/dev/null | while read -r file; do
            local original_size=$(stat -c %s "$file")

            # Compress file
            if gzip -9 "$file" 2>/dev/null; then
                local compressed_file="${file}.gz"
                local compressed_size=$(stat -c %s "$compressed_file")
                local saved=$((original_size - compressed_size))

                print_success "Compressed: $(basename "$file") (saved $(numfmt --to=iec-i --suffix=B $saved))"

                ((total_compressed++))
                total_saved=$((total_saved + saved))
            else
                print_error "Failed to compress: $(basename "$file")"
            fi
        done
    done

    echo ""
    print_success "Compression complete!"
    echo "Files compressed: $total_compressed"
    echo "Space saved: $(numfmt --to=iec-i --suffix=B $total_saved)"
}

# Cleanup old CDR files
cleanup_old_files() {
    local retention_days="${1:-90}"
    local protocol="${2:-}"
    local dry_run="${3:-no}"

    print_header
    echo "Cleaning up CDR files older than $retention_days days"

    if [ "$dry_run" = "yes" ]; then
        echo "(DRY RUN - no files will be deleted)"
    fi

    echo ""

    if [ -z "$protocol" ]; then
        protocols=("MAP" "CAP" "INAP" "Diameter" "GTP" "PFCP" "HTTP2" "NGAP" "S1AP" "NAS")
    else
        protocols=("$protocol")
    fi

    local total_deleted=0
    local total_size_freed=0

    for proto in "${protocols[@]}"; do
        local proto_dir="$CDR_BASE_DIR/$proto"

        if [ ! -d "$proto_dir" ]; then
            continue
        fi

        print_info "Processing $proto..."

        # Find files older than retention period
        find "$proto_dir" -type f -mtime +$retention_days 2>/dev/null | while read -r file; do
            local file_size=$(stat -c %s "$file")

            if [ "$dry_run" = "yes" ]; then
                echo "Would delete: $(basename "$file") ($(numfmt --to=iec-i --suffix=B $file_size))"
            else
                if rm -f "$file" 2>/dev/null; then
                    print_success "Deleted: $(basename "$file")"
                    ((total_deleted++))
                    total_size_freed=$((total_size_freed + file_size))
                else
                    print_error "Failed to delete: $(basename "$file")"
                fi
            fi
        done
    done

    echo ""
    if [ "$dry_run" = "yes" ]; then
        print_info "Dry run complete - no files were deleted"
    else
        print_success "Cleanup complete!"
        echo "Files deleted: $total_deleted"
        echo "Space freed: $(numfmt --to=iec-i --suffix=B $total_size_freed)"
    fi
}

# Generate CDR statistics
generate_statistics() {
    local protocol="${1:-}"
    local start_date="${2:-}"
    local end_date="${3:-}"

    print_header
    echo "CDR Statistics"
    echo ""

    if [ -z "$protocol" ]; then
        protocols=("MAP" "CAP" "INAP" "Diameter" "GTP" "PFCP" "HTTP2" "NGAP" "S1AP" "NAS")
    else
        protocols=("$protocol")
    fi

    echo -e "${YELLOW}Protocol${NC}\t${YELLOW}Files${NC}\t${YELLOW}Total Size${NC}\t${YELLOW}Compressed${NC}\t${YELLOW}Oldest${NC}\t\t${YELLOW}Newest${NC}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

    for proto in "${protocols[@]}"; do
        local proto_dir="$CDR_BASE_DIR/$proto"

        if [ ! -d "$proto_dir" ]; then
            continue
        fi

        local file_count=$(find "$proto_dir" -type f 2>/dev/null | wc -l)

        if [ $file_count -eq 0 ]; then
            echo -e "$proto\t0\t-\t-\t-\t-"
            continue
        fi

        local total_size=$(du -sb "$proto_dir" 2>/dev/null | cut -f1)
        local compressed_count=$(find "$proto_dir" -type f -name "*.gz" 2>/dev/null | wc -l)
        local oldest=$(find "$proto_dir" -type f -printf '%T+ %p\n' 2>/dev/null | sort | head -n1 | cut -d' ' -f1 | cut -d'T' -f1)
        local newest=$(find "$proto_dir" -type f -printf '%T+ %p\n' 2>/dev/null | sort -r | head -n1 | cut -d' ' -f1 | cut -d'T' -f1)

        echo -e "$proto\t$file_count\t$(numfmt --to=iec-i --suffix=B $total_size)\t$compressed_count\t$oldest\t$newest"
    done

    echo ""

    # Overall statistics
    local total_files=$(find "$CDR_BASE_DIR" -type f 2>/dev/null | wc -l)
    local total_size=$(du -sb "$CDR_BASE_DIR" 2>/dev/null | cut -f1)
    local total_compressed=$(find "$CDR_BASE_DIR" -type f -name "*.gz" 2>/dev/null | wc -l)

    echo "Overall Statistics:"
    echo "  Total CDR files: $total_files"
    echo "  Total size: $(numfmt --to=iec-i --suffix=B $total_size)"
    echo "  Compressed files: $total_compressed"
    echo "  Compression ratio: $((total_compressed * 100 / (total_files > 0 ? total_files : 1)))%"
}

# Export CDR data to CSV
export_cdr_data() {
    local protocol="$1"
    local start_date="$2"
    local end_date="$3"
    local output_file="$4"

    print_header
    echo "Exporting CDR data"
    echo ""

    if [ -z "$protocol" ] || [ -z "$output_file" ]; then
        print_error "Protocol and output file required"
        return 1
    fi

    local proto_dir="$CDR_BASE_DIR/$protocol"

    if [ ! -d "$proto_dir" ]; then
        print_error "Protocol directory not found: $proto_dir"
        return 1
    fi

    print_info "Searching for CDR files..."

    # Find matching files
    local files=""
    if [ -n "$start_date" ] && [ -n "$end_date" ]; then
        files=$(find "$proto_dir" -type f -newermt "$start_date" ! -newermt "$end_date" 2>/dev/null || true)
    else
        files=$(find "$proto_dir" -type f 2>/dev/null || true)
    fi

    if [ -z "$files" ]; then
        print_warning "No CDR files found"
        return 0
    fi

    print_info "Exporting to: $output_file"

    # Combine all CSV files
    local first_file=true
    local record_count=0

    while IFS= read -r file; do
        if [[ "$file" == *.gz ]]; then
            # Decompress and read
            if [ "$first_file" = true ]; then
                zcat "$file" > "$output_file"
                first_file=false
            else
                # Skip header for subsequent files
                zcat "$file" | tail -n +2 >> "$output_file"
            fi
        elif [[ "$file" == *.csv ]]; then
            # Read CSV directly
            if [ "$first_file" = true ]; then
                cat "$file" > "$output_file"
                first_file=false
            else
                # Skip header for subsequent files
                tail -n +2 "$file" >> "$output_file"
            fi
        fi

        # Count records (excluding header)
        local lines=$(wc -l < "$file" 2>/dev/null || echo 0)
        record_count=$((record_count + lines - 1))
    done <<< "$files"

    print_success "Export complete!"
    echo "Records exported: $record_count"
    echo "Output file: $output_file"
}

# Verify CDR file integrity
verify_cdr_files() {
    local protocol="${1:-}"

    print_header
    echo "Verifying CDR file integrity"
    echo ""

    if [ -z "$protocol" ]; then
        protocols=("MAP" "CAP" "INAP" "Diameter" "GTP" "PFCP" "HTTP2" "NGAP" "S1AP" "NAS")
    else
        protocols=("$protocol")
    fi

    local total_files=0
    local corrupted_files=0

    for proto in "${protocols[@]}"; do
        local proto_dir="$CDR_BASE_DIR/$proto"

        if [ ! -d "$proto_dir" ]; then
            continue
        fi

        print_info "Verifying $proto files..."

        find "$proto_dir" -type f -name "*.gz" 2>/dev/null | while read -r file; do
            ((total_files++))

            if ! gzip -t "$file" 2>/dev/null; then
                print_error "Corrupted: $(basename "$file")"
                ((corrupted_files++))
            fi
        done
    done

    echo ""
    if [ $corrupted_files -eq 0 ]; then
        print_success "All CDR files verified successfully!"
    else
        print_warning "$corrupted_files out of $total_files files are corrupted"
    fi
}

# Usage information
usage() {
    cat <<EOF
Usage: $0 <command> [options]

Commands:
    list [protocol] [date]           - List CDR files
    compress <days> [protocol]       - Compress files older than N days
    cleanup <days> [protocol] [dry]  - Delete files older than N days
    stats [protocol]                 - Generate statistics
    export <protocol> <output>       - Export CDR data to CSV
    verify [protocol]                - Verify file integrity

Examples:
    # List all MAP CDR files
    $0 list MAP

    # List CDR files for specific date
    $0 list MAP 20240101

    # Compress files older than 7 days
    $0 compress 7

    # Compress only GTP files
    $0 compress 7 GTP

    # Cleanup files older than 90 days (dry run)
    $0 cleanup 90 "" yes

    # Cleanup files older than 90 days (actual deletion)
    $0 cleanup 90

    # Generate statistics for all protocols
    $0 stats

    # Generate statistics for Diameter only
    $0 stats Diameter

    # Export Diameter CDR data
    $0 export Diameter /tmp/diameter_export.csv

    # Verify all compressed files
    $0 verify

    # Verify only MAP files
    $0 verify MAP

EOF
}

# Main
main() {
    local command="${1:-}"

    case "$command" in
        list)
            list_cdr_files "$2" "$3"
            ;;
        compress)
            if [ -z "$2" ]; then
                print_error "Days parameter required"
                usage
                exit 1
            fi
            compress_old_files "$2" "$3"
            ;;
        cleanup)
            if [ -z "$2" ]; then
                print_error "Days parameter required"
                usage
                exit 1
            fi
            cleanup_old_files "$2" "$3" "$4"
            ;;
        stats)
            generate_statistics "$2" "$3" "$4"
            ;;
        export)
            if [ -z "$2" ] || [ -z "$3" ]; then
                print_error "Protocol and output file required"
                usage
                exit 1
            fi
            export_cdr_data "$2" "" "" "$3"
            ;;
        verify)
            verify_cdr_files "$2"
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
