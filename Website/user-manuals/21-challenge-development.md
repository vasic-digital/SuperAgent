# User Manual 21: Challenge Development

## Overview
Guide for creating HelixAgent challenges.

## Challenge Structure
```bash
challenges/scripts/
├── challenge_name.sh
└── README.md
```

## Template
```bash
#!/bin/bash
set -e

CHALLENGE_NAME="Challenge Name"
POINTS=10

log_info() { echo "[INFO] $1"; }
log_success() { echo "[SUCCESS] $1"; }
log_error() { echo "[ERROR] $1"; }

test_1() {
    # Test implementation
}

main() {
    test_1
}

main "$@"
```

## Scoring
- Easy: 5-10 points
- Medium: 15-25 points
- Hard: 30-50 points
