#!/bin/sh

# Compress or decompress the image directory after CRIU completes a dump.
set -eu

if [ "${CRTOOLS_SCRIPT_ACTION:-}" != "post-dump" ]; then
	exit 0
fi

: "${CRTOOLS_IMAGE_DIR:?CRTOOLS_IMAGE_DIR is required}"
: "${CRIT_COMPRESSION_ACTION:?CRIT_COMPRESSION_ACTION is required}"

script_dir=$(CDPATH='' cd -- "$(dirname -- "$0")" && pwd)
crit_bin=${CRIT_BIN:-"${script_dir}/../../crit/bin/crit"}

run_compression_action()
{
	operation=$1
	progress=$2
	output=$("$crit_bin" "$operation" --in-place "$CRTOOLS_IMAGE_DIR")
	printf '%s\n' "$output"
	if ! printf '%s\n' "$output" | grep -Fq "$progress checkpoint in "; then
		echo "crit $operation did not update the checkpoint" >&2
		exit 1
	fi
}

case "$CRIT_COMPRESSION_ACTION" in
decompress)
	run_compression_action decompress Decompressing
	;;
compress)
	run_compression_action compress Compressing
	;;
roundtrip)
	run_compression_action decompress Decompressing
	run_compression_action compress Compressing
	;;
*)
	echo "unsupported CRIT_COMPRESSION_ACTION: $CRIT_COMPRESSION_ACTION" >&2
	exit 1
	;;
esac
