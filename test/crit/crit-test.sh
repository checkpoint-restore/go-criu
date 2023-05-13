#!/bin/bash

set -x

CRIT=../../crit/bin/crit
TEST_IMG_DIR=test-imgs

function gen_img_list {
	images_list=$(find "$TEST_IMG_DIR" -regex '^[^\.]*\.img$')
	if [ -z "$images_list" ]; then
		echo "Failed to generate images"
		exit 1
	fi
}

function recode_test {
	for x in $images_list
	do
		echo "=== $x"
		if [[ $x == *pages* ]]; then
			echo "=== SKIP"
			continue
		fi

		echo "  -- to json"
		$CRIT decode -i "$x" -o "$x"".json" --pretty || exit $?
		echo "  -- to img"
		$CRIT encode -i "$x"".json" -o "$x"".json.img" || exit $?
		echo "  -- cmp"
		cmp "$x" "$x"".json.img" || exit $?

		echo "=== done"
	done
}

function command_test {
	PROTO_IN="$TEST_IMG_DIR"/inventory.img
	JSON_IN=$(mktemp -p "$TEST_IMG_DIR" tmp.XXXXXXXXXX.json)
	OUT=$(mktemp -p "$TEST_IMG_DIR" tmp.XXXXXXXXXX.img)

	# prepare
	$CRIT decode -i "$PROTO_IN" -o "$JSON_IN"

	# proto in - json out decode
	$CRIT decode -i "$PROTO_IN" || exit 1
	$CRIT decode -i "$PROTO_IN" -o "$OUT" || exit 1
	$CRIT decode -i "$PROTO_IN" > "$OUT" || exit 1
	$CRIT decode < "$PROTO_IN" || exit 1
	$CRIT decode -o "$OUT" < "$PROTO_IN" || exit 1
	$CRIT decode < "$PROTO_IN" > "$OUT" || exit 1

	# json in - proto out decode -> should fail
	$CRIT decode -i "$JSON_IN" || true
	$CRIT decode -i "$JSON_IN" -o "$OUT" || true
	$CRIT decode -i "$JSON_IN" > "$OUT" || true

	# json in - proto out encode
	$CRIT encode -i "$JSON_IN" || exit 1
	$CRIT encode -i "$JSON_IN" -o "$OUT" || exit 1
	$CRIT encode -i "$JSON_IN" > "$OUT" || exit 1
	$CRIT encode < "$JSON_IN" || exit 1
	$CRIT encode -o "$OUT" < "$JSON_IN" || exit 1
	$CRIT encode < "$JSON_IN" > "$OUT" || exit 1

	# proto in - json out encode -> should fail
	$CRIT encode -i "$PROTO_IN" || true
	$CRIT encode -i "$PROTO_IN" -o "$OUT" || true
	$CRIT encode -i "$PROTO_IN" > "$OUT" || true

	# test info and show commands
	$CRIT info "$PROTO_IN" || exit 1
	$CRIT show "$PROTO_IN" || exit 1

	# explore image directory
	$CRIT x "$TEST_IMG_DIR" ps || exit 1
	$CRIT x "$TEST_IMG_DIR" fds || exit 1
	$CRIT x "$TEST_IMG_DIR" mems || exit 1
	$CRIT x "$TEST_IMG_DIR" rss || exit 1
}

echo "Generating image list..."
gen_img_list
echo "Testing recode..."
recode_test
echo "Testing commands..."
command_test
