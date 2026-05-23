#!/usr/bin/env bash
# Prints the next version in YYYY.M.PATCH format.
# PATCH is the number of existing tags for the current year+month, starting at 0.
set -euo pipefail

YEAR=$(date +%Y)
MONTH=$(date +%-m)    # no leading zero

PREFIX="${YEAR}.${MONTH}."

# Count tags matching the prefix to determine next PATCH
PATCH=$(git tag --list "${PREFIX}*" 2>/dev/null | wc -l | tr -d ' ')

echo "${PREFIX}${PATCH}"
