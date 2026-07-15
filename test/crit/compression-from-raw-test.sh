#!/bin/sh

# Verify that CRIU restores an initially raw checkpoint compressed by Go CRIT.
set -eu

: "${CRIU_TREE:?CRIU_TREE is required}"
: "${CRIU_BIN:?CRIU_BIN is required}"
: "${CRIT_BIN:?CRIT_BIN is required}"

script_dir=$(CDPATH='' cd -- "$(dirname -- "$0")" && pwd)
zdtm_dir=${CRIU_TREE}/test/zdtm/static
test_bin=${zdtm_dir}/pages_content02
work_dir=$(mktemp -d)
image_dir=${work_dir}/images
pid=
restored_pid=

cleanup()
{
	for task in "$restored_pid" "$pid"; do
		if [ -n "$task" ]; then
			pkill -KILL -P "$task" 2>/dev/null || true
			kill -KILL "$task" 2>/dev/null || true
		fi
	done
	rm -rf "$work_dir"
}
trap cleanup EXIT
trap 'exit 1' HUP INT TERM

make -C "$zdtm_dir" pages_content02
mkdir "$image_dir"
(
	cd "$work_dir"
	"$test_bin" \
		--pidfile=test.pid \
		--outfile=test.out \
		--filename=test.file
)

pid=$(cat "$work_dir/test.pid")
if ! kill -0 "$pid" 2>/dev/null; then
	echo "pages_content02 failed to start" >&2
	cat "$work_dir/test.out" >&2
	exit 1
fi

if ! "$CRIU_BIN" dump --no-default-config -v4 \
		-o dump.log -D "$image_dir" -t "$pid" --shell-job; then
	echo "CRIU failed to create the uncompressed checkpoint" >&2
	cat "$image_dir/dump.log" >&2
	exit 1
fi
pid=

CRTOOLS_SCRIPT_ACTION=post-dump \
	CRTOOLS_IMAGE_DIR="$image_dir" \
	CRIT_COMPRESSION_ACTION=compress \
	CRIT_BIN="$CRIT_BIN" \
	"$script_dir/compression-action.sh"

if ! "$CRIU_BIN" restore --no-default-config -v4 \
		-o restore.log -D "$image_dir" --shell-job -d \
		--pidfile "$work_dir/restored.pid"; then
	echo "CRIU failed to restore the Go-compressed checkpoint" >&2
	cat "$image_dir/restore.log" >&2
	exit 1
fi

restored_pid=$(cat "$work_dir/restored.pid")
if ! kill -TERM "$restored_pid" 2>/dev/null; then
	echo "restored pages_content02 process is not running" >&2
	exit 1
fi

verified=
i=0
while [ "$i" -lt 100 ]; do
	if grep -q PASS "$work_dir/test.out" 2>/dev/null; then
		verified=1
		break
	fi
	if ! kill -0 "$restored_pid" 2>/dev/null; then
		break
	fi
	sleep 0.1
	i=$((i + 1))
done

if [ -z "$verified" ]; then
	echo "pages_content02 did not verify its restored memory" >&2
	cat "$work_dir/test.out" >&2
	exit 1
fi
echo "pages_content02 verified its restored memory"
restored_pid=
