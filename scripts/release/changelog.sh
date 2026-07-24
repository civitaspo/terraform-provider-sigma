#!/bin/sh

set -eu

version=${1:?usage: changelog.sh VERSION [BASE_TAG]}
base_tag=${2:-}

if ! printf '%s\n' "$version" | grep -Eq '^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$'; then
  printf 'Invalid release version: %s\n' "$version" >&2
  exit 2
fi

if [ -n "$base_tag" ] && git rev-parse --verify --quiet "$base_tag^{commit}" >/dev/null; then
  range=$base_tag..HEAD
else
  range=HEAD
fi

work_dir=$(mktemp -d)
trap 'rm -rf "$work_dir"' EXIT HUP INT TERM

for group in features fixes documentation maintenance; do
  : >"$work_dir/$group"
done

git log "$range" --format='%s' | while IFS= read -r subject; do
  case "$subject" in
    chore\(release\):*)
      continue
      ;;
    feat:* | feat!:* | feat\(*\):* | feat\(*\)!:*)
      group=features
      ;;
    fix:* | fix!:* | fix\(*\):* | fix\(*\)!:*)
      group=fixes
      ;;
    docs:* | docs!:* | docs\(*\):* | docs\(*\)!:*)
      group=documentation
      ;;
    *)
      group=maintenance
      ;;
  esac

  title=$(printf '%s\n' "$subject" | sed 's/^[a-z][a-z]*\(([^)]*)\)\{0,1\}!\{0,1\}: *//')
  pull_request=$(printf '%s\n' "$title" | sed -n 's/.* \(#[0-9][0-9]*\)$/\1/p')
  if [ -n "$pull_request" ]; then
    title=${title% \("$pull_request"\)}
    printf -- '- %s (%s)\n' "$title" "$pull_request" >>"$work_dir/$group"
  else
    printf -- '- %s\n' "$title" >>"$work_dir/$group"
  fi
done

printf '## [%s] - %s\n' "$version" "$(date -u +%Y-%m-%d)"

for definition in \
  'features:Features' \
  'fixes:Bug Fixes' \
  'documentation:Documentation' \
  'maintenance:Maintenance'
do
  group=${definition%%:*}
  heading=${definition#*:}
  if [ -s "$work_dir/$group" ]; then
    printf '\n### %s\n\n' "$heading"
    while IFS= read -r entry; do
      printf '%s\n' "$entry"
    done <"$work_dir/$group"
  fi
done
