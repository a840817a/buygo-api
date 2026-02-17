#!/usr/bin/env bash

set -euo pipefail

MIN_COVERAGE="${1:-20}"
PROFILE_FILE="${2:-coverage.out}"

if [[ ! -f "$PROFILE_FILE" ]]; then
  echo "[coverage] missing profile: $PROFILE_FILE"
  exit 1
fi

TOTAL_COVERAGE="$(
  go tool cover -func="$PROFILE_FILE" \
    | awk '/^total:/{gsub("%","",$3); print $3}'
)"

if [[ -z "${TOTAL_COVERAGE}" ]]; then
  echo "[coverage] failed to parse total coverage from $PROFILE_FILE"
  exit 1
fi

echo "[coverage] backend total statements: ${TOTAL_COVERAGE}% (required >= ${MIN_COVERAGE}%)"

awk -v current="$TOTAL_COVERAGE" -v minimum="$MIN_COVERAGE" 'BEGIN { exit !(current + 0 >= minimum + 0) }'

echo "[coverage] backend gate passed"
