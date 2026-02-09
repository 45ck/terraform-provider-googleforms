#!/usr/bin/env bash
set -euo pipefail

THRESHOLD=60

# Check if gremlins is installed
if ! command -v gremlins &> /dev/null; then
    echo "gremlins not installed. Install with: go install github.com/go-gremlins/gremlins/cmd/gremlins@latest"
    echo "Skipping mutation testing."
    exit 0
fi

# Detect changed Go packages (compared to main branch)
CHANGED_PKGS=$(git diff --name-only origin/main...HEAD 2>/dev/null | \
    grep '\.go$' | \
    grep -v '_test\.go$' | \
    xargs -I{} dirname {} 2>/dev/null | \
    sort -u | \
    sed 's|^|./|' || echo "")

if [ -z "$CHANGED_PKGS" ]; then
    echo "No changed Go packages detected. Skipping mutation testing."
    exit 0
fi

echo "Running mutation tests on changed packages:"
echo "$CHANGED_PKGS"

for pkg in $CHANGED_PKGS; do
    echo ""
    echo "--- Mutating: $pkg ---"
    gremlins unleash "$pkg" --threshold "$THRESHOLD" || {
        echo "FAIL: Mutation score below ${THRESHOLD}% for $pkg"
        exit 1
    }
done

echo ""
echo "Mutation testing passed (threshold: ${THRESHOLD}%)"
