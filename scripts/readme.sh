#!/usr/bin/env bash
# readme.sh: regenerates results/tables.md and the generated blocks of README.md from the
# saved results and prices, so no number in the README is typed by hand (constitution 1.3).
#
# Usage: scripts/readme.sh           rewrite the blocks
#        scripts/readme.sh --check   exit 1 if the README differs from what would be generated
set -euo pipefail
cd "$(dirname "$0")/.."
[ -f results/runs.tsv ] || { echo "readme: results/runs.tsv is missing; run scripts/measure.sh on the Mac first." >&2; exit 1; }
if [ "${1:-}" = "--check" ]; then
  go run ./cmd/tables -check
else
  go run ./cmd/tables
fi
