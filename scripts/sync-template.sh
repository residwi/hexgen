#!/usr/bin/env bash
# Ports go-api-project-template commits made since the last sync into the
# embedded skeleton. See "Syncing from the template" in README.md.
#
# Usage: scripts/sync-template.sh [template-checkout] [rev]
#   defaults: ../go-api-project-template, its HEAD
set -euo pipefail

root=$(git rev-parse --show-toplevel)
src=$(cd "${1:-$root/../go-api-project-template}" && pwd)
mark=$root/internal/template/UPSTREAM
td=internal/template/testdata
from=$(cat "$mark")
to=$(git -C "$src" rev-parse "${2:-HEAD}")

if [ "$from" = "$to" ]; then
	echo "Already synced at $to."
	exit 0
fi
if [ -n "$(git -C "$root" status --porcelain)" ]; then
	echo "hexgen has uncommitted changes. Commit or stash them, then run make sync-template." >&2
	exit 1
fi
git -C "$src" log --oneline "$from..$to"
echo

# The template's identity and shop words, as the skeleton spells them.
neutralize='s#github.com/residwi/go-api-project-template#__MODULE__#g; s#ecommerce-api#__PROJECT_NAME__#g; s#"customer"#"user"#g'

# target prints where a template path lives in the skeleton, or nothing when
# hexgen keeps no copy of it (a shop module, a doc, a migration it reshaped).
# A file new upstream goes into an existing internal/ package.
target() {
	for t in "$td/auth/$1" "$td/skeleton/$1" "$td/skeleton/$1.tmpl"; do
		[ -e "$root/$t" ] && { echo "$t"; return; }
	done
	case $1 in internal/*) ;; *) return ;; esac
	for t in "$td/skeleton/$1" "$td/auth/$1"; do
		[ -d "$root/$(dirname "$t")" ] && { echo "$t"; return; }
	done
}

for f in $(git -C "$src" diff --no-renames --name-only "$from..$to"); do
	case $f in
	go.mod)
		{ echo 'module {{.Module}}'; git -C "$src" show "$to:go.mod" | sed 1d; } >"$root/$td/skeleton/go.mod.tmpl"
		continue ;;
	go.sum)
		git -C "$src" show "$to:go.sum" >"$root/$td/skeleton/go.sum"
		continue ;;
	esac
	t=$(target "$f") || true
	if [ -z "$t" ]; then
		echo "skipped, no copy in hexgen: $f"
		continue
	fi
	git -C "$src" diff --no-renames "$from..$to" -- "$f" |
		sed -E -e "/^(diff --git|--- a\/|\+\+\+ b\/)/s#([ab])/$f#\1/$t#g" -e "$neutralize" |
		(cd "$root" && git apply --reject) || true
done

echo "$to" >"$mark"

rejects=$(cd "$root" && git ls-files --others --exclude-standard -- "$td" | grep '\.rej$' || true)
if [ -n "$rejects" ]; then
	echo
	echo "Hunks that did not apply (port each by hand, then delete the .rej):"
	printf '  %s\n' $rejects
	exit 1
fi
echo "Synced to $to. Next: make test, then commit."
