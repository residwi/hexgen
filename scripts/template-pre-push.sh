#!/bin/sh
# pre-push hook for the go-api-project-template checkout: when main is pushed,
# ports it into hexgen's skeleton with scripts/sync-template.sh. It never
# blocks the push. Install it from the template checkout:
#
#   ln -s ../../../hexgen/scripts/template-pre-push.sh .git/hooks/pre-push
#
# HEXGEN_DIR overrides the hexgen checkout (default: ../hexgen).

template=$(pwd)
hexgen=${HEXGEN_DIR:-$template/../hexgen}

while read -r _ local_sha remote_ref _; do
	[ "$remote_ref" = refs/heads/main ] || continue
	case $local_sha in *[!0]*) ;; *) continue ;; esac # main is being deleted
	# Git exports GIT_DIR to hooks; unset it so the sync's git commands act on hexgen.
	(unset GIT_DIR GIT_WORK_TREE GIT_INDEX_FILE && cd "$hexgen" &&
		scripts/sync-template.sh "$template" "$local_sha") ||
		echo "hexgen sync needs attention (see above): $hexgen"
done
exit 0
