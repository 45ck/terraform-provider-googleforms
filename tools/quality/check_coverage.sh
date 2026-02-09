#!/usr/bin/env bash
set -euo pipefail

TOTAL_THRESHOLD=85
PACKAGE_THRESHOLD=75
PROFILE="coverprofile.txt"

if [ ! -f "$PROFILE" ]; then
    echo "Coverage profile not found. Running tests..."
    go test ./... -short -coverprofile="$PROFILE" -covermode=atomic
fi

# Check total coverage
TOTAL=$(go tool cover -func="$PROFILE" | grep "^total:" | awk '{print $3}' | tr -d '%')

if [ -z "$TOTAL" ]; then
    echo "WARN: Could not determine total coverage (no code to cover yet?)"
    exit 0
fi

echo "Total coverage: ${TOTAL}% (threshold: ${TOTAL_THRESHOLD}%)"

TOTAL_INT=$(echo "$TOTAL" | cut -d. -f1)
if [ "$TOTAL_INT" -lt "$TOTAL_THRESHOLD" ]; then
    echo "FAIL: Total coverage ${TOTAL}% is below threshold ${TOTAL_THRESHOLD}%"
    exit 1
fi

# Check per-package coverage
FAILED=0
go tool cover -func="$PROFILE" | grep -v "^total:" | \
    awk '{print $1}' | sort -u | while read -r pkg; do
    PKG_COV=$(go tool cover -func="$PROFILE" | grep "^${pkg}" | tail -1 | awk '{print $3}' | tr -d '%')
    if [ -n "$PKG_COV" ]; then
        PKG_INT=$(echo "$PKG_COV" | cut -d. -f1)
        if [ "$PKG_INT" -lt "$PACKAGE_THRESHOLD" ]; then
            echo "WARN: ${pkg} coverage ${PKG_COV}% is below threshold ${PACKAGE_THRESHOLD}%"
        fi
    fi
done

echo "Coverage checks passed."
