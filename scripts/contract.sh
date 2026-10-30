#!/usr/bin/env bash
# Read one part of openapi.yaml instead of the whole file.
# Usage: scripts/contract.sh list              paths and schemas with line numbers
#        scripts/contract.sh '/topics/{topicSlug}'   one path with its operations
#        scripts/contract.sh Error             one schema
set -eu
[ $# -eq 1 ] || { echo "usage: scripts/contract.sh list|<path>|<schema>" >&2; exit 2; }
f="$(dirname "$0")/../openapi.yaml"
if [ "$1" = list ]; then
  awk '/^paths:/{s="path"} /^components:/{s=""} /^  schemas:/{s="schema"}
       s=="path"   && /^  \//            { sub(/:$/,""); print NR": "$0 }
       s=="schema" && /^    [A-Za-z]+:$/ { sub(/:$/,""); print NR": "$0 }' "$f"
  exit 0
fi
awk -v want="$1" '
  /^paths:/ { in_paths = 1 } /^components:/ { in_paths = 0 } /^  schemas:/ { in_schemas = 1 }
  want ~ /^\// && in_paths {
    if ($0 == "  " want ":") { p = 1; print NR": "$0; next }
    if (p && /^  \//) exit
    if (p && /^[a-z]/) exit
    if (p) print NR": "$0
  }
  want !~ /^\// && in_schemas {
    if ($0 == "    " want ":") { p = 1; print NR": "$0; next }
    if (p && (/^    [A-Za-z]+:$/ || /^  [a-z]/ || /^[a-z]/)) exit
    if (p) print NR": "$0
  }
  END { if (!p) { print "not found: " want > "/dev/stderr"; exit 1 } }
' "$f"
