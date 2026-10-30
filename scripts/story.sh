#!/usr/bin/env bash
# Print one user story from product-spec.md.
# Usage: scripts/story.sh <n>     (for example: scripts/story.sh 3)
set -eu
[ $# -eq 1 ] || { echo "usage: scripts/story.sh <n>" >&2; exit 2; }
awk -v n="$1" '
  $0 ~ "^### US-" n "([^0-9]|$)" { p = 1; print; next }
  p && /^##/ { exit }
  p { print }
  END { if (!p) { print "user story US-" n " not found" > "/dev/stderr"; exit 1 } }
' "$(dirname "$0")/../product-spec.md"
