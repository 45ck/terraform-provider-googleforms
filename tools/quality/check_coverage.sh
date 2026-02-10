#!/usr/bin/env bash
set -euo pipefail

# check_coverage.sh enforces coverage in a way that's practical for providers:
# - A hard "total repo coverage" percentage is not meaningful when many packages
#   are thin Terraform-framework glue and/or have no tests yet.
# - Instead we enforce stricter minimums on the core packages that carry most of
#   the logic and risk.
#
# The workflow can be tightened over time by raising thresholds or adding more
# critical packages.

PROFILE="${PROFILE:-coverprofile.txt}"

# Optional total threshold (0 disables). Defaults to 0 since "total" includes
# packages with no tests and punishes new packages unfairly.
TOTAL_THRESHOLD="${TOTAL_THRESHOLD:-0}"

# Format: "<package>=<min_percent_int>"
CRITICAL_PKGS=(
  "./internal/provider=85"
  "./internal/convert=60"
  "./internal/resource_form=65"
  "./internal/client=20"
  "./internal/resource_sheet_values=20"
  "./internal/resource_sheets_batch_update=15"
  "./internal/resource_drive_permission=10"
)

ensure_profile() {
  if [ -f "$PROFILE" ]; then
    return 0
  fi
  echo "Coverage profile not found. Running tests to generate it..."
  go test ./... -short -coverprofile="$PROFILE" -covermode=atomic
}

extract_total_pct() {
  local prof="$1"
  go tool cover -func="$prof" | awk '/^total:/{print $3}' | tr -d '%'
}

ensure_profile

if [ "$TOTAL_THRESHOLD" != "0" ]; then
  TOTAL="$(extract_total_pct "$PROFILE")"
  if [ -z "$TOTAL" ]; then
    echo "WARN: could not determine total coverage"
    exit 0
  fi
  echo "Total coverage: ${TOTAL}% (threshold: ${TOTAL_THRESHOLD}%)"
  TOTAL_INT="$(echo "$TOTAL" | cut -d. -f1)"
  if [ "$TOTAL_INT" -lt "$TOTAL_THRESHOLD" ]; then
    echo "FAIL: total coverage ${TOTAL}% is below threshold ${TOTAL_THRESHOLD}%"
    exit 1
  fi
fi

fail=0
for entry in "${CRITICAL_PKGS[@]}"; do
  pkg="${entry%%=*}"
  min="${entry##*=}"
  tmp="coverprofile.$(echo "$pkg" | tr '/.' '__').txt"

  # Generate a per-package profile so the threshold reflects that package only.
  go test "$pkg" -short -coverprofile="$tmp" -covermode=atomic >/dev/null

  cov="$(extract_total_pct "$tmp")"
  if [ -z "$cov" ]; then
    echo "FAIL: could not determine coverage for $pkg"
    fail=1
    continue
  fi

  cov_int="$(echo "$cov" | cut -d. -f1)"
  echo "Coverage $pkg: ${cov}% (min: ${min}%)"
  if [ "$cov_int" -lt "$min" ]; then
    echo "FAIL: $pkg coverage ${cov}% is below minimum ${min}%"
    fail=1
  fi
done

if [ "$fail" -ne 0 ]; then
  exit 1
fi

echo "Coverage checks passed."

