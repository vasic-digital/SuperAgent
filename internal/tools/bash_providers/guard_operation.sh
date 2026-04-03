#!/usr/bin/env bash
# SPDX-FileCopyrightText: 2026 Milos Vasic
# SPDX-License-Identifier: Apache-2.0
#
# Guard operation - confirm dangerous operations

set -e

# List of dangerous commands/patterns
DANGEROUS_PATTERNS=(
    "rm -rf /"
    "rm -rf /*"
    "> /dev/sda"
    "mkfs"
    "dd if=/dev/zero"
    ":(){ :|:& };:"
    "chmod -R 777 /"
    "chown -R"
    "curl.*|.*sh"
    "wget.*|.*sh"
)

# Check if operation is dangerous
check_dangerous() {
    local cmd="$1"
    for pattern in "${DANGEROUS_PATTERNS[@]}"; do
        if [[ "$cmd" =~ $pattern ]]; then
            echo "ERROR: Dangerous operation detected: $pattern" >&2
            echo "Operation blocked for security." >&2
            exit 1
        fi
    done
}

# Confirm operation with user
confirm_operation() {
    local operation="$1"
    echo "This operation requires confirmation:"
    echo "  $operation"
    echo ""
    read -p "Proceed? [y/N] " response
    if [[ ! "$response" =~ ^[Yy]$ ]]; then
        echo "Operation cancelled."
        exit 1
    fi
}

# Main guard check
main() {
    # Get command from environment or arguments
    local cmd="${GUARD_COMMAND:-$*}"
    
    # Check for dangerous patterns
    check_dangerous "$cmd"
    
    # Check if confirmation required
    if [[ -n "$GUARD_CONFIRM" ]]; then
        confirm_operation "$cmd"
    fi
}

# Run guard if executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
