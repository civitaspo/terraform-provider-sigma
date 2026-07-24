#!/bin/sh

set -eu

base_tag=${1:-}
if [ -z "$base_tag" ]; then
  base_tag=$(git describe --tags --abbrev=0 --match 'v[0-9]*' 2>/dev/null || printf '%s\n' v0.0.0)
fi

if ! printf '%s\n' "$base_tag" | grep -Eq '^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$'; then
  printf 'Invalid release tag: %s\n' "$base_tag" >&2
  exit 2
fi

version=${base_tag#v}
major=${version%%.*}
remainder=${version#*.}
minor=${remainder%%.*}
patch=${remainder#*.}

if git rev-parse --verify --quiet "$base_tag^{commit}" >/dev/null; then
  commits=$(git log "$base_tag"..HEAD --format='%B%x1e')
else
  # No real tag yet: consider the full history reachable from HEAD.
  commits=$(git log HEAD --format='%B%x1e')
fi
releasable=$(printf '%s\n' "$commits" | awk '
  BEGIN { RS = "\036" }
  {
    first = $0
    sub(/\n.*/, "", first)
    if (first != "" && first !~ /^chore\(release\):/) {
      found = 1
    }
  }
  END { if (found) print "yes" }
')

if [ -z "$releasable" ]; then
  exit 0
fi

breaking=$(printf '%s\n' "$commits" | awk '
  BEGIN { RS = "\036" }
  {
    first = $0
    sub(/\n.*/, "", first)
    if (first ~ /^[a-z]+(\([^)]*\))?!:/ || $0 ~ /(^|\n)BREAKING CHANGE:/) {
      found = 1
    }
  }
  END { if (found) print "yes" }
')

feature=$(printf '%s\n' "$commits" | awk '
  BEGIN { RS = "\036" }
  {
    first = $0
    sub(/\n.*/, "", first)
    if (first ~ /^feat(\([^)]*\))?:/) {
      found = 1
    }
  }
  END { if (found) print "yes" }
')

if [ -n "$breaking" ]; then
  if [ "$major" -eq 0 ]; then
    minor=$((minor + 1))
    patch=0
  else
    major=$((major + 1))
    minor=0
    patch=0
  fi
elif [ -n "$feature" ]; then
  minor=$((minor + 1))
  patch=0
else
  patch=$((patch + 1))
fi

printf '%s.%s.%s\n' "$major" "$minor" "$patch"
